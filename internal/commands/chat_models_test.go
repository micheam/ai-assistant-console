package commands_test

import (
	"bytes"
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"

	commands "micheam.com/aico/internal/commands"
	"micheam.com/aico/internal/config"
)

func TestModelsCommand_Plain(t *testing.T) {
	// Setup
	t.Setenv(config.EnvKeyConfigPath, filepath.Join("testdata", "config.yaml"))
	_, require := assert.New(t), require.New(t)

	var buf bytes.Buffer
	cmd := &cli.Command{
		Writer:   &buf,
		Commands: []*cli.Command{commands.ChatModels},
	}

	// Exercise
	err := cmd.Run(context.Background(), []string{"chat", "models"})

	// Verify
	require.NoError(err)
	output := buf.String()
	
	// Check that the selected model (gpt-4o-mini) has a mark
	require.Contains(output, "gpt-4o-mini *", "selected model should have an asterisk mark")
	
	// Check that other models don't have marks (basic verification)
	require.Contains(output, "gpt-4o\n", "gpt-4o should not have an asterisk mark")
}

func TestModelsCommand_JSON(t *testing.T) {
	// Setup
	t.Setenv(config.EnvKeyConfigPath, filepath.Join("testdata", "config.yaml"))
	_, require := assert.New(t), require.New(t)

	var buf bytes.Buffer
	cmd := &cli.Command{
		Writer:   &buf,
		Commands: []*cli.Command{commands.ChatModels},
	}

	// Exercise
	ctx := context.Background()
	err := cmd.Run(ctx, []string{"chat", "models", "--json"})

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
	cmd := &cli.Command{
		Writer:   &buf,
		Commands: []*cli.Command{commands.ChatModels},
	}

	// Exercise
	ctx := context.Background()
	err := cmd.Run(ctx, []string{"chat", "models", "--model", "claude-3-7-sonnet"})

	// Verify
	require.NoError(err)
	expected := `gpt-4o
gpt-4o-mini
o1
o1-mini
o3-mini
claude-opus-4
claude-sonnet-4
claude-3-7-sonnet *
claude-3-5-haiku
`
	if diff := cmp.Diff(expected, buf.String()); diff != "" {
		t.Errorf("Unexpected output: (-got +want)\n%s", diff)
	}
}
