package constants

const (
	MaxRelays                   = 1000
	NumRelayCounters            = 128
	MaxRelayVersionStringLength = 32
	RelayTimeout                = 10
	RelayHistorySize            = 60

	MaxNearRelays  = 32
	MaxRouteRelays = 5

	MaxTags = 8

	CostBias = 3
	MaxRoutesPerEntry = 64
	JitterThreshold = 15      // todo: poorly named

	// todo: convert these to golang style
	NEXT_MAX_NODES = 7
	NEXT_ADDRESS_BYTES = 19
	NEXT_ADDRESS_BYTES_SHORT = 7
	NEXT_ROUTE_TOKEN_BYTES = 76
	NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES = 116
	NEXT_CONTINUE_TOKEN_BYTES = 17
	NEXT_ENCRYPTED_CONTINUE_TOKEN_BYTES = 57

	RelayFlags_ShuttingDown = uint64(1)
)
