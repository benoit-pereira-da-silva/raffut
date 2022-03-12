package streams

import (
	"github.com/klauspost/compress/s2"
	"io"
)

type S2Compressor struct {
	Compressor
}

func (c *S2Compressor) Compress(src []byte, dst io.Writer) error {
	enc := s2.NewWriter(dst)
	defer enc.Close()
	return enc.EncodeBuffer(src)
}

func (c *S2Compressor) Decompress(src io.Reader, dst io.Writer) error {
	dec := s2.NewReader(src)
	_, err := io.Copy(dst, dec)
	return err
}
