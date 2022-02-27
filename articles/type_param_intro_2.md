---
title: "Go言語のジェネリクス入門(2) インスタンス化と型推論" # 記事のタイトル
emoji: "😸" # アイキャッチとして使われる絵文字（1文字だけ）
type: "tech" # tech: 技術記事 / idea: アイデア記事
topics: ["go"] # タグ。["markdown", "rust", "aws"]のように指定する
published: true # 公開設定（falseにすると下書き）
---
# はじめに

この記事は[Go言語のジェネリクス入門(1)](https://zenn.dev/nobishii/articles/type_param_intro)の続編で、インスタンス化や型推論について解説します。

実用上はコンパイルエラーになったら直せばいいのでここに書かれているようなことを知らなくてもそれほど困らないと思います。しかし、型推論の仕組みについて正確に知ることで初めて思いつくコーディングパターンも時々はあるかもしれません。

Go言語仕様書は非常に読みやすい言語仕様書ですが、それでもジェネリクス関係の仕様を正確に理解するのは骨が折れるはずなので、正確な仕様を理解したくなった方の助けになればと思います。


# 型セットについての資料

この記事では型セット(Type set)についての理解を前提とします。Type setについては、ひとまず次のポイントを理解してください。

- Go言語のinterface型はすべて、「型の集合（型セット）」を定めるものである。
- 型`T`がinterface`I`を実装するとは、`I`の型セットに`T`が属するということである。

型セットについては前編では説明していないのですが、現在のところ次のような資料が利用できます。

英語でよければ1番上のGriesemer氏の解説がよく、日本語で読む場合は[Go の "Type Sets" proposal を読む - Zenn](https://zenn.dev/nobishii/articles/99a2b55e2d3e50)の後半部分が良いかと思います。

| リンク | 内容紹介 |
| ---- | ---- |
| [GopherCon 2021: Robert Griesemer & Ian Lance Taylor - Generics!](https://www.youtube.com/watch?v=Pa_e9EeCdy8) | [英語] Go言語開発者によるジェネリクス解説です。前半のgriesemer氏の発表部分に型セットの説明があります。| 
| [Type Parameters Proposal](https://go.googlesource.com/proposal/+/refs/heads/master/design/43651-type-parameters.md) | [英語] ジェネリクスのプロポーザルドキュメントです。型セットについては[ここ](https://go.googlesource.com/proposal/+/refs/heads/master/design/43651-type-parameters.md#type-sets)で説明されています。 |
| [Go の "Type Sets" proposal を読む - Zenn](https://zenn.dev/nobishii/articles/99a2b55e2d3e50)| Type Setsのプロポーザルが出たときに書いた記事です。前半は経緯の解説なので今は読む必要ありません。Type setの仕様は後半で解説しています。| 
| [初めての型セット](https://speakerdeck.com/nobishino/introduction-to-type-sets) | Go1.17リリースパーティの発表スライドです。「型セット」と「実装」の概念理解にフォーカスしています。 すこし図があります。|
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
| 制約型推論 | すでに決定できた型引数と、型引数が従う型制約| 

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

## unification(unify)の直感的説明

型推論の厳密な説明には、unification(unify)の理解が必要です。

しかし、unificationの厳密な説明をすると前置きが長くなってしまうので、ここでは

- unificationとは、型パラメータを含む2つの型をパターンマッチングする仕組みである
- パターンマッチングした結果、substitution map entryが作られる

ことを具体例から感じ取ってください。

| 型1 | 型2 | つくられるsubstitution map entry |
| ---- | ---- | ---- |
| `[]map[int]bool`| `T1` | `T1 -> []map[int]bool` |
| `[]map[int]bool`| `[]T1` | `T1 -> map[int]bool` |
| `[]map[int]bool`| `[]map[T1]T2` | `T1 -> int, T2 -> bool` |
| `[]map[int]bool`| `*T` | unification失敗し、entryはつくられない |

ここで、**substitution mapとは、型推論によって作られるkey->valueストアであって、未知の型パラメータをkeyとし、他の型をvalueとするもの**です。

型推論の目的は、substitution mapを完成させて、未知の型パラメータを具体的な型引数に対応付けることだと言えます。

ここまでの言葉を使って少しフォーマルに言い直すと、「**unificationとは、2つの型を受け取って動作するルーチン**であり、その結果として**substitution mapに0個以上のエントリを追加する**もの」だと言えます。

ですから、今後unificationが出てきたときは、「受け取る2つの型は何なのか？」ということを問いながら読みすすめると理解がしやすいとおもいます。

以下、substitution map entryのことを単に「エントリー」と書きます。
## 関数引数型推論

関数引数型推論は、「関数に渡された実引数の型」と、「関数の引数の型」をunifyします。ただし、このunificationは、関数の引数の型が型パラメータを含むときにのみ行われます。

例を挙げましょう。


```go
func f[T any](x T) {...}

var x int
f(x)
```

この関数呼び出し`f(x)`では型パラメータ`T`が未知なので型推論が起動します。`x`の型`int`と, パラメータの型`T`がunificationルーチンの「引数」となります。`T`と`int`のunificationによりentry`T -> int`がsubstitution mapに追加され、ここですべての型パラメータが既知となるため型推論が完了します。

## 制約型推論

制約型推論は、「型パラメータ」と、「その型パラメータに課された型制約のcore type」をunifyします。
ただし、型制約がcore typeを持つ場合にのみ制約型推論が起動されます。

:::message

core typeについては[前編](https://zenn.dev/nobishii/articles/type_param_intro#core-type)を参照ください。

:::

:::message

型制約はインタフェース型であり、よってもちろんそれ自体が「型」であることを注意しておきます。「unificationは2つの型を受け取るルーチンである」と書いたとおりですね。

:::

:::message
次のような「ただひとつのdefined typeを型セットに含む」型制約の場合のみ、「core typeではなくadjusted core typeというものにunifyされる」という例外があります。

ただし実用上の重要性が低いため、この記事では一旦詳細を略します。

```go
type MyInt int

type C interface {
  MyInt
}
```

:::

仕様書にある例を使って説明します。

https://gotipplay.golang.org/p/G77uiNe_taU

```go
type T[A any, B []C, C *A] struct {
	A A
	B B
	C C
}

func main() {
	var t T[int]
	fmt.Printf("A: %T, B: %T, C:%T\n", t.A, t.B, t.C)
	// A: int, B: []*int, C:*int
}
```

このジェネリックな`T`関数は3つの型パラメータ`A, B, C`を持ちますが、変数`t`の宣言において1つの型引数`A = int`しか特定していません。しかし、制約型推論によって残り2つの型引数を決定できるため、コンパイルが成功します。

流れとしてはつぎのようになります。


1. `var t T[int]`は関数呼び出しではないため関数引数型推論は行われず、スキップされます。
1. 未知の型パラメータ`B, C`があるため、制約型推論が起動します。
   1. `C`はcore type`*A`をもつため、`C`と`*A`をunifyして`C -> *A`というエントリーができます。
   2. `B`はcore type`[]C`をもつため、`B`と`[]C`をunifyして`B -> []C`というエントリーができます。
   3. 明示された型引数`int`により、`A -> int`がすでにあるため、`C -> *A`における`A`を`int`で置き換えます。これで`C -> *int`というエントリーができます。
   4. さらに`B -> []C`における`C`を`*int`で置き換えることで、`B -> []*int`というエントリーができます。
   5. これ以上置き換えはできないので、制約型推論が終了します。
1. 未知の型パラメータがなくなったので、型推論を終了します。

最終的なsubstitution mapは次のようになっています:

```
A -> int
B -> []*int
C -> *int
```

# unification/unify

ここまでの話で、型推論(関数引数型推論と制約型推論)がどのように行われるかを説明してきました。

しかし、その説明のなかで「unifyする」とか「unification」という手続き（ルーチン）がどのようなものなのかは厳密に説明していませんでした。

このセクションでは、いままで直感的にしか説明していなかったunify/unificationを厳密に定義することで、型推論の解説を完成させます。

## unificationの厳密な定義

**定義**

2つの型をunifyするとは、その2つの型を **等価(equivalent)** にするようなsubstitution map entryを見つけることである。

つまり、`X`と`Y`という2つの型をunifyするとは、エントリー`P -> A`を適当に追加して、**`X`と`Y`に含まれている`P`を`A`に置き換えれば`X`と`Y`が等価になるようにする**ということです。

:::message

仕様書上は次のように書かれています。

https://tip.golang.org/ref/spec#Type_unification

> Unification is the process of finding substitution map entries that make the two types equivalent.

:::

この定義を完全なものとするには、「等価(equivalent)」とは何かも定義する必要があります。

「型が等価である」ことの厳密な定義は少しあとに回したいのですが、すこしだけ述べておくと、「型が等価である」というのは「型が同一である」よりも少し広い（ゆるい）条件です。
つまり、2つの型が同一(identical)であればそれらは等価でもあるのですが、2つの型が等価であっても同一ではない場合があります。

まずunificationの具体例をみておきましょう。
### 例


- `[B []C]`において、型`B`と型`[]C`をunifyするとは、エントリー`B -> []C`を追加することです。なぜなら、`B`には`B`が「含まれて」おり、`B`をエントリー`B -> []C`に従って置き換えることで`[]C`と`[]C`が「等価」になるからです。
- `[]map[int]bool`と`[]map[T1]T2`をunifyするとは、`T1 -> int, T2 -> bool`という２つのエントリーを追加することです。なぜなら、`[]map[T1]T2`に含まれる`T1`と`T2`をそれぞれエントリーに従って置き換えることで、2つの型はどちらも`[]map[int]bool`となり、等価となるからです。


![unification_1](/images/unification_1.jpeg)

unificationが失敗する例をあげておきます。

`[]map[int]bool`と`*T`のunificationを試みると、どのようなエントリ`T -> ?`を追加してもこの2つの型を等価にすることはできないので、unificationが失敗します。

型推論の中でunificationが失敗すれば、コンパイルエラーとなります。

https://gotipplay.golang.org/p/C1kepqzqWKJ

```go
func f[T any](x *T) {}

func main() {
	f(make([]map[int]bool, 0))
	// type []map[int]bool of make([]map[int]bool, 0) does not match *T (cannot infer T)
}
```

関数引数型推論により`[]map[int]bool`と`*T`のunificationが試みられますが、この2つを等価にするようなエントリーは存在しないのでunificationが失敗し、型推論の失敗によりコンパイルが失敗します。

:::message

unificationの厳密な定義をみたとき、腑に落ちなさを感じる方もいるとおもいます。その腑に落ちなさは、この定義が「エントリーを求める方法」について何も述べていないからではないかと思います。

実際、Go言語仕様書はunificationにおいて「エントリーを求める方法・アルゴリズム」については述べていません。しかし、unificationによって見つけるべきエントリーとはどのようなものなのか、については正確に述べているので、仕様書としては十分なのです。

高校相当の数学に例えると、「unificationの定義」は、「(0個以上の)エントリー`x`の組が満たすべき方程式」のようなものです。2次方程式`x^2 + x - 1 = 0`の解とは何か、という「定義」は、その方程式を「解く方法」がわからなくても、それとは無関係に行うことができます。これと同じように、unificationの定義においては、「エントリーが満たすべき性質」を述べているだけで、「そのようなエントリーを求める方法」については述べていませんし、それで十分だということです。

方程式のたとえによるならば、「方程式」を満たすエントリーが存在すればunificationが可能ですし、方程式が「解なし」であればunificationは失敗し、コンパイルエラーになるわけです。

:::
## 型の同一性(identity)と等価性(equivalence)

「型の同一性」はGo1.17以前から定義されている用語です。

https://go.dev/ref/spec#Type_identity

ここで厳密な定義を解説することはしませんが、`type A B`というように型定義したときに`A`と`B`とは異なる型であるということだけ注意してください。

**定義(型の等価性)**

2つの型`X, Y`が等価であるとは、つぎの3つのいずれかが成り立つことです。

- `X, Y`が同一(identical)であるとき
- `X, Y`がどちらもchannel型であり、方向を無視すれば同一であるとき
- `X, Y`のunderlying typeが等価であるとき

### 等価性の例

- `type X = int`と`type Y = int`は同一なので等価です。
- `X = <-chan int`と`Y = chan int`は方向を無視すれば同一なので等価です。
- `type X <-chan int`と`type Y = chan int`は、`X`のunderlying type`<-chan int`と`Y`のunderlying type`chan int`が等価なので、等価です。

等価ではない例も挙げておきます。

```go
type MyInt int

type X chan int

type Y chan MyInt
```

このように定義した`X, Y`は等価ではありません。

## 等価性とunificationの例

等価性があってはじめてunificationが成功する例を挙げましょう。

https://gotipplay.golang.org/p/ckSANEXiR9c

```go
package main

import "fmt"

type A chan int
type X[T ~<-chan U | ~chan U, U any] struct {
	T T
	U U
}

func main() {
	var x X[A]
	fmt.Printf("T: %T, U: %T\n", x.T, x.U)
	// T: main.A, U: int
}
```

ジェネリックな型`X`は型パラメータ`T, U`を持ちますが、宣言`var x X[A]`において１つの型引数しか渡されていません。
よって、制約型推論によって`U`が決定できないとコンパイルエラーになってしまいます。

この型推論は次のように進みます。

- エントリー`T -> A`は明示的に与えられます。
- 型パラメータ`U`の制約`any`はcore typeが存在しない型なので、`U`についての制約型推論は行われません。
- 型パラメータ`T`の制約`~<-chan U | ~chan U`はcore type`<-chan U`を持ちます。よって制約型推論により、`T`と`<-chan U`のunificationを行いますが、`T`は`A`であることがすでにわかっているので、`A`と`<-chan U`のunificationを行います。
- つまり、エントリー`U -> ?`を追加することで`type A chan int`と`<-chan U`を等価にすることができるかどうかが問題になります。
- 試みとして、`U -> int`というエントリーを追加すると仮定しましょう。
- `type A chan int`と`<-chan int`が等価かどうか？を考えます。
- この2つは同一ではないので、1つ目の条件にはあてはまりません。
- `A`は`chan int`とは異なる型なので、channelの方向を無視しても同一の型にはなりません。よって2つ目の条件にもあてはまりません。
- 最後にunderlying typeを考えると、`A`のunderlying typeは`chan int`であり、これは方向を無視すれば`<-chan int`と同一の型になります。よって等価性の3番めの条件を満たしています。
- つまり、エントリー`U -> int`を追加することで、型`A`と`<-chan int`を等価にすることができたので、このエントリーの追加をもってunificationが成功しました。
- これですべての型パラメータが確定したので、型推論が完了します。

:::message

このcore typeは難しいパターンなので、[前編のcore type部分](https://zenn.dev/nobishii/articles/type_param_intro#core-type)を見直してください。

:::

:::message

上記の例だとあたかもエントリーを「あてずっぽう」に決めているように見えるので違和感をおぼえるかもしれません。

もちろん実際のコンパイラは決まったアルゴリズムでエントリーを発見しているはずですが、前述の通りunificationが可能であることを示すためには実際に等価性を満たすことのできるエントリーが存在することを言えば十分です。

:::

# 具体例や未解決の問題

いくつか型推論にまつわる面白い話題を挙げてこの記事を終わります。

## 公式ドキュメントに見る制約型推論の活用例

Type Parameters Proposalに出てくる、[Pointer method example](https://go.googlesource.com/proposal/+/refs/heads/master/design/43651-type-parameters.md#pointer-method-example)を解説します。

:::message

このセクションはほとんどproposalにかかれている通りのことを書きます。ここで挙げるのは、制約型推論の応用として面白い例だからです。

:::

次のようなジェネリックな関数を考えます。

```go
// Setterはstringの値をSetできるインタフェース
type Setter interface {
	Set(string)
}

// FromStringsは[]stringを受け取ってSetterのスライスを返す
// その際にSetメソッドでsの内容をSetする
func FromStrings[T Setter](s []string) []T {
	result := make([]T, len(s))
	for i, v := range s {
		result[i].Set(v)
	}
	return result
}
```

これを次のように使いたいのですが、これはコンパイルできません。

```go
type Settable int

func (p *Settable) Set(s string) {
	i, _ := strconv.Atoi(s) // real code should not ignore the error
	*p = Settable(i)
}

func F() {
	nums := FromStrings[Settable]([]string{"1", "2"})
  _ = nums
}
```

https://gotipplay.golang.org/p/g2GkggqE7e0

> Settable does not implement Setter (Set method has pointer receiver)

これは、`Set`メソッドを実装しているのはあくまで`*Settable`型であり、`Settable`型ではないので、`Settable`が型制約`Setter`を実装しないからです。

では、`*Settable`型を渡すとどうなるでしょうか。

```go
func F() {
	nums := FromStrings[*Settable]([]string{"1", "2"})
	_ = nums
}
```

https://gotipplay.golang.org/p/TYI_tmQ06tZ

> panic: runtime error: invalid memory address or nil pointer dereference
> [signal SIGSEGV: segmentation violation code=0x1 addr=0x0 pc=0x45810e]

今度はコンパイルできるのですがpanicしてしまいます。

```go
func FromStrings[T Setter](s []string) []T {
	result := make([]T, len(s))
	for i, v := range s {
		result[i].Set(v)
	}
	return result
}
```

をみると、`result[i]`は`*Settable`のゼロ値である`nil`が入っているので、`Set`呼び出しのときにnil pointer dereferenceになるからです。かといって、`Settable`という型を`T`から取り出すことはできません。

素朴な解決策は、2つの型パラメータを用いることです。

```go
// パラメータ付きの型制約。何かのBに対するポインタ型であり、かつ、Set(string)を実装するという型制約になっている。
type Setter2[B any] interface {
	Set(string)
	*B
}

// 2つの型パラメータを持つようにする
func FromStrings2[T any, PT Setter2[T]](s []string) []T {
	result := make([]T, len(s)) // T型のスライスとして初期化する
	for i, v := range s {
		// PT型へのコンバージョン
		p := PT(&result[i])
		// PTはSetメソッドを持つ
		p.Set(v)
	}
	return result
}
```

これを利用して`F`を書き直せます:

```go
func F2() {
	nums := FromStrings2[Settable, *Settable]([]string{"1", "2"})
	fmt.Println(nums) // [1 2]
}
```

https://gotipplay.golang.org/p/VFxDjHrE7N6

これは意図通りに動作しますが、2つの型引数を渡すところが億劫です。

そこで、制約型推論を活用して次のようにすることができます。

```go
func F3() {
	nums := FromStrings2[Settable]([]string{"1", "2"})
	fmt.Println(nums) // [1, 2]
}
```
https://gotipplay.golang.org/p/01aOdAHrXut

型引数を2つ渡すのではなく、１つだけ渡しました。これでも結果は同一となります。

起きているのは次のようなことです。

- `T -> Settable`が確定する
- `PT`の型制約`Setter2[T]`の`T`に`Settable`を渡してインスタンス化する。
- `PT`の型制約を擬似コードで書くと次のようになる:

```go
type Setter2Instantiated interface {
	Set(string)
	*Settable
}
```

- 制約型推論が起動する。`Setter2Instantiated`はcore type `*Settable`を持つので、`PT`と`*Settable`をunifyする。
- エントリー`PT -> *Settable`が作られ、全ての型引数が確定して型推論が終了する。

このように、制約型推論を使うことで、ポインタレシーバメソッドをジェネリックに扱う関数を少し短いコードで呼び出すことができました。
## 関数引数型推論と引数の順序

https://github.com/golang/go/issues/43056

関数引数型推論において、unifyしうる型のペアが複数あるとき、どの順序でunifyを行うかは未定義です。
ほとんどの場合、順序によって結果が変わることはありませんが、結果が変わる厄介な例が挙げられています。

## 制約型推論とdefined type、型推論インタリービング

https://github.com/golang/go/issues/51139

制約型推論とdefined type、代入可能性の兼ね合いでコンパイルできないコードが挙げられています。
Go1.18仕様では関数引数型推論と制約型推論を別のステップで行いますが、関数引数型推論の対象となるペアが複数ある場合に間に制約型推論を挟み込むことでコンパイルが可能になるのではないかという提案がなされています。(issueでいうところのinterleaved world)

問題はこれがGo1.18との後方互換性を持つかどうかで、griesemer氏は「おそらく後方互換性があるだろう」と述べています。

# おわりに

筆者が特に書きたかったところがunificationの厳密な定義であるため、それ以外の実用上重要な仕様への言及漏れがあるとおもいます。たとえば型引数の部分的省略、ジェネリックな型制約、相互参照する型制約、また型エイリアスなどについて言及できていません。

そうした不足分に限らず、誤りやわかりにくい点、古くなった用語なども含めて出来得る限りメンテナンスしたいと考えています。GitHubのPRやIssueなど歓迎します。（いきなりPR出していただいて大丈夫です）