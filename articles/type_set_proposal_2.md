---
title: "Type Sets Proposalを読む(2)"
emoji: "📖"
type: "tech" # tech: 技術記事 / idea: アイデア
topics: [go, generics]
published: true
---

- [はじめに](#はじめに)
- [Type Sets Proposalとは何か](#type-sets-proposalとは何か)
- [interface/constraintに対して制限を追加する](#interfaceconstraintに対して制限を追加する)
  - [EBNFによる表現](#ebnfによる表現)
- [なぜこのように制限するのか](#なぜこのように制限するのか)
- [具体例](#具体例)
  - [unionsを標準形に変形する](#unionsを標準形に変形する)
  - [標準形のunionsを1つにまとめる](#標準形のunionsを1つにまとめる)
  - [メソッドのインライン化](#メソッドのインライン化)
  - [最終形](#最終形)
- [ある型が型制約を満たすかどうかの判定](#ある型が型制約を満たすかどうかの判定)
- [ある型制約が別な型制約に含まれるかどうかの判定](#ある型制約が別な型制約に含まれるかどうかの判定)
- [この制限がないとどうなるか](#この制限がないとどうなるか)
- [最後に](#最後に)

# はじめに

この記事は、https://github.com/golang/go/issues/45346 に加えられた修正内容とその意味について説明するもので、[Goの"Type Sets" Proposalを読む](https://zenn.dev/nobishii/articles/99a2b55e2d3e50)の続編です。前編を読んでいない方は先に読んでからこの記事を読んだ方がわかりやすいと思います。

前提となる知識は次のようなものです。

- Go言語についての初歩的な知識と実装経験([A Tour of Go](https://tour.golang.org/)をやったことがあるくらいで大丈夫)
- [Type Parameters Proposal](https://go.googlesource.com/proposal/+/refs/heads/master/design/43651-type-parameters.md)の概要
- [Goの"Type Sets" Proposalを読む](https://zenn.dev/nobishii/articles/99a2b55e2d3e50)の後半の内容
- [underlying type](https://golang.org/ref/spec#Types)や[method sets](https://golang.org/ref/spec#Method_sets)の理解

# Type Sets Proposalとは何か

この記事では、https://github.com/golang/go/issues/45346 のことを簡単に"Type Sets Proposal"と呼ぶことにします。

Type Sets Proposalは、Go言語のGenericsの実現方法に関わるProposalです。2021年7月22日(JST)にAcceptされました。

より具体的に書くと、Type Sets Proposalとは、Type Parameters Proposalにおける型制約の表現手段であった"type list"を置き換えて改善するProposalです。つまり、現在のType Parameters Proposalの内容の一部がこのProposalの内容に置き換えられて採用されます。

# interface/constraintに対して制限を追加する

その具体的な内容はdescriptionにあるのですが、この記事で紹介したいのはそこからさらに加えられた変更内容です。その内容は、griesemer氏による次のコメントで詳しく説明されています。

https://github.com/golang/go/issues/45346#issuecomment-862505803

非常に丁寧に説明はされているのですが、それでも十分に難しいので、より具体的に理解しやすく紹介することを試みたいと思います。

変更内容を一言で言うと、「interface/constraintとして許容されるパターンが当初のType Sets Proposalよりも狭く限定される」と言う変更です。具体的には、次のように制限されます。

interface定義において、union element(以下、unionsと書きます)の項(原文では"term"のため以下termと記載します)となる型は、methodをもつinterface型であってはいけません。言い換えると、methodを持つinterface型は、スタンドアローンで現れなければいけません。

言っていることがわかりにくいと思いますが、proposalのコメントで具体例を書いてくれていますのでこれを借りて説明します。


```go
// OKな例
type Stringer interface {
    // そもそもunionsがないので問題なし
    String() string 
}

type Number interface {
    // unionsがあるが、termであるintとfloatはいずれもnon-interface型なので問題なし
    ~int | ~float64 
}

type C1 interface {
    // NumberはMethodを持たないInterfaceなので、unionsの項(term)になることができる。	
    Number | ~string
    // StringerはMethodを含むInterfaceだが、"stand-alone"で埋め込まれているのでOK
    Stringer		
    m()
}

type C2 interface {
    // C1 はMethodをもつInterfaceだが、"stand-alone"で埋め込まれているのでOK
    C1			
}
```

対して、次の例は禁止されます。

```go
// ダメな例
type C2 interface {
    // invalid: Stringerはmethodを定義しているinterface型なので、unionsのtermとして使ってはいけない
    ~int | Stringer		
}
```

当初のType Sets Proposalでは「ダメな例」の書き方も許されていました。ですが、最新のType Sets Proposalではこれは許されなくなります。※コンパイルエラーになると思われます。

## EBNFによる表現

EBNFを用いてもう少し厳密に述べましょう。Type Sets Proposalの下では、Goのinterface型は次のように定義されます。

```ebnf
InterfaceType = "interface" "{" { ( MethodSpec | InterfaceTypeName | ConstraintElem ) ";" } "}" .
ConstraintElem = ConstraintTerm { "|" ConstraintTerm } .
ConstraintTerm = [ "~" ] Type .
```

比較して、Version of Feb 10, 2021(Go1.16)の仕様書では、`InterfaceType`は次のように定義されています。`ConstraintElem`が差分となっていることがわかります。

```ebnf
InterfaceType      = "interface" "{" { ( MethodSpec | InterfaceTypeName ) ";" } "}" .
```

なお、上記において、次の定義は共通です。

```ebnf
MethodSpec         = MethodName Signature .
MethodName         = identifier .
InterfaceTypeName  = TypeName .
```

具体例を挙げます。`InterfaceType`とは次の全体を指します。`type Hoge`のところは含まれないことに気をつけてください。

```go
interface {
	Number | ~string	
	Stringer	
	m()	
}
```

そして、`m()`は`MethodSpec`、`Stringer`は`InterfaceTypeName`、そして`Number | ~string`は`ConstraintElem`に対応します。

生成規則`ConstraintElem = ConstraintTerm { "|" ConstraintTerm } .`は、`ConstraintElem`が1つ以上の`ConstraintTerm`からなることを表します。

さらにこの`ConstraintTerm`としては`Type`を取ることができます。この`Type`として「『MethodSpecを含むinterface型、及びMethodSpecを含むinterface型をInterfaceTypeNameに埋め込んだinterface型』を使うことはできない」というのが今回の変更内容だと言えます。

※`ConstraintTerm`が1つだけの場合は許可されるべきではないかという気もしますが、そのケースは`ConstraintElem`ではなく`InterfaceType`として許可されるので、`ConstraintTerm`の`Type`に対して上記の`interface`が禁止される、という規定の仕方で良いと思います。ここは新しい仕様書でどういう記述になるかはわかりません。

# なぜこのように制限するのか

なぜこのような制限が追加されたのでしょうか？要約すると次のようになります。

- この形のunion element(unions)は、「標準形」に変形することができる
- 「標準形」のunions同士は、型セットの包含関係を比較的簡単に計算できる
- ゆえに、次のような判定問題の解決が容易になる
    - ある型がある型制約を満たすかどうかの判定問題
    - ある型制約がある型制約に「含まれる」かどうかの判定問題

# 具体例

次のinterface型(型制約)を使って説明しましょう。少し複雑ですみませんが、「複雑なものを簡単に変形する」ための具体例なのでご容赦ください。

```go
type C interface {
    Number | ~string | MyFloat
    ~int8 | ~int16 | ~int32 | ~int64 | ~int | ~string | MyFloat
    Stringer
    ToInt() int
}

type Number {
	~int | ~float64 
}

type MyFloat float64 

func (f MyFloat) ToInt() int {return int(f)}
func (f MyFloat) String() string {return strconv.FormatFloat(f, 'E', -1, 64)}
```

以下、ステップを踏んでこの`C`を簡約していきます。
## unionsを標準形に変形する

```go
type C interface {
    Number | ~string | MyFloat
    ~int8 | ~int16 | ~int32 | ~int64 | ~int | ~string | MyFloat
    Stringer
    ToInt() int
}
```

この`C`には2行のunionsがあります。まずそれぞれ簡単な形にします。

`Number | ~string | MyFloat`の`Number`はmethodを持たないinterface型です。しかも、この場合は1行のunionsだけからなっています。

```go
type Number {
    ~int | ~float64
}
``` 

そこで、`Number`は`~int | ~float64`に置き換えてしまいます。

`Number | ~string | MyFloat` = `~int | ~float64 | ~string | MyFloat`です。このようにnon-interface型もしくはapproximation要素だけになったら「unionsの標準化」は完了です。

2行目の`~int8 | ~int16 | ~int32 | ~int64 | ~int | ~string | MyFloat`は初めからnon-interface型もしくはapproximation要素だけになっているので今のところ何もしなくて良いです。

## 標準形のunionsを1つにまとめる

次は、2行に分かれているunionsを1行にまとめたいです。

unionsが2行書かれているとき、その型セットはそれぞれの型セットの共通部分(intersection)になるのでした。つまり、論理演算風に書くと`(A or B) and (C or D)`のような計算をすることになります。
これは`(A and C) or (A and D) or (B and C) or (B and D)`のように「展開」することができます。
```go

~int | ~float64 | ~string | MyFloat
~int8 | ~int16 | ~int32 | ~int64 | ~int | ~string | MyFloat

=

~int & ~int8 | ~int & ~int16 | ...(中略)... | MyFloat & MyFloat

```

のように、$4 \times 7$個のtermからなるunionsに展開できます。ここで`&`はこの記事独自の記号で、型セットの共通部分を取ることを意味するものとします。

複雑な式になったようにも見えますが、Goのnon-interface型には階層性のようなものがないので、`A`と`B`が同一の型でなければ`A&B`は空集合ですし、同一の型ならば`A&B = A`と単純に言い切ることができます。`~`がついている場合も`~A & ~B`は`A=B`ならば`~A`に等しく、そうでなければ空集合に等しいです。よって、

```go
~int & ~int8 | ~int & ~int16 | ...(中略)... | MyFloat & MyFloat

=

~int | ~string | MyFloat
```

と1行の標準形に直すことができます。これではじめに挙げたinterface型`C`は次のinterfaceと等価であることがわかりました。

```go
type C interface {
    ~int | ~string | MyFloat
    Stringer
    ToInt() int
}
```

## メソッドのインライン化

最後にスタンドアローンで埋め込まれている`Stringer`を「インライン化」して終わりです。

```go
type C interface {
    ~int | ~string | MyFloat
    String() string
    ToInt() int
}
```

## 最終形

以上により、次の最終形までinterface型を単純化することができました。

```go
type C interface {
    ~int | ~string | MyFloat
    String() string
    ToInt() int
}
```

この具体例に限らず一般的に、Type Sets Proposalに付け加えられた制限の下では、全てのinterface型は、

```go
interface {
    A | B | ~C | ... // 標準形のunions
    MethodA() // メソッド
    MethodB() // メソッド
    MethodC() // メソッド
    ...
}
```

のように、「標準形のunions」が0個または1個と、メソッドが0個以上定義されている形に変形することができます。EBNFで表すと、

```ebnf
InterfaceType = "interface" "{" [ ConstraintElem ";" ] { MethodSpec ";" } "}" .
ConstraintElem = ConstraintTerm { "|" ConstraintTerm } .
ConstraintTerm = [ "~" ] Type .
```

です。ここで`Type`は全てnon-interface型という条件がつきます。

# ある型が型制約を満たすかどうかの判定

このように単純化すると、ある型がある型制約を満たすかどうかは、unionsで表されるいわば「型セット部分」と、`{ MethodSpec }`で表されるいわば「メソッド部分」とを別々に考えることで判断できるようになります。

例えば、型`T`があるとするとき、これが次のインターフェースを満たすかどうかを考えます。

```go
type C interface {
    A | B | ~C | ... 
    MethodA() 
    MethodB() 
    MethodC() 
    ...
}
```

これは次のように簡単に判定できます。

- `T`がunionsに挙げられているいずれかのtermとマッチするかどうか?
- `T`が`MethodA, MethodB, MethodC...`を全て実装するかどうか？

両方がYesならば`T`は型制約`C`を満たしますが、どちらかが満たされなければ`T`は型制約`C`を満たしません。

# ある型制約が別な型制約に含まれるかどうかの判定

今度は型制約`C1`が`C2`に含まれるかどうかを考えてみましょう。これも、「型セット部分」と「メソッド部分」を分けて考えることができます。

```go
type C1 interface {
    A | B | ~C | ... 
    MethodA() 
    MethodB() 
    MethodC() 
    ...
}

type C2 interface {
    D | E | ~F | ... 
    MethodA() 
    MethodB() 
    MethodC() 
    ...
}
```

この時、`C1`が`C2`に含まれるのは、次の両方を満たすときです。

- `A, B, ~C...`のそれぞれが、`D, E, ~F...`のいずれかに含まれる
    - （※論理的には`D, E, ~F...`の和集合に含まれる、になりますが、Goの型の性質から「複数のtermにまたがって初めて包含される」ような場合はありません）
- `C1`のmethod setが`C2`のmethod setに含まれる

もし、`C1, C2`がそれぞれ上のような形に単純化されておらず、unionsのtermにメソッド定義が「混ざって」いたら、`C1`が`C2`の一部かどうかの判定はこれほど簡単にはいかないでしょう。

# この制限がないとどうなるか

ここで「禁止される」interface定義をもう一度見てみましょう。

```go
// ダメな例
type Invalid1 interface {
    // invalid: Stringerはmethodを定義しているinterface型なので、unionsのtermとして使ってはいけない
    ~int | Stringer		
    ToInt() int
}
```

`Stringer`がunionsのtermに現れているところが新しい言語仕様に違反しているポイントです。この`Stringer`で定義されている`String() string`メソッドをインライン化することはできないのでしょうか？例えば次のように「変形」してみましょう。

```go
// Invalid1を変形したインタフェース？？？
type Invalid2 interface {
    ~int 
    String() string
    ToInt() int
}
```

この「同値変形」は見るからに怪しいですが、実際に誤りです。誤りであることを言うには、`Invalid1`に含まれるが`Invalid2`に含まれない型、あるいは逆に`Invalid2`に含まれるが`Invalid1`に含まれない型を具体的に構成すれば良いです。ちょっと考えてみると、

```go
type MyFloat float64

func(MyFloat) String() string {...}
func(MyFloat) ToInt() int {...}
```

のような型を定義すれば、`MyFloat`が`Invalid1`を満たすけれども`Invalid2`は満たさないことがわかります。

結局どうやってみても、`Invalid1`のtermである`Stringer`をインライン化することはできないのです。そうすると、型制約を満たすかどうかの判定も先程のように単純に2つの問題に分けて考えよう、とはいかなくなります。これをみると、Type Sets Proposalに加えられた変更のメリットがわかります。

# 最後に

おそらく、この制限が気になるようなコードを書くことはあまりないような気がしています（想像力がないだけかもしれませんが）。

この記事の内容を勉強していて改めて感じたことは、Goのnon-interface型には階層性(サブクラスのようなもの）がないということです。複数行のunionsを1行のunionsに「まとめる」ときにこの性質が計算を簡単にしてくれていて、そのことが個人的に面白かったポイントでした。

荒削りな記述になっていると思いますので、質問・指摘・改善提案その他お気軽にいただければと思います。