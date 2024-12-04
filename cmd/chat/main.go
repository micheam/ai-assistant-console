package main

import (
	"bufio"
	_ "embed"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/urfave/cli/v2"

	"micheam.com/aico/internal/config"
	"micheam.com/aico/internal/logging"
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
				Usage:   "GPT model to use, default is set in config file",
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
			logger, err := logging.SetupLogger(c.Bool("debug"), cfg.Logfile())
			if err != nil {
				return fmt.Errorf("setup logger: %w", err)
			}
			if c.Bool("debug") {
				fmt.Println(theme.Info("Debug mode is on")) // TODO: promote to Logger or Presenter
				fmt.Printf(theme.Info("You can find logs in %q\n"), cfg.Logfile())
				fmt.Println()
			}

			// Override model if specified with flag
			if model := c.String("model"); model != "" {
				cfg.Chat.Model = model
			}
			if cfg.Chat.Model == "" {
				cfg.Chat.Model = config.DefaultModel
			}

			// Setup context
			ctx = config.WithConfig(ctx, cfg)
			ctx = logging.WithLogger(ctx, logger)
			c.Context = ctx

			return nil
		},
		Commands: commands,
	}
}

// --------------------------------------------------------------------
// Commands
// --------------------------------------------------------------------

var commands = []*cli.Command{
	configCommand,
	startTUICommand,
	sendMessageCommand,
}

var configCommand = &cli.Command{
	Name:        "config",
	Usage:       "Show config file path",
	Description: "Show config file path",
	Action: func(c *cli.Context) error {
		path := config.ConfigFilePath(c.Context)
		fmt.Println(path)
		return nil
	},
}

var startTUICommand = &cli.Command{
	Name:        "tui",
	Usage:       "Chat with AI in TUI",
	Description: "Start TUI application to chat with AI",
	Action: func(c *cli.Context) error {
		ctx := c.Context
		cfg := config.ConfigFrom(ctx)
		logger := logging.LoggerFrom(ctx)
		logger.SetPrefix("[CHAT][TUI] ")

		handler := tui.New(cfg, logger)
		if persona, ok := cfg.Chat.GetPersona(c.String("persona")); ok {
			handler = handler.WithPersona(persona)
		}
		return handler.Run(ctx)
	},
}

var sendMessageCommand = &cli.Command{
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
	Action: func(c *cli.Context) error {
		var (
			ctx       = c.Context
			conf      = config.ConfigFrom(ctx)
			logger    = logging.LoggerFrom(ctx)
			fileInput = c.String("input")
			msg       = c.Args().First()
			persona   = c.String("persona")
		)

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
		if persona, ok := conf.Chat.GetPersona(persona); ok {
			for _, msg := range persona.Messages {
				messages = append(messages, openai.Message{
					Role:    openai.RoleSystem,
					Content: msg,
				})
			}
		}

		// Handle user message
		var inputMsg io.Reader
		if msg != "" {
			inputMsg = strings.NewReader(msg)

		} else if fileInput != "" {
			f, err := os.Open(fileInput)
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

		cnt := 0
		err := chat.DoStream(ctx, req, func(resp *openai.ChatResponse) error {
			logger.Printf("ChatCompletion response: %+v", resp)
			if cnt == 0 {
				fmt.Println("Assistant:")
				fmt.Println()
			}
			if len(resp.Choices) == 0 {
				return nil
			}
			if len(resp.Choices) > 1 {
				logger.Printf("[WARN]: Got %d choices", len(resp.Choices))
			}
			if msg := resp.Choices[0].Delta; msg != nil {
				fmt.Fprintf(os.Stdout, "%s", msg.Content)
			}
			cnt++
			return nil
		})
		fmt.Println()

		if err != nil {
			logger.Printf("Got error: %+v", err)
			return fmt.Errorf("chat: %w", err)
		}
		return nil
	},
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
			continue

		case strings.HasPrefix(line, "Assistant:"):
			role = openai.RoleAssistant
			continue

		case strings.HasPrefix(line, "System:"):
			role = openai.RoleSystem
			continue
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
