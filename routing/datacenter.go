package routing

var UnknownDatacenter = Datacenter{
	ID:      0,
	Name:    "unknown",
	Enabled: false,
}

type Datacenter struct {
	ID        uint64
	Name      string
	AliasName string
	Enabled   bool
	Location  Location
}
