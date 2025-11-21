package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/urfave/cli/v3"
	"golang.org/x/term"

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

	// Read source from stdin if available (piped input)
	source, err := readSource(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to read source from stdin: %w", err)
	}

	ctx = logging.ContextWith(ctx, logger)
	model, err := detectModel(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to detect model: %w", err)
	}

	// Build the message content
	messageText := buildMessageText(prompt, source)
	msg := assistant.NewUserMessage(assistant.NewTextContent(messageText))
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

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------

// readSource reads content from stdin if it's piped (not a terminal).
// Returns empty string if stdin is a terminal.
func readSource(r io.Reader) (string, error) {
	// Check if stdin is a terminal
	if f, ok := r.(*os.File); ok {
		if term.IsTerminal(int(f.Fd())) {
			return "", nil
		}
	}

	// Read all content from stdin
	data, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// buildMessageText constructs the final message text combining prompt and source.
func buildMessageText(prompt, source string) string {
	if source == "" {
		return prompt
	}
	return fmt.Sprintf("%s\n\n---\n<source>\n%s</source>", prompt, source)
}
