package gomemcached

type MemcachedClient struct {
	Cluster *Cluster
}

func NewMemcachedClient(addrs []string, maxConnPerServer uint32) *MemcachedClient {
	m := &MemcachedClient{}
	m.Cluster = createCluster(addrs, maxConnPerServer)
	return m
}

func (m *MemcachedClient) SetServerErrorCallback(call ServerErrorCallback) {
	m.Cluster.serverErrCallback = call
}

func (m *MemcachedClient) Exit() {
	m.Cluster.exit()
}

func (m *MemcachedClient) Get(key string, value interface{}) (uint64, error) {
	server, cmder, err := m.Cluster.ChooseServerCommand(key)
	if err != nil {
		return 0, err
	}

	cas, err := cmder.get(key, value)
	if err == nil {
		m.Cluster.ReleaseServerCommand(server, cmder)
	} else if _, ok := err.(*StatusError); ok {
		m.Cluster.ReleaseServerCommand(server, cmder)
	}
	return cas, err
}

func (m *MemcachedClient) Set(key string, value interface{}, expiration uint32) (uint64, error) {
	server, cmder, err := m.Cluster.ChooseServerCommand(key)
	if err != nil {
		return 0, err
	}

	modifyCAS, err := cmder.store(OPCODE_SET, key, value, expiration, 0)
	if err == nil {
		m.Cluster.ReleaseServerCommand(server, cmder)
	} else if _, ok := err.(*StatusError); ok {
		m.Cluster.ReleaseServerCommand(server, cmder)
	}
	return modifyCAS, err
}

func (m *MemcachedClient) Add(key string, value interface{}, expiration uint32) (uint64, error) {
	server, cmder, err := m.Cluster.ChooseServerCommand(key)
	if err != nil {
		return 0, err
	}

	modifyCAS, err := cmder.store(OPCODE_ADD, key, value, expiration, 0)
	if err == nil {
		m.Cluster.ReleaseServerCommand(server, cmder)
	} else if _, ok := err.(*StatusError); ok {
		m.Cluster.ReleaseServerCommand(server, cmder)
	}
	return modifyCAS, err
}

func (m *MemcachedClient) Replace(key string, value interface{}, expiration uint32, cas uint64) (uint64, error) {
	server, cmder, err := m.Cluster.ChooseServerCommand(key)
	if err != nil {
		return 0, err
	}

	modifyCAS, err := cmder.store(OPCODE_REPLACE, key, value, expiration, cas)
	if err == nil {
		m.Cluster.ReleaseServerCommand(server, cmder)
	} else if _, ok := err.(*StatusError); ok {
		m.Cluster.ReleaseServerCommand(server, cmder)
	}
	return modifyCAS, err
}

func (m *MemcachedClient) Append(key string, value interface{}, cas uint64) (uint64, error) {
	server, cmder, err := m.Cluster.ChooseServerCommand(key)
	if err != nil {
		return 0, err
	}

	modifyCAS, err := cmder.append(OPCODE_APPEND, key, value, cas)
	if err == nil {
		m.Cluster.ReleaseServerCommand(server, cmder)
	} else if _, ok := err.(*StatusError); ok {
		m.Cluster.ReleaseServerCommand(server, cmder)
	}
	return modifyCAS, err
}

func (m *MemcachedClient) Prepend(key string, value interface{}, cas uint64) (uint64, error) {
	server, cmder, err := m.Cluster.ChooseServerCommand(key)
	if err != nil {
		return 0, err
	}

	modifyCAS, err := cmder.append(OPCODE_PREPEND, key, value, cas)
	if err == nil {
		m.Cluster.ReleaseServerCommand(server, cmder)
	} else if _, ok := err.(*StatusError); ok {
		m.Cluster.ReleaseServerCommand(server, cmder)
	}
	return modifyCAS, err
}

func (m *MemcachedClient) Increment(key string, delta uint64, expiration uint32, cas uint64) (uint64, uint64, error) {
	server, cmder, err := m.Cluster.ChooseServerCommand(key)
	if err != nil {
		return 0, 0, err
	}

	value, modifyCAS, err := cmder.atomic(OPCODE_INCR, key, delta, expiration, cas)
	if err == nil {
		m.Cluster.ReleaseServerCommand(server, cmder)
	} else if _, ok := err.(*StatusError); ok {
		m.Cluster.ReleaseServerCommand(server, cmder)
	}
	return value, modifyCAS, err
}

func (m *MemcachedClient) Decrement(key string, delta uint64, expiration uint32, cas uint64) (uint64, uint64, error) {
	server, cmder, err := m.Cluster.ChooseServerCommand(key)
	if err != nil {
		return 0, 0, err
	}

	value, cas, err := cmder.atomic(OPCODE_DECR, key, delta, expiration, cas)
	if err == nil {
		m.Cluster.ReleaseServerCommand(server, cmder)
	} else if _, ok := err.(*StatusError); ok {
		m.Cluster.ReleaseServerCommand(server, cmder)
	}
	return value, cas, err
}

func (m *MemcachedClient) TouchAtomicValue(key string) (uint64, error) {
	server, cmder, err := m.Cluster.ChooseServerCommand(key)
	if err != nil {
		return 0, err
	}

	value, err := cmder.touchAtomicValue(key)
	if err == nil {
		m.Cluster.ReleaseServerCommand(server, cmder)
	} else if _, ok := err.(*StatusError); ok {
		m.Cluster.ReleaseServerCommand(server, cmder)
	}
	return value, err
}
