package billing

import (
	"fmt"
	"net"

	"github.com/golang/protobuf/proto"
)

const BillingSliceSeconds = 10

type RouteSliceFlag uint64

const (
	RouteSliceFlagNone                RouteSliceFlag = 0
	RouteSliceFlagNext                RouteSliceFlag = 1 << 1
	RouteSliceFlagReported            RouteSliceFlag = 1 << 2
	RouteSliceFlagVetoed              RouteSliceFlag = 1 << 3
	RouteSliceFlagFallbackToDirect    RouteSliceFlag = 1 << 4
	RouteSliceFlagPacketLossMultipath RouteSliceFlag = 1 << 5
	RouteSliceFlagJitterMultipath     RouteSliceFlag = 1 << 6
	RouteSliceFlagRTTMultipath        RouteSliceFlag = 1 << 7
)

type RouteState struct {
	SessionID       uint64 `protobuf:"fixed64,1,opt,name=sessionId,proto3" json:"sessionId,omitempty"`
	SessionVersion  uint32 `protobuf:"varint,2,opt,name=sessionVersion,proto3" json:"sessionVersion,omitempty"`
	SessionFlags    uint32 `protobuf:"varint,3,opt,name=sessionFlags,proto3" json:"sessionFlags,omitempty"`
	RouteHash       uint64 `protobuf:"fixed64,4,opt,name=routeHash,proto3" json:"routeHash,omitempty"`
	TimestampStart  uint64 `protobuf:"fixed64,5,opt,name=timestampStart,proto3" json:"timestampStart,omitempty"`
	TimestampExpire uint64 `protobuf:"fixed64,6,opt,name=timestampExpire,proto3" json:"timestampExpire,omitempty"`
	PacketsLost     uint64 `protobuf:"varint,7,opt,name=packetsLost,proto3" json:"packetsLost,omitempty"`
}

type Entry struct {
	Request              *RouteRequest `protobuf:"bytes,1,opt,name=request,proto3" json:"request,omitempty"`
	Route                []*RouteHop   `protobuf:"bytes,2,rep,name=route,proto3" json:"route,omitempty"`
	RouteDecision        uint64        `protobuf:"varint,3,opt,name=routeDecision,proto3" json:"routeDecision,omitempty"`
	Duration             uint64        `protobuf:"varint,4,opt,name=duration,proto3" json:"duration,omitempty"`
	UsageBytesUp         uint64        `protobuf:"varint,5,opt,name=usageBytesUp,proto3" json:"usageBytesUp,omitempty"`
	UsageBytesDown       uint64        `protobuf:"varint,6,opt,name=usageBytesDown,proto3" json:"usageBytesDown,omitempty"`
	Timestamp            uint64        `protobuf:"varint,7,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	TimestampStart       uint64        `protobuf:"varint,8,opt,name=timestampStart,proto3" json:"timestampStart,omitempty"`
	PredictedRTT         float32       `protobuf:"fixed32,9,opt,name=predictedRtt,proto3" json:"predictedRtt,omitempty"`
	PredictedJitter      float32       `protobuf:"fixed32,10,opt,name=predictedJitter,proto3" json:"predictedJitter,omitempty"`
	PredictedPacketLoss  float32       `protobuf:"fixed32,11,opt,name=predictedPacketLoss,proto3" json:"predictedPacketLoss,omitempty"`
	RouteChanged         bool          `protobuf:"varint,12,opt,name=routeChanged,proto3" json:"routeChanged,omitempty"`
	NetworkNextAvailable bool          `protobuf:"varint,13,opt,name=networkNextAvailable,proto3" json:"networkNextAvailable,omitempty"`
	Initial              bool          `protobuf:"varint,14,opt,name=initial,proto3" json:"initial,omitempty"`
	EnvelopeBytesUp      uint64        `protobuf:"varint,15,opt,name=envelopeBytesUp,proto3" json:"envelopeBytesUp,omitempty"`
	EnvelopeBytesDown    uint64        `protobuf:"varint,16,opt,name=envelopeBytesDown,proto3" json:"envelopeBytesDown,omitempty"`
	ConsideredRoutes     []*Route      `protobuf:"bytes,17,rep,name=consideredRoutes,proto3" json:"consideredRoutes,omitempty"`
	AcceptableRoutes     []*Route      `protobuf:"bytes,18,rep,name=acceptableRoutes,proto3" json:"acceptableRoutes,omitempty"`
	SameRoute            bool          `protobuf:"varint,19,opt,name=sameRoute,proto3" json:"sameRoute,omitempty"`
	OnNetworkNext        bool          `protobuf:"varint,20,opt,name=onNetworkNext,proto3" json:"onNetworkNext,omitempty"`
	SliceFlags           uint64        `protobuf:"varint,21,opt,name=sliceFlags,proto3" json:"sliceFlags,omitempty"`
}

func (entry *Entry) Reset() {
	*entry = Entry{}
}
func (entry *Entry) String() string {
	return proto.CompactTextString(entry)
}
func (entry *Entry) ProtoMessage() {}

type Route struct {
	Route []*RouteHop `protobuf:"bytes,1,rep,name=route,proto3" json:"route,omitempty"`
}

func (route *Route) Reset() {
	*route = Route{}
}
func (route *Route) String() string {
	return proto.CompactTextString(route)
}
func (route *Route) ProtoMessage() {}

type RouteHop struct {
	RelayID      *EntityID `protobuf:"bytes,1,opt,name=relayId,proto3" json:"relayId,omitempty"`
	SellerID     *EntityID `protobuf:"bytes,2,opt,name=sellerId,proto3" json:"sellerId,omitempty"`
	PriceIngress int64     `protobuf:"varint,3,opt,name=priceIngress,proto3" json:"priceIngress,omitempty"`
	PriceEgress  int64     `protobuf:"varint,4,opt,name=priceEgress,proto3" json:"priceEgress,omitempty"`
}

func (hop *RouteHop) Reset() {
	*hop = RouteHop{}
}
func (hop *RouteHop) String() string {
	return proto.CompactTextString(hop)
}
func (hop *RouteHop) ProtoMessage() {}

type RouteRequest struct {
	BuyerID                   *EntityID             `protobuf:"bytes,1,opt,name=buyerId,proto3" json:"buyerId,omitempty"`
	SessionID                 uint64                `protobuf:"varint,2,opt,name=sessionId,proto3" json:"sessionId,omitempty"`
	UserHash                  uint64                `protobuf:"varint,3,opt,name=userHash,proto3" json:"userHash,omitempty"`
	PlatformID                uint64                `protobuf:"varint,4,opt,name=platformId,proto3" json:"platformId,omitempty"`
	DirectRTT                 float32               `protobuf:"fixed32,5,opt,name=directRtt,proto3" json:"directRtt,omitempty"`
	DirectJitter              float32               `protobuf:"fixed32,6,opt,name=directJitter,proto3" json:"directJitter,omitempty"`
	DirectPacketLoss          float32               `protobuf:"fixed32,7,opt,name=directPacketLoss,proto3" json:"directPacketLoss,omitempty"`
	NextRTT                   float32               `protobuf:"fixed32,8,opt,name=nextRtt,proto3" json:"nextRtt,omitempty"`
	NextJitter                float32               `protobuf:"fixed32,9,opt,name=nextJitter,proto3" json:"nextJitter,omitempty"`
	NextPacketLoss            float32               `protobuf:"fixed32,10,opt,name=nextPacketLoss,proto3" json:"nextPacketLoss,omitempty"`
	ClientIpAddress           *Address              `protobuf:"bytes,11,opt,name=clientIpAddress,proto3" json:"clientIpAddress,omitempty"`
	ServerIpAddress           *Address              `protobuf:"bytes,12,opt,name=serverIpAddress,proto3" json:"serverIpAddress,omitempty"`
	ClientRoutePublicKey      []byte                `protobuf:"bytes,13,opt,name=clientRoutePublicKey,proto3" json:"clientRoutePublicKey,omitempty"`
	Tag                       uint64                `protobuf:"varint,14,opt,name=tag,proto3" json:"tag,omitempty"`
	NearRelays                []*NearRelay          `protobuf:"bytes,15,rep,name=nearRelays,proto3" json:"nearRelays,omitempty"`
	ConnectionType            SessionConnectionType `protobuf:"varint,16,opt,name=connectionType,proto3,enum=session.SessionConnectionType" json:"connectionType,omitempty"`
	ServerRoutePublicKey      []byte                `protobuf:"bytes,17,opt,name=serverRoutePublicKey,proto3" json:"serverRoutePublicKey,omitempty"`
	DatacenterID              *EntityID             `protobuf:"bytes,18,opt,name=datacenterId,proto3" json:"datacenterId,omitempty"`
	ServerPrivateIpAddress    *Address              `protobuf:"bytes,19,opt,name=serverPrivateIpAddress,proto3" json:"serverPrivateIpAddress,omitempty"`
	SequenceNumber            uint64                `protobuf:"varint,20,opt,name=sequenceNumber,proto3" json:"sequenceNumber,omitempty"`
	FallbackToDirect          bool                  `protobuf:"varint,21,opt,name=fallbackToDirect,proto3" json:"fallbackToDirect,omitempty"`
	VersionMajor              int32                 `protobuf:"varint,22,opt,name=versionMajor,proto3" json:"versionMajor,omitempty"`
	VersionMinor              int32                 `protobuf:"varint,23,opt,name=versionMinor,proto3" json:"versionMinor,omitempty"`
	VersionPatch              int32                 `protobuf:"varint,24,opt,name=versionPatch,proto3" json:"versionPatch,omitempty"`
	UsageKbpsUp               uint32                `protobuf:"varint,25,opt,name=usageKbpsUp,proto3" json:"usageKbpsUp,omitempty"`
	UsageKbpsDown             uint32                `protobuf:"varint,26,opt,name=usageKbpsDown,proto3" json:"usageKbpsDown,omitempty"`
	Location                  *Location             `protobuf:"bytes,27,opt,name=location,proto3" json:"location,omitempty"`
	OnNetworkNext             bool                  `protobuf:"varint,29,opt,name=onNetworkNext,proto3" json:"onNetworkNext,omitempty"`
	Flagged                   bool                  `protobuf:"varint,30,opt,name=flagged,proto3" json:"flagged,omitempty"`
	TryBeforeYouBuy           bool                  `protobuf:"varint,31,opt,name=tryBeforeYouBuy,proto3" json:"tryBeforeYouBuy,omitempty"`
	PacketsLostClientToServer uint64                `protobuf:"varint,32,opt,name=packetsLostClientToServer,proto3" json:"packetsLostClientToServer,omitempty"`
	PacketsLostServerToClient uint64                `protobuf:"varint,33,opt,name=packetsLostServerToClient,proto3" json:"packetsLostServerToClient,omitempty"`
	FallbackFlags             uint32                `protobuf:"varint,34,opt,name=fallbackFlags,proto3" json:"fallbackFlags,omitempty"`
	IssuedNearRelays          []*IssuedNearRelay    `protobuf:"bytes,35,rep,name=issuedNearRelays,proto3" json:"issuedNearRelays,omitempty"`
	Committed                 bool                  `protobuf:"varint,36,opt,name=committed,proto3" json:"committed,omitempty"`
	UserFlags                 uint64                `protobuf:"varint,37,opt,name=userFlags,proto3" json:"userFlags,omitempty"`
}

func (req *RouteRequest) Reset() {
	*req = RouteRequest{}
}
func (req *RouteRequest) String() string {
	return proto.CompactTextString(req)
}
func (req *RouteRequest) ProtoMessage() {}

type EntityID struct {
	Kind string `protobuf:"bytes,1,opt,name=kind,proto3" json:"kind,omitempty"`
	Name string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
}

func (id *EntityID) Reset() {
	*id = EntityID{}
}
func (id *EntityID) String() string {
	return fmt.Sprintf("%s/%s", id.Kind, id.Name)
}
func (id *EntityID) ProtoMessage() {}

type Address_Type uint32

const (
	Address_NONE Address_Type = 0
	Address_IPV4 Address_Type = 1
	Address_IPV6 Address_Type = 2
)

// Max length of an IPv4-mapped IPv6 address is 45 characters
const Address_FORMATTED_MAX_LENGTH = 45

type Address struct {
	Ip        []byte       `protobuf:"bytes,1,opt,name=ip,proto3" json:"ip,omitempty"`
	Type      Address_Type `protobuf:"varint,2,opt,name=type,proto3,enum=address.Address_Type" json:"type,omitempty"`
	Port      uint32       `protobuf:"varint,3,opt,name=port,proto3" json:"port,omitempty"`
	Formatted string       `protobuf:"bytes,4,opt,name=formatted,proto3" json:"formatted,omitempty"`
}

func (addr *Address) Reset() {
	*addr = Address{}
}
func (addr *Address) String() string {
	return fmt.Sprintf("%s:%d", net.IP(addr.Ip).String(), addr.Port)
}
func (addr *Address) ProtoMessage() {}

type NearRelay struct {
	RelayID    *EntityID `protobuf:"bytes,1,opt,name=relayId,proto3" json:"relayId,omitempty"`
	RTT        float64   `protobuf:"fixed64,2,opt,name=rtt,proto3" json:"rtt,omitempty"`
	Jitter     float64   `protobuf:"fixed64,3,opt,name=jitter,proto3" json:"jitter,omitempty"`
	PacketLoss float64   `protobuf:"fixed64,4,opt,name=packetLoss,proto3" json:"packetLoss,omitempty"`
}

type IssuedNearRelay struct {
	Index          int32     `protobuf:"varint,1,opt,name=index,proto3" json:"index,omitempty"`
	RelayID        *EntityID `protobuf:"bytes,2,opt,name=relayId,proto3" json:"relayId,omitempty"`
	RelayIpAddress *Address  `protobuf:"bytes,3,opt,name=relayIpAddress,proto3" json:"relayIpAddress,omitempty"`
}

func (near *NearRelay) Reset() {
	*near = NearRelay{}
}
func (near *NearRelay) String() string {
	return proto.CompactTextString(near)
}
func (near *NearRelay) ProtoMessage() {}

// This mirrors next_internal.h
type SessionConnectionType int32

const (
	SessionConnectionType_SESSION_CONNECTION_TYPE_UNKNOWN  SessionConnectionType = 0
	SessionConnectionType_SESSION_CONNECTION_TYPE_WIRED    SessionConnectionType = 1
	SessionConnectionType_SESSION_CONNECTION_TYPE_WIFI     SessionConnectionType = 2
	SessionConnectionType_SESSION_CONNECTION_TYPE_CELLULAR SessionConnectionType = 3
)

type Location struct {
	CountryCode string  `protobuf:"bytes,1,opt,name=countryCode,proto3" json:"countryCode,omitempty"`
	Country     string  `protobuf:"bytes,2,opt,name=country,proto3" json:"country,omitempty"`
	Region      string  `protobuf:"bytes,3,opt,name=region,proto3" json:"region,omitempty"`
	City        string  `protobuf:"bytes,4,opt,name=city,proto3" json:"city,omitempty"`
	Latitude    float32 `protobuf:"fixed32,5,opt,name=latitude,proto3" json:"latitude,omitempty"`
	Longitude   float32 `protobuf:"fixed32,6,opt,name=longitude,proto3" json:"longitude,omitempty"`
	Isp         string  `protobuf:"bytes,7,opt,name=isp,proto3" json:"isp,omitempty"`
	Asn         int64   `protobuf:"varint,8,opt,name=asn,proto3" json:"asn,omitempty"`
	Continent   string  `protobuf:"bytes,9,opt,name=continent,proto3" json:"continent,omitempty"`
}

func (loc *Location) Reset() {
	*loc = Location{}
}
func (loc *Location) String() string {
	return proto.CompactTextString(loc)
}
func (loc *Location) ProtoMessage() {}
