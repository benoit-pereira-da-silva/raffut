package streams

import "net"

// MonitorUDPReception can be used on devices that does have audio support to test.
func MonitorUDPReception(address string) (err error) {
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
	c := Console{}
	c.print(conn, udpChunkSize, nil)
	return nil
}
