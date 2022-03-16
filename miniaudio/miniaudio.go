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
	stack      *streams.CompressionStack
}

// ReadStreamFrom receive the stream
// plays the audio via miniaudio.
func (p *Miniaudio) ReadStreamFrom(c io.ReadCloser) error {
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
	// Instantiate the compression stack.
	p.stack = streams.NewCompressionStack(p.maxDataLength())

	// This is the function that's used for sending more data to the device for playback.
	onSamples := func(out, in []byte, frameCount uint32) {
		if p.Compressor != nil {
			_, _ = p.stack.FillWith(c)
			compressedDataLength, framesId, checksum := p.stack.ReadHeader()
			// Take a valid compressed slice including the header.
			currentBytes, _ := p.stack.Take(streams.HeaderLength + int(compressedDataLength))
			// Exclude the Header from Decompression.
			decompressed, decErr := p.Compressor.Decompress(currentBytes[streams.HeaderLength:])
			println("[ReadStreamFrom] framesId:", framesId, "checksum:", checksum, "=", crc32.ChecksumIEEE(decompressed), "compressedDataLength:", compressedDataLength, "len(currentBytes):", len(currentBytes), "len(decompressed):", len(decompressed), "currentBytes.crc32", crc32.ChecksumIEEE(currentBytes))

			if decErr != nil {
				// shall we log the decompression error?
				println("ERROR:", compressedDataLength, "->", len(currentBytes), decErr.Error())
			} else {
				copy(out, decompressed)
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
func (p *Miniaudio) WriteStreamTo(c io.WriteCloser) error {
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
			// ts := time.Now()
			compressed, _ := p.Compressor.Compress(in)
			// elapsed := time.Since(ts)
			// Compute the length header.
			cl := uint32(len(compressed))
			header := streams.EncodeCompressionHeader(cl, framesId, checksum)

			//println("Compression:",  len(b), "<-", len(in), elapsed.Microseconds())
			// Prepend the uint32 len header
			currentBytes := append(header, compressed...)

			percent := (len(in) - len(currentBytes)) * 100 / len(in)
			println("[WriteStreamTo] framesId:", framesId, "checksum", checksum, "cl:", cl, "len(currentBytes):", len(currentBytes), "len(in):", len(in), "%", percent, "currentBytes.crc32:", crc32.ChecksumIEEE(currentBytes))

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
