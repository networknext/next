package database

import (
	// "bytes"
	// "encoding/gob"
	// "io/ioutil"
	// "os"
)

type Relay struct {
	// todo
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
	RelayMap       map[uint64]Relay
	BuyerMap       map[uint64]Buyer
	SellerMap      map[string]Seller
	DatacenterMap  map[uint64]Datacenter
	DatacenterMaps map[uint64]map[uint64]DatacenterMap              // todo: datacenter maps design strikes me as bad
	//                 ^ Buyer.ID   ^ DatacenterMap map index
}

type Overlay struct {
	CreationTime string
	BuyerMap     map[uint64]Buyer
}

/*
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

func (wrapper DatabaseBinWrapper) WriteDatabaseBinFile(outputPath string) error {
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

// This function is essentially the same as DecodeDatabaseWrapper in modules/backend/helpers.go
func (wrapper *DatabaseBinWrapper) ReadDatabaseBinFile(databaseFilePath string) error {
	databaseFile, err := os.Open(databaseFilePath)
	if err != nil {
		return err
	}
	defer databaseFile.Close()

	return gob.NewDecoder(databaseFile).Decode(wrapper)
}
*/

// overlay

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