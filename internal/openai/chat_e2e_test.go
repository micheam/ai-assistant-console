//go:build e2e

package openai_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
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
	httpClient := &http.Client{Transport: &openai.DebugTransport{Transport: http.DefaultTransport}}
	client.SetHTTPClient(httpClient)
	chat := openai.NewChatClient(client)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	url, _ := url.Parse("https://upload.wikimedia.org/wikipedia/commons/thumb/d/dd/Gfp-wisconsin-madison-the-nature-boardwalk.jpg/2560px-Gfp-wisconsin-madison-the-nature-boardwalk.jpg")

	req := openai.NewChatRequest(
		openai.DefaultChatModel,
		[]openai.Message{
			&openai.SystemMessage{Content: "Today is 01/02 and the weather is nice."},
			&openai.UserMessage{
				Content: []openai.Content{
					&openai.TextContent{"What's in this image?"},
					&openai.ImageContent{*url},
				},
			},
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
	httpClient := &http.Client{Transport: &openai.DebugTransport{Transport: http.DefaultTransport}}
	client.SetHTTPClient(httpClient)
	chat := openai.NewChatClient(client)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req := openai.NewChatRequest(
		openai.DefaultChatModel,
		[]openai.Message{
			&openai.SystemMessage{Content: "Today is 01/02 and the weather is nice."},
			&openai.UserMessage{
				Content: []openai.Content{
					&openai.TextContent{"What is the day after tomorrow?"},
				},
			},
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
