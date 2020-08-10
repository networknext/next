package routing

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
)

type Buyer struct {
	ID                   uint64
	Name                 string
	Domain               string
	Active               bool
	Live                 bool
	PublicKey            []byte
	RoutingRulesSettings RoutingRulesSettings
}

func (b *Buyer) EncodedPublicKey() string {
	totalPubkey := make([]byte, 0)
	buyerIDBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(buyerIDBytes, b.ID)

	totalPubkey = append(totalPubkey, buyerIDBytes...)
	totalPubkey = append(totalPubkey, b.PublicKey...)
	return base64.StdEncoding.EncodeToString(totalPubkey)
}

func (b *Buyer) DecodedPublicKey(key string) error {
	var err error
	b.PublicKey, err = base64.StdEncoding.DecodeString(key)

	return err
}

type Envelope struct {
	Up   int64 `json:"up"`
	Down int64 `json:"down"`
}

func (e Envelope) RedisString() string {
	return fmt.Sprintf("%d|%d", e.Up, e.Down)
}
