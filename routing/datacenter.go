package routing

import (
	"fmt"
	"strconv"
)

var UnknownDatacenter = Datacenter{
	ID:      0,
	Name:    "unknown",
	Enabled: false,
}

type Datacenter struct {
	ID            uint64
	SignedID      int64
	Name          string
	AliasName     string // convenience field in server_handlers.go
	Enabled       bool
	Location      Location
	SupplierName  string
	StreetAddress string
	SellerID      int64 // sql FK
	DatacenterID  int64 // sql PK
}

func (d *Datacenter) String() string {

	datacenter := "\nrouting.Datacenter:\n"
	datacenter += "\tID (hex)     : " + fmt.Sprintf("%16x", d.ID) + "\n"
	datacenter += "\tID           : " + fmt.Sprintf("%d", d.ID) + "\n"
	datacenter += "\tName         : " + d.Name + "\n"
	datacenter += "\tEnabled      : " + strconv.FormatBool(d.Enabled) + "\n"
	datacenter += "\tLocation     : TBD\n"
	datacenter += "\tSupplierName : " + d.SupplierName + "\n"
	datacenter += "\tStreetAddress: " + d.StreetAddress + "\n"
	datacenter += "\tSellerID     : " + fmt.Sprintf("%d", d.SellerID) + "\n"
	datacenter += "\tDatacenterID : " + fmt.Sprintf("%d", d.DatacenterID) + "\n"

	return datacenter
}
