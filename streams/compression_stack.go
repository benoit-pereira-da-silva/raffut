package streams

import (
	"fmt"
	"io"
)

// CompressionStack may be a sort of Ring Buffer (i m not sure).
// It should be filled by calling FillWith
// It can be read by calling ReadTo
// And more importantly when calling Take the taken bytes are ejected and the remaining bytes moved to the start.
type CompressionStack struct {
	data   []byte
	size   int
	cursor int
}

// NewCompressionStack is the designated CompressionStack constructor.
func NewCompressionStack(size int) *CompressionStack {
	return &CompressionStack{
		data:   make([]byte, size),
		size:   size,
		cursor: 0,
	}
}

// FillWith complement the data by reading from the io.Reader
func (r *CompressionStack) FillWith(reader io.Reader) (n int, err error) {
	lenToRead := r.size - r.cursor
	if lenToRead > 0 {
		complement := make([]byte, lenToRead)
		n, err = io.ReadFull(reader, complement)
		if err != nil {
			return n, err
		}
		fillIdx := 0
		for _, oneByte := range complement {
			r.data[fillIdx] = oneByte
			fillIdx++
		}
		r.cursor = 0
		return n, nil
	} else {
		return 0, nil
	}
}

/*
// ReadTo reads the data from 0 to the index.
// The underlining data are not modified
// can be used for inspection.
func (r *CompressionStack) ReadTo(index int) []byte {
	return r.data[:index]
}*/

// Take return a copy of the n first bytes
//
func (r *CompressionStack) Take(n int) (b []byte, err error) {
	if n > r.size {
		return b, fmt.Errorf("byte overflow")
	}
	// copy the taken bytes to the new response slice.
	// and copy the end of the slice to its beginning
	b = make([]byte, n)
	for idx, oneByte := range r.data {
		if idx < n {
			b[idx] = oneByte // copy the value to the response.
			r.data[idx] = 0  // reset data to zero for inspection.
		} else {
			if idx == n {
				r.cursor = 0
			}
			r.data[r.cursor] = oneByte
			r.cursor++
		}
	}
	return b, nil
}

// HeaderLength  correspond to the compressed Header length
// Byte 0 to 3: the compressed bytes Length (Big Endian Uint32)
// Byte 4 to 11: the incremental frame identifier  (Big Endian Uint64)
// Byte 12 to 15: the crc32 checksum of the uncompressed payload (Big Endian Uint32)
const HeaderLength = 16 // 16 bits

// ReadHeader returns the Header's infos.
func (r *CompressionStack) ReadHeader() (compressedLen uint32, framesId uint64, checksum uint32) {
	data := r.data
	compressedLen = uint32(data[3]) |
		uint32(data[2])<<8 |
		uint32(data[1])<<16 |
		uint32(data[0])<<24
	framesId = uint64(data[11]) |
		uint64(data[10])<<8 |
		uint64(data[9])<<16 |
		uint64(data[8])<<24 |
		uint64(data[7])<<32 |
		uint64(data[6])<<40 |
		uint64(data[5])<<48 |
		uint64(data[4])<<56
	checksum = uint32(data[15]) |
		uint32(data[14])<<8 |
		uint32(data[13])<<16 |
		uint32(data[12])<<24
	return compressedLen, framesId, checksum
}

// EncodeCompressionHeader is a global function that does not require a CompressionStack
func EncodeCompressionHeader(compressedLen uint32, framesId uint64, checksum uint32) []byte {
	header := make([]byte, HeaderLength)
	// Compressed bytes Length Header (Big Endian Uint32)
	header[0] = byte(compressedLen >> 24)
	header[1] = byte(compressedLen >> 16)
	header[2] = byte(compressedLen >> 8)
	header[3] = byte(compressedLen)
	// Incremental Frame Identifier (Big Endian encoded Uint64).
	header[4] = byte(framesId >> 56)
	header[5] = byte(framesId >> 48)
	header[6] = byte(framesId >> 40)
	header[7] = byte(framesId >> 32)
	header[8] = byte(framesId >> 24)
	header[9] = byte(framesId >> 16)
	header[10] = byte(framesId >> 8)
	header[11] = byte(framesId)
	// The data Checksum
	header[12] = byte(checksum >> 24)
	header[13] = byte(checksum >> 16)
	header[14] = byte(checksum >> 8)
	header[15] = byte(checksum)
	return header
}
