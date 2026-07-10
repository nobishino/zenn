# transcript — Zenn book版「よくわかるThe Go Memory Model」

`slide.pptx`（Go Conference 2023 A3-L発表スライド）をZennのbook形式に文字起こししたものです。本文はスライドのスピーカーノートをもとにし、ノートのないスライドは記載内容から本文を起こしています。

## 構成

```
transcript/
├── config.yaml                 # book設定
├── 1.gomm-motivation.md        # 第1章 発表のモチベーションと発表の流れ
├── 2.sequential-consistency.md # 第2章 並行処理の難しさと逐次一貫モデルの破綻
├── 3.happens-before-relation.md# 第3章 観測可能性とhappens-before関係
├── 4.synchronized-before.md    # 第4章 happens beforeの正体：sequenced beforeとsynchronized before
├── 5.synchronization.md        # 第5章 同期演算(Synchronization)を読み解く
├── 6.go119-sync-atomic.md      # 第6章 Go1.19メモリーモデルとsync/atomicパッケージ
├── 7.gomm-summary.md           # 第7章 まとめ
├── 8.gomm-appendix.md          # 付録: Message Passing Testの詳しい解き方
└── images/                     # スライドから抽出した画像
    ├── glossary.png                  # 用語集のスクリーンショット (image1)
    ├── messagepassing-experiment.png # Message Passing Test実験結果 (image2)
    ├── gopher.png                    # Gopherくんのイラスト (image3)
    └── atomics-experiment.png        # atomic版実験の様子 (image5)
```

## Zennで公開する手順

ZennのGitHub連携リポジトリに配置する場合:

1. このディレクトリを `books/<slug>/` にコピーする（例: `books/go-memory-model/`）
2. `images/` の中身はリポジトリ**ルート**の `/images/` に置く必要があります（Zennの画像パス仕様）。本文中の参照は `/images/xxx.png` になっているので、`transcript/images/*.png` をリポジトリルートの `images/` にコピーしてください
3. 公開するときは `config.yaml` の `published: false` を `true` に変更

## 補足

- スライド上で図形として描かれていた図（happens-beforeグラフ、概念相関図など）は、画像として抽出できないためmermaid図・テキスト図で再現しています
- Venn図（並行処理の難しさの2レベル）はテキストアートで再現しています
- スピーカーノート中の進行メモ（【予定ペース】など）は本文から除いています
