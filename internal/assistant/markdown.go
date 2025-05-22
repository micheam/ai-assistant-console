package assistant

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"

	"micheam.com/aico/internal/markdown"
)

// LoadMarkdown parses a markdown file and loads its content into a ChatSession.
// The markdown file should follow a specific format with frontmatter, system instructions,
// and chat history sections.
func LoadMarkdown(sess *ChatSession, r io.Reader) error {
	// Read the entire file
	source, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	// Create a new parser with meta extension for frontmatter
	md := goldmark.New(
		goldmark.WithExtensions(meta.Meta),
	)

	// Parse the markdown content
	context := parser.NewContext()
	reader := text.NewReader(source)
	node := md.Parser().Parse(reader, parser.WithContext(context))

	// Process the frontmatter metadata
	if err := processFrontmatter(sess, meta.Get(context)); err != nil {
		return err
	}

	// Process the document content (system instructions and chat history)
	err = processDocument(sess, node, source)
	if err != nil {
		return fmt.Errorf("process document: %w", err)
	}

	return nil
}

// processFrontmatter extracts metadata from the frontmatter and sets it on the ChatSession.
func processFrontmatter(sess *ChatSession, metaData map[string]any) error {
	// Process title if present
	if title, ok := metaData["title"]; ok {
		sess.Title = title.(string)
	}

	// Process creation timestamp if present
	if createdAt, ok := metaData["created_at"]; ok {
		strCreatedAt := createdAt.(string)
		parsedTime, err := time.Parse(time.RFC3339, strCreatedAt)
		if err != nil {
			return fmt.Errorf("parse created_at: %w", err)
		}
		sess.CreatedAt = parsedTime
	}

	// Process session ID if present
	if sessionID, ok := metaData["session_id"]; ok {
		sess.ID = sessionID.(string)
	}

	return nil
}

// processDocument walks the AST and processes the markdown document content.
func processDocument(sess *ChatSession, node ast.Node, source []byte) error {
	// State for tracking the document traversal
	state := &processState{
		source:         source,
		currentSection: "",
		historyNumber:  0,
		currentAuthor:  "",
		currentMessage: nil,
		currentContent: strings.Builder{},
		session:        sess,
	}

	// Walk the AST and process each node
	err := ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		// TODO: ここでの処理は state.currentSection で振り分ける方が見通しが良い希ガス

		switch v := n.(type) {

		case *ast.Heading:
			return processHeading(state, v)

		case *ast.FencedCodeBlock:
			return processFencedCodeBlock(state, v)

		case *ast.Paragraph:
			return processParagraph(state, v)
		}

		return ast.WalkContinue, nil
	})

	if err != nil {
		return fmt.Errorf("walk error: %w", err)
	}

	// Add any remaining content to the last message
	finalizeCurrentContent(state)

	return nil
}

// processState tracks the state while traversing the AST.
type processState struct {
	source         []byte
	currentSection string
	historyNumber  int
	currentAuthor  MessageAuthor
	currentMessage *Message
	currentContent strings.Builder
	session        *ChatSession
}

// processHeading handles heading nodes in the AST.
func processHeading(state *processState, v *ast.Heading) (ast.WalkStatus, error) {
	if v.Level == 2 {
		// Level 2 headings define main sections (System Instructions, History)
		header := markdown.ExtractTextFromChildren(v, state.source)
		state.currentSection = header
		state.historyNumber = 0
		state.currentMessage = nil
	} else if v.Level == 3 && state.currentSection == "History" {
		// Level 3 headings in the History section define message authors
		header := markdown.ExtractTextFromChildren(v, state.source)

		// Finalize any pending content for the previous message
		finalizeCurrentContent(state)

		state.historyNumber++

		// Parse the message author from the heading
		if strings.Contains(header, "User") {
			state.currentAuthor = MessageAuthorUser
		} else if strings.Contains(header, "Assistant") {
			state.currentAuthor = MessageAuthorAssistant
		}

		// Create a new message
		state.currentMessage = &Message{
			Author:   state.currentAuthor,
			Contents: []MessageContent{},
		}
		state.session.History = append(state.session.History, state.currentMessage)
	}

	return ast.WalkContinue, nil
}

// processFencedCodeBlock handles code block nodes in the AST.
func processFencedCodeBlock(state *processState, v *ast.FencedCodeBlock) (ast.WalkStatus, error) {
	if state.currentSection == "System Instructions" {
		content := markdown.ExtractLinesText(v.Lines(), state.source)
		state.session.SystemInstruction = NewTextContent(content)
	} else if state.currentSection == "History" && state.currentMessage != nil {
		// Finalize any pending content before adding the code block
		finalizeCurrentContent(state)

		// For code blocks in messages, create a new content with proper formatting
		content := markdown.ExtractLinesText(v.Lines(), state.source)
		info := ""
		if v.Info != nil {
			info = string(v.Info.Value(state.source))
		}

		// Format the code block with proper markdown syntax
		var codeBlock strings.Builder
		if info != "" {
			codeBlock.WriteString("```")
			codeBlock.WriteString(info)
			codeBlock.WriteString("\n")
		} else {
			codeBlock.WriteString("```\n")
		}
		codeBlock.WriteString(content)
		codeBlock.WriteString("\n```\n")

		// Add the code block as a separate content
		codeContent := NewTextContent(codeBlock.String())
		state.currentMessage.Contents = append(state.currentMessage.Contents, codeContent)
	}

	return ast.WalkContinue, nil
}

// processParagraph handles paragraph nodes in the AST.
func processParagraph(state *processState, v *ast.Paragraph) (ast.WalkStatus, error) {
	if state.currentSection == "System Instructions" {
		// Extract the text content from paragraph
		content := ""
		if v.Lines().Len() > 0 {
			content = markdown.ExtractLinesText(v.Lines(), state.source)
		} else {
			content = markdown.ExtractTextFromChildren(v, state.source)
		}
		// Add to the system instruction content
		if state.session.SystemInstruction == nil {
			state.session.SystemInstruction = NewTextContent(content)
		} else {
			// Append to the existing system instruction
			state.session.SystemInstruction.Text += "\n" + content
		}
	} else if state.currentSection == "History" && state.currentMessage != nil {
		// Extract the text content from paragraph
		content := ""
		if v.Lines().Len() > 0 {
			content = markdown.ExtractLinesText(v.Lines(), state.source)
		} else {
			content = markdown.ExtractTextFromChildren(v, state.source)
		}

		// Add to the current content builder with proper spacing
		if state.currentContent.Len() > 0 {
			state.currentContent.WriteString("\n")
		}
		state.currentContent.WriteString(content)
	}

	return ast.WalkContinue, nil
}

// finalizeCurrentContent adds any accumulated content to the current message.
func finalizeCurrentContent(state *processState) {
	if state.currentSection == "History" &&
		state.currentMessage != nil &&
		state.currentContent.Len() > 0 {
		content := NewTextContent(strings.TrimSpace(state.currentContent.String()))
		state.currentMessage.Contents = append(state.currentMessage.Contents, content)
		state.currentContent.Reset()
	}
}
