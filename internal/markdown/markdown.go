package markdown

import (
	"io"
	"iter"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// パッケージレベルのパーサー設定
var defaultParser = goldmark.New()

// io.Reader から AST を構築するヘルパー
func parseFromReader(src io.Reader) (ast.Node, []byte, error) {
	content, err := io.ReadAll(src)
	if err != nil {
		return nil, nil, err
	}

	reader := text.NewReader(content)
	doc := defaultParser.Parser().Parse(reader)

	return doc, content, nil
}

// H2などの特定レベルのヘッダーを取得
func Headings(src io.Reader, level int) iter.Seq2[*ast.Heading, string] {
	doc, content, err := parseFromReader(src)
	if err != nil {
		// エラーハンドリング: 空のイテレータを返す
		return func(yield func(*ast.Heading, string) bool) {}
	}

	return func(yield func(*ast.Heading, string) bool) {
		ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
			if entering {
				if heading, ok := node.(*ast.Heading); ok && heading.Level == level {
					text := string(node.Text(content))
					if !yield(heading, text) {
						return ast.WalkStop, nil
					}
				}
			}
			return ast.WalkContinue, nil
		})
	}
}

// 全てのヘッダーを取得
func AllHeadings(src io.Reader) iter.Seq2[*ast.Heading, string] {
	doc, content, err := parseFromReader(src)
	if err != nil {
		return func(yield func(*ast.Heading, string) bool) {}
	}

	return func(yield func(*ast.Heading, string) bool) {
		ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
			if entering {
				if heading, ok := node.(*ast.Heading); ok {
					text := string(node.Text(content))
					if !yield(heading, text) {
						return ast.WalkStop, nil
					}
				}
			}
			return ast.WalkContinue, nil
		})
	}
}

// セクション情報
type Section struct {
	Heading *ast.Heading
	Title   string
	Level   int
	Content []ast.Node
}

// セクション単位でのイテレータ
func Sections(src io.Reader, level int) iter.Seq[*Section] {
	doc, content, err := parseFromReader(src)
	if err != nil {
		return func(yield func(*Section) bool) {}
	}

	return func(yield func(*Section) bool) {
		var currentSection *Section
		var collecting bool

		ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
			if !entering {
				return ast.WalkContinue, nil
			}

			if heading, ok := node.(*ast.Heading); ok {
				// 現在のセクションを出力してから新しいセクション開始
				if currentSection != nil && heading.Level <= level {
					if !yield(currentSection) {
						return ast.WalkStop, nil
					}
					collecting = false
				}

				// 指定レベルのヘッダーで新しいセクション開始
				if heading.Level == level {
					currentSection = &Section{
						Heading: heading,
						Title:   string(node.Text(content)),
						Level:   level,
						Content: make([]ast.Node, 0),
					}
					collecting = true
				}
			} else if collecting && currentSection != nil {
				// コンテンツを収集
				currentSection.Content = append(currentSection.Content, node)
			}

			return ast.WalkContinue, nil
		})

		// 最後のセクションを出力
		if currentSection != nil {
			yield(currentSection)
		}
	}
}

// ジェネリック版: 特定の型のノードを取得
func Nodes[T ast.Node](src io.Reader) iter.Seq2[T, string] {
	doc, content, err := parseFromReader(src)
	if err != nil {
		return func(yield func(T, string) bool) {}
	}

	return func(yield func(T, string) bool) {
		ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
			if entering {
				if typedNode, ok := node.(T); ok {
					text := string(node.Text(content))
					if !yield(typedNode, text) {
						return ast.WalkStop, nil
					}
				}
			}
			return ast.WalkContinue, nil
		})
	}
}

// 便利関数: 特定の型のノードだけを取得
func Paragraphs(src io.Reader) iter.Seq2[*ast.Paragraph, string] {
	return Nodes[*ast.Paragraph](src)
}

func Lists(src io.Reader) iter.Seq2[*ast.List, string] {
	return Nodes[*ast.List](src)
}

func Links(src io.Reader) iter.Seq2[*ast.Link, string] {
	return Nodes[*ast.Link](src)
}

// エラーハンドリング付きの関数も提供
func HeadingsWithError(src io.Reader, level int) (iter.Seq2[*ast.Heading, string], error) {
	doc, content, err := parseFromReader(src)
	if err != nil {
		return nil, err
	}

	seq := func(yield func(*ast.Heading, string) bool) {
		ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
			if entering {
				if heading, ok := node.(*ast.Heading); ok && heading.Level == level {
					text := string(node.Text(content))
					if !yield(heading, text) {
						return ast.WalkStop, nil
					}
				}
			}
			return ast.WalkContinue, nil
		})
	}

	return seq, nil
}

// カスタマイズ用の設定構造体
type Config struct {
	EnableGFM       bool
	EnableTables    bool
	EnableFootnotes bool
}

type Option func(*Config)

func WithGFM() Option {
	return func(c *Config) { c.EnableGFM = true }
}

func WithTables() Option {
	return func(c *Config) { c.EnableTables = true }
}

func WithFootnotes() Option {
	return func(c *Config) { c.EnableFootnotes = true }
}

// カスタマイズされたパーサーを作成
func NewParser(opts ...Option) *Parser {
	config := &Config{}
	for _, opt := range opts {
		opt(config)
	}

	// goldmarkのオプションを設定
	var extensions []goldmark.Extender
	if config.EnableGFM {
		// 実際には extension.GFM を追加
	}

	md := goldmark.New(goldmark.WithExtensions(extensions...))

	return &Parser{md: md}
}

type Parser struct {
	md goldmark.Markdown
}

// カスタムパーサーでのヘッダー取得
func (p *Parser) Headings(src io.Reader, level int) iter.Seq2[*ast.Heading, string] {
	content, err := io.ReadAll(src)
	if err != nil {
		return func(yield func(*ast.Heading, string) bool) {}
	}

	reader := text.NewReader(content)
	doc := p.md.Parser().Parse(reader)

	return func(yield func(*ast.Heading, string) bool) {
		ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
			if entering {
				if heading, ok := node.(*ast.Heading); ok && heading.Level == level {
					text := string(node.Text(content))
					if !yield(heading, text) {
						return ast.WalkStop, nil
					}
				}
			}
			return ast.WalkContinue, nil
		})
	}
}
