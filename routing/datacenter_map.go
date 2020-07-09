package routing

import "fmt"

// DatacenterMap maps buyer's alias names to the actual Datacenter
type DatacenterMap struct {
	BuyerID    string
	Datacenter string
	Alias      string
}

func (dcm *DatacenterMap) String() string {
	return fmt.Sprintf("{\n\tBuyerID      : %s\n\tDatacenter ID: %s\n\tAlias        : %s\n}", dcm.BuyerID, dcm.Datacenter, dcm.Alias)
}
