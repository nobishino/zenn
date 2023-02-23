---
title: "[書評+α] Harsanyi \"100 Go Mistakes and How to Avoid Them\""
emoji: "👻"
type: "idea" # tech: 技術記事 / idea: アイデア
topics: [Go]
published: false
---

Teiva Harsany "100 Go Mistakes and How to Avoid Them"の書評(レビュー)+勉強会のお誘いです。

結論から言うとすごく良い本です。リンクは↓にあります。

https://www.manning.com/books/100-go-mistakes-and-how-to-avoid-them

# 要約

- Harsanyi "100 Go Mistakes and How to Avoid Them"は気軽に読めてGo中級者からのステップアップに最適なすばらしい本です
- 内容を人に話して理解を深めたいので、本を所持している人向けの勉強会をやろうと思っています

# 本の概要

タイトルを直訳すると「Go言語における100個の間違いと、それを回避する方法」のようになります。

Go言語をよりうまく・正しく使うためのTipsを100個集めた本で、Mistake(間違い)ごとにセクションが別れています。

- ひとつのセクションが数ページで完結していて、ほとんどの場合独立して読める。よってつまみ食いでも得るものがある。
- mistakeの例を挙げ、それがなぜmistakeなのかを原理原則から説明し、どう直せばよいのかを述べるというフォーマットになっている。それにより、具体的な場面のイメージしやすさと、問題を基礎から理解できる本質性を両立している。
- 100個のセクションはジャンルごとに章に分けられているので関心のあるものを続けて読むこともできる。

著者のTeiva Harsanyiさん([@teivah](https://twitter.com/teivah))はTwitterプロフィールによるとDockerのソフトウェアエンジニアだそうです。

# 章目次とその直訳

具体的には次のような章があります。これらの章の中に"mistakes"単位で別れたセクションがあります。

| 原題 | 省タイトルの拙訳 | "mistake"の例 |
| ---- | ---- | ---- | 
| Go: Simple to learn but hard to master | | |
| Code and project organization | コードとプロジェクトの組織化 | #6: Interface on the producer side |
| Data types | データ型 | |
| Control structures | 制御構造 | |
| Strings | 文字列 | |
| Function and methods | 関数とメソッド | |
| Error management | エラー管理 | |
| Concurrency: Foundations | 並行性(基礎編) | |
| Concurrency: Practice |　並行性(実践編) | |
| The standard library | 標準ライブラリ | |
| Testing | テスティング | |
| Optimizations | 最適化 | |

# チャプターごとに軽く紹介

## Go: Simple to learn but hard to master 

## Code and project organization 

型やパッケージなどの抽象化の単位ごとにどう使うかを述べている章です。

テーマはいろいろありますが、Goのinterfaceの特徴(implicit)とRob Pikeによる"Don’t design with interfaces, discover them"の警句から論理的にinterfaceの良い使い方を述べているところがお気に入りです。

## Data types 

数値型やスライス・マップ型それぞれに特有の注意点を述べています。スライスのcapacity leakやmapの内部構造など普通にGoを書いているだけだと知る機会のない情報もあります。

## Control structures 

## Strings 

## Function and methods 

## Error management 

## Concurrency: Foundations 

## Concurrency: Practice 

また、`sync.Cond`の説明が秀逸でした。

## The standard library 

## Testing 

## Optimizations 

- CPUキャッシュと局所性、stack/heapとエスケープ解析、GCといった最適化の前提となる低レイヤー知識の基礎 
- Goプログラムの最適化のための実際のプロファイリング方法

の2つを学べます。

メモリの局所性がなぜパフォーマンスに重要なのかをCPUキャッシュの仕組みからわかりやすく説明しています。また、stackとheapの違いや具体的にどのようにheapへの「エスケープ」が起きるのかをサンプルコードと豊富な図解とともに説明しています。筆者の低レイヤーへの知見が伺える章です。