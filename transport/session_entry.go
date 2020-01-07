package transport

// SessionEntry is the struct for the entries into the session database (game client connections?)
type SessionEntry struct {
	id              uint64
	version         uint8
	expireTimestamp uint64
	route           []uint64
	next            bool
	slice           uint64
}
