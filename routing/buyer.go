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
