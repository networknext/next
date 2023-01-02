package database

import (
	"bytes"
	"encoding/gob"
	"io/ioutil"
	"net"
	"os"

	"github.com/networknext/backend/modules/core"
)

type Relay struct {
	ID           uint64
	Name         string
	Addr         net.UDPAddr
	InternalAddr net.UDPAddr
	Version      string
	PublicKey    []byte
	Seller       Seller
	Datacenter   Datacenter
	MaxSessions  uint32
	NICSpeedMbps int32
}

type Buyer struct {
	ID             uint64
	Live           bool
	Debug          bool
	PublicKey      []byte
	RouteShader    core.RouteShader
	InternalConfig core.InternalConfig
}

type Seller struct {
	ID   string // todo: needs to be a string for compatibility with old database, but should be a uint64 in future
	Name string
}

type Datacenter struct {
	ID        uint64
	Name      string
	Latitude  float32 // todo: need to put in Latitude and Longitude fields in BinWrapper struct to make this work
	Longitude float32
}

// todo: what's really going on here? It's a per-buyer, per-datacenter struct? or something?
type DatacenterMap struct {
	BuyerID            uint64
	DatacenterID       uint64
	EnableAcceleration bool
}

type Database struct {
	CreationTime   string
	Creator        string
	Relays         []Relay
	RelayMap       map[uint64]Relay
	BuyerMap       map[uint64]Buyer
	SellerMap      map[string]Seller
	DatacenterMap  map[uint64]Datacenter
	DatacenterMaps map[uint64]map[uint64]DatacenterMap         // todo: this is really just a map from (buyerId,datacenterId) -> true/false for enabling a datacenter. surely there is a better way to express this
	//                   ^ BuyerId  ^ DatacenterId
}

func CreateDatabase() *Database {

	database := &Database{
		CreationTime:   "",
		Creator:        "",
		Relays:         []Relay{},
		RelayMap:       make(map[uint64]Relay),
		BuyerMap:       make(map[uint64]Buyer),
		SellerMap:      make(map[string]Seller),
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

func (database *Database) IsEmpty() bool {

	if database.CreationTime == "" {
		return false
	}

	if database.Creator == "" {
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
