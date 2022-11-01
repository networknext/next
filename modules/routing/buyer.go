package routing

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"strconv"

	"github.com/networknext/backend/modules/core"
)

type Buyer struct {
	CompanyCode         string // TODO: chopping block - defined by the parent customer
	Alias               string
	ID                  uint64
	HexID               string // needed by external (non-go) clients
	Live                bool
	Debug               bool
	Analytics           bool
	Billing             bool
	Trial               bool
	Verified            bool
	ExoticLocationFee   float64
	StandardLocationFee float64
	PublicKey           []byte
	RouteShader         core.RouteShader
	InternalConfig      core.InternalConfig
	DatabaseID          int64 // sql PK
	CustomerID          int64 // sql FK
	LookerSeats         int64
}

func (b *Buyer) String() string {

	buyer := "\nrouting.Buyer:\n"
	buyer += "\tID (hex)              : " + fmt.Sprintf("%16x", b.ID) + "\n"
	buyer += "\tID                    : " + fmt.Sprintf("%d", b.ID) + "\n"
	buyer += "\tAlias                 : '" + b.Alias + "'\n"
	buyer += "\tCompanyCode           : '" + b.CompanyCode + "'\n"
	buyer += "\tLive                  : " + strconv.FormatBool(b.Live) + "\n"
	buyer += "\tDebug                 : " + strconv.FormatBool(b.Debug) + "\n"
	buyer += "\tAnalytics             : " + strconv.FormatBool(b.Analytics) + "\n"
	buyer += "\tBilling               : " + strconv.FormatBool(b.Billing) + "\n"
	buyer += "\tTrial                 : " + strconv.FormatBool(b.Trial) + "\n"
	buyer += "\tVerified                : " + strconv.FormatBool(b.Verified) + "\n"
	buyer += "\tExotic Location Fee   : " + fmt.Sprintf("%f", b.ExoticLocationFee) + "\n"
	buyer += "\tStandard Location Fee : " + fmt.Sprintf("%f", b.StandardLocationFee) + "\n"
	buyer += "\tPublicKey             : " + string(b.PublicKey) + "\n"
	buyer += "\tRouteShader           : TBD\n"
	buyer += "\tInternalConfig        : TBD\n"
	buyer += "\tDatabaseID            : " + fmt.Sprintf("%d", b.DatabaseID) + "\n"
	buyer += "\tCustomerID            : " + fmt.Sprintf("%d", b.CustomerID) + "\n"
	buyer += "\tLooker Seats          : " + fmt.Sprintf("%d", b.LookerSeats) + "\n"

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
