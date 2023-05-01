package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/sashabaranov/go-openai"
)

const authEnvKey = "CHATGPT_API_KEY"

var availableModels = []string{
	openai.GPT3Dot5Turbo,
	openai.GPT4,
}

func main() {
	var (
		ctx     = context.Background()
		cmdList = flag.Bool("L", false, "list available models")

		model       = flag.String("m", availableModels[0], "model to use")
		prompt      = flag.String("p", "> ", "prompt to use")
		temperature = flag.Float64("t", 0.9, "temperature to use")
	)

	flag.Parse()

	authToken := os.Getenv(authEnvKey)
	if authToken == "" {
		fmt.Printf("%s is not set", authEnvKey)
		os.Exit(1)
	}

	client := openai.NewClient(authToken)

	if *cmdList {
		for _, m := range availableModels {
			fmt.Printf("%s\n", m)
		}
		os.Exit(0)
	}

	messages := make([]openai.ChatCompletionMessage, 0)
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("Conversation with %s\n", *model)
	fmt.Println("------------------------------")

	for {
		fmt.Print(*prompt)
		text, _ := reader.ReadString('\n')
		// convert CRLF to LF
		text = strings.Replace(text, "\n", "", -1)
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: text,
		})

		resp, err := client.CreateChatCompletion(
			ctx,
			openai.ChatCompletionRequest{
				Model:       *model,
				Messages:    messages,
				Temperature: float32(*temperature),
			},
		)

		if err != nil {
			fmt.Printf("ChatCompletion error: %v\n", err)
			continue
		}

		content := resp.Choices[0].Message.Content
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: content,
		})
		fmt.Println(content)
	}
}

// defaultModel returns default model to use
func defaultModel() string {
	if os.Getenv("CHATGPT_MODEL") != "" {
		return os.Getenv("CHATGPT_MODEL")
	}
	return openai.GPT3Dot5Turbo
}
