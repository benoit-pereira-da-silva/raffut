package main

import (
	"fmt"
	"github.com/benoit-pereira-da-silva/raffut/console"
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
		streamer := &console.Console{}
		streamer.Configure(address, udpChunkSize, sampleRate, false, nil)
		switch subCmd {
		case "receive":
			// "raffut receive-udp "192.168.1.4:8383"
			err = streams.ReceiveUDP(streamer)
		case "send":
			// raffut send-udp "192.168.1.4:8383"
			// can be used on devices that does have audio support to test.
			streamer.Simulate = true
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
