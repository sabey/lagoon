package lagoon

import (
	"fmt"
	"net"
	"time"
)

const (
	TICKEVERY_MIN     = time.Second * 5
	TICKEVERY_DEFAULT = time.Second * 15
	TICKEVERY_MAX     = time.Minute
)

var (
	ERR_CONFIG_NIL              = fmt.Errorf("Config NIL")
	ERR_DIAL_NIL                = fmt.Errorf("Dial NIL")
	ERR_DIAL_INITIAL            = fmt.Errorf("Dial Initial < 0")
	ERR_DIAL_BUFFER_NIL         = fmt.Errorf("Dial Buffer NIL")
	ERR_DIAL_INITIAL_BUFFER_MAX = fmt.Errorf("Dial Initial More Than Buffer Max")
)

type Config struct {
	Dial        func() (net.Conn, error)
	DialInitial int
	IdleTimeout time.Duration
	TickEvery   time.Duration
	Buffer      *Buffer
}

func (self *Config) IsValid() bool {
	if self == nil {
		return false
	}
	return true
}
func (self *Config) Clone() *Config {
	if self == nil {
		return nil
	}
	config := &Config{
		Dial:        self.Dial,
		DialInitial: self.DialInitial,
		IdleTimeout: self.IdleTimeout,
		TickEvery:   self.TickEvery,
		Buffer:      self.Buffer,
	}
	return config
}
func (self *Config) Validate() error {
	if self == nil {
		return ERR_CONFIG_NIL
	}
	if self.Dial == nil {
		return ERR_DIAL_NIL
	}
	if self.DialInitial < 0 {
		return ERR_DIAL_INITIAL
	}
	if self.Buffer == nil {
		return ERR_DIAL_BUFFER_NIL
	}
	if self.DialInitial > self.Buffer.GetMax() {
		return ERR_DIAL_INITIAL_BUFFER_MAX
	}
	if self.TickEvery == 0 {
		self.TickEvery = TICKEVERY_DEFAULT
	} else if self.TickEvery < TICKEVERY_MIN {
		self.TickEvery = TICKEVERY_MIN
	} else if self.TickEvery > TICKEVERY_MAX {
		self.TickEvery = TICKEVERY_MAX
	}
	return nil
}
