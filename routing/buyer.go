package routing

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"strconv"

	"github.com/networknext/backend/modules/core"
)

type Buyer struct {
	CompanyCode          string
	ID                   uint64
	Live                 bool
	Debug                bool
	PublicKey            []byte
	RouteShader          core.RouteShader
	InternalConfig       core.InternalConfig
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

func (e *Envelope) ParseRedisString(values []string) error {
	var index int
	var err error

	if e.Up, err = strconv.ParseInt(values[index], 10, 64); err != nil {
		return fmt.Errorf("[Envelope] failed to read up from redis data: %v", err)
	}
	index++

	if e.Down, err = strconv.ParseInt(values[index], 10, 64); err != nil {
		return fmt.Errorf("[Envelope] failed to read down from redis data: %v", err)
	}
	index++

	return nil
}
