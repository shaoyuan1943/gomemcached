package gomemcached

import (
	"bufio"
	"encoding/binary"
	"net"
	"time"

	"github.com/vmihailenco/msgpack/v4"

	"github.com/karlseguin/bytepool"
)

var (
	ConnectTimeout = time.Duration(5) * time.Second
	ReadTimeout    = time.Duration(5) * time.Second
	WriterTimeout  = time.Duration(5) * time.Second
)

type Commander struct {
	ID     int64
	conn   net.Conn
	rw     bufio.ReadWriter
	pool   *bytepool.Pool
	server *Server
	giveup bool
}

func (cmder *Commander) waitForResponse(req *bytepool.Bytes) (*bytepool.Bytes, uint8, uint64, error) {
	if err := cmder.write(req); err != nil {
		cmder.Giveup()
		return nil, 0, 0, err
	}

	if err := cmder.flush(); err != nil {
		cmder.Giveup()
		return nil, 0, 0, err
	}

	rsp := cmder.pool.Checkout()
	defer rsp.Release()

	if err := cmder.readN(rsp, RSP_HEADER_LEN); err != nil {
		cmder.Giveup()
		return nil, 0, 0, err
	}

	b := rsp.Bytes()
	extLen := b[4]
	status := binary.BigEndian.Uint16(b[6:8])
	bodyLen := binary.BigEndian.Uint32(b[8:12])
	cas := binary.BigEndian.Uint64(b[16:24])

	if err := checkStatus(status); err != nil {
		return nil, 0, 0, err
	}

	if bodyLen > 0 {
		body := cmder.pool.Checkout()
		if err := cmder.readN(body, bodyLen); err != nil {
			cmder.Giveup()
			return nil, 0, 0, err
		}

		return body, extLen, cas, nil
	}

	return nil, extLen, cas, nil
}

func (cmder *Commander) set(key string, value interface{}, expiration uint32, cas uint64) error {
	r := &requestHeader{
		magic:    MAGIC_REQUEST,
		opcode:   OPCODE_SET,
		keyLen:   (uint16)(len(key)),
		extLen:   0x00,
		dataType: RAW_DATA,
		status:   0x00,
		bodyLen:  0x00,
		opaque:   0x00,
		cas:      cas,
	}

	req := cmder.pool.Checkout()
	defer req.Release()

	// type value --> raw value
	rawValue, err := msgpack.Marshal(value)
	if err != nil {
		return err
	}

	r.extLen = 0x08
	// extra len, key len, value len
	r.bodyLen = uint32(0x08 + len(key) + len(rawValue))
	// request header
	cmder.writeRequestHeader(r, req)
	// extra:8byte |----flag:4----|----expiration:4----|
	req.WriteUint32(0)
	req.WriteUint32(expiration)
	// key
	if _, err = req.WriteString(key); err != nil {
		return err
	}
	// value
	if _, err = req.Write(rawValue); err != nil {
		return err
	}

	_, _, _, err = cmder.waitForResponse(req)
	return err
}

func (cmder *Commander) get(key string, value interface{}) (uint64, error) {
	r := &requestHeader{
		magic:    MAGIC_REQUEST,
		opcode:   OPCODE_GET,
		keyLen:   (uint16)(len(key)),
		extLen:   0x00,
		dataType: RAW_DATA,
		status:   0x00,
		bodyLen:  (uint32)(len(key)),
		opaque:   0x00,
		cas:      0x00,
	}

	req := cmder.pool.Checkout()
	defer req.Release()

	// request header
	cmder.writeRequestHeader(r, req)
	// key
	if _, err := req.WriteString(key); err != nil {
		return 0, err
	}

	// flush to memcached server
	rawValue, extLen, cas, err := cmder.waitForResponse(req)
	if err != nil {
		return 0, err
	}

	if rawValue != nil {
		err = msgpack.Unmarshal(rawValue.Bytes()[extLen:], value)
		rawValue.Release()
		if err != nil {
			return 0, err
		}

		return cas, nil
	}

	return cas, nil
}

func (cmder *Commander) noop() error {
	r := &requestHeader{
		magic:    MAGIC_REQUEST,
		opcode:   OPCODE_NOOP,
		keyLen:   0x00,
		extLen:   0x00,
		dataType: RAW_DATA,
		status:   0x00,
		bodyLen:  0x00,
		opaque:   0x00,
		cas:      0x00,
	}

	req := cmder.pool.Checkout()
	defer req.Release()

	cmder.writeRequestHeader(r, req)
	_, _, _, err := cmder.waitForResponse(req)
	return err
}

func (cmder *Commander) writeRequestHeader(r *requestHeader, b *bytepool.Bytes) {
	b.WriteByte((byte)(r.magic))
	b.WriteByte((byte)(r.opcode))
	b.WriteUint16(r.keyLen)
	b.WriteByte((byte)(r.extLen))
	b.WriteByte((byte)(r.dataType))
	b.WriteUint16((uint16)(r.status))
	b.WriteUint32((uint32)(r.bodyLen))
	b.WriteUint32((uint32)(r.opaque))
	b.WriteUint64((uint64)(r.cas))
}
