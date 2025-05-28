package commands_test

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"

	commands "micheam.com/aico/internal/commands"
	"micheam.com/aico/internal/config"
)

func TestChatSendCommand(t *testing.T) {
	// Setup
	t.Setenv(config.EnvKeyConfigPath, filepath.Join("testdata", "config.yaml"))
	_, require := assert.New(t), require.New(t)

	var buf bytes.Buffer
	app := &cli.Command{
		Writer:   &buf,
		Commands: []*cli.Command{commands.ChatSend},
	}

	// Exercise
	ctx := context.Background()
	err := app.Run(ctx, []string{"chat", "send", "'Hello, How are you doing?'"})

	// Verify
	require.NoError(err)
	expected := "I am doing great, thank you!\n"
	require.Equal(expected, buf.String())
}

func TestChatSendCommand_WithNoArguments(t *testing.T) {
	// Setup
	t.Setenv(config.EnvKeyConfigPath, filepath.Join("testdata", "config.yaml"))
	_, require := assert.New(t), require.New(t)

	var buf bytes.Buffer
	app := &cli.Command{
		Writer:   &buf,
		Commands: []*cli.Command{commands.ChatSend},
	}

	// Exercise
	ctx := context.Background()
	err := app.Run(ctx, []string{"chat", "send"})

	// Verify
	require.Error(err)
	require.Contains(err.Error(), "message")
}

func TestChatSendCommand_WithModel(t *testing.T) {
	// Setup
	t.Setenv(config.EnvKeyConfigPath, filepath.Join("testdata", "config.yaml"))
	_, require := assert.New(t), require.New(t)

	var buf bytes.Buffer
	app := &cli.Command{
		Writer:   &buf,
		Commands: []*cli.Command{commands.ChatSend},
	}

	// Exercise
	ctx := context.Background()
	err := app.Run(ctx, []string{
		"chat", "send",
		"--model=o3-mini",
		"'Hello, How are you doing?'",
	})

	// Verify
	require.NoError(err)
	expected := "I am doing great, thank you!\n"
	require.Equal(expected, buf.String())
}

func TestChatSendCommand_WithPersona(t *testing.T) {
	// Setup
	t.Setenv(config.EnvKeyConfigPath, filepath.Join("testdata", "config.yaml"))
	_, require := assert.New(t), require.New(t)

	var buf bytes.Buffer
	app := &cli.Command{
		Writer:   &buf,
		Commands: []*cli.Command{commands.ChatSend},
	}

	// Exercise
	ctx := context.Background()
	err := app.Run(ctx, []string{
		"chat", "send",
		"--persona=someotherpersona",
		"'Hello, who are you?'",
	})

	// Verify
	require.NoError(err)
	expected := "I am a someotherpersona!\n"
	require.Equal(expected, buf.String())
}

func TestChatSendCommand_WithPersona_Unknown(t *testing.T) {
	// Setup
	t.Setenv(config.EnvKeyConfigPath, filepath.Join("testdata", "config.yaml"))
	_, require := assert.New(t), require.New(t)

	var buf bytes.Buffer
	app := &cli.Command{
		Writer:   &buf,
		Commands: []*cli.Command{commands.ChatSend},
	}

	// Exercise
	ctx := context.Background()
	err := app.Run(ctx, []string{
		"chat", "send",
		"--persona=unknownpersona",
		"'Hello, who are you?'",
	})

	// Verify
	require.Error(err)
	require.Contains(err.Error(), "unknownpersona")
}

func TestChatSendCommand_SessionCreation(t *testing.T) {
	// Setup
	conf := commands.PrepareConfig(t)
	_, require := assert.New(t), require.New(t)

	var buf bytes.Buffer
	app := &cli.Command{
		Writer:   &buf,
		Commands: []*cli.Command{commands.ChatSend},
	}

	// Exercise
	ctx := context.Background()
	err := app.Run(ctx, []string{
		"chat", "send", "'Hello, How are you doing?'",
	})

	// Verify
	require.NoError(err)

	// Check if the session is created under
	require.FileExists(conf.Location())
	require.DirExists(filepath.Join(filepath.Dir(conf.Location()), conf.Chat.Session.Directory))
}
