package main

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"

	"micheam.com/aico/internal/config"
)

func TestConfigCommand_NoArguments(t *testing.T) {
	// Setup
	t.Setenv(config.EnvKeyConfigPath, filepath.Join("testdata", "config.yaml"))
	_, require := assert.New(t), require.New(t)

	var buf bytes.Buffer
	app := &cli.Command{
		Writer:   &buf,
		Commands: []*cli.Command{CmdConfig},
	}

	// Exercise
	ctx := context.Background()
	err := app.Run(ctx, []string{"chat", "config", "path"})

	// Verify
	require.NoError(err)
	expected := config.ConfigFilePath() + "\n"
	require.Equal(expected, buf.String())
}

func TestConfigCommand_Initialize(t *testing.T) {
	// Setup
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	t.Setenv(config.EnvKeyConfigPath, configPath) // non-existent file
	_, require := assert.New(t), require.New(t)

	var buf bytes.Buffer
	app := &cli.Command{
		Writer:   &buf,
		Commands: []*cli.Command{CmdConfig},
	}

	// Exercise
	ctx := context.Background()
	err := app.Run(ctx, []string{"chat", "config", "init"})

	// Verify
	require.NoError(err)
	require.FileExists(configPath)
	require.Contains(buf.String(), "Configuration file initialized\n")
}

func TestConfigCommand_InitializeWithSpecifiedPath(t *testing.T) {
	// Setup
	configPath := filepath.Join(t.TempDir(), "config2.yaml")
	_, require := assert.New(t), require.New(t)

	var buf bytes.Buffer
	app := &cli.Command{
		Writer:   &buf,
		Commands: []*cli.Command{CmdConfig},
	}

	// Exercise
	ctx := context.Background()
	err := app.Run(ctx, []string{"config", "init", "-path", configPath})

	// Verify
	require.NoError(err)
	require.FileExists(configPath)
	require.Contains(buf.String(), "Configuration file initialized\n")
}
