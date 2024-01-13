package runnerv2alpha1

import (
	"go.uber.org/zap"

	runnerv2alpha1 "github.com/stateful/runme/internal/gen/proto/go/runme/runner/v2alpha1"
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
