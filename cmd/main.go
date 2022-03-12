package main

import (
	"fmt"
	"github.com/benoit-pereira-da-silva/malgo"
	"github.com/benoit-pereira-da-silva/raffut/console"
	"github.com/benoit-pereira-da-silva/raffut/miniaudio"
	"github.com/benoit-pereira-da-silva/raffut/streams"
	"log"
	"os"
	"strings"
)

const sampleRate = 44100

func main() {
	var err error
	if len(os.Args) > 2 {
		subCmd := strings.ToLower(os.Args[1])
		address := os.Args[2]

		switch subCmd {
		case "receive":
			// "raffut receive"192.168.1.4:8383"
			//streamer := &miniaudio.Miniaudio{Format: malgo.FormatS16, Compressor: &streams.NoCompressor{}}
			streamer := &miniaudio.Miniaudio{Format: malgo.FormatS16, Compressor: nil}
			streamer.Configure(address, sampleRate, 2, true, nil)
			err = streams.ReceiveUDP(streamer)
		case "send":
			// raffut send "192.168.1.4:8383"
			//streamer := &miniaudio.Miniaudio{Format: malgo.FormatS16, Compressor: &streams.NoCompressor{}}
			streamer := &miniaudio.Miniaudio{Format: malgo.FormatS16, Compressor: nil}
			streamer.Configure(address, sampleRate, 2, false, nil)
			err = streams.SendUDP(streamer)
		case "send-noise":
			// raffut send-noise "192.168.1.4:8383"
			// can be used on devices that does have audio support to test.
			streamer := &console.Console{ChunkSize: 256}
			streamer.Configure(address, sampleRate, 2, false, nil)
			streamer.Simulate = true
			err = streams.SendUDP(streamer)
		case "show-in-console":
			// raffut show-in-console "192.168.1.4:8383"
			// can be used to test the UDP connection visually without sound.
			streamer := &console.Console{ChunkSize: 256}
			streamer.Configure(address, sampleRate, 2, true, nil)
			err = streams.ReceiveUDP(streamer)
		default:
			err = fmt.Errorf("unsupported sub command \"%s\"", os.Args[1])
		}
	} else {
		err = fmt.Errorf("no sub command")
	}
	if err != nil {
		log.Println(err)
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}
