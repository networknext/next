package routing

import "encoding/base64"

type Buyer struct {
	ID                   uint64
	Name                 string
	Active               bool
	Live                 bool
	PublicKey            []byte
	RoutingRulesSettings RoutingRulesSettings
}

func (b *Buyer) EncodedPublicKey() string {
	return base64.StdEncoding.EncodeToString(b.PublicKey)
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
