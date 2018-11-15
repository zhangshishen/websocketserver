package stun

import (
	"fmt"
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
	conn      *net.UDPConn
	port      int
	address   int
}

func (a *Agent) bind(domainName string, port int, laddr *net.UDPAddr) int {
	if a.conn != nil {
		wlog.Out("[stun] bind failed, already has conn\n")
		return -1
	}
	fmt.Println(domainName + ":" + strconv.Itoa(port))
	addrs, err := net.ResolveIPAddr("ip", domainName)
	if err != nil {
		wlog.Out("[stun] dns resolve failed")
		return -1
	}
	var raddr net.UDPAddr
	raddr.IP = net.ParseIP(addrs.String())
	raddr.Port = port
	conn, err := net.DialUDP("udp", laddr, &raddr)
	if err != nil {
		wlog.Out("[stun] udp dail failed\n")
		return -1
	}
	a.conn = conn
	wlog.Out("[stun] bind success\n")
	return 0

}

func (a *Agent) listen(callback attributeCallback) {
	buf := make([]byte, 1024)

	l, err := a.conn.Read(buf)
	if err != nil {

		wlog.Out("[stun] read udp error\n")
		return
	}
	if l < HEADERSIZE {
		wlog.Out("[stun] size too small")
		return
	}

	msg := Message{}
	msg.hcallback = callback

	len := msg.parseMessageHeader(buf[0:HEADERSIZE])

	if (l - HEADERSIZE) != len {
		wlog.Out("[stun] parse failed, length not match")
		return
	}
	/*
		for last > 0 {
			wlog.Out("[stun] start read")
			l, err := a.conn.Read(buf[len-last : len])
			if err != nil {
				wlog.Out("[stun] read udp error\n")
			}
			wlog.Out("[stun] read size", l)
			last -= l
		}*/

	msg.parseMessageAttribute(buf[HEADERSIZE : HEADERSIZE+len])

}

//make a sync message and send to server
func (a *Agent) request(method int, class int) int {
	if a.conn == nil {
		wlog.Out("[stun] bind failed, no conn exist\n")
		return -1
	}
	m := Message{}
	m.attribute = make([]Attribute, 0)
	m.make(method, class)

	//m.addAttribute(Attribute{})
	buf := m.generateBuffer(nil)
	fmt.Printf("message is %v\n", buf)

	l, err := a.conn.Write(buf)
	if l < len(buf) || err != nil {
		wlog.Out("[stun] send msg failed!\n")
	}
	wlog.Out("[stun] request success,send size %d\n", l)
	return 0
}
