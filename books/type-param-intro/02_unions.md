---
title: "unions"
---

# genericなMax関数とunions

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

そこでGo言語は、型パラメータの導入と同時にインタフェース型として定義できる型も拡張することにしました。

```go
type Number interface {
    int | int32 | int64 | float32 | float64
}
```

この`Number`というインタフェースは、`int, int32, int64, float32, float64`という5種類の型によって「実装」され、これ以外の型によっては実装されません。
この文法要素`int | int32 | int64 | float32 | float64`のことを`unions`や`union element`と呼びます。

:::message

`|`を使わずに一つだけの型を書けば、その一つだけの型によってのみ実装されるインタフェースを定義できます。

```go
type Int interface {
    int
}
```

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

# まとめ