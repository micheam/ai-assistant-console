package commands

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"

	"micheam.com/aico/internal/assistant"
	"micheam.com/aico/internal/config"
	"micheam.com/aico/internal/logging"
	"micheam.com/aico/internal/openai"
)

var ChatSend = &cli.Command{
	Name:      "send",
	Usage:     "Send a message to the AI assistant",
	ArgsUsage: "<message>",
	Flags: []cli.Flag{
		flagPersona,
		flagModel,
		flagChatSession,
		flagChatInstant,
	},
	Before: LoadConfig,
	Action: func(c *cli.Context) error {
		msgs := c.Args().Slice()
		if len(msgs) == 0 {
			// Try to read from stdin
			var err error
			msgs, err = readLines(os.Stdin)
			if err != nil {
				return fmt.Errorf("read from stdin: %w", err)
			}
			if len(msgs) == 0 {
				return fmt.Errorf("no message provided")
			}
		}

		logger := logging.LoggerFrom(c.Context)
		conf := config.MustFromContext(c.Context)

		// Resolve session directory
		confLocationDir := filepath.Dir(conf.Location())
		sessStoreDir, err := filepath.Abs(path.Join(confLocationDir, conf.Chat.Session.Directory))
		if err != nil {
			return fmt.Errorf("resolve session directory: %w", err)
		}

		// Load GenerativeModel
		m, err := setupGenerativeModel(*conf)
		if err != nil {
			return fmt.Errorf("create model: %w", err)
		}

		// Create or Restore ChatSession
		var sess *assistant.ChatSession
		if sessID := c.String("session"); sessID != "" {
			sess, err = assistant.RestoreChat(sessStoreDir, sessID, m)
			if err != nil {
				return fmt.Errorf("restore chat: %s: %w", sessID, err)
			}
		} else {
			sess, err = assistant.StartChat(m)
			if err != nil {
				return fmt.Errorf("start chat: %w", err)
			}
		}

		// Resolve persona
		persona, found := resolvePersona(conf, c.String("persona"))
		if !found {
			return fmt.Errorf("persona not found: %s", c.String("persona"))
		}
		sess.SetSystemInstruction(
			assistant.NewTextContent(strings.Join(persona.Messages, "\n")))

		// Send message to assistant
		ctx := logging.ContextWith(c.Context, logger)
		message := strings.Join(msgs, " ")
		resp, err := sess.SendMessage(ctx, assistant.NewTextContent(message))
		if err != nil {
			return fmt.Errorf("send message: %w", err)
		}

		// Store session
		if !c.Bool("instant") {
			if err := sess.Save(sessStoreDir); err != nil {
				return fmt.Errorf("save session: %w", err)
			}
			fmt.Fprintf(c.App.Writer, "Session saved: %s\n", sess.ID)
		}

		// Print response
		switch v := resp.Content.(type) {
		case *assistant.TextContent:
			fmt.Fprintln(c.App.Writer, v.Text)
		default:
			logger.Error("unexpected response type", "type", fmt.Sprintf("%T", v))
		}

		return nil
	},
}

func setupGenerativeModel(conf config.Config) (assistant.GenerativeModel, error) {
	openaiAPIKey, found := os.LookupEnv("OPENAI_API_KEY")
	if !found {
		return nil, fmt.Errorf("missing environment variable: OPENAI_API_KEY")
	}
	model, err := openai.NewGenerativeModel(conf.Chat.Model, openaiAPIKey)
	if err != nil {
		return nil, fmt.Errorf("OpenAI: %w", err)
	}
	return model, nil
}

// resolvePersona resolves the persona to use for the chat
// If the personaName is empty, the default persona will be used.
func resolvePersona(conf *config.Config, personaName string) (*config.Personality, bool) {
	if personaName != "" {
		return conf.Chat.GetPersona(personaName)
	}
	return conf.Chat.GetDefaultPersona(), true
}
