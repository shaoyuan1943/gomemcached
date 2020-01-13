package gomemcached

type connPool struct {
}

func newConnPool(addrs []string) *connPool {
	pool := &connPool{}
	for _, addr := range addrs {

	}
	return pool
}
