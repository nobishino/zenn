# Goの型推論を仕様書から読む

## 現在のGoでできる型推論

### 通常の関数呼び出し

### 既知の関数型の変数への代入

### 別の関数の引数として渡す

### 関数の戻り値として返す

### 関数型への変換

### ジェネリックメソッド

### 型引数を一部だけ明示する

### 制約から別の型引数を推論する

### 型推論が行われないもの

## 型推論の全体像

### インスタンス化

### 型推論が行われる場面

### 型推論に使われる情報

### 型方程式

### type unification

### 型制約の確認

### 型推論が成功する条件

### 仕様書の構成

## 型同士の関係と型方程式

### 代入可能性から得られる関係

### 型制約から得られる関係

### 型方程式の両辺

### ≡A

### ≡C

## dedupの例

### サンプルコード

### SliceとSの型方程式

### Sと型制約の型方程式

### Sの型引数

### Eの型引数

## bound type parameter

### bound type parameterとは

### unbound type parameterとは

### 今回の型推論で解く型パラメータ

### 複数のジェネリック関数をまとめて推論する場合

## 型方程式の作り方

### 関数呼び出し

#### 型の付いた実引数

#### untyped constant

#### ジェネリック関数を実引数にする場合

### ジェネリック関数と既知の関数型

#### 変数への代入

#### 引数として渡す

#### 戻り値として返す

#### 関数型への変換

### 型制約

## 型推論の2つのフェーズ

### 第1フェーズ

#### 型方程式をunificationする

### 第2フェーズ

#### untyped constantのconstant kind

#### default type

### constant kindが競合する場合

#### サンプルコード

#### 型推論の流れ

### 型引数が決まらない場合

## 型引数の簡約

### bound type parameterを含む型引数

### bound type parameterの置換

### 循環参照

#### サンプルコード

#### P1の型推論

#### P2の型推論

#### 型引数を簡約する

#### 型推論が失敗する

## type unification

### bound type parameterと型引数のmap

### 空のmapから始める

### 型引数をmapに追加する

### 既知の型引数を使う

### 複合型を再帰的に比較する

#### 仕様書のサンプル

#### 配列型

#### 構造体型

#### Pとstring

#### []Pと[]string

### type unificationが成功する条件

## exactとloose

### exact

### loose

### ≡Aのmatching mode

### element matching mode

## 型制約のunification

### 共通のunderlying type

#### サンプルコード

#### 型推論の流れ

### channel type

#### サンプルコード

#### 型推論の流れ

### 1つだけのtype term

#### サンプルコード

#### 型推論の流れ

### method

#### サンプルコード

#### 型推論の流れ

### 繰り返し処理

## Goのバージョンによる変更

### Go 1.18

### Go 1.21

### Go 1.27

## type unification rules

### matching mode

### bound type parameterではない型同士

#### identicalな型

#### 同じ構造の型

#### unbound type parameterを含む場合

### bound type parameter同士

#### 同じ型パラメータ

#### 型パラメータをjoinする場合

#### どちらにも型引数がある場合

### bound type parameterとその他の型

#### 型引数がまだない場合

#### 型引数がすでにある場合

#### interface typeの場合

#### defined typeを残す場合

### loose unification

#### exact unificationできる場合

#### defined typeとtype literal

#### interface type同士

#### interface typeとその他の型

#### 同じ構造の型

## まとめ

### 型推論の流れ

### 仕様書を読み直す
