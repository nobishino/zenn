----
title: ""
----

# `slices`パッケージ

本章からは、Goのexperimentalリポジトリにあるパッケージのコードリーディングをします。

- https://github.com/golang/exp/blob/master/slices/slices.go
- https://github.com/golang/exp/blob/master/slices/slices_test.go

# `Equal[E comparable](s1, s2 []E) bool`

まず最初の関数を読んでみましょう。

https://github.com/golang/exp/blob/master/slices/slices.go#L12-L28

```go
// Equal reports whether two slices are equal: the same length and all
// elements equal. If the lengths are different, Equal returns false.
// Otherwise, the elements are compared in increasing index order, and the
// comparison stops at the first unequal pair.
// Floating point NaNs are not considered equal.
func Equal[E comparable](s1, s2 []E) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i, v1 := range s1 {
		v2 := s2[i]
		if v1 != v2 {
			return false
		}
	}
	return true
}
```

# `Delete`

```go
// Delete removes the elements s[i:j] from s, returning the modified slice.
// Delete panics if s[i:j] is not a valid slice of s.
// Delete modifies the contents of the slice s; it does not create a new slice.
// Delete is O(len(s)-(j-i)), so if many items must be deleted, it is better to
// make a single call deleting them all together than to delete one at a time.
func Delete[S ~[]E, E any](s S, i, j int) S {
	return append(s[:i], s[j:]...)
}
```

# `~`の有無

さて、先程のEqualでは`~`を使っていなかったのですが、Deleteでは~を使って型パラメータ`S`を導入しています。これはなぜでしょうか？

答えは、`Equal`の場合は`~`を使うまでもなかったからです。次のコードを考えましょう。

```go
// Equal[E comparable](s1, s2 []E) bool
type S []int
s1 := S{1,2,3}
s2 := S{1,2,3}

Equal(s1,s2) {
```

一見すると`S`は`[]int`とは異なる型なので、`S`型の値`s1, s2`を`Equal`に渡すことはできないようにも見えます。しかし、これはGo言語の代入可能性のルールにより許可されているため上のコードは動作します。

この仕様は型パラメータとは関係なく従来からあるもので、例えば次のようなコードも動作します。

```go


```
