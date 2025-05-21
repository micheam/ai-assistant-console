package markdown

import (
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// ExtractTextFromChildren recursively extracts text from child nodes.
//
// This function traverses the first child and its siblings of
// the given node, extracting text from each node.
func ExtractTextFromChildren(node ast.Node, source []byte) string {
	var result strings.Builder

	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		// Handle different node types
		switch v := child.(type) {

		case *ast.Text:
			result.WriteString(string(v.Value(source)))

		case *ast.String:
			result.WriteString(string(v.Value))

		case *ast.CodeSpan, *ast.CodeBlock, *ast.FencedCodeBlock:
			// NOTE: These nodes have a Lines() method
			if lines := v.Lines(); lines != nil {
				result.WriteString(ExtractLinesText(lines, source))
			}

		default:
			// Recursively extract text from child nodes
			result.WriteString(ExtractTextFromChildren(v, source))
		}
	}
	return result.String()
}

// ExtractLinesText extracts text from line segments.
func ExtractLinesText(lines *text.Segments, source []byte) string {
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
