package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"./wlog"
)

const (
	//start    = 0
	received = 2
	closing  = 3

	//needPing = 5
)

var keyGUID = []byte("258EAFA5-E914-47DA-95CA-C5AB0DC85B11")

type UpgradeHandler func(w http.ResponseWriter, r *http.Request)
type groupSelector func(r *http.Request) string
type idSelector func(r *http.Request) string

//type broadcastHandler func(message []byte, group string) int

type Websocket struct {
	mu             sync.Mutex
	upgradeHandler UpgradeHandler
	//id map, send msg by id
	connMap map[string]*Connect
	//group map
	group map[string]*Group
	//queue

	//selector
	gs groupSelector
	is idSelector
}

func (w *Websocket) removeConnect(c *Connect) {
	w.mu.Lock()
	defer w.mu.Unlock()
	//retrive group map TODO
	//retrive ws map

}

func defaultMsgSelector(r *http.Request) string {
	return "Default"
}

func defaultIDSelector(r *http.Request) string {
	return ""
}

func (w *Websocket) addConn(c *Connect, g string, id string) {
	//todo
	if g == "" {
		g = "Default"
	}

	//create map between group and conn
	w.addToGroup(c, g)
	//create map between conn and ws
	w.mu.Lock()
	w.connMap[id] = c
	w.mu.Unlock()
	c.ws = w
	//init handler

	c.mh = echoHandler

	//reader goroutine
	go readRoutine(c, c.outQueue)

	for {

		wmsg := <-c.inQueue
		if wmsg == nil {
			//close connect
			continue
		}

		n, err := c.conn.Write(wmsg.data)

		if err != nil || wmsg.op == connClosed { //passive close or adjective close
			c.conn.Close()
			close(c.ctx)
			return
		}

		if n != len(wmsg.data) {
			fmt.Printf("fatal error write %d byte\n", n)
		}

	}
}

func readRoutine(c *Connect, outQueue chan *Message) {

	bufc := make(chan []byte)

	go fillMsg(bufc, outQueue, c)

	for {
		n, err := c.conn.Read(c.wbuf)
		if err != nil {
			c.conn.Close()
			close(bufc)
			return
		}
		if n == 0 {
			continue
		}

		bufc <- c.wbuf[:n]

		t := c.mbuf
		c.mbuf = c.wbuf
		c.wbuf = t
	}
}

func pingPong(c *Connect, t time.Duration, context chan int) {

	for {
		select {
		case <-context:
			return
		case <-time.After(t * time.Second):
			c.state &= needPing
		}
	}

}
func (w *Websocket) addToGroup(c *Connect, group string) {
	//todo
	g := w.group[group]

	if g == nil {
		g = new(Group)
		g.conn = make(map[*Connect]bool)
		w.mu.Lock()
		w.group[group] = g
		w.mu.Unlock()
	}

	g.addConn(c)

}
func (w *Websocket) removeFromGroupByPointer(c *Connect) {
	//todo
	g := c.group
	if g == nil {
		wlog.Out("group name %d not found\n", g.name)
		return
	}

	ref := g.removeConn(c)

	if ref == 0 {
		w.mu.Lock()
		delete(w.group, g.name)
		w.mu.Unlock()
	}

}
func (w *Websocket) removeFromGroup(c *Connect, group string) {
	//todo
	g := w.group[group]
	if g == nil {
		wlog.Out("group name %d not found\n", group)
		return
	}

	ref := g.removeConn(c)

	if ref == 0 {
		w.mu.Lock()
		delete(w.group, g.name)
		w.mu.Unlock()
	}

}
func (w *Websocket) addGroupHandler(Group string, mh messageHandler) {

}
func (w *Websocket) addHandler(URL string, mh messageHandler) {

}
func (w *Websocket) listen(url string) {

	//fmt.Println("listening ...")
	if len(w.group) == 0 {
		w.group = make(map[string]*Group)
	}
	if len(w.connMap) == 0 {
		w.connMap = make(map[string]*Connect)
	}
	if url == "" || url == "/" {
		addWs("Default", w)
	} else {
		addWs(url, w)
	}

	mux := http.DefaultServeMux
	mux.HandleFunc("/", upgradeHandler)
	log.Fatal(http.ListenAndServe(":8080", mux))

}
