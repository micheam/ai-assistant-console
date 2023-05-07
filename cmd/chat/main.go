package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/sashabaranov/go-openai"

	"github.com/nsf/termbox-go"
)

const authEnvKey = "OPENAI_API_KEY"

var (
	white  = color.New(color.FgWhite).SprintFunc()
	gray   = color.New(color.FgHiBlack).SprintFunc()
	red    = color.New(color.FgRed).SprintFunc()
	green  = color.New(color.FgGreen).SprintFunc()
	blue   = color.New(color.FgBlue).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()

	info  = white
	reply = blue
	error = red
)

func main() {
	var (
		ctx     = context.Background()
		cmdList = flag.Bool("M", false, "list available models")

		model       = flag.String("m", defaultModel(), "model to use")
		prompt      = flag.String("p", defaultPrompt(), "prompt to use")
		temperature = flag.Float64("t", defaultTemperature(), "temperature to use")

		stream = flag.Bool("s", false, "streaming mode")
		debug  = flag.Bool("debug", false, "debug mode")
	)

	flag.Parse()

	logger := log.New(io.Discard, "", log.LstdFlags|log.Lshortfile)
	if *debug {
		lfile := logfile()
		defer lfile.Close()
		logger.SetOutput(lfile)
		fmt.Println(info("Debug mode is on"))
		fmt.Printf(info("You can find logs in %q\n"), lfile.Name())
		fmt.Println()
	}

	authToken := os.Getenv(authEnvKey)
	if authToken == "" {
		fmt.Printf(error("%s is not set"), authEnvKey)
		os.Exit(1)
	}

	client := openai.NewClient(authToken)

	if *cmdList {
		for _, m := range availableModels {
			fmt.Printf(error("%s\n"), m)
		}
		os.Exit(0)
	}

	messages := make([]openai.ChatCompletionMessage, 0)
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf(info("Conversation with %s\n"), *model)
	logger.Printf("Conversation Starts with %s\n", *model)
	fmt.Println(info("------------------------------"))

	for {
		fmt.Print(info(*prompt))
		text, _ := reader.ReadString('\n')
		text = strings.ReplaceAll(text, "\n", "") // convert CRLF to LF

		switch text {

		default: // store user input
			logger.Printf("User input: %s\n", text)
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: text,
			})

		case "": // empty input
			continue

		case ":quit", ":q", ":exit":
			return

		case ":send", ";;":
			fmt.Println()

			if *stream {
				logger.Printf("Handling ChatCompletion [Stream] request\n")
				req := openai.ChatCompletionRequest{
					Model:       *model,
					Messages:    messages,
					Temperature: float32(*temperature),
					Stream:      true, // Stream Mode
				}
				logger.Printf("ChatCompletion request: %+v\n", req)
				st, err := client.CreateChatCompletionStream(ctx, req)
				if err != nil {
					logger.Printf("ChatCompletion [Stream] error: %v\n", err)
					fmt.Printf(error("ChatCompletion [Stream] error: %v\n"), err)
					continue
				}
				defer st.Close()

				contents := []string{}
				buf := []string{}
				for {
					resp, err := st.Recv()
					if err != nil {
						if errors.Is(err, io.EOF) {
							fmt.Print(reply(strings.Join(buf, "")))
							messages = append(messages, openai.ChatCompletionMessage{
								Role:    openai.ChatMessageRoleAssistant,
								Content: strings.Join(contents, ""),
							})
							fmt.Println()
							break
						}

						logger.Printf("ChatCompletion [Stream] error: %v\n", err)
						fmt.Printf(error("ChatCompletion [Stream] error: %v\n"), err)
						break
					}
					delta := resp.Choices[0].Delta.Content
					contents = append(contents, delta)

					buf = append(buf, delta)
					logger.Printf("[stream]: %+v\n", delta)

					if strings.Contains(delta, "\n") {
						fmt.Print(reply(strings.Join(buf, "")))
						buf = []string{}
					}
				}
				continue
			}

			req := openai.ChatCompletionRequest{
				Model:       *model,
				Messages:    messages,
				Temperature: float32(*temperature),
				Stream:      false, // Batch Mode
			}
			logger.Printf("ChatCompletion request: %+v\n", req)
			resp, err := client.CreateChatCompletion(ctx, req)

			if err != nil {
				logger.Printf("ChatCompletion error: %v\n", err)
				fmt.Printf(error("ChatCompletion error: %v\n"), err)
				continue
			}
			logger.Printf("ChatCompletion response: %+v\n", resp)

			content := resp.Choices[0].Message.Content
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleAssistant,
				Content: content,
			})

			fmt.Println(reply(content))
			fmt.Println()
		}
	}
}

func getTerminalWidth() int {
	if err := termbox.Init(); err != nil {
		panic(err)
	}
	w, _ := termbox.Size()
	termbox.Close()
	return w
}

var availableModels = []string{
	openai.GPT3Dot5Turbo,
	openai.GPT4,
}

// defaultModel returns default model to use
func defaultModel() string {
	if os.Getenv("CHATGPT_MODEL") != "" {
		return os.Getenv("CHATGPT_MODEL")
	}
	return availableModels[0]
}

// defaultPrompt returns default prompt to use
func defaultPrompt() string {
	if os.Getenv("CHATGPT_PROMPT") != "" {
		return os.Getenv("CHATGPT_PROMPT")
	}
	return "> "
}

// defaultTemperature returns default temperature to use
func defaultTemperature() float64 {
	if os.Getenv("CHATGPT_TEMPERATURE") != "" {
		return 0.9
	}
	return 0.9
}

// datadir returns default data directory
//
// We determin data directory by the rules below:
// 1. If CHATGPT_DATA_DIR environment variable is set, use it
// 2. If XDG_DATA_HOME environment variable is set, use it
// 3. otherwise, use $HOME/.local/share
func datadir() string {
	if os.Getenv("CHATGPT_DATA_DIR") != "" {
		return os.Getenv("CHATGPT_DATA_DIR")
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
