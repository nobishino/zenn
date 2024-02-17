// GOEXPERIMENT=rangefunc
// (*)実行するとgo vetがエラーを出しますが、実行はできています
package main

import (
	"fmt"

	"github.com/nobishino/gocoro/iter"
)

func main() {
	s1 := createIntSeq(1, 2, 3, 4, 5)
	s2 := createIntSeq(1, 2, 3, 5, 4)
	s3 := createIntSeq(1, 2, 3, 4, 5)

	fmt.Println(equalIntValues(s1, s2)) // false
	fmt.Println(equalIntValues(s1, s3)) // true
}

// a, bからえられる値が全て同じかどうかを判定する
func equalIntValues(a, b iter.Seq[int]) bool {
	nextA, stopA := iter.Pull(a)
	defer stopA()
	nextB, stopB := iter.Pull(b)
	defer stopB()
	for {
		av, aok := nextA()
		bv, bok := nextB()
		if av != bv || aok != bok {
			return false
		}
		if !aok {
			return true
		}
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
