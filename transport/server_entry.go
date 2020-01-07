package transport

import "net"

// ServerEntry is the struct for entries in the server database
type ServerEntry struct {
	address    *net.UDPAddr
	publicKey  []byte
	lastUpdate int64
}
