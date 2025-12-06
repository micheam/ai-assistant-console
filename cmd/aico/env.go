package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli/v3"

	"micheam.com/aico/internal/config"
)

var CmdEnv = &cli.Command{
	Name:   "env",
	Usage:  "show environment information",
	Action: runShowEnv,
}

// -----------------------------------------------------------------------------
// Actions
// -----------------------------------------------------------------------------

func runShowEnv(ctx context.Context, cmd *cli.Command) error {
	var model string
	if conf, err := config.Load(); err != nil {
		model = "Not-loaded"
	} else {
		model = conf.Model
	}

	fmt.Printf("Default Model: %s\n", model)
	fmt.Printf("Config file: %s\n", config.ConfigFilePath())

	fmt.Printf("ANTHROPIC_API_KEY: %s\n", maskAPIKey(os.Getenv("ANTHROPIC_API_KEY")))
	fmt.Printf("OPENAI_API_KEY: %s\n", maskAPIKey(os.Getenv("OPENAI_API_KEY")))
	fmt.Printf("CEREBRAS_API_KEY: %s\n", maskAPIKey(os.Getenv("CEREBRAS_API_KEY")))
	return nil
}

func maskAPIKey(key string) string {
	if key == "" {
		return "(not set)"
	}
	if len(key) <= 8 {
		return "***"
	}
	return key[:4] + "..." + key[len(key)-4:]
}

// -----------------------------------------------------------------------------
// Helper functions
// -----------------------------------------------------------------------------

func envKeyWithPrefix(prefix, key string) string {
	prefix = strings.ToUpper(prefix)
	key = strings.ToUpper(key)
	return strings.Join([]string{prefix, key}, "_")
}
