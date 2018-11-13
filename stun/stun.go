package stun

import (
	"net"
	"strconv"

	"../wlog"
)

const (
	client = iota
	server
)

const (
	HEADERSIZE = 20
)

type Agent struct {
	agentType int
	method    int
	conn      net.Conn
	port      int
	address   int
}

func (a *Agent) bind(domainName string, port int) int {
	if a.conn != nil {
		wlog.Out("bind failed, already has conn\n")
		return -1
	}
	conn, err := net.Dial("udp", domainName+":"+strconv.Itoa(port))
	if err != nil {
		wlog.Out("[stun] udp dail failed\n")
		return -1
	}
	a.conn = conn
	return 0

}
func (a *Agent) listen(callback attributeCallback) {
	buf := make([]byte, 1024)

	last := HEADERSIZE
	for last > 0 {
		l, err := a.conn.Read(buf[HEADERSIZE-last : HEADERSIZE])
		if err != nil {
			wlog.Out("[stun] read udp error\n")
		}
		last -= l
	}

	msg := Message{}
	msg.hcallback = callback

	len := msg.parseMessageHeader(buf)

	last = len

	for last > 0 {
		l, err := a.conn.Read(buf[len-last : len])
		if err != nil {
			wlog.Out("[stun] read udp error\n")
		}
		last -= l
	}
	msg.parseMessageAttribute(buf)

}

//make a sync message and send to server
func (a *Agent) request(method int, class int) int {
	if a.conn == nil {
		wlog.Out("[stun] bind failed, no conn exist\n")
		return -1
	}
	m := Message{}
	m.make(method, class)
	//m.addAttribute(Attribute{})
	buf := m.generateBuffer(nil)

	l, err := a.conn.Write(buf)
	if l < len(buf) || err != nil {
		wlog.Out("[stun] send msg failed!\n")
	}
	return 0
}
