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
	Id              uint64
	Name            string
	DatacenterId    uint64
	PublicAddress   net.UDPAddr
	HasInternalAddress bool
	InternalAddress net.UDPAddr
	SSHAddress      net.UDPAddr
	SSHUser         string
	PublicKey       []byte
	PrivateKey      []byte
	MaxSessions     int
	PortSpeed       int
	Version         string
	Seller          Seller
	Datacenter      Datacenter
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
