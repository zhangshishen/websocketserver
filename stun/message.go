package stun

type Message struct {
	mtype         uint16
	mlength       uint16
	magicCokkie   int
	transactionID [3]int
}

func (m *Message) makeMessage() []byte {
	return nil
}
