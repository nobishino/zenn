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

- `T`が事前宣言されたboolean, 数値, 文字列型や型リテラルのとき、`T`のunderlying typeは`T`自身である
- それ以外の場合、(`T`は`type T X`のように定義された型なので)`T`のunderlying typeは`X`のunderlying typeである

のように再帰的な定義になっています。

::: message
なお、`T`が型パラメータ型の場合も、`T`のunderlying typeは`T`自身です。上記引用箇所はGo1.17の仕様書なのでこれが言及されていません。
:::

より丁寧な解説を見たいかたは、DQNEOさんによる次の発表を見るのが良いと思います。

- https://www.youtube.com/watch?v=mlg1Scnm44Q&t=3148s
- [上記発表のスライド](https://speakerdeck.com/dqneo/go-language-underlying-type)

# `constraints`パッケージと`unions`

`<, >`で比較可能な型を`unions`で列挙できることは分かりましたが、実際に全ての型を書こうとすると面倒だなと思われた方もいると思います。

そこで、順序付けできるとか、数値型である、などの基本的な型制約は標準パッケージ`constraints`で提供されることになりました。

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

※この節の内容はスキップしても大丈夫です。

underlying typeと関連の深い新しい概念であるstructural typeについてここで説明します。
## 定義

### 型制約がstructuralである

「型制約がstructuralである」という概念を次のように定義します。

- 型制約がstructuralであるとは、次のいずれかが成り立つことをいう。
  - その型制約を満たす（実装する）全ての型のunderlying typeが同一である
  - その型制約を満たす（実装する）全ての型のunderlying typeが同一の要素型を持つchannel型であり、その中に受信専用チャネルと送信専用チャネルが混ざっていないこと

### 型制約の"structural type"

また、型制約がstructuralであるとき、その共通のunderlying typeのことをstructural typeと呼びます。そうでないとき、structural typeは存在しません。

:::message
ここでは2つの概念を定義していることに気をつけてください。

- 型制約がstructuralであるとはどういうことか
- 型制約「の」structural typeとは何か
:::
## 具体例

次の型制約はstructuralでしょうか？またその場合structural typeは何でしょうか？

```go
type C1 interface {
	~[]int
}

type C2 interface {
	int | string
}
```

- `C1`を実装する全ての型のunderlying typeは`[]int`なので`C1`はstructuralで、そのstructural typeは`[]int`です。
- `C2`を実装する型は`int`, `string`なのでunderlying typeは同一でなく`C2`はstructuralではありません。よって、`C2`のstructural typeは存在しません。

## structural typeの登場場面

このstructural typeですが、次のような場面で登場します。

- 代入可能性
- composite literals
- for range
- 制約型推論(後述)

本章では最初の3つについて説明します。

## 代入可能性

型パラメータ型の変数に対する代入を考えてみましょう。味気ない例ですが、次のコードは動作します。

https://gotipplay.golang.org/p/UJz2aE7bqH_e

```go
type C interface {
	~[]int
}

func F[T C](x T) {
	x = []int{} // 型Tの変数xに[]intの値を代入している
}
```

しかし、次のコードは動作しません。

https://gotipplay.golang.org/p/5f8DKNv6Hq_E

```go
type C interface {
	~[]int | string // stringを増やした
}

func F[T C](x T) {
	x = []int{} // 代入できない
}
```

`C`の型の「範囲」は広がっているのに代入ができなくなるのは不思議な気もします。これは次のように考えてください。

`T`と`[]int`は異なる型なので、`T`型の変数`x`に`[]int{}`を代入するためには、`T`のunderlying typeが`[]int`であることが必要だ、というのがGo1.17の仕様です。
ところが今`T`は型パラメータなので、`T`のunderlying typeは`T`自身です。
このような場合、`T`のunderlying typeの代わりに、`T`の制約のstructural typeを使う、というのが新しく追加される仕様です。

つまり、`T`の制約である`C`のstructural typeが`[]int`なので最初の例は代入できましたが、二つ目の例は`C`のstructural typeが存在しないために代入できなくなったということです。

## composite literals

型パラメータ型を使って、composite literalsを書くことができる場合があります。
その条件にもstructural typeが関係しています。

例えば次のコードは動作します。

https://gotipplay.golang.org/p/RFvZrv_hp6T

```go
type C interface { // structural type = struct { Field int }
	struct{ Field int }
}

func F[T C](T) {
	_ = T{Field: 1} // composite literalを作れる
}
```

しかし次のコードは動作しません。

https://gotipplay.golang.org/p/7UY0hO2-rlj

```go
type C interface {
	struct{ Field int } | struct { Field int `tag` }
}

func F[T C](T) {
	_ = T{Field: 1}
}
```

composite literalを作るのに使う型名が型パラメータ型の名前である場合、その制約はstructural typeを持っていなければいけません。
2つ目の例は、全く同じ構造のstruct型のunionsであるにもかかわらず、struct tagの有無によって型の同一性が満たされず、`C`のstructural typeが存在しないため、`T`のcomposite literalは作れません。

## `for range`ループ

前章`for range`ループについて扱いましたが、実は型パラメータ型に対して`for range`ループを回すためには、制約がstructural typeを持っていないといけません。

例えば次の例は動作します。

https://gotipplay.golang.org/p/ec6KpsOHgHv

```go
// 動作する例
type I interface {
	[]int 
}

func f[T I](x T) {
	for range x {
	}
}
```

しかし次の例は（どちらもスライスであるにもかかわらず）動作しません。

https://gotipplay.golang.org/p/biQ_41vglso

```go
type I interface {
	[]int | []string
}

func f[T I](x T) {
	for range x {
	}
}
```

> ./prog.go:11:12: cannot range over x (variable of type T constrained by I) (T has no structural type)