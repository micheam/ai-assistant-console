//go:build e2e

package openai_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

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
		openai.DefaultChatModel,
		[]openai.Message{
			{Role: openai.RoleUser, Content: fmt.Sprintf("Today is %s", time.Now().Format("01/02"))},
			{Role: openai.RoleUser, Content: "What is the day after tomorrow?"},
		},
	)

	res, err := chat.Do(ctx, req)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Response: %+v", res)
	for _, choice := range res.Choices {
		b := &bytes.Buffer{}
		if err := json.NewEncoder(b).Encode(choice); err != nil {
			t.Fatal(err)
		}
		t.Logf("Choice: %s", b.String())
	}
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
		openai.DefaultChatModel,
		[]openai.Message{
			{Role: openai.RoleUser, Content: fmt.Sprintf("Today is %s", time.Now().Format("01/02"))},
			{Role: openai.RoleUser, Content: "What is the day after tomorrow?"},
		},
	)

	err := chat.DoStream(ctx, req, func(resp *openai.ChatResponse) error {
		t.Logf("Response: %+v", resp)
		for _, choice := range resp.Choices {
			b := &bytes.Buffer{}
			if err := json.NewEncoder(b).Encode(choice); err != nil {
				t.Fatal(err)
			}
			t.Logf("Choice: %s", b.String())
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

}
