package transport

import (
	"net"
)

func AnonymizeAddr(addr net.UDPAddr) net.UDPAddr {
	// If it is IPv4 make a []byte of 4 "segments" of an IP so they are all "0"
	// then copy only the first 3 "segments" leaving the last "segment" as a "0"
	if ipv4 := addr.IP.To4(); ipv4 != nil {
		buf := make([]byte, 4)
		copy(buf, ipv4[0:3])
		return net.UDPAddr{
			IP:   buf,
			Zone: addr.Zone,
		}
	}

	// Do the same just in case it is IPv6 for whatever reason
	if ipv6 := addr.IP.To16(); ipv6 != nil {
		buf := make([]byte, 16)
		copy(buf, ipv6[0:6])
		return net.UDPAddr{
			IP:   buf,
			Zone: addr.Zone,
		}
	}
	return net.UDPAddr{}
}
