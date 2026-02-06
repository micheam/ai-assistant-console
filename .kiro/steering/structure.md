# Project Structure

## Organization Philosophy

フラットな内部パッケージ構造を採用。各パッケージは自己完結的で、深いネストを避ける。プロバイダーパッケージは外部サービスのアイデンティティを反映した命名。

## Directory Patterns

### CLI Application (`cmd/aico/`)
**Purpose**: アプリケーションのエントリーポイントとコマンド定義
**Pattern**: コマンドごとに1ファイル（`generate.go`, `models.go`, `persona.go`）
**Example**: `main.go` でグローバルフラグとコマンドレジストリを定義

### Core Abstractions (`internal/assistant/`)
**Purpose**: AI モデルとメッセージの共通インターフェース定義
**Pattern**: `GenerativeModel` インターフェースを中心とした抽象化
**Key Interface**:
```go
type GenerativeModel interface {
    Name() string
    Description() string
    Provider() string
    SetSystemInstruction(...*TextContent)
    GenerateContent(ctx, ...*Message) (*GenerateContentResponse, error)
    GenerateContentStream(ctx, ...*Message) (iter.Seq[*GenerateContentResponse], error)
}
```

### Provider Implementations (`internal/providers/{name}/`)
**Purpose**: 各 AI プロバイダーの具体的な実装
**Pattern**: プロバイダーごとにディレクトリ、モデルごとに1ファイル
**Example**:
```
providers/anthropic/
├── anthropic.go          # AvailableModels(), NewGenerativeModel(), DescribeModel()
├── claude_opus_4_6.go    # 個別モデル実装
└── claude_sonnet_4_5.go
```
**Registry Pattern**: 各プロバイダーは3つのファクトリ関数をエクスポート:
- `AvailableModels() []assistant.ModelDescriptor`
- `NewGenerativeModel(name, apiKey string) (GenerativeModel, error)`
- `DescribeModel(name string) (string, bool)`

### Configuration (`internal/config/`)
**Purpose**: TOML 設定の読み込みとコンテキスト伝搬
**Pattern**: XDG 準拠のパス解決 + `context.Context` ベースの設定受け渡し

### Utilities (`internal/logging/`, `internal/spinner/`, `internal/theme/`, `internal/pointer/`)
**Purpose**: 横断的関心事のユーティリティ
**Pattern**: 薄いラッパーで stdlib を拡張

### Vim Plugin (`plugin/`, `autoload/`)
**Purpose**: Vim エディタとの統合
**Pattern**: Vim9script + autoload パターン。`aico` バイナリをサブプロセスとして呼び出し

## Naming Conventions

- **Files**: `snake_case.go`（Go 標準）。モデルファイルは `{model_name}.go`（例: `claude_opus_4_6.go`）
- **Packages**: 小文字、アンダースコアなし（`assistant`, `config`, `openai`）
- **Types**: PascalCase。インターフェースは概念名（`GenerativeModel`）、実装は `{Provider}{Model}`（`ClaudeOpus4_6`）
- **Functions**: `New{Type}` コンストラクタ、`{Verb}{Noun}` アクション（`GenerateContent`, `LoadConfig`）

## Import Organization

```go
import (
    // stdlib
    "context"
    "fmt"

    // internal packages
    "micheam.com/aico/internal/assistant"
    "micheam.com/aico/internal/providers/openai"

    // external dependencies
    "github.com/urfave/cli/v3"
)
```

**Module Path**: `micheam.com/aico`
**Internal Packages**: すべて `internal/` 配下で外部からのインポートを防止

## Code Organization Principles

1. **インターフェース定義は `assistant` パッケージ**: プロバイダーは実装のみ提供
2. **OpenAI 互換の共有**: Groq/Cerebras は `openai.APIClient` を埋め込み、プロトコル実装を再利用
3. **1モデル1ファイル**: 新モデル追加時は新ファイルを作成し、プロバイダーのレジストリに登録
4. **CLI コマンドの独立性**: 各コマンドは独立したファイルで定義、`main.go` で登録

---
_Document patterns, not file trees. New files following patterns shouldn't require updates_
