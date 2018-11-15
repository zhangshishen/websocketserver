package stun

import (
	"encoding/binary"
	"fmt"
	"net"
)

func callback(h *Attribute, m *Message) {
	h.print()
	if len(h.value) != 8 {
		fmt.Printf("[stun] xor addvice not correct\n")
	}
	switch h.htype {
	case XOR_MAPPED_ADDRESS:
		port := binary.BigEndian.Uint16(h.value[2:4])
		ip := binary.BigEndian.Uint32(h.value[4:8])
		port = port ^ (uint16)(MAGICCOOKIE>>16)
		ip = ip ^ MAGICCOOKIE
		var ipaddr net.IP
		ipaddr = make(net.IP, 4)
		binary.BigEndian.PutUint32(ipaddr, ip)

		fmt.Printf("port = %d,ip = %s\n", port, ipaddr.String())
	}

}

func Stun_test() {
	agent := Agent{}
	agent.bind("stun.l.google.com", 19302, nil)
	agent.request(binding, request)
	agent.listen(callback)
}
