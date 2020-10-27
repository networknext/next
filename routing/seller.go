package routing

import "fmt"

type Seller struct {
	ID                        string // internal use, this is identically equal to the parent Customer.CompanyCode
	Name                      string // TODO: drop - defined by parent customer
	CompanyCode               string // TODO: drop - defined by parent customer
	IngressPriceNibblinsPerGB Nibblin
	EgressPriceNibblinsPerGB  Nibblin
	SellerID                  int64 // seller_id db PK
	CustomerID                int64 // customer_id FK
}

func (s *Seller) String() string {

	seller := "\nrouting.Seller:\n"
	seller += "\tID                       : '" + s.ID + "'\n"
	seller += "\tCompanyCode              : '" + s.CompanyCode + "'\n"
	seller += "\tIngressPriceNibblinsPerGB: " + fmt.Sprintf("%v", s.IngressPriceNibblinsPerGB) + "\n"
	seller += "\tEgressPriceNibblinsPerGB : " + fmt.Sprintf("%v", s.EgressPriceNibblinsPerGB) + "\n"
	seller += "\tSellerID                 : " + fmt.Sprintf("%d", s.SellerID) + "\n"
	seller += "\tCustomerID               : " + fmt.Sprintf("%d", s.CustomerID) + "\n"

	return seller
}
