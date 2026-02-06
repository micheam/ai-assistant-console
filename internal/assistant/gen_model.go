package assistant

import (
	"context"
	"iter"
)

type ModelDescriptor interface {
	Name() string
	Description() string
	Provider() string
	Deprecated() bool
	DeprecatedRemovedIn() string
}

// GenerativeModel represents a generative model.
type GenerativeModel interface {
	ModelDescriptor
	SetSystemInstruction(...*TextContent)
	GenerateContent(context.Context, ...*Message) (*GenerateContentResponse, error)
	GenerateContentStream(context.Context, ...*Message) (iter.Seq[*GenerateContentResponse], error)
}

type GenerateContentResponse struct {
	Content MessageContent
}
