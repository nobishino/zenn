// GOEXPERIMENT=rangefunc
// (*)実行するとgo vetがエラーを出しますが、実行はできています
package main

import "github.com/nobishino/gocoro/iter"

func main() {
	for v := range seq() {
		println(v)
	}
}

func seq() iter.Seq[string] {
	values := []string{"a", "b", "c"}
	return func(yield func(string) bool) {
		for i := range 10 {
			if !yield(values[i%3]) {
				break
			}
		}
	}
}
