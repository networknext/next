package constants

const (
	MajorVersion = 5
	MinorVersion = 0
	PatchVersion = 0

	MaxRelays        = 1000
	NumRelayCounters = 128
	RelayTimeout     = 10
	RelayHistorySize = 60

	MaxNearRelays  = 16
	MaxRouteRelays = 5

	MaxRelayNameLength      = 63
	MaxRelayVersionLength   = 32
	MaxDatacenterNameLength = 256

	MaxTags = 8

	MaxMatchValues = 64

	MagicBytes = 8

	MaxConnectionType = 3

	MaxPlatformType = 10

	CostBias          = 3
	MaxIndirects      = 8
	MaxRoutesPerEntry = 16
	JitterThreshold   = 15
	CostThreshold     = 1

	MaxRouteCost = 255

	NEXT_MAX_NODES                      = 7
	NEXT_ADDRESS_BYTES                  = 19
	NEXT_ADDRESS_IPV4_BYTES             = 6
	NEXT_ROUTE_TOKEN_BYTES              = 71
	NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES    = 111
	NEXT_CONTINUE_TOKEN_BYTES           = 17
	NEXT_ENCRYPTED_CONTINUE_TOKEN_BYTES = 57

	// todo: these need to go somewhere. they are useful and I want to see them in slice data, and in aggregate in the summary
	/*
		SessionFlags_Reported                        = (1 << 1)
		SessionFlags_LatencyReduction                = (1 << 13)
		SessionFlags_PacketLossReduction             = (1 << 14)
		SessionFlags_LongDuration                    = (1 << 18)
	*/

	SessionError_FallbackToDirect                = (1 << 1)
	SessionError_NoRoute                         = (1 << 2)
	SessionError_UnknownDatacenter               = (1 << 3)
	SessionError_DatacenterNotEnabled            = (1 << 4)
	SessionError_StaleRouteMatrix                = (1 << 5)
	SessionError_Aborted                         = (1 << 6)
	SessionError_SessionDataSignatureCheckFailed = (1 << 7)
	SessionError_FailedToReadSessionData         = (1 << 8)
	SessionError_BadSessionId                    = (1 << 9)
	SessionError_BadSliceNumber                  = (1 << 10)
	SessionError_NoRelaysInDatacenter            = (1 << 11)
	SessionError_NoNearRelays                    = (1 << 12)
	SessionError_NoRouteRelays                   = (1 << 13)
	SessionError_RouteRelayNoLongerExists        = (1 << 14)
	SessionError_RouteNoLongerExists             = (1 << 15)
	SessionError_FailedToWriteResponsePacket     = (1 << 16)
	SessionError_FailedToWriteSessionData        = (1 << 17)

	RelayFlags_ShuttingDown = uint64(1)

	RelayStatus_Offline      = 0
	RelayStatus_Online       = 1
	RelayStatus_ShuttingDown = 2

	PingTokenBytes = 32
)
