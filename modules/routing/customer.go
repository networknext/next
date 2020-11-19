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
	DatabaseID             int64                  // customer_id - sql PK
}

func (c *Customer) String() string {

	customer := "\nrouting.Customer      :\n"
	customer += "\tCode                  : '" + c.Code + "'\n"
	customer += "\tName                  : '" + c.Name + "'\n"
	customer += "\tAutomaticSignInDomains: '" + c.AutomaticSignInDomains + "'\n"
	customer += "\tActive                : " + strconv.FormatBool(c.Active) + "\n"
	customer += "\tDebug                 : " + strconv.FormatBool(c.Debug) + "\n"
	customer += "\tDatabaseID            : " + fmt.Sprintf("%d", c.DatabaseID) + "\n"

	return customer
}
