//go:build !windows

package runnerv2alpha1

import (
	"bytes"
	"context"
	"io"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

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

var testReadme = bytes.Join(
	[][]byte{
		[]byte("```sh {\"name\":\"basic\"}\necho -n test\n```"),
		[]byte("```sh {\"name\":\"basic-sleep\"}\necho 1\nsleep 1\necho 2\n```"),
		[]byte("```sh {\"name\":\"basic-input\"}\nread name\necho \"My name is $name\"\n```"),
	},
	[]byte("\n"),
)

func TestRunnerServiceServerExecute(t *testing.T) {
	t.Parallel()

	lis, stop := testStartRunnerServiceServer(t)
	t.Cleanup(stop)
	_, client := testCreateRunnerServiceClient(t, lis)

	tmpDir := t.TempDir()
	readmeFile := filepath.Join(tmpDir, "README.md")

	err := os.WriteFile(readmeFile, testReadme, 0o644)
	require.NoError(t, err)

	t.Run("Basic", func(t *testing.T) {
		t.Parallel()

		stream, err := client.Execute(context.Background())
		require.NoError(t, err)

		execResult := make(chan executeResult)
		go getExecuteResult(stream, execResult)

		err = stream.Send(&runnerv2alpha1.ExecuteRequest{
			Project: &runnerv2alpha1.Project{
				Root: tmpDir,
			},
			Block: &runnerv2alpha1.ExecuteRequest_BlockName{
				BlockName: "basic",
			},
		})
		assert.NoError(t, err)

		result := <-execResult

		assert.NoError(t, result.Err)
		assert.Equal(t, "test", string(result.Stdout))
		assert.EqualValues(t, 0, result.ExitCode)
	})

	t.Run("BasicSleep", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		stream, err := client.Execute(ctx)
		require.NoError(t, err)

		execResult := make(chan executeResult)
		go getExecuteResult(stream, execResult)

		err = stream.Send(&runnerv2alpha1.ExecuteRequest{
			Project: &runnerv2alpha1.Project{
				Root: tmpDir,
			},
			Block: &runnerv2alpha1.ExecuteRequest_BlockName{
				BlockName: "basic-sleep",
			},
		})
		assert.NoError(t, err)

		result := <-execResult

		assert.NoError(t, result.Err)
		assert.Equal(t, "1\r\n2\r\n", string(result.Stdout))
		assert.EqualValues(t, 0, result.ExitCode)
	})

	t.Run("BasicInput", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		stream, err := client.Execute(ctx)
		require.NoError(t, err)

		execResult := make(chan executeResult)
		go getExecuteResult(stream, execResult)

		err = stream.Send(&runnerv2alpha1.ExecuteRequest{
			Project: &runnerv2alpha1.Project{
				Root: tmpDir,
			},
			Block: &runnerv2alpha1.ExecuteRequest_BlockName{
				BlockName: "basic-input",
			},
			Interactive: true,
			InputData:   []byte("Frank\n"),
		})
		assert.NoError(t, err)

		result := <-execResult

		assert.NoError(t, result.Err)
		assert.Equal(t, "My name is Frank\r\n", string(result.Stdout))
		assert.EqualValues(t, 0, result.ExitCode)
	})
}
