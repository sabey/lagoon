package lagoon

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

var (
	ERR_TIMEDOUT   = fmt.Errorf("Timed-out")
	ERR_DIAL_EMPTY = fmt.Errorf("Dial Available Was Empty!!!")
)

type Lagoon struct {
	// safe
	config *Config
	// unsafe
	available      map[*Connection]struct{}
	active         map[*Connection]struct{}
	ticker_running time.Time
	ticker_stop    bool
	mu             sync.RWMutex
}

func CreateLagoon(
	config *Config,
) (
	*Lagoon,
	error,
) {
	// dereference
	config = config.Clone()
	if !config.IsValid() {
		return nil, ERR_CONFIG_NIL
	}
	if err := config.Validate(); err != nil {
		return nil, err
	}
	l := &Lagoon{
		config:    config,
		available: make(map[*Connection]struct{}),
		active:    make(map[*Connection]struct{}),
	}
	if config.DialInitial > 0 {
		// if there's an initial amount of connections we will attempt to create them
		// there is no guarantee that we can create an initial amount of connections since we allow shared buffers
		var wg sync.WaitGroup
		wg.Add(config.DialInitial)
		for i := 0; i < config.DialInitial; i++ {
			go func() {
				defer wg.Done()
				// we're going to discard any dial errors
				l.DialInitialize()
			}()
		}
		wg.Wait()
	}
	return l, nil
}
func (self *Lagoon) IsValid() bool {
	if self == nil {
		return false
	}
	return true
}
func (self *Lagoon) dial() error {
	// acquire
	if !self.config.Buffer.acquire() {
		// failed to acquire
		return &dialError{ERR_TIMEDOUT}
	}
	conn, err := self.config.Dial()
	if err != nil {
		// failed to dial - release
		self.config.Buffer.release()
		return err
	}
	// dialed
	// wrap connection and store in available
	self.available[self.createConnection(conn)] = struct{}{}
	// toggle tick
	self.toggleTick()
	return nil
}
func (self *Lagoon) DialInitialize() error {
	// dial initialize will allow us to allocate a new connection
	// if successful, connection will be moved to the available connections
	self.mu.Lock()
	err := self.dial()
	self.mu.Unlock()
	if err != nil {
		return err
	}
	return nil
}
func (self *Lagoon) Dial() (
	net.Conn,
	error,
) {
	// get connection
	self.mu.Lock()
	defer self.mu.Unlock()
	if len(self.available) == 0 {
		// dial new connection
		if err := self.dial(); err != nil {
			// failed to dial
			return nil, err
		}
	}
	// acquired something
	for conn, _ := range self.available {
		// take the first result
		// remove from available
		delete(self.available, conn)
		// store in active
		conn.idle = time.Time{}
		self.active[conn] = struct{}{}
		// toggle tick
		self.toggleTick()
		return conn, nil
	}
	log.Panicln("./lagoon.Dial(): available container was empty, wtf???")
	return nil, ERR_DIAL_EMPTY
}
func (self *Lagoon) Connections() int {
	self.mu.RLock()
	defer self.mu.RUnlock()
	return len(self.available) + len(self.active)
}
func (self *Lagoon) ConnectionsAvailable() int {
	self.mu.RLock()
	defer self.mu.RUnlock()
	return len(self.available)
}
func (self *Lagoon) ConnectionsActive() int {
	self.mu.RLock()
	defer self.mu.RUnlock()
	return len(self.active)
}
func (self *Lagoon) Close() {
	// pool will remain usable even once closed!
	// we will only CLOSE and REMOVE all connections!
	self.mu.Lock()
	self.closeAvailable()
	self.closeActive()
	self.mu.Unlock()
}
func (self *Lagoon) CloseAvailable() {
	// pool will remain usable even once closed!
	// we will only CLOSE and REMOVE all available connections!
	self.mu.Lock()
	self.closeAvailable()
	self.mu.Unlock()
}
func (self *Lagoon) closeAvailable() {
	// close available
	for c, _ := range self.available {
		c.Disable()
		c.close()
	}
	// clean containers
	self.available = make(map[*Connection]struct{})
}
func (self *Lagoon) CloseActive() {
	// pool will remain usable even once closed!
	// we will only CLOSE and REMOVE all active connections!
	self.mu.Lock()
	self.closeActive()
	self.mu.Unlock()
}
func (self *Lagoon) closeActive() {
	// close active
	for c, _ := range self.active {
		c.Disable()
		c.close()
	}
	// clean containers
	self.active = make(map[*Connection]struct{})
}
