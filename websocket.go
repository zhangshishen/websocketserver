package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

const (
	//start    = 0
	received = 2
	closing  = 3

	//needPing = 5
)

var keyGUID = []byte("258EAFA5-E914-47DA-95CA-C5AB0DC85B11")

type UpgradeHandler func(w http.ResponseWriter, r *http.Request)

//type broadcastHandler func(message []byte, group string) int
type multicastHandler func(message []byte, multiGroup []string) int
type unicastHandler func(message []byte, id string) int

type Websocket struct {
	mu             sync.Mutex
	upgradeHandler UpgradeHandler
	//id map, send msg by id
	connMap map[string]*Connect
	//group map
	group map[string]*Group
	//queue

	//callback
	//broadcastHandler
	unicastHandler
}

func (w *Websocket) removeConnect(c *Connect) {
	w.mu.Lock()
	defer w.mu.Unlock()
	//retrive group map TODO
	//retrive ws map

}

func (w *Websocket) addConn(c *Connect, g string, id string) {
	//todo
	if g == "" {
		g = "Default"
	}
	//fmt.Println("#connect: create success ")
	//create map between group and conn
	group := w.group[g]
	if group == nil {
		group = new(Group)
		group.conn = make(map[*Connect]bool)
	}
	c.group = group
	group.addConn(c)
	//create map between conn and ws
	w.mu.Lock()
	w.connMap[id] = c
	w.mu.Unlock()
	//init var

	c.ctx = make(chan int)

	c.mh = echoHandler
	c.ws = w
	outQueue := make(chan *Message, 256)
	c.inQueue = make(chan *Message, 256)
	//reader goroutine
	go readRoutine(c, outQueue)

	for {

		wmsg := <-c.inQueue
		if wmsg == nil {
			//close connect
			return
		}

		n, err := c.conn.Write(wmsg.data)

		if err != nil { //passive close
			close(c.ctx)
			return
		}
		if wmsg.op == connClosed { //adjective close
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
}

func (w *Websocket) removeFromGroup(c *Connect, group string) {
	//todo
}

func (w *Websocket) listen(url string) {

	//fmt.Println("listening ...")
	mux := http.DefaultServeMux
	mux.HandleFunc("/", upgradeHandler)
	log.Fatal(http.ListenAndServe(":8080", mux))
}
