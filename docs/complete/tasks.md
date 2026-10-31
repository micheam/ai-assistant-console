# Vim でのコード補完(complete): タスク

設計は [design.md](design.md) を参照。作業ブランチは aico が `feature/fim-serve`、vim-aico が `feature/fim-completion`。

最初に 1 から 5 の最小形を縦に 1 本通す。最小形では、補完を 1 件取得し、ゴーストテキストとして表示し、全体を受け入れるところまでを作る。

## aico

- [x] 1. プロファイル設定: config.toml の `[complete]` 節と `aico complete profiles [--json]`
- [x] 2. llamacpp プロバイダ: エンジンに依存しない complete のインターフェースと `/infill` のクライアント。ユニットテストと integration タグのテスト
- [x] 3. `aico serve`: 待ち受け、フレーミング、initialize、shutdown、`complete`、`$/cancelRequest`、新しい要求の優先、ロックと古いソケットの処理、`aico serve path`

## vim-aico

- [x] 4. 接続管理: 補完用の autoload ファイル、接続と起動、initialize での版の照合、再接続
- [x] 5. 補完本体: `CursorMovedI` とデバウンスによるトリガ、キー入力での取り消し、文脈の収集(現在のバッファの前後と、開いている他のバッファの丸ごと)、ゴーストテキスト、受け入れ(全体、1 行、1 単語)、回し直し、候補の切り替え、filetype による有効化、timings の表示
- [x] 6. `:AssistantCompleteProfile`
- [x] 7. README の追記

## 検証

- [ ] 8. PoC の Phase 1 で確かめる 4 点の結果を design.md の未決事項に反映する
