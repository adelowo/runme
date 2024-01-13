package command

import (
	"bytes"
	"context"
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

var testConfigBasic = &Config{ProgramName: "echo", Arguments: []string{"-n", "test"}}

func TestNewNative(t *testing.T) {
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

func TestNewVirtual(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	defer logger.Sync()

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
