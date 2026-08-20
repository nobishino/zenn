# zenncode

記事に埋め込まれた Go のサンプルコードを検証し、Go Playground のリンクを維持するツール。

課題と全体設計は [issue #8](https://github.com/nobishino/zenn/issues/8) を参照。
実装済みはフェーズ1（検証）、フェーズ2（リンクの自動生成・同期）、
フェーズ3（期待結果の指定と照合）。

## 使い方

```sh
make verify      # 全記事のサンプルコードをコンパイルし、リンクを照合する
make fix         # 検証を通ったスニペットを共有し、リンクを記事に書き込む
make fix-dry     # fix が何をするかだけ表示する
make baseline    # 現時点で失敗するものを baseline に記録する
make list        # 1スニペット1行で一覧
make test        # zenncode 自身のテスト
```

記事を書いたら `make fix` → `make verify` が緑、が通常の流れ。

`make test` と `make verify` は push と pull request のたびに GitHub Actions でも
走る（[.github/workflows/zenncode.yml](../../.github/workflows/zenncode.yml)）。
`verify` はネットワークを使わないので、CI が赤いときは記事かツールの問題。
別に週次で最新の Go を使った検証も回していて（[zenncode-latest.yml](../../.github/workflows/zenncode-latest.yml)）、
そちらは PR のゲートではなく「Go が上がって記事が古くなった」ことに気づくためのもの。

パス引数はリポジトリルートからの相対パスで書く（ツールは `articles/` を含む
祖先ディレクトリを自動で探してそこへ移動する）。

```sh
make verify ARGS="articles/go-encoding-csv-rfc.md"
cd tools/zenncode
go run . show articles/go-encoding-csv-rfc.md:52   # 正規化後のプログラムを表示
go run . list -v articles                          # 失敗の詳細つき一覧
```

## 何をしているか

記事中の ` ```go ` ブロックそれぞれについて:

1. **正規化** — `package main` が無ければ足し、goimports で import を補い、
   `func main` が無ければ空の `func main() {}` を足す。記事本文は import を
   省いた抜粋であることがほとんどなので、この復元が要になる。
2. **ビルド** — 一時ディレクトリに module を作ってコンパイルする。
   `GOPROXY=off` なので標準ライブラリのみ、ネットワーク不要。
3. **共有**（`fix` のみ）— 正規化後のプログラムを Playground に投稿し、
   得られた URL を記事の該当ブロックの隣に書き込む。
4. **照合**（`verify`）— 記事のリンクが、そのブロックのコードに対応する
   URL になっているかを見る。

正規化後のプログラムのハッシュがスニペットの ID で、`zenncode-lock.json` が
「ハッシュ → Playground URL」を保持する。`verify` はこのファイルだけを見るので
**ネットワークに一切アクセスしない**。共有するのは `fix` だけ。

### 共有が冪等であること

Playground の共有エンドポイントは内容アドレスで、同じバイト列を投稿すると
同じ ID が返る。ただし `Content-Type: text/plain` で送った場合に限る
（フォーム形式で送るとサーバがパーセントエンコードされた本文を保存してしまい、
別の ID になる）。この性質のおかげで:

- コードを変えていないスニペットを再共有してもリンクは変わらない
- 記事に既にあるリンクが正しければ `fix` は何も書き換えない

実際、既存記事の82スニペットを再共有したところ、全て記事にあるURLと一致した。

## ディレクティブ

ブロックの直前に HTML コメントで書く。Zenn の表示には出ない。

```markdown
<!-- zenncode: skip -->
```

| キー | 状態 | 意味 |
| ---- | ---- | ---- |
| `skip` | 実装済 | このブロックを検証もリンク付けもしない |
| `imports=math/rand,fmt` | 実装済 | goimports が迷う場合に import を指定する |
| `goversion=1.18` | 実装済 | 一時 module の `go` ディレクティブ |
| `playground=none` | 実装済 | 検証はするが Playground リンクを付けない。既にあるリンクは `fix` が消す |
| `expect=build` | 実装済 | 既定。コンパイルが通ればよい（実行はしない） |
| `expect=run` | 実装済 | 実行して正常終了することまで見る |
| `expect=compile-error` | 実装済 | **意図的にコンパイルエラーになる**サンプル。通ってしまったら失敗 |
| `expect=panic` | 実装済 | **意図的に実行時 panic する**サンプル。正常終了したら失敗 |
| `error="cannot use"` | 実装済 | コンパイルエラーの文言を正規表現で照合する（`expect=compile-error` を含意） |
| `panic="uncomparable type"` | 実装済 | panic の文言を正規表現で照合する（`expect=panic` を含意） |
| `output=next` | 実装済 | 直後のコードブロックを期待 stdout として照合する（`expect=run` を含意） |
| `timeout=5s` | 実装済 | 実行の制限時間（既定 30s）。超えたら失敗 |
| `file=` | 未実装 | 設計だけ済み。書くとエラーになる（黙って無視されるのを避けるため） |

### 期待結果の指定

既定は「コンパイルが通ること」だけを見る。この記事群には意図的に止まる/落ちる
サンプルが多いので、**実行は明示的に頼まれたときだけ**行う。

```markdown
<!-- zenncode: expect=compile-error error="does not implement comparable" -->
<!-- zenncode: expect=panic panic="comparing uncomparable type" -->
<!-- zenncode: expect=run output=next -->
```

`expect=panic` は「プログラムが自分の都合で落ちること」を要求する。判定は
標準エラーの `panic: ` または `fatal error: `（デッドロック検出など、recover
できない側の異常終了）で行うので、`os.Exit(1)` のような単なる非ゼロ終了は
panic として認めない。

`expect=compile-error` や `expect=panic` のブロックにも Playground リンクは付く。
「期待どおりの結果になった」ことが検証できているなら、Playground を開けば
記事が言っているコンパイルエラーや panic がそのまま再現するからで、
クイズ形式の記事（`goquiz_20230817.md`）はこれで成り立っている。
リンクが付かないのは、そもそも Go として構文解析できずプログラムを組み立て
られなかったブロックと、`playground=none` を指定したブロック。

### リンクを付けたくないとき

コードは載せるが Playground リンクは要らない、という場合は `playground=none` を
書く。検証は通常どおり行われ、リンクだけが対象外になる。

```markdown
<!-- zenncode: playground=none -->
```

これは「リンクを付けない」ではなく「リンクが無い状態にする」指定なので、
**既に本文にあるリンクは `fix` が削除する**（`verify` はそれを link problem として
報告する）。lock の該当エントリも full run の `fix` で prune される。

ただし、2つのコードブロックの間にリンクが1行だけある場合、それがどちらの
ブロックのものかは本文からは決まらず、ツールは記事の多数派から推測している
（下記「リンクの位置」）。この推測で削除まで行うと隣のブロックのリンクを
黙って失う可能性があるので、**曖昧なリンクは `fix` は消さず、`verify` が
手で消すよう促す**。

## リンクの位置

記事によってリンクをコードブロックの上に置いていたり下に置いていたりする。
`fix` は記事ごとに多数派の側を判定してそちらに新しいリンクを置く
（判断材料が無い記事はブロックの下）。既にリンクがあるブロックは
その位置のまま URL だけ更新するので、記法は記事ごとに保たれる。

ブロックとブロックの間にリンクが1つだけある場合、上下どちらのブロックのものとも
解釈できる。この曖昧なケースは記事の多数派の側に従って帰属を決める。

## baseline

`zenncode-baseline.json`（リポジトリルート）に、現時点で検証を通らない
スニペットを記録してある。`verify` はここに載っているものを許容し、
それ以外の失敗だけを報告する。エントリは**正規化後のプログラムのハッシュ**で
照合するので、スニペットを編集すると baseline から外れ、単体で通る必要がある。

165個中55個が失敗している。フェーズ3の前は75個で、意図的にコンパイルエラーに
なるサンプルを `expect=compile-error` に移したぶんが減った。残りの内訳は

- 記事本文が宣言の断片（`~int | ~string` など、それ単体では Go でない）
- 前のブロックで定義した型を参照している（`undefined: MyFloat` 系）

の2つ。後者を扱うには、前のブロックの宣言を取り込む `deps=` のような
ディレクティブが要る（未着手）。

検証を通らないスニペットには `fix` はリンクを付けない（リンクは「記事の言うとおりの
結果になる」という約束なので、その確認が取れていないコードには付けない）。
