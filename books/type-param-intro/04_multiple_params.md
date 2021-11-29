---
title: "TBW"
---

この章では、型パラメータを複数使うことでどのようなプログラムが書けるかを見ていきます。

# 複数の型パラメータを持つ関数

複数の型パラメータを持つ関数の代表例として、いわゆる`map`関数があります。

https://gotipplay.golang.org/p/EhW6WUiqvRx

```go
func main() {
	xs := []int{1, 2, 3, 4}
	ys := Map(xs, func(x int) int {
		return 3 * x
	})
	fmt.Println(ys)
	// Output:
	// [3 6 9 12]
}
func Map[U, V any](us []U, f func(U) V) []V {
	var result []V
	for _, u := range us {
		result = append(result, f(u))
	}
	return result
}
```

`U, V`の型制約はどちらも`any`なので、`[U, V any]`と省略することができます。
これは省略せずに書けば`[U any, V any]`と同じ意味です。

# 相互参照する型パラメータ

型パラメータは型制約を通して相互に参照することもできます。

型制約にはインタフェース型が使えますが、このインタフェース型自体が型パラメータを持つこともできます。

```go


```


https://gotipplay.golang.org/p/gBjT5abOk1n