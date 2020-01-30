package gomemcached

type MemcachedClient struct {
	cluster *Cluster
}

func NewMemcachedClient(addrs []string, maxConnPerServer uint32) *MemcachedClient {
	m := &MemcachedClient{}
	m.cluster = createCluster(addrs, maxConnPerServer)
	return m
}

func (m *MemcachedClient) AddServer(addr string, maxConnPerServer uint32) error {
	return m.cluster.AddServer2Cluster(addr, maxConnPerServer)
}

func (m *MemcachedClient) SetServerErrorCallback(call ServerErrorCallback) {
	m.cluster.serverErrCallback = call
}

func (m *MemcachedClient) Exit() {
	m.cluster.exit()
}

func (m *MemcachedClient) exec(key string, cmdFunc func(cmder *Commander) error) error {
	var err error
	server, cmder, err := m.cluster.ChooseServerCommand(key)
	if err != nil {
		return err
	}

	defer func() {
		if err == nil {
			m.cluster.ReleaseServerCommand(server, cmder)
		} else if _, ok := err.(*StatusError); ok {
			m.cluster.ReleaseServerCommand(server, cmder)
		}
	}()

	err = cmdFunc(cmder)
	return err
}

func (m *MemcachedClient) Get(key string, value interface{}) (uint64, error) {
	var modifyCAS uint64
	var resErr error

	err := m.exec(key, func(cmder *Commander) error {
		modifyCAS, resErr = cmder.get(key, value)
		return resErr
	})

	return modifyCAS, err
}

func (m *MemcachedClient) Set(key string, value interface{}, expiration uint32) (uint64, error) {
	var modifyCAS uint64
	var resErr error

	err := m.exec(key, func(cmder *Commander) error {
		modifyCAS, resErr = cmder.store(OPCODE_SET, key, value, expiration, 0, true)
		return resErr
	})

	return modifyCAS, err
}

func (m *MemcachedClient) SetRawData(key string, value []byte, expiration uint32) (uint64, error) {
	var modifyCAS uint64
	var resErr error

	err := m.exec(key, func(cmder *Commander) error {
		modifyCAS, resErr = cmder.store(OPCODE_SET, key, value, expiration, 0, false)
		return resErr
	})

	return modifyCAS, err
}

func (m *MemcachedClient) Add(key string, value interface{}, expiration uint32) (uint64, error) {
	var modifyCAS uint64
	var resErr error

	err := m.exec(key, func(cmder *Commander) error {
		modifyCAS, resErr = cmder.store(OPCODE_ADD, key, value, expiration, 0, true)
		return resErr
	})

	return modifyCAS, err
}

func (m *MemcachedClient) AddRawData(key string, value []byte, expiration uint32) (uint64, error) {
	var modifyCAS uint64
	var resErr error

	err := m.exec(key, func(cmder *Commander) error {
		modifyCAS, resErr = cmder.store(OPCODE_ADD, key, value, expiration, 0, false)
		return resErr
	})

	return modifyCAS, err
}

func (m *MemcachedClient) Replace(key string, value interface{}, expiration uint32, cas uint64) (uint64, error) {
	var modifyCAS uint64
	var resErr error

	err := m.exec(key, func(cmder *Commander) error {
		modifyCAS, resErr = cmder.store(OPCODE_REPLACE, key, value, expiration, cas, true)
		return resErr
	})

	return modifyCAS, err
}

func (m *MemcachedClient) ReplaceRawData(key string, value []byte, expiration uint32, cas uint64) (uint64, error) {
	var modifyCAS uint64
	var resErr error

	err := m.exec(key, func(cmder *Commander) error {
		modifyCAS, resErr = cmder.store(OPCODE_REPLACE, key, value, expiration, cas, false)
		return resErr
	})

	return modifyCAS, err
}

func (m *MemcachedClient) Append(key string, value []byte, cas uint64) (uint64, error) {
	var modifyCAS uint64
	var resErr error

	err := m.exec(key, func(cmder *Commander) error {
		modifyCAS, resErr = cmder.append(OPCODE_APPEND, key, value, cas)
		return resErr
	})

	return modifyCAS, err
}

func (m *MemcachedClient) Prepend(key string, value []byte, cas uint64) (uint64, error) {
	var modifyCAS uint64
	var resErr error

	err := m.exec(key, func(cmder *Commander) error {
		modifyCAS, resErr = cmder.append(OPCODE_PREPEND, key, value, cas)
		return resErr
	})

	return modifyCAS, err
}

func (m *MemcachedClient) Increment(key string, delta uint64, expiration uint32, cas uint64) (uint64, uint64, error) {
	var value uint64
	var modifyCAS uint64
	var resErr error

	err := m.exec(key, func(cmder *Commander) error {
		value, modifyCAS, resErr = cmder.atomic(OPCODE_INCR, key, delta, expiration, cas)
		return resErr
	})

	return value, modifyCAS, err
}

func (m *MemcachedClient) Decrement(key string, delta uint64, expiration uint32, cas uint64) (uint64, uint64, error) {
	var value uint64
	var modifyCAS uint64
	var resErr error

	err := m.exec(key, func(cmder *Commander) error {
		value, modifyCAS, resErr = cmder.atomic(OPCODE_DECR, key, delta, expiration, cas)
		return resErr
	})

	return value, modifyCAS, err
}

func (m *MemcachedClient) TouchAtomicValue(key string) (uint64, error) {
	var value uint64
	var resErr error

	err := m.exec(key, func(cmder *Commander) error {
		value, resErr = cmder.touchAtomicValue(key)
		return resErr
	})

	return value, err
}
