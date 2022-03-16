package miniaudio

import (
	"github.com/benoit-pereira-da-silva/malgo"
	"github.com/benoit-pereira-da-silva/raffut/console"
	"github.com/benoit-pereira-da-silva/raffut/streams"
	"hash/crc32"
	"io"
	"math"
)

// headerLength  correspond to the compressed Header length
// Byte 0 to 3: the compressed bytes Length (Big Endian Uint32)
// Byte 4 to 11: the incremental frame identifier  (Big Endian Uint64)
// Byte 12 to 15: the crc32 checksum of the uncompressed payload (Big Endian Uint32)
const headerLength = 16 // 16 bits

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
	// This is the function that's used for sending more data to the device for playback.
	onSamples := func(out, in []byte, frameCount uint32) {
		if p.Compressor != nil {

			// Decode the Header.
			header := make([]byte, headerLength)
			_, _ = io.ReadAtLeast(c, header, headerLength)

			compressedDataLength := uint32(header[3]) | uint32(header[2])<<8 | uint32(header[1])<<16 | uint32(header[0])<<24
			framesId := uint64(header[11]) | uint64(header[10])<<8 | uint64(header[9])<<16 | uint64(header[8])<<24 |
				uint64(header[7])<<32 | uint64(header[6])<<40 | uint64(header[5])<<48 | uint64(header[4])<<56
			checksum := uint32(header[15]) | uint32(header[14])<<8 | uint32(header[13])<<16 | uint32(header[12])<<24

			// Then read all the data including the length header.
			bl := int(compressedDataLength) + headerLength
			allBytes := make([]byte, bl)
			_, _ = io.ReadFull(c, allBytes)

			// Decompress ignoring the length header.
			decompressed, decErr := p.Compressor.Decompress(allBytes[headerLength:])
			println("[ReadStreamFrom] framesId:", framesId, "checksum:", checksum, "=", crc32.ChecksumIEEE(decompressed), "compressedDataLength:", compressedDataLength, "len(allBytes):", len(allBytes), "len(decompressed):", len(decompressed), "allBytes.crc32", crc32.ChecksumIEEE(allBytes))

			if decErr != nil {
				// shall we log the decompression error?
				println("ERROR:", compressedDataLength, "->", len(allBytes), decErr.Error())
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

			header := make([]byte, headerLength)
			// Compressed bytes Length Header (Big Endian Uint32)
			header[0] = byte(cl >> 24)
			header[1] = byte(cl >> 16)
			header[2] = byte(cl >> 8)
			header[3] = byte(cl)
			// Incremental Frame Identifier (Big Endian encoded Uint64).
			header[4] = byte(framesId >> 56)
			header[5] = byte(framesId >> 48)
			header[6] = byte(framesId >> 40)
			header[7] = byte(framesId >> 32)
			header[8] = byte(framesId >> 24)
			header[9] = byte(framesId >> 16)
			header[10] = byte(framesId >> 8)
			header[11] = byte(framesId)
			// The data Checksum
			header[12] = byte(checksum >> 24)
			header[13] = byte(checksum >> 16)
			header[14] = byte(checksum >> 8)
			header[15] = byte(checksum)

			//println("Compression:",  len(b), "<-", len(in), elapsed.Microseconds())
			// Prepend the uint32 len header
			allBytes := append(header, compressed...)

			percent := (len(in) - len(allBytes)) * 100 / len(in)
			println("[WriteStreamTo] framesId:", framesId, "checksum", checksum, "cl:", cl, "len(allData):", len(allBytes), "len(in):", len(in), "%", percent, "allBytes.crc32:", crc32.ChecksumIEEE(allBytes))

			// Write the header and the compressed frames
			_, err = c.Write(allBytes)
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
