package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"

	"micheam.com/aico/internal/config"
)

// runCompleteApp runs CmdComplete with the global --json flag registered on the root,
// capturing stdout into buf.
func runCompleteApp(t *testing.T, buf *bytes.Buffer, args ...string) error {
	t.Helper()
	app := &cli.Command{
		Name:     "aico",
		Writer:   buf,
		Flags:    []cli.Flag{flagJSON},
		Commands: []*cli.Command{CmdComplete},
	}
	return app.Run(context.Background(), append([]string{"aico"}, args...))
}

// writeCompleteConfig writes src as the config file and points
// AI_ASSISTANT_CONFIG_PATH at it.
func writeCompleteConfig(t *testing.T, src string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.toml")
	require.NoError(t, os.WriteFile(path, []byte(src), 0o600))
	t.Setenv(config.EnvKeyConfigPath, path)
}

const completeProfilesConfig = `
[complete]
default_profile = "rack"

[complete.profiles.rack]
endpoint = "http://rack.local:8012/infill"
model = "qwen2.5-coder-7b"

[complete.profiles.local]
endpoint = "http://127.0.0.1:8012/infill"
`

func TestCompleteProfiles_JSON(t *testing.T) {
	writeCompleteConfig(t, completeProfilesConfig)

	var buf bytes.Buffer
	require.NoError(t, runCompleteApp(t, &buf, "--json", "complete", "profiles"))
	assert.JSONEq(t, `[
		{"name": "local", "endpoint": "http://127.0.0.1:8012/infill", "model": "", "default": false},
		{"name": "rack", "endpoint": "http://rack.local:8012/infill", "model": "qwen2.5-coder-7b", "default": true}
	]`, buf.String())
}

func TestCompleteProfiles_Text(t *testing.T) {
	writeCompleteConfig(t, completeProfilesConfig)

	var buf bytes.Buffer
	require.NoError(t, runCompleteApp(t, &buf, "complete", "profiles"))
	assert.Equal(t,
		"local http://127.0.0.1:8012/infill\n"+
			"rack (qwen2.5-coder-7b) http://rack.local:8012/infill *\n",
		buf.String())
}

func TestCompleteProfiles_JSON_NoConfigFile(t *testing.T) {
	t.Setenv(config.EnvKeyConfigPath, filepath.Join(t.TempDir(), "no-such-config.toml"))

	var buf bytes.Buffer
	require.NoError(t, runCompleteApp(t, &buf, "--json", "complete", "profiles"))
	assert.Equal(t, "[]\n", buf.String())
}
