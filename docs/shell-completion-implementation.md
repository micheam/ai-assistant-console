# urfave/cli v3 を使った動的な補完候補の実装方法

## 概要

このドキュメントでは、urfave/cli v3 を使用してシェル補完機能、特に動的な補完候補（実行時に決定される値）を実装する方法について説明します。

## urfave/cli v3 のシェル補完機能

### 基本的な仕組み

urfave/cli v3 は、アプリケーション自体を特殊なフラグ付きで実行することで、動的に補完候補を生成する仕組みを採用しています。これにより、以下のような利点があります：

- 実行時の状態に基づいた補完候補の提供
- 外部リソース（API、データベース、設定ファイルなど）からの動的な候補取得
- アプリケーションのロジックを活用した文脈依存の補完

### サポートされているシェル

- Bash
- Zsh
- Fish
- PowerShell

## 実装方法

### 1. EnableShellCompletion の設定

まず、ルートコマンドで `EnableShellCompletion: true` を設定します：

```go
app := &cli.Command{
    Name:                  "myapp",
    Usage:                 "My Application",
    EnableShellCompletion: true,
    // ... その他の設定
}
```

この設定により、`myapp completion bash` などの補完スクリプト生成コマンドが自動的に利用可能になります。

### 2. ShellComplete 関数の実装

各コマンドに `ShellComplete` フィールドを設定することで、そのコマンドの引数に対する補完候補を提供できます。

#### 基本的な実装パターン

```go
ShellComplete: func(ctx context.Context, cmd *cli.Command) {
    // 補完候補を改行区切りで標準出力に書き出す
    fmt.Fprintln(cmd.Root().Writer, "option1")
    fmt.Fprintln(cmd.Root().Writer, "option2")
    fmt.Fprintln(cmd.Root().Writer, "option3")
}
```

#### 動的な補完候補の例

```go
{
    Name:      "show",
    Usage:     "show model information",
    ArgsUsage: "MODEL",
    ShellComplete: func(ctx context.Context, cmd *cli.Command) {
        // 利用可能なモデル名を動的に取得
        for _, model := range allAvailableModels() {
            fmt.Fprintln(cmd.Root().Writer, model.Name())
        }
    },
    Action: runShowModelInfo,
}
```

### 3. 補完スクリプトの生成と有効化

#### Bash の場合

```bash
# 補完スクリプトを生成
myapp completion bash > /usr/local/etc/bash_completion.d/myapp

# またはホームディレクトリに配置
myapp completion bash > ~/.bash_completion.d/myapp

# シェルを再起動または source
source ~/.bash_completion.d/myapp
```

#### Zsh の場合

```bash
# 補完スクリプトを生成
myapp completion zsh > /usr/local/share/zsh/site-functions/_myapp

# または fpath に含まれるディレクトリに配置
myapp completion zsh > ~/.zsh/completion/_myapp

# .zshrc に以下を追加（必要な場合）
fpath=(~/.zsh/completion $fpath)
autoload -Uz compinit && compinit
```

#### Fish の場合

```bash
# 補完スクリプトを生成
myapp completion fish > ~/.config/fish/completions/myapp.fish
```

## 実装例：モデル名の補完

### 要件

`aico models show` コマンドで Tab キーを押したときに、利用可能なモデル名（claude-sonnet-4-5, claude-opus-4-5, gpt-4o など）を補完候補として表示する。

### 実装

```go
var CmdModels = &cli.Command{
    Name:  "models",
    Usage: "manage AI models",
    Commands: []*cli.Command{
        {
            Name:      "show",
            Usage:     "show model information",
            ArgsUsage: "MODEL",
            ShellComplete: func(ctx context.Context, cmd *cli.Command) {
                // すべての利用可能なモデル名を補完候補として出力
                for _, model := range allAvailableModels() {
                    fmt.Fprintln(cmd.Root().Writer, model.Name())
                }
            },
            Action: runShowModelInfo,
        },
    },
}

func allAvailableModels() []assistant.ModelDescriptor {
    return append(anthropic.AvailableModels(), openai.AvailableModels()...)
}
```

### 使用例

```bash
$ aico models show <Tab>
claude-sonnet-4-5  claude-opus-4-5  claude-haiku-4-5  gpt-4o  gpt-4o-mini  o1  o1-mini  o3-mini
```

## トラブルシューティング

### 補完が動作しない場合

1. **補完スクリプトが正しく source されているか確認**
   ```bash
   # Bash の場合
   echo $BASH_COMPLETION_COMPAT_DIR

   # Zsh の場合
   echo $fpath
   ```

2. **補完スクリプトを再生成**
   ```bash
   myapp completion bash > /path/to/completion/file
   source /path/to/completion/file
   ```

3. **シェルを再起動**
   ```bash
   exec $SHELL
   ```

### 補完候補が古い場合

補完候補は動的に生成されるため、アプリケーションを更新した後も補完スクリプトの再生成は不要です。ただし、補完スクリプトの生成ロジック自体が変更された場合は、補完スクリプトを再生成する必要があります。

## 制限事項

### フラグの補完

現在の urfave/cli v3（v3.3.3 時点）では、コマンドの引数に対する補完は `ShellComplete` で実装できますが、フラグの値に対する動的な補完はサポートされていません。

この機能は [Issue #1905](https://github.com/urfave/cli/issues/1905) で議論されており、将来のバージョンで実装される可能性があります。

### 文脈依存の補完

現在の `ShellComplete` 関数のシグネチャでは、すでに入力されている引数の情報にアクセスする方法が限られています。より高度な文脈依存の補完を実装する場合は、`cmd.Args()` を使用して現在の引数を取得する必要があります。

## 参考リンク

- [Shell Completions - urfave/cli](https://cli.urfave.org/v3/examples/completions/shell-completions/)
- [urfave/cli GitHub Repository](https://github.com/urfave/cli)
- [urfave/cli v3 Package Documentation](https://pkg.go.dev/github.com/urfave/cli/v3)
- [Issue #1905: More powerful shell completion](https://github.com/urfave/cli/issues/1905)
- [command_test.go - ShellComplete examples](https://github.com/urfave/cli/blob/main/command_test.go)

## まとめ

urfave/cli v3 を使った動的な補完候補の実装は、以下の3つのステップで実現できます：

1. **EnableShellCompletion を有効化** - ルートコマンドで設定
2. **ShellComplete 関数を実装** - 各コマンドで補完候補を出力
3. **補完スクリプトを生成・source** - `completion` サブコマンドを使用

この仕組みにより、実行時の状態に基づいた柔軟な補完機能を提供できます。
