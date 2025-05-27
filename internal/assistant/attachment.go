package assistant

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// AttachmentContent represents an attachment content in a message.
type AttachmentContent struct {
	Name    string
	Syntax  string
	Content []byte
}

var _ MessageContent = (*AttachmentContent)(nil)

func (*AttachmentContent) isMessageContent() {}

// NewAttachmentContent creates a new attachment content.
func NewAttachmentContent(name, syntax, content string) MessageContent {
	return &AttachmentContent{
		Name:    name,
		Syntax:  syntax,
		Content: []byte(content),
	}
}

// String returns the string representation of the attachment content.
func (c *AttachmentContent) String() string {
	return fmt.Sprintf("<Attachment: %s>", c.Name)
}

// Example:
//
//		input
//		{
//			"Name": "file.txt",
//			"Syntax: "plaintext",
//			"Content": "Hello, world!"
//		}
//
//	 output
//		<details><summary>Attachment: file.txt</summary>
//
//		```plaintext
//		Hello, world!
//		```
//
//		</details>
func (a AttachmentContent) ToText() string {
	var sb = &strings.Builder{}
	sb.WriteString("<details><summary>Attachment: ")
	sb.WriteString(a.Name)
	sb.WriteString("</summary>\n\n")
	sb.WriteString("```")
	sb.WriteString(a.Syntax)
	sb.WriteString("\n")
	sb.WriteString(string(a.Content))
	sb.WriteString("\n```\n\n")
	sb.WriteString("</details>")
	return sb.String()
}

func (a *AttachmentContent) LoadFile(filePath string) error {
	var (
		name, syntax string
		content      []byte
	)

	name = filepath.Base(filePath)
	syntax = detectLauguage(filePath)

	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer f.Close()

	content, err = io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	a.Name = name
	a.Syntax = syntax
	a.Content = content

	return nil
}

func detectLauguage(filePath string) string {
	ext := filepath.Ext(filePath)
	switch ext {
	case ".txt":
		return "plaintext"
	case ".json":
		return "json"
	case ".yaml", ".yml":
		return "yaml"
	case ".md":
		return "markdown"
	case ".go":
		return "go"
	case ".zig":
		return "zig"
	case ".py":
		return "python"
	case ".js":
		return "javascript"
	case ".html":
		return "html"
	case ".css":
		return "css"
	case ".java":
		return "java"
	case ".cpp", ".cxx", ".cc":
		return "cpp"
	case ".c":
		return "c"
	case ".rs":
		return "rust"
	case ".sh":
		return "shell"
	case ".sql":
		return "sql"
	case ".xml":
		return "xml"
	case ".php":
		return "php"
	case ".rb":
		return "ruby"
	case ".swift":
		return "swift"
	case ".ts":
		return "typescript"
	case ".kt":
		return "kotlin"
	case ".lua":
		return "lua"
	case ".pl":
		return "perl"
	case ".hs":
		return "haskell"
	case ".dart":
		return "dart"
	default:
		return "plaintext" // Default syntax if not recognized
	}
}
