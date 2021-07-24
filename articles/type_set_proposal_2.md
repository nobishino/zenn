---
title: "Type Sets Proposalを読む(2) カノニカル形式編"
emoji: "💬"
type: "tech" # tech: 技術記事 / idea: アイデア
topics: [go, generics]
published: false
---

- [はじめに](#はじめに)
- [Type Sets Proposalとは何か](#type-sets-proposalとは何か)
- [この記事のテーマ: Type Sets Proposalに加えられた変更](#この記事のテーマ-type-sets-proposalに加えられた変更)
- [interface/constraintに対して制約を追加する](#interfaceconstraintに対して制約を追加する)
- [なぜこのように制約するのか](#なぜこのように制約するのか)
- [具体例](#具体例)
  - [unionsを標準形に変形する](#unionsを標準形に変形する)
  - [標準形のunionsを1つにまとめる](#標準形のunionsを1つにまとめる)
  - [メソッドのインライン化](#メソッドのインライン化)
- [constraintの包含関係](#constraintの包含関係)
  - [命題1](#命題1)
    - [導出1](#導出1)
  - [命題2](#命題2)
    - [導出2](#導出2)

# はじめに

この記事は、https://github.com/golang/go/issues/45346 に加えられた修正内容とその意味について説明するもので、[Goの"Type Sets" Proposalを読む](https://zenn.dev/nobishii/articles/99a2b55e2d3e50)の続編です。前提となる知識は次のようなものですので、前編を読んでいない方は先に読んでからこの記事を読んだ方がわかりやすいと思います。

- Go言語についての初歩的な知識と実装経験(A Tour of Go程度)
- Type Parameters Proposalの概要
- [Goの"Type Sets" Proposalを読む](https://zenn.dev/nobishii/articles/99a2b55e2d3e50)の後半の内容
- underlying typeやmethod setsの理解

この記事では、https://github.com/golang/go/issues/45346 のことを簡単に"Type Sets Proposal"と呼ぶことにします。

内容メモ

- 制約の内容説明
- モチベーション
    - インタフェース型の標準形への変換ができるようになる
    - 標準形に変換されたインタフェース型同士は包含関係を容易に計算できる
- 標準形への変形の具体例
- 標準形へ変換できると何が嬉しいか


# Type Sets Proposalとは何か

Type Sets Proposalは、Go言語のGenericsの実現方法に関わるProposalです。2021年7月22日(JST)にAcceptされました。

より具体的に書くと、Type Sets Proposalとは、Type Parameters Proposalのおける型制約の表現手段であった"type list"を置き換えて改善するProposalです。つまり、現在のType Parameters Proposalの内容の一部がこのProposalの内容に置き換えられて採用されることになります。

# この記事のテーマ: Type Sets Proposalに加えられた変更

その具体的な内容はdescriptionにあるのですが、この記事で紹介したいのはそこからさらに加えられた変更内容です。その内容は、griesemer氏による次のコメントで詳しく説明されています。

https://github.com/golang/go/issues/45346#issuecomment-862505803

非常に丁寧に説明はされているのですが、それでも十分に難しいので、より具体的に理解しやすく紹介することを試みたいと思います。
# interface/constraintに対して制約を追加する

変更内容を一言で言うと、「interface/constraintとして許容されるパターンが当初のType Sets Proposalよりも狭く限定される」と言う変更です。具体的には、次のように制約されます。

interface定義において、union element(以下、unionsと書きます)の項(原文では"term")となる型は、methodをもつinterface型であってはいけません。言い換えると、methodを持つinterface型は、スタンドアローンで現れなければいけません。

言っていることがわかりにくいと思いますが、proposalのコメントで具体例を書いてくれていますのでこれを借りて説明します。


```go
// OKな例
type Stringer {
	String() string // そもそもunionsがないので問題なし
}

type Number {
	~int | ~float64 // unionがあるが、termであるintとfloatはいずれもnon-interface型なので問題なし
}

type C1 interface {
	Number | ~string	// NumberはMethodを持たないInterfaceなので、unionsの項(term)になることができる。
	Stringer		// StringerはMethodを含むInterfaceだが、"stand-alone"で埋め込まれているのでOK
    m()
}

type C2 interface {
	C1			// C1 はMethodをもつInterfaceだが、"stand-alone"で埋め込まれているのでOK
}
```

// ダメな例
```go
type C2 interface {
	~int | Stringer		// invalid: Stringerはmethodを定義しているinterface型なので、unionsのtermとして使ってはいけない
}
```

当初のType Sets Proposalでは「ダメな例」の書き方も許されていました。ですが、最新のType Sets Proposalではこれは許されなくなります。※コンパイルエラーになると思われます。

EBNFを用いてもう少し厳密に述べましょう。

# なぜこのように制約するのか

なぜこのような制約が追加されたのでしょうか？要約すると次のようになります。

- この形のunion element(unions)は、「標準形」に変形することができる
- 「標準形」のunions同士は、型セットの包含関係を比較的簡単に計算できる
- ゆえに、次のような判定問題の解決が容易になる
    - ある型がある型制約を満たすかどうかの判定問題
    - ある型制約がある型制約に「含まれる」かどうかの判定問題

この記事を最後まで読むと、これらの判定問題が容易であることの理由の1つに、「Go言語のnon-interface型には階層性がない」という事実が「効いている」ことも見えてきます。

TBW:コード例
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
	~int | ~float64 // unionがあるが、termであるintとfloatはいずれもnon-interface型なので問題なし
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
	~int | ~float64 // unionがあるが、termであるintとfloatはいずれもnon-interface型なので問題なし
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

のように、$4 \dot 7$個のtermに展開できます。ここで`&`はこの記事独自の記号で、型セットの共通部分を取ることを意味するものとします。

複雑な式になったようにも見えますが、Goのnon-interface型には階層性のようなものがないので、`A`と`B`が同一の型でなければ`A&B`は空集合ですし、同一の型ならば`A&B = A`と単純に言い切ることができます。よって、

```go
~int & ~int8 | ~int & ~int16 | ...(中略)... | MyFloat & MyFloat

=

~int | ~string | MyFloat
```

と1行の標準形に直すことができます。これで元の`C`は次のinterfaceと等価であることがわかりました。

```go
type C interface {
    ~int | ~string | MyFloat
    Stringer
    ToInt() int
}
```

## メソッドのインライン化

TBW
# constraintの包含関係

要素$a$が要素$b$に含まれるとは、要素$a$の型セットが$b$の型セットに含まれることをいうものとし、$a \leq b$という記号で表すものとします。

要素には、次の種類があります。

- 型$A$
- approximation element $\tilde A$
- union element $a|b|c \dots$ ただし$a$は要素とする

まず、型がinterface typeではない場合だけを考えることにします。そのため、型を表す記号$A, B, C, \dots$は全てnon-interface typeを表すものとします。

## 命題1

$$ A \leq B \Longleftrightarrow A = B $$

### 導出1

まず、

$$ A \leq B \Longleftrightarrow \rm{typeset}(A) \subset \rm{typeset}(B) \Longleftrightarrow \{A\} \subset \{B\} $$ 

です. これは $A \leq B$ の定義を当てはめ、またnon-interface typeである$A$の型セットが$\{ A\}$であることを使いました。あとは通常の集合の包含関係を考えれば、

$$ \{A\} \subset \{B\} \Longleftrightarrow A = B $$

であることがわかります。

## 命題2

$$ A \leq \tilde B \Longleftrightarrow \rm{underlying}(A) = B $$

### 導出2

$$ A \leq \tilde B \Longleftrightarrow \rm{typeset}(A) \subset \rm{typeset}(\tilde B) \Longleftrightarrow \{A\} \subset \rm{typeset}(\tilde B) $$

$$ \Longleftrightarrow A \in \rm{typeset}(\tilde B)\Longleftrightarrow \rm{underlying}(A) = B $$
