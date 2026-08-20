// Message Passing型リトマステスト（half-atomic版・ガード付き）
//
// 検証の趣旨:
// 第6章「ガードすればdata race freeになる」の実験。ふつうの変数xの読み込みを
// 「atomicなフラグyのLoadが1を観測したときだけ」実行するようにガードすると、
// どの実行にもdata raceがなくなる（= プログラムがdata race freeになる）ことを
// 確認する。publicationパターンの最小例。
//
// 期待される結果:
//   - 通常実行: 何度実行しても panic しない
//   - go run -race: 何も報告されない（ガードなし版との対比）
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

var x int64        // ふつうの変数
var y atomic.Int64 // atomicな変数
var wg sync.WaitGroup

func f() {
	defer wg.Done()
	x = 1      // (1) ふつうの書き込み
	y.Store(1) // (2) atomicな書き込み
}

func g() {
	defer wg.Done()
	if y.Load() == 1 { // (3) atomicな読み込み
		r1 := x // (4) ふつうの読み込み（(3)が1を観測したときだけ実行される）
		if r1 == 0 {
			panic("(r2, r1) = (1, 0) が観測されました")
		}
	}
}

// 実験を1回行う関数
func exec() {
	x = 0 // 前回の実験の結果をリセットする
	y.Store(0)
	wg.Add(2)
	defer wg.Wait()
	go f()
	go g()
}

func main() {
	for i := 0; i < 1_000_000; i++ {
		exec()
	}
	fmt.Println("(r2, r1) = (1, 0) は一度も観測されませんでした")
}
