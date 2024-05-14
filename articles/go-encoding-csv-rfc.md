---
title: "Goのencoding/csvのオプションとRFC 4180
emoji: "😽"
type: "tech" # tech: 技術記事 / idea: アイデア
topics: [Go, CSV, TSV, RFC]
published: false
---

# はじめに

この記事は、Goのencoding/csvのやや詳しい入門であり、CSVの比較的よく知られた仕様である[RFC 4180](https://www.rfc-editor.org/rfc/rfc4180.html)の入門にもなっています。

Goのencoding/csvのデフォルト動作を[RFC 4180](https://www.rfc-editor.org/rfc/rfc4180.html)の記述と対応させながら、オプションによってRFCの要求仕様をどのように緩和・変更できるかをみていきます。

特に、二重引用符`"`を含むフィールドの扱いについては開発上出会うことがあるので、何が起きているのかを把握する役に立つかもしれません。

## 基本資料

- https://www.rfc-editor.org/rfc/rfc4180.html
- https://pkg.go.dev/encoding/csv

# encoding/csvの基本デザイン

Go1.22時点のパッケージドキュメントには次のようにあります。

> Package csv reads and writes comma-separated values (CSV) files. There are many kinds of CSV files; this package supports the format described in [[RFC 4180](https://www.rfc-editor.org/rfc/rfc4180.html)](https://www.rfc-editor.org/rfc/rfc4180.html).

> csvパッケージはCSVファイルの読み書きを行います。CSVファイルにはたくさんの種類がありますが、このパッケージは[[RFC 4180](https://www.rfc-editor.org/rfc/rfc4180.html)](https://www.rfc-editor.org/rfc/rfc4180.html)で説明されるフォーマットをサポートします。

https://pkg.go.dev/encoding/csv

これを実現するように、`encoding/csv`は次のような基本デザインになっています。

- CSVのよく知られた仕様である[[RFC 4180](https://www.rfc-editor.org/rfc/rfc4180.html)](https://www.rfc-editor.org/rfc/rfc4180.html)をサポートする。
  - (概ね)デフォルトでは厳格にRFCに準拠し、オプションで緩和できるようになっている
- 区切り文字をデフォルトの`,`から変えることでTSVなども扱える。ただしその場合もCSVの仕様である[[RFC 4180](https://www.rfc-editor.org/rfc/rfc4180.html)](https://www.rfc-editor.org/rfc/rfc4180.html)に準拠した動作になる。
- APIとしては読み取り用の`csv.Reader`と書き込み用の`csv.Writer`を提供し、これらは`io.Reader, io.Writer`を渡して作る。
  - この点は他のGo標準パッケージたちと同様で、例えば他にはencoding/json, compress/gzipなども同様のデザインになっている

# encoding/csvの典型的な使用例

まず、おさらいとしてシンプルな使用例をみておきましょう。慣れている人はこのセクションを飛ばして構いません。

ここではpackageドキュメントの例をほとんどそのまま引用しつつ、コメントで解説を付け加えておきます。

## Readerを使用してCSVを読み取る

https://pkg.go.dev/encoding/csv#example-Reader

```go
func main() {
	in := `first_name,last_name,username
"Rob","Pike",rob
Ken,Thompson,ken
"Robert","Griesemer","gri"
`
    // csv.NewReaderはio.Readerを受け取るので、
    // strings.NewReaderでstringからio.Readerにしている
	r := csv.NewReader(strings.NewReader(in)) 

	for {
        // 1行を[]stringで返す
		record, err := r.Read()
		if err == io.EOF {
            // 終了した時はio.EOFを返す
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(record)
	}
}
```

## Writerを使用してCSVを書き込む


https://pkg.go.dev/encoding/csv#example-Writer

```go
func main() {
	records := [][]string{
		{"first_name", "last_name", "username"},
		{"Rob", "Pike", "rob"},
		{"Ken", "Thompson", "ken"},
		{"Robert", "Griesemer", "gri"},
	}

    // CSVを書き込みたい先をio.Writerとして渡す
    // ここでは標準出力を渡している
	w := csv.NewWriter(os.Stdout)

	for _, record := range records {
		if err := w.Write(record); err != nil {
			log.Fatalln("error writing record to csv:", err)
		}
	}

	// Write any buffered data to the underlying writer (standard output).
    // Flushを呼ばないと一部のデータが書き込まれないままになってしまうので注意
	w.Flush()

	if err := w.Error(); err != nil {
		log.Fatal(err)
	}
}
```

# [[RFC 4180](https://www.rfc-editor.org/rfc/rfc4180.html)](https://www.rfc-editor.org/rfc/rfc4180.html) の仕様と`encoding/csv`の関係

はじめに見た通り、`encoding/csv`は[[RFC 4180](https://www.rfc-editor.org/rfc/rfc4180.html)](https://www.rfc-editor.org/rfc/rfc4180.html)をサポートします。
`encoding/csv`はCSVの読み取りと書き込みを行うので、「サポートする」というのは具体的には次のようなことを指しています。

- CSVを読み取るとき、読み取るファイルが[[RFC 4180](https://www.rfc-editor.org/rfc/rfc4180.html)](https://www.rfc-editor.org/rfc/rfc4180.html)のフォーマットに従っていることを期待し、そうでない場合はエラーとする。
    - ただし、いくつかのRFCの仕様は`csv.Reader`のオプションによって緩めることもできる。
- CSVを書き込むとき、[[RFC 4180](https://www.rfc-editor.org/rfc/rfc4180.html)](https://www.rfc-editor.org/rfc/rfc4180.html)のフォーマットで書き込む。
    - ただし、デフォルトの改行文字は`\n`であって`\r\n`ではない。([[RFC 4180](https://www.rfc-editor.org/rfc/rfc4180.html)](https://www.rfc-editor.org/rfc/rfc4180.html)ではCRLF、つまり`\r\n`となっている)   
    - フォーマットの一部を`csv.Writer`のオプションで変更できる。
      - 変更できるのは、区切り文字(デフォルトは`,`)と改行文字(デフォルトは`\n`)の2つ

そこで、`csv.Reader`と`csv.Writer`のそれぞれについて[[RFC 4180](https://www.rfc-editor.org/rfc/rfc4180.html)](https://www.rfc-editor.org/rfc/rfc4180.html)との関係を見ていきます。


# [[RFC 4180](https://www.rfc-editor.org/rfc/rfc4180.html)](https://www.rfc-editor.org/rfc/rfc4180.html) の仕様と`csv.Reader`の関係

すでに述べたように、`csv.Reader`は読み取るファイルが[[RFC 4180](https://www.rfc-editor.org/rfc/rfc4180.html)](https://www.rfc-editor.org/rfc/rfc4180.html)のフォーマットに従っていることを期待し、そうでない場合はエラーとします。

そこで、RFCの仕様のそれぞれについて、

- 破っている時にどうエラーになるか
- どのオプションによってそのエラーを抑制できるか

を見ていくことにします。

[[RFC 4180](https://www.rfc-editor.org/rfc/rfc4180.html)](https://www.rfc-editor.org/rfc/rfc4180.html)の中で、CSVのフォーマットについて記載しているのはSection 2のみで、仕様は7項目からなっています。

::: message 
仕様について日本語への拙訳を記載しますが、直訳ではないこともあることに気をつけてください。
:::


https://www.rfc-editor.org/rfc/rfc4180.html#section-2

## 仕様1: Each record is located on a separate line, delimited by a line break (CRLF).  

まず1.です。

>   1.  Each record is located on a separate line, delimited by a line break (CRLF).  

> For example:

       aaa,bbb,ccc CRLF
       zzz,yyy,xxx CRLF

> それぞれのレコードは改行(CRLF)によって分かれた行にあります。

これは、どちらかというと、改行で分けられた1つの行が「レコード」であると定義しているような文で、これに違反する入力というのは特に想定されません。

## 仕様2: The last record in the file may or may not have an ending line break.  

次に2.です。

>  2.  The last record in the file may or may not have an ending line break.  

> 最後のレコードは末尾の改行を持っても持たなくても良い

ファイル最後の改行はあってもなくても有効なCSVだと述べています。

実際、先に見た例では最後の改行がありましたが、次のように最後の改行をなくしても読み取りできています。

https://go.dev/play/p/q57da13RzUE

```go
func main() {
	in := `first_name,last_name,username
"Rob","Pike",rob
Ken,Thompson,ken
"Robert","Griesemer","gri"`
	r := csv.NewReader(strings.NewReader(in))

	for {
		// 1行を[]stringで返す
		record, err := r.Read()
		if err == io.EOF {
			// 終了した時はio.EOFを返す
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(record)
	}
}
```

::: message
最後の改行があってもなくても良いということは、`csv.Reader`から見ると「あってもなくても読み取れなければいけない」ということです。

このように、データの立場とデータを処理するプログラムの立場では、一方から「〜しても良い」となっている仕様を他方から見ると「〜できなければいけない」となるように、一方にとって寛容なルールが他方から見ると厳格なルールになることがよくあります。

これは経験のあるプログラマーには暗黙知のようになっていることですが、あえて書く価値のある面白いことだと個人的には思います。

ソフトウェアの介在しない書類事務でも同じような現象があるかもしれません。
:::

## 仕様3: ヘッダー行について

次に3.をみていきます。

> 3.  There maybe an optional header line appearing as the first line
> of the file with the same format as normal record lines.  This
> header will contain names corresponding to the fields in the file
> and should contain the same number of fields as the records in
> the rest of the file (the presence or absence of the header line
> should be indicated via the optional "header" parameter of this
> MIME type). 

> ファイルの先頭にはオプショナルなヘッダー行があっても良い。これは通常のレコード行と同じフォーマットを持つ。
> このヘッダーはファイルのフィールドに対応する名前を持つだろう。
> また、ファイルの残りのレコードと同じ数のフィールドを持つべきである。
> (ヘッダー行があるかないかは、 このMIME typeのオプショナルなheaderパラメータで示されるべきである)

ちょっと長いですがヘッダー行があっても良いということを述べています。

そして「ヘッダー行はファイルの残りのレコードと同じ数のフィールドを持つべき(should)である」と述べているので、ヘッダー行のフィールド数と残りのレコードのフィールド数が異なる場合はこの部分に違反することになります。

違反する場合のコード例を出したいのですが、次の使用でまとめて具体例を出したいので、一旦4.に進みます。

## 仕様4: フィールドの数などについて

>   4.  Within the header and each record, there may be one or more
>       fields, separated by commas.  Each line should contain the same
>       number of fields throughout the file.  Spaces are considered part
>       of a field and should not be ignored.  The last field in the
>       record must not be followed by a comma.

> ヘッダーおよび各レコード内には、カンマで区切られた1つ以上のフィールドが存在することがある。
> ファイル全体で各行は同じ数のフィールドを含むべきである。
> スペースはフィールドの一部と見なし、無視されるべきではない。
> レコードの最後のフィールドの後にはカンマを付けてはならない。

### 仕様: Each line should contain the same number of fields throughout the file. 

この「ファイル全体で各行は同じ数のフィールドを含むべきである。」に違反する入力をデフォルト設定の`Reader`に与えると、エラーを返します。

https://go.dev/play/p/k0e-16GimEs

```
first_name,last_name,username
"Rob","Pike",rob
Ken,Thompson
"Robert","Griesemer","gri"`
```

3行目だけフィールドが2つしかありません。

> record on line 3: wrong number of fields

この振る舞いは緩めることもできます。https://pkg.go.dev/encoding/csv#Reader の`FieldsPerRecords int`フィールドを設定して、`-1`などの負の数を与えると、RFCの前記仕様に違反した入力も読めるようになります。

https://go.dev/play/p/rMrbuLDlgyR

```go

func main() {
	in := `first_name,last_name,username
"Rob","Pike",rob
Ken,Thompson
"Robert","Griesemer","gri"`
	r := csv.NewReader(strings.NewReader(in))
	r.FieldsPerRecord = -1 // 条件を緩和

	for {
		// 1行を[]stringで返す
		record, err := r.Read()
		if err == io.EOF {
			// 終了した時はio.EOFを返す
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(record)
	}
}
```

逆に`12`などの正の整数を設定すると、フィールド数が揃っているだけでなく、`12`で揃っていない場合にエラーになります。

デフォルトは`int`のゼロ値である`0`になっていて、その場合はフィールド数はなんでも良いが全てのレコードで等しくなっていることを仮定します。これはRFC通りの仕様です。

### 仕様: Spaces are considered part of a field and should not be ignored. 

この仕様は、フィールドにあるスペース` `はフィールドの一部として解釈されることを示しています。

Goの`csv.Reader`もこれに従った動作をしますが、この動作は`csv.Reader.TrimLeadingSpace`を`true`に設定することで変更できます。設定すると、フィールド先頭のスペースはフィールドの一部とはみなされなくなります。

### 仕様: The last field in the record must not be followed by a comma.

また、「レコードの最後のフィールドの後にはカンマを付けてはならない。」を破ってみましょう。

https://go.dev/play/p/MzEM7mle7wX

```go
func main() {
	in := `first_name,last_name,username
"Rob","Pike",
Ken,Thompson,ken
"Robert","Griesemer","gri"`
	r := csv.NewReader(strings.NewReader(in))

	for {
		// 1行を[]stringで返す
		record, err := r.Read()
		if err == io.EOF {
			// 終了した時はio.EOFを返す
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(record, len(record))
	}
}
```

> [first_name last_name username] 3
> [Rob Pike ] 3
> [Ken Thompson ken] 3
> [Robert Griesemer gri] 3

予想に反して（？）エラーにはなりませんでした。これはカンマの後の空文字列が1つのフィールドとみなされているためです。

2行目に空ではないフィールドを3つ設定して、その最後にカンマをつけた場合は、今度は2行目のフィールド数だけが異なることによりエラーになります。

つまり、「最後のフィールドの後にカンマがついている」という理由のエラーは返らないのですが、その後の空文字列がフィールドとみなされてしまうことにより、間接的に「最後のフィールドにカンマをつけられない」というRFCの仕様が満たされています。


## 仕様5: 二重引用符`"`について

5. をみていきます。

> 5.  Each field may or may not be enclosed in double quotes (however
>       some programs, such as Microsoft Excel, do not use double quotes
>       at all).  If fields are not enclosed with double quotes, then
>       double quotes may not appear inside the fields.

> 各フィールドは二重引用符で囲まれている場合もあれば、囲まれていない場合もある（ただし、Microsoft Excelのような一部のプログラムは、まったく二重引用符を使用しない）。
> フィールドが二重引用符で囲まれていない場合、フィールド内に二重引用符を含めてはならない。

このあたりが開発で遭遇しやすい問題かもしれません。

先ほどから例に出している有効な入力は、確かに二重引用符`"`が使われていたり使われていなかったりします。

```go
first_name,last_name,username
"Rob","Pike",rob
Ken,Thompson,ken
"Robert","Griesemer","gri"
```

レコードごとに違うばかりか、同一レコードでもフィールドごとに違ったりします。これもRFCには準拠したフォーマットであり、実際に`csv.Reader`はこれを読み取れます。

### 仕様:  If fields are not enclosed with double quotes, then double quotes may not appear inside the fields.

これに違反するのは次のようなレコードを含む場合です。実際に、`csv.Reader`はこれをエラーにします。

```go
"Rob","Pike",r"ob
```

>  parse error on line 2, column 15: bare " in non-quoted-field

この条件は`csv.Reader.LazyQuotes`フィールドを`true`に設定すると緩和できます。

```go
// If LazyQuotes is true, a quote may appear in an unquoted field and a
// non-doubled quote may appear in a quoted field.
LazyQuotes bool
```

https://go.dev/play/p/cwQaKwyR8i0

```go
func main() {
	in := `first_name,last_name,username
"Rob","Pike",r"ob
Ken,Thompson,ken
"Robert","Griesemer","gri"`
	r := csv.NewReader(strings.NewReader(in))
	r.LazyQuotes = true

	for {
		// 1行を[]stringで返す
		record, err := r.Read()
		if err == io.EOF {
			// 終了した時はio.EOFを返す
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(record, len(record))
	}
}
```

> [first_name last_name username] 3
> [Rob Pike r"ob] 3
> [Ken Thompson ken] 3
> [Robert Griesemer gri] 3

## 仕様6: Fields containing line breaks (CRLF), double quotes, and commas should be enclosed in double-quotes. 

> 改行（CRLF）、二重引用符、カンマを含むフィールドは、二重引用符で囲まなければならない。

逆にいうと、次のようにすればフィールドに改行やカンマを使えます。

https://go.dev/play/p/J2Gvi90-3y7

```
"aaa","b
bb","ccc"
zzz,yyy,xxx
```

## 仕様7: If double-quotes are used to enclose fields, then a double-quote appearing inside a field must be escaped by preceding it with another double quote.

> フィールドを囲むために二重引用符が使用される場合、フィールド内に現れる二重引用符は、もう一つの二重引用符を前に置いてエスケープしなければならない。

```
"aaa","b""bb","ccc"
```

は、`csv.Reader`で読み取ることができて、次のように読み取られます。

```
[aaa b"bb ccc]
```

これに違反した入力は次のようなもので、デフォルトの`Reader`ではエラーになります。

```
"aaa","b"bb","ccc"
```

> extraneous or missing " in quoted-field
> exit status 1

https://go.dev/play/p/RsuP4m66hd6

これも、`csv.Reader.LazyQuotes`フィールドを`true`に設定すると緩和できます。

https://go.dev/play/p/2hBKDPtcxfL

## `LazyQuotes = true`でエラーを抑制する際の注意点

次のCSVファイルはRFC 4180違反で、デフォルトの`csv.Reader`ではエラーになります。

```
a,b,c
1,2,3
4,"5"five,6
7,"8",9
```

3行目の`"5"five`が次の仕様に違反しているからです。

> Fields containing line breaks (CRLF), double quotes, and commas should be enclosed in double-quotes.

これは次のように`LazyQuotes = true`で抑制できます。しかし結果の行数を見ると`3`になっており、期待される`4`と異なります。

https://go.dev/play/p/FoSgK1T-Z6G

実は、このとき3行目の2列目は意図通りに（？）解釈されておらず、次のような1つのフィールドとして読み取られています。

```
5"five,6
7,"8
```

どういうことでしょうか？このフィールドは`"`から始まっているので、`"`によってencloseされたフィールドとして扱われます。よってその終わりは`"`であり、その後には`,`が来るはずです。
よって、最後の行の`"8",`に含まれる最後の`",`に到達して初めて1つのフィールドが終わったものと解釈されているのです。

これだけだと幾つか疑問が残ると思います。

*疑問1: 改行はどうなっているのか？*

`4,"5"five,6`のあとの改行はどうなっているのか？というと、これはRFCにより`"`でencloseされたフィールドでは改行文字をフィールドの一部として使えるので、フィールドの一部として扱われています。

> Fields containing line breaks (CRLF), double quotes, and commas should be enclosed in double-quotes. 

*疑問2: 途中の`"`はどうなっているのか?*

`4,"5"five,6`の`"five`の部分の`"`は、RFCのフォーマットには違反しています。しかし今は`LazyQuotes`でフォーマット制限を緩めているので、エラーになりません。しかも、この`"`の直後には`,`がないのでフィールドの終端とは判断されず、フィールドの一部と判断されます。

*疑問3: 途中の`,`はどうなっているのか?*

`4,"5"five,6`の`five,`の部分のカンマ`,`は、フィールドの一部として扱われています。

> Fields containing line breaks (CRLF), double quotes, and commas should be enclosed in double-quotes. 

今のフィールドは`"`でenclosedなのでRFCの仕様としてもこの`,`がフィールドの一部になるのは正しいです。このカンマが区切り文字であるならば直前に`"`があるはずなので、`",`がセットで出現したときにはじめてこのフィールドが終わったと解釈されます。

このように、`LazyQuotes`は便利なようですがある種の入力に対しては意図しない結果を生む原因にもなるので、`LazyQuotes`を使えば読み込めるからといって使って良いかどうかは個別に判断が必要だと思います。

:::message
今扱った入力だと、意図しない（？）結果になりかつカラム数がおかしくならないためエラーにもなってくれないという結果になります。

似たようなケースで`LazyQuotes`を`true`にしていると、多くの場合は結果のカラム数がおかしくなって"wrong number of fields"のエラーになると思います。しかしたまたまカラム数が「正しくなってしまう」可能性もあるので、エラーになることを期待すべきではないと思います。
:::


# [[RFC 4180](https://www.rfc-editor.org/rfc/rfc4180.html)](https://www.rfc-editor.org/rfc/rfc4180.html) の仕様と`csv.Writer`の関係

`csv.Writer`はオプションフィールドが2つしかなく、`Reader`に比べると単純です。

- CSVを書き込むとき、[[RFC 4180](https://www.rfc-editor.org/rfc/rfc4180.html)](https://www.rfc-editor.org/rfc/rfc4180.html)のフォーマットで書き込む。
    - ただし、デフォルトの改行文字は`\n`であって`\r\n`ではない。([[RFC 4180](https://www.rfc-editor.org/rfc/rfc4180.html)](https://www.rfc-editor.org/rfc/rfc4180.html)ではCRLF、つまり`\r\n`となっている)   
    - フォーマットの一部を`csv.Writer`のオプションで変更できる。
      - 変更できるのは、区切り文字(デフォルトは`,`)と改行文字(デフォルトは`\n`)の2つ

具体的には、`Writer.Comma`で区切り文字をデフォルトから変更できて、`Writer.UseCRLF`で改行文字を`\n`にするか`\r\n`にするか選択できます。

この`Writer.UseCRLF`は、`encoding/csv`のデフォルト動作が[[RFC 4180](https://www.rfc-editor.org/rfc/rfc4180.html)](https://www.rfc-editor.org/rfc/rfc4180.html)と異なっている唯一のポイントです。

# TSVを扱うときの注意点

`encoding/csv`は区切り文字を変更することでTSVのライブラリとしても使用できます。

ただし、その場合も準拠する仕様はあくまでCSVの仕様である[[RFC 4180](https://www.rfc-editor.org/rfc/rfc4180.html)](https://www.rfc-editor.org/rfc/rfc4180.html)です。

::: message
TSVについては（筆者の知る限り）CSVの[[RFC 4180](https://www.rfc-editor.org/rfc/rfc4180.html)](https://www.rfc-editor.org/rfc/rfc4180.html)に相当するような広く知られた仕様がありません。
:::

# この記事へのフィードバックについて

この記事についてフィードバックやご意見がある場合、[GitHubリポジトリ](https://github.com/nobishino/zenn)にissueを立てるか、PRを直接立てていただけると助かります。