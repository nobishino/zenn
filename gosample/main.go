// GOEXPERIMENT=rangefunc
// (*)実行するとgo vetがエラーを出しますが、実行はできています
package main

import "github.com/nobishino/gocoro/iter"

func main() {
	for k, v := range seq2() {
		println(k, v)
	}
}

func seq2() iter.Seq2[string, int] {
	values := []string{"a", "b", "c"}
	return func(yield func(string, int) bool) {
		for i := range 10 {
			if !yield(values[i%3], i) {
				break
			}
		}
	}
}
