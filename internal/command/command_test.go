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

		localOptions := &LocalOptions{
			Stdout: stdout,
		}
		command, err := NewLocal(blocks[0], localOptions)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		require.NoError(t, command.Start(ctx))
		require.NoError(t, command.Wait())
		require.Equal(t, "test", stdout.String())
	})
}