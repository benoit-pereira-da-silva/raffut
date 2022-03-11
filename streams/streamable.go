package streams

import "io"

// Streamable the interface to implement to be streamable over a network.
type Streamable interface {

	// Configure is the designated configuration method
	Configure(address string, chunkSize int, sampleRate float64, echo bool, done chan interface{})

	// ReadStreamFrom reads the chunks from the stream.
	ReadStreamFrom(c io.ReadWriteCloser) error

	// WriteStreamTo writes to the given connection stream
	WriteStreamTo(c io.ReadWriteCloser) error

	// Address correspond to the <IP or Name:PORT>
	Address() string

	// ChunkSize is the size of the stream packet
	ChunkSize() int

	// SampleRate is the sample rate :)
	SampleRate() float64

	// Echo if responding true prints the flow in the stdio
	Echo() bool

	// Done is the cancellation channel
	Done() chan interface{}
}
