package kernel

import (
	"context"
	"errors"
	"fmt"
	"time"

	kernelv1 "github.com/stateful/runme/internal/gen/proto/go/runme/kernel/v1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func NewKernelServiceServer(logger *zap.Logger) kernelv1.KernelServiceServer {
	return &kernelServiceServer{logger: logger}
}

type kernelServiceServer struct {
	kernelv1.UnimplementedKernelServiceServer

	sessions *sessionsContainer
	logger   *zap.Logger
}

func (s *kernelServiceServer) PostSession(ctx context.Context, req *kernelv1.PostSessionRequest) (*kernelv1.PostSessionResponse, error) {
	promptStr := req.Prompt
	if promptStr == "" {
		prompt, err := DetectPrompt(req.Command)
		s.logger.Info("detected prompt", zap.Error(err), zap.ByteString("prompt", prompt))
		if err != nil {
			return nil, fmt.Errorf("failed to detect prompt: %w", err)
		}
	}

	session, data, err := newSession(req.Command, promptStr, s.logger)
	if err != nil {
		return nil, err
	}

	s.sessions.AddSession(session)

	return &kernelv1.PostSessionResponse{
		Session:   &kernelv1.Session{Id: session.id},
		IntroData: data,
	}, nil
}

func (s *kernelServiceServer) DeleteSession(ctx context.Context, req *kernelv1.DeleteSessionRequest) (*kernelv1.DeleteSessionResponse, error) {
	session := s.sessions.FindSession(req.SessionId)
	if session == nil {
		return nil, errors.New("session not found")
	}

	s.sessions.DeleteSession(session)

	if err := session.Close(); err != nil {
		return nil, err
	}

	return nil, errors.New("session does not exist")
}

func (s *kernelServiceServer) ListSessions(ctx context.Context, req *kernelv1.ListSessionsRequest) (*kernelv1.ListSessionsResponse, error) {
	sessions := s.sessions.Sessions()
	resp := kernelv1.ListSessionsResponse{
		Sessions: make([]*kernelv1.Session, len(sessions)),
	}
	for idx, s := range sessions {
		resp.Sessions[idx] = &kernelv1.Session{
			Id: s.ID(),
		}
	}
	return &resp, nil
}

func (s *kernelServiceServer) Execute(ctx context.Context, req *kernelv1.ExecuteRequest) (*kernelv1.ExecuteResponse, error) {
	session := s.sessions.FindSession(req.SessionId)
	if session == nil {
		return nil, errors.New("session not found")
	}

	data, exitCode, err := session.Execute(req.Command, time.Second*10)
	if err != nil {
		return nil, err
	}

	return &kernelv1.ExecuteResponse{
		Data:     data,
		ExitCode: wrapperspb.UInt32(uint32(exitCode)),
	}, nil
}