package streams

// Compressor defines the interface for compression providers
type Compressor interface {
	Compress(src []byte) (res []byte, err error)
	Decompress(src []byte) (res []byte, err error)
}

type NoCompressor struct {
	Compressor
}

func NewNoCompressor() *NoCompressor {
	return &NoCompressor{}
}

func (c *NoCompressor) Compress(src []byte) (res []byte, err error) {
	return src, nil
}

func (c *NoCompressor) Decompress(src []byte) (res []byte, err error) {
	return src, nil
}
