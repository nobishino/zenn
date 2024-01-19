---
title: "Go言語でdata raceを発生させたときに起きること"
emoji: "😽"
type: "tech" # tech: 技術記事 / idea: アイデア
topics: [Go, Memory Model]
published: false
---

# はじめに

この記事は、プログラミングにおいて特に難しいことの1つである「並行処理」に関する記事です。

特に、「並行処理」を行うときに意図せず発生させてしまいやすい「data race」について書きます。data raceがどのような驚くべき問題を引き起こすかを、簡単に動かせるサンプルコードで具体的に見ていきます。

プログラム言語としてGoを使いますが、Goに限らず当てはまる内容も多いと思います。ただし、data raceに関してはプログラム言語ごとに微妙なアプローチの違いがあるので、それについては最後に少しだけ補足します。

大前提として、ソフトウェア開発では、data raceを一切発生させない状態を目指すべきだと思います。[Data Race Detector](https://go.dev/doc/articles/race_detector)を使って十分なテストを行えば、そのような状態に近づくことができます。しかし、実際にdata raceが存在するとどのようなことが起こりうるのかを詳しく知っている人は少ないのではないでしょうか。そこでこの記事ではそうした例をいくつも挙げることで、data raceをなくすことへのモチベーションを高めたいと思います。

## 注意: 誤解してほしくないポイント2つ

記事が長くなるので、最初に誤解してほしくないポイントを2つ書いておきます。

- この記事では「驚くような動き」を挙げていきますが、これはdata raceが存在する状況ではプログラム言語に関わらず発生するものです。Go言語だから発生する訳ではありません。
- この記事で挙げる「驚くような動き」を心配しなければならないのはdata raceが存在する場合であって、「並行処理を使うといつもこのようなことが起こりうる」訳ではありません。並行処理を使っていても、data raceを発生させていなければ「驚くような動き」は起きません。

# data raceとは何か

「data race」について、この記事を読むのに必要十分な程度に説明します。

::: message
本当に詳しく知りたい方は、Goの公式ドキュメントである[The Go Memory Model](https://go.dev/ref/mem)と、この2022年のアップデートをするにあたって[Russ Cox氏が書いたブログシリーズ](https://research.swtch.com/mm)を読むのが良いと思います。もしくは、日本語資料では筆者の作成したスライド「[よくわかるThe Go Memory Model](https://docs.google.com/presentation/d/1UjL5jTqreNrFpulVi6l_H5vY_Bcz9jQriL65gZs1zFM/edit?usp=sharing)」があります。
:::



## data raceと間違われやすいもの



# データの一貫性が壊れる例

データの一貫性が壊れるとは、総じていえば、次のような代入文が意図した通りに働かないことです。

```go
variable = value
```

私たちが普通にプログラミングするとき、代入文の前後の変数`variable`は「全く代入がされていないか、完全に代入が終わっているか」のどちらかであることを期待すると思います。

当たり前すぎて何を言っているかわからないかもしれませんが、「誰かが`variable`を読み取ったとき、上記の代入が中途半端に行われた状態を観測することはないだろう」と期待しているという意味です。

私たちが当たり前に依拠しているこの前提は、data raceのあるプログラムでは必ずしも成り立ちません。そのことを具体的に見ていきましょう。

## 中途半端に書き込まれた構造体を読み取る

次の関数を見てください。構造体型`Pair`の変数`p`があります。また、メインのgoroutineと`go`文で起動されるもう1つのgoroutineがあります。片方のgoroutineで`p`に書き込み、メインのgoroutineで`p`を読み取っています。

```go
func structCorrption() string {
	type Pair struct {
		X int
		Y int
	}
	arr := []Pair{{X: 0, Y: 0}, {X: 1, Y: 1}} // arrにあるのは{0, 0}か{1, 1}だけ
	var p Pair // この変数を使う
	done := make(chan struct{})
	go func() { // writer goroutine
		for i := 0; ; i++ {
			select {
			case <-done:
				return
			default:
				p = arr[i%2]
			}
		}
	}()
    // reader
	for {
		read := p // この瞬間の変数pの観測値をローカル変数readにコピーする
		switch read.X + read.Y {
		case 0, 2: 
        // {X:0, Y:0}または{X:1, Y:1}の場合はここに入るので、何もしない
		default:
			close(done)
            // 
			return fmt.Sprintf("struct corruption detected: %+v", read)
		}
	}
}
```

このサンプルに限らず、この記事のサンプルコードでは2つのgoroutineを使い、片方で書き込み、もう片方で読み込みを行います。そこで書き込む方をwriter goroutine、読み込む方をreader goroutineと呼ぶことにしましょう。

writer goroutineが`p`に代入するのは`Pair{X: 0, Y: 0}`か`Pair{X: 1, Y: 1}`のどちらかです。reader goroutineはこれ以外の値を観測したときにメッセージを返して終了するようになっています。

readerが終了しない限りwriterも終了しないようになっていますから、writerが書き込む2通りの値だけがreaderによって読まれている限り、このプログラムは無限ループするでしょう。実際にはどうなるでしょうか？

次のplaygroundを動かしてみてください。

https://go.dev/play/p/EYXe5UWzKEZ

メッセージを返して終了したと思います。

> struct corruption detected: {X:0 Y:1}

> Program exited.

readerが読み取った値は驚くべきことに`{X:0, Y:1}`というものです。

このサンプルコードには何の害もありませんが、構造体の意味によっては、そもそも存在してはいけない状態というものがあって、それを意図せず読み取ってしまうかもしれません。

## 文字列を`print`すると`panic`する

この例はbudougumi0617さんのブログ[[Go] stringの比較でヌルポのpanicが発生する（こともある） #横浜Go読書会](https://budougumi0617.github.io/2021/03/31/go-string-null-pointer-panic/)で説明されているものを参考に作成しました。


```go
package main

import "fmt"

func main() {
	var s string
	// writer
	go func() {
		arr := [2]string{"", "hello"}
		for i := 0; ; i++ {
			s = arr[i%2]
		}
	}()
	// reader
	for {
		fmt.Println(s)
	}
}
```

https://go.dev/play/p/KLR5U0rbzEN 

上記のPlaygroundで実行すると、次のように`panic`するのではないでしょうか。

> panic: runtime error: invalid memory address or nil pointer dereference
> [signal SIGSEGV: segmentation violation code=0x1 addr=0x0 pc=0x45d33c]
> goroutine 1 [running]:
> fmt.(*buffer).writeString(...)

`string`型の値は複数の部分からなっており、文字列の長さを表す部分とバイト列の先頭へのポインタを持っています。

長さを表す部分とそのポインタ部分が一緒に更新されれば問題ないのですが、reader側から中途半端に片方だけ更新された状態を観測してしまうと、nil pointer dereferenceが発生します。

## あるはずのスライスの要素の参照で`panic`する

## mapの読み書きで`panic`する

次に`map`型を扱います。実は`map`型は少し特別で、race detectorを使うまでもなく、data raceが発生したらその時点で`panic`するようになっています。

## 型assertしたはずのinterfaceの動的値がおかしい

inteface型の例として、`any`型の変数の例をあげます。writer側では、異なる型の値を交互に代入してみましょう。reader側では型スイッチ文を使って動的型を確かめてから、動的値が期待通りかどうかチェックします。

```go
func interfaceCorruption() string {
	var x any

	done := make(chan struct{})
	go func() {
		arr := []any{1, "hello"}
		for i := 0; ; i++ {
			select {
			case <-done:
				return
			default:
				x = arr[i%2]
			}
		}
	}()
	for {
		read := x
		switch r := read.(type) {
		case int:
			if r != 1 {
				return fmt.Sprintf("unexpected int value: %d", r)
			}
		case string:
			if len(r) != 5 {
				return fmt.Sprintf("unexpected string length :%d", len(r))
			}
		case nil:
		default:
			close(done)
			return fmt.Sprintf("strange type detected: %+v", read)
		}
	}
}
```

`int`型の`1`と`string`型の`"hello"`だけを交互に代入しているのですから、reader側で`int`と判定すれば値は`1`だし、`string`型と判定すれば長さは`5`になりそうなものですが、次のPlaygroundで実行するとそうならないケースがレポートされます。

これは、interface型の値には「型の情報(動的型)」と「値の情報(動的値)」の2つの部分があり、この2つの部分を中途半端に更新した状態をreaderが観測することによって起こっています。


# その他直感に反する結果

このセクションでは一貫性とは別な観点で直感に反する結果をもたらすdata raceサンプルコードを挙げます。

それぞれのサンプルにはよく使われる名前がついているので、その名前を見出しにしています。興味があれば調べてみてください。

## Message Passing

## Store Buffer

# 参考資料

## Race Detector関係

- 公式
- Looking Inside
- Go Mistakes

## メモリーモデル関係

- The Go Memory Model
- RSC
- 発表資料



