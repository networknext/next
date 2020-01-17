package routing

import "net"

type Server struct {
	Addr      net.UDPAddr
	PublicKey []byte
}
