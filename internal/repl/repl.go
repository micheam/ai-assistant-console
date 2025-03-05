package repl

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"micheam.com/aico/internal/assistant"
	"micheam.com/aico/internal/config"
	"micheam.com/aico/internal/logging"
)

var defaultUserPrompt = func(ctx context.Context) string { return "USER: " }
var defaultAssistantPrompt = func(ctx context.Context) string { return "ASSISTANT: " }

type Repl struct {
	Config          *config.Config
	Model           assistant.GenerativeModel
	PersonaName     string
	UserPrompt      func(ctx context.Context) string
	AssistantPrompt func(ctx context.Context) string

	Out io.Writer
	Err io.Writer
}

func Init(conf *config.Config, personaName string, model assistant.GenerativeModel) *Repl {
	return &Repl{
		Config:          conf,
		Model:           model,
		PersonaName:     personaName,
		UserPrompt:      defaultUserPrompt,
		AssistantPrompt: defaultAssistantPrompt,

		Out: os.Stdout,
		Err: os.Stderr,
	}
}

// Run starts the interactive Read-Eval-Print Loop (REPL) mode.
// This will block until the user sends EOF (Ctrl-D) or types :quit or :exit.
func (r *Repl) Run(ctx context.Context) error {
	logger := logging.LoggerFrom(ctx)

	r.outputf("Chat with an AI assistant.\n")
	r.outputf("Type :quit or :exit to exit.\n")
	r.outputf("Type a lessage and end with a double semicolon (;;).\n")

	r.outputf("Model: %s\n", r.Config.Chat.Model)
	r.outputf("Persona: %s\n", r.PersonaName)
	r.outputf("\n")

	reader := bufio.NewReader(os.Stdin)
	var lines []string
	for {
		fmt.Print(defaultUserPrompt(ctx))
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		line = strings.TrimSpace(line)
		if line == ":quit" || line == ":exit" {
			return nil
		}
		if strings.HasSuffix(line, ";;") {
			lines = append(lines, line)
			query := strings.Join(lines, " ")

			// Start New Chat Session
			sess, err := assistant.StartChat(r.Model)
			if err != nil {
				return fmt.Errorf("start chat: %w", err)
			}
			sess.SetSystemInstruction(
				assistant.NewTextContent(strings.Join(r.resolvePersona().Messages, "\n")))
			// Send message to assistant
			ctx := logging.ContextWith(ctx, logger)
			resp, err := sess.SendMessage(ctx, assistant.NewTextContent(query))
			if err != nil {
				r.errorf("send message: %v\n", err)
			}

			// Print response
			switch v := resp.Content.(type) {
			case *assistant.TextContent:
				r.output(r.AssistantPrompt(ctx))
				r.outputf("%s\n", v.Text)
			default:
				r.errorf("unexpected response type: %T\n", v)
			}

			lines = nil
		} else {
			lines = append(lines, line)
		}
	}
	return nil
}

func (r *Repl) resolvePersona() *config.Personality {
	p, ok := r.Config.Chat.GetPersona(r.PersonaName)
	if !ok {
		return r.Config.Chat.GetDefaultPersona()
	}
	return p
}

func (r *Repl) output(msg string) {
	fmt.Fprint(r.Out, msg)
}

func (r *Repl) outputf(format string, a ...any) {
	fmt.Fprintf(r.Out, format, a...)
}

func (r *Repl) errorf(format string, a ...any) {
	fmt.Fprintf(r.Err, "ERROR: "+format, a...)
}
