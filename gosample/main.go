// GOEXPERIMENT=rangefunc
// (*)実行するとgo vetがエラーを出しますが、実行はできています
package main

import (
	"fmt"
)

func main() {
	s1 := createIntSeq(1, 3, 5, 8)
	s2 := createIntSeq(2, 4, 6, 7)

	for v := range mergeSortedIntSeq(s1, s2) {
		if v >= 7 {
			break
		}
		fmt.Println(v) // 1,2,3,4,5,6
	}

}

// a, bという2つのシーケンスを受け取り、それをマージしたシーケンスを返す
// その際、a ,bが昇順ソート済みだと仮定して、そのソート順序を維持する
func mergeSortedIntSeq(a, b Seq[int]) Seq[int] {
	nextA, stopA := Pull(a)
	nextB, stopB := Pull(b)
	av, aok := nextA()
	bv, bok := nextB()
	return func(yield func(int) bool) {
		for more := true; more; {
			switch {
			case aok && (av <= bv || !bok):
				more = yield(av)
				av, aok = nextA()
			case bok:
				more = yield(bv)
				bv, bok = nextB()
			default:
				return
			}
		}
		stopA()
		stopB()
	}
}

// 可変長引数を受け取り、それを要素とするシーケンスを返す
func createIntSeq(xs ...int) Seq[int] {
	return func(yield func(int) bool) {
		for _, x := range xs {
			if !yield(x) {
				break
			}
		}
	}
}
