# Bluesky CAR Post Extractor

`myapp/` は GitHub Pages 向けの静的サイトです。Bluesky の repo export `.car` をブラウザ内だけで解析し、投稿レコードの `createdAt` と `text` だけを含む JSON / CSV を生成します。

## Build

```bash
./myapp/build.sh
```

このコマンドは以下を生成します。

- `myapp/wasm_exec.js`
- `myapp/car-extractor.wasm`

## Local preview

```bash
python3 -m http.server 8000
```

その後 `http://localhost:8000/myapp/` を開きます。

## GitHub Pages

`myapp/` の内容を Pages の配信対象に置けば、そのまま静的サイトとして動きます。`Jekyll` を経由させないため、`.nojekyll` も同梱しています。

## How to get a CAR file from Bluesky

2026-03-24 時点で確認できた公式 Bluesky アプリの現行ソース上の文言では、取得導線は次のとおりです。

1. `Settings`
2. `Account`
3. `Export my data`
4. `Download CAR file`

補足:

- Bluesky の説明では、CAR には public data records が入り、画像などの media embeds や private data は含まれません。
- GUI 上にも同じ案内と公式リンクを表示しています。

## License

`myapp/` を単体の repository として公開する場合も、このディレクトリは `indigo` と同じく MIT または Apache-2.0 のデュアルライセンスとして扱うのが安全です。`cmd/gosky/car.go` を参考にした派生部分があるため、`NOTICE.md` と upstream ライセンス本文もこのディレクトリに含めています。

- `NOTICE.md`
- `LICENSE-MIT`
- `LICENSE-APACHE`

## Output format

ダウンロードされる JSON は配列のみで、CSV は `createdAt,text` のヘッダ付きです。

```json
[
  {
    "createdAt": "2025-03-06T20:08:56.020Z",
    "text": "結局、見栄と話題性の世界なんだすこ"
  }
]
```

```csv
createdAt,text
2025-03-06T20:08:56.020Z,"結局、見栄と話題性の世界なんだすこ"
```
