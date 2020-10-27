package routing

type Seller struct {
	ID                        string // internal use
	Name                      string
	CompanyCode               string
	IngressPriceNibblinsPerGB Nibblin
	EgressPriceNibblinsPerGB  Nibblin
	SellerID                  int64 // seller_id db PK
}
