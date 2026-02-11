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

	prompt := cmd.Args().First()
	if prompt == "" {
		return fmt.Errorf("prompt is required")
	}

	// Session Management:
	sess, err := loadSession(logging.ContextWith(ctx, logger), cmd)
	if err != nil {
		return fmt.Errorf("failed to load session: %w", err)
	}
	logger = logger.With(slog.String("session_id", sess.ID))
	ctx = logging.ContextWith(ctx, logger)

	// Source: --source flag or stdin (mutually exclusive)
	srcFlag := cmd.String(flagSource.Name)
	source, err := getSource(srcFlag, os.Stdin)
	if err != nil {
		return err
	}
	if srcFlag != "" {
		logger = logger.With(slog.String("source", srcFlag))
	}

	model, err := detectModel(ctx, cmd)
	logger = logger.With(slog.String("model", model.Name()))
	if err != nil {
		return fmt.Errorf("failed to detect model: %w", err)
	}

	// Persona / System Instruction:
	systemInstruction := make([]*assistant.TextContent, 0)
	conf, err := config.Load()
	if err != nil {
		conf = config.DefaultConfig()
	}
	isResumedSession := cmd.String(flagSessionID.Name) != ""
	if isResumedSession && len(sess.SystemInstruction) > 0 {
		// Existing session: use saved instructions (maintains persona consistency)
		systemInstruction = append(systemInstruction, sess.SystemInstruction...)
	} else {
		// New session or legacy session: build from config
		persona, ok := conf.PersonaMap[cmd.String(flagPersona.Name)]
		if !ok {
			return fmt.Errorf("persona %q not found", cmd.String(flagPersona.Name))
		}
		systemInstruction = append(systemInstruction, assistant.NewTextContent(persona.Message))
		logger = logger.With(slog.String("persona", cmd.String(flagPersona.Name)))
	}

	// Save base system instruction (without per-invocation context) for session persistence
	baseSystemInstruction := make([]*assistant.TextContent, len(systemInstruction))
	copy(baseSystemInstruction, systemInstruction)

	// Context (per-invocation, not persisted in session):
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
	logger.Debug("prepared system instruction", "system_instruction", systemInstruction)

	// Generate Content
	// Source is included as the first content of the user message so it persists in session history
	userContents := make([]assistant.MessageContent, 0, 2)
	if source != "" {
		userContents = append(userContents, assistant.NewTextContent(source))
	}
	userContents = append(userContents, assistant.NewTextContent(prompt))
	userMsg := assistant.NewUserMessage(userContents...)
	sess.AddMessage(userMsg)
	logger.Debug("sending generate request", "history_len", len(sess.Messages))
	iter, err := model.GenerateContentStream(ctx, sess.GetMessages()...)
	if err != nil {
		return fmt.Errorf("failed to generate content: %w", err)
	}

	// Print the generated content
	acc := strings.Builder{}
	for resp, err := range iter {
		if err != nil {
			fmt.Fprintf(cmd.ErrWriter, "\nError: %v\n", err)
			return fmt.Errorf("stream error: %w", err)
		}
		switch content := resp.Content.(type) {
		case *assistant.TextContent:
			fmt.Fprintf(cmd.Writer, "%s", content.Text)
			acc.WriteString(content.Text)
		default:
			// Ignore other content types for now
			logger.Warn("ignore unsupported content type", "type", fmt.Sprintf("%T", content))
		}
	}
	fmt.Fprintln(cmd.Writer)

	if acc.Len() > 0 {
		sess.AddMessage(assistant.NewAssistantMessage(assistant.NewTextContent(acc.String())))
	}
	// Persist base system instruction for new sessions (not overwriting existing saved instruction)
	if len(sess.SystemInstruction) == 0 {
		sess.SystemInstruction = baseSystemInstruction
	}
	if err := sess.Save(ctx, model); err != nil {
		return err
	}
	fmt.Fprintf(cmd.ErrWriter, "session: %s\n", sess.ID)
	return nil
}

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------

// getSource returns the source content from either --source flag or stdin.
// It returns an error if both are specified.
func getSource(srcFlag string, stdin io.Reader) (string, error) {
	stdinContent, err := readSource(stdin)
	if err != nil {
		return "", fmt.Errorf("failed to read source from stdin: %w", err)
	}

	if srcFlag != "" && stdinContent != "" {
		return "", fmt.Errorf("cannot specify both --source flag and stdin input")
	}

	if srcFlag != "" {
		return resolveSource(srcFlag)
	}

	if stdinContent == "" {
		return "", nil
	}

	sb := new(strings.Builder)
	sb.WriteString("<source>\n")
	sb.WriteString(stdinContent)
	sb.WriteString("\n</source>")
	return sb.String(), nil
}

// resolveSource resolves a source string.
// If the string starts with '@', it reads from the file path after '@'.
// Otherwise, it returns the string as-is wrapped in <source> tags.
func resolveSource(src string) (string, error) {
	if strings.HasPrefix(src, "@") {
		filePath := strings.TrimPrefix(src, "@")
		data, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to read file %q: %w", filePath, err)
		}
		sb := new(strings.Builder)
		fmt.Fprintf(sb, "<source file=%q>\n", filePath)
		fmt.Fprint(sb, string(data))
		fmt.Fprintln(sb, "\n</source>")
		return sb.String(), nil
	}
	// Direct string source
	sb := new(strings.Builder)
	sb.WriteString("<source>\n")
	sb.WriteString(src)
	sb.WriteString("\n</source>")
	return sb.String(), nil
}

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

func loadSession(ctx context.Context, cmd *cli.Command) (*assistant.Session, error) {
	logger := logging.LoggerFrom(ctx)
	conf, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("can't load config: %w", err)
	}
	dir := conf.GetSessionDir()
	if cmd.String(flagSessionID.Name) == "" {
		logger.Debug("creating new session")
		return assistant.NewSession(dir), nil
	}
	id := cmd.String(flagSessionID.Name)
	existing, err := assistant.LoadSession(ctx, dir, id)
	if err != nil {
		return nil, fmt.Errorf("load existing session %q: %w", id, err)
	}
	logger.Debug("loaded existing session", "session_id", id)
	return existing, nil
}
