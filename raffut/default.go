package raffut

import (
	"io"
	"log"
	"os"
)

const sampleRate = 44100
const chunkSize = sampleRate / 1
const udpChunkSize = 256

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
