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
	cmder.server.badCmders <- cmder
}

func (cmder *Commander) flush() error {
	cmder.conn.SetWriteDeadline(time.Now().Add(WriterTimeout))
	return cmder.rw.Flush()
}

func (cmder *Commander) readN(b *bytepool.Bytes, n uint32) error {
	if n <= 0 {
		return ErrInvalidArguments
	}

	_, err := b.ReadNFrom(int64(n), cmder.rw)
	return err
}

func (cmder *Commander) read(b *bytepool.Bytes) error {
	cmder.conn.SetReadDeadline(time.Now().Add(ReadTimeout))
	_, err := io.ReadFull(cmder.rw, b.Bytes())
	return err
}

func (cmder *Commander) write(b *bytepool.Bytes) error {
	_, err := cmder.rw.Write(b.Bytes())
	return err
}
