package main

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
)

var keyGUID = []byte("258EAFA5-E914-47DA-95CA-C5AB0DC85B11")

type UpgradeHandler func(w http.ResponseWriter, r *http.Request)
type Websocket struct {
	upgradeHandler UpgradeHandler
}

func upgradeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("connection building")
	res := upgrade(w, r)
	if res != "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println(res)
	}

}

func computeKey(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	h.Write(keyGUID)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func upgrade(w http.ResponseWriter, r *http.Request) string {
	if r.Method != "GET" {
		return "Method is not get"
	}
	if r.Header.Get("Upgrade") != "websocket" {
		return " upgrade not websocket"
	}
	if r.Header.Get("Connection") != "Upgrade" {
		return "connection not upgrade"
	}

	fmt.Println("Receive correct connection")
	wkey := r.Header.Get("Sec-WebSocket-Key")
	rkey := computeKey(wkey)

	hj, ok := w.(http.Hijacker)

	if !ok {
		return "hijack failed"
	}

	_, rwb, err := hj.Hijack()
	if err != nil {
		return "hijack failed"
	}
	var p []byte
	p = append(p, "HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Accept: "...)
	p = append(p, rkey...)
	p = append(p, "\r\n\r\n"...)

	_, err = rwb.Write(p)
	rwb.Flush()
	fmt.Println("send response correct")
	if err != nil {
		return "response failed"
	}

	return ""
}
func (w *Websocket) listen(url string) {
	fmt.Println("listening ...")
	mux := http.DefaultServeMux
	mux.HandleFunc("/", upgradeHandler)
	log.Fatal(http.ListenAndServe(":8080", mux))
}
