package streams

import (
	"github.com/klauspost/compress/fse"
)

// FSECompressor is a
// NOT FUNCTIONAL.
type FSECompressor struct {
	Compressor
	scratch *fse.Scratch
}

func NewFSECompressor() *FSECompressor {
	return &FSECompressor{}
}

func (c *FSECompressor) Compress(src []byte) (res []byte, err error) {
	return fse.Compress(src, c.scratch)
}

func (c *FSECompressor) Decompress(src []byte) (res []byte, err error) {
	return c.Decompress(src)
}
