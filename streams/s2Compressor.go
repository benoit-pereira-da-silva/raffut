package streams

import "github.com/klauspost/compress/s2"

// S2Compressor S2
type S2Compressor struct {
	Compressor
}

func NewS2Compressor() *S2Compressor {
	return &S2Compressor{}
}

func (c *S2Compressor) Compress(src []byte) (res []byte, err error) {
	res = s2.Encode(nil, src)
	return res, nil
}

func (c *S2Compressor) Decompress(src []byte) (res []byte, err error) {
	return s2.Decode(nil, src)
}
