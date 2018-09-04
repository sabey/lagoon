package lagoon

import (
	"fmt"
	"log"
	"net"
	"sabey.co/unittest"
	"testing"
	"time"
)

func TestLagoon(t *testing.T) {
	log.Println("TestLagoon")

	go func() {
		<-time.After(time.Second * 15)
		log.Fatalln("unitests failed")
	}()

	buffer := CreateBuffer(10, time.Second*2)
	unittest.NotNil(t, buffer)

	config := &Config{
		Dial: func() (net.Conn, error) {
			return &fakeConnection{}, nil
		},
		DialInitial: 5,
		Buffer:      buffer,
	}

	l, err := CreateLagoon(config)
	unittest.IsNil(t, err)
	unittest.NotNil(t, l)

	// dial
	fmt.Println("dial")
	c, err := l.Dial()
	unittest.IsNil(t, err)
	unittest.Equals(t, len(l.available), config.DialInitial-1)
	unittest.Equals(t, l.ConnectionsAvailable(), config.DialInitial-1)
	unittest.Equals(t, len(l.active), 1)
	unittest.Equals(t, l.ConnectionsActive(), 1)
	unittest.Equals(t, len(l.available)+len(l.active), config.DialInitial)
	unittest.Equals(t, l.Connections(), config.DialInitial)

	// dial max
	fmt.Println("dial max")
	for i := 0; i < config.Buffer.GetMax()-1; i++ {
		_, err = l.Dial()
		unittest.IsNil(t, err)
	}
	unittest.Equals(t, len(l.available), 0)
	unittest.Equals(t, l.ConnectionsAvailable(), 0)
	unittest.Equals(t, len(l.active), config.Buffer.GetMax())
	unittest.Equals(t, l.ConnectionsActive(), config.Buffer.GetMax())
	unittest.Equals(t, len(l.available)+len(l.active), config.Buffer.GetMax())
	unittest.Equals(t, l.Connections(), config.Buffer.GetMax())

	// dial
	fmt.Println("dial timeout")
	_, err = l.Dial()
	unittest.NotNil(t, err)
	e, ok := err.(*dialError)
	unittest.Equals(t, ok, true)
	unittest.Equals(t, e.error, ERR_TIMEDOUT)
	unittest.Equals(t, e.Temporary(), true)
	unittest.Equals(t, e.Timeout(), true)
	unittest.Equals(t, len(l.available), 0)
	unittest.Equals(t, l.ConnectionsAvailable(), 0)
	unittest.Equals(t, len(l.active), config.Buffer.GetMax())
	unittest.Equals(t, l.ConnectionsActive(), config.Buffer.GetMax())
	unittest.Equals(t, len(l.available)+len(l.active), config.Buffer.GetMax())
	unittest.Equals(t, l.Connections(), config.Buffer.GetMax())

	// close connection
	fmt.Println("close connection")
	unittest.IsNil(t, c.Close())
	unittest.Equals(t, len(l.available), 1)
	unittest.Equals(t, l.ConnectionsAvailable(), 1)
	unittest.Equals(t, len(l.active), config.Buffer.GetMax()-1)
	unittest.Equals(t, l.ConnectionsActive(), config.Buffer.GetMax()-1)
	unittest.Equals(t, len(l.available)+len(l.active), config.Buffer.GetMax())
	unittest.Equals(t, l.Connections(), config.Buffer.GetMax())

	// dial
	fmt.Println("dial")
	_, err = l.Dial()
	unittest.IsNil(t, err)
	unittest.Equals(t, len(l.available), 0)
	unittest.Equals(t, l.ConnectionsAvailable(), 0)
	unittest.Equals(t, len(l.active), config.Buffer.GetMax())
	unittest.Equals(t, l.ConnectionsActive(), config.Buffer.GetMax())
	unittest.Equals(t, len(l.available)+len(l.active), config.Buffer.GetMax())
	unittest.Equals(t, l.Connections(), config.Buffer.GetMax())

	// close
	fmt.Println("l.Close")
	l.Close()
	fmt.Println("closed")
	unittest.Equals(t, len(l.available), 0)
	unittest.Equals(t, l.ConnectionsAvailable(), 0)
	unittest.Equals(t, len(l.active), 0)
	unittest.Equals(t, l.ConnectionsActive(), 0)
	unittest.Equals(t, len(l.available)+len(l.active), 0)
	unittest.Equals(t, l.Connections(), 0)

	// dial available
	fmt.Println("DialInitialize")
	l.DialInitialize()
	unittest.Equals(t, len(l.available), 1)
	unittest.Equals(t, l.ConnectionsAvailable(), 1)
	unittest.Equals(t, len(l.active), 0)
	unittest.Equals(t, l.ConnectionsActive(), 0)
	unittest.Equals(t, len(l.available)+len(l.active), 1)
	unittest.Equals(t, l.Connections(), 1)

	// close active
	fmt.Println("l.CloseActive")
	l.CloseActive()
	unittest.Equals(t, len(l.available), 1)
	unittest.Equals(t, l.ConnectionsAvailable(), 1)
	unittest.Equals(t, len(l.active), 0)
	unittest.Equals(t, l.ConnectionsActive(), 0)
	unittest.Equals(t, len(l.available)+len(l.active), 1)
	unittest.Equals(t, l.Connections(), 1)

	// close available
	fmt.Println("l.CloseAvailable")
	l.CloseAvailable()
	unittest.Equals(t, len(l.available), 0)
	unittest.Equals(t, l.ConnectionsAvailable(), 0)
	unittest.Equals(t, len(l.active), 0)
	unittest.Equals(t, l.ConnectionsActive(), 0)
	unittest.Equals(t, len(l.available)+len(l.active), 0)
	unittest.Equals(t, l.Connections(), 0)

	// dial max
	fmt.Println("dial max")
	for i := 0; i < config.Buffer.GetMax(); i++ {
		_, err = l.Dial()
		unittest.IsNil(t, err)
	}
	unittest.Equals(t, len(l.available), 0)
	unittest.Equals(t, l.ConnectionsAvailable(), 0)
	unittest.Equals(t, len(l.active), config.Buffer.GetMax())
	unittest.Equals(t, l.ConnectionsActive(), config.Buffer.GetMax())
	unittest.Equals(t, len(l.available)+len(l.active), config.Buffer.GetMax())
	unittest.Equals(t, l.Connections(), config.Buffer.GetMax())

	// disable + close
	fmt.Println("disable + active c.Close")
	for c, _ := range l.active {
		c.Disable()
		c.Close()
	}
	unittest.Equals(t, len(l.available), 0)
	unittest.Equals(t, l.ConnectionsAvailable(), 0)
	unittest.Equals(t, len(l.active), 0)
	unittest.Equals(t, l.ConnectionsActive(), 0)
	unittest.Equals(t, len(l.available)+len(l.active), 0)
	unittest.Equals(t, l.Connections(), 0)

	// dial max
	fmt.Println("dial max")
	for i := 0; i < config.Buffer.GetMax(); i++ {
		_, err = l.Dial()
		unittest.IsNil(t, err)
	}
	unittest.Equals(t, len(l.available), 0)
	unittest.Equals(t, l.ConnectionsAvailable(), 0)
	unittest.Equals(t, len(l.active), config.Buffer.GetMax())
	unittest.Equals(t, l.ConnectionsActive(), config.Buffer.GetMax())
	unittest.Equals(t, len(l.available)+len(l.active), config.Buffer.GetMax())
	unittest.Equals(t, l.Connections(), config.Buffer.GetMax())

	// close
	fmt.Println("active c.Close")
	for c, _ := range l.active {
		c.Close()
	}
	unittest.Equals(t, len(l.available), config.Buffer.GetMax())
	unittest.Equals(t, l.ConnectionsAvailable(), config.Buffer.GetMax())
	unittest.Equals(t, len(l.active), 0)
	unittest.Equals(t, l.ConnectionsActive(), 0)
	unittest.Equals(t, len(l.available)+len(l.active), config.Buffer.GetMax())
	unittest.Equals(t, l.Connections(), config.Buffer.GetMax())

	fmt.Println("l.Close")
	l.Close()

	// DialInitialize max
	fmt.Println("DialInitialize max")
	for i := 0; i < config.Buffer.GetMax(); i++ {
		err = l.DialInitialize()
		unittest.IsNil(t, err)
	}
	unittest.Equals(t, len(l.available), config.Buffer.GetMax())
	unittest.Equals(t, l.ConnectionsAvailable(), config.Buffer.GetMax())
	unittest.Equals(t, len(l.active), 0)
	unittest.Equals(t, l.ConnectionsActive(), 0)
	unittest.Equals(t, len(l.available)+len(l.active), config.Buffer.GetMax())
	unittest.Equals(t, l.Connections(), config.Buffer.GetMax())

	// disable + close
	fmt.Println("disable + available c.Close")
	for c, _ := range l.available {
		c.Close()
	}
	unittest.Equals(t, len(l.available), 0)
	unittest.Equals(t, l.ConnectionsAvailable(), 0)
	unittest.Equals(t, len(l.active), 0)
	unittest.Equals(t, l.ConnectionsActive(), 0)
	unittest.Equals(t, len(l.available)+len(l.active), 0)
	unittest.Equals(t, l.Connections(), 0)

	// DialInitialize max
	fmt.Println("DialInitialize max")
	for i := 0; i < config.Buffer.GetMax(); i++ {
		err = l.DialInitialize()
		unittest.IsNil(t, err)
	}
	unittest.Equals(t, len(l.available), config.Buffer.GetMax())
	unittest.Equals(t, l.ConnectionsAvailable(), config.Buffer.GetMax())
	unittest.Equals(t, len(l.active), 0)
	unittest.Equals(t, l.ConnectionsActive(), 0)
	unittest.Equals(t, len(l.available)+len(l.active), config.Buffer.GetMax())
	unittest.Equals(t, l.Connections(), config.Buffer.GetMax())

	// close
	fmt.Println("available c.Close")
	for c, _ := range l.available {
		c.Disable()
		c.Close()
	}
	unittest.Equals(t, len(l.available), 0)
	unittest.Equals(t, l.ConnectionsAvailable(), 0)
	unittest.Equals(t, len(l.active), 0)
	unittest.Equals(t, l.ConnectionsActive(), 0)
	unittest.Equals(t, len(l.available)+len(l.active), 0)
	unittest.Equals(t, l.Connections(), 0)
}

type fakeConnection struct{}

func (self *fakeConnection) Read(b []byte) (n int, err error) {
	return 0, nil
}
func (self *fakeConnection) Write(b []byte) (n int, err error) {
	return 0, nil
}
func (self *fakeConnection) Close() error {
	return nil
}
func (self *fakeConnection) LocalAddr() net.Addr {
	return nil
}
func (self *fakeConnection) RemoteAddr() net.Addr {
	return nil
}
func (self *fakeConnection) SetDeadline(t time.Time) error {
	return nil
}
func (self *fakeConnection) SetReadDeadline(t time.Time) error {
	return nil
}
func (self *fakeConnection) SetWriteDeadline(t time.Time) error {
	return nil
}
