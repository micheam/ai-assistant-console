package commands

import (
	"path/filepath"
	"testing"

	"micheam.com/aico/internal/config"
)

var ConfigInit = configInit

func PrepareConfig(t *testing.T) *config.Config {
	t.Helper()
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	t.Setenv(config.EnvKeyConfigPath, configPath) // non-existent file
	conf, err := config.InitAndLoad()
	if err != nil {
		t.Fatal(err)
	}
	return conf
}
