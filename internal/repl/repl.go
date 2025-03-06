package repl

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"micheam.com/aico/internal/assistant"
	"micheam.com/aico/internal/config"
	"micheam.com/aico/internal/logging"
	"micheam.com/aico/internal/spinner"
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
	Spinner          *spinner.Spinner

	In  io.Reader
	Out io.Writer
	Err io.Writer
}

// Init returns a new Repl configured with the given settings.
func Init(conf *config.Config, personaName string, model assistant.GenerativeModel) *Repl {
	spinner := spinner.New(
		100*time.Millisecond,
		[]string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
	)
	return &Repl{
		Config:      conf,
		Model:       model,
		PersonaName: personaName,

		Prompt1: func(ctx context.Context) string { return theme.Bold(model.Name() + "=> ") },
		Prompt2: func(ctx context.Context) string { return theme.Bold(model.Name() + "-> ") },
		Spinner: spinner,

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

	// Determine the session storage directory from configuration.
	confLocationDir := filepath.Dir(r.Config.Location())
	sessStoreDir, err := filepath.Abs(filepath.Join(confLocationDir, r.Config.Chat.Session.Directory))
	if err != nil {
		return fmt.Errorf("resolve session directory: %w", err)
	}

	// Start a persistent chat session.
	sess, err := assistant.StartChat(r.Model)
	if err != nil {
		return fmt.Errorf("start chat session: %w", err)
	}

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

		// Include a file as the context
		// Currently, we support only the textual content.
		if strings.HasPrefix(line, COMMAND_SOURCE_FILE) {
			// Extract the file path (anything after [COMMAND_SOURCE_FILE])
			fileLines, err := r.loadExternalFile(line)
			if err != nil {
				if err == ErrAbort {
					continue
				}
				return err
			}

			// Append the sourced file lines to the current query lines.
			lines = append(lines, fileLines...)
			continue
		}

		// If the line indicates that the user wants to edit the current query.
		if strings.HasSuffix(line, COMMAND_EDIT_QUERY) {
			editedLines, err := r.editQuery(lines)
			if err != nil {
				fmt.Fprintf(r.Err, theme.Error("[ERR] editing query: %v\n"), err)
			} else {
				lines = editedLines
			}
			continue
		}

		// If the line indicates the end of a query, process it.
		if strings.HasSuffix(line, COMMAND_END_QUERY) {
			r.Spinner.Start()
			defer r.Spinner.Stop()

			persona := r.resolvePersona()
			systemText := strings.Join(persona.Messages, "\n")
			sess.SetSystemInstruction(assistant.NewTextContent(systemText))

			lines = append(lines, line)
			query := r.buildQuery(lines)
			ctxWithLogger := logging.ContextWith(ctx, logger)
			iter, err := sess.SendMessageStream(ctxWithLogger, assistant.NewTextContent(query))
			if err != nil {
				// NOTE: We don't return here, as we want to continue the REPL.
				fmt.Fprintf(r.Err, theme.Error("[ERR] send message: %v\n"), err)
				continue
			}

			// Accumulate the stream response for storing
			//
			// TODO: Process the response stream and store only the final reply,
			//       possibly within ChatSession.SendMessageStream.
			var replyBuilder strings.Builder

			// Handle response stream
			for resp := range iter {
				r.Spinner.Stop()
				switch content := resp.Content.(type) {
				case *assistant.TextContent:
					fmt.Fprint(r.Out, theme.Reply(content.Text))
					replyBuilder.WriteString(content.Text)
				default:
					fmt.Fprintf(r.Err, theme.Error("[ERR] unexpected response type: %T\n"), content)
				}
			}
			// After streaming, add the complete reply to session history.
			completeReply := replyBuilder.String()
			c := assistant.NewTextContent(completeReply)
			sess.AddHistory(&assistant.GenerateContentResponse{Content: c})
			fmt.Fprint(r.Out, "\n\n")

			// Save the conversation (inputs and replies)
			if err := sess.Save(sessStoreDir); err != nil {
				fmt.Fprintf(r.Err, theme.Error("[ERR] saving session: %v\n"), err)
			}

			lines = nil // Clear the collected lines for the next input
			continue
		}

		// Append the line to the buffer and continue reading.
		lines = append(lines, line)
	}

	return nil
}

var ErrAbort = fmt.Errorf("aborted")

func (r *Repl) loadExternalFile(line string) ([]string, error) {
	filePath := strings.TrimSpace(strings.TrimPrefix(line, COMMAND_SOURCE_FILE))
	if filePath == "" {
		fmt.Fprintln(r.Err, theme.Error("No file path provided for source command."))
		return nil, ErrAbort
	}

	// Read the file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(r.Err, theme.Error("[ERR] could not read file %s: %v\n"), filePath, err)
		return nil, ErrAbort
	}

	// Split the file content into lines.
	fileLines := strings.Split(string(content), "\n")
	// Remove a trailing empty line if present.
	if len(fileLines) > 0 && fileLines[len(fileLines)-1] == "" {
		fileLines = fileLines[:len(fileLines)-1]
	}

	// sandwich with code block
	//
	// file: <file-path>
	// ```<ext>
	// ...
	// <file-content>
	// ...
	// ```
	//
	fileLines = append([]string{
		fmt.Sprintf("file: %s", filePath),
		fmt.Sprintf("```%s", filepath.Ext(filePath)),
	}, fileLines...)
	fileLines = append(fileLines, "```")

	return fileLines, nil
}

const (
	COMMAND_END_QUERY   = `;;`
	COMMAND_SHOW_HELP   = `\?`
	COMMAND_QUIT        = `\q`
	COMMAND_CLEAR_LINES = `\c`
	COMMAND_EDIT_QUERY  = `\e`
	COMMAND_SOURCE_FILE = `\i`
)

func (r *Repl) Help() {
	fmt.Fprintln(r.Out)
	fmt.Fprintln(r.Out, theme.Info(";;          Execute the query"))
	fmt.Fprintf(r.Out, theme.Info("%s          Show this help message\n"), COMMAND_SHOW_HELP)
	fmt.Fprintf(r.Out, theme.Info("%s          Quit the application\n"), COMMAND_QUIT)
	fmt.Fprintf(r.Out, theme.Info("%s          Clear the query buffer\n"), COMMAND_CLEAR_LINES)
	fmt.Fprintf(r.Out, theme.Info("%s          Edit the current query\n"), COMMAND_EDIT_QUERY)
	fmt.Fprintf(r.Out, theme.Info("%s <PATH>   include a file as the context\n"), COMMAND_SOURCE_FILE)
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
	return strings.Join(lines, "\n")
}

// resolvePersona returns the matching personality or the default one.
func (r *Repl) resolvePersona() *config.Personality {
	if p, ok := r.Config.Chat.GetPersona(r.PersonaName); ok {
		return p
	}
	return r.Config.Chat.GetDefaultPersona()
}

// editQuery opens the user's preferred editor to edit the query.
func (r *Repl) editQuery(lines []string) ([]string, error) {
	// Create a temporary file.
	tmpFile, err := os.CreateTemp("", "AICO_CHAT_QUERY_*.txt")
	if err != nil {
		return nil, fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write current query into the temporary file.
	initialQuery := r.buildQuery(lines)
	if _, err := tmpFile.WriteString(initialQuery); err != nil {
		tmpFile.Close()
		return nil, fmt.Errorf("write to temp file: %w", err)
	}
	tmpFile.Close()

	// Determine the editor from the environment variable; fallback to "vim"
	editor, ok := os.LookupEnv("EDITOR")
	if !ok {
		editor = "vim"
	}

	// Launch the editor.
	cmd := exec.Command(editor, tmpFile.Name())
	// Connect the editor's stdio to the user's terminal.
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("editor error: %w", err)
	}

	// Open the temporary file for reading the updated content.
	editedContent, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return nil, fmt.Errorf("read edited file: %w", err)
	}

	// Split the edited content into lines. Adjust the separator if necessary.
	editedLines := strings.Split(string(editedContent), "\n")
	// Optionally remove any trailing empty element if the file ended with newline.
	if len(editedLines) > 0 && editedLines[len(editedLines)-1] == "" {
		editedLines = editedLines[:len(editedLines)-1]
	}
	return editedLines, nil
}
