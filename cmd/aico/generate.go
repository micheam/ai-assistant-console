package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

// -----------------------------------------------------------------------------
// Actions
// -----------------------------------------------------------------------------

func runGenerate(_ context.Context, cmd *cli.Command, args []string) error {
	prompt := args[0]
	model := cmd.String("model")
	contextFile := cmd.String("context-file")

	fmt.Printf("Generating text for: %s\n", prompt)
	fmt.Printf("Model: %s\n", model)
	fmt.Printf("Context file: %s\n", contextFile)

	// TODO: 実装
	return nil
}
