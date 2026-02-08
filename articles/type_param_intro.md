---
title: "Go言語のジェネリクス入門" # 記事のタイトル
emoji: "😸" # アイキャッチとして使われる絵文字（1文字だけ）
type: "tech" # tech: 技術記事 / idea: アイデア記事
topics: ["go"] # タグ。["markdown", "rust", "aws"]のように指定する
published: true # 公開設定（falseにすると下書き）
---

Go1.18は2022年3月にリリースされました。このリリースはGo言語へのジェネリクスの実装を含んでいます。
この記事ではできるだけ最新の仕様と用語法にもとづいてジェネリクスの言語仕様について解説していきます。

:::message
この記事のタイトルは最初の投稿時のまま「入門」となっているのですが、元々少し進んだ内容も書いていた上に、更新に伴って進んだ内容が増えてきています。

本当に典型的な使用法と、基本的な考え方を把握したい方は、記事の序盤だけを読むのが良いと思います。
:::

## 更新履歴

- 2026/02/xx [Go1.25](https://go.dev/doc/go1.25)で言語仕様書から"core type"の用語が廃止されたことに対応し、関連箇所を大きく加筆しました。
- 2024/01/03: [Go1.21(2023-08-08)](https://go.dev/doc/devel/release#go1.21.0)で`cmp`パッケージが標準ライブラリに追加されたことに対応しました。
- 2023/02/23: [Go1.20(2023-02-01)](https://go.dev/doc/devel/release#go1.20)の[`comparable`の仕様変更](https://golang.org/doc/go1.20#language)に対応しました。
  - 次の関連資料があります:
    - [The Go Blog - All your comparable types](https://go.dev/blog/comparable) Griesemer氏によるGo公式ブログです。
	- [Go言語のBasic Interfaceはcomparableを満たすようになる(でも実装するようにはならない)](https://zenn.dev/nobishii/articles/basic-interface-is-comparable) 上記の内容に対する筆者の解説記事です。Go1.20リリース前に書いたので用語が使えてないところがあります。

**シリーズ**

| タイトル | 内容 | 
| ---- | ---- |
| Go言語のジェネリクス入門(1) | この記事です。基本的なジェネリクスの使用法とunion, ~について説明します。 |
| [Go言語のジェネリクス入門(2) インスタンス化と型推論](https://zenn.dev/nobishii/articles/type_param_intro_2) | この記事の続編です。インスタンス化と型推論、そこで使われるunificationというルーチンについてできるだけ厳密に説明します。 |

## 章目次

実用上は最初の「基本原則とシンプルな例」というセクションの内容で十分なことが多いと思います。とりあえずここだけ読むのをおすすめします。

| タイトル | 内容 | 
| ---- | ---- |
| 基本原則とシンプルな例 | Goジェネリクスの基本原則と典型的な例を解説します。 |
| `unions` | メソッドの実装以外の性質をジェネリックに扱いたい場合に使える機能を解説します。|
| `~`とunderlying type | `unions`をさらに使いこなすための文法`~`を解説します。|
| core type | 言語仕様書読み込み勢(?)向けです。続編の前提知識になります。|

# 基本原則とシンプルな例

Goジェネリクスの基本原則とシンプルな例を説明します。シンプルな例と言っても、ユースケースの大半はこれで尽くされると思いますので、この節だけ読んで終わりにするのもおすすめです。

## Goのジェネリクスの基本原則

Goのジェネリクスの基本事項については[Type Parameters Proposal](https://go.googlesource.com/proposal/+/refs/heads/master/design/43651-type-parameters.md)の冒頭に挙げられています。このうち特に重要なのは次の2つです。この2つを覚えればGoのジェネリクスを十分に使うことができると思います。

- 「関数」と「型」は「型パラメータ」を持つことができる。
- 「型パラメータ」の満たすべき性質は「インタフェース型」を「型制約」として使うことで表す。

## 具体例1: 型パラメータを持つ関数`f[T Stringer]`

まず「型パラメータを持つ関数」の具体例を見てみましょう。

https://gotipplay.golang.org/p/NWxONCa85DL

```go
func main() {
	fmt.Println(f([]MyInt{1, 2, 3, 4}))
    // Output:
    // [1 2 3 4]
}

// fは型パラメータを持つ関数
// Tは型パラメータ
// インタフェースStringerは、Tに対する型制約として使われている
func f[T Stringer](xs []T) []string {
	var result []string
	for _, x := range xs {
        // xは型制約StringerによりString()メソッドが使える
		result = append(result, x.String())
	}
	return result
}

type Stringer interface {
	String() string
}

type MyInt int

// MyIntはStringerを実装する
func (i MyInt) String() string {
	return strconv.Itoa(int(i))
}
```

関数`f`の宣言時に`f[T Stringer]`という四角カッコの文法要素がついていますね。これが型パラメータと一緒に導入される新しい文法です。この意味は、

- 関数`f`において型パラメータ`T`を宣言する
- `T`型はインタフェース`Stringer`を満たす型である、という型制約を設ける

という意味です。このように宣言した型パラメータは関数の他の部分で参照できます。例えば引数の型として`T`を使うことができます。よって、`f[T Stringer](xs []T)`というのは、引数`xs`として型`T`のスライス型`[]T`を受け取る、という意味になります。

## 具体例2: 型パラメータを持つ型`Stack[T any]`

関数だけでなく、型も「型パラメータ」を持つことができます。

一例として、データ構造「スタック」を実装してみます。

https://gotipplay.golang.org/p/jCS7vhCe_XC

```go
type Stack[T any] []T

func New[T any]() *Stack[T] {
	v := make(Stack[T], 0)
	return &v
}

func (s *Stack[T]) Push(x T) {
	(*s) = append((*s), x)
}

func (s *Stack[T]) Pop() T {
	v := (*s)[len(*s)-1]
	(*s) = (*s)[:len(*s)-1]
	return v
}

func main() {
	s := New[string]()
	s.Push("hello")
	s.Push("world")
	fmt.Println(s.Pop()) // world
	fmt.Println(s.Pop()) // hello
}
```

:::message

この`Stack[T]`は`Pop`したあともstackインスタンスが生きているとメモリリークしますが、簡単のためこのままにします。

:::

まず型定義において、`type Stack[T any] []T`としています。もし`string`に限定したスタックであれば`type Stack []string`と定義するところです。この`string`の部分をパラメータ化するために`[T any]`を追加したわけです。

ここで、`any`は新しく導入される識別子で、空インタフェース`interface{}`の別名です。`any`を書けるところには代わりに`interface{}`を書いても構いませんし、その逆もOKです。`Stack`の内容になる要素型は何の型であっても良いですから、`any`を型制約にするのが適切です。

:::message
ちなみに、`any`は「事前宣言された識別子(predeclared identifier)」であって「予約語(keyword)」ではありません。
なので、`any := 1`のように同じ名前の識別子で新たに変数定義することもできます。
:::

次にコンストラクタである`New`関数を見てみます。型自体がパラメータ化されているので、コンストラクタも型パラメータを持つ関数としています。

Stackはメソッド`Push`と`Pop`を持ちます。型パラメータを持つ型に対してメソッドを宣言するときは、次のような構文を使います。

```go
func(s *Stack[T]) Push(x T)
```

`*`とポインタにしてあるのはポインタレシーバにするためで、これは従来通りの文法です。少し覚えにくいのはレシーバの型を`Stack[T]`のようにして型パラメータをつける必要があるところです。この`T`をメソッド内の別な場所で参照することができます。`Push`の場合は引数の型として`(x T)`と使っていますね。

最後に`main`を見てみましょう。

```go
	s := New[string]()
```

という行がありますね。関数宣言ではなく、関数呼び出しの方に`[string]`がついています。これは型引数(type argument)で、型パラメータに具体的な型を渡すための構文です。

型引数は通常は省略可能です。それは、関数の引数の型と型パラメータをマッチングさせて型パラメータを**型推論**できる場合が多いからです。しかし、`New`関数には引数がないため、具体的な型引数を渡さないと型推論ができず、コンパイルが失敗します。



### メソッド宣言において新たな型パラメータは宣言できない

メソッドも関数の一種なので、メソッド宣言時に新たな型パラメータを宣言できるのかという疑問が湧くかもしれません。これはできないことになっています。

例えば次のようなコードは書けません。

```go
// これは書けない
func (s *Stack[T]) ZipWith[S,U any](x *Stack[S], func(T, S) U) *Stack[U] {
    // ...
}
```

こういうことをしたければメソッドではない関数として定義すべきです。

```go
// これは書ける
func ZipWith[S,T,U any](x *Stack[T], y *Stack[S], func(T, S) U) *Stack[U] {
    // ...
}
```

:::message

具体的な実装例はこちらにあります: https://gotipplay.golang.org/p/-_HxaTjE_Zi

:::

## 具体例3: 型パラメータを持つ型`Set[T comparable]`

次に、いわゆるSet型を定義してみましょう。

https://gotipplay.golang.org/p/ht_akn1eCGy

```go
type Set[T comparable] map[T]struct{}

func New[T comparable](xs ...T) Set[T] {
	s := make(Set[T])
	for _, xs := range xs {
		s.Add(xs)
	}
	return s
}

func (s Set[T]) Add(x T) {
	s[x] = struct{}{}
}

func (s Set[T]) Includes(x T) bool {
	_, ok := s[x]
	return ok
}

func (s Set[T]) Remove(x T) {
	delete(s, x)
}

func main() {
	s := New(1, 2, 3)
	s.Add(5)
	fmt.Println(s.Includes(3)) // true
	s.Remove(3)
	fmt.Println(s.Includes(3)) // false
}
```

型定義に注目してください。

```go
type Set[T comparable] map[T]struct{}
```

ここで、`comparable`という新しいインタフェース型が型制約に使われています。なぜ`any`ではダメなのでしょうか？

それは、`T`を`map`のkeyとして使いたいからです。`map`はkeyの値に重複がないように値を保管していくデータ構造なので、重複しているかどうかを判定できる必要があります。その判定には`==`及び`!=`演算子による比較が用いられます。Go言語ではこの2つの演算子により比較できる型と比較できない型があるため、「比較可能なすべての型により満たされるインタフェース」が必要なのです。

しかしそのようなインタフェースをユーザが定義することはできません。そこでGo言語は`comparable`というインタフェースを予め定義されたものとして提供することにしました。これを使えば、genericに使える`Set`型を簡単に作ることができます。

## Go1.17でできなかったこと

ここで型パラメータのモチベーションを知るために、Go1.17でのコードを考えてみます。

### インタフェース型のスライスを受け取る関数

まず、Go1.17において次のインタフェースと関数を考えます。

```go
type Stringer interface {
    String() string
}

func f(xs []Stringer) []string {
    var result []string
    for _, x := range xs {
        result = append(result, x.String())
    }
    return result
}
```

また、次のように`Stringer`を実装する型を用意します。

```go
type MyInt int

// MyIntはStringerを実装する
func(i MyInt) String() string {
    return strconv.Itoa(int(i))
}
```

このとき次のように、`MyInt`のスライスを`f`に渡すことはできるでしょうか？

```go
xs := []MyInt{0,1,2}
f(xs) // fは[]Stringerを受け付ける
```

このようなコードは書けません。`MyInt`は`Stringer`を満たすので`MyInt`型の値は`Stringer`型の変数に代入可能ですが、`[]MyInt`型の値は`[]Stringer`型の変数に代入できないためです。

Go1.17で`[]Stringer`を一般的に扱う関数を書くには、次の`f2`のように空インタフェース型`interface{}`を受け取るようにするしかありませんでした。この関数`f2`にはどんな型の値でも渡せてしまうので、関数の利用側で間違った値を渡さないように気をつけなければいけません。

```go
// 【注意】 Stringerを実装する型Tのスライス[]Tだけを渡すこと
func f2(xs interface{}) {
    if vs,ok := xs.([]MyInt); ok {
        // vsに関する処理
    }
    // ... 
}
```

:::message
`f2`が`[]MyInt`以外のスライス型を受け取るようにするには、それぞれの型についての[型アサーション](https://go.dev/ref/spec##Type_assertions)を書く必要があります。

```go
if vs, ok := xs.([]Stringer); ok
```

のようなアサーションを書くこと自体はできますが、こう書いても`[]MyInt`型の値を渡したときには`!ok`となります。

型スイッチ文を使う場合も、渡すかもしれない具体的な型ごとにcase節が必要です。
:::

### 型パラメータによる記述

> Stringerインタフェースを実装する型Tのスライス[]Tだけを渡すこと

という条件付けは、型パラメータを使うと次のように記述できます。

```go
// fは型パラメータを持つ関数
// Tは型パラメータ
// インタフェースStringerは、Tに対する型制約として使われている
func f[T Stringer](xs []T) []string {
	var result []string
	for _, x := range xs {
        // xは型制約StringerによりString()メソッドが使える
		result = append(result, x.String())
	}
	return result
}
```

この`f`には`[]MyInt`型の値だけでなく、何のコードの変更もなしに`Stringer`を実装する型`T`のスライス`[]T`を渡せます。また、そうでない型の値を渡した場合はコンパイルエラーとして検出できます。

## まとめ

- 「関数」と「型」は「型パラメータ」を持つことができる。
- 「型パラメータ」の満たすべき性質は「インタフェース型」を「型制約」として使うことで表す。
- 関数や型の後に`[T constraint]`という文法要素をつけると、「型パラメータTを宣言する。Tは`constraint`を満たさなければならない」という意味になる。`constraint`は型制約と呼ばれ、インタフェース型を用いる。
- メソッドに追加で型パラメータを宣言することはできない。
- `any`インタフェースは空インタフェース`interface{}`の別名である
- 比較可能、つまり`==, !=`による等値判定が可能な型により満たされる`comparable`が提供される。
- 型パラメータの重要な使い方の1つは、スライスやマップなどのいわゆるコレクション型の抽象化である。

# `unions`

Go1.18では`unions`という文法要素を使って、従来のインタフェースでは表現できない型制約を定義することができます。

## genericなMax関数とunions

標準パッケージの`math.Max`関数は`func Max(x, y float64) float64`というシグネチャを持ち、`float64`の値しか渡すことができません。

せっかく型パラメータが使えるようになるので、genericなMax関数を作ってみたいと思います。まず初めに次のようなコードを考えました。

```go
func Max(T any) (x, y T) T {
	if x >= y {
		return x
	}
	return y
}
```

ところが、このコードは動作しません。実行すると次のエラーメッセージが出力されます。`T`の型制約は`any`なので、演算子`>=`で比較できるとは限らないからです。


```
invalid operation: cannot compare x >= y (operator >= not defined on T)
```

> 無効な演算: `x >= y`という比較はできません。(演算子 >= は 型Tで定義されていません)

それでは、適当なインタフェース型を定義して演算子`>=`で比較できるような型制約にすることはできるでしょうか？

Go1.17までは、できませんでした。なぜなら、Go1.17までのインタフェース型とは「メソッドセット」すなわちメソッドの集合（集まり）を定義するものであって、「ある演算子が使える」というようなメソッド以外の型の性質を表すことはできないからです。

そこでGo言語は、「インタフェース型」として次のようなものも定義できるように機能を拡張することにしました。

```go
type Number interface {
    int | int32 | int64 | float32 | float64
}
```

この`Number`というインタフェースは、`int, int32, int64, float32, float64`という5種類の型によって **「満たされ」ます**。かつ、これ以外の型によっては満たされません。
この文法要素`int | int32 | int64 | float32 | float64`のことを`unions`や`union element`と呼びます。

:::message

`|`を使わずに一つだけの型を書けば、その**一つの型によってのみ満たされるインタフェース**を定義できます。

```go
type Int interface {
    int
}
```

この`Int`インタフェースを実装するのは`int`型のみです。

:::

大切なことは、`Number`を実装する全ての型は、演算子`>=`をサポートしていることです。これにより、次のような関数を書くことができます。

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

::: message

`>=`などで順序づけられる型はこの5つ以外にもありますが、全て書き出すと大変なため5つだけ書きました。

:::

## まとめ

- Goの型パラメータは型制約をインタフェース型によって表現するが、型の性質には「メソッドを持つ」以外の性質もある。その性質の一部は`unions`を利用した新しいインタフェース型によって表現できる。
- `<, >, <=, >=`による順序付可能性は`unions`を使って順序づけられる型のみを列挙することで表現できる。
- `==, !=`による比較可能性は`comparable`インタフェースで表現する(再掲)。

# `~`とunderlying type

## モチベーション

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

## `~`をつかってunderlying typeをマッチングする

`NewInt`や`NewNewInt`も数値型であることに変わりはなく、`>=`などの演算子で比較することができるのですから、このような型を許すインタフェースを作りたいです。

もちろん、`NewInt`を直接unionsに加えれば`NewInt`に`Number`を実装させることはできます:

```go
// NewIntとNewNewIntがNumberを実装するようになった
type Number interface { 
    int | int32 | int64 | float32 | float64 | NewInt | NewNewInt
}
```

しかし、「`int`を元にして型定義で作られる新しい型」は無限にあるので、それら全てが`Number`を実装するようにしたいです。そのための文法として、Go言語は`~`を導入しました。

```go
type Number interface { 
    ~int | ~int32 | ~int64 | ~float32 | ~float64
}
```

このように定義すると、「`int, int32, int64, float32, float64`のうちいずれかをunderlying typeとする型」すべてが`Number`を実装するようになります。

::: message

underlying typeについては次の節で説明します。

:::

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

:::message

Proposal段階では`~int`のような要素は近似要素(approximation element)と呼ばれていました。

これはつまり`type MyInt int`のような型は`int`に「似ている」型であるという気持ちが込められていると思います。
※物理学や工学では`x ~ 1000`のような記法で「`x`は1000に近い」という意味を表したりすることがあります。

:::

## underlying type

Go言語の全ての型は、それに対応する"underlying type"という型を持っています。

1つの型に対して、対応するunderlying typeは必ず1つだけ存在します。underlying typeを持たない型や、underlying typeを2つ以上持つ型は存在しません。

### 具体例

まず具体例を見てみます。

```go
type NewInt int // NewIntのunderlying typeはint

type NewNewInt NewInt // NewNewIntのunderlying typeもint

// intのunderlying typeはint

type IntSlice []int // IntSliceのunderlying typeは[]int

// []intのunderlying typeは[]int
```

大まかにいうと、`type A B`という形の型定義を左から右に遡ってゆき、それ以上遡れないところにある型がunderlying typeです。

### 厳密な定義(Go 1.17)

https://go.dev/ref/spec##Types によると、

> Each type T has an underlying type: If T is one of the predeclared boolean, numeric, or string types, or a type literal, the corresponding underlying type is T itself. Otherwise, T's underlying type is the underlying type of the type to which T refers in its type declaration.

とあります。つまり、

- `T`が事前宣言されたboolean, 数値, 文字列型や型リテラルのとき、`T`のunderlying typeは`T`自身である
- それ以外の場合、(`T`は`type T X`のように定義された型なので)`T`のunderlying typeは`X`のunderlying typeである

のように再帰的な定義になっています。

::: message
なお、`T`が型パラメータ型の場合、`T`のunderlying typeその型制約のunderlying typeです。上記引用箇所はGo1.17の仕様書なのでこれが言及されていません。
:::

より丁寧な解説を見たいかたは、DQNEOさんによる次の発表を見るのが良いと思います。

- https://www.youtube.com/watch?v=mlg1Scnm44Q&t=3148s
- [上記発表のスライド](https://speakerdeck.com/dqneo/go-language-underlying-type)

## `cmp`パッケージ

`<, >`で順序づけできる型を`unions`で列挙できることは分かりましたが、実際に全ての型を書こうとすると面倒だなと思われた方もいると思います。

そこで、`<, >`で順序づけできる型によって満たされる型制約は標準パッケージ`cmp`の[`cmp.Ordered`](https://pkg.go.dev/cmp#Ordered)として提供されています。

実装は次のようになっています。

```go
// Ordered is a constraint that permits any ordered type: any type
// that supports the operators < <= >= >.
// If future releases of Go add new ordered types,
// this constraint will be modified to include them.
//
// Note that floating-point types may contain NaN ("not-a-number") values.
// An operator such as == or < will always report false when
// comparing a NaN value with any other value, NaN or not.
// See the [Compare] function for a consistent way to compare NaN values.
type Ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64 |
		~string
}
```

これを使って、一般的な`Max`関数を定義できます。

https://go.dev/play/p/-WB97e8w2NC

```go
package main

import (
	"cmp"
	"fmt"
)

type NewInt int

var x, y NewInt = 3, 2

func main() {
	fmt.Println(Max(x, y)) // 3
}

func Max[T cmp.Ordered](x, y T) T {
	if y > x {
		return y
	}
	return x
}
```

## `unions`の要素としてどんな型でも書いていいのか

細かい話になりますが、`unions`の要素として使える型については、少し制約があります。

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

## まとめ

- `~`をつかうと型定義によって作りうる無限の型にインタフェースを実装させることができる
- `~T`は`T`をunderlying typeに持つすべての型を表す

# できないこと(1): メソッドは独自の型パラメータを持つことができない

Go1.25現在のジェネリクスでできないことの一つに、メソッドに独自の型パラメータを定義することがあります。例えば、次のようなコードを書けません。

```go
type S struct { … }  
func (*S) m[P any](x P) { … }
```

これができないのはいくつかの点で不便です。メソッドチェーンのようなAPIを作れなかったり、型によるコードのグループ化の妨げになることがあるからです。

この制約は将来解除される可能性が高いと思います。次のプロポーザルで検討が進んでいるからです。ただし、2026年2月現在では未承認のプロポーザルなので、まだどうなるかはわからないことに留意してください。

https://github.com/golang/go/issues/77273

# できないこと(2)

このセクションでは、その他の「できそうでできないこと」を説明しましょう。ただ列挙しても良いのですが、より良い理解のため、 **Goのジェネリクスで「できそうなこと」とはそもそも何なのかを考えてみます。** 

そのために、Goのジェネリクスがどういうものだったかおさらいします。ジェネリックな関数の場合で考えると、型制約として渡されたインターフェースで宣言されているメソッドは、関数のbody(実装)のなかであたかも型パラメータ型のメソッドであるかのように使って良いのでした。これを示すため、この記事の最初のサンプルコードを再掲します:

```go
// fは型パラメータを持つ関数
// Tは型パラメータ
// インタフェースStringerは、Tに対する型制約として使われている
func f[T Stringer](xs []T) []string {
	var result []string
	for _, x := range xs {
        // xは型制約StringerによりString()メソッドが使える
		result = append(result, x.String())
	}
	return result
}

type Stringer interface {
	String() string
}

type MyInt int

// MyIntはStringerを満たす
func (i MyInt) String() string {
	return strconv.Itoa(int(i))
}
```

言い換えると、 **「型制約を満たすすべての型について`String()`が使えるならば、型パラメータ`T`に対しても`String()`が使える」** というのがGoのジェネリック関数だと言っても良さそうです。

もしも、この文を一般化した **「型制約を満たすすべての型について操作Xが可能ならば、型パラメータ`T`に対しても操作Xが可能である」** というテーゼが成り立つなら非常に分かりやすく、ある意味で理想的です。Goのジェネリクスに対して「できそうなこと」だとプログラマーが期待することの多くは、このテーゼが成り立つという期待に基づいていると思います。そこでこの記事では、これをジェネリクスの **「理想のテーゼ」** と呼ぶことにします。

しかし、2026年2月(Go1.25)現在、この「理想のテーゼ」は必ずしも成り立ちません。"操作X"に様々なものを当てはめて、それをみていきましょう。


:::message
Goの公式ブログ https://go.dev/blog/coretypes においても、元々のジェネリクスの設計思想は次のようなものであった、と書かれています。

> an operation involving operands of generic type should be valid if it is valid for any type permitted by the respective type constraint.

> ジェネリックな型のオペランドに係る演算は、その演算が、対応する型制約によって許される全ての型に対して有効ならば、有効であるべきだ。
:::

以下、これらを具体的にみていきますが、入門段階では細かすぎる仕様だと思うので、気になったときに参照する程度でご利用いただければ良いと思います。

## 「理想のテーゼ」が成り立つ操作

次の表のそれぞれの行にある「操作X」については、 **「型制約を満たすすべての型について操作Xが可能ならば、型パラメータ`T`に対しても操作Xが可能である」** というテーゼが成り立ちます。

| 操作X | 
| ---- | 
| `nil`の代入 |
| 型リテラルの値の代入 |
| Representability(表現可能性) |
| 算術演算 |
| 比較演算(`==, !=`)と順序演算(`<=`など) |
| チャネル受信演算 |
| 型変換 |
| `clear`関数の適用 |
| `len`,`cap`関数の適用 |
| `unsafe.Pointer`と`uintptr`の間の型変換 |


### `nil`の代入

型制約`Constraint`を満たすすべての型について`nil`を代入できるならば、型パラメータ`T`の変数にも`nil`を代入可能です。

よって、次のコードはコンパイルできます。

```go
type Constraint interface {
	[]int | ~[]string | *bool
}

func f[T Constraint]() {
	var _ T = nil
}
```
https://go.dev/play/p/cH5ktnw3Yck


:::message
仕様上の根拠は https://go.dev/ref/spec#Assignability にあります。
:::

### 型リテラルの値の代入

型制約`Constraint`を満たすすべての型について型リテラルで表される型`V`の値が代入可能ならば、型パラメータ`T`の変数にも`V`の値を代入可能です。

よって、次のコードはコンパイルできます。

```go
type Constraint interface {
	DefinedIntSliceA | DefinedIntSliceB
}

type DefinedIntSliceA []int
type DefinedIntSliceB []int

func f[T Constraint]() {
	var _ T = []int{}
}
```
https://go.dev/play/p/R21YGC2O8nV

型制約`Constraint`を満たすすべての型について、その値が型リテラルで表される型`V`の変数に代入可能ならば、型パラメータ`T`の値は`V`の変数に代入可能です。

よって、次のコードはコンパイルできます。

```go
type Constraint interface {
	DefinedIntSliceA | DefinedIntSliceB
}

type DefinedIntSliceA []int
type DefinedIntSliceB []int

func f[T Constraint]() {
	var t T
	var _ []int = t
}
```

https://go.dev/play/p/8ifyI02FINX

### Representability(表現可能性)

型制約`Constraint`を満たすすべての型によって、ある型なし定数`c`が表現可能ならば、その型なし定数`c`は型パラメータ`T`によって表現可能です。

よって、次のコードはコンパイルできます。

```go
type Constraint interface {
	complex128 | float64
}

func f[T Constraint]() {
	const c = 1.1
	var _ T = c // 表現可能なので代入可能である
}
```
https://go.dev/play/p/FJO4JhKl09x

### 算術演算

型制約`Constraint`を満たすすべての型についてある算術演算が可能ならば、型パラメータ`T`の値に対してもその算術演算が可能です。

よって、次のコードはコンパイルできます。

```go
type Constraint interface {
	int | float32 | float64
}

func f[T Constraint](t1, t2 T) T {
	return t1 * t2
}
```

https://go.dev/play/p/BDeKlBse44u

:::message
これはこの記事の前半ですでに扱った内容ですね。

言語仕様上の根拠は次の箇所にあります。

https://go.dev/ref/spec#Arithmetic_operators
:::

### 比較演算(`==, !=`)と順序演算(`<=`など)

型制約`Constraint`を満たすすべての型が比較可能ならば、型パラメータ`T`の値もその比較可能です。

よって、次のコードはコンパイルできます。

```go
package main

func f[T comparable](t1, t2 T) bool { return t1 == t2 }

func main() {
	var x1, x2 any
	f(x1, x2)
}
```
https://go.dev/play/p/wagDyQk6xRp

:::message
比較演算(`==, !=`)とインタフェース型に関して少し難解な仕様があります。

正確に知りたい方は、次の資料があります。

- https://go.dev/blog/comparable 公式ブログ、英語
- https://zenn.dev/nobishii/articles/basic-interface-is-comparable 

:::

型制約`Constraint`を満たすすべての型が`<`などで順序づけできるならば、`T`の値に対しても順序づけできます。

よって、次のコードはコンパイルできます。

```go
type Constraint interface {
	int | string | float32
}

func f[T Constraint](t1, t2 T) bool { return t1 < t2 }
```

https://go.dev/play/p/JqPmRpRYgkN
:::message

言語仕様上の根拠は次の箇所にあります。
https://go.dev/ref/spec#Comparison_operators
:::

### チャネル受信演算

型制約`Constraint`を満たすすべての型が、型`S`の値を受信できるチャネル型ならば、型パラメータ`T`の値からも型`S`の値を受信できます。

よって、次のコードはコンパイルできます。

```go
// 型Sはこの場合intに相当する
type MyChanInt <-chan int

type Constraint interface {
	chan int | <-chan int | MyChanInt
}

func f[T Constraint](t T) int {
	return <-t
}
```
https://go.dev/play/p/YKXhTLD6Uwy

:::message

言語仕様上の根拠は次の箇所にあります。
https://go.dev/ref/spec#Receive_operator
:::

### 型変換

型変換については3つのパターンを説明します。

型制約`Constraint`を満たすすべての型`V`について、型`V`から別な型`W`への型変換ができるならば、型パラメータ`T`から`W`への型変換が可能です。

よって、次のコードはコンパイルできます。

```go
type Constraint interface {
	int | int32 | int64 // 全てfloat64への型変換が可能
}

func f[T Constraint](t T) {
	var _ = float64(t)
}
```
https://go.dev/play/p/CGVytLgeERL

型制約`Constraint`を満たすすべての型`V`について、別な型`W`から`V`への型変換ができるならば、`W`から型パラメータ`T`への型変換が可能です。

よって、次のコードはコンパイルできます。
```go
type Constraint interface {
	int | int32 | int64 // 全てfloat64からの型変換が可能
}

func f[T Constraint]() {
	var v float64 = 1.1
	var _ = T(v)
}
```
https://go.dev/play/p/EghdRMpCycD

型制約`Constraint1`を満たす全ての型`V`と、型制約`Constraint2`を満たす全ての型`W`について、`V`から`W`への型変換が可能ならば、それぞれ対応する型パラメータ`T1`から`T2`への型変換が可能です。

よって、次のコードはコンパイルできます。
```go
type Constraint1 interface {
	int | int32 | int64 // 全てfloat32, float64への型変換が可能
}

type Constraint2 interface {
	float32 | float64
}

func f[T1 Constraint1, T2 Constraint2](t1 T1) T2 {
	return T2(t1)
}
```
https://go.dev/play/p/u01fTfEmoJc

:::message
言語仕様上の根拠は次の箇所にあります。
https://go.dev/ref/spec#Conversions
:::

### clear関数の適用

型制約`Constraint`を満たす全ての型について、`clear`関数の適用が可能ならば、型パラメータ`T`の値に対しても`clear`関数の適用が可能です。

よって、次のコードはコンパイルできます。

```go
type Constraint interface {
	[]int | map[int]int | []string
}

func f[T Constraint](t T) {
	clear(t)
}
```
https://go.dev/play/p/pzsv02pBSaH
:::message
言語仕様上の根拠は次の箇所にあります。
https://go.dev/ref/spec#Clear
:::

### `len`,`cap`関数の適用

型制約`Constraint`を満たす全ての型について、`len`,`cap`関数の適用が可能ならば、型パラメータ`T`の値に対してもこれらの関数の適用が可能です。

よって、次のコードはコンパイルできます。

```go
type MyInt int

type Constraint interface {
	map[int]bool | map[MyInt]bool | []MyInt
}

func f[T Constraint](m T) int {
	return len(m)
}
```
https://go.dev/play/p/ZO8mpMzLukQ
:::message
言語仕様上の根拠は次の箇所にあります。
https://go.dev/ref/spec#Length_and_capacity
:::

### `unsafe.Pointer`と`uintptr`の間の型変換

- 型制約`Constraint`を満たす全ての型について、`unsafe.Pointer`への型変換ができるならば、型パラメータ`T`の値も`unsafe.Pointer`への型変換ができます。
- 型制約`Constraint`を満たす全ての型について、`uintptr`への型変換ができるならば、型パラメータ`T`の値も`uintptr`への型変換ができます。

よって、次のコードはコンパイルできます。

```go
import "unsafe"

type MyUintPtr uintptr

type Constraint interface {
	uintptr | MyUintPtr
}

func f[T Constraint](ptr T) unsafe.Pointer {
	return unsafe.Pointer(ptr)
}
```

https://go.dev/play/p/_INs7vJ5TKb
:::message
言語仕様上の根拠は次の箇所にあります。
https://go.dev/ref/spec#Package_unsafe
:::


## 「理想のテーゼ」が成り立たない操作 = できそうでできないこと

次の表のそれぞれの行にある「操作X」については、 **「型制約を満たすすべての型について操作Xが可能ならば、型パラメータ`T`に対しても操作Xが可能である」** というテーゼが **必ずしも成り立ちません。**

| 操作X | 
| ---- | 
| フィールドの読み取り |
| 定数宣言 |
| コンポジットリテラルの使用 |
| インデックス式の使用 |
| スライス式の使用 |
| 関数呼び出し(型パラメータ型自体が関数型というケース) |
| チャネルへの送信 |
| range句を使ったfor文 |
| append関数による要素の追加 |
| channelの`close`関数 |
| 複素数に関する操作 |
| mapの`delete`関数による特定エントリーの削除 |
| `make`関数による作成 |

### フィールドの読み取り

型制約`Constraint`を満たすすべての型について、その値`x`のフィールドのセレクタ式`x.F`が有効だとしても、型パラメータ`T`の値`t`について`t.F`は有効ではありません。

よって、次のコードはコンパイルできません。

```go
type AB struct {
	A int
	B int
}
type ABC struct {
	AB
	C int
}

type Constraint interface {
	AB | ABC
}

func f[T Constraint](t T) int {
	return t.A
}

```
https://go.dev/play/p/IUcO6kAVYu3

### 定数宣言

型制約`Constraint`を満たすすべての型について、その型を持つ定数を定数式`exp`で宣言できるとしても、型パラメータ`T`の型を持つ定数を宣言することはできません。

定数宣言の型として型パラメータ`T`を使うこと自体ができないためです。

よって、次のコードはコンパイルできません。

```go
type Constraint interface {
	complex128 | float64
}

func f[T Constraint]() {
	const exp = 1.1
	const _ T = exp
}
```
https://go.dev/play/p/HKUPvDpkLmm

### コンポジットリテラルの使用

型制約`Constraint`を満たすすべての型について、その型のコンポジットリテラルが使えるとしても、型パラメータ`T`のコンポジットリテラルが使えるとは限りません。

追加条件として、`T`を満たすすべての型が、同一のunderlying typeを持つ必要があります。

よって、次のコードはコンパイルできません。

```go
type Constraint interface {
	[]int | [1]int
}

func f[T Constraint]() {
	var _ = T{}
}
```

https://go.dev/play/p/Ogs7lmQL3Cj



### インデックス式の使用

型制約`Constraint`を満たすすべての型について、その型の式からインデックス式が作れるとしても、型パラメータ`T`の式にインデックス式が使えるとは限りません。

追加条件として、`T`を満たすすべての型が、同一の要素型を持つ必要があります。

よって、次のコードはコンパイルできません。
```go
type Constraint interface {
	[1]int | [1]string // どちらもインデックス式が作れるが、要素型がintとstringで異なる
}

func f[T Constraint]() {
	var t T
	_ = t[0] // このようなインデックス式は無効
}
```
https://go.dev/play/p/G1JrWC1UQKm

:::message
言語仕様上の根拠は次の箇所にあります。
https://go.dev/ref/spec#Index_expressions

> The element types of all types in P's type set must be identical.
:::

### スライス式の使用

型制約`Constraint`を満たすすべての型について、その型の式からスライス式が作れるとしても、型パラメータ`T`の式にスライス式が使えるとは限りません。

追加条件として、`T`を満たすすべての型が同一のunderlying typeを持つ必要があります。ただし、`string`型と`[]byte`型はこのルールの適用上は同一視して良いことになっています。

よって、次のコードはコンパイルできません。
```go
type Constraint interface {
	[10]int | [11]int // どちらもインデックス式が作れるが、underlying typeが異なる
}

func f[T Constraint]() {
	var t T
	_ = t[:] // このようなスライス式は無効
}
```
https://go.dev/play/p/y0ZsHgjBtre

:::message
言語仕様上の根拠は次の箇所にあります。
https://go.dev/ref/spec#Slice_expressions

> If the operand type is a type parameter, unless its type set contains string types, all types in the type set must have the same underlying type, and the slice expression must be valid for an operand of that type. If the type set contains string types it may also contain byte slices with underlying type []byte. In this case, the slice expression must be valid for an operand of string type.

また、この例から分かるように、underlying typeが同一というのは要素型が同一であるよりも強い条件です。
:::

### 関数呼び出し(型パラメータ型自体が関数型というケース)

型制約`Constraint`を満たすすべての型について、その型が関数型であり、特定の引数`(a)`に対して関数呼び出しが可能だとしても、型パラメータ`F`の値である関数についてその呼び出しが可能だとは限りません。

追加条件として、`F`を満たすすべての型が同一のunderlying typeを持つ必要があります。

よって、次のコードはコンパイルできません。
```go
type MyIntPointer *int

type Constraint interface {
	func() *int | func() MyIntPointer
}

func f[F Constraint]() {
	var f F
	var _ *int = f() // 無効: MyIntPointerは*intに代入可能であるにもかかわらず。
}
```
https://go.dev/play/p/gUfpOnaSmKj

:::message
言語仕様上の根拠は次の箇所にあります。
https://go.dev/ref/spec#Calls

> If the type of f is a type parameter, all types in its type set must have the same underlying type, which must be a function type, and the function call must be valid for that type.
:::


### チャネルへの送信

型制約`Constraint`を満たすすべての型について、その型がチャネル型であり、ある値をその型のチャネルに送信可能だとしても、型パラメータ`T`の値であるチャネルにその値を送信可能であるとは限りません。

追加条件として、`Constraint`を満たすすべての型について、その要素型が同一でなければいけません。

よって、次のコードはコンパイルできません。
```go
type MyInt int
type MyChanInt chan<- MyInt

type Constraint interface {
	chan int | chan<- int | MyChanInt // 要素型がintとMyIntで一致しない
}

func f[T Constraint](t T) {
	t <- 1 // 無効
}
```
https://go.dev/play/p/J2wFa_fZ39n

:::message
言語仕様上の根拠は次の箇所にあります。
https://go.dev/ref/spec#Send_statements
:::

### range句を使ったfor文

型制約`Constraint`を満たすすべての型について、その型の値vを使って`for ... range v`という`for`文が使えたとしても、型パラメータ`T`の値に対して同様にrange句を使ったfor文が使えるとは限りません。

追加条件として、`Constraint`を満たす全ての型のunderlying typeが同一である必要があります。

よって、次のコードはコンパイルできません。
```go
type Constraint interface {
	string | []byte
}

func f[T Constraint](t T) {
	for _, v := range t {
	}
}
```

https://go.dev/play/p/NQsa3XWYU8M

:::message
言語仕様上の根拠は次の箇所にあります。
https://go.dev/ref/spec#For_range
:::

### append関数による要素の追加

型制約`Constraint`を満たすすべての型について、append関数である値を追加することができたとしても、型パラメータ`T`について同じことができるとは限りません。

追加条件として、`Constraint`を満たす全ての型のunderlying typeが同一である必要があります。

よって、次のコードはコンパイルできません。
```go
type MyInt int

type Constraint interface {
	[]int | []MyInt
}

func f[T Constraint](t T) {
	append(t, 1)
}
```
https://go.dev/play/p/UxrSm8UIz_o

:::message
言語仕様上の根拠は次の箇所にあります。
https://go.dev/ref/spec#Appending_and_copying_slices
:::


### channelの`close`関数

型制約`Constraint`を満たすすべての型について、その値を`close`関数に渡すことができるとしても、型パラメータ`T`の値を`close`関数に渡せるとは限りません。

追加条件として、`Constraint`を満たす全ての型の要素型が同一である必要があります。

よって、次のコードはコンパイルできません。
```go
type Constraint interface {
	chan int | chan string
}

func f[T Constraint](ch T) {
	close(ch)
}
```
https://go.dev/play/p/5huphqnb64r

:::message
言語仕様上の根拠は次の箇所にあります。
https://go.dev/ref/spec#Close
:::

### 複素数に関する操作

型制約`Constraint`を満たすすべての型について、その値を`real`,`imag`,`complex`のそれぞれの関数に渡せるとしても、型パラメータ`T`の値をこれらの関数に渡すことはできません。

これらの関数はそもそも型パラメータ型を受け取らないようになっているからです。

よって、次のコードはコンパイルできません。
```go
type Constraint interface {
	float32 | float64
}

func f[T Constraint](v T) {
	_ = complex(v, v)
}
```

https://go.dev/play/p/7PMcp7Q91oM

:::message
言語仕様上の根拠は次の箇所にあります。
https://go.dev/ref/spec#Complex_numbers
:::

### mapの`delete`関数による特定エントリーの削除

型制約`Constraint`を満たすすべての型について、その値`m`とあるキー値`k`について`delete(m,k)`によるエントリー削除ができるとしても、型パラメータ`T`の値`m`に対して`delete(m,k)`ができるとは限りません。

追加条件として、`Constarint`を満たす全ての型についてキーの型が同一である必要があります。

よって、次のコードはコンパイルできません。

```go
type MyInt int

type Constraint interface {
	map[int]bool | map[MyInt]bool
}

func f[T Constraint](m T) {
	delete(m, 1)
}
```
https://go.dev/play/p/__j2DhnYrUn

:::message
言語仕様上の根拠は次の箇所にあります。
https://go.dev/ref/spec#Deletion_of_map_elements
:::

### `make`関数による作成

型制約`Constraint`を満たすすべての型について、その値を`make`関数で作れるとしても、型パラメータ`T`についてその値を`make`関数で作れるとは限りません。

追加条件として、次のいずれかに当てはまる必要があるからです。
- `Constraint`を満たす全ての型のunderlying typeが同一のスライス型またはmap型である
- `Constraint`を満たす全ての型がチャネル型であり、その要素の型が同一で、方向が矛盾しない

よって、次のコードはコンパイルできません。
```go
type Constraint interface {
	MyChan | chan<- int
}

func f[T Constraint]() {
	_ = make(T)
}
```

https://go.dev/play/p/QTggKJPwmlW

:::message
言語仕様上の根拠は次の箇所にあります。
https://go.dev/ref/spec#Making_slices_maps_and_channels
:::



# 最後に

この記事をかくにあたり[#gospecreading](https://gospecreading.connpass.com/)から得られた理解が本質的でした。いつもありがとうございます。