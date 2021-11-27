---
title: "TBW"
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

# `constraints`パッケージ

# structralなインタフェースとそのstructural type 

https://gospec-previewer.vercel.app/refs/0bacee18fda5733fe0bcf5c15e095f16abce3252#For_statements

> The expression on the right in the "range" clause is called the range expression, which may be an array, pointer to an array, slice, string, map, or channel permitting receive operations. The range expression may also be of type parameter type with a structural constraint in which case the rules below consider the constraint's structural type as the type of the range expression.

