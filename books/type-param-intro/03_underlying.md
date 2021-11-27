---
title: "~とunderlying type"
---

# モチベーション

前章では`unions`を使ったインタフェース定義により、複数の数値型に適用できるgenericな`Max`関数を作れることを見ました。

```go
type Number interface {
    int | int32 | int64 | float32 | float64
}

func Max[T Number] (x, y T) T {
	if x >= y {
		return x
	}
	return y
}
```

では、次のように定義した`NewInt`や`NewNewInt`に対して`Max`関数を使用できるでしょうか？

```go
type NewInt int

type NewNewInt NewInt
```

「できない」というのが答えです。`int, NewInt, NewNewInt`はそれぞれ相異なる型であり、したがって`NewInt`と`NewNewInt`は`Number`インタフェースを実装しないからです。

# `~`(approximation element)

しかし、`NewInt`や`NewNewInt`も数値型であることに変わりはなく、`>=`などの演算子で比較することができるのですから、このような型を許すインタフェースを作りたいです。
もちろん、`NewInt`を直接unionsに加えれば`NewInt`に`Number`を実装させることはできます:

```go
// NewIntとNewNewIntがNumberを実装するようになった
type Number interface { 
    int | int32 | int64 | float32 | float64 | NewInt | NewNewInt
}
```

しかし、「`int`を元にして型定義で作られる新しい型」は無限にあるので、それら全てが`Number`を実装するようにしたいです。
そのための文法として、Go言語は`~`で表されるapproximation element(近似要素)を導入しました。

```go
type Number interface { 
    ~int | ~int32 | ~int64 | ~float32 | ~float64
}
```

このように定義すると、「`int, int32, int64, float32, float64`のうちいずれかをunderlying typeとする型」すべてが`Number`を実装するようになります。
(underlying typeについては次の節で説明します。)

これにより、次のようなコードが書けるようになります。

```go
type Number interface { 
    ~int | ~int32 | ~int64 | ~float32 | ~float64
}

func Max[T Number] (x, y T) T {
	if x >= y {
		return x
	}
	return y
}

var x y NewInt = 1, 2

max := Max(x, y) // max == NewInt(2)
```

# underlying type

Go言語の全ての型は、それに対応する"underlying type"という型を持っています。

1つの型に対して、対応するunderlying typeは必ず1つだけ存在します。underlying typeを持たない型や、underlying typeを2つ以上持つ型は存在しません。

## 具体例

まず具体例を見てみます。

```go
type NewInt int // NewIntのunderlying typeはint

type NewNewInt NewInt // NewNewIntのunderlying typeもint

// intのunderlying typeはint

type IntSlice []int // IntSliceのunderlying typeは[]int

// []intのunderlying typeは[]int
```

大まかにいうと、`type A B`という形の型定義を左から右に遡ってゆき、それ以上遡れないところにある型がunderlying typeです。

## 厳密な定義

https://go.dev/ref/spec#Types によると、

> Each type T has an underlying type: If T is one of the predeclared boolean, numeric, or string types, or a type literal, the corresponding underlying type is T itself. Otherwise, T's underlying type is the underlying type of the type to which T refers in its type declaration.

とあります。つまり、

- `T`が事前宣言されたboolean, 数値, 文字列型や型リテラルの場合は`T`のunderlying typeは`T`自身
- それ以外の場合、(`T`は`type T X`のように定義された型なので)`T`のunderlying typeは`X`のunderlying type

のように再帰的な定義になっています。

より丁寧な解説を見たいかたは、DQNEOさんによる次の発表を見るのが良いと思います。

- https://www.youtube.com/watch?v=mlg1Scnm44Q&t=3148s
- [上記発表のスライド](https://speakerdeck.com/dqneo/go-language-underlying-type)

# `constraints`パッケージと`unions`

`<, >`で比較可能な型を`unions`で列挙できることは分かりましたが、実際に全ての方を書こうとすると面倒だなと思われた方もいると思います。

そこで、順序付できるとか、足し算ができるなどの基本的な型制約は標準パッケージ`constraints`で提供されることになりました。

https://github.com/golang/go/blob/master/src/constraints/constraints.go

例えは、順序付できる型を表すインタフェースは`constraints.Ordered`です:

```go
// Ordered is a constraint that permits any ordered type: any type
// that supports the operators < <= >= >.
// If future releases of Go add new ordered types,
// this constraint will be modified to include them.
type Ordered interface {
	Integer | Float | ~string
}
```

`Integer`, `Float`も同じ`constraings`パッケージで定義されているインタフェースです。

ここで`unions`の要素として別なインタフェース型が初めて出てきましたね。

`Integer`を`unions`の一部として使った場合どういう意味になるかというと、`Integer`の`unions`に列挙されている型を`Integer`の代わりに書いたのと同じ意味になります。`Float`も同様です。
従って、`Ordered`は次のように書いても同じ意味です。

```go
type Ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | 
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
	~float32 | ~float64 | ~string
}
```

これを使って、一般的な`Max`関数を定義できます。

https://gotipplay.golang.org/p/fBQ5QUeLFb3

```go
type NewInt int

var x, y NewInt = 3, 2

func main() {
	fmt.Println(Max(x, y)) // 3
}

func Max[T constraints.Ordered](x, y T) T {
	if y > x {
		return y
	}
	return x
}
```

## `unions`の要素としてどんな型でも書いていいのか

`unions`が複数要素からなるとき、その要素になれるのは

- 非インタフェース型
- メソッド定義を含まないインタフェース型

です。つまり、`fmt.Stringer`のように`String() string`というメソッド定義を含んでいる型は複数要素からなる`unions`の要素とすることができません。

:::message
厳密にいうと、

- メソッド定義を含むインタフェース型は許可されない
- 許可されないインタフェース型を埋め込んだインタフェースは許可されない

となります。例えば次のインタフェースも複数要素`unions`の要素になれません。

```go
type I interface { // 許可されないインタフェースを埋め込んだインタフェースなので許可されない
	fmt.Stringer // メソッド定義を含むインタフェースなので許可されない
}
```
:::

:::message
ここで「複数要素の」と断ったのは、単一要素、つまり`|`を含まない`unions`にインタフェース型を使うのは、従来からあるインタフェース型の「埋め込み」と同じことだからです。

```go
type I interface {
	fmt.Stringer // 単一要素のunionsにインタフェースを使うのは、インタフェースの埋め込みと同じこと
}
```
:::

:::message
このような制限を設けている理由は、これを許可するとコンパイラの実装が複雑になるわりにそれほど有用性はないからだと考えられます。
ただし、実装が不可能と考えられているわけではないようで、もしこのような定義が必要だと将来判断されれば許可されるようになる可能性もあります。

この点については筆者資料の[Type Sets Proposalを読む(2)](https://zenn.dev/nobishii/articles/type_set_proposal_2)で議論しています。
:::

# structralなインタフェースとそのstructural type 

さて、前章で扱った`for range`ループ

https://gospec-previewer.vercel.app/refs/0bacee18fda5733fe0bcf5c15e095f16abce3252#For_statements

> The expression on the right in the "range" clause is called the range expression, which may be an array, pointer to an array, slice, string, map, or channel permitting receive operations. The range expression may also be of type parameter type with a structural constraint in which case the rules below consider the constraint's structural type as the type of the range expression.

