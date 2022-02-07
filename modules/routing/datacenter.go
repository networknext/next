package routing

import (
	"fmt"
)

const (
	MaxDatacenterNameLength = 63
)

var UnknownDatacenter = Datacenter{
	ID:   0,
	Name: "unknown",
}

type Datacenter struct {
	ID         uint64
	Name       string
	AliasName  string // TODO: remove (convenience field in server_handlers.go)
	Location   Location
	SellerID   int64 // sql FK
	DatabaseID int64 // sql PK
}

func (d *Datacenter) String() string {

	datacenter := "\nrouting.Datacenter:\n"
	datacenter += "\tID (hex)     : " + fmt.Sprintf("%16x", d.ID) + "\n"
	datacenter += "\tID           : " + fmt.Sprintf("%d", d.ID) + "\n"
	datacenter += "\tName         : " + d.Name + "\n"
	datacenter += "\tLocation     : TBD\n"
	datacenter += "\tSellerID     : " + fmt.Sprintf("%d", d.SellerID) + "\n"
	datacenter += "\tDatabaseID   : " + fmt.Sprintf("%d", d.DatabaseID) + "\n"

	return datacenter
}
