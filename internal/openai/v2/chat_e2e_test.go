// //go:build e2e

package openai_test

import (
	"context"
	"os"
	"testing"

	openai "micheam.com/aico/internal/openai/v2"
)

func TestChat_Do_EndToEnd(t *testing.T) {
	apikey := os.Getenv("OPENAI_API_KEY")
	if apikey == "" {
		t.Fatal("OPENAI_API_KEY is not set")
	}
	client, err := openai.NewClient(apikey)
	if err != nil {
		t.Fatal(err)
	}
	chat := openai.NewChatClient(client)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req := openai.NewChatRequest(
		"gpt-4o-mini",
		[]openai.Message{
			{Role: openai.RoleUser, Content: "Hello, How are you?"},
		},
	)
	cb := func(resp *openai.ChatResponse) error {
		t.Logf("chunk: %+v", resp)
		return nil
	}
	if err := chat.DoStream(ctx, req, cb); err != nil {
		t.Fatal(err)
	}
}
