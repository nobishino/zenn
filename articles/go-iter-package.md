
---
title: "Go言語のiterパッケージ入門"
emoji: "😽"
type: "tech" # tech: 技術記事 / idea: アイデア
topics: [Go, Concurrency, MemoryModel]
published: false
---

# メモ:

iterProposal = [iter: new package for iterators](https://github.com/golang/go/issues/61897#issuecomment-1945059401)

# はじめに

2024/02/15に、Go言語の標準パッケージについてのProposalである[iter: new package for iterators](https://github.com/golang/go/issues/61897#issuecomment-1945059401)がacceptedになりました。

リリース時期は未確定ですが、最速であれば、2024年8月にリリースされるGo1.23から`iter`パッケージが利用できるようになります。

この記事はその`iter`パッケージについての入門記事です。具体的には、次のような問いに答えます。

- `iter`パッケージの典型的な使い方は何か？
- `iter`パッケージはなんのために作られるのか？
- `iter.Pull`はいつ使うのか？
- range over functionとの関係は？
- 
- `iter`パッケージは他のGoライブラリにどのように影響していくのか？

# 要約

# iterに関係するproposalたち

iterProposalは、GoのコレクションAPI

# 典型的な使い方

`iter`パッケージは、2つの型と2つの関数を提供します。iterProposalから引用したgodocを日本語訳すると、次のようになります。

```
Package iter provides basic definitions and operations related to iteration in
Go.

FUNCTIONS

func Pull[V any](seq Seq[V]) (next func() (V, bool), stop func())
    Pull converts the “push-style” iterator sequence seq into a “pull-style”
    iterator accessed by the two functions next and stop.

    Next returns the next value in the sequence and a boolean indicating whether
    the value is valid. When the sequence is over, next returns the zero V and
    false. It is valid to call next after reaching the end of the sequence
    or after calling stop. These calls will continue to return the zero V and
    false.

    Stop ends the iteration. It must be called when the caller is no longer
    interested in next values and next has not yet signaled that the sequence is
    over (with a false boolean return). It is valid to call stop multiple times
    and when next has already returned false.

    It is an error to call next or stop from multiple goroutines simultaneously.

func Pull2[K, V any](seq Seq2[K, V]) (next func() (K, V, bool), stop func())
    Pull2 converts the “push-style” iterator sequence seq into a “pull-style”
    iterator accessed by the two functions next and stop.

    Next returns the next pair in the sequence and a boolean indicating whether
    the pair is valid. When the sequence is over, next returns a pair of zero
    values and false. It is valid to call next after reaching the end of the
    sequence or after calling stop. These calls will continue to return a pair
    of zero values and false.

    Stop ends the iteration. It must be called when the caller is no longer
    interested in next values and next has not yet signaled that the sequence is
    over (with a false boolean return). It is valid to call stop multiple times
    and when next has already returned false.

    It is an error to call next or stop from multiple goroutines simultaneously.


TYPES

type Seq[V any] func(yield func(V) bool)
    Seq is an iterator over sequences of individual values. When called as
    seq(yield), seq calls yield(v) for each value v in the sequence, stopping
    early if yield returns false.

type Seq2[K, V any] func(yield func(K, V) bool)
    Seq2 is an iterator over sequences of pairs of values, most commonly
    key-value pairs. When called as seq(yield), seq calls yield(k, v) for each
    pair (k, v) in the sequence, stopping early if yield returns false.
```

# 参考資料

筆者が参考にした資料と、参考になりそうな資料を挙げておきます。

| タイトルとリンク                                            | 概要                                                                                                    |
|------------------------------------------------------------|---------------------------------------------------------------------------------------------------------|
| | |

# 最後に

執筆にあたり次の方から情報やフィードバックをいただきました。ありがとうございます。


もちろん、記述の誤りなどについてのすべての責任は筆者にあります。
