//go:build e2e

package chat_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"micheam.com/aico/internal/openai"
	"micheam.com/aico/internal/openai/chat"
	"micheam.com/aico/internal/openai/models"
)

func TestChat_Do_EndToEnd(t *testing.T) {
	apikey := os.Getenv("OPENAI_API_KEY")
	if apikey == "" {
		t.Fatal("OPENAI_API_KEY is not set")
	}
	openaiClient := openai.NewClient(apikey)
	httpClient := &http.Client{Transport: &openai.DebugTransport{Transport: http.DefaultTransport}}
	openaiClient.SetHTTPClient(httpClient)
	client := chat.NewWithOpenAIClient(openaiClient)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	url, _ := url.Parse("https://upload.wikimedia.org/wikipedia/commons/thumb/d/dd/Gfp-wisconsin-madison-the-nature-boardwalk.jpg/2560px-Gfp-wisconsin-madison-the-nature-boardwalk.jpg")
	req := chat.NewRequest([]openai.Message{
		&openai.SystemMessage{Content: "Today is 01/02 and the weather is nice."},
		&openai.UserMessage{
			Content: []openai.Content{
				&openai.TextContent{Text: "What's in this image?"},
				&openai.ImageContent{URL: *url},
			},
		},
	}, chat.WithModel(models.GPT4OMini))

	res, err := client.Do(ctx, req)
	require.NoError(t, err)

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
	openaiClient := openai.NewClient(apikey)
	httpClient := &http.Client{Transport: &openai.DebugTransport{Transport: http.DefaultTransport}}
	openaiClient.SetHTTPClient(httpClient)
	client := chat.NewWithOpenAIClient(openaiClient)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req := chat.NewRequest([]openai.Message{
		&openai.SystemMessage{Content: "Today is 01/02 and the weather is nice."},
		&openai.UserMessage{
			Content: []openai.Content{
				&openai.TextContent{Text: "What is the day after tomorrow?"},
			},
		},
	}, chat.WithModel(models.GPT4OMini))

	err := client.DoStream(ctx, req, func(resp *chat.Response) error {
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
