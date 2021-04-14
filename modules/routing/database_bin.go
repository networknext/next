package routing

const (
    MaxDatabaseBinWrapperSize = 100000000
)

// DatabaseBinWrapper contains all the data from the database for
// static use by the relay_backend and server_backend
type DatabaseBinWrapper struct {
	Relays         []Relay
	Buyers         []Buyer
	Sellers        []Seller
	Datacenters    []Datacenter
	DatacenterMaps map[uint64]map[uint64]DatacenterMap
	//                 ^ Buyer.ID   ^ DatacenterMap map index
}
