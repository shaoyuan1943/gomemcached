package gomemcached

import (
	"bufio"
	"encoding/binary"
	"net"
	"time"

	"github.com/karlseguin/bytepool"
)

var (
	ConnectTimeout = time.Duration(5) * time.Second
	ReadTimeout    = time.Duration(5) * time.Second
	WriterTimeout  = time.Duration(5) * time.Second
)

type Conn struct {
	conn net.Conn
	rw   bufio.ReadWriter

	bytesPool *bytepool.Pool
}

func connect(addr string) (*Conn, error) {
	if len(addr) <= 0 {
		return nil, ErrInvalidArguments
	}

	conn, err := net.DialTimeout("tcp", addr, ConnectTimeout)
	if err != nil {
		return nil, ErrNotConnected
	}

	c := &Conn{
		conn: conn,
		rw: bufio.ReadWriter{
			Reader: bufio.NewReader(conn),
			Writer: bufio.NewWriter(conn),
		},
	}

	c.bytesPool = bytepool.New(24, 1024)
	return c, nil
}

func (c *Conn) fillRequestHeader(header *requestHeader) error {
	b := c.bytesPool.Checkout()
	defer b.Release()

	buff := b.Bytes()
	buff[0] = uint8(header.magic)
	buff[1] = uint8(header.opcode)

	binary.BigEndian.PutUint16(buff[2:4], header.keyLen)

	return nil
}
