package command

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/stateful/runme/internal/document"
	"github.com/stateful/runme/internal/document/identity"
	runnerv2alpha1 "github.com/stateful/runme/internal/gen/proto/go/runme/runner/v2alpha1"
)

var testConfigBasic = &Config{
	ProgramName: "echo",
	Arguments:   []string{"-n", "test"},
	Mode:        runnerv2alpha1.CommandMode_COMMAND_MODE_INLINE,
}

func TestNativeCommand(t *testing.T) {
	t.Run("OptionsIsNil", func(t *testing.T) {
		cmd, err := NewNative(testConfigBasic, nil)
		require.NoError(t, err)
		require.NoError(t, cmd.Start(context.Background()))
		require.NoError(t, cmd.Wait())
	})

	t.Run("Output", func(t *testing.T) {
		stdout := bytes.NewBuffer(nil)
		opts := &NativeCommandOptions{
			Stdout: stdout,
		}
		cmd, err := NewNative(testConfigBasic, opts)
		require.NoError(t, err)
		require.NoError(t, cmd.Start(context.Background()))
		require.NoError(t, cmd.Wait())
		assert.Equal(t, "test", stdout.String())
	})
}

func TestVirtualCommand(t *testing.T) {
	t.Run("OptionsIsNil", func(t *testing.T) {
		cmd, err := NewVirtual(testConfigBasic, nil)
		require.NoError(t, err)
		require.NoError(t, cmd.Start(context.Background()))
		require.NoError(t, cmd.Wait())
	})

	t.Run("Output", func(t *testing.T) {
		stdout := bytes.NewBuffer(nil)
		opts := &VirtualCommandOptions{
			Stdout: stdout,
		}
		cmd, err := NewVirtual(testConfigBasic, opts)
		require.NoError(t, err)
		require.NoError(t, cmd.Start(context.Background()))
		require.NoError(t, cmd.Wait())
		assert.Equal(t, "test", stdout.String())
	})
}

func TestExecutionCommandFromCodeBlocks(t *testing.T) {
	t.Parallel()

	idResolver := identity.NewResolver(identity.AllLifecycleIdentity)

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	defer logger.Sync()

	testCases := []struct {
		name           string
		source         string
		env            []string
		input          io.Reader
		expectedStdout string
		expectedStderr string
	}{
		{
			name:           "BasicShell",
			source:         "```bash\necho -n test\n```",
			expectedStdout: "test",
		},
		{
			name:           "ShellScript",
			source:         "```shellscript\n#!/usr/local/bin/bash\n\nset -x -e -o pipefail\n\necho -n test\n```",
			expectedStdout: "test",
			expectedStderr: "+ echo -n test\n", // due to -x
		},
		{
			name:           "Python",
			source:         "```python\nprint('test')\n```",
			expectedStdout: "test\n",
		},
		{
			name:           "JavaScript",
			source:         "```js\nconsole.log('1'); console.log('2')\n```",
			expectedStdout: "1\n2\n",
		},
		{
			name:   "Empty",
			source: "```sh\n```",
		},
		{
			name:           "WithInput",
			source:         "```bash\ncat - | tr a-z A-Z\n```",
			input:          bytes.NewReader([]byte("test\n")),
			expectedStdout: "TEST\n",
		},
		{
			name:           "Env",
			source:         "```bash\necho -n $MY_ENV\n```",
			env:            []string{"MY_ENV=hello"},
			expectedStdout: "hello",
		},
		{
			name:           "Interpreter",
			source:         "```sh { \"interpreter\": \"bash\" }\necho -n test\n```",
			expectedStdout: "test",
		},
		{
			name:           "FrontmatterShell",
			source:         "---\nshell: bash\n---\n```sh\necho -n test\n```",
			expectedStdout: "test",
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run("NativeCommand", func(t *testing.T) {
			t.Parallel()

			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				testExecuteNativeCommand(
					t,
					idResolver,
					[]byte(tc.source),
					tc.env,
					tc.input,
					tc.expectedStdout,
					tc.expectedStderr,
					logger,
				)
			})
		})

		t.Run("VirtualCommand", func(t *testing.T) {
			t.Parallel()

			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				testExecuteVirtualCommand(
					t,
					idResolver,
					[]byte(tc.source),
					tc.env,
					tc.input,
					tc.expectedStdout,
					logger,
				)
			})
		})
	}
}

func testExecuteNativeCommand(
	t *testing.T,
	idResolver *identity.IdentityResolver,
	source []byte,
	env []string,
	input io.Reader,
	expectedStdout string,
	expectedStderr string,
	logger *zap.Logger,
) {
	t.Helper()

	doc := document.New(source, idResolver)
	node, err := doc.Root()
	require.NoError(t, err)

	blocks := document.CollectCodeBlocks(node)
	require.Len(t, blocks, 1)

	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)

	cfg, err := NewConfigFromCodeBlock(blocks[0])
	require.NoError(t, err)

	options := &NativeCommandOptions{
		Env:    env,
		Stdout: stdout,
		Stderr: stderr,
		Stdin:  input,
		Logger: logger,
	}
	require.NoError(t, err)

	command, err := NewNative(cfg, options)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	require.NoError(t, command.Start(ctx))
	require.NoError(t, command.Wait())
	require.Equal(t, expectedStdout, stdout.String())
	require.Equal(t, expectedStderr, stderr.String())
}

func testExecuteVirtualCommand(
	t *testing.T,
	idResolver *identity.IdentityResolver,
	source []byte,
	env []string,
	input io.Reader,
	expectedStdout string,
	logger *zap.Logger,
) {
	t.Helper()

	doc := document.New(source, idResolver)
	node, err := doc.Root()
	require.NoError(t, err)

	blocks := document.CollectCodeBlocks(node)
	require.Len(t, blocks, 1)

	cfg, err := NewConfigFromCodeBlock(blocks[0])
	require.NoError(t, err)

	stdout := bytes.NewBuffer(nil)

	options := &VirtualCommandOptions{
		Env:    env,
		Stdout: stdout,
		Stdin:  input,
		Logger: logger,
	}
	command, err := NewVirtual(cfg, options)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	require.NoError(t, command.Start(ctx))
	require.NoError(t, command.Wait())
	require.Equal(t, expectedStdout, stdout.String())
}

func TestVirtualCommandFromCodeBlocksWithInputUsingPipe(t *testing.T) {
	idResolver := identity.NewResolver(identity.AllLifecycleIdentity)

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	defer logger.Sync()

	t.Run("Cat", func(t *testing.T) {
		t.Parallel()

		source := "```sh\ncat - | tr a-z A-Z\n```"

		doc := document.New([]byte(source), idResolver)
		node, err := doc.Root()
		require.NoError(t, err)

		blocks := document.CollectCodeBlocks(node)
		require.Len(t, blocks, 1)

		cfg, err := NewConfigFromCodeBlock(blocks[0])
		require.NoError(t, err)

		stdinR, stdinW := io.Pipe()
		stdout := bytes.NewBuffer(nil)

		remoteOptions := &VirtualCommandOptions{
			Stdin:  stdinR,
			Stdout: stdout,
			Logger: logger,
		}
		command, err := NewVirtual(cfg, remoteOptions)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		require.NoError(t, command.Start(ctx))

		_, err = stdinW.Write([]byte("unit tests\n"))
		require.NoError(t, err)
		_, err = stdinW.Write([]byte{0x04}) // EOT
		require.NoError(t, err)

		require.NoError(t, command.Wait())
		require.Equal(t, "UNIT TESTS\r\n", stdout.String())
	})

	t.Run("Read", func(t *testing.T) {
		t.Parallel()

		source := "```sh\nread name\necho \"My name is $name\"\n```"

		doc := document.New([]byte(source), idResolver)
		node, err := doc.Root()
		require.NoError(t, err)

		blocks := document.CollectCodeBlocks(node)
		require.Len(t, blocks, 1)

		cfg, err := NewConfigFromCodeBlock(blocks[0])
		require.NoError(t, err)

		stdinR, stdinW := io.Pipe()
		stdout := bytes.NewBuffer(nil)

		remoteOptions := &VirtualCommandOptions{
			Stdin:  stdinR,
			Stdout: stdout,
			Logger: logger,
		}
		command, err := NewVirtual(cfg, remoteOptions)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		require.NoError(t, command.Start(ctx))

		_, err = stdinW.Write([]byte("Unit Test\n"))
		require.NoError(t, err)

		require.NoError(t, command.Wait())
		require.Equal(t, "My name is Unit Test\r\n", stdout.String())
	})

	t.Run("SimulateCtrlC", func(t *testing.T) {
		t.Parallel()

		// Using sh start bash. We need to go deeper...
		source := "```sh\nbash\n```"

		doc := document.New([]byte(source), idResolver)
		node, err := doc.Root()
		require.NoError(t, err)

		blocks := document.CollectCodeBlocks(node)
		require.Len(t, blocks, 1)

		cfg, err := NewConfigFromCodeBlock(blocks[0])
		require.NoError(t, err)

		stdinR, stdinW := io.Pipe()
		stdout := bytes.NewBuffer(nil)

		remoteOptions := &VirtualCommandOptions{
			Stdin:  stdinR,
			Stdout: stdout,
			Logger: logger,
		}
		command, err := NewVirtual(cfg, remoteOptions)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		require.NoError(t, command.Start(ctx))

		errc := make(chan error)
		go func() {
			defer close(errc)

			time.Sleep(time.Millisecond * 500)
			_, err = stdinW.Write([]byte("sleep 30\n"))
			errc <- err

			// cancel sleep
			time.Sleep(time.Millisecond * 500)
			_, err = stdinW.Write([]byte{3})
			errc <- err

			// terminate shell
			time.Sleep(time.Millisecond * 500)
			_, err = stdinW.Write([]byte{4})
			errc <- err

			// close writer; it's not needed
			errc <- stdinW.Close()
		}()
		for err := range errc {
			require.NoError(t, err)
		}

		require.EqualError(t, command.Wait(), "exit status 130")
	})
}
