package gomemcached

import (
	"encoding/binary"

	"github.com/karlseguin/bytepool"
	"github.com/vmihailenco/msgpack/v4"
)

type Command struct {
	conn     *Conn
	bytePool *bytepool.Pool
}

func newCommand(c *Conn) *Command {
	cmd := &Command{conn: c}
	cmd.bytePool = bytepool.New(24, 1024)
	return cmd
}

func (cmd *Command) writeRequestHeader(r *requestHeader, b *bytepool.Bytes) {
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

func (cmd *Command) waitForResponse(req *bytepool.Bytes) (*bytepool.Bytes, uint8, uint64, error) {
	if err := cmd.conn.write(req); err != nil {
		return nil, 0, 0, err
	}

	if err := cmd.conn.flush(); err != nil {
		return nil, 0, 0, err
	}

	rsp := cmd.bytePool.Checkout()
	defer rsp.Release()

	if err := cmd.conn.read(rsp); err != nil {
		return nil, 0, 0, err
	}

	b := rsp.Bytes()
	extLen := b[4]
	status := binary.BigEndian.Uint16(b[6:8])
	bodyLen := binary.BigEndian.Uint32(b[8:12])
	cas := binary.BigEndian.Uint64(b[16:24])

	if err := checkError(status); err != nil {
		return nil, 0, 0, err
	}

	if bodyLen > 0 {
		body := cmd.bytePool.Checkout()
		if err := cmd.conn.read(body); err != nil {
			return nil, 0, 0, err
		}

		return body, extLen, cas, nil
	}

	return nil, extLen, cas, nil
}

func (cmd *Command) set(key string, value interface{}, expiration uint32, cas uint64) error {
	r := &requestHeader{
		magic:    MAGIC_REQUEST,
		opcode:   OPCODE_SET,
		keyLen:   (uint16)(len(key)),
		extLen:   0x00,
		dataType: RAW_DATA,
		status:   0x00,
		bodyLen:  0x00,
		opaque:   0x00,
		cas:      0x00,
	}

	req := cmd.bytePool.Checkout()
	defer req.Release()

	// type value --> raw value
	rawValue, err := msgpack.Marshal(value)
	if err != nil {
		return err
	}
	r.extLen = 0x08
	// extra len, key len, value len
	r.bodyLen = uint32(0x08 + len(key) + len(rawValue))

	cmd.writeRequestHeader(r, req)
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

	_, _, _, err = cmd.waitForResponse(req)
	if err != nil {
		return err
	}

	return nil
}

func (cmd *Command) get(key string, value interface{}) (uint64, error) {
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

	req := cmd.bytePool.Checkout()
	defer req.Release()

	// request header
	cmd.writeRequestHeader(r, req)
	// key
	if _, err := req.WriteString(key); err != nil {
		return 0, err
	}

	// flush to memcached server
	rawValue, extLen, cas, err := cmd.waitForResponse(req)
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
