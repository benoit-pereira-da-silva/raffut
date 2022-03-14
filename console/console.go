package console

import (
	"encoding/binary"
	"fmt"
	"github.com/benoit-pereira-da-silva/raffut/streams"
	"io"
	"math/rand"
	"time"
)

// A bunch of function to print frames and evaluate silence detection algorithms
var frameMax = 0

const fbLen = 100

var frameBuffer = make([]int, fbLen)
var precision = 1000
var silenceThreshold = 1
var sIdx = 0

type Console struct {
	streams.Streamable
	address    string
	ChunkSize  int
	sampleRate float64
	nbChannels int
	echo       bool
	done       chan interface{}
	Simulate   bool // If set to true we generate random noise.
}

func (p *Console) ReadStreamFrom(c io.ReadCloser) error {
	p.print(c)
	return nil
}

func (p *Console) WriteStreamTo(c io.WriteCloser) error {
	p.echo = p.Simulate
	if p.Simulate {
		// Simulation mode
		st := time.Duration(p.ChunkSize) * (time.Second / time.Duration(p.sampleRate))
		buffer := make([]float32, p.ChunkSize)
		for {
			for idx := 0; idx < p.ChunkSize; idx++ {
				buffer[idx] = rand.Float32()
			}
			binary.Write(c, binary.BigEndian, &buffer)
			if p.echo {
				sum := float32(0)
				for i := range buffer {
					v := buffer[i]
					buffer[i] = v
					sum += v
				}
				PrintFrame(sum)
			}
			time.Sleep(st) // Sleep according to the sample rate
		}
		return nil
	} else {
		//
		//p.print(c)
	}
	return nil
}

func (p *Console) print(c io.ReadCloser) {
	buffer := make([]float32, p.ChunkSize)
	for {
		select {
		case <-p.done:
			return
		default:
			err := binary.Read(c, binary.BigEndian, &buffer)
			if err != nil {
				println(err.Error())
			} else {
				sum := float32(0)
				for i := range buffer {
					sum += buffer[i]
				}
				PrintFrame(sum)
			}
		}
	}
}

func PrintFrame(f float32) {
	green := "\033[32m"
	displayLen := 50
	v := int(f * 1)
	if v < 0 {
		// abs
		v = -v
	}
	fillFrameBuffer(v)
	fm := vMax(v)
	threshold := 0
	if fm != 0 {
		threshold = v * displayLen / fm
	}
	s := make([]rune, displayLen)
	idx := 0
	for i := 0; i < displayLen; i++ {
		if i <= threshold {
			s[idx] = '='
		} else {
			s[idx] = ' '
		}
		idx++
	}
	silence := ""

	if isSilent(silenceThreshold) {
		silence = "Silence"
	}
	str := fmt.Sprintf("%s%s [%04d<%04d] %s", green, string(s), v, fm, silence)
	println(str)
}

func fillFrameBuffer(v int) {
	// Fill the buffer.
	frameBuffer[sIdx] = v * precision
	sIdx++
	if sIdx == fbLen {
		sIdx = 0
	}
}

func isSilent(threshold int) bool {
	for _, sv := range frameBuffer {
		if sv > threshold {
			return false
		}
	}
	return true
}

func vMax(v int) int {
	if v > frameMax {
		frameMax = v
	}
	return frameMax
}

func (p *Console) Configure(address string, sampleRate float64, nbChannels int, echo bool, done chan interface{}) {
	p.address = address
	p.nbChannels = nbChannels
	p.sampleRate = sampleRate
	p.echo = echo
	p.done = done
}

// Address correspond to the <IP or Name:PORT>
func (p *Console) Address() string {
	return p.address
}

// SampleRate is the sample rate :)
func (p *Console) SampleRate() float64 {
	return p.sampleRate
}

// NbChannels stereo = 2
func (p *Console) NbChannels() int {
	return p.nbChannels
}

// Echo if responding true prints the flow in the stdio
func (p *Console) Echo() bool {
	return p.echo
}

// Done is the cancellation channel
func (p *Console) Done() chan interface{} {
	return p.done
}
