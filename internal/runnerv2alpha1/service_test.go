//go:build !windows

package runnerv2alpha1

import (
	"bytes"
	"context"
	"io"
	"net"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	runnerv2alpha1 "github.com/stateful/runme/internal/gen/proto/go/runme/runner/v2alpha1"
)

func testStartRunnerServiceServer(t *testing.T) (
	interface{ Dial() (net.Conn, error) },
	func(),
) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	lis := bufconn.Listen(1024 << 10)
	server := grpc.NewServer()
	runnerService, err := newRunnerService(logger)
	require.NoError(t, err)
	runnerv2alpha1.RegisterRunnerServiceServer(server, runnerService)
	go server.Serve(lis)
	return lis, server.Stop
}

func testCreateRunnerServiceClient(
	t *testing.T,
	lis interface{ Dial() (net.Conn, error) },
) (*grpc.ClientConn, runnerv2alpha1.RunnerServiceClient) {
	conn, err := grpc.Dial(
		"",
		grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	return conn, runnerv2alpha1.NewRunnerServiceClient(conn)
}

type executeResult struct {
	Stdout   []byte
	Stderr   []byte
	ExitCode int
	Err      error
}

func getExecuteResult(
	stream runnerv2alpha1.RunnerService_ExecuteClient,
	resultc chan<- executeResult,
) {
	result := executeResult{ExitCode: -1}

	for {
		r, rerr := stream.Recv()
		if rerr != nil {
			if rerr == io.EOF {
				rerr = nil
			}
			result.Err = rerr
			break
		}
		result.Stdout = append(result.Stdout, r.StdoutData...)
		result.Stderr = append(result.Stderr, r.StderrData...)
		if r.ExitCode != nil {
			result.ExitCode = int(r.ExitCode.Value)
		}
	}

	resultc <- result
}

func TestRunnerServiceServer(t *testing.T) {
	t.Parallel()

	lis, stop := testStartRunnerServiceServer(t)
	t.Cleanup(stop)
	_, client := testCreateRunnerServiceClient(t, lis)

	tmpDir := t.TempDir()
	err := os.WriteFile(tmpDir+"/test.md", []byte("```sh {\"name\":\"echo\"}\necho -n test\n```"), 0o644)
	require.NoError(t, err)

	stream, err := client.Execute(context.Background())
	require.NoError(t, err)

	execResult := make(chan executeResult)
	go getExecuteResult(stream, execResult)

	err = stream.Send(&runnerv2alpha1.ExecuteRequest{
		DocumentPath: tmpDir + "/test.md",
		Block: &runnerv2alpha1.ExecuteRequest_BlockName{
			BlockName: "echo",
		},
	})
	assert.NoError(t, err)

	result := <-execResult

	assert.NoError(t, result.Err)
	assert.Equal(t, "test", string(result.Stdout))
	assert.EqualValues(t, 0, result.ExitCode)
}

func Test_readLoop(t *testing.T) {
	const dataSize = 10 * 1024 * 1024

	stdout := make([]byte, dataSize)
	stderr := make([]byte, dataSize)
	results := make(chan output)
	stdoutN, stderrN := 0, 0

	done := make(chan struct{})
	go func() {
		for data := range results {
			stdoutN += len(data.Stdout)
			stderrN += len(data.Stderr)
		}
		close(done)
	}()

	err := readLoop(bytes.NewReader(stdout), bytes.NewReader(stderr), results)
	assert.NoError(t, err)
	close(results)
	<-done
	assert.Equal(t, dataSize, stdoutN)
	assert.Equal(t, dataSize, stderrN)
}
