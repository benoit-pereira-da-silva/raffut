package streams

// Compressor defines the interface for compression providers
type Compressor interface {
	Compress(src []byte) (res []byte, err error)
	Decompress(src []byte) (res []byte, err error)
}
