package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/urfave/cli/v3"

	"micheam.com/aico/internal/assistant"
	"micheam.com/aico/internal/config"
)

var CmdSession = &cli.Command{
	Name:  "session",
	Usage: "Manage chat sessions",
	// See the comment on the root command's DisableSliceFlagSeparator in
	// main.go: this must be repeated here and on "resume" because
	// urfave/cli resets the (process-wide) slice separator setting from
	// each command's own field as it dispatches into subcommands.
	DisableSliceFlagSeparator: true,
	Commands: []*cli.Command{
		{
			Name:    "list",
			Aliases: []string{"ls"},
			Usage:   "List saved sessions",
			Action:  runSessionList,
			Flags: []cli.Flag{
				&cli.IntFlag{
					Name:    "limit",
					Aliases: []string{"n"},
					Usage:   "Maximum number of sessions to show",
					Value:   20,
				},
			},
		},
		{
			Name:      "show",
			Usage:     "Show messages of a session",
			ArgsUsage: "<session-id>",
			Action:    runSessionShow,
		},
		{
			Name:                      "resume",
			Usage:                     "Resume an existing session with a new prompt",
			ArgsUsage:                 "<session-id> <prompt>",
			Action:                    runSessionResume,
			DisableSliceFlagSeparator: true,
			Flags: []cli.Flag{
				flagSource,
				flagContext,
				flagModel,
				flagNoStream,
				flagDebug,
				flagPersona,
				flagTool,
			},
		},
	},
}

// sessionListItemView is the JSON representation of one row of `session list`.
type sessionListItemView struct {
	ID           string `json:"id"`
	UpdatedAt    string `json:"updated_at"`
	MessageCount int    `json:"message_count"`
	Preview      string `json:"preview"`
}

func runSessionList(ctx context.Context, cmd *cli.Command) error {
	conf, err := config.Load()
	if err != nil {
		conf = config.DefaultConfig()
	}
	dir := conf.GetSessionDir()

	summaries, err := assistant.ListSessions(dir)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("list sessions: %w", err)
	}

	limit := min(int(cmd.Int("limit")), len(summaries))
	summaries = summaries[:limit]

	if cmd.Bool(flagJSON.Name) {
		items := make([]sessionListItemView, 0, len(summaries))
		for _, s := range summaries {
			items = append(items, sessionListItemView{
				ID:           s.ID,
				UpdatedAt:    s.ModTime.Format(time.RFC3339),
				MessageCount: s.MsgCount,
				Preview:      s.Preview,
			})
		}
		encoder := json.NewEncoder(cmd.Root().Writer)
		encoder.SetIndent("", "  ")
		return encoder.Encode(items)
	}

	if len(summaries) == 0 {
		fmt.Fprintln(cmd.Root().Writer, "No sessions found.")
		return nil
	}

	w := tabwriter.NewWriter(cmd.Root().Writer, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "ID\tUPDATED\tMSGS\tPREVIEW\n")
	for _, s := range summaries {
		fmt.Fprintf(w, "%s\t%s\t%d\t%s\n",
			s.ID,
			s.ModTime.Format("2006-01-02 15:04"),
			s.MsgCount,
			s.Preview,
		)
	}
	return w.Flush()
}

// sessionShowView is the JSON representation of `session show`.
type sessionShowView struct {
	ID        string                `json:"id"`
	Model     string                `json:"model"`
	UpdatedAt string                `json:"updated_at"`
	Source    *assistant.SourceInfo `json:"source,omitempty"`
	Messages  []assistant.Message   `json:"messages"`
}

func runSessionShow(ctx context.Context, cmd *cli.Command) error {
	id := cmd.Args().First()
	if id == "" {
		return fmt.Errorf("session ID is required: aico session show <session-id>")
	}

	conf, err := config.Load()
	if err != nil {
		conf = config.DefaultConfig()
	}

	sess, err := assistant.LoadSession(conf.GetSessionDir(), id)
	if err != nil {
		return err
	}

	if cmd.Bool(flagJSON.Name) {
		info, err := os.Stat(sess.FilePath())
		if err != nil {
			return fmt.Errorf("stat session file: %w", err)
		}
		view := sessionShowView{
			ID:        sess.ID,
			Model:     sess.Model,
			UpdatedAt: info.ModTime().Format(time.RFC3339),
			Source:    sess.Source,
			Messages:  sess.Messages,
		}
		encoder := json.NewEncoder(cmd.Root().Writer)
		encoder.SetIndent("", "  ")
		return encoder.Encode(view)
	}

	w := cmd.Root().Writer
	if s := sess.Source.String(); s != "" {
		fmt.Fprintf(w, "Source: %s\n\n", s)
	}
	for i, msg := range sess.Messages {
		if i > 0 {
			fmt.Fprintln(w)
		}
		fmt.Fprintf(w, "--- %s ---\n", msg.GetAuthor())
		for _, c := range msg.GetContents() {
			switch c := c.(type) {
			case *assistant.TextContent:
				fmt.Fprintln(w, c.Text)
			case *assistant.URLImageContent:
				fmt.Fprintf(w, "[image] %s\n", c.URL.String())
			case *assistant.ToolUseContent:
				printToolUse(w, c)
			case *assistant.ToolResultContent:
				fmt.Fprintf(w, "[tool_result] %s\n", c.Content)
			}
		}
	}
	return nil
}

// printToolUse renders a tool_use content block for the plain-text `session
// show` output. propose_edit is rendered the same way as
// ConsoleLineStreamWriter.EmitToolUse; any other tool name falls back to
// its raw input.
func printToolUse(w io.Writer, c *assistant.ToolUseContent) {
	if c.Name != toolNameProposeEdit {
		fmt.Fprintf(w, "[tool_use %s] %s\n", c.Name, string(c.Input))
		return
	}
	description, oldString, newString, ok := proposeEditFields(c.Input)
	if !ok {
		fmt.Fprintf(w, "[propose_edit] %s\n", string(c.Input))
		return
	}
	fmt.Fprintf(w, "[propose_edit] %s\n", description)
	for _, l := range strings.Split(oldString, "\n") {
		fmt.Fprintf(w, "-%s\n", l)
	}
	for _, l := range strings.Split(newString, "\n") {
		fmt.Fprintf(w, "+%s\n", l)
	}
}

func runSessionResume(ctx context.Context, cmd *cli.Command) error {
	sessionID := cmd.Args().First()
	if sessionID == "" {
		return fmt.Errorf("session ID is required: aico session resume <session-id> <prompt>")
	}
	prompt := cmd.Args().Get(1)
	if prompt == "" {
		return fmt.Errorf("prompt is required: aico session resume <session-id> <prompt>")
	}

	if err := cmd.Root().Set("session", sessionID); err != nil {
		return fmt.Errorf("set session flag: %w", err)
	}

	return doGenerate(ctx, cmd, prompt)
}
