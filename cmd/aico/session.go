package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"text/tabwriter"
	"time"

	"github.com/urfave/cli/v3"

	"micheam.com/aico/internal/assistant"
	"micheam.com/aico/internal/config"
)

var CmdSession = &cli.Command{
	Name:  "session",
	Usage: "Manage chat sessions",
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
			Name:      "resume",
			Usage:     "Resume an existing session with a new prompt",
			ArgsUsage: "<session-id> <prompt>",
			Action:    runSessionResume,
			Flags: []cli.Flag{
				flagSource,
				flagContext,
				flagModel,
				flagNoStream,
				flagDebug,
				flagPersona,
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
