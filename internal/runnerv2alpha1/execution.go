package runnerv2alpha1

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/stateful/runme/internal/command"
	"github.com/stateful/runme/internal/document"
	"github.com/stateful/runme/internal/document/identity"
	runnerv2alpha1 "github.com/stateful/runme/internal/gen/proto/go/runme/runner/v2alpha1"
	"github.com/stateful/runme/internal/project"
	"github.com/stateful/runme/internal/rbuffer"
)

const (
	// ringBufferSize limits the size of the ring buffers
	// that sit between a command and the handler.
	ringBufferSize = 8192 << 10 // 8 MiB

	// msgBufferSize limits the size of data chunks
	// sent by the handler to clients. It's smaller
	// intentionally as typically the messages are
	// small.
	// In the future, it might be worth to implement
	// variable-sized buffers.
	msgBufferSize = 2048 << 10 // 2 MiB
)

type execution struct {
	ID      string
	Block   *document.CodeBlock
	Project *project.Project
	Req     *runnerv2alpha1.ExecuteRequest

	Cmd *command.VirtualCommand

	stdin       io.Reader
	stdinWriter io.WriteCloser
	stdout      *rbuffer.RingBuffer

	logger *zap.Logger
}

func newExecutionFrom(ctx context.Context, id string, req *runnerv2alpha1.ExecuteRequest, logger *zap.Logger) (*execution, error) {
	proj, err := getProjectFromRequest(req, logger)
	if err != nil {
		return nil, err
	}

	block, err := getCodeBlockFromRequest(ctx, req, proj)
	if err != nil {
		return nil, err
	}

	var (
		stdin       io.Reader
		stdinWriter io.WriteCloser
	)

	if req.Interactive {
		stdin, stdinWriter = io.Pipe()
	}

	stdout := rbuffer.NewRingBuffer(ringBufferSize)

	cmd, err := command.NewVirtual(
		block,
		&command.VirtualCommandOptions{
			Stdin:  stdin,
			Stdout: stdout,
			Logger: logger,
		},
	)
	if err != nil {
		return nil, err
	}

	exec := &execution{
		ID:      id,
		Block:   block,
		Project: proj,
		Req:     req,

		Cmd: cmd,

		stdin:       stdin,
		stdinWriter: stdinWriter,
		stdout:      stdout,

		logger: logger,
	}

	return exec, nil
}

func (e *execution) Wait(ctx context.Context, sender sender) (int, error) {
	exitCode := -1

	// Write initial input data.
	if len(e.Req.InputData) > 0 {
		if e.stdinWriter == nil {
			e.logger.Warn("input data provided but stdin is not available")
		} else {
			if _, err := e.stdinWriter.Write(e.Req.InputData); err != nil {
				return exitCode, errors.WithStack(err)
			}
		}
	}

	waitErr := e.Cmd.Wait()
	exitCode = exitCodeFromErr(waitErr)

	e.closeIO()

	errc := make(chan error, 1)
	go func() {
		errc <- readSendLoop(e.stdout, sender)
	}()

	// If waitErr is not nil, only log the errors but return waitErr.
	if waitErr != nil {
		select {
		case err := <-errc:
			e.logger.Info("readSendLoop finished; ignoring any errors because there was a wait error", zap.Error(err))
		case <-ctx.Done():
			e.logger.Info("context canceled while waiting for the readSendLoop finish; ignoring any errors because there was a wait error")
		}
		return exitCode, waitErr
	}

	// If waitErr is nil, wait for the readSendLoop to finish,
	// or the context being canceled.
	select {
	case err := <-errc:
		return exitCode, err
	case <-ctx.Done():
		return exitCode, ctx.Err()
	}
}

func (e *execution) Write(p []byte) (int, error) {
	return e.stdinWriter.Write(p)
}

func (e *execution) closeIO() {
	var err error

	if e.stdinWriter != nil {
		err = e.stdinWriter.Close()
		e.logger.Debug("closed stdin writer", zap.Error(err))
	}

	err = e.stdout.Close()
	e.logger.Debug("closed stdout writer", zap.Error(err))
}

type sender interface {
	Send(*runnerv2alpha1.ExecuteResponse) error
}

func readSendLoop(reader io.Reader, sender sender) error {
	limitedReader := io.LimitReader(reader, msgBufferSize)

	for {
		buf := make([]byte, msgBufferSize)
		n, err := limitedReader.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return errors.WithStack(err)
		}
		if n == 0 {
			continue
		}

		err = sender.Send(&runnerv2alpha1.ExecuteResponse{StdoutData: buf[:n]})
		if err != nil {
			return errors.WithStack(err)
		}
	}
}

func getProjectFromRequest(req *runnerv2alpha1.ExecuteRequest, logger *zap.Logger) (*project.Project, error) {
	var (
		proj *project.Project
		err  error
	)

	if req.Project != nil {
		proj, err = getDirProjectFromRequest(req, logger)
	} else if path := req.DocumentPath; path != "" {
		proj, err = getFileProjectFromRequest(req, logger)
	}

	return proj, err
}

func getDirProjectFromRequest(req *runnerv2alpha1.ExecuteRequest, logger *zap.Logger) (*project.Project, error) {
	idResolver := identity.NewResolver(identity.DefaultLifecycleIdentity)

	opts := []project.ProjectOption{
		project.WithIdentityResolver(idResolver),
		project.WithFindRepoUpward(),
		project.WithRespectGitignore(),
		project.WithEnvFilesReadOrder(req.Project.EnvLoadOrder),
		project.WithLogger(logger),
	}

	return project.NewDirProject(
		req.Project.Root,
		opts...,
	)
}

func getFileProjectFromRequest(req *runnerv2alpha1.ExecuteRequest, logger *zap.Logger) (*project.Project, error) {
	idResolver := identity.NewResolver(identity.DefaultLifecycleIdentity)

	path := req.DocumentPath

	if !filepath.IsAbs(path) {
		path = filepath.Join(req.Directory, req.DocumentPath)
	}

	return project.NewFileProject(
		path,
		project.WithIdentityResolver(idResolver),
		project.WithLogger(logger),
	)
}

func getCodeBlockFromRequest(ctx context.Context, req *runnerv2alpha1.ExecuteRequest, proj *project.Project) (*document.CodeBlock, error) {
	tasks, err := project.LoadTasks(ctx, proj)
	if err != nil {
		return nil, err
	}

	tasks, err = project.FilterTasksByFileAndTaskName(tasks, req.DocumentPath, req.GetBlockName())
	if err != nil {
		return nil, err
	}

	switch len(tasks) {
	case 0:
		return nil, errors.New("no tasks found")
	case 1:
		return tasks[0].CodeBlock, nil
	default:
		return nil, errors.New("multiple tasks found")
	}
}

func exitCodeFromErr(err error) int {
	if err == nil {
		return 0
	}
	var exiterr *exec.ExitError
	if errors.As(err, &exiterr) {
		status, ok := exiterr.ProcessState.Sys().(syscall.WaitStatus)
		if ok && status.Signaled() {
			// TODO(adamb): will like need to be improved.
			if status.Signal() == os.Interrupt {
				return 130
			} else if status.Signal() == os.Kill {
				return 137
			}
		}
		return exiterr.ExitCode()
	}
	return -1
}
