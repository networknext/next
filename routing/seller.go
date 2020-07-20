package routing

type Seller struct {
	ID                        string
	Name                      string
	IngressPriceNibblinsPerGB Nibblin
	EgressPriceNibblinsPerGB  Nibblin
}
