package commands_test

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	"micheam.com/aico/internal/commands"
	"micheam.com/aico/internal/config"
)

func TestModelsCommand_Plain(t *testing.T) {
	// Setup
	t.Setenv(config.EnvKeyConfigPath, filepath.Join("testdata", "config.yaml"))
	_, require := assert.New(t), require.New(t)

	var buf bytes.Buffer
	app := &cli.App{
		Writer:   &buf,
		Commands: []*cli.Command{commands.ChatModels},
	}

	// Exercise
	err := app.Run([]string{"chat", "models"})

	// Verify
	require.NoError(err)
	expected := `gpt-4o
gpt-4o-mini *
o1
o1-mini
o3-mini
claude-3-7-sonnet
`
	if diff := cmp.Diff(expected, buf.String()); diff != "" {
		t.Errorf("Unexpected output: (-got +want)\n%s", diff)
	}
}

func TestModelsCommand_JSON(t *testing.T) {
	// Setup
	t.Setenv(config.EnvKeyConfigPath, filepath.Join("testdata", "config.yaml"))
	_, require := assert.New(t), require.New(t)

	var buf bytes.Buffer
	app := &cli.App{
		Writer:   &buf,
		Commands: []*cli.Command{commands.ChatModels},
	}

	// Exercise
	err := app.Run([]string{"chat", "models", "--json"})

	// Verify
	require.NoError(err)
	expected := `{"name":"gpt-4o","selected":false}
{"name":"gpt-4o-mini","selected":true}
{"name":"o1","selected":false}
{"name":"o1-mini","selected":false}
{"name":"o3-mini","selected":false}
{"name":"claude-3-7-sonnet","selected":false}
`
	if diff := cmp.Diff(expected, buf.String()); diff != "" {
		t.Errorf("Unexpected output: (-got +want)\n%s", diff)
	}
}

func TestModelsCommand_RespectSelected(t *testing.T) {
	// Setup
	t.Setenv(config.EnvKeyConfigPath, filepath.Join("testdata", "config.yaml"))
	_, require := assert.New(t), require.New(t)

	var buf bytes.Buffer
	app := &cli.App{
		Writer:   &buf,
		Commands: []*cli.Command{commands.ChatModels},
	}

	// Exercise
	err := app.Run([]string{"chat", "models", "--model", "o1"})

	// Verify
	require.NoError(err)
	expected := `gpt-4o
gpt-4o-mini
o1 *
o1-mini
o3-mini
claude-3-7-sonnet
`
	if diff := cmp.Diff(expected, buf.String()); diff != "" {
		t.Errorf("Unexpected output: (-got +want)\n%s", diff)
	}
}
