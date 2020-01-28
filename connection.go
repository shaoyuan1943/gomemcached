package gomemcached

import (
	"fmt"
	"io"
	"time"

	"github.com/karlseguin/bytepool"
)

func (cmder *Commander) Giveup() {
	fmt.Printf("something wrong with this connection: %v\n", cmder.ID)
	cmder.conn.Close()
	cmder.giveup = true
	cmder.server.badCmders = append(cmder.server.badCmders, cmder)
	cmder.server.cluster.badServerNoticer <- cmder.server
}

func (cmder *Commander) flush() error {
	cmder.conn.SetWriteDeadline(time.Now().Add(WriterTimeout))
	return cmder.rw.Flush()
}

func (cmder *Commander) readN(b *bytepool.Bytes, n uint32) error {
	if n <= 0 {
		return ErrInvalidArguments
	}

	cmder.conn.SetReadDeadline(time.Now().Add(ReadTimeout))
	_, err := b.ReadNFrom(int64(n), cmder.rw.Reader)
	return err
}

func (cmder *Commander) read(n int) ([]byte, error) {
	cmder.conn.SetReadDeadline(time.Now().Add(ReadTimeout))
	recv := make([]byte, n)
	_, err := io.ReadFull(cmder.rw.Reader, recv)
	return recv, err
}

func (cmder *Commander) write(b *bytepool.Bytes) error {
	_, err := cmder.rw.Write(b.Bytes())
	return err
}
