package command

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/stateful/runme/internal/document"
	"github.com/stateful/runme/internal/document/identity"
	"github.com/stretchr/testify/require"
)

func TestNewLocal(t *testing.T) {
	t.Parallel()

	idResolver := identity.NewResolver(identity.AllLifecycleIdentity)

	t.Run("Basic", func(t *testing.T) {
		t.Parallel()

		source := "```sh\necho -n test\n```"

		doc := document.New([]byte(source), idResolver)
		node, err := doc.Root()
		require.NoError(t, err)

		blocks := document.CollectCodeBlocks(node)
		require.Len(t, blocks, 1)

		stdout := bytes.NewBuffer(nil)

		localOptions := &NativeCommandOptions{
			Stdout: stdout,
		}
		command, err := NewNative(blocks[0], localOptions)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		require.NoError(t, command.Start(ctx))
		require.NoError(t, command.Wait())
		require.Equal(t, "test", stdout.String())
	})
}

func TestNewRemote(t *testing.T) {
	t.Parallel()

	idResolver := identity.NewResolver(identity.AllLifecycleIdentity)

	t.Run("SimpleEcho", func(t *testing.T) {
		t.Parallel()

		source := "```sh\necho -n test\n```"

		doc := document.New([]byte(source), idResolver)
		node, err := doc.Root()
		require.NoError(t, err)

		blocks := document.CollectCodeBlocks(node)
		require.Len(t, blocks, 1)

		stdout := bytes.NewBuffer(nil)

		remoteOptions := &VirtualCommandOptions{
			Stdout: stdout,
		}
		command, err := NewVirtual(blocks[0], remoteOptions)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		require.NoError(t, command.Start(ctx))
		require.NoError(t, command.Wait())
		require.Equal(t, "test", stdout.String())
	})

	t.Run("ReadInput", func(t *testing.T) {
		t.Parallel()

		source := "```sh\nread name\necho \"My name is $name\"\n```"

		doc := document.New([]byte(source), idResolver)
		node, err := doc.Root()
		require.NoError(t, err)

		blocks := document.CollectCodeBlocks(node)
		require.Len(t, blocks, 1)

		stdin := bytes.NewReader([]byte("Unit Test\n"))
		stdout := bytes.NewBuffer(nil)

		remoteOptions := &VirtualCommandOptions{
			Stdin:  stdin,
			Stdout: stdout,
		}
		command, err := NewVirtual(blocks[0], remoteOptions)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		require.NoError(t, command.Start(ctx))
		require.NoError(t, command.Wait())
		require.Equal(t, "My name is Unit Test\r\n", stdout.String())
	})
}
