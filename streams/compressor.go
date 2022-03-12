package streams

import "io"

// Compressor defines the interface for compression providers
type Compressor interface {
	Compress(src []byte, dst io.Writer) error
	Decompress(src io.Reader, dst io.Writer) error
}
