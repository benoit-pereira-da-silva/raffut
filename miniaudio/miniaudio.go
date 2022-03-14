package miniaudio

import (
	"encoding/binary"
	"github.com/benoit-pereira-da-silva/malgo"
	"github.com/benoit-pereira-da-silva/raffut/console"
	"github.com/benoit-pereira-da-silva/raffut/streams"
	"io"
)

// Miniaudio Streamable support.
// Source: [Miniaudio](https://miniaud.io)
// "Miniaudio is an audio playback and capture library for C and C++.
// It's made up of a single source file, has no external dependencies and is released into the public domain."
// We use the embedded go bindings [malgo](https://github.com/gen2brain/malgo)
//	+---------------+----------------------------------------+---------------------------+
//	| ma_format_f32 | 32-bit floating point                  | [-1, 1]                   |
//	| ma_format_s16 | 16-bit signed integer                  | [-32768, 32767]           |
//	| ma_format_s24 | 24-bit signed integer (tightly packed) | [-8388608, 8388607]       |
//	| ma_format_s32 | 32-bit signed integer                  | [-2147483648, 2147483647] |
//	| ma_format_u8  | 8-bit unsigned integer                 | [0, 255]                  |
//	+---------------+----------------------------------------+---------------------------+
type Miniaudio struct {
	streams.Streamable
	address    string
	sampleRate float64          // Default 441000hz
	nbChannels int              // Default Stereo = 2 two channels
	Format     malgo.FormatType // Default malgo.FormatS16
	// If defined each packet is Compressed / Decompressed.
	Compressor streams.Compressor
	echo       bool
	done       chan interface{}
}

// ReadStreamFrom receive the stream
// plays the audio via miniaudio.
func (p *Miniaudio) ReadStreamFrom(c io.ReadCloser) error {
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, func(message string) {
		println(message)
	})
	if err != nil {
		return err
	}
	defer func() {
		_ = ctx.Uninit()
		ctx.Free()
	}()
	deviceConfig := malgo.DefaultDeviceConfig(malgo.Playback)
	deviceConfig.Playback.Format = p.Format
	deviceConfig.Playback.Channels = uint32(p.nbChannels)
	deviceConfig.SampleRate = uint32(p.sampleRate)
	deviceConfig.Alsa.NoMMap = 2
	// This is the function that's used for sending more data to the device for playback.
	onSamples := func(out, in []byte, frameCount uint32) {
		if p.Compressor != nil {
			// Read the length header
			lenHeader := make([]byte, 4)
			_, _ = io.ReadAtLeast(c, lenHeader, 4)
			cl := binary.BigEndian.Uint32(lenHeader)
			// Then read the data.
			bl := int(cl) + 4
			var b = make([]byte, bl)
			_, err = io.ReadFull(c, b)
			// Decompress ignoring the length header.
			b = b[4:]
			b, err = p.Compressor.Decompress(b)
			if err != nil {
				// shall we log the decompression error?
			} else {
				copy(out, b)
			}
		} else {
			io.ReadFull(c, out)
		}
		if p.echo {
			p.printBytes(out)
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

// WriteStreamTo captures the audio in using miniaudio.
// then send them to the stream.
func (p *Miniaudio) WriteStreamTo(c io.WriteCloser) error {
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, func(message string) {
		println(message)
	})
	if err != nil {
		return err
	}
	defer func() {
		_ = ctx.Uninit()
		ctx.Free()
	}()
	deviceConfig := malgo.DefaultDeviceConfig(malgo.Capture)
	deviceConfig.Capture.Format = p.Format
	deviceConfig.Capture.Channels = uint32(p.nbChannels)
	deviceConfig.SampleRate = uint32(p.SampleRate())
	deviceConfig.Alsa.NoMMap = 1
	onRecFrames := func(out, in []byte, frameCount uint32) {
		if p.Compressor != nil {
			// ts := time.Now()
			b, _ := p.Compressor.Compress(in)
			// elapsed := time.Since(ts)
			// Compute the length header.
			cl := uint32(len(b))
			lenHeader := make([]byte, 4)
			binary.BigEndian.PutUint32(lenHeader, cl)
			//println("Compression:",  len(b), "<-", len(in), elapsed.Microseconds())
			// Prepend the uint32 len header
			b = append(lenHeader, b...)
			// Write the length header and the compressed frames
			_, err = c.Write(b)
		} else {
			_, err = c.Write(in)
		}
		if err != nil {
			// After one write there is always an error
			// Explanation: https://stackoverflow.com/questions/46697799/golang-udp-connection-refused-on-every-other-write
			// " Because UDP has no real connection and there is no ACK for any packets sent,
			// the best a "connected" UDP socket can do to simulate a failure is to save the ICMP response,
			// and return it as an error on the next write."
		} else {
			if p.echo {
				p.printBytes(in)
			}
		}
	}
	captureCallbacks := malgo.DeviceCallbacks{
		Data: onRecFrames,
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

func (p *Miniaudio) printBytes(bsl []byte) {
	sum := float32(0)
	b := make([]byte, 4)
	rx := 0
	for idx, v := range bsl {
		b[rx] = v
		rx += 1
		if rx == 4 {
			u := uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24 // LittleEndian
			sum += float32(u)
			rx = 0
		}
		// 441 is divisible by : 1, 3, 7, 9, 21, 49, 63, 147, 441.
		if idx%(147) == 0 {
			console.PrintFrame(sum)
			sum = 0
		}
	}
}

func (p *Miniaudio) Configure(address string, sampleRate float64, nbChannels int, echo bool, done chan interface{}) {
	p.address = address
	p.sampleRate = sampleRate
	p.nbChannels = nbChannels
	p.echo = echo
	p.done = done
}

// Address correspond to the <IP or Name:PORT>
func (p *Miniaudio) Address() string {
	return p.address
}

// SampleRate is the sample rate :)
func (p *Miniaudio) SampleRate() float64 {
	return p.sampleRate
}

// NbChannels stereo = 2
func (p *Miniaudio) NbChannels() int {
	return p.nbChannels
}

// Echo if responding true prints the flow in the stdio
func (p *Miniaudio) Echo() bool {
	return p.echo
}

// Done is the cancellation channel
func (p *Miniaudio) Done() chan interface{} {
	return p.done
}
