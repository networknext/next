package database

import (
	"bytes"
	"encoding/gob"
	"encoding/base64"
	"encoding/binary"
	"io/ioutil"
	"errors"
	"net"
	"os"
	"fmt"
	"sort"

	"github.com/networknext/backend/modules/core"

	"github.com/modood/table"
)

type Relay struct {
	Id                 uint64
	Name               string
	DatacenterId       uint64
	PublicAddress      net.UDPAddr
	HasInternalAddress bool
	InternalAddress    net.UDPAddr
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
	RelayMap       map[uint64]Relay
	BuyerMap       map[uint64]Buyer
	SellerMap      map[uint64]Seller
	DatacenterMap  map[uint64]Datacenter
	DatacenterMaps map[uint64]map[uint64]DatacenterMap
	//                   ^ BuyerId  ^ DatacenterId
}

func CreateDatabase() *Database {

	database := &Database{
		CreationTime:   "",
		Creator:        "",
		Relays:         []Relay{},
		RelayMap:       make(map[uint64]Relay),
		BuyerMap:       make(map[uint64]Buyer),
		SellerMap:      make(map[uint64]Seller),
		DatacenterMap:  make(map[uint64]Datacenter),
		DatacenterMaps: make(map[uint64]map[uint64]DatacenterMap),
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
		return errors.New("database is empty")
	}

	// todo: each relay must have a valid datacenter

	// todo: each datacenter must have valid seller

	// todo: buyer must have a valid non-zero buyer id

	// todo: if a relay has an internal ip address, it must not be "0.0.0.0"

	// todo: relay ssh address must not be 0.0.0.0 -- it should be set to the 

	// todo: relay internal address, external address *and* ssh address ports must not be zero

	// todo: if a relay has both public and private keypair specified, decrypt/encrypt something with them, to make sure the keypair is valid -- catch errors early

	// todo: each relay must have a unique name

	// todo: each relay must have a unique public ip:port

	// todo: datacenter maps must have valid datacenter id and buyer ids (would have caught an error I just found...)

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

	output += "\n\nBuyers:\n\n"

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
			Id:    fmt.Sprintf("%0x", v.Id),
			Name:  v.Name,
			Live:  fmt.Sprintf("%v", v.Live),
			Debug: fmt.Sprintf("%v", v.Debug),
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
		Buyers []string
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

	sort.SliceStable(destinationDatacenters, func(i, j int) bool { return destinationDatacenters[i].Datacenter < destinationDatacenters[j].Datacenter })

	output += table.Table(destinationDatacenters)

	return output
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
