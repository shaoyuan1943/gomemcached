package gomemcached

type MagicType uint8

const (
	MagicRequest  MagicType = 0x80
	MagicResponse MagicType = 0x81
)

type StatusType uint16

const (
	StatusOK                            StatusType = 0x0000
	StatusKeyNotFound                   StatusType = 0x0001
	StatusKeyExists                     StatusType = 0x0002
	StatusValueTooLarge                 StatusType = 0x0003
	StatusInvalidArguments              StatusType = 0x0004
	StatusItemNotStored                 StatusType = 0x0005
	StatusNonNumberValue                StatusType = 0x0006
	StatusVbucketBelongsToAnotherServer StatusType = 0x0007
	StatusAuthenticationError           StatusType = 0x0008
	StatusAuthenticationContinue        StatusType = 0x0009
	StatusUnknownCommand                StatusType = 0x0081
	StatusOutOfMemory                   StatusType = 0x0082
	StatusNotSupported                  StatusType = 0x0083
	StatusInternalError                 StatusType = 0x0084
	StatusBusy                          StatusType = 0x0085
	StatusTemporaryFailure              StatusType = 0x0086
)

type OpcodeType uint8

const (
	OpcodeGet       OpcodeType = 0x00
	OpcodeSet       OpcodeType = 0x01
	OpcodeAdd       OpcodeType = 0x02
	OpcodeReplace   OpcodeType = 0x03
	OpcodeDelete    OpcodeType = 0x04
	OpcodeIncrement OpcodeType = 0x05
	OpcodeDecrement OpcodeType = 0x06
	OpcodeQuit      OpcodeType = 0x07
	OpcodeFlush     OpcodeType = 0x08
	OpcodeNoop      OpcodeType = 0x0a
	OpcodeVersion   OpcodeType = 0x0b
	OpcodeGetK      OpcodeType = 0x0c
	OpcodeAppend    OpcodeType = 0x0e
	OpcodePrepend   OpcodeType = 0x0f
	OpcodeStat      OpcodeType = 0x10
)

type DataType uint8

var (
	DataRawBytes DataType = 0x00
)

var (
	REQ_HEADER_LEN int = 24
)

type requestHeader struct {
	magic    MagicType
	opcode   OpcodeType
	keyLen   uint16
	extLen   uint8
	dataType DataType
	status   StatusType
	bodyLen  uint32
	opaque   uint32
	cas      uint64
}
type responseHeader struct {
	magic    MagicType
	opcode   OpcodeType
	keyLen   uint16
	extLen   uint8
	dataType DataType
	status   StatusType
	bodyLen  uint32
	opaque   uint32
	cas      uint64
}
