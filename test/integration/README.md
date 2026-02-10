# AICO Integration Tests

このディレクトリには、AICO のインストールスクリプトの統合テストが含まれています。

## 概要

統合テストはコンテナ内でインストールスクリプトを実行し、以下を検証します：

1. インストールスクリプトが正常に実行される
2. バイナリが正しい場所にインストールされる
3. バイナリが実行可能である
4. 基本的なコマンド（`--version`, `--help`）が動作する

## 前提条件

### コンテナランタイム

以下のいずれかが必要です（自動検出されます）：

- **Apple Container** (macOS 推奨): https://github.com/apple/container
  - Apple Silicon Mac 専用
  - macOS 26 以上が必要
  - `brew install apple/apple/container` でインストール可能

- **Docker**: https://www.docker.com/
  - クロスプラットフォーム対応
  - Intel Mac や Linux でも動作

### その他の前提条件

- GitHub Releases に少なくとも1つのリリースが公開されている
- インターネット接続

## 実行方法

### 基本的な使い方

```bash
# プロジェクトルートから実行（自動検出）
./test/integration/run-integration-test.sh
```

スクリプトは自動的に利用可能なコンテナランタイムを検出します：
1. `container` コマンド（Apple Container）
2. `docker` コマンド（Docker）

### コンテナランタイムを明示的に指定

```bash
# Apple Container を強制使用
CONTAINER_RUNTIME=container ./test/integration/run-integration-test.sh

# Docker を強制使用
CONTAINER_RUNTIME=docker ./test/integration/run-integration-test.sh
```

## テスト環境

- **ベースイメージ**: Ubuntu 22.04
- **ユーザー**: 非 root ユーザー（testuser）
- **インストール先**: `$HOME/.local/bin/aico`

## テスト項目

1. **バイナリの存在確認**: `/home/testuser/.local/bin/aico` が存在する
2. **実行権限の確認**: バイナリに実行権限がある
3. **バージョン表示**: `aico --version` が正常に動作する
4. **ヘルプ表示**: `aico --help` が正常に動作する

## トラブルシューティング

### コンテナランタイムが見つからない

**エラー**: `No container runtime found`

**解決方法**:

```bash
# Apple Container をインストール（Apple Silicon Mac）
brew install apple/apple/container

# または Docker をインストール（全プラットフォーム）
brew install --cask docker  # macOS
# または https://www.docker.com/ からダウンロード
```

### Apple Container が動作しない

Apple Container は以下の要件があります：
- Apple Silicon Mac（M1/M2/M3）
- macOS 26 以上

要件を満たさない場合は Docker を使用してください。

### テストが失敗する

1. GitHub Releases にリリースが公開されているか確認
2. インターネット接続を確認
3. コンテナランタイムが正常に動作しているか確認

```bash
# Apple Container の場合
container system status

# Docker の場合
docker ps
```

### クリーンアップ

テストコンテナは自動的にクリーンアップされますが、手動でクリーンアップする場合：

**Apple Container の場合**:
```bash
# テストイメージを削除
container image delete aico-integration-test

# 残っているテストコンテナを削除
container list | grep aico-test | awk '{print $1}' | xargs -I {} container delete {}
```

**Docker の場合**:
```bash
# テストイメージを削除
docker rmi aico-integration-test

# 残っているテストコンテナを削除
docker ps -a | grep aico-test | awk '{print $1}' | xargs docker rm -f
```

## 制限事項

- 現在は Linux amd64 プラットフォームのみテスト（コンテナ内は Ubuntu 22.04）
- 実際の GitHub Releases を使用するため、リリース前はテストできない
- ネットワーク接続が必要
- Apple Container は Apple Silicon Mac + macOS 26 以上が必要

## 今後の拡張

- 複数プラットフォーム対応（arm64, Alpine Linux など）
- ローカルビルドを使用したテストモード
- GitHub Actions での自動実行
- チェックサム検証のテスト
- PATH 設定のテスト
