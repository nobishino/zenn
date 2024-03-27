package main

import (
	"fmt"
)

type MyStruct struct {
	X int
	Y int
}

var a *MyStruct = &as[0]

var as = [2]MyStruct{
	{0, 0},
	{1, 1},
}

func main() {
	go readerGoroutine(a)
	for i := 0; ; i++ {
		a = &as[i%2] // 設定更新
	}
}

func readerGoroutine(value *MyStruct) {
	for {
		v := *value
		if v.X != v.Y {
			fmt.Println("Value: ", *value)
		}
	}
}
