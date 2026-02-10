# Research & Design Decisions

---
**Purpose**: Capture discovery findings, architectural investigations, and rationale that inform the technical design.

**Usage**:
- Log research activities and outcomes during the discovery phase.
- Document design decision trade-offs that are too detailed for `design.md`.
- Provide references and evidence for future audits or reuse.
---

## Summary
- **Feature**: `install-script`
- **Discovery Scope**: Simple Addition (new shell script with external API integration)
- **Key Findings**:
  - GitHub Releases API は `/repos/{owner}/{repo}/releases/latest` で最新リリース情報を JSON で提供
  - 現在のリリースワークフローは tar.gz アーカイブを生成しておらず、生バイナリのみをアップロード
  - チェックサム検証は現在実装されていないため、リリースワークフローの拡張が必要
  - `set -euo pipefail` によるエラーハンドリングと `curl -fsSL` によるセキュアなダウンロードがベストプラクティス

## Research Log

### GitHub Releases API 仕様
- **Context**: 最新リリース情報の取得方法とレスポンスフォーマットの確認
- **Sources Consulted**:
  - [GitHub REST API - Releases](https://docs.github.com/en/rest/releases/releases)
- **Findings**:
  - エンドポイント: `GET /repos/{owner}/{repo}/releases/latest`
  - レスポンスに `tag_name` (例: `v1.2.3`)、`assets[]` (ダウンロード URL を含む) が含まれる
  - レート制限: 認証なしで 60 req/hour、認証ありで 5000 req/hour
- **Implications**: スクリプトは認証なしで動作する必要があるため、レート制限を考慮した設計が必要

### 現在のリリースワークフロー分析
- **Context**: 既存の GitHub Actions リリース設定を確認し、必要な変更を特定
- **Sources Consulted**: `.github/workflows/release.yml`
- **Findings**:
  - 現在は生バイナリファイル (`aico-{os}-{arch}`) を直接アップロード
  - tar.gz アーカイブは生成されていない
  - チェックサムファイルは生成されていない
  - バイナリ名のパターンは `aico-{os}-{arch}` (ハイフン区切り)
- **Implications**:
  - リリースワークフローに tar.gz 圧縮ステップの追加が必要
  - SHA256 チェックサムファイル生成ステップの追加が必要
  - インストールスクリプトのバイナリ名パターンを `aico_{os}_{arch}.tar.gz` に合わせる

### Shell Script セキュリティベストプラクティス
- **Context**: `curl | sh` パターンのセキュリティリスクと緩和策の調査
- **Sources Consulted**:
  - [5 Ways to Deal With the install.sh Curl Pipe Bash problem](https://www.chef.io/blog/5-ways-to-deal-with-the-install-sh-curl-pipe-bash-problem)
  - [Best practices when using Curl in shell scripts](https://www.joyfulbikeshedding.com/blog/2020-05-11-best-practices-when-using-curl-in-shell-scripts.html)
  - [Writing Safe Shell Scripts – MIT SIPB](https://sipb.mit.edu/doc/safe-shell/)
- **Findings**:
  - **リスク**: 不完全なダウンロード時の部分実行、ブラウザ表示内容とダウンロード内容の不一致
  - **ベストプラクティス**:
    - `curl -fsSL`: `-f` (HTTP エラーで失敗)、`-s` (サイレント)、`-S` (エラー表示)、`-L` (リダイレクト追従)
    - `set -euo pipefail`: エラー時即座終了、未定義変数でエラー、パイプラインの途中失敗を検出
    - 全スクリプトコードを関数でラップして部分実行を防止
    - 2ステップインストール (ダウンロード → 確認 → 実行) を推奨
- **Implications**:
  - スクリプト冒頭で `set -euo pipefail` を設定
  - `curl -fsSL` オプションを統一使用
  - README にセキュリティ推奨事項を記載

### SHA256 チェックサム検証パターン
- **Context**: バイナリ整合性検証の実装方法
- **Sources Consulted**:
  - [Compare SHA256 checksum of a GitHub Release](https://gist.github.com/thanoskoutr/12afbd6b87d8c0126f344cfae75769e3)
  - [Shell script to validate SHA256 hashes](https://gist.github.com/onnimonni/b49779ebc96216771a6be3de46449fa1)
- **Findings**:
  - macOS: `shasum -a 256 -c <checksum_file>`
  - Linux: `sha256sum -c <checksum_file>`
  - チェックサムファイル形式: `<hash> <filename>` (例: `abc123... aico_darwin_arm64.tar.gz`)
  - コマンドが存在しない場合のフォールバック処理が必要
- **Implications**:
  - OS 検出と共にチェックサムコマンドの存在確認を実装
  - チェックサムファイルが取得できない場合の警告とユーザー確認フローを追加

### Bash シェバングと互換性
- **Context**: スクリプトの可搬性とシェル互換性の確認
- **Sources Consulted**:
  - [set -euo pipefail explanation](https://gist.github.com/mohanpedala/1e2ff5661761d3abd0385e8223e16425)
  - [Illegal option -o pipefail](https://www.baeldung.com/linux/illegal-option-o-pipefail)
- **Findings**:
  - `set -o pipefail` は bash 固有の機能で sh では使用不可
  - シェバングは `#!/bin/bash` を使用する必要がある
  - POSIX sh との互換性は犠牲になるが、エラーハンドリングの堅牢性を優先
- **Implications**: スクリプトは `#!/bin/bash` で開始し、bash の存在を前提とする

## Architecture Pattern Evaluation

シンプルな単一スクリプトのため、アーキテクチャパターンの評価は不要。以下の単純なステップフローで実装:

1. 環境チェック (OS/アーキテクチャ/コマンド存在確認)
2. GitHub API から最新バージョン情報取得
3. バイナリとチェックサムファイルのダウンロード
4. チェックサム検証
5. tar.gz 展開とインストール
6. クリーンアップ

## Design Decisions

### Decision: tar.gz アーカイブ形式の採用

- **Context**: 現在のリリースワークフローは生バイナリを直接アップロードしているが、チェックサム検証や将来的な複数ファイル配布を考慮
- **Alternatives Considered**:
  1. 生バイナリのまま配布 (現状維持)
  2. tar.gz アーカイブで配布
  3. zip アーカイブで配布
- **Selected Approach**: tar.gz アーカイブ形式
- **Rationale**:
  - Unix/Linux 環境での標準的な配布形式
  - チェックサムファイルと共にアップロードする際の一貫性
  - 将来的に設定ファイルやドキュメントを同梱する拡張性
  - macOS/Linux の両方で標準的に解凍可能
- **Trade-offs**:
  - リリースワークフローの修正が必要
  - アーカイブ展開ステップが追加されるため、インストール時間がわずかに増加
- **Follow-up**: リリースワークフローで圧縮率を確認 (バイナリサイズが小さい場合は圧縮効果が限定的)

### Decision: チェックサムファイル取得失敗時のフォールバック

- **Context**: 古いリリースや移行期間中のリリースでチェックサムファイルが存在しない可能性
- **Alternatives Considered**:
  1. チェックサムファイルが必須 (取得失敗で即座終了)
  2. チェックサムファイルがない場合は警告を表示し、ユーザー確認で続行
  3. チェックサムファイルがない場合は警告のみで自動続行
- **Selected Approach**: オプション 2 (警告 + ユーザー確認)
- **Rationale**:
  - セキュリティ重視のユーザーは確認段階で中断可能
  - 古いリリースやテスト環境でも動作可能
  - リリースワークフロー移行期間中の互換性を確保
- **Trade-offs**: インストールフローが対話的になるため、完全自動化スクリプトには不向き
- **Follow-up**: 環境変数 `AICO_SKIP_CHECKSUM=1` などでチェックサムスキップを許可するオプションを検討

### Decision: インストールディレクトリ `$HOME/.local/bin` の選択

- **Context**: ユーザー権限でインストール可能なディレクトリが必要
- **Alternatives Considered**:
  1. `/usr/local/bin` (システム全体、sudo 必要)
  2. `$HOME/.local/bin` (XDG Base Directory 準拠)
  3. `$HOME/bin` (伝統的なユーザーディレクトリ)
- **Selected Approach**: `$HOME/.local/bin`
- **Rationale**:
  - XDG Base Directory Specification 準拠
  - sudo 権限不要でインストール可能
  - 多くの Linux ディストリビューションでデフォルトの PATH に含まれる
  - macOS でも問題なく使用可能
- **Trade-offs**: ユーザーが PATH に追加していない場合は手動設定が必要
- **Follow-up**: インストール完了メッセージで PATH 設定手順を表示

## Risks & Mitigations

- **Risk 1**: GitHub API レート制限による取得失敗
  - **Mitigation**: エラーメッセージに HTTP ステータスコードを表示し、レート制限の可能性を示唆
- **Risk 2**: リリースワークフロー変更の影響
  - **Mitigation**: 既存リリースとの互換性を維持するため、バイナリ名パターンを慎重に設計
- **Risk 3**: ネットワーク不安定時の部分ダウンロード
  - **Mitigation**: `curl -f` でエラー時に非ゼロ終了コード、`set -e` で即座終了
- **Risk 4**: チェックサムコマンド未インストール環境
  - **Mitigation**: `shasum` / `sha256sum` の存在確認を実装し、どちらもない場合は警告とスキップオプション提供

## References

- [GitHub REST API - Releases](https://docs.github.com/en/rest/releases/releases)
- [5 Ways to Deal With the install.sh Curl Pipe Bash problem](https://www.chef.io/blog/5-ways-to-deal-with-the-install-sh-curl-pipe-bash-problem)
- [Best practices when using Curl in shell scripts](https://www.joyfulbikeshedding.com/blog/2020-05-11-best-practices-when-using-curl-in-shell-scripts.html)
- [Writing Safe Shell Scripts – MIT SIPB](https://sipb.mit.edu/doc/safe-shell/)
- [set -euo pipefail explanation](https://gist.github.com/mohanpedala/1e2ff5661761d3abd0385e8223e16425)
- [Compare SHA256 checksum of a GitHub Release](https://gist.github.com/thanoskoutr/12afbd6b87d8c0126f344cfae75769e3)
