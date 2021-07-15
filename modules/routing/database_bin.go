package routing

const (
	MaxDatabaseBinWrapperSize = 100000000
)

// DatabaseBinWrapper contains all the data from the database for
// static use by the relay_gateway, relay_backend, and server_backend
type DatabaseBinWrapper struct {
	CreationTime   string
	Creator        string
	Relays         []Relay
	RelayMap       map[uint64]Relay
	BuyerMap       map[uint64]Buyer
	SellerMap      map[string]Seller
	DatacenterMap  map[uint64]Datacenter
	DatacenterMaps map[uint64]map[uint64]DatacenterMap
	//                 ^ Buyer.ID   ^ DatacenterMap map index
}

func CreateEmptyDatabaseBinWrapper() *DatabaseBinWrapper {
	wrapper := &DatabaseBinWrapper{
		CreationTime:   "",
		Creator:        "",
		Relays:         []Relay{},
		RelayMap:       make(map[uint64]Relay),
		BuyerMap:       make(map[uint64]Buyer),
		SellerMap:      make(map[string]Seller),
		DatacenterMap:  make(map[uint64]Datacenter),
		DatacenterMaps: make(map[uint64]map[uint64]DatacenterMap),
	}

	return wrapper
}

func (wrapper DatabaseBinWrapper) IsEmpty() bool {
	if len(wrapper.RelayMap) != 0 {
		return false
	} else if len(wrapper.BuyerMap) != 0 {
		return false
	} else if len(wrapper.SellerMap) != 0 {
		return false
	} else if len(wrapper.DatacenterMap) != 0 {
		return false
	} else if len(wrapper.DatacenterMaps) != 0 {
		return false
	} else if wrapper.CreationTime == "" {
		return false
	} else if wrapper.Creator == "" {
		return false
	} else if len(wrapper.Relays) != 0 {
		return false
	}

	return true
}
