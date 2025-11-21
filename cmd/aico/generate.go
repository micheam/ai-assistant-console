package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"

	"micheam.com/aico/internal/assistant"
	"micheam.com/aico/internal/logging"
)

// -----------------------------------------------------------------------------
// Actions
// -----------------------------------------------------------------------------

func runGenerate(ctx context.Context, cmd *cli.Command) error {
	logger, cleanup, err := initializeLogger(ctx, cmd)
	if err != nil {
		return err
	}
	defer cleanup()

	prompt := cmd.Args().First()
	if prompt == "" {
		return fmt.Errorf("prompt is required")
	}

	ctx = logging.ContextWith(ctx, logger)
	model, err := detectModel(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to detect model: %w", err)
	}

	msg := assistant.NewUserMessage(assistant.NewTextContent(prompt))
	iter, err := model.GenerateContentStream(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to generate content: %w", err)
	}

	// Print the generated content
	for resp := range iter {
		switch content := resp.Content.(type) {
		case *assistant.TextContent:
			fmt.Fprintf(cmd.Writer, "%s", content.Text)
		default:
			// Ignore other content types for now
			logger.Warn("ignore unsupported content type", "type", fmt.Sprintf("%T", content))
		}
	}
	return nil
}
