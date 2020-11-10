package routing

type Seller struct {
	ID                        string
	Name                      string
	ShortName                 string
	CompanyCode               string
	IngressPriceNibblinsPerGB Nibblin
	EgressPriceNibblinsPerGB  Nibblin
}
