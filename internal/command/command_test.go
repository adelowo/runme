package command

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/stateful/runme/internal/document"
	"github.com/stateful/runme/internal/document/identity"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNativeCommand(t *testing.T) {
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
			source:         "```sh\necho -n test\n```",
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
			source:         "```sh\ncat - | tr a-z A-Z\n```",
			input:          bytes.NewReader([]byte("test\n")),
			expectedStdout: "TEST\n",
		},
		{
			name:           "Env",
			source:         "```sh\necho -n $MY_ENV\n```",
			env:            []string{"MY_ENV=hello"},
			expectedStdout: "hello",
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			doc := document.New([]byte(tc.source), idResolver)
			node, err := doc.Root()
			require.NoError(t, err)

			blocks := document.CollectCodeBlocks(node)
			require.Len(t, blocks, 1)

			stdout := bytes.NewBuffer(nil)
			stderr := bytes.NewBuffer(nil)

			options := &NativeCommandOptions{
				Env:    tc.env,
				Stdout: stdout,
				Stderr: stderr,
				Stdin:  tc.input,
				Logger: logger,
			}
			command, err := NewNative(blocks[0], options)
			require.NoError(t, err)

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			require.NoError(t, command.Start(ctx))
			require.NoError(t, command.Wait())
			require.Equal(t, tc.expectedStdout, stdout.String())
			require.Equal(t, tc.expectedStderr, stderr.String())
		})
	}
}

func TestVirtualCommand(t *testing.T) {
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
	}{
		{
			name:           "BasicShell",
			source:         "```sh\necho -n test\n```",
			expectedStdout: "test",
		},
		{
			name:           "ShellScript",
			source:         "```shellscript\n#!/usr/local/bin/bash\n\nset -x -e -o pipefail\n\necho -n test\n```",
			expectedStdout: "+ echo -n test\r\ntest", // due to -x
		},
		{
			name:           "Python",
			source:         "```python\nprint('test')\n```",
			expectedStdout: "test\r\n",
		},
		{
			name:           "JavaScript",
			source:         "```js\nconsole.log('1'); console.log('2')\n```",
			expectedStdout: "1\r\n2\r\n",
		},
		{
			name:   "Empty",
			source: "```sh\n```",
		},
		{
			name:           "WithInput",
			source:         "```sh\nread name\necho \"My name is $name\"\n```",
			input:          bytes.NewReader([]byte("Test\n")),
			expectedStdout: "My name is Test\r\n",
		},
		{
			name:           "Env",
			source:         "```sh\necho -n $MY_ENV\n```",
			env:            []string{"MY_ENV=hello"},
			expectedStdout: "hello",
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			doc := document.New([]byte(tc.source), idResolver)
			node, err := doc.Root()
			require.NoError(t, err)

			blocks := document.CollectCodeBlocks(node)
			require.Len(t, blocks, 1)

			stdout := bytes.NewBuffer(nil)

			options := &VirtualCommandOptions{
				Env:    tc.env,
				Stdout: stdout,
				Stdin:  tc.input,
				Logger: logger,
			}
			command, err := NewVirtual(blocks[0], options)
			require.NoError(t, err)

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			require.NoError(t, command.Start(ctx))
			require.NoError(t, command.Wait())
			require.Equal(t, tc.expectedStdout, stdout.String())
		})
	}
}

func TestVirtualCommandInputUsingPipe(t *testing.T) {
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

		stdinR, stdinW := io.Pipe()
		stdout := bytes.NewBuffer(nil)

		remoteOptions := &VirtualCommandOptions{
			Stdin:  stdinR,
			Stdout: stdout,
			Logger: logger,
		}
		command, err := NewVirtual(blocks[0], remoteOptions)
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

		stdinR, stdinW := io.Pipe()
		stdout := bytes.NewBuffer(nil)

		remoteOptions := &VirtualCommandOptions{
			Stdin:  stdinR,
			Stdout: stdout,
			Logger: logger,
		}
		command, err := NewVirtual(blocks[0], remoteOptions)
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

		stdinR, stdinW := io.Pipe()
		stdout := bytes.NewBuffer(nil)

		remoteOptions := &VirtualCommandOptions{
			Stdin:  stdinR,
			Stdout: stdout,
			Logger: logger,
		}
		command, err := NewVirtual(blocks[0], remoteOptions)
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
