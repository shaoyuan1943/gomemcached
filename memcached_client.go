package gomemcached

type MemcachedClient struct {
	cluster *cluster
}

func NewMemcachedClient(addrs []string, maxConnPerServer uint32) *MemcachedClient {
	m := &MemcachedClient{}
	m.cluster = createCluster(addrs, maxConnPerServer)
	return m
}

func (m *MemcachedClient) Exit() {
	m.cluster.exit()
}

func (m *MemcachedClient) Get(key string, value interface{}) (uint64, error) {
	server, cmder, err := m.cluster.ChooseServerCommand(key)
	if err != nil {
		return 0, err
	}

	cas, err := cmder.get(key, value)
	if err == nil {
		m.cluster.ReleaseServerCommand(server, cmder)
	} else if _, ok := err.(*StatusError); ok {
		m.cluster.ReleaseServerCommand(server, cmder)
	}
	return cas, err
}

func (m *MemcachedClient) Set(key string, value interface{}, expiration uint32) (uint64, error) {
	server, cmder, err := m.cluster.ChooseServerCommand(key)
	if err != nil {
		return 0, err
	}

	modifyCAS, err := cmder.store(OPCODE_SET, key, value, expiration, 0)
	if err == nil {
		m.cluster.ReleaseServerCommand(server, cmder)
	} else if _, ok := err.(*StatusError); ok {
		m.cluster.ReleaseServerCommand(server, cmder)
	}
	return modifyCAS, err
}

func (m *MemcachedClient) Add(key string, value interface{}, expiration uint32) (uint64, error) {
	server, cmder, err := m.cluster.ChooseServerCommand(key)
	if err != nil {
		return 0, err
	}

	modifyCAS, err := cmder.store(OPCODE_ADD, key, value, expiration, 0)
	if err == nil {
		m.cluster.ReleaseServerCommand(server, cmder)
	} else if _, ok := err.(*StatusError); ok {
		m.cluster.ReleaseServerCommand(server, cmder)
	}
	return modifyCAS, err
}

func (m *MemcachedClient) Replace(key string, value interface{}, expiration uint32, cas uint64) (uint64, error) {
	server, cmder, err := m.cluster.ChooseServerCommand(key)
	if err != nil {
		return 0, err
	}

	modifyCAS, err := cmder.store(OPCODE_REPLACE, key, value, expiration, cas)
	if err == nil {
		m.cluster.ReleaseServerCommand(server, cmder)
	} else if _, ok := err.(*StatusError); ok {
		m.cluster.ReleaseServerCommand(server, cmder)
	}
	return modifyCAS, err
}

func (m *MemcachedClient) Append(key string, value interface{}, cas uint64) (uint64, error) {
	server, cmder, err := m.cluster.ChooseServerCommand(key)
	if err != nil {
		return 0, err
	}

	modifyCAS, err := cmder.append(OPCODE_APPEND, key, value, cas)
	if err == nil {
		m.cluster.ReleaseServerCommand(server, cmder)
	} else if _, ok := err.(*StatusError); ok {
		m.cluster.ReleaseServerCommand(server, cmder)
	}
	return modifyCAS, err
}

func (m *MemcachedClient) Prepend(key string, value interface{}, cas uint64) (uint64, error) {
	server, cmder, err := m.cluster.ChooseServerCommand(key)
	if err != nil {
		return 0, err
	}

	modifyCAS, err := cmder.append(OPCODE_PREPEND, key, value, cas)
	if err == nil {
		m.cluster.ReleaseServerCommand(server, cmder)
	} else if _, ok := err.(*StatusError); ok {
		m.cluster.ReleaseServerCommand(server, cmder)
	}
	return modifyCAS, err
}

func (m *MemcachedClient) Increment(key string, delta uint64, expiration uint32, cas uint64) (uint64, uint64, error) {
	server, cmder, err := m.cluster.ChooseServerCommand(key)
	if err != nil {
		return 0, 0, err
	}

	value, cas, err := cmder.atomic(OPCODE_INCR, key, delta, expiration, cas)
	if err == nil {
		m.cluster.ReleaseServerCommand(server, cmder)
	} else if _, ok := err.(*StatusError); ok {
		m.cluster.ReleaseServerCommand(server, cmder)
	}
	return value, cas, err
}

func (m *MemcachedClient) Decrement(key string, delta uint64, expiration uint32, cas uint64) (uint64, uint64, error) {
	server, cmder, err := m.cluster.ChooseServerCommand(key)
	if err != nil {
		return 0, 0, err
	}

	value, cas, err := cmder.atomic(OPCODE_DECR, key, delta, expiration, cas)
	if err == nil {
		m.cluster.ReleaseServerCommand(server, cmder)
	} else if _, ok := err.(*StatusError); ok {
		m.cluster.ReleaseServerCommand(server, cmder)
	}
	return value, cas, err
}

func (m *MemcachedClient) TouchAtomicValue(key string) (uint64, error) {
	server, cmder, err := m.cluster.ChooseServerCommand(key)
	if err != nil {
		return 0, err
	}

	value, err := cmder.touchAtomicValue(key)
	if err == nil {
		m.cluster.ReleaseServerCommand(server, cmder)
	} else if _, ok := err.(*StatusError); ok {
		m.cluster.ReleaseServerCommand(server, cmder)
	}
	return value, err
}
