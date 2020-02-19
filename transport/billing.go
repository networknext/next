package transport

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
	"github.com/networknext/backend/encoding"
	"github.com/networknext/backend/routing"
)

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

func (entry *BillingEntry) Reset() {
	*entry = BillingEntry{}
}
func (entry *BillingEntry) String() string {
	return proto.CompactTextString(entry)
}
func (entry *BillingEntry) ProtoMessage() {}

type BillingRouteHop struct {
	RelayOrRelayLinkId *EntityId
	SellerId           *EntityId
	PriceIngress       int64
	PriceEgress        int64
}

func (hop *BillingRouteHop) Reset() {
	*hop = BillingRouteHop{}
}
func (hop *BillingRouteHop) String() string {
	return proto.CompactTextString(hop)
}
func (hop *BillingRouteHop) ProtoMessage() {}

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

func (req *RouteRequest) Reset() {
	*req = RouteRequest{}
}
func (req *RouteRequest) String() string {
	return proto.CompactTextString(req)
}
func (req *RouteRequest) ProtoMessage() {}

type EntityId struct {
	Kind string
	Name string
}

type Address_Type uint32

const (
	Address_NONE Address_Type = 0
	Address_IPV4 Address_Type = 1
	Address_IPV6 Address_Type = 2
)

// Max length of an IPv4-mapped IPv6 address is 45 characters
const Address_FORMATTED_MAX_LENGTH = 45

type Address struct {
	Ip        []byte
	Type      Address_Type
	Port      uint32
	Formatted string
}

func (addr Address) MarshalBinary() ([]byte, error) {
	length := len(addr.Ip) + 4 + 4 + len(addr.Formatted)
	data := make([]byte, length)

	index := 0
	encoding.WriteBytes(data, &index, addr.Ip, len(addr.Ip))
	encoding.WriteUint32(data, &index, uint32(addr.Type))
	encoding.WriteUint32(data, &index, addr.Port)
	encoding.WriteString(data, &index, addr.Formatted, Address_FORMATTED_MAX_LENGTH)

	return data, nil
}

func (addr *Address) UnmarshalBinary(data []byte) error {
	index := 0
	if !(encoding.ReadBytes(data, &index, &addr.Ip, len) &&
		encoding.ReadString(data, &index, &loc.Country, math.MaxInt32) && // TODO define an actual limit on this
		encoding.ReadString(data, &index, &loc.Region, math.MaxInt32) &&
		encoding.ReadString(data, &index, &loc.City) &&
		encoding.ReadFloat32(data, &index, &loc.Latitude) &&
		encoding.ReadFloat32(data, &index, &loc.Longitude) {
		return errors.New("Invalid Location")
	}

	return nil
}

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
		} else {
			return &Address{
				Ip:        []byte(ipv6),
				Type:      Address_IPV6,
				Port:      uint32(addr.Port),
				Formatted: addr.String(),
			}
		}
	} else {
		return &Address{
			Ip:        []byte(ipv4),
			Type:      Address_IPV4,
			Port:      uint32(addr.Port),
			Formatted: addr.String(),
		}
	}
}

type NearRelay struct {
	RelayId    *EntityId
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

type Location struct {
	Country   string
	Region    string
	City      string
	Latitude  float64
	Longitude float64
	Isp 	  string
	Asn 	  int64
	Continent string
}

type BillingClient interface {
	Init()
	Send(ctx context.Context, sessionID uint64, entry *BillingEntry) error
}

type GooglePubSubClient struct {
	topics  []*pubsub.Topic
	results []chan *pubsub.PublishResult
}

func (billing *GooglePubSubClient) Init(logger log.Logger) {
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
		client, err := pubsub.NewClient(context.Background(), projectID)
		if err != nil {
			level.Error(logger).Log("msg", "could not create pubsub client", "index", i, "err", err)
			continue
		}

		billing.topics[i] = client.Topic(billingTopicID)
		billing.topics[i].PublishSettings.NumGoroutines = (25 * runtime.GOMAXPROCS(0)) / pubsubClients
		billing.results[i] = make(chan *pubsub.PublishResult, 10000*60*10) // 10,000 messages per second for 10 minutes
		go printPubSubErrors(logger, billing.results[i])
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

func printPubSubErrors(logger log.Logger, results chan *pubsub.PublishResult) {
	for result := range results {
		_, err := result.Get(context.Background())
		if err != nil {
			level.Error(logger).Log("billing", "failed to publish to pub/sub", "err", err)
		}
	}
}
