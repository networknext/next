package packets

import (
	"github.com/networknext/next/modules/constants"
	"github.com/networknext/next/modules/crypto"
)

// -------------------------------------------------

const (
	SDK_MaxPacketBytes = constants.MaxPacketBytes

	SDK_SessionDataVersion_Min   = 1
	SDK_SessionDataVersion_Max   = 6
	SDK_SessionDataVersion_Write = 6

	SDK_CRYPTO_SIGN_BYTES             = 64
	SDK_CRYPTO_SIGN_PUBLIC_KEY_BYTES  = 32
	SDK_CRYPTO_SIGN_PRIVATE_KEY_BYTES = 64

	SDK_SERVER_INIT_REQUEST_PACKET     = 50
	SDK_SERVER_INIT_RESPONSE_PACKET    = 51
	SDK_SERVER_UPDATE_REQUEST_PACKET   = 52
	SDK_SERVER_UPDATE_RESPONSE_PACKET  = 53
	SDK_SESSION_UPDATE_REQUEST_PACKET  = 54
	SDK_SESSION_UPDATE_RESPONSE_PACKET = 55
	SDK_CLIENT_RELAY_REQUEST_PACKET    = 56
	SDK_CLIENT_RELAY_RESPONSE_PACKET   = 57
	SDK_SERVER_RELAY_REQUEST_PACKET    = 58
	SDK_SERVER_RELAY_RESPONSE_PACKET   = 59

	SDK_MaxDatacenterNameLength = 256
	SDK_MaxSessionDataSize      = 256
	SDK_MaxTokens               = constants.NextMaxNodes
	SDK_MaxRelaysPerRoute       = constants.NextMaxNodes - 2
	SDK_MaxClientRelays         = int(constants.MaxClientRelays)
	SDK_MaxServerRelays         = int(constants.MaxServerRelays)
	SDK_MaxDestRelays           = int(constants.MaxDestRelays)
	SDK_MaxSessionUpdateRetries = 10

	SDK_ServerInitResponseOK                   = 0
	SDK_ServerInitResponseUnknownBuyer         = 1
	SDK_ServerInitResponseUnknownDatacenter    = 2
	SDK_ServerInitResponseSDKVersionTooOld     = 3
	SDK_ServerInitResponseSignatureCheckFailed = 4
	SDK_ServerInitResponseBuyerNotActive       = 5
	SDK_ServerInitResponseDatacenterNotEnabled = 6

	SDK_PlatformTypeUnknown     = 0
	SDK_PlatformTypeWindows     = 1
	SDK_PlatformTypeMac         = 2
	SDK_PlatformTypeLinux       = 3
	SDK_PlatformTypeSwitch      = 4
	SDK_PlatformTypePS4         = 5
	SDK_PlatformTypeIOS         = 6
	SDK_PlatformTypeXBoxOne     = 7
	SDK_PlatformTypeXBoxSeriesX = 8
	SDK_PlatformTypePS5         = 9
	SDK_PlatformTypeGDK         = 10
	SDK_PlatformTypeMax         = 10

	SDK_ConnectionTypeUnknown  = 0
	SDK_ConnectionTypeWired    = 1
	SDK_ConnectionTypeWifi     = 2
	SDK_ConnectionTypeCellular = 3
	SDK_ConnectionTypeMax      = 3

	SDK_FallbackFlagsBadRouteToken              = (1 << 0)
	SDK_FallbackFlagsNoNextRouteToContinue      = (1 << 1)
	SDK_FallbackFlagsPreviousUpdateStillPending = (1 << 2)
	SDK_FallbackFlagsBadContinueToken           = (1 << 3)
	SDK_FallbackFlagsRouteExpired               = (1 << 4)
	SDK_FallbackFlagsRouteRequestTimedOut       = (1 << 5)
	SDK_FallbackFlagsContinueRequestTimedOut    = (1 << 6)
	SDK_FallbackFlagsClientTimedOut             = (1 << 7)
	SDK_FallbackFlagsUpgradeResponseTimedOut    = (1 << 8)
	SDK_FallbackFlagsRouteUpdateTimedOut        = (1 << 9)
	SDK_FallbackFlagsDirectPongTimedOut         = (1 << 10)
	SDK_FallbackFlagsNextPongTimedOut           = (1 << 11)
	SDK_FallbackFlagsCount                      = 12

	SDK_RouteTypeDirect   = 0
	SDK_RouteTypeNew      = 1
	SDK_RouteTypeContinue = 2

	SDK_NextRouteTokenSize          = 71
	SDK_EncryptedNextRouteTokenSize = 111

	SDK_ContinueRouteTokenSize          = 17
	SDK_EncryptedContinueRouteTokenSize = 57

	SDK_InvalidRouteValue = 10000

	SDK_MaxCityLength    = 128
	SDK_MaxISPNameLength = 64

	SDK_MaxLocationSize = 128

	SDK_SliceSeconds = 10

	SDK_MinPacketBytes = 18 + 3 + 8 + SDK_CRYPTO_SIGN_BYTES + 2

	SDK_MacBytes        = crypto.Box_MacSize
	SDK_NonceBytes      = crypto.Box_NonceSize
	SDK_PublicKeyBytes  = crypto.Box_PublicKeySize
	SDK_PrivateKeyBytes = crypto.Box_PublicKeySize

	SDK_SignatureBytes = crypto.Sign_SignatureSize

	SDK_MaxRelayRTT        = 255
	SDK_MaxRelayJitter     = 255
	SDK_MaxRelayPacketLoss = 100
)
