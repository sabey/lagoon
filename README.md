# Lagoon
### Golang net.Conn Connection Pool
[![GoDoc](https://godoc.org/sabey.co/lagoon?status.svg)](https://godoc.org/sabey.co/lagoon)

This package was inspired by [github.com/fatih/pool](https://github.com/fatih/pool) - **This is NOT a fork, it will function differently!!!**

The main differences are:
* The connection pool is capped and Dial will block once the pool is full!
* Connections are capped with a Buffer that can be optionally shared between pools.
* Dial can Timeout if we fail to acquire a spot from the Buffer queue! (DialTimeout or similar should be used within the Dial function)
* DialInitial can not garauntee that we will always dial the initial amount due to the possibility of a shared Buffer.

## Install
```bash
go get -t -u sabey.co/lagoon
```

## Usage
```golang
// create a shared buffer
buffer := CreateBuffer(10, time.Second*2)

// create a config for our lagoon instance
config := &Config{
	Dial: func() (net.Conn, error) {
		return net.DialTimeout("tcp", "service.local:25", time.Second*30)
	},
	DialInitial: 5,
	Buffer:      buffer,
}

// create a lagoon instance
l, err := CreateLagoon(config)

// dial
c, err := l.Dial()

// return connection back to the pool
c.Close()

// dial
c, err := l.Dial()

// remove connection from the pool on close
c.(*Connection).Disable()

// close connection
c.Close()

```
