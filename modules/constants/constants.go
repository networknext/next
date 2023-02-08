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
	
	RelayFlags_ShuttingDown = uint64(1)
)
