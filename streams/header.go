package streams

// HeaderLength  correspond to the compressed Header length
// Byte 0 to 3: the compressed bytes Length (Big Endian Uint32)
// Byte 4 to 11: the incremental frame identifier  (Big Endian Uint64)
// Byte 12 to 15: the crc32 checksum of the uncompressed payload (Big Endian Uint32)
const HeaderLength = 16 // 16 bits

// ReadHeader returns the Header's infos.
func ReadHeader(data []byte) (compressedLen uint32, framesId uint64, checksum uint32) {
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
