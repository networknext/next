package routing

var UnknownDatacenter = Datacenter{
	ID:      0,
	Name:    "Unknown",
	Enabled: false,
}

type Datacenter struct {
	ID       uint64
	Name     string
	Enabled  bool
	Location Location
}
