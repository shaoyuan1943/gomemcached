package gomemcached

import "errors"

var (
	ErrInvalidArguments        = errors.New("Invalid arguments")
	ErrNotConnected            = errors.New("Not connected server")
	ErrConnError               = errors.New("Connection error")
	ErrFillRequestHeaderFailed = errors.New("Fill request header failed")
	ErrNotFoundServerNode      = errors.New("Not found server node")
	ErrNoUsableConnection      = errors.New("No usable connection")
	ErrBadConnection           = errors.New("Bad connection")
	// memcached status
	ErrKeyNotFound      = NewStatusError(errors.New("Key not found"))
	ErrKeyExists        = NewStatusError(errors.New("Key exists"))
	ErrValueTooLarge    = NewStatusError(errors.New("Value too large"))
	ErrItemNotStored    = NewStatusError(errors.New("Item not stored"))
	ErrNoNumericValue   = NewStatusError(errors.New("Incr/Decr on non-numeric value"))
	ErrVbucketNotFound  = NewStatusError(errors.New("The vbucket belongs to another server"))
	ErrAuthFailed       = NewStatusError(errors.New("Authentication error"))
	ErrAuthContinue     = NewStatusError(errors.New("Authentication continue"))
	ErrUnknownCommand   = NewStatusError(errors.New("Unknown command"))
	ErrOutOfMemory      = NewStatusError(errors.New("Out of memory"))
	ErrNotSupported     = NewStatusError(errors.New("Not supported"))
	ErrInternalError    = NewStatusError(errors.New("Internal error"))
	ErrBusy             = NewStatusError(errors.New("Internal error"))
	ErrTemporaryFailure = NewStatusError(errors.New("Temporary failure"))
)

type StatusError struct {
	Err error
}

func (s *StatusError) Error() string {
	return s.Err.Error()
}

func (s *StatusError) Is(err error) bool {
	e, ok := err.(*StatusError)
	if !ok {
		return false
	}

	return e == s.Err
}

func (s *StatusError) Unwrap() error {
	return s.Err
}

func NewStatusError(err error) *StatusError {
	return &StatusError{Err: err}
}

func checkStatus(status uint16) error {
	switch status {
	case STATUS_AUTH_CONTINUE:
		return ErrAuthContinue
	case STATUS_AUTH_ERROR:
		return ErrAuthFailed
	case STATUS_BUSY:
		return ErrBusy
	case STATUS_INTERNAL_ERROR:
		return ErrInternalError
	case STATUS_INVALID_ARGS:
		return ErrInvalidArguments
	case STATUS_KEY_EXISTS:
		return ErrKeyExists
	case STATUS_KEY_NOT_FOUND:
		return ErrKeyNotFound
	case STATUS_NON_NUMERIC_VALUE:
		return ErrNoNumericValue
	case STATUS_NOT_STORED:
		return ErrItemNotStored
	case STATUS_NOT_SUPPORTED:
		return ErrNotSupported
	case STATUS_OUT_OF_MEMORY:
		return ErrOutOfMemory
	case STATUS_TEMPORARY_FAILURE:
		return ErrTemporaryFailure
	case STATUS_VALUE_TOO_LARGE:
		return ErrValueTooLarge
	case STATUS_UNKNOWN_COMMAND:
		return ErrUnknownCommand
	case STATUS_VBUCKET_NOT_FOUND:
		return ErrVbucketNotFound
	}

	return nil
}
