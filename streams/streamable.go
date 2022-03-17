package streams

import "io"

// Streamable the interface to implement to be streamable over a network.
type Streamable interface {

	// Configure is the designated configuration method
	Configure(address string, sampleRate float64, nbChannels int, echo bool, done chan interface{})

	// ReadStreamFrom reads the chunks from the stream.
	ReadStreamFrom(c io.Reader) error

	// WriteStreamTo writes to the given connection stream
	WriteStreamTo(c io.Writer) error

	// Address correspond to the <IP or Name:PORT>
	Address() string

	// SampleRate is the sample rate :)
	SampleRate() float64

	// NbChannels stereo = 2
	NbChannels() int

	// Echo if responding true prints the flow in the stdio
	Echo() bool

	// Done is the cancellation channel
	Done() chan interface{}
}
