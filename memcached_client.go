package gomemcached

type MemcachedClient struct {
	cluster *cluster
}

func NewMemcachedClient(addrs []string, maxConnPerServer uint32) *MemcachedClient {
	m := &MemcachedClient{}
	m.cluster = createCluster(addrs, maxConnPerServer)
	return m
}

func (m *MemcachedClient) Get(key string, value interface{}) (uint64, error) {
	server, cmder, err := m.cluster.ChooseServerCommand(key)
	if err != nil {
		return 0, err
	}

	cas, err := cmder.get(key, value)
	if err == nil {
		m.cluster.ReleaseServerCommand(server, cmder)
	}
	return cas, err
}

func (m *MemcachedClient) Set(key string, value string, expiration uint32, cas uint64) error {
	server, cmder, err := m.cluster.ChooseServerCommand(key)
	if err != nil {
		return err
	}

	err = cmder.set(key, value, expiration, cas)
	if err == nil {
		m.cluster.ReleaseServerCommand(server, cmder)
	}
	return err
}

/*func (m *MemcachedClient) Add(key string, value string, expiration uint32, cas uint64) error {
	server, cmder, err := m.cluster.ChooseServerCommand(key)
	if err != nil {
		return err
	}

	err = cmder.add(key, value, expiration, cas)
	if err == nil {
		m.cluster.ReleaseServerCommand(server, cmder)
	}
	return err
}*/
