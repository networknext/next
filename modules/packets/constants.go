package packets

import (
	"github.com/networknext/backend/modules/core"

	"github.com/networknext/backend/modules-old/crypto"
)

// -------------------------------------------------

const (
	NEXT_CRYPTO_SIGN_BYTES = 64
	NEXT_CRYPTO_SIGN_PUBLIC_KEY_BYTES = 32
	NEXT_CRYPTO_SIGN_PRIVATE_KEY_BYTES = 64
)

// -------------------------------------------------

const (
	// todo: clean up as per sdk5 below
	/*
		SDK4_PACKET_TYPE_ServerUpdate       = 220
		SDK4_PACKET_TYPE_SessionUpdate      = 221
		SDK4_PACKET_TYPE_SessionResponse    = 222
		SDK4_PACKET_TYPE_ServerInitRequest  = 223
		SDK4_PACKET_TYPE_ServerInitResponse = 224
		SDK4_PACKET_TYPE_MatchDataRequest   = 225
		SDK4_PACKET_TYPE_MatchDataResponse  = 226
	*/

	SDK4_MaxDatacenterNameLength = 256
	SDK4_MaxSessionDataSize      = 511
	SDK4_MaxTags                 = 8
	SDK4_MaxTokens               = core.NEXT_MAX_NODES
	SDK4_MaxRelaysPerRoute       = core.NEXT_MAX_NODES
	SDK4_MaxNearRelays           = core.MaxNearRelays
	SDK4_MaxSessionUpdateRetries = 10
	SDK4_MaxSessionDebug         = 1024

	SDK4_ServerInitResponseOK                   = 0
	SDK4_ServerInitResponseUnknownBuyer         = 1
	SDK4_ServerInitResponseUnknownDatacenter    = 2
	SDK4_ServerInitResponseOldSDKVersion        = 3
	SDK4_ServerInitResponseSignatureCheckFailed = 4
	SDK4_ServerInitResponseBuyerNotActive       = 5
	SDK4_ServerInitResponseDatacenterNotEnabled = 6

	SDK4_PlatformTypeUnknown     = 0
	SDK4_PlatformTypeWindows     = 1
	SDK4_PlatformTypeMac         = 2
	SDK4_PlatformTypeLinux       = 3
	SDK4_PlatformTypeSwitch      = 4
	SDK4_PlatformTypePS4         = 5
	SDK4_PlatformTypeIOS         = 6
	SDK4_PlatformTypeXBoxOne     = 7
	SDK4_PlatformTypeXBoxSeriesX = 8
	SDK4_PlatformTypePS5         = 9
	SDK4_PlatformTypeGDK         = 10
	SDK4_PlatformTypeMax         = 10

	SDK4_ConnectionTypeUnknown  = 0
	SDK4_ConnectionTypeWired    = 1
	SDK4_ConnectionTypeWifi     = 2
	SDK4_ConnectionTypeCellular = 3
	SDK4_ConnectionTypeMax      = 3

	SDK4_FallbackFlagsBadRouteToken              = (1 << 0)
	SDK4_FallbackFlagsNoNextRouteToContinue      = (1 << 1)
	SDK4_FallbackFlagsPreviousUpdateStillPending = (1 << 2)
	SDK4_FallbackFlagsBadContinueToken           = (1 << 3)
	SDK4_FallbackFlagsRouteExpired               = (1 << 4)
	SDK4_FallbackFlagsRouteRequestTimedOut       = (1 << 5)
	SDK4_FallbackFlagsContinueRequestTimedOut    = (1 << 6)
	SDK4_FallbackFlagsClientTimedOut             = (1 << 7)
	SDK4_FallbackFlagsUpgradeResponseTimedOut    = (1 << 8)
	SDK4_FallbackFlagsRouteUpdateTimedOut        = (1 << 9)
	SDK4_FallbackFlagsDirectPongTimedOut         = (1 << 10)
	SDK4_FallbackFlagsNextPongTimedOut           = (1 << 11)
	SDK4_FallbackFlagsCount                      = 12

	SDK4_RouteTypeDirect   = 0
	SDK4_RouteTypeNew      = 1
	SDK4_RouteTypeContinue = 2

	SDK4_NextRouteTokenSize          = 100
	SDK4_EncryptedNextRouteTokenSize = SDK4_NextRouteTokenSize + crypto.MACSize

	SDK4_ContinueRouteTokenSize          = 41
	SDK4_EncryptedContinueRouteTokenSize = SDK4_ContinueRouteTokenSize + crypto.MACSize

	SDK4_SessionDataVersion = 15

	SDK4_MaxMatchValues = 64

	SDK4_InvalidRouteValue = 10000

	SDK4_LocationVersion = 1

	SDK4_MaxContinentLength   = 16
	SDK4_MaxCountryLength     = 64
	SDK4_MaxCountryCodeLength = 16
	SDK4_MaxRegionLength      = 64
	SDK4_MaxCityLength        = 128
	SDK4_MaxISPNameLength     = 64

	SDK4_MaxLocationSize = 128
)

// -------------------------------------------------

const (
	SDK5_SERVER_INIT_REQUEST_PACKET     = 50
	SDK5_SERVER_INIT_RESPONSE_PACKET    = 51
	SDK5_SERVER_UPDATE_REQUEST_PACKET   = 52
	SDK5_SERVER_UPDATE_RESPONSE_PACKET  = 53
	SDK5_SESSION_UPDATE_REQUEST_PACKET  = 54
	SDK5_SESSION_UPDATE_RESPONSE_PACKET = 55
	SDK5_MATCH_DATA_REQUEST_PACKET      = 56
	SDK5_MATCH_DATA_RESPONSE_PACKET     = 57

	SDK5_MaxDatacenterNameLength = 256
	SDK5_MaxSessionDataSize      = 511
	SDK5_MaxTags                 = 8
	SDK5_MaxTokens               = core.NEXT_MAX_NODES
	SDK5_MaxRelaysPerRoute       = core.NEXT_MAX_NODES
	SDK5_MaxNearRelays           = core.MaxNearRelays
	SDK5_MaxSessionUpdateRetries = 10
	SDK5_MaxSessionDebug         = 1024

	SDK5_ServerInitResponseOK                   = 0
	SDK5_ServerInitResponseUnknownBuyer         = 1
	SDK5_ServerInitResponseBuyerNotActive       = 2
	SDK5_ServerInitResponseOldSDKVersion        = 3

	SDK5_PlatformTypeUnknown     = 0
	SDK5_PlatformTypeWindows     = 1
	SDK5_PlatformTypeMac         = 2
	SDK5_PlatformTypeLinux       = 3
	SDK5_PlatformTypeSwitch      = 4
	SDK5_PlatformTypePS4         = 5
	SDK5_PlatformTypeIOS         = 6
	SDK5_PlatformTypeXBoxOne     = 7
	SDK5_PlatformTypeXBoxSeriesX = 8
	SDK5_PlatformTypePS5         = 9
	SDK5_PlatformTypeGDK         = 10
	SDK5_PlatformTypeMax         = 10

	SDK5_ConnectionTypeUnknown  = 0
	SDK5_ConnectionTypeWired    = 1
	SDK5_ConnectionTypeWifi     = 2
	SDK5_ConnectionTypeCellular = 3
	SDK5_ConnectionTypeMax      = 3

	SDK5_FallbackFlagsBadRouteToken              = (1 << 0)
	SDK5_FallbackFlagsNoNextRouteToContinue      = (1 << 1)
	SDK5_FallbackFlagsPreviousUpdateStillPending = (1 << 2)
	SDK5_FallbackFlagsBadContinueToken           = (1 << 3)
	SDK5_FallbackFlagsRouteExpired               = (1 << 4)
	SDK5_FallbackFlagsRouteRequestTimedOut       = (1 << 5)
	SDK5_FallbackFlagsContinueRequestTimedOut    = (1 << 6)
	SDK5_FallbackFlagsClientTimedOut             = (1 << 7)
	SDK5_FallbackFlagsUpgradeResponseTimedOut    = (1 << 8)
	SDK5_FallbackFlagsRouteUpdateTimedOut        = (1 << 9)
	SDK5_FallbackFlagsDirectPongTimedOut         = (1 << 10)
	SDK5_FallbackFlagsNextPongTimedOut           = (1 << 11)
	SDK5_FallbackFlagsCount                      = 12

	SDK5_RouteTypeDirect   = 0
	SDK5_RouteTypeNew      = 1
	SDK5_RouteTypeContinue = 2

	SDK5_NextRouteTokenSize          = 100
	SDK5_EncryptedNextRouteTokenSize = SDK4_NextRouteTokenSize + crypto.MACSize

	SDK5_ContinueRouteTokenSize          = 41
	SDK5_EncryptedContinueRouteTokenSize = SDK4_ContinueRouteTokenSize + crypto.MACSize

	SDK5_SessionDataVersion = 15

	SDK5_MaxMatchValues = 64

	SDK5_InvalidRouteValue = 10000

	SDK5_LocationVersion = 1

	SDK5_MaxContinentLength   = 16
	SDK5_MaxCountryLength     = 64
	SDK5_MaxCountryCodeLength = 16
	SDK5_MaxRegionLength      = 64
	SDK5_MaxCityLength        = 128
	SDK5_MaxISPNameLength     = 64

	SDK5_MaxLocationSize = 128
)

// -------------------------------------------------
