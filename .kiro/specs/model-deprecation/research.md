# Research & Design Decisions

## Summary
- **Feature**: `model-deprecation`
- **Discovery Scope**: Extension（既存システムの拡張）
- **Key Findings**:
  - 現在の非推奨管理は `Description()` 内のテキストベースで、構造化されていない
  - Go の struct embedding を活用することで、非推奨でないモデルのボイラープレートを最小化できる
  - `ModelDescriptor` インターフェースへのメソッド追加は全モデル実装に影響するが、embedding パターンで軽減可能

## Research Log

### ModelDescriptor インターフェース拡張の影響範囲
- **Context**: `ModelDescriptor` に `Deprecated()` / `DeprecatedRemovedIn()` メソッドを追加すると、全モデル実装（約13構造体）に影響
- **Sources Consulted**: コードベース内の全モデル実装を調査
- **Findings**:
  - Anthropic: 4モデル（1非推奨）
  - OpenAI: 10モデル（4非推奨）
  - Groq: 3モデル（0非推奨）
  - Cerebras: 3モデル（0非推奨）
  - 合計20モデル中5モデルが非推奨
- **Implications**: 非推奨でない15モデルに対してもメソッド実装が必要。embedding パターンで対応可能

### 既存の非推奨パターン分析
- **Context**: 現在のテキストベースの非推奨表現の一貫性を確認
- **Findings**:
  - `[Deprecated]` プレフィックス + `superseded by` テキストが共通パターン
  - `AvailableModels()` では非推奨モデルもコメント付きで列挙
  - `selectModel()` / `NewGenerativeModel()` では非推奨モデルも引き続き利用可能
- **Implications**: 後方互換性を維持しつつ、構造化メタデータへ移行する

### CLI フラグの既存パターン
- **Context**: `models list` コマンドの既存フラグ設計を確認
- **Findings**:
  - 既存フラグ: `--json` のみ
  - `cli.BoolFlag` パターンで統一
  - `cmd.Bool(flagName)` で参照
- **Implications**: `--all` フラグは同一パターンで追加可能

### 警告出力の挿入ポイント
- **Context**: 非推奨モデル使用時の警告をどこに挿入するか
- **Findings**:
  - `detectModel()` in `cmd/aico/models.go` がモデル解決の中心
  - `generate.go` から呼ばれ、`GenerativeModel` を返す
  - 警告は `detectModel()` 内でモデル解決後、return 前に `fmt.Fprintf(os.Stderr, ...)` で出力可能
  - `ModelDescriptor` 情報は `GenerativeModel` 経由で取得可能
- **Implications**: `detectModel()` の返却値を変更する必要はなく、副作用として stderr 出力のみ追加

## Architecture Pattern Evaluation

| Option | Description | Strengths | Risks / Limitations | Notes |
|--------|-------------|-----------|---------------------|-------|
| A: インターフェース直接拡張 | `ModelDescriptor` に `Deprecated()` / `DeprecatedRemovedIn()` を追加、各モデルで直接実装 | 明確、型安全 | 全モデル（20ファイル）に2メソッド追加のボイラープレート | シンプルだがコード量が多い |
| B: 別インターフェース + 型アサーション | `DeprecatableModel` インターフェースを新設、型アサーションで判定 | 既存コード変更不要 | 呼び出し側に型アサーションが散在、コンパイル時安全性が低い | Go のイディオムとしてはあり得るが煩雑 |
| **C: Embedding パターン** | `DeprecationInfo` 構造体を定義し各モデルに embed | ゼロ値で非推奨でない扱い、ボイラープレート最小 | 若干の間接性 | **採用**: Go イディオムに沿い、最も変更量が少ない |

## Design Decisions

### Decision: DeprecationInfo Embedding パターンの採用
- **Context**: `ModelDescriptor` にメソッドを追加する際、全モデルへの影響を最小化する手法が必要
- **Alternatives Considered**:
  1. 各モデルに直接メソッド追加 — 明確だが20ファイル変更
  2. 別インターフェース — 型安全性が低下
  3. Embedding — ゼロ値で非推奨でない扱い
- **Selected Approach**: `assistant.DeprecationInfo` 構造体を定義し、全モデル構造体に embed。非推奨モデルのみコンストラクタで値を設定
- **Rationale**: Go のゼロ値セマンティクスを活用し、非推奨でないモデルは変更不要（embed 追加のみ）。インターフェースメソッドはコンパイル時に保証
- **Trade-offs**: 構造体の embed は若干の間接性があるが、明示的で理解しやすい
- **Follow-up**: 各プロバイダーの `AvailableModels()` が返すインスタンスの非推奨設定を確認

### Decision: `--all` フラグによるフィルタリング制御
- **Context**: 非推奨モデルのデフォルト非表示と明示的表示の切り替え
- **Selected Approach**: `--all` フラグ追加。デフォルトで非推奨を除外、`--all` で全表示
- **Rationale**: `--all` は CLI ツールで広く使われるイディオム（`docker images -a`、`brew list --all` 等）

### Decision: スキル命名規則 `/model-*`
- **Context**: 既存 `add-model` スキルとの命名一貫性
- **Selected Approach**: `add-model` → `model-add` にリネーム、新規 `model-deprecate` を作成
- **Rationale**: `/model-*` プレフィックスでモデル関連操作をグループ化。Tab 補完との親和性も高い

## Risks & Mitigations
- **Risk**: インターフェース変更により全モデル実装のコンパイルエラー → **Mitigation**: embedding で変更量を最小化、一括適用
- **Risk**: 非推奨フィルタリングによりユーザーが既存モデルを見失う → **Mitigation**: `--all` フラグの存在を help に明記、警告メッセージで `--all` を案内
- **Risk**: スキルリネームにより既存ユーザーの参照が切れる → **Mitigation**: CLAUDE.md の参照箇所を同時更新

## References
- Go embedding pattern: https://go.dev/doc/effective_go#embedding
- urfave/cli v3 flag handling: https://cli.urfave.org/v3/
