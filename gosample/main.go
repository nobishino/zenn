package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"strings"
)

func main() {
	in := `"aaa","b"bb","ccc"
ddd,ee"e,e"ee,f
`
	r := csv.NewReader(strings.NewReader(in))
	r.LazyQuotes = true

	for {
		// 1行を[]stringで返す
		record, err := r.Read()
		if err == io.EOF {
			// 終了した時はio.EOFを返す
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(record, len(record))
	}
}

const csvStr = `first_name,last_name,username
"Rob","Pike",rob
Ken,Thompson,ken
"Robert","Griesemer","gri"
`
