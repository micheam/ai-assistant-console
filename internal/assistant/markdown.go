package assistant

import (
	"io"
	"strings"

	_ "github.com/yuin/goldmark"
	_ "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/ast"
	_ "github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

func LoadMarkdown(sess *ChatSession, r io.Reader) error {

	// 	return fmt.Errorf("read file: %w", err)
	// }

	// md := goldmark.New(
	// 	goldmark.WithExtensions(meta.Meta),
	// )
	// context := parser.NewContext()
	// node := md.Parser().Parse(text.NewReader(source), parser.WithContext(context))
	// metaData := meta.Get(context)

	// // Frontmatter -----------------------------------------------------------------
	// if title, ok := metaData["title"]; ok {
	// 	sess.Title = title.(string)
	// }
	// if createdAt, ok := metaData["created_at"]; ok {
	// 	strCreatedAt := createdAt.(string)
	// 	sess.CreatedAt, err = time.Parse(time.RFC3339, strCreatedAt)
	// 	if err != nil {
	// 		return fmt.Errorf("parse created_at: %w", err)
	// 	}
	// }

	// // Handle the body of the document ----------------------------------------------
	// ast.Walk(node, func(currentNode ast.Node, entering bool) (ast.WalkStatus, error) {
	// 	switch n := currentNode.(type) {
	// 	case *ast.Heading:
	// 		if n.Level == 2 {
	// 			// H2 ヘッダーにより、セクションを分ける
	// 			// 1. Sistem Instructions
	// 			// 2. History（User, Assistant）
	// 			header := extractText(n, source)
	// 			fmt.Printf("%q\n", header)
	// 			switch header {
	// 			case "System Instructions":
	// 				// システムインストラクションを取得
	// 			case "History":
	// 				// ヒストリーを取得
	// 				// ユーザーとアシスタントのメッセージを取得
	// 				for child := n.NextSibling(); child != nil; child = child.NextSibling() {
	// 					if child.Kind() == ast.KindHeading {
	// 						// 次のヘッダーに到達したら終了
	// 						break
	// 					}
	// 					if child.Kind() == ast.KindBlockquote {
	// 						fmt.Println(extractText(child, source))
	// 					}
	// 				}
	// 			default:
	// 				// その他のヘッダーは無視
	// 				// 何もしない
	// 			}
	// 		}
	// 	}
	// 	return ast.WalkContinue, nil
	// })

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
