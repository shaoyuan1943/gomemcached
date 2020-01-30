package gomemcached

import (
	"bufio"
	"encoding/binary"
	"net"
	"strconv"
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
	rw     *bufio.ReadWriter
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

	body := cmder.pool.Checkout()
	if bodyLen > 0 {
		if err := cmder.readN(body, bodyLen); err != nil {
			cmder.Giveup()
			body.Release()
			return nil, 0, 0, err
		}
	}

	if err := checkStatus(status); err != nil {
		body.Release()
		return nil, 0, 0, err
	}

	return body, extLen, cas, nil
}

func (cmder *Commander) store(opCode uint8, key string, value interface{}, expiration uint32, cas uint64, useMsgp bool) (uint64, error) {
	r := &RequestHeader{
		Magic:    MAGIC_REQUEST,
		Opcode:   opCode,
		KeyLen:   (uint16)(len(key)),
		ExtLen:   0x00,
		DataType: RAW_DATA,
		Status:   0x00,
		BodyLen:  0x00,
		Opaque:   0x00,
		CAS:      cas,
	}

	var err error
	req := cmder.pool.Checkout()
	defer req.Release()

	var byteData []byte
	if useMsgp {
		// type value --> raw value
		byteData, err = msgpack.Marshal(value)
		if err != nil {
			return 0, err
		}
	} else {
		byteData = value.([]byte)
	}

	// extra len, key len, value len
	r.ExtLen = 0x08
	r.BodyLen = uint32(0x08 + len(key) + len(byteData))

	// request header
	writeRequestHeader(r, req)
	// request end

	// extra:8byte |----flag:4----|----expiration:4----|
	if useMsgp {
		req.WriteUint32(USE_MSGP_FLAG)
	} else {
		req.WriteUint32(0)
	}
	req.WriteUint32(expiration)
	// extra end

	// key
	if _, err = req.WriteString(key); err != nil {
		return 0, err
	}

	// value
	if _, err = req.Write(byteData); err != nil {
		return 0, err
	}

	var modifyCAS uint64
	rawValue, _, modifyCAS, err := cmder.waitForResponse(req)
	defer func() {
		if rawValue != nil {
			rawValue.Release()
		}
	}()

	return modifyCAS, err
}

func (cmder *Commander) get(key string, value interface{}) (uint64, error) {
	r := &RequestHeader{
		Magic:    MAGIC_REQUEST,
		Opcode:   OPCODE_GET,
		KeyLen:   (uint16)(len(key)),
		ExtLen:   0x00,
		DataType: RAW_DATA,
		Status:   0x00,
		BodyLen:  (uint32)(len(key)),
		Opaque:   0x00,
		CAS:      0x00,
	}

	req := cmder.pool.Checkout()
	defer req.Release()

	// request header
	writeRequestHeader(r, req)
	// key
	if _, err := req.WriteString(key); err != nil {
		return 0, err
	}

	// flush to memcached server
	rawValue, extLen, cas, err := cmder.waitForResponse(req)
	defer func() {
		if rawValue != nil {
			rawValue.Release()
		}
	}()

	if err != nil {
		return 0, err
	}

	if rawValue == nil {
		return cas, nil
	}

	flag := binary.BigEndian.Uint32(rawValue.Bytes()[:extLen])
	if flag == USE_MSGP_FLAG {
		err = msgpack.Unmarshal(rawValue.Bytes()[extLen:], value)
		if err != nil {
			return 0, err
		}
	} else {
		switch value.(type) {
		case *[]byte:
			v := value.(*[]byte)
			*v = append((*v), rawValue.Bytes()[extLen:]...)
		default:
			return 0, ErrUnpackTypeInvalid
		}
	}

	return cas, nil
}

func (cmder *Commander) noop() error {
	r := &RequestHeader{
		Magic:    MAGIC_REQUEST,
		Opcode:   OPCODE_NOOP,
		KeyLen:   0x00,
		ExtLen:   0x00,
		DataType: RAW_DATA,
		Status:   0x00,
		BodyLen:  0x00,
		Opaque:   0x00,
		CAS:      0x00,
	}

	req := cmder.pool.Checkout()
	defer req.Release()

	writeRequestHeader(r, req)
	rawValue, _, _, err := cmder.waitForResponse(req)
	defer func() {
		if rawValue != nil {
			rawValue.Release()
		}
	}()
	return err
}

func (cmder *Commander) delete(key string, cas uint64) error {
	r := &RequestHeader{
		Magic:    MAGIC_REQUEST,
		Opcode:   OPCODE_DEL,
		KeyLen:   uint16(len(key)),
		ExtLen:   0x00,
		DataType: RAW_DATA,
		Status:   0x00,
		BodyLen:  uint32(len(key)),
		Opaque:   0x00,
		CAS:      0x00,
	}

	req := cmder.pool.Checkout()
	defer req.Release()
	// request header
	writeRequestHeader(r, req)
	// body: key
	req.WriteString(key)
	rawValue, _, _, err := cmder.waitForResponse(req)
	defer func() {
		if rawValue != nil {
			rawValue.Release()
		}
	}()
	return err
}

func (cmder *Commander) append(opCode uint8, key string, value []byte, cas uint64) (uint64, error) {
	r := &RequestHeader{
		Magic:    MAGIC_REQUEST,
		Opcode:   opCode,
		KeyLen:   uint16(len(key)),
		ExtLen:   0x00,
		DataType: RAW_DATA,
		Status:   0x00,
		BodyLen:  0x00,
		Opaque:   0x00,
		CAS:      cas,
	}

	req := cmder.pool.Checkout()
	defer req.Release()

	r.BodyLen = uint32(len(key) + len(value))
	writeRequestHeader(r, req)
	req.WriteString(key)
	req.Write(value)

	rawValue, _, modifyCAS, err := cmder.waitForResponse(req)
	defer func() {
		if rawValue != nil {
			rawValue.Release()
		}
	}()
	return modifyCAS, err
}

func (cmder *Commander) atomic(opCode uint8, key string, delta uint64, expiration uint32, cas uint64) (uint64, uint64, error) {
	r := &RequestHeader{
		Magic:    MAGIC_REQUEST,
		Opcode:   opCode,
		KeyLen:   uint16(len(key)),
		ExtLen:   0x14,
		DataType: RAW_DATA,
		Status:   0x00,
		BodyLen:  uint32(len(key) + 0x14),
		Opaque:   0x00,
		CAS:      cas,
	}

	req := cmder.pool.Checkout()
	defer req.Release()

	extData := cmder.pool.Checkout()
	defer extData.Release()

	extData.WriteUint64(delta)
	extData.WriteUint64(0x0000000000000000)
	extData.WriteUint32(expiration)

	writeRequestHeader(r, req)
	req.Write(extData.Bytes())
	req.WriteString(key)

	rawValue, extLen, cas, err := cmder.waitForResponse(req)
	defer func() {
		if rawValue != nil {
			rawValue.Release()
		}
	}()

	if err != nil {
		return 0, 0, err
	}

	atomicValue := binary.BigEndian.Uint64(rawValue.Bytes()[extLen:])
	return atomicValue, cas, err
}

func (cmder *Commander) touchAtomicValue(key string) (uint64, error) {
	r := &RequestHeader{
		Magic:    MAGIC_REQUEST,
		Opcode:   OPCODE_GET,
		KeyLen:   (uint16)(len(key)),
		ExtLen:   0x00,
		DataType: RAW_DATA,
		Status:   0x00,
		BodyLen:  (uint32)(len(key)),
		Opaque:   0x00,
		CAS:      0x00,
	}

	req := cmder.pool.Checkout()
	defer req.Release()

	// request header
	writeRequestHeader(r, req)
	// key
	if _, err := req.WriteString(key); err != nil {
		return 0, err
	}

	// flush to memcached server
	rawValue, extLen, _, err := cmder.waitForResponse(req)
	defer func() {
		if rawValue != nil {
			rawValue.Release()
		}
	}()
	if err != nil {
		return 0, err
	}

	if rawValue != nil {
		value, err := strconv.Atoi(string(rawValue.Bytes()[extLen:]))
		if err != nil {
			return 0, err
		}

		return uint64(value), nil
	}

	return 0, nil
}

func writeRequestHeader(r *RequestHeader, b *bytepool.Bytes) {
	b.WriteByte(r.Magic)     // 0
	b.WriteByte(r.Opcode)    // 1
	b.WriteUint16(r.KeyLen)  // 2,3
	b.WriteByte(r.ExtLen)    // 4
	b.WriteByte(r.DataType)  // 5
	b.WriteUint16(r.Status)  // 6,7
	b.WriteUint32(r.BodyLen) // 8,9,10,11
	b.WriteUint32(r.Opaque)  // 12,13,14,15
	b.WriteUint64(r.CAS)     // 16, 23
}
