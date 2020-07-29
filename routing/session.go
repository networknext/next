package routing

import (
	"strings"
	"time"
	"unicode/utf8"

	jsoniter "github.com/json-iterator/go"
)

type SessionCountData struct {
	TotalNumDirectSessions    uint64
	TotalNumNextSessions      uint64
	NumDirectSessionsPerBuyer map[uint64]uint64
	NumNextSessionsPerBuyer   map[uint64]uint64
}

func (s *SessionCountData) UnmarshalBinary(data []byte) error {
	return jsoniter.Unmarshal(data, s)
}

func (s SessionCountData) MarshalBinary() ([]byte, error) {
	return jsoniter.Marshal(s)
}

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
	ID              string   `json:"id"`
	UserHash        string   `json:"user_hash"`
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
	Connection      string   `json:"connection"`
	NearbyRelays    []Relay  `json:"nearby_relays"`
	Platform        string   `json:"platform"`
	BuyerID         string   `json:"customer_id"`
}

func (s *SessionMeta) UnmarshalBinary(data []byte) error {
	return jsoniter.Unmarshal(data, s)
}

func (s SessionMeta) MarshalBinary() ([]byte, error) {
	return jsoniter.Marshal(s)
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
	return jsoniter.Unmarshal(data, s)
}

func (s SessionSlice) MarshalBinary() ([]byte, error) {
	return jsoniter.Marshal(s)
}

type SessionMapPoint struct {
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	OnNetworkNext bool    `json:"on_network_next"`
}

func (s *SessionMapPoint) UnmarshalBinary(data []byte) error {
	return jsoniter.Unmarshal(data, s)
}

func (s SessionMapPoint) MarshalBinary() ([]byte, error) {
	return jsoniter.Marshal(s)
}
