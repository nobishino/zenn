// Store Buffering型リトマステスト（half-atomic版）
//
// 検証の趣旨:
// 第6章「全順序が縛るのはatomic演算だけ」の実験。Store Bufferingの2変数のうち
// yだけをatomicにしても (ry, rx) = (0, 0) は防げないことを確認する。
//
// もし「観測しないLoadからStoreへ」happens-before辺を描いてよいなら、
// ry == 0 のとき [x = 1] < [y.Load] < [y.Store] < [rx = x] となって
// rx == 1 が保証されるはずである。しかし実機では (0, 0) が観測されるので、
// 暗黙的全順序上の前後関係をhappens-before辺として扱う方法は誤りだとわかる
// （synchronized beforeは「Wが直接観測した部分」に限られる: Requirement 2）。
//
// 期待される結果（環境依存）:
//   - amd64では (ry, rx) = (0, 0) が少数回観測される
//   - go run -race を付けるとx上のdata raceが報告される
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

var x int64        // ふつうの変数
var y atomic.Int64 // atomicな変数
var rx, ry int64
var wg sync.WaitGroup

func f() {
	defer wg.Done()
	y.Store(1) // (1) yへのatomicな書き込み
	rx = x     // (2) xからのふつうの読み込み
}

func g() {
	defer wg.Done()
	x = 1         // (3) xへのふつうの書き込み
	ry = y.Load() // (4) yからのatomicな読み込み
}

// 実験を1回行う関数
func exec() (int64, int64) {
	x = 0
	y.Store(0)
	rx, ry = 0, 0
	wg.Add(2)
	go f()
	go g()
	wg.Wait()
	return ry, rx
}

func main() {
	counts := map[[2]int64]int{}
	for i := 0; i < 1_000_000; i++ {
		a, b := exec()
		counts[[2]int64{a, b}]++
	}
	for _, k := range [][2]int64{{0, 0}, {0, 1}, {1, 0}, {1, 1}} {
		fmt.Printf("(ry, rx) = (%d, %d): %d回\n", k[0], k[1], counts[k])
	}
}
