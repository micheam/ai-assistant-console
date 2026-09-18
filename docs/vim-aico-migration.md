# vim-aico 側で検討してほしい変更

`aico` 本体の `--source` / `--context` 設計方針への追従（`feature/source-context-alignment` ブランチ）に伴い、
[vim-aico](https://github.com/micheam/vim-aico)（private リポジトリ）側で検討・対応してほしい変更点をまとめる。
このドキュメントは指示のみで、vim-aico 自体のコードは変更していない。

対象コミット時点の vim-aico: `autoload/ai_assistant.vim`（`c06d636`）。

## 1. 履歴描画のフィルタ正規表現を緩める（互換性の注意点）

`autoload/ai_assistant.vim:263` と `:293` は

```vim
if text->empty() || text =~# '^<source>'
```

で `<source>` ブロックを履歴表示・タイトル抽出から除外している。この正規表現は **属性なしの `<source>` にしかマッチしない**。

`aico` 側は既に `@file` 指定の source に `<source file="...">` を付けていた（旧実装から）ため、この不一致は今回の変更で新しく生まれたものではないが、
`label:@-` のようなラベル付き stdin source を使うと `<source name="...">` も同様にフィルタをすり抜けるようになる。

**推奨**: 正規表現を `Session.Messages` 側の Go 実装（`internal/assistant/session.go` の `sessionPreview` は `strings.HasPrefix(text, "<source")` で判定している）に合わせて、前方一致に緩める。

```diff
- if text->empty() || text =~# '^<source>'
+ if text->empty() || text =~# '^<source\>'
```

（Vim の `\>` は単語境界。`<source`, `<source file="...">`, `<source name="...">` のいずれにもマッチし、`<sourceXxx>` のような偶然の別タグとは区別される。）

該当箇所: `ai_assistant.vim:263`, `:293`。

## 2. `session show --json` に `source` フィールドが増えた

`aico session show --json` のレスポンスに、任意で次のフィールドが追加された（値がなければキー自体が省略される）。

```json
{
  "id": "...",
  "model": "...",
  "updated_at": "...",
  "source": { "file": "/path/to/file", "name": "buffer.go" },
  "messages": [...]
}
```

`source.file` と `source.name` はどちらか一方だけのこともある（inline source のみを使ったセッションでは `source` 自体が省略される）。

**活用案**（任意、今回は指示のみ）:

- `OpenSessionBuffer`（`ai_assistant.vim:235` 付近）でのバッファ名生成に、最初のユーザーメッセージのテキストから拾う代わりに `session.source.name` / `.file` を優先して使う。
- 将来的に「生成結果を source ファイルへ書き戻す」機能を作る場合、`session.source.file` が書き戻し先の候補になる。ただし書き戻し自体は `aico` 側・vim-aico 側ともに未実装であり、今回のスコープには含まれていない。

## 3. バッファ内容にラベルを付けたい場合は `label:@-` を使う

これまで vim-aico は `job_start` + `ch_sendraw` で stdin にバッファ全文を流し込み、`aico` 側の「stdin が非対話的なら暗黙に source として使う」フォールバックに乗っていた（`ai_assistant.vim:565-566`, `:631-635`）。この経路は**そのまま動作し続ける**（後方互換）。

新しく、`--source` / `--context` に `label:@-` という形式を渡すと、stdin の内容に人間可読なラベルを付けられるようになった（`<source name="label">` として埋め込まれる）。vim-aico から呼ぶコマンドに

```vim
cmd->add('--source')->add($'{bufname_label}:@-')
```

のように追加すれば、たとえば現在のバッファ名を `name` 属性として LLM に渡せる。これは任意の改善であり、既存の無指定 stdin 経路を置き換える必要はない。

**注意 1**: `label:` はラベルの直後が `@` で始まるときだけ認識される。ラベルにコロンや `@` を含めると誤動作するため、バッファ名をそのままラベルに使う場合はコロンを含まないことを確認するか、`fnamemodify()` 等でサニタイズすること。

**注意 2**: `--source 'label:@-'` を追加する場合は、従来どおり `job_start` + `ch_sendraw` で stdin にバッファ全文を流し込む経路と**組み合わせて**使うこと（`--source` 自体が `@-` で stdin を読む指定になるため、そのまま動作する）。一方で、stdin へバッファを流し込んだまま `--context @-` のように**別の** stdin 読み取りを追加すると、stdin は 1 回しか読めないため 2 番目の `@-` がエラーになるか、意図した内容が読めなくなる。stdin を要求する `@-` 指定は `--source` と `--context` を通じて呼び出し全体で 1 箇所だけにすること。

## 4. `--context` の値がカンマで分割されなくなった

`aico` 側で `DisableSliceFlagSeparator` を有効にしたため、`--context` に渡す値がカンマを含んでいても分割されなくなった（従来は `urfave/cli` のデフォルトでカンマ区切りだった）。vim-aico が将来 `--context` で選択範囲や `go doc` 相当の出力を渡す機能を追加する場合、カンマの有無を気にする必要がなくなる。

## 5. `--session <id> <prompt>` 系のフローで `--context` が効くようになった

これまで `--last` / `--session` / `session resume` で会話を再開する際、`--context` は無視されていた。今回の変更で `--context` は毎ターン解決されるようになったため、`ContinueConversation`（`ai_assistant.vim:396` 付近、`BuildCommand(['--session', session_id, prompt])` を呼ぶ箇所）から `--context` を追加で渡せば、フォローアップの質問に資料を添付できる。現状 vim-aico はこの経路で `--context` を渡していないため、対応は完全に任意。

---

## 優先度の目安

| # | 内容 | 緊急度 |
| --- | --- | --- |
| 1 | 履歴フィルタ正規表現の緩和 | should consider（ラベル付き source を使わない限り実害なし） |
| 2 | `session.source` の活用 | 任意 |
| 3 | `label:@-` の導入 | 任意 |
| 4 | カンマ分割の考慮不要化 | 情報共有のみ |
| 5 | resume 時の `--context` 対応 | 任意 |

いずれも `aico` 側の後方互換は維持されているため、vim-aico を一切変更しなくても動作は壊れない。
