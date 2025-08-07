package constants

const (
	MajorVersion = 1
	MinorVersion = 0
	PatchVersion = 0

	MaxPacketBytes = 1384

	MaxBuyers = 1024

	MaxRelays        = 1000
	NumRelayCounters = 150
	RelayTimeout     = 10
	RelayHistorySize = 300

	MaxRouteRelays  = 5
	MaxClientRelays = 16
	MaxServerRelays = 8
	MaxDestRelays   = MaxServerRelays

	MaxRelayNameLength      = 63
	MaxRelayVersionLength   = 32
	MaxDatacenterNameLength = 256

	MagicBytes = 8

	MaxConnectionType = 3

	MaxPlatformType = 10

	CostBias          = 3
	MaxIndirects      = 8
	MaxRoutesPerEntry = 16

	MaxRouteCost = 255

	NextMaxNodes = MaxRouteRelays + 2

	NextAddressBytes      = 19
	NextAddressBytes_IPv4 = 6

	RouteTokenBytes          = 71
	EncryptedRouteTokenBytes = 111

	ContinueTokenBytes          = 17
	EncryptedContinueTokenBytes = 57

	SessionError_FallbackToDirect                = (1 << 0)
	SessionError_NoRoute                         = (1 << 1)
	SessionError_UnknownDatacenter               = (1 << 2)
	SessionError_DatacenterNotEnabled            = (1 << 3)
	SessionError_StaleRouteMatrix                = (1 << 4)
	SessionError_Aborted                         = (1 << 5)
	SessionError_SessionDataSignatureCheckFailed = (1 << 6)
	SessionError_FailedToReadSessionData         = (1 << 7)
	SessionError_BadSessionId                    = (1 << 8)
	SessionError_BadSliceNumber                  = (1 << 9)
	SessionError_NoClientRelays                  = (1 << 10)
	SessionError_NoServerRelays                  = (1 << 11)
	SessionError_NoRouteRelays                   = (1 << 12)
	SessionError_RouteRelayNoLongerExists        = (1 << 13)
	SessionError_RouteNoLongerExists             = (1 << 14)
	SessionError_FailedToWriteResponsePacket     = (1 << 15)
	SessionError_FailedToWriteSessionData        = (1 << 16)

	RelayFlags_ShuttingDown = uint64(1)

	RelayStatus_Offline      = 0
	RelayStatus_Online       = 1
	RelayStatus_ShuttingDown = 2

	PingKeyBytes = 32

	PingTokenBytes = 32

	SecretKeyBytes = 32

	MaxDatabaseSize = 256 * 1024
)
