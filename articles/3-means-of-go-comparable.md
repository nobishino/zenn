---
title: "Go言語のcomparableには3つの意味がある"
emoji: "👋"
type: "tech" # tech: 技術記事 / idea: アイデア
topics: [Go]
published: false
---

タイトルの通り、Go言語のcomparableには3つの意味があります。

普段意識する必要はないと思いますが、混乱することがあるかもしれないのでまとめました。

とりあえずGo言語のcomparableが多義的になっているということだけ頭の片隅においておき、実際に混乱したときにこの記事の残りの部分を読めば十分だとおもいます。

# 要約

- Go言語には次の3箇所で"comparable"ということばが現れますが、それぞれ指す範囲が少しずつ異なります。
  - [言語仕様書](https://go.dev/ref/spec#Comparison_operators)
  - [型制約](https://go.dev/ref/spec#Type_constraints)
  - [reflectパッケージ](https://github.com/golang/go/issues/46746)
- この微妙な違いは、interface型のcomparabilityの複雑さからやむを得ず生じているものです。

# ３つのcomparableの関係

厳密に書くとまどろっこしくなるので、少し不正確な書き方が混じるかもしれません。ご了承ください。

Go言語には次の3箇所で"comparable"ということばが現れますが、それぞれ指す範囲が少しずつ異なります。それぞれを次のように記載することにします。

- [comparable(言語仕様)](https://go.dev/ref/spec#Comparison_operators)
- [comparable(型制約)](https://go.dev/ref/spec#Type_constraints)
- [comparable(reflect)](https://github.com/golang/go/issues/46746)

これらの包含関係をVenn図で表すと次のようになります:

![3つのcomparable](/images/venn-comparable-go.png)

では、それぞれの内容を見ていきましょう。

# comparable(言語仕様)

[言語仕様](https://go.dev/ref/spec#Comparison_operators)上の"comparable"な値とは、次の値のことです。

- boolean、整数、浮動小数点数、複素数、文字列、ポインター、チャネル、配列の値
- interface型の値

これらの値は同じ型同士で`==`による比較ができます。

一方で、次の値はcomparableではありません。

- 関数、スライス、マップの値

これを例示したのが、次のサンプルプログラムです。


```go
// https://go.dev/play/p/0du6Ya70CtL
func main() {
	var a int
	var b string
	var c fmt.Stringer
	var d []int

	fmt.Println(a == a) // true
	fmt.Println(b == b) // true
	fmt.Println(c == c) // true
	// fmt.Println(d == d) // compile error
	_ = d
}
```

## interface型同士の比較は`panic`を引き起こす場合がある

ここで厄介なのは、interface型同士の比較は`panic`を引き起こす場合があることです。

> A comparison of two interface values with identical dynamic types causes a run-time panic if values of that type are not comparable.

> 同一の動的型をもつ2つのインタフェース値の比較は、その型の値がcomparableではないとき、run-time panicを引き起こす。

例えば次のサンプルプログラムを実行するとrun-time panicになります。

```go
// https://go.dev/play/p/gNmPDq0pl2X
func main() {
	var e interface{}
	e = map[int]int{}
	fmt.Println(e == e) // panic
}
```

つまり、言語仕様上の「comparableな値」はinterface型の値も含みますが、そのような値を比較したときはpanicを引き起こすことがあります。

標語的にまとめると、**comparable(言語仕様)は静的に判断され、panicを引き起こす値も含みます。**

# comparable(型制約)

Go1.18のジェネリクス導入により、事前宣言された型制約`comparable`が導入されました。型制約はすなわちインタフェースなので、「型`X`は`comparable`を実装する」という言い方ができます。

:::message
comparable型制約について詳しくは[Go言語のジェネリクス入門(1)](https://zenn.dev/nobishii/articles/type_param_intro#%E5%85%B7%E4%BD%93%E4%BE%8B3%3A-%E5%9E%8B%E3%83%91%E3%83%A9%E3%83%A1%E3%83%BC%E3%82%BF%E3%82%92%E6%8C%81%E3%81%A4%E5%9E%8Bset%5Bt-comparable%5D)に書きました。
:::

言語仕様書によると、型`T`が`comparable`を実装するのは次の場合です:

> * T is not an interface type and T supports the operations == and !=; or
> * T is an interface type and each type in T's type set implements comparable.

つまり、型`X`が`comparable`を実装するのは、次のときです。

- `X`がboolean、整数、浮動小数点数、複素数、文字列、ポインター、チャネル、配列型であるとき
- `X`がinterface型であってその型セットがcomparableを実装する型のみからなるとき

ややこしく書いてあるのはunions（後述)を考慮した記述なのでこの記事の本筋とは関係ありません。重要なのは、comparable(言語仕様)と違い、`X`がふつうのinterface型であるとき、comparable(型制約)は`X`を**含まない**ということです。これを示すのが次のサンプルプログラムです。

```go
// https://go.dev/play/p/WzCU9sh__fD
func main() {
	var x int
	var y any
	f(x) // ok
	// f(y) // compile error
	_ = y
}

func f[T comparable](x T) {}
```

つまり、**comparable(型制約)はcomparable(言語仕様)と比べて、ふつうのinterface型を一切含まない分だけ狭い概念になっています。**

:::message
「ふつうの」interface型の正確な意味は次節で補足します。
:::

このようになった理由は、`comparable`型制約を満たしてインスタンス化された関数の中で`==`による比較を行ったときにrun-time panicが起きないことを保証したほうがよいと判断されたからとおもわれます。

:::message
この議論をしていたissueが探せてないので見つかったら追記したいです。PR歓迎しています。
:::

標語的にまとめると、**comparable(型制約)は静的に判断され、panicを引き起こさずに比較できる型だけを含みます。**

## 少し進んだ補足: unionsを含むinterface型はcomparableを実装しうる

上記で「ふつうの」interface型と限定して書いたのは、正確には **「unionsを含まないinterface型」** のことでした。

unionsを含むinterface型はcomparableを実装できます。これを例示したのが次のサンプルプログラムです。

```go
// https://go.dev/play/p/Z456wQiTfum
type C interface {
	~int | ~string
}

func main() {}

func f[T C](t T) {
	g(t) // compileできる。つまりCがcomparableを実装することがわかる
}

func g[S comparable](s S) {}
```

:::message
unionsについて詳しくは[Go言語のジェネリクス入門(1)](https://zenn.dev/nobishii/articles/type_param_intro#unions)を参照してください。
:::

unionsを含まないinterface型を便利に表すことばがないので、この記事やこの記事の冒頭に掲載したVenn図ではinterfaceというのをGo1.17以前の**「unionsを含まないinterface」**の意味で使わせてもらっています。もしこの点で混乱させていたらすみません。。

# comparable(reflect)

reflectパッケージに`reflect.Value.Comparable() bool`というAPIを追加する[proposal](https://github.com/golang/go/issues/46746)がacceptedになりました。

```go
// Comparable reports whether the type of v is comparable.
// If the type of v is an interface, this checks the dynamic type.
// If this reports true then v.Interface() == x will not panic for any x.
func (v Value) Comparable() bool
```

このコメントによると、

> vがinterface型の場合、このメソッドは動的型をチェックする。
> このメソッドが`true`を返す場合、`v.Interface() == x`はいかなる`x`に対してもpanicしない。

と書かれています。

つまり、comparable(reflect)は次の値を含むと読み取れます。

- boolean、整数、浮動小数点数、複素数、文字列、ポインター、チャネル、配列の値
- interface型の値であって、その動的型の値がcomparable(言語仕様)であるもの
  - つまり、比較時にpanicを引き起こさないinterface型の値

したがって、包含関係でいうと、comparable(reflect)はcomparable(型制約)よりも広く、comparable(言語仕様)よりも狭いということになります。

標語的に言うと、**comparable(reflect)は動的に判断され、panicせずに比較できるすべての値を含みます。**

# まとめ

- comparable(言語仕様)は静的に判断され、panicを引き起こす値も含みます。
- comparable(型制約)は静的に判断され、panicを引き起こさずに比較できる型だけを含みます。
- comparable(reflect)は動的に判断され、panicせずに比較できるすべての値を含みます。

![3つのcomparable](/images/venn-comparable-go.png)