package main

import (
	"fmt"
)

// A BoundedBuffer is a fixed-capacity buffer of bytes.
type BoundedBuffer struct {
	writeSlice []byte
	readSlice  []byte
}

func NewBoundedBuffer(cap int) *BoundedBuffer {
	buf := make([]byte, 0, cap)
	return &BoundedBuffer{
		writeSlice: buf[:cap],
		readSlice:  buf,
	}
}

func (self *BoundedBuffer) Write(p []byte) (int, error) {
	n := copy(self.writeSlice, p)
	self.writeSlice = self.writeSlice[n:]
	self.readSlice = self.readSlice[:len(self.readSlice)+n]

	var err error
	if n < len(p) {
		err = fmt.Errorf("Buffer is full")
	}
	return n, err
}

func (self *BoundedBuffer) String() string {
	return string(self.readSlice)
}
