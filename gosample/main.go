package main

import (
	"encoding/csv"
	"fmt"
	"strings"
)

func main() {
	in := `a,b,c
1,2,3
4,"5"five,6
7,"8",9
`
	r := csv.NewReader(strings.NewReader(in))
	r.LazyQuotes = true

	record, err := r.ReadAll()
	if err != nil {
		panic(err)
	}
	fmt.Println(len(record))

	fmt.Println(record[2][1])
}
