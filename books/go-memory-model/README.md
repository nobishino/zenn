# Zenn book「よくわかるThe Go Memory Model 〜行間を図解で埋め尽くす〜」

Go Conference 2023 (A3-L) の発表「よくわかるThe Go Memory Model」を書籍化したZenn bookです。当初はスライドのスピーカーノートの文字起こしとして出発し、その後、書籍として独立した章（第4章・第5章）の追加や全体の添削を行っています。

## 構成

```
go-memory-model/
├── config.yaml                  # book設定（published: falseの間は非公開）
├── 1.gomm-motivation.md         # 第1章 本書のモチベーションと流れ
├── 2.sequential-consistency.md  # 第2章 並行処理の難しさと逐次一貫モデルの破綻
├── 3.happens-before-relation.md # 第3章 観測可能性とhappens-before関係
├── 4.synchronized-before.md     # 第4章 happens beforeのフォーマルな定義：sequenced beforeとsynchronized before
├── 5.synchronization.md         # 第5章 同期演算(Synchronization)を読み解く
├── 6.go119-sync-atomic.md       # 第6章 Go1.19メモリーモデルとsync/atomicパッケージ
├── 7.gomm-summary.md            # 第7章 まとめ
├── README.md                    # このファイル（Zenn上には表示されない）
└── review-note.md               # 通読レビューの記録（Zenn上には表示されない）
```

章のファイルは`<番号>.<slug>.md`の命名で、番号順に表示されます。

## 編集上のメモ

- 本文中の画像はリポジトリ**ルート**の `/images/` に置きます（Zennの画像パス仕様）。本文からは `/images/xxx.png` で参照します
- 原文引用には「拙訳:」を添える形式で統一しています
- 図はmermaid記法で描いています。happens-before関係を示す矢印は太字（`==>`）で強調します
- サンプルコードは完全な`package main`プログラムとしてGo Playgroundにshareし、コードブロック直後に「👉 Go Playgroundで実行する」リンクを添えます
- 公開するときは `config.yaml` の `published: false` を `true` に変更します
