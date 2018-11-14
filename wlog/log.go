package wlog

import (
	"fmt"
)

func Out(a ...interface{}) {
	fmt.Println(a)
}

func ASSERT(b bool, out string) {
	if b == false {
		fmt.Println("ASSERT FAILED!", out)
	}
}
