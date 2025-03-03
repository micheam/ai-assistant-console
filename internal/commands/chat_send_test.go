package commands_test

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	"micheam.com/aico/internal/commands"
	"micheam.com/aico/internal/config"
)

func TestChatSendCommand(t *testing.T) {
	// Setup
	t.Setenv(config.EnvKeyConfigPath, filepath.Join("testdata", "config.yaml"))
	_, require := assert.New(t), require.New(t)

	var buf bytes.Buffer
	app := &cli.App{
		Writer:   &buf,
		Commands: []*cli.Command{commands.ChatSend},
	}

	// Exercise
	err := app.Run([]string{"chat", "send", "'Hello, How are you doing?'"})

	// Verify
	require.NoError(err)
	expected := "I am doing great, thank you!\n"
	require.Equal(expected, buf.String())

	// Verify: Session must be stored
}

func TestChatSendCommand_WithNoArguments(t *testing.T) {
	// Setup
	t.Setenv(config.EnvKeyConfigPath, filepath.Join("testdata", "config.yaml"))
	_, require := assert.New(t), require.New(t)

	var buf bytes.Buffer
	app := &cli.App{
		Writer:   &buf,
		Commands: []*cli.Command{commands.ChatSend},
	}

	// Exercise
	err := app.Run([]string{"chat", "send"})

	// Verify
	require.Error(err)
	require.Contains(err.Error(), "message")
}

func TestChatSendCommand_WithModel(t *testing.T) {
	// Setup
	t.Setenv(config.EnvKeyConfigPath, filepath.Join("testdata", "config.yaml"))
	_, require := assert.New(t), require.New(t)

	var buf bytes.Buffer
	app := &cli.App{
		Writer:   &buf,
		Commands: []*cli.Command{commands.ChatSend},
	}

	// Exercise
	err := app.Run([]string{
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
	app := &cli.App{
		Writer:   &buf,
		Commands: []*cli.Command{commands.ChatSend},
	}

	// Exercise
	err := app.Run([]string{
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
	app := &cli.App{
		Writer:   &buf,
		Commands: []*cli.Command{commands.ChatSend},
	}

	// Exercise
	err := app.Run([]string{
		"chat", "send",
		"--persona=unknownpersona",
		"'Hello, who are you?'",
	})

	// Verify
	require.Error(err)
	require.Contains(err.Error(), "unknownpersona")
}
