// Message Passing型リトマステスト（half-atomic版・ガードなし）
//
// 検証の趣旨:
// 第6章の【本題】の実験。データ役のxはふつうの変数のまま、フラグ役のyだけを
// atomicにしたとき、「フラグは立っているのにデータが見えない」
// (r2, r1) = (1, 0) が起きないこと（yの観測がsynchronized beforeを作り、
// x = 1 の順序をr1 := xへ伝播させること）を確認する。
//
// 期待される結果:
//   - 通常実行: 何度実行しても panic しない（(1, 0)は観測されない）
//   - go run -race: yのLoadが0を観測する実行（パターンB）が混ざると
//     x上のdata raceが報告される。「panicは絶対に起きないのに、
//     data raceはある」＝ data raceが実行ごとの性質であることの実証。
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
	r2 := y.Load() // (3) atomicな読み込み
	r1 := x        // (4) ふつうの読み込み
	if r2 == 1 && r1 == 0 {
		panic("(r2, r1) = (1, 0) が観測されました")
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
