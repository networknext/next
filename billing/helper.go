package billing

import "net"

func udpAddrToAddress(addr net.UDPAddr) *Address {
	if addr.IP == nil {
		return &Address{
			Ip:        nil,
			Type:      Address_NONE,
			Port:      0,
			Formatted: "",
		}
	}

	ipv4 := addr.IP.To4()
	if ipv4 == nil {
		ipv6 := addr.IP.To16()
		if ipv6 == nil {
			return &Address{
				Ip:        nil,
				Type:      Address_NONE,
				Port:      0,
				Formatted: "",
			}
		}

		return &Address{
			Ip:        []byte(ipv6),
			Type:      Address_IPV6,
			Port:      uint32(addr.Port),
			Formatted: addr.String(),
		}
	}

	return &Address{
		Ip:        []byte(ipv4),
		Type:      Address_IPV4,
		Port:      uint32(addr.Port),
		Formatted: addr.String(),
	}
}
