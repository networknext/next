package routing

import (
	"net"

	"github.com/networknext/backend/crypto"
)

const (
	TokenSize = crypto.KeySize + crypto.MACSize
)

type Relay struct {
	ID uint64

	Addr      net.UDPAddr
	PublicKey []byte

	Latitude  float64
	Longitude float64

	RTT        float64
	Jitter     float64
	PacketLoss float64
}
