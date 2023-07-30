package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"micheam.com/aico/internal/openai"
)

var SendMessageCommand = &cli.Command{
	Name:        "send",
	Usage:       "Send message to AI",
	Description: "Send message to AI and get response",
	ArgsUsage:   "MESSAGE",
	Action:      sendMessage,
}

func sendMessage(c *cli.Context) error {
	ctx := c.Context
	conf := ConfigFrom(ctx)
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

	// TODO: Choose Personality from Config

	if c.NArg() == 0 {
		return fmt.Errorf("message is not set")
	}
	messages = append(messages, openai.Message{
		Content: c.Args().First(),
		Role:    openai.RoleUser,
	})

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

	if len(resp.Choices) > 0 {
		msg := resp.Choices[0].Message
		fmt.Println(msg.Content)
	}

	return nil
}
