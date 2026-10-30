# Vim でのコード補完(complete): 設計

> 改訂案(2026-10-29)。「aico コンセプトの再考」での 3 つの決定(FIM から complete への改名、追加文脈を初版に含める、エージェント化の上限)を反映した。コードの改名はこの改訂案のレビュー後に行う。改訂前との差分は末尾の「改訂前からの変更」にまとめた。

ローカルの llama-server で動く FIM(Fill-in-the-Middle)モデルを使い、Vim のバッファにインライン補完(ゴーストテキスト)を表示する。目的はスループットの最大化で、推論サーバには M5 Max の Mac Studio を使う前提である。モデルと推論サーバの性能検証は別の PoC 計画で扱い、本書はエディタから推論サーバまでの構成だけを扱う。

## aico における位置づけ

aico と vim-aico が提供するのは、自律的に探索し実行するエージェントではなく、利用者の作業を底上げする 2 つの機能である。

- **ask**: 作用対象(`--source`)、参照資料(`--context`)、立場(persona)を明示して問い、結果を受け取る。現在の `aico` の既定動作と `:Assistant` がこれにあたる。
- **complete**: 開いているバッファと、明示した文脈から求まるコードを高速に提案する。文脈を広くとった snippet エンジンであり、本書の対象である。

ask と complete は意図の名前であり、generate や FIM は仕組みの名前である。利用者向けの面(コマンド名、設定のキー、RPC のメソッド名)には意図の名前を出し、仕組みの名前はプロバイダの内側に閉じる。

エージェント的な振る舞いの上限は、ask の `propose_edit` と `:AssistantApply`(提案を表示し、人が選んで適用する)に定める。complete でゴーストテキストを人が受け入れるのと同じ水準である。ツールの結果をモデルに返して続きを生成する、モデルが自分でファイルを探索する、といった仕組みは作らない。

## 作らないもの

次のものは作らない。

- ask(generate)のデーモン経由化。ask は従来どおり 1 回のプロセス起動で完結させる。
- `aico serve` への complete 以外のメソッドの追加。

次のものは初版では作らない。

- 補完結果のキャッシュ。入れる場合、回し直しはキャッシュを素通りする必要がある
- 追加文脈の大きさの最適化(断片への分割、大きさの上限、バッファ単位の除外)
- stdio での接続、launchd との連携、アイドル時の自動終了

## 構成

```
Vim (vim-aico) ─┐
                ├─ Unix ソケット / JSON-RPC ─→ aico serve ─ HTTP ─→ llama-server /infill
Vim (別の起動) ─┘
```

- **aico serve**: aico と同じバイナリのサブコマンドで、Unix ドメインソケットで待ち受ける常駐デーモン。プロファイルの解決、推論サーバへの要求、要求の取り消しを担う。
- **vim-aico**: デーモンへの接続と起動、補完のトリガ、文脈の収集、ゴーストテキストの表示と受け入れを担う。
- **llama-server**: llama.cpp に付属する推論サーバ。デーモンからは HTTP で呼ぶ。

## 決定事項と理由

1. **常駐デーモンと Unix ソケットにする。** 複数の Vim が同じ llama-server を使うとき、要求の順序と取り消しを 1 か所で管理するためである。プロセス起動の時間は理由ではない。
2. **aico と同じバイナリで提供する。** install.sh、GitHub Releases、vim-aico の `g:ai_assistant_bin` という配布単位を増やさないためである。コードはデーモン用の internal パッケージに分け、generate 側のコードに依存させない。
3. **デーモンの範囲は complete に限る。** 名前は `aico serve` のままとするが、メソッドは complete のものだけを持つ。ask をデーモンに載せる動機はツールの結果を同じターンでモデルに返すことにあり、それは上記の上限の外にある。
4. **推論エンジンは llama.cpp を前提とするが、プロトコルとプロバイダのインターフェースは FIM 形のバックエンドだけを前提にしない。** 要求は prefix と suffix と追加文脈の形で受け、`/infill` への対応付けは llamacpp プロバイダの中に閉じる。`/v1/completions` やテンプレートを噛ませた chat モデルを裏側にする余地を残すが、初版で実装するのは llamacpp だけである。
5. **接続先、モデル、時間予算はプロファイルとして config.toml に置く。** complete のモデルは ask のモデルとは別物なので、`aico models` の一覧には混ぜない。
6. **デーモンは 1 つとし、プロファイルは接続ごとに initialize で選ぶ。** デーモンの中の状態はプロファイル別に持つ。
7. **実行中の補完要求は、プロファイルごとに 1 つに限る。** 新しい要求が来たら、どの接続からのものであっても古い要求を取り消す。利用者は 1 人で、補完が必要なのは入力中の Vim だけだからである。
8. **Vim とデーモンの間の認証は、ソケットのファイル権限だけにする。** llama-server の API キーはプロファイル側で扱う。
9. **追加文脈は初版から送り、その出所は Vim で開いているバッファとする。** 文脈を 1 つのバッファに限ると snippet エンジンとしての文脈が狭く、llama.vim の代わりにこの構成を作る意味がなくなるためである。開いているバッファは利用者が意図して開いたものなので、明示した参照として扱える。内容は丸ごと送り、大きさの最適化は別途行う。

## 命名

| 対象 | 名前 |
| --- | --- |
| サブコマンド | `aico complete profiles [--json]` |
| 設定の節 | `[complete]`、`[complete.profiles.<名前>]` |
| RPC メソッド | `complete` |
| Go パッケージ | `internal/complete` |
| 設定の型 | `config.CompleteConfig`、`config.CompleteProfile` |
| プロバイダ | `internal/providers/llamacpp`(変更なし) |

補足:

- RPC メソッドは平坦な `complete` とする。`aico serve` のメソッドを complete に限ると決めたので名前空間は要らず、`complete.complete` は冗長になる。
- urfave/cli がシェル補完スクリプトのために生成する `aico completion` と並ぶ。この `completion` はヘルプに表示されない隠しコマンドで、名前の一致は前方一致ではなく完全一致で判定されるので、動作上の衝突はない。両方を残す。紛らわしさが問題になれば、urfave/cli の設定でシェル補完側の名前を変えられる。
- FIM という語は llamacpp プロバイダの中と、PoC 計画にだけ残る。

## 設定

```toml
[complete]
default_profile = "local"

[complete.profiles.local]
endpoint = "http://127.0.0.1:8012/infill"
model = "qwen2.5-coder-7b"
```

| キー | 既定値 |
| --- | --- |
| endpoint | なし(必須) |
| model | 空 |
| api_key_env | 空 |
| n_predict | 128 |
| t_max_prompt_ms | 500 |
| t_max_predict_ms | 1000 |
| prefix_lines | 256 |
| suffix_lines | 64 |
| candidates | 1 |

補足:

- `endpoint` は `/infill` までを含む URL で書く。プロキシ経由でパスが変わる場合にも対応できるようにするためである。
- `model` は要求の `model` にそのまま載せる。単一モデルで起動した llama-server がこの値をどう扱うかは PoC で確認する。
- `api_key_env` には API キーそのものではなく、キーを入れた環境変数の名前を書く。
- `prefix_lines` と `suffix_lines` は Vim 側が現在のバッファから文脈を集めるときの行数で、initialize の応答で Vim に渡す。デーモンはこの上限を超えた要求を切り詰めず、そのまま推論サーバに送る。
- `candidates` は 1 回の要求で生成する候補の数で、要求の `count` が省略されたときに使う。候補が揃うまで応答が返らないので、増やすと最初の表示までの時間が生成の分だけ延びる。既定は 1 にし、実測を見て決める。
- 既定値は llama.vim の既定値に合わせている。
- `default_profile` を省略した場合、プロファイルが 1 つだけならそれを使う。

`aico complete profiles` でプロファイルの一覧を表示する。`--json` を付けると vim-aico 向けの JSON を出力する。

## 文脈の収集

補完の要求に載せる文脈は 2 種類ある。収集はすべて Vim 側で行い、デーモンは受け取ったものをそのまま推論サーバに渡す。

| 種類 | 出所 | 要求の項目 |
| --- | --- | --- |
| 前後のテキスト | 現在のバッファ | prefix、suffix |
| 参照 | 同じ Vim で開いている他のバッファ | extra |

補足:

- 参照にするバッファは、バッファ一覧に載っていて(listed)、読み込み済みで、`buftype` が空で、名前があり、現在のバッファではないものとする。名前のないバッファは、推論サーバに渡すファイル名がないので送らない。
- 各バッファは `extra` の 1 要素として内容を丸ごと送る。`file` には Vim の作業ディレクトリからの相対パスを入れる。
- `extra` の順序はバッファ番号の昇順とする。要求ごとに順序が変わらないので、推論サーバのプロンプトキャッシュが前方一致で当たりやすい。
- バッファ単位で参照から外す仕組み、大きさの上限、断片への分割は初版では作らない。

PoC の Phase 2 は llama.vim で品質を評価する。llama.vim は既定で追加文脈を送るので、vim-aico 初版が追加文脈を送れば品質の数字は転用しやすい。ただし集め方が違うので、Phase 2 の日次メモには「どの文脈が効いたか」を残す。

## ソケットの場所と寿命管理

ソケットのパスは次の順で決める。`aico serve path` で、解決したパスを表示できる。`aico serve --socket <path>` を渡すと、この順序を使わずに指定したパスで待ち受ける。

1. `$XDG_RUNTIME_DIR/com.micheam.aico/serve.sock`
2. `$TMPDIR/com.micheam.aico/serve.sock`(macOS には `XDG_RUNTIME_DIR` がない)
3. `/tmp/com.micheam.aico-<uid>/serve.sock`

ディレクトリの権限は 0700 にする。macOS ではソケットのパスに 104 バイトの上限がある。`$TMPDIR` を使った場合は 80 バイト前後に収まる。

デーモンの起動手順は次のとおり。

1. 同じディレクトリの `serve.lock` に排他ロックをかける。ロックが取れなければ、別のデーモンが動いているとみなして正常終了する。
2. ロックを持ったまま、古いソケットファイルがあれば消してから待ち受ける。
3. ロックはデーモンの終了まで持ち続ける。

ロックを持つプロセスだけがソケットを作るので、2 つの Vim が同時にデーモンを起動しても競合しない。

vim-aico 側の手順は次のとおり。

1. ソケットに接続する。失敗したら `aico serve` を起動し、短い間隔で接続を再試行する。
2. デーモンを起動するときは、Vim の終了時にジョブを止めない設定にし、標準入出力を使わない。デーモンは SIGHUP を無視する。
3. 接続できたら initialize を送る。プロトコルの版が合わなければ shutdown を送り、デーモンを起動し直す。
4. 接続が切れたら、次の補完要求の時点で接続し直す。

デーモンのログは aico のログファイルにだけ書き、標準出力と標準エラーは使わない。

## プロトコル(版 1)

### フレーミング

LSP と同じく、`Content-Length: <バイト数>\r\n\r\n` のヘッダに続けて JSON の本文を送る。本文は JSON-RPC 2.0 のメッセージである。Vim は `ch_open()` に `mode: 'lsp'` を渡せばこの形式をそのまま扱える。

Vim の lsp モードの挙動として、次の点を確認している。

- `ch_sendexpr()` は `"jsonrpc": "2.0"` を自動で付ける。
- コールバックを渡した要求には、1 から始まる整数の `id` を Vim が振る。要求に書いた `id` は上書きされる。
- コールバックを渡さないメッセージは、`id` を付けずにそのまま送られる。

### initialize

接続して最初に送る要求。これより前に他の要求を送ると、エラー -32002 になる。例外は shutdown で、initialize なしでも受け付ける。プロトコルの版が合わずに initialize が失敗した接続から、古いデーモンを止めるためである。

```json
{"method": "initialize", "params": {
  "protocol_version": 1,
  "client": {"name": "vim-aico", "version": "0.1.0"},
  "profile": "local"
}}
```

`profile` を省略すると既定のプロファイルを使う。デーモンは initialize のたびに設定ファイルを読み直すので、プロファイルの変更は接続し直せば反映される(デーモンの再起動は要らない)。応答は次の形になる。

```json
{"result": {
  "protocol_version": 1,
  "server": {"name": "aico", "version": "v20261009.01"},
  "methods": ["complete"],
  "profile": {
    "name": "local",
    "model": "qwen2.5-coder-7b",
    "endpoint": "http://127.0.0.1:8012/infill",
    "prefix_lines": 256,
    "suffix_lines": 64,
    "candidates": 1
  }
}}
```

版が一致しなければエラー 1001 を返し、`data.protocol_version` にデーモン側の版を入れる。1 つの接続で initialize を 2 回送るとエラー -32600 になる。プロファイルを変えるときは接続し直す。

### complete

```json
{"method": "complete", "params": {
  "prefix": "func add(a, b int) int {\n\treturn ",
  "suffix": "\n}\n",
  "indent": 1,
  "extra": [{"file": "util.go", "text": "..."}],
  "count": 1
}}
```

- `prefix` はカーソルより前のテキストで、現在行のカーソルまでを含む。
- `suffix` はカーソルより後のテキストで、現在行のカーソル以降を含む。
- `indent` は現在行の字下げの文字数で、省略できる。
- `extra` は他のファイルから集めた参考の断片で、省略できる。`file` はファイル名で、推論サーバに渡される。
- `count` は生成する候補の数で、省略するとプロファイルの `candidates` を使う。回し直し(同じ文脈でもう一度要求する)は、この要求をもう一度送るだけでよい。候補の多様性は推論サーバのサンプリングから出るので、seed は扱わない。

応答は次の形になる。

```json
{"result": {
  "candidates": [
    {"text": "a + b", "truncated": false}
  ],
  "timings": {
    "prompt_tokens": 12, "prompt_ms": 35.2,
    "predicted_tokens": 4, "predicted_ms": 48.0,
    "cached_tokens": 210
  }
}}
```

同じプロファイルで新しい `complete` が届くと、実行中の要求は取り消され、エラー -32800 で応答する。推論サーバへの要求に失敗した場合はエラー 1003 を返す。

### $/cancelRequest

要求を取り消す通知。LSP と同じく、取り消す要求の `id` を `params` に入れる。

```json
{"method": "$/cancelRequest", "params": {"id": 3}}
```

取り消された要求にはエラー -32800 で応答する。すでに応答済みの `id` は無視する。Vim のヘルプの例は `id` を最上位に書いているが、その形で送るとコールバックなしでも `id` がそのまま残り、元の要求と同じ `id` を持つ要求として届く。そのため vim-aico は `params.id` の形で送る。

### shutdown

デーモンを止める要求。応答を返したあと、すべての接続を閉じ、ソケットとロックを解放して終了する。デーモンは複数の Vim で共有しているので、プロトコルの版が合わないときにだけ使う。

### エラーコード

| コード | 意味 |
| --- | --- |
| -32700 | JSON の解析に失敗 |
| -32600 | 不正な要求 |
| -32601 | 未知のメソッド |
| -32602 | 不正なパラメータ |
| -32603 | 内部エラー |
| -32002 | 未初期化 |
| -32800 | 取り消し |
| 1001 | プロトコルの版の不一致 |
| 1002 | 未知のプロファイル |
| 1003 | 推論サーバのエラー |

## llama-server への対応付け

llamacpp プロバイダは、`complete` を次のように `/infill` の要求に変換する。

| complete | /infill |
| --- | --- |
| prefix の最後の改行まで | input_prefix |
| prefix の最後の改行より後 | prompt |
| suffix | input_suffix |
| extra | input_extra |
| indent | n_indent |
| count | n_cmpl |
| プロファイルの n_predict | n_predict |
| プロファイルの t_max_prompt_ms | t_max_prompt_ms |
| プロファイルの t_max_predict_ms | t_max_predict_ms |
| プロファイルの model | model |

補足:

- `cache_prompt` は常に true、`stream` は常に false で送る。
- サンプラーは llama.vim と同じ設定(top_k 40、top_p 0.90、最後に infill)で送る。PoC の Phase 2 では llama.vim で品質を評価するので、その結果をそのまま当てはめられるようにするためである。
- `api_key_env` が指す環境変数に値があれば、`Authorization: Bearer` ヘッダで送る。
- 応答は `n_cmpl` が 1 ならオブジェクト、2 以上なら配列で返る。各要素の `content` と `truncated` を `candidates` の要素に移す。`timings` と `tokens_cached` は先頭の要素から取る(全候補が同じプロンプトを共有するため)。llama-server の README は 1 プロンプトで `n_cmpl` が 2 以上のときの応答の形を明記しておらず、llama.vim の実装に合わせている。PoC で実物を確かめる。
- `id_slot` は送らない。llama-server はプロンプトの類似度でスロットを選ぶ。スロットを固定すべきかは実測で判断する。

## テスト

- **vim-aico の回帰確認**: vim-aico の `test/complete_test.sh` が、偽の `/infill` サーバと本物の `aico serve` を使って、表示、候補の切り替え、受け入れ、他バッファの文脈送信を headless の Vim で確かめる。
- **ユニットテスト**: `go test ./...` で実行する。llamacpp プロバイダは `httptest` の偽サーバを相手に、送る要求の本文と応答の変換を確かめる。
- **インテグレーションテスト**: `//go:build integration` を付け、実際の llama-server を相手に実行する。環境変数 `AICO_TEST_LLAMACPP_ENDPOINT` に `/infill` の URL を入れたときだけ動き、未設定ならスキップする。

```
AICO_TEST_LLAMACPP_ENDPOINT=http://127.0.0.1:8012/infill \
  go test -tags=integration ./internal/providers/llamacpp/...
```

Makefile の `test` は `-tags e2e` で実行しており、このインテグレーションテストは含まれない。

## 未決事項

- [ ] 開いているバッファのどれかを編集するたびに、そのバッファの `extra` が変わり、その分の prefill がやり直しになる。PoC Phase 1 の「周辺コンテキスト大」の計測で、この影響の大きさを見る
- [ ] `id_slot` を固定するかどうか
- [ ] プロトコルの版は変わらずに aico だけを更新したとき、古いデーモンをどう入れ替えるか
- [ ] プロトコルの版が異なる vim-aico を同時に使うと、互いにデーモンを再起動し合う。これを許容するか
- [ ] PoC の Phase 1 で確かめる 4 点(LiteLLM Proxy の中継、`model` の扱い、`/v1/completions` の項目、切断時の生成停止)

## 改訂前からの変更

- 複数候補に対応した。要求に `count`、応答に `candidates`(常に配列)、プロファイルに `candidates` を足した。回し直しと候補の切り替えは Vim 側の操作で、デーモンには新しい要求として届く。
- 「aico における位置づけ」の節を足し、ask と complete の区別とエージェント化の上限を書いた。
- 利用者向けの名前を FIM から complete に改めた。「命名」の節に対応表を置いた。RPC メソッド名は平坦な `complete` とした。
- `generate.*` の予約と、ask のデーモン経由化の構想を取り下げた。「作らないもの」の節に移した。
- 追加文脈を「初版では作らない」から外し、「文脈の収集」の節で集め方を定めた。出所は同じ Vim で開いている他のバッファとし、内容を丸ごと送る。大きさの最適化は初版では行わない。
- 決定事項 4 を、FIM 形のバックエンドだけを前提にしない表現に改めた。
- 未決事項から、デーモンとプロファイルの関係を外した(確定)。追加文脈の編集によるキャッシュへの影響を足した。
