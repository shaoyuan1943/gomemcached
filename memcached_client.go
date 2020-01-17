package gomemcached

type MemcachedClient struct {
	cluster *cluster
}

func NewMemcachedClient(addrs []string, maxConnPerServer int) *MemcachedClient {
	m := &MemcachedClient{}
	m.cluster = createCluster(addrs, maxConnPerServer)
	return m
}

func (m *MemcachedClient) choose(key string) (*server, *Command, error) {
	if len(key) <= 0 {
		return nil, nil, ErrInvalidArguments
	}

	server := m.cluster.chooseServer(key)
	if server == nil {
		return nil, nil, ErrNotFoundServerNode
	}

	cmd, err := server.GetCmd()
	if err != nil {
		return nil, nil, err
	}

	return server, cmd, nil
}

func (m *MemcachedClient) Get(key string, value interface{}) (uint64, error) {
	server, cmd, err := m.choose(key)
	if err != nil {
		return 0, err
	}

	cas, err := cmd.get(key, value)
	server.PutCmd(cmd)
	return cas, err
}

func (m *MemcachedClient) Set(key string, value string, expiration uint32, cas uint64) error {
	server, cmd, err := m.choose(key)
	if err != nil {
		return err
	}

	err = cmd.set(key, value, expiration, cas)
	server.PutCmd(cmd)
	return err
}
