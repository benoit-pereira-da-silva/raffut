package streams

import (
	"net"
)

func ReceiveUDP(s Streamable) (err error) {
	rAddr, err := net.ResolveUDPAddr("udp", s.Address())
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
	return (s).ReadStreamFrom(conn)
}

func SendUDP(s Streamable) (err error) {
	rAddr, err := net.ResolveUDPAddr("udp", s.Address())
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
	return (s).WriteStreamTo(conn)
}
