package main

import (
	"context"
	"encoding/json"
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
	return doGenerate(ctx, cmd, cmd.Args().First())
}

func doGenerate(ctx context.Context, cmd *cli.Command, prompt string) error {
	logger, cleanup, err := initializeLogger(ctx, cmd)
	if err != nil {
		return err
	}
	defer cleanup()

	sess, err := loadSession(cmd)
	if err != nil {
		return fmt.Errorf("failed to load session: %w", err)
	}

	logger = logger.With(slog.String("session_id", sess.ID))
	ctx = logging.ContextWith(ctx, logger)

	{
		userContents := []assistant.MessageContent{}
		source, err := detectSource(cmd.String(flagSource.Name), os.Stdin)
		if err != nil {
			return err
		}
		if source != "" {
			userContents = append(userContents, assistant.NewTextContent(source))
		}
		if prompt := cmd.Args().First(); prompt != "" {
			userContents = append(userContents, assistant.NewTextContent(prompt))
		}
		userMsg := assistant.NewUserMessage(userContents...)
		sess.AddMessage(userMsg)
	}

	model, err := modelByName(cmd, sess.Model)
	if err != nil {
		return fmt.Errorf("model by name: %w", err)
	}
	model.SetSystemInstruction(sess.SystemInstruction...)
	defer sess.Save(ctx, model)

	iter, err := model.GenerateContentStream(ctx, sess.GetMessages()...)
	if err != nil {
		return fmt.Errorf("failed to generate content: %w", err)
	}

	// Stream content and accumulate text for session history
	var (
		acc    = new(strings.Builder)
		writer = detectWriter(cmd, *sess)
	)
	defer writer.Close()
	for resp, err := range iter {
		if err != nil {
			fmt.Fprintf(cmd.ErrWriter, "\nError: %v\n", err)
			return fmt.Errorf("stream error: %w", err)
		}
		switch content := resp.Content.(type) {
		case *assistant.TextContent:
			_, err := writer.Write([]byte(content.Text))
			if err != nil {
				return fmt.Errorf("failed to write content: %w", err)
			}
			acc.WriteString(content.Text)
		default:
			// Ignore other content types for now
			logger.Warn("ignore unsupported content type",
				"type", fmt.Sprintf("%T", content))
		}
	}
	if acc.Len() > 0 {
		sess.AddMessage(assistant.NewAssistantMessage(assistant.NewTextContent(acc.String())))
	}
	return nil
}

// generateView is the JSON output shape for the generate action when --json is set.
type generateView struct {
	Session string `json:"session"`
	Model   string `json:"model"`
	Content string `json:"content"`
}

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------

// detectSource returns the source content from either --source flag or stdin.
// It returns an error if both are specified.
func detectSource(srcFlag string, stdin io.Reader) (string, error) {
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
	if after, ok := strings.CutPrefix(src, "@"); ok {
		filePath := after
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
	if after, ok := strings.CutPrefix(ctx, "@"); ok {
		filePath := after
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

type SessionMode int

const (
	SessionModeNew SessionMode = iota
	SessionModeLast
	SessionModeExisting
)

func detectSessionMode(cmd *cli.Command) (SessionMode, error) {
	givenSessionID := cmd.String(flagSessionID.Name)
	useLast := cmd.Bool(flagLast.Name)

	if givenSessionID != "" && useLast {
		return SessionMode(0), fmt.Errorf("--session and --last are mutually exclusive")
	}
	if givenSessionID != "" {
		return SessionModeExisting, nil
	}
	if useLast {
		return SessionModeLast, nil
	}
	return SessionModeNew, nil
}

func loadSession(cmd *cli.Command) (*assistant.Session, error) {
	conf, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("can't load config: %w", err)
	}

	sessMode, err := detectSessionMode(cmd)
	if err != nil {
		return nil, err
	}
	givenSessionID := cmd.String(flagSessionID.Name)

	switch sessMode {
	case SessionModeLast:
		return assistant.LoadLatestSession(conf.GetSessionDir())
	case SessionModeExisting:
		return assistant.LoadSession(conf.GetSessionDir(), givenSessionID)
	case SessionModeNew:
		sess := assistant.NewSession(conf.GetSessionDir())
		{ // Model
			model, err := detectModel(cmd)
			if err != nil {
				return nil, fmt.Errorf("detect model: %w", err)
			}
			sess.Model = QualifiedName(model.Provider(), model.Name())
		}
		{ // Persona
			personaName := cmd.String(flagPersona.Name)
			persona, ok := conf.PersonaMap[personaName]
			if !ok {
				return nil, fmt.Errorf("persona %q not found", cmd.String(flagPersona.Name))
			}
			sess.SystemInstruction = append(sess.SystemInstruction, assistant.NewTextContent(persona.Message))
		}
		{ // Contexts
			contexts := cmd.StringSlice(flagContext.Name)
			instructions := make([]*assistant.TextContent, 0)
			if len(contexts) > 0 {
				instructions = append(instructions,
					assistant.NewTextContent("The following context is provided for the prompt."))
			}
			for _, ctx := range contexts {
				content, err := resolveContext(ctx)
				if err != nil {
					return nil, fmt.Errorf("failed to resolve context %q: %w", ctx, err)
				}
				instructions = append(instructions, assistant.NewTextContent(content))
			}
			sess.SystemInstruction = append(sess.SystemInstruction, instructions...)
		}
		return sess, nil
	default:
		return nil, fmt.Errorf("unsupported session_mode(%v)", sessMode)
	}
}

func detectWriter(cmd *cli.Command, sess assistant.Session) io.WriteCloser {
	if cmd.Bool(flagJSON.Name) {
		return &JSONLineStreamWriter{
			enc: json.NewEncoder(cmd.Writer),
			metaData: struct {
				Session string
				Model   string
			}{
				Session: sess.ID,
				Model:   sess.Model,
			},
		}
	}
	return &ConsoleLineStreamWriter{
		out: cmd.Writer,
	}
}
