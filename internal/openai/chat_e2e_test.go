//go:build e2e

package openai_test

import (
	"context"
	"os"
	"testing"

	"micheam.com/aico/internal/openai"
)

func TestChat_Do_EndToEnd(t *testing.T) {
	apikey := os.Getenv("OPENAI_API_KEY")
	if apikey == "" {
		t.Fatal("OPENAI_API_KEY is not set")
	}
	client := openai.NewClient(apikey)
	chat := openai.NewChatClient(client)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req := openai.NewChatRequest(
		"",
		[]openai.Message{
			{Role: openai.RoleUser, Content: "Hello, How are you?"},
		},
	)

	res, err := chat.Do(ctx, req)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Response: %+v", res)
}

func TestChat_DoStream_EndToEnd(t *testing.T) {
	apikey := os.Getenv("OPENAI_API_KEY")
	if apikey == "" {
		t.Fatal("OPENAI_API_KEY is not set")
	}
	client := openai.NewClient(apikey)
	chat := openai.NewChatClient(client)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req := openai.NewChatRequest(
		"",
		[]openai.Message{
			{Role: openai.RoleUser, Content: "Hello, How are you?"},
		},
	)

	err := chat.DoStream(ctx, req, func(resp *openai.ChatResponse) error {
		t.Logf("%s: %+v", resp.Choices[0].Delta.Content, resp.Choices[0])
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
