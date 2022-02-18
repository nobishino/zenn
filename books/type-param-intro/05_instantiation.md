---
title:"TBW"
---

# インスタンス化と型推論

次のコードは以前の章で出てきた`Map`関数です。

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

この`main`関数で呼び出されている`Map`関数は、**暗黙的に**`U = int, V = func(int) int)`という型引数によってインスタンス化されています。
型パラメータ化された関数は、呼び出す際には必ずインスタンス化しなければいけません。

型引数は明示的に指定することもでき、上記の`Map`呼び出しは次のコードと等価です。

```go
	ys := Map[int, func(int) int](xs, func(x int) int {
		return 3 * x
	})
```

`Map[int ,func(int) int]`の部分で明示的に型引数を指定していますね。
この型引数が省略可能なのは、型推論により自動的に型引数を決定できているからです。
型推論はいつでも成功するわけではなく、型引数を明示的に書かないと動作しないコードもあります。

この章では、型推論のアルゴリズムについて説明し、型推論がうまくいく場合といかない場合についてその理由を理解したいと思います。

# 型推論アルゴリズムの概要

- 制約型推論: すでにわかっている型引数から残りの型引数を推論する仕組み
- 関数型推論: 関数の引数の型から型引数を推論する仕組み

どちらの型推論も、型結合(type unification)というものを利用します。
「型結合」というのは筆者が今考えた訳なので、今後は訳さずにtype unificationと書くことにします。

型推論のinput/output

- 型パラメータのリスト
- すでに知られている型引数で初期化されたsubstitution map
- 関数呼び出しの場合、その関数の引数

1. 制約型推論を行う
1. 型のある引数、つまり型なし定数ではない引数を使った関数型推論を行う
1. 制約型推論を行う
1. 型なし定数の引数を使った関数型推論を行う
1. 制約型推論を行う

# substitution map

substitution mapとは、型パラメータ`P`を型引数`A`に対応づけるエントリ`P->A`の集まりです。

はじめ、substitution mapは明示的に決められた型引数に対応するエントリだけを持っています。

例えば、`Map[U, V any]`という関数に対してMap[int]`のように型引数を最初の1つだけ指定することができるのですが、この場合のsubstition mapは

```
U -> int
```

としてスタートします。

substitution mapが全ての型パラメータのエントリを含むとき、substitution mapが"full"であると言います。
上記のsubstitution mapはfullではありません。

型推論のゴールは、残りの型引数を推論することで、substitution mapをfullにすることです。
型推論の途中のステップが「失敗」したり、最後まで進んだけれどもsubstitution mapがfull出なかった場合、型推論は失敗です。

# type unification

type unificationは型推論アルゴリズムの一部であり、1回の型推論において一般には複数回のtype unificationが行われます。
type unificationの入力は2つの型であり、出力はsubstitution mapのエントリです。

**入力**

- 2つの型(型パラメータを含んでいても良い)

**出力**

- substitution mapのエントリ

type unificationは、2つの型を「等価」にするようなsubstition mapエントリを見つけるためのアルゴリズムです。

## equivalence(等価性)

2つの型が**等価(equivalent)**であるのは、次のいずれかが成り立つときです。

どちらの型も型パラメータを含まず、

- 2つの型が同一(identical)であるとき
- 2つの型がchannel型であって、それらが方向を無視すれば同一(identical)であるとき
- 2つの型のunderlying typeが等価であるとき

等価性を考えるときには、すでに知られている(substitution mapに含まれる)型引数は対応する型パラメータに代入した上で等価性を判定します。

## type unificationの手順

まず、型のペアの構造を比較します。肩パラメータを無視した時の型の構造は同一でなければならず、かつ、型パラメータでない型は等価でなければいけません。
型パラメータはもう一方の型の一部にマッチさせることができます。そのようなマッチングは1つの新しいsubstition mapエントリになります。

構造が異なったり、型パラメータ以外の型が等価でない場合、type unificationは失敗します。

### 具体例

`T1, T2`を型パラメータとします。

# 制約型推論(constraint type inference)

# 関数型推論(function type inference)

# 具体例

## 具体例1

```go
func f[V ~T, T constraints.Number](v V) {}

func main(){
	_ = f[int]
}
```

まず明示的に与えた型引数`int`によりエントリが作られます。

```
V -> int
```

次にステップ1の制約型推論が行われます。

`V`はstructural type`T`を持つので`V`と`T`をunifyします。これにより`V -> T, T -> V`です。




# 型推論が成功した後でコンパイルエラーが発生することもある

