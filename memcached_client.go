package gomemcached

type MemcachedClient struct {
	cluster *cluster
}

func NewMemcachedClient(addrs []string, maxConnPerServer int) *MemcachedClient {
	m := &MemcachedClient{}
	m.cluster = createCluster(addrs, maxConnPerServer)
	return m
}

func (m *MemcachedClient) choose(key string) (*Command, error) {
	if len(key) <= 0 {
		return nil, ErrInvalidArguments
	}

	server := m.cluster.chooseServer(key)
	if server == nil {
		return nil, ErrNotFoundServerNode
	}

	cmd, err := server.GetCmd()
	if err != nil {
		return nil, err
	}

	return cmd, nil
}

func (m *MemcachedClient) Get(key string, value interface{}) (uint64, error) {
	cmd, err := m.choose(key)
	if err != nil {
		return 0, err
	}

	cas, err := cmd.get(key, value)
	if err != nil {
		return 0, err
	}

	return cas, nil
}

func (m *MemcachedClient) Set(key string, value string, expiration uint32, cas uint64) error {
	cmd, err := m.choose(key)
	if err != nil {
		return err
	}

	return cmd.set(key, value, expiration, cas)
}
