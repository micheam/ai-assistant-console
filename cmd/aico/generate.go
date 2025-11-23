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

	// Persona:
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

	// Source:
	if source != "" {
		sb := new(strings.Builder)
		sb.WriteString("In the following <source> block is the context information for the prompt.\n\n")
		sb.WriteString("<source>\n")
		sb.WriteString(source)
		sb.WriteString("\n</source>\n")
		systemInstruction = append(systemInstruction, assistant.NewTextContent(sb.String()))
	}

	// Context:
	contexts := cmd.StringSlice(flagContext.Name)
	logger = logger.With(slog.String("contexts", strings.Join(contexts, ",")))
	if len(contexts) > 0 {
		systemInstruction = append(systemInstruction,
			assistant.NewTextContent("The following context is provided for the prompt."))
	}
	for _, ctx := range contexts {
		content, err := resolveContext(ctx)
		if err != nil {
			return fmt.Errorf("failed to resolve context %q: %w", ctx, err)
		}
		systemInstruction = append(systemInstruction, assistant.NewTextContent(content))
	}

	// Wrap up system instructions:
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

// resolveContext resolves a context string.
// If the string starts with '@', it reads from the file path after '@'.
// Otherwise, it returns the string as-is.
func resolveContext(ctx string) (string, error) {
	if strings.HasPrefix(ctx, "@") {
		filePath := strings.TrimPrefix(ctx, "@")
		data, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to read file %q: %w", filePath, err)
		}
		sb := new(strings.Builder)
		fmt.Fprintf(sb, "<context file=%q>\n", filePath)
		fmt.Fprint(sb, string(data))
		fmt.Fprintln(sb, "\n</context>")
		return sb.String(), nil
	}
	// Direct string context
	sb := new(strings.Builder)
	sb.WriteString("<context>\n")
	sb.WriteString(ctx)
	sb.WriteString("\n</context>")
	return sb.String(), nil
}

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
