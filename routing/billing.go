package routing

import (
	"context"
	"fmt"
	"net"
	"os"
	"runtime"

	"cloud.google.com/go/pubsub"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/golang/protobuf/proto"
)

const BillingSliceSeconds = 10

type RouteState struct {
	SessionId       uint64 `protobuf:"fixed64,1,opt,name=sessionId,proto3" json:"sessionId,omitempty"`
	SessionVersion  uint32 `protobuf:"varint,2,opt,name=sessionVersion,proto3" json:"sessionVersion,omitempty"`
	SessionFlags    uint32 `protobuf:"varint,3,opt,name=sessionFlags,proto3" json:"sessionFlags,omitempty"`
	RouteHash       uint64 `protobuf:"fixed64,4,opt,name=routeHash,proto3" json:"routeHash,omitempty"`
	TimestampStart  uint64 `protobuf:"fixed64,5,opt,name=timestampStart,proto3" json:"timestampStart,omitempty"`
	TimestampExpire uint64 `protobuf:"fixed64,6,opt,name=timestampExpire,proto3" json:"timestampExpire,omitempty"`
	PacketsLost     uint64 `protobuf:"varint,7,opt,name=packetsLost,proto3" json:"packetsLost,omitempty"`
}

type BillingEntry struct {
	Request              *RouteRequest      `protobuf:"bytes,1,opt,name=request,proto3" json:"request,omitempty"`
	Route                []*BillingRouteHop `protobuf:"bytes,2,rep,name=route,proto3" json:"route,omitempty"`
	RouteDecision        uint64             `protobuf:"varint,3,opt,name=routeDecision,proto3" json:"routeDecision,omitempty"`
	Duration             uint64             `protobuf:"varint,4,opt,name=duration,proto3" json:"duration,omitempty"`
	UsageBytesUp         uint64             `protobuf:"varint,5,opt,name=usageBytesUp,proto3" json:"usageBytesUp,omitempty"`
	UsageBytesDown       uint64             `protobuf:"varint,6,opt,name=usageBytesDown,proto3" json:"usageBytesDown,omitempty"`
	Timestamp            uint64             `protobuf:"varint,7,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	TimestampStart       uint64             `protobuf:"varint,8,opt,name=timestampStart,proto3" json:"timestampStart,omitempty"`
	PredictedRtt         float32            `protobuf:"fixed32,9,opt,name=predictedRtt,proto3" json:"predictedRtt,omitempty"`
	PredictedJitter      float32            `protobuf:"fixed32,10,opt,name=predictedJitter,proto3" json:"predictedJitter,omitempty"`
	PredictedPacketLoss  float32            `protobuf:"fixed32,11,opt,name=predictedPacketLoss,proto3" json:"predictedPacketLoss,omitempty"`
	RouteChanged         bool               `protobuf:"varint,12,opt,name=routeChanged,proto3" json:"routeChanged,omitempty"`
	NetworkNextAvailable bool               `protobuf:"varint,13,opt,name=networkNextAvailable,proto3" json:"networkNextAvailable,omitempty"`
	Initial              bool               `protobuf:"varint,14,opt,name=initial,proto3" json:"initial,omitempty"`
	EnvelopeBytesUp      uint64             `protobuf:"varint,15,opt,name=envelopeBytesUp,proto3" json:"envelopeBytesUp,omitempty"`
	EnvelopeBytesDown    uint64             `protobuf:"varint,16,opt,name=envelopeBytesDown,proto3" json:"envelopeBytesDown,omitempty"`
	ConsideredRoutes     []*BillingRoute    `protobuf:"bytes,17,rep,name=consideredRoutes,proto3" json:"consideredRoutes,omitempty"`
	AcceptableRoutes     []*BillingRoute    `protobuf:"bytes,18,rep,name=acceptableRoutes,proto3" json:"acceptableRoutes,omitempty"`
	SameRoute            bool               `protobuf:"varint,19,opt,name=sameRoute,proto3" json:"sameRoute,omitempty"`
	OnNetworkNext        bool               `protobuf:"varint,20,opt,name=onNetworkNext,proto3" json:"onNetworkNext,omitempty"`
	SliceFlags           uint64             `protobuf:"varint,21,opt,name=sliceFlags,proto3" json:"sliceFlags,omitempty"`
}

func (entry *BillingEntry) Reset() {
	*entry = BillingEntry{}
}
func (entry *BillingEntry) String() string {
	return proto.CompactTextString(entry)
}
func (entry *BillingEntry) ProtoMessage() {}

type BillingRoute struct {
	Route []*BillingRouteHop `protobuf:"bytes,1,rep,name=route,proto3" json:"route,omitempty"`
}

func (route *BillingRoute) Reset() {
	*route = BillingRoute{}
}
func (route *BillingRoute) String() string {
	return proto.CompactTextString(route)
}
func (route *BillingRoute) ProtoMessage() {}

type BillingRouteHop struct {
	RelayId      *EntityId `protobuf:"bytes,1,opt,name=relayId,proto3" json:"relayId,omitempty"`
	SellerId     *EntityId `protobuf:"bytes,2,opt,name=sellerId,proto3" json:"sellerId,omitempty"`
	PriceIngress int64     `protobuf:"varint,3,opt,name=priceIngress,proto3" json:"priceIngress,omitempty"`
	PriceEgress  int64     `protobuf:"varint,4,opt,name=priceEgress,proto3" json:"priceEgress,omitempty"`
}

func (hop *BillingRouteHop) Reset() {
	*hop = BillingRouteHop{}
}
func (hop *BillingRouteHop) String() string {
	return proto.CompactTextString(hop)
}
func (hop *BillingRouteHop) ProtoMessage() {}

type RouteRequest struct {
	BuyerId                   *EntityId             `protobuf:"bytes,1,opt,name=buyerId,proto3" json:"buyerId,omitempty"`
	SessionId                 uint64                `protobuf:"varint,2,opt,name=sessionId,proto3" json:"sessionId,omitempty"`
	UserHash                  uint64                `protobuf:"varint,3,opt,name=userHash,proto3" json:"userHash,omitempty"`
	PlatformId                uint64                `protobuf:"varint,4,opt,name=platformId,proto3" json:"platformId,omitempty"`
	DirectRtt                 float32               `protobuf:"fixed32,5,opt,name=directRtt,proto3" json:"directRtt,omitempty"`
	DirectJitter              float32               `protobuf:"fixed32,6,opt,name=directJitter,proto3" json:"directJitter,omitempty"`
	DirectPacketLoss          float32               `protobuf:"fixed32,7,opt,name=directPacketLoss,proto3" json:"directPacketLoss,omitempty"`
	NextRtt                   float32               `protobuf:"fixed32,8,opt,name=nextRtt,proto3" json:"nextRtt,omitempty"`
	NextJitter                float32               `protobuf:"fixed32,9,opt,name=nextJitter,proto3" json:"nextJitter,omitempty"`
	NextPacketLoss            float32               `protobuf:"fixed32,10,opt,name=nextPacketLoss,proto3" json:"nextPacketLoss,omitempty"`
	ClientIpAddress           *Address              `protobuf:"bytes,11,opt,name=clientIpAddress,proto3" json:"clientIpAddress,omitempty"`
	ServerIpAddress           *Address              `protobuf:"bytes,12,opt,name=serverIpAddress,proto3" json:"serverIpAddress,omitempty"`
	ClientRoutePublicKey      []byte                `protobuf:"bytes,13,opt,name=clientRoutePublicKey,proto3" json:"clientRoutePublicKey,omitempty"`
	Tag                       uint64                `protobuf:"varint,14,opt,name=tag,proto3" json:"tag,omitempty"`
	NearRelays                []*NearRelay          `protobuf:"bytes,15,rep,name=nearRelays,proto3" json:"nearRelays,omitempty"`
	ConnectionType            SessionConnectionType `protobuf:"varint,16,opt,name=connectionType,proto3,enum=session.SessionConnectionType" json:"connectionType,omitempty"`
	ServerRoutePublicKey      []byte                `protobuf:"bytes,17,opt,name=serverRoutePublicKey,proto3" json:"serverRoutePublicKey,omitempty"`
	DatacenterId              *EntityId             `protobuf:"bytes,18,opt,name=datacenterId,proto3" json:"datacenterId,omitempty"`
	ServerPrivateIpAddress    *Address              `protobuf:"bytes,19,opt,name=serverPrivateIpAddress,proto3" json:"serverPrivateIpAddress,omitempty"`
	SequenceNumber            uint64                `protobuf:"varint,20,opt,name=sequenceNumber,proto3" json:"sequenceNumber,omitempty"`
	FallbackToDirect          bool                  `protobuf:"varint,21,opt,name=fallbackToDirect,proto3" json:"fallbackToDirect,omitempty"`
	VersionMajor              int32                 `protobuf:"varint,22,opt,name=versionMajor,proto3" json:"versionMajor,omitempty"`
	VersionMinor              int32                 `protobuf:"varint,23,opt,name=versionMinor,proto3" json:"versionMinor,omitempty"`
	VersionPatch              int32                 `protobuf:"varint,24,opt,name=versionPatch,proto3" json:"versionPatch,omitempty"`
	UsageKbpsUp               uint32                `protobuf:"varint,25,opt,name=usageKbpsUp,proto3" json:"usageKbpsUp,omitempty"`
	UsageKbpsDown             uint32                `protobuf:"varint,26,opt,name=usageKbpsDown,proto3" json:"usageKbpsDown,omitempty"`
	BillingLocation           *BillingLocation      `protobuf:"bytes,27,opt,name=BillingLocation,proto3" json:"BillingLocation,omitempty"`
	OnNetworkNext             bool                  `protobuf:"varint,29,opt,name=onNetworkNext,proto3" json:"onNetworkNext,omitempty"`
	Flagged                   bool                  `protobuf:"varint,30,opt,name=flagged,proto3" json:"flagged,omitempty"`
	TryBeforeYouBuy           bool                  `protobuf:"varint,31,opt,name=tryBeforeYouBuy,proto3" json:"tryBeforeYouBuy,omitempty"`
	PacketsLostClientToServer uint64                `protobuf:"varint,32,opt,name=packetsLostClientToServer,proto3" json:"packetsLostClientToServer,omitempty"`
	PacketsLostServerToClient uint64                `protobuf:"varint,33,opt,name=packetsLostServerToClient,proto3" json:"packetsLostServerToClient,omitempty"`
	FallbackFlags             uint32                `protobuf:"varint,34,opt,name=fallbackFlags,proto3" json:"fallbackFlags,omitempty"`
	IssuedNearRelays          []*IssuedNearRelay    `protobuf:"bytes,35,rep,name=issuedNearRelays,proto3" json:"issuedNearRelays,omitempty"`
}

func (req *RouteRequest) Reset() {
	*req = RouteRequest{}
}
func (req *RouteRequest) String() string {
	return proto.CompactTextString(req)
}
func (req *RouteRequest) ProtoMessage() {}

type EntityId struct {
	Kind string `protobuf:"bytes,1,opt,name=kind,proto3" json:"kind,omitempty"`
	Name string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
}

func (id *EntityId) Reset() {
	*id = EntityId{}
}
func (id *EntityId) String() string {
	return proto.CompactTextString(id)
}
func (id *EntityId) ProtoMessage() {}

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
	return proto.CompactTextString(addr)
}
func (addr *Address) ProtoMessage() {}

func udpAddrToAddress(addr net.UDPAddr) *Address {
	if addr.IP == nil {
		return &Address{
			Ip:        nil,
			Type:      Address_NONE,
			Port:      0,
			Formatted: "",
		}
	}

	ipv4 := addr.IP.To4()
	if ipv4 == nil {
		ipv6 := addr.IP.To16()
		if ipv6 == nil {
			return &Address{
				Ip:        nil,
				Type:      Address_NONE,
				Port:      0,
				Formatted: "",
			}
		}

		return &Address{
			Ip:        []byte(ipv6),
			Type:      Address_IPV6,
			Port:      uint32(addr.Port),
			Formatted: addr.String(),
		}
	}

	return &Address{
		Ip:        []byte(ipv4),
		Type:      Address_IPV4,
		Port:      uint32(addr.Port),
		Formatted: addr.String(),
	}
}

type NearRelay struct {
	RelayId    *EntityId `protobuf:"bytes,1,opt,name=relayId,proto3" json:"relayId,omitempty"`
	Rtt        float64   `protobuf:"fixed64,2,opt,name=rtt,proto3" json:"rtt,omitempty"`
	Jitter     float64   `protobuf:"fixed64,3,opt,name=jitter,proto3" json:"jitter,omitempty"`
	PacketLoss float64   `protobuf:"fixed64,4,opt,name=packetLoss,proto3" json:"packetLoss,omitempty"`
}

type IssuedNearRelay struct {
	Index          int32     `protobuf:"varint,1,opt,name=index,proto3" json:"index,omitempty"`
	RelayId        *EntityId `protobuf:"bytes,2,opt,name=relayId,proto3" json:"relayId,omitempty"`
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

type BillingLocation struct {
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

func (loc *BillingLocation) Reset() {
	*loc = BillingLocation{}
}
func (loc *BillingLocation) String() string {
	return proto.CompactTextString(loc)
}
func (loc *BillingLocation) ProtoMessage() {}

type BillingClient interface {
	Init(ctx context.Context, logger log.Logger)
	Send(ctx context.Context, sessionID uint64, entry *BillingEntry) error
}

type GooglePubSubClient struct {
	topics  []*pubsub.Topic
	results []chan *pubsub.PublishResult
}

func (billing *GooglePubSubClient) Init(ctx context.Context, logger log.Logger) {
	projectID, ok := os.LookupEnv("GOOGLE_CLOUD_BILLING_PROJECT")
	if !ok {
		level.Warn(logger).Log("msg", "billing GCP project env var not set, billing will not be processed.")
		return
	}

	pubsubClients := 4
	billing.topics = make([]*pubsub.Topic, pubsubClients)
	billing.results = make([]chan *pubsub.PublishResult, pubsubClients)

	billingTopicID, ok := os.LookupEnv("BILLING_PUBSUB_TOPIC")
	if !ok {
		level.Warn(logger).Log("msg", "billing GCP topic env var not set, billing will not be processed.")
		return
	}

	for i := 0; i < pubsubClients; i++ {
		client, err := pubsub.NewClient(ctx, projectID)
		if err != nil {
			level.Error(logger).Log("msg", "could not create pubsub client", "index", i, "err", err)
			continue
		}

		billing.topics[i] = client.Topic(billingTopicID)
		billing.topics[i].PublishSettings.NumGoroutines = (25 * runtime.GOMAXPROCS(0)) / pubsubClients
		billing.results[i] = make(chan *pubsub.PublishResult, 10000*60*10) // 10,000 messages per second for 10 minutes
		go printPubSubErrors(ctx, logger, billing.results[i])
	}
}

func (billing *GooglePubSubClient) Send(ctx context.Context, sessionID uint64, entry *BillingEntry) error {
	data, err := proto.Marshal(entry)
	if err != nil {
		return err
	}

	index := sessionID % uint64(len(billing.topics))
	topic := billing.topics[index]
	if topic == nil {
		return fmt.Errorf("billing: topic %v not initialized", index)
	}

	resultChannel := billing.results[index]
	if resultChannel == nil {
		return fmt.Errorf("billing: result channel %v not initialized", index)
	}

	result := topic.Publish(ctx, &pubsub.Message{Data: data})
	resultChannel <- result
	return nil
}

func printPubSubErrors(ctx context.Context, logger log.Logger, results chan *pubsub.PublishResult) {
	for result := range results {
		_, err := result.Get(ctx)
		if err != nil {
			level.Error(logger).Log("billing", "failed to publish to pub/sub", "err", err)
		} else {
			level.Debug(logger).Log("billing", "successfully pushed billing data")
		}
	}
}
