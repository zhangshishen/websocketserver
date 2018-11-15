package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
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
	mu             sync.RWMutex
	cmu            sync.Mutex
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
	w.cmu.Lock()
	defer w.cmu.Unlock()
	//retrive group map TODO
	delete(w.connMap, c.id)
	//retrive ws map

}
func (w *Websocket) broadcastWithoutSelf(c *Connect, m *Message) {

}

func (w *Websocket) broadcast(c *Connect, m *Message) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	g := c.group

	if g == nil {
		fmt.Printf("connect has no group!\n")
		os.Exit(3)
	}

	g.broadcast(m)
}

func (w *Websocket) init() {
	w.group = make(map[string]*Group)
	w.connMap = make(map[string]*Connect)

	w.gs = defaultGroupSelector
	w.is = defaultIDSelector
}

func defaultGroupSelector(r *http.Request) string {
	return "Default"
}

var incrementID int32 = 0

func defaultIDSelector(r *http.Request) string {
	i := atomic.AddInt32(&incrementID, 1)
	return strconv.Itoa(int(i))
}

func (w *Websocket) getGroup(s string) *Group {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return w.group[s]
}

func (w *Websocket) addToMap(c *Connect, id string) {
	w.cmu.Lock()
	defer w.cmu.Unlock()
	w.connMap[id] = c
}

func (w *Websocket) addConn(c *Connect, g string, id string, mh messageHandler) {
	//todo
	if g == "" {
		g = "Default"
	}

	//create map between group and conn

	//create map between conn and ws
	w.addToMap(c, id)

	c.ws = w
	//init handler

	c.mh = mh
	w.addToGroup(c, g)
	//reader goroutine
	go readRoutine(c, c.outQueue)
	//tbuf := make([]byte, 1024)
	for {

		wmsg := <-c.inQueue

		if wmsg == nil {

			w.releaseConn(c)
			wlog.Out("connection closed")
			return

		} else {
			tbuf := make([]byte, 0)
			tbuf = append(tbuf, wmsg.head...)
			tbuf = append(tbuf, wmsg.data...)
			//fmt.Println("out message\n")
			_, err := c.conn.Write(tbuf)

			if err != nil || wmsg.op == connClosed { //passive close or adjective close
				w.releaseConn(c)
				wlog.Out("connection closed")
				return
			}
		}

		/*
			if n != len(wmsg.data) {
				fmt.Printf("fatal error write %d byte\n", n)
			}*/

	}
}

func (w *Websocket) releaseConn(c *Connect) {
	//fmt.Println("close conn\n")
	c.conn.Close()
	close(c.ctx)
	//remove from group
	w.removeFromGroupByPointer(c)
	//remove from map
	w.removeConnect(c)
}

func readRoutine(c *Connect, outQueue chan *Message) {

	bufc := make(chan []byte)

	go fillMsg(bufc, outQueue, c)

	for {
		n, err := c.conn.Read(c.wbuf)
		fmt.Printf("%d receive package,length = %d\n", c.num, n)
		if err != nil {

			c.conn.Close()
			close(bufc)
			c.Write(nil)

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
	w.mu.Lock()
	defer w.mu.Unlock()

	g := w.group[group]

	if g == nil {

		g = new(Group)
		g.conn = make(map[*Connect]bool)

		w.group[group] = g
	}

	g.addConn(c)

}
func (w *Websocket) removeFromGroupByPointer(c *Connect) {
	//todo
	w.mu.Lock()
	defer w.mu.Unlock()

	g := c.group

	if g == nil {
		wlog.Out("group name %d not found\n", g.name)
		return
	}

	if g.removeConn(c) == 0 { //group ref count equal 0
		delete(w.group, g.name)
	}

}
func (w *Websocket) removeFromGroup(c *Connect, group string) {
	//todo
	w.mu.Lock()
	defer w.mu.Unlock()

	g := w.group[group]

	if g == nil {
		wlog.Out("group name %d not found\n", group)
		return
	}

	if g.removeConn(c) == 0 {
		delete(w.group, g.name)
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
	//if url == "" || url == "/" {
	addWs("Default", w)
	//} else {
	//	addWs(url, w)
	//}

	mux := http.DefaultServeMux
	mux.HandleFunc(url, upgradeHandler)
	//mux.HandleFunc("/", home)
	log.Fatal(http.ListenAndServe(":8080", mux))

}
