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

func (c *Conn) makeResponseHeaderFromConn(data []byte) *responseHeader {
	return &responseHeader{
		magic:    (MagicType)(data[0]),
		opcode:   (OpcodeType)(data[1]),
		keyLen:   (uint16)(binary.BigEndian.Uint16(data[2:4])),
		extLen:   (uint8)(data[4]),
		dataType: (DataType)(data[5]),
		status:   (StatusType)(binary.BigEndian.Uint16(data[6:8])),
		bodyLen:  (uint32)(binary.BigEndian.Uint32(data[8:10])),
		opaque:   (uint32)(binary.BigEndian.Uint32(data[10:12])),
		cas:      (uint64)(binary.BigEndian.Uint64(data[12:24])),
	}
}

func (c *Conn) makeRequestHeader2Conn(header *requestHeader) error {
	b := c.bytesPool.Checkout()
	defer b.Release()

	b.WriteByte((byte)(header.magic))
	b.WriteByte((byte)(header.opcode))
	b.WriteUint16(header.keyLen)
	b.WriteByte((byte)(header.extLen))
	b.WriteByte((byte)(header.dataType))
	b.WriteUint16((uint16)(header.status))
	b.WriteUint32((uint32)(header.bodyLen))
	b.WriteUint32((uint32)(header.opaque))
	b.WriteUint64((uint64)(header.cas))

	len, err := c.rw.Write(b.Bytes())
	if err != nil {
		return err
	}

	if len < REQ_HEADER_LEN {
		return ErrFillRequestHeaderFailed
	}

	return nil
}
