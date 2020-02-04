package gomemcached

type ServerErrorCallback func(addr string)

type KeyArgs struct {
	Key        string
	Value      interface{}
	Expiration uint32
	CAS        uint64
	Delta      uint64

	useMsgpack bool
}

type Client interface {
	// Add a memcached server.
	AddServer(addr string, maxConnPerServer uint32) error

	// Set callback when memcached server failed.
	// The callback's parameter is server address.
	SetServerErrorCallback(errCall ServerErrorCallback)

	// Exit client by manual control.
	// In theory that client will not be available after this function is called.
	Exit()

	// Get the value of key.
	// `value` is a pointer to a value variable.
	// Return value is the CAS corresponding to the key,
	// the error is nil when the operation is successful.
	Get(key string, value interface{}) (uint64, error)

	// Set the value of key.
	// Return value is the CAS corresponding to the key,
	// the error is nil when operation is successful.
	Set(args *KeyArgs) (uint64, error)

	// Same as `Set`.
	// When increase or decrease part of the data, must use this function.
	// Return value is CAS,
	// the error is nil when operation is successful, the function does not serialize data.
	SetRawData(args *KeyArgs) (uint64, error)

	// Add the value of key.
	// Return value is CAS, the error is nil when operation is successful.
	Add(args *KeyArgs) (uint64, error)

	// Same as `Add`.
	// When increase or decrease part of the data, must use this function.
	AddRawData(args *KeyArgs) (uint64, error)

	// Replace the value of key.
	// Return value is CAS, the error is nil when operation is successful.
	Replace(args *KeyArgs) (uint64, error)

	// Same as `Replace`
	// When increase or decrease part of the data, must use this function.
	ReplaceRawData(args *KeyArgs) (uint64, error)

	// Appends data to the tail/head of an existing value.
	// Return value is the CAS, and the error is nil when the operation is successful.
	// This function does not serialize data
	Append(args *KeyArgs) (uint64, error)
	Prepend(args *KeyArgs) (uint64, error)

	// Atomic operation, the delta of the existing value is increased/decreased.
	// If the key does not exist, the operation returns the initial value of the key.
	// The error is nil when the operation is successful
	Increment(args *KeyArgs) (uint64, uint64, error)
	Decrement(args *KeyArgs) (uint64, uint64, error)

	// Returns the current value of an atom.
	// The error is nil when the operation is successful
	TouchAtomicValue(key string) (uint64, error)
}
