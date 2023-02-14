package database

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"hash/fnv"
	"io"
	"io/ioutil"
	"net"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/networknext/backend/modules/core"

	_ "github.com/lib/pq"
	"github.com/modood/table"
)

func HashString(s string) uint64 {
	hash := fnv.New64a()
	hash.Write([]byte(s))
	return hash.Sum64()
}

type Relay struct {
	Id                 uint64
	Name               string
	DatacenterId       uint64
	PublicAddress      net.UDPAddr
	HasInternalAddress bool
	InternalAddress    net.UDPAddr
	InternalGroup      uint64
	SSHAddress         net.UDPAddr
	SSHUser            string
	PublicKey          []byte
	PrivateKey         []byte
	MaxSessions        int
	PortSpeed          int
	Version            string
	Seller             Seller
	Datacenter         Datacenter
}

type Buyer struct {
	Id          uint64
	Name        string
	Live        bool
	Debug       bool
	PublicKey   []byte
	RouteShader core.RouteShader
}

type Seller struct {
	Id   uint64
	Name string
}

type Datacenter struct {
	Id        uint64
	Name      string
	Latitude  float32
	Longitude float32
}

type DatacenterMap struct {
	BuyerId            uint64
	DatacenterId       uint64
	EnableAcceleration bool
}

type Database struct {
	CreationTime   string
	Creator        string
	Relays         []Relay
	RelayMap       map[uint64]*Relay
	BuyerMap       map[uint64]*Buyer
	SellerMap      map[uint64]*Seller
	DatacenterMap  map[uint64]*Datacenter
	DatacenterMaps map[uint64]map[uint64]*DatacenterMap
	//                   ^ BuyerId  ^ DatacenterId
}

func CreateDatabase() *Database {

	database := &Database{
		CreationTime:   "",
		Creator:        "",
		Relays:         []Relay{},
		RelayMap:       make(map[uint64]*Relay),
		BuyerMap:       make(map[uint64]*Buyer),
		SellerMap:      make(map[uint64]*Seller),
		DatacenterMap:  make(map[uint64]*Datacenter),
		DatacenterMaps: make(map[uint64]map[uint64]*DatacenterMap),
	}

	return database
}

func LoadDatabase(filename string) (*Database, error) {

	databaseFile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer databaseFile.Close()

	database := &Database{}

	err = gob.NewDecoder(databaseFile).Decode(database)

	return database, err
}

func (database *Database) Save(filename string) error {

	var buffer bytes.Buffer

	err := gob.NewEncoder(&buffer).Encode(database)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(filename, buffer.Bytes(), 0644); err != nil {
		return err
	}

	return nil
}

func (database *Database) Validate() error {

	if database.IsEmpty() {
		return fmt.Errorf("database is empty")
	}

	for id, datacenter := range database.DatacenterMap {
		if id != datacenter.Id {
			return fmt.Errorf("datacenter %s has mismatched id (%d vs. %d)", datacenter.Name, datacenter.Id, id)
		}
	}

	for id, seller := range database.SellerMap {
		if id != seller.Id {
			return fmt.Errorf("seller %s has mismatched id (%d vs. %d)", seller.Name, seller.Id, id)
		}
	}

	for id, buyer := range database.BuyerMap {
		if id != buyer.Id {
			return fmt.Errorf("buyer %s has mismatched id (%d vs. %d)", buyer.Name, buyer.Id, id)
		}
	}

	for id, relay := range database.RelayMap {
		if id != relay.Id {
			return fmt.Errorf("relay %s has mismatched id (%d vs. %d)", relay.Name, relay.Id, id)
		}
		_, sellerExists := database.SellerMap[relay.Seller.Id]
		if !sellerExists {
			return fmt.Errorf("relay %s seller does not exist", relay.Name)
		}
		_, datacenterExists := database.DatacenterMap[relay.Datacenter.Id]
		if !datacenterExists {
			return fmt.Errorf("relay %s datacenter does not exist", relay.Name)
		}
		if relay.PublicAddress.IP.String() == "0.0.0.0" {
			return fmt.Errorf("relay %s public address is 0.0.0.0", relay.Name)
		}
		if relay.PublicAddress.Port == 0 {
			return fmt.Errorf("relay %s public address port is zero", relay.Name)
		}
		if relay.HasInternalAddress {
			if relay.InternalAddress.IP.String() == "0.0.0.0" {
				return fmt.Errorf("relay %s internal address is 0.0.0.0", relay.Name)
			}
			if relay.InternalAddress.Port == 0 {
				return fmt.Errorf("relay %s internal address port is zero", relay.Name)
			}
		}
		if relay.SSHAddress.IP.String() == "0.0.0.0" {
			return fmt.Errorf("relay %s ssh address is 0.0.0.0", relay.Name)
		}
		if relay.SSHAddress.Port == 0 {
			return fmt.Errorf("relay %s ssh address port is zero", relay.Name)
		}
	}

	relayNames := make(map[string]int)
	relayAddresses := make(map[string]int)

	for _, relay := range database.RelayMap {
		if _, nameAlreadyExists := relayNames[relay.Name]; nameAlreadyExists {
			return fmt.Errorf("there is more than one relay with the name '%s'", relay.Name)
		}
		if _, addressAlreadyExists := relayAddresses[relay.PublicAddress.String()]; addressAlreadyExists {
			return fmt.Errorf("there is more than one relay with the public address '%s'", relay.PublicAddress.String())
		}
		relayNames[relay.Name] = 1
		relayAddresses[relay.PublicAddress.String()] = 1
	}

	for buyerId, buyerMap := range database.DatacenterMaps {
		for datacenterId, entry := range buyerMap {
			if entry.DatacenterId != datacenterId {
				return fmt.Errorf("bad datacenter id in datacenter maps: %d vs %d", datacenterId, entry.DatacenterId)
			}
			if entry.BuyerId != buyerId {
				return fmt.Errorf("bad buyer id in datacenter maps: %d vs %d", buyerId, entry.BuyerId)
			}
		}
	}

	return nil
}

func (database *Database) IsEmpty() bool {

	if database.CreationTime != "" {
		return false
	}

	if database.Creator != "" {
		return false
	}

	if len(database.Relays) != 0 {
		return false
	}

	if len(database.RelayMap) != 0 {
		return false
	}

	if len(database.BuyerMap) != 0 {
		return false
	}

	if len(database.SellerMap) != 0 {
		return false
	}

	if len(database.DatacenterMap) != 0 {
		return false
	}

	if len(database.DatacenterMaps) != 0 {
		return false
	}

	return true
}

func (database *Database) String() string {

	output := "Headers:\n\n"

	// header

	type HeaderRow struct {
		Creator      string
		CreationTime string
	}

	header := [1]HeaderRow{}

	header[0] = HeaderRow{Creator: database.Creator, CreationTime: database.CreationTime}

	output += table.Table(header[:])

	// buyers

	output += "\nBuyers:\n\n"

	type BuyerRow struct {
		Name            string
		Id              string
		Live            string
		Debug           string
		PublicKeyBase64 string
	}

	buyers := []BuyerRow{}

	for _, v := range database.BuyerMap {

		data := make([]byte, 8+32)
		binary.LittleEndian.PutUint64(data, v.Id)
		copy(data[8:], v.PublicKey)

		row := BuyerRow{
			Id:              fmt.Sprintf("%0x", v.Id),
			Name:            v.Name,
			Live:            fmt.Sprintf("%v", v.Live),
			Debug:           fmt.Sprintf("%v", v.Debug),
			PublicKeyBase64: base64.StdEncoding.EncodeToString(data),
		}

		buyers = append(buyers, row)
	}

	sort.SliceStable(buyers, func(i, j int) bool { return buyers[i].Name < buyers[j].Name })

	output += "\n\n"

	output += table.Table(buyers)

	// sellers

	output += "\n\nSellers:\n\n"

	type SellerRow struct {
		Name string
		Id   string
	}

	sellers := []SellerRow{}

	for _, v := range database.SellerMap {

		row := SellerRow{
			Id:   fmt.Sprintf("%0x", v.Id),
			Name: v.Name,
		}

		sellers = append(sellers, row)
	}

	sort.SliceStable(sellers, func(i, j int) bool { return sellers[i].Id < sellers[j].Id })

	output += table.Table(sellers)

	// datacenters

	output += "\n\nDatacenters:\n\n"

	type DatacenterRow struct {
		Name      string
		Id        string
		Latitude  string
		Longitude string
	}

	datacenters := []DatacenterRow{}

	for _, v := range database.DatacenterMap {

		row := DatacenterRow{
			Id:        fmt.Sprintf("%0x", v.Id),
			Name:      v.Name,
			Latitude:  fmt.Sprintf("%+3.2f", v.Latitude),
			Longitude: fmt.Sprintf("%+3.2f", v.Longitude),
		}

		datacenters = append(datacenters, row)
	}

	sort.SliceStable(datacenters, func(i, j int) bool { return datacenters[i].Name < datacenters[j].Name })

	output += table.Table(datacenters)

	// relays

	output += "\n\nRelays:\n\n"

	type RelayRow struct {
		Name            string
		Id              string
		PublicAddress   string
		InternalAddress string
		PublicKey       string
		PrivateKey      string
	}

	relays := []RelayRow{}

	for _, v := range database.RelayMap {

		row := RelayRow{
			Id:            fmt.Sprintf("%0x", v.Id),
			Name:          v.Name,
			PublicAddress: v.PublicAddress.String(),
			PublicKey:     base64.StdEncoding.EncodeToString(v.PublicKey),
			PrivateKey:    base64.StdEncoding.EncodeToString(v.PrivateKey),
		}

		if v.HasInternalAddress {
			row.InternalAddress = v.InternalAddress.String()
		}

		relays = append(relays, row)
	}

	sort.SliceStable(relays, func(i, j int) bool { return relays[i].Name < relays[j].Name })

	output += table.Table(relays)

	// route shaders

	type PropertyRow struct {
		Property string
		Value    string
	}

	for _, v := range database.BuyerMap {

		output += fmt.Sprintf("\n\nRoute Shader for '%s'\n\n", v.Name)

		routeShader := v.RouteShader

		properties := []PropertyRow{}

		properties = append(properties, PropertyRow{"Disable Network Next", fmt.Sprintf("%v", routeShader.DisableNetworkNext)})
		properties = append(properties, PropertyRow{"Analysis Only", fmt.Sprintf("%v", routeShader.AnalysisOnly)})
		properties = append(properties, PropertyRow{"AB Test", fmt.Sprintf("%v", routeShader.ABTest)})
		properties = append(properties, PropertyRow{"Reduce Latency", fmt.Sprintf("%v", routeShader.ReduceLatency)})
		properties = append(properties, PropertyRow{"Reduce Packet Loss", fmt.Sprintf("%v", routeShader.ReducePacketLoss)})
		properties = append(properties, PropertyRow{"Multipath", fmt.Sprintf("%v", routeShader.Multipath)})
		properties = append(properties, PropertyRow{"Force Next", fmt.Sprintf("%v", routeShader.ForceNext)})
		properties = append(properties, PropertyRow{"Selection Percent", fmt.Sprintf("%d%%", routeShader.SelectionPercent)})
		properties = append(properties, PropertyRow{"Acceptable Latency", fmt.Sprintf("%dms", routeShader.AcceptableLatency)})
		properties = append(properties, PropertyRow{"Latency Threshold", fmt.Sprintf("%dms", routeShader.LatencyThreshold)})
		properties = append(properties, PropertyRow{"Acceptable Packet Loss", fmt.Sprintf("%.1f%%", routeShader.AcceptablePacketLoss)})
		properties = append(properties, PropertyRow{"Packet Loss Sustained", fmt.Sprintf("%.1f%%", routeShader.PacketLossSustained)})
		properties = append(properties, PropertyRow{"Bandwidth Envelope Up", fmt.Sprintf("%dkbps", routeShader.BandwidthEnvelopeUpKbps)})
		properties = append(properties, PropertyRow{"Bandwidth Envelope Down", fmt.Sprintf("%dkbps", routeShader.BandwidthEnvelopeDownKbps)})
		properties = append(properties, PropertyRow{"Route Select Threshold", fmt.Sprintf("%dms", routeShader.RouteSelectThreshold)})
		properties = append(properties, PropertyRow{"Route Switch Threshold", fmt.Sprintf("%dms", routeShader.RouteSwitchThreshold)})
		properties = append(properties, PropertyRow{"Max Latency Trade Off", fmt.Sprintf("%dms", routeShader.MaxLatencyTradeOff)})
		properties = append(properties, PropertyRow{"RTT Veto (Default)", fmt.Sprintf("%dms", routeShader.RTTVeto_Default)})
		properties = append(properties, PropertyRow{"RTT Veto (Multipath)", fmt.Sprintf("%dms", routeShader.RTTVeto_Multipath)})
		properties = append(properties, PropertyRow{"RTT Veto (PacketLoss)", fmt.Sprintf("%dms", routeShader.RTTVeto_PacketLoss)})
		properties = append(properties, PropertyRow{"Max Next RTT", fmt.Sprintf("%dms", routeShader.MaxNextRTT)})
		properties = append(properties, PropertyRow{"Route Diversity", fmt.Sprintf("%d", routeShader.RouteDiversity)})

		output += table.Table(properties)
	}

	// destination datacenters

	output += "\n\nDestination datacenters:\n\n"

	type DestinationDatacenterRow struct {
		Datacenter string
		Buyers     []string
	}

	destinationDatacenterMap := make(map[uint64]*DestinationDatacenterRow)

	for _, v1 := range database.DatacenterMaps {
		for _, v2 := range v1 {
			if !v2.EnableAcceleration {
				continue
			}
			buyerId := v2.BuyerId
			datacenterId := v2.DatacenterId
			entry := destinationDatacenterMap[datacenterId]
			if entry == nil {
				entry = &DestinationDatacenterRow{}
				entry.Datacenter = database.DatacenterMap[datacenterId].Name
				destinationDatacenterMap[datacenterId] = entry
			}
			entry.Buyers = append(entry.Buyers, database.BuyerMap[buyerId].Name)
		}
	}

	destinationDatacenters := make([]DestinationDatacenterRow, 0)

	for _, v := range destinationDatacenterMap {
		sort.SliceStable(v.Buyers, func(i, j int) bool { return v.Buyers[i] < v.Buyers[j] })
		destinationDatacenters = append(destinationDatacenters, *v)
	}

	sort.SliceStable(destinationDatacenters, func(i, j int) bool {
		return destinationDatacenters[i].Datacenter < destinationDatacenters[j].Datacenter
	})

	output += table.Table(destinationDatacenters)

	return output
}

// ---------------------------------------------------------------------

func (database *Database) WriteHTML(w io.Writer) {

	const htmlHeader = `<!DOCTYPE html>
	<html lang="en">
	<head>
	  <meta charset="utf-8">
	  <meta http-equiv="refresh" content="1">
	  <title>Database</title>
	  <style>
		table, th, td {
	      border: 1px solid black;
	      border-collapse: collapse;
	      text-align: center;
	      padding: 10px;
	    }
		*{
		  font-family:Courier;
		}	  
	  </style>
	</head>
	<body>`

	fmt.Fprintf(w, "%s\n", htmlHeader)

	// header

	fmt.Fprintf(w, "Header:<br><br>")
	fmt.Fprintf(w, "<table>\n")
	fmt.Fprintf(w, "<tr><td><b>%s</b></td><td><b>%s</b></td></tr>\n", "Creator", "Creation Time")
	fmt.Fprintf(w, "<tr><td>%s</td><td>%s</td></tr>\n", database.Creator, database.CreationTime)
	fmt.Fprintf(w, "</table>\n")

	// buyers

	type BuyerRow struct {
		Name            string
		Id              string
		Live            string
		Debug           string
		PublicKeyBase64 string
	}

	buyers := []BuyerRow{}

	for _, v := range database.BuyerMap {

		data := make([]byte, 8+32)
		binary.LittleEndian.PutUint64(data, v.Id)
		copy(data[8:], v.PublicKey)

		row := BuyerRow{
			Id:              fmt.Sprintf("%0x", v.Id),
			Name:            v.Name,
			Live:            fmt.Sprintf("%v", v.Live),
			Debug:           fmt.Sprintf("%v", v.Debug),
			PublicKeyBase64: base64.StdEncoding.EncodeToString(data),
		}

		buyers = append(buyers, row)
	}

	sort.SliceStable(buyers, func(i, j int) bool { return buyers[i].Name < buyers[j].Name })

	fmt.Fprintf(w, "<br><br>Buyers:<br><br>")
	fmt.Fprintf(w, "<table>\n")
	fmt.Fprintf(w, "<tr><td><b>%s</b></td><td><b>%s</b></td><td><b>%s</b></td><td><b>%s</b></td><td><b>%s</b></td></tr>\n", "Name", "Id", "Live", "Debug", "Public Key Base64")
	for i := range buyers {
		fmt.Fprintf(w, "<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>\n", buyers[i].Name, buyers[i].Id, buyers[i].Live, buyers[i].Debug, buyers[i].PublicKeyBase64)
	}
	fmt.Fprintf(w, "</table>\n")

	// sellers

	type SellerRow struct {
		Name string
		Id   string
	}

	sellers := []SellerRow{}

	for _, v := range database.SellerMap {

		row := SellerRow{
			Id:   fmt.Sprintf("%0x", v.Id),
			Name: v.Name,
		}

		sellers = append(sellers, row)
	}

	sort.SliceStable(sellers, func(i, j int) bool { return sellers[i].Id < sellers[j].Id })

	fmt.Fprintf(w, "<br><br>Sellers:<br><br>")
	fmt.Fprintf(w, "<table>\n")
	fmt.Fprintf(w, "<tr><td><b>%s</b></td><td><b>%s</b></td></tr>\n", "Name", "Id")
	for i := range sellers {
		fmt.Fprintf(w, "<tr><td>%s</td><td>%s</td></tr>\n", sellers[i].Name, sellers[i].Id)
	}
	fmt.Fprintf(w, "</table>\n")

	// datacenters

	type DatacenterRow struct {
		Name      string
		Id        string
		Latitude  string
		Longitude string
	}

	datacenters := []DatacenterRow{}

	for _, v := range database.DatacenterMap {

		row := DatacenterRow{
			Id:        fmt.Sprintf("%0x", v.Id),
			Name:      v.Name,
			Latitude:  fmt.Sprintf("%+3.2f", v.Latitude),
			Longitude: fmt.Sprintf("%+3.2f", v.Longitude),
		}

		datacenters = append(datacenters, row)
	}

	sort.SliceStable(datacenters, func(i, j int) bool { return datacenters[i].Name < datacenters[j].Name })

	fmt.Fprintf(w, "<br><br>Datacenters:<br><br>")
	fmt.Fprintf(w, "<table>\n")
	fmt.Fprintf(w, "<tr><td><b>%s</b></td><td><b>%s</b></td><td><b>%s</b></td><td><b>%s</b></td></tr>\n", "Name", "Id", "Latitude", "Longitude")
	for i := range datacenters {
		fmt.Fprintf(w, "<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>\n", datacenters[i].Name, datacenters[i].Id, datacenters[i].Latitude, datacenters[i].Longitude)
	}
	fmt.Fprintf(w, "</table>\n")

	// relays

	type RelayRow struct {
		Name            string
		Id              string
		PublicAddress   string
		InternalAddress string
		PublicKey       string
		PrivateKey      string
	}

	relays := []RelayRow{}

	for _, v := range database.RelayMap {

		row := RelayRow{
			Id:            fmt.Sprintf("%0x", v.Id),
			Name:          v.Name,
			PublicAddress: v.PublicAddress.String(),
			PublicKey:     base64.StdEncoding.EncodeToString(v.PublicKey),
			PrivateKey:    base64.StdEncoding.EncodeToString(v.PrivateKey),
		}

		if v.HasInternalAddress {
			row.InternalAddress = v.InternalAddress.String()
		}

		relays = append(relays, row)
	}

	sort.SliceStable(relays, func(i, j int) bool { return relays[i].Name < relays[j].Name })

	fmt.Fprintf(w, "<br><br>Relays:<br><br>")
	fmt.Fprintf(w, "<table>\n")
	fmt.Fprintf(w, "<tr><td><b>%s</b></td><td><b>%s</b></td><td><b>%s</b></td><td><b>%s</b></td><td><b>%s</b></td><td><b>%s</b></td></tr>\n", "Id", "Name", "Public Address", "Internal Address", "Public Key", "Private Key")
	for i := range relays {
		fmt.Fprintf(w, "<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>\n", relays[i].Id, relays[i].Name, relays[i].PublicAddress, relays[i].InternalAddress, relays[i].PublicKey, relays[i].PrivateKey)
	}
	fmt.Fprintf(w, "</table>\n")

	// route shaders

	for _, v := range database.BuyerMap {

		routeShader := v.RouteShader

		fmt.Fprintf(w, "<br><br>Route shader for '%s':<br><br>", v.Name)
		fmt.Fprintf(w, "<table>\n")
		fmt.Fprintf(w, "<tr><td><b>%s</b></td><td><b>%v</b></td>\n", "Property", "Value")
		fmt.Fprintf(w, "<tr><td>%s</td><td>%v</td>\n", "Disable Network Next", routeShader.DisableNetworkNext)
		fmt.Fprintf(w, "<tr><td>%s</td><td>%v</td>\n", "Analysis Only", routeShader.AnalysisOnly)
		fmt.Fprintf(w, "<tr><td>%s</td><td>%v</td>\n", "AB Test", routeShader.ABTest)
		fmt.Fprintf(w, "<tr><td>%s</td><td>%v</td>\n", "Reduce Latency", routeShader.ReduceLatency)
		fmt.Fprintf(w, "<tr><td>%s</td><td>%v</td>\n", "Reduce Packet Loss", routeShader.ReducePacketLoss)
		fmt.Fprintf(w, "<tr><td>%s</td><td>%v</td>\n", "Multipath", routeShader.Multipath)
		fmt.Fprintf(w, "<tr><td>%s</td><td>%v</td>\n", "Force Next", routeShader.ForceNext)
		fmt.Fprintf(w, "<tr><td>%s</td><td>%d%%</td>\n", "Selection Percent", routeShader.SelectionPercent)
		fmt.Fprintf(w, "<tr><td>%s</td><td>%dms</td>\n", "Acceptable Latency", routeShader.AcceptableLatency)
		fmt.Fprintf(w, "<tr><td>%s</td><td>%dms</td>\n", "Latency Threshold", routeShader.LatencyThreshold)
		fmt.Fprintf(w, "<tr><td>%s</td><td>%.1f%%</td>\n", "Acceptable Packet Loss", routeShader.AcceptablePacketLoss)
		fmt.Fprintf(w, "<tr><td>%s</td><td>%.1f%%</td>\n", "Packet Loss Sustained", routeShader.PacketLossSustained)
		fmt.Fprintf(w, "<tr><td>%s</td><td>%dkbps</td>\n", "Bandwidth Envelope Up", routeShader.BandwidthEnvelopeUpKbps)
		fmt.Fprintf(w, "<tr><td>%s</td><td>%dkbps</td>\n", "Bandwidth Envelope Down", routeShader.BandwidthEnvelopeDownKbps)
		fmt.Fprintf(w, "<tr><td>%s</td><td>%dms</td>\n", "Route Select Threshold", routeShader.RouteSelectThreshold)
		fmt.Fprintf(w, "<tr><td>%s</td><td>%dms</td>\n", "Route Switch Threshold", routeShader.RouteSwitchThreshold)
		fmt.Fprintf(w, "<tr><td>%s</td><td>%dms</td>\n", "Max Latency Trade Off", routeShader.MaxLatencyTradeOff)
		fmt.Fprintf(w, "<tr><td>%s</td><td>%dms</td>\n", "RTT Veto (Default)", routeShader.RTTVeto_Default)
		fmt.Fprintf(w, "<tr><td>%s</td><td>%dms</td>\n", "RTT Veto (Multipath)", routeShader.RTTVeto_Multipath)
		fmt.Fprintf(w, "<tr><td>%s</td><td>%dms</td>\n", "RTT Veto (PacketLoss)", routeShader.RTTVeto_PacketLoss)
		fmt.Fprintf(w, "<tr><td>%s</td><td>%dms</td>\n", "Max Next RTT", routeShader.MaxNextRTT)
		fmt.Fprintf(w, "<tr><td>%s</td><td>%d</td>\n", "Route Diversity", routeShader.RouteDiversity)
		fmt.Fprintf(w, "</table>\n")
	}

	// destination datacenters

	type DestinationDatacenterRow struct {
		Datacenter string
		Buyers     []string
	}

	destinationDatacenterMap := make(map[uint64]*DestinationDatacenterRow)

	for _, v1 := range database.DatacenterMaps {
		for _, v2 := range v1 {
			if !v2.EnableAcceleration {
				continue
			}
			buyerId := v2.BuyerId
			datacenterId := v2.DatacenterId
			entry := destinationDatacenterMap[datacenterId]
			if entry == nil {
				entry = &DestinationDatacenterRow{}
				entry.Datacenter = database.DatacenterMap[datacenterId].Name
				destinationDatacenterMap[datacenterId] = entry
			}
			entry.Buyers = append(entry.Buyers, database.BuyerMap[buyerId].Name)
		}
	}

	destinationDatacenters := make([]DestinationDatacenterRow, 0)

	for _, v := range destinationDatacenterMap {
		sort.SliceStable(v.Buyers, func(i, j int) bool { return v.Buyers[i] < v.Buyers[j] })
		destinationDatacenters = append(destinationDatacenters, *v)
	}

	sort.SliceStable(destinationDatacenters, func(i, j int) bool {
		return destinationDatacenters[i].Datacenter < destinationDatacenters[j].Datacenter
	})

	fmt.Fprintf(w, "<br><br>Destination datacenters:<br><br>")
	fmt.Fprintf(w, "<table>\n")
	fmt.Fprintf(w, "<tr><td><b>%s</b></td><td><b>%s</b></td></tr>\n", "Datacenter", "Buyers")
	for i := range destinationDatacenters {
		fmt.Fprintf(w, "<tr><td>%s</td><td>%s</td></tr>\n", destinationDatacenters[i].Datacenter, destinationDatacenters[i].Buyers)
	}
	fmt.Fprintf(w, "</table>\n")

	const htmlFooter = `</body></html>`

	fmt.Fprintf(w, "%s\n", htmlFooter)
}

// ---------------------------------------------------------------------

type Overlay struct {
	CreationTime string
	BuyerMap     map[uint64]Buyer
}

func CreateOverlay() *Overlay {

	overlay := &Overlay{
		CreationTime: "",
		BuyerMap:     make(map[uint64]Buyer),
	}

	return overlay
}

func LoadOverlay(filename string) (*Overlay, error) {

	overlayFile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer overlayFile.Close()

	overlay := &Overlay{}

	err = gob.NewDecoder(overlayFile).Decode(overlay)

	return overlay, err
}

func (overlay *Overlay) Save(filename string) error {

	var buffer bytes.Buffer

	err := gob.NewEncoder(&buffer).Encode(overlay)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(filename, buffer.Bytes(), 0644); err != nil {
		return err
	}

	return nil
}

func (overlay Overlay) IsEmpty() bool {
	return len(overlay.BuyerMap) == 0
}

// ---------------------------------------------------------------------

func ExtractDatabase(config string) (*Database, error) {

	pgsql, err := sql.Open("postgres", config)
	if err != nil {
		return nil, fmt.Errorf("could not connect to postgres: %v\n", err)
	}

	err = pgsql.Ping()
	if err != nil {
		return nil, fmt.Errorf("could not ping postgres: %v\n", err)
	}

	fmt.Printf("successfully connected to postgres\n")

	// relays

	type RelayRow struct {
		id                 uint64
		name               string
		datacenter         uint64
		public_ip          string
		public_port        int
		internal_ip        string
		internal_port      int
		internal_group     string
		ssh_ip             string
		ssh_port           int
		ssh_user           string
		public_key_base64  string
		private_key_base64 string
		version            sql.NullString
		mrc                int
		port_speed         int
		max_sessions       int
	}

	relayRows := make([]RelayRow, 0)
	{
		rows, err := pgsql.Query("SELECT id, display_name, datacenter, public_ip, public_port, internal_ip, internal_port, internal_group, ssh_ip, ssh_port, ssh_user, public_key_base64, private_key_base64, version, mrc, port_speed, max_sessions FROM relays")
		if err != nil {
			return nil, fmt.Errorf("could not extract relays: %v\n", err)
		}

		defer rows.Close()

		for rows.Next() {
			row := RelayRow{}
			if err := rows.Scan(&row.id, &row.name, &row.datacenter, &row.public_ip, &row.public_port, &row.internal_ip, &row.internal_port, &row.internal_group, &row.ssh_ip, &row.ssh_port, &row.ssh_user, &row.public_key_base64, &row.private_key_base64, &row.version, &row.mrc, &row.port_speed, &row.max_sessions); err != nil {
				return nil, fmt.Errorf("failed to scan relay row: %v\n", err)
			}
			relayRows = append(relayRows, row)
		}
	}

	// datacenters

	type DatacenterRow struct {
		id        uint64
		name      string
		latitude  float32
		longitude float32
		seller_id uint64
	}

	datacenterRows := make([]DatacenterRow, 0)
	{
		rows, err := pgsql.Query("SELECT id, display_name, latitude, longitude, seller_id FROM datacenters")
		if err != nil {
			return nil, fmt.Errorf("could not extract datacenters: %v\n", err)
		}

		defer rows.Close()

		for rows.Next() {
			row := DatacenterRow{}
			if err := rows.Scan(&row.id, &row.name, &row.latitude, &row.longitude, &row.seller_id); err != nil {
				return nil, fmt.Errorf("failed to scan datacenter row: %v\n", err)
			}
			datacenterRows = append(datacenterRows, row)
		}
	}

	// buyers

	type BuyerRow struct {
		id                uint64
		name              string
		public_key_base64 string
		customer_id       uint64
		route_shader_id   uint64
	}

	buyerRows := make([]BuyerRow, 0)
	{
		rows, err := pgsql.Query("SELECT id, short_name, public_key_base64, customer_id, route_shader_id FROM buyers")
		if err != nil {
			return nil, fmt.Errorf("could not extract buyers: %v\n", err)
		}

		defer rows.Close()

		for rows.Next() {
			row := BuyerRow{}
			if err := rows.Scan(&row.id, &row.name, &row.public_key_base64, &row.customer_id, &row.route_shader_id); err != nil {
				return nil, fmt.Errorf("failed to scan buyer row: %v\n", err)
			}
			buyerRows = append(buyerRows, row)
		}
	}

	// sellers

	type SellerRow struct {
		id          uint64
		name        string
		customer_id sql.NullInt64
	}

	sellerRows := make([]SellerRow, 0)
	{
		rows, err := pgsql.Query("SELECT id, short_name, customer_id FROM sellers")
		if err != nil {
			return nil, fmt.Errorf("could not extract sellers: %v\n", err)
		}

		defer rows.Close()

		for rows.Next() {
			row := SellerRow{}
			if err := rows.Scan(&row.id, &row.name, &row.customer_id); err != nil {
				return nil, fmt.Errorf("failed to scan seller row: %v\n", err)
			}
			sellerRows = append(sellerRows, row)
		}
	}

	// customers

	type CustomerRow struct {
		id            uint64
		customer_name string
		customer_code string
		live          bool
		debug         bool
	}

	customerRows := make([]CustomerRow, 0)
	{
		rows, err := pgsql.Query("SELECT id, customer_name, customer_code, live, debug FROM customers")
		if err != nil {
			return nil, fmt.Errorf("could not extract customers: %v\n", err)
		}

		defer rows.Close()

		for rows.Next() {
			row := CustomerRow{}
			if err := rows.Scan(&row.id, &row.customer_name, &row.customer_code, &row.live, &row.debug); err != nil {
				return nil, fmt.Errorf("failed to scan customer row: %v\n", err)
			}
			customerRows = append(customerRows, row)
		}
	}

	// route shaders

	type RouteShaderRow struct {
		id                           uint64
		ab_test                      bool
		acceptable_latency           int
		acceptable_packet_loss       float32
		packet_loss_sustained        float32
		analysis_only                bool
		bandwidth_envelope_down_kbps int
		bandwidth_envelope_up_kbps   int
		disable_network_next         bool
		latency_threshold            int
		multipath                    bool
		reduce_latency               bool
		reduce_packet_loss           bool
		selection_percent            int
		max_latency_tradeoff         int
		max_next_rtt                 int
		route_switch_threshold       int
		route_select_threshold       int
		rtt_veto_default             int
		rtt_veto_multipath           int
		rtt_veto_packetloss          int
		force_next                   bool
		route_diversity              int
	}

	routeShaderRows := make([]RouteShaderRow, 0)
	{
		rows, err := pgsql.Query("SELECT id, ab_test, acceptable_latency, acceptable_packet_loss, packet_loss_sustained, analysis_only, bandwidth_envelope_down_kbps, bandwidth_envelope_up_kbps, disable_network_next, latency_threshold, multipath, reduce_latency, reduce_packet_loss, selection_percent, max_latency_tradeoff, max_next_rtt, route_switch_threshold, route_select_threshold, rtt_veto_default, rtt_veto_multipath, rtt_veto_packetloss, force_next, route_diversity FROM route_shaders")
		if err != nil {
			return nil, fmt.Errorf("could not extract route shaders: %v\n", err)
		}

		defer rows.Close()

		for rows.Next() {
			row := RouteShaderRow{}
			if err := rows.Scan(&row.id, &row.ab_test, &row.acceptable_latency, &row.acceptable_packet_loss, &row.packet_loss_sustained, &row.analysis_only, &row.bandwidth_envelope_down_kbps, &row.bandwidth_envelope_up_kbps, &row.disable_network_next, &row.latency_threshold, &row.multipath, &row.reduce_latency, &row.reduce_packet_loss, &row.selection_percent, &row.max_latency_tradeoff, &row.max_next_rtt, &row.route_switch_threshold, &row.route_select_threshold, &row.rtt_veto_default, &row.rtt_veto_multipath, &row.rtt_veto_packetloss, &row.force_next, &row.route_diversity); err != nil {
				return nil, fmt.Errorf("failed to scan route shader row: %v\n", err)
			}
			routeShaderRows = append(routeShaderRows, row)
		}
	}

	// datacenter maps

	type DatacenterMapRow struct {
		buyer_id            uint64
		datacenter_id       uint64
		enable_acceleration bool
	}

	datacenterMapRows := make([]DatacenterMapRow, 0)
	{
		rows, err := pgsql.Query("SELECT buyer_id, datacenter_id, enable_acceleration FROM datacenter_maps")
		if err != nil {
			return nil, fmt.Errorf("could not extract datacenter maps: %v\n", err)
		}

		defer rows.Close()

		for rows.Next() {
			row := DatacenterMapRow{}
			if err := rows.Scan(&row.buyer_id, &row.datacenter_id, &row.enable_acceleration); err != nil {
				return nil, fmt.Errorf("failed to scan datacenter map row: %v\n", err)
			}
			datacenterMapRows = append(datacenterMapRows, row)
		}
	}

	// print out rows

	fmt.Printf("\nrelays:\n")
	for _, row := range relayRows {
		fmt.Printf("%d: %s, %d, %s, %d, %s, %d, %s, %d, %s, %s, %s, %s, %d, %d, %d\n", row.id, row.name, row.datacenter, row.public_ip, row.public_port, row.internal_ip, row.internal_port, row.ssh_ip, row.ssh_port, row.ssh_user, row.public_key_base64, row.private_key_base64, row.version.String, row.mrc, row.port_speed, row.max_sessions)
	}

	fmt.Printf("\ndatacenters:\n")
	for _, row := range datacenterRows {
		fmt.Printf("%d: %s, %.1f, %.1f, %d\n", row.id, row.name, row.latitude, row.longitude, row.seller_id)
	}

	fmt.Printf("\nbuyers:\n")
	for _, row := range buyerRows {
		fmt.Printf("%d: %s, %s, %d\n", row.id, row.name, row.public_key_base64, row.customer_id)
	}

	fmt.Printf("\nsellers:\n")
	for _, row := range sellerRows {
		fmt.Printf("%d: %s, %d\n", row.id, row.name, row.customer_id.Int64)
	}

	fmt.Printf("\ncustomers:\n")
	for _, row := range customerRows {
		fmt.Printf("%d: %s, %s, %v, %v\n", row.id, row.customer_name, row.customer_code, row.live, row.debug)
	}

	fmt.Printf("\nroute shaders:\n")
	for _, row := range routeShaderRows {
		fmt.Printf("%d: %v, %d, %.1f, %v, %d, %d, %v, %d, %v, %v, %v, %d, %d, %d, %d, %d, %d, %d, %d, %v\n", row.id, row.ab_test, row.acceptable_latency, row.acceptable_packet_loss, row.analysis_only, row.bandwidth_envelope_down_kbps, row.bandwidth_envelope_up_kbps, row.disable_network_next, row.latency_threshold, row.multipath, row.reduce_latency, row.reduce_packet_loss, row.selection_percent, row.max_latency_tradeoff, row.max_next_rtt, row.route_switch_threshold, row.route_select_threshold, row.rtt_veto_default, row.rtt_veto_multipath, row.rtt_veto_packetloss, row.force_next)
	}

	fmt.Printf("\ndatacenter maps:\n")
	for _, row := range datacenterMapRows {
		fmt.Printf("(%d,%d): %v\n", row.buyer_id, row.datacenter_id, row.enable_acceleration)
	}

	// index datacenters by postgres id

	datacenterIndex := make(map[uint64]DatacenterRow)
	for _, row := range datacenterRows {
		datacenterIndex[row.id] = row
	}

	// index customers by postgres id

	customerIndex := make(map[uint64]CustomerRow)
	for _, row := range customerRows {
		customerIndex[row.id] = row
	}

	// index buyers by postgres id

	buyerIndex := make(map[uint64]BuyerRow)
	for _, row := range buyerRows {
		buyerIndex[row.id] = row
	}

	// index sellers by postgres id

	sellerIndex := make(map[uint64]SellerRow)
	for _, row := range sellerRows {
		sellerIndex[row.id] = row
	}

	// index route shaders by postgres id

	routeShaderIndex := make(map[uint64]RouteShaderRow)
	for _, row := range routeShaderRows {
		routeShaderIndex[row.id] = row
	}

	// build database

	fmt.Printf("\nbuilding network next database...\n\n")

	database := CreateDatabase()

	database.CreationTime = time.Now().Format("Monday 02 January 2006 15:04:05 MST")
	database.Creator = "extract_database"

	database.Relays = make([]Relay, len(relayRows))

	for i, row := range sellerRows {

		seller := Seller{}

		seller.Id = row.id
		seller.Name = row.name

		database.SellerMap[seller.Id] = &seller

		fmt.Printf("seller %d: %s [%d]\n", i, seller.Name, seller.Id)
	}

	for i, row := range buyerRows {

		buyer := Buyer{}

		buyer.Name = row.name

		data, err := base64.StdEncoding.DecodeString(row.public_key_base64)
		if err != nil {
			return nil, fmt.Errorf("could not decode public key base64 for buyer %s: %v\n", buyer.Name, err)
		}

		if len(data) != 40 {
			return nil, fmt.Errorf("buyer public key data must be 40 bytes\n")
		}

		buyer.Id = binary.LittleEndian.Uint64(data[:8])
		buyer.PublicKey = data[8:40]

		customer_row, customer_exists := customerIndex[row.customer_id]
		if !customer_exists {
			return nil, fmt.Errorf("buyer %s does not have a customer?!\n", buyer.Name)
		}

		buyer.Live = customer_row.live
		buyer.Debug = customer_row.debug

		route_shader_row, route_shader_exists := routeShaderIndex[row.route_shader_id]
		if !route_shader_exists {
			return nil, fmt.Errorf("buyer %s does not have a route shader?!\n", buyer.Name)
		}

		buyer.RouteShader.DisableNetworkNext = route_shader_row.disable_network_next
		buyer.RouteShader.AnalysisOnly = route_shader_row.analysis_only
		buyer.RouteShader.SelectionPercent = route_shader_row.selection_percent
		buyer.RouteShader.ABTest = route_shader_row.ab_test
		buyer.RouteShader.ReduceLatency = route_shader_row.reduce_latency
		buyer.RouteShader.ReducePacketLoss = route_shader_row.reduce_packet_loss
		buyer.RouteShader.Multipath = route_shader_row.multipath
		buyer.RouteShader.AcceptableLatency = int32(route_shader_row.acceptable_latency)
		buyer.RouteShader.LatencyThreshold = int32(route_shader_row.latency_threshold)
		buyer.RouteShader.AcceptablePacketLoss = route_shader_row.acceptable_packet_loss
		buyer.RouteShader.PacketLossSustained = route_shader_row.packet_loss_sustained
		buyer.RouteShader.BandwidthEnvelopeUpKbps = int32(route_shader_row.bandwidth_envelope_up_kbps)
		buyer.RouteShader.BandwidthEnvelopeDownKbps = int32(route_shader_row.bandwidth_envelope_down_kbps)
		buyer.RouteShader.RouteSelectThreshold = int32(route_shader_row.route_select_threshold)
		buyer.RouteShader.RouteSwitchThreshold = int32(route_shader_row.route_switch_threshold)
		buyer.RouteShader.MaxLatencyTradeOff = int32(route_shader_row.max_latency_tradeoff)
		buyer.RouteShader.RTTVeto_Default = int32(route_shader_row.rtt_veto_default)
		buyer.RouteShader.RTTVeto_Multipath = int32(route_shader_row.rtt_veto_multipath)
		buyer.RouteShader.RTTVeto_PacketLoss = int32(route_shader_row.rtt_veto_packetloss)
		buyer.RouteShader.MaxNextRTT = int32(route_shader_row.max_next_rtt)
		buyer.RouteShader.ForceNext = route_shader_row.force_next
		buyer.RouteShader.RouteDiversity = int32(route_shader_row.route_diversity)

		database.BuyerMap[buyer.Id] = &buyer

		fmt.Printf("buyer %d: %s [%x] (live=%v, debug=%v)\n", i, buyer.Name, buyer.Id, buyer.Live, buyer.Debug)
	}

	for i, row := range datacenterRows {

		datacenter := Datacenter{}

		datacenter.Id = HashString(row.name)
		datacenter.Name = row.name
		datacenter.Latitude = row.latitude
		datacenter.Longitude = row.longitude

		seller_row, seller_exists := sellerIndex[row.seller_id]
		if !seller_exists {
			return nil, fmt.Errorf("datacenter %s doesn't have a seller?!\n", datacenter.Name)
		}

		if !strings.Contains(datacenter.Name, seller_row.name) {
			return nil, fmt.Errorf("datacenter '%s' does not contain the seller name '%s' as a substring. are you sure this datacenter has the right seller?\n", datacenter.Name, seller_row.name)
		}

		database.DatacenterMap[datacenter.Id] = &datacenter

		fmt.Printf("datacenter %d: %s [%x] (%.1f,%.1f)\n", i, datacenter.Name, datacenter.Id, datacenter.Latitude, datacenter.Longitude)
	}

	for i, row := range relayRows {

		relay := &database.Relays[i]

		relay.Name = row.name

		relay.PublicAddress = core.ParseAddress(row.public_ip)
		relay.PublicAddress.Port = row.public_port

		relay.Id = HashString(relay.PublicAddress.String())

		relay.InternalAddress = core.ParseAddress(row.internal_ip)
		relay.InternalAddress.Port = row.internal_port

		if relay.InternalAddress.String() != "0.0.0.0:0" {
			relay.HasInternalAddress = true
			if relay.InternalAddress.Port == 0 {
				relay.InternalAddress.Port = relay.PublicAddress.Port
			}
		}

		if row.internal_group == "" {
			relay.InternalGroup = 0
		} else {
			relay.InternalGroup = HashString(row.internal_group)
		}

		relay.SSHAddress = core.ParseAddress(row.ssh_ip)
		relay.SSHUser = row.ssh_user

		if relay.SSHAddress.String() == "0.0.0.0:0" {
			relay.SSHAddress = relay.PublicAddress
		}

		relay.PublicKey, err = base64.StdEncoding.DecodeString(row.public_key_base64)
		if err != nil {
			return nil, fmt.Errorf("could not decode public key base64 for relay %s: %v\n", relay.Name, err)
		}
		if len(relay.PublicKey) != 32 {
			return nil, fmt.Errorf("relay public key must be 32 bytes\n")
		}

		relay.PrivateKey, err = base64.StdEncoding.DecodeString(row.private_key_base64)
		if err != nil {
			return nil, fmt.Errorf("could not decode private key base64 for relay %s: %v\n", relay.Name, err)
		}
		if len(relay.PrivateKey) != 32 {
			return nil, fmt.Errorf("relay private key must be 32 bytes\n")
			os.Exit(1)
		}

		relay.MaxSessions = row.max_sessions
		relay.PortSpeed = row.port_speed
		relay.Version = row.version.String

		datacenter_row, datacenter_exists := datacenterIndex[row.datacenter]
		if !datacenter_exists {
			return nil, fmt.Errorf("relay %s doesn't have a datacenter?!\n", relay.Name)
		}

		relay.DatacenterId = HashString(datacenter_row.name)

		if !strings.Contains(relay.Name, datacenter_row.name) {
			return nil, fmt.Errorf("relay '%s' does not contain the datacenter name '%s' as a substring. are you sure this relay has the right datacenter?\n", relay.Name, datacenter_row.name)
		}

		relay.Datacenter = *database.DatacenterMap[relay.DatacenterId]
		if relay.Datacenter.Id != relay.DatacenterId {
			return nil, fmt.Errorf("relay '%s' has a bad datacenter?!\n", relay.Name)
		}

		seller_row, seller_exists := sellerIndex[datacenter_row.seller_id]
		if !seller_exists {
			return nil, fmt.Errorf("relay %s doesn't have a seller?!\n", relay.Name)
		}

		relay.Seller = *database.SellerMap[seller_row.id]

		fmt.Printf("relay %d: %s -> %s [%x]\n", i, relay.Name, datacenter_row.name, relay.DatacenterId)

		database.RelayMap[relay.Id] = relay
	}

	for i, row := range datacenterMapRows {

		buyer_row, buyer_exists := buyerIndex[row.buyer_id]
		if !buyer_exists {
			return nil, fmt.Errorf("datacenter map doesn't have a buyer?!\n")
		}

		datacenter_row, datacenter_exists := datacenterIndex[row.datacenter_id]
		if !datacenter_exists {
			return nil, fmt.Errorf("datacenter map doesn't have a datacenter?!\n")
		}

		buyerName := buyer_row.name
		buyerId := uint64(0)
		datacenterName := datacenter_row.name
		datacenterId := HashString(datacenterName)

		for _, v := range database.BuyerMap {
			if v.Name == buyerName {
				buyerId = v.Id
			}
		}

		if buyerId == 0 {
			return nil, fmt.Errorf("could not find runtime buyer id for buyer %s?!\n", buyerName)
		}

		fmt.Printf("datacenter map %d: %s [%x] -> %s [%x] enabled\n", i, buyerName, buyerId, datacenterName, datacenterId)

		datacenterMap := DatacenterMap{}
		datacenterMap.BuyerId = buyerId
		datacenterMap.DatacenterId = datacenterId
		datacenterMap.EnableAcceleration = row.enable_acceleration
		if database.DatacenterMaps[buyerId] == nil {
			database.DatacenterMaps[buyerId] = make(map[uint64]*DatacenterMap)
		}
		database.DatacenterMaps[buyerId][datacenterId] = &datacenterMap
	}

	return database, nil
}

// -----------------------------------------------------------------------------------------------------------

// -----------------------------------------------------------------------------------------------------------
