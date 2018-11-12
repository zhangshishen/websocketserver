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

type Agent struct {
	agentType int

	conn    net.Conn
	port    int
	address int
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
func (a *Agent) listen(port int) {

}
func (a *Agent) request() int {
	if a.conn == nil {
		wlog.Out("[stun] bind failed, no conn exist\n")
		return -1
	}
	m := Message{}
	m.mtype = 1
	return 0
}
