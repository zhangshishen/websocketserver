package main

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/http"
)

func upgradeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("connection building")
	conn, err := upgrade(w, r)
	if err != "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println(err)
	}

	ws := getWs()

	go ws.addConn(conn, "", "")
}

func computeKey(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	h.Write(keyGUID)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func upgrade(w http.ResponseWriter, r *http.Request) (*Connect, string) {
	if r.Method != "GET" {
		return nil, "Method is not get"
	}
	if r.Header.Get("Upgrade") != "websocket" {
		return nil, " upgrade not websocket"
	}
	if r.Header.Get("Connection") != "Upgrade" {
		return nil, "connection not upgrade"
	}

	fmt.Println("upgrade: Receive correct connection")
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

	//fmt.Println("send response correct")

	if err != nil {
		return nil, "response failed"
	}
	conn := new(Connect)
	conn.conn = c

	conn.wbuf = make([]byte, BUFSIZE)
	conn.mbuf = make([]byte, BUFSIZE)
	return conn, ""
}
