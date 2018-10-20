package main

import (
	"sync"
)

type ctx <-chan int
type Group struct {
	mu   sync.Mutex
	conn map[*Connect]bool
}

func (g *Group) addConn(c *Connect) {
	//todo
	g.mu.Lock()
	defer g.mu.Unlock()

	g.conn[c] = true
}
func (g *Group) broadCast(m *Message) {
	g.mu.Lock()
	defer g.mu.Unlock()

	for k, v := range g.conn {
		if v {
			k.Write(m)
		}
	}
	//todo

}
