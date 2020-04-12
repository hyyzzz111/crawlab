package gob

import (
	"encoding/gob"
	"github.com/oxtoacart/bpool"
)

type Gob struct{}

// create buffer pool with 16 instances each preallocated with 256 bytes
var bufferPool = bpool.NewSizedBufferPool(16, 256)

func (g Gob) Marshal(i interface{}) ([]byte, error) {
	buf := bufferPool.Get()
	defer bufferPool.Put(buf)
	encoder := gob.NewEncoder(buf)
	err := encoder.Encode(i)
	return buf.Bytes(), err
}

func (g Gob) Unmarshal(b []byte, i interface{}) error {
	buf := bufferPool.Get()
	defer bufferPool.Put(buf)
	buf.Write(b)
	decoder := gob.NewDecoder(buf)
	return decoder.Decode(i)
}

func (g Gob) String() string {
	return "gob"
}
