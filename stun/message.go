package stun

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"time"

	"../wlog"
)

//class
const (
	request  = 0
	response = 2
)

//method
const (
	binding = 1
)

//attribute
const (
	MAPPED_ADDRESS     = 1
	XOR_MAPPED_ADDRESS = 0x20
)

type attributeCallback func(h *Attribute)
type messageCallback func(m *Message)

type Attribute struct {
	htype  uint16
	length uint16
	value  []byte
}

type Message struct {
	mtype         uint16
	mlength       uint16
	magicCookie   int
	transactionID [3]int
	attribute     []Attribute

	//mcallback messageCallback
	hcallback attributeCallback
}

func (a *Attribute) print() {
	fmt.Printf("Attribute Type: %d\n", a.htype)
	fmt.Printf("Attribute value: %v\n", a.value)
}

func (m *Message) generateBuffer(reuse []byte) []byte {

	var length uint16 = HEADERSIZE
	var res []byte

	for _, v := range m.attribute {
		length += v.length
	}
	m.mlength = length - HEADERSIZE

	if reuse == nil || len(reuse) < (int)(length) {
		res = make([]byte, length)
	} else {
		res = reuse
	}

	binary.BigEndian.PutUint16(res[0:], m.mtype)
	binary.BigEndian.PutUint16(res[2:], m.mlength)
	binary.BigEndian.PutUint32(res[4:], (uint32)(m.magicCookie))
	binary.BigEndian.PutUint32(res[8:], (uint32)(m.transactionID[0]))
	binary.BigEndian.PutUint32(res[12:], (uint32)(m.transactionID[1]))
	binary.BigEndian.PutUint32(res[16:], (uint32)(m.transactionID[2]))

	attr := res[HEADERSIZE:]

	for _, v := range m.attribute {
		binary.BigEndian.PutUint16(attr[0:], v.htype)
		binary.BigEndian.PutUint16(attr[0:], v.length)

		length := copy(attr[4:], v.value)
		attr = attr[4+length:]
	}

	return res
}

func (m *Message) make(class int, method int) {
	//TODO
	m.mtype = makeType(class, method)

	rand.Seed((int64)(time.Now().UnixNano()))

	m.transactionID[0] = rand.Int()
	m.transactionID[1] = rand.Int()
	m.transactionID[2] = rand.Int()

	m.magicCookie = 0x2112a442

	return
}

func (m *Message) addAttribute(a Attribute) {
	m.attribute = append(m.attribute, a)
}

func makeType(method int, class int) uint16 {

	var res uint16 = 0
	res |= (uint16)((class & 1) << 4)
	res |= (uint16)((class & 2) << 8)
	res |= (uint16)((method & 0xf))
	res |= (uint16)((method & 0x70) << 1)
	res |= (uint16)((method & 0xf80) << 2)

	return res
}

//parse header and return length
func (m *Message) parseMessageHeader(buf []byte) int {
	if m.attribute == nil {
		m.attribute = make([]Attribute, 0)
	}
	if len(buf) != 20 {
		wlog.Out("[stun] parse failed, header too short\n")
		return -1
	}

	if buf[0]&0xC0 != 0 {
		wlog.Out("[stun] parse failed, not stun header\n")
		return -1
	}
	m.mtype = 0
	m.mtype |= ((uint16)(0x3F & buf[0])) << 8
	m.mtype |= ((uint16)(0xFF & buf[1]))

	m.mlength = binary.BigEndian.Uint16(buf[2:4])
	m.magicCookie = (int)(binary.BigEndian.Uint32(buf[4:8]))

	if m.magicCookie != 0x2112a442 {
		wlog.Out("[stun] magic cookie is incorrect!\n")
		return -1
	}

	m.transactionID[0] = (int)(binary.BigEndian.Uint32(buf[8:12]))
	m.transactionID[1] = (int)(binary.BigEndian.Uint32(buf[12:16]))
	m.transactionID[2] = (int)(binary.BigEndian.Uint32(buf[16:20]))

	//m.mcallback(m)
	wlog.Out("[stun] parse header success length is ", m.mlength)
	return int(m.mlength)

}

func (m *Message) parseMessageAttribute(buf []byte) int {
	//m.mtype =
	for {
		//wlog.Out("[stun] Attribute buffer length", len(buf))
		if len(buf) <= 4 {
			return 0
		}

		h := Attribute{}
		h.htype = binary.BigEndian.Uint16(buf[0:2])
		h.length = binary.BigEndian.Uint16(buf[2:4])

		if h.length%4 != 0 {
			wlog.Out("[stun] parse Attribute header failed, boundary error\n")
			return -1
		}

		buf = buf[4:]

		if len(buf) < (int)(h.length) {
			wlog.Out("[stun] parse Attribute header failed, buffer too small\n")
			return -1
		}

		h.value = buf[:h.length]
		m.hcallback(&h)

		buf = buf[h.length:]
	}
}
