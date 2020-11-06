package routing

import "fmt"

// Seller
// TODO: ID, CompanyCode and ShortName all serve the same purpose here, though ID
//       is assigned as shown in syncSellers(). Clean this up.
type Seller struct {
	ID                        string // internal use, this is assigned to the parent Customer.CompanyCode in syncSellers()
	Name                      string // TODO: drop - defined by parent customer
	CompanyCode               string // TODO: drop - defined by parent customer
	ShortName                 string // WIP: independent but unqique within sellers - same as ID?
	IngressPriceNibblinsPerGB Nibblin
	EgressPriceNibblinsPerGB  Nibblin
	DatabaseID                int64 // seller_id db PK
	CustomerID                int64 // customer_id FK
}

func (s *Seller) String() string {

	seller := "\nrouting.Seller:\n"
	seller += "\tID                       : '" + s.ID + "'\n"
	seller += "\tName                     : '" + s.Name + "'\n"
	seller += "\tCompanyCode              : '" + s.CompanyCode + "'\n"
	seller += "\tShortName                : '" + s.ShortName + "'\n"
	seller += "\tIngressPriceNibblinsPerGB: " + fmt.Sprintf("%v", s.IngressPriceNibblinsPerGB) + "\n"
	seller += "\tEgressPriceNibblinsPerGB : " + fmt.Sprintf("%v", s.EgressPriceNibblinsPerGB) + "\n"
	seller += "\tDatabaseID               : " + fmt.Sprintf("%d", s.DatabaseID) + "\n"
	seller += "\tCustomerID               : " + fmt.Sprintf("%d", s.CustomerID) + "\n"

	return seller
}
