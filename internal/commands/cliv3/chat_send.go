package commands

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/urfave/cli/v3"

	"micheam.com/aico/internal/anthropic"
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
		flagDebug,
		flagPersona,
		flagModel,
		flagChatSession,
		flagChatInstant,
		&cli.BoolFlag{
			Name:  "stream",
			Usage: "Stream the conversation",
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		msgs := cmd.Args().Slice()
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

		// Resolve session directory
		confLocationDir := filepath.Dir(conf.Location())
		sessStoreDir, err := filepath.Abs(path.Join(confLocationDir, conf.Chat.Session.Directory))
		if err != nil {
			return fmt.Errorf("resolve session directory: %w", err)
		}

		// Load GenerativeModel
		m, err := setupGenerativeModel(conf.Chat.Model)
		if err != nil {
			return fmt.Errorf("create model: %w", err)
		}

		// Create or Restore ChatSession
		var sess *assistant.ChatSession
		if sessID := cmd.String("session"); sessID != "" {
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
		persona, found := resolvePersona(conf, cmd.String("persona"))
		if !found {
			return fmt.Errorf("persona not found: %s", cmd.String("persona"))
		}
		sess.SetSystemInstruction(
			assistant.NewTextContent(strings.Join(persona.Messages, "\n")))

		// Send message to assistant
		ctx = logging.ContextWith(ctx, logger)
		message := strings.Join(msgs, " ")

		if !cmd.Bool("stream") {
			resp, err := sess.SendMessage(ctx, assistant.NewTextContent(message))
			if err != nil {
				return fmt.Errorf("send message: %w", err)
			}

			// Store session
			if !cmd.Bool("instant") {
				if err := sess.Save(sessStoreDir); err != nil {
					return fmt.Errorf("save session: %w", err)
				}
				logger.Debug("Session saved", "session-id", sess.ID)
			}

			// Print response
			switch v := resp.Content.(type) {
			case *assistant.TextContent:
				fmt.Fprintln(cmd.Root().Writer, v.Text)
			default:
				logger.Error("unexpected response type", "type", fmt.Sprintf("%T", v))
			}
			return nil
		}

		// Stream conversation
		iter, err := sess.SendMessageStream(ctx, assistant.NewTextContent(message))
		if err != nil {
			return fmt.Errorf("send message: %w", err)
		}

		var replyBuilder strings.Builder
		for resp := range iter {
			switch content := resp.Content.(type) {
			case *assistant.TextContent:
				fmt.Fprint(cmd.Root().Writer, content.Text)
				replyBuilder.WriteString(content.Text)
			default:
				logger.Error("unexpected response type", "type", fmt.Sprintf("%T", content))
			}
		}
		completeReply := replyBuilder.String()
		sess.AddHistory(&assistant.GenerateContentResponse{
			Content: assistant.NewTextContent(completeReply),
		})
		fmt.Fprint(cmd.Root().Writer, "\n\n")

		if !cmd.Bool("instant") {
			if err := sess.Save(sessStoreDir); err != nil {
				return fmt.Errorf("save session: %w", err)
			}
			logger.Debug("Session saved", "session-id", sess.ID)
		}
		return nil
	},
}

func setupGenerativeModel(model string) (assistant.GenerativeModel, error) {

	// OPEN_AI API
	if slices.Contains(openai.AvailableModels(), model) {
		openaiAPIKey, found := os.LookupEnv("OPENAI_API_KEY")
		if !found {
			return nil, fmt.Errorf("missing environment variable: OPENAI_API_KEY")
		}
		model, err := openai.NewGenerativeModel(model, openaiAPIKey)
		if err != nil {
			return nil, fmt.Errorf("OpenAI: %w", err)
		}
		return model, nil
	}

	// ANTHROPIC API
	if slices.Contains(anthropic.AvailableModels(), model) {
		anthropicAPIKey, found := os.LookupEnv("ANTHROPIC_API_KEY")
		if !found {
			return nil, fmt.Errorf("missing environment variable: ANTHROPIC_API_KEY")
		}
		model, err := anthropic.NewGenerativeModel(model, anthropicAPIKey)
		if err != nil {
			return nil, fmt.Errorf("anthropic: %w", err)
		}
		return model, nil
	}

	// UNSUPPORTED...
	return nil, fmt.Errorf("unfortunately, the model %s is not supported", model)
}

// resolvePersona resolves the persona to use for the chat
// If the personaName is empty, the default persona will be used.
func resolvePersona(conf *config.Config, personaName string) (*config.Personality, bool) {
	if personaName != "" {
		return conf.Chat.GetPersona(personaName)
	}
	return conf.Chat.GetDefaultPersona(), true
}
