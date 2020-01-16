package gomemcached

const (
	MAGIC_REQUEST  uint8 = 0x80
	MAGIC_RESPONSE uint8 = 0x81
)

const (
	STATUS_OK                uint16 = 0x0000
	STATUS_KEY_NOT_FOUND     uint16 = 0x0001
	STATUS_KEY_EXISTS        uint16 = 0x0002
	STATUS_VALUE_TOO_LARGE   uint16 = 0x0003
	STATUS_INVALID_ARGS      uint16 = 0x0004
	STATUS_NOT_STORED        uint16 = 0x0005
	STATUS_NON_NUMERIC_VALUE uint16 = 0x0006
	STATUS_VBUCKET_NOT_FOUND uint16 = 0x0007
	STATUS_AUTH_ERROR        uint16 = 0x0008
	STATUS_AUTH_CONTINUE     uint16 = 0x0009
	STATUS_UNKNOWN_COMMAND   uint16 = 0x0081
	STATUS_OUT_OF_MEMORY     uint16 = 0x0082
	STATUS_NOT_SUPPORTED     uint16 = 0x0083
	STATUS_INTERNAL_ERROR    uint16 = 0x0084
	STATUS_BUSY              uint16 = 0x0085
	STATUS_TEMPORARY_FAILURE uint16 = 0x0086
)

const (
	OPCODE_GET     uint8 = 0x00
	OPCODE_SET     uint8 = 0x01
	OPCODE_ADD     uint8 = 0x02
	OPCODE_REPLACE uint8 = 0x03
	OPCODE_DEL     uint8 = 0x04
	OPCODE_INCR    uint8 = 0x05
	OPCODE_DECR    uint8 = 0x06
	OPCODE_QUIT    uint8 = 0x07
	OPCODE_FLUSH   uint8 = 0x08
	OPCODE_NOOP    uint8 = 0x0a
	OPCODE_VERSION uint8 = 0x0b
	OPCODE_GETK    uint8 = 0x0c
	OPCODE_APPEND  uint8 = 0x0e
	OPCODE_PREPEND uint8 = 0x0f
	OPCODE_STAT    uint8 = 0x10
)

const (
	RAW_DATA uint8 = 0x00
)

const (
	REQ_HEADER_LEN int = 24
	RSP_HEADER_LEN int = 24
)

type requestHeader struct {
	magic    uint8
	opcode   uint8
	keyLen   uint16
	extLen   uint8
	dataType uint8
	status   uint16
	bodyLen  uint32
	opaque   uint32
	cas      uint64
}
type responseHeader struct {
	magic    uint8
	opcode   uint8
	keyLen   uint16
	extLen   uint8
	dataType uint8
	status   uint16
	bodyLen  uint32
	opaque   uint32
	cas      uint64
}

const (
	DATA_TYPE_INT     uint32 = 0x01
	DATA_TYPE_INT8    uint32 = 0x02
	DATA_TYPE_INT16   uint32 = 0x03
	DATA_TYPE_INT32   uint32 = 0x04
	DATA_TYPE_INT64   uint32 = 0x05
	DATA_TYPE_FLOAT32 uint32 = 0x06
	DATA_TYPE_FLOAT64 uint32 = 0x07
	DATA_TYPE_UINT    uint32 = 0x08
	DATA_TYPE_UINT8   uint32 = 0x09
	DATA_TYPE_UINT16  uint32 = 0x0A
	DATA_TYPE_UINT32  uint32 = 0x0B
	DATA_TYPE_UINT64  uint32 = 0x0C
	DATA_TYPE_BOOL    uint32 = 0x0D
	DATA_TYPE_STRING  uint32 = 0x0E
	DATA_TYPE_BINARY  uint32 = 0x0F
)
