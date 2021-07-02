package transport

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/networknext/backend/modules/encoding"
	"github.com/networknext/backend/modules/routing"
)

const (
	SessionCountDataVersion    = 1
	SessionPortalDataVersion   = 1
	SessionMetaVersion         = 2
	SessionSliceVersion        = 2
	SessionMapPointVersion     = 1
	RelayHopVersion            = 1
	NearRelayPortalDataVersion = 1

	MaxSessionIDLength  = 1024
	MaxAddressLength    = 256
	MaxSDKVersionLength = 11
)

type SessionCountData struct {
	Version     uint32
	ServerID    uint64
	BuyerID     uint64
	NumSessions uint32
}

func (s *SessionCountData) UnmarshalBinary(data []byte) error {
	index := 0

	if !encoding.ReadUint32(data, &index, &s.Version) {
		return errors.New("[SessionCountData] invalid read at version number")
	}

	if s.Version > SessionCountDataVersion {
		return fmt.Errorf("unknown session count version: %d", s.Version)
	}

	if !encoding.ReadUint64(data, &index, &s.ServerID) {
		return errors.New("[SessionCountData] invalid read at server ID")
	}

	if !encoding.ReadUint64(data, &index, &s.BuyerID) {
		return errors.New("[SessionCountData] invalid read at buyer ID")
	}

	if !encoding.ReadUint32(data, &index, &s.NumSessions) {
		return errors.New("[SessionCountData] invalid read at num sessions")
	}

	return nil
}

func (s SessionCountData) MarshalBinary() ([]byte, error) {
	data := make([]byte, s.Size())
	index := 0

	encoding.WriteUint32(data, &index, SessionCountDataVersion)
	encoding.WriteUint64(data, &index, s.ServerID)
	encoding.WriteUint64(data, &index, s.BuyerID)
	encoding.WriteUint32(data, &index, s.NumSessions)

	return data, nil
}

func (s *SessionCountData) Serialize(stream encoding.Stream) error {
	stream.SerializeUint32(&s.Version)

	stream.SerializeUint64(&s.ServerID)

	stream.SerializeUint64(&s.BuyerID)

	stream.SerializeUint32(&s.NumSessions)

	return stream.Error()
}

func (s *SessionCountData) Size() uint64 {
	return 4 + 8 + 8 + 4
}

func WriteSessionCountData(entry *SessionCountData) ([]byte, error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("recovered from panic during SessionCountData packet entry write: %v\n", r)
		}
	}()

	size := entry.Size()
	buffer := make([]byte, size)

	ws, err := encoding.CreateWriteStream(buffer[:])
	if err != nil {
		return nil, err
	}

	if err := entry.Serialize(ws); err != nil {
		return nil, err
	}
	ws.Flush()

	return buffer[:ws.GetBytesProcessed()], nil
}

func ReadSessionCountData(entry *SessionCountData, data []byte) error {
	if err := entry.Serialize(encoding.CreateReadStream(data)); err != nil {
		return err
	}
	return nil
}

type SessionPortalData struct {
	Version uint32
	Meta    SessionMeta     `json:"meta"`
	Slice   SessionSlice    `json:"slice"`
	Point   SessionMapPoint `json:"point"`

	LargeCustomer bool `json:"largeCustomer"`
	EverOnNext    bool `json:"everOnNext"`
}

func (s *SessionPortalData) UnmarshalBinary(data []byte) error {
	index := 0

	if !encoding.ReadUint32(data, &index, &s.Version) {
		return errors.New("[SessionPortalData] invalid read at version number")
	}

	if s.Version > SessionPortalDataVersion {
		return fmt.Errorf("unknown session data version: %d", s.Version)
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

	if s.Version >= 1 {
		if !encoding.ReadBool(data, &index, &s.LargeCustomer) {
			return errors.New("[SessionPortalData] invalid read at large customer bool")
		}

		if !encoding.ReadBool(data, &index, &s.EverOnNext) {
			return errors.New("[SessionPortalData] invalid read at ever on next bool")
		}
	}

	return nil
}

func (s SessionPortalData) MarshalBinary() ([]byte, error) {
	data := make([]byte, s.Size())
	index := 0

	encoding.WriteUint32(data, &index, SessionPortalDataVersion)

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

	encoding.WriteBool(data, &index, s.LargeCustomer)
	encoding.WriteBool(data, &index, s.EverOnNext)

	return data, nil
}

func (s *SessionPortalData) Serialize(stream encoding.Stream) error {
	stream.SerializeUint32(&s.Version)

	var err error
	var meta []byte

	if stream.IsWriting() {
		meta, err = WriteSessionMeta(&s.Meta)
		if err != nil {
			return err
		}
	}
	stream.SerializeBytes(meta)
	if stream.IsReading() {
		err = ReadSessionMeta(&s.Meta, meta)
		if err != nil {
			return err
		}
	}

	var slice []byte

	if stream.IsWriting() {
		slice, err = WriteSessionSlice(&s.Slice)
		if err != nil {
			return err
		}
	}
	stream.SerializeBytes(slice)
	if stream.IsReading() {
		err = ReadSessionSlice(&s.Slice, slice)
		if err != nil {
			return err
		}
	}

	var point []byte

	if stream.IsWriting() {
		point, err = WriteSessionMapPoint(&s.Point)
		if err != nil {
			return err
		}
	}
	stream.SerializeBytes(point)
	if stream.IsReading() {
		err = ReadSessionMapPoint(&s.Point, point)
		if err != nil {
			return err
		}
	}

	stream.SerializeBool(&s.LargeCustomer)

	stream.SerializeBool(&s.EverOnNext)

	return stream.Error()
}

func (s *SessionPortalData) Size() uint64 {
	return 4 + 4 + s.Meta.Size() + 4 + s.Slice.Size() + 4 + s.Point.Size() + 1 + 1
}

func WriteSessionPortalData(entry *SessionPortalData) ([]byte, error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("recovered from panic during SessionPortalData packet entry write: %v\n", r)
		}
	}()

	size := entry.Size()
	buffer := make([]byte, size)

	ws, err := encoding.CreateWriteStream(buffer[:])
	if err != nil {
		return nil, err
	}

	if err := entry.Serialize(ws); err != nil {
		return nil, err
	}
	ws.Flush()

	return buffer[:ws.GetBytesProcessed()], nil
}

func ReadSessionPortalData(entry *SessionPortalData, data []byte) error {
	if err := entry.Serialize(encoding.CreateReadStream(data)); err != nil {
		return err
	}
	return nil
}

type RelayHop struct {
	Version uint32
	ID      uint64 `json:"id"`
	Name    string `json:"name"`
}

func (h RelayHop) RedisString() string {
	return fmt.Sprintf("%d|%016x|%s", h.Version, h.ID, h.Name)
}

func (h *RelayHop) ParseRedisString(values []string) error {
	var index int
	var err error

	var version int64
	if version, err = strconv.ParseInt(values[index], 10, 64); err != nil {
		return fmt.Errorf("[RelayHop] failed to read version from redis data: %v", err)
	}

	// The original data didn't have the version serialized
	if math.Abs(float64(version)) > RelayHopVersion {
		version = 0
	} else {
		h.Version = uint32(version)
		index++
	}

	if h.ID, err = strconv.ParseUint(values[index], 16, 64); err != nil {
		return fmt.Errorf("[RelayHop] failed to read relay ID from redis data: %v", err)
	}
	index++

	h.Name = values[index]
	index++

	return nil
}

func (h *RelayHop) Serialize(stream encoding.Stream) error {
	stream.SerializeUint32(&h.Version)

	stream.SerializeUint64(&h.ID)

	stream.SerializeString(&h.Name, routing.MaxRelayNameLength)

	return stream.Error()
}

func (h *RelayHop) Size() uint64 {
	return 4 + // Version
		8 + // ID
		routing.MaxRelayNameLength + 1 // Name
}

func WriteRelayHop(entry *RelayHop) ([]byte, error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("recovered from panic during RelayHop packet entry write: %v\n", r)
		}
	}()

	size := entry.Size()
	buffer := make([]byte, size)

	ws, err := encoding.CreateWriteStream(buffer[:])
	if err != nil {
		return nil, err
	}

	if err := entry.Serialize(ws); err != nil {
		return nil, err
	}
	ws.Flush()

	return buffer[:ws.GetBytesProcessed()], nil
}

func ReadRelayHop(entry *RelayHop, data []byte) error {
	if err := entry.Serialize(encoding.CreateReadStream(data)); err != nil {
		return err
	}
	return nil
}

type NearRelayPortalData struct {
	Version     uint32
	ID          uint64        `json:"id"`
	Name        string        `json:"name"`
	ClientStats routing.Stats `json:"client_stats"`
}

func (n NearRelayPortalData) RedisString() string {
	return fmt.Sprintf("%d|%016x|%s|%s", n.Version, n.ID, n.Name, n.ClientStats.RedisString())
}

func (n *NearRelayPortalData) ParseRedisString(values []string) error {
	var index int
	var err error

	var version int64
	if version, err = strconv.ParseInt(values[index], 10, 64); err != nil {
		return fmt.Errorf("[NearRelayPortalData] failed to read version from redis data: %v", err)
	}

	// The original data didn't have the version serialized
	if math.Abs(float64(version)) > NearRelayPortalDataVersion {
		version = 0
	} else {
		n.Version = uint32(version)
		index++
	}

	if n.ID, err = strconv.ParseUint(values[index], 16, 64); err != nil {
		return fmt.Errorf("[NearRelayPortalData] failed to read relay ID from redis data: %v", err)
	}
	index++

	n.Name = values[index]
	index++

	if err := n.ClientStats.ParseRedisString([]string{values[index], values[index+1], values[index+2]}); err != nil {
		return fmt.Errorf("[NearRelayPortalData] failed to read client stats from redis data: %v", err)
	}
	index += 3

	return nil
}

func (n *NearRelayPortalData) Serialize(stream encoding.Stream) error {
	stream.SerializeUint32(&n.Version)

	stream.SerializeUint64(&n.ID)

	stream.SerializeString(&n.Name, routing.MaxRelayNameLength)

	if stream.IsReading() {
		n.ClientStats = routing.Stats{}
	}

	stats := &n.ClientStats
	stream.SerializeFloat64(&stats.RTT)
	stream.SerializeFloat64(&stats.Jitter)
	stream.SerializeFloat64(&stats.PacketLoss)

	return stream.Error()
}

func (n *NearRelayPortalData) Size() uint64 {
	return 4 + // Version
		8 + // ID
		routing.MaxRelayNameLength + 1 + // Name
		8 + // Client Stats RTT
		8 + // Client Stats Jitter
		8 // Client Stats Packet Loss
}

func WriteNearRelayPortalData(entry *NearRelayPortalData) ([]byte, error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("recovered from panic during NearRelayPortalData packet entry write: %v\n", r)
		}
	}()

	size := entry.Size()
	buffer := make([]byte, size)

	ws, err := encoding.CreateWriteStream(buffer[:])
	if err != nil {
		return nil, err
	}

	if err := entry.Serialize(ws); err != nil {
		return nil, err
	}
	ws.Flush()

	return buffer[:ws.GetBytesProcessed()], nil
}

func ReadNearRelayPortalData(entry *NearRelayPortalData, data []byte) error {
	if err := entry.Serialize(encoding.CreateReadStream(data)); err != nil {
		return err
	}
	return nil
}

type SessionMeta struct {
	Version         uint32
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

	if !encoding.ReadUint32(data, &index, &s.Version) {
		return errors.New("[SessionMeta] invalid read at version number")
	}

	if s.Version > SessionMetaVersion {
		return fmt.Errorf("unknown session meta version: %d", s.Version)
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
		if s.Version >= 2 {
			if !encoding.ReadUint32(data, &index, &s.Hops[i].Version) {
				return errors.New("[SessionMeta] invalid read at relay hops relay version")
			}
		}

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
		if s.Version >= 2 {
			if !encoding.ReadUint32(data, &index, &s.NearbyRelays[i].Version) {
				return errors.New("[SessionMeta] invalid read at nearby relay version")
			}
		}

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

	encoding.WriteUint32(data, &index, s.Version)
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
		if s.Version >= 2 {
			encoding.WriteUint32(data, &index, hop.Version)
		}
		encoding.WriteUint64(data, &index, hop.ID)
		encoding.WriteString(data, &index, hop.Name, uint32(len(hop.Name)))
	}

	encoding.WriteString(data, &index, s.SDK, uint32(len(s.SDK)))
	encoding.WriteUint8(data, &index, s.Connection)

	encoding.WriteUint32(data, &index, uint32(len(s.NearbyRelays)))
	for _, nearRelayData := range s.NearbyRelays {
		if s.Version >= 2 {
			encoding.WriteUint32(data, &index, nearRelayData.Version)
		}
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

func (s *SessionMeta) Serialize(stream encoding.Stream) error {
	stream.SerializeUint32(&s.Version)

	stream.SerializeUint64(&s.ID)

	stream.SerializeUint64(&s.UserHash)

	stream.SerializeString(&s.DatacenterName, MaxDatacenterNameLength)
	stream.SerializeString(&s.DatacenterAlias, MaxDatacenterNameLength)

	stream.SerializeBool(&s.OnNetworkNext)

	stream.SerializeFloat64(&s.NextRTT)
	stream.SerializeFloat64(&s.DirectRTT)
	stream.SerializeFloat64(&s.DeltaRTT)

	if stream.IsReading() {
		s.Location = routing.Location{}
	}
	location := &s.Location

	stream.SerializeString(&location.Continent, routing.MaxContinentLength)
	stream.SerializeString(&location.Country, routing.MaxCountryLength)
	stream.SerializeString(&location.CountryCode, routing.MaxCountryCodeLength)
	stream.SerializeString(&location.Region, routing.MaxRegionLength)
	stream.SerializeString(&location.City, routing.MaxCityLength)
	stream.SerializeFloat32(&location.Latitude)
	stream.SerializeFloat32(&location.Longitude)
	stream.SerializeString(&location.ISP, routing.MaxISPNameLength)
	var asn uint64
	if stream.IsWriting() {
		asn = uint64(location.ASN)
	}
	stream.SerializeUint64(&asn)
	if stream.IsReading() {
		location.ASN = int(asn)
	}

	stream.SerializeString(&s.ClientAddr, MaxAddressLength)
	stream.SerializeString(&s.ServerAddr, MaxAddressLength)

	numHops := uint32(len(s.Hops))
	stream.SerializeUint32(&numHops)

	if stream.IsReading() {
		s.Hops = make([]RelayHop, numHops)
	}

	for i := uint32(0); i < numHops; i++ {
		hop := &s.Hops[i]

		if s.Version >= 2 {
			stream.SerializeUint32(&hop.Version)
		}

		stream.SerializeUint64(&hop.ID)
		stream.SerializeString(&hop.Name, routing.MaxRelayNameLength)
	}

	stream.SerializeString(&s.SDK, MaxSDKVersionLength)

	var conn uint32
	if stream.IsWriting() {
		conn = uint32(s.Connection)
	}
	stream.SerializeBits(&conn, 8)
	if stream.IsReading() {
		s.Connection = uint8(conn)
	}

	numNearbyRelays := uint32(len(s.NearbyRelays))
	stream.SerializeUint32(&numNearbyRelays)

	if stream.IsReading() {
		s.NearbyRelays = make([]NearRelayPortalData, numNearbyRelays)
	}

	for i := uint32(0); i < numNearbyRelays; i++ {
		nearbyRelay := &s.NearbyRelays[i]

		if s.Version >= 2 {
			stream.SerializeUint32(&nearbyRelay.Version)
		}

		stream.SerializeUint64(&nearbyRelay.ID)
		stream.SerializeString(&nearbyRelay.Name, routing.MaxRelayNameLength)

		if stream.IsReading() {
			nearbyRelay.ClientStats = routing.Stats{}
		}

		stats := &nearbyRelay.ClientStats
		stream.SerializeFloat64(&stats.RTT)
		stream.SerializeFloat64(&stats.Jitter)
		stream.SerializeFloat64(&stats.PacketLoss)
	}

	var platform uint32
	if stream.IsWriting() {
		platform = uint32(s.Platform)
	}
	stream.SerializeBits(&platform, 8)
	if stream.IsReading() {
		s.Platform = uint8(platform)
	}

	stream.SerializeUint64(&s.BuyerID)

	return stream.Error()
}

func (s SessionMeta) Size() uint64 {
	var relayHopsSize uint64
	for _, hop := range s.Hops {
		relayHopsSize += 8 + 4 + 4 + uint64(len(hop.Name))
	}

	var nearbyRelaysSize uint64
	for _, nearRelayData := range s.NearbyRelays {
		nearbyRelaysSize += 8 + 4 + 4 + uint64(len(nearRelayData.Name)) + 8 + 8 + 8
	}

	return 4 + 8 + 8 + 4 + uint64(len(s.DatacenterName)) + 4 + uint64(len(s.DatacenterAlias)) + 1 + 8 + 8 + 8 + 4 + s.Location.Size() +
		4 + uint64(len(s.ClientAddr)) + 4 + uint64(len(s.ServerAddr)) + (4 + relayHopsSize) + 4 + uint64(len(s.SDK)) + 1 + (4 + nearbyRelaysSize) + 1 + 8
}

func (s SessionMeta) SerializeSize() uint64 {
	hopSize := uint64(0)
	for i := 0; i < len(s.Hops); i++ {
		hopSize += s.Hops[i].Size()
	}

	nearbyRelaySize := uint64(0)
	for i := 0; i < len(s.NearbyRelays); i++ {
		nearbyRelaySize += s.NearbyRelays[i].Size()
	}

	return 4 + // Version
		8 + // ID
		8 + // UserHash
		MaxDatacenterNameLength + // DataceterName
		MaxDatacenterNameLength + // DatacenterAlias
		1 + // OnNetworkNext
		8 + // NextRTT
		8 + // DirectRTT
		8 + // DeltaRTT
		routing.MaxContinentLength + // Continent
		routing.MaxCountryLength + // Country
		routing.MaxCountryCodeLength + // CountryCode
		routing.MaxRegionLength + // Region
		routing.MaxCityLength + // City
		4 + // Latitude
		4 + // Longitude
		routing.MaxISPNameLength + // ISP
		8 + // ASN
		MaxAddressLength + // ClientAddr
		MaxAddressLength + // ServerAddr
		hopSize + // Hops
		MaxSDKVersionLength + 1 + // SDK
		1 + // Connection
		nearbyRelaySize + // NearbyRelays
		1 + // Platform
		8 + // BuyerID
		1 // extra bytes to be divisible by 4
}

func WriteSessionMeta(entry *SessionMeta) ([]byte, error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("recovered from panic during SessionMeta packet entry write: %v\n", r)
		}
	}()

	size := entry.SerializeSize()
	buffer := make([]byte, size)

	ws, err := encoding.CreateWriteStream(buffer[:])
	if err != nil {
		return nil, err
	}

	if err := entry.Serialize(ws); err != nil {
		return nil, err
	}
	ws.Flush()

	return buffer[:ws.GetBytesProcessed()], nil
}

func ReadSessionMeta(entry *SessionMeta, data []byte) error {
	if err := entry.Serialize(encoding.CreateReadStream(data)); err != nil {
		return err
	}
	return nil
}

func (s *SessionMeta) UnmarshalJSON(data []byte) error {
	fields := map[string]interface{}{}
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}

	id, err := strconv.ParseUint(fields["id"].(string), 16, 64)
	if err != nil {
		return err
	}
	s.ID = id

	userHash, err := strconv.ParseUint(fields["user_hash"].(string), 16, 64)
	if err != nil {
		return err
	}
	s.UserHash = userHash

	s.DatacenterName = fields["datacenter_name"].(string)
	s.DatacenterAlias = fields["datacenter_alias"].(string)
	s.OnNetworkNext = fields["on_network_next"].(bool)
	s.NextRTT = fields["next_rtt"].(float64)
	s.DirectRTT = fields["direct_rtt"].(float64)
	s.DeltaRTT = fields["delta_rtt"].(float64)

	// Have to grossly remarshal and unmarshal since you can only unmarshal from a []byte and not a map[string]interface{}
	locationMap := fields["location"].(map[string]interface{})
	locationData, err := json.Marshal(locationMap)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(locationData, &s.Location); err != nil {
		return err
	}

	s.ClientAddr = fields["client_addr"].(string)
	s.ServerAddr = fields["server_addr"].(string)

	// Have to grossly remarshal and unmarshal since you can only unmarshal from a []byte and not a []interface{}
	hopsArray := fields["hops"].([]interface{})
	hopsData, err := json.Marshal(hopsArray)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(hopsData, &s.Hops); err != nil {
		return err
	}

	s.SDK = fields["sdk"].(string)
	s.Connection = ParseConnectionType(fields["connection"].(string))

	// Have to grossly remarshal and unmarshal since you can only unmarshal from a []byte and not a []interface{}
	nearbyRelaysArray := fields["nearby_relays"].([]interface{})
	nearbyRelaysData, err := json.Marshal(nearbyRelaysArray)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(nearbyRelaysData, &s.NearbyRelays); err != nil {
		return err
	}

	s.Platform = ParsePlatformType(fields["platform"].(string))

	customerID, err := strconv.ParseUint(fields["customer_id"].(string), 16, 64)
	if err != nil {
		return err
	}
	s.BuyerID = customerID

	return nil
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
	fields["connection"] = ConnectionTypeText(s.Connection)
	fields["nearby_relays"] = s.NearbyRelays
	fields["platform"] = PlatformTypeText(s.Platform)
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

func (s SessionMeta) RedisString() string {
	onNetworkNextString := "0"
	if s.OnNetworkNext {
		onNetworkNextString = "1"
	}

	result := fmt.Sprintf("%d|%016x|%016x|%s|%s|%s|%.2f|%.2f|%.2f|%s|%s|%s|", s.Version, s.ID, s.UserHash, s.DatacenterName, s.DatacenterAlias, onNetworkNextString,
		s.NextRTT, s.DirectRTT, s.DeltaRTT, s.Location.RedisString(), s.ClientAddr, s.ServerAddr)
	result += fmt.Sprintf("%d|", len(s.Hops))
	for i := 0; i < len(s.Hops); i++ {
		result += fmt.Sprintf("%s|", s.Hops[i].RedisString())
	}
	result += fmt.Sprintf("%s|%d|", s.SDK, s.Connection)
	result += fmt.Sprintf("%d|", len(s.NearbyRelays))
	for i := 0; i < len(s.NearbyRelays); i++ {
		result += fmt.Sprintf("%s|", s.NearbyRelays[i].RedisString())
	}
	result += fmt.Sprintf("%d|%016x", s.Platform, s.BuyerID)
	return result
}

func (s *SessionMeta) ParseRedisString(values []string) error {
	var index int
	var err error

	var version int64
	if version, err = strconv.ParseInt(values[index], 10, 64); err != nil {
		return fmt.Errorf("[SessionMeta] failed to read version from redis data: %v", err)
	}

	// The original data didn't have the version serialized
	if math.Abs(float64(version)) > SessionMetaVersion {
		version = 0
	} else {
		s.Version = uint32(version)
		index++
	}

	if s.ID, err = strconv.ParseUint(values[index], 16, 64); err != nil {
		return fmt.Errorf("[SessionMeta] failed to read session ID from redis data: %v", err)
	}
	index++

	if s.UserHash, err = strconv.ParseUint(values[index], 16, 64); err != nil {
		return fmt.Errorf("[SessionMeta] failed to read user hash from redis data: %v", err)
	}
	index++

	s.DatacenterName = values[index]
	index++
	s.DatacenterAlias = values[index]
	index++

	if s.OnNetworkNext, err = strconv.ParseBool(values[index]); err != nil {
		return fmt.Errorf("[SessionMeta] failed to read on network next from redis data: %v", err)
	}
	index++

	if s.NextRTT, err = strconv.ParseFloat(values[index], 64); err != nil {
		return fmt.Errorf("[SessionMeta] failed to read next RTT from redis data: %v", err)
	}
	index++

	if s.DirectRTT, err = strconv.ParseFloat(values[index], 64); err != nil {
		return fmt.Errorf("[SessionMeta] failed to read direct RTT from redis data: %v", err)
	}
	index++

	if s.DeltaRTT, err = strconv.ParseFloat(values[index], 64); err != nil {
		return fmt.Errorf("[SessionMeta] failed to read delta RTT from redis data: %v", err)
	}
	index++

	if err := s.Location.ParseRedisString([]string{values[index], values[index+1], values[index+2]}); err != nil {
		return fmt.Errorf("[SessionMeta] failed to read location from redis data: %v", err)
	}
	index += 3

	s.ClientAddr = values[index]
	index++
	s.ServerAddr = values[index]
	index++

	var numHops int64
	if numHops, err = strconv.ParseInt(values[index], 10, 32); err != nil {
		return fmt.Errorf("[SessionMeta] failed to read number of relay hops from redis data: %v", err)
	}
	index++

	s.Hops = make([]RelayHop, numHops)
	for i := 0; i < int(numHops); i++ {
		var hop RelayHop
		if err := hop.ParseRedisString([]string{values[index], values[index+1], values[index+2]}); err != nil {
			return fmt.Errorf("[SessionMeta] failed to read relay hop from redis data: %v", err)
		}
		index += 3

		s.Hops[i] = hop
	}

	s.SDK = values[index]
	index++

	var connection uint64
	if connection, err = strconv.ParseUint(values[index], 10, 8); err != nil {
		return fmt.Errorf("[SessionMeta] failed to read connection type from redis data: %v", err)
	}
	s.Connection = uint8(connection)
	index++

	var numNearRelays int64
	if numNearRelays, err = strconv.ParseInt(values[index], 10, 32); err != nil {
		return fmt.Errorf("[SessionMeta] failed to read number of near relays from redis data: %v", err)
	}
	index++

	s.NearbyRelays = make([]NearRelayPortalData, numNearRelays)
	for i := 0; i < int(numNearRelays); i++ {
		var nearRelay NearRelayPortalData
		if err := nearRelay.ParseRedisString([]string{values[index], values[index+1], values[index+2], values[index+3], values[index+4], values[index+5]}); err != nil {
			return fmt.Errorf("[SessionMeta] failed to read near relay from redis data: %v", err)
		}
		index += 6

		s.NearbyRelays[i] = nearRelay
	}

	var platform uint64
	if platform, err = strconv.ParseUint(values[index], 10, 8); err != nil {
		return fmt.Errorf("[SessionMeta] failed to read platform type from redis data: %v", err)
	}
	s.Platform = uint8(platform)
	index++

	if s.BuyerID, err = strconv.ParseUint(values[index], 16, 64); err != nil {
		return fmt.Errorf("[SessionMeta] failed to read buyer ID from redis data: %v", err)
	}
	index++

	return nil
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
	Version             uint32           `json:"version"`
	Timestamp           time.Time        `json:"timestamp"`
	Next                routing.Stats    `json:"next"`
	Direct              routing.Stats    `json:"direct"`
	Predicted           routing.Stats    `json:"predicted"`
	ClientToServerStats routing.Stats    `json:"client_to_server_stats"`
	ServerToClientStats routing.Stats    `json:"server_to_client_stats"`
	RouteDiversity      uint32           `json:"route_diversity"`
	Envelope            routing.Envelope `json:"envelope"`
	OnNetworkNext       bool             `json:"on_network_next"`
	IsMultiPath         bool             `json:"is_multipath"`
	IsTryBeforeYouBuy   bool             `json:"is_try_before_you_buy"`
}

func (s *SessionSlice) UnmarshalBinary(data []byte) error {
	index := 0

	if !encoding.ReadUint32(data, &index, &s.Version) {
		return errors.New("[SessionSlice] invalid read at version number")
	}

	if s.Version > SessionSliceVersion {
		return fmt.Errorf("unknown session slice version: %d", s.Version)
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

	if s.Version >= 1 {
		if !encoding.ReadFloat64(data, &index, &s.Predicted.RTT) {
			return errors.New("[SessionSlice] invalid read at predicted RTT")
		}
	}

	if s.Version >= 2 {
		if !encoding.ReadFloat64(data, &index, &s.ClientToServerStats.Jitter) {
			return errors.New("[SessionSlice] invalid read at client to server jitter")
		}

		if !encoding.ReadFloat64(data, &index, &s.ClientToServerStats.PacketLoss) {
			return errors.New("[SessionSlice] invalid read at client to server packet loss")
		}

		if !encoding.ReadFloat64(data, &index, &s.ServerToClientStats.Jitter) {
			return errors.New("[SessionSlice] invalid read at server to client jitter")
		}

		if !encoding.ReadFloat64(data, &index, &s.ServerToClientStats.PacketLoss) {
			return errors.New("[SessionSlice] invalid read at server to client packet loss")
		}

		if !encoding.ReadUint32(data, &index, &s.RouteDiversity) {
			return errors.New("[SessionSlice] invalid read at route diversity")
		}
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
	encoding.WriteFloat64(data, &index, s.Predicted.RTT)
	encoding.WriteFloat64(data, &index, s.ClientToServerStats.Jitter)
	encoding.WriteFloat64(data, &index, s.ClientToServerStats.PacketLoss)
	encoding.WriteFloat64(data, &index, s.ServerToClientStats.Jitter)
	encoding.WriteFloat64(data, &index, s.ServerToClientStats.PacketLoss)
	encoding.WriteUint32(data, &index, s.RouteDiversity)
	encoding.WriteUint64(data, &index, uint64(s.Envelope.Up))
	encoding.WriteUint64(data, &index, uint64(s.Envelope.Down))
	encoding.WriteBool(data, &index, s.OnNetworkNext)
	encoding.WriteBool(data, &index, s.IsMultiPath)
	encoding.WriteBool(data, &index, s.IsTryBeforeYouBuy)

	return data[:index], nil
}

func (s *SessionSlice) Serialize(stream encoding.Stream) error {
	stream.SerializeUint32(&s.Version)

	var unixTime uint64
	if stream.IsWriting() {
		unixTime = uint64(s.Timestamp.UnixNano())
	}
	stream.SerializeUint64(&unixTime)

	if stream.IsReading() {
		s.Timestamp = time.Unix(0, int64(unixTime))
	}

	if stream.IsReading() {
		s.Next = routing.Stats{}
		s.Direct = routing.Stats{}

		if s.Version >= 1 {
			s.Predicted = routing.Stats{}
		}

		if s.Version >= 2 {
			s.ClientToServerStats = routing.Stats{}
			s.ServerToClientStats = routing.Stats{}
		}
	}

	nextStats := &s.Next
	stream.SerializeFloat64(&nextStats.RTT)
	stream.SerializeFloat64(&nextStats.Jitter)
	stream.SerializeFloat64(&nextStats.PacketLoss)

	directStats := &s.Direct
	stream.SerializeFloat64(&directStats.RTT)
	stream.SerializeFloat64(&directStats.Jitter)
	stream.SerializeFloat64(&directStats.PacketLoss)

	if s.Version >= 1 {
		predictedStats := &s.Predicted
		stream.SerializeFloat64(&predictedStats.RTT)
	}

	if s.Version >= 2 {
		clientToServerStats := &s.ClientToServerStats
		stream.SerializeFloat64(&clientToServerStats.Jitter)
		stream.SerializeFloat64(&clientToServerStats.PacketLoss)

		serverToClientStats := &s.ServerToClientStats
		stream.SerializeFloat64(&serverToClientStats.Jitter)
		stream.SerializeFloat64(&serverToClientStats.PacketLoss)

		stream.SerializeUint32(&s.RouteDiversity)
	}

	if stream.IsReading() {
		s.Envelope = routing.Envelope{}
	}

	var envelopeUp uint64
	var envelopeDown uint64
	if stream.IsWriting() {
		envelopeUp = uint64(s.Envelope.Up)
		envelopeDown = uint64(s.Envelope.Down)
	}
	stream.SerializeUint64(&envelopeUp)
	stream.SerializeUint64(&envelopeDown)
	if stream.IsReading() {
		s.Envelope = routing.Envelope{
			Up:   int64(envelopeUp),
			Down: int64(envelopeDown),
		}
	}

	stream.SerializeBool(&s.OnNetworkNext)

	stream.SerializeBool(&s.IsMultiPath)

	stream.SerializeBool(&s.IsTryBeforeYouBuy)

	return stream.Error()
}

func (s SessionSlice) Size() uint64 {
	return 4 + 8 + (3 * 8) + (3 * 8) + 8 + (2 * 8) + (2 * 8) + 4 + (2 * 8) + 1 + 1 + 1
}

func WriteSessionSlice(entry *SessionSlice) ([]byte, error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("recovered from panic during SessionSlice packet entry write: %v\n", r)
		}
	}()

	size := entry.Size()
	buffer := make([]byte, size)

	ws, err := encoding.CreateWriteStream(buffer[:])
	if err != nil {
		return nil, err
	}

	if err := entry.Serialize(ws); err != nil {
		return nil, err
	}
	ws.Flush()

	return buffer[:ws.GetBytesProcessed()], nil
}

func ReadSessionSlice(entry *SessionSlice, data []byte) error {
	if err := entry.Serialize(encoding.CreateReadStream(data)); err != nil {
		return err
	}
	return nil
}

func (s SessionSlice) RedisString() string {
	onNetworkNextString := "0"
	if s.OnNetworkNext {
		onNetworkNextString = "1"
	}

	isMultipathString := "0"
	if s.IsMultiPath {
		isMultipathString = "1"
	}

	isTryBeforeYouBuyString := "0"
	if s.IsTryBeforeYouBuy {
		isTryBeforeYouBuyString = "1"
	}

	return fmt.Sprintf("%d|%d|%s|%s|%s|%s|%s|%d|%s|%s|%s|%s",
		SessionSliceVersion,
		s.Timestamp.Unix(),
		s.Next.RedisString(),
		s.Direct.RedisString(),
		s.Predicted.RedisString(),
		s.ClientToServerStats.RedisString(),
		s.ServerToClientStats.RedisString(),
		s.RouteDiversity,
		s.Envelope.RedisString(),
		onNetworkNextString,
		isMultipathString,
		isTryBeforeYouBuyString,
	)
}

func (s *SessionSlice) ParseRedisString(values []string) error {
	var index int
	var err error

	var version int64
	if version, err = strconv.ParseInt(values[index], 10, 64); err != nil {
		return fmt.Errorf("[SessionSlice] failed to read version from redis data: %v", err)
	}
	index++

	// The original data didn't have the version serialized, so it was actually the timestamp
	if version > SessionSliceVersion {
		s.Timestamp = time.Unix(version, 0)
		version = 0
	} else {
		s.Version = uint32(version)

		var timestamp int64
		if timestamp, err = strconv.ParseInt(values[index], 10, 64); err != nil {
			return fmt.Errorf("[SessionSlice] failed to read timestamp from redis data: %v", err)
		}
		index++

		s.Timestamp = time.Unix(timestamp, 0)
	}

	if err := s.Next.ParseRedisString([]string{values[index], values[index+1], values[index+2]}); err != nil {
		return fmt.Errorf("[SessionSlice] failed to read next stats from redis data: %v", err)
	}
	index += 3

	if err := s.Direct.ParseRedisString([]string{values[index], values[index+1], values[index+2]}); err != nil {
		return fmt.Errorf("[SessionSlice] failed to read direct stats from redis data: %v", err)
	}
	index += 3

	if err := s.Predicted.ParseRedisString([]string{values[index], values[index+1], values[index+2]}); err != nil {
		return fmt.Errorf("[SessionSlice] failed to read predicted stats from redis data: %v", err)
	}
	index += 3

	if s.Version >= 2 {
		if err := s.ClientToServerStats.ParseRedisString([]string{values[index], values[index+1], values[index+2]}); err != nil {
			return fmt.Errorf("[SessionSlice] failed to read client to server stats from redis data: %v", err)
		}
		index += 3

		if err := s.ServerToClientStats.ParseRedisString([]string{values[index], values[index+1], values[index+2]}); err != nil {
			return fmt.Errorf("[SessionSlice] failed to read server to client stats from redis data: %v", err)
		}
		index += 3

		var routeDiversity uint64
		if routeDiversity, err = strconv.ParseUint(values[index], 10, 32); err != nil {
			return fmt.Errorf("[SessionSlice] failed to read route diversity from redis data: %v", err)
		}
		index++

		s.RouteDiversity = uint32(routeDiversity)
	}

	if err := s.Envelope.ParseRedisString([]string{values[index], values[index+1]}); err != nil {
		return fmt.Errorf("[SessionSlice] failed to read envelope from redis data: %v", err)
	}
	index += 2

	if s.OnNetworkNext, err = strconv.ParseBool(values[index]); err != nil {
		return fmt.Errorf("[SessionSlice] failed to read on network next from redis data: %v", err)
	}
	index++

	if s.IsMultiPath, err = strconv.ParseBool(values[index]); err != nil {
		return fmt.Errorf("[SessionSlice] failed to read is multipath from redis data: %v", err)
	}
	index++

	if s.IsTryBeforeYouBuy, err = strconv.ParseBool(values[index]); err != nil {
		return fmt.Errorf("[SessionSlice] failed to read is try before you buy from redis data: %v", err)
	}
	index++

	return nil
}

type SessionMapPoint struct {
	Version   uint32
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	SessionID uint64  `json:"id"`
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

	s.Version = version

	if !encoding.ReadFloat64(data, &index, &s.Latitude) {
		return errors.New("[SessionMapPoint] invalid read at latitude")
	}

	if !encoding.ReadFloat64(data, &index, &s.Longitude) {
		return errors.New("[SessionMapPoint] invalid read at longitude")
	}

	if version >= 1 {
		if !encoding.ReadUint64(data, &index, &s.SessionID) {
			return errors.New("[SessionMapPoint] invalid read at session ID")
		}
	}

	return nil
}

func (s SessionMapPoint) MarshalBinary() ([]byte, error) {
	data := make([]byte, s.Size())
	index := 0

	encoding.WriteUint32(data, &index, SessionMapPointVersion)
	encoding.WriteFloat64(data, &index, s.Latitude)
	encoding.WriteFloat64(data, &index, s.Longitude)
	if s.Version >= 1 {
		encoding.WriteUint64(data, &index, s.SessionID)
	}

	return data, nil
}

func (s *SessionMapPoint) Serialize(stream encoding.Stream) error {
	stream.SerializeUint32(&s.Version)

	stream.SerializeFloat64(&s.Latitude)
	stream.SerializeFloat64(&s.Longitude)

	if s.Version >= 1 {
		stream.SerializeUint64(&s.SessionID)
	}

	return stream.Error()
}

func (s SessionMapPoint) Size() uint64 {
	return 4 + 8 + 8 + 8
}

func WriteSessionMapPoint(entry *SessionMapPoint) ([]byte, error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("recovered from panic during SessionMapPoint packet entry write: %v\n", r)
		}
	}()

	size := entry.Size()
	buffer := make([]byte, size)

	ws, err := encoding.CreateWriteStream(buffer[:])
	if err != nil {
		return nil, err
	}

	if err := entry.Serialize(ws); err != nil {
		return nil, err
	}
	ws.Flush()

	return buffer[:ws.GetBytesProcessed()], nil
}

func ReadSessionMapPoint(entry *SessionMapPoint, data []byte) error {
	if err := entry.Serialize(encoding.CreateReadStream(data)); err != nil {
		return err
	}
	return nil
}

func (s SessionMapPoint) RedisString() string {
	return fmt.Sprintf("%d|%.2f|%.2f|%016x", s.Version, s.Latitude, s.Longitude, s.SessionID)
}

func (s *SessionMapPoint) ParseRedisString(values []string) error {
	var index int
	var err error

	var version int64
	if version, err = strconv.ParseInt(values[index], 10, 64); err != nil {
		return fmt.Errorf("[SessionMapPoint] failed to read version from redis data: %v", err)
	}

	// The original data didn't have the version serialized
	if math.Abs(float64(version)) > SessionMapPointVersion {
		version = 0
	} else {
		s.Version = uint32(version)
		index++
	}

	if s.Latitude, err = strconv.ParseFloat(values[index], 64); err != nil {
		return fmt.Errorf("[SessionMapPoint] failed to read latitude from redis data: %v", err)
	}
	index++

	if s.Longitude, err = strconv.ParseFloat(values[index], 64); err != nil {
		return fmt.Errorf("[SessionMapPoint] failed to read longitude from redis data: %v", err)
	}
	index++

	// The original data didn't have the version serialized
	if s.Version >= 1 {
		if s.SessionID, err = strconv.ParseUint(values[index], 16, 64); err != nil {
			return fmt.Errorf("[SessionMapPoint] failed to read session ID from redis data: %v", err)
		}
		index++
	}

	return nil
}
