package transport

import "net"

// RelayEntry is a struct for the entries into the relay database (game servers?)
type RelayEntry struct {
	id         uint64
	name       string
	address    *net.UDPAddr
	lastUpdate int64
	token      []byte
}
