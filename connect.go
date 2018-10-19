package main

import (
	"bufio"
	"fmt"
	"net"
	"sync"
)

var ws *Websocket

const (
	BUFSIZE = 4096 * 4
)

const (
	needPing   = 1
	connClosed = 2
)

type messageHandler func(caller *Connect, m *Message) int
type Connect struct {
	mu    sync.Mutex //lock for write
	conn  net.Conn
	state int
	group *Group
	ws    *Websocket
	id    string
	//write buffer and message buffer
	wbuf      []byte
	wbufIndex int
	mbuf      []byte
	mbufIndex int
	//write field
	bw bufio.Writer

	//read field
	br bufio.Reader

	//callback for user
	mh messageHandler
	//queue
	inQueue chan *Message
}

func (c *Connect) Ping() int {
	return 0
}
func echoHandler(caller *Connect, m *Message) int {
	fmt.Println("...echo")
	r := make([]byte, 0)
	r = append(r, m.head...)
	r = append(r, m.data...)
	m.data = r
	fmt.Printf("length = %d\n", len(m.data))
	caller.inQueue <- m

	return 1
}
func (c *Connect) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.conn.Close()
	c.ws.removeConnect(c)
}

func getWs() *Websocket {
	return ws
}
