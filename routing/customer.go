package routing

import (
	"fmt"
	"strconv"

	"cloud.google.com/go/firestore"
)

type Customer struct {
	Code                   string
	Name                   string
	AutomaticSignInDomains string
	Active                 bool
	Debug                  bool
	BuyerRef               *firestore.DocumentRef // TODO: chopping block
	SellerRef              *firestore.DocumentRef // TODO: chopping block
	CustomerID             int64                  // customer_id - sql PK
	// BuyerID                uint64 // binary.LittleEndian.Uint64(publicKey[:8]),
	// SellerID               string // ID: name
}

func (c *Customer) String() string {

	customer := "\nrouting.Customer      :\n"
	customer += "\tCode                  : '" + c.Code + "'\n"
	customer += "\tName                  : '" + c.Name + "'\n"
	customer += "\tAutomaticSignInDomains: '" + c.AutomaticSignInDomains + "'\n"
	customer += "\tActive                : " + strconv.FormatBool(c.Active) + "\n"
	customer += "\tDebug                 : " + strconv.FormatBool(c.Debug) + "\n"
	customer += "\tCustomerID            : " + fmt.Sprintf("%d", c.CustomerID) + "\n"

	return customer
}
