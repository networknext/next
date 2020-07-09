package routing

import "fmt"

// DatacenterMap maps buyer's alias names to the actual Datacenter
type DatacenterMap struct {
	BuyerID    string `json:"buyer_id"`
	Datacenter string `json:"datacenter"`
	Alias      string `json:"alias"`
}

func (dcm DatacenterMap) String() string {
	return fmt.Sprintf("{\n\tBuyer ID     : %v\n\tDatacenter ID: %v\n\tAlias        : %v\n}", dcm.BuyerID, dcm.Datacenter, dcm.Alias)
}
