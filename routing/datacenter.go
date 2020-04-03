package routing

type Datacenter struct {
	ID       uint64
	Name     string
	Enabled  bool
	Location Location
}
