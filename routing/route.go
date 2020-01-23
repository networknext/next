package routing

import "encoding/json"

const (
	DecisionTypeDirect   = 0
	DecisionTypeNew      = 1
	DecisionTypeContinue = 2
)

var (
	DecisionDirect = Decision{Type: DecisionTypeDirect}
)

type Decision struct {
	Type      int
	NumTokens int
	Tokens    []byte
	Multipath bool
}

func (r Decision) MarshalBinary() ([]byte, error) {
	return json.Marshal(r)
}

func (r *Decision) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, r)
}

type Route struct {
	Relays []Relay
	Stats  Stats
}
