package gomemcached

import (
	. "github.com/valyala/bytebufferpool"
)

func WriteUint16(buffer *ByteBuffer, value uint16) {
	buffer.WriteByte(byte(value >> 8))
	buffer.WriteByte(byte(value))
}

func WriteUint32(buffer *ByteBuffer, value uint32) {
	buffer.WriteByte(byte(value >> 24))
	buffer.WriteByte(byte(value >> 16))
	buffer.WriteByte(byte(value >> 8))
	buffer.WriteByte(byte(value))
}

func WriteUint64(buffer *ByteBuffer, value uint64) {
	buffer.WriteByte(byte(value >> 56))
	buffer.WriteByte(byte(value >> 48))
	buffer.WriteByte(byte(value >> 40))
	buffer.WriteByte(byte(value >> 32))
	buffer.WriteByte(byte(value >> 24))
	buffer.WriteByte(byte(value >> 16))
	buffer.WriteByte(byte(value >> 8))
	buffer.WriteByte(byte(value))
}
