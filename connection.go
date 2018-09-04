package lagoon

import (
	"net"
	"sync"
)

type Connection struct {
	// safe
	l *Lagoon
	// unsafe
	net.Conn
	disabled bool
	mu       sync.Mutex
}

func (self *Connection) IsValid() bool {
	if self == nil {
		return false
	}
	return true
}
func (self *Lagoon) createConnection(
	conn net.Conn,
) *Connection {
	return &Connection{
		l:    self,
		Conn: conn,
	}
}
func (self *Connection) Disable() {
	self.mu.Lock()
	self.disabled = true
	self.mu.Unlock()
}
func (self *Connection) Close() error {
	self.l.mu.Lock()
	defer self.l.mu.Unlock()
	return self.close()
}
func (self *Connection) close() error {
	var err error
	self.mu.Lock()
	defer self.mu.Unlock()
	// check if connection is in active
	if _, ok := self.l.active[self]; ok {
		// remove from active
		delete(self.l.active, self)
		if self.disabled {
			// close connection
			err = self.Conn.Close()
			// release buffer
			self.l.config.Buffer.release()
		} else {
			// return to available
			self.l.available[self] = struct{}{}
			// DO NOT RELEASE BUFFER!!!
		}
	} else {
		// check if connection is in available
		if _, ok := self.l.available[self]; ok {
			// remove from available
			delete(self.l.available, self)
			// close connection
			err = self.Conn.Close()
			// release buffer
			self.l.config.Buffer.release()
		} else {
			// not found, wtf?
			// close connection
			err = self.Conn.Close()
			// DO NOT RELEASE BUFFER!!!
		}
	}
	// closed
	if err != nil {
		return err
	}
	return nil
}
