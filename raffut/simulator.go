package raffut

import (
	"encoding/binary"
	"io"
	"math/rand"
	"time"
)

type Simulator struct {
	Console
	Echo bool
}

func (p *Simulator) WriteStreamTo(c io.ReadWriteCloser, chunkSize int, done chan interface{}) {
	st := time.Duration(chunkSize) * (time.Second / sampleRate)
	buffer := make([]float32, chunkSize)
	// Simulate an audio buffer for devices that does not have PortAudio.
	for {
		for idx := 0; idx < chunkSize; idx++ {
			buffer[idx] = rand.Float32()
		}
		binary.Write(c, binary.BigEndian, &buffer)
		if p.Echo {
			sum := float32(0)
			for i := range buffer {
				v := buffer[i]
				buffer[i] = v
				sum += v
			}
			printFrame(sum)
		}
		time.Sleep(st) // Sleep according to the sample rate?
	}
}
