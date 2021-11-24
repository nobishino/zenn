---
title: "Draft" # 記事のタイトル
emoji: "😸" # アイキャッチとして使われる絵文字（1文字だけ）
type: "tech" # tech: 技術記事 / idea: アイデア記事
topics: ["go"] # タグ。["markdown", "rust", "aws"]のように指定する
published: false # 公開設定（falseにすると下書き）
---

# formula

- 型$\rm{T}$のメソッドセットを$\rm{ms}(T)$
- 型$\rm{T}$の型セットを$\rm{ts}(T)$

という記号を使うと、インタフェース型 $\rm{I}$のメソッドセットは次の式で定義される。

$$ \rm{ms}(I) = \bigcap_{x \in \rm{ts}(I)} \rm{ms}(x) $$
$$ = \rm{ms}(x_1) \cap \rm{ms}(x_2) \cap \dots $$

# 例題

```go
type MyInt int 
func (MyInt) F()
type MyIntIF {
    MyInt
}
```
と定義するときの`MyIntIF`のメソッドセットは？

## 答え

先に型セットを求める.

$$ \rm{ts}(MyIntIF) = \rm{ts}(MyInt) = \{ MyInt \} $$

これを使うと,

$$
\rm{ms}(MyIntIF)
= \bigcap_{x \in \rm{ts}(MyIntIF)}\rm{ms}(x) 
= \bigcap_{x \in \{\rm{MyInt\}}}\rm{ms}(x)
= \rm{ms}(MyInt)
= \{\rm{F}()\}
$$