---
title: "Go言語でdata raceが起きるときに起きる（かもしれない）こと"
emoji: "😽"
type: "tech" # tech: 技術記事 / idea: アイデア
topics: [Go, Concurrency, MemoryModel]
published: false
---

# はじめに

この記事は、プログラミングにおいて特に難しいことの1つである「並行処理」に関する記事です。特に、「並行処理」を行うときに意図せず発生させてしまいやすい「data race」について書きます。data raceがどのような驚くべき問題を引き起こすかを、簡単に動かせるサンプルコードで具体的に見ていきます。

プログラム言語としてGoを使いますが、内容的にはGoに限らず当てはまると思います。ただし、data raceに関してはプログラム言語ごとに微妙なアプローチの違いがあるので、それについては最後に少しだけ補足します。

ところで、ソフトウェア開発では、data raceを一切発生させない状態を目指すべきだと筆者は考えています。[Data Race Detector](https://go.dev/doc/articles/race_detector)を使って十分なテストを行えば、そのような状態に近づくことができます。

しかし、実際にdata raceが存在するとどのようなことが起こりうるのかを詳しく知っている人は少ないのではないでしょうか。そこで、data raceによって起こる「驚くような動き」をいくつも挙げることで、data raceをなくすことへのモチベーションを高めたいと思います。

## 注意: 誤解してほしくないポイント2つ

記事が長くなるので、誤解してほしくないポイントを最初に2つ書いておきます。

:::message
この記事で挙げる「驚くような動き」は、ほとんどの（全ての？）プログラム言語で発生します。Go言語だから発生するというわけではありません。
:::


:::message
この記事で挙げる「驚くような動き」を心配しなければならないのはdata raceが存在する場合であって、「並行処理を使うといつもこのようなことが起こりうる」わけではありません。

つまり、並行処理を使っていても、data raceを発生させていなければ「驚くような動き」は起きません。
:::

# data raceとは何か

「data race」について、この記事を読むのに必要十分な程度に説明します。

> A data race is defined as a write to a memory location happening concurrently with another read or write to that same location, unless all the accesses involved are atomic data accesses as provided by the sync/atomic package. 

https://go.dev/ref/mem#overview

これを訳すと概ね次のようになります:

**data raceは、あるメモリー位置への書き込みであって、その同じ位置に対する他の読み込みまたは書き込みと並行に起きるものとして定義されます。** ただし、すべてのアクセスが`sync/atomic`パッケージで提供されるアトミックなデータアクセスである場合を除きます。


もっと簡単に言ってしまうと、次の2つのいずれかに当てはまるものはdata raceです。

- 同一のメモリー位置に対する並行な2つの書き込み
- 同一のメモリー位置に対する並行な読み込みと書き込み

それぞれについて、シンプルな例を挙げておきます。

```go
// 同一メモリー位置に対する並行な2つの書き込み
package main

var x int

func main() {
	go func() {
		x = 1 // 書き込み1
	}()
	x = 2 // 書き込み2
}
```

https://go.dev/play/p/wtuAR68yt8B

```go
// 同一メモリー位置に対する並行な読み込みと書き込み
package main

import "fmt"

var x int

func main() {
	go func() {
		x = 1 // 書き込み
	}()
	fmt.Println(x) // 読み込み

}
```

https://go.dev/play/p/wmCAMpPYLVV

この記事を読むにはこの2つがdata raceであることがわかれば十分です。一応細かい補足をいくつか書いておきます。

:::message
「同一のメモリー位置」というのは「同一の変数」と似た意味ですが、正確にはそれよりも粒度が細かいです。

構造体の変数は複数の「メモリー位置」にまたがっていますし、スライス型の変数も複数のメモリー位置にまたがっています。
:::

:::message
「並行(concurrent)」の正確な意味は、[The Go Memory Model](https://go.dev/ref/mem)に書かれています。

本当に詳しく知りたい方は、Goの公式ドキュメントである[The Go Memory Model](https://go.dev/ref/mem)と、この2022年のアップデートをするにあたって[Russ Cox氏が書いたブログシリーズ](https://research.swtch.com/mm)を読むのが良いと思います。

もしくは、日本語資料では筆者の作成したスライド「[よくわかるThe Go Memory Model](https://docs.google.com/presentation/d/e/2PACX-1vS2FIFiNgrpRpm1bPO3KpVzJYX4vEFhpyttvBTsTq15BwFmWvW0Q0W4udSf3pJMQyzZJicE5LcR5_cY/pub?start=false&loop=false&delayms=3000)」があります。
:::

## data raceと間違われやすいもの

### 並行な2つの読み込み

次の2つの読み込みは並行ですが、2つの読み込みの組み合わせはdata raceにはなりません。

```go
// 同一メモリー位置に対する並行な2つの読み込み
package main

import "fmt"

var x int

func main() {
	x = 1
	go func() {
		fmt.Println(x) // 読み込み1
	}()
	fmt.Println(x) // 読み込み2

}
```

https://go.dev/play/p/goEYaakTEat

:::message
なお、この例はプログラム全体としてもdata raceを発生させません。
:::

### 「競争」しているけどdata raceではない例

次の2つの書き込みはどちらが先に行われるかわかりませんが、data raceではありません。

```go
package main

import (
	"sync"
)

var x int
var mu sync.Mutex

func main() {
	go func() {
		mu.Lock()
		x = 2 // 書き込み1
		mu.Unlock()
	}()
	mu.Lock()
	x = 1 // 書き込み2
	mu.Unlock()
}
```

https://go.dev/play/p/7aV0gncOpR_O


# data raceによってデータの一貫性が壊れる例

データの一貫性が壊れるとは、総じていえば、次のような代入文の結果を意図した通りに読み取れないことです。

```go
variable = value
```

私たちが普通にプログラミングするとき、代入文の前後の変数`variable`は「全く代入がされていないか、完全に代入が終わっているか」のどちらかであることを期待すると思います。

当たり前すぎて何を言っているかわからないかもしれませんが、 **「誰かが`variable`を読み取ったとき、上記の代入が中途半端に行われた状態を観測することはないだろう」と期待している** という意味です。

私たちが当たり前に依拠しているこの前提は、data raceのあるプログラムでは必ずしも成り立ちません。そのことを具体的に見ていきましょう。

## 中途半端に書き込まれた構造体を読み取る

次の関数を見てください。構造体型`Pair`の変数`p`があります。また、メインのgoroutineと`go`文で起動されるもう1つのgoroutineがあります。片方のgoroutineで`p`に書き込み、メインのgoroutineで`p`を読み取っています。

```go
func structCorruption() string {
	type Pair struct {
		X int
		Y int
	}
	arr := []Pair{{X: 0, Y: 0}, {X: 1, Y: 1}}
	var p Pair // 共有変数
	
	// writer
	go func() {
		for i := 0; ; i++ {
			// 代入するのは{X: 0, Y: 0}, {X: 1, Y: 1}のどちらかのみ
			p = arr[i%2] 
		}
	}()
	
	// reader
	for {
		read := p
		switch read.X + read.Y {
		case 0, 2: 
			// {X: 0, Y: 0}, {X: 1, Y: 1}のどちらかならば、
			// このケースに入るので何も起きない。
		default:
			return fmt.Sprintf("struct corruption detected: %+v", read)
		}
	}
}
```

このサンプルに限らず、この記事のサンプルコードでは2つのgoroutineを使い、片方で書き込み、もう片方で読み込みを行います。そこで書き込む方をwriter、読み込む方をreaderと呼ぶことにしましょう。

writerが`p`に代入するのは`Pair{X: 0, Y: 0}`か`Pair{X: 1, Y: 1}`のどちらかです。readerはこれ以外の値を観測したときにメッセージを返して終了するようになっています。

readerが終了しない限りwriterも終了しないようになっていますから、writerが書き込む2通りの値だけがreaderによって読まれている限り、このプログラムは無限ループするでしょう。実際にはどうなるでしょうか？

次のplaygroundを動かしてみてください。

https://go.dev/play/p/lWtoA_ikaQG

メッセージを返して終了したと思います。

```
struct corruption detected: {X:0 Y:1}

Program exited.
```

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

上記のPlaygroundで実行すると、次のように`panic`することがあります。

```
panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x1 addr=0x0 pc=0x45d33c]
goroutine 1 [running]:
fmt.(*buffer).writeString(...)
```

`string`型の値は複数の部分からなっており、文字列の長さを表す部分とバイト列の先頭へのポインタを持っています。

https://github.com/golang/go/blob/97daa6e94296980b4aa2dac93a938a5edd95ce93/src/runtime/string.go#L232-L2351

長さを表す部分とそのポインタ部分が一緒に更新されれば問題ないのですが、reader側から中途半端に片方だけ更新された状態を観測してしまうと、nil pointer dereferenceが発生します。

## スライスの`len`と`cap`が中途半端に更新される

次のプログラムで、writerは常にlenとcapが等しいようなスライスをsに代入しています。sの初期値(`nil`)も`len(s) == cap(s) == 0`ですから、一見するとこのプログラムの全体にわたって`len(s) == cap(s)`になりそうです。

```go
func sliceCorruption() {
	underlying := [5]int{1, 2, 3, 4, 5}
	var s []int

	go func() { // writer
		for i := 0; ; i++ {
			// rは1から5までの整数
			r := i%5 + 1
			// len == capであるようなスライスを新たに作り、
			// sに代入する
			s = underlying[:r:r]
		}
	}()
	// reader
	for {
		// len(s) == cap(s)は常に成り立つと期待する？
		if len(s) != cap(s) {
			panic(fmt.Sprintf("len(s) == %d and cap(s) == %d", len(s), cap(s)))
		}
	}
}
```

次のPlaygroundでこの関数を実行してみます。

https://go.dev/play/p/CSEvhIpGqtv

```
panic: len(s) == 2 and cap(s) == 1

goroutine 1 [running]:
main.sliceCorruption()
	/tmp/sandbox2438933514/prog.go:27 +0x139
main.main()
	/tmp/sandbox2438933514/prog.go:6 +0xf
```

具体的な実行結果は毎回変わりますが、`len`が`cap`と異なるばかりか、`len`が`cap`よりも大きい状態（！）をreaderが観測しました。

**補足**

sliceの実装を見ると、3つのフィールドからなるstructになっています。sliceの要素を保存する配列(underlying arrayと言います)と`len`と`cap`です。

https://github.com/golang/go/blob/master/src/runtime/slice.go#L15-L19

これらが別々のメモリー位置にあることから、data raceが起きているときにはその一部のフィールドだけが更新された状態を観測する可能性があることがわかります。

## 型assertしたはずのinterfaceの動的値がおかしい

inteface型の例として、`any`型の変数の例をあげます。writer側では、異なる型の値を交互に代入してみましょう。reader側では型スイッチ文を使って動的型を確かめてから、動的値が期待通りかどうかチェックします。

```go
func interfaceCorruption() string {
	var x any

	go func() { // writer
		arr := []any{1, "hello"}
		for i := 0; ; i++ {
			x = arr[i%2]
		}
	}()
	// reader
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
			return fmt.Sprintf("strange type detected: %+v", read)
		}
	}
}
```

`int`型の`1`と`string`型の`"hello"`だけを交互に代入しているのですから、reader側で`int`と判定すれば値は`1`だし、`string`型と判定すれば長さは`5`になりそうなものですが、次のPlaygroundで実行するとそうならないケースがレポートされます。

https://go.dev/play/p/dT7SDd4becu

```
unexpected string length :-9223372036854775808
```

interface型の値には「型の情報(動的型など)」と「値の情報(動的値)」の2つの部分があります。この2つの部分を中途半端に更新した状態をreaderが観測することによって、このような結果が起こります。

**補足**

Goのruntimeにおけるinterface型の実装はおそらく次の箇所にあります。

https://github.com/golang/go/blob/master/src/runtime/runtime2.go#L205-L208

`data`の部分が動的値に対応して、`tab`の部分が型に関する情報になっています。

## mapのdata raceは`panic`するようになっている?

最後に`map`型を扱います。実は`map`型は少し特別で、race detectorを使うまでもなく、data raceが発生したらその時点で`panic`するようになっています。

例えば、次の関数を実行すると`panic`します。

```go
func mapCorruption() {
	// 共有変数
	m := map[int]int{}

	// writer
	go func() {
		for i := 0; ; i++ {
			m[i] = i
		}
	}()
	// reader
	for {
		if m[len(m)] > 10000 {
			break
		}
	}
}
```

https://go.dev/play/p/lLXLPicqXQJ

ただし、readerからの`map`へのアクセスの仕方を変えて、要素へのアクセス`m[key]`を行わずに、`m`の大きさである`len(m)`にのみアクセスした場合は、`panic`しませんでした。

```go
func mapCorruption2() {
	// 共有変数
	m := map[int]int{}

	// writer
	go func() {
		for i := 0; ; i++ {
			m[i] = i
		}
	}()
	// reader
	for {
		// len(m)にだけアクセスする
		// 要素にはアクセスしない
		if len(m) > 10000 {
			break
		}
	}
}
```

https://go.dev/play/p/TtBIoccdk2s

これもdata raceであることに変わりはなく、`-race`つきでローカル実行するとdata raceが報告されます。

`len(m)`と`m`の中身で矛盾があるようなサンプルコードを書こうと思ったのですが、`m`の中身にアクセスしようとすると`panic`してしまいますから、そのようなコードは書けませんでした。


# その他直感に反する結果

このセクションでは一貫性とは別な観点で直感に反する結果をもたらすdata raceサンプルコードを挙げます。

:::message
他にも書きたい例があるのですが、動作確認でき次第追加していきます。
:::

それぞれのサンプルにはよく使われる名前がついているので、その名前を見出しにしています。興味があれば調べてみてください。

## Store Buffering

```go
// メモリーモデル上はpanicする可能性があり実際panicすることがある
func storeBuffer() {
	var eg errgroup.Group
	// 共有変数
	x, y := 0, 0
	r1WasZero, r2WasZero := false, false
	eg.Go(func() error {
		x = 1
		r1 := y
		r1WasZero = r1 == 0
		return nil
	})
	eg.Go(func() error {
		y = 1
		r2 := x
		r2WasZero = r2 == 0
		return nil
	})
	eg.Wait() // エラー処理略
	if r1WasZero && r2WasZero {
		panic("Store Buffer Test Failed")
	}
}
```

素直に考えると、`r1 == 0`だったなら`y = 1`よりも先に`x = 1`の書き込みをしていると考えるので、`r2 := x`の時点で`x == 1`になっているはずだと思えます。しかし、次のPlaygroundでこの関数を繰り返し呼び出すと、`panicします。`

https://go.dev/play/p/_XpsxYfh8X5

```
panic: Store Buffer Test Failed

goroutine 1 [running]:
main.storeBuffer()
	/tmp/sandbox1591362111/main.go:29 +0x192
main.main()
	/tmp/sandbox1591362111/main.go:7 +0xf
```

# まとめと開発上の個人的な考え方

この記事では、data raceが存在するときには通常のプログラマーの自然な期待を裏切るような結果が起こりうることを見てきました。

最初に述べたように、これはあくまでdata raceが存在するときにのみ起こりうることです。並行処理を使っていても、data raceが起きないようにしていれば、この記事で挙げたような不思議な事象は起こりません。それでは、data raceが起きないようにするにはどうすれば良いでしょうか？

data raceが起きていないことについて自信を持つには、Race Detectorを使ったテストをするのが有効です。ただし、Race Detectorは静的解析ではなく、動的にdata raceを検知する技術です。つまり、実際にプログラムを動かして、実際に起きたメモリー読み書きがdata raceであればそれを報告します。ですから、テストがdata raceを引き起こすようなシナリオをカバーしていなければ、Race Detectorはそれを見逃してしまいます。

:::message
ここでいうテストとは`go test -race`で実行される自動テストに限らず、`go build -race`でビルドしたバイナリを使って行う、より大きなテストも含みます。
:::

:::message
なお、Race Detectorはdata raceを見落とすことはありますが、「data raceを報告したけれどもそれが本当はdata raceではない」という方向の誤りはしません。

短くいうと、false negativeはありますがfalse positiveはないとされています。

ですから、data raceが報告されたらそれが本当のdata raceであることを疑う必要はありません。
:::

個人的には、data raceを引き起こすかもしれないようなテストケース・テストシナリオをすべてカバーするというのは簡単ではないと思います。ですから、例えばチーム開発であれば、どのようなコードがdata raceになりうるかを理解したメンバーがレビューやモブプロに参加するといった地道な取り組みも重要だと思います。

ところで、data raceは絶対に避けるべきものなのでしょうか？Go言語に関する限り、絶対に避けるべきだとは言い切れないと思っています。Go言語のメモリーモデルにおいてdata raceは未定義動作ではなく、起こりうる結果は有限通りのパターンしかないとされているので、原理的にはすべての起こりうるパターンをプログラマーが確認できるからです。

しかし、個人的にはdata raceは極力見つけ次第解消したいと思っています。実践的には、data raceは無条件でバグとして取り扱う、くらいのスタンスが良いのではないでしょうか。というのも、この記事で挙げたような短い関数でも驚くような挙動があるので、現実的な大きさのソースコードにdata raceが紛れこんでいるとき、それが「無害なdata race」であることを確信するのは非常に難しいと思うからです。

以上をまとめると、個人的にはdata raceについて次のように考えています:

- data raceがあるプログラムはとても理解が難しくなるので、data raceは極力完全に無くした方が良い
- data raceをなくすには、Race Detectorを活用し、goroutineに慣れているメンバーを含むレビューやモブプロも行うのが良い

## 補足: 他言語におけるdata raceの取り扱い

ついさきほど、「並行処理を使っていても、data raceが起きないようにしていれば、この記事で挙げたような不思議な事象は起こりません。」と書きました。この性質をより専門的には、"DRF-SC"と呼んでいます。もちろん、DRF-SCにはもっと正確な定義がありますが、とりあえず「data raceさえなければ素直な動きをするという性質」くらいに捉えて構わないと思います。

多くの現代的プログラム言語(のメモリーモデル)がDRF-SCを満たしていて、例えばGo, C, C++, Rust, Java, JavaScript(ECMAScript)が当てはまります。

一方で、「data raceが起きた場合に何が起こりうるか」の部分は、DRF-SCを満たす言語の間でも違いがあります。

例えばC, C++などはdata raceが発生した場合の動きは未定義動作で、「何が起きてもおかしくない」と言えます。

一方、Go, Java, JavaScriptはそうではなく、data raceが発生した場合の動きは有限通りのパターンとして定義されています。非常に理解が難しいとはいえ、徹底分析すれば起こりうる可能性は列挙できるはずだと言えます。

:::message
このあたりの情報は、Russ Cox氏の https://research.swtch.com/plmm#fire の受け売りです。
:::

:::message
DRF-SCという用語についていくつか補足します。

- DRF-SCのことをSC-DRFとも言うようです。
- DRFの部分は"data-race-free"の略で、「data raceがないこと」の意味です。
- SCの部分は"Sequentially Consitent"の略で、「逐次一貫的」と訳されます。
- DRF-SCは、「DRFならばSCである」というように書き下して理解すると覚えやすいと思います。
:::

# 参考資料

筆者が参考にした資料と、参考になりそうな資料を挙げておきます。

| タイトルとリンク                                            | 概要                                                                                                    |
|------------------------------------------------------------|---------------------------------------------------------------------------------------------------------|
| [The Go Memory Model - The Go Programming Language](https://go.dev/ref/mem) | Goのメモリーモデルです。メモリーモデルとはメモリーへの並行アクセスをしたときに起きることを定めた言語仕様のことで、この記事で扱った内容の基礎となるドキュメントです。|
| [research!rsc: Memory Models](https://research.swtch.com/mm)                      | GoのメンバーであるRuss Cox氏による、Goに限らないメモリーモデル全般についての解説・論文です。2022年に行われたGoメモリーモデルのアップデートのために書かれたものなのですが、その意義を理解するために必要な前提知識から丁寧に説明しています。|
| [データ競合と happens-before 関係](https://uchan.hateblo.jp/entry/2020/06/19/185152) | uchan氏による、data race(データ競合)についての詳しい解説です。日本語です。|
| [よくわかるThe Go Memory Model](https://docs.google.com/presentation/d/e/2PACX-1vS2FIFiNgrpRpm1bPO3KpVzJYX4vEFhpyttvBTsTq15BwFmWvW0Q0W4udSf3pJMQyzZJicE5LcR5_cY/pub?start=false&loop=false&delayms=3000)                          | 筆者によるGoメモリーモデル解説です。                                                                    |
| [Data Race Detector - The Go Programming Language](https://go.dev/doc/articles/race_detector)   | Go公式によるData Race Detectorの解説です。                                                            |
| [Looking inside a Race Detector](https://www.infoq.com/presentations/go-race-detector/)                    | Race Detectorの仕組みであるVector Clockについての非常にわかりやすい解説です。                             |
| [Go Slices: usage and internals - The Go Programming Language](https://go.dev/blog/slices-intro) | Go公式によるスライスの使い方と内部の解説です。                                                          |
| [The Laws of Reflection - The Go Programming Language](https://go.dev/blog/laws-of-reflection) | Go公式によるreflectionの解説なのですが、interface型についての解説も含んでいます。                      |
| [research!rsc: Go Data Structures: Interfaces](https://research.swtch.com/interfaces)       | GoのメンバーであるRuss Cox氏によるinterface型についての解説です。                                        |

# 最後に

サンプルコードの動作確認をしつつ正確を期して記述しましたが、data raceというテーマ自体がかなり難しいものなので、記述に誤りがないとは言い切れません。何か気づいたことがありましたらGitHubリポジトリのIssueやPull Requestなどでご連絡いただけると助かります。

また、執筆にあたり次の方から情報やフィードバックをいただきました。ありがとうございます。

- DQNEOさん

もちろん、記述の誤りなどについてのすべての責任は筆者にあります。