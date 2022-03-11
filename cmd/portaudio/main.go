package main

import (
	"fmt"
	"github.com/benoit-pereira-da-silva/raffut/portaudio"
	"github.com/benoit-pereira-da-silva/raffut/streams"
	"log"
	"os"
	"strings"
)

const sampleRate = 44100
const udpChunkSize = 256

func main() {
	var err error
	if len(os.Args) > 2 {
		subCmd := strings.ToLower(os.Args[1])
		address := os.Args[2]
		streamer := &portaudio.PortAudio{}
		streamer.Configure(address, udpChunkSize, sampleRate, false, nil)
		switch subCmd {
		case "receive":
			// "raffut receive "192.168.1.4:8383"
			err = streams.ReceiveUDP(streamer)
		case "send":
			// raffut send "192.168.1.4:8383"
			err = streams.SendUDP(streamer)
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
