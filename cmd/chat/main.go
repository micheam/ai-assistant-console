package main

import (
	"bufio"
	"context"
	_ "embed"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/urfave/cli/v2"

	"micheam.com/aico/internal/config"
	"micheam.com/aico/internal/openai"
	"micheam.com/aico/internal/theme"
	tui "micheam.com/aico/internal/tui/chat"
)

const authEnvKey = "OPENAI_API_KEY"

//go:embed version.txt
var version string

func main() {
	if err := app().Run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func app() *cli.App {
	return &cli.App{
		Name:    "chat",
		Usage:   "Chat with AI",
		Version: version,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "debug",
				Usage:   "Enable debug mode",
				EnvVars: []string{"AICO_DEBUG"},
			},
			&cli.StringFlag{
				Name:    "model",
				Aliases: []string{"m"},
				Usage:   "GPT model to use",
				Value:   config.DefaultModel,
			},
			&cli.StringFlag{
				Name:    "persona",
				Aliases: []string{"p"},
				Usage:   "Persona to use",
				Value:   "default",
			},
		},
		Before: func(c *cli.Context) error {
			ctx := c.Context

			// Load Config
			cfg, err := config.InitAndLoad(ctx)
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			// Setup logger
			logger := log.New(io.Discard, "", log.LstdFlags|log.LUTC)
			if c.Bool("debug") {
				lfile := cfg.Logfile()
				f, err := os.OpenFile(
					lfile,
					os.O_APPEND|os.O_CREATE|os.O_WRONLY,
					0644,
				)
				if os.IsNotExist(err) {
					if f, err = os.Create(lfile); err != nil {
						return fmt.Errorf("prepare log file(%q): %w", lfile, err)
					}
				}

				logger.SetOutput(f)
				fmt.Println(theme.Info("Debug mode is on")) // TODO: promote to Logger or Presenter
				fmt.Printf(theme.Info("You can find logs in %q\n"), lfile)
				fmt.Println()
			}

			// Override model if specified with flag
			model := c.String("model")
			cfg.Chat.Model = model

			// Setup context
			ctx = WithConfig(ctx, cfg)
			ctx = WithLogger(ctx, logger)

			c.Context = ctx
			return nil
		},
		Commands: []*cli.Command{

			{
				Name:        "config",
				Usage:       "Show config file path",
				Description: "Show config file path",
				Action: func(c *cli.Context) error {
					path := config.ConfigFilePath(c.Context)
					fmt.Println(path)
					return nil
				},
			},

			{
				Name:        "tui",
				Usage:       "Chat with AI in TUI",
				Description: "Start TUI application to chat with AI",
				Action: func(c *cli.Context) error {
					ctx := c.Context
					cfg := ConfigFrom(ctx)
					logger := LoggerFrom(ctx)
					logger.SetPrefix("[CHAT][TUI] ")

					handler := tui.New(cfg, logger)
					return handler.Run(ctx)
				},
			},

			SendMessageCommand,
		},
	}
}

// --------------------------------------------------------------------
// Handle context
// --------------------------------------------------------------------

type contextKey int

const (
	contextKeyConfig contextKey = iota
	contextKeyLogger
)

func WithConfig(ctx context.Context, cfg *config.Config) context.Context {
	return context.WithValue(ctx, contextKeyConfig, cfg)
}

func ConfigFrom(ctx context.Context) *config.Config {
	return ctx.Value(contextKeyConfig).(*config.Config)
}

func WithLogger(ctx context.Context, logger *log.Logger) context.Context {
	return context.WithValue(ctx, contextKeyLogger, logger)
}

func LoggerFrom(ctx context.Context) *log.Logger {
	return ctx.Value(contextKeyLogger).(*log.Logger)
}

// --------------------------------------------------------------------
// Commands
// --------------------------------------------------------------------

var SendMessageCommand = &cli.Command{
	Name:        "send",
	Usage:       "Send message to AI",
	Description: "Send message to AI and get response",
	ArgsUsage:   "MESSAGE",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:      "input",
			Aliases:   []string{"i"},
			Usage:     "Input file path",
			TakesFile: true,
		},
	},
	Action: sendMessage,
}

func sendMessage(c *cli.Context) error {
	var (
		err  error
		ctx  = c.Context
		conf = ConfigFrom(ctx)
	)

	logger := LoggerFrom(ctx)
	logger.SetPrefix("[CHAT][CLI] ")

	var chat *openai.ChatClient
	{
		var apikey string
		if apikey = os.Getenv(authEnvKey); apikey == "" {
			logger.Printf("[ERROR] API Key (env: %s) is not set", authEnvKey)
			return fmt.Errorf("API Key is not set, please set %s", authEnvKey)
		}
		client := openai.NewClient(apikey)
		chat = openai.NewChatClient(client)
	}

	messages := make([]openai.Message, 0)

	// System messages from persona
	if persona, ok := conf.Chat.GetPersona(c.String("persona")); ok {
		for _, msg := range persona.Messages {
			messages = append(messages, openai.Message{
				Role:    openai.RoleSystem,
				Content: msg,
			})
		}
	}

	// Handle user message
	var inputMsg io.Reader
	if c.Args().Len() > 0 {
		inputMsg = strings.NewReader(c.Args().First())

	} else if c.String("input") != "" {
		f, err := os.Open(c.String("input"))
		if err != nil {
			return fmt.Errorf("open input file: %w", err)
		}
		defer f.Close()
		inputMsg = f

	} else if !isStdinEmpty() {
		inputMsg = os.Stdin

	} else {
		return fmt.Errorf("no input message")
	}

	msgs := ParseInputMessage(inputMsg)
	messages = append(messages, msgs...)

	// Send chat request
	model := conf.Chat.Model
	req := openai.NewChatRequest(model, messages)
	req.Temperature = conf.Chat.Temperature

	logger.Printf("ChatCompletion request: %+v", req)
	resp, err := chat.Do(ctx, req)
	if err != nil {
		logger.Printf("Got error: %+v", err)
		return fmt.Errorf("chat: %w", err)
	}
	logger.Printf("ChatCompletion response: %+v", resp)

	// Show response
	if len(resp.Choices) > 0 {
		msg := resp.Choices[0].Message
		fmt.Println("Assistant:")
		fmt.Println(msg.Content + "\n")
	}

	return nil
}

// --------------------------------------------------------------------
// Helpers
// --------------------------------------------------------------------

// ParseInputMessage parses input messages.
func ParseInputMessage(src io.Reader) []openai.Message {

	var (
		messages = make([]openai.Message, 0)
		scanner  = bufio.NewScanner(src)

		role    = openai.RoleUser
		content = ""
	)

	for scanner.Scan() {
		line := scanner.Text()

		switch { // detect Role with prompt

		case strings.HasPrefix(line, "User:"):
			role = openai.RoleUser
			line = strings.TrimPrefix(line, "User:")
			line = strings.TrimPrefix(line, " ")

		case strings.HasPrefix(line, "Assistant:"):
			role = openai.RoleAssistant
			line = strings.TrimPrefix(line, "Assistant:")
			line = strings.TrimPrefix(line, " ")

		case strings.HasPrefix(line, "System:"):
			role = openai.RoleSystem
			line = strings.TrimPrefix(line, "System:")
			line = strings.TrimPrefix(line, " ")
		}

		if line == "" && content != "" { // empty line means end of message section
			messages = append(messages, openai.Message{
				Role:    role,
				Content: content,
			})
			content = ""
			continue
		}
		if content != "" { // soft break
			content += "\n"
		}
		content += line
	}

	if content != "" {
		messages = append(messages, openai.Message{
			Role:    role,
			Content: content,
		})
	}

	return messages
}

func isStdinEmpty() bool {
	stat, _ := os.Stdin.Stat()
	return (stat.Mode() & os.ModeCharDevice) != 0
}
