package main

import (
	"bufio"
	_ "embed"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/mattn/go-runewidth"
	"github.com/urfave/cli/v2"
	"golang.org/x/term"

	"micheam.com/aico/openai"
)

const authEnvKey = "OPENAI_API_KEY"

var (
	// Basic colors
	white  = color.New(color.FgWhite).SprintFunc()
	gray   = color.New(color.FgHiBlack).SprintFunc()
	red    = color.New(color.FgRed).SprintFunc()
	green  = color.New(color.FgGreen).SprintFunc()
	blue   = color.New(color.FgBlue).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()

	// Color themes
	Info  = white
	Reply = blue
	Error = red

	//go:embed version.txt
	version string

	// Spinner settings
	spinnerFrames   = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	spinnerInterval = 100 * time.Millisecond
)

func main() {
	if err := app().Run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func app() *cli.App {
	return &cli.App{
		Name:                 "aico",
		Usage:                "AI Assistant Console",
		Version:              version,
		EnableBashCompletion: true,
		Flags: []cli.Flag{
			// Debug Enabled
			&cli.BoolFlag{
				Name:    "debug",
				Usage:   "Enable debug mode",
				EnvVars: []string{"AICO_DEBUG"},
			},
		},

		Commands: []*cli.Command{
			{
				Name:  "version",
				Usage: "Show version",
				Action: func(_ *cli.Context) error {
					fmt.Println(version)
					return nil
				},
			},
			{
				Name:   "chat",
				Usage:  "Chat with AI",
				Action: chat,
				Flags: []cli.Flag{
					// gpt model to use
					&cli.StringFlag{
						Name:    "model",
						Aliases: []string{"m"},
						Usage:   "GPT model to use",
						Value:   openai.DefaultChatModel(),
					},
				},
			},
		},
	}
}

func chat(c *cli.Context) error {
	logger := log.New(io.Discard, "", log.LstdFlags|log.LUTC)
	if c.Bool("debug") {
		lfile := logfile()
		defer lfile.Close()
		logger.SetOutput(lfile)
		fmt.Println(Info("Debug mode is on"))
		fmt.Printf(Info("You can find logs in %q\n"), lfile.Name())
		fmt.Println()
	}

	// Spinner settings
	spinner := NewSpinner(spinnerInterval, spinnerFrames)

	ctx := c.Context
	model := c.String("model")
	prompt := "> "

	authToken := os.Getenv(authEnvKey)
	if authToken == "" {
		fmt.Printf(Error("%s is not set"), authEnvKey)
		os.Exit(1)
	}

	client := openai.NewClient(authToken)
	chat := openai.NewChatClient(client)

	messages := make([]openai.Message, 0)
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf(Info("Conversation with %s\n"), model)
	fmt.Println(Info("------------------------------"))
	logger.Printf("Conversation Starts with %s\n", model)

	for {
		fmt.Print(Info(prompt))
		text, _ := reader.ReadString('\n')
		text = strings.ReplaceAll(text, "\n", "") // convert CRLF to LF

		switch text {

		default: // store user input
			logger.Printf("User input: %s\n", text)
			messages = append(messages, openai.Message{
				Role:    openai.RoleUser,
				Content: text,
			})

		case "": // empty input
			continue

		case ".quit", ".q", ".exit":
			return nil

		case ".send", ";;":
			fmt.Println()
			spinner.Start()
			defer spinner.Stop()

			req := openai.NewChatRequest(model, messages)
			logger.Printf("ChatCompletion request: %+v\n", req)

			// width of terminal
			width, _, err := term.GetSize(0)
			if err != nil {
				width = 100
			}

			var cnt int // Current width of output
			content := new(strings.Builder)

			if err := chat.DoStream(ctx, req, func(resp *openai.ChatResponse) error { // Block until completion DONE
				spinner.Stop()
				delta := resp.Choices[0].Delta
				if delta == nil {
					return nil
				}

				if cnt+runewidth.StringWidth(delta.Content) > width {
					fmt.Printf("\n")
					cnt = 0
				}
				fmt.Printf(Reply("%s"), delta.Content)
				cnt += runewidth.StringWidth(delta.Content)

				_, err := content.WriteString(delta.Content)
				return err
			}); err != nil {
				logger.Printf("ChatCompletion error: %v\n", err)
				fmt.Printf(Error("ChatCompletion error: %v\n"), err)
				spinner.Stop()
				continue
			}
			fmt.Printf("\n\n")

			messages = append(messages, openai.Message{
				Role:    openai.RoleAssistant,
				Content: content.String(),
			})

		}
	}
}

// datadir returns default data directory
//
// We determin data directory by the rules below:
// 1. If AICO_DATA_DIR environment variable is set, use it
// 2. If XDG_DATA_HOME environment variable is set, use it
// 3. otherwise, use $HOME/.local/share
func datadir() string {
	if os.Getenv("AICO_DATA_DIR") != "" {
		return os.Getenv("AICO_DATA_DIR")
	}

	if os.Getenv("XDG_DATA_HOME") != "" {
		return os.Getenv("XDG_DATA_HOME")
	}

	return fmt.Sprintf("%s/.local/share", os.Getenv("HOME"))
}

// logfile returns logfile with location based on datadir.
func logfile() *os.File {
	logfile, err := os.OpenFile(
		fmt.Sprintf("%s/chatgpt.log", datadir()),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)
	if err != nil {
		log.Fatal(err)
	}
	return logfile
}
