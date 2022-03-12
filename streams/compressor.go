package streams

import (
	"io"
)

// Compressor defines the interface for compression providers
type Compressor interface {
	Compress(src []byte, dst io.Writer) error
	Decompress(src io.Reader, dst io.Writer) error
}

type NoCompressor struct {
	Compressor
}

func (c *NoCompressor) Compress(src []byte, dst io.Writer) error {
	_, err := dst.Write(src)
	return err
}

func (c *NoCompressor) Decompress(src io.Reader, dst io.Writer) error {
	_, err := io.Copy(dst, src)
	return err
}
