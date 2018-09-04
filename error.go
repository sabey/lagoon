package lagoon

type dialError struct {
	error
}

func (self *dialError) Timeout() bool {
	// net.Error
	// Is the error a timeout?
	return true
}
func (self *dialError) Temporary() bool {
	// net.Error
	// Is the error temporary?
	return true
}
