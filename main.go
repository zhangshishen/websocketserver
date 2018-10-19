package main

import (
	"fmt"
)

func main() {
	fmt.Println("run server")
	w := Websocket{}
	w.listen("")

}
