package transport

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/networknext/backend/encoding"
	"github.com/networknext/backend/routing"
)

const (
	SessionCountDataVersion = 0
	SessionDataVersion      = 0
	SessionMetaVersion      = 0
	SessionSliceVersion     = 0
	SessionMapPointVersion  = 0
)

type SessionCountData struct {
	InstanceID                uint64
	TotalNumDirectSessions    uint64
	TotalNumNextSessions      uint64
	NumDirectSessionsPerBuyer map[uint64]uint64
	NumNextSessionsPerBuyer   map[uint64]uint64
}

func (s *SessionCountData) UnmarshalBinary(data []byte) error {
	index := 0

	var version uint32
	if !encoding.ReadUint32(data, &index, &version) {
		return errors.New("[SessionCountData] invalid read at version number")
	}

	if version > SessionCountDataVersion {
		return fmt.Errorf("unknown session count data version: %d", version)
	}

	if !encoding.ReadUint64(data, &index, &s.InstanceID) {
		return errors.New("[SessionCountData] invalid read at instance ID")
	}

	if !encoding.ReadUint64(data, &index, &s.TotalNumDirectSessions) {
		return errors.New("[SessionCountData] invalid read at total number of direct sessions")
	}

	if !encoding.ReadUint64(data, &index, &s.TotalNumNextSessions) {
		return errors.New("[SessionCountData] invalid read at total number of next sessions")
	}

	var directPerBuyerSize uint32
	if !encoding.ReadUint32(data, &index, &directPerBuyerSize) {
		return errors.New("[SessionCountData] invalid read at direct per buyer size")
	}

	directSessionPerBuyer := make(map[uint64]uint64)
	for i := uint32(0); i < directPerBuyerSize; i++ {
		var buyerID uint64
		if !encoding.ReadUint64(data, &index, &buyerID) {
			return errors.New("[SessionCountData] invalid read at direct buyer ID")
		}

		var count uint64
		if !encoding.ReadUint64(data, &index, &count) {
			return errors.New("[SessionCountData] invalid read at direct count")
		}

		directSessionPerBuyer[buyerID] = count
	}
	s.NumDirectSessionsPerBuyer = directSessionPerBuyer

	var nextPerBuyerSize uint32
	if !encoding.ReadUint32(data, &index, &nextPerBuyerSize) {
		return errors.New("[SessionCountData] invalid read at next per buyer size")
	}

	nextSessionPerBuyer := make(map[uint64]uint64)
	for i := uint32(0); i < nextPerBuyerSize; i++ {
		var buyerID uint64
		if !encoding.ReadUint64(data, &index, &buyerID) {
			return errors.New("[SessionCountData] invalid read at next buyer ID")
		}

		var count uint64
		if !encoding.ReadUint64(data, &index, &count) {
			return errors.New("[SessionCountData] invalid read at next count")
		}

		nextSessionPerBuyer[buyerID] = count
	}
	s.NumNextSessionsPerBuyer = nextSessionPerBuyer

	return nil
}

func (s SessionCountData) MarshalBinary() ([]byte, error) {
	data := make([]byte, s.Size())
	index := 0

	encoding.WriteUint32(data, &index, SessionCountDataVersion)

	encoding.WriteUint64(data, &index, s.InstanceID)
	encoding.WriteUint64(data, &index, s.TotalNumDirectSessions)
	encoding.WriteUint64(data, &index, s.TotalNumNextSessions)

	encoding.WriteUint32(data, &index, uint32(len(s.NumDirectSessionsPerBuyer)))
	for buyerID, count := range s.NumDirectSessionsPerBuyer {
		encoding.WriteUint64(data, &index, buyerID)
		encoding.WriteUint64(data, &index, count)
	}

	encoding.WriteUint32(data, &index, uint32(len(s.NumNextSessionsPerBuyer)))
	for buyerID, count := range s.NumNextSessionsPerBuyer {
		encoding.WriteUint64(data, &index, buyerID)
		encoding.WriteUint64(data, &index, count)
	}

	return data, nil
}

func (s SessionCountData) Size() uint64 {
	return 4 + 8 + 8 + 8 + 4 + 2*8*uint64(len(s.NumDirectSessionsPerBuyer)) + 4 + 2*8*uint64(len(s.NumNextSessionsPerBuyer))
}

type SessionPortalData struct {
	Meta  SessionMeta     `json:"meta"`
	Slice SessionSlice    `json:"slice"`
	Point SessionMapPoint `json:"point"`
}

func (s *SessionPortalData) UnmarshalBinary(data []byte) error {
	index := 0

	var version uint32
	if !encoding.ReadUint32(data, &index, &version) {
		return errors.New("[SessionPortalData] invalid read at version number")
	}

	if version > SessionDataVersion {
		return fmt.Errorf("unknown session data version: %d", version)
	}

	var metaSize uint32
	if !encoding.ReadUint32(data, &index, &metaSize) {
		return errors.New("[SessionPortalData] invalid read at meta size")
	}

	var metaBytes []byte
	if !encoding.ReadBytes(data, &index, &metaBytes, metaSize) {
		return errors.New("[SessionPortalData] invalid read at meta bytes")
	}

	if err := s.Meta.UnmarshalBinary(metaBytes); err != nil {
		return err
	}

	var sliceSize uint32
	if !encoding.ReadUint32(data, &index, &sliceSize) {
		return errors.New("[SessionPortalData] invalid read at slice size")
	}

	var sliceBytes []byte
	if !encoding.ReadBytes(data, &index, &sliceBytes, sliceSize) {
		return errors.New("[SessionPortalData] invalid read at slice bytes")
	}

	if err := s.Slice.UnmarshalBinary(sliceBytes); err != nil {
		return err
	}

	var pointSize uint32
	if !encoding.ReadUint32(data, &index, &pointSize) {
		return errors.New("[SessionPortalData] invalid read at map point size")
	}

	var pointBytes []byte
	if !encoding.ReadBytes(data, &index, &pointBytes, pointSize) {
		return errors.New("[SessionPortalData] invalid read at map point bytes")
	}

	if err := s.Point.UnmarshalBinary(pointBytes); err != nil {
		return err
	}

	return nil
}

func (s SessionPortalData) MarshalBinary() ([]byte, error) {
	data := make([]byte, s.Size())
	index := 0

	encoding.WriteUint32(data, &index, SessionDataVersion)

	metaBytes, err := s.Meta.MarshalBinary()
	if err != nil {
		return nil, err
	}

	encoding.WriteUint32(data, &index, uint32(len(metaBytes)))
	encoding.WriteBytes(data, &index, metaBytes, len(metaBytes))

	sliceBytes, err := s.Slice.MarshalBinary()
	if err != nil {
		return nil, err
	}

	encoding.WriteUint32(data, &index, uint32(len(sliceBytes)))
	encoding.WriteBytes(data, &index, sliceBytes, len(sliceBytes))

	pointBytes, err := s.Point.MarshalBinary()
	if err != nil {
		return nil, err
	}

	encoding.WriteUint32(data, &index, uint32(len(pointBytes)))
	encoding.WriteBytes(data, &index, pointBytes, len(pointBytes))

	return data, nil
}

func (s *SessionPortalData) Size() uint64 {
	return 4 + 4 + s.Meta.Size() + 4 + s.Slice.Size() + 4 + s.Point.Size()
}

type RelayHop struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

type NearRelayPortalData struct {
	ID          uint64        `json:"id"`
	Name        string        `json:"name"`
	ClientStats routing.Stats `json:"client_stats"`
}

type SessionMeta struct {
	ID              uint64                `json:"id"`
	UserHash        uint64                `json:"user_hash"`
	DatacenterName  string                `json:"datacenter_name"`
	DatacenterAlias string                `json:"datacenter_alias"`
	OnNetworkNext   bool                  `json:"on_network_next"`
	NextRTT         float64               `json:"next_rtt"`
	DirectRTT       float64               `json:"direct_rtt"`
	DeltaRTT        float64               `json:"delta_rtt"`
	Location        routing.Location      `json:"location"`
	ClientAddr      string                `json:"client_addr"`
	ServerAddr      string                `json:"server_addr"`
	Hops            []RelayHop            `json:"hops"`
	SDK             string                `json:"sdk"`
	Connection      uint8                 `json:"connection"`
	NearbyRelays    []NearRelayPortalData `json:"nearby_relays"`
	Platform        uint8                 `json:"platform"`
	BuyerID         uint64                `json:"customer_id"`
}

func (s *SessionMeta) UnmarshalBinary(data []byte) error {
	index := 0

	var version uint32
	if !encoding.ReadUint32(data, &index, &version) {
		return errors.New("[SessionMeta] invalid read at version number")
	}

	if version > SessionMetaVersion {
		return fmt.Errorf("unknown session meta version: %d", version)
	}

	if !encoding.ReadUint64(data, &index, &s.ID) {
		return errors.New("[SessionMeta] invalid read at session ID")
	}

	if !encoding.ReadUint64(data, &index, &s.UserHash) {
		return errors.New("[SessionMeta] invalid read at user hash")
	}

	if !encoding.ReadString(data, &index, &s.DatacenterName, math.MaxInt32) {
		return errors.New("[SessionMeta] invalid read at datacenter name")
	}

	if !encoding.ReadString(data, &index, &s.DatacenterAlias, math.MaxInt32) {
		return errors.New("[SessionMeta] invalid read at datacenter alias")
	}

	if !encoding.ReadBool(data, &index, &s.OnNetworkNext) {
		return errors.New("[SessionMeta] invalid read at on network next")
	}

	if !encoding.ReadFloat64(data, &index, &s.NextRTT) {
		return errors.New("[SessionMeta] invalid read at next RTT")
	}

	if !encoding.ReadFloat64(data, &index, &s.DirectRTT) {
		return errors.New("[SessionMeta] invalid read at direct RTT")
	}

	if !encoding.ReadFloat64(data, &index, &s.DeltaRTT) {
		return errors.New("[SessionMeta] invalid read at delta RTT")
	}

	var locationSize uint32
	if !encoding.ReadUint32(data, &index, &locationSize) {
		return errors.New("[SessionMeta] invalid read at location size")
	}

	var locationBytes []byte
	if !encoding.ReadBytes(data, &index, &locationBytes, locationSize) {
		return errors.New("[SessionMeta] invalid read at location bytes")
	}

	if err := s.Location.UnmarshalBinary(locationBytes); err != nil {
		return err
	}

	if !encoding.ReadString(data, &index, &s.ClientAddr, math.MaxInt32) {
		return errors.New("[SessionMeta] invalid read at client address")
	}

	if !encoding.ReadString(data, &index, &s.ServerAddr, math.MaxInt32) {
		return errors.New("[SessionMeta] invalid read at server address")
	}

	var hopsCount uint32
	if !encoding.ReadUint32(data, &index, &hopsCount) {
		return errors.New("[SessionMeta] invalid read at relay hops count")
	}

	s.Hops = make([]RelayHop, hopsCount)
	for i := uint32(0); i < hopsCount; i++ {
		if !encoding.ReadUint64(data, &index, &s.Hops[i].ID) {
			return errors.New("[SessionMeta] invalid read at relay hops relay ID")
		}

		if !encoding.ReadString(data, &index, &s.Hops[i].Name, math.MaxInt32) {
			return errors.New("[SessionMeta] invalid read at relay hops relay name")
		}
	}

	if !encoding.ReadString(data, &index, &s.SDK, math.MaxInt32) {
		return errors.New("[SessionMeta] invalid read at SDK version")
	}

	if !encoding.ReadUint8(data, &index, &s.Connection) {
		return errors.New("[SessionMeta] invalid read at connection type")
	}

	var nearbyRelaysCount uint32
	if !encoding.ReadUint32(data, &index, &nearbyRelaysCount) {
		return errors.New("[SessionMeta] invalid read at nearby relays count")
	}

	s.NearbyRelays = make([]NearRelayPortalData, nearbyRelaysCount)
	for i := uint32(0); i < nearbyRelaysCount; i++ {
		if !encoding.ReadUint64(data, &index, &s.NearbyRelays[i].ID) {
			return errors.New("[SessionMeta] invalid read at nearby relays relay ID")
		}

		if !encoding.ReadString(data, &index, &s.NearbyRelays[i].Name, math.MaxInt32) {
			return errors.New("[SessionMeta] invalid read at nearby relays relay name")
		}

		if !encoding.ReadFloat64(data, &index, &s.NearbyRelays[i].ClientStats.RTT) {
			return errors.New("[SessionMeta] invalid read at nearby relays relay RTT")
		}

		if !encoding.ReadFloat64(data, &index, &s.NearbyRelays[i].ClientStats.Jitter) {
			return errors.New("[SessionMeta] invalid read at nearby relays relay jitter")
		}

		if !encoding.ReadFloat64(data, &index, &s.NearbyRelays[i].ClientStats.PacketLoss) {
			return errors.New("[SessionMeta] invalid read at nearby relays relay packet loss")
		}
	}

	if !encoding.ReadUint8(data, &index, &s.Platform) {
		return errors.New("[SessionMeta] invalid read at platform type")
	}

	if !encoding.ReadUint64(data, &index, &s.BuyerID) {
		return errors.New("[SessionMeta] invalid read at buyer ID")
	}

	return nil
}

func (s SessionMeta) MarshalBinary() ([]byte, error) {
	data := make([]byte, s.Size())
	index := 0

	encoding.WriteUint32(data, &index, SessionMetaVersion)
	encoding.WriteUint64(data, &index, s.ID)
	encoding.WriteUint64(data, &index, s.UserHash)
	encoding.WriteString(data, &index, s.DatacenterName, uint32(len(s.DatacenterName)))
	encoding.WriteString(data, &index, s.DatacenterAlias, uint32(len(s.DatacenterAlias)))
	encoding.WriteBool(data, &index, s.OnNetworkNext)
	encoding.WriteFloat64(data, &index, s.NextRTT)
	encoding.WriteFloat64(data, &index, s.DirectRTT)
	encoding.WriteFloat64(data, &index, s.DeltaRTT)

	locationBytes, err := s.Location.MarshalBinary()
	if err != nil {
		return nil, err
	}
	encoding.WriteUint32(data, &index, uint32(len(locationBytes)))
	encoding.WriteBytes(data, &index, locationBytes, len(locationBytes))

	encoding.WriteString(data, &index, s.ClientAddr, uint32(len(s.ClientAddr)))
	encoding.WriteString(data, &index, s.ServerAddr, uint32(len(s.ServerAddr)))

	encoding.WriteUint32(data, &index, uint32(len(s.Hops)))
	for _, hop := range s.Hops {
		encoding.WriteUint64(data, &index, hop.ID)
		encoding.WriteString(data, &index, hop.Name, uint32(len(hop.Name)))
	}

	encoding.WriteString(data, &index, s.SDK, uint32(len(s.SDK)))
	encoding.WriteUint8(data, &index, s.Connection)

	encoding.WriteUint32(data, &index, uint32(len(s.NearbyRelays)))
	for _, nearRelayData := range s.NearbyRelays {
		encoding.WriteUint64(data, &index, nearRelayData.ID)
		encoding.WriteString(data, &index, nearRelayData.Name, uint32(len(nearRelayData.Name)))
		encoding.WriteFloat64(data, &index, nearRelayData.ClientStats.RTT)
		encoding.WriteFloat64(data, &index, nearRelayData.ClientStats.Jitter)
		encoding.WriteFloat64(data, &index, nearRelayData.ClientStats.PacketLoss)
	}

	encoding.WriteUint8(data, &index, s.Platform)
	encoding.WriteUint64(data, &index, s.BuyerID)

	return data, nil
}

func (s SessionMeta) Size() uint64 {
	var relayHopsSize uint64
	for _, hop := range s.Hops {
		relayHopsSize += 8 + 4 + uint64(len(hop.Name))
	}

	var nearbyRelaysSize uint64
	for _, nearRelayData := range s.NearbyRelays {
		nearbyRelaysSize += 8 + 4 + uint64(len(nearRelayData.Name)) + 8 + 8 + 8
	}

	return 4 + 8 + 8 + 4 + uint64(len(s.DatacenterName)) + 4 + uint64(len(s.DatacenterAlias)) + 1 + 8 + 8 + 8 + 4 + s.Location.Size() +
		4 + uint64(len(s.ClientAddr)) + 4 + uint64(len(s.ServerAddr)) + (4 + relayHopsSize) + 4 + uint64(len(s.SDK)) + 1 + (4 + nearbyRelaysSize) + 1 + 8
}

func (s SessionMeta) MarshalJSON() ([]byte, error) {
	fields := map[string]interface{}{}

	fields["id"] = fmt.Sprintf("%016x", s.ID)
	fields["user_hash"] = fmt.Sprintf("%016x", s.UserHash)
	fields["datacenter_name"] = s.DatacenterName
	fields["datacenter_alias"] = s.DatacenterAlias
	fields["on_network_next"] = s.OnNetworkNext
	fields["next_rtt"] = s.NextRTT
	fields["direct_rtt"] = s.DirectRTT
	fields["delta_rtt"] = s.DeltaRTT
	fields["location"] = s.Location
	fields["client_addr"] = s.ClientAddr
	fields["server_addr"] = s.ServerAddr
	fields["hops"] = s.Hops
	fields["sdk"] = s.SDK
	fields["connection"] = ConnectionTypeText(int32(s.Connection))
	fields["nearby_relays"] = s.NearbyRelays
	fields["platform"] = PlatformTypeText(uint64(s.Platform))
	fields["customer_id"] = fmt.Sprintf("%016x", s.BuyerID)

	return json.Marshal(fields)
}

func (s *SessionMeta) Anonymise() {
	s.ServerAddr = ObscureString(s.ServerAddr, ".", -1)
	s.BuyerID = 0
	s.NearbyRelays = []NearRelayPortalData{}
	s.Hops = []RelayHop{}
	s.DatacenterAlias = ""
}

func ObscureString(source string, delim string, count int) string {
	numPieces := count
	pieces := strings.Split(source, delim)

	if numPieces == -1 {
		numPieces = len(pieces)
	}

	for i := 0; i < numPieces; i++ {
		pieces[i] = strings.Repeat("*", utf8.RuneCountInString(pieces[i]))
	}
	return strings.Join(pieces, delim)
}

type SessionSlice struct {
	Timestamp         time.Time        `json:"timestamp"`
	Next              routing.Stats    `json:"next"`
	Direct            routing.Stats    `json:"direct"`
	Envelope          routing.Envelope `json:"envelope"`
	OnNetworkNext     bool             `json:"on_network_next"`
	IsMultiPath       bool             `json:"is_multipath"`
	IsTryBeforeYouBuy bool             `json:"is_try_before_you_buy"`
}

func (s *SessionSlice) UnmarshalBinary(data []byte) error {
	index := 0

	var version uint32
	if !encoding.ReadUint32(data, &index, &version) {
		return errors.New("[SessionSlice] invalid read at version number")
	}

	if version > SessionSliceVersion {
		return fmt.Errorf("unknown session slice version: %d", version)
	}

	var timestamp uint64
	if !encoding.ReadUint64(data, &index, &timestamp) {
		return errors.New("[SessionSlice] invalid read at timestamp")
	}
	s.Timestamp = time.Unix(0, int64(timestamp))

	if !encoding.ReadFloat64(data, &index, &s.Next.RTT) {
		return errors.New("[SessionSlice] invalid read at next RTT")
	}
	if !encoding.ReadFloat64(data, &index, &s.Next.Jitter) {
		return errors.New("[SessionSlice] invalid read at next jitter")
	}
	if !encoding.ReadFloat64(data, &index, &s.Next.PacketLoss) {
		return errors.New("[SessionSlice] invalid read at next packet loss")
	}

	if !encoding.ReadFloat64(data, &index, &s.Direct.RTT) {
		return errors.New("[SessionSlice] invalid read at direct RTT")
	}
	if !encoding.ReadFloat64(data, &index, &s.Direct.Jitter) {
		return errors.New("[SessionSlice] invalid read at direct jitter")
	}
	if !encoding.ReadFloat64(data, &index, &s.Direct.PacketLoss) {
		return errors.New("[SessionSlice] invalid read at direct packet loss")
	}

	var up uint64
	if !encoding.ReadUint64(data, &index, &up) {
		return errors.New("[SessionSlice] invalid read at envelope up")
	}

	var down uint64
	if !encoding.ReadUint64(data, &index, &down) {
		return errors.New("[SessionSlice] invalid read at envelope down")
	}

	s.Envelope = routing.Envelope{Up: int64(up), Down: int64(down)}

	if !encoding.ReadBool(data, &index, &s.OnNetworkNext) {
		return errors.New("[SessionSlice] invalid read at on network next")
	}

	if !encoding.ReadBool(data, &index, &s.IsMultiPath) {
		return errors.New("[SessionSlice] invalid read at is multipath")
	}

	if !encoding.ReadBool(data, &index, &s.IsTryBeforeYouBuy) {
		return errors.New("[SessionSlice] invalid read at is try before you buy")
	}

	return nil
}

func (s SessionSlice) MarshalBinary() ([]byte, error) {
	data := make([]byte, s.Size())
	index := 0

	encoding.WriteUint32(data, &index, SessionSliceVersion)
	encoding.WriteUint64(data, &index, uint64(s.Timestamp.UnixNano()))
	encoding.WriteFloat64(data, &index, s.Next.RTT)
	encoding.WriteFloat64(data, &index, s.Next.Jitter)
	encoding.WriteFloat64(data, &index, s.Next.PacketLoss)
	encoding.WriteFloat64(data, &index, s.Direct.RTT)
	encoding.WriteFloat64(data, &index, s.Direct.Jitter)
	encoding.WriteFloat64(data, &index, s.Direct.PacketLoss)
	encoding.WriteUint64(data, &index, uint64(s.Envelope.Up))
	encoding.WriteUint64(data, &index, uint64(s.Envelope.Down))
	encoding.WriteBool(data, &index, s.OnNetworkNext)
	encoding.WriteBool(data, &index, s.IsMultiPath)
	encoding.WriteBool(data, &index, s.IsTryBeforeYouBuy)

	return data, nil
}

func (s SessionSlice) Size() uint64 {
	return 4 + 8 + (3 * 8) + (3 * 8) + (2 * 8) + 1 + 1 + 1
}

type SessionMapPoint struct {
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	OnNetworkNext bool    `json:"on_network_next"`
}

func (s *SessionMapPoint) UnmarshalBinary(data []byte) error {
	index := 0

	var version uint32
	if !encoding.ReadUint32(data, &index, &version) {
		return errors.New("[SessionMapPoint] invalid read at version number")
	}

	if version > SessionMapPointVersion {
		return fmt.Errorf("unknown session map point version: %d", version)
	}

	if !encoding.ReadFloat64(data, &index, &s.Latitude) {
		return errors.New("[SessionMapPoint] invalid read at latitude")
	}

	if !encoding.ReadFloat64(data, &index, &s.Longitude) {
		return errors.New("[SessionMapPoint] invalid read at longitude")
	}

	if !encoding.ReadBool(data, &index, &s.OnNetworkNext) {
		return errors.New("[SessionMapPoint] invalid read at on network next")
	}

	return nil
}

func (s SessionMapPoint) MarshalBinary() ([]byte, error) {
	data := make([]byte, s.Size())
	index := 0

	encoding.WriteUint32(data, &index, SessionMapPointVersion)
	encoding.WriteFloat64(data, &index, s.Latitude)
	encoding.WriteFloat64(data, &index, s.Longitude)
	encoding.WriteBool(data, &index, s.OnNetworkNext)

	return data, nil
}

func (s SessionMapPoint) Size() uint64 {
	return 4 + 8 + 8 + 1
}
