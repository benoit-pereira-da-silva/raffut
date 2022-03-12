package streams

import (
	"github.com/klauspost/compress/s2"
	"io"
)

type S2Compressor struct {
	Compressor
}

func (c S2Compressor) Compress(src []byte, dst io.Writer) error {
	enc := s2.NewWriter(dst)
	// The encoder owns the buffer until Flush or Close is called.
	err := enc.EncodeBuffer(src)
	if err != nil {
		enc.Close()
		return err
	}
	// Blocks until compression is done.
	return enc.Close()
}

func (c S2Compressor) Decompress(src io.Reader, dst io.Writer) error {
	dec := s2.NewReader(src)
	_, err := io.Copy(dst, dec)
	return err
}
