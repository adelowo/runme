package runnerv2alpha1

import (
	"io"
	"os"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"

	runnerv2alpha1 "github.com/stateful/runme/internal/gen/proto/go/runme/runner/v2alpha1"
	"github.com/stateful/runme/internal/ulid"
)

func (r *runnerService) Execute(srv runnerv2alpha1.RunnerService_ExecuteServer) error {
	id := ulid.GenerateID()
	logger := r.logger.With(zap.String("id", id))

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

	// Create the execution.
	exec, err := newExecutionFrom(ctx, id, req, logger)
	if err != nil {
		return err
	}

	// Start the command and send the initial response with PID.
	if err := exec.Cmd.Start(ctx); err != nil {
		return err
	}
	if err := srv.Send(&runnerv2alpha1.ExecuteResponse{
		Pid: &runnerv2alpha1.ProcessPID{
			Pid: int64(exec.Cmd.PID()),
		},
	}); err != nil {
		return err
	}

	go func() {
		for {
			req, err := srv.Recv()
			logger.Info("received request", zap.Any("req", req), zap.Error(err))
			switch {
			case err == nil:
				// continue
			case err == io.EOF:
				logger.Info("client closed the send direction; ignoring")
				return
			case status.Convert(err).Code() == codes.Canceled || status.Convert(err).Code() == codes.DeadlineExceeded:
				if !exec.Cmd.IsRunning() {
					logger.Info("stream canceled after the process finished; ignoring")
				} else {
					logger.Info("stream canceled while the process is still running; program will be stopped if non-background")
					if err := exec.Cmd.StopWithSignal(os.Kill); err != nil {
						logger.Info("failed to stop program", zap.Error(err))
					}
				}
				return
			}

			// only update winsize when field is explicitly set
			// if req.ProtoReflect().Has(
			// 	req.ProtoReflect().Descriptor().Fields().ByName("winsize"),
			// ) {
			// 	cmd.setWinsize(runnerWinsizeToPty(req.Winsize))
			// }

			// First, handle the input data. It does not prevent to stop the program right after.
			if l := len(req.InputData); l > 0 {
				logger.Debug("received input data", zap.Int("len", l))
				_, err = exec.Write(req.InputData)
				if err != nil {
					logger.Info("failed to write to stdin", zap.Error(err))
				}
			}

			switch req.Stop {
			case runnerv2alpha1.ExecuteStop_EXECUTE_STOP_UNSPECIFIED:
				// continue
			case runnerv2alpha1.ExecuteStop_EXECUTE_STOP_INTERRUPT:
				err = exec.Cmd.StopWithSignal(os.Interrupt)
			case runnerv2alpha1.ExecuteStop_EXECUTE_STOP_KILL:
				err = exec.Cmd.StopWithSignal(os.Kill)
			default:
				err = errors.New("unknown stop signal")
			}
			if err != nil {
				logger.Info("failed to stop program", zap.Error(err))
				return
			}
		}
	}()

	exitCode, waitErr := exec.Wait(ctx, srv)

	logger.Info("command finished", zap.Int("exitCode", exitCode), zap.Error(waitErr))

	var finalExitCode *wrapperspb.UInt32Value
	if exitCode > -1 {
		finalExitCode = wrapperspb.UInt32(uint32(exitCode))
	}

	if err := srv.Send(&runnerv2alpha1.ExecuteResponse{
		ExitCode: finalExitCode,
	}); err != nil {
		logger.Info("failed to send exit code", zap.Error(err))
	}

	return waitErr
}
