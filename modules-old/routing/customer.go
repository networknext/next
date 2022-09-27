package routing

import (
	"fmt"

	"cloud.google.com/go/firestore"
)

type Customer struct {
	Code                    string
	Name                    string
	AutomaticSignInDomains  string
	BuyerTOSSignerFirstName string
	BuyerTOSSignerLastName  string
	BuyerTOSSignerEmail     string
	BuyerTOSSignedTimestamp string
	ShowAnalytics           bool
	ShowBilling             bool
	BuyerRef                *firestore.DocumentRef // TODO: chopping block
	SellerRef               *firestore.DocumentRef // TODO: chopping block
	DatabaseID              int64                  // customer_id - sql PK
}

func (c *Customer) String() string {

	customer := "\nrouting.Customer      :\n"
	customer += "\tCode                  : '" + c.Code + "'\n"
	customer += "\tName                  : '" + c.Name + "'\n"
	customer += "\tTOSSigner             : '" + c.BuyerTOSSignerFirstName + " " + c.BuyerTOSSignerLastName + ": " + c.BuyerTOSSignerEmail + "'\n"
	customer += "\tTOSTimestamp          : '" + c.BuyerTOSSignedTimestamp + "'\n"
	customer += "\tAutomaticSignInDomains: '" + c.AutomaticSignInDomains + "'\n"
	customer += "\tDatabaseID            : " + fmt.Sprintf("%d", c.DatabaseID) + "\n"

	return customer
}
