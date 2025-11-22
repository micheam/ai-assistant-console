package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/urfave/cli/v3"
	"golang.org/x/term"

	"micheam.com/aico/internal/assistant"
	"micheam.com/aico/internal/config"
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
	ctx = logging.ContextWith(ctx, logger)

	prompt := cmd.Args().First()
	if prompt == "" {
		return fmt.Errorf("prompt is required")
	}

	source, err := readSource(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to read source from stdin: %w", err)
	}

	model, err := detectModel(ctx, cmd)
	logger = logger.With(slog.String("model", model.Name()))
	if err != nil {
		return fmt.Errorf("failed to detect model: %w", err)
	}

	// Set Personalities as System Instructions
	systemInstruction := make([]*assistant.TextContent, 0)
	conf, err := config.Load()
	if err != nil {
		conf = config.DefaultConfig()
	}
	persona, ok := conf.PersonaMap[cmd.String(flagPersona.Name)]
	if !ok {
		return fmt.Errorf("persona %q not found", cmd.String(flagPersona.Name))
	}
	systemInstruction = append(systemInstruction, assistant.NewTextContent(persona.Message))
	logger = logger.With(slog.String("persona", cmd.String(flagPersona.Name)))

	// Append Source Context as System Instruction
	if source != "" {
		sb := new(strings.Builder)
		sb.WriteString("In the following <source> block is the context information for the prompt.\n\n")
		sb.WriteString("<source>\n")
		sb.WriteString(source)
		sb.WriteString("\n</source>\n")
		// model.SetSystemInstruction(assistant.NewTextContent(sb.String()))
		systemInstruction = append(systemInstruction, assistant.NewTextContent(sb.String()))
	}
	model.SetSystemInstruction(systemInstruction...)

	// Generate Content
	msg := assistant.NewUserMessage(assistant.NewTextContent(prompt))
	logger.Debug("sending generate request")
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
