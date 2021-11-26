---
title: "簡単な例"
---

# Goの型パラメータの基本原則

型パラメータの基本事項についてはType Parameters Proposalの冒頭に挙げられていますが、その中でも筆者が特に重要と考えるのは次の2つです。この2つを覚えればGoの型パラメータを十分に使うことができると思います。

- 「関数」と「型」は「型パラメータ」を持つことができる。
- 「型パラメータ」の満たすべき性質は「インタフェース型」を「型制約」として使うことで表す。

# 具体例1: 型パラメータを持つ関数`f[T Stringer]`

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

# 具体例2: 型パラメータを持つ型`Stack[T any]`

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



## メソッド宣言において新たな型パラメータは宣言できない

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


- Setを書く？

# 具体例3: 型パラメータを持つ型`Set[T comparable]`

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
# Go1.17でできなかったこと

ここで少し型パラメータのモチベーションを知るためにGo1.17でのコードを考えてみます。

## インタフェース型のスライスを受け取る関数

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
`f2`が`[]MyInt`以外のスライス型を受け取るようにするには、それぞれの型についての[型アサーション](https://go.dev/ref/spec#Type_assertions)を書く必要があります。

```go
if vs, ok := xs.([]Stringer); ok
```

のようなアサーションを書くこと自体はできますが、こう書いても`[]MyInt`型の値を渡したときには`!ok`となります。

型スイッチ文を使う場合も、渡すかもしれない具体的な型ごとにcase節が必要です。
:::

## 型パラメータによる記述

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

# まとめ

- 「関数」と「型」は「型パラメータ」を持つことができる。
- 「型パラメータ」の満たすべき性質は「インタフェース型」を「型制約」として使うことで表す。
- 関数や型の後に`[T constraint]`という文法要素をつけると、「型パラメータTを宣言する。Tは`constraint`を実装しなければならない」という意味になる。`constraint`は型制約と呼ばれ、インタフェース型を用いる。
- メソッドに追加で型パラメータを宣言することはできない。
- `any`インタフェースは空インタフェース`interface{}`の別名である
- 比較可能、つまり`==, !=`による等値判定が可能なすべての型により実装される`comparable`が提供される。
- 型パラメータの重要な使い方の1つは、スライスやマップなどのいわゆるコレクション型の抽象化である。