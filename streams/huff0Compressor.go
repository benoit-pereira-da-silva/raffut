package streams

import (
	"github.com/klauspost/compress/huff0"
)

// Huff0Compressor is a Huffman encoding and decoding as used in "zstd".
// "Huff0, a Huffman codec designed for modern CPU, featuring OoO (Out of Order) operations on multiple ALU (Arithmetic Logic Unit),
// achieving extremely fast compression and decompression speeds."
// ```
//	compressor := streams.NewHuff0Compressor()
// 	streamer := &miniaudio.Miniaudio{Format: malgo.FormatS16, Compressor: compressor}
// ```
// With a modern processor this compressor
// delivers an average lossless compression of 20% on PCM data
// with a latency < 0.5 milliseconds for encoding + decoding.
type Huff0Compressor struct {
	Compressor
	scratch *huff0.Scratch
}

func NewHuff0Compressor() *Huff0Compressor {
	return &Huff0Compressor{
		scratch: &huff0.Scratch{Reuse: huff0.ReusePolicyAllow},
	}
}

func (c *Huff0Compressor) Compress(src []byte) (res []byte, err error) {
	res, _, err = huff0.Compress1X(src, nil)
	return res, err
}

func (c *Huff0Compressor) Decompress(src []byte) (res []byte, err error) {
	var remain []byte
	c.scratch, remain, err = huff0.ReadTable(src, nil)
	if err != nil {
		return nil, err
	}
	return c.scratch.Decompress1X(remain)
}
