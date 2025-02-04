package main

import (
	"bufio"
	_ "embed"
	"fmt"
	"io"
	"net/url"
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
			level := logging.LevelInfo
			logfile := cfg.Logfile()
			if c.Bool("debug") {
				level = logging.LevelDebug
				fmt.Println(theme.Info("Debug mode is on")) // TODO: promote to Logger or Presenter
				fmt.Printf(theme.Info("You can find logs in %q\n"), logfile)
				fmt.Println()
			}
			f, err := os.OpenFile(logfile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				return fmt.Errorf("open log file: %w", err)
			}
			logger := logging.New(f, &logging.Options{Level: level, AddSource: true})
			logger.Info("Starting chat application", "version", version)

			// Override model if specified with flag
			if model := c.String("model"); model != "" {
				cfg.Chat.Model = model
			}
			if cfg.Chat.Model == "" {
				cfg.Chat.Model = config.DefaultModel
			}

			// Setup context
			c.Context = logging.ContextWith(config.WithConfig(ctx, cfg), logger)

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
		logger := logging.LoggerFrom(ctx).With("component", "tui")

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
	Description: "Send message to AI and get response.",
	ArgsUsage:   "MESSAGE",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:      "input",
			Aliases:   []string{"i"},
			Usage:     "Input file path. If not set, read from stdin.",
			TakesFile: true,
		},
		&cli.BoolFlag{
			Name:  "show-input-format",
			Usage: "Show the expected input format and exit.",
		},
	},
	Action: func(c *cli.Context) error {
		var (
			ctx       = c.Context
			conf      = config.ConfigFrom(ctx)
			logger    = logging.LoggerFrom(ctx).With("component", "chat")
			fileInput = c.String("input")
			msg       = c.Args().First()
			persona   = c.String("persona")
		)

		// Show input format and exit
		if c.Bool("show-input-format") {
			msg := `
Expected input format:

[Role]:
[Message]

	Roles can be 'User:', 'Assistant:', or 'System:'.

Example:

	System:
	語尾に『にゃ』をつけて、可愛い猫ちゃんのように話すにゃ。

	User:
	この画像には、何が写っていますか？

    <https://upload.wikimedia.org/wikipedia/commons/thumb/d/dd/Gfp-wisconsin-madison-the-nature-boardwalk.jpg/2560px-Gfp-wisconsin-madison-the-nature-boardwalk.jpg>
`
			fmt.Fprintf(os.Stdout, "%s\n", msg)
			return nil
		}

		var chat *openai.ChatClient
		{
			var apikey string
			if apikey = os.Getenv(authEnvKey); apikey == "" {
				logger.Error(fmt.Sprintf("API Key (env: %s) is not set", authEnvKey))
				return fmt.Errorf("API Key is not set, please set %s", authEnvKey)
			}
			client := openai.NewClient(apikey)
			chat = openai.NewChatClient(client)
		}

		messages := make([]openai.Message, 0)

		// System messages from persona
		if persona, ok := conf.Chat.GetPersona(persona); ok {
			for _, msg := range persona.Messages {
				messages = append(messages, &openai.SystemMessage{
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
		if t := conf.Chat.Temperature; t != nil {
			req.Temperature = *t
		}

		logger.Debug("Sending ChatCompletion request", "request", req)

		cnt := 0
		err := chat.DoStream(ctx, req, func(resp *openai.ChatResponse) error {
			logger.Debug("Got ChatCompletion response", "response", resp)
			if cnt == 0 {
				fmt.Println("Assistant:")
				fmt.Println()
			}
			if len(resp.Choices) == 0 {
				return nil
			}
			if len(resp.Choices) > 1 {
				logger.Warn("Got multiple choices, using the first one", "choices", resp.Choices)
			}
			if msg := resp.Choices[0].Delta; msg != nil {
				fmt.Fprintf(os.Stdout, "%s", msg.Content)
			}
			cnt++
			return nil
		})
		fmt.Println()

		if err != nil {
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
		role     = openai.RoleUser
		content  []openai.Content
	)

	for scanner.Scan() {
		line := scanner.Text()

		switch { // detect Role with prompt
		case strings.HasPrefix(line, "User:"):
			role = openai.RoleUser
			content = nil
			continue
		case strings.HasPrefix(line, "Assistant:"):
			role = openai.RoleAssistant
			content = nil
			continue
		case strings.HasPrefix(line, "System:"):
			role = openai.RoleSystem
			content = nil
			continue
		}

		// Detect image URL
		if strings.HasPrefix(line, "<") && strings.HasSuffix(line, ">") {
			urlStr := line[1 : len(line)-1]
			if u, err := url.Parse(urlStr); err == nil {
				content = append(content, &openai.ImageContent{URL: *u})
			}
			continue
		}

		// Detect end of message section
		if line == "" && len(content) > 0 {
			switch role {
			case openai.RoleUser:
				messages = append(messages, &openai.UserMessage{Content: content})
			case openai.RoleAssistant:
				messages = append(messages, &openai.AssistantMessage{Content: content})
			case openai.RoleSystem:
				if len(content) == 1 {
					if textContent, ok := content[0].(*openai.TextContent); ok {
						messages = append(messages, &openai.SystemMessage{Content: textContent.Text})
					}
				}
			}
			content = nil
			continue
		}

		// Append text content
		if len(content) > 0 {
			if last, ok := content[len(content)-1].(*openai.TextContent); ok {
				last.Text += "\n" + line
			} else {
				content = append(content, &openai.TextContent{Text: line})
			}
		} else {
			content = append(content, &openai.TextContent{Text: line})
		}
	}

	if len(content) > 0 {
		switch role {
		case openai.RoleUser:
			messages = append(messages, &openai.UserMessage{Content: content})
		case openai.RoleAssistant:
			messages = append(messages, &openai.AssistantMessage{Content: content})
		case openai.RoleSystem:
			if len(content) == 1 {
				if textContent, ok := content[0].(*openai.TextContent); ok {
					messages = append(messages, &openai.SystemMessage{Content: textContent.Text})
				}
			}
		}
	}
	return messages
}

func isStdinEmpty() bool {
	stat, _ := os.Stdin.Stat()
	return (stat.Mode() & os.ModeCharDevice) != 0
}
