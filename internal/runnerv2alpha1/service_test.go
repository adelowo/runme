//go:build !windows

package runnerv2alpha1

import (
	"context"
	"io"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	runnerv2alpha1 "github.com/stateful/runme/internal/gen/proto/go/runme/runner/v2alpha1"
)

func TestRunnerServiceServerExecute(t *testing.T) {
	t.Parallel()

	lis, stop := testStartRunnerServiceServer(t)
	t.Cleanup(stop)
	_, client := testCreateRunnerServiceClient(t, lis)

	// data := bytes.Join(
	// 	[][]byte{
	// 		[]byte("```sh {\"name\":\"basic\"}\necho -n test\n```"),
	// 		[]byte("```sh {\"name\":\"basic-sleep\"}\necho 1\nsleep 1\necho 2\n```"),
	// 		[]byte("```sh {\"name\":\"basic-input\"}\nread name\necho \"My name is $name\"\n```"),
	// 		[]byte("```python {\"name\": \"py\"}\nprint('test')\n```"),
	// 		[]byte("```js {\"name\": \"js\"}\nconsole.log('1'); console.log('2')\n```"),
	// 	},
	// 	[]byte("\n"),
	// )

	testCases := []struct {
		name           string
		programConfig  *runnerv2alpha1.ProgramConfig
		inputData      []byte
		expectedOutput string
	}{
		{
			name: "Basic",
			programConfig: &runnerv2alpha1.ProgramConfig{
				ProgramName: "bash",
				Source: &runnerv2alpha1.ProgramConfig_Commands{
					Commands: &runnerv2alpha1.CommandList{
						Commands: []string{
							"echo -n test",
						},
					},
				},
			},
			expectedOutput: "test",
		},

		// {
		// 	name:           "BasicSleep",
		// 	blockName:      "basic-sleep",
		// 	expectedOutput: "1\r\n2\r\n",
		// },
		// {
		// 	name:           "BasicSleepWithInputData",
		// 	blockName:      "basic-sleep",
		// 	inputData:      []byte("unused input\n"),
		// 	expectedOutput: "1\r\n2\r\n",
		// },
		// {
		// 	name:           "BasicInput",
		// 	blockName:      "basic-input",
		// 	inputData:      []byte("Frank\n"),
		// 	expectedOutput: "My name is Frank\r\n",
		// },
		// {
		// 	name:           "Python",
		// 	blockName:      "py",
		// 	expectedOutput: "test\r\n",
		// },
		// {
		// 	name:           "JavaScript",
		// 	blockName:      "js",
		// 	expectedOutput: "1\r\n2\r\n",
		// },
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
				Config: tc.programConfig,
			}

			if tc.inputData != nil {
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

// func TestRunnerServiceServerExecute_Input(t *testing.T) {
// 	t.Parallel()

// 	lis, stop := testStartRunnerServiceServer(t)
// 	t.Cleanup(stop)
// 	_, client := testCreateRunnerServiceClient(t, lis)

// 	tmpDir := t.TempDir()

// 	t.Run("ContinuousInput", func(t *testing.T) {
// 		t.Parallel()

// 		readmeFile := filepath.Join(tmpDir, "ContinuousInput.md")
// 		data := []byte("```sh {\"name\":\"cat\"}\ncat - | tr a-z A-Z\n```")
// 		err := os.WriteFile(readmeFile, data, 0o644)
// 		require.NoError(t, err)

// 		stream, err := client.Execute(context.Background())
// 		require.NoError(t, err)

// 		execResult := make(chan executeResult)
// 		go getExecuteResult(stream, execResult)

// 		req := &runnerv2alpha1.ExecuteRequest{
// 			Project: &runnerv2alpha1.Project{
// 				Root: tmpDir,
// 			},
// 			Block: &runnerv2alpha1.ExecuteRequest_BlockName{
// 				BlockName: "cat",
// 			},
// 			InputData:   []byte("a\n"),
// 			Interactive: true,
// 		}

// 		err = stream.Send(req)
// 		assert.NoError(t, err)

// 		for _, data := range [][]byte{[]byte("b\n"), []byte("c\n"), []byte("d\n"), {0x04}} {
// 			req := &runnerv2alpha1.ExecuteRequest{InputData: data}
// 			err = stream.Send(req)
// 			assert.NoError(t, err)
// 		}

// 		result := <-execResult

// 		assert.NoError(t, result.Err)
// 		assert.Equal(t, "A\r\nB\r\nC\r\nD\r\n", string(result.Stdout))
// 		assert.EqualValues(t, 0, result.ExitCode)
// 	})

// 	t.Run("SimulateCtrlC", func(t *testing.T) {
// 		t.Parallel()

// 		readmeFile := filepath.Join(tmpDir, "ContinuousInput.md")
// 		data := []byte("```sh {\"name\":\"bash\"}\nbash\n```")
// 		err := os.WriteFile(readmeFile, data, 0o644)
// 		require.NoError(t, err)

// 		stream, err := client.Execute(context.Background())
// 		require.NoError(t, err)

// 		execResult := make(chan executeResult)
// 		go getExecuteResult(stream, execResult)

// 		req := &runnerv2alpha1.ExecuteRequest{
// 			Project: &runnerv2alpha1.Project{
// 				Root: tmpDir,
// 			},
// 			Block: &runnerv2alpha1.ExecuteRequest_BlockName{
// 				BlockName: "bash",
// 			},
// 			Interactive: true,
// 		}

// 		err = stream.Send(req)
// 		assert.NoError(t, err)

// 		time.Sleep(time.Millisecond * 500)
// 		err = stream.Send(&runnerv2alpha1.ExecuteRequest{InputData: []byte("sleep 30")})
// 		assert.NoError(t, err)

// 		// cancel sleep
// 		time.Sleep(time.Millisecond * 500)
// 		err = stream.Send(&runnerv2alpha1.ExecuteRequest{InputData: []byte{0x03}})
// 		assert.NoError(t, err)

// 		// terminate shell
// 		time.Sleep(time.Millisecond * 500)
// 		err = stream.Send(&runnerv2alpha1.ExecuteRequest{InputData: []byte{0x04}})
// 		assert.NoError(t, err)

// 		result := <-execResult

// 		// TODO(adamb): This should be a specific gRPC error rather than Unknown.
// 		assert.Contains(t, result.Err.Error(), "exit status 130")
// 		assert.Equal(t, 130, result.ExitCode)
// 	})

// 	t.Run("CloseSendDirection", func(t *testing.T) {
// 		t.Parallel()

// 		readmeFile := filepath.Join(tmpDir, "CloseSendDirection.md")
// 		data := []byte("```sh {\"name\":\"bash\"}\nbash\n```")
// 		err := os.WriteFile(readmeFile, data, 0o644)
// 		require.NoError(t, err)

// 		stream, err := client.Execute(context.Background())
// 		require.NoError(t, err)

// 		execResult := make(chan executeResult)
// 		go getExecuteResult(stream, execResult)

// 		req := &runnerv2alpha1.ExecuteRequest{
// 			Project: &runnerv2alpha1.Project{
// 				Root: tmpDir,
// 			},
// 			Block: &runnerv2alpha1.ExecuteRequest_BlockName{
// 				BlockName: "bash",
// 			},
// 			Interactive: true,
// 		}

// 		err = stream.Send(req)
// 		assert.NoError(t, err)

// 		// Close the send direction.
// 		assert.NoError(t, stream.CloseSend())

// 		result := <-execResult
// 		// TODO(adamb): This should be a specific gRPC error rather than Unknown.
// 		assert.Contains(t, result.Err.Error(), "signal: interrupt")
// 		assert.Equal(t, 130, result.ExitCode)
// 	})
// }

// func TestRunnerServiceServerExecute_WithStop(t *testing.T) {
// 	t.Parallel()

// 	lis, stop := testStartRunnerServiceServer(t)
// 	t.Cleanup(stop)
// 	_, client := testCreateRunnerServiceClient(t, lis)

// 	tmpDir := t.TempDir()

// 	readmeFile := filepath.Join(tmpDir, "WithStop.md")
// 	data := []byte("```sh {\"name\":\"sleep\"}\necho 1\nsleep 30\n```")
// 	err := os.WriteFile(readmeFile, data, 0o644)
// 	require.NoError(t, err)

// 	stream, err := client.Execute(context.Background())
// 	require.NoError(t, err)

// 	execResult := make(chan executeResult)
// 	go getExecuteResult(stream, execResult)

// 	req := &runnerv2alpha1.ExecuteRequest{
// 		Project: &runnerv2alpha1.Project{
// 			Root: tmpDir,
// 		},
// 		Block: &runnerv2alpha1.ExecuteRequest_BlockName{
// 			BlockName: "sleep",
// 		},
// 		Interactive: true,
// 	}

// 	err = stream.Send(req)
// 	require.NoError(t, err)

// 	errc := make(chan error)
// 	go func() {
// 		defer close(errc)
// 		time.Sleep(500 * time.Millisecond)
// 		err := stream.Send(&runnerv2alpha1.ExecuteRequest{
// 			Stop: runnerv2alpha1.ExecuteStop_EXECUTE_STOP_INTERRUPT,
// 		})
// 		errc <- err
// 	}()
// 	require.NoError(t, <-errc)

// 	result := <-execResult

// 	// TODO(adamb): There should be no error.
// 	assert.Contains(t, result.Err.Error(), "signal: interrupt")
// 	assert.Equal(t, 130, result.ExitCode)
// }

// func TestRunnerServiceServerExecute_Winsize(t *testing.T) {
// 	t.Parallel()

// 	lis, stop := testStartRunnerServiceServer(t)
// 	t.Cleanup(stop)
// 	_, client := testCreateRunnerServiceClient(t, lis)

// 	tmpDir := t.TempDir()

// 	readmeFile := filepath.Join(tmpDir, "Winsize.md")
// 	data := [][]byte{
// 		[]byte("```sh {\"name\":\"get-size\"}\ntput lines -T linux\ntput cols -T linux\n```"),
// 	}
// 	err := os.WriteFile(readmeFile, bytes.Join(data, []byte{'\n'}), 0o644)
// 	require.NoError(t, err)

// 	t.Run("DefaultWinsize", func(t *testing.T) {
// 		stream, err := client.Execute(context.Background())
// 		require.NoError(t, err)

// 		execResult := make(chan executeResult)
// 		go getExecuteResult(stream, execResult)

// 		req := &runnerv2alpha1.ExecuteRequest{
// 			Project: &runnerv2alpha1.Project{
// 				Root: tmpDir,
// 			},
// 			Block: &runnerv2alpha1.ExecuteRequest_BlockName{
// 				BlockName: "get-size",
// 			},
// 			Interactive: true,
// 		}
// 		require.NoError(t, stream.Send(req))

// 		result := <-execResult

// 		assert.NoError(t, result.Err)
// 		assert.Equal(t, "24\r\n80\r\n", string(result.Stdout))
// 		assert.EqualValues(t, 0, result.ExitCode)
// 	})

// 	t.Run("SetWinsize", func(t *testing.T) {
// 		stream, err := client.Execute(context.Background())
// 		require.NoError(t, err)

// 		execResult := make(chan executeResult)
// 		go getExecuteResult(stream, execResult)

// 		req := &runnerv2alpha1.ExecuteRequest{
// 			Project: &runnerv2alpha1.Project{
// 				Root: tmpDir,
// 			},
// 			Block: &runnerv2alpha1.ExecuteRequest_BlockName{
// 				BlockName: "get-size",
// 			},
// 			Interactive: true,
// 			Winsize: &runnerv2alpha1.Winsize{
// 				Cols: 200,
// 				Rows: 64,
// 			},
// 		}
// 		require.NoError(t, stream.Send(req))

// 		result := <-execResult

// 		assert.NoError(t, result.Err)
// 		assert.Equal(t, "24\r\n80\r\n", string(result.Stdout))
// 		assert.EqualValues(t, 0, result.ExitCode)
// 	})
// }

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
