package miniaudio

import (
	"fmt"
	"github.com/benoit-pereira-da-silva/raffut/console"
	"github.com/benoit-pereira-da-silva/raffut/streams"
	"github.com/gen2brain/malgo"
	"io"
	"math"
	"os"
)

// Miniaudio Streamable support.
// Source: [Miniaudio](http://https://miniaud.io)
// "Miniaudio is an audio playback and capture library for C and C++.
// It's made up of a single source file, has no external dependencies and is released into the public domain."
// We use the embedded go bindings [malgo](https://github.com/gen2brain/malgo)
type Miniaudio struct {
	streams.Streamable
	address    string
	chunkSize  int
	sampleRate float64
	echo       bool
	done       chan interface{}
}

func (p *Miniaudio) ReadStreamFrom(c io.ReadWriteCloser) error {
	var reader io.Reader = c
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, func(message string) {
		fmt.Printf("LOG <%v>\n", message)
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer func() {
		_ = ctx.Uninit()
		ctx.Free()
	}()
	deviceConfig := malgo.DefaultDeviceConfig(malgo.Playback)
	deviceConfig.Playback.Format = malgo.FormatS16
	deviceConfig.Playback.Channels = 2
	deviceConfig.SampleRate = uint32(p.sampleRate)
	deviceConfig.Alsa.NoMMap = 2
	// This is the function that's used for sending more data to the device for playback.
	onSamples := func(pOutputSample, pInputSamples []byte, framecount uint32) {
		io.ReadFull(reader, pOutputSample)
		if p.echo {
			sum := float32(0)
			for _, v := range pOutputSample {
				sum += float32(math.MaxUint8 - v)
			}
			console.PrintFrame(sum)
		}
	}
	deviceCallbacks := malgo.DeviceCallbacks{
		Data: onSamples,
	}
	device, err := malgo.InitDevice(ctx.Context, deviceConfig, deviceCallbacks)
	if err != nil {
		return err
	}
	defer device.Uninit()
	err = device.Start()
	if err != nil {
		return err
	} else {
		for {
			select {
			case <-p.done:
				return nil
			}
		}
	}
}

// WriteStreamTo captures the frame using miniaudio.
// then write them to the stream.
func (p *Miniaudio) WriteStreamTo(c io.ReadWriteCloser) error {
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, func(message string) {
		fmt.Printf("LOG <%v>\n", message)
	})
	if err != nil {
		return err
	}
	defer func() {
		_ = ctx.Uninit()
		ctx.Free()
	}()
	deviceConfig := malgo.DefaultDeviceConfig(malgo.Capture)
	deviceConfig.Capture.Format = malgo.FormatS16
	deviceConfig.Capture.Channels = 2
	deviceConfig.SampleRate = uint32(p.SampleRate())
	deviceConfig.Alsa.NoMMap = 1
	onRecvFrames := func(pSample2, pSample []byte, framecount uint32) {
		_, err = c.Write(pSample)
		if err != nil {
			// After one write there is always an error
			// Explanation: https://stackoverflow.com/questions/46697799/golang-udp-connection-refused-on-every-other-write
			// " Because UDP has no real connection and there is no ACK for any packets sent,
			// the best a "connected" UDP socket can do to simulate a send failure is to save the ICMP response,
			// and return it as an error on the next write."
		} else {
			if p.echo {
				if err != nil {
					println(err.Error())
					<-p.done
				} else {
					sum := float32(0)
					for _, v := range pSample {
						sum += float32(math.MaxUint8 - v)
					}
					console.PrintFrame(sum)
				}
			}
		}
	}
	captureCallbacks := malgo.DeviceCallbacks{
		Data: onRecvFrames,
	}
	device, err := malgo.InitDevice(ctx.Context, deviceConfig, captureCallbacks)
	if err != nil {
		return err
	}
	err = device.Start()
	if err != nil {
		return err
	} else {
		defer device.Uninit()
		for {
			select {
			case <-p.done:
				return nil
			}
		}
	}
}

func (p *Miniaudio) Configure(address string, chunkSize int, sampleRate float64, echo bool, done chan interface{}) {
	p.address = address
	p.chunkSize = chunkSize
	p.sampleRate = sampleRate
	p.echo = echo
	p.done = done
}

// Address correspond to the <IP or Name:PORT>
func (p *Miniaudio) Address() string {
	return p.address
}

// ChunkSize is the size of the stream packet
func (p *Miniaudio) ChunkSize() int {
	return p.chunkSize
}

// SampleRate is the sample rate :)
func (p *Miniaudio) SampleRate() float64 {
	return p.sampleRate
}

// Echo if responding true prints the flow in the stdio
func (p *Miniaudio) Echo() bool {
	return p.echo
}

// Done is the cancellation channel
func (p *Miniaudio) Done() chan interface{} {
	return p.done
}
