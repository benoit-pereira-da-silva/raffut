package streams

import (
	"net"
)

func ReceiveUDP(address string) (err error) {
	rAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return err
	}
	conn, err := net.ListenUDP("udp", rAddr)
	if err != nil {
		return err
	}
	println("LocalAddr", conn.LocalAddr().String())
	if rAddr != nil {
		println("RemoteAddr", rAddr.String())
	}
	p := PortAudio{Echo: false, SampleRate: sampleRate}
	p.ReadStreamFrom(conn, udpChunkSize, nil)
	return nil
}

func SendUDP(address string) (err error) {
	rAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return err
	}
	conn, err := net.DialUDP("udp", nil, rAddr)
	if err != nil {
		return err
	}
	println("LocalAddr", conn.LocalAddr().String())
	if rAddr != nil {
		println("RemoteAddr", rAddr.String())
	}
	//p := Simulator{Echo: true}
	p := PortAudio{Echo: false, SampleRate: sampleRate}
	p.WriteStreamTo(conn, udpChunkSize, nil)
	return
}
