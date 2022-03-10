package streams

import (
	"github.com/gordonklaus/portaudio"
	"io"
	"net"
)

// PortAudio Streamable support.
type PortAudio struct {
	Streamable
	SampleRate float64
	Echo       bool
}

func (p *PortAudio) ReadStreamFrom(c *net.UDPConn, chunkSize int, done chan interface{}) {
	portaudio.Initialize()
	defer portaudio.Terminate()
	bs := make([]byte, chunkSize*4)
	floatBuffer := make([]float32, len(bs)/4)
	stream, err := portaudio.OpenDefaultStream(0, 1, p.SampleRate, chunkSize, func(out []float32) {
		_, err := c.Read(bs)
		if err != nil {
			println(err.Error())
			<-done
		} else {
			err = bigEndianBytesToFloat32(bs, &floatBuffer)
			if err != nil {
				println(err.Error())
				<-done
			} else {
				sum := float32(0)
				for i := range out {
					v := floatBuffer[i]
					out[i] = v
					sum += v
				}
				if p.Echo {
					printFrame(sum)
				}
			}
		}
	})
	chk(err)
	chk(stream.Start())
	defer stream.Close()
	for {
		select {
		case <-done:
			chk(stream.Stop())
			return
		}
	}
}

func (p *PortAudio) WriteStreamTo(c io.ReadWriteCloser, chunkSize int, done chan interface{}) {
	buffer := make([]float32, chunkSize)
	byteBuffer := make([]byte, len(buffer)*4)
	portaudio.Initialize()
	defer portaudio.Terminate()
	stream, err := portaudio.OpenDefaultStream(1, 0, p.SampleRate, chunkSize, func(in []float32) {
		sum := float32(0)
		for i := range buffer {
			v := in[i]
			buffer[i] = v
			sum += v
		}
		err := bigEndianFloat32ToBytes(buffer, &byteBuffer)
		if err != nil {
			println(err.Error())
			<-done
		} else {
			_, err = c.Write(byteBuffer)
			if err != nil {
				// After one write there is always an error
				// Explanation: https://stackoverflow.com/questions/46697799/golang-udp-connection-refused-on-every-other-write
				// " Because UDP has no real connection and there is no ACK for any packets sent,
				// the best a "connected" UDP socket can do to simulate a send failure is to save the ICMP response,
				// and return it as an error on the next write."
			} else {
				if p.Echo {
					printFrame(sum)
				}
			}
		}

	})
	chk(err)
	chk(stream.Start())
	defer stream.Close()
	for {
		select {
		case <-done:
			chk(stream.Stop())
			return
		}
	}
}
