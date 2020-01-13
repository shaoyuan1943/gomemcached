package gomemcached

import "errors"

var (
	ErrInvalidArguments        = errors.New("Invalid arguments")
	ErrNotConnected            = errors.New("Not connected server")
	ErrFillRequestHeaderFailed = errors.New("Fill request header failed")

	ErrKeyNotFound = errors.New("Key not found")
	ErrKeyExists   = errors.New("Key exists")
)
