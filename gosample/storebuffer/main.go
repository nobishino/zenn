// Store Buffering型リトマステスト（ふつうの変数版）
//
// 検証の趣旨:
// 第6章【ルール2の意味】の実験。ふつうの変数だけを使ったStore Buffering型の
// プログラムで、逐次一貫モデルでは説明できない (ry, rx) = (0, 0) が
// 実機で観測されることを確認する。
// 第2章のMessage Passing Testはamd64(TSO)ではまず再現しないが、
// Store Bufferingによる並べ替えはamd64でも起きるため、この例は
// amd64でSC破れを観測できる。名前の由来はハードウェアのストアバッファ。
//
// 期待される結果（環境依存）:
//   - amd64では (ry, rx) = (0, 0) が少数回観測される
//   - go run -race を付けるとdata raceが報告される（x, yともにdata race）
package main

import (
	"fmt"
	"sync"
)

var x, y int64 // どちらもふつうの変数
var rx, ry int64
var wg sync.WaitGroup

func f() {
	defer wg.Done()
	y = 1  // (1) yへの書き込み
	rx = x // (2) xからの読み込み
}

func g() {
	defer wg.Done()
	x = 1  // (3) xへの書き込み
	ry = y // (4) yからの読み込み
}

// 実験を1回行う関数
func exec() (int64, int64) {
	x, y, rx, ry = 0, 0, 0, 0
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
