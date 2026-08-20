// Store Buffering型リトマステスト（全atomic版）
//
// 検証の趣旨:
// 第6章【ルール2の意味】の実験。Store Bufferingの2変数を両方atomicにすると、
// 暗黙的全順序（逐次一貫性）により (ry, rx) = (0, 0) が不可能になることを
// 確認する。(0, 0)を仮定すると全順序S上で
//   y.Load < y.Store < x.Load < x.Store < y.Load
// という循環ができて矛盾する、というのが仕様上の理由。
// この結論はhappens-beforeグラフからは導けない（(0,0)の実行には
// synchronized before辺が1本もなく、hb上の矛盾がない）ことに注意。
// 全順序がhbとは独立の制約であることを示す例になっている。
//
// 期待される結果:
//   - どの環境でも (ry, rx) = (0, 0) は0回
//   - go run -race でも何も報告されない（全アクセスがatomic）
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

var x, y atomic.Int64 // どちらもatomicな変数
var rx, ry int64
var wg sync.WaitGroup

func f() {
	defer wg.Done()
	y.Store(1)    // (1) yへのatomicな書き込み
	rx = x.Load() // (2) xからのatomicな読み込み
}

func g() {
	defer wg.Done()
	x.Store(1)    // (3) xへのatomicな書き込み
	ry = y.Load() // (4) yからのatomicな読み込み
}

// 実験を1回行う関数
func exec() (int64, int64) {
	x.Store(0)
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
