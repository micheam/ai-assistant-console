package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"

	"micheam.com/aico/internal/assistant"
	"micheam.com/aico/internal/config"
)

var ChatSession = &cli.Command{
	Name:      "session",
	Usage:     "Manage chat sessions",
	ArgsUsage: "<subcommand>",
	Before:    loadConfig,
	Subcommands: []*cli.Command{
		{
			Name:    "list",
			Usage:   "List saved chat sessions",
			Aliases: []string{"ls"},
			Action:  listChatSessions,
		},
		{
			Name:      "show",
			Usage:     "Show a chat session",
			ArgsUsage: "<session-id>",
			Action:    showChatSession,
		},
		{
			Name:      "delete",
			Usage:     "Delete a chat session",
			Aliases:   []string{"rm"},
			ArgsUsage: "<session-id>",
			Action:    deleteChatSession,
		},
	},
}

func listChatSessions(c *cli.Context) error {
	conf := config.MustFromContext(c.Context)
	confLocationDir := filepath.Dir(conf.Location())
	sessStoreDir, err := filepath.Abs(filepath.Join(confLocationDir, conf.Chat.Session.Directory))
	if err != nil {
		return fmt.Errorf("resolve session directory: %w", err)
	}

	files, err := os.ReadDir(sessStoreDir)
	if err != nil {
		return fmt.Errorf("read session directory: %w", err)
	}

	for _, f := range files {
		if filepath.Ext(f.Name()) != ".pb" || !strings.HasPrefix(f.Name(), "sess-") {
			continue // Skip if not a session file
		}
		id := strings.TrimSuffix(f.Name(), ".pb")
		fmt.Fprintln(c.App.Writer, id)
	}
	return nil
}

func showChatSession(c *cli.Context) error {
	if c.Args().Len() == 0 {
		return fmt.Errorf("session-id is required")
	}

	conf := config.MustFromContext(c.Context)
	confLocationDir := filepath.Dir(conf.Location())
	sessStoreDir, err := filepath.Abs(filepath.Join(confLocationDir, conf.Chat.Session.Directory))
	if err != nil {
		return fmt.Errorf("resolve session directory: %w", err)
	}

	sessID := c.Args().First()
	sess, err := assistant.RestoreChat(sessStoreDir, sessID, nil)
	if err != nil {
		return fmt.Errorf("restore chat: %s: %w", sessID, err)
	}
	b, err := json.Marshal(sess)
	if err != nil {
		return fmt.Errorf("marshal chat session: %w", err)
	}
	fmt.Fprintln(c.App.Writer, string(b))
	return nil
}

func deleteChatSession(c *cli.Context) error {
	if c.Args().Len() == 0 {
		return fmt.Errorf("session-id is required")
	}

	conf := config.MustFromContext(c.Context)
	confLocationDir := filepath.Dir(conf.Location())
	sessStoreDir, err := filepath.Abs(filepath.Join(confLocationDir, conf.Chat.Session.Directory))
	if err != nil {
		return fmt.Errorf("resolve session directory: %w", err)
	}

	sessID := c.Args().First()
	if err := os.Remove(filepath.Join(sessStoreDir, sessID+".pb")); err != nil {
		return fmt.Errorf("delete session: %s: %w", sessID, err)
	}
	return nil
}
