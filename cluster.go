package gomemcached

import (
	"sort"
	"sync"
)

var (
	NodeRepetitions = 160
	RingPosition    = 4
)

type server struct {
	Addr string
}

type cluster struct {
	servers  map[uint32]*server
	nodeList []uint32

	sync.RWMutex
}

func createCluster(addrs []string) *cluster {
	cl := &cluster{
		servers: make(map[uint32]*server),
	}

	for _, addr := range addrs {
		s := &server{Addr: addr}
		for i := 0; i < NodeRepetitions/4; i++ {
			hashs := KetamaHash(s.Addr, (uint32)(i))
			cl.nodeList = append(cl.nodeList, hashs...)
			for _, hashValue := range hashs {
				cl.servers[hashValue] = s
			}
		}
	}

	sort.Sort((SortList)(cl.nodeList))
	return cl
}

func (c *cluster) FindServerByKey(key string) *server {
	c.Lock()
	defer c.Unlock()

	if len(c.nodeList) < 1 {
		return nil
	}

	var targetNode uint32
	hashValue := MakeHash(key)
	if hashValue > c.nodeList[len(c.nodeList)-1] {
		targetNode = c.nodeList[0]
	} else {
		l := 0
		r := len(c.nodeList)
		i := 0
		for {
			mid := (l + r) / 2
			if hashValue == c.nodeList[mid] {
				i = mid
				break
			}

			if hashValue > c.nodeList[mid] {
				l = mid
			} else {
				r = mid
			}

			if r-l == 1 {
				i = r
				break
			}
		}

		targetNode = c.nodeList[i]
	}

	if targetNode <= 0 {
		return nil
	}

	s, ok := c.servers[targetNode]
	if !ok {
		panic("virtual node not found in cluster")
	}

	return s
}
