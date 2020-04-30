package routing

import (
	"time"

	jsoniter "github.com/json-iterator/go"
)

type SessionMeta struct {
	ID         string   `json:"id"`
	UserHash   string   `json:"user_hash"`
	Datacenter string   `json:"datacenter"`
	NextRTT    float64  `json:"next_rtt"`
	DirectRTT  float64  `json:"direct_rtt"`
	DeltaRTT   float64  `json:"delta_rtt"`
	Location   Location `json:"location"`
	ClientAddr string   `json:"client_addr"`
	ServerAddr string   `json:"server_addr"`
	Hops       int      `json:"hops"`
	SDK        string   `json:"sdk"`
	Connection string   `json:"connection"`
}

func (s *SessionMeta) UnmarshalBinary(data []byte) error {
	return jsoniter.Unmarshal(data, s)
}

func (s SessionMeta) MarshalBinary() ([]byte, error) {
	return jsoniter.Marshal(s)
}

type SessionSlice struct {
	Timestamp time.Time `json:"timestamp"`
	Next      Stats     `json:"next"`
	Direct    Stats     `json:"direct"`
	Envelope  Envelope  `json:"envelope"`
}

func (s *SessionSlice) UnmarshalBinary(data []byte) error {
	return jsoniter.Unmarshal(data, s)
}

func (s SessionSlice) MarshalBinary() ([]byte, error) {
	return jsoniter.Marshal(s)
}
