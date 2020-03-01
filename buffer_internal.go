package gomemcached

import (
	"encoding/binary"
	. "github.com/valyala/bytebufferpool"
)

func WriteUint16(buffer *ByteBuffer, value uint16, scratch []byte) {
	binary.BigEndian.PutUint16(scratch, value)
	buffer.Write(scratch[:2])
}

func WriteUint32(buffer *ByteBuffer, value uint32, scratch []byte) {
	binary.BigEndian.PutUint32(scratch, value)
	buffer.Write(scratch[:4])
}

func WriteUint64(buffer *ByteBuffer, value uint64, scratch []byte) {
	binary.BigEndian.PutUint64(scratch, value)
	buffer.Write(scratch)
}
