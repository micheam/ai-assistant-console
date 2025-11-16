package main

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
)

var CmdEnv = &cli.Command{
	Name:   "env",
	Usage:  "show environment information",
	Action: runShowEnv,
}

// -----------------------------------------------------------------------------
// Actions
// -----------------------------------------------------------------------------

func runShowEnv(_ context.Context, cmd *cli.Command) error {
	fmt.Println("Environment Information:")
	fmt.Printf("  AI_MODEL: %s\n", os.Getenv("AI_MODEL"))
	fmt.Printf("  ANTHROPIC_API_KEY: %s\n", maskAPIKey(os.Getenv("ANTHROPIC_API_KEY")))
	fmt.Printf("  OPENAI_API_KEY: %s\n", maskAPIKey(os.Getenv("OPENAI_API_KEY")))
	fmt.Printf("  CEREBRAS_API_KEY: %s\n", maskAPIKey(os.Getenv("CEREBRAS_API_KEY")))
	fmt.Printf("  Config file: %s\n", "~/.config/aico/config.toml")
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
