package transport

import (
	"fmt"
	"net"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/encoding"
	"github.com/networknext/backend/routing"
)

const (
	DefaultMaxPacketSize = 4096

	MaxDatacenterNameLength = 256
	MaxSessionUpdateRetries = 10

	SessionDataVersion = 0
	MaxSessionDataSize = 511

	MaxNearRelays = 32
	MaxTokens     = 7

	PacketTypeServerUpdate       = 220
	PacketTypeSessionUpdate      = 221
	PacketTypeSessionResponse    = 222
	PacketTypeServerInitRequest  = 223
	PacketTypeServerInitResponse = 224

	InitResponseOK                   = 0
	InitResponseUnknownCustomer      = 1
	InitResponseUnknownDatacenter    = 2
	InitResponseOldSDKVersion        = 3
	InitResponseSignatureCheckFailed = 4

	ConnectionTypeUnknown  = 0
	ConnectionTypeWired    = 1
	ConnectionTypeWifi     = 2
	ConnectionTypeCellular = 3
	ConnectionTypeMax      = 3

	PlatformTypeUnknown = 0
	PlatformTypeWindows = 1
	PlatformTypeMac     = 2
	PlatformTypeUnix    = 3
	PlatformTypeSwitch  = 4
	PlatformTypePS4     = 5
	PlatformTypeIOS     = 6
	PlatformTypeXBOXOne = 7
	PlatformTypeMax     = 7

	FallbackFlagsBadRouteToken              = (1 << 0)
	FallbackFlagsNoNextRouteToContinue      = (1 << 1)
	FallbackFlagsPreviousUpdateStillPending = (1 << 2)
	FallbackFlagsBadContinueToken           = (1 << 3)
	FallbackFlagsRouteExpired               = (1 << 4)
	FallbackFlagsRouteRequestTimedOut       = (1 << 5)
	FallbackFlagsContinueRequestTimedOut    = (1 << 6)
	FallbackFlagsClientTimedOut             = (1 << 7)
	FallbackFlagsUpgradeResponseTimedOut    = (1 << 8)
	FallbackFlagsRouteUpdateTimedOut        = (1 << 9)
	FallbackFlagsDirectPongTimedOut         = (1 << 10)
	FallbackFlagsNextPongTimedOut           = (1 << 11)
	FallbackFlagsCount_400                  = 11
	FallbackFlagsCount_401                  = 12

	MaxTags = 8
)

// ConnectionTypeText is similar to http.StatusText(int) which converts the code to a readable text format
func ConnectionTypeText(conntype uint8) string {
	switch conntype {
	case ConnectionTypeWired:
		return "wired"
	case ConnectionTypeWifi:
		return "wifi"
	case ConnectionTypeCellular:
		return "cellular"
	default:
		return "unknown"
	}
}

func ParseConnectionType(conntype string) uint8 {
	switch conntype {
	case "wired":
		return ConnectionTypeWired
	case "wifi":
		return ConnectionTypeWifi
	case "cellular":
		return ConnectionTypeCellular
	default:
		return ConnectionTypeUnknown
	}
}

// PlatformTypeText is similar to http.StatusText(int) which converts the code to a readable text format
func PlatformTypeText(platformType uint8) string {
	switch platformType {
	case PlatformTypeWindows:
		return "Windows"
	case PlatformTypeMac:
		return "Mac"
	case PlatformTypeUnix:
		return "Unix"
	case PlatformTypeSwitch:
		return "Switch"
	case PlatformTypePS4:
		return "PS4"
	case PlatformTypeIOS:
		return "IOS"
	case PlatformTypeXBOXOne:
		return "XBOXOne"
	default:
		return "unknown"
	}
}

func ParsePlatformType(conntype string) uint8 {
	switch conntype {
	case "Windows":
		return PlatformTypeWindows
	case "Mac":
		return PlatformTypeMac
	case "Unix":
		return PlatformTypeUnix
	case "Switch":
		return PlatformTypeSwitch
	case "PS4":
		return PlatformTypePS4
	case "IOS":
		return PlatformTypeIOS
	case "XBOXOne":
		return PlatformTypeXBOXOne
	default:
		return PlatformTypeUnknown
	}
}

// FallbackFlagText is similar to http.StatusText(int) which converts the code to a readable text format
func FallbackFlagText(fallbackFlag uint32) string {
	switch fallbackFlag {
	case FallbackFlagsBadRouteToken:
		return "bad route token"
	case FallbackFlagsNoNextRouteToContinue:
		return "no next route to continue"
	case FallbackFlagsPreviousUpdateStillPending:
		return "previous update still pending"
	case FallbackFlagsBadContinueToken:
		return "bad continue token"
	case FallbackFlagsRouteExpired:
		return "route expired"
	case FallbackFlagsRouteRequestTimedOut:
		return "route request timed out"
	case FallbackFlagsContinueRequestTimedOut:
		return "continue request timed out"
	case FallbackFlagsClientTimedOut:
		return "client timed out"
	case FallbackFlagsUpgradeResponseTimedOut:
		return "upgrade response timed out"
	case FallbackFlagsRouteUpdateTimedOut:
		return "route update timed out"
	case FallbackFlagsDirectPongTimedOut:
		return "direct pong timed out"
	default:
		return "unknown"
	}
}

type Packet interface {
	Serialize(stream encoding.Stream) error
}

func UnmarshalPacket(packet Packet, data []byte) error {
	if err := packet.Serialize(encoding.CreateReadStream(data)); err != nil {
		return err
	}
	return nil
}

func MarshalPacket(packet Packet) ([]byte, error) {
	ws, err := encoding.CreateWriteStream(DefaultMaxPacketSize)
	if err != nil {
		return nil, err
	}

	if err := packet.Serialize(ws); err != nil {
		return nil, err
	}
	ws.Flush()

	return ws.GetData()[:ws.GetBytesProcessed()], nil
}

type ServerInitRequestPacket struct {
	Version        SDKVersion
	CustomerID     uint64
	DatacenterID   uint64
	RequestID      uint64
	DatacenterName string
}

func (packet *ServerInitRequestPacket) Serialize(stream encoding.Stream) error {
	versionMajor := uint32(packet.Version.Major)
	versionMinor := uint32(packet.Version.Minor)
	versionPatch := uint32(packet.Version.Patch)
	stream.SerializeBits(&versionMajor, 8)
	stream.SerializeBits(&versionMinor, 8)
	stream.SerializeBits(&versionPatch, 8)
	packet.Version = SDKVersion{int32(versionMajor), int32(versionMinor), int32(versionPatch)}
	stream.SerializeUint64(&packet.CustomerID)
	stream.SerializeUint64(&packet.DatacenterID)
	stream.SerializeUint64(&packet.RequestID)
	stream.SerializeString(&packet.DatacenterName, MaxDatacenterNameLength)
	return stream.Error()
}

type ServerInitResponsePacket struct {
	RequestID uint64
	Response  uint32
}

func (packet *ServerInitResponsePacket) Serialize(stream encoding.Stream) error {
	stream.SerializeUint64(&packet.RequestID)
	stream.SerializeBits(&packet.Response, 8)
	return stream.Error()
}

type ServerUpdatePacket struct {
	Version       SDKVersion
	CustomerID    uint64
	DatacenterID  uint64
	NumSessions   uint32
	ServerAddress net.UDPAddr
}

func (packet *ServerUpdatePacket) Serialize(stream encoding.Stream) error {
	versionMajor := uint32(packet.Version.Major)
	versionMinor := uint32(packet.Version.Minor)
	versionPatch := uint32(packet.Version.Patch)
	stream.SerializeBits(&versionMajor, 8)
	stream.SerializeBits(&versionMinor, 8)
	stream.SerializeBits(&versionPatch, 8)
	packet.Version = SDKVersion{int32(versionMajor), int32(versionMinor), int32(versionPatch)}
	stream.SerializeUint64(&packet.CustomerID)
	stream.SerializeUint64(&packet.DatacenterID)
	stream.SerializeUint32(&packet.NumSessions)
	stream.SerializeAddress(&packet.ServerAddress)
	return stream.Error()
}

type SessionUpdatePacket struct {
	Version                         SDKVersion
	CustomerID                      uint64
	DatacenterID                    uint64
	SessionID                       uint64
	SliceNumber                     uint32
	RetryNumber                     int32
	SessionDataBytes                int32
	SessionData                     [MaxSessionDataSize]byte
	ClientAddress                   net.UDPAddr
	ServerAddress                   net.UDPAddr
	ClientRoutePublicKey            []byte
	ServerRoutePublicKey            []byte
	UserHash                        uint64
	PlatformType                    int32
	ConnectionType                  int32
	Next                            bool
	Committed                       bool
	Reported                        bool
	FallbackToDirect                bool
	ClientBandwidthOverLimit        bool
	ServerBandwidthOverLimit        bool
	ClientPingTimedOut              bool
	NumTags                         int32
	Tags                            [MaxTags]uint64
	Flags                           uint32
	UserFlags                       uint64
	DirectRTT                       float32
	DirectJitter                    float32
	DirectPacketLoss                float32
	NextRTT                         float32
	NextJitter                      float32
	NextPacketLoss                  float32
	NumNearRelays                   int32
	NearRelayIDs                    []uint64
	NearRelayRTT                    []float32
	NearRelayJitter                 []float32
	NearRelayPacketLoss             []float32
	NextKbpsUp                      uint32
	NextKbpsDown                    uint32
	PacketsSentClientToServer       uint64
	PacketsSentServerToClient       uint64
	PacketsLostClientToServer       uint64
	PacketsLostServerToClient       uint64
	PacketsOutOfOrderClientToServer uint64
	PacketsOutOfOrderServerToClient uint64
	JitterClientToServer            float32
	JitterServerToClient            float32
}

func (packet *SessionUpdatePacket) Serialize(stream encoding.Stream) error {
	
	versionMajor := uint32(packet.Version.Major)
	versionMinor := uint32(packet.Version.Minor)
	versionPatch := uint32(packet.Version.Patch)
	
	stream.SerializeBits(&versionMajor, 8)
	stream.SerializeBits(&versionMinor, 8)
	stream.SerializeBits(&versionPatch, 8)
	
	packet.Version = SDKVersion{int32(versionMajor), int32(versionMinor), int32(versionPatch)}
	
	stream.SerializeUint64(&packet.CustomerID)
	stream.SerializeUint64(&packet.DatacenterID)
	stream.SerializeUint64(&packet.SessionID)
	stream.SerializeUint32(&packet.SliceNumber)
	stream.SerializeInteger(&packet.RetryNumber, 0, MaxSessionUpdateRetries)
	
	stream.SerializeInteger(&packet.SessionDataBytes, 0, MaxSessionDataSize)
	if packet.SessionDataBytes > 0 {
		sessionData := packet.SessionData[:packet.SessionDataBytes]
		stream.SerializeBytes(sessionData)
	}
	
	stream.SerializeAddress(&packet.ClientAddress)
	
	stream.SerializeAddress(&packet.ServerAddress)
	
	if stream.IsReading() {
		packet.ClientRoutePublicKey = make([]byte, crypto.KeySize)
		packet.ServerRoutePublicKey = make([]byte, crypto.KeySize)
	}
	
	stream.SerializeBytes(packet.ClientRoutePublicKey)

	stream.SerializeBytes(packet.ServerRoutePublicKey)
	
	stream.SerializeUint64(&packet.UserHash)
	
	stream.SerializeInteger(&packet.PlatformType, PlatformTypeUnknown, PlatformTypeMax)
	
	stream.SerializeInteger(&packet.ConnectionType, ConnectionTypeUnknown, ConnectionTypeMax)
	
	stream.SerializeBool(&packet.Next)
	stream.SerializeBool(&packet.Committed)
	stream.SerializeBool(&packet.Reported)
	stream.SerializeBool(&packet.FallbackToDirect)
	stream.SerializeBool(&packet.ClientBandwidthOverLimit)
	stream.SerializeBool(&packet.ServerBandwidthOverLimit)
	if core.ProtocolVersionAtLeast(versionMajor, versionMinor, versionPatch, 4, 0, 2) {
		stream.SerializeBool(&packet.ClientPingTimedOut)
	}
	
	hasTags := stream.IsWriting() && packet.NumTags > 0
	hasFlags := stream.IsWriting() && packet.Flags != 0
	hasUserFlags := stream.IsWriting() && packet.UserFlags != 0
	hasLostPackets := stream.IsWriting() && (packet.PacketsLostClientToServer+packet.PacketsLostServerToClient) > 0
	hasOutOfOrderPackets := stream.IsWriting() && (packet.PacketsOutOfOrderClientToServer+packet.PacketsOutOfOrderServerToClient) > 0
	
	stream.SerializeBool(&hasTags)
	stream.SerializeBool(&hasFlags)
	stream.SerializeBool(&hasUserFlags)
	stream.SerializeBool(&hasLostPackets)
	stream.SerializeBool(&hasOutOfOrderPackets)
	
	if hasTags {
		if core.ProtocolVersionAtLeast(versionMajor, versionMinor, versionPatch, 4, 0, 3) {
			// multiple tags (SDK 4.0.3 and above)
			stream.SerializeInteger(&packet.NumTags, 0, MaxTags)
			for i := 0; i < int(packet.NumTags); i++ {
				stream.SerializeUint64(&packet.Tags[i])
			}
		} else {
			// single tag (< SDK 4.0.3)
			stream.SerializeUint64(&packet.Tags[0])
			if stream.IsWriting() {
				packet.NumTags = 1
			}
		}
	}
	
	if hasFlags {
		if core.ProtocolVersionAtLeast(versionMajor, versionMinor, versionPatch, 4, 0, 1) {
			// flag added in SDK 4.0.1 for fallback to new direct reason
			stream.SerializeBits(&packet.Flags, FallbackFlagsCount_401)
		} else {
			stream.SerializeBits(&packet.Flags, FallbackFlagsCount_400)
		}
	}
	
	if hasUserFlags {
		stream.SerializeUint64(&packet.UserFlags)
	}
	
	stream.SerializeFloat32(&packet.DirectRTT)
	stream.SerializeFloat32(&packet.DirectJitter)
	stream.SerializeFloat32(&packet.DirectPacketLoss)
	
	if packet.Next {
		stream.SerializeFloat32(&packet.NextRTT)
		stream.SerializeFloat32(&packet.NextJitter)
		stream.SerializeFloat32(&packet.NextPacketLoss)
	}
	
	stream.SerializeInteger(&packet.NumNearRelays, 0, MaxNearRelays)
	
	if stream.IsReading() {
		packet.NearRelayIDs = make([]uint64, packet.NumNearRelays)
		packet.NearRelayRTT = make([]float32, packet.NumNearRelays)
		packet.NearRelayJitter = make([]float32, packet.NumNearRelays)
		packet.NearRelayPacketLoss = make([]float32, packet.NumNearRelays)
	}
	
	for i := int32(0); i < packet.NumNearRelays; i++ {
		stream.SerializeUint64(&packet.NearRelayIDs[i])
		stream.SerializeFloat32(&packet.NearRelayRTT[i])
		stream.SerializeFloat32(&packet.NearRelayJitter[i])
		stream.SerializeFloat32(&packet.NearRelayPacketLoss[i])
	}
	
	if packet.Next {
		stream.SerializeUint32(&packet.NextKbpsUp)
		stream.SerializeUint32(&packet.NextKbpsDown)
	}
	
	stream.SerializeUint64(&packet.PacketsSentClientToServer)
	stream.SerializeUint64(&packet.PacketsSentServerToClient)
	
	if hasLostPackets {
		stream.SerializeUint64(&packet.PacketsLostClientToServer)
		stream.SerializeUint64(&packet.PacketsLostServerToClient)
	}
	
	if hasOutOfOrderPackets {
		stream.SerializeUint64(&packet.PacketsOutOfOrderClientToServer)
		stream.SerializeUint64(&packet.PacketsOutOfOrderServerToClient)
	}
	
	stream.SerializeFloat32(&packet.JitterClientToServer)
	stream.SerializeFloat32(&packet.JitterServerToClient)

	return stream.Error()
}

type SessionResponsePacket struct {
	SessionID          uint64
	SliceNumber        uint32
	SessionDataBytes   int32
	SessionData        [MaxSessionDataSize]byte
	RouteType          int32
	NumNearRelays      int32
	NearRelayIDs       []uint64
	NearRelayAddresses []net.UDPAddr
	NumTokens          int32
	Tokens             []byte
	Multipath          bool
	Committed          bool
}

func (packet *SessionResponsePacket) Serialize(stream encoding.Stream) error {
	stream.SerializeUint64(&packet.SessionID)
	stream.SerializeUint32(&packet.SliceNumber)
	stream.SerializeInteger(&packet.SessionDataBytes, 0, MaxSessionDataSize)
	if packet.SessionDataBytes > 0 {
		sessionData := packet.SessionData[:packet.SessionDataBytes]
		stream.SerializeBytes(sessionData)
	}
	stream.SerializeInteger(&packet.RouteType, 0, routing.RouteTypeContinue)
	stream.SerializeInteger(&packet.NumNearRelays, 0, MaxNearRelays)
	if stream.IsReading() {
		packet.NearRelayIDs = make([]uint64, packet.NumNearRelays)
		packet.NearRelayAddresses = make([]net.UDPAddr, packet.NumNearRelays)
	}
	for i := int32(0); i < packet.NumNearRelays; i++ {
		stream.SerializeUint64(&packet.NearRelayIDs[i])
		stream.SerializeAddress(&packet.NearRelayAddresses[i])
	}
	if packet.RouteType != routing.RouteTypeDirect {
		stream.SerializeBool(&packet.Multipath)
		stream.SerializeBool(&packet.Committed)
		stream.SerializeInteger(&packet.NumTokens, 0, MaxTokens)
	}
	if packet.RouteType == routing.RouteTypeNew {
		if stream.IsReading() {
			packet.Tokens = make([]byte, packet.NumTokens*routing.EncryptedNextRouteTokenSize)
		}
		stream.SerializeBytes(packet.Tokens)
	}
	if packet.RouteType == routing.RouteTypeContinue {
		if stream.IsReading() {
			packet.Tokens = make([]byte, packet.NumTokens*routing.EncryptedContinueRouteTokenSize)
		}
		stream.SerializeBytes(packet.Tokens)
	}

	return stream.Error()
}

type SessionData struct {
	Version          uint32
	SessionID        uint64
	SessionVersion   uint32
	SliceNumber      uint32
	ExpireTimestamp  uint64
	Initial          bool
	Location         routing.Location
	RouteNumRelays   int32
	RouteCost        int32
	RouteRelayIDs    [routing.MaxRelays]uint64
	RouteState       core.RouteState
	EverOnNext       bool
	FellBackToDirect bool
}

func UnmarshalSessionData(sessionData *SessionData, data []byte) error {
	if err := sessionData.Serialize(encoding.CreateReadStream(data)); err != nil {
		return err
	}
	return nil
}

func MarshalSessionData(sessionData *SessionData) ([]byte, error) {
	ws, err := encoding.CreateWriteStream(DefaultMaxPacketSize)
	if err != nil {
		return nil, err
	}

	if err := sessionData.Serialize(ws); err != nil {
		return nil, err
	}
	ws.Flush()

	return ws.GetData()[:ws.GetBytesProcessed()], nil
}

func (sessionData *SessionData) Serialize(stream encoding.Stream) error {
	stream.SerializeBits(&sessionData.Version, 8)
	if stream.IsReading() && sessionData.Version != SessionDataVersion {
		return fmt.Errorf("bad session data version %d, expected %d", sessionData.Version, SessionDataVersion)
	}
	stream.SerializeUint64(&sessionData.SessionID)
	stream.SerializeBits(&sessionData.SessionVersion, 8)
	stream.SerializeUint32(&sessionData.SliceNumber)
	stream.SerializeUint64(&sessionData.ExpireTimestamp)
	stream.SerializeBool(&sessionData.Initial)
	locationSize := uint32(sessionData.Location.Size())
	stream.SerializeUint32(&locationSize)
	if stream.IsReading() {
		locationBytes := make([]byte, locationSize)
		stream.SerializeBytes(locationBytes)
		if err := sessionData.Location.UnmarshalBinary(locationBytes); err != nil {
			return err
		}
	} else {
		locationBytes, err := sessionData.Location.MarshalBinary()
		if err != nil {
			return err
		}
		stream.SerializeBytes(locationBytes)
	}
	hasRoute := sessionData.RouteNumRelays > 0
	stream.SerializeBool(&hasRoute)
	if hasRoute {
		stream.SerializeInteger(&sessionData.RouteNumRelays, 0, routing.MaxRelays)
		stream.SerializeInteger(&sessionData.RouteCost, 0, routing.InvalidRouteValue)
		for i := int32(0); i < sessionData.RouteNumRelays; i++ {
			stream.SerializeUint64(&sessionData.RouteRelayIDs[i])
		}
	}
	stream.SerializeUint64(&sessionData.RouteState.UserID)
	stream.SerializeBool(&sessionData.RouteState.Next)
	stream.SerializeBool(&sessionData.RouteState.Veto)
	stream.SerializeBool(&sessionData.RouteState.Banned)
	stream.SerializeBool(&sessionData.RouteState.Disabled)
	stream.SerializeBool(&sessionData.RouteState.NotSelected)
	stream.SerializeBool(&sessionData.RouteState.ABTest)
	stream.SerializeBool(&sessionData.RouteState.A)
	stream.SerializeBool(&sessionData.RouteState.B)
	stream.SerializeBool(&sessionData.RouteState.ForcedNext)
	stream.SerializeBool(&sessionData.RouteState.ReduceLatency)
	stream.SerializeBool(&sessionData.RouteState.ReducePacketLoss)
	stream.SerializeBool(&sessionData.RouteState.ProMode)
	stream.SerializeBool(&sessionData.RouteState.Multipath)
	stream.SerializeBool(&sessionData.RouteState.Committed)
	stream.SerializeBool(&sessionData.RouteState.CommitPending)
	stream.SerializeInteger(&sessionData.RouteState.CommitCounter, 0, 3)
	stream.SerializeBool(&sessionData.RouteState.LatencyWorse)
	stream.SerializeBool(&sessionData.RouteState.MultipathOverload)
	stream.SerializeBool(&sessionData.RouteState.NoRoute)
	stream.SerializeBool(&sessionData.RouteState.CommitVeto)
	stream.SerializeBool(&sessionData.EverOnNext)
	stream.SerializeBool(&sessionData.FellBackToDirect)

	return stream.Error()
}