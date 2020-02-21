package routing

type Route struct {
	Relays []Relay
	Stats  Stats
}

func GetRouteHash(relayIds []RelayId) uint64 {
	hash := fnv.New64a()
	for _, v := range relayIds {
		a := make([]byte, 4)
		binary.LittleEndian.PutUint32(a, uint32(v))
		hash.Write(a)
	}
	return hash.Sum64()
}