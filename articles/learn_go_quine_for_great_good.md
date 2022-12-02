---
title: "QuineであそんでまなぶGo言語"
emoji: "👻"
type: "idea" # tech: 技術記事 / idea: アイデア
topics: [Go]
published: false
---

# 対象読者

この記事は次のような人を対象としています。

- Go言語の基本的な文法を知っている
- パズルとしてプログラミングを楽しむことに興味がある

# Quine(クワイン) - 自分自身を出力するプログラム

Quine(クワイン)というものをご存知でしょうか？Quineとは、文字列を出力するプログラムであって、その出力される文字列がそのプログラム自身を表す文字列と同一であるもののことを言います。

https://ja.wikipedia.org/wiki/%E3%82%AF%E3%83%AF%E3%82%A4%E3%83%B3_(%E3%83%97%E3%83%AD%E3%82%B0%E3%83%A9%E3%83%9F%E3%83%B3%E3%82%B0)

ある日、「Go言語のクワインであって、できるだけコードの長さが短いものを募集してみよう」と思い立ちました。そこで[GitHubリポジトリ](https://github.com/nobishino/goquine)を用意し、クワインの判定器をGitHub Actionに書き込んでQuineを募集したところ、何名かの方から投稿をいただけました。

これは完全に遊びだったのですが、思いのほか「学び」もありました。そこでこの記事では、投稿されたQuineを紹介しながらそうした「学び」をシェアしたいと思います。

# ルール

一応次のようなルールで募集していました。

- `main.go`で完結していること
- `go run main.go`を実行したとき、自分自身と`main.go`同一の文字列を標準出力すること
- `gofmt`で整形済みであること(`gofmt`にかけても変化しないこと)
- `os.Open`などで外部からデータを入力しないこと

このルールが一番面白いルールだと思ったわけではなく、単にCIを作るときにそれに合わせて定義しておいただけです。実際、このルールは満たさないが面白いQuineも投稿していただいたので、この記事ではルールを満たすものもそうでないものも紹介していきます。

# 筆者(nobishii)が最初に書いたクワイン(200 Bytes)

まず、筆者が最初に書いたQuineです。大きさは200 Bytesです。

```go
package main

import "fmt"

func main() {
	fmt.Println(q + fmt.Sprintf("%q", q))
}

var q = "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(q + fmt.Sprintf(\"%q\", q))\n}\n\nvar q = "
```

https://go.dev/play/p/ruPgcxLx1kr

これを書くにあたり、WikipediaにあったHaskellのQuineを参考にしました。

```haskell
main = putStrLn $ q ++ show q where q = "main = putStrLn $ q ++ show q where q = "
```

自分が知っていた別な言語のQuineでは、引用符`"`を`""`に置換するなどの複雑なテクニックが必要とされていましたが、このHaskellのコードはとても簡潔ですね。この簡潔さの秘密は文字列である`q`をプログラムの末尾で定義できていることにあると思いました。そこで、Go言語のパッケージ変数をつかって、文字列をプログラムの最後に置くことにしました。

パッケージ変数はそれが宣言されたり使われている行番号に関係なく、依存関係を勝手に解決して初期化してくれるので、このように書くことが出来ました。

https://go.dev/ref/spec#Package_initialization

# `.`によるimport(151 Bytes)  by @tenntenn 

次は@tenntennさんに投稿いただいたQuineです。

```
package main

import . "fmt"

var s = "package main\n\nimport . \"fmt\"\n\nvar s = %q\n\nfunc main() { Printf(s, s) }\n"

func main() { Printf(s, s) }
```

https://go.dev/play/p/Ht7aRBmaLWt

`.`によるimportによりパッケージ名を指定せずに`fmt.Printf`を呼べるようになりました。
その他にも、フォーマット文字列に自分自身を渡す`Printf(s,s)`形になっていてスマートです。

サイズは*151 Bytes*とかなり短いですね。

# asciiコードの技巧(137 Bytes) by @cia-rana

次に紹介するのは @cia-ranaさんによる異なるアプローチのQuineです。

```go
package main

func main() { a += "\x60"; println(a + a) }

var a = `package main

func main() { a += "\x60"; println(a + a) }

var a = `
```

https://go.dev/play/p/UGAQqcNJk5d

何をしているかわかるでしょうか？`"\x60"`はasciiコードの60番、つまりバッククオーテーションを表します。`println(a + a)`がこの`main.go`と一致するということは、この*`main.go`は前半と後半で全く同じ文字列を2回繰り返す文字列になっている*わけです。

最後の行で定義した`a`自体にはバッククオーテーションが含まれていないことに気をつけます。そこに`a += "\x60"`でバッククオーテーションを1つだけ後ろにくっつけます。それを2回出力すると、ひとつめの`a`の末尾にあるバッククオーテーションは5行目のバッククオーテーションとなり、ふたつめの`a`の末尾にあるバッククオーテーションは最後のバッククオーテーションになってくれるというわけです。おもしろいですね！

`println`は標準エラー出力なのでルールを満たしていませんが、それでも非常に面白いし、長さは**137 Bytes**とかなり短くなりました。標準パッケージを使っていないのもかっこいいですね。

# 脇道: 標準パッケージを使わずに標準出力するには

さて、当初「標準パッケージを使わずにルール通りのQuineを書ける」と思い込んでいたのですが、よく考えてみると標準パッケージを使わないとシステムコールができないので、`println`のような組み込み関数が別にない限り標準出力自体が一切できないことに気づきました。

ルールの設定をミスったなあと反省していたところ、DQNEOさんから次の情報を頂きました。

https://twitter.com/DQNEO/status/1594939354874798082

`.s`ファイルにアセンブリを書いてビルドすれば、標準パッケージなしでも標準出力ができるということですね。`main.s`もありというルールにしておけばよかったかもしれません。

# 禁じ手 - `embed`パッケージ() by tenntenn

最後に禁じ手っぽいものを紹介します（？）。

```go
package main

import (
	_ "embed"
	. "fmt"
)

//go:embed main.go
var src string

func main() { Print(src) }
```

https://go.dev/play/p/A2iC2R_VSxA

`main.go`をembedして`src`に格納しているので、それを出力すればQuineになります。
反則のように見えますが、embedは外部入力ではなく、あくまでビルド時にバイナリに含める機能ですから反則ではありません。

長さは *108 Bytes* です！


# おわりに

遊びではじめたリポジトリでしたが、投稿いただいたコードから学びがありました。
もっと短いQuineが書けるぞという方は、次のリポジトリにPRを出してみてください。

https://github.com/nobishino/goquine