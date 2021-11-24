---
title: "簡単な例"
---

# Goの型パラメータの基本原則

型パラメータの基本事項についてはType Parameters Proposalの冒頭に挙げられていますが、その中でも筆者が特に重要と考えるのは次の2つです。この2つを覚えればGoの型パラメータを十分に使うことができると思います。

- 「関数」と「型」は「型パラメータ」を持つことができる。
- 「型パラメータ」の満たすべき性質は「インタフェース型」を「型制約」として使うことで表す。

# 具体例1: 型パラメータを持つ関数

まず「関数」と「型」について具体例を見てみましょう。型パラメータを持つ関数の例は次のようなものです。

https://gotipplay.golang.org/p/NWxONCa85DL

```go
func main() {
	fmt.Println(f([]MyInt{1, 2, 3, 4}))
    // Output:
    // [1 2 3 4]
}

// fは型パラメータを持つ関数
// Tは型パラメータ
// インタフェースStringerは、Tに対する型制約として使われている
func f[T Stringer](xs []T) []string {
	var result []string
	for _, x := range xs {
        // xは型制約StringerによりString()メソッドが使える
		result = append(result, x.String())
	}
	return result
}

type Stringer interface {
	String() string
}

type MyInt int

// MyIntはStringerを実装する
func (i MyInt) String() string {
	return strconv.Itoa(int(i))
}
```

# 具体例2: 型パラメータを持つ型

TBW
Setを書く？

# Go1.17でできなかったこと

ここで少し型パラメータのモチベーションを知るためにGo1.17でのコードを考えてみます。

## インタフェース型のスライスを受け取る関数

まず、Go1.17において次のインタフェースと関数を考えます。

```go
type Stringer interface {
    String() string
}

func f(xs []Stringer) []string {
    var result []string
    for _, x := range xs {
        result = append(result, x.String())
    }
    return result
}
```

また、次のように`Stringer`を実装する型を用意します。

```go
type MyInt int

// MyIntはStringerを実装する
func(i MyInt) String() string {
    return strconv.Itoa(int(i))
}
```

このとき次のように、`MyInt`のスライスを`f`に渡すことはできるでしょうか？

```go
xs := []MyInt{0,1,2}
f(xs) // fは[]Stringerを受け付ける
```

このようなコードは書けません。`MyInt`は`Stringer`を実装するので`MyInt`型の値は`Stringer`型の変数に代入可能ですが、`[]MyInt`型の値は`[]Stringer`型の変数に代入できないためです。

Go1.17で`[]Stringer`を一般的に扱う関数を書くには、次の`f2`のように空インタフェース型`interface{}`を受け取るようにするしかありませんでした。この関数`f2`にはどんな型の値でも渡せてしまうので、関数の利用側で間違った値を渡さないように気をつけなければいけません。

```go
// 【注意】 Stringerを実装する型Tのスライス[]Tだけを渡すこと
func f2(xs interface{}) {
    if vs,ok := xs.([]MyInt); ok {
        // vsに関する処理
    }
    // ... 
}
```

:::message
`f2`が`[]MyInt`以外のスライス型を受け取るようにするには、それぞれの型についての[型アサーション](https://go.dev/ref/spec#Type_assertions)を書く必要があります。

```go
if vs, ok := xs.([]Stringer); ok
```

のようなアサーションを書くこと自体はできますが、こう書いても`[]MyInt`型の値を渡したときには`!ok`となります。

型スイッチ文を使う場合も、渡すかもしれない具体的な型ごとにcase節が必要です。
:::

## 型パラメータによる記述

> Stringerインタフェースを実装する型Tのスライス[]Tだけを渡すこと

という条件付けは、型パラメータを使うと次のように記述できます。

```go
// fは型パラメータを持つ関数
// Tは型パラメータ
// インタフェースStringerは、Tに対する型制約として使われている
func f[T Stringer](xs []T) []string {
	var result []string
	for _, x := range xs {
        // xは型制約StringerによりString()メソッドが使える
		result = append(result, x.String())
	}
	return result
}
```

この`f`には`[]MyInt`型の値だけでなく、何のコードの変更もなしに`Stringer`を実装する型`T`のスライス`[]T`を渡せますし、そうでない型の値は渡すことができないため、安心してプログラミングすることができます。

# まとめ




# TL;DRs

- 「関数」と「型」は「型パラメータ」を持つことができる。
- 「型パラメータ」の満たすべき性質は「インタフェース型」を「型制約」として使うことで表す。
- 