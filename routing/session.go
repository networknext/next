package routing

import (
	"time"

	jsoniter "github.com/json-iterator/go"
)

type SessionMeta struct {
	Location   Location `json:"location"`
	ClientAddr string   `json:"client_addr"`
	ServerAddr string   `json:"server_addr"`
	Stats      Stats    `json:"stats"`
	Hops       int      `json:"hops"`
	SDK        string   `json:"sdk"`
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
