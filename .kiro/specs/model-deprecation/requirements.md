# Requirements Document

## Introduction
プロバイダーごとにモデルを Deprecated としてマークし、モデル一覧の表示制御および削除予定バージョンの管理を行う機能。現状、非推奨情報は `Description()` メソッド内のテキストとしてのみ管理されており、構造化されたメタデータや一覧のフィルタリング機構は存在しない。本仕様により、非推奨モデルを型安全かつ統一的に管理可能とする。

## Project Description (Input)
providor 毎にモデルを Deprecated としてマークできるようにしたい。Deprecated としてマークされたモデルは models list からはデフォルトでは除外され、特別にフラグを付けた場合にのみ表示されるようにしたい。また、Deprecated としてマークする場合には、「どのバージョンで消されるか」を明示しておけるようにしたい。

## Requirements

### Requirement 1: モデルの非推奨メタデータ定義
**Objective:** As a 開発者, I want モデルに構造化された非推奨メタデータを持たせたい, so that 非推奨状態をプログラマティックに判定・管理できる。

#### Acceptance Criteria
1. The `ModelDescriptor` shall 非推奨状態を表す `Deprecated()` メソッドを提供し、`bool` 型の値を返す。
2. The `ModelDescriptor` shall 削除予定バージョンを返す `DeprecatedRemovedIn()` メソッドを提供し、`string` 型の値を返す（例: `"v2.0.0"`）。
3. While モデルが非推奨でない場合, the `ModelDescriptor` shall `Deprecated()` で `false` を、`DeprecatedRemovedIn()` で空文字列を返す。
4. While モデルが非推奨である場合, the `ModelDescriptor` shall `Deprecated()` で `true` を返し、`DeprecatedRemovedIn()` で削除予定バージョンの文字列を返す。

### Requirement 2: 既存モデルへの非推奨メタデータ適用
**Objective:** As a 開発者, I want 既存の非推奨モデルに構造化メタデータを適用したい, so that Description 内のテキストベースの管理から型安全な管理へ移行できる。

#### Acceptance Criteria
1. The Anthropic プロバイダー shall `claude-opus-4-5` を非推奨としてマークし、削除予定バージョンを設定する。
2. The OpenAI プロバイダー shall `gpt-4o`、`gpt-4o-mini`、`o1`、`o1-mini` を非推奨としてマークし、各モデルに削除予定バージョンを設定する。
3. The 各プロバイダー shall 非推奨でないモデルについて `Deprecated()` が `false` を返すことを保証する。

### Requirement 3: モデル一覧のフィルタリング
**Objective:** As a ユーザー, I want `models list` コマンドでデフォルトで非推奨モデルを非表示にしたい, so that 現行の推奨モデルのみを確認できる。

#### Acceptance Criteria
1. When `aico models list` を実行した場合, the CLI shall 非推奨としてマークされたモデルを一覧から除外する。
2. When `aico models list --all` を実行した場合, the CLI shall 非推奨モデルを含む全モデルを一覧に表示する。
3. While 非推奨モデルが `--all` フラグ付きで表示される場合, the CLI shall 非推奨であることを視覚的に区別できるように表示する（例: `[deprecated]` ラベル付与）。
4. When `aico models list --json` を実行した場合, the CLI shall JSON 出力にも非推奨フィルタリングをデフォルトで適用する。
5. When `aico models list --json --all` を実行した場合, the CLI shall JSON 出力に `deprecated` および `deprecated_removed_in` フィールドを含める。

### Requirement 4: モデル詳細表示での非推奨情報
**Objective:** As a ユーザー, I want `models describe` コマンドで非推奨情報を確認したい, so that 使用中のモデルの非推奨状況と移行先を把握できる。

#### Acceptance Criteria
1. When 非推奨モデルに対して `aico models describe MODEL` を実行した場合, the CLI shall 非推奨である旨と削除予定バージョンを表示する。
2. When 非推奨でないモデルに対して `aico models describe MODEL` を実行した場合, the CLI shall 非推奨関連の情報を表示しない。

### Requirement 5: 非推奨モデルの使用時の警告
**Objective:** As a ユーザー, I want 非推奨モデルを使用しようとした場合に警告を受けたい, so that 非推奨モデルの使用を意識的に判断できる。

#### Acceptance Criteria
1. When 非推奨モデルが設定ファイルまたはコマンドライン引数で指定された場合, the CLI shall 標準エラー出力に警告メッセージを表示する。
2. The 警告メッセージ shall モデル名と削除予定バージョンを含む。
3. While 非推奨モデルが警告された場合でも, the CLI shall モデルの使用をブロックせず、正常に処理を継続する。

### Requirement 6: モデル関連 Claude Code スキルの命名規則統一
**Objective:** As a 開発者, I want モデル関連のスキルを `/model-*` の命名規則で統一したい, so that スキル名から機能を直感的に把握でき、一貫性のある開発体験を得られる。

#### Acceptance Criteria
1. The 既存の `add-model` スキル shall `model-add` にリネームされる（`.claude/skills/add-model/` → `.claude/skills/model-add/`）。
2. The `model-add` スキル shall 既存の `add-model` と同等の機能を維持する。
3. The スキル名の変更 shall CLAUDE.md 等の参照箇所にも反映される。

### Requirement 7: モデル非推奨化 Claude Code スキル（`/model-deprecate`）
**Objective:** As a 開発者, I want `/model-deprecate provider:model-id` スキルを使ってモデルを非推奨としてマークしたい, so that 手動でコードを編集せずに一貫した手順でモデルを非推奨化できる。

#### Acceptance Criteria
1. The スキル shall `.claude/skills/model-deprecate/SKILL.md` に定義される。
2. When `/model-deprecate provider:model-id` を実行した場合, the スキル shall 指定されたプロバイダーの該当モデルに非推奨メタデータを設定する。
3. The スキル shall 引数として `provider:model-id` 形式のモデル指定を受け付ける（例: `anthropic:claude-opus-4-5`）。
4. If プロバイダーのみが指定された場合またはモデルIDのみが指定された場合, the スキル shall `model-add` スキルと同様のプロバイダー推定ロジックを適用する。
5. The スキル shall 削除予定バージョン（`removed-in`）をユーザーに確認してから設定する。
6. When 非推奨化が完了した場合, the スキル shall ビルドとテストを実行し、`models describe` の出力を表示して結果を確認する。
7. The スキル shall Conventional Commits 形式でコミットメッセージを生成する（例: `chore(api): deprecate model-name`）。
