package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v3"
	"gopkg.in/yaml.v3"

	"micheam.com/aico/internal/assistant"
	"micheam.com/aico/internal/logging"
)

var ChatSession = &cli.Command{
	Name:      "session",
	Usage:     "Manage chat sessions",
	ArgsUsage: "<subcommand>",
	Flags:     []cli.Flag{flagDebug},
	Commands: []*cli.Command{
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
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "format",
					Usage: "Output format: markdown, json, yaml",
					Value: "markdown",
				},
			},
			Action: showChatSession,
		},
		{
			Name:      "delete",
			Usage:     "Delete a chat session",
			Aliases:   []string{"rm"},
			ArgsUsage: "<session-id>",
			Action:    deleteChatSession,
		},
		{
			Name:   "prepare",
			Usage:  "Prepare an empty chat session",
			Action: prepareChatSession,
		},
	},
}

func listChatSessions(ctx context.Context, cmd *cli.Command) error {
	conf, err := LoadConfig(ctx, cmd)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Setup logger
	logLevel := logging.LevelInfo
	if cmd.Bool("debug") {
		logLevel = logging.LevelDebug
	}
	logger, cleanup, err := setupLogger(conf.Logfile(), logLevel)
	if err != nil {
		return err
	}
	defer cleanup()

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
		fmt.Fprintln(cmd.Root().Writer, id)
	}
	_ = logger
	return nil
}

func showChatSession(ctx context.Context, cmd *cli.Command) error {
	if cmd.Args().Len() == 0 {
		return fmt.Errorf("session-id is required")
	}
	conf, err := LoadConfig(ctx, cmd)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Setup logger
	logLevel := logging.LevelInfo
	if cmd.Bool("debug") {
		logLevel = logging.LevelDebug
	}
	logger, cleanup, err := setupLogger(conf.Logfile(), logLevel)
	if err != nil {
		return err
	}
	defer cleanup()

	confLocationDir := filepath.Dir(conf.Location())
	sessStoreDir, err := filepath.Abs(filepath.Join(confLocationDir, conf.Chat.Session.Directory))
	if err != nil {
		return fmt.Errorf("resolve session directory: %w", err)
	}

	sessID := cmd.Args().First()
	sess, err := assistant.RestoreChat(sessStoreDir, sessID, nil)
	if err != nil {
		return fmt.Errorf("restore chat: %s: %w", sessID, err)
	}
	
	format := cmd.String("format")
	switch format {
	case "markdown":
		markdown, err := sess.ToMarkdown()
		if err != nil {
			return fmt.Errorf("convert to markdown: %w", err)
		}
		fmt.Fprintln(cmd.Root().Writer, markdown)
	case "json":
		jsonBytes, err := json.MarshalIndent(sess, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal to JSON: %w", err)
		}
		fmt.Fprintln(cmd.Root().Writer, string(jsonBytes))
	case "yaml":
		yamlBytes, err := yaml.Marshal(sess)
		if err != nil {
			return fmt.Errorf("marshal to YAML: %w", err)
		}
		fmt.Fprint(cmd.Root().Writer, string(yamlBytes))
	default:
		return fmt.Errorf("unsupported format: %s (supported: markdown, json, yaml)", format)
	}

	_ = logger
	return nil
}

func deleteChatSession(ctx context.Context, cmd *cli.Command) error {
	if cmd.Args().Len() == 0 {
		return fmt.Errorf("session-id is required")
	}

	conf, err := LoadConfig(ctx, cmd)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	// Setup logger
	logLevel := logging.LevelInfo
	if cmd.Bool("debug") {
		logLevel = logging.LevelDebug
	}
	logger, cleanup, err := setupLogger(conf.Logfile(), logLevel)
	if err != nil {
		return err
	}
	defer cleanup()

	confLocationDir := filepath.Dir(conf.Location())
	sessStoreDir, err := filepath.Abs(filepath.Join(confLocationDir, conf.Chat.Session.Directory))
	if err != nil {
		return fmt.Errorf("resolve session directory: %w", err)
	}

	sessID := cmd.Args().First()
	if err := os.Remove(filepath.Join(sessStoreDir, sessID+".pb")); err != nil {
		return fmt.Errorf("delete session: %s: %w", sessID, err)
	}

	_ = logger
	return nil
}

// prepareChatSession prepares empty chat session, and show the id
func prepareChatSession(ctx context.Context, cmd *cli.Command) error {
	conf, err := LoadConfig(ctx, cmd)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	// Setup logger
	logLevel := logging.LevelInfo
	if cmd.Bool("debug") {
		logLevel = logging.LevelDebug
	}
	logger, cleanup, err := setupLogger(conf.Logfile(), logLevel)
	if err != nil {
		return err
	}
	defer cleanup()

	confLocationDir := filepath.Dir(conf.Location())
	sessStoreDir, err := filepath.Abs(filepath.Join(confLocationDir, conf.Chat.Session.Directory))
	if err != nil {
		return fmt.Errorf("resolve session directory: %w", err)
	}

	// Load GenerativeModel
	m, err := setupGenerativeModel(conf.Chat.Model)
	if err != nil {
		return fmt.Errorf("create model: %w", err)
	}

	// Create ChatSession
	sess, err := assistant.StartChat(m)
	if err != nil {
		return fmt.Errorf("create chat session: %w", err)
	}
	if err := sess.Save(sessStoreDir); err != nil {
		return fmt.Errorf("save chat session: %w", err)
	}
	fmt.Fprintln(cmd.Root().Writer, sess.ID)

	logger.Debug("Chat session created", "session-id", sess.ID)
	return nil
}

