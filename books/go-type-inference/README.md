# Goの型推論を仕様書から読む

## この本の方針

第1章と第2章で、現在のGoでできる型推論と処理の全体像を先に確認します。

第3章以降は、次の2つの仕様書を上から順に読みます。

- [Type inference](https://go.dev/ref/spec#Type_inference)
- [Type unification rules](https://go.dev/ref/spec#Type_unification_rules)

仕様書の各センテンスは、割り当てられた章で少なくとも一度取り上げます。READMEでは、各章が担当する原文の連続範囲を段落単位で示します。仕様書の改訂時にはこの対応を見直します。

## 第1章 現在のGoでできる型推論

## 第2章 型推論の全体像

## 第3章 型同士の関係と型方程式

カバーする原文の全範囲:

[Type inference](https://go.dev/ref/spec#Type_inference)の冒頭4段落。型引数を省略できる条件と成功条件から始まり、型同士の関係を型方程式の集合にして解くという説明まで。

> “Otherwise, type inference fails and the program is invalid.”

この文を含む冒頭から、`type equations`を解くという説明までを扱います。

## 第4章 dedupの例

カバーする原文の全範囲:

[Type inference](https://go.dev/ref/spec#Type_inference)の`dedup`の例全体。サンプルコードの提示から、`S ➞ Slice`と`E ➞ int`を得る説明まで。

> `Slice ≡A S`  
> `S ≡C ~[]E`  
> `S ➞ Slice`  
> `E ➞ int`

## 第5章 bound type parameter

カバーする原文の全範囲:

[Type inference](https://go.dev/ref/spec#Type_inference)のbound type parameterを定義する2段落。解く対象の定義から、型方程式はbound type parameterについてだけ解かれるという説明まで。

> `bound type parameters`

## 第6章 型方程式の作り方

カバーする原文の全範囲:

[Type inference](https://go.dev/ref/spec#Type_inference)の型推論が対応する利用場面と、場面ごとに作られる入力を列挙する範囲。ジェネリック関数の呼び出しと関数型が要求される文脈の説明から、`Pₖ ≡C Cₖ`の生成規則まで。

> `typeof(pᵢ) ≡A typeof(aᵢ)`  
> `(cⱼ, Pₖ)`  
> `typeof(f) ≡A T`  
> `Pₖ ≡C Cₖ`

## 第7章 型推論の2つのフェーズ

カバーする原文の全範囲:

[Type inference](https://go.dev/ref/spec#Type_inference)の型付きオペランドを優先するという説明から、第1フェーズ、第2フェーズ、全型引数が見つからない場合の失敗、`Pₖ ➞ Aₖ`という結果の提示まで。

> `two phases`  
> `Pₖ ➞ Aₖ`

## 第8章 型引数の簡約

カバーする原文の全範囲:

[Type inference](https://go.dev/ref/spec#Type_inference)の型引数がbound type parameterを含み得るという説明から、繰り返し置換による簡約と、循環参照による失敗の説明まで。

> `cyclic references`

## 第9章 type unificationとmap

カバーする原文の全範囲:

[Type inferenceのType unification](https://go.dev/ref/spec#Type_inference)の冒頭3段落。型方程式の左右を再帰的に比較する説明から、bound type parameterと推論済み型引数のmapを参照・更新し、成功または失敗へ進む説明まで。

> `P ➞ A`

## 第10章 複合型のunification

カバーする原文の全範囲:

[Type inferenceのType unification](https://go.dev/ref/spec#Type_inference)の配列と構造体を使った例全体。型方程式の提示から、空のmapを更新しながら`P ➞ string`を得て型推論に成功する説明まで。

> `[10]struct{ elem P, list []P } ≡A [10]struct{ elem string; list []string }`

## 第11章 exactとlooseと型制約

カバーする原文の全範囲:

[Type inferenceのType unification](https://go.dev/ref/spec#Type_inference)のexactとlooseの導入から、`≡A`の比較、`≡C`に対する4つの規則、型制約由来の式から新しい型引数が得られる限り処理を繰り返すという説明まで。

> `exact` / `loose`  
> `X ≡A Y`  
> `P ≡C C`

## 第12章 type unification rulesの読み方

カバーする原文の全範囲:

[Type unification rules](https://go.dev/ref/spec#Type_unification_rules)の冒頭2段落。規則の目的と位置づけから、exactとlooseというmatching mode、および`≡A`でelement matching modeがexactへ変化する説明まで。

> `matching mode` / `element matching mode`  
> `exact` / `loose`

## 第13章 bound type parameterではない型のexact unification

カバーする原文の全範囲:

[Type unification rules](https://go.dev/ref/spec#Type_unification_rules)のbound type parameterではない2型がexactに一致する条件の全体。identical、同一構造、片方だけがunbound type parameterである場合の3条件。

> `identical`  
> `unbound type parameter`

## 第14章 bound type parameterのunification

カバーする原文の全範囲:

[Type unification rules](https://go.dev/ref/spec#Type_unification_rules)のbound type parameterを含む比較の全体。両辺がbound type parameterである場合の3条件から、片方がbound type parameterである場合、interfaceの扱い、defined typeによる推論結果の置換まで。

> `P` / `T` / `A`  
> `interface types` / `defined types`

## 第15章 loose unification

カバーする原文の全範囲:

[Type unification rules](https://go.dev/ref/spec#Type_unification_rules)のloose unificationの全体。exactに一致する場合から、defined typeとtype literal、interface同士、interfaceとその他の型、同一構造の型を比較する各条件まで。

> `defined type` / `type literal`  
> `interface` / `type terms` / `method set`

## 第16章 まとめ

第3章から第15章までの流れをまとめ、仕様書を上から読み直します。
