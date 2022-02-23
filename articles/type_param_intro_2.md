---
title: "Go言語のジェネリクス入門(2) インスタンス化と型推論" # 記事のタイトル
emoji: "😸" # アイキャッチとして使われる絵文字（1文字だけ）
type: "tech" # tech: 技術記事 / idea: アイデア記事
topics: ["go"] # タグ。["markdown", "rust", "aws"]のように指定する
published: false # 公開設定（falseにすると下書き）
---
- [はじめに](#はじめに)
- [型セットについての資料](#型セットについての資料)
- [インスタンス化とは](#インスタンス化とは)
  - [インスタンス化は2ステップで行われる](#インスタンス化は2ステップで行われる)
    - [具体例(インスタンス化の失敗)](#具体例インスタンス化の失敗)
  - [型推論はインスタンス化の前に行われる](#型推論はインスタンス化の前に行われる)
    - [型推論が成功してもインスタンス化が失敗することはある](#型推論が成功してもインスタンス化が失敗することはある)
  - [全体像](#全体像)
- [型推論の概要](#型推論の概要)
  - [関数引数型推論(概要)](#関数引数型推論概要)
  - [制約型推論(概要)](#制約型推論概要)
- [unification/unify](#unificationunify)
  - [unificationの厳密な定義](#unificationの厳密な定義)
    - [例](#例)
    - [たとえ話: 解の求め方がわからなくても方程式は定義できる](#たとえ話-解の求め方がわからなくても方程式は定義できる)
  - [substitution mapとentry](#substitution-mapとentry)
  - [型の同一性(identity)と等価性(equivalence)](#型の同一性identityと等価性equivalence)
- [関数引数型推論の厳密な定式化](#関数引数型推論の厳密な定式化)
- [制約型推論の厳密な定式化](#制約型推論の厳密な定式化)
- [具体例や未解決の問題](#具体例や未解決の問題)
  - [公式ドキュメントに見る制約型推論の活用例](#公式ドキュメントに見る制約型推論の活用例)
  - [関数引数型推論と引数の順序](#関数引数型推論と引数の順序)
  - [制約型推論とdefined type、型推論インタリービング](#制約型推論とdefined-type型推論インタリービング)

# はじめに

[Go言語のジェネリクス入門(1)](https://zenn.dev/nobishii/articles/type_param_intro)

の続編で、インスタンス化や型推論について解説します。

実用上はコンパイルエラーになったら直せばいいのでここに書かれているようなことを知る必要はあまりなく、知っていると時々こう書けばよいというアイディアが出て便利かもしれない、くらいです。

好奇心を満たしたり、細かい仕様を正確に理解したくなったときにこの記事をご活用ください。Go言語仕様書は非常に読みやすい言語仕様書ですが、それでもジェネリクス関係の仕様を正確に理解するのは骨が折れるはずなので、理解の助けになればと思います。

# 型セットについての資料

この記事では型セット(Type set)についての知識を前提とします。Type setについては前編では説明していないのですが、次の記事やスライドで説明しています。

| リンク | 内容紹介 |
| ---- | ---- |
| [初めての型セット](https://speakerdeck.com/nobishino/introduction-to-type-sets) | Go1.17リリースパーティの発表スライドです。「型セット」と「実装」の概念理解にフォーカスしています。 すこし図があります。|
| [Go の "Type Sets" proposal を読む - Zenn](https://zenn.dev/nobishii/articles/99a2b55e2d3e50)| Type Setsのプロポーザルが出たときに書いた記事です。前半は経緯の解説なので今は読む必要ありません。Type setの仕様は後半で解説しています。| 
| [GopherCon 2021: Robert Griesemer & Ian Lance Taylor - Generics!](https://www.youtube.com/watch?v=Pa_e9EeCdy8) | [英語] Go言語開発者によるジェネリクス解説です。前半のgriesemer氏の発表部分に型セットの説明があります。| 
| [Go言語仕様書(Go1.18ドラフト) - Interface types](https://tip.golang.org/ref/spec#Interface_types) | 言語仕様書の型セット該当部分です。 |

# インスタンス化とは

ジェネリックな関数と型は、使う前に必ずインスタンス化して普通の関数・型にする必要があります。

インスタンス化とは、それぞれの型パラメータに具体的な型引数(type argument)を代入することです。

:::message

この記事では、type parameterの訳語として「型パラメータ」、type argumentの訳語として「型引数」を使います。

意味的には疑問もあるところと思いますが、口頭での言いやすさなども加味すると便利なため、個人的に採用しています。

:::

次の例では、`Print[T any]`関数の`T`という型パラメータに`string`という型引数が代入されることで、`Print`関数のインスタンス化が行われています。


```go
package main

import (
	"fmt"
)

// This playground uses a development build of Go:
// devel go1.18-c9fe126c8b Mon Feb 21 21:28:40 2022 +0000

func Print[T any](s ...T) {
	for _, v := range s {
		fmt.Print(v)
	}
}

func main() {
	Print("Hello, ", "playground\n")
}
```

https://gotipplay.golang.org/p/ZRx0SE4Q1Yi

`T`が型推論により自動決定されているので、あたかも`Print`というジェネリックな関数をそのまま使っているようにも見えます。
しかし、**実際には型推論がされていてもいなくてもインスタンス化は必ず行われています。**
つまり、上記のコードは次のように書き換えても同じ意味です。

```go
func main() {
	Print[string]("Hello, ", "playground\n")
}
```

:::message

言語仕様上は次のように明記されています。

https://tip.golang.org/ref/spec#Function_declarations

> If the function declaration specifies type parameters, the function name denotes a generic function. Generic functions must be instantiated when they are used.

https://tip.golang.org/ref/spec#Type_declarations

> If the type definition specifies type parameters, the type name denotes a generic type. Generic types must be instantiated when they are used.

:::

## インスタンス化は2ステップで行われる

インスタンス化は、次の2ステップで行われます。

- 型引数を対応する型パラメータに代入する。
- 代入された型引数が、対応する型パラメータの型制約を実装することを検証する。満たしていなければ、**インスタンス化が**失敗する。

:::message
型セットの復習として、次の2つが全く同じ意味であることを確認しておきます。

- 型引数が型制約を実装する
- 型引数が型制約の型セットに属する(型セットの要素である)

この記事ではこれと同じ意味で「型制約を満たす」という表現を用いることがあります。直感的なためです。
:::

### 具体例(インスタンス化の失敗)

https://gotipplay.golang.org/p/FUdYlX-a6oH

```go
package main

import "fmt"

type S[T fmt.Stringer] struct{}

type s = S[int]
// ./prog.go:7:12: int does not implement fmt.Stringer (missing String method)
```

- Step1: 型パラメータ`T`に型引数`int`を代入する
- Step2: intが`T`の型制約`fmt.Stringer`を実装することを検証する

これは`int`が`fmt.Stringer`を実装しないため、インスタンス化が失敗します。

## 型推論はインスタンス化の前に行われる

型引数が欠けているとき、Go言語処理系は型推論により欠けている型引数の決定を試みます。

インスタンス化をするまえに型引数は全て決定している必要があるため、型推論が必要な場合には、型推論はインスタンス化の前に行われます。

### 型推論が成功してもインスタンス化が失敗することはある

https://gotipplay.golang.org/p/t4n8HllorSt

```go
package main

func main() {
	var ch chan int
	f(ch)
}

func f[T <-chan int](ch T) {}
```

このコードは次のようにコンパイル失敗します。

> ./prog.go:5:3: chan int does not implement <-chan int

`T`を`chan int`と推論することには成功しているのですが、その推論結果である`chan int`という型引数が型制約`<-chan int`を実装しないため、インスタンス化のStep 2で失敗しています。

## 全体像

以上により、ジェネリックな型・関数を使うときには

- 型引数が欠けている場合には型推論を試みる
- インスタンス化を行う
- 関数呼び出しの場合には、引数をインスタンス化後の関数にわたす

というように処理が進みます。このそれぞれの段階でコンパイルエラーが発生し得ます。

これを図にすると次のようになります。型推論についてまだ説明していない詳細が含まれていますが、これは後ほど説明します。

![インスタンス化フロー](/images/instantiation_inference_flow.jpeg)

# 型推論の概要

Goジェネリクスにおける型推論とは、未知の型引数を既知の情報から推論し、決定することです。

既知の情報には2種類あり、それに応じて型推論メカニズムも2種類あります。この両方を決まった順序で行うというのが型推論の概要です。

| 型推論メカニズム | 使う情報 |
| ---- | ---- |
| 関数引数型推論 | 関数呼び出しで、引数として渡された値の型 | 
| 制約型推論 | すでに決定できた型引数と、未知の型引数が従う型制約| 

さらに関数引数型推論が、「型あり引数」をもちいるものと「型なし引数」を用いるものの2種類あります。

これらを合わせて、型推論は次のような4つのステップにより行われます。

1. 型あり引数を用いた関数引数型推論
1. 制約型推論
1. 型なし引数を用いた関数引数型推論
1. 制約型推論

型なし引数とは、[型なし定数(untyped constant)](https://tip.golang.org/ref/spec#Constants)の引数のことです。`f(1)`や`fmt.Println("hello world")`の引数が該当します。

型あり引数とは、それ以外の全ての引数のことです。

```go
x := 1 // xの型はintになる(※default type)
f(x) // xは型あり引数
```

:::message

型なし定数について詳しくは、つぎのDQNEOさんによる発表とスライドをみるとよくわかります。

- [入門Go言語仕様輪読会 Untyped Constants](https://youtu.be/bZZd_762zGA?t=752)
- [発表スライド](https://speakerdeck.com/dqneo/go-specification-untyped-constants)

:::

## 関数引数型推論(概要)

## 制約型推論(概要)

# unification/unify

## unificationの厳密な定義

**定義**

2つの型をunifyするとは、その2つの型を等価にするようなsubstitution map entryを見つけることである。

### 例

### たとえ話: 解の求め方がわからなくても方程式は定義できる

## substitution mapとentry

## 型の同一性(identity)と等価性(equivalence)


# 関数引数型推論の厳密な定式化

# 制約型推論の厳密な定式化

# 具体例や未解決の問題

## 公式ドキュメントに見る制約型推論の活用例

## 関数引数型推論と引数の順序

## 制約型推論とdefined type、型推論インタリービング