package gomemcached

import "errors"

var (
	ErrInvalidArguments = errors.New("Invalid arguments")
	ErrNotConnected     = errors.New("Not connected server")
)
