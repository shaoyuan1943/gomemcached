package gomemcached

import (
	"bufio"
	"encoding/binary"
	"net"
	"strconv"
	"time"

	"github.com/valyala/bytebufferpool"

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

func newCommander(ID int64, conn net.Conn, s *Server) *Commander {
	return &Commander{
		ID:   ID,
		conn: conn,
		rw: bufio.NewReadWriter(
			bufio.NewReader(conn),
			bufio.NewWriter(conn),
		),
		server: s,
		giveup: false,
	}
}

func (cmder *Commander) wait4Rsp(req *bytebufferpool.ByteBuffer) (*bytebufferpool.ByteBuffer, uint8, uint64, error) {
	if err := cmder.write(req); err != nil {
		return nil, 0, 0, err
	}

	if err := cmder.flush2Server(); err != nil {
		return nil, 0, 0, err
	}

	rsp := bytebufferpool.Get()
	defer bytebufferpool.Put(rsp)

	rsp.Reset()
	if _, err := cmder.readN(rsp, RSP_HEADER_LEN); err != nil {
		return nil, 0, 0, err
	}

	extLen := rsp.B[4]
	status := binary.BigEndian.Uint16(rsp.B[6:8])
	bodyLen := binary.BigEndian.Uint32(rsp.B[8:12])
	cas := binary.BigEndian.Uint64(rsp.B[16:24])

	body := bytebufferpool.Get()
	if bodyLen > 0 {
		if _, err := cmder.readN(body, (int)(bodyLen)); err != nil {
			bytebufferpool.Put(body)
			return nil, 0, 0, err
		}
	}

	if err := checkStatus(status); err != nil {
		return nil, 0, 0, err
	}

	return body, extLen, cas, nil
}

func (cmder *Commander) store(opCode uint8, args *KeyArgs) (uint64, error) {
	req := bytebufferpool.Get()
	defer bytebufferpool.Put(req)

	var encoder Encoder
	var rawValue []byte
	var err error
	if args.useMsgpack {
		// type value --> raw value
		encoder = getEncoder()
		rawValue, err = encoder.Encode(args.Value)
		if err != nil {
			return 0, ErrMarshalFailed
		}
	} else {
		rawValue = args.Value.([]byte)
	}

	// request header
	writeReqHeader(req, MAGIC_REQUEST, opCode, (uint16)(len(args.Key)), 0x08, RAW_DATA, 0x00,
		uint32(0x08+len(args.Key)+len(rawValue)), 0x00, args.CAS)

	// extra:8byte |----flag:4----|----expiration:4----|
	if args.useMsgpack {
		WriteUint32(req, (uint32)(USE_MSGP_FLAG))
	} else {
		WriteUint32(req, 0)
	}
	WriteUint32(req, args.Expiration)
	// extra end

	// key
	req.WriteString(args.Key)
	// value
	req.Write(rawValue)

	if encoder != nil {
		putEncoder(encoder)
	}

	body, _, modifyCAS, err := cmder.wait4Rsp(req)
	defer func() {
		if body != nil {
			bytebufferpool.Put(body)
		}
	}()

	return modifyCAS, err
}

func (cmder *Commander) get(key string, value interface{}) (uint64, error) {
	req := bytebufferpool.Get()
	defer bytebufferpool.Put(req)

	// request header
	writeReqHeader(req, MAGIC_REQUEST, OPCODE_GET, (uint16)(len(key)), 0x00, RAW_DATA, 0x00,
		(uint32)(len(key)), 0x00, 0x00)
	// key
	req.WriteString(key)

	// flush to memcached server
	body, extLen, cas, err := cmder.wait4Rsp(req)
	defer func() {
		if body != nil {
			bytebufferpool.Put(body)
		}
	}()
	if err != nil {
		return 0, err
	}

	flag := binary.BigEndian.Uint32(body.Bytes()[:extLen])
	if flag == USE_MSGP_FLAG {
		decoder := getDecoder()
		defer putDecoder(decoder)
		err = decoder.Decode(body.Bytes()[extLen:], value)
		if err != nil {
			return 0, ErrUnmarshalFailed
		}
	} else {
		switch value.(type) {
		case *[]byte:
			v := value.(*[]byte)
			*v = append((*v), body.Bytes()[extLen:]...)
		default:
			return 0, ErrTypeInvalid
		}
	}

	return cas, nil
}

func (cmder *Commander) noop() error {
	req := bytebufferpool.Get()
	defer bytebufferpool.Put(req)

	// request header
	writeReqHeader(req, MAGIC_REQUEST, OPCODE_NOOP, 0x00, 0x00, RAW_DATA, 0x00,
		0x00, 0x00, 0x00)

	body, _, _, err := cmder.wait4Rsp(req)
	defer func() {
		if body != nil {
			bytebufferpool.Put(body)
		}
	}()

	return err
}

func (cmder *Commander) delete(key string, cas uint64) error {
	req := bytebufferpool.Get()
	defer bytebufferpool.Put(req)

	// request header
	writeReqHeader(req, MAGIC_REQUEST, OPCODE_DEL, uint16(len(key)), 0x00, RAW_DATA, 0x00,
		uint32(len(key)), 0x00, 0x00)

	// key
	req.WriteString(key)

	body, _, _, err := cmder.wait4Rsp(req)
	defer func() {
		if body != nil {
			bytebufferpool.Put(body)
		}
	}()

	return err
}

func (cmder *Commander) append(opCode uint8, args *KeyArgs) (uint64, error) {
	value, ok := args.Value.([]byte)
	if !ok {
		return 0, ErrCommandArgumentsInvalid
	}

	req := bytebufferpool.Get()
	defer bytebufferpool.Put(req)

	// request header
	writeReqHeader(req, MAGIC_REQUEST, opCode, uint16(len(args.Key)), 0x00, RAW_DATA, 0x00,
		uint32(len(args.Key)+len(value)), 0x00, args.CAS)
	// key
	req.WriteString(args.Key)
	// value
	req.Write(value)

	body, _, modifyCAS, err := cmder.wait4Rsp(req)
	defer func() {
		if body != nil {
			bytebufferpool.Put(body)
		}
	}()

	return modifyCAS, err
}

func (cmder *Commander) atomic(opCode uint8, args *KeyArgs) (uint64, uint64, error) {
	extData := bytebufferpool.Get()
	defer bytebufferpool.Put(extData)

	WriteUint64(extData, args.Delta)
	WriteUint64(extData, 0x0000000000000000)
	WriteUint32(extData, args.Expiration)

	req := bytebufferpool.Get()
	defer bytebufferpool.Put(req)

	// request header
	writeReqHeader(req, MAGIC_REQUEST, opCode, uint16(len(args.Key)), 0x14, RAW_DATA, 0x00,
		uint32(len(args.Key)+0x14), 0x00, args.CAS)
	// ext data
	req.Write(extData.Bytes())
	// key
	req.WriteString(args.Key)

	body, extLen, cas, err := cmder.wait4Rsp(req)
	defer func() {
		if body != nil {
			bytebufferpool.Put(body)
		}
	}()

	if err != nil {
		return 0, 0, err
	}

	atomicValue := binary.BigEndian.Uint64(body.Bytes()[extLen:])
	return atomicValue, cas, err
}

func (cmder *Commander) touchAtomicValue(key string) (uint64, error) {
	req := bytebufferpool.Get()
	defer bytebufferpool.Put(req)

	// request header
	writeReqHeader(req, MAGIC_REQUEST, OPCODE_GET, (uint16)(len(key)), 0x00, RAW_DATA, 0x00,
		(uint32)(len(key)), 0x00, 0x00)

	// key
	req.WriteString(key)

	body, extLen, _, err := cmder.wait4Rsp(req)
	defer func() {
		if body != nil {
			bytebufferpool.Put(body)
		}
	}()

	if err != nil {
		return 0, err
	}

	value, err := strconv.Atoi(string(body.Bytes()[extLen:]))
	if err != nil {
		return 0, err
	}

	return uint64(value), nil
}

func (cmder *Commander) flush(args *KeyArgs) error {
	req := bytebufferpool.Get()
	defer bytebufferpool.Put(req)

	extData := bytebufferpool.Get()
	defer bytebufferpool.Put(req)
	WriteUint32(extData, args.Expiration)

	// header
	writeReqHeader(req, MAGIC_REQUEST, OPCODE_FLUSH, 0x00, 0x04, RAW_DATA, 0x00,
		0x04, 0x00, 0x00)
	// ext data
	req.Write(extData.Bytes())

	body, _, _, err := cmder.wait4Rsp(req)
	defer func() {
		if body != nil {
			bytebufferpool.Put(body)
		}
	}()

	return err
}

func writeReqHeader(buffer *bytebufferpool.ByteBuffer, magic uint8, opcode uint8, keyLen uint16,
	extLen uint8, dataType uint8, status uint16, bodyLen uint32, opaque uint32, cas uint64) {
	if buffer == nil {
		panic("target buffer invalid")
	}

	buffer.WriteByte(magic)      // 0
	buffer.WriteByte(opcode)     // 1
	WriteUint16(buffer, keyLen)  // 2,3
	buffer.WriteByte(extLen)     // 4
	buffer.WriteByte(dataType)   // 5
	WriteUint16(buffer, status)  // 6,7
	WriteUint32(buffer, bodyLen) // 8,9,10,11
	WriteUint32(buffer, opaque)  // 12,13,14,15
	WriteUint64(buffer, cas)     // 16, 23
}
