package raffut

import (
	"fmt"
	"github.com/gordonklaus/portaudio"
	"io"
	"math"
	"net"
)

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
				// After one write there is always an error.
				// Explanation from https://stackoverflow.com/questions/46697799/golang-udp-connection-refused-on-every-other-write
				// "Because UDP has no real connection and there is no ACK for any packets sent,
				// the best a "connected" UDP socket can do to simulate a "send" failure is to save the ICMP response,
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
