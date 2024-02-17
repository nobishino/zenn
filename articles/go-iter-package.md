
---
title: "Go言語でdata raceが起きるときに起きる（かもしれない）こと"
emoji: "😽"
type: "tech" # tech: 技術記事 / idea: アイデア
topics: [Go, Concurrency, MemoryModel]
published: true
---

# はじめに

:::message
この記事で挙げる「驚くような動き」を心配しなければならないのはdata raceが存在する場合であって、「並行処理を使うといつもこのようなことが起こりうる」わけではありません。

つまり、並行処理を使っていても、data raceを発生させていなければ「驚くような動き」は起きません。
:::

https://go.dev/play/p/KLR5U0rbzEN

上記のPlaygroundで実行すると、次のように`panic`することがあります。

```
panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x1 addr=0x0 pc=0x45d33c]
goroutine 1 [running]:
fmt.(*buffer).writeString(...)
```

`string`型の値は複数の部分からなっており、文字列の長さを表す部分とバイト列の先頭へのポインタを持っています。

https://github.com/golang/go/blob/97daa6e94296980b4aa2dac93a938a5edd95ce93/src/runtime/string.go#L232-L235

長さを表す部分とそのポインタ部分が一緒に更新されれば問題ないのですが、reader側から中途半端に片方だけ更新された状態を観測してしまうと、nil pointer dereferenceが発生します。

# 参考資料

筆者が参考にした資料と、参考になりそうな資料を挙げておきます。

| タイトルとリンク                                            | 概要                                                                                                    |
|------------------------------------------------------------|---------------------------------------------------------------------------------------------------------|
# 最後に

執筆にあたり次の方から情報やフィードバックをいただきました。ありがとうございます。


もちろん、記述の誤りなどについてのすべての責任は筆者にあります。
