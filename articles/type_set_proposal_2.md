---
title: "Type Sets Proposalを読む(2) カノニカル形式編"
emoji: "💬"
type: "tech" # tech: 技術記事 / idea: アイデア
topics: [go, generics]
published: false
---

# はじめに

TBW

# interface/constraintに対して追加された制約の内容

interface定義において、union elementの要素となる型は、method set部分を持つ型であってはいけません。言い換えると、method set部分を持つ型は、次のようにスタンドアローンで現れなければいけません。

```go
// OKな例
type ConstraintGood interface {
    interface { // Method setを持つinterfaceはこの形でなら使える
        Method()
    }
}

// ダメな例
type ConstraintBad interface {
    int | interface { Method() } // methodをもつinterface型を、union elementの要素としてはいけない
}
```

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
