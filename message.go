package main

import (
	"encoding/binary"
	"fmt"
)

const (
	//return value

	fin      = 1
	ping     = 2
	shutdown = 3
	succ     = 4
	small    = 6
	ferror   = 7
)

const (
	continuation = 100
	text         = 101
	bin          = 102
	closed       = 103
)

type Message struct {
	head       []byte
	op         int
	fin        bool
	maskingKey []byte
	length     uint64
	data       []byte
}

func testBuf(buf []byte, bc chan []byte) ([]byte, bool) {

	if len(buf) == 0 {
		buf = <-bc
		if len(buf) == 0 {
			return buf, false
		}

	}
	return buf, true
}
func fillMsg(bc chan []byte, mc chan<- *Message, c *Connect) {

	var buf []byte
	err := true
	var m *Message

	defer fmt.Println("#msg: fill end")
	for {
		//head := 0
		if m == nil {
			m = new(Message)
			m.head = make([]byte, 0)
			m.data = make([]byte, 0)
		}
		if buf, err = testBuf(buf, bc); !err {
			return
		}

		fin := (buf[0] & 0x80) >> 7
		op := buf[0] & 0xf
		//rsv := buf[0] & 0x70 >> 4
		m.head = append(m.head, buf[0])
		if op == 0x8 {
			m.op = connClosed
			c.conn.Close()
		}
		buf = buf[1:]

		if buf, err = testBuf(buf, bc); !err {
			return
		}

		mask := (buf[0] & 0x80) >> 7
		payload := buf[0] & 0x7f
		if fin != 0 {
			m.fin = true
		} else {
			m.fin = false
		}
		m.head = append(m.head, buf[0]&0x7f)
		buf = buf[1:]
		if mask == 0 {
			c.conn.Close()
			return
		}

		if payload == 126 {
			t := make([]byte, 0)
			for i := 0; i < 2; i++ {
				if len(buf) > 0 {
					t = append(t, buf[0])
					buf = buf[1:]
				} else {
					if buf, err = testBuf(buf, bc); !err {
						return
					}
				}
			}
			m.head = append(m.head, t...)
			m.length += uint64(binary.BigEndian.Uint16(t))

		} else if payload == 127 {
			t := make([]byte, 0)
			for i := 0; i < 8; i++ {
				if len(buf) > 0 {
					t = append(t, buf[0])
					buf = buf[1:]
				} else {
					if buf, err = testBuf(buf, bc); !err {
						return
					}
				}
			}
			//fmt.Println("#msg:length ", binary.BigEndian.Uint64(t))
			m.length += binary.BigEndian.Uint64(t)
			m.head = append(m.head, t...)
		} else {
			m.length += uint64(payload)

		}

		t := make([]byte, 0)

		for i := 0; i < 4; i++ {
			if len(buf) > 0 {
				t = append(t, buf[0])
				buf = buf[1:]
			} else {
				if buf, err = testBuf(buf, bc); !err {
					return
				}
			}
		}
		m.maskingKey = t
		//fmt.Println("#msg: m.length = %d, buflen = %d,datalen = %d", m.length, len(buf), len(m.data))
		for len(buf) < int(m.length)-len(m.data) {
			m.data = append(m.data, buf...)
			buf = <-bc
			if len(buf) == 0 {
				return
			}
		}
		clen := int(m.length) - len(m.data)
		m.data = append(m.data, buf[:clen]...)
		buf = buf[clen:]
		if m.fin == true {

			for i := 0; i < len(m.data); i++ {
				m.data[i] ^= m.maskingKey[i%4]
			}
			fmt.Println(string(m.data))
			c.mh(c, m)

			m = nil
		}

	}
	return

}

func min(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}
