# Implementation Plan

## Task Overview

このドキュメントは `install-script` 機能の実装タスクを定義します。全ての要件をカバーし、論理的な順序で実装を進めます。

## Implementation Tasks

### Phase 1: インストールスクリプトのコア実装

- [x] 1. スクリプト基盤とロギングシステムの構築
- [x] 1.1 (P) スクリプトのセットアップとエラーハンドリング基盤を構築
  - `install.sh` ファイルを作成し、シェバング `#!/bin/bash` を設定
  - `set -euo pipefail` でエラーハンドリングを有効化
  - `trap 'cleanup' EXIT` で一時ファイルのクリーンアップを設定
  - グローバル変数 (OS, ARCH, VERSION, TEMP_DIR) を定義
  - _Requirements: 6_

- [x] 1.2 (P) ロギング関数の実装
  - `log_info()` 関数を実装（標準出力にメッセージを出力）
  - `log_error()` 関数を実装（標準エラー出力にメッセージを出力）
  - `log_success()` 関数を実装（成功メッセージを標準出力に出力）
  - メッセージプレフィックス `[INFO]`, `[ERROR]`, `[SUCCESS]` を追加
  - _Requirements: 4_

- [x] 2. プラットフォーム検出機能の実装 (P)
  - `detect_platform()` 関数を作成
  - OS を検出し、`Darwin` を `darwin`、`Linux` を `linux` に正規化
  - アーキテクチャを検出し、`x86_64` を `amd64`、`aarch64` と `arm64` を `arm64` に正規化
  - サポート対象: darwin/amd64, darwin/arm64, linux/amd64, linux/arm64
  - サポート外の場合はエラーメッセージとサポート対象一覧を表示し終了
  - グローバル変数 `OS` と `ARCH` に結果を格納
  - _Requirements: 1_

- [x] 3. GitHub Releases から最新バージョンを取得する機能の実装 (P)
  - `fetch_latest_version()` 関数を作成
  - GitHub API エンドポイント `https://api.github.com/repos/micheam/ai-assistant-console/releases/latest` に GET リクエスト
  - `curl -fsSL --max-time 10` でリクエストを実行
  - `jq` が利用可能な場合は `jq -r .tag_name` で `tag_name` を抽出
  - `jq` がない場合は `grep` と `sed` で `"tag_name": "v1.2.3"` からバージョン番号を抽出
  - API エラー時（HTTP 4xx/5xx）はステータスコードを表示し終了
  - レート制限（403）の場合は特別なメッセージを表示
  - グローバル変数 `VERSION` に結果を格納
  - _Requirements: 2_

- [x] 4. セキュアなファイルダウンロード機能の実装
- [x] 4.1 (P) ダウンロード関数の実装
  - `download_file()` 関数を作成（引数: `url`, `output`）
  - `curl -fsSL -o "$output" "$url"` でダウンロード
  - HTTP エラー時（4xx/5xx）は非ゼロ終了コード
  - エラー時は `curl -w "%{http_code}"` で HTTP ステータスコードを取得し表示
  - HTTPS のみを使用し、HTTP フォールバックは許可しない
  - _Requirements: 6_

- [x] 4.2 バイナリとチェックサムファイルのダウンロード
  - バイナリファイル URL を構築: `https://github.com/micheam/ai-assistant-console/releases/download/${VERSION}/aico_${OS}_${ARCH}.tar.gz`
  - チェックサムファイル URL を構築: `https://github.com/micheam/ai-assistant-console/releases/download/${VERSION}/aico_${OS}_${ARCH}.tar.gz.sha256`
  - 一時ディレクトリに両ファイルをダウンロード
  - ダウンロード失敗時はエラーメッセージを表示し終了
  - タスク 3 の完了後に実行（VERSION 変数が必要）
  - _Requirements: 1, 6_

- [x] 5. チェックサム検証機能の実装 (P)
  - `verify_checksum()` 関数を作成（引数: `binary_file`, `checksum_file`）
  - OS に応じて `shasum -a 256 -c`（macOS）または `sha256sum -c`（Linux）を使用
  - チェックサムコマンドが存在しない場合は警告を表示し、スキップオプションを提供
  - チェックサム一致時は正常終了（戻り値 0）
  - チェックサム不一致時はエラーメッセージを表示し終了（戻り値 1）
  - チェックサムファイルが取得できない場合（404）は警告を表示し、ユーザーに確認プロンプトを表示
  - ユーザー確認プロンプト: `read -p "Continue without verification? (y/N): " response`
  - _Requirements: 6_

- [x] 6. バイナリ展開とインストール機能の実装
- [x] 6.1 (P) tar.gz ファイルの展開
  - `extract_archive()` 関数を作成（引数: `archive_file`, `extract_dir`）
  - `tar -xzf "$archive_file" -C "$extract_dir"` で展開
  - 展開失敗時はエラーメッセージを表示し終了
  - アーカイブ内のファイル数を確認（`tar -tzf "$archive_file" | wc -l`）
  - 予期しない内容の場合は警告を表示
  - _Requirements: 3_

- [x] 6.2 (P) バイナリのインストールディレクトリへの配置
  - `install_binary()` 関数を作成（引数: `source_binary`）
  - インストールディレクトリ `$HOME/.local/bin` が存在しない場合は `mkdir -p` で作成
  - 既存バイナリが存在する場合は `aico --version` で現在のバージョンを取得し表示
  - `cp "$source_binary" "$HOME/.local/bin/aico"` でバイナリをコピー
  - `chmod +x "$HOME/.local/bin/aico"` で実行権限を付与
  - コピー失敗時（ディスク容量不足、権限エラー）は詳細なエラーメッセージを表示
  - _Requirements: 3, 5_

- [x] 7. メインフローの統合と完成
- [x] 7.1 メインフローの実装
  - スクリプト開始時に開始メッセージを表示（例: "Installing aico..."）
  - `detect_platform()` を呼び出し
  - `fetch_latest_version()` を呼び出し
  - 各ステップの進行状況を標準出力に表示
  - バイナリとチェックサムファイルをダウンロード
  - `verify_checksum()` を呼び出し
  - `extract_archive()` を呼び出し
  - `install_binary()` を呼び出し
  - 一時ファイルのクリーンアップ（`trap` による自動実行）
  - タスク 1〜6 の完了後に実行
  - _Requirements: 4_

- [x] 7.2 インストール完了メッセージの実装
  - インストール成功時にバージョンとパスを表示（例: "Successfully installed aico v1.2.3 to $HOME/.local/bin/aico"）
  - `$HOME/.local/bin` が PATH に含まれているか確認
  - PATH に含まれていない場合は、PATH 設定手順を表示
  - 例: "Add $HOME/.local/bin to your PATH by adding this line to your ~/.bashrc or ~/.zshrc:"
  - 例: "  export PATH=\"$HOME/.local/bin:$PATH\""
  - _Requirements: 4_

### Phase 2: リリースワークフローの拡張

- [x] 8. GitHub Actions リリースワークフローの更新
- [x] 8.1 tar.gz アーカイブ生成ステップの追加
  - `.github/workflows/release.yml` を編集
  - ビルドステップ後に各プラットフォームのバイナリを tar.gz 形式で圧縮
  - 例: `tar -czf aico_darwin_arm64.tar.gz aico-darwin-arm64`（各プラットフォームで実行）
  - バイナリ名を `aico` にリネームしてから圧縮
  - タスク 7 の完了後に実行（インストールスクリプトのテストに必要）
  - _Requirements: 6_

- [x] 8.2 SHA256 チェックサム生成ステップの追加
  - 各 tar.gz ファイルに対して `shasum -a 256 aico_{os}_{arch}.tar.gz > aico_{os}_{arch}.tar.gz.sha256` を実行
  - チェックサムファイルを GitHub Releases にアップロード
  - タスク 8.1 の完了後に実行
  - _Requirements: 6_

- [x] 8.3 リリース成果物に tar.gz とチェックサムファイルを追加
  - `softprops/action-gh-release@v1` の `files` セクションに tar.gz とチェックサムファイルを追加
  - 既存の生バイナリも維持（後方互換性のため）
  - タスク 8.2 の完了後に実行
  - _Requirements: 6_

### Phase 3: ドキュメント

- [x] 9. README にインストール手順を追加 (P)
  - README のインストールセクションを更新
  - ワンライナーインストール方法を記載: `curl -fsSL https://raw.githubusercontent.com/micheam/ai-assistant-console/main/install.sh | bash`
  - セキュリティ推奨事項として 2ステップインストール方法を記載
  - 例: `curl -fsSL https://raw.githubusercontent.com/micheam/ai-assistant-console/main/install.sh -o install.sh`
  - 例: `less install.sh`（内容確認）
  - 例: `bash install.sh`
  - PATH 設定が必要な場合の手順を記載
  - タスク 7 の完了後に実行可能（インストールスクリプトが完成している必要がある）
  - _Requirements: 6_

### Phase 4: テスト（オプション）

- [ ]* 10. インストールスクリプトのテスト実装
- [ ]* 10.1 単体テストの実装（bats-core 使用）
  - `detect_platform()` のテスト: 各 OS/アーキテクチャの組み合わせをモック
  - `verify_checksum()` のテスト: 正常ケース、不一致ケース、チェックサムファイル欠落ケース
  - `download_file()` のテスト: `curl` の成功/失敗をモック（ネットワークアクセスなし）
  - タスク 1〜6 の完了後に実行可能
  - _Requirements: 1, 6_

- [ ]* 10.2 統合テストの実装
  - 成功フロー: 最新版を実際にダウンロードしてインストール
  - チェックサム検証エラー: 不正なチェックサムファイルを提供し、エラーで終了することを確認
  - プラットフォーム検出: Docker コンテナで複数の OS/アーキテクチャをシミュレート
  - タスク 7 と 8 の完了後に実行可能（リリース成果物が必要）
  - _Requirements: 1, 6_

## Requirements Coverage

全ての要件がタスクでカバーされています:

- **Requirement 1**: タスク 2, 4.2, 10.1, 10.2
- **Requirement 2**: タスク 3
- **Requirement 3**: タスク 6.1, 6.2
- **Requirement 4**: タスク 1.2, 7.1, 7.2
- **Requirement 5**: タスク 6.2
- **Requirement 6**: タスク 1.1, 4.1, 4.2, 5, 8.1, 8.2, 8.3, 9, 10.1, 10.2

## Parallel Execution Notes

- **Phase 1** の以下のタスクは並行実行可能（`(P)` マーク付き）:
  - タスク 1.1, 1.2: スクリプト基盤（共有リソースなし）
  - タスク 2, 3, 5, 6.1, 6.2: 独立した関数の実装（相互依存なし）
  - タスク 4.1: ダウンロード関数の実装（他のタスクに依存しない）
- タスク 4.2 は VERSION 変数が必要なため、タスク 3 の完了後に実行
- タスク 7 は統合タスクのため、タスク 1〜6 の完了後に実行
- **Phase 2** のタスク 8 は Phase 1 の完了後に順次実行（8.1 → 8.2 → 8.3）
- **Phase 3** のタスク 9 は Phase 1 完了後に実行可能（Phase 2 とは並行実行可能）
- **Phase 4** のテストタスクはオプション（`*` マーク付き）で、MVP 後に実装可能
