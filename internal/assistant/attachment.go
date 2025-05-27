package assistant

import (
	"fmt"
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
