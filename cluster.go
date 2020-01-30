package gomemcached

import (
	"bufio"
	"context"
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

type ServerErrorCallback func(addr string)

type Server struct {
	Addr              string
	VirtualHashs      []uint32
	MaxCommanderCount uint32
	cmders            map[int64]*Commander
	badCmders         []*Commander
	cluster           *Cluster
}

type Cluster struct {
	hash2Servers      map[uint32]*Server
	addr2Servers      map[string]*Server
	nodeList          []uint32
	ctx               context.Context
	quitF             context.CancelFunc
	serverErrCallback ServerErrorCallback
	badServerNoticer  chan *Server
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

func createCluster(addrs []string, maxConnPerServer uint32) *Cluster {
	cl := &Cluster{
		hash2Servers:     make(map[uint32]*Server),
		nodeList:         make([]uint32, len(addrs)*int(maxConnPerServer)),
		addr2Servers:     make(map[string]*Server, len(addrs)),
		badServerNoticer: make(chan *Server),
	}

	for _, addr := range addrs {
		s := &Server{
			Addr:              addr,
			MaxCommanderCount: maxConnPerServer,
			cmders:            make(map[int64]*Commander, maxConnPerServer),
			cluster:           cl,
		}
		cl.hashServer(s)
	}

	sort.Sort(SortList(cl.nodeList))

	cl.ctx, cl.quitF = context.WithCancel(context.Background())
	go cl.checkClusterServerNode()
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

func (cl *Cluster) exit() {
	cl.quitF()
}

func (cl *Cluster) hashServer(s *Server) {
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
				rw: bufio.NewReadWriter(
					bufio.NewReader(conn),
					bufio.NewWriter(conn),
				),
				pool:   bytepool.New(24, 256),
				server: s,
				giveup: false,
			}
			s.cmders[ID] = cmder
		}
	}
}

func (cl *Cluster) chooseServer(key string) *Server {
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
		panic("Virtual node not found in Cluster")
	}

	return s
}

func (cl *Cluster) ChooseServerCommand(key string) (*Server, *Commander, error) {
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

func (cl *Cluster) ReleaseServerCommand(s *Server, cmder *Commander) {
	cl.Lock()
	defer cl.Unlock()

	s.putCmder(cmder)
}

func (cl *Cluster) AddServer2Cluster(addr string, maxConnPerServer uint32) error {
	cl.Lock()
	defer cl.Unlock()

	if _, ok := cl.addr2Servers[addr]; ok {
		return ErrServerAlreadyInCluster
	}

	s := &Server{
		Addr:              addr,
		MaxCommanderCount: maxConnPerServer,
		cmders:            make(map[int64]*Commander, maxConnPerServer),
	}

	cl.hashServer(s)
	
	return nil
}

func (cl *Cluster) checkClusterServerNode() {
	heartbeatTimer := time.After(time.Second * 3)
	for {
		select {
		case <-cl.ctx.Done():
			return
		case s := <-cl.badServerNoticer:
			cl.doCheckServer(s)
		case <-heartbeatTimer:
			cl.doCheckHeartbeat()
		default:
		}
	}
}

func (cl *Cluster) doCheckServer(s *Server) {
	if len(s.badCmders) >= int(s.MaxCommanderCount) {
		cl.Lock()
		defer cl.Unlock()

		cl.cleanBadServer(s)
		if cl.serverErrCallback != nil {
			cl.serverErrCallback(s.Addr)
		}

		// rebuild nodeList
		nodeList := cl.nodeList[:0]
		for _, s := range cl.addr2Servers {
			nodeList = append(nodeList, s.VirtualHashs...)
		}
		cl.nodeList = nodeList
		sort.Sort(SortList(cl.nodeList))

		fmt.Printf("rebuild Cluster, nodeList: %v\n", len(cl.nodeList))
	}
}

func (cl *Cluster) cleanBadServer(s *Server) {
	// remove sever from c.servers
	for _, v := range s.VirtualHashs {
		delete(cl.hash2Servers, v)
	}

	// remove from c.allServer
	delete(cl.addr2Servers, s.Addr)
}

func (cl *Cluster) doCheckHeartbeat() {
	cl.RLock()
	defer cl.RUnlock()

	for _, server := range cl.addr2Servers {
		for _, cmder := range server.cmders {
			cmder.noop()
		}
	}
}
