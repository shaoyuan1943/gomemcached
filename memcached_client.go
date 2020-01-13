package gomemcached

type MemcachedClient struct {
	cluster *cluster
}

func NewMemcachedClient(addrs []string) *MemcachedClient {
	m := &MemcachedClient{}
	m.cluster = createCluster(addrs)
	return m
}

func (m *MemcachedClient) Get(key string, format interface{}) {

}
