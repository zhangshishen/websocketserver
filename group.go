package main

import (
	"sync"
)

type ctx <-chan int
type Group struct {
	mu   sync.Mutex
	conn map[*Connect]int
}

func (g *Group) addConn(c *Connect) {
	//todo
}
func (g *Group) broadCast(m *Message, context ctx) {
	g.mu.Lock()
	defer g.mu.Unlock()
	//todo

}
