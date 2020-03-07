package gomemcached

import (
	"io"
	"time"

	"github.com/valyala/bytebufferpool"
)

func (cmder *Commander) Giveup() {
	if cmder.giveup {
		return
	}

	cmder.conn.Close()
	cmder.giveup = true
	cmder.server.badCmders = append(cmder.server.badCmders, cmder)
	cmder.server.cluster.badServerNoticer <- cmder.server
}
func (cmder *Commander) flush2Server() error {
	cmder.conn.SetWriteDeadline(time.Now().Add(WriterTimeout))
	return cmder.rw.Flush()
}

func (cmder *Commander) readN(buffer *bytebufferpool.ByteBuffer, count int) (int, error) {
	if count <= 0 {
		return 0, ErrInvalidArguments
	}

	cmder.conn.SetReadDeadline(time.Now().Add(ReadTimeout))
	start := len(buffer.B)
	max := cap(buffer.B)
	if max == 0 {
		max = count
		buffer.B = make([]byte, max)
	} else {
		if max-start < count {
			newBytes := make([]byte, max+count)
			copy(newBytes, buffer.B)
			buffer.B = newBytes
		} else {
			buffer.B = buffer.B[:max]
		}
	}

	n, err := io.ReadFull(cmder.rw, buffer.B[start:start+count])
	return n, err
}

func (cmder *Commander) write(buffer *bytebufferpool.ByteBuffer) error {
	_, err := cmder.rw.Write(buffer.Bytes())
	return err
}
