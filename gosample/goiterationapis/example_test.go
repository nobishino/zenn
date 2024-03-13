package goiterationapis

import (
	"fmt"
	"net"
)

func ExampleIPv4() {
	fmt.Println(net.IPv4(8, 8, 8, 8))

	// Output:
	// 8.8.8.8
}
