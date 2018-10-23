package main

var wsMap map[string]*Websocket = make(map[string]*Websocket)

func getWs(name string) *Websocket {
	return wsMap[name]
}

func addWs(name string, ws *Websocket) {
	wsMap[name] = ws
}
