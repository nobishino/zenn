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

Type Sets Proposalは、Go言語のGenericsの実現方法に関わるProposalです。

より具体的に書くと、Type Sets Proposalとは、Type Parameters Proposalのおける型制約の表現手段であった"type list"を置き換えて改善するProposalです。つまり、これがAcceptedになると、現在のType Parameters Proposalの内容の一部がこのProposalの内容に置き換えられて採用されることになります。

そして、このType Sets ProposalはAcceptされることがおそらく確実([likely accept](https://github.com/golang/go/issues/45346#issuecomment-880098162))で、"Proposal-FinalCommentPeriod"のラベルがつけられています。

# この記事のテーマ: Type Sets Proposalに加えられた変更

その具体的な内容はdescriptionにあるのですが、この記事で紹介したいのはそこからさらに加えられた変更内容です。その内容は、griesemer氏による次のコメントで詳しく説明されています。

https://github.com/golang/go/issues/45346#issuecomment-862505803

丁寧に説明はされているのですが、無駄がなさすぎて少し行間の空いたように感じられる記述でもあるので、より理解しやすく紹介することを試みたいと思います。
# interface/constraintに対して制約を追加する

変更内容を一言で言うと、「interface/constraintとして許容されるパターンが当初のType Sets Proposalよりも狭く限定される」と言う変更です。具体的には、次のように制約されます。

interface定義において、union element(以下、unionsと書きます)の項となる型は、method set部分を持つ型であってはいけません。言い換えると、method set部分を持つ型は、スタンドアローンで現れなければいけません。

言っていることがわかりにくいと思いますが、proposalのコメントで具体例を書いてくれていますので借りてきます。


```go
// OKな例
type ConstraintGood interface {
    interface { // Method set部分を持つinterfaceはこの形でなら使える
        Method()
    }
}

// ダメな例
type ConstraintBad interface {
    int | interface { Method() } // methodをもつinterface型を、union elementの要素としてはいけない
}
```

ここに2つの例を挙げましたが、当初のType Sets Proposalでは後者の書き方も許されていました。ですが、最新のType Sets Proposalではこれは許されなくなります。(コンパイルエラーになると思われます)

# なぜこのように制約するのか

なぜこのような制約が追加されたのでしょうか？要約すると次のようになります。

- この形のインタフェースは、「標準形」に変形することができる
- 「標準形」のインタフェース同士は、型セットの包含関係を比較的簡単に計算できる
- ゆえに、次のようなコードがコンパイルできるかどうかの判定が容易になる

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
