package gomemcached

import (
	"bufio"
	"fmt"
	"net"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/karlseguin/bytepool"
)

var (
	NodeRepetitions       = 160
	RingPosition          = 4
	CommanderID     int64 = 10000
)

type Server struct {
	Addr              string
	VirtualHashs      []uint32
	MaxCommanderCount uint32
	cmders            map[int64]*Commander
	badCmders         chan *Commander
}

type cluster struct {
	hash2Servers map[uint32]*Server
	addr2Servers map[string]*Server
	nodeList     []uint32
	sync.RWMutex
}

func connect(addr string) (net.Conn, error) {
	if len(addr) <= 0 {
		return nil, ErrInvalidArguments
	}

	conn, err := net.DialTimeout("tcp", addr, ConnectTimeout)
	if err != nil {
		return nil, ErrNotConnected
	}

	return conn, nil
}

func createCluster(addrs []string, maxConnPerServer uint32) *cluster {
	cl := &cluster{
		hash2Servers: make(map[uint32]*Server),
		nodeList:     make([]uint32, len(addrs)*int(maxConnPerServer)),
		addr2Servers: make(map[string]*Server, len(addrs)),
	}

	for _, addr := range addrs {
		s := &Server{
			Addr:              addr,
			MaxCommanderCount: maxConnPerServer,
			cmders:            make(map[int64]*Commander, maxConnPerServer),
			badCmders:         make(chan *Commander, maxConnPerServer),
		}
		cl.hashServer(s)
	}

	sort.Sort(SortList(cl.nodeList))
	go cl.checkServer()
	return cl
}

func (s *Server) getCmder() (*Commander, error) {
	if len(s.cmders) <= 0 {
		return nil, ErrNoUsableConnection
	}

	var targetCmder *Commander
	for _, cmder := range s.cmders {
		if cmder != nil {
			targetCmder = cmder
			break
		}
	}

	// if found enable commander, delete it from 's.cmders'
	// then use commander completed, put it to 's.cmders'
	if targetCmder != nil {
		delete(s.cmders, targetCmder.ID)
		return targetCmder, nil
	}

	return nil, ErrNoUsableConnection
}
func (s *Server) putCmder(cmder *Commander) {
	if cmder != nil && !cmder.giveup {
		s.cmders[cmder.ID] = cmder
	}
}

func (cl *cluster) hashServer(s *Server) {
	cl.addr2Servers[s.Addr] = s
	for i := 0; i < NodeRepetitions/4; i++ {
		hashs := KetamaHash(s.Addr, (uint32)(i))
		s.VirtualHashs = append(s.VirtualHashs, hashs...)
		cl.nodeList = append(cl.nodeList, hashs...)
		for _, hashValue := range hashs {
			cl.hash2Servers[hashValue] = s
		}
	}

	for i := 0; i < int(s.MaxCommanderCount); i++ {
		conn, err := connect(s.Addr)
		if err == nil {
			ID := atomic.AddInt64(&CommanderID, 1)
			cmder := &Commander{
				ID:   ID,
				conn: conn,
				rw: bufio.ReadWriter{
					Reader: bufio.NewReader(conn),
					Writer: bufio.NewWriter(conn),
				},
				pool:   bytepool.New(24, 256),
				server: s,
				giveup: false,
			}
			s.cmders[ID] = cmder
		}
	}
}

func (cl *cluster) chooseServer(key string) *Server {
	if len(cl.nodeList) <= 0 {
		return nil
	}

	var targetHash uint32
	hashValue := MakeHash(key)
	if hashValue > cl.nodeList[len(cl.nodeList)-1] {
		targetHash = cl.nodeList[0]
	} else {
		l := 0
		r := len(cl.nodeList)
		i := 0
		for {
			mid := (l + r) / 2
			if hashValue == cl.nodeList[mid] {
				i = mid
				break
			} else if hashValue > cl.nodeList[mid] {
				l = mid
			} else {
				r = mid
			}

			if r-l == 1 {
				i = r
				break
			}
		}

		targetHash = cl.nodeList[i]
	}

	if targetHash <= 0 {
		return nil
	}

	s, ok := cl.hash2Servers[targetHash]
	if !ok {
		panic("virtual node not found in cluster")
	}

	return s
}

func (cl *cluster) ChooseServerCommand(key string) (*Server, *Commander, error) {
	cl.RLock()
	defer cl.RUnlock()

	s := cl.chooseServer(key)
	if s == nil {
		return nil, nil, ErrInvalidArguments
	}

	cmder, err := s.getCmder()
	if err == ErrNoUsableConnection {
		for _, s = range cl.addr2Servers {
			cmder, err = s.getCmder()
			if err == nil {
				return s, cmder, nil
			}
		}
	}

	return s, cmder, err
}

func (cl *cluster) ReleaseServerCommand(s *Server, cmder *Commander) {
	cl.Lock()
	defer cl.Unlock()

	s.putCmder(cmder)
}

func (cl *cluster) ReloadCluster(addr string, maxConnPerServer uint32) error {
	cl.Lock()
	defer cl.Unlock()

	if _, ok := cl.addr2Servers[addr]; ok {
		return ErrInvalidArguments
	}

	s := &Server{
		Addr:              addr,
		MaxCommanderCount: maxConnPerServer,
		cmders:            make(map[int64]*Commander, maxConnPerServer),
		badCmders:         make(chan *Commander, maxConnPerServer),
	}

	cl.hashServer(s)
	return nil
}

func (cl *cluster) checkServer() {
	keepTimer := time.After(time.Second * 5)
	for {
		select {
		case <-keepTimer:
			cl.doCheckServer()
		default:
		}
	}
}

func (cl *cluster) doCheckServer() {
	cl.Lock()
	defer cl.Unlock()

	needRebuild := false
	for _, s := range cl.addr2Servers {
		if len(s.badCmders) >= int(s.MaxCommanderCount) {
			// if all commanders of server failed, remove this server from cluster
			fmt.Printf("some server(%v) failed.\n", s.Addr)

			cl.cleanBadServer(s)
			needRebuild = true
			continue
		}

		// heartbeat
		for _, cmder := range s.cmders {
			cmder.noop()
		}
	}

	if needRebuild {
		// rebuild nodeList
		nodeList := cl.nodeList[:0]
		for _, s := range cl.addr2Servers {
			nodeList = append(nodeList, s.VirtualHashs...)
		}
		cl.nodeList = nodeList
		sort.Sort(SortList(cl.nodeList))

		fmt.Printf("rebuild cluster, nodeList: %v\n", len(cl.nodeList))
	}
}

func (cl *cluster) cleanBadServer(s *Server) {
	// remove sever from c.servers
	for _, v := range s.VirtualHashs {
		delete(cl.hash2Servers, v)
	}

	// remove from c.allServer
	delete(cl.addr2Servers, s.Addr)
}
