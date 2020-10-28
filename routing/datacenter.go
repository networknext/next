package routing

var UnknownDatacenter = Datacenter{
	ID:      0,
	Name:    "unknown",
	Enabled: false,
}

type Datacenter struct {
	ID            uint64
	SignedID      int64
	Name          string
	AliasName     string
	Enabled       bool
	Location      Location
	SupplierName  string
	StreetAddress string
	SellerID      int64 // sql FK
	DatacenterID  int64 // sql PK
}
