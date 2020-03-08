package gomemcached

import (
	"bytes"
	"sync"

	"github.com/vmihailenco/msgpack/v4"
)

type Encoder interface {
	Encode(v interface{}) ([]byte, error)
}

type Decoder interface {
	Decode(data []byte, v interface{}) error
}

type CodecEncoder struct {
	enc          *msgpack.Encoder
	encodeBuffer bytes.Buffer
}

func newEncoder() interface{} {
	encoder := &CodecEncoder{}
	encoder.enc = msgpack.NewEncoder(&encoder.encodeBuffer)
	return encoder
}

// return new slice from encodeBuffer

// encoder := GetEncoder()
// data, err := encoder.Encode(v)
// PutEncoder(encoder)
// use data....
//
// Dont do this, must put encoder to pool after data(new slice from encodeBuffer) use completed
// to avoid data conflict
func (encoder *CodecEncoder) Encode(v interface{}) ([]byte, error) {
	encoder.encodeBuffer.Reset()
	err := encoder.enc.Encode(v)
	if err != nil {
		return nil, err
	}

	return encoder.encodeBuffer.Bytes(), nil
}

type CodecDecoder struct {
	dec          *msgpack.Decoder
	decodeBuffer bytes.Buffer
}

func newDecoder() interface{} {
	decoder := &CodecDecoder{}
	decoder.dec = msgpack.NewDecoder(&decoder.decodeBuffer)
	return decoder
}

func (decoder *CodecDecoder) Decode(data []byte, v interface{}) error {
	decoder.decodeBuffer.Reset()
	_, err := decoder.decodeBuffer.Write(data)
	if err != nil {
		return err
	}

	err = decoder.dec.Decode(v)
	return err
}

var (
	encoderPool = sync.Pool{
		New: newEncoder,
	}
	decoderPool = sync.Pool{
		New: newDecoder,
	}
)

func getEncoder() Encoder {
	v := encoderPool.Get()
	return v.(Encoder)
}

func putEncoder(v Encoder) {
	encoderPool.Put(v)
}

func getDecoder() Decoder {
	v := decoderPool.Get()
	return v.(Decoder)
}

func putDecoder(v Decoder) {
	decoderPool.Put(v)
}
