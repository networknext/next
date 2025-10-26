package database

import (
	"bytes"
	"compress/gzip"
	"database/sql"
	"encoding/base64"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"hash/fnv"
	"io"
	"net"
	"os"
	"sort"
	"time"

	"github.com/networknext/next/modules/core"
	"github.com/networknext/next/modules/crypto"

	_ "github.com/lib/pq"
	"github.com/modood/table"
)

func HashString(s string) uint64 {
	hash := fnv.New64a()
	hash.Write([]byte(s))
	return hash.Sum64()
}

type Relay struct {
	Id                 uint64      `json:"id,string"`
	Name               string      `json:"name"`
	DatacenterId       uint64      `json:"datacenter_id,string"`
	PublicAddress      net.UDPAddr `json:"public_address,string"`
	HasInternalAddress bool        `json:"has_internal_address"`
	InternalAddress    net.UDPAddr `json:"internal_address,string"`
	InternalGroup      uint64      `json:"internal_group,string"`
	SSHAddress         net.UDPAddr `json:"ssh_address,string"`
	SSHUser            string      `json:"ssh_user,string"`
	PublicKey          []byte      `json:"public_key"`
	PrivateKey         []byte      `json:"private_key"`
	MaxSessions        int         `json:"max_sessions"`
	PortSpeed          int         `json:"port_speed"`
	BandwidthPrice     int         `json:"bandwidth_price"`
	Version            string      `json:"version"`
	Seller             *Seller     `json:"seller"`
	Datacenter         *Datacenter `json:"datacenter"`
}

type Buyer struct {
	Id          uint64           `json:"id,string"`
	Name        string           `json:"name"`
	Code        string           `json:"code"`
	Live        bool             `json:"live"`
	Debug       bool             `json:"debug"`
	PublicKey   []byte           `json:"public_key"`
	RouteShader core.RouteShader `json:"route_shader"`
}

type Seller struct {
	Id   uint64 `json:"id,string"`
	Name string `json:"name"`
	Code string `json:"code"`
}

type Datacenter struct {
	Id        uint64  `json:"id,string"`
	Name      string  `json:"name"`
	Native    string  `json:"native"`
	Latitude  float32 `json:"latitude"`
	Longitude float32 `json:"longitude"`
	SellerId  uint64  `json:"seller_id,string"`
}

type BuyerDatacenterSettings struct {
	BuyerId            uint64 `json:"buyer_id,string"`
	DatacenterId       uint64 `json:"datacenter_id,string"`
	EnableAcceleration bool   `json:"enable_acceleration"`
}

type Database struct {
	CreationTime            string                                         `json:"creation_time"`
	Creator                 string                                         `json:"creator"`
	Relays                  []Relay                                        `json:"relays"`
	BuyerMap                map[uint64]*Buyer                              `json:"buyer_map"`
	SellerMap               map[uint64]*Seller                             `json:"seller_map"`
	DatacenterMap           map[uint64]*Datacenter                         `json:"datacenter"`
	DatacenterRelays        map[uint64][]uint64                            `json:"datacenter_relays"`
	BuyerDatacenterSettings map[uint64]map[uint64]*BuyerDatacenterSettings `json:"buyer_datacenter_settings"`
	RelayMap                map[uint64]*Relay
	RelayNameMap            map[string]*Relay
	BuyerCodeMap            map[string]*Buyer
	SellerCodeMap           map[string]*Seller
	DatacenterNameMap       map[string]*Datacenter
	RelaySecretKeys         map[uint64][]byte
}

func CreateDatabase() *Database {

	database := &Database{
		CreationTime:            "",
		Creator:                 "",
		Relays:                  []Relay{},
		BuyerMap:                make(map[uint64]*Buyer),
		SellerMap:               make(map[uint64]*Seller),
		DatacenterMap:           make(map[uint64]*Datacenter),
		DatacenterRelays:        make(map[uint64][]uint64),
		BuyerDatacenterSettings: make(map[uint64]map[uint64]*BuyerDatacenterSettings),
		RelayMap:                make(map[uint64]*Relay),
		RelayNameMap:            make(map[string]*Relay),
		BuyerCodeMap:            make(map[string]*Buyer),
		SellerCodeMap:           make(map[string]*Seller),
		DatacenterNameMap:       make(map[string]*Datacenter),
		RelaySecretKeys:         make(map[uint64][]byte),
	}

	return database
}

func LoadDatabase(filename string) (*Database, error) {

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	database := &Database{}

	return database, database.LoadBinary(data)
}

func (database *Database) Fixup() {

	if len(database.RelayMap) != len(database.Relays) {
		database.RelayMap = make(map[uint64]*Relay, len(database.Relays))
		for i := range database.Relays {
			database.RelayMap[database.Relays[i].Id] = &database.Relays[i]
		}
	}

	if len(database.RelayNameMap) != len(database.Relays) {
		database.RelayNameMap = make(map[string]*Relay, len(database.Relays))
		for i := range database.Relays {
			database.RelayNameMap[database.Relays[i].Name] = &database.Relays[i]
		}
	}

	if len(database.BuyerCodeMap) != len(database.BuyerMap) {
		database.BuyerCodeMap = make(map[string]*Buyer, len(database.BuyerMap))
		for _, v := range database.BuyerMap {
			database.BuyerCodeMap[v.Code] = v
		}
	}

	if len(database.SellerCodeMap) != len(database.SellerMap) {
		database.SellerCodeMap = make(map[string]*Seller, len(database.SellerMap))
		for _, v := range database.SellerMap {
			database.SellerCodeMap[v.Code] = v
		}
	}

	if len(database.DatacenterNameMap) != len(database.DatacenterMap) {
		database.DatacenterNameMap = make(map[string]*Datacenter, len(database.DatacenterMap))
		for _, v := range database.DatacenterMap {
			database.DatacenterNameMap[v.Name] = v
		}
	}

	if len(database.RelaySecretKeys) == 0 {
		database.RelaySecretKeys = make(map[uint64][]byte)
	}
}

func (database *Database) GenerateRelaySecretKeys(relayBackendPublicKey []byte, relayBackendPrivateKey []byte) int {
	numWarnings := 0
	for i := range database.Relays {
		relay := &database.Relays[i]
		var err error
		database.RelaySecretKeys[relay.Id], err = crypto.SecretKey_GenerateRemote(relayBackendPublicKey, relayBackendPrivateKey, relay.PublicKey)
		if err != nil {
			core.Warn("failed to generate secret key for relay %s", relay.Name)
			numWarnings++
		}
	}
	return numWarnings
}

func (database *Database) Save(filename string) error {

	var buffer bytes.Buffer

	err := gob.NewEncoder(&buffer).Encode(database)
	if err != nil {
		return err
	}

	var compressed_buffer bytes.Buffer
	gz, err := gzip.NewWriterLevel(&compressed_buffer, gzip.BestCompression)
	if err != nil {
		return err
	}

	if _, err := gz.Write(buffer.Bytes()); err != nil {
		return err
	}

	if err := gz.Close(); err != nil {
		return err
	}

	if err := os.WriteFile(filename, compressed_buffer.Bytes(), 0644); err != nil {
		return err
	}

	return nil
}

func (database *Database) Validate() error {

	if database.IsEmpty() {
		return fmt.Errorf("database is empty")
	}

	if len(database.Relays) != len(database.RelayMap) {
		return fmt.Errorf("mismatch between number of relays in relay list vs. relay map (%d vs. %d)", len(database.Relays), len(database.RelayMap))
	}

	if len(database.Relays) != len(database.RelayNameMap) {
		return fmt.Errorf("mismatch between number of relays in relay list vs. relay name map (%d vs. %d)", len(database.Relays), len(database.RelayNameMap))
	}

	if len(database.BuyerMap) != len(database.BuyerCodeMap) {
		return fmt.Errorf("mismatch between number of buyers in buyer map vs. buyer code map (%d vs. %d)", len(database.BuyerMap), len(database.BuyerCodeMap))
	}

	if len(database.SellerMap) != len(database.SellerCodeMap) {
		return fmt.Errorf("mismatch between number of sellers in seller map vs. seller code map (%d vs. %d)", len(database.SellerMap), len(database.SellerCodeMap))
	}

	for k, v := range database.RelayNameMap {
		if v.Name != k {
			return fmt.Errorf("relay %s has wrong key %s in relay name map", v.Name, k)
		}
	}

	for id, datacenter := range database.DatacenterMap {
		if id != datacenter.Id {
			return fmt.Errorf("datacenter %s has mismatched id (%d vs. %d)", datacenter.Name, datacenter.Id, id)
		}
	}

	for k, v := range database.DatacenterNameMap {
		if v.Name != k {
			return fmt.Errorf("datacenter %s has wrong key %s in datacenter name map", v.Name, k)
		}
	}

	for id, seller := range database.SellerMap {
		if id != seller.Id {
			return fmt.Errorf("seller %s has mismatched id (%d vs. %d)", seller.Name, seller.Id, id)
		}
	}

	for code, seller := range database.SellerCodeMap {
		if code != seller.Code {
			return fmt.Errorf("seller %s has mismatched code ('%s' vs. '%s')", seller.Name, seller.Code, code)
		}
	}

	for id, buyer := range database.BuyerMap {
		if id != buyer.Id {
			return fmt.Errorf("buyer %s has mismatched id (%d vs. %d)", buyer.Name, buyer.Id, id)
		}
	}

	for code, buyer := range database.BuyerCodeMap {
		if code != buyer.Code {
			return fmt.Errorf("buyer %s has mismatched code ('%s' vs. '%s')", buyer.Name, buyer.Code, code)
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
			return fmt.Errorf("relay %s ssh address port is zero: '%s'", relay.Name, relay.SSHAddress.String())
		}
	}

	for i := range database.Relays {
		relayId := database.Relays[i].Id
		datacenterId := database.Relays[i].Datacenter.Id
		datacenterRelays := database.GetDatacenterRelays(datacenterId)
		found := false
		for j := range datacenterRelays {
			if datacenterRelays[j] == relayId {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("datacenter relays map is invalid")
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

	for buyerId, buyerMap := range database.BuyerDatacenterSettings {
		for datacenterId, settings := range buyerMap {
			if settings.DatacenterId != datacenterId {
				return fmt.Errorf("bad datacenter id in buyer datacenter settings: %d vs %d", datacenterId, settings.DatacenterId)
			}
			if settings.BuyerId != buyerId {
				return fmt.Errorf("bad buyer id in buyer datacenter settings: %d vs %d", buyerId, settings.BuyerId)
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

	if len(database.BuyerDatacenterSettings) != 0 {
		return false
	}

	return true
}

func (database *Database) DatacenterExists(datacenterId uint64) bool {
	return database.DatacenterMap[datacenterId] != nil
}

func (database *Database) DatacenterEnabled(buyerId uint64, datacenterId uint64) bool {
	buyerEntry := database.BuyerDatacenterSettings[buyerId]
	if buyerEntry == nil {
		return false
	}
	settings := buyerEntry[datacenterId]
	if settings == nil {
		return false
	}
	return settings.EnableAcceleration
}

func (database *Database) GetRelay(relayId uint64) *Relay {
	return database.RelayMap[relayId]
}

func (database *Database) GetRelayByName(relayName string) *Relay {
	return database.RelayNameMap[relayName]
}

func (database *Database) GetRelayIds() []uint64 {
	relayIds := make([]uint64, len(database.RelayMap))
	index := 0
	for k := range database.RelayMap {
		relayIds[index] = k
		index++
	}
	return relayIds
}

func (database *Database) GetBuyer(buyerId uint64) *Buyer {
	return database.BuyerMap[buyerId]
}

func (database *Database) GetBuyerByCode(buyerCode string) *Buyer {
	return database.BuyerCodeMap[buyerCode]
}

func (database *Database) GetBuyerIds() []uint64 {
	buyerIds := make([]uint64, len(database.BuyerMap))
	index := 0
	for k := range database.BuyerMap {
		buyerIds[index] = k
		index++
	}
	return buyerIds
}

func (database *Database) GetSeller(sellerId uint64) *Seller {
	return database.SellerMap[sellerId]
}

func (database *Database) GetSellerByCode(sellerCode string) *Seller {
	return database.SellerCodeMap[sellerCode]
}

func (database *Database) GetDatacenter(datacenterId uint64) *Datacenter {
	return database.DatacenterMap[datacenterId]
}

func (database *Database) GetDatacenterByName(datacenterName string) *Datacenter {
	return database.DatacenterNameMap[datacenterName]
}

func (database *Database) GetDatacenterRelays(datacenterId uint64) []uint64 {
	return database.DatacenterRelays[datacenterId]
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
			Id:              fmt.Sprintf("%016x", v.Id),
			Name:            v.Name,
			Live:            fmt.Sprintf("%v", v.Live),
			Debug:           fmt.Sprintf("%v", v.Debug),
			PublicKeyBase64: base64.StdEncoding.EncodeToString(data),
		}

		buyers = append(buyers, row)
	}

	sort.SliceStable(buyers, func(i, j int) bool { return buyers[i].Name < buyers[j].Name })

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
			Id:   fmt.Sprintf("%d", v.Id),
			Name: v.Name,
		}

		sellers = append(sellers, row)
	}

	sort.SliceStable(sellers, func(i, j int) bool { return sellers[i].Id < sellers[j].Id })

	output += table.Table(sellers)

	// datacenters

	output += "\n\nDatacenters:\n\n"

	type DatacenterRow struct {
		Id        string
		Name      string
		Native    string
		Seller    string
		Latitude  string
		Longitude string
	}

	datacenters := []DatacenterRow{}

	for _, v := range database.DatacenterMap {

		row := DatacenterRow{
			Id:        fmt.Sprintf("%016x", v.Id),
			Name:      v.Name,
			Native:    v.Native,
			Latitude:  fmt.Sprintf("%+3.2f", v.Latitude),
			Longitude: fmt.Sprintf("%+3.2f", v.Longitude),
			Seller:    fmt.Sprintf("%d", v.SellerId),
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
		InternalGroup   string
		BandwidthPrice  string
		Version         string
		PublicKey       string
		PrivateKey      string
	}

	relays := []RelayRow{}

	for _, v := range database.RelayMap {

		row := RelayRow{
			Id:             fmt.Sprintf("%016x", v.Id),
			Name:           v.Name,
			Version:        v.Version,
			PublicAddress:  v.PublicAddress.String(),
			BandwidthPrice: fmt.Sprintf("%d", v.BandwidthPrice),
			PublicKey:      base64.StdEncoding.EncodeToString(v.PublicKey),
			PrivateKey:     base64.StdEncoding.EncodeToString(v.PrivateKey),
		}

		if v.HasInternalAddress {
			row.InternalAddress = v.InternalAddress.String()
			if v.InternalGroup != 0 {
				row.InternalGroup = fmt.Sprintf("%016x", v.InternalGroup)
			}
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
		properties = append(properties, PropertyRow{"AB Test", fmt.Sprintf("%v", routeShader.ABTest)})
		properties = append(properties, PropertyRow{"Force Next", fmt.Sprintf("%v", routeShader.ForceNext)})
		properties = append(properties, PropertyRow{"Selection Percent", fmt.Sprintf("%d%%", routeShader.SelectionPercent)})
		properties = append(properties, PropertyRow{"Acceptable Latency", fmt.Sprintf("%dms", routeShader.AcceptableLatency)})
		properties = append(properties, PropertyRow{"Latency Threshold", fmt.Sprintf("%dms", routeShader.LatencyReductionThreshold)})
		properties = append(properties, PropertyRow{"Acceptable Packet Loss", fmt.Sprintf("%.1f%%", routeShader.AcceptablePacketLoss)})
		properties = append(properties, PropertyRow{"Bandwidth Envelope Up", fmt.Sprintf("%dkbps", routeShader.BandwidthEnvelopeUpKbps)})
		properties = append(properties, PropertyRow{"Bandwidth Envelope Down", fmt.Sprintf("%dkbps", routeShader.BandwidthEnvelopeDownKbps)})
		properties = append(properties, PropertyRow{"Route Select Threshold", fmt.Sprintf("%dms", routeShader.RouteSelectThreshold)})
		properties = append(properties, PropertyRow{"Route Switch Threshold", fmt.Sprintf("%dms", routeShader.RouteSwitchThreshold)})
		properties = append(properties, PropertyRow{"Max Latency Trade Off", fmt.Sprintf("%dms", routeShader.MaxLatencyTradeOff)})

		output += table.Table(properties)
	}

	// destination datacenters

	output += "\n\nDestination datacenters:\n\n"

	type DestinationDatacenterRow struct {
		Datacenter string
		Buyers     []string
	}

	destinationDatacenterMap := make(map[uint64]*DestinationDatacenterRow)

	for _, v1 := range database.BuyerDatacenterSettings {
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
			entry.Buyers = append(entry.Buyers, database.BuyerMap[buyerId].Code)
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
			Id:              fmt.Sprintf("%016x", v.Id),
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
			Id:   fmt.Sprintf("%016x", v.Id),
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
			Id:        fmt.Sprintf("%016x", v.Id),
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
		InternalGroup   string
		PublicKey       string
		PrivateKey      string
		Price           string
	}

	relays := []RelayRow{}

	for _, v := range database.RelayMap {

		row := RelayRow{
			Id:            fmt.Sprintf("%016x", v.Id),
			Name:          v.Name,
			PublicAddress: v.PublicAddress.String(),
			PublicKey:     base64.StdEncoding.EncodeToString(v.PublicKey),
			PrivateKey:    base64.StdEncoding.EncodeToString(v.PrivateKey),
			Price:         fmt.Sprintf("%d", v.BandwidthPrice),
		}

		if v.HasInternalAddress {
			row.InternalAddress = v.InternalAddress.String()
			if v.InternalGroup != 0 {
				row.InternalGroup = fmt.Sprintf("%016x", v.InternalGroup)
			}
		}

		relays = append(relays, row)
	}

	sort.SliceStable(relays, func(i, j int) bool { return relays[i].Name < relays[j].Name })

	fmt.Fprintf(w, "<br><br>Relays:<br><br>")
	fmt.Fprintf(w, "<table>\n")
	fmt.Fprintf(w, "<tr><td><b>%s</b></td><td><b>%s</b></td><td><b>%s</b></td><td><b>%s</b></td><td><b>%s</b></td><td><b>%s</b></td><td><b>%s</b></td></tr>\n", "Id", "Name", "Public Address", "Internal Address", "Internal Group", "Price", "Public Key")
	for i := range relays {
		fmt.Fprintf(w, "<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>\n", relays[i].Id, relays[i].Name, relays[i].PublicAddress, relays[i].InternalAddress, relays[i].InternalGroup, relays[i].Price, relays[i].PublicKey)
	}
	fmt.Fprintf(w, "</table>\n")

	// route shaders

	for _, v := range database.BuyerMap {

		routeShader := v.RouteShader

		fmt.Fprintf(w, "<br><br>Route shader for '%s':<br><br>", v.Name)
		fmt.Fprintf(w, "<table>\n")
		fmt.Fprintf(w, "<tr><td><b>%s</b></td><td><b>%v</b></td>\n", "Property", "Value")
		fmt.Fprintf(w, "<tr><td>%s</td><td>%v</td>\n", "Disable Network Next", routeShader.DisableNetworkNext)
		fmt.Fprintf(w, "<tr><td>%s</td><td>%v</td>\n", "AB Test", routeShader.ABTest)
		fmt.Fprintf(w, "<tr><td>%s</td><td>%v</td>\n", "Force Next", routeShader.ForceNext)
		fmt.Fprintf(w, "<tr><td>%s</td><td>%d%%</td>\n", "Selection Percent", routeShader.SelectionPercent)
		fmt.Fprintf(w, "<tr><td>%s</td><td>%dms</td>\n", "Acceptable Latency", routeShader.AcceptableLatency)
		fmt.Fprintf(w, "<tr><td>%s</td><td>%dms</td>\n", "Latency Threshold", routeShader.LatencyReductionThreshold)
		fmt.Fprintf(w, "<tr><td>%s</td><td>%.1f%%</td>\n", "Acceptable Packet Loss", routeShader.AcceptablePacketLoss)
		fmt.Fprintf(w, "<tr><td>%s</td><td>%dkbps</td>\n", "Bandwidth Envelope Up", routeShader.BandwidthEnvelopeUpKbps)
		fmt.Fprintf(w, "<tr><td>%s</td><td>%dkbps</td>\n", "Bandwidth Envelope Down", routeShader.BandwidthEnvelopeDownKbps)
		fmt.Fprintf(w, "<tr><td>%s</td><td>%dms</td>\n", "Route Select Threshold", routeShader.RouteSelectThreshold)
		fmt.Fprintf(w, "<tr><td>%s</td><td>%dms</td>\n", "Route Switch Threshold", routeShader.RouteSwitchThreshold)
		fmt.Fprintf(w, "<tr><td>%s</td><td>%dms</td>\n", "Max Latency Trade Off", routeShader.MaxLatencyTradeOff)
		fmt.Fprintf(w, "</table>\n")
	}

	// destination datacenters

	type DestinationDatacenterRow struct {
		Datacenter string
		Buyers     []string
	}

	destinationDatacenterMap := make(map[uint64]*DestinationDatacenterRow)

	for _, v1 := range database.BuyerDatacenterSettings {
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
		relay_id           uint64
		relay_name         string
		datacenter_id      uint64
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
		bandwidth_price    int
	}

	relayRows := make([]RelayRow, 0)
	{
		rows, err := pgsql.Query("SELECT relay_id, relay_name, datacenter_id, public_ip, public_port, internal_ip, internal_port, internal_group, ssh_ip, ssh_port, ssh_user, public_key_base64, private_key_base64, version, mrc, port_speed, max_sessions, bandwidth_price FROM relays")
		if err != nil {
			return nil, fmt.Errorf("could not extract relays: %v\n", err)
		}

		defer rows.Close()

		for rows.Next() {
			row := RelayRow{}
			if err := rows.Scan(&row.relay_id, &row.relay_name, &row.datacenter_id, &row.public_ip, &row.public_port, &row.internal_ip, &row.internal_port, &row.internal_group, &row.ssh_ip, &row.ssh_port, &row.ssh_user, &row.public_key_base64, &row.private_key_base64, &row.version, &row.mrc, &row.port_speed, &row.max_sessions, &row.bandwidth_price); err != nil {
				return nil, fmt.Errorf("failed to scan relay row: %v\n", err)
			}
			relayRows = append(relayRows, row)
		}
	}

	// datacenters

	type DatacenterRow struct {
		datacenter_id   uint64
		datacenter_name string
		native_name     string
		latitude        float32
		longitude       float32
		seller_id       uint64
	}

	datacenterRows := make([]DatacenterRow, 0)
	{
		rows, err := pgsql.Query("SELECT datacenter_id, datacenter_name, native_name, latitude, longitude, seller_id FROM datacenters")
		if err != nil {
			return nil, fmt.Errorf("could not extract datacenters: %v\n", err)
		}

		defer rows.Close()

		for rows.Next() {
			row := DatacenterRow{}
			if err := rows.Scan(&row.datacenter_id, &row.datacenter_name, &row.native_name, &row.latitude, &row.longitude, &row.seller_id); err != nil {
				return nil, fmt.Errorf("failed to scan datacenter row: %v\n", err)
			}
			datacenterRows = append(datacenterRows, row)
		}
	}

	// buyers

	type BuyerRow struct {
		buyer_id          uint64
		buyer_name        string
		buyer_code        string
		public_key_base64 string
		route_shader_id   uint64
		live              bool
		debug             bool
	}

	buyerRows := make([]BuyerRow, 0)
	{
		rows, err := pgsql.Query("SELECT buyer_id, buyer_name, buyer_code, public_key_base64, route_shader_id, live, debug FROM buyers")
		if err != nil {
			return nil, fmt.Errorf("could not extract buyers: %v\n", err)
		}

		defer rows.Close()

		for rows.Next() {
			row := BuyerRow{}
			if err := rows.Scan(&row.buyer_id, &row.buyer_name, &row.buyer_code, &row.public_key_base64, &row.route_shader_id, &row.live, &row.debug); err != nil {
				return nil, fmt.Errorf("failed to scan buyer row: %v\n", err)
			}
			buyerRows = append(buyerRows, row)
		}
	}

	// sellers

	type SellerRow struct {
		seller_id   uint64
		seller_name string
		seller_code string
	}

	sellerRows := make([]SellerRow, 0)
	{
		rows, err := pgsql.Query("SELECT seller_id, seller_name, seller_code FROM sellers")
		if err != nil {
			return nil, fmt.Errorf("could not extract sellers: %v\n", err)
		}

		defer rows.Close()

		for rows.Next() {
			row := SellerRow{}
			if err := rows.Scan(&row.seller_id, &row.seller_name, &row.seller_code); err != nil {
				return nil, fmt.Errorf("failed to scan seller row: %v\n", err)
			}
			sellerRows = append(sellerRows, row)
		}
	}

	// route shaders

	type RouteShaderRow struct {
		route_shader_id                  uint64
		ab_test                          bool
		acceptable_latency               int
		acceptable_packet_loss           float32
		bandwidth_envelope_down_kbps     int
		bandwidth_envelope_up_kbps       int
		disable_network_next             bool
		latency_reduction_threshold      int
		selection_percent                int
		max_latency_trade_off            int
		route_switch_threshold           int
		route_select_threshold           int
		force_next                       bool
	}

	routeShaderRows := make([]RouteShaderRow, 0)
	{
		rows, err := pgsql.Query("SELECT route_shader_id, ab_test, acceptable_latency, acceptable_packet_loss, bandwidth_envelope_down_kbps, bandwidth_envelope_up_kbps, disable_network_next, latency_reduction_threshold, selection_percent, max_latency_trade_off, route_switch_threshold, route_select_threshold, force_next FROM route_shaders")
		if err != nil {
			return nil, fmt.Errorf("could not extract route shaders: %v\n", err)
		}

		defer rows.Close()

		for rows.Next() {
			row := RouteShaderRow{}
			if err := rows.Scan(&row.route_shader_id, &row.ab_test, &row.acceptable_latency, &row.acceptable_packet_loss, &row.bandwidth_envelope_down_kbps, &row.bandwidth_envelope_up_kbps, &row.disable_network_next, &row.latency_reduction_threshold, &row.selection_percent, &row.max_latency_trade_off, &row.route_switch_threshold, &row.route_select_threshold, &row.force_next); err != nil {
				return nil, fmt.Errorf("failed to scan route shader row: %v\n", err)
			}
			routeShaderRows = append(routeShaderRows, row)
		}
	}

	// buyer datacenter settings

	type BuyerDatacenterSettingsRow struct {
		buyer_id            uint64
		datacenter_id       uint64
		enable_acceleration bool
	}

	buyerDatacenterSettingsRows := make([]BuyerDatacenterSettingsRow, 0)
	{
		rows, err := pgsql.Query("SELECT buyer_id, datacenter_id, enable_acceleration FROM buyer_datacenter_settings")
		if err != nil {
			return nil, fmt.Errorf("could not extract buyer datacenter settings: %v\n", err)
		}

		defer rows.Close()

		for rows.Next() {
			row := BuyerDatacenterSettingsRow{}
			if err := rows.Scan(&row.buyer_id, &row.datacenter_id, &row.enable_acceleration); err != nil {
				return nil, fmt.Errorf("failed to scan buyer datacenter settings row: %v\n", err)
			}
			buyerDatacenterSettingsRows = append(buyerDatacenterSettingsRows, row)
		}
	}

	// print out rows

	fmt.Printf("\nrelays:\n")
	for _, row := range relayRows {
		fmt.Printf("%d: %s, %d, %s, %d, %s, %d, %s, %d, %s, %s, %s, %s, %d, %d, %d, %d\n", row.relay_id, row.relay_name, row.datacenter_id, row.public_ip, row.public_port, row.internal_ip, row.internal_port, row.ssh_ip, row.ssh_port, row.ssh_user, row.public_key_base64, row.private_key_base64, row.version.String, row.mrc, row.port_speed, row.max_sessions, row.bandwidth_price)
	}

	fmt.Printf("\ndatacenters:\n")
	for _, row := range datacenterRows {
		fmt.Printf("%d: %s, %s, %.1f, %.1f, %d\n", row.datacenter_id, row.datacenter_name, row.native_name, row.latitude, row.longitude, row.seller_id)
	}

	fmt.Printf("\nbuyers:\n")
	for _, row := range buyerRows {
		fmt.Printf("%d: %s, %s, %s\n", row.buyer_id, row.buyer_name, row.buyer_code, row.public_key_base64)
	}

	fmt.Printf("\nsellers:\n")
	for _, row := range sellerRows {
		fmt.Printf("%d: %s, %s\n", row.seller_id, row.seller_name, row.seller_code)
	}

	fmt.Printf("\nroute shaders:\n")
	for _, row := range routeShaderRows {
		fmt.Printf("%d: %v, %d, %.1f, %d, %d, %v, %d, %d, %d, %d, %d, %v\n",
			row.route_shader_id,
			row.ab_test,
			row.acceptable_latency,
			row.acceptable_packet_loss,
			row.bandwidth_envelope_down_kbps,
			row.bandwidth_envelope_up_kbps,
			row.disable_network_next,
			row.latency_reduction_threshold,
			row.selection_percent,
			row.max_latency_trade_off,
			row.route_switch_threshold,
			row.route_select_threshold,
			row.force_next)
	}

	fmt.Printf("\nbuyer datacenter settings:\n")
	for _, row := range buyerDatacenterSettingsRows {
		fmt.Printf("(%d,%d): %v\n", row.buyer_id, row.datacenter_id, row.enable_acceleration)
	}

	// index datacenters by postgres id

	datacenterIndex := make(map[uint64]DatacenterRow)
	for _, row := range datacenterRows {
		datacenterIndex[row.datacenter_id] = row
	}

	// index buyers by postgres id

	buyerIndex := make(map[uint64]BuyerRow)
	for _, row := range buyerRows {
		buyerIndex[row.buyer_id] = row
	}

	// index sellers by postgres id

	sellerIndex := make(map[uint64]SellerRow)
	for _, row := range sellerRows {
		sellerIndex[row.seller_id] = row
	}

	// index route shaders by postgres id

	routeShaderIndex := make(map[uint64]RouteShaderRow)
	for _, row := range routeShaderRows {
		routeShaderIndex[row.route_shader_id] = row
	}

	// build database

	fmt.Printf("\nbuilding network next database...\n\n")

	database := CreateDatabase()

	database.CreationTime = time.Now().Format("Monday 02 January 2006 15:04:05 MST")
	database.Creator = "extract_database"

	database.Relays = make([]Relay, len(relayRows))

	for i, row := range sellerRows {

		seller := Seller{}

		seller.Id = row.seller_id
		seller.Name = row.seller_name
		seller.Code = row.seller_code

		database.SellerMap[seller.Id] = &seller

		fmt.Printf("seller %d: %s %s [%d]\n", i, seller.Name, seller.Code, seller.Id)
	}

	for i, row := range buyerRows {

		buyer := Buyer{}

		buyer.Name = row.buyer_name
		buyer.Code = row.buyer_code

		data, err := base64.StdEncoding.DecodeString(row.public_key_base64)
		if err != nil {
			return nil, fmt.Errorf("could not decode public key base64 for buyer %s: %v\n", buyer.Name, err)
		}

		if len(data) != 40 {
			// IMPORTANT: Downgrade this to a warning because otherwise the API service can get stuck in a broken state
			fmt.Printf("warning: buyer '%s' public key data is invalid. Expected 40 bytes, got %d\n", buyer.Name, len(data))
			data = make([]byte, 40)
		}

		buyer.Id = binary.LittleEndian.Uint64(data[:8])
		buyer.PublicKey = data[8:40]

		buyer.Live = row.live
		buyer.Debug = row.debug

		route_shader_row, route_shader_exists := routeShaderIndex[row.route_shader_id]
		if !route_shader_exists {
			return nil, fmt.Errorf("buyer %s does not have a route shader\n", buyer.Name)
		}

		buyer.RouteShader.DisableNetworkNext = route_shader_row.disable_network_next
		buyer.RouteShader.SelectionPercent = route_shader_row.selection_percent
		buyer.RouteShader.ABTest = route_shader_row.ab_test
		buyer.RouteShader.AcceptableLatency = int32(route_shader_row.acceptable_latency)
		buyer.RouteShader.LatencyReductionThreshold = int32(route_shader_row.latency_reduction_threshold)
		buyer.RouteShader.AcceptablePacketLoss = route_shader_row.acceptable_packet_loss
		buyer.RouteShader.BandwidthEnvelopeUpKbps = int32(route_shader_row.bandwidth_envelope_up_kbps)
		buyer.RouteShader.BandwidthEnvelopeDownKbps = int32(route_shader_row.bandwidth_envelope_down_kbps)
		buyer.RouteShader.RouteSelectThreshold = int32(route_shader_row.route_select_threshold)
		buyer.RouteShader.RouteSwitchThreshold = int32(route_shader_row.route_switch_threshold)
		buyer.RouteShader.MaxLatencyTradeOff = int32(route_shader_row.max_latency_trade_off)
		buyer.RouteShader.ForceNext = route_shader_row.force_next

		database.BuyerMap[buyer.Id] = &buyer

		fmt.Printf("buyer %d: %s %s [%x] (live=%v, debug=%v)\n", i, buyer.Name, buyer.Code, buyer.Id, buyer.Live, buyer.Debug)
	}

	for i, row := range datacenterRows {

		datacenter := Datacenter{}

		datacenter.Id = HashString(row.datacenter_name)
		datacenter.Name = row.datacenter_name
		datacenter.Native = row.native_name
		datacenter.Latitude = row.latitude
		datacenter.Longitude = row.longitude
		datacenter.SellerId = row.seller_id

		_, seller_exists := sellerIndex[row.seller_id]
		if !seller_exists {
			return nil, fmt.Errorf("datacenter %s doesn't have a seller\n", datacenter.Name)
		}

		database.DatacenterMap[datacenter.Id] = &datacenter

		fmt.Printf("datacenter %d: %s [%x] (%.1f,%.1f)\n", i, datacenter.Name, datacenter.Id, datacenter.Latitude, datacenter.Longitude)
	}

	for i, row := range relayRows {

		relay := &database.Relays[i]

		relay.Name = row.relay_name

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
		if relay.SSHAddress.String() == "0.0.0.0:0" {
			relay.SSHAddress = relay.PublicAddress
		}
		relay.SSHAddress.Port = row.ssh_port
		relay.SSHUser = row.ssh_user

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
		relay.BandwidthPrice = row.bandwidth_price

		datacenter_row, datacenter_exists := datacenterIndex[row.datacenter_id]
		if !datacenter_exists {
			return nil, fmt.Errorf("relay %s doesn't have a datacenter\n", relay.Name)
		}

		relay.DatacenterId = HashString(datacenter_row.datacenter_name)

		relay.Datacenter = database.DatacenterMap[relay.DatacenterId]
		if relay.Datacenter.Id != relay.DatacenterId {
			return nil, fmt.Errorf("relay '%s' has a bad datacenter\n", relay.Name)
		}

		seller_row, seller_exists := sellerIndex[datacenter_row.seller_id]
		if !seller_exists {
			return nil, fmt.Errorf("relay %s doesn't have a seller\n", relay.Name)
		}

		relay.Seller = database.SellerMap[seller_row.seller_id]

		fmt.Printf("relay %d: %s -> %s [%x]\n", i, relay.Name, datacenter_row.datacenter_name, relay.DatacenterId)

		database.RelayMap[relay.Id] = relay
		database.RelayNameMap[relay.Name] = relay
	}

	for i, row := range buyerDatacenterSettingsRows {

		buyer_row, buyer_exists := buyerIndex[row.buyer_id]
		if !buyer_exists {
			return nil, fmt.Errorf("buyer datacenter settings don't have a buyer\n")
		}

		datacenter_row, datacenter_exists := datacenterIndex[row.datacenter_id]
		if !datacenter_exists {
			return nil, fmt.Errorf("buyer datacenter settings don't have a datacenter\n")
		}

		buyerName := buyer_row.buyer_name
		buyerId := uint64(0)
		datacenterName := datacenter_row.datacenter_name
		datacenterId := HashString(datacenterName)

		for _, v := range database.BuyerMap {
			if v.Name == buyerName {
				buyerId = v.Id
			}
		}

		if buyerId == 0 {
			// IMPORTANT: Downgrade to a warning otherwise the API service can get stuck in a broken state
			fmt.Printf("warning: could not find runtime buyer id for buyer '%s'\n", buyerName)
		}

		fmt.Printf("buyer datacenter settings %d: %s [%x] -> %s [%x] enabled\n", i, buyerName, buyerId, datacenterName, datacenterId)

		settings := BuyerDatacenterSettings{}
		settings.BuyerId = buyerId
		settings.DatacenterId = datacenterId
		settings.EnableAcceleration = row.enable_acceleration
		if database.BuyerDatacenterSettings[buyerId] == nil {
			database.BuyerDatacenterSettings[buyerId] = make(map[uint64]*BuyerDatacenterSettings)
		}
		database.BuyerDatacenterSettings[buyerId][datacenterId] = &settings
	}

	for i := range database.Relays {
		relayId := database.Relays[i].Id
		datacenterId := database.Relays[i].Datacenter.Id
		datacenterRelays := database.DatacenterRelays[datacenterId]
		datacenterRelays = append(datacenterRelays, relayId)
		database.DatacenterRelays[datacenterId] = datacenterRelays
	}

	database.Fixup()

	return database, nil
}

// -----------------------------------------------------------------------------------------------------------

func (database *Database) LoadBinary(data []byte) error {

	compressed_buffer := bytes.NewReader(data)

	gz_reader, err := gzip.NewReader(compressed_buffer)
	if err != nil {
		return err
	}

	err = gob.NewDecoder(gz_reader).Decode(database)

	if err == nil {
		database.Fixup()
	}

	return err
}

func (database *Database) GetBinary() []byte {

	var buffer bytes.Buffer

	err := gob.NewEncoder(&buffer).Encode(database)
	if err != nil {
		return nil
	}

	var compressed_buffer bytes.Buffer
	gz, err := gzip.NewWriterLevel(&compressed_buffer, gzip.BestCompression)
	if err != nil {
		return nil
	}

	if _, err := gz.Write(buffer.Bytes()); err != nil {
		return nil
	}

	if err := gz.Close(); err != nil {
		return nil
	}

	return compressed_buffer.Bytes()
}

type HeaderResponse struct {
	CreationTime   string `json:"creation_time"`
	Creator        string `json:"creator"`
	NumRelays      int    `json:"num_relays"`
	NumBuyers      int    `json:"num_buyers"`
	NumSellers     int    `json:"num_sellers"`
	NumDatacenters int    `json:"num_datacenters"`
}

func (database *Database) GetHeader() *HeaderResponse {
	header := HeaderResponse{}
	header.CreationTime = database.CreationTime
	header.Creator = database.Creator
	header.NumRelays = len(database.Relays)
	header.NumBuyers = len(database.BuyerMap)
	header.NumSellers = len(database.SellerMap)
	header.NumDatacenters = len(database.DatacenterMap)
	return &header
}

type RelaysResponse struct {
	Relays []Relay `json:"relays"`
}

func (database *Database) GetRelays() *RelaysResponse {
	response := RelaysResponse{}
	response.Relays = database.Relays
	return &response
}

type BuyersResponse struct {
	Buyers []Buyer `json:"buyers"`
}

func (database *Database) GetBuyers() *BuyersResponse {
	response := BuyersResponse{}
	response.Buyers = make([]Buyer, len(database.BuyerMap))
	index := 0
	for _, v := range database.BuyerMap {
		response.Buyers[index] = *v
		index++
	}
	return &response
}

type SellersResponse struct {
	Sellers []Seller `json:"sellers"`
}

func (database *Database) GetSellers() *SellersResponse {
	response := SellersResponse{}
	response.Sellers = make([]Seller, len(database.SellerMap))
	index := 0
	for _, v := range database.SellerMap {
		response.Sellers[index] = *v
		index++
	}
	return &response
}

type DatacentersResponse struct {
	Datacenters []Datacenter `json:"datacenters"`
}

func (database *Database) GetDatacenters() *DatacentersResponse {
	response := DatacentersResponse{}
	response.Datacenters = make([]Datacenter, len(database.DatacenterMap))
	index := 0
	for _, v := range database.DatacenterMap {
		response.Datacenters[index] = *v
		index++
	}
	return &response
}

type BuyerDatacenterSettingsResponse struct {
	BuyerDatacenterSettings []BuyerDatacenterSettings `json:"buyer_datacenter_settings"`
}

func (database *Database) GetBuyerDatacenterSettings() *BuyerDatacenterSettingsResponse {
	response := BuyerDatacenterSettingsResponse{}
	response.BuyerDatacenterSettings = make([]BuyerDatacenterSettings, 0)
	for _, datacenterMap := range database.BuyerDatacenterSettings {
		for _, settings := range datacenterMap {
			response.BuyerDatacenterSettings = append(response.BuyerDatacenterSettings, *settings)
		}
	}
	return &response
}

// -----------------------------------------------------------------------------------------------------------
