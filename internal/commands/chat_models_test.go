package commands_test

import (
	"bytes"
	"encoding/json"
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
claude-3-5-sonnet
claude-3-5-haiku
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

	// Verify - each line is a JSON object
	require.NoError(err)
	for _, line := range bytes.Split(buf.Bytes(), []byte{'\n'}) {
		if len(line) == 0 {
			continue
		}
		data := map[string]any{}
		require.NoError(json.Unmarshal(line, &data))
		require.Contains(data, "name")
		require.Contains(data, "selected")
		require.Contains(data, "description")
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
	err := app.Run([]string{"chat", "models", "--model", "claude-3-7-sonnet"})

	// Verify
	require.NoError(err)
	expected := `gpt-4o
gpt-4o-mini
o1
o1-mini
o3-mini
claude-3-7-sonnet *
claude-3-5-sonnet
claude-3-5-haiku
`
	if diff := cmp.Diff(expected, buf.String()); diff != "" {
		t.Errorf("Unexpected output: (-got +want)\n%s", diff)
	}
}
