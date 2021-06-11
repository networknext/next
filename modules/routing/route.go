package routing

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"net"
	"strings"

	"github.com/networknext/backend/modules/core"
)

// todo: is this still used?
type Route struct {
	NumRelays       int                                 `json:"numRelays"`
	RelayIDs        [core.MaxRelaysPerRoute]uint64      `json:"relayIDs"`
	RelayNames      [core.MaxRelaysPerRoute]string      `json:"relayNames"`
	RelayAddrs      [core.MaxRelaysPerRoute]net.UDPAddr `json:"relayAddrs"`
	RelayPublicKeys [core.MaxRelaysPerRoute][]byte      `json:"relayPublicKeys"`
	RelaySellers    [core.MaxRelaysPerRoute]Seller      `json:"relaySellers"`
	Stats           Stats                               `json:"stats"`
}

func (r Route) String() string {
	var sb strings.Builder
	sb.WriteString("stats=")
	sb.WriteString(r.Stats.String())
	sb.WriteString(" ")

	sb.WriteString("hash=")
	sb.WriteString(fmt.Sprintf("%d", r.Hash64()))
	sb.WriteString(" ")

	sb.WriteString("relayIDs=")
	for idx, relayID := range r.RelayIDs {
		sb.WriteString(fmt.Sprintf("%d", relayID))
		if idx < len(r.RelayIDs) {
			sb.WriteString(" ")
		}
	}
	return sb.String()
}

func (r *Route) Hash64() uint64 {
	fnv64 := fnv.New64()
	id := make([]byte, 8)

	for i := 0; i < r.NumRelays; i++ {
		binary.LittleEndian.PutUint64(id, r.RelayIDs[i])
		fnv64.Write(id)
	}

	return fnv64.Sum64()
}

func (r Route) Equals(other Route) bool {
	if r.NumRelays != other.NumRelays {
		return false
	}

	for i := 0; i < r.NumRelays; i++ {
		if r.RelayIDs[i] != other.RelayIDs[i] {
			return false
		}
	}

	return true
}
