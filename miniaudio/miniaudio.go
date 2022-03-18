package miniaudio

import (
	"github.com/benoit-pereira-da-silva/malgo"
	"github.com/benoit-pereira-da-silva/raffut/console"
	"github.com/benoit-pereira-da-silva/raffut/streams"
	"hash/crc32"
	"io"
	"math"
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
	echo       bool
	done       chan interface{}
	// If defined each packet is Compressed / Decompressed.
	Compressor streams.Compressor
}

// ReadStreamFrom receive the stream
// plays the audio via miniaudio.
func (p *Miniaudio) ReadStreamFrom(c io.Reader) error {
	ctx, initErr := malgo.InitContext(nil, malgo.ContextConfig{}, func(message string) {
		println(message)
	})
	if initErr != nil {
		return initErr
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

	data := make([]byte, p.maxDataLength())
	//compression := 0
	//counter := 0
	// This is the function that's used for sending more data to the device for playback.
	onSamples := func(out, in []byte, frameCount uint32) {
		if p.Compressor != nil {
			// It is very important to understand that the underlining UDP connection at a given time returns a packet.
			// Reading twice would read the next packet.
			// So we use the number of reading bytes to create a new slice (currentBytes).
			n, _ := c.Read(data)
			if n == 16 {
				// Perfect silence.
			} else {
				compressedDataLength, _, _ := streams.ReadHeader(data)
				compressedSegment := data[streams.HeaderLength : compressedDataLength+streams.HeaderLength]
				decompressed, decErr := p.Compressor.Decompress(compressedSegment)
				//compression += 100 - (100 * (int(compressedDataLength) + streams.HeaderLength) / p.maxDataLength())
				//counter++
				//if counter == 50 {
				//	println(fmt.Sprintf("Lossless compression: %d%s", compression/counter, "%"))
				//	counter = 0
				//	compression = 0
				//}
				if decErr != nil {
					// shall we log the decompression error?
					println("Compressor did trigger an error on Decompress:", decErr.Error())
				} else {
					// Copy the PCM data to the audio output
					copy(out, decompressed)
				}
			}
		} else {
			_, err := io.ReadFull(c, out)
			if err != nil {
				println(err.Error())
			}
		}
		if p.echo {
			p.printBytes(out)
		}
	}
	deviceCallbacks := malgo.DeviceCallbacks{
		Data: onSamples,
	}
	device, initDevErr := malgo.InitDevice(ctx.Context, deviceConfig, deviceCallbacks)
	if initDevErr != nil {
		return initDevErr
	}
	defer device.Uninit()
	startErr := device.Start()
	if startErr != nil {
		return startErr
	} else {
		for {
			select {
			case <-p.done:
				return nil
			}
		}
	}
}

// maxDataLength correspond to the uncompressed length including the compression header.
// the division 100 is empirically deduced from tests.
// We are not sure that it is constant.
func (p *Miniaudio) maxDataLength() int {
	sampleSize := 0
	switch p.Format {
	case malgo.FormatU8:
		sampleSize = 1
	case malgo.FormatS16:
		sampleSize = 2
	case malgo.FormatS24:
		sampleSize = 3
	case malgo.FormatS32, malgo.FormatF32:
		sampleSize = 4
	}
	return (int(p.sampleRate) / 100 * p.nbChannels * sampleSize) + streams.HeaderLength
}

// WriteStreamTo captures the audio in using miniaudio.
// then send them to the stream.
func (p *Miniaudio) WriteStreamTo(c io.Writer) error {
	ctx, initErr := malgo.InitContext(nil, malgo.ContextConfig{}, func(message string) {
		println(message)
	})
	if initErr != nil {
		return initErr
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
	framesId := uint64(0)
	onRecFrames := func(out, in []byte, frameCount uint32) {
		var err error
		if p.Compressor != nil {
			checksum := crc32.ChecksumIEEE(in)
			compressed, _ := p.Compressor.Compress(in)
			cl := uint32(len(compressed))
			header := streams.EncodeCompressionHeader(cl, framesId, checksum)
			// Prepend the uint32 len header
			currentBytes := append(header, compressed...)
			// Write the header and the compressed frames
			_, err = c.Write(currentBytes)
		} else {
			_, err = c.Write(in)
		}
		if err != nil {
			// UDP has no real connection and no Acknowledgement on any packet transmission.
			// If there is no receiver c.Write you get a "connection refused"
			// This is not always the case.
			println("ERROR:", err.Error())
		} else {
			if p.echo {
				p.printBytes(in)
			}
		}
		if framesId == math.MaxUint64 {
			// Reinitialize the framesId
			framesId = 0
		} else {
			framesId++
		}
	}
	captureCallbacks := malgo.DeviceCallbacks{
		Data: onRecFrames,
	}
	device, initDevErr := malgo.InitDevice(ctx.Context, deviceConfig, captureCallbacks)
	if initDevErr != nil {
		return initDevErr
	}
	startErr := device.Start()
	if startErr != nil {
		return startErr
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
