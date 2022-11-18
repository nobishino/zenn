---
title: "Basic Interfaceはcomparableを満たすようになる(でも実装するようにはならない)"
emoji: "👋"
type: "tech" # tech: 技術記事 / idea: アイデア
topics: [Go]
published: false
---


前回の記事[Go言語のcomparableには3つの意味がある](./3-means-of-go-comparable.md)において、言語仕様のcomparableと型制約のcomparableは指す範囲が異なるということを説明しました。たとえば、`any`型は言語仕様上comparableですが、`comparable`型制約を満たしていません。

このギャップをなくすProposalがacceptされそうです。今回はその内容を説明します。

https://github.com/golang/go/issues/56548

言語仕様としての理屈にそれほど関心がない読者の人は要約だけ読めば十分だと思います。

# 要約

- unionsをふくまないinterface型(basic interface)について、Go1.18時点では`comparable`型制約を満たさないようになっていた
- このproposalが実装されると、basic interfaceは`comparable`型制約を満たすことができるようになる
- 特別な場合として、`any`は`comparable`型制約を満たすようになるので、次のようなコードが書けるようになる(今は書けない)

https://go.dev/play/p/_TyieBbyzXx

```go
func f[T comparable](T) {}

func main() {
	var x any
	f(x)
}
```



# 用語の整理

authorのgriesemerさんは議論のために次の用語を用いています。

| 用語 | 前回の記事における対応する用語 | 意味 |
| ---- | ---- | ---- |
| spec-comparable | comparable(言語仕様) | 言語仕様上`==`を使ってもコンパイルができるすべての型 |
| strictly comparable | comparable(型制約) | `==`で`panic`せずに比較できる型 |

前回の記事のvenn図とは次のように対応します:

![comparableの種類](/images/venn-comparable-go-2.png)

# proposalで何が変わるのか

## 現在の状態(Go1.18)

Go1.18では、あらゆる型制約`C`について、

- 型`T`が`C`を実装する(implement)
- 型`T`が`C`を満たす(satisfy)
- 型`T`が`C`の型集合に属する

という3つの文章は全く同じ意味です。

そして`comparable`の型集合は、strictly comparableな型全体からなる集合です。つまり、

- spec-comparableな非インタフェース型
- strictry-comparableな型のみをフィールドに持つstruct型

だけが`comparable`の型集合に属します。これを示すのが次のサンプルコードです。

https://go.dev/play/p/ULeOmhRP6s3

```go
func f[T comparable](T) {}

type S1 struct {
	int
}

type S2 struct {
	fmt.Stringer
}

func main() {
	// spec-comparableな非インタフェース型なのでstrictly comparableである
	f(1) 

	// spec-comparableだがインタフェース型なのでstrictly comparableではない
	var x any 
	f(x) // compile error

	// strictly comparableな型であるintのみをフィールドにもつstruct型はstrictly comparableである
	var s1 S1 
	f(s1) // OK

	// strictly comparableではないfmt.Stringerをフィールドに持つstruct型はstrictly comparableではない
	var s2 S2 
	f(s2) // compile error
}
```

## Proposal採用後の状態(Go1.xx)

Proposal採用後は次のようになります。

- 型`T`が`C`を実装する(implement)
- 型`T`が`C`の型集合に属する

の2つの文章は全く同じ意味です。この点はGo1.18と変わりありません。

しかし、

- 型`T`が`C`を満たす(satisfy)

はこの2つよりも広い場合にあてはまることがあります。

また、先程のサンプルコードはすべてコンパイルできるようになります(執筆時点ではできません)

https://go.dev/play/p/ULeOmhRP6s3

```go
func f[T comparable](T) {}

type S1 struct {
	int
}

type S2 struct {
	fmt.Stringer
}

func main() {
	f(1) 

	var x any 
	f(x)

	var s1 S1 
	f(s1)

	var s2 S2 
	f(s2) 
}
```

一方、次のようなコードはGo1.18でもProposal採用後でもコンパイルできません。

https://go.dev/play/p/vTpk2lXlA6L

```go
func f[T comparable](T) {}

func main() {
	var s []int
	f(s) // compile error: sliceはspec-comparableですらない
}
```

### 型制約を満たす(satisfy)の定義

- 型`T`が`C`を実装する(implement)
- 型`T`が`C`を満たす(satisfy)

この2つは異なる意味を持つようになると書きました。では、型`T`が`C`を満たす(satisfy)とは正確にはどのような意味になるのでしょうか？

proposalによると、型`T`が型制約`C`を満たす(satisfy)のは次の2つのいずれかのときです。

- 型`T`が`C`を実装する(implement)とき
- `C`が`interface{comparable; E}`の形で書けて、`T`がspec-comparableであり、かつ`T`が`E`を実装する(implement)とき
  - ここで、`E`はbasic interfaceつまりunionsを含まないinterfaceであるものとする

具体例を見る前に、なぜこの2つを別概念にする必要があるかを説明します。

## なぜimplementとsatisfyを別概念にする必要があるか

要請として、`any`が`comparable`型制約を満たす(satisfy)ようにしなければいけないとしましょう。
その上でsatisfyとimplementが全く同じ意味であると仮定すると、次のようにまずいことになります。

守るべき前提として、`[]int`のようにspec-comparableではない型が`comparable`型制約を満たさないようにしなければいけません。
ここで、`[]int`は`any`を実装することに注意すると、次のようなことが言えます。

- `[]int`は`any`を実装する (`any`の定義から当然)
- `any`は`comparable`を実装する (要請と、satisfyとimplementの同義性から)
- `[]int`は`comparable`を実装しない (守るべき前提)

すると、**`[]int`は`any`を実装し、`any`は`comparable`を実装するのに、`[]int`は`comparable`を実装しない**という奇妙な結論になってしまいます。
つまり、`implement`するという関係は本来[推移法則](https://ja.wikipedia.org/wiki/%E6%8E%A8%E7%A7%BB%E9%96%A2%E4%BF%82)を満たさないといけないはずなのに、推移法則を満たさない結果になってしまっています。

これは言語仕様として困るので、ここまでで使った前提のどれかは諦めないといけません。
最も諦めがつくのは、**implementとsatisfyが同義であるという前提**です。

## 具体例

### `any`は`comparable`を満たす

proposalによると、型`T`が型制約`C`を満たす(satisfy)のは次の2つのいずれかのときです。

- 型`T`が`C`を実装する(implement)とき
- `C`が`interface{comparable; E}`の形で書けて、`T`がspec-comparableであり、かつ`T`が`E`を実装する(implement)とき
  - ここで、`E`はbasic interfaceつまりunionsを含まないinterfaceであるものとする

`C`に`comparable`を代入して整理すると次のようになります。`comparable`はbasic intefaceである`any`をつかって次のように書けることに気をつけます:

```go
// これは疑似コードです
type comparable inteface {
	comparable 
	any
}
```

**型`T`が型制約`comparable`を満たす(satisfy)のは次の2つのいずれかのときである:**

- 型`T`が`comparable`を実装する(implement)とき
- `T`がspec-comparableであり、かつ`T`が`any`を実装する(implement)とき

どんな型も`any`を実装することを使えば、この条件は次と同じことです:

**型`T`が型制約`comparable`を満たす(satisfy)のは次の2つのいずれかのときである:**

- 型`T`が`comparable`を実装する(implement)とき
- `T`がspec-comparableであるとき

`T`に`any`を代入すれば`any`はspec-comparableなので、`any`は`comparable`を満たします。

### より複雑な制約

```go
type C interface {
	comparable
	String() string
}
```

`fmt.Stringer`がこれを実装するかを考えてみましょう。`fmt.Stringer`は`comparable`を実装しません。しかし、`fmt.Stringer`はspec-comparableであり、かつ`String() string`を実装します。したがって**`fmt.Stringer`は`C`を実装はしませんが`C`を満たします。**

`[]int`はどうでしょうか？`[]int`はspec-comparableではないので`C`を満たしません。

## ある型がstrictly comparableであることをコンパイル時にチェックする方法

Go1.18ではstrictly comparableな型だけが`comparable`を満たしていたので、`panic`せずに`==, !=`で比較できる型であることをコンパイル時にチェックできました。

proposal採用後は、ある型`T`が`panic`せずに`==, !=`で比較できる型であるのをコンパイル時にチェックできなくなるのでしょうか？

これにはauthorのgriesemerさんが次のような方法を提示しています。

https://github.com/golang/go/issues/56548#issuecomment-1317673963

```go
// we want to ensure that T is strictly comparable
type T struct {
	x int
}

// define a helper function with a type parameter P constrained by T
// and use that type parameter with isComparable
func TisComparable[P T]() {
	_ = isComparable[P]
}

func isComparable[_ comparable]() {}
```

どういうことなのでしょうか？実はそもそも型パラメータ型のspec-comparabilityは現在の言語仕様上明確に定義が書いてないのですが、これは次のように修正されるだろうと書かれています。

https://github.com/golang/go/issues/56548#issuecomment-1319052631

> Type parameters are comparable if each type in the type parameter's type set implements comparable.

> 型パラメータは、その型セットに属するそれぞれの型が`comparable`を実装するときにcomparableである。

:::message

筆者は型パラメータの型集合というのも不正確な表現な気がしていて、言いたいことは**型パラメータ`T`の型制約の型集合**だと思います。

:::

これを踏まえてもう一度コードをみてみます。

```go
// define a helper function with a type parameter P constrained by T
// and use that type parameter with isComparable
func TisComparable[P T]() {
	_ = isComparable[P]
}
func isComparable[_ comparable]() {}
```

型パラメータ型である`P`が`comparable`を実装するのは`P`の型制約である`T`の型セットに属するすべての型が`comparable`を実装するときですが、`T`は普通の型なので`T`の型集合は`T`飲みからなる集合です。それが`comparable`を実装するということは`T`はstrictly comparableです。よって、このコードのコンパイルができるならば`T`はspec-comparableであるばかりかstrictly comparableでもあるということが保証できます。

:::message

griesemerさんの書いているのがspec-comparabilityなのかちょっとはっきりしませんが、一つの可能な解釈で議論しました。このあたりはさすがに正式なspecのCLを読まないとわからなさそうです。

:::

# まとめ

以上をまとめると、つぎのVenn図のようになるとおもわれます:

- spec-comparableはstrictly comparableよりも広い
- spec-comparableな範囲は"satisfy comparable"な範囲と一致する
- strictly comparableな範囲は"implement comparable"な範囲と一致する

![comparableの種類](/images/venn-comparable-go-3.png)

:::message
ちょっと自信ない...情報歓迎中です
:::