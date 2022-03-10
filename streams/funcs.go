package streams

import (
	"fmt"
	"io"
	"log"
	"math"
	"os"
)

const sampleRate = 44100
const udpChunkSize = 256

// Streamable the interface to implement to be streamable over a network.
type Streamable interface {

	// ReadStreamFrom reads the chunks from the stream.
	ReadStreamFrom(c io.ReadWriteCloser, chunkSize int, done chan interface{})

	// WriteStreamTo writes to the given connection stream
	WriteStreamTo(c io.ReadWriteCloser, chunkSize int, done chan interface{})
}

func chk(err error) {
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

// bigEndianFloat32ToBytes should be faster than binary.Write(c, binary.BigEndian, &buffer)
// It does not rely on reflexion.
// when dealing with sound faster is always better.
func bigEndianFloat32ToBytes(data []float32, result *[]byte) error {
	if len(data) != len(*result)/4 {
		return fmt.Errorf("length missmatch in bigEndianFloat32ToBytes []float32 len should be equal to []byte len / 4")
	}
	for i, x := range data {
		v := math.Float32bits(x)
		(*result)[4*i] = byte(v >> 24)
		(*result)[4*i+1] = byte(v >> 16)
		(*result)[4*i+2] = byte(v >> 8)
		(*result)[4*i+3] = byte(v)
	}
	return nil
}

// bigEndianBytesToFloat32 should be faster than binary.Read(c, binary.BigEndian, &buffer)
// It does not rely on reflexion.
func bigEndianBytesToFloat32(data []byte, result *[]float32) error {
	if len(data)/4 != len(*result) {
		return fmt.Errorf("length missmatch in bigEndianBytesToFloat32 []float32 len should be equal to []byte len / 4")
	}
	for i, _ := range *result {
		v := uint32(data[4*i+3]) | uint32(data[4*i+2])<<8 | uint32(data[4*i+1])<<16 | uint32(data[4*i])<<24
		(*result)[i] = math.Float32frombits(v)
	}
	return nil
}
