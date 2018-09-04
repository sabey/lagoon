package lagoon

import (
	"log"
	"time"
)

type Buffer struct {
	// safe
	buffer  chan struct{}
	timeout time.Duration
}

func CreateBuffer(
	max int,
	timeout time.Duration,
) *Buffer {
	if max <= 0 {
		log.Fatalf("./lagoon.CreateBuffer(): max < 1: %d", max)
		return nil
	}
	if timeout <= 0 {
		log.Fatalf("./lagoon.CreateBuffer(): timeout < 1: %d", timeout)
		return nil
	}
	return &Buffer{
		buffer:  make(chan struct{}, max),
		timeout: timeout,
	}
}
func (self *Buffer) GetMax() int {
	return cap(self.buffer)
}
func (self *Buffer) GetTimeout() time.Duration {
	return self.timeout
}
func (self *Buffer) acquire() bool {
	// acquire from the buffer
	select {
	// this will block when the buffer becomes full
	case self.buffer <- struct{}{}:
		// successfully acquired!
		// BUFFER MUST BE RELEASED
		return true
	case <-time.After(self.timeout):
		// failed to acquire
		// buffer will not have to be released!!!
		return false
	}
}
func (self *Buffer) release() {
	// release from the buffer
	<-self.buffer
}
