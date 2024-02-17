// GOEXPERIMENT=rangefunc
// (*)実行するとgo vetがエラーを出しますが、実行はできています
package main

import (
	"fmt"

	"github.com/nobishino/gocoro/iter"
)

func main() {
	s1 := createIntSeq(1, 3, 5, 8)
	s2 := createIntSeq(2, 4, 6, 7)

	for v := range mergeSortedIntSeq(s1, s2) {
		fmt.Println(v) // 1,2,4,5,6,7,8
	}

}

// a, bという2つのシーケンスを受け取り、それをマージしたシーケンスを返す
// その際、a ,bが昇順ソート済みだと仮定して、そのソート順序を維持する
func mergeSortedIntSeq(a, b iter.Seq[int]) iter.Seq[int] {
	nextA, stopA := iter.Pull(a)
	nextB, stopB := iter.Pull(b)
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

func createIntSeq(xs ...int) iter.Seq[int] {
	return func(yield func(int) bool) {
		for _, x := range xs {
			if !yield(x) {
				break
			}
		}
	}
}
