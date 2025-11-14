//go:build integration

package openai_test

import (
	"context"
	"fmt"
	"os"

	"micheam.com/aico/internal/assistant"
	openai "micheam.com/aico/internal/providers/openai"
)

func ExampleGPT4O_GenerateContent() {
	ctx := context.TODO()
	model := openai.NewGPT4O(os.Getenv("OPENAI_API_KEY"))
	sess, err := assistant.StartChat(model)
	if err != nil {
		fmt.Printf("error: %v", err)
		return
	}
	sess.SetSystemInstruction(assistant.NewTextContent(`
		You are a bot assistant that can answer questions.
		Please provide an answer to the following question,
		in "YES" or "NO".
	`))

	resp, err := sess.SendMessage(ctx, assistant.NewTextContent(`
		Are you a bot assistant?
	`))
	if err != nil {
		fmt.Printf("error: %v", err)
		return
	}

	fmt.Printf("%v", resp.Content.(*assistant.TextContent).Text)

	// Output:
	// YES
}

func ExampleGPT4O_GenerateContentStream() {
	ctx := context.TODO()
	model := openai.NewGPT4O(os.Getenv("OPENAI_API_KEY"))
	sess, err := assistant.StartChat(model)
	if err != nil {
		fmt.Printf("error: %v", err)
		return
	}

	iter, err := sess.SendMessageStream(ctx, assistant.NewTextContent(`
		Say this is a test.
	`))
	if err != nil {
		fmt.Printf("error: %v", err)
		return
	}

	for resp := range iter {
		fmt.Printf("%v", resp.Content.(*assistant.TextContent).Text)
	}

	// Output:
	// This is a test.
}
