package stun

func callback(h *Attribute) {
	h.print()
}
func Stun_test() {
	agent := Agent{}
	agent.bind("stun.l.google.com", 19302)
	agent.request(binding, request)
	agent.listen(callback)
}
