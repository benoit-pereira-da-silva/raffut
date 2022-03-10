package streams

import (
	"encoding/binary"
	"fmt"
	"io"
)

// A bunch of function to print frames and evaluate silence detection algorithms
var frameMax = 0

const fbLen = 100

var frameBuffer = make([]int, fbLen)
var precision = 1000
var silenceThreshold = 1
var sIdx = 0

type Console struct {
	Streamable
}

func (p *Console) ReadStreamFrom(c io.ReadWriteCloser, chunkSize int, done chan interface{}) {
	p.print(c, chunkSize, done)
}

func (p *Console) WriteStreamTo(c io.ReadWriteCloser, chunkSize int, done chan interface{}) {
	p.print(c, chunkSize, done)
}

func (p *Console) print(c io.ReadWriteCloser, chunkSize int, done chan interface{}) {
	buffer := make([]float32, chunkSize)
	for {
		select {
		case <-done:
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
				printFrame(sum)
			}

		}
	}
}

func printFrame(f float32) {
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
