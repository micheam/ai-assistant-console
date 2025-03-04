package assistant

import (
	"context"
	"iter"
)

// GenerativeModel represents a generative model.
type GenerativeModel interface {
	Description() string
	SetSystemInstruction(*TextContent)
	GenerateContent(context.Context, ...*Message) (*GenerateContentResponse, error)
	GenerateContentStream(context.Context, ...*Message) (iter.Seq[*GenerateContentResponse], error)
}

type GenerateContentResponse struct {
	Content MessageContent
}
