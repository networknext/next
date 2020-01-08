package relayport

// PingRelayDataJSON is a struct for holding ping data
type PingRelayDataJSON struct {
	ID        uint64
	Group     uint64
	Role      string
	Address   []byte
	PublicKey []byte
	PingKey   []byte
}
