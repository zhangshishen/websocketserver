package main

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
)

func upgradeHandler(w http.ResponseWriter, r *http.Request) {
	//fmt.Printf("start handshaking %d\n", icrementID)
	conn, err := upgrade(w, r)
	if err != "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println(err)
		return
	}
	//todo get websocket with url
	ws := getWs("Default")
	//default selector,
	g := ws.gs(r)
	id := ws.is(r)
	//define your handler
	go ws.addConn(conn, g, id, broadcastHandler)

}

func computeKey(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	h.Write(keyGUID)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

var (
	icrementID = 1
)

func upgrade(w http.ResponseWriter, r *http.Request) (*Connect, string) {
	if r.Method != "GET" {
		return nil, "Method is not get"
	}
	if r.Header.Get("Upgrade") != "websocket" {
		return nil, " upgrade not websocket"
	}
	if !strings.Contains(r.Header.Get("Connection"), "Upgrade") {
		for k, v := range r.Header {
			fmt.Printf("%s :%s\n", k, v)
		}
		return nil, "connection not upgrade"
	}

	//fmt.Println("upgrade: Receive correct connection")
	wkey := r.Header.Get("Sec-WebSocket-Key")
	rkey := computeKey(wkey)

	hj, ok := w.(http.Hijacker)

	if !ok {
		return nil, "hijack failed"
	}

	c, rwb, err := hj.Hijack()
	rwb.Flush()
	if err != nil {
		return nil, "hijack failed"
	}
	var p []byte
	p = append(p, "HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Accept: "...)
	p = append(p, rkey...)
	p = append(p, "\r\n\r\n"...)

	_, err = rwb.Write(p)
	rwb.Flush()

	//fmt.Printf("%d handshake success\n", icrementID)

	if err != nil {
		return nil, "response failed"
	}
	//create default conn
	conn := new(Connect)
	conn.conn = c //native socket fd
	conn.num = icrementID
	icrementID++
	conn.wbuf = make([]byte, BUFSIZE)
	conn.mbuf = make([]byte, BUFSIZE)

	conn.outQueue = make(chan *Message, 256)
	conn.inQueue = make(chan *Message, 256)
	conn.ctx = make(chan int)

	return conn, ""
}
