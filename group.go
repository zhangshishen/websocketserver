package main

import (
	"fmt"

	"./wlog"
)

type Group struct {
	conn map[*Connect]bool
	mh   messageHandler
	ref  int
	name string
}

func (g *Group) addConn(c *Connect) {
	//todo

	g.conn[c] = true
	c.group = g
	g.ref++

}
func (g *Group) removeConn(c *Connect) int { //return the reference
	//todo

	_, ok := g.conn[c]
	if ok {
		delete(g.conn, c)
	} else {
		wlog.Out("remove connection from group %d failed! [group has no connection]\n", g.name)
		return -1
	}
	c.group = nil
	g.ref--
	if g.ref < 0 {
		g.ref++
		wlog.Out("remove connection from group %d failed! [ref less than zero]\n", g.name)
		return -1
	}
	return g.ref

}

func (g *Group) broadcast(m *Message) {
	fmt.Printf("b message %s\n", m.data)
	for k, v := range g.conn {
		if v {
			k.Write(m)
		}
	}
	//todo

}
