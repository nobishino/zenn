---
title: "go:wasmimportを使ってHello Worldする"
emoji: "😽"
type: "tech" # tech: 技術記事 / idea: アイデア
topics: [Go]
published: true
---

# go:wasmimportとは

Go1.21ではWASIの実験的サポートが追加されましたが、それと同時に追加されたコンパイラディレクティブです。該当のリリースノートは次の箇所にあります。

https://golang.org/doc/go1.21#wasm

ドキュメントは https://pkg.go.dev/cmd/compile にあり、次のように書かれています。

> The //go:wasmimport directive is wasm-only and must be followed by a function declaration. It specifies that the function is provided by a wasm module identified by “importmodule“ and “importname“.

例えば次のように、"importmodule"と"importname"を指定して使います。

<!-- zenncode: goos=wasip1 goarch=wasm -->
```go
//go:wasmimport a_module f
func g()
```

このとき、関数`g`の実装は書く必要がなく、実装はa_moduleという[Wasmモジュール](https://webassembly.github.io/spec/core/syntax/modules.html)の`f`という関数によって提供されます。

この説明だと抽象的でよくわからないと思いますが、例えば次のようなことができます。

- JavaScriptで書いた関数をwasm向けのGoコードからgo:wasmimportで使う
- `wasmtime`などのWasmランタイムが提供している関数をGoの関数として使う

この記事の目標は、後者を行うコードを自分で動かしてみることで`go:wasmimport`への最初の一歩を踏み出すことです。

:::message

前者の例については次の記事が参考になります。
https://zenn.dev/askua/articles/9b614c377cc1e0#go%3Awasmimport

:::

# 行うこと

DQNEOさんによるライブコーディング「自力でシステムコールを叩いてhello worldを出力しよう」を同じことをwasmimportを使って行います。

https://docs.google.com/presentation/d/10ru3LdbofJqgdmD8pprZuZyWbGvOFC8rKxb6q5Q46Xc/edit#slide=id.gcf4887a11e_1_317

このライブコーディングは、`fmt.Println`を使って文字列を標準出力するときに内部で行われていることを掘ってゆき、より低レベルのAPIを使うようにリファクタリングしていくというものです。これと同じことを、wasmimportについて行ってみます。

ただし、wasmランタイムを使ってGoプログラムを実行する場合には、システムコールを叩く代わりにwasmランタイムの関数を呼び出す点が異なります。表で比較すると次のようになります。

| | Goアセンブリからシステムコールを叩く場合 | wasmimportした関数を叩く場合 | 
| ---- | ---- | ---- |
| 最終的に呼びたいもの | システムコール(`write(2)`)など | Wasmランタイムが提供するfs_write関数 |
| どうやるか | システムコールを叩くGoアセンブリを書く | go:wasmimportを使ってWasmランタイムの関数をGoから呼び出す | 

# `fmt.Println`を`syscall.Write`にリファクタリングする

実際にやってみましょう。出発点は次のコードです。

<!-- zenncode: expect=run -->
```go
func main() {
    fmt.Println("Hello, Wasm")
}
```

https://go.dev/play/p/fC8spIZnTY1

まず、オリジナルのライブコーディング同様に、これを`syscall.Write`を使うようにリファクタリングします。

<!-- zenncode: expect=run -->
```go
func main() {
	syscall.Write(1, []byte("Hello world\n"))
}
```

https://go.dev/play/p/Vnr97M6-m83

:::message
この引数で使っている`1`はファイルディスクリプタと呼ばれるもので、`1`だと標準出力(`os.Stdout`)の意味になります。
:::

`syscall`パッケージとは、ドキュメントに

> Package syscall contains an interface to the low-level operating system primitives.

> syscallパッケージは低レベルのOSプリミティブへのインタフェースを持っています。

とあるように、下にあるシステムの違いを吸収する役割を持っているパッケージです。その役割を果たすために、Goプログラムをビルドするターゲットのシステムごとに異なる実装を持っています。

例えばunix向けの実装は https://cs.opensource.google/go/go/+/refs/tags/go1.23.2:src/syscall/syscall_unix.go;l=201 にあります。

今回のハンズオンで使いたいのはwasip1向けの https://github.com/golang/go/blob/go1.23.2/src/syscall/fs_wasip1.go#L890-L895 です。

# `syscall.Write`を、そのwasip1向け実装に置き換える

`syscall.Write`に任せていた部分を自前のコードに置き換えます。そのために、https://github.com/golang/go/blob/go1.23.2/src/syscall/fs_wasip1.go#L890-L895 の実装を、依存する型や関数含めて丸ごと`main.go`にコピー＆ペーストします。

少しコード量が多くなりますが、`syscall`パッケージへの依存が消えたことがわかると思います。

<!-- zenncode: goos=wasip1 goarch=wasm -->
```go
//go:build wasip1 && wasm

package main

import (
	"fmt"
	"runtime"
	"unsafe"
)

// GOOS=wasip1 GOARCH=wasm go build -o main
// WASMTIME_LOG=wasmtime_wasi=trace wasmtime main
func main() {
	var buf = []byte("Hello, Wasm\n")
	_, err := write(1, buf)
	if err != nil {
		panic(err)
	}
}

// syscall.Writeのwasip1実装を(ほとんど)コピーしてきたもの
// 参考元は https://github.com/golang/go/blob/master/src/syscall/fs_wasip1.go#L910-L915
func write(fd int, b []byte) (int, error) {
	var nwritten size
	errno := fd_write(int32(fd), makeIOVec(b), 1, unsafe.Pointer(&nwritten))
	runtime.KeepAlive(b)
	return int(nwritten), errnoErr(errno)
}

type size = uint32
type Errno uint32

func (e Errno) Error() string {
	return fmt.Sprintf("errno %d", e)
}

type uintptr32 = uint32

func makeIOVec(b []byte) unsafe.Pointer {
	return unsafe.Pointer(&iovec{
		buf:    uintptr32(uintptr(bytesPointer(b))),
		bufLen: size(len(b)),
	})
}
func bytesPointer(b []byte) unsafe.Pointer {
	return unsafe.Pointer(unsafe.SliceData(b))
}

type iovec struct {
	buf    uintptr32
	bufLen size
}

// 本質的でないので簡略化した
func errnoErr(e Errno) error {
	switch e {
	case 0:
		return nil
	}
	return e
}

//go:wasmimport wasi_snapshot_preview1 fd_write
//go:noescape
func fd_write(fd int32, iovs unsafe.Pointer, iovsLen size, nwritten unsafe.Pointer) Errno
```


コード量は多いですが、重要なのは、結局次の関数が呼び出されているということです:

<!-- zenncode: goos=wasip1 goarch=wasm -->
```go
//go:wasmimport wasi_snapshot_preview1 fd_write
//go:noescape
func fd_write(fd int32, iovs unsafe.Pointer, iovsLen size, nwritten unsafe.Pointer) Errno
```

例えばLinux上で`fmt.Println`を実行したときには、最終的に`write(2)`システムコールが呼ばれます。それと類似して、WASI向けのGoプログラムにおいては、最終的にWasmランタイムが提供する[`fd_write`関数](https://github.com/WebAssembly/WASI/blob/main/legacy/preview1/docs.md#fd_write)を呼び出したいです。

しかし、Wasmランタイムの関数はもちろんGoの関数ではありませんから、それをGoの関数として呼び出せないと`syscall.Write`をGoで書くことができません。
そこでGoの関数宣言だけを書き、`go:wasmimport`ディレクティブを使って`wasi_snapshot_preview1`モジュールの`fd_write`関数をimportします。すると、Wasmランタイムが提供している該当の関数をGoの関数として、Goのプログラムから呼び出せるようになります。

:::message
この機能は`fd_write`関数以外にも`syscall`パッケージのwasiサポートにおいて広く使われています。

この記事のハンズオンでは、あえてその機能をmain.goから直接使ってみたということになります。
:::

# 実行する

実行するには、goの他にWasmランタイムが必要です。この記事では`wasmtime`を使います。

https://wasmtime.dev/

上記に従って`wasmtime`をインストールした後、次の2つを行います。

- wasi向けにmain.goをコンパイルして、wasi向けのバイナリ`main`を作る
- Wasmランタイムである`wasmtime`で、`main`を実行する

```
GOOS=wasip1 GOARCH=wasm go build -o main
wasmtime main
```

すると、次のようにターミナルに文字列が表示されるので、実験成功です。

```sh
wasmtime main
Hello, Wasm
```

# 使用したコードの全体

コードの全体は次のファイルにあります。

https://github.com/nobishino/wasmimport-study/blob/main/main.go

手元で実行したい場合の手順は次のようになります。

```
git clone https://github.com/nobishino/wasmimport-study
cd wasmimport-study
GOOS=wasip1 GOARCH=wasm go build -o main
wasmtime main
```

# この記事へのフィードバックについて

- この記事についてフィードバックやご意見がある場合、[GitHubリポジトリ](https://github.com/nobishino/zenn)にissueかPRを立てていただけると助かります。
  - ZennのコメントよりもGitHub上でのやり取りが好ましいです。
  - GitHub上ではissueを立てずにいきなりPRを立てても大丈夫です。
  - とりあえずカジュアルに聞きたい場合はXや[Gophers Slack](https://gophers.slack.com/join/shared_invite/zt-1vukscera-OjamkAvBRDw~qgPh~q~cxQ#/shared-invite/email)の`@Nobishii`にコンタクトしてもらえればと思います。

:::message
WasmやWASIを学び始めて間もないので、結構本質的な勘違いをしていたりして言葉遣いなどもおかしい可能性がありそうなので何かあったらぜひお願いします
:::
