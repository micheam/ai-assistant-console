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
)

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

	// Get the frontmatter metadata
	metaData := meta.Get(context)

	// Process the frontmatter
	if title, ok := metaData["title"]; ok {
		sess.Title = title.(string)
	}

	if createdAt, ok := metaData["created_at"]; ok {
		strCreatedAt := createdAt.(string)
		sess.CreatedAt, err = time.Parse(time.RFC3339, strCreatedAt)
		if err != nil {
			return fmt.Errorf("parse created_at: %w", err)
		}
	}

	if sessionID, ok := metaData["session_id"]; ok {
		sess.ID = sessionID.(string)
	}

	// Process the document content
	var currentSection string
	var historyNumber int
	var currentAuthor MessageAuthor
	var currentMessage *Message // Track current message
	var currentContent strings.Builder // For accumulating text within a content block

	err = ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch v := n.(type) {
		case *ast.Heading:
			// Process headings to identify sections and messages
			if v.Level == 2 {
				// Level 2 headings define main sections (System Instructions, History)
				header := extractText(v, source)
				currentSection = header
				historyNumber = 0
				currentMessage = nil
			} else if v.Level == 3 && currentSection == "History" {
				// Level 3 headings in the History section define message authors
				header := extractText(v, source)

				// If we have a current message being processed, finalize any pending content
				if currentMessage != nil && currentContent.Len() > 0 {
					content := NewTextContent(strings.TrimSpace(currentContent.String()))
					currentMessage.Contents = append(currentMessage.Contents, content)
					currentContent.Reset()
				}

				historyNumber++

				// Parse the message author from the heading
				if strings.Contains(header, "User") {
					currentAuthor = MessageAuthorUser
				} else if strings.Contains(header, "Assistant") {
					currentAuthor = MessageAuthorAssistant
				}

				// Create a new message
				currentMessage = &Message{
					Author:   currentAuthor,
					Contents: []MessageContent{},
				}
				sess.History = append(sess.History, currentMessage)
			}

		case *ast.FencedCodeBlock:
			// Process code blocks
			if currentSection == "System Instructions" && entering {
				// For system instructions, we extract the code block content
				content := extractLinesText(v.Lines(), source)
				sess.SystemInstruction = NewTextContent(content)
			} else if currentSection == "History" && entering && currentMessage != nil {
				// If we have accumulated text content, add it first
				if currentContent.Len() > 0 {
					content := NewTextContent(strings.TrimSpace(currentContent.String()))
					currentMessage.Contents = append(currentMessage.Contents, content)
					currentContent.Reset()
				}

				// For code blocks in messages, create a new content with proper formatting
				content := extractLinesText(v.Lines(), source)
				info := ""
				if v.Info != nil {
					info = string(v.Info.Value(source))
				}

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
				currentMessage.Contents = append(currentMessage.Contents, codeContent)
			}

		case *ast.Paragraph:
			// Process paragraphs in the History section
			if currentSection == "History" && entering && currentMessage != nil {
				// We start a new content if needed
				content := extractText(v, source)
				
				// Store in the current content builder
				if currentContent.Len() > 0 {
					currentContent.WriteString("\n")
				}
				currentContent.WriteString(content)
			}
		}

		return ast.WalkContinue, nil
	})

	if err != nil {
		return fmt.Errorf("walk error: %w", err)
	}

	// Add any remaining content to the last message
	if currentSection == "History" && currentMessage != nil && currentContent.Len() > 0 {
		content := NewTextContent(strings.TrimSpace(currentContent.String()))
		currentMessage.Contents = append(currentMessage.Contents, content)
	}

	return nil
}

// extractText はノードからテキストを抽出する
func extractText(node ast.Node, source []byte) string {
	switch v := node.(type) {
	case *ast.Text:
		return string(v.Value(source))

	case *ast.String:
		return string(v.Value)

	case *ast.Heading:
		// ヘッダーは子ノードからテキストを取得
		return extractTextFromChildren(v, source)

	case *ast.Paragraph:
		// Paragraphは Lines() を使用
		if v.Lines() != nil {
			return extractLinesText(v.Lines(), source)
		}
		return extractTextFromChildren(v, source)

	case *ast.Link:
		// リンクは子ノードのテキストが表示テキスト
		return extractTextFromChildren(v, source)

	case *ast.Image:
		// 画像はaltテキストを使用
		return extractTextFromChildren(v, source)

	case *ast.CodeSpan:
		return extractLinesText(v.Lines(), source)

	case *ast.CodeBlock:
		return extractLinesText(v.Lines(), source)

	case *ast.FencedCodeBlock:
		return extractLinesText(v.Lines(), source)

	case *ast.Blockquote:
		return extractTextFromChildren(v, source)

	case *ast.List:
		return extractTextFromChildren(v, source)

	case *ast.ListItem:
		return extractTextFromChildren(v, source)

	case *ast.Emphasis:
		return extractTextFromChildren(v, source)

	default:
		// その他のノードは子ノードから再帰的に取得
		return extractTextFromChildren(v, source)
	}
}

func extractTextFromChildren(node ast.Node, source []byte) string {
	var result strings.Builder

	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if child.Kind() == ast.KindText {
			if textNode, ok := child.(*ast.Text); ok {
				result.Write(textNode.Value(source))
			}
		} else {
			text := extractText(child, source)
			result.WriteString(text)
		}

		// ソフトラインブレイクの処理
		if child.NextSibling() != nil {
			segment := child.NextSibling()
			if segment.Kind() == ast.KindText {
				// TODO: 改行の処理など
			}
		}
	}

	return result.String()
}

func extractLinesText(lines *text.Segments, source []byte) string {
	if lines == nil || lines.Len() == 0 {
		return ""
	}

	var result strings.Builder
	for i := range lines.Len() {
		segment := lines.At(i)
		result.Write(segment.Value(source))
	}

	return strings.TrimSpace(result.String())
}
