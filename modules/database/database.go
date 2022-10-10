package database

import (
	"bytes"
	"encoding/gob"
	"io/ioutil"
	"net"
	"os"
)

type Relay struct {
	ID           uint64
	Name         string
	Addr         net.UDPAddr
	InternalAddr net.UDPAddr
	PublicKey    []byte
	Seller       Seller
	Datacenter   Datacenter
	MaxSessions  uint32
}

type Buyer struct {
	// todo
}

type Seller struct {
	// todo
}

type Datacenter struct {
	// todo
}

type DatacenterMap struct {
	// todo: whut
}

type Database struct {
	CreationTime   string
	Creator        string
	Relays         []Relay
	/*
	RelayMap       map[uint64]Relay
	BuyerMap       map[uint64]Buyer
	SellerMap      map[string]Seller
	DatacenterMap  map[uint64]Datacenter
	DatacenterMaps map[uint64]map[uint64]DatacenterMap // todo: datacenter maps design strikes me as bad
	*/
	//                 ^ Buyer.ID   ^ DatacenterMap map index
}

func CreateDatabase() *Database {

	database := &Database{
		CreationTime:   "",
		Creator:        "",
		Relays:         []Relay{},
		/*
		RelayMap:       make(map[uint64]Relay),
		BuyerMap:       make(map[uint64]Buyer),
		SellerMap:      make(map[string]Seller),
		DatacenterMap:  make(map[uint64]Datacenter),
		DatacenterMaps: make(map[uint64]map[uint64]DatacenterMap),
		*/
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

	/*
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
	*/

	return true
}

// ---------------------------------------------------------------------

/*
type Overlay struct {
	CreationTime string
	BuyerMap     map[uint64]Buyer
}
*/

/*
func CreateEmptyOverlayBinWrapper() *OverlayBinWrapper {
	wrapper := &OverlayBinWrapper{
		BuyerMap: make(map[uint64]Buyer),
	}

	return wrapper
}

func (wrapper OverlayBinWrapper) IsEmpty() bool {
	return len(wrapper.BuyerMap) == 0
}

func (wrapper OverlayBinWrapper) WriteOverlayBinFile(outputPath string) error {
	var buffer bytes.Buffer

	err := gob.NewEncoder(&buffer).Encode(wrapper)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(outputPath, buffer.Bytes(), 0644); err != nil {
		return err
	}

	return nil
}

// This function is essentially the same as DecodeOverlayWrapper in modules/backend/helpers.go
func (wrapper *OverlayBinWrapper) ReadOverlayBinFile(overlayFilePath string) error {
	overlayFile, err := os.Open(overlayFilePath)
	if err != nil {
		return err
	}
	defer overlayFile.Close()

	return gob.NewDecoder(overlayFile).Decode(wrapper)
}
*/
