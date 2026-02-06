# Implementation Plan

- [x] 1. DeprecationInfo 構造体と ModelDescriptor インターフェースの拡張
- [x] 1.1 非推奨メタデータを保持する embeddable 構造体を定義する
  - `assistant` パッケージに、非推奨フラグと削除予定バージョンを保持する構造体を追加する
  - ゼロ値が「非推奨でない」状態を表すようにし、`Deprecated()` と `DeprecatedRemovedIn()` メソッドを実装する
  - 構造体のゼロ値テストと設定値テストを記述する
  - _Requirements: 1.1, 1.2, 1.3, 1.4_

- [x] 1.2 ModelDescriptor インターフェースに非推奨メソッドを追加する
  - `ModelDescriptor` インターフェースに `Deprecated()` と `DeprecatedRemovedIn()` の2メソッドを追加する
  - コンパイルエラーが発生することを確認し、次タスクで全モデルに対応する
  - _Requirements: 1.1, 1.2_

- [x] 2. 全プロバイダーのモデル構造体に DeprecationInfo を embed する
- [x] 2.1 (P) Anthropic プロバイダーのモデルに非推奨メタデータを適用する
  - 全 Anthropic モデル構造体に非推奨メタデータ構造体を embed する
  - `claude-opus-4-5` を非推奨としてマークし、削除予定バージョンを設定する
  - 非推奨でないモデル（`claude-opus-4-6`, `claude-sonnet-4-5`, `claude-haiku-4-5`）はゼロ値のまま
  - `AvailableModels()` が返すインスタンスの非推奨メタデータが正しいことをテストする
  - _Requirements: 2.1, 2.3_

- [x] 2.2 (P) OpenAI プロバイダーのモデルに非推奨メタデータを適用する
  - 全 OpenAI モデル構造体に非推奨メタデータ構造体を embed する
  - `gpt-4o`, `gpt-4o-mini`, `o1`, `o1-mini` を非推奨としてマークし、各モデルに削除予定バージョンを設定する
  - 非推奨でないモデル（`gpt-5.2`, `gpt-4.1`, `gpt-4.1-mini`, `o3`, `o4-mini`, `o3-mini`）はゼロ値のまま
  - _Requirements: 2.2, 2.3_

- [x] 2.3 (P) Groq プロバイダーのモデルに非推奨メタデータ構造体を embed する
  - 全 Groq モデル構造体に非推奨メタデータ構造体を embed する（全モデル非推奨でないためゼロ値）
  - _Requirements: 2.3_

- [x] 2.4 (P) Cerebras プロバイダーのモデルに非推奨メタデータ構造体を embed する
  - 全 Cerebras モデル構造体に非推奨メタデータ構造体を embed する（全モデル非推奨でないためゼロ値）
  - _Requirements: 2.3_

- [x] 3. models list コマンドに非推奨フィルタリングを追加する
- [x] 3.1 一覧表示のビューモデルに非推奨情報を追加し、`--all` フラグでフィルタリングを制御する
  - 一覧ビューに非推奨フラグと削除予定バージョンのフィールドを追加する
  - `--all` フラグを `models list` サブコマンドとデフォルトアクションの両方に追加する
  - デフォルトでは非推奨モデルを一覧から除外し、`--all` 指定時は全モデルを表示する
  - `--all` で表示される非推奨モデルには `[deprecated]` ラベルを付与する
  - JSON 出力にも同じフィルタリングルールを適用し、`--all` 時は `deprecated` / `deprecated_removed_in` フィールドを出力する
  - フィルタリング動作とラベル表示のテストを記述する
  - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5_

- [x] 4. models describe コマンドに非推奨情報表示を追加する
- [x] 4.1 モデル詳細表示に非推奨ステータスと削除予定バージョンを追加する
  - 非推奨モデルの場合、テキスト出力に `Deprecated:` と `Removed In:` 行を追加する
  - 非推奨でないモデルの場合、非推奨関連の行を表示しない
  - JSON 出力にも `deprecated` と `deprecated_removed_in` フィールドを追加する
  - _Requirements: 4.1, 4.2_

- [x] 5. 非推奨モデル使用時の警告を追加する
- [x] 5.1 モデル解決時に非推奨モデルへの警告メッセージを stderr に出力する
  - モデル解決関数内で、解決されたモデルが非推奨の場合に stderr へ警告メッセージを出力する
  - 警告メッセージにはモデル名と削除予定バージョンを含める
  - 警告出力後も処理を中断せず、モデルの利用を継続する
  - _Requirements: 5.1, 5.2, 5.3_

- [x] 6. Claude Code スキルの命名規則統一とリネーム
- [x] 6.1 (P) `add-model` スキルを `model-add` にリネームする
  - スキルのディレクトリを `.claude/skills/add-model/` から `.claude/skills/model-add/` に移動する
  - YAML front-matter の `name:` フィールドを `model-add` に更新する
  - スキル本文中のスキル名参照を更新する
  - 既存の機能が変わらないことを確認する
  - _Requirements: 6.1, 6.2_

- [x] 6.2 (P) CLAUDE.md 等の参照箇所を更新する
  - プロジェクト内で `add-model` を参照している箇所を検索し、`model-add` に置き換える
  - _Requirements: 6.3_

- [x] 7. model-deprecate スキルを作成する
- [x] 7.1 `/model-deprecate` スキルの SKILL.md を作成する
  - `.claude/skills/model-deprecate/SKILL.md` を新規作成する
  - `provider:model-id` 形式の引数パースと、プロバイダー推定ロジックを手順に含める
  - 削除予定バージョン（`removed-in`）をユーザーに確認する手順を含める
  - 対象モデルの非推奨メタデータ設定、Description の更新、doc comment の更新手順を記述する
  - ビルド・テスト実行と `models describe` での結果確認手順を含める
  - Conventional Commits 形式のコミット手順を含める
  - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5, 7.6, 7.7_

- [x] 8. ビルド検証と全体結合テスト
- [x] 8.1 全体のビルド成功と既存テストの通過を確認する
  - `go build ./...` と `go test ./...` が成功することを確認する
  - `go run ./cmd/aico models` でデフォルト出力に非推奨モデルが含まれないことを確認する
  - `go run ./cmd/aico models --all` で非推奨モデルが `[deprecated]` ラベル付きで表示されることを確認する
  - `go run ./cmd/aico models describe <deprecated-model>` で非推奨情報が表示されることを確認する
  - _Requirements: 1.1, 1.2, 2.1, 2.2, 3.1, 3.2, 3.3, 4.1, 5.1_
