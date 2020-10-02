package routing

type Seller struct {
	ID                        string
	Name                      string
	CompanyCode               string
	IngressPriceNibblinsPerGB Nibblin
	EgressPriceNibblinsPerGB  Nibblin
}
