package routing

import "fmt"

// DatacenterMap maps buyer's alias names to the actual Datacenter
type DatacenterMap struct {
	BuyerID      uint64 `json:"buyer_id"`
	DatacenterID uint64 `json:"datacenter_id"`
	Alias        string `json:"alias"`
}

func (dcm DatacenterMap) String() string {

	dcMap := "\nrouting.DatacenterMap\n"
	dcMap += "\tAlias              : " + dcm.Alias + "\n"
	dcMap += "\tBuyer ID (hex)     : " + fmt.Sprintf("%016x", dcm.BuyerID) + "\n"
	dcMap += "\tBuyer ID           : " + fmt.Sprintf("%d", dcm.BuyerID) + "\n"
	dcMap += "\tdatacenter ID (hex): " + fmt.Sprintf("%016x", dcm.DatacenterID) + "\n"
	dcMap += "\tdatacenter ID      : " + fmt.Sprintf("%d", dcm.DatacenterID) + "\n"

	return dcMap
}
