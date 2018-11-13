package main

type parser struct {
	frame      []int
	byteStream chan []byte
}
