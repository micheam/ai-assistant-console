# Technology Stack

## Architecture

CLI アプリケーションとして構築。インターフェース駆動のプロバイダー抽象化レイヤーにより、複数の AI プロバイダーを統一的に扱う。ストリーミングファーストの設計で Go 1.23+ イテレータを活用。

## Core Technologies

- **Language**: Go 1.24
- **CLI Framework**: `github.com/urfave/cli/v3`
- **Module Path**: `micheam.com/aico`
- **Build System**: Makefile (build, test, install, clean)

## Key Libraries

- `github.com/anthropics/anthropic-sdk-go` — Anthropic Claude の公式 SDK
- `github.com/BurntSushi/toml` — TOML 設定ファイルの解析
- `github.com/fatih/color` — ターミナルカラー出力
- `golang.org/x/term` — ターミナル検出と制御
- `github.com/stretchr/testify` — テストアサーション

## Development Standards

### Code Quality
- Go 標準のフォーマッティング (`gofmt`)
- `log/slog` による構造化ログ（JSON 形式）
- コンテキストベースの設定・ロガー伝搬パターン

### Error Handling
- `fmt.Errorf("...: %w", err)` によるエラーラッピング
- センチネルエラー（`ErrConfigFileNotFound` など）
- グレースフルデグラデーション（設定失敗時はデフォルト使用）

### Testing
- ユニットテスト: `go test ./...`
- インテグレーションテスト: `go test -tags=integration ./...`（API キー必要）
- `export_test.go` パターンで内部関数をテスト用にエクスポート
- `t.TempDir()` による一時ファイルテスト

## Development Environment

### Required Tools
- Go 1.24+
- Make
- Git

### Common Commands
```bash
# Build: make build
# Test: make test
# Install: make install
# Clean: make clean
# Protobuf: make protogen
```

## Key Technical Decisions

1. **OpenAI 互換プロバイダーの再利用**: Groq/Cerebras は `internal/providers/openai` の HTTP クライアントを共有。エンドポイント URL とモデル名のみが異なる
2. **Go イテレータによるストリーミング**: `iter.Seq[*GenerateContentResponse]` で全プロバイダーの統一インターフェース
3. **XDG Base Directory 準拠**: 設定ファイルは `$XDG_CONFIG_HOME/com.micheam.aico/config.toml`
4. **最小限の外部依存**: stdlib を最大限活用し、戦略的に SDK/ライブラリを追加
5. **コンテキスト伝搬パターン**: `config.WithConfig(ctx, cfg)` / `config.FromContext(ctx)` による設定の受け渡し

---
_Document standards and patterns, not every dependency_
