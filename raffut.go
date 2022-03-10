package main

import (
	"Raffut/raffut"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	if len(os.Args) > 2 {
		subCmd := strings.ToLower(os.Args[1])
		switch subCmd {
		case "send", "send-udp":
			// raffut send-udp "developer.home:8383"
			chk(raffut.SendUDP(os.Args[2]))
		case "receive", "receive-udp":
			// "raffut receive-udp "developer.home:8383"
			chk(raffut.ReceiveUDP(os.Args[2]))
		default:
			log.Println(fmt.Sprintf("unsupported sub command \"%s\"", os.Args[1]))
			os.Exit(1)
		}
	} else {
		log.Println("no sub command")
		os.Exit(1)
	}
	os.Exit(0)
}

func chk(err error) {
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
