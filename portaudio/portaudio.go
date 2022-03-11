package portaudio

import (
	"github.com/benoit-pereira-da-silva/raffut/console"
	"github.com/benoit-pereira-da-silva/raffut/streams"
	"github.com/gordonklaus/portaudio"
	"io"
)

// PortAudio Streamable support.
// Source: [Portaudio](http://www.portaudio.com)
// "PortAudio is a free, cross-platform, open-source, audio I/O library.
// It lets you write simple audio programs in 'C' or C++ that will compile and run on many platforms including Windows, Macintosh OS X, and Unix (OSS/ALSA).
// It is intended to promote the exchange of audio software between developers on different platforms.
// Many applications use PortAudio for Audio I/O."
type PortAudio struct {
	streams.Streamable
	address    string
	chunkSize  int
	sampleRate float64
	echo       bool
	done       chan interface{}
}

func (p *PortAudio) ReadStreamFrom(c io.ReadWriteCloser) error {
	portaudio.Initialize()
	defer portaudio.Terminate()
	bs := make([]byte, p.chunkSize*4)
	floatBuffer := make([]float32, len(bs)/4)
	stream, err := portaudio.OpenDefaultStream(0, 1, p.sampleRate, p.chunkSize, func(out []float32) {
		_, err := c.Read(bs)
		if err != nil {
			println(err.Error())
			<-p.done
		} else {
			err = streams.BigEndianBytesToFloat32(bs, &floatBuffer)
			if err != nil {
				println(err.Error())
				<-p.done
			} else {
				sum := float32(0)
				for i := range out {
					v := floatBuffer[i]
					out[i] = v
					sum += v
				}
				if p.echo {
					console.PrintFrame(sum)
				}
			}
		}
	})
	streams.Chk(err)
	streams.Chk(stream.Start())
	defer stream.Close()
	for {
		select {
		case <-p.done:
			streams.Chk(stream.Stop())
			return nil
		}
	}
}

func (p *PortAudio) WriteStreamTo(c io.ReadWriteCloser) error {
	buffer := make([]float32, p.chunkSize)
	byteBuffer := make([]byte, len(buffer)*4)
	portaudio.Initialize()
	defer portaudio.Terminate()
	stream, err := portaudio.OpenDefaultStream(1, 0, p.sampleRate, p.chunkSize, func(in []float32) {
		sum := float32(0)
		for i := range buffer {
			v := in[i]
			buffer[i] = v
			sum += v
		}
		err := streams.BigEndianFloat32ToBytes(buffer, &byteBuffer)
		if err != nil {
			println(err.Error())
			<-p.done
		} else {
			_, err = c.Write(byteBuffer)
			if err != nil {
				// After one write there is always an error
				// Explanation: https://stackoverflow.com/questions/46697799/golang-udp-connection-refused-on-every-other-write
				// " Because UDP has no real connection and there is no ACK for any packets sent,
				// the best a "connected" UDP socket can do to simulate a send failure is to save the ICMP response,
				// and return it as an error on the next write."
			} else {
				if p.echo {
					console.PrintFrame(sum)
				}
			}
		}

	})
	streams.Chk(err)
	streams.Chk(stream.Start())
	defer stream.Close()
	for {
		select {
		case <-p.done:
			streams.Chk(stream.Stop())
			return nil
		}
	}
}

func (p *PortAudio) Configure(address string, chunkSize int, sampleRate float64, echo bool, done chan interface{}) {
	p.address = address
	p.chunkSize = chunkSize
	p.sampleRate = sampleRate
	p.echo = echo
	p.done = done
}

// Address correspond to the <IP or Name:PORT>
func (p *PortAudio) Address() string {
	return p.address
}

// ChunkSize is the size of the stream packet
func (p *PortAudio) ChunkSize() int {
	return p.chunkSize
}

// SampleRate is the sample rate :)
func (p *PortAudio) SampleRate() float64 {
	return p.sampleRate
}

// Echo if responding true prints the flow in the stdio
func (p *PortAudio) Echo() bool {
	return p.echo
}

// Done is the cancellation channel
func (p *PortAudio) Done() chan interface{} {
	return p.done
}
