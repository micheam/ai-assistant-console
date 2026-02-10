# Requirements Document

## Project Description (Input)
pre-build バイナリをインストールするスクリプトを用意したい

## Introduction

AICO のプリビルドバイナリを簡単にインストールするためのシェルスクリプトを提供します。ユーザーは `curl | sh` パターンで、最新リリースのバイナリを自動的にダウンロードし、システムにインストールできます。GitHub Releases から適切なプラットフォーム向けのバイナリを取得し、`$HOME/.local/bin` に配置します。

## Requirements

### Requirement 1: プラットフォーム検出とバイナリダウンロード
**Objective:** As a ユーザー, I want 自分の OS/アーキテクチャに対応したバイナリを自動的に取得したい, so that 手動でプラットフォームを選択する必要がない

#### Acceptance Criteria
1. When インストールスクリプトが実行された時, the インストールスクリプト shall OS (`uname -s`) とアーキテクチャ (`uname -m`) を検出する
2. When 検出されたプラットフォームが Darwin (macOS) かつ arm64 の場合, the インストールスクリプト shall `aico_darwin_arm64.tar.gz` をダウンロードする
3. When 検出されたプラットフォームが Darwin (macOS) かつ x86_64 の場合, the インストールスクリプト shall `aico_darwin_amd64.tar.gz` をダウンロードする
4. When 検出されたプラットフォームが Linux かつ x86_64 の場合, the インストールスクリプト shall `aico_linux_amd64.tar.gz` をダウンロードする
5. When 検出されたプラットフォームが Linux かつ arm64 の場合, the インストールスクリプト shall `aico_linux_arm64.tar.gz` をダウンロードする
6. If サポート外の OS またはアーキテクチャが検出された場合, then the インストールスクリプト shall エラーメッセージを表示し終了する

### Requirement 2: GitHub Releases からの最新バージョン取得
**Objective:** As a ユーザー, I want 常に最新の安定版バイナリを取得したい, so that 最新機能とバグフィックスが利用できる

#### Acceptance Criteria
1. The インストールスクリプト shall GitHub API (`https://api.github.com/repos/micheam/ai-assistant-console/releases/latest`) から最新リリース情報を取得する
2. When API レスポンスが正常に取得できた時, the インストールスクリプト shall `tag_name` フィールドからバージョン番号を抽出する
3. When バージョン番号が取得できた時, the インストールスクリプト shall `https://github.com/micheam/ai-assistant-console/releases/download/{version}/aico_{platform}_{arch}.tar.gz` 形式で URL を構築する
4. If GitHub API へのアクセスが失敗した場合, then the インストールスクリプト shall エラーメッセージを表示し終了する

### Requirement 3: バイナリの展開とインストール
**Objective:** As a ユーザー, I want ダウンロードしたバイナリを適切なディレクトリに配置したい, so that PATH 環境変数からコマンドを実行できる

#### Acceptance Criteria
1. When tar.gz ファイルがダウンロードされた時, the インストールスクリプト shall 一時ディレクトリに展開する
2. When バイナリが展開された時, the インストールスクリプト shall `$HOME/.local/bin` ディレクトリの存在を確認する
3. If `$HOME/.local/bin` が存在しない場合, then the インストールスクリプト shall ディレクトリを作成する
4. When インストール先ディレクトリが準備できた時, the インストールスクリプト shall `aico` バイナリを `$HOME/.local/bin/aico` にコピーする
5. When バイナリがコピーされた時, the インストールスクリプト shall 実行権限 (`chmod +x`) を付与する
6. The インストールスクリプト shall インストール完了後に一時ファイルをクリーンアップする

### Requirement 4: エラーハンドリングと情報表示
**Objective:** As a ユーザー, I want インストールプロセスの進行状況とエラーを把握したい, so that 問題が発生した際に適切に対処できる

#### Acceptance Criteria
1. When スクリプトが開始された時, the インストールスクリプト shall "Installing aico..." のような開始メッセージを表示する
2. When 各主要ステップが完了した時, the インストールスクリプト shall 進行状況を標準出力に表示する
3. If いずれかのコマンドが失敗した場合, then the インストールスクリプト shall エラーメッセージを標準エラー出力に表示する
4. If ダウンロードに失敗した場合, then the インストールスクリプト shall HTTP ステータスコードまたはネットワークエラーの詳細を表示する
5. When インストールが正常に完了した時, the インストールスクリプト shall インストールされたバージョンとパスを表示する
6. When インストールが完了した時, the インストールスクリプト shall PATH 環境変数への追加が必要な場合は、その手順を表示する

### Requirement 5: 既存インストールの上書き
**Objective:** As a ユーザー, I want 既存のインストールを簡単に更新したい, so that 再インストール時に手動で削除する必要がない

#### Acceptance Criteria
1. When `$HOME/.local/bin/aico` が既に存在する場合, the インストールスクリプト shall 既存のバイナリを上書きする
2. When 上書きが実行される時, the インストールスクリプト shall 既存バージョンが上書きされることを通知する
3. If 既存バイナリが別プロセスで使用中の場合, then the インストールスクリプト shall 上書きを試み、失敗時にエラーメッセージを表示する

### Requirement 6: セキュリティと検証
**Objective:** As a セキュリティ重視のユーザー, I want インストールスクリプトとバイナリの安全性を検証したい, so that 悪意あるコードの実行を防げる

#### Acceptance Criteria
1. The インストールスクリプト shall HTTPS のみを使用してファイルをダウンロードする（HTTP フォールバックを許可しない）
2. The インストールスクリプト shall スクリプト冒頭で `set -euo pipefail` を設定し、エラー時に即座に終了する
3. When バイナリをダウンロードした時, the インストールスクリプト shall 対応する SHA256 チェックサムファイル (`aico_{platform}_{arch}.tar.gz.sha256`) もダウンロードする
4. When チェックサムファイルが取得できた時, the インストールスクリプト shall `shasum -a 256 -c` または `sha256sum -c` を使用してバイナリの整合性を検証する
5. If チェックサム検証が失敗した場合, then the インストールスクリプト shall エラーメッセージを標準エラー出力に表示し、非ゼロの終了コードで終了する
6. If チェックサムファイルが取得できない場合, then the インストールスクリプト shall 警告メッセージを表示し、ユーザーに検証なしで続行するかを確認する
7. When ダウンロードコマンドを実行する時, the インストールスクリプト shall `curl -fsSL` オプションを使用する（`-f`: HTTP エラーで失敗, `-s`: サイレント, `-S`: エラー表示, `-L`: リダイレクト追従）
8. The README shall 2ステップインストール方法（スクリプトをダウンロード後に内容確認してから実行）をセキュリティ推奨事項として記載する

