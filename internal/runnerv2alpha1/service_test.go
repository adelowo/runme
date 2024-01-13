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
	t.Helper()

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
	t.Helper()

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

	testCases := []struct {
		name           string
		blockName      string
		inputData      []byte
		expectedOutput string
	}{
		{
			name:           "Basic",
			blockName:      "basic",
			expectedOutput: "test",
		},

		{
			name:           "BasicSleep",
			blockName:      "basic-sleep",
			expectedOutput: "1\r\n2\r\n",
		},
		{
			name:           "BasicInput",
			blockName:      "basic-input",
			inputData:      []byte("Frank\n"),
			expectedOutput: "My name is Frank\r\n",
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			stream, err := client.Execute(context.Background())
			require.NoError(t, err)

			execResult := make(chan executeResult)
			go getExecuteResult(stream, execResult)

			req := &runnerv2alpha1.ExecuteRequest{
				Project: &runnerv2alpha1.Project{
					Root: tmpDir,
				},
				Block: &runnerv2alpha1.ExecuteRequest_BlockName{
					BlockName: tc.blockName,
				},
			}

			if tc.inputData != nil {
				req.Interactive = true
				req.InputData = tc.inputData
			}

			err = stream.Send(req)
			assert.NoError(t, err)

			result := <-execResult

			assert.NoError(t, result.Err)
			assert.Equal(t, tc.expectedOutput, string(result.Stdout))
			assert.EqualValues(t, 0, result.ExitCode)
		})
	}
}
