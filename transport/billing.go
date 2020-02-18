package transport

const BillingSliceSeconds = 10

type BillingEntry struct {
	Request             *RouteRequest
	Route               []*BillingRouteHop
	RouteDecision       uint64
	Duration            uint64
	UsageBytesUp        uint64
	UsageBytesDown      uint64
	Timestamp           uint64
	TimestampStart      uint64
	PredictedRtt        float32
	PredictedJitter     float32
	PredictedPacketLoss float32
	RouteChanged        bool
	NetworkNext         bool
	Initial             bool
	EnvelopeBytesUp     uint64
	EnvelopeBytesDown   uint64
}

type BillingRouteHop struct {
	RelayOrRelayLinkId *EntityId
	SellerId           *EntityId
	PriceIngress       int64
	PriceEgress        int64
}

type RouteRequest struct {
	BuyerId                   *EntityId
	SessionId                 uint64
	UserId                    uint64
	PlatformId                uint64
	DirectRtt                 float32
	DirectJitter              float32
	DirectPacketLoss          float32
	NextRtt                   float32
	NextJitter                float32
	NextPacketLoss            float32
	ClientIpAddress           *Address
	ServerIpAddress           *Address
	ClientRoutePublicKey      []byte
	Tag                       uint64
	NearRelays                []*NearRelay
	ConnectionType            SessionConnectionType
	ServerRoutePublicKey      []byte
	DatacenterId              *EntityId
	ServerPrivateIpAddress    *Address
	SequenceNumber            uint64
	FallbackToDirect          bool
	VersionMajor              int32
	VersionMinor              int32
	VersionPatch              int32
	UsageKbpsUp               uint32
	UsageKbpsDown             uint32
	Location                  *Location
	IssuedNearRelays          []*EntityId
	OnNetworkNext             bool
	Flagged                   bool
	TryBeforeYouBuy           bool
	PacketsLostClientToServer uint64
	PacketsLostServerToClient uint64
}

type EntityId struct {
	Kind string
	Name string
}

type Address struct {
	Ip        []byte
	Type      Address_Type
	Port      uint32
	Formatted string
}

type NearRelay struct {
	RelayId    *id.EntityId
	Rtt        float64
	Jitter     float64
	PacketLoss float64
}

// This mirrors next_internal.h
type SessionConnectionType int32

const (
	SessionConnectionType_SESSION_CONNECTION_TYPE_UNKNOWN  SessionConnectionType = 0
	SessionConnectionType_SESSION_CONNECTION_TYPE_WIRED    SessionConnectionType = 1
	SessionConnectionType_SESSION_CONNECTION_TYPE_WIFI     SessionConnectionType = 2
	SessionConnectionType_SESSION_CONNECTION_TYPE_CELLULAR SessionConnectionType = 3
)

type Address_Type int32

const (
	Address_NONE Address_Type = 0
	Address_IPV4 Address_Type = 1
	Address_IPV6 Address_Type = 2
)

type Location struct {
	CountryCode string
	Country     string
	Region      string
	City        string
	Latitude    float32
	Longitude   float32
	Isp         string
	Asn         int64
}
