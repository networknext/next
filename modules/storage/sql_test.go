package storage_test

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/networknext/backend/modules/backend"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/transport/looker"
	"github.com/stretchr/testify/assert"
)

// TODO: first clean and reset postgresql

// TODO: sort environment in/from Makefile
func SetupEnv() {
	os.Setenv("ENV", "local")
	os.Setenv("FEATURE_POSTGRESQL", "false")
	os.Setenv("DB_SYNC_INTERVAL", "10s")
}

func TestInsertSQL(t *testing.T) {

	SetupEnv()

	ctx := context.Background()
	logger := log.NewNopLogger()

	env, err := backend.GetEnv()
	assert.NoError(t, err)
	db, err := backend.GetStorer(ctx, logger, "local", env)
	assert.NoError(t, err)

	var outerCustomer routing.Customer
	var outerBuyer routing.Buyer
	var outerSeller routing.Seller
	var outerDatacenter routing.Datacenter
	var initialRelayVersion string

	customerShortname := "compcode"

	t.Run("AddCustomer", func(t *testing.T) {
		customer := routing.Customer{
			Code:                   customerShortname,
			Name:                   "Company, Ltd.",
			AutomaticSignInDomains: "fredscuttle.com",
		}

		err = db.AddCustomer(ctx, customer)
		assert.NoError(t, err)

		outerCustomer, err = db.Customer(ctx, customerShortname)
		assert.NoError(t, err)
		assert.Equal(t, customer.Code, outerCustomer.Code)
		assert.Equal(t, customer.Name, outerCustomer.Name)
		assert.Equal(t, customer.AutomaticSignInDomains, outerCustomer.AutomaticSignInDomains)
	})

	t.Run("AddSeller", func(t *testing.T) {
		seller := routing.Seller{
			ID:                       customerShortname,
			ShortName:                customerShortname,
			CompanyCode:              customerShortname,
			Secret:                   true,
			EgressPriceNibblinsPerGB: 20,
			CustomerID:               outerCustomer.DatabaseID,
		}

		err = db.AddSeller(ctx, seller)
		assert.NoError(t, err)

		outerSeller, err = db.Seller(ctx, "compcode")
		assert.NoError(t, err)
		assert.Equal(t, seller.ID, outerSeller.ID)
		assert.Equal(t, true, outerSeller.Secret)
		assert.Equal(t, seller.ShortName, outerSeller.ShortName)
		assert.Equal(t, seller.EgressPriceNibblinsPerGB, outerSeller.EgressPriceNibblinsPerGB)
		assert.Equal(t, seller.CustomerID, outerSeller.CustomerID)
	})

	t.Run("AddDatacenter", func(t *testing.T) {

		datacenter := routing.Datacenter{
			ID:   crypto.HashID("some.locale.name"),
			Name: "some.locale.name",
			Location: routing.Location{
				Latitude:  70.5,
				Longitude: 120.5,
			},
			SellerID: outerSeller.DatabaseID,
		}

		err = db.AddDatacenter(ctx, datacenter)
		assert.NoError(t, err)

		outerDatacenter, err = db.Datacenter(ctx, datacenter.ID)
		assert.NoError(t, err)
		assert.Equal(t, outerDatacenter.ID, datacenter.ID)
		assert.Equal(t, outerDatacenter.Name, datacenter.Name)
		assert.Equal(t, outerDatacenter.Location.Latitude, datacenter.Location.Latitude)
		assert.Equal(t, outerDatacenter.Location.Longitude, datacenter.Location.Longitude)
		assert.Equal(t, outerDatacenter.SellerID, datacenter.SellerID)
	})

	t.Run("AddBuyer", func(t *testing.T) {

		publicKey := make([]byte, crypto.KeySize)
		_, err := rand.Read(publicKey)
		assert.NoError(t, err)

		internalID := uint64(3142537350691193170)

		buyer := routing.Buyer{
			ID:          internalID,
			ShortName:   outerCustomer.Code,
			CompanyCode: outerCustomer.Code,
			Live:        true,
			Debug:       true,
			PublicKey:   publicKey,
			// CustomerID:  outerCustomer.DatabaseID,
		}

		err = db.AddBuyer(ctx, buyer)
		assert.NoError(t, err)

		outerBuyer, err = db.Buyer(ctx, internalID)
		assert.NoError(t, err)

		assert.Equal(t, internalID, outerBuyer.ID)
		assert.Equal(t, buyer.Live, outerBuyer.Live)
		assert.Equal(t, buyer.Debug, outerBuyer.Debug)
		assert.Equal(t, publicKey, outerBuyer.PublicKey)
		// assert.Equal(t, buyer.CustomerID, outerBuyer.CustomerID)
		assert.Equal(t, buyer.ShortName, outerBuyer.ShortName)
		assert.Equal(t, buyer.CompanyCode, outerBuyer.CompanyCode)
	})

	t.Run("AddRelay", func(t *testing.T) {

		// relay with no null values (except dc to trip an error)
		addr, err := net.ResolveUDPAddr("udp", "127.0.0.2:40000")
		assert.NoError(t, err)

		internalAddr, err := net.ResolveUDPAddr("udp", "172.20.2.6:40000")
		assert.NoError(t, err)

		rid := crypto.HashID(addr.String())

		publicKey := make([]byte, crypto.KeySize)
		_, err = rand.Read(publicKey)
		assert.NoError(t, err)

		initialRelayVersion = "2.0.6"

		// fields not stored in the database are not tested here
		relay := routing.Relay{
			ID:              rid,
			Name:            "test.1",
			Addr:            *addr,
			InternalAddr:    *internalAddr,
			ManagementAddr:  "1.2.3.4",
			SSHPort:         22,
			SSHUser:         "fred",
			MaxSessions:     1000,
			PublicKey:       publicKey,
			BillingSupplier: outerSeller.ShortName,
			// Datacenter:     outerDatacenter,
			EgressPriceOverride:           10000000000000,
			MRC:                           19700000000000,
			Overage:                       26000000000000,
			BWRule:                        routing.BWRuleBurst,
			ContractTerm:                  12,
			StartDate:                     time.Now(),
			EndDate:                       time.Now(),
			Type:                          routing.BareMetal,
			IncludedBandwidthGB:           10000,
			NICSpeedMbps:                  1000,
			MaxBandwidthMbps:              900,
			Notes:                         "the original notes",
			Version:                       initialRelayVersion,
			DestFirst:                     false,
			InternalAddressClientRoutable: true,
		}

		// adding a relay w/o a valid datacenter should return an FK violation error
		err = db.AddRelay(ctx, relay)
		assert.Error(t, err)
		assert.EqualError(t, err, "FOREIGN KEY constraint failed")

		relay.Datacenter = outerDatacenter

		// adding a relay w/o an internal address should return an error if InternalAddressClientRoutable is true
		relay.InternalAddr = net.UDPAddr{}
		err = db.AddRelay(ctx, relay)
		assert.Error(t, err)
		assert.EqualError(t, err, "relay flag InternalAddressClientRoutable cannot be true without valid internal IP")

		relay.InternalAddr = *internalAddr

		// TODO: repeat the above test with bwrule, type and state

		err = db.AddRelay(ctx, relay)
		assert.NoError(t, err)

		// Trying to add this relay again should throw an error
		err = db.AddRelay(ctx, relay)
		assert.Error(t, err)
		assert.EqualError(t, err, fmt.Sprintf("relay %s (%016x) (state: %s) already exists with this IP address. please reuse this relay.", relay.Name, relay.ID, relay.State.String()))

		// check only the fields set above
		checkRelay, err := db.Relay(ctx, rid)
		assert.NoError(t, err)

		assert.Equal(t, relay.Name, checkRelay.Name)
		assert.Equal(t, relay.Addr, checkRelay.Addr)
		assert.Equal(t, relay.InternalAddr, checkRelay.InternalAddr)
		assert.Equal(t, relay.ManagementAddr, checkRelay.ManagementAddr)
		assert.Equal(t, relay.SSHPort, checkRelay.SSHPort)
		assert.Equal(t, relay.SSHUser, checkRelay.SSHUser)
		assert.Equal(t, relay.MaxSessions, checkRelay.MaxSessions)
		assert.Equal(t, relay.PublicKey, checkRelay.PublicKey)
		assert.Equal(t, relay.Datacenter.DatabaseID, checkRelay.Datacenter.DatabaseID)
		assert.Equal(t, relay.EgressPriceOverride, checkRelay.EgressPriceOverride)
		assert.Equal(t, relay.MRC, checkRelay.MRC)
		assert.Equal(t, relay.Overage, checkRelay.Overage)
		assert.Equal(t, relay.BWRule, checkRelay.BWRule)
		assert.Equal(t, relay.ContractTerm, checkRelay.ContractTerm)
		assert.Equal(t, relay.StartDate.Format("01/02/06"), checkRelay.StartDate.Format("01/02/06"))
		assert.Equal(t, relay.EndDate.Format("01/02/06"), checkRelay.EndDate.Format("01/02/06"))
		assert.Equal(t, relay.Type, checkRelay.Type)
		assert.Equal(t, routing.RelayStateEnabled, checkRelay.State)
		assert.Equal(t, int32(10000), checkRelay.IncludedBandwidthGB)
		assert.Equal(t, int32(1000), checkRelay.NICSpeedMbps)
		assert.Equal(t, int32(900), checkRelay.MaxBandwidthMbps)
		assert.Equal(t, customerShortname, checkRelay.Seller.ID)
		assert.Equal(t, customerShortname, checkRelay.Seller.ShortName)
		assert.Equal(t, customerShortname, checkRelay.Seller.CompanyCode)
		assert.Equal(t, routing.Nibblin(20), checkRelay.Seller.EgressPriceNibblinsPerGB)
		assert.Equal(t, outerCustomer.DatabaseID, checkRelay.Seller.CustomerID)
		assert.Equal(t, relay.Notes, checkRelay.Notes)
		assert.Equal(t, outerSeller.ShortName, checkRelay.BillingSupplier)
		assert.Equal(t, initialRelayVersion, checkRelay.Version)
		assert.Equal(t, relay.DestFirst, checkRelay.DestFirst)
		assert.Equal(t, relay.InternalAddressClientRoutable, checkRelay.InternalAddressClientRoutable)

		// overwrite with SetRelay - test nullable fields, possible in relay_backend
		var relayMod routing.Relay

		relayMod.ID = checkRelay.ID
		relayMod.Name = checkRelay.Name
		// Addr
		// InternalAddr
		relayMod.ManagementAddr = checkRelay.ManagementAddr
		relayMod.SSHPort = checkRelay.SSHPort
		relayMod.SSHUser = checkRelay.SSHUser
		relayMod.MaxSessions = checkRelay.MaxSessions
		relayMod.PublicKey = checkRelay.PublicKey
		relayMod.Datacenter = checkRelay.Datacenter
		relayMod.EgressPriceOverride = checkRelay.EgressPriceOverride
		relayMod.MRC = checkRelay.MRC
		relayMod.Overage = checkRelay.Overage
		relayMod.BWRule = checkRelay.BWRule
		relayMod.ContractTerm = checkRelay.ContractTerm
		// StartDate
		// EndDate
		relayMod.Type = checkRelay.Type
		relayMod.State = checkRelay.State
		relayMod.IncludedBandwidthGB = checkRelay.IncludedBandwidthGB
		relayMod.NICSpeedMbps = checkRelay.NICSpeedMbps
		relayMod.MaxBandwidthMbps = checkRelay.MaxBandwidthMbps
		relayMod.Notes = checkRelay.Notes
		relayMod.DatabaseID = checkRelay.DatabaseID

		relayMod.Seller = checkRelay.Seller

		err = db.SetRelay(ctx, relayMod)
		assert.NoError(t, err)

		checkRelayMod, err := db.Relay(ctx, relay.ID)
		assert.NoError(t, err)

		assert.Equal(t, relayMod.Name, checkRelayMod.Name)
		assert.Equal(t, net.UDPAddr{IP: net.IP(nil), Port: 0, Zone: ""}, checkRelayMod.Addr)
		assert.Equal(t, net.UDPAddr{IP: net.IP(nil), Port: 0, Zone: ""}, checkRelayMod.InternalAddr)
		assert.Equal(t, relayMod.ManagementAddr, checkRelayMod.ManagementAddr)
		assert.Equal(t, relayMod.SSHPort, checkRelayMod.SSHPort)
		assert.Equal(t, relayMod.SSHUser, checkRelayMod.SSHUser)
		assert.Equal(t, relayMod.MaxSessions, checkRelayMod.MaxSessions)
		assert.Equal(t, relayMod.PublicKey, checkRelayMod.PublicKey)
		assert.Equal(t, relayMod.Datacenter.DatabaseID, checkRelayMod.Datacenter.DatabaseID)
		assert.Equal(t, relayMod.EgressPriceOverride, checkRelayMod.EgressPriceOverride)
		assert.Equal(t, relayMod.MRC, checkRelayMod.MRC)
		assert.Equal(t, relayMod.Overage, checkRelayMod.Overage)
		assert.Equal(t, relayMod.BWRule, checkRelayMod.BWRule)
		assert.Equal(t, relayMod.ContractTerm, checkRelayMod.ContractTerm)
		assert.True(t, checkRelayMod.StartDate.IsZero())
		assert.True(t, checkRelayMod.EndDate.IsZero())
		assert.Equal(t, relayMod.Type, checkRelayMod.Type)
		assert.Equal(t, relayMod.State, checkRelayMod.State)
		assert.Equal(t, int32(10000), checkRelayMod.IncludedBandwidthGB)
		assert.Equal(t, int32(1000), checkRelayMod.NICSpeedMbps)
		assert.Equal(t, int32(900), checkRelayMod.MaxBandwidthMbps)

		assert.Equal(t, customerShortname, checkRelayMod.Seller.ID)
		assert.Equal(t, customerShortname, checkRelayMod.Seller.ShortName)
		assert.Equal(t, customerShortname, checkRelayMod.Seller.CompanyCode)
		assert.Equal(t, routing.Nibblin(20), checkRelayMod.Seller.EgressPriceNibblinsPerGB)
		assert.Equal(t, outerCustomer.DatabaseID, checkRelayMod.Seller.CustomerID)
		assert.Equal(t, relayMod.Notes, checkRelayMod.Notes)

		// relay with some null values null values (except dc to trip an error)
		addr2, err := net.ResolveUDPAddr("udp", "127.0.0.3:40000")
		assert.NoError(t, err)

		rid2 := crypto.HashID(addr2.String())

		publicKey = make([]byte, crypto.KeySize)
		_, err = rand.Read(publicKey)
		assert.NoError(t, err)

		// fields not stored in the database are not tested here
		relay3 := routing.Relay{
			ID:   rid2,
			Name: "test.3",
			Addr: *addr2,
			// InternalAddr:   *internalAddr, <-- nullable
			ManagementAddr:      "1.2.3.4",
			SSHPort:             22,
			SSHUser:             "fred",
			MaxSessions:         1000,
			PublicKey:           publicKey,
			Datacenter:          outerDatacenter,
			EgressPriceOverride: 10000000000000,
			MRC:                 19700000000000,
			Overage:             26000000000000,
			BWRule:              routing.BWRuleBurst,
			ContractTerm:        12,
			// StartDate:           time.Now(), <-- nullable
			// EndDate:             time.Now(), <-- nullable
			Type:                          routing.BareMetal,
			State:                         routing.RelayStateMaintenance,
			IncludedBandwidthGB:           10000,
			NICSpeedMbps:                  1000,
			MaxBandwidthMbps:              900,
			Notes:                         "the original notes",
			Version:                       initialRelayVersion,
			DestFirst:                     true,
			InternalAddressClientRoutable: false,
		}

		err = db.AddRelay(ctx, relay3)
		assert.NoError(t, err)

		// check only the fields *not* set above
		checkRelay2, err := db.Relay(ctx, rid2)
		assert.NoError(t, err)

		assert.Equal(t, net.UDPAddr{IP: net.IP(nil), Port: 0, Zone: ""}, checkRelay2.InternalAddr)
		assert.True(t, checkRelay2.StartDate.IsZero())
		assert.True(t, checkRelay2.EndDate.IsZero())

		relay4 := routing.Relay{
			ID:   rid2,
			Name: "test.3.a",
			Addr: *addr2,
			// InternalAddr:   *internalAddr, <-- nullable
			ManagementAddr:      "1.2.3.4",
			SSHPort:             22,
			SSHUser:             "fred",
			MaxSessions:         1000,
			PublicKey:           publicKey,
			Datacenter:          outerDatacenter,
			EgressPriceOverride: 10000000000000,
			MRC:                 19700000000000,
			Overage:             26000000000000,
			BWRule:              routing.BWRuleBurst,
			ContractTerm:        12,
			// StartDate:           time.Now(), <-- nullable
			// EndDate:             time.Now(), <-- nullable
			Type:                          routing.BareMetal,
			State:                         routing.RelayStateMaintenance,
			IncludedBandwidthGB:           10000,
			NICSpeedMbps:                  1000,
			MaxBandwidthMbps:              900,
			Notes:                         "the original notes",
			Version:                       initialRelayVersion,
			DestFirst:                     true,
			InternalAddressClientRoutable: false,
		}

		err = db.AddRelay(ctx, relay4)
		assert.Error(t, err)

		relayMod = relay3

		relayMod.State = routing.RelayStateDecommissioned

		err = db.UpdateRelay(ctx, rid2, "State", float64(routing.RelayStateDecommissioned))
		assert.NoError(t, err)

		// Don't allow a relay to be readded with the same ID, even if it is decommissioned
		err = db.AddRelay(ctx, relay4)
		assert.Error(t, err)
		assert.EqualError(t, err, fmt.Sprintf("relay %s (%016x) (state: %s) already exists with this IP address. please reuse this relay.", relayMod.Name, relayMod.ID, relayMod.State.String()))

	})

	t.Run("AddRelayWithNullables", func(t *testing.T) {

		addr, err := net.ResolveUDPAddr("udp", "127.3.4.5:40000")
		assert.NoError(t, err)

		rid := crypto.HashID(addr.String())

		publicKey := make([]byte, crypto.KeySize)
		_, err = rand.Read(publicKey)
		assert.NoError(t, err)

		// fields not stored in the database are not tested here
		relay := routing.Relay{
			ID:   rid,
			Name: "nullable.test.1",
			Addr: *addr,
			// InternalAddr:   *internalAddr,
			ManagementAddr: "1.2.3.5",
			SSHPort:        22,
			SSHUser:        "fred",
			MaxSessions:    1000,
			PublicKey:      publicKey,
			// Datacenter:     outerDatacenter,
			EgressPriceOverride: 10000000000000,
			MRC:                 19700000000000,
			Overage:             26000000000000,
			BWRule:              routing.BWRuleBurst,
			ContractTerm:        12,
			// StartDate:           time.Now(),
			// EndDate:             time.Now(),
			Type:                routing.BareMetal,
			State:               routing.RelayStateMaintenance,
			IncludedBandwidthGB: 10000,
			NICSpeedMbps:        1000,
			MaxBandwidthMbps:    900,
			// Notes: "the original notes"
			Version:                       initialRelayVersion,
			DestFirst:                     true,
			InternalAddressClientRoutable: false,
		}

		// adding a relay w/o a valid datacenter should return an FK violation error
		err = db.AddRelay(ctx, relay)
		assert.Error(t, err)

		// TODO repeat the above test with bwrule, type and state

		relay.Datacenter = outerDatacenter
		err = db.AddRelay(ctx, relay)
		assert.NoError(t, err)

		// check only the fields set above
		checkRelay, err := db.Relay(ctx, rid)
		assert.NoError(t, err)

		assert.Equal(t, relay.Name, checkRelay.Name)
		assert.Equal(t, relay.Addr, checkRelay.Addr)
		assert.Equal(t, relay.ManagementAddr, checkRelay.ManagementAddr)
		assert.Equal(t, relay.SSHPort, checkRelay.SSHPort)
		assert.Equal(t, relay.SSHUser, checkRelay.SSHUser)
		assert.Equal(t, relay.MaxSessions, checkRelay.MaxSessions)
		assert.Equal(t, relay.PublicKey, checkRelay.PublicKey)
		assert.Equal(t, relay.Datacenter.DatabaseID, checkRelay.Datacenter.DatabaseID)
		assert.Equal(t, relay.EgressPriceOverride, checkRelay.EgressPriceOverride)
		assert.Equal(t, relay.MRC, checkRelay.MRC)
		assert.Equal(t, relay.Overage, checkRelay.Overage)
		assert.Equal(t, relay.BWRule, checkRelay.BWRule)
		assert.Equal(t, relay.ContractTerm, checkRelay.ContractTerm)

		// dates are null, though no "zero" value for InternalAddr to test
		assert.Equal(t, time.Time{}.Format("01/02/06"), checkRelay.StartDate.Format("01/02/06"))
		assert.Equal(t, time.Time{}.Format("01/02/06"), checkRelay.EndDate.Format("01/02/06"))
		assert.Equal(t, relay.Type, checkRelay.Type)
		assert.Equal(t, relay.State, checkRelay.State)
		assert.Equal(t, int32(10000), checkRelay.IncludedBandwidthGB)
		assert.Equal(t, int32(1000), checkRelay.NICSpeedMbps)
		assert.Equal(t, int32(900), checkRelay.MaxBandwidthMbps)
		assert.Equal(t, relay.DestFirst, checkRelay.DestFirst)
		assert.Equal(t, relay.InternalAddressClientRoutable, checkRelay.InternalAddressClientRoutable)

		assert.Equal(t, customerShortname, checkRelay.Seller.ID)
		assert.Equal(t, customerShortname, checkRelay.Seller.ShortName)
		assert.Equal(t, customerShortname, checkRelay.Seller.CompanyCode)
		assert.Equal(t, routing.Nibblin(20), checkRelay.Seller.EgressPriceNibblinsPerGB)
		assert.Equal(t, outerCustomer.DatabaseID, checkRelay.Seller.CustomerID)
	})

	t.Run("AddDatacenterMap", func(t *testing.T) {
		dcMap := routing.DatacenterMap{
			BuyerID:      outerBuyer.ID,
			DatacenterID: outerDatacenter.ID,
		}

		err := db.AddDatacenterMap(ctx, dcMap)
		assert.NoError(t, err)

		checkDCMaps := db.GetDatacenterMapsForBuyer(ctx, outerBuyer.ID)

		assert.Equal(t, 1, len(checkDCMaps))
		assert.Equal(t, dcMap.BuyerID, checkDCMaps[outerDatacenter.ID].BuyerID)
		assert.Equal(t, dcMap.DatacenterID, checkDCMaps[outerDatacenter.ID].DatacenterID)
	})
}

func TestDeleteSQL(t *testing.T) {

	SetupEnv()

	ctx := context.Background()
	logger := log.NewNopLogger()

	// db, err := storage.NewSQLStorage(ctx, logger)
	env, err := backend.GetEnv()
	assert.NoError(t, err)
	db, err := backend.GetStorer(ctx, logger, "local", env)
	assert.NoError(t, err)

	time.Sleep(1000 * time.Millisecond) // allow time for sync functions to complete
	assert.NoError(t, err)

	var outerCustomer routing.Customer
	var outerBuyer routing.Buyer
	var outerSeller routing.Seller
	var outerDatacenter routing.Datacenter
	var outerDatacenterMap routing.DatacenterMap

	t.Run("ExerciseFKs", func(t *testing.T) {

		customerCode := "compcode"
		customer := routing.Customer{
			Code:                   customerCode,
			Name:                   "Company, Ltd.",
			AutomaticSignInDomains: "fredscuttle.com",
		}

		err = db.AddCustomer(ctx, customer)
		assert.NoError(t, err)

		outerCustomer, err = db.Customer(ctx, customerCode)
		assert.NoError(t, err)

		publicKey := make([]byte, crypto.KeySize)
		_, err := rand.Read(publicKey)
		assert.NoError(t, err)

		internalID := uint64(3142537350691193170)

		buyer := routing.Buyer{
			ID:          internalID,
			ShortName:   outerCustomer.Code,
			CompanyCode: outerCustomer.Code,
			Live:        true,
			Debug:       true,
			PublicKey:   publicKey,
			// CustomerID:  outerCustomer.DatabaseID,
		}

		err = db.AddBuyer(ctx, buyer)
		assert.NoError(t, err)

		outerBuyer, err = db.Buyer(ctx, internalID)
		assert.NoError(t, err)

		seller := routing.Seller{
			ID:                       "compcode",
			ShortName:                "compcode",
			EgressPriceNibblinsPerGB: 20,
			Secret:                   true,
			CustomerID:               outerCustomer.DatabaseID,
			CompanyCode:              outerCustomer.Code,
		}

		err = db.AddSeller(ctx, seller)
		assert.NoError(t, err)

		outerSeller, err = db.Seller(ctx, "compcode")
		assert.NoError(t, err)

		datacenter := routing.Datacenter{
			ID:   crypto.HashID("some.locale.name"),
			Name: "some.locale.name",
			Location: routing.Location{
				Latitude:  70.5,
				Longitude: 120.5,
			},
			SellerID: outerSeller.DatabaseID,
		}

		err = db.AddDatacenter(ctx, datacenter)
		assert.NoError(t, err)

		outerDatacenter, err = db.Datacenter(ctx, datacenter.ID)
		assert.NoError(t, err)

		dcMap := routing.DatacenterMap{
			BuyerID:      outerBuyer.ID,
			DatacenterID: outerDatacenter.ID,
		}

		err = db.AddDatacenterMap(ctx, dcMap)
		assert.NoError(t, err)

		dcMaps := db.GetDatacenterMapsForBuyer(ctx, outerBuyer.ID)
		assert.Equal(t, 1, len(dcMaps))
		outerDatacenterMap = dcMaps[outerDatacenter.ID]

		addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
		assert.NoError(t, err)

		internalAddr, err := net.ResolveUDPAddr("udp", "172.20.2.6:40000")
		assert.NoError(t, err)

		rid := crypto.HashID(addr.String())

		relayPublicKey := make([]byte, crypto.KeySize)
		_, err = rand.Read(relayPublicKey)
		assert.NoError(t, err)

		relay := routing.Relay{
			ID:                  rid,
			Name:                "test.1",
			Addr:                *addr,
			InternalAddr:        *internalAddr,
			ManagementAddr:      "1.2.3.4",
			SSHPort:             22,
			SSHUser:             "fred",
			MaxSessions:         1000,
			PublicKey:           relayPublicKey,
			Datacenter:          outerDatacenter,
			EgressPriceOverride: 10000000000000,
			MRC:                 19700000000000,
			Overage:             26000000000000,
			BWRule:              routing.BWRuleBurst,
			ContractTerm:        12,
			StartDate:           time.Now(),
			EndDate:             time.Now(),
			Type:                routing.BareMetal,
			State:               routing.RelayStateMaintenance,
			Notes:               "the original notes",
			Version:             "2.0.6",
		}

		err = db.AddRelay(ctx, relay)
		assert.NoError(t, err)

		// Attempting to remove the customer should return a foreign
		// key violation error (for buyer and/or seller)
		// sqlite3: FOREIGN KEY constraint failed
		err = db.RemoveCustomer(ctx, "compcode")
		assert.Error(t, err)

		// Attempting to remove the buyer should return an FK
		// violation error (for datacenter maps and banned users (TBD))
		err = db.RemoveBuyer(ctx, outerBuyer.ID)
		assert.Error(t, err)

		// Attempting to remove the seller should return an FK
		// violation error (for the datacenter)
		err = db.RemoveSeller(ctx, outerSeller.ID)
		assert.Error(t, err)

		// Attempting to remove the datacenter should return an FK
		// violation error (for the datacenter map)
		err = db.RemoveDatacenter(ctx, outerDatacenter.ID)
		assert.Error(t, err)

		err = db.RemoveDatacenterMap(ctx, outerDatacenterMap)
		assert.NoError(t, err)

		dcMapsCheck := db.GetDatacenterMapsForBuyer(ctx, outerBuyer.ID)
		assert.Equal(t, 0, len(dcMapsCheck))

		err = db.RemoveBuyer(ctx, outerBuyer.ID)
		assert.NoError(t, err)

		_, err = db.Buyer(ctx, outerBuyer.ID)
		assert.Error(t, err)

		err = db.RemoveRelay(ctx, relay.ID)
		assert.NoError(t, err)

		_, err = db.Relay(ctx, relay.ID)
		assert.Error(t, err)

		err = db.RemoveDatacenter(ctx, outerDatacenter.ID)
		assert.NoError(t, err)

		_, err = db.Datacenter(ctx, outerDatacenter.ID)
		assert.Error(t, err)

		err = db.RemoveSeller(ctx, outerSeller.ID)
		assert.NoError(t, err)

		_, err = db.Seller(ctx, outerSeller.ID)
		assert.Error(t, err)

		err = db.RemoveCustomer(ctx, "compcode")
		assert.NoError(t, err)

		_, err = db.Customer(ctx, "compcode")
		assert.Error(t, err)
	})
}

func TestUpdateSQL(t *testing.T) {

	SetupEnv()

	ctx := context.Background()
	logger := log.NewNopLogger()

	// db, err := storage.NewSQLStorage(ctx, logger)
	env, err := backend.GetEnv()
	assert.NoError(t, err)
	db, err := backend.GetStorer(ctx, logger, "local", env)
	assert.NoError(t, err)

	time.Sleep(1000 * time.Millisecond) // allow time for sync functions to complete
	assert.NoError(t, err)

	// var customerWithID, customerWithID2 routing.Customer
	var customerWithID routing.Customer
	var buyerWithID routing.Buyer
	var sellerWithID, sellerWithID2 routing.Seller
	var datacenterWithID routing.Datacenter
	// var outerDatacenterMap routing.DatacenterMap

	t.Run("SetCustomer", func(t *testing.T) {
		customer := routing.Customer{
			Code:                   "compcode",
			Name:                   "Company, Ltd.",
			AutomaticSignInDomains: "fredscuttle.com",
		}

		err = db.AddCustomer(ctx, customer)
		assert.NoError(t, err)

		// the CustomerID field is the PK and is set by AddCustomer(). In
		// production usage this field would already be set and sync'd.
		customerWithID, err = db.Customer(ctx, "compcode")

		customerWithID.Name = "No Longer The Company, Ltd."
		customerWithID.AutomaticSignInDomains = "fredscuttle.com,swampthing.com"

		err = db.SetCustomer(ctx, customerWithID)
		assert.NoError(t, err)

		checkCustomer, err := db.Customer(ctx, "compcode")
		assert.NoError(t, err)

		assert.Equal(t, customerWithID.AutomaticSignInDomains, checkCustomer.AutomaticSignInDomains)
		assert.Equal(t, customerWithID.Name, checkCustomer.Name)

		// we need a second customer to check Relay.BillingSupplier
		customer2 := routing.Customer{
			Code:                   "DifferentSupplier",
			Name:                   "Different Supplier, Ltd.",
			AutomaticSignInDomains: "differentsupplier.com",
		}

		err = db.AddCustomer(ctx, customer2)
		assert.NoError(t, err)

		_, err = db.Customer(ctx, "DifferentSupplier")
		assert.NoError(t, err)

	})

	t.Run("UpdateDatacenter", func(t *testing.T) {

		seller := routing.Seller{
			ID:                       "compcode",
			ShortName:                "compcode",
			EgressPriceNibblinsPerGB: 20,
			Secret:                   true,
			CustomerID:               customerWithID.DatabaseID,
			CompanyCode:              customerWithID.Code,
		}

		err = db.AddSeller(ctx, seller)
		assert.NoError(t, err)

		sellerWithID, err = db.Seller(ctx, "compcode")
		assert.NoError(t, err)

		did := crypto.HashID("some.locale.name")
		datacenter := routing.Datacenter{
			ID:   did,
			Name: "some.locale.name",
			Location: routing.Location{
				Latitude:  70.5,
				Longitude: 120.5,
			},
			SellerID: sellerWithID.DatabaseID,
		}

		err = db.AddDatacenter(ctx, datacenter)
		assert.NoError(t, err)

		err = db.UpdateDatacenter(ctx, did, "Latitude", float32(130.3))
		assert.NoError(t, err)

		err = db.UpdateDatacenter(ctx, did, "Longitude", float32(80.3))
		assert.NoError(t, err)

		checkDatacenter, err := db.Datacenter(ctx, did)
		assert.NoError(t, err)
		assert.Equal(t, float32(80.3), checkDatacenter.Location.Longitude)
		assert.Equal(t, float32(130.3), checkDatacenter.Location.Latitude)
	})

	t.Run("UpdateDatacenterMap", func(t *testing.T) {

		publicKey := make([]byte, crypto.KeySize)
		_, err = rand.Read(publicKey)
		assert.NoError(t, err)

		internalID := uint64(3142537350691193170)

		buyer := routing.Buyer{
			ID:          internalID,
			ShortName:   customerWithID.Code,
			CompanyCode: customerWithID.Code,
			Live:        true,
			Debug:       true,
			PublicKey:   publicKey,
			// CustomerID:  customerWithID.DatabaseID,
		}

		err = db.AddBuyer(ctx, buyer)
		assert.NoError(t, err)

		buyerWithID, err = db.Buyer(ctx, internalID)
		assert.NoError(t, err)

		did1 := crypto.HashID("some.locale.name.1")
		datacenter1 := routing.Datacenter{
			ID:   did1,
			Name: "some.locale.name.1",
			Location: routing.Location{
				Latitude:  73.5,
				Longitude: 10.5,
			},
			SellerID: sellerWithID.DatabaseID,
		}

		err = db.AddDatacenter(ctx, datacenter1)
		assert.NoError(t, err)

		_, err := db.Datacenter(ctx, did1)
		assert.NoError(t, err)

		did2 := crypto.HashID("some.locale.name.2")
		datacenter2 := routing.Datacenter{
			ID:   did2,
			Name: "some.locale.name.2",
			Location: routing.Location{
				Latitude:  73.5,
				Longitude: 10.5,
			},
			SellerID: sellerWithID.DatabaseID,
		}

		err = db.AddDatacenter(ctx, datacenter2)
		assert.NoError(t, err)

		_, err = db.Datacenter(ctx, did2)
		assert.NoError(t, err)

		dcMap := routing.DatacenterMap{
			BuyerID:      buyerWithID.ID,
			DatacenterID: datacenter1.ID,
		}

		err = db.AddDatacenterMap(ctx, dcMap)
		assert.NoError(t, err)

		// hexDcID := fmt.Sprintf("%016x", did2)
		// err = db.UpdateDatacenterMap(ctx, buyerWithID.ID, datacenter1.ID, "HexDatacenterID", hexDcID)
		// assert.NoError(t, err)

		// checkDcMaps := db.GetDatacenterMapsForBuyer(buyerWithID.ID)
		// assert.Equal(t, 1, len(checkDcMaps))

		// assert.Equal(t, did2, checkDcMaps[did2].DatacenterID)
		// assert.Equal(t, buyerWithID.ID, checkDcMaps[did2].BuyerID)

	})

	t.Run("UpdateCustomer", func(t *testing.T) {
		err := db.UpdateCustomer(ctx, customerWithID.Code, "Name", "A Brand New Name")
		assert.NoError(t, err)

		err = db.UpdateCustomer(ctx, customerWithID.Code, "AutomaticSigninDomains", "somewhere.com,somewhere.else.com")
		assert.NoError(t, err)

		checkCustomer, err := db.Customer(ctx, customerWithID.Code)
		assert.NoError(t, err)

		assert.Equal(t, "A Brand New Name", checkCustomer.Name)
		assert.Equal(t, "somewhere.com,somewhere.else.com", checkCustomer.AutomaticSignInDomains)
	})

	t.Run("UpdateBuyer", func(t *testing.T) {
		err := db.UpdateBuyer(ctx, buyerWithID.ID, "Live", false)
		assert.NoError(t, err)

		err = db.UpdateBuyer(ctx, buyerWithID.ID, "Debug", false)
		assert.NoError(t, err)

		err = db.UpdateBuyer(ctx, buyerWithID.ID, "ShortName", "newname")
		assert.NoError(t, err)

		err = db.UpdateBuyer(ctx, buyerWithID.ID, "LookerSeats", int64(100))
		assert.NoError(t, err)

		err = db.UpdateBuyer(ctx, buyerWithID.ID, "ExoticLocationFee", float64(100))
		assert.NoError(t, err)

		err = db.UpdateBuyer(ctx, buyerWithID.ID, "StandardLocationFee", float64(100))
		assert.NoError(t, err)

		newPublicKeyStr := "YFWQjOJfHfOqsCMM/1pd+c5haMhsrE2Gm05bVUQhCnG7YlPUrI/d1g=="
		newPublicKeyEncoded, err := base64.StdEncoding.DecodeString(newPublicKeyStr)
		assert.NoError(t, err)
		newBuyerID := binary.LittleEndian.Uint64(newPublicKeyEncoded[:8])

		err = db.UpdateBuyer(ctx, buyerWithID.ID, "PublicKey", newPublicKeyStr)
		assert.NoError(t, err)

		// the changed public key also changes the buyer ID
		checkBuyer, err := db.Buyer(ctx, newBuyerID)
		assert.NoError(t, err)

		assert.Equal(t, false, checkBuyer.Live)
		assert.Equal(t, false, checkBuyer.Debug)
		assert.Equal(t, "newname", checkBuyer.ShortName)
		assert.Equal(t, int64(100), checkBuyer.LookerSeats)
		assert.Equal(t, float64(100), checkBuyer.ExoticLocationFee)
		assert.Equal(t, float64(100), checkBuyer.StandardLocationFee)
		assert.Equal(t, newBuyerID, checkBuyer.ID)
		assert.Equal(t, newPublicKeyEncoded[8:], checkBuyer.PublicKey)
		assert.Equal(t, newPublicKeyStr, checkBuyer.EncodedPublicKey())

		// a datacenter map for this buyer were added above and the UpdateBuyer method
		// must modify it for the new ID
		dcMaps := db.GetDatacenterMapsForBuyer(ctx, newBuyerID)
		assert.Equal(t, 1, len(dcMaps))
	})

	t.Run("UpdateSeller", func(t *testing.T) {

		err := db.UpdateSeller(ctx, sellerWithID.ID, "EgressPriceNibblinsPerGB", 133.44)
		assert.NoError(t, err)

		err = db.UpdateSeller(ctx, sellerWithID.ID, "Secret", false)
		assert.NoError(t, err)

		checkSeller, err := db.Seller(ctx, sellerWithID.ID)
		assert.NoError(t, err)

		assert.Equal(t, routing.Nibblin(13344000000000), checkSeller.EgressPriceNibblinsPerGB)
		assert.Equal(t, false, checkSeller.Secret)
	})

	t.Run("UpdateRelay", func(t *testing.T) {

		addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
		assert.NoError(t, err)

		rid := crypto.HashID(addr.String())

		internalAddr, err := net.ResolveUDPAddr("udp", "172.20.2.6:40000")
		assert.NoError(t, err)

		publicKey := make([]byte, crypto.KeySize)
		_, err = rand.Read(publicKey)
		assert.NoError(t, err)

		initialRelayVersion := "2.0.6"

		did := crypto.HashID("some.locale.name")
		datacenterWithID, err = db.Datacenter(ctx, did)
		assert.NoError(t, err)

		relay := routing.Relay{
			ID:                            rid,
			Name:                          "test.1",
			Addr:                          *addr,
			InternalAddr:                  *internalAddr,
			ManagementAddr:                "1.2.3.4",
			BillingSupplier:               sellerWithID.ShortName,
			SSHPort:                       22,
			SSHUser:                       "fred",
			MaxSessions:                   1000,
			PublicKey:                     publicKey,
			Datacenter:                    datacenterWithID,
			EgressPriceOverride:           10000000000000,
			MRC:                           19700000000000,
			Overage:                       26000000000000,
			BWRule:                        routing.BWRuleBurst,
			NICSpeedMbps:                  1000,
			MaxBandwidthMbps:              900,
			IncludedBandwidthGB:           10000,
			ContractTerm:                  12,
			StartDate:                     time.Now(),
			EndDate:                       time.Now(),
			Type:                          routing.BareMetal,
			State:                         routing.RelayStateMaintenance,
			Notes:                         "the original notes",
			Version:                       initialRelayVersion,
			DestFirst:                     false,
			InternalAddressClientRoutable: false,
		}

		err = db.AddRelay(ctx, relay)
		assert.NoError(t, err)

		_, err = db.Relay(ctx, rid)
		assert.NoError(t, err)

		// relay.Name
		err = db.UpdateRelay(ctx, rid, "Name", "test.2")
		assert.NoError(t, err)
		checkRelay, err := db.Relay(ctx, rid)
		assert.NoError(t, err)
		assert.Equal(t, "test.2", checkRelay.Name)

		// relay.Addr
		newAddr, err := net.ResolveUDPAddr("udp", "192.168.0.1:40000")
		assert.NoError(t, err)
		err = db.UpdateRelay(ctx, rid, "Addr", "192.168.0.1:40000")
		assert.NoError(t, err)
		checkRelay, err = db.Relay(ctx, rid)
		assert.NoError(t, err)
		assert.Equal(t, *newAddr, checkRelay.Addr)

		// relay.Addr (zeroed-out address e.g. relay removal)
		err = db.UpdateRelay(ctx, rid, "Addr", "")
		assert.NoError(t, err)
		checkRelay, err = db.Relay(ctx, rid)
		assert.NoError(t, err)
		assert.Equal(t, ":0", checkRelay.Addr.String())

		// relay.InternalAddr
		intAddr, err := net.ResolveUDPAddr("udp", "192.168.0.2:40000")
		assert.NoError(t, err)
		err = db.UpdateRelay(ctx, rid, "InternalAddr", "192.168.0.2:40000")
		assert.NoError(t, err)
		checkRelay, err = db.Relay(ctx, rid)
		assert.NoError(t, err)
		assert.Equal(t, *intAddr, checkRelay.InternalAddr)

		// relay.InternalAddr (null)
		err = db.UpdateRelay(ctx, rid, "InternalAddressClientRoutable", true)
		assert.NoError(t, err)
		err = db.UpdateRelay(ctx, rid, "InternalAddr", "")
		assert.Error(t, err)
		assert.EqualError(t, err, "cannot remove internal address while InternalAddressClientRoutable is true")
		err = db.UpdateRelay(ctx, rid, "InternalAddressClientRoutable", false)
		assert.NoError(t, err)

		err = db.UpdateRelay(ctx, rid, "InternalAddr", "")
		assert.NoError(t, err)
		checkRelay, err = db.Relay(ctx, rid)
		assert.NoError(t, err)
		assert.Equal(t, net.UDPAddr{}, checkRelay.InternalAddr)

		// relay.ManagementAddr
		err = db.UpdateRelay(ctx, rid, "ManagementAddr", "9.8.7.6")
		assert.NoError(t, err)
		checkRelay, err = db.Relay(ctx, rid)
		assert.NoError(t, err)
		assert.Equal(t, "9.8.7.6", checkRelay.ManagementAddr)

		// relay.SSHPort
		// Note: ints in json are unmarshalled as float64
		err = db.UpdateRelay(ctx, rid, "SSHPort", float64(13))
		assert.NoError(t, err)
		checkRelay, err = db.Relay(ctx, rid)
		assert.NoError(t, err)
		assert.Equal(t, int64(13), checkRelay.SSHPort)

		// checkRelay.SSHUser
		err = db.UpdateRelay(ctx, rid, "SSHUser", "Abercrombie")
		assert.NoError(t, err)
		checkRelay, err = db.Relay(ctx, rid)
		assert.NoError(t, err)
		assert.Equal(t, "Abercrombie", checkRelay.SSHUser)

		// relay.MaxSessions
		err = db.UpdateRelay(ctx, rid, "MaxSessions", float64(25000))
		assert.NoError(t, err)
		checkRelay, err = db.Relay(ctx, rid)
		assert.NoError(t, err)
		assert.Equal(t, uint32(25000), checkRelay.MaxSessions)

		// relay.PublicKey
		err = db.UpdateRelay(ctx, rid, "PublicKey", "1AKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=")
		assert.NoError(t, err)
		checkRelay, err = db.Relay(ctx, rid)
		assert.NoError(t, err)
		newPublicKey, err := base64.StdEncoding.DecodeString("1AKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=")
		assert.NoError(t, err)
		assert.Equal(t, newPublicKey, checkRelay.PublicKey)

		// relay.Datacenter = only one datacenter available...

		// relay.EgressPriceOverride
		err = db.UpdateRelay(ctx, rid, "EgressPriceOverride", float64(200))
		assert.NoError(t, err)
		checkRelay, err = db.Relay(ctx, rid)
		assert.NoError(t, err)
		assert.Equal(t, routing.Nibblin(20000000000000), checkRelay.EgressPriceOverride)

		// relay.MRC
		err = db.UpdateRelay(ctx, rid, "MRC", float64(397))
		assert.NoError(t, err)
		checkRelay, err = db.Relay(ctx, rid)
		assert.NoError(t, err)
		assert.Equal(t, routing.Nibblin(39700000000000), checkRelay.MRC)

		// relay.Overage
		err = db.UpdateRelay(ctx, rid, "Overage", float64(260))
		assert.NoError(t, err)
		checkRelay, err = db.Relay(ctx, rid)
		assert.NoError(t, err)
		assert.Equal(t, routing.Nibblin(26000000000000), checkRelay.Overage)

		// relay.BWRule
		err = db.UpdateRelay(ctx, rid, "BWRule", float64(3))
		assert.NoError(t, err)
		checkRelay, err = db.Relay(ctx, rid)
		assert.NoError(t, err)
		assert.Equal(t, routing.BWRulePool, checkRelay.BWRule)

		// relay.ContractTerm
		err = db.UpdateRelay(ctx, rid, "ContractTerm", float64(1))
		assert.NoError(t, err)
		checkRelay, err = db.Relay(ctx, rid)
		assert.NoError(t, err)
		assert.Equal(t, int32(1), checkRelay.ContractTerm)

		// relay.StartDate
		// We use a string as type-switching (in UpdateRelay()) doesn't work with a time.Time type
		startDate := "July 7, 2023"
		err = db.UpdateRelay(ctx, rid, "StartDate", startDate)
		assert.NoError(t, err)

		checkRelay, err = db.Relay(ctx, rid)
		assert.NoError(t, err)

		startDateFormatted, err := time.Parse("January 2, 2006", startDate)
		assert.NoError(t, err)
		assert.Equal(t, startDateFormatted, checkRelay.StartDate)

		// relay.StartDate (null)
		err = db.UpdateRelay(ctx, rid, "StartDate", "")
		assert.NoError(t, err)
		checkRelay, err = db.Relay(ctx, rid)
		assert.NoError(t, err)
		assert.Equal(t, time.Time{}, checkRelay.StartDate)

		// relay.EndDate
		endDate := "July 7, 2025"
		err = db.UpdateRelay(ctx, rid, "EndDate", endDate)
		assert.NoError(t, err)
		checkRelay, err = db.Relay(ctx, rid)
		assert.NoError(t, err)
		endDateFormatted, err := time.Parse("January 2, 2006", endDate)
		assert.NoError(t, err)
		assert.Equal(t, endDateFormatted, checkRelay.EndDate)

		// relay.EndDate (null)
		err = db.UpdateRelay(ctx, rid, "EndDate", "")
		assert.NoError(t, err)
		checkRelay, err = db.Relay(ctx, rid)
		assert.NoError(t, err)
		assert.Equal(t, time.Time{}, checkRelay.EndDate)

		// relay.Type
		err = db.UpdateRelay(ctx, rid, "Type", float64(2))
		assert.NoError(t, err)
		checkRelay, err = db.Relay(ctx, rid)
		assert.NoError(t, err)
		assert.Equal(t, routing.VirtualMachine, checkRelay.Type)

		// relay.State
		err = db.UpdateRelay(ctx, rid, "State", float64(0))
		assert.NoError(t, err)
		checkRelay, err = db.Relay(ctx, rid)
		assert.NoError(t, err)
		assert.Equal(t, routing.RelayStateEnabled, checkRelay.State)

		// relay.NICSpeedMbps
		err = db.UpdateRelay(ctx, rid, "NICSpeedMbps", float64(20000))
		assert.NoError(t, err)
		checkRelay, err = db.Relay(ctx, rid)
		assert.NoError(t, err)
		assert.Equal(t, int32(20000), checkRelay.NICSpeedMbps)

		// relay.IncludedBandwidthGB
		err = db.UpdateRelay(ctx, rid, "IncludedBandwidthGB", float64(25000))
		assert.NoError(t, err)
		checkRelay, err = db.Relay(ctx, rid)
		assert.NoError(t, err)
		assert.Equal(t, int32(25000), checkRelay.IncludedBandwidthGB)

		// relay.MaxBandwidthMbps
		err = db.UpdateRelay(ctx, rid, "MaxBandwidthMbps", float64(19000))
		assert.NoError(t, err)
		checkRelay, err = db.Relay(ctx, rid)
		assert.NoError(t, err)
		assert.Equal(t, int32(19000), checkRelay.MaxBandwidthMbps)

		// relay.Notes
		err = db.UpdateRelay(ctx, rid, "Notes", "not the original notes")
		assert.NoError(t, err)
		checkRelay, err = db.Relay(ctx, rid)
		assert.NoError(t, err)
		assert.Equal(t, "not the original notes", checkRelay.Notes)

		// relay.Notes (null)
		err = db.UpdateRelay(ctx, rid, "Notes", "")
		assert.NoError(t, err)
		checkRelay, err = db.Relay(ctx, rid)
		assert.NoError(t, err)
		assert.Equal(t, "", checkRelay.Notes)

		// relay.BillingSupplier
		err = db.UpdateRelay(ctx, rid, "BillingSupplier", sellerWithID2.ID)
		assert.NoError(t, err)
		checkRelay, err = db.Relay(ctx, rid)
		assert.NoError(t, err)
		assert.Equal(t, sellerWithID2.ID, checkRelay.BillingSupplier)

		// relay.BillingSupplier (null)
		err = db.UpdateRelay(ctx, rid, "BillingSupplier", "")
		assert.NoError(t, err)
		checkRelay, err = db.Relay(ctx, rid)
		assert.NoError(t, err)
		assert.Equal(t, "", checkRelay.BillingSupplier)

		// relay.Version
		err = db.UpdateRelay(ctx, rid, "Version", "")
		assert.Error(t, err)

		err = db.UpdateRelay(ctx, rid, "Version", "7.6.4")
		assert.NoError(t, err)
		checkRelay, err = db.Relay(ctx, rid)
		assert.NoError(t, err)
		assert.Equal(t, "7.6.4", checkRelay.Version)

		// relay.DestFirst
		err = db.UpdateRelay(ctx, rid, "DestFirst", "not a bool")
		assert.Error(t, err)
		err = db.UpdateRelay(ctx, rid, "DestFirst", "")
		assert.Error(t, err)

		err = db.UpdateRelay(ctx, rid, "DestFirst", true)
		assert.NoError(t, err)
		checkRelay, err = db.Relay(ctx, rid)
		assert.NoError(t, err)
		assert.Equal(t, true, checkRelay.DestFirst)

		// relay.InternalAddressClientRoutable
		err = db.UpdateRelay(ctx, rid, "InternalAddressClientRoutable", "not a bool")
		assert.Error(t, err)
		err = db.UpdateRelay(ctx, rid, "InternalAddressClientRoutable", "")
		assert.Error(t, err)

		err = db.UpdateRelay(ctx, rid, "InternalAddr", "")
		assert.NoError(t, err)
		checkRelay, err = db.Relay(ctx, rid)
		assert.NoError(t, err)
		assert.Equal(t, net.UDPAddr{}, checkRelay.InternalAddr)

		err = db.UpdateRelay(ctx, rid, "InternalAddressClientRoutable", true)
		assert.Error(t, err)
		assert.EqualError(t, err, "relay must have valid internal address before InternalAddressClientRoutable is true")
		err = db.UpdateRelay(ctx, rid, "InternalAddr", "192.168.0.2:40000")
		assert.NoError(t, err)

		err = db.UpdateRelay(ctx, rid, "InternalAddressClientRoutable", true)
		assert.NoError(t, err)
		checkRelay, err = db.Relay(ctx, rid)
		assert.NoError(t, err)
		assert.Equal(t, true, checkRelay.InternalAddressClientRoutable)
	})
}

func TestInternalConfig(t *testing.T) {

	SetupEnv()

	ctx := context.Background()
	logger := log.NewNopLogger()

	env, err := backend.GetEnv()
	assert.NoError(t, err)
	db, err := backend.GetStorer(ctx, logger, "local", env)
	assert.NoError(t, err)

	time.Sleep(1000 * time.Millisecond) // allow time for sync functions to complete
	assert.NoError(t, err)

	var outerCustomer routing.Customer
	var outerBuyer routing.Buyer
	var outerInternalConfig core.InternalConfig

	t.Run("AddInternalConfig", func(t *testing.T) {

		customerCode := "compcode"
		customer := routing.Customer{
			Code:                   customerCode,
			Name:                   "Company, Ltd.",
			AutomaticSignInDomains: "fredscuttle.com",
		}

		err = db.AddCustomer(ctx, customer)
		assert.NoError(t, err)

		outerCustomer, err = db.Customer(ctx, customerCode)
		assert.NoError(t, err)

		publicKey := make([]byte, crypto.KeySize)
		_, err := rand.Read(publicKey)
		assert.NoError(t, err)

		internalID := uint64(3142537350691193170)

		buyer := routing.Buyer{
			ID:          internalID,
			ShortName:   outerCustomer.Code,
			CompanyCode: outerCustomer.Code,
			Live:        true,
			Debug:       true,
			PublicKey:   publicKey,
		}

		err = db.AddBuyer(ctx, buyer)
		assert.NoError(t, err)

		outerBuyer, err = db.Buyer(ctx, internalID)
		assert.NoError(t, err)

		internalConfig := core.InternalConfig{
			RouteSelectThreshold:           2,
			RouteSwitchThreshold:           5,
			MaxLatencyTradeOff:             10,
			RTTVeto_Default:                -10,
			RTTVeto_PacketLoss:             -20,
			RTTVeto_Multipath:              -20,
			MultipathOverloadThreshold:     500,
			TryBeforeYouBuy:                true,
			ForceNext:                      true,
			LargeCustomer:                  true,
			Uncommitted:                    true,
			MaxRTT:                         300,
			HighFrequencyPings:             true,
			RouteDiversity:                 10,
			MultipathThreshold:             35,
			EnableVanityMetrics:            true,
			ReducePacketLossMinSliceNumber: 10,
		}

		err = db.AddInternalConfig(ctx, internalConfig, outerBuyer.ID)
		assert.NoError(t, err)

		outerInternalConfig, err = db.InternalConfig(ctx, outerBuyer.ID)
		assert.NoError(t, err)

		assert.Equal(t, int32(2), outerInternalConfig.RouteSelectThreshold)
		assert.Equal(t, int32(5), outerInternalConfig.RouteSwitchThreshold)
		assert.Equal(t, int32(10), outerInternalConfig.MaxLatencyTradeOff)
		assert.Equal(t, int32(-10), outerInternalConfig.RTTVeto_Default)
		assert.Equal(t, int32(-20), outerInternalConfig.RTTVeto_PacketLoss)
		assert.Equal(t, int32(-20), outerInternalConfig.RTTVeto_Multipath)
		assert.Equal(t, int32(500), outerInternalConfig.MultipathOverloadThreshold)
		assert.Equal(t, true, outerInternalConfig.TryBeforeYouBuy)
		assert.Equal(t, true, outerInternalConfig.ForceNext)
		assert.Equal(t, true, outerInternalConfig.LargeCustomer)
		assert.Equal(t, true, outerInternalConfig.Uncommitted)
		assert.Equal(t, true, outerInternalConfig.HighFrequencyPings)
		assert.Equal(t, true, outerInternalConfig.EnableVanityMetrics)
		assert.Equal(t, int32(10), outerInternalConfig.RouteDiversity)
		assert.Equal(t, int32(35), outerInternalConfig.MultipathThreshold)
		assert.Equal(t, int32(300), outerInternalConfig.MaxRTT)
		assert.Equal(t, true, outerInternalConfig.EnableVanityMetrics)
		assert.Equal(t, int32(10), outerInternalConfig.ReducePacketLossMinSliceNumber)
	})

	t.Run("UpdateInternalConfig", func(t *testing.T) {
		// t.Skip() // working on it

		// RouteSelectThreshold
		err = db.UpdateInternalConfig(ctx, outerBuyer.ID, "RouteSelectThreshold", int32(1))
		assert.NoError(t, err)
		checkInternalConfig, err := db.InternalConfig(ctx, outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, int32(1), checkInternalConfig.RouteSelectThreshold)

		// RouteSwitchThreshold
		err = db.UpdateInternalConfig(ctx, outerBuyer.ID, "RouteSwitchThreshold", int32(4))
		assert.NoError(t, err)
		checkInternalConfig, err = db.InternalConfig(ctx, outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, int32(4), checkInternalConfig.RouteSwitchThreshold)

		// MaxLatencyTradeOff
		err = db.UpdateInternalConfig(ctx, outerBuyer.ID, "MaxLatencyTradeOff", int32(11))
		assert.NoError(t, err)
		checkInternalConfig, err = db.InternalConfig(ctx, outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, int32(11), checkInternalConfig.MaxLatencyTradeOff)

		// RTTVeto_Default
		err = db.UpdateInternalConfig(ctx, outerBuyer.ID, "RTTVeto_Default", int32(-20))
		assert.NoError(t, err)
		checkInternalConfig, err = db.InternalConfig(ctx, outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, int32(-20), checkInternalConfig.RTTVeto_Default)

		// RTTVeto_PacketLoss
		err = db.UpdateInternalConfig(ctx, outerBuyer.ID, "RTTVeto_PacketLoss", int32(-30))
		assert.NoError(t, err)
		checkInternalConfig, err = db.InternalConfig(ctx, outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, int32(-30), checkInternalConfig.RTTVeto_PacketLoss)

		// RTTVeto_Multipath
		err = db.UpdateInternalConfig(ctx, outerBuyer.ID, "RTTVeto_Multipath", int32(-40))
		assert.NoError(t, err)
		checkInternalConfig, err = db.InternalConfig(ctx, outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, int32(-40), checkInternalConfig.RTTVeto_Multipath)

		// MultipathOverloadThreshold
		err = db.UpdateInternalConfig(ctx, outerBuyer.ID, "MultipathOverloadThreshold", int32(600))
		assert.NoError(t, err)
		checkInternalConfig, err = db.InternalConfig(ctx, outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, int32(600), checkInternalConfig.MultipathOverloadThreshold)

		// TryBeforeYouBuy
		err = db.UpdateInternalConfig(ctx, outerBuyer.ID, "TryBeforeYouBuy", false)
		assert.NoError(t, err)
		checkInternalConfig, err = db.InternalConfig(ctx, outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, false, checkInternalConfig.TryBeforeYouBuy)

		// ForceNext
		err = db.UpdateInternalConfig(ctx, outerBuyer.ID, "ForceNext", false)
		assert.NoError(t, err)
		checkInternalConfig, err = db.InternalConfig(ctx, outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, false, checkInternalConfig.ForceNext)

		// LargeCustomer
		err = db.UpdateInternalConfig(ctx, outerBuyer.ID, "LargeCustomer", false)
		assert.NoError(t, err)
		checkInternalConfig, err = db.InternalConfig(ctx, outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, false, checkInternalConfig.LargeCustomer)

		// Uncommitted
		err = db.UpdateInternalConfig(ctx, outerBuyer.ID, "Uncommitted", false)
		assert.NoError(t, err)
		checkInternalConfig, err = db.InternalConfig(ctx, outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, false, checkInternalConfig.Uncommitted)

		// MaxRTT
		err = db.UpdateInternalConfig(ctx, outerBuyer.ID, "MaxRTT", int32(400))
		assert.NoError(t, err)
		checkInternalConfig, err = db.InternalConfig(ctx, outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, int32(400), checkInternalConfig.MaxRTT)

		// HighFrequencyPings
		err = db.UpdateInternalConfig(ctx, outerBuyer.ID, "HighFrequencyPings", false)
		assert.NoError(t, err)
		checkInternalConfig, err = db.InternalConfig(ctx, outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, false, checkInternalConfig.HighFrequencyPings)

		// RouteDiversity
		err = db.UpdateInternalConfig(ctx, outerBuyer.ID, "RouteDiversity", int32(40))
		assert.NoError(t, err)
		checkInternalConfig, err = db.InternalConfig(ctx, outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, int32(40), checkInternalConfig.RouteDiversity)

		// MultipathThreshold
		err = db.UpdateInternalConfig(ctx, outerBuyer.ID, "MultipathThreshold", int32(50))
		assert.NoError(t, err)
		checkInternalConfig, err = db.InternalConfig(ctx, outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, int32(50), checkInternalConfig.MultipathThreshold)

		// EnableVanityMetrics
		err = db.UpdateInternalConfig(ctx, outerBuyer.ID, "EnableVanityMetrics", false)
		assert.NoError(t, err)
		checkInternalConfig, err = db.InternalConfig(ctx, outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, false, checkInternalConfig.EnableVanityMetrics)

		// EnableVanityMetrics
		err = db.UpdateInternalConfig(ctx, outerBuyer.ID, "EnableVanityMetrics", false)
		assert.NoError(t, err)
		checkInternalConfig, err = db.InternalConfig(ctx, outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, false, checkInternalConfig.EnableVanityMetrics)

		// ReducePacketLossMinSliceNumber
		err = db.UpdateInternalConfig(ctx, outerBuyer.ID, "ReducePacketLossMinSliceNumber", int32(50))
		assert.NoError(t, err)
		checkInternalConfig, err = db.InternalConfig(ctx, outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, int32(50), checkInternalConfig.ReducePacketLossMinSliceNumber)

	})

	t.Run("RemoveInternalConfig", func(t *testing.T) {

		err := db.RemoveInternalConfig(context.Background(), outerBuyer.ID)
		assert.NoError(t, err)

		_, err = db.InternalConfig(ctx, outerBuyer.ID)
		assert.Error(t, err)

	})

}

func TestRouteShaders(t *testing.T) {

	SetupEnv()

	ctx := context.Background()
	logger := log.NewNopLogger()

	// db, err := storage.NewSQLStorage(ctx, logger)
	env, err := backend.GetEnv()
	assert.NoError(t, err)
	db, err := backend.GetStorer(ctx, logger, "local", env)
	assert.NoError(t, err)

	time.Sleep(1000 * time.Millisecond) // allow time for sync functions to complete
	assert.NoError(t, err)

	var outerCustomer routing.Customer
	var outerBuyer routing.Buyer
	var outerRouteShader core.RouteShader

	t.Run("AddRouteShader", func(t *testing.T) {

		customerCode := "compcode"
		customer := routing.Customer{
			Code:                   customerCode,
			Name:                   "Company, Ltd.",
			AutomaticSignInDomains: "fredscuttle.com",
		}

		err = db.AddCustomer(ctx, customer)
		assert.NoError(t, err)

		outerCustomer, err = db.Customer(ctx, customerCode)
		assert.NoError(t, err)

		publicKey := make([]byte, crypto.KeySize)
		_, err := rand.Read(publicKey)
		assert.NoError(t, err)

		internalID := uint64(3142537350691193170)

		buyer := routing.Buyer{
			ID:          internalID,
			ShortName:   outerCustomer.Code,
			CompanyCode: outerCustomer.Code,
			Live:        true,
			Debug:       true,
			PublicKey:   publicKey,
		}

		err = db.AddBuyer(ctx, buyer)
		assert.NoError(t, err)

		outerBuyer, err = db.Buyer(ctx, internalID)
		assert.NoError(t, err)

		routeShader := core.RouteShader{
			ABTest:                    true,
			AcceptableLatency:         int32(25),
			AcceptablePacketLoss:      float32(1),
			AnalysisOnly:              true,
			BandwidthEnvelopeDownKbps: int32(1200),
			BandwidthEnvelopeUpKbps:   int32(500),
			DisableNetworkNext:        true,
			LatencyThreshold:          int32(5),
			Multipath:                 true,
			ProMode:                   true,
			ReduceLatency:             true,
			ReducePacketLoss:          true,
			ReduceJitter:              true,
			SelectionPercent:          int(100),
			PacketLossSustained:       float32(10),
		}

		err = db.AddRouteShader(ctx, routeShader, outerBuyer.ID)
		assert.NoError(t, err)

		outerRouteShader, err = db.RouteShader(ctx, outerBuyer.ID)
		assert.NoError(t, err)

		assert.Equal(t, true, outerRouteShader.ABTest)
		assert.Equal(t, int32(25), outerRouteShader.AcceptableLatency)
		assert.Equal(t, float32(1), outerRouteShader.AcceptablePacketLoss)
		assert.Equal(t, true, outerRouteShader.AnalysisOnly)
		assert.Equal(t, int32(1200), outerRouteShader.BandwidthEnvelopeDownKbps)
		assert.Equal(t, int32(500), outerRouteShader.BandwidthEnvelopeUpKbps)
		assert.Equal(t, true, outerRouteShader.DisableNetworkNext)
		assert.Equal(t, int32(5), outerRouteShader.LatencyThreshold)
		assert.Equal(t, true, outerRouteShader.Multipath)
		assert.Equal(t, true, outerRouteShader.ProMode)
		assert.Equal(t, true, outerRouteShader.ReduceLatency)
		assert.Equal(t, true, outerRouteShader.ReducePacketLoss)
		assert.Equal(t, true, outerRouteShader.ReduceJitter)
		assert.Equal(t, int(100), outerRouteShader.SelectionPercent)
		assert.Equal(t, float32(10), outerRouteShader.PacketLossSustained)
	})

	t.Run("UpdateRouteShader", func(t *testing.T) {

		time.Sleep(1000 * time.Millisecond) // allow time for sync functions to complete

		// ABTest
		err = db.UpdateRouteShader(ctx, outerBuyer.ID, "ABTest", false)
		assert.NoError(t, err)
		checkRouteShader, err := db.RouteShader(ctx, outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, false, checkRouteShader.ABTest)

		// AcceptableLatency
		err = db.UpdateRouteShader(ctx, outerBuyer.ID, "AcceptableLatency", int32(35))
		assert.NoError(t, err)
		checkRouteShader, err = db.RouteShader(ctx, outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, int32(35), checkRouteShader.AcceptableLatency)

		// AcceptablePacketLoss
		err = db.UpdateRouteShader(ctx, outerBuyer.ID, "AcceptablePacketLoss", float32(10))
		assert.NoError(t, err)
		checkRouteShader, err = db.RouteShader(ctx, outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, float32(10), checkRouteShader.AcceptablePacketLoss)

		// AnalysisOnly
		err = db.UpdateRouteShader(ctx, outerBuyer.ID, "AnalysisOnly", false)
		assert.NoError(t, err)
		checkRouteShader, err = db.RouteShader(ctx, outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, false, checkRouteShader.AnalysisOnly)

		// BandwidthEnvelopeDownKbps
		err = db.UpdateRouteShader(ctx, outerBuyer.ID, "BandwidthEnvelopeDownKbps", int32(1000))
		assert.NoError(t, err)
		checkRouteShader, err = db.RouteShader(ctx, outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, int32(1000), checkRouteShader.BandwidthEnvelopeDownKbps)

		// BandwidthEnvelopeUpKbps
		err = db.UpdateRouteShader(ctx, outerBuyer.ID, "BandwidthEnvelopeUpKbps", int32(400))
		assert.NoError(t, err)
		checkRouteShader, err = db.RouteShader(ctx, outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, int32(400), checkRouteShader.BandwidthEnvelopeUpKbps)

		// DisableNetworkNext
		err = db.UpdateRouteShader(ctx, outerBuyer.ID, "DisableNetworkNext", false)
		assert.NoError(t, err)
		checkRouteShader, err = db.RouteShader(ctx, outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, false, checkRouteShader.DisableNetworkNext)

		// LatencyThreshold
		err = db.UpdateRouteShader(ctx, outerBuyer.ID, "LatencyThreshold", int32(15))
		assert.NoError(t, err)
		checkRouteShader, err = db.RouteShader(ctx, outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, int32(15), checkRouteShader.LatencyThreshold)

		// Multipath
		err = db.UpdateRouteShader(ctx, outerBuyer.ID, "Multipath", false)
		assert.NoError(t, err)
		checkRouteShader, err = db.RouteShader(ctx, outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, false, checkRouteShader.Multipath)

		// ProMode
		err = db.UpdateRouteShader(ctx, outerBuyer.ID, "ProMode", false)
		assert.NoError(t, err)
		checkRouteShader, err = db.RouteShader(ctx, outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, false, checkRouteShader.ProMode)

		// ReduceLatency
		err = db.UpdateRouteShader(ctx, outerBuyer.ID, "ReduceLatency", false)
		assert.NoError(t, err)
		checkRouteShader, err = db.RouteShader(ctx, outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, false, checkRouteShader.ReduceLatency)

		// ReducePacketLoss
		err = db.UpdateRouteShader(ctx, outerBuyer.ID, "ReducePacketLoss", false)
		assert.NoError(t, err)
		checkRouteShader, err = db.RouteShader(ctx, outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, false, checkRouteShader.ReducePacketLoss)

		// ReduceJitter
		err = db.UpdateRouteShader(ctx, outerBuyer.ID, "ReduceJitter", false)
		assert.NoError(t, err)
		checkRouteShader, err = db.RouteShader(ctx, outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, false, checkRouteShader.ReduceJitter)

		// SelectionPercent
		err = db.UpdateRouteShader(ctx, outerBuyer.ID, "SelectionPercent", int(90))
		assert.NoError(t, err)
		checkRouteShader, err = db.RouteShader(ctx, outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, int(90), checkRouteShader.SelectionPercent)

		// PacketLossSustained
		err = db.UpdateRouteShader(ctx, outerBuyer.ID, "PacketLossSustained", float32(10))
		assert.NoError(t, err)
		checkRouteShader, err = db.RouteShader(ctx, outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, float32(10), checkRouteShader.PacketLossSustained)

	})

	t.Run("BannedUser tests", func(t *testing.T) {
		// random user IDs scraped from the portal
		userID1, err := strconv.ParseUint("77c556007df7c02e", 16, 64)
		assert.NoError(t, err)
		userID2, err := strconv.ParseUint("a731e14c521514a4", 16, 64)
		assert.NoError(t, err)
		userID3, err := strconv.ParseUint("fb6fa90ad67bc76a", 16, 64)
		assert.NoError(t, err)

		err = db.AddBannedUser(ctx, outerBuyer.ID, userID1)
		assert.NoError(t, err)
		err = db.AddBannedUser(ctx, outerBuyer.ID, userID2)
		assert.NoError(t, err)
		err = db.AddBannedUser(ctx, outerBuyer.ID, userID3)
		assert.NoError(t, err)

		bannedUserList, err := db.BannedUsers(ctx, outerBuyer.ID)
		assert.NoError(t, err)

		assert.True(t, bannedUserList[userID1])
		assert.True(t, bannedUserList[userID2])
		assert.True(t, bannedUserList[userID3])

		checkRouteShader, err := db.RouteShader(ctx, outerBuyer.ID)
		assert.NoError(t, err)
		assert.True(t, len(checkRouteShader.BannedUsers) > 0)
		assert.True(t, checkRouteShader.BannedUsers[userID1])
		assert.True(t, checkRouteShader.BannedUsers[userID2])
		assert.True(t, checkRouteShader.BannedUsers[userID3])

		err = db.RemoveBannedUser(ctx, outerBuyer.ID, userID1)
		assert.NoError(t, err)

		bannedUserList2, err := db.BannedUsers(ctx, outerBuyer.ID)
		assert.NoError(t, err)
		assert.False(t, bannedUserList2[userID1])

	})

	t.Run("RemoveRouteShader", func(t *testing.T) {
		err := db.RemoveRouteShader(context.Background(), outerBuyer.ID)
		assert.NoError(t, err)

		_, err = db.RouteShader(ctx, outerBuyer.ID)
		assert.Error(t, err)

	})

}

func TestDatabaseBinMetaData(t *testing.T) {

	SetupEnv()

	ctx := context.Background()
	logger := log.NewNopLogger()

	// db, err := storage.NewSQLStorage(ctx, logger)
	env, err := backend.GetEnv()
	assert.NoError(t, err)
	db, err := backend.GetStorer(ctx, logger, "local", env)
	assert.NoError(t, err)

	time.Sleep(1000 * time.Millisecond) // allow time for sync functions to complete
	assert.NoError(t, err)

	t.Run("AddDatabaseBinMetaData", func(t *testing.T) {

		testTime := time.Now()
		ctx := context.Background()

		metaData := routing.DatabaseBinFileMetaData{
			DatabaseBinFileAuthor:       "Arthur Dent",
			DatabaseBinFileCreationTime: testTime,
		}

		err = db.UpdateDatabaseBinFileMetaData(ctx, metaData)
		assert.NoError(t, err)

		checkMetaData, err := db.GetDatabaseBinFileMetaData(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "Arthur Dent", checkMetaData.DatabaseBinFileAuthor)
		assert.Equal(t, testTime.Format("01/02/06"), checkMetaData.DatabaseBinFileCreationTime.Format("01/02/06"))

		// should only return the most recent record
		testTime2 := time.Now()
		metaData2 := routing.DatabaseBinFileMetaData{
			DatabaseBinFileAuthor:       "Brian Cohen",
			DatabaseBinFileCreationTime: testTime2,
		}

		err = db.UpdateDatabaseBinFileMetaData(ctx, metaData2)
		assert.NoError(t, err)

		checkMetaData2, err := db.GetDatabaseBinFileMetaData(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "Brian Cohen", checkMetaData2.DatabaseBinFileAuthor)
		assert.Equal(t, testTime2.Format("01/02/06"), checkMetaData2.DatabaseBinFileCreationTime.Format("01/02/06"))

	})
}

func TestAnalyticsDashboards(t *testing.T) {
	SetupEnv()

	ctx := context.Background()
	logger := log.NewNopLogger()

	env, err := backend.GetEnv()
	assert.NoError(t, err)
	db, err := backend.GetStorer(ctx, logger, "local", env)
	assert.NoError(t, err)

	time.Sleep(1000 * time.Millisecond) // allow time for sync functions to complete
	assert.NoError(t, err)

	dashboards, err := db.GetAnalyticsDashboards(ctx)
	assert.NoError(t, err)
	for _, dashboard := range dashboards {
		err := db.RemoveAnalyticsDashboardByID(ctx, dashboard.ID)
		assert.NoError(t, err)
	}

	categories, err := db.GetAnalyticsDashboardCategories(ctx)
	assert.NoError(t, err)
	for _, category := range categories {
		subCategories, err := db.GetAnalyticsDashboardSubCategoriesByCategoryID(ctx, category.ID)
		assert.NoError(t, err)

		for _, category := range subCategories {
			err := db.RemoveAnalyticsDashboardCategoryByID(ctx, category.ID)
			assert.NoError(t, err)
		}

		err = db.RemoveAnalyticsDashboardCategoryByID(ctx, category.ID)
		assert.NoError(t, err)
	}

	err = db.AddCustomer(ctx, routing.Customer{
		Code: "test-company",
		Name: "Test Company",
	})
	assert.NoError(t, err)

	err = db.AddCustomer(ctx, routing.Customer{
		Code: "another-test-company",
		Name: "Another Test Company",
	})
	assert.NoError(t, err)

	customers := db.Customers(ctx)

	t.Run("AddAnalyticsDashboardCategory", func(t *testing.T) {
		category := looker.AnalyticsDashboardCategory{
			Label: "Test Category",
			Order: 10,
		}

		err := db.AddAnalyticsDashboardCategory(ctx, category.Order, category.Label, -1)
		assert.NoError(t, err)

		dashboardCategories, err := db.GetAnalyticsDashboardCategories(ctx)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(dashboardCategories))
		assert.Equal(t, category.Label, dashboardCategories[0].Label)

		category2 := looker.AnalyticsDashboardCategory{
			Label: "Another Test Category",
			Order: 5,
		}

		err = db.AddAnalyticsDashboardCategory(ctx, category2.Order, category2.Label, -1)
		assert.NoError(t, err)

		dashboardCategories, err = db.GetAnalyticsDashboardCategories(ctx)
		assert.NoError(t, err)

		assert.Equal(t, 2, len(dashboardCategories))
		assert.Equal(t, category.Label, dashboardCategories[0].Label)
		assert.Equal(t, category2.Label, dashboardCategories[1].Label)

		dbCategory, err := db.GetAnalyticsDashboardCategoryByID(ctx, dashboardCategories[0].ID)
		assert.NoError(t, err)
		assert.Equal(t, dashboardCategories[0].ID, dbCategory.ID)
		assert.Equal(t, dashboardCategories[0].Label, dbCategory.Label)

		dbCategory, err = db.GetAnalyticsDashboardCategoryByLabel(ctx, dashboardCategories[0].Label)
		assert.NoError(t, err)
		assert.Equal(t, dashboardCategories[0].ID, dbCategory.ID)
		assert.Equal(t, dashboardCategories[0].Label, dbCategory.Label)
	})

	t.Run("AddAnalyticsDashboard", func(t *testing.T) {
		dashboard := looker.AnalyticsDashboard{
			Name:     "Test Dashboard",
			LookerID: 10,
			Order:    10,
		}

		dashboardCategories, err := db.GetAnalyticsDashboardCategories(ctx)
		assert.NoError(t, err)

		err = db.AddAnalyticsDashboard(ctx, dashboard.Order, dashboard.Name, false, true, dashboard.LookerID, customers[0].DatabaseID, dashboardCategories[0].ID)
		assert.NoError(t, err)

		dashboards, err := db.GetAnalyticsDashboards(ctx)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(dashboards))
		assert.Equal(t, dashboard.Name, dashboards[0].Name)
		assert.Equal(t, dashboard.Order, dashboards[0].Order)
		assert.Equal(t, dashboard.LookerID, dashboards[0].LookerID)
		assert.Equal(t, customers[0].Code, dashboards[0].CustomerCode)
		assert.Equal(t, dashboardCategories[0].ID, dashboards[0].Category.ID)
		assert.Equal(t, dashboardCategories[0].Label, dashboards[0].Category.Label)

		dashboard2 := looker.AnalyticsDashboard{
			Name:     "Another Test Dashboard",
			LookerID: 5,
			Order:    5,
		}

		err = db.AddAnalyticsDashboard(ctx, dashboard2.Order, dashboard2.Name, false, false, dashboard2.LookerID, customers[0].DatabaseID, dashboardCategories[1].ID)
		assert.NoError(t, err)

		dashboards, err = db.GetAnalyticsDashboards(ctx)
		assert.NoError(t, err)

		assert.Equal(t, 2, len(dashboards))
		assert.Equal(t, dashboard.Name, dashboards[0].Name)
		assert.Equal(t, dashboard.Order, dashboards[0].Order)
		assert.Equal(t, dashboard.LookerID, dashboards[0].LookerID)
		assert.Equal(t, customers[0].Code, dashboards[0].CustomerCode)
		assert.Equal(t, dashboardCategories[0].ID, dashboards[0].Category.ID)
		assert.Equal(t, dashboardCategories[0].Label, dashboards[0].Category.Label)

		assert.Equal(t, dashboard2.Name, dashboards[1].Name)
		assert.Equal(t, dashboard2.Order, dashboards[1].Order)
		assert.Equal(t, dashboard2.LookerID, dashboards[1].LookerID)
		assert.Equal(t, customers[0].Code, dashboards[1].CustomerCode)
		assert.Equal(t, dashboardCategories[1].ID, dashboards[1].Category.ID)
		assert.Equal(t, dashboardCategories[1].Label, dashboards[1].Category.Label)

		dashboards, err = db.GetFreeAnalyticsDashboards(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(dashboards))

		dashboards, err = db.GetPremiumAnalyticsDashboards(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(dashboards))

		dashboards, err = db.GetAnalyticsDashboardsByCategoryID(ctx, dashboardCategories[0].ID)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(dashboards))
		assert.Equal(t, dashboard.Name, dashboards[0].Name)
		assert.Equal(t, dashboard.LookerID, dashboards[0].LookerID)
		assert.Equal(t, customers[0].Code, dashboards[0].CustomerCode)
		assert.Equal(t, dashboardCategories[0].ID, dashboards[0].Category.ID)
		assert.Equal(t, dashboardCategories[0].Label, dashboards[0].Category.Label)

		dashboards, err = db.GetAnalyticsDashboardsByCategoryID(ctx, dashboardCategories[1].ID)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(dashboards))

		assert.Equal(t, dashboard2.Name, dashboards[0].Name)
		assert.Equal(t, dashboard2.LookerID, dashboards[0].LookerID)
		assert.Equal(t, customers[0].Code, dashboards[0].CustomerCode)
		assert.Equal(t, dashboardCategories[1].ID, dashboards[0].Category.ID)
		assert.Equal(t, dashboardCategories[1].Label, dashboards[0].Category.Label)

		dashboards, err = db.GetAnalyticsDashboardsByCategoryLabel(ctx, dashboardCategories[1].Label)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(dashboards))

		dashboards, err = db.GetAnalyticsDashboardsByCategoryLabel(ctx, dashboardCategories[0].Label)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(dashboards))

		dbDashboard, err := db.GetAnalyticsDashboardByID(ctx, dashboards[0].ID)
		assert.NoError(t, err)

		assert.Equal(t, dashboard.Name, dbDashboard.Name)
		assert.Equal(t, dashboard.LookerID, dbDashboard.LookerID)
		assert.Equal(t, customers[0].Code, dbDashboard.CustomerCode)
		assert.Equal(t, dashboardCategories[0].ID, dbDashboard.Category.ID)
		assert.Equal(t, dashboardCategories[0].Label, dbDashboard.Category.Label)

		dbDashboard, err = db.GetAnalyticsDashboardByName(ctx, dashboard.Name)
		assert.NoError(t, err)

		assert.Equal(t, dashboard.Name, dbDashboard.Name)
		assert.Equal(t, dashboard.LookerID, dbDashboard.LookerID)
		assert.Equal(t, customers[0].Code, dbDashboard.CustomerCode)
		assert.Equal(t, dashboardCategories[0].ID, dbDashboard.Category.ID)
		assert.Equal(t, dashboardCategories[0].Label, dbDashboard.Category.Label)
	})

	t.Run("RemoveAnalyticsDashboards", func(t *testing.T) {

		dashboards, err := db.GetAnalyticsDashboards(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(dashboards))

		removedDashboard := dashboards[0]
		removedDashboardCustomer, err := db.Customer(ctx, removedDashboard.CustomerCode)
		assert.NoError(t, err)

		err = db.RemoveAnalyticsDashboardByID(ctx, dashboards[0].ID)
		assert.NoError(t, err)

		dashboards, err = db.GetAnalyticsDashboards(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(dashboards))

		assert.NotEqual(t, removedDashboard.ID, dashboards[0].ID)
		assert.NotEqual(t, removedDashboard.Name, dashboards[0].Name)
		assert.NotEqual(t, removedDashboard.LookerID, dashboards[0].LookerID)
		assert.NotEqual(t, removedDashboard.Category.ID, dashboards[0].Category.ID)
		assert.NotEqual(t, removedDashboard.Category.Label, dashboards[0].Category.Label)

		err = db.AddAnalyticsDashboard(ctx, 0, removedDashboard.Name, false, false, removedDashboard.LookerID, removedDashboardCustomer.DatabaseID, removedDashboard.Category.ID)
		assert.NoError(t, err)

		dashboards, err = db.GetAnalyticsDashboards(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(dashboards))

		err = db.RemoveAnalyticsDashboardByName(ctx, dashboards[1].Name)
		assert.NoError(t, err)

		dashboards, err = db.GetAnalyticsDashboards(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(dashboards))

		assert.NotEqual(t, removedDashboard.ID, dashboards[0].ID)
		assert.NotEqual(t, removedDashboard.Name, dashboards[0].Name)
		assert.NotEqual(t, removedDashboard.LookerID, dashboards[0].LookerID)
		assert.NotEqual(t, removedDashboard.Category.ID, dashboards[0].Category.ID)
		assert.NotEqual(t, removedDashboard.Category.Label, dashboards[0].Category.Label)

	})

	t.Run("RemoveAnalyticsCategories", func(t *testing.T) {

		categories, err := db.GetAnalyticsDashboardCategories(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(categories))

		dashboards, err := db.GetAnalyticsDashboards(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(dashboards))

		assert.Equal(t, dashboards[0].Category.ID, categories[1].ID)
		assert.Equal(t, dashboards[0].Category.Label, categories[1].Label)

		err = db.RemoveAnalyticsDashboardCategoryByID(ctx, dashboards[0].Category.ID)
		assert.Error(t, err)

		removedCategory := categories[0]

		err = db.RemoveAnalyticsDashboardCategoryByID(ctx, removedCategory.ID)
		assert.NoError(t, err)

		categories, err = db.GetAnalyticsDashboardCategories(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(categories))

		assert.NotEqual(t, categories[0].ID, removedCategory.ID)
		assert.NotEqual(t, categories[0].Label, removedCategory.Label)

		err = db.AddAnalyticsDashboardCategory(ctx, 0, removedCategory.Label, -1)
		assert.NoError(t, err)

		categories, err = db.GetAnalyticsDashboardCategories(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(categories))

		err = db.RemoveAnalyticsDashboardCategoryByLabel(ctx, dashboards[0].Category.Label)
		assert.Error(t, err)

		err = db.RemoveAnalyticsDashboardCategoryByLabel(ctx, removedCategory.Label)
		assert.NoError(t, err)

		categories, err = db.GetAnalyticsDashboardCategories(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(categories))

		assert.NotEqual(t, categories[0].ID, removedCategory.ID)
		assert.NotEqual(t, categories[0].Label, removedCategory.Label)
	})

	t.Run("UpdateAnalyticsDashboards", func(t *testing.T) {

		dashboards, err := db.GetAnalyticsDashboards(ctx)
		assert.NoError(t, err)

		updatedDashboard := dashboards[0]

		assert.NotEqual(t, int64(120), updatedDashboard.LookerID)

		err = db.UpdateAnalyticsDashboardByID(ctx, updatedDashboard.ID, "LookerID", int64(120))
		assert.NoError(t, err)

		dashboard, err := db.GetAnalyticsDashboardByID(ctx, updatedDashboard.ID)
		assert.NoError(t, err)

		assert.Equal(t, int64(120), dashboard.LookerID)

		err = db.UpdateAnalyticsDashboardByID(ctx, updatedDashboard.ID, "Name", "This is the new dashboard name")
		assert.NoError(t, err)

		dashboard, err = db.GetAnalyticsDashboardByID(ctx, updatedDashboard.ID)
		assert.NoError(t, err)

		assert.Equal(t, "This is the new dashboard name", dashboard.Name)

		oldOrder := dashboard.Order

		err = db.UpdateAnalyticsDashboardByID(ctx, updatedDashboard.ID, "Order", oldOrder+10)
		assert.NoError(t, err)

		dashboard, err = db.GetAnalyticsDashboardByID(ctx, updatedDashboard.ID)
		assert.NoError(t, err)

		assert.Equal(t, oldOrder+10, dashboard.Order)

		customers := db.Customers(ctx)

		err = db.UpdateAnalyticsDashboardByID(ctx, updatedDashboard.ID, "CustomerCode", customers[1].Code)
		assert.NoError(t, err)

		dashboard, err = db.GetAnalyticsDashboardByID(ctx, updatedDashboard.ID)
		assert.NoError(t, err)

		assert.Equal(t, customers[1].Code, dashboard.CustomerCode)

		err = db.AddAnalyticsDashboardCategory(ctx, 0, "My Test Category", -1)
		assert.NoError(t, err)

		categories, err := db.GetAnalyticsDashboardCategories(ctx)
		assert.NoError(t, err)

		err = db.UpdateAnalyticsDashboardByID(ctx, updatedDashboard.ID, "Category", categories[1].ID)
		assert.NoError(t, err)

		dashboard, err = db.GetAnalyticsDashboardByID(ctx, updatedDashboard.ID)
		assert.NoError(t, err)

		assert.Equal(t, categories[1].ID, dashboard.Category.ID)
		assert.Equal(t, categories[1].Label, dashboard.Category.Label)

		err = db.UpdateAnalyticsDashboardByID(ctx, updatedDashboard.ID, "Name", "")
		assert.Error(t, err)

		err = db.UpdateAnalyticsDashboardByID(ctx, updatedDashboard.ID, "Name", nil)
		assert.Error(t, err)

		err = db.UpdateAnalyticsDashboardByID(ctx, updatedDashboard.ID, "Order", "")
		assert.Error(t, err)

		err = db.UpdateAnalyticsDashboardByID(ctx, updatedDashboard.ID, "Order", nil)
		assert.Error(t, err)

		err = db.UpdateAnalyticsDashboardByID(ctx, updatedDashboard.ID, "LookerID", "")
		assert.Error(t, err)

		err = db.UpdateAnalyticsDashboardByID(ctx, updatedDashboard.ID, "LookerID", nil)
		assert.Error(t, err)

		err = db.UpdateAnalyticsDashboardByID(ctx, updatedDashboard.ID, "CompanyCode", "")
		assert.Error(t, err)

		err = db.UpdateAnalyticsDashboardByID(ctx, updatedDashboard.ID, "CompanyCode", nil)
		assert.Error(t, err)

		err = db.UpdateAnalyticsDashboardByID(ctx, updatedDashboard.ID, "CompanyCode", "not a valid customer code")
		assert.Error(t, err)

		err = db.UpdateAnalyticsDashboardByID(ctx, updatedDashboard.ID, "Category", nil)
		assert.Error(t, err)

		err = db.UpdateAnalyticsDashboardByID(ctx, updatedDashboard.ID, "Category", int64(12343523452))
		assert.Error(t, err)
	})

	t.Run("UpdateAnalyticsCategories", func(t *testing.T) {

		categories, err := db.GetAnalyticsDashboardCategories(ctx)
		assert.NoError(t, err)

		updatedCategory := categories[0]

		assert.NotEqual(t, "New Category Label", updatedCategory.Label)

		err = db.UpdateAnalyticsDashboardCategoryByID(ctx, updatedCategory.ID, "Label", "New Category Label")
		assert.NoError(t, err)

		category, err := db.GetAnalyticsDashboardCategoryByID(ctx, updatedCategory.ID)
		assert.NoError(t, err)

		assert.Equal(t, "New Category Label", category.Label)

		oldOrder := category.Order

		err = db.UpdateAnalyticsDashboardCategoryByID(ctx, updatedCategory.ID, "Order", oldOrder+10)
		assert.NoError(t, err)

		category, err = db.GetAnalyticsDashboardCategoryByID(ctx, updatedCategory.ID)
		assert.NoError(t, err)

		assert.Equal(t, oldOrder+10, category.Order)

		err = db.UpdateAnalyticsDashboardCategoryByID(ctx, updatedCategory.ID, "Label", "")
		assert.Error(t, err)

		err = db.UpdateAnalyticsDashboardCategoryByID(ctx, updatedCategory.ID, "Label", nil)
		assert.Error(t, err)

		err = db.UpdateAnalyticsDashboardCategoryByID(ctx, updatedCategory.ID, "Order", "")
		assert.Error(t, err)

		err = db.UpdateAnalyticsDashboardCategoryByID(ctx, updatedCategory.ID, "Order", nil)
		assert.Error(t, err)
	})
}
