package assistant

import (
	"fmt"
	"io"
	"regexp"
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

		inAttachment:      false,
		attachmentName:    "",
		attachmentSyntax:  "",
		attachmentContent: strings.Builder{},
		contentOrder:      []MessageContent{},
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

		case *ast.HTMLBlock:
			return processHTMLBlock(state, v)
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

	inAttachment      bool
	attachmentName    string
	attachmentSyntax  string
	attachmentContent strings.Builder

	// Track content order for proper sequencing
	contentOrder []MessageContent
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

		// Reset content order for the new message
		state.contentOrder = []MessageContent{}
	}

	return ast.WalkContinue, nil
}

// processFencedCodeBlock handles code block nodes in the AST.
func processFencedCodeBlock(state *processState, v *ast.FencedCodeBlock) (ast.WalkStatus, error) {
	if state.currentSection == "System Instructions" {
		content := markdown.ExtractLinesText(v.Lines(), state.source)
		state.session.SystemInstruction = NewTextContent(content)
	} else if state.currentSection == "History" && state.currentMessage != nil {
		content := markdown.ExtractLinesText(v.Lines(), state.source)
		info := ""
		if v.Info != nil {
			info = string(v.Info.Value(state.source))
		}

		if state.inAttachment {
			// This is an attachment content, store it
			state.attachmentSyntax = info
			state.attachmentContent.WriteString(content)
			return ast.WalkContinue, nil
		}

		// Add any accumulated text content first
		if state.currentContent.Len() > 0 {
			textContent := NewTextContent(strings.TrimSpace(state.currentContent.String()))
			state.contentOrder = append(state.contentOrder, textContent)
			state.currentContent.Reset()
		}

		// For code blocks in messages, create a new content with proper formatting
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

		// Add the code block to the content order
		codeContent := NewTextContent(codeBlock.String())
		state.contentOrder = append(state.contentOrder, codeContent)
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
		// Skip paragraphs inside attachments
		if state.inAttachment {
			return ast.WalkContinue, nil
		}

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

		// Immediately add text content to maintain order
		if state.currentContent.Len() > 0 {
			textContent := NewTextContent(strings.TrimSpace(state.currentContent.String()))
			state.contentOrder = append(state.contentOrder, textContent)
			state.currentContent.Reset()
		}
	}

	return ast.WalkContinue, nil
}

// Regular expressions for parsing artifact details
var (
	// Match <details> tag (with or without newlines/spaces)
	detailsStartRegex = regexp.MustCompile(`(?s)<details>`)
	// Match <summary>{name}</summary> tag (with or without newlines/spaces)
	summaryRegex = regexp.MustCompile(`(?s)<summary>(.*?)</summary>`)
	// Match </details> tag
	detailsEndRegex = regexp.MustCompile(`</details>`)
)

// processHTMLBlock handles HTML block nodes in the AST, specifically for attachments.
func processHTMLBlock(state *processState, v *ast.HTMLBlock) (ast.WalkStatus, error) {
	if state.currentSection != "History" || state.currentMessage == nil {
		return ast.WalkContinue, nil
	}

	content := markdown.ExtractLinesText(v.Lines(), state.source)

	// If we're not in an attachment yet, check if this block starts an attachment
	if !state.inAttachment {
		if detailsStartRegex.MatchString(content) {
			// We found the opening <details> tag
			state.inAttachment = true

			// Look for the summary tag which might be in the same HTML block
			if matches := summaryRegex.FindStringSubmatch(content); len(matches) > 1 {
				state.attachmentName = normalizeAttachmentName(matches[1])
			}

			state.attachmentContent.Reset()
			return ast.WalkContinue, nil
		}
	} else {
		// If we're already in an attachment, we might need to look for the summary tag
		if state.attachmentName == "" {
			if matches := summaryRegex.FindStringSubmatch(content); len(matches) > 1 {
				state.attachmentName = normalizeAttachmentName(matches[1])
				return ast.WalkContinue, nil
			}
		}

		// Check if this is the end of an attachment
		if detailsEndRegex.MatchString(content) {
			// Create a new attachment content
			attachment := NewAttachmentContent(
				state.attachmentName,
				state.attachmentSyntax,
				strings.TrimSpace(state.attachmentContent.String()),
			)

			// Add the attachment to the content order
			state.contentOrder = append(state.contentOrder, attachment)

			// Reset attachment state
			state.inAttachment = false
			state.attachmentName = ""
			state.attachmentSyntax = ""
			state.attachmentContent.Reset()
		}
	}

	return ast.WalkContinue, nil
}

func normalizeAttachmentName(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "Attachment:")
	return strings.TrimSpace(s)
}

// finalizeCurrentContent adds any accumulated content to the current message.
func finalizeCurrentContent(state *processState) {
	if state.currentSection == "History" && state.currentMessage != nil {
		// Add any remaining accumulated text content (though it should already be processed)
		if state.currentContent.Len() > 0 {
			content := NewTextContent(strings.TrimSpace(state.currentContent.String()))
			state.contentOrder = append(state.contentOrder, content)
			state.currentContent.Reset()
		}

		// Add all content in the correct order
		state.currentMessage.Contents = append(state.currentMessage.Contents, state.contentOrder...)
		state.contentOrder = []MessageContent{}
	}
}

