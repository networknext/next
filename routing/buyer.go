package routing

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"strconv"

	"github.com/networknext/backend/modules/core"
)

// TODO: The RouteShader and InternalConfig fields need to be
// converted to slices or moved to different objects.
type Buyer struct {
	CompanyCode    string // TODO: remove field? It is defined by the parent customer
	ID             uint64
	Live           bool
	Debug          bool
	PublicKey      []byte
	RouteShader    core.RouteShader
	InternalConfig core.InternalConfig
	BuyerID        int64 // sql PK
	CustomerID     int64 // sql FK
}

func (b *Buyer) String() string {

	buyer := "\nrouting.Buyer:\n"
	buyer += "\tID            : " + fmt.Sprintf("%d", b.ID) + "\n"
	buyer += "\tLive          : " + strconv.FormatBool(b.Live) + "\n"
	buyer += "\tDebug         : " + strconv.FormatBool(b.Debug) + "\n"
	buyer += "\tPublicKey     : " + string(b.PublicKey) + "\n"
	buyer += "\tRouteShader   : TBD\n"
	buyer += "\tInternalConfig: TBD\n"
	buyer += "\tBuyerID       : " + fmt.Sprintf("%d", b.BuyerID) + "\n"
	buyer += "\tCustomerID    : " + fmt.Sprintf("%d", b.CustomerID) + "\n"

	return buyer
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
