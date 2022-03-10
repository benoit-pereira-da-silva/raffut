package main

import (
	"fmt"
	"github.com/benoit-pereira-da-silva/raffut/streams"
	"log"
	"os"
	"strings"
)

func main() {
	var err error
	if len(os.Args) > 2 {
		subCmd := strings.ToLower(os.Args[1])
		switch subCmd {
		case "monitor":
			// "raffut monitor "192.168.1.4:8383"
			// can be used on devices that does have audio support to test.
			err = streams.MonitorUDPReception(os.Args[2])
		case "receive", "receive-udp":
			// "raffut receive-udp "192.168.1.4:8383"
			err = streams.ReceiveUDP(os.Args[2])
		case "send", "send-udp":
			// raffut send-udp "192.168.1.4:8383"
			err = streams.SendUDP(os.Args[2])
		case "simulate":
			// "raffut simulate "192.168.1.4:8383"
			// can be used on devices that does have audio support to test.
			err = streams.SimulateUDPSending(os.Args[2])
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
