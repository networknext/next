package routing

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	jsoniter "github.com/json-iterator/go"
	"github.com/networknext/backend/encoding"
)

const (
	SessionMetaVersion     = 0
	SessionSliceVersion    = 0
	SessionMapPointVersion = 0
)

type SessionData struct {
	Meta  SessionMeta     `json:"meta"`
	Slice SessionSlice    `json:"slice"`
	Point SessionMapPoint `json:"point"`
}

func (s *SessionData) UnmarshalBinary(data []byte) error {
	return jsoniter.Unmarshal(data, s)
}

func (s SessionData) MarshalBinary() ([]byte, error) {
	return jsoniter.Marshal(s)
}

type SessionMeta struct {
	ID              uint64   `json:"id"`
	UserHash        uint64   `json:"user_hash"`
	DatacenterName  string   `json:"datacenter_name"`
	DatacenterAlias string   `json:"datacenter_alias"`
	OnNetworkNext   bool     `json:"on_network_next"`
	NextRTT         float64  `json:"next_rtt"`
	DirectRTT       float64  `json:"direct_rtt"`
	DeltaRTT        float64  `json:"delta_rtt"`
	Location        Location `json:"location"`
	ClientAddr      string   `json:"client_addr"`
	ServerAddr      string   `json:"server_addr"`
	Hops            []Relay  `json:"hops"`
	SDK             string   `json:"sdk"`
	Connection      uint8    `json:"connection"`
	NearbyRelays    []Relay  `json:"nearby_relays"`
	Platform        uint8    `json:"platform"`
	BuyerID         uint64   `json:"customer_id"`
}

func (s *SessionMeta) UnmarshalBinary(data []byte) error {
	return jsoniter.Unmarshal(data, s)
}

func (s SessionMeta) MarshalBinary() ([]byte, error) {
	return jsoniter.Marshal(s)
}

func (s SessionMeta) Size() uint64 {
	return uint64(4 + 4*len(s.ID) + 4*len(s.UserHash) + 4*len(s.DatacenterName) + 4*len(s.DatacenterAlias) + 1 + 8 + 8 + 8 +
		/*LOCATION*/ +4*len(s.ClientAddr) + 4*len(s.ServerAddr) /*HOPS*/ + 4*len(s.SDK) + 4*len(s.Connection) /*NEARBY_RELAYS*/ + 4*len(s.Platform) + 4*len(s.BuyerID))
}

func (s *SessionMeta) Anonymise() {
	s.ServerAddr = ObscureString(s.ServerAddr, ".", -1)
	s.BuyerID = ""
	s.NearbyRelays = []Relay{}
	s.Hops = []Relay{}
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
	Timestamp         time.Time `json:"timestamp"`
	Next              Stats     `json:"next"`
	Direct            Stats     `json:"direct"`
	Envelope          Envelope  `json:"envelope"`
	OnNetworkNext     bool      `json:"on_network_next"`
	IsMultiPath       bool      `json:"is_multipath"`
	IsTryBeforeYouBuy bool      `json:"is_try_before_you_buy"`
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

	var next Stats
	if !encoding.ReadFloat64(data, &index, &next.RTT) {
		return errors.New("[SessionSlice] invalid read at next RTT")
	}
	if !encoding.ReadFloat64(data, &index, &next.Jitter) {
		return errors.New("[SessionSlice] invalid read at next jitter")
	}
	if !encoding.ReadFloat64(data, &index, &next.PacketLoss) {
		return errors.New("[SessionSlice] invalid read at next packet loss")
	}

	var direct Stats
	if !encoding.ReadFloat64(data, &index, &direct.RTT) {
		return errors.New("[SessionSlice] invalid read at direct RTT")
	}
	if !encoding.ReadFloat64(data, &index, &direct.Jitter) {
		return errors.New("[SessionSlice] invalid read at direct jitter")
	}
	if !encoding.ReadFloat64(data, &index, &direct.PacketLoss) {
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

	s.Envelope = Envelope{Up: int64(up), Down: int64(down)}

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
