package constants

const (
	MajorVersion = 5
	MinorVersion = 0
	PatchVersion = 0

	MaxRelays        = 1000
	NumRelayCounters = 128
	RelayTimeout     = 10
	RelayHistorySize = 60

	MaxNearRelays  = 32
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

	// todo: convert these to golang style
	NEXT_MAX_NODES                      = 7
	NEXT_ADDRESS_BYTES                  = 19
	NEXT_ADDRESS_BYTES_SHORT            = 7
	NEXT_ROUTE_TOKEN_BYTES              = 76
	NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES    = 116
	NEXT_CONTINUE_TOKEN_BYTES           = 17
	NEXT_ENCRYPTED_CONTINUE_TOKEN_BYTES = 57

	SessionFlags_Next                            = (1 << 0)
	SessionFlags_Reported                        = (1 << 1)
	SessionFlags_Summary                         = (1 << 2)
	SessionFlags_FallbackToDirect                = (1 << 3)
	SessionFlags_Mispredict                      = (1 << 4)
	SessionFlags_LatencyWorse                    = (1 << 5)
	SessionFlags_NoRoute                         = (1 << 6)
	SessionFlags_NextLatencyTooHigh              = (1 << 7)
	SessionFlags_UnknownDatacenter               = (1 << 8)
	SessionFlags_DatacenterNotEnabled            = (1 << 9)
	SessionFlags_StaleRouteMatrix                = (1 << 10)
	SessionFlags_ABTest                          = (1 << 11)
	SessionFlags_Aborted                         = (1 << 12)
	SessionFlags_LatencyReduction                = (1 << 13)
	SessionFlags_PacketLossReduction             = (1 << 14)
	SessionFlags_EverOnNext                      = (1 << 15)
	SessionFlags_SessionDataSignatureCheckFailed = (1 << 16)
	SessionFlags_FailedToReadSessionData         = (1 << 17)
	SessionFlags_LongDuration                    = (1 << 18)
	SessionFlags_ClientPingTimedOut              = (1 << 19)
	SessionFlags_BadSessionId                    = (1 << 20)
	SessionFlags_BadSliceNumber                  = (1 << 21)
	SessionFlags_AnalysisOnly                    = (1 << 22)
	SessionFlags_NoRelaysInDatacenter            = (1 << 23)
	SessionFlags_NoNearRelays                    = (1 << 24)
	SessionFlags_NoRouteRelays                   = (1 << 25)
	SessionFlags_RouteRelayNoLongerExists        = (1 << 26)
	SessionFlags_RouteChanged                    = (1 << 27)
	SessionFlags_RouteContinued                  = (1 << 28)
	SessionFlags_RouteNoLongerExists             = (1 << 29)
	SessionFlags_TakeNetworkNext                 = (1 << 30)
	SessionFlags_StayDirect                      = (1 << 31)
	SessionFlags_LeftNetworkNext                 = (1 << 32)
	SessionFlags_FailedToWriteResponsePacket     = (1 << 33)
	SessionFlags_FailedToWriteSessionData        = (1 << 34)
	SessionFlags_LocationVeto                    = (1 << 35)
	SessionFlags_ClientNextBandwidthOverLimit    = (1 << 36)
	SessionFlags_ServerNextBandwidthOverLimit    = (1 << 37)

	RelayFlags_ShuttingDown = uint64(1)

	RelayStatus_Offline      = 0
	RelayStatus_Online       = 1
	RelayStatus_ShuttingDown = 2
)
