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
	"micheam.com/aico/internal/theme"
)

// PromptFunc generates the prompt string based on context.
type PromptFunc func(ctx context.Context) string

// Repl represents the interactive Read-Eval-Print Loop.
type Repl struct {
	Config      *config.Config
	Model       assistant.GenerativeModel
	PersonaName string

	Prompt1, Prompt2 PromptFunc

	In  io.Reader
	Out io.Writer
	Err io.Writer
}

// Init returns a new Repl configured with the given settings.
func Init(conf *config.Config, personaName string, model assistant.GenerativeModel) *Repl {
	return &Repl{
		Config:      conf,
		Model:       model,
		PersonaName: personaName,

		Prompt1: func(ctx context.Context) string { return theme.Bold(model.Name() + "=> ") },
		Prompt2: func(ctx context.Context) string { return theme.Bold(model.Name() + "-> ") },

		In:  os.Stdin,
		Out: os.Stdout,
		Err: os.Stderr,
	}
}

// Run starts the interactive Read-Eval-Print Loop.
// It blocks until the user sends EOF (Ctrl-D) or types \q.
func (r *Repl) Run(ctx context.Context) error {
	logger := logging.LoggerFrom(ctx)
	fmt.Fprintf(r.Out, theme.Info("type %s for help\n"), COMMAND_SHOW_HELP)

	reader := bufio.NewReader(r.In)
	var lines []string

	for {
		// Print the appropriate prompt
		r.printPrompt(ctx, lines)

		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		line = strings.TrimSpace(line)

		// Quit the application if requested
		if strings.HasSuffix(line, COMMAND_QUIT) {
			return nil
		}

		// Show help if requested
		if strings.HasSuffix(line, COMMAND_SHOW_HELP) {
			r.Help()
			continue
		}

		// Clear the lines if requested
		if strings.HasSuffix(line, COMMAND_CLEAR_LINES) {
			fmt.Fprintln(r.Out)
			lines = nil
			continue
		}

		// If the line indicates the end of a query, process it.
		if strings.HasSuffix(line, `;;`) {
			lines = append(lines, line)
			query := r.buildQuery(lines)

			// Start a new chat session.
			sess, err := assistant.StartChat(r.Model)
			if err != nil {
				return fmt.Errorf("start chat: %w", err)
			}

			// Prepare system instructions from the chosen persona.
			persona := r.resolvePersona()
			systemText := strings.Join(persona.Messages, "\n")
			sess.SetSystemInstruction(assistant.NewTextContent(systemText))

			// Use the same context with logger information.
			ctxWithLogger := logging.ContextWith(ctx, logger)
			iter, err := sess.SendMessageStream(ctxWithLogger, assistant.NewTextContent(query))
			if err != nil {
				fmt.Fprintf(r.Err, theme.Error("[ERR] send message: %v\n"), err)
			} else {
				// Print streamed response.
				for resp := range iter {
					switch content := resp.Content.(type) {
					case *assistant.TextContent:
						fmt.Fprint(r.Out, theme.Reply(content.Text))
					default:
						fmt.Fprintf(r.Err, theme.Error("[ERR] unexpected response type: %T\n"), content)
					}
				}
			}
			fmt.Fprint(r.Out, "\n\n")
			lines = nil // Clear the collected lines for the next input
			continue
		}

		// Append the line to the buffer and continue reading.
		lines = append(lines, line)
	}

	return nil
}

const (
	COMMAND_SHOW_HELP   = `\?`
	COMMAND_QUIT        = `\q`
	COMMAND_CLEAR_LINES = `\c`
)

func (r *Repl) Help() {
	fmt.Fprintln(r.Out)
	fmt.Fprintf(r.Out, ";;  Execute the query\n")
	fmt.Fprintf(r.Out, "%s  Show this help message\n", COMMAND_SHOW_HELP)
	fmt.Fprintf(r.Out, "%s  Quit the application\n", COMMAND_QUIT)
	fmt.Fprintf(r.Out, "%s  Clear the query buffer\n", COMMAND_CLEAR_LINES)
	fmt.Fprintln(r.Out)
}

// printPrompt prints the primary or secondary prompt based on whether we are in multiline mode.
func (r *Repl) printPrompt(ctx context.Context, lines []string) {
	if len(lines) == 0 {
		fmt.Fprint(r.Out, r.Prompt1(ctx))
	} else {
		fmt.Fprint(r.Out, r.Prompt2(ctx))
	}
}

// buildQuery uses a strings.Builder to join all the lines into one query.
func (r *Repl) buildQuery(lines []string) string {
	var builder strings.Builder
	for i, line := range lines {
		// Add a space between lines (or newline if preferable).
		if i > 0 {
			builder.WriteString(" ")
		}
		builder.WriteString(line)
	}
	return builder.String()
}

// resolvePersona returns the matching personality or the default one.
func (r *Repl) resolvePersona() *config.Personality {
	if p, ok := r.Config.Chat.GetPersona(r.PersonaName); ok {
		return p
	}
	return r.Config.Chat.GetDefaultPersona()
}
