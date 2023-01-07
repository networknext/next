package packets

import (
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/crypto"
)

// -------------------------------------------------

const (
	SDK5_SessionDataVersion_Min   = 1
	SDK5_SessionDataVersion_Max   = 1
	SDK5_SessionDataVersion_Write = 1

	SDK5_LocationVersion_Min   = 1
	SDK5_LocationVersion_Max   = 1
	SDK5_LocationVersion_Write = 1

	SDK5_CRYPTO_SIGN_BYTES             = 64
	SDK5_CRYPTO_SIGN_PUBLIC_KEY_BYTES  = 32
	SDK5_CRYPTO_SIGN_PRIVATE_KEY_BYTES = 64

	SDK5_SERVER_INIT_REQUEST_PACKET     = 50
	SDK5_SERVER_INIT_RESPONSE_PACKET    = 51
	SDK5_SERVER_UPDATE_REQUEST_PACKET   = 52
	SDK5_SERVER_UPDATE_RESPONSE_PACKET  = 53
	SDK5_SESSION_UPDATE_REQUEST_PACKET  = 54
	SDK5_SESSION_UPDATE_RESPONSE_PACKET = 55
	SDK5_MATCH_DATA_REQUEST_PACKET      = 56
	SDK5_MATCH_DATA_RESPONSE_PACKET     = 57

	SDK5_MaxDatacenterNameLength = 256
	SDK5_MaxSessionDataSize      = 500
	SDK5_MaxTags                 = 8
	SDK5_MaxTokens               = core.NEXT_MAX_NODES
	SDK5_MaxRelaysPerRoute       = core.NEXT_MAX_NODES - 2
	SDK5_MaxNearRelays           = int(core.MaxNearRelays)
	SDK5_MaxSessionUpdateRetries = 10
	SDK5_MaxSessionDebug         = 1024

	SDK5_ServerInitResponseOK             = 0
	SDK5_ServerInitResponseUnknownBuyer   = 1
	SDK5_ServerInitResponseBuyerNotActive = 2
	SDK5_ServerInitResponseOldSDKVersion  = 3

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
	SDK5_EncryptedNextRouteTokenSize = SDK5_NextRouteTokenSize + crypto.Box_MacSize

	SDK5_ContinueRouteTokenSize          = 41
	SDK5_EncryptedContinueRouteTokenSize = SDK5_ContinueRouteTokenSize + crypto.Box_MacSize

	SDK5_MaxMatchValues = 64

	SDK5_InvalidRouteValue = 10000

	SDK5_MaxContinentLength   = 16
	SDK5_MaxCountryLength     = 64
	SDK5_MaxCountryCodeLength = 16
	SDK5_MaxRegionLength      = 64
	SDK5_MaxCityLength        = 128
	SDK5_MaxISPNameLength     = 64

	SDK5_MaxLocationSize = 128

	SDK5_BillingSliceSeconds = 10

	SDK5_MinPacketBytes = 16 + 3 + 8 + SDK5_CRYPTO_SIGN_BYTES + 2
	SDK5_MaxPacketBytes = 4096

	SDK5_MacBytes   = crypto.Box_MacSize
	SDK5_NonceBytes = crypto.Box_NonceSize
	SDK5_KeyBytes   = crypto.Box_KeySize

	SDK5_SignatureBytes = crypto.Sign_SignatureSize

	SDK5_MaxNearRelayRTT        = 255
	SDK5_MaxNearRelayJitter     = 255
	SDK5_MaxNearRelayPacketLoss = 100
)
