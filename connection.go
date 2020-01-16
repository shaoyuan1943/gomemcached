package gomemcached

import (
	"bufio"
	"io"
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

	return c, nil
}

func (c *Conn) flush() error {
	c.conn.SetWriteDeadline(time.Now().Add(WriterTimeout))
	return c.rw.Flush()
}

func (c *Conn) read(b *bytepool.Bytes) error {
	c.conn.SetReadDeadline(time.Now().Add(ReadTimeout))
	_, err := io.ReadFull(c.rw, b.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func (c *Conn) write(b *bytepool.Bytes) error {
	_, err := c.rw.Write(b.Bytes())
	return err
}
