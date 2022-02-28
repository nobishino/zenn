---
title: "Go言語のジェネリクス入門(1)" # 記事のタイトル
emoji: "😸" # アイキャッチとして使われる絵文字（1文字だけ）
type: "tech" # tech: 技術記事 / idea: アイデア記事
topics: ["go"] # タグ。["markdown", "rust", "aws"]のように指定する
published: true # 公開設定（falseにすると下書き）
---

Go1.18のリリースは2022年３月の予定になっており、ジェネリクスを含む言語仕様書はかなりの頻度で加筆修正されています。
この記事ではできるだけ最新の仕様と用語法にもとづいてジェネリクスの言語仕様について解説していきます。

実用上は最初の「基本原則とシンプルな例」というセクションの内容で十分なことが多い（そうであることが望ましい）のではないかと思います。
それ以降の章は、言語仕様に関心のある人に向けた内容です。

**シリーズ**

| タイトル | 内容 | 
| ---- | ---- |
| Go言語のジェネリクス入門(1) | この記事です。基本的なジェネリクスの使用法と`|, ~`について説明します。 |
| [Go言語のジェネリクス入門(2) インスタンス化と型推論](https://zenn.dev/nobishii/articles/type_param_intro_2) | この記事の続編です。インスタンス化と型推論、そこで使われるunificationというルーチンについてできるだけ厳密に説明します。 |

# 基本原則とシンプルな例

Goジェネリクスの基本原則とシンプルな例を説明します。シンプルな例と言っても、ユースケースの大半はこれで尽くされると思いますので、この節だけ読んで終わりにするのもおすすめです。

## Goのジェネリクスの基本原則

型パラメータの基本事項については[Type Parameters Proposal](https://go.googlesource.com/proposal/+/refs/heads/master/design/43651-type-parameters.md)の冒頭に挙げられていますが、特に重要なのは次の2つです。この2つを覚えればGoのジェネリクスを十分に使うことができると思います。

- 「関数」と「型」は「型パラメータ」を持つことができる。
- 「型パラメータ」の満たすべき性質は「インタフェース型」を「型制約」として使うことで表す。

## 具体例1: 型パラメータを持つ関数`f[T Stringer]`

まず「関数」と「型」について具体例を見てみましょう。型パラメータを持つ関数の例は次のようなものです。

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
- `T`型はインタフェース`Stringer`を実装する型である、という型制約を設ける

という意味です。このように宣言した型パラメータは関数の他の部分、例えば引数の型として使うことができます。よって、`f[T Stringer](xs []T)`というのは、引数`xs`として型`T`のスライス型`[]T`を受け取る、という意味になります。

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

まず型定義において、`type Stack[T any] []T`としています。もし`string`に限定したスタックであれば`type Stack []string`と定義するところです。
この`string`の部分をパラメータ化するために`[T any]`を追加したわけです。

ここで、`any`は新しく導入される識別子で、空インタフェース`interface{}`の別名です。
`any`を書けるところには代わりに`interface{}`を書いても構いませんし、その逆もOKです。
`Stack`の内容になる要素型は何の型であっても良いですから、`any`を型制約にするのが適切です。

:::message
ちなみに、`any`は「事前宣言された識別子(predeclared identifier)」であって「予約語(keyword)」ではありません。
なので、`any := 1`のように同じ名前の識別子で新たに変数定義することもできます。
:::

次にコンストラクタである`New`関数を見てみます。型自体がパラメータ化されているので、コンストラクタも型パラメータを持つ関数としています。

Stackはメソッド`Push`と`Pop`を持ちます。
型パラメータを持つ型に対してメソッドを宣言するときは、次のような構文を使います。

```go
func(s *Stack[T]) Push(x T)
```

`*`とポインタにしてあるのはポインタレシーバにするためで、これは従来通りの文法です。少し覚えにくいのはレシーバの型を`Stack[T]`のようにして型パラメータをつける必要があるところです。
この`T`をメソッド内の別な場所で参照することができます。
`Push`の場合は引数の型として`(x T)`と使っていますね。

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
func (s Stack[T]) ZipWith[S,U any](x Stack[S], func(T, S) U) Stack[U] {
    // ...
}
```

こういうことをしたければメソッドではない関数として定義すべきです。

```go
// これは書ける
func ZipWith[S,T,U any](x Stack[T], y Stack[S], func(T, S) U) Stack[U] {
    // ...
}
```

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

ここで、`comparable`という新しいインタフェース型が型制約に使われています。
なぜ`any`ではダメなのでしょうか？

それは、`T`を`map`のkeyとして使いたいからです。`map`はkeyの値に重複がないように値を保管していくデータ構造なので、重複しているかどうかを判定できる必要があります。
その判定には`==`及び`!=`演算子による比較が用いられます。
Go言語ではこの2つの演算子により比較できる型と比較できない型があるため、「比較可能なすべての型により実装されるインタフェース」が必要なのです。

しかしそのようなインタフェースをユーザが定義することはできないため、Go言語は`comparable`というインタフェースを予め定義されたものとして提供することにしました。
これを利用することで、genericに使える`Set`型を簡単に作ることができます。

:::message

厳密には`comparable`インタフェースは言語仕様上比較可能なすべての型ではなく、「比較可能な非インタフェース型」によって実装されます。

言語仕様上の比較可能性については[言語仕様書のComparison Operators](https://go.dev/ref/spec#Comparison_operators)に詳しく書いてあります。

これは、インタフェース型の比較においてはruntime panicが発生する可能性があることを考慮したものと考えられます。
:::
## Go1.17でできなかったこと

ここで少し型パラメータのモチベーションを知るためにGo1.17でのコードを考えてみます。

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

このようなコードは書けません。`MyInt`は`Stringer`を実装するので`MyInt`型の値は`Stringer`型の変数に代入可能ですが、`[]MyInt`型の値は`[]Stringer`型の変数に代入できないためです。

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

この`f`には`[]MyInt`型の値だけでなく、何のコードの変更もなしに`Stringer`を実装する型`T`のスライス`[]T`を渡せますし、そうでない型の値は渡すことができないため、安心してプログラミングすることができます。

## まとめ

- 「関数」と「型」は「型パラメータ」を持つことができる。
- 「型パラメータ」の満たすべき性質は「インタフェース型」を「型制約」として使うことで表す。
- 関数や型の後に`[T constraint]`という文法要素をつけると、「型パラメータTを宣言する。Tは`constraint`を実装しなければならない」という意味になる。`constraint`は型制約と呼ばれ、インタフェース型を用いる。
- メソッドに追加で型パラメータを宣言することはできない。
- `any`インタフェースは空インタフェース`interface{}`の別名である
- 比較可能、つまり`==, !=`による等値判定が可能な非インタフェース型により実装される`comparable`が提供される。
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

ところが、このコードは動作せず、次のようなエラーメッセージを出力します。

```
invalid operation: cannot compare x >= y (operator >= not defined on T)
```

`T`の型制約は`any`なので、演算子`>=`で比較できるとは限らないからです。それでは、適当なインタフェース型を定義して演算子`>=`で比較できるような型制約にすることはできるでしょうか？

Go1.17までのインタフェース型では、これはできませんでした。なぜなら、Go1.17までのインタフェース型とは「メソッドセット」すなわちメソッドの集合（集まり）を定義するものであって、「ある演算子が使える」というようなメソッド以外の型の性質を表すことはできないからです。

そこでGo言語は、「インタフェース型」として次のようなものも定義できるように機能を拡張することにしました。

```go
type Number interface {
    int | int32 | int64 | float32 | float64
}
```

この`Number`というインタフェースは、`int, int32, int64, float32, float64`という5種類の型によって **「実装」** され、これ以外の型によっては実装されません。
この文法要素`int | int32 | int64 | float32 | float64`のことを`unions`や`union element`と呼びます。

:::message

`|`を使わずに一つだけの型を書けば、その**一つの型によってのみ実装されるインタフェース**を定義できます。

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
この辺りは後程触れる`constraints`パッケージで解決されます。

:::

## for~rangeが使えるインタフェース型

型の満たす性質にはいくつか種類があります。ここでいくつか挙げてみましょう。

- あるメソッドを持っているという性質
- `==, !=`で比較できるという性質
- `<, >, >=, <=`で順序づけられるという性質
- `for ~ range`文でループを回すことができるという性質
- ある名前のフィールドを持っているという性質
- etc

「あるメソッドを持っているという性質」は従来のインタフェース型で表現できます。
`==, !=`で比較可能な性質は、組み込みの`comparable`インタフェースで表現できるのでしたね。
そして`<, >`などで順序づけられる性質は`unions`を利用した新しいインタフェースで表現できることをみました。

次に、`for ~ range`でループを回せるという性質をみてみましょう。
面白みのない例ですが、次のコードはコンパイルできます。

https://gotipplay.golang.org/p/ec6KpsOHgHv

```go
type I interface {
	[]int 
}

func f[T I](x T) {
	for range x {
	}
}
```

`I`を実装する型は`[]int`のみで、かつこの型は`for range`でループすることができる型です。
このような場合、`I`を型制約とする型パラメータの値に対して`for range`ループを書くことができます。

::: message

ここで代わりに
```go
type I interface {
	[]int | []string
}
```
とするとコンパイルが通らなくなります。

> ./prog.go:11:12: cannot range over x (variable of type T constrained by I) (T has no core type)

ここで言われているのは`I`が"core type"を持たないということです。この"core type"の説明は次章で行います。
:::


## `unions`を含むインタフェースは型制約でしか使えない

型制約ではなく通常の変数の型として`unions`を使うと、いわゆるsum typeのようなものが定義できそうに見えます。
しかし、現在のところこれは許可されません。

```go
type IntString interface {
	int | string
}

var x IntString // これはできない
```

この制限は将来的に取り除かれる可能性があります。型パラメータの導入だけでも非常に大きな変更であるため、安全を期するためにまずは最低限の機能でリリースし、実際の使われ方からフィードバックを得て判断していくのだと思います。

## フィールドを持つという性質は型制約で扱えない

できそうでできないことを1つ挙げておきます。ある名前のフィールドを持つという性質を型制約で表現することはできません。

https://gotipplay.golang.org/p/WEM-yelirK1

```go
type I interface {
	X
}

type X struct {
	SomeField int
}

func f[T I](x T) {
	fmt.Println(x.X) // これはできない
}
```

## まとめ

- Goの型パラメータは型制約をインタフェース型によって表現するが、型の性質には「メソッドを持つ」以外の性質もある。その性質の一部は`unions`を利用した新しいインタフェース型によって表現できる。
- `<, >, <=, >=`による順序付可能性は`unions`を使って順序づけられる型のみを列挙することで表現できる。
- `for range`ループができる型を使って、その1つの型だけからなる`unions`による型制約を作ると、その型制約に従う型パラメータ型の値について`for range`ループができる。
- `==, !=`による比較可能性は`comparable`インタフェースで表現する(再掲)。
- フィールドを持つという性質を型制約で表して、型パラメータ型の値のフィールドを参照することはできない。

# 近似要素`~`、underlying type、core type

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

## `~`(approximation element)

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

## `constraints`パッケージと`unions`

`<, >`で比較可能な型を`unions`で列挙できることは分かりましたが、実際に全ての型を書こうとすると面倒だなと思われた方もいると思います。

そこで、順序付けできるとか、数値型である、などの基本的な型制約はパッケージ`constraints`で提供されることになりました。

https://github.com/golang/exp/blob/master/constraints/constraints.go

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

# core type

underlying typeを一般化した新しい概念であるcore typeと、それがどのように仕様に関わるかについて説明します。

## 定義

https://tip.golang.org/ref/spec#Core_types

定義はこちらにあるのですが、

- この定義どおりに読むと型セット(type set)の概念が必要なので型セットを使わずに読み替えたい
- 型パラメータ型のcore typeについて言語仕様書の記述に疑問がある

ことから、筆者が妥当と考えている定義を以下記載します。言語仕様書がアップデートされ次第こちらもできるだけメンテするつもりです。

## core typeが存在する場合としない場合

Go言語の型はcore typeという型を持つ場合と持たない場合があります。型`T`がcore typeを持つのは、次の場合です。

- `T`がインタフェース型でも型パラメータでもないとき
- `T`がインタフェース型であり、つぎのいずれかに該当するとき
	- `T`を実装する全ての型のunderlying type`U`が同一であるとき
	- `T`を実装する全ての型は、同一の要素型`E`のチャネル型であり、かつ、それらが方向付きチャネルを含む場合にはその方向が同一であるとき
		- つまり、`E`の受信チャネル`<-chan E`と送信チャネル`chan<- E`の両方は含んでいないとき
- `T`が型パラメータであり、その型制約(常にインタフェース型)がcore typeをもつとき

これ以外のすべての場合、`T`はcore typeを持ちません。

`T`がcore typeをもつ場合、それぞれのケースにおいてcore typeは次のように決まります。

- 型`T`がinterface型でも型パラメータでもないとき、`T`のcore typeは`T`のunderlying typeである
- 型`T`がinterface型のとき
  - その`T`を実装する全ての型のunderlying type`U`が同一であるとき、`T`のcore typeは`U`である
  - その`T`を実装する全ての型のunderlying typeが同一の要素型`E`を持つchannel型であり、
    - underlying typeが双方向チャネル型のみであれば、その双方向チャネル型`chan E`が`T`のcore typeである
    - 受信チャネル`<-chan E`か送信チャネル`chan<- E`のどちらか一方のみがunderlying typeに含まれていれば、それが`T`のcore typeである
- 型`T`が型パラメータ型で、`T`の型制約がcore typeを持つとき、それが`T`のcore typeである

channelの場合には例外的な規定が必要なのでややこしくなっていますが、大雑把に言えば次のような理解で大丈夫です。

### 大雑把なcore typeの理解

- `T`がインタフェース型でも型パラメータでもないとき、`T`のcore typeは`T`のunderlying type
- `T`がインタフェース型であり、`T`を実装する全ての型のunderlying type`U`が同一であれば`T`はcore typeを持ち、`T`のcore typeは`U`
- `T`がインタフェース型であり、`T`を実装する全ての型のunderlying type`U`が同一でなければ`T`はcore typeを持たない(これはchannelの例外があり厳密には正しくない)
- `T`が型パラメータであるとき、`T`のcore typeは`T`の型制約のcore type(core typeが存在するかしないか含めて型制約に従う)

:::message

言語仕様上、型パラメータ`T`のcore typeは文言通りに読むと`T`のunderlying typeである型制約になります。しかし、仕様書の随所で「型パラメータのcore type」という言い方を「制約のcore type」の意味で用いている箇所が見られ、かつその用語法がむしろ便利である（だからこそ厳密ではないのに使われてしまっている）と考えたため、本記事では型パラメータのcore typeは型制約のcore typeであるとしました。

そのような用法の一例を挙げます:

https://tip.golang.org/ref/spec#Composite_literals

> The LiteralType's core type T must be a struct, array, slice, or map type

これは型パラメータ`T`が"LiteralType"に相当する場合にでコンポジットリテラル`T{}`が作れることへの言及であり、`T`のcore typeは明らかに`T`の型制約のcore typeと解釈されています。

一方、core typeの文言上の定義によれば型パラメータ`T`はインタフェース型ではないのでunderlying type = core typeのはずです。

:::

## 具体例

次の型制約はcore typeをもつでしょうか？またその場合core typeは何でしょうか？

```go
type C1 interface {
	~[]int
}

type C2 interface {
	int | string
}
```

- `C1`を実装する全ての型のunderlying typeは`[]int`なので`C1`はcore typeをもち、core typeは`[]int`です。
- `C2`を実装する型は`int`, `string`なのでunderlying typeは同一でなく`C2`はcore typeをもちません。

## core typeの登場場面

このcore typeですが、次のような場面で登場します。

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
ところが今`T`は型パラメータなので、`T`のunderlying typeはその型制約です。
このような場合、代入可能性を判断するために「`T`のunderlying typeの代わりに`T`の制約のcore typeを使う」というのが新しく追加される仕様です。

つまり、`T`の制約である`C`のcore typeが`[]int`なので最初の例は代入できましたが、二つ目の例は`C`のcore typeが存在しないために代入できなくなったということです。

## composite literals

型パラメータ型を使って、composite literalsを書くことができる場合があります。
その条件にもcore typeが関係しています。

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

composite literalを作るのに使う型名が型パラメータ型の名前である場合、その制約はcore typeを持っていなければいけません。
2つ目の例は、全く同じ構造のstruct型のunionsであるにもかかわらず、struct tagの有無によって型の同一性が満たされず、`C`のcore typeが存在しないため、`T`のcomposite literalは作れません。

## `for range`ループ

前章`for range`ループについて扱いましたが、実は型パラメータ型に対して`for range`ループを回すためには、制約がcore typeを持っていないといけません。

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

./prog.go:11:12: cannot range over x (variable of type T constrained by I) (T has no core type)

## 制約型推論とcore type

他にcore typeの関連仕様として欠かせないのは型推論アルゴリズムの一部である「制約型推論」です。
「制約型推論」が適用されるためには、型制約がcore typeをもつことが必要となっています。これを理解することで、「なんでこれは型推論できるのにこれはできないの？」という疑問にスッキリ答えられるようになるでしょう。

型推論については、続編で解説できればと思います。

## まとめ

- `~`近似要素をつかうと型定義によって作りうる無限の型にインタフェースを実装させることができる
- `~T`は`T`をunderlying typeに持つすべての型を表す
- core typeはunderlying typeを拡張したような概念で、型推論をはじめ型パラメータの関わる言語仕様の随所に現れる重要概念

## 最後に

この記事をかくにあたり[#gospecreading](https://gospecreading.connpass.com/)から得られた理解が本質的でした。いつもありがとうございます。