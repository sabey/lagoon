package lagoon

import (
	"time"
)

func (self *Lagoon) toggleTick() {
	if self.config.IdleTimeout < 1 {
		// we don't tick
		return
	}
	if len(self.available) == 0 {
		self.ticker_stop = true
	} else {
		// idle connections exist
		self.ticker_stop = false
		if self.ticker_running.IsZero() {
			// ticker is not running, start ticker
			self.ticker_running = time.Now()
			go self.tick()
		}
	}
}
func (self *Lagoon) tick() {
	defer func() {
		// close our ticker
		self.mu.Lock()
		self.ticker_stop = false
		self.ticker_running = time.Time{}
		self.mu.Unlock()
	}()
	for {
		// tick
		self.mu.Lock()
		if self.ticker_stop {
			self.mu.Unlock()
			// ticker is stopped
			return
		}
		// check idle
		now := time.Now()
		for c, _ := range self.available {
			c.mu.Lock()
			if now.After(c.idle.Add(self.config.IdleTimeout)) {
				// timedout - mark as disabled
				c.disabled = true
				// remove from lagoon
				c.remove()
			}
			c.mu.Unlock()
		}
		self.mu.Unlock()
		// sleep
		<-time.After(self.config.TickEvery)
	}
}
