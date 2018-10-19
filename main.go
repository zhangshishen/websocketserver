package main

import (
	"fmt"
)

func main() {
	fmt.Println("run server")

	w := Websocket{}
	ws = &w
	ws.group = make(map[string]*Group)
	ws.connMap = make(map[string]*Connect)
	w.listen("")

}
