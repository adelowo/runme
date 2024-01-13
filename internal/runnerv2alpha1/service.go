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
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/stateful/runme/internal/command"
	"github.com/stateful/runme/internal/document"
	"github.com/stateful/runme/internal/document/identity"
	runnerv2alpha1 "github.com/stateful/runme/internal/gen/proto/go/runme/runner/v2alpha1"
	"github.com/stateful/runme/internal/project"
	"github.com/stateful/runme/internal/rbuffer"
	"github.com/stateful/runme/internal/ulid"
)

const (
	MaxMsgSize = 4096 << 10 // 4 MiB

	// ringBufferSize limits the size of the ring buffers
	// that sit between a command and the handler.
	ringBufferSize = 8192 << 10 // 8 MiB

	// msgBufferSize limits the size of data chunks
	// sent by the handler to clients. It's smaller
	// intentionally as typically the messages are
	// small.
	// In the future, it might be worth to implement
	// variable-sized buffers.
	msgBufferSize = 32 << 10 // 32 KiB
)

type runnerService struct {
	runnerv2alpha1.UnimplementedRunnerServiceServer

	logger *zap.Logger
}

func NewRunnerService(logger *zap.Logger) (runnerv2alpha1.RunnerServiceServer, error) {
	return newRunnerService(logger)
}

func newRunnerService(logger *zap.Logger) (*runnerService, error) {
	return &runnerService{
		logger: logger,
	}, nil
}

func (r *runnerService) stopCommand(
	req *runnerv2alpha1.ExecuteRequest,
	cmd command.Command,
) error {
	if req.Stop == runnerv2alpha1.ExecuteStop_EXECUTE_STOP_UNSPECIFIED {
		return nil
	}

	switch req.Stop {
	case runnerv2alpha1.ExecuteStop_EXECUTE_STOP_INTERRUPT:
		return cmd.StopWithSignal(os.Interrupt)
	case runnerv2alpha1.ExecuteStop_EXECUTE_STOP_KILL:
		return cmd.StopWithSignal(os.Kill)
	default:
		return errors.New("unknown stop signal")
	}
}

func (r *runnerService) receiveLoop(
	srv runnerv2alpha1.RunnerService_ExecuteServer,
	cmd command.Command,
	stdinWriter io.WriteCloser,
	logger *zap.Logger,
) error {
	for {
		req, err := srv.Recv()
		if err == io.EOF {
			logger.Info("client closed the send direction; ignoring")
			return nil
		}
		if err != nil && status.Convert(err).Code() == codes.Canceled {
			if !cmd.IsRunning() {
				logger.Info("stream canceled after the process finished; ignoring")
			} else {
				logger.Info("stream canceled while the process is still running; program will be stopped if non-background")
			}
			return nil
		}
		if err != nil {
			logger.Info("error while receiving a request; stopping the program", zap.Error(err))
			if err := cmd.StopWithSignal(os.Kill); err != nil {
				logger.Info("failed to stop program", zap.Error(err))
			}
			return err
		}

		if err := r.stopCommand(req, cmd); err != nil {
			logger.Info("failed to stop the command", zap.Error(err))
			return err
		}

		if len(req.InputData) != 0 {
			logger.Debug("received input data", zap.Int("len", len(req.InputData)))
			_, err = stdinWriter.Write(req.InputData)
			if err != nil {
				logger.Info("failed to write to stdin", zap.Error(err))
				return err
			}
		}

		// only update winsize when field is explicitly set
		// if req.ProtoReflect().Has(
		// 	req.ProtoReflect().Descriptor().Fields().ByName("winsize"),
		// ) {
		// 	cmd.setWinsize(runnerWinsizeToPty(req.Winsize))
		// }
	}
}

func (r *runnerService) Execute(srv runnerv2alpha1.RunnerService_ExecuteServer) error {
	logger := r.logger.With(zap.String("_id", ulid.GenerateID()))

	logger.Info("running Execute in runnerService")

	// Get the initial request.
	req, err := srv.Recv()
	if err != nil {
		if errors.Is(err, io.EOF) {
			logger.Info("client closed the connection while getting initial request")
			return nil
		}
		logger.Info("failed to receive a request", zap.Error(err))
		return errors.WithStack(err)
	}

	logger.Info("received initial request", zap.Any("req", req))

	ctx := srv.Context()

	block, err := getCodeBlockFromRequest(ctx, req, logger)
	if err != nil {
		return err
	}

	var (
		stdin       io.Reader
		stdinWriter io.WriteCloser
	)

	if req.Interactive {
		stdin, stdinWriter = io.Pipe()
	}

	stdout := rbuffer.NewRingBuffer(ringBufferSize)
	stderr := rbuffer.NewRingBuffer(ringBufferSize)
	// Close buffers so that the readers will be notified about EOF.
	// It's ok to close the buffers multiple times.
	defer func() { _ = stdout.Close() }()
	defer func() { _ = stderr.Close() }()

	cmdOpts := &command.VirtualCommandOptions{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
	}

	cmd, err := command.NewVirtual(block, cmdOpts)
	if err != nil {
		return err
	}

	if err := cmd.Start(ctx); err != nil {
		return err
	}

	if err := srv.Send(&runnerv2alpha1.ExecuteResponse{
		Pid: &runnerv2alpha1.ProcessPID{
			Pid: int64(cmd.PID()),
		},
	}); err != nil {
		return err
	}

	// Write initial input data.
	if len(req.InputData) > 0 {
		if _, err := stdinWriter.Write(req.InputData); err != nil {
			logger.Info("failed to write initial input to stdin", zap.Error(err))
			return err
		}
	}

	// This goroutine will be closed when the handler exits or earlier.
	go func() {
		err := r.receiveLoop(srv, cmd, stdinWriter, logger)
		if err != nil {
			logger.Info("receiveLoop exited with error", zap.Error(err))
		}
	}()

	g := new(errgroup.Group)
	datac := make(chan output)

	g.Go(func() error {
		err := readLoop(stdout, stderr, datac)
		close(datac)
		if errors.Is(err, io.EOF) {
			err = nil
		}
		return err
	})

	g.Go(func() error {
		for data := range datac {
			logger.Debug("sending data", zap.Int("lenStdout", len(data.Stdout)), zap.Int("lenStderr", len(data.Stderr)))
			err := srv.Send(&runnerv2alpha1.ExecuteResponse{
				StdoutData: data.Stdout,
				StderrData: data.Stderr,
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	werr := cmd.Wait()

	exitCode := exitCodeFromErr(werr)

	logger.Info("command finished", zap.Int("exitCode", exitCode))

	// Close the stdinWriter so that the loops in the `cmd` will finish.
	// The problem occurs only with TTY.
	if stdinWriter != nil {
		_ = stdinWriter.Close()
	}

	logger.Info("command was finalized successfully")

	// Close buffers so that the readLoop() can exit.
	_ = stdout.Close()
	_ = stderr.Close()

	werr = g.Wait()
	if werr != nil {
		logger.Info("failed to wait for goroutines to finish", zap.Error(err))
	}

	var finalExitCode *wrapperspb.UInt32Value
	if exitCode > -1 {
		finalExitCode = wrapperspb.UInt32(uint32(exitCode))
		logger.Info("sending the final response with exit code", zap.Int("exitCode", int(finalExitCode.GetValue())))
	} else {
		logger.Info("sending the final response without exit code since its unknown", zap.Int("exitCode", exitCode))
	}

	if err := srv.Send(&runnerv2alpha1.ExecuteResponse{
		ExitCode: finalExitCode,
	}); err != nil {
		logger.Info("failed to send exit code", zap.Error(err))
		if werr == nil {
			werr = err
		}
	}

	return werr
}

func getCodeBlockFromRequest(ctx context.Context, req *runnerv2alpha1.ExecuteRequest, logger *zap.Logger) (*document.CodeBlock, error) {
	var proj *project.Project

	if req.Project != nil {
		var err error
		proj, err = getProjectFromRequestProject(req.Project, logger)
		if err != nil {
			return nil, err
		}
	} else if path := req.DocumentPath; path != "" {
		if !filepath.IsAbs(req.DocumentPath) {
			path = filepath.Join(req.Directory, req.DocumentPath)
		}

		var err error
		proj, err = getProjectFromRequestFile(path, logger)
		if err != nil {
			return nil, err
		}
	}

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

func getProjectFromRequestProject(reqProj *runnerv2alpha1.Project, logger *zap.Logger) (*project.Project, error) {
	idResolver := identity.NewResolver(identity.DefaultLifecycleIdentity)

	opts := []project.ProjectOption{
		project.WithIdentityResolver(idResolver),
		project.WithFindRepoUpward(),
		project.WithRespectGitignore(),
		project.WithEnvFilesReadOrder(reqProj.EnvLoadOrder),
		project.WithLogger(logger),
	}

	return project.NewDirProject(
		reqProj.Root,
		opts...,
	)
}

func getProjectFromRequestFile(path string, logger *zap.Logger) (*project.Project, error) {
	idResolver := identity.NewResolver(identity.DefaultLifecycleIdentity)

	return project.NewFileProject(
		path,
		project.WithIdentityResolver(idResolver),
		project.WithLogger(logger),
	)
}

type output struct {
	Stdout []byte
	Stderr []byte
}

// readLoop uses two sets of buffers in order to avoid allocating
// new memory over and over and putting more presure on GC.
// When the first set is read, it is sent to a channel called `results`.
// `results` should be an unbuffered channel. When a consumer consumes
// from the channel, the loop is unblocked and it moves on to read
// into the second set of buffers and blocks. During this time,
// the consumer has a chance to do something with the data stored
// in the first set of buffers.
func readLoop(
	stdout io.Reader,
	stderr io.Reader,
	results chan<- output,
) error {
	if cap(results) > 0 {
		panic("readLoop requires unbuffered channel")
	}

	read := func(reader io.Reader, fn func(p []byte) output) error {
		for {
			buf := make([]byte, msgBufferSize)
			n, err := reader.Read(buf)
			if err != nil {
				if errors.Is(err, io.EOF) {
					return nil
				}
				return errors.WithStack(err)
			} else if n > 0 {
				results <- fn(buf[:n])
			}
		}
	}

	g := new(errgroup.Group)

	g.Go(func() error {
		return read(stdout, func(p []byte) output {
			return output{Stdout: p}
		})
	})

	g.Go(func() error {
		return read(stderr, func(p []byte) output {
			return output{Stderr: p}
		})
	})

	return g.Wait()
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
