package jsonrpc_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/storage"
	"github.com/networknext/backend/modules/transport/jsonrpc"
	"github.com/networknext/backend/modules/transport/middleware"
	"github.com/stretchr/testify/assert"
)

func TestBuyers(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}

	err := storer.AddBuyer(context.Background(), routing.Buyer{ID: 1, CompanyCode: "local", ShortName: "local.1"})
	assert.NoError(t, err)
	err = storer.AddBuyer(context.Background(), routing.Buyer{ID: 2, CompanyCode: "local-local", ShortName: "local.local.2"})
	assert.NoError(t, err)

	svc := jsonrpc.OpsService{
		Storage: &storer,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("insufficient privileges", func(t *testing.T) {
		var reply jsonrpc.BuyersReply
		err := svc.Buyers(req, &jsonrpc.BuyersArgs{}, &reply)
		assert.Equal(t, jsonrpc.JSONRPCErrorCodes[int(jsonrpc.ERROR_INSUFFICIENT_PRIVILEGES)].Message, err.Error())
	})

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
		"Admin",
	})
	req = req.WithContext(reqContext)

	t.Run("no customer for buyer", func(t *testing.T) {
		var reply jsonrpc.BuyersReply
		err := svc.Buyers(req, &jsonrpc.BuyersArgs{}, &reply)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Buyers() could not find Customer")
	})

	err = storer.AddCustomer(context.Background(), routing.Customer{Name: "Local", Code: "local"})
	assert.NoError(t, err)
	err = storer.AddCustomer(context.Background(), routing.Customer{Name: "Local-Local", Code: "local-local"})
	assert.NoError(t, err)

	t.Run("success - list", func(t *testing.T) {
		var reply jsonrpc.BuyersReply
		err := svc.Buyers(req, &jsonrpc.BuyersArgs{}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, reply.Buyers[0].ID, uint64(1))
		assert.Equal(t, reply.Buyers[0].ShortName, "local.1")
		assert.Equal(t, reply.Buyers[1].ID, uint64(2))
		assert.Equal(t, reply.Buyers[1].ShortName, "local.local.2")
	})
}

func TestJSAddBuyer(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}

	svc := jsonrpc.OpsService{
		Storage: &storer,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("insufficient privileges", func(t *testing.T) {
		var reply jsonrpc.JSAddBuyerReply
		err := svc.JSAddBuyer(req, &jsonrpc.JSAddBuyerArgs{}, &reply)
		assert.Equal(t, jsonrpc.JSONRPCErrorCodes[int(jsonrpc.ERROR_INSUFFICIENT_PRIVILEGES)].Message, err.Error())
	})

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
		"Admin",
	})
	req = req.WithContext(reqContext)

	t.Run("bad public key", func(t *testing.T) {
		var reply jsonrpc.JSAddBuyerReply
		err := svc.JSAddBuyer(req, &jsonrpc.JSAddBuyerArgs{
			ShortName:           "local",
			ExoticLocationFee:   "100",
			StandardLocationFee: "100",
			PublicKey:           "KcZ+NlIAkrMfc9ir79ZMGJxLnPEDuHkf6Yi0akyyWWcR3JaMY+yp2A=",
		}, &reply)
		assert.Error(t, err)
	})

	t.Run("bad exotic location fee", func(t *testing.T) {
		var reply jsonrpc.JSAddBuyerReply
		err := svc.JSAddBuyer(req, &jsonrpc.JSAddBuyerArgs{
			ShortName:           "local",
			ExoticLocationFee:   "a",
			StandardLocationFee: "100",
			PublicKey:           "KcZ+NlIAkrMfc9ir79ZMGJxLnPEDuHkf6Yi0akyyWWcR3JaMY+yp2A==",
		}, &reply)
		assert.Error(t, err)
	})

	t.Run("bad standard location fee", func(t *testing.T) {
		var reply jsonrpc.JSAddBuyerReply
		err := svc.JSAddBuyer(req, &jsonrpc.JSAddBuyerArgs{
			ShortName:           "local",
			ExoticLocationFee:   "100",
			StandardLocationFee: "a",
			PublicKey:           "KcZ+NlIAkrMfc9ir79ZMGJxLnPEDuHkf6Yi0akyyWWcR3JaMY+yp2A==",
		}, &reply)
		assert.Error(t, err)
	})

	t.Run("bad looker seats", func(t *testing.T) {
		var reply jsonrpc.JSAddBuyerReply
		err := svc.JSAddBuyer(req, &jsonrpc.JSAddBuyerArgs{
			ShortName:           "local",
			ExoticLocationFee:   "100",
			StandardLocationFee: "100",
			LookerSeats:         "a",
			PublicKey:           "KcZ+NlIAkrMfc9ir79ZMGJxLnPEDuHkf6Yi0akyyWWcR3JaMY+yp2A==",
		}, &reply)
		assert.Error(t, err)
	})

	t.Run("add buyer for non-existant customer", func(t *testing.T) {
		var reply jsonrpc.JSAddBuyerReply
		err := svc.JSAddBuyer(req, &jsonrpc.JSAddBuyerArgs{
			ShortName:           "local",
			ExoticLocationFee:   "100",
			StandardLocationFee: "100",
			PublicKey:           "KcZ+NlIAkrMfc9ir79ZMGJxLnPEDuHkf6Yi0akyyWWcR3JaMY+yp2A==",
		}, &reply)
		assert.Error(t, err)
	})

	err := storer.AddCustomer(context.Background(), routing.Customer{Name: "Local", Code: "local"})
	assert.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		var reply jsonrpc.JSAddBuyerReply
		err := svc.JSAddBuyer(req, &jsonrpc.JSAddBuyerArgs{
			ShortName:           "local",
			ExoticLocationFee:   "100",
			StandardLocationFee: "100",
			LookerSeats:         "100",
			PublicKey:           "KcZ+NlIAkrMfc9ir79ZMGJxLnPEDuHkf6Yi0akyyWWcR3JaMY+yp2A==",
		}, &reply)
		assert.NoError(t, err)

		buyers := storer.Buyers(context.Background())
		assert.Equal(t, 1, len(buyers))
		assert.Equal(t, "local", buyers[0].ShortName)
	})
}

func TestRemoveBuyer(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}

	svc := jsonrpc.OpsService{
		Storage: &storer,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("insufficient privileges", func(t *testing.T) {
		var reply jsonrpc.RemoveBuyerReply
		err := svc.RemoveBuyer(req, &jsonrpc.RemoveBuyerArgs{}, &reply)
		assert.Equal(t, jsonrpc.JSONRPCErrorCodes[int(jsonrpc.ERROR_INSUFFICIENT_PRIVILEGES)].Message, err.Error())
	})

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
		"Admin",
	})
	req = req.WithContext(reqContext)

	t.Run("bad buyer ID", func(t *testing.T) {
		var reply jsonrpc.RemoveBuyerReply
		err := svc.RemoveBuyer(req, &jsonrpc.RemoveBuyerArgs{ID: "-"}, &reply)
		assert.Contains(t, err.Error(), "RemoveBuyer() could not convert buyer ID - to uint64:")
	})

	t.Run("buyer does not exist", func(t *testing.T) {
		var reply jsonrpc.RemoveBuyerReply
		err := svc.RemoveBuyer(req, &jsonrpc.RemoveBuyerArgs{ID: "0"}, &reply)
		assert.Contains(t, err.Error(), "buyer with reference 0 not found")
	})

	err := storer.AddBuyer(context.Background(), routing.Buyer{ID: 1, CompanyCode: "local", ShortName: "local.1"})
	assert.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		var reply jsonrpc.RemoveBuyerReply
		err := svc.RemoveBuyer(req, &jsonrpc.RemoveBuyerArgs{ID: "1"}, &reply)
		assert.NoError(t, err)

		buyers := storer.Buyers(context.Background())
		assert.NoError(t, err)
		assert.Zero(t, len(buyers))
	})
}

func TestSellers(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}

	err := storer.AddSeller(context.Background(), routing.Seller{ID: "1", CompanyCode: "local", Name: "local.1"})
	assert.NoError(t, err)
	err = storer.AddSeller(context.Background(), routing.Seller{ID: "2", CompanyCode: "local-local", Name: "local.local.2"})
	assert.NoError(t, err)

	svc := jsonrpc.OpsService{
		Storage: &storer,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("insufficient privileges", func(t *testing.T) {
		var reply jsonrpc.SellersReply
		err := svc.Sellers(req, &jsonrpc.SellersArgs{}, &reply)
		assert.Equal(t, jsonrpc.JSONRPCErrorCodes[int(jsonrpc.ERROR_INSUFFICIENT_PRIVILEGES)].Message, err.Error())
	})

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
		"Admin",
	})
	req = req.WithContext(reqContext)

	t.Run("success - list", func(t *testing.T) {
		var reply jsonrpc.SellersReply
		err := svc.Sellers(req, &jsonrpc.SellersArgs{}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, reply.Sellers[0].ID, "1")
		assert.Equal(t, reply.Sellers[0].Name, "local.1")
		assert.Equal(t, reply.Sellers[1].ID, "2")
		assert.Equal(t, reply.Sellers[1].Name, "local.local.2")
	})
}

func TestCustomers(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}

	svc := jsonrpc.OpsService{
		Storage: &storer,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("insufficient privileges", func(t *testing.T) {
		var reply jsonrpc.CustomersReply
		err := svc.Customers(req, &jsonrpc.CustomersArgs{}, &reply)
		assert.Equal(t, jsonrpc.JSONRPCErrorCodes[int(jsonrpc.ERROR_INSUFFICIENT_PRIVILEGES)].Message, err.Error())
	})

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
		"Admin",
	})
	req = req.WithContext(reqContext)

	err := storer.AddCustomer(context.Background(), routing.Customer{Name: "Local", Code: "local"})
	assert.NoError(t, err)
	err = storer.AddCustomer(context.Background(), routing.Customer{Name: "Local-Local", Code: "local-local"})
	assert.NoError(t, err)

	t.Run("success - list", func(t *testing.T) {
		var reply jsonrpc.CustomersReply
		err := svc.Customers(req, &jsonrpc.CustomersArgs{}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, reply.Customers[0].Code, "local")
		assert.Equal(t, reply.Customers[0].Name, "Local")
		assert.Equal(t, reply.Customers[1].Code, "local-local")
		assert.Equal(t, reply.Customers[1].Name, "Local-Local")
	})
}

func TestJSAddCustomers(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}

	svc := jsonrpc.OpsService{
		Storage: &storer,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("insufficient privileges", func(t *testing.T) {
		var reply jsonrpc.JSAddCustomerReply
		err := svc.JSAddCustomer(req, &jsonrpc.JSAddCustomerArgs{}, &reply)
		assert.Equal(t, jsonrpc.JSONRPCErrorCodes[int(jsonrpc.ERROR_INSUFFICIENT_PRIVILEGES)].Message, err.Error())
	})

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
		"Admin",
	})
	req = req.WithContext(reqContext)

	t.Run("success", func(t *testing.T) {
		var reply jsonrpc.JSAddCustomerReply
		err := svc.JSAddCustomer(req, &jsonrpc.JSAddCustomerArgs{Code: "local", Name: "Local"}, &reply)
		assert.NoError(t, err)

		customers := storer.Customers(context.Background())
		assert.Equal(t, 1, len(customers))
		assert.Equal(t, "local", customers[0].Code)
		assert.Equal(t, "Local", customers[0].Name)
	})
}

func TestAddCustomers(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}

	svc := jsonrpc.OpsService{
		Storage: &storer,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("insufficient privileges", func(t *testing.T) {
		var reply jsonrpc.AddCustomerReply
		err := svc.AddCustomer(req, &jsonrpc.AddCustomerArgs{}, &reply)
		assert.Equal(t, jsonrpc.JSONRPCErrorCodes[int(jsonrpc.ERROR_INSUFFICIENT_PRIVILEGES)].Message, err.Error())
	})

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
		"Admin",
	})
	req = req.WithContext(reqContext)

	t.Run("success", func(t *testing.T) {
		expectedCustomer := routing.Customer{
			Code: "local",
			Name: "Local",
		}

		var reply jsonrpc.AddCustomerReply
		err := svc.AddCustomer(req, &jsonrpc.AddCustomerArgs{Customer: expectedCustomer}, &reply)
		assert.NoError(t, err)

		customers := storer.Customers(context.Background())
		assert.Equal(t, 1, len(customers))
		assert.Equal(t, expectedCustomer.Code, customers[0].Code)
		assert.Equal(t, expectedCustomer.Name, customers[0].Name)
	})
}

func TestCustomer(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}

	svc := jsonrpc.OpsService{
		Storage: &storer,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("insufficient privileges", func(t *testing.T) {
		var reply jsonrpc.CustomerReply
		err := svc.Customer(req, &jsonrpc.CustomerArg{}, &reply)
		assert.Equal(t, jsonrpc.JSONRPCErrorCodes[int(jsonrpc.ERROR_INSUFFICIENT_PRIVILEGES)].Message, err.Error())
	})

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
		"Admin",
	})
	req = req.WithContext(reqContext)

	t.Run("customer does not exist", func(t *testing.T) {
		var reply jsonrpc.CustomerReply
		err := svc.Customer(req, &jsonrpc.CustomerArg{CustomerID: "local"}, &reply)
		assert.Error(t, err)
	})

	expectedCustomer := routing.Customer{
		Code: "local",
		Name: "Local",
	}
	err := storer.AddCustomer(context.Background(), expectedCustomer)
	assert.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		var reply jsonrpc.CustomerReply
		err := svc.Customer(req, &jsonrpc.CustomerArg{CustomerID: "local"}, &reply)
		assert.NoError(t, err)
		assert.Equal(t, expectedCustomer, reply.Customer)
	})
}

func TestSeller(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}

	svc := jsonrpc.OpsService{
		Storage: &storer,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("insufficient privileges", func(t *testing.T) {
		var reply jsonrpc.SellerReply
		err := svc.Seller(req, &jsonrpc.SellerArg{}, &reply)
		assert.Equal(t, jsonrpc.JSONRPCErrorCodes[int(jsonrpc.ERROR_INSUFFICIENT_PRIVILEGES)].Message, err.Error())
	})

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
		"Admin",
	})
	req = req.WithContext(reqContext)

	t.Run("seller does not exist", func(t *testing.T) {
		var reply jsonrpc.SellerReply
		err := svc.Seller(req, &jsonrpc.SellerArg{ID: "1"}, &reply)
		assert.Error(t, err)
	})

	expectedSeller := routing.Seller{
		ID:          "1",
		CompanyCode: "local",
		Name:        "Local",
	}
	err := storer.AddSeller(context.Background(), expectedSeller)
	assert.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		var reply jsonrpc.SellerReply
		err := svc.Seller(req, &jsonrpc.SellerArg{ID: "1"}, &reply)
		assert.NoError(t, err)
		assert.Equal(t, expectedSeller, reply.Seller)
	})
}

func TestJSAddSeller(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}

	svc := jsonrpc.OpsService{
		Storage: &storer,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("insufficient privileges", func(t *testing.T) {
		var reply jsonrpc.JSAddSellerReply
		err := svc.JSAddSeller(req, &jsonrpc.JSAddSellerArgs{}, &reply)
		assert.Equal(t, jsonrpc.JSONRPCErrorCodes[int(jsonrpc.ERROR_INSUFFICIENT_PRIVILEGES)].Message, err.Error())
	})

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
		"Admin",
	})
	req = req.WithContext(reqContext)

	t.Run("success", func(t *testing.T) {
		var reply jsonrpc.JSAddSellerReply
		err := svc.JSAddSeller(req, &jsonrpc.JSAddSellerArgs{
			ShortName:   "local",
			EgressPrice: 100,
		}, &reply)
		assert.NoError(t, err)

		sellers := storer.Sellers(context.Background())
		assert.Equal(t, 1, len(sellers))
		assert.Equal(t, "local", sellers[0].ShortName)
		assert.Equal(t, routing.Nibblin(100), sellers[0].EgressPriceNibblinsPerGB)
	})
}

func TestAddSeller(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}

	svc := jsonrpc.OpsService{
		Storage: &storer,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("insufficient privileges", func(t *testing.T) {
		var reply jsonrpc.AddSellerReply
		err := svc.AddSeller(req, &jsonrpc.AddSellerArgs{}, &reply)
		assert.Equal(t, jsonrpc.JSONRPCErrorCodes[int(jsonrpc.ERROR_INSUFFICIENT_PRIVILEGES)].Message, err.Error())
	})

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
		"Admin",
	})
	req = req.WithContext(reqContext)

	t.Run("success", func(t *testing.T) {
		expectedSeller := routing.Seller{
			ID:          "1",
			CompanyCode: "local",
			Name:        "Local",
		}

		var reply jsonrpc.AddSellerReply
		err := svc.AddSeller(req, &jsonrpc.AddSellerArgs{Seller: expectedSeller}, &reply)
		assert.NoError(t, err)

		sellers := storer.Sellers(context.Background())
		assert.Equal(t, 1, len(sellers))
		assert.Equal(t, "local", sellers[0].CompanyCode)
	})
}

func TestRemoveSeller(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}

	svc := jsonrpc.OpsService{
		Storage: &storer,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("insufficient privileges", func(t *testing.T) {
		var reply jsonrpc.RemoveSellerReply
		err := svc.RemoveSeller(req, &jsonrpc.RemoveSellerArgs{}, &reply)
		assert.Equal(t, jsonrpc.JSONRPCErrorCodes[int(jsonrpc.ERROR_INSUFFICIENT_PRIVILEGES)].Message, err.Error())
	})

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
		"Admin",
	})
	req = req.WithContext(reqContext)

	t.Run("seller does not exist", func(t *testing.T) {
		var reply jsonrpc.RemoveSellerReply
		err := svc.RemoveSeller(req, &jsonrpc.RemoveSellerArgs{ID: "0"}, &reply)
		assert.Contains(t, err.Error(), "seller with reference 0 not found")
	})

	err := storer.AddSeller(context.Background(), routing.Seller{ID: "1", CompanyCode: "local", Name: "Local"})
	assert.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		var reply jsonrpc.RemoveSellerReply
		err := svc.RemoveSeller(req, &jsonrpc.RemoveSellerArgs{ID: "1"}, &reply)
		assert.NoError(t, err)

		sellers := storer.Sellers(context.Background())
		assert.Zero(t, len(sellers))
	})
}

func TestRelays(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}

	svc := jsonrpc.OpsService{
		Storage: &storer,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("insufficient privileges", func(t *testing.T) {
		var reply jsonrpc.RelaysReply
		err := svc.Relays(req, &jsonrpc.RelaysArgs{}, &reply)
		assert.Equal(t, jsonrpc.JSONRPCErrorCodes[int(jsonrpc.ERROR_INSUFFICIENT_PRIVILEGES)].Message, err.Error())
	})

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
		"Admin",
	})
	req = req.WithContext(reqContext)

	seller := routing.Seller{
		ID:   "sellerID",
		Name: "seller name",
	}

	datacenter := routing.Datacenter{
		ID:   crypto.HashID("datacenter name"),
		Name: "datacenter name",
	}

	relay1 := routing.Relay{
		ID:         1,
		Name:       "local.local.1",
		Seller:     seller,
		Datacenter: datacenter,
	}

	relay2 := routing.Relay{
		ID:         2,
		Name:       "local.local.2",
		Seller:     seller,
		Datacenter: datacenter,
	}

	relay3 := routing.Relay{
		ID:         3,
		Name:       "local.local.3",
		Seller:     seller,
		Datacenter: datacenter,
	}

	err := storer.AddSeller(context.Background(), seller)
	assert.NoError(t, err)
	err = storer.AddDatacenter(context.Background(), datacenter)
	assert.NoError(t, err)
	err = storer.AddRelay(context.Background(), relay1)
	assert.NoError(t, err)
	err = storer.AddRelay(context.Background(), relay2)
	assert.NoError(t, err)
	err = storer.AddRelay(context.Background(), relay3)
	assert.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		var reply jsonrpc.RelaysReply
		err := svc.Relays(req, &jsonrpc.RelaysArgs{}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, 3, len(reply.Relays))
		assert.Equal(t, relay1.Name, reply.Relays[0].Name)
		assert.Equal(t, relay2.Name, reply.Relays[1].Name)
		assert.Equal(t, relay3.Name, reply.Relays[2].Name)
	})

	t.Run("success - regex", func(t *testing.T) {
		var reply jsonrpc.RelaysReply
		err := svc.Relays(req, &jsonrpc.RelaysArgs{Regex: "local.3"}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(reply.Relays))
		assert.Equal(t, relay3.Name, reply.Relays[0].Name)
	})
}

func TestRelaysWithEgressPriceOverride(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}

	svc := jsonrpc.OpsService{
		Storage: &storer,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("insufficient privileges", func(t *testing.T) {
		var reply jsonrpc.RelayEgressPriceOverrideReply
		err := svc.RelaysWithEgressPriceOverride(req, &jsonrpc.RelayEgressPriceOverrideArgs{}, &reply)
		assert.Equal(t, jsonrpc.JSONRPCErrorCodes[int(jsonrpc.ERROR_INSUFFICIENT_PRIVILEGES)].Message, err.Error())
	})

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
		"Admin",
	})
	req = req.WithContext(reqContext)

	seller := routing.Seller{
		ID:        "sellerID",
		Name:      "seller name",
		ShortName: "sname",
	}

	datacenter := routing.Datacenter{
		ID:   crypto.HashID("datacenter name"),
		Name: "datacenter name",
	}

	relay1 := routing.Relay{
		ID:         1,
		Name:       "local.local.1",
		Seller:     seller,
		Datacenter: datacenter,
	}

	relay2 := routing.Relay{
		ID:         2,
		Name:       "local.local.2",
		Seller:     seller,
		Datacenter: datacenter,
	}

	relay3 := routing.Relay{
		ID:                  3,
		Name:                "local.local.3",
		Seller:              seller,
		Datacenter:          datacenter,
		EgressPriceOverride: routing.Nibblin(100),
	}

	err := storer.AddSeller(context.Background(), seller)
	assert.NoError(t, err)
	err = storer.AddDatacenter(context.Background(), datacenter)
	assert.NoError(t, err)
	err = storer.AddRelay(context.Background(), relay1)
	assert.NoError(t, err)
	err = storer.AddRelay(context.Background(), relay2)
	assert.NoError(t, err)
	err = storer.AddRelay(context.Background(), relay3)
	assert.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		var reply jsonrpc.RelayEgressPriceOverrideReply
		err := svc.RelaysWithEgressPriceOverride(req, &jsonrpc.RelayEgressPriceOverrideArgs{SellerShortName: "sname"}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(reply.Relays))
		assert.Equal(t, relay3.Name, reply.Relays[0].Name)
	})
}

func TestAddRelay(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}

	svc := jsonrpc.OpsService{
		Storage: &storer,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("insufficient privileges", func(t *testing.T) {
		var reply jsonrpc.AddRelayReply
		err := svc.AddRelay(req, &jsonrpc.AddRelayArgs{}, &reply)
		assert.Equal(t, jsonrpc.JSONRPCErrorCodes[int(jsonrpc.ERROR_INSUFFICIENT_PRIVILEGES)].Message, err.Error())
	})

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
		"Admin",
	})
	req = req.WithContext(reqContext)

	t.Run("add relay without seller", func(t *testing.T) {
		var reply jsonrpc.AddRelayReply
		err := svc.AddRelay(req, &jsonrpc.AddRelayArgs{
			routing.Relay{
				ID:   1,
				Name: "local.local.11",
				Seller: routing.Seller{
					ID: "0",
				},
			},
		}, &reply)
		assert.Contains(t, err.Error(), "seller with reference 0 not found")
	})

	seller := routing.Seller{
		ID:        "sellerID",
		Name:      "seller name",
		ShortName: "sname",
	}

	err := storer.AddSeller(context.Background(), seller)
	assert.NoError(t, err)

	t.Run("add relay without datacenter", func(t *testing.T) {
		var reply jsonrpc.AddRelayReply
		err := svc.AddRelay(req, &jsonrpc.AddRelayArgs{
			routing.Relay{
				ID:     1,
				Name:   "local.local.11",
				Seller: seller,
				Datacenter: routing.Datacenter{
					ID:   0,
					Name: "unknown",
				},
			},
		}, &reply)
		assert.Contains(t, err.Error(), "datacenter with reference 0 not found")
	})

	datacenter := routing.Datacenter{
		ID:   crypto.HashID("datacenter name"),
		Name: "datacenter name",
	}

	err = storer.AddDatacenter(context.Background(), datacenter)
	assert.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		var reply jsonrpc.AddRelayReply
		err := svc.AddRelay(req, &jsonrpc.AddRelayArgs{
			routing.Relay{
				ID:         11,
				Name:       "local.local.11",
				Seller:     seller,
				Datacenter: datacenter,
			},
		}, &reply)
		assert.NoError(t, err)

		relays := storer.Relays(context.Background())
		assert.Equal(t, 1, len(relays))
		assert.Equal(t, "local.local.11", relays[0].Name)
	})
}

func TestJSAddRelay(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}

	svc := jsonrpc.OpsService{
		Storage: &storer,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("insufficient privileges", func(t *testing.T) {
		var reply jsonrpc.JSAddRelayReply
		err := svc.JSAddRelay(req, &jsonrpc.JSAddRelayArgs{}, &reply)
		assert.Equal(t, jsonrpc.JSONRPCErrorCodes[int(jsonrpc.ERROR_INSUFFICIENT_PRIVILEGES)].Message, err.Error())
	})

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
		"Admin",
	})
	req = req.WithContext(reqContext)

	t.Run("add relay without datacenter", func(t *testing.T) {
		var reply jsonrpc.JSAddRelayReply
		err := svc.JSAddRelay(req, &jsonrpc.JSAddRelayArgs{
			Name:         "local.local.11",
			SellerID:     "0",
			DatacenterID: "0",
		}, &reply)
		assert.Contains(t, err.Error(), "datacenter with reference 0 not found")
	})

	datacenter := routing.Datacenter{
		ID:   0,
		Name: "datacenter name",
	}

	err := storer.AddDatacenter(context.Background(), datacenter)
	assert.NoError(t, err)

	t.Run("add relay without seller", func(t *testing.T) {
		var reply jsonrpc.JSAddRelayReply
		err := svc.JSAddRelay(req, &jsonrpc.JSAddRelayArgs{
			Name:            "local.local.11",
			BillingSupplier: "0",
			DatacenterID:    "0",
		}, &reply)
		assert.Contains(t, err.Error(), "seller with reference  not found")
	})

	seller := routing.Seller{
		ID:        "sname",
		Name:      "seller name",
		ShortName: "sname",
	}

	err = storer.AddSeller(context.Background(), seller)
	assert.NoError(t, err)

	t.Run("bad public key", func(t *testing.T) {
		var reply jsonrpc.JSAddRelayReply
		err := svc.JSAddRelay(req, &jsonrpc.JSAddRelayArgs{
			Name:            "local.local.11",
			BillingSupplier: "sname",
			DatacenterID:    "0",
			PublicKey:       "KcZ+NlIAkrMfc9ir79ZMGJxLnPEDuHkf6Yi0akyyWWcR3JaMY+yp2",
		}, &reply)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "could not decode base64 public key")
	})

	t.Run("bad internal addr", func(t *testing.T) {
		var reply jsonrpc.JSAddRelayReply
		err := svc.JSAddRelay(req, &jsonrpc.JSAddRelayArgs{
			Name:            "local.local.11",
			BillingSupplier: "sname",
			DatacenterID:    "0",
			PublicKey:       "KcZ+NlIAkrMfc9ir79ZMGJxLnPEDuHkf6Yi0akyyWWcR3JaMY+yp2A==",
			InternalAddr:    "127.0.0.1.1",
		}, &reply)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "address 127.0.0.1.1: missing port in address")
	})

	t.Run("bad start date", func(t *testing.T) {
		var reply jsonrpc.JSAddRelayReply
		err := svc.JSAddRelay(req, &jsonrpc.JSAddRelayArgs{
			Name:            "local.local.11",
			BillingSupplier: "sname",
			DatacenterID:    "0",
			PublicKey:       "KcZ+NlIAkrMfc9ir79ZMGJxLnPEDuHkf6Yi0akyyWWcR3JaMY+yp2A==",
			EndDate:         "2021-13-51",
		}, &reply)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), `parsing time "2021-13-51": month out of range`)
	})

	t.Run("bad end date", func(t *testing.T) {
		var reply jsonrpc.JSAddRelayReply
		err := svc.JSAddRelay(req, &jsonrpc.JSAddRelayArgs{
			Name:            "local.local.11",
			BillingSupplier: "sname",
			DatacenterID:    "0",
			PublicKey:       "KcZ+NlIAkrMfc9ir79ZMGJxLnPEDuHkf6Yi0akyyWWcR3JaMY+yp2A==",
			EndDate:         "2021-13-51",
		}, &reply)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), `parsing time "2021-13-51": month out of range`)
	})

	t.Run("success", func(t *testing.T) {
		var reply jsonrpc.JSAddRelayReply
		err := svc.JSAddRelay(req, &jsonrpc.JSAddRelayArgs{
			Name:            "local.local.11",
			BillingSupplier: "sname",
			DatacenterID:    "0",
			PublicKey:       "KcZ+NlIAkrMfc9ir79ZMGJxLnPEDuHkf6Yi0akyyWWcR3JaMY+yp2A==",
		}, &reply)
		assert.NoError(t, err)

		relays := storer.Relays(context.Background())
		assert.Equal(t, 1, len(relays))
		assert.Equal(t, "local.local.11", relays[0].Name)
	})
}

func TestRemoveRelay(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}

	svc := jsonrpc.OpsService{
		Storage: &storer,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("insufficient privileges", func(t *testing.T) {
		var reply jsonrpc.RemoveRelayReply
		err := svc.RemoveRelay(req, &jsonrpc.RemoveRelayArgs{}, &reply)
		assert.Equal(t, jsonrpc.JSONRPCErrorCodes[int(jsonrpc.ERROR_INSUFFICIENT_PRIVILEGES)].Message, err.Error())
	})

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
		"Admin",
	})
	req = req.WithContext(reqContext)

	t.Run("remove non-existant relay", func(t *testing.T) {
		var reply jsonrpc.RemoveRelayReply
		err := svc.RemoveRelay(req, &jsonrpc.RemoveRelayArgs{RelayID: 0}, &reply)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "RemoveRelay() Storage.Relay error")
	})

	seller := routing.Seller{
		ID:        "sellerID",
		Name:      "seller name",
		ShortName: "sname",
	}

	datacenter := routing.Datacenter{
		ID:   crypto.HashID("datacenter name"),
		Name: "datacenter name",
	}

	relay1 := routing.Relay{
		ID:         1,
		Name:       "local.local.1",
		Seller:     seller,
		Datacenter: datacenter,
	}

	err := storer.AddSeller(context.Background(), seller)
	assert.NoError(t, err)
	err = storer.AddDatacenter(context.Background(), datacenter)
	assert.NoError(t, err)
	err = storer.AddRelay(context.Background(), relay1)
	assert.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		var reply jsonrpc.RemoveRelayReply
		err := svc.RemoveRelay(req, &jsonrpc.RemoveRelayArgs{RelayID: relay1.ID}, &reply)
		assert.NoError(t, err)

		relay, err := storer.Relay(context.Background(), relay1.ID)
		assert.Contains(t, relay.Name, "local.local.1-removed-")
		assert.Equal(t, routing.RelayStateDecommissioned, relay.State)
		assert.Equal(t, net.UDPAddr{}, relay.Addr)
	})
}

func TestRelayNameUpdate(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}

	svc := jsonrpc.OpsService{
		Storage: &storer,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("insufficient privileges", func(t *testing.T) {
		var reply jsonrpc.RelayNameUpdateReply
		err := svc.RelayNameUpdate(req, &jsonrpc.RelayNameUpdateArgs{}, &reply)
		assert.Equal(t, jsonrpc.JSONRPCErrorCodes[int(jsonrpc.ERROR_INSUFFICIENT_PRIVILEGES)].Message, err.Error())
	})

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
		"Admin",
	})
	req = req.WithContext(reqContext)

	t.Run("update non-existant relay", func(t *testing.T) {
		var reply jsonrpc.RelayNameUpdateReply
		err := svc.RelayNameUpdate(req, &jsonrpc.RelayNameUpdateArgs{RelayID: 0}, &reply)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "RelayNameUpdate() Storage.Relay error")
	})

	seller := routing.Seller{
		ID:        "sellerID",
		Name:      "seller name",
		ShortName: "sname",
	}

	datacenter := routing.Datacenter{
		ID:   crypto.HashID("datacenter name"),
		Name: "datacenter name",
	}

	relay1 := routing.Relay{
		ID:         1,
		Name:       "local.local.1",
		Seller:     seller,
		Datacenter: datacenter,
	}

	err := storer.AddSeller(context.Background(), seller)
	assert.NoError(t, err)
	err = storer.AddDatacenter(context.Background(), datacenter)
	assert.NoError(t, err)
	err = storer.AddRelay(context.Background(), relay1)
	assert.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		var reply jsonrpc.RelayNameUpdateReply
		err := svc.RelayNameUpdate(req, &jsonrpc.RelayNameUpdateArgs{RelayID: relay1.ID, RelayName: "new_name"}, &reply)
		assert.NoError(t, err)

		relay, err := storer.Relay(context.Background(), relay1.ID)
		assert.Equal(t, relay.Name, "new_name")
	})
}

func TestRelayStateUpdate(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}

	svc := jsonrpc.OpsService{
		Storage: &storer,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("insufficient privileges", func(t *testing.T) {
		var reply jsonrpc.RelayStateUpdateReply
		err := svc.RelayStateUpdate(req, &jsonrpc.RelayStateUpdateArgs{}, &reply)
		assert.Equal(t, jsonrpc.JSONRPCErrorCodes[int(jsonrpc.ERROR_INSUFFICIENT_PRIVILEGES)].Message, err.Error())
	})

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
		"Admin",
	})
	req = req.WithContext(reqContext)

	t.Run("update non-existant relay", func(t *testing.T) {
		var reply jsonrpc.RelayStateUpdateReply
		err := svc.RelayStateUpdate(req, &jsonrpc.RelayStateUpdateArgs{RelayID: 0}, &reply)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "RelayStateUpdate() Storage.Relay error")
	})

	seller := routing.Seller{
		ID:        "sellerID",
		Name:      "seller name",
		ShortName: "sname",
	}

	datacenter := routing.Datacenter{
		ID:   crypto.HashID("datacenter name"),
		Name: "datacenter name",
	}

	relay1 := routing.Relay{
		ID:         1,
		Name:       "local.local.1",
		Seller:     seller,
		Datacenter: datacenter,
		State:      routing.RelayStateEnabled,
	}

	err := storer.AddSeller(context.Background(), seller)
	assert.NoError(t, err)
	err = storer.AddDatacenter(context.Background(), datacenter)
	assert.NoError(t, err)
	err = storer.AddRelay(context.Background(), relay1)
	assert.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		var reply jsonrpc.RelayStateUpdateReply
		err := svc.RelayStateUpdate(req, &jsonrpc.RelayStateUpdateArgs{RelayID: relay1.ID, RelayState: routing.RelayStateDecommissioned}, &reply)
		assert.NoError(t, err)

		relay, err := storer.Relay(context.Background(), relay1.ID)
		assert.Equal(t, relay.State, routing.RelayStateDecommissioned)
	})
}

func TestRelayPublicKeyUpdate(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}

	svc := jsonrpc.OpsService{
		Storage: &storer,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("insufficient privileges", func(t *testing.T) {
		var reply jsonrpc.RelayPublicKeyUpdateReply
		err := svc.RelayPublicKeyUpdate(req, &jsonrpc.RelayPublicKeyUpdateArgs{}, &reply)
		assert.Equal(t, jsonrpc.JSONRPCErrorCodes[int(jsonrpc.ERROR_INSUFFICIENT_PRIVILEGES)].Message, err.Error())
	})

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
		"Admin",
	})
	req = req.WithContext(reqContext)

	t.Run("update non-existant relay", func(t *testing.T) {
		var reply jsonrpc.RelayPublicKeyUpdateReply
		err := svc.RelayPublicKeyUpdate(req, &jsonrpc.RelayPublicKeyUpdateArgs{RelayID: 0}, &reply)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "RelayPublicKeyUpdate()")
	})

	seller := routing.Seller{
		ID:        "sellerID",
		Name:      "seller name",
		ShortName: "sname",
	}

	datacenter := routing.Datacenter{
		ID:   crypto.HashID("datacenter name"),
		Name: "datacenter name",
	}

	relay1 := routing.Relay{
		ID:         1,
		Name:       "local.local.1",
		Seller:     seller,
		Datacenter: datacenter,
	}

	err := storer.AddSeller(context.Background(), seller)
	assert.NoError(t, err)
	err = storer.AddDatacenter(context.Background(), datacenter)
	assert.NoError(t, err)
	err = storer.AddRelay(context.Background(), relay1)
	assert.NoError(t, err)

	t.Run("bad public key", func(t *testing.T) {
		var reply jsonrpc.RelayPublicKeyUpdateReply
		err := svc.RelayPublicKeyUpdate(req, &jsonrpc.RelayPublicKeyUpdateArgs{RelayID: relay1.ID, RelayPublicKey: "KcZ+NlIAkrMfc9ir79ZMGJxLnPEDuHkf6Yi0akyyWWcR3JaMY+yp2A"}, &reply)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "RelayPublicKeyUpdate() could not decode relay public key")
	})

	t.Run("success", func(t *testing.T) {
		var reply jsonrpc.RelayPublicKeyUpdateReply
		err := svc.RelayPublicKeyUpdate(req, &jsonrpc.RelayPublicKeyUpdateArgs{RelayID: relay1.ID, RelayPublicKey: "KcZ+NlIAkrMfc9ir79ZMGJxLnPEDuHkf6Yi0akyyWWcR3JaMY+yp2A=="}, &reply)
		assert.NoError(t, err)

		relay, err := storer.Relay(context.Background(), relay1.ID)

		expectedPublicKey, err := base64.StdEncoding.DecodeString("KcZ+NlIAkrMfc9ir79ZMGJxLnPEDuHkf6Yi0akyyWWcR3JaMY+yp2A==")
		assert.NoError(t, err)
		assert.Equal(t, relay.PublicKey, expectedPublicKey)
	})
}

func TestDatacenter(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}

	svc := jsonrpc.OpsService{
		Storage: &storer,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("insufficient privileges", func(t *testing.T) {
		var reply jsonrpc.DatacenterReply
		err := svc.Datacenter(req, &jsonrpc.DatacenterArg{}, &reply)
		assert.Equal(t, jsonrpc.JSONRPCErrorCodes[int(jsonrpc.ERROR_INSUFFICIENT_PRIVILEGES)].Message, err.Error())
	})

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
		"Admin",
	})
	req = req.WithContext(reqContext)

	t.Run("get non-existant datacenter", func(t *testing.T) {
		var reply jsonrpc.DatacenterReply
		err := svc.Datacenter(req, &jsonrpc.DatacenterArg{ID: 0}, &reply)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Datacenter() error")
	})

	datacenter := routing.Datacenter{
		ID:   crypto.HashID("datacenter name"),
		Name: "datacenter name",
	}

	err := storer.AddDatacenter(context.Background(), datacenter)
	assert.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		var reply jsonrpc.DatacenterReply
		err := svc.Datacenter(req, &jsonrpc.DatacenterArg{ID: crypto.HashID("datacenter name")}, &reply)
		assert.NoError(t, err)
		assert.Equal(t, datacenter, reply.Datacenter)
	})
}

func TestDatacenters(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}

	svc := jsonrpc.OpsService{
		Storage: &storer,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("insufficient privileges", func(t *testing.T) {
		var reply jsonrpc.DatacentersReply
		err := svc.Datacenters(req, &jsonrpc.DatacentersArgs{}, &reply)
		assert.Equal(t, jsonrpc.JSONRPCErrorCodes[int(jsonrpc.ERROR_INSUFFICIENT_PRIVILEGES)].Message, err.Error())
	})

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
		"Admin",
	})
	req = req.WithContext(reqContext)

	datacenter1 := routing.Datacenter{
		ID:   crypto.HashID("datacenter name 1"),
		Name: "datacenter name 1",
	}

	datacenter2 := routing.Datacenter{
		ID:   crypto.HashID("datacenter name 2"),
		Name: "datacenter name 2",
	}

	err := storer.AddDatacenter(context.Background(), datacenter1)
	assert.NoError(t, err)
	err = storer.AddDatacenter(context.Background(), datacenter2)
	assert.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		var reply jsonrpc.DatacentersReply
		err := svc.Datacenters(req, &jsonrpc.DatacentersArgs{}, &reply)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(reply.Datacenters))
		assert.Equal(t, datacenter1.Name, reply.Datacenters[0].Name)
		assert.Equal(t, datacenter2.Name, reply.Datacenters[1].Name)
	})

	t.Run("success - name filter", func(t *testing.T) {
		var reply jsonrpc.DatacentersReply
		err := svc.Datacenters(req, &jsonrpc.DatacentersArgs{Name: "datacenter name 1"}, &reply)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(reply.Datacenters))
		assert.Equal(t, datacenter1.Name, reply.Datacenters[0].Name)
	})
}

func TestAddDatacenter(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}

	svc := jsonrpc.OpsService{
		Storage: &storer,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("insufficient privileges", func(t *testing.T) {
		var reply jsonrpc.AddDatacenterReply
		err := svc.AddDatacenter(req, &jsonrpc.AddDatacenterArgs{}, &reply)
		assert.Equal(t, jsonrpc.JSONRPCErrorCodes[int(jsonrpc.ERROR_INSUFFICIENT_PRIVILEGES)].Message, err.Error())
	})

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
		"Admin",
	})
	req = req.WithContext(reqContext)

	datacenter1 := routing.Datacenter{
		ID:   crypto.HashID("datacenter name 1"),
		Name: "datacenter name 1",
	}

	err := storer.AddDatacenter(context.Background(), datacenter1)
	assert.NoError(t, err)

	t.Run("add datacenter that already exists", func(t *testing.T) {
		var reply jsonrpc.AddDatacenterReply
		err := svc.AddDatacenter(req, &jsonrpc.AddDatacenterArgs{Datacenter: datacenter1}, &reply)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "AddDatacenter() error: ")
	})

	err = storer.RemoveDatacenter(context.Background(), datacenter1.ID)
	assert.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		var reply jsonrpc.AddDatacenterReply
		err := svc.AddDatacenter(req, &jsonrpc.AddDatacenterArgs{Datacenter: datacenter1}, &reply)
		assert.NoError(t, err)

		dc, err := storer.Datacenter(context.Background(), datacenter1.ID)
		assert.NoError(t, err)
		assert.Equal(t, datacenter1, dc)
	})
}

func TestJSAddDatacenter(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}

	svc := jsonrpc.OpsService{
		Storage: &storer,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("insufficient privileges", func(t *testing.T) {
		var reply jsonrpc.JSAddDatacenterReply
		err := svc.JSAddDatacenter(req, &jsonrpc.JSAddDatacenterArgs{}, &reply)
		assert.Equal(t, jsonrpc.JSONRPCErrorCodes[int(jsonrpc.ERROR_INSUFFICIENT_PRIVILEGES)].Message, err.Error())
	})

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
		"Admin",
	})
	req = req.WithContext(reqContext)

	datacenter1 := routing.Datacenter{
		ID:   crypto.HashID("datacenter name 1"),
		Name: "datacenter name 1",
	}

	t.Run("seller does not exist", func(t *testing.T) {
		var reply jsonrpc.JSAddDatacenterReply
		err := svc.JSAddDatacenter(req, &jsonrpc.JSAddDatacenterArgs{Name: datacenter1.Name, SellerID: ""}, &reply)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "seller with reference  not found")
	})

	seller := routing.Seller{
		ID:         "sname",
		Name:       "seller name",
		ShortName:  "sname",
		DatabaseID: 0,
	}

	err := storer.AddSeller(context.Background(), seller)
	assert.NoError(t, err)

	err = storer.AddDatacenter(context.Background(), datacenter1)
	assert.NoError(t, err)

	t.Run("add datacenter that already exists", func(t *testing.T) {
		var reply jsonrpc.JSAddDatacenterReply
		err := svc.JSAddDatacenter(req, &jsonrpc.JSAddDatacenterArgs{Name: datacenter1.Name, SellerID: seller.ShortName}, &reply)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "AddDatacenter() error:")
	})

	err = storer.RemoveDatacenter(context.Background(), datacenter1.ID)
	assert.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		var reply jsonrpc.JSAddDatacenterReply
		err := svc.JSAddDatacenter(req, &jsonrpc.JSAddDatacenterArgs{Name: datacenter1.Name, SellerID: seller.ShortName}, &reply)
		assert.NoError(t, err)

		dc, err := storer.Datacenter(context.Background(), datacenter1.ID)
		assert.NoError(t, err)
		assert.Equal(t, datacenter1, dc)
	})
}

func TestRemoveDatacenter(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}

	svc := jsonrpc.OpsService{
		Storage: &storer,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("insufficient privileges", func(t *testing.T) {
		var reply jsonrpc.RemoveDatacenterReply
		err := svc.RemoveDatacenter(req, &jsonrpc.RemoveDatacenterArgs{}, &reply)
		assert.Equal(t, jsonrpc.JSONRPCErrorCodes[int(jsonrpc.ERROR_INSUFFICIENT_PRIVILEGES)].Message, err.Error())
	})

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
		"Admin",
	})
	req = req.WithContext(reqContext)

	datacenter1 := routing.Datacenter{
		ID:   crypto.HashID("datacenter name 1"),
		Name: "datacenter name 1",
	}

	t.Run("remove non-existant datacenter", func(t *testing.T) {
		var reply jsonrpc.RemoveDatacenterReply
		err := svc.RemoveDatacenter(req, &jsonrpc.RemoveDatacenterArgs{Name: datacenter1.Name}, &reply)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "RemoveDatacenter() error:")
	})

	err := storer.AddDatacenter(context.Background(), datacenter1)
	assert.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		var reply jsonrpc.RemoveDatacenterReply
		err := svc.RemoveDatacenter(req, &jsonrpc.RemoveDatacenterArgs{Name: datacenter1.Name}, &reply)
		assert.NoError(t, err)

		dcs := storer.Datacenters(context.Background())
		assert.Zero(t, len(dcs))
	})
}

func TestListDatacenterMaps(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}

	svc := jsonrpc.OpsService{
		Storage: &storer,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("insufficient privileges", func(t *testing.T) {
		var reply jsonrpc.ListDatacenterMapsReply
		err := svc.ListDatacenterMaps(req, &jsonrpc.ListDatacenterMapsArgs{}, &reply)
		assert.Equal(t, jsonrpc.JSONRPCErrorCodes[int(jsonrpc.ERROR_INSUFFICIENT_PRIVILEGES)].Message, err.Error())
	})

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
		"Admin",
	})
	req = req.WithContext(reqContext)

	datacenter1 := routing.Datacenter{
		ID:   crypto.HashID("datacenter name 1"),
		Name: "datacenter name 1",
	}

	datacenter2 := routing.Datacenter{
		ID:   crypto.HashID("datacenter name 2"),
		Name: "datacenter name 2",
	}

	dcMap1 := routing.DatacenterMap{
		BuyerID:      0,
		DatacenterID: datacenter1.ID,
	}

	dcMap2 := routing.DatacenterMap{
		BuyerID:      1,
		DatacenterID: datacenter2.ID,
	}

	err := storer.AddDatacenterMap(context.Background(), dcMap1)
	assert.NoError(t, err)
	err = storer.AddDatacenterMap(context.Background(), dcMap2)
	assert.NoError(t, err)

	t.Run("buyer for datacenter map does not exist", func(t *testing.T) {
		var reply jsonrpc.ListDatacenterMapsReply
		err := svc.ListDatacenterMaps(req, &jsonrpc.ListDatacenterMapsArgs{DatacenterID: 0}, &reply)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ListDatacenterMaps() could not parse buyer")
	})

	customer1 := routing.Customer{
		Code: "local1",
		Name: "Local1",
	}

	customer2 := routing.Customer{
		Code: "local2",
		Name: "Local2",
	}

	buyer1 := routing.Buyer{
		ID:        0,
		ShortName: "Local1",
	}

	buyer2 := routing.Buyer{
		ID:        1,
		ShortName: "Local2",
	}

	err = storer.AddCustomer(context.Background(), customer1)
	assert.NoError(t, err)
	err = storer.AddCustomer(context.Background(), customer2)
	assert.NoError(t, err)
	err = storer.AddBuyer(context.Background(), buyer1)
	assert.NoError(t, err)
	err = storer.AddBuyer(context.Background(), buyer2)
	assert.NoError(t, err)

	t.Run("datacenter in datacenter map does not exist", func(t *testing.T) {
		var reply jsonrpc.ListDatacenterMapsReply
		err := svc.ListDatacenterMaps(req, &jsonrpc.ListDatacenterMapsArgs{DatacenterID: 0}, &reply)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ListDatacenterMaps() could not parse datacenter")
	})

	err = storer.AddDatacenter(context.Background(), datacenter1)
	assert.NoError(t, err)
	err = storer.AddDatacenter(context.Background(), datacenter2)
	assert.NoError(t, err)

	t.Run("buyer in datacenter map does not have company code", func(t *testing.T) {
		var reply jsonrpc.ListDatacenterMapsReply
		err := svc.ListDatacenterMaps(req, &jsonrpc.ListDatacenterMapsArgs{DatacenterID: 0}, &reply)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ListDatacenterMaps() failed to find buyer company")
	})

	err = storer.RemoveBuyer(context.Background(), buyer1.ID)
	assert.NoError(t, err)
	err = storer.RemoveBuyer(context.Background(), buyer2.ID)
	assert.NoError(t, err)

	buyer1.CompanyCode = "local1"
	buyer2.CompanyCode = "local2"

	err = storer.AddBuyer(context.Background(), buyer1)
	assert.NoError(t, err)
	err = storer.AddBuyer(context.Background(), buyer2)
	assert.NoError(t, err)

	t.Run("success - all maps", func(t *testing.T) {
		var reply jsonrpc.ListDatacenterMapsReply
		err := svc.ListDatacenterMaps(req, &jsonrpc.ListDatacenterMapsArgs{DatacenterID: 0}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, 2, len(reply.DatacenterMaps))
	})

	t.Run("success - single map", func(t *testing.T) {
		var reply jsonrpc.ListDatacenterMapsReply
		err := svc.ListDatacenterMaps(req, &jsonrpc.ListDatacenterMapsArgs{DatacenterID: datacenter1.ID}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(reply.DatacenterMaps))
		assert.Equal(t, fmt.Sprintf("%016x", datacenter1.ID), reply.DatacenterMaps[0].DatacenterID)
	})
}

func TestCheckRelayIPAddress(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}

	svc := jsonrpc.OpsService{
		Storage: &storer,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("insufficient privileges", func(t *testing.T) {
		var reply jsonrpc.CheckRelayIPAddressReply
		err := svc.CheckRelayIPAddress(req, &jsonrpc.CheckRelayIPAddressArgs{}, &reply)
		assert.Equal(t, jsonrpc.JSONRPCErrorCodes[int(jsonrpc.ERROR_INSUFFICIENT_PRIVILEGES)].Message, err.Error())
	})

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
		"Admin",
	})
	req = req.WithContext(reqContext)

	t.Run("bad hex ID", func(t *testing.T) {
		var reply jsonrpc.CheckRelayIPAddressReply
		err := svc.CheckRelayIPAddress(req, &jsonrpc.CheckRelayIPAddressArgs{HexID: "-"}, &reply)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), `strconv.ParseUint: parsing "-": invalid syntax`)
		assert.False(t, reply.Valid)
	})

	t.Run("bad IP address", func(t *testing.T) {
		var reply jsonrpc.CheckRelayIPAddressReply
		err := svc.CheckRelayIPAddress(req, &jsonrpc.CheckRelayIPAddressArgs{HexID: "0", IpAddress: "127.0.0.1.1"}, &reply)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "address 127.0.0.1.1: missing port in address")
		assert.False(t, reply.Valid)
	})

	t.Run("internal ID mismatch", func(t *testing.T) {
		var reply jsonrpc.CheckRelayIPAddressReply
		err := svc.CheckRelayIPAddress(req, &jsonrpc.CheckRelayIPAddressArgs{HexID: "0", IpAddress: "127.0.0.1:40000"}, &reply)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CheckRelayIPAddress(): internal ID from Hex ID")
		assert.False(t, reply.Valid)
	})

	t.Run("success", func(t *testing.T) {
		var reply jsonrpc.CheckRelayIPAddressReply
		err := svc.CheckRelayIPAddress(req, &jsonrpc.CheckRelayIPAddressArgs{HexID: fmt.Sprintf("%x", crypto.HashID("127.0.0.1:40000")), IpAddress: "127.0.0.1:40000"}, &reply)
		assert.NoError(t, err)
		assert.True(t, reply.Valid)
	})
}

func TestUpdateRelay(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}

	svc := jsonrpc.OpsService{
		Storage: &storer,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("insufficient privileges", func(t *testing.T) {
		var reply jsonrpc.UpdateRelayReply
		err := svc.UpdateRelay(req, &jsonrpc.UpdateRelayArgs{}, &reply)
		assert.Equal(t, jsonrpc.JSONRPCErrorCodes[int(jsonrpc.ERROR_INSUFFICIENT_PRIVILEGES)].Message, err.Error())
	})

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
		"Admin",
	})
	req = req.WithContext(reqContext)

	t.Run("bad hex relay ID", func(t *testing.T) {
		var reply jsonrpc.UpdateRelayReply
		err := svc.UpdateRelay(req, &jsonrpc.UpdateRelayArgs{HexRelayID: "-"}, &reply)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "UpdateRelay() failed to parse HexRelayID")
	})

	t.Run("relay does not exist", func(t *testing.T) {
		var reply jsonrpc.UpdateRelayReply
		err := svc.UpdateRelay(req, &jsonrpc.UpdateRelayArgs{RelayID: 0}, &reply)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "UpdateRelay() failed to modify relay record for field")
	})

	seller := routing.Seller{
		ID:        "sellerID",
		Name:      "seller name",
		ShortName: "sname",
	}

	datacenter := routing.Datacenter{
		ID:   crypto.HashID("datacenter name"),
		Name: "datacenter name",
	}

	relay1 := routing.Relay{
		ID:         1,
		Name:       "local.local.1",
		Seller:     seller,
		Datacenter: datacenter,
	}

	err := storer.AddSeller(context.Background(), seller)
	assert.NoError(t, err)
	err = storer.AddDatacenter(context.Background(), datacenter)
	assert.NoError(t, err)
	err = storer.AddRelay(context.Background(), relay1)
	assert.NoError(t, err)

	t.Run("bad hex ID", func(t *testing.T) {
		var reply jsonrpc.UpdateRelayReply
		err := svc.UpdateRelay(req, &jsonrpc.UpdateRelayArgs{HexRelayID: "-"}, &reply)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), `strconv.ParseUint: parsing "-": invalid syntax`)
	})

	// Storer test takes care of all invalid fields
	t.Run("field is invalid", func(t *testing.T) {
		var reply jsonrpc.UpdateRelayReply
		err := svc.UpdateRelay(req, &jsonrpc.UpdateRelayArgs{RelayID: 1, Field: ""}, &reply)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "UpdateRelay() failed to modify relay record for field")
	})

	// Storer test takes care of all invalid values
	t.Run("value is invalid", func(t *testing.T) {
		var reply jsonrpc.UpdateRelayReply
		err := svc.UpdateRelay(req, &jsonrpc.UpdateRelayArgs{RelayID: 1, Field: "Version", Value: ""}, &reply)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "UpdateRelay() failed to modify relay record for field")
	})

	t.Run("success", func(t *testing.T) {
		var reply jsonrpc.UpdateRelayReply
		err := svc.UpdateRelay(req, &jsonrpc.UpdateRelayArgs{RelayID: 1, Field: "Version", Value: "2.0.8"}, &reply)
		assert.NoError(t, err)
	})
}

func TestGetRelay(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}

	svc := jsonrpc.OpsService{
		Storage: &storer,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("insufficient privileges", func(t *testing.T) {
		var reply jsonrpc.GetRelayReply
		err := svc.GetRelay(req, &jsonrpc.GetRelayArgs{}, &reply)
		assert.Equal(t, jsonrpc.JSONRPCErrorCodes[int(jsonrpc.ERROR_INSUFFICIENT_PRIVILEGES)].Message, err.Error())
	})

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
		"Admin",
	})
	req = req.WithContext(reqContext)

	t.Run("get non-existant relay", func(t *testing.T) {
		var reply jsonrpc.GetRelayReply
		err := svc.GetRelay(req, &jsonrpc.GetRelayArgs{RelayID: 0}, &reply)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), fmt.Sprintf("Error retrieving relay ID %016x: ", 0))
	})

	seller := routing.Seller{
		ID:        "sellerID",
		Name:      "seller name",
		ShortName: "sname",
	}

	datacenter := routing.Datacenter{
		ID:   crypto.HashID("datacenter name"),
		Name: "datacenter name",
	}

	relay1 := routing.Relay{
		ID:         1,
		Name:       "local.local.1",
		Seller:     seller,
		Datacenter: datacenter,
	}

	err := storer.AddSeller(context.Background(), seller)
	assert.NoError(t, err)
	err = storer.AddDatacenter(context.Background(), datacenter)
	assert.NoError(t, err)
	err = storer.AddRelay(context.Background(), relay1)
	assert.NoError(t, err)

	t.Run("get non-existant relay", func(t *testing.T) {
		var reply jsonrpc.GetRelayReply
		err := svc.GetRelay(req, &jsonrpc.GetRelayArgs{RelayID: 1}, &reply)
		assert.NoError(t, err)
		assert.Equal(t, relay1.ID, reply.Relay.ID)
		assert.Equal(t, relay1.Name, reply.Relay.Name)
	})
}

func TestModifyRelayField(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}

	svc := jsonrpc.OpsService{
		Storage: &storer,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("insufficient privileges", func(t *testing.T) {
		var reply jsonrpc.ModifyRelayFieldReply
		err := svc.ModifyRelayField(req, &jsonrpc.ModifyRelayFieldArgs{}, &reply)
		assert.Equal(t, jsonrpc.JSONRPCErrorCodes[int(jsonrpc.ERROR_INSUFFICIENT_PRIVILEGES)].Message, err.Error())
	})

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
		"Admin",
	})
	req = req.WithContext(reqContext)

	seller := routing.Seller{
		ID:         "sellerID",
		Name:       "seller name",
		ShortName:  "sname",
		DatabaseID: 1,
	}

	datacenter := routing.Datacenter{
		ID:   crypto.HashID("datacenter name"),
		Name: "datacenter name",
	}

	relay1 := routing.Relay{
		ID:         1,
		Name:       "local.local.1",
		Seller:     seller,
		Datacenter: datacenter,
	}

	err := storer.AddSeller(context.Background(), seller)
	assert.NoError(t, err)
	err = storer.AddDatacenter(context.Background(), datacenter)
	assert.NoError(t, err)
	err = storer.AddRelay(context.Background(), relay1)
	assert.NoError(t, err)

	float64Fields := []string{"NICSpeedMbps", "IncludedBandwidthGB", "MaxBandwidthMbps", "ContractTerm", "SSHPort", "MaxSessions"}
	addressFields := []string{"Addr", "InternalAddr"}
	timeFields := []string{"StartDate", "EndDate"}
	boolFields := []string{"DestFirst", "InternalAddressClientRoutable"}
	stringFields := []string{"ManagementAddr", "SSHUser", "Version"}
	nibblinFields := []string{"EgressPriceOverride", "MRC", "Overage"}

	// special cases: PublicKey, State, BWRule, Type, BillingSupplier

	t.Run("invalid float64 fields", func(t *testing.T) {
		for _, field := range float64Fields {
			var reply jsonrpc.ModifyRelayFieldReply
			err := svc.ModifyRelayField(req, &jsonrpc.ModifyRelayFieldArgs{RelayID: 1, Field: field, Value: "a"}, &reply)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), fmt.Sprintf("Value: %s is not a valid numeric type", "a"))
		}
	})

	t.Run("invalid address fields", func(t *testing.T) {
		for _, field := range addressFields {
			var reply jsonrpc.ModifyRelayFieldReply
			err := svc.ModifyRelayField(req, &jsonrpc.ModifyRelayFieldArgs{RelayID: 1, Field: field, Value: "-"}, &reply)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), fmt.Sprintf("UpdateRelay() error updating field for relay %016x", 1))
		}
	})

	t.Run("invalid time fields", func(t *testing.T) {
		for _, field := range timeFields {
			var reply jsonrpc.ModifyRelayFieldReply
			err := svc.ModifyRelayField(req, &jsonrpc.ModifyRelayFieldArgs{RelayID: 1, Field: field, Value: "-"}, &reply)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), fmt.Sprintf("UpdateRelay() error updating field for relay %016x", 1))
		}
	})

	t.Run("invalid bool fields", func(t *testing.T) {
		for _, field := range boolFields {
			var reply jsonrpc.ModifyRelayFieldReply
			err := svc.ModifyRelayField(req, &jsonrpc.ModifyRelayFieldArgs{RelayID: 1, Field: field, Value: "-"}, &reply)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), fmt.Sprintf("Value: %s is not a valid boolean type", "-"))
		}
	})

	t.Run("invalid nibblin fields", func(t *testing.T) {
		for _, field := range nibblinFields {
			var reply jsonrpc.ModifyRelayFieldReply
			err := svc.ModifyRelayField(req, &jsonrpc.ModifyRelayFieldArgs{RelayID: 1, Field: field, Value: "a"}, &reply)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), fmt.Sprintf("value '%s' is not a valid float64 number", "a"))
		}
	})

	t.Run("invalid public key", func(t *testing.T) {
		var reply jsonrpc.ModifyRelayFieldReply
		err := svc.ModifyRelayField(req, &jsonrpc.ModifyRelayFieldArgs{RelayID: 1, Field: "PublicKey", Value: "a"}, &reply)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), fmt.Sprintf("UpdateRelay() error updating field for relay %016x", 1))
	})

	// Storer tests handle edge cases
	t.Run("invalid relay state", func(t *testing.T) {
		var reply jsonrpc.ModifyRelayFieldReply
		err := svc.ModifyRelayField(req, &jsonrpc.ModifyRelayFieldArgs{RelayID: 1, Field: "State", Value: "a"}, &reply)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), fmt.Sprintf("value '%s' is not a valid relay state", "a"))
	})

	// Storer tests handle edge cases
	t.Run("invalid bw rule", func(t *testing.T) {
		var reply jsonrpc.ModifyRelayFieldReply
		err := svc.ModifyRelayField(req, &jsonrpc.ModifyRelayFieldArgs{RelayID: 1, Field: "BWRule", Value: "a"}, &reply)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), fmt.Sprintf("value '%s' is not a valid bandwidth rule", "a"))
	})

	// Storer tests handle edge cases
	t.Run("invalid machine type", func(t *testing.T) {
		var reply jsonrpc.ModifyRelayFieldReply
		err := svc.ModifyRelayField(req, &jsonrpc.ModifyRelayFieldArgs{RelayID: 1, Field: "Type", Value: "a"}, &reply)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), fmt.Sprintf("value '%s' is not a valid machine type", "a"))
	})

	t.Run("invalid billing supplier", func(t *testing.T) {
		var reply jsonrpc.ModifyRelayFieldReply
		err := svc.ModifyRelayField(req, &jsonrpc.ModifyRelayFieldArgs{RelayID: 1, Field: "BillingSupplier", Value: "a"}, &reply)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), fmt.Sprintf("UpdateRelay() error updating field for relay %016x", 1))
	})

	t.Run("success float64 fields", func(t *testing.T) {
		for _, field := range float64Fields {
			var reply jsonrpc.ModifyRelayFieldReply
			err := svc.ModifyRelayField(req, &jsonrpc.ModifyRelayFieldArgs{RelayID: 1, Field: field, Value: "1"}, &reply)
			assert.NoError(t, err)
		}
	})

	t.Run("success address fields", func(t *testing.T) {
		for _, field := range addressFields {
			var reply jsonrpc.ModifyRelayFieldReply
			err := svc.ModifyRelayField(req, &jsonrpc.ModifyRelayFieldArgs{RelayID: 1, Field: field, Value: "127.0.0.1:40000"}, &reply)
			assert.NoError(t, err)
		}
	})

	t.Run("success time fields", func(t *testing.T) {
		for _, field := range timeFields {
			var reply jsonrpc.ModifyRelayFieldReply
			err := svc.ModifyRelayField(req, &jsonrpc.ModifyRelayFieldArgs{RelayID: 1, Field: field, Value: "November 17, 2021"}, &reply)
			assert.NoError(t, err)
		}
	})

	t.Run("success string fields", func(t *testing.T) {
		for _, field := range stringFields {
			var reply jsonrpc.ModifyRelayFieldReply
			err := svc.ModifyRelayField(req, &jsonrpc.ModifyRelayFieldArgs{RelayID: 1, Field: field, Value: "some string"}, &reply)
			assert.NoError(t, err)
		}
	})

	t.Run("success bool fields", func(t *testing.T) {
		for _, field := range boolFields {
			var reply jsonrpc.ModifyRelayFieldReply
			err := svc.ModifyRelayField(req, &jsonrpc.ModifyRelayFieldArgs{RelayID: 1, Field: field, Value: "true"}, &reply)
			assert.NoError(t, err)
		}
	})

	t.Run("success nibblin fields", func(t *testing.T) {
		for _, field := range nibblinFields {
			var reply jsonrpc.ModifyRelayFieldReply
			err := svc.ModifyRelayField(req, &jsonrpc.ModifyRelayFieldArgs{RelayID: 1, Field: field, Value: "100"}, &reply)
			assert.NoError(t, err)
		}
	})

	t.Run("success public key", func(t *testing.T) {
		var reply jsonrpc.ModifyRelayFieldReply
		err := svc.ModifyRelayField(req, &jsonrpc.ModifyRelayFieldArgs{RelayID: 1, Field: "PublicKey", Value: "YFWQjOJfHfOqsCMM/1pd+c5haMhsrE2Gm05bVUQhCnG7YlPUrI/d1g=="}, &reply)
		assert.NoError(t, err)
	})

	t.Run("success relay state", func(t *testing.T) {
		var reply jsonrpc.ModifyRelayFieldReply
		err := svc.ModifyRelayField(req, &jsonrpc.ModifyRelayFieldArgs{RelayID: 1, Field: "State", Value: "enabled"}, &reply)
		assert.NoError(t, err)
	})

	t.Run("success bw rule", func(t *testing.T) {
		var reply jsonrpc.ModifyRelayFieldReply
		err := svc.ModifyRelayField(req, &jsonrpc.ModifyRelayFieldArgs{RelayID: 1, Field: "BWRule", Value: "burst"}, &reply)
		assert.NoError(t, err)
	})

	t.Run("success machine type", func(t *testing.T) {
		var reply jsonrpc.ModifyRelayFieldReply
		err := svc.ModifyRelayField(req, &jsonrpc.ModifyRelayFieldArgs{RelayID: 1, Field: "Type", Value: "vm"}, &reply)
		assert.NoError(t, err)
	})

	t.Run("success billing supplier", func(t *testing.T) {
		var reply jsonrpc.ModifyRelayFieldReply
		err := svc.ModifyRelayField(req, &jsonrpc.ModifyRelayFieldArgs{RelayID: 1, Field: "BillingSupplier", Value: "sellerID"}, &reply)
		assert.NoError(t, err)
	})
}

func TestUpdateCustomer(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}

	svc := jsonrpc.OpsService{
		Storage: &storer,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("insufficient privileges", func(t *testing.T) {
		var reply jsonrpc.UpdateCustomerReply
		err := svc.UpdateCustomer(req, &jsonrpc.UpdateCustomerArgs{}, &reply)
		assert.Equal(t, jsonrpc.JSONRPCErrorCodes[int(jsonrpc.ERROR_INSUFFICIENT_PRIVILEGES)].Message, err.Error())
	})

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
		"Admin",
	})
	req = req.WithContext(reqContext)

	err := storer.AddCustomer(context.Background(), routing.Customer{Name: "Local", Code: "local"})
	assert.NoError(t, err)

	stringFields := []string{"AutomaticSigninDomains", "Name"}

	immutableFields := []string{"Code", "BuyerRef", "SellerRef", "DatabaseID"}

	t.Run("failed immutable fields", func(t *testing.T) {
		for _, field := range immutableFields {
			var reply jsonrpc.UpdateCustomerReply
			err := svc.UpdateCustomer(req, &jsonrpc.UpdateCustomerArgs{CustomerID: "local", Field: field, Value: ""}, &reply)
			assert.Error(t, err)
			assert.EqualError(t, err, fmt.Sprintf("Field '%v' does not exist (or is not editable) on the Customer type", field))
		}
	})

	t.Run("success mutable fields", func(t *testing.T) {
		for _, field := range stringFields {
			var reply jsonrpc.UpdateCustomerReply
			err := svc.UpdateCustomer(req, &jsonrpc.UpdateCustomerArgs{CustomerID: "local", Field: field, Value: "some value"}, &reply)
			assert.NoError(t, err)
		}
	})
}

func TestRemoveCustomer(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}

	svc := jsonrpc.OpsService{
		Storage: &storer,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("insufficient privileges", func(t *testing.T) {
		var reply jsonrpc.RemoveCustomerReply
		err := svc.RemoveCustomer(req, &jsonrpc.RemoveCustomerArgs{}, &reply)
		assert.Equal(t, jsonrpc.JSONRPCErrorCodes[int(jsonrpc.ERROR_INSUFFICIENT_PRIVILEGES)].Message, err.Error())
	})

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
		"Admin",
	})
	req = req.WithContext(reqContext)

	t.Run("customer does not exist", func(t *testing.T) {
		var reply jsonrpc.RemoveCustomerReply
		err := svc.RemoveCustomer(req, &jsonrpc.RemoveCustomerArgs{CustomerCode: "0"}, &reply)
		assert.Contains(t, err.Error(), "customer with reference 0 not found")
	})

	err := storer.AddCustomer(context.Background(), routing.Customer{Name: "Local", Code: "local"})
	assert.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		var reply jsonrpc.RemoveCustomerReply
		err := svc.RemoveCustomer(req, &jsonrpc.RemoveCustomerArgs{CustomerCode: "local"}, &reply)
		assert.NoError(t, err)

		Customers := storer.Customers(context.Background())
		assert.Zero(t, len(Customers))
	})
}

func TestUpdateSeller(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}

	svc := jsonrpc.OpsService{
		Storage: &storer,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("insufficient privileges", func(t *testing.T) {
		var reply jsonrpc.UpdateSellerReply
		err := svc.UpdateSeller(req, &jsonrpc.UpdateSellerArgs{}, &reply)
		assert.Equal(t, jsonrpc.JSONRPCErrorCodes[int(jsonrpc.ERROR_INSUFFICIENT_PRIVILEGES)].Message, err.Error())
	})

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
		"Admin",
	})
	req = req.WithContext(reqContext)

	t.Run("seller does not exist", func(t *testing.T) {
		var reply jsonrpc.UpdateSellerReply
		err := svc.UpdateSeller(req, &jsonrpc.UpdateSellerArgs{SellerID: "0", Field: "ShortName", Value: "-"}, &reply)
		assert.Contains(t, err.Error(), "seller with reference 0 not found")
	})

	err := storer.AddSeller(context.Background(), routing.Seller{ID: "1", CompanyCode: "local", Name: "Local"})
	assert.NoError(t, err)

	stringFields := []string{"ShortName"}

	boolFields := []string{"Secret"}

	float64Fields := []string{"EgressPrice", "EgressPriceNibblinsPerGB"}

	immutableFields := []string{"ID", "Name", "CompanyCode", "DatabaseID", "CustomerID"}

	t.Run("failed immutable fields", func(t *testing.T) {
		for _, field := range immutableFields {
			var reply jsonrpc.UpdateSellerReply
			err := svc.UpdateSeller(req, &jsonrpc.UpdateSellerArgs{SellerID: "1", Field: field, Value: "-"}, &reply)
			assert.EqualError(t, err, fmt.Sprintf("Field '%v' does not exist (or is not editable) on the Seller type", field))
		}
	})

	t.Run("failed bool fields", func(t *testing.T) {
		for _, field := range boolFields {
			var reply jsonrpc.UpdateSellerReply
			err := svc.UpdateSeller(req, &jsonrpc.UpdateSellerArgs{SellerID: "1", Field: field, Value: "-"}, &reply)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), fmt.Sprintf("UpdateSeller() value '%s' is not a valid Secret/boolean:", "-"))
		}
	})

	t.Run("failed float64 fields", func(t *testing.T) {
		for _, field := range float64Fields {
			var reply jsonrpc.UpdateSellerReply
			err := svc.UpdateSeller(req, &jsonrpc.UpdateSellerArgs{SellerID: "1", Field: field, Value: "-"}, &reply)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), fmt.Sprintf("UpdateSeller() value '%s' is not a valid price:", "-"))
		}
	})

	t.Run("success strings fields", func(t *testing.T) {
		for _, field := range stringFields {
			var reply jsonrpc.UpdateSellerReply
			err := svc.UpdateSeller(req, &jsonrpc.UpdateSellerArgs{SellerID: "1", Field: field, Value: "-"}, &reply)
			assert.NoError(t, err)
		}
	})

	t.Run("success bool fields", func(t *testing.T) {
		for _, field := range boolFields {
			var reply jsonrpc.UpdateSellerReply
			err := svc.UpdateSeller(req, &jsonrpc.UpdateSellerArgs{SellerID: "1", Field: field, Value: "true"}, &reply)
			assert.NoError(t, err)
		}
	})

	t.Run("success float64 fields", func(t *testing.T) {
		for _, field := range float64Fields {
			var reply jsonrpc.UpdateSellerReply
			err := svc.UpdateSeller(req, &jsonrpc.UpdateSellerArgs{SellerID: "1", Field: field, Value: "1"}, &reply)
			assert.NoError(t, err)
		}
	})
}

func TestResetSellerEgressPriceOverride(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}

	svc := jsonrpc.OpsService{
		Storage: &storer,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("insufficient privileges", func(t *testing.T) {
		var reply jsonrpc.ResetSellerEgressPriceOverrideReply
		err := svc.ResetSellerEgressPriceOverride(req, &jsonrpc.ResetSellerEgressPriceOverrideArgs{}, &reply)
		assert.Equal(t, jsonrpc.JSONRPCErrorCodes[int(jsonrpc.ERROR_INSUFFICIENT_PRIVILEGES)].Message, err.Error())
	})

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
		"Admin",
	})
	req = req.WithContext(reqContext)

	seller1 := routing.Seller{
		ID:         "sellerID1",
		Name:       "seller name 1",
		ShortName:  "sname1",
		DatabaseID: 1,
	}

	seller2 := routing.Seller{
		ID:         "sellerID2",
		Name:       "seller name 2",
		ShortName:  "sname2",
		DatabaseID: 2,
	}

	datacenter := routing.Datacenter{
		ID:   crypto.HashID("datacenter name"),
		Name: "datacenter name",
	}

	relay1 := routing.Relay{
		ID:         1,
		Name:       "local.local.1",
		Seller:     seller1,
		Datacenter: datacenter,
	}

	relay2 := routing.Relay{
		ID:                  2,
		Name:                "local.local.2",
		Seller:              seller1,
		Datacenter:          datacenter,
		EgressPriceOverride: routing.Nibblin(100),
	}

	relay3 := routing.Relay{
		ID:                  3,
		Name:                "local.local.3",
		Seller:              seller2,
		Datacenter:          datacenter,
		EgressPriceOverride: routing.Nibblin(100),
	}

	err := storer.AddSeller(context.Background(), seller1)
	assert.NoError(t, err)
	err = storer.AddSeller(context.Background(), seller2)
	assert.NoError(t, err)
	err = storer.AddDatacenter(context.Background(), datacenter)
	assert.NoError(t, err)
	err = storer.AddRelay(context.Background(), relay1)
	assert.NoError(t, err)
	err = storer.AddRelay(context.Background(), relay2)
	assert.NoError(t, err)
	err = storer.AddRelay(context.Background(), relay3)
	assert.NoError(t, err)

	t.Run("unknown field", func(t *testing.T) {
		var reply jsonrpc.ResetSellerEgressPriceOverrideReply
		err := svc.ResetSellerEgressPriceOverride(req, &jsonrpc.ResetSellerEgressPriceOverrideArgs{Field: "unknown"}, &reply)
		assert.Error(t, err)
		assert.EqualError(t, err, fmt.Sprintf("Field '%s' is not a valid Relay type for resetting seller egress price override", "unknown"))
	})

	t.Run("success - only reset for seller 1", func(t *testing.T) {
		var reply jsonrpc.ResetSellerEgressPriceOverrideReply
		err := svc.ResetSellerEgressPriceOverride(req, &jsonrpc.ResetSellerEgressPriceOverrideArgs{SellerID: "sname1", Field: "EgressPriceOverride"}, &reply)
		assert.NoError(t, err)

		r1, err := storer.Relay(context.Background(), relay1.ID)
		assert.NoError(t, err)
		r2, err := storer.Relay(context.Background(), relay2.ID)
		assert.NoError(t, err)
		r3, err := storer.Relay(context.Background(), relay3.ID)
		assert.NoError(t, err)

		assert.Zero(t, r1.EgressPriceOverride)
		assert.Zero(t, r2.EgressPriceOverride)
		assert.Equal(t, r3.EgressPriceOverride, routing.Nibblin(100))
	})
}

func TestUpdateDatacenter(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}

	svc := jsonrpc.OpsService{
		Storage: &storer,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("insufficient privileges", func(t *testing.T) {
		var reply jsonrpc.UpdateDatacenterReply
		err := svc.UpdateDatacenter(req, &jsonrpc.UpdateDatacenterArgs{}, &reply)
		assert.Equal(t, jsonrpc.JSONRPCErrorCodes[int(jsonrpc.ERROR_INSUFFICIENT_PRIVILEGES)].Message, err.Error())
	})

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
		"Admin",
	})
	req = req.WithContext(reqContext)

	t.Run("bad hex datacenter ID", func(t *testing.T) {
		var reply jsonrpc.UpdateDatacenterReply
		err := svc.UpdateDatacenter(req, &jsonrpc.UpdateDatacenterArgs{HexDatacenterID: "-"}, &reply)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "UpdateDatacenter() failed to parse hex datacenter ID")
	})

	t.Run("datacenter does not exist", func(t *testing.T) {
		var reply jsonrpc.UpdateDatacenterReply
		err := svc.UpdateDatacenter(req, &jsonrpc.UpdateDatacenterArgs{HexDatacenterID: fmt.Sprintf("%016x", 0), Field: "Latitude", Value: float64(23.32)}, &reply)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "datacenter with reference 0 not found")
	})

	datacenter := routing.Datacenter{
		ID:   crypto.HashID("datacenter name"),
		Name: "datacenter name",
	}

	err := storer.AddDatacenter(context.Background(), datacenter)
	assert.NoError(t, err)

	float32Fields := []string{"Latitude", "Longitude"}
	immutableFields := []string{"ID", "Name", "AliasName", "SellerID", "DatabaseID"}

	t.Run("failed immutable fields", func(t *testing.T) {
		for _, field := range immutableFields {
			var reply jsonrpc.UpdateDatacenterReply
			err := svc.UpdateDatacenter(req, &jsonrpc.UpdateDatacenterArgs{HexDatacenterID: fmt.Sprintf("%016x", crypto.HashID("datacenter name")), Field: field, Value: float64(23.32)}, &reply)
			assert.EqualError(t, err, fmt.Sprintf("Field '%v' does not exist (or is not editable) on the Datacenter type", field))
		}
	})

	t.Run("failed float32 fields", func(t *testing.T) {
		for _, field := range float32Fields {
			var reply jsonrpc.UpdateDatacenterReply
			err := svc.UpdateDatacenter(req, &jsonrpc.UpdateDatacenterArgs{HexDatacenterID: fmt.Sprintf("%016x", crypto.HashID("datacenter name")), Field: field, Value: "-"}, &reply)
			assert.EqualError(t, err, fmt.Sprintf("UpdateDatacenter() value '%v' is not a valid float32 type", "-"))
		}
	})

	t.Run("success float32 fields", func(t *testing.T) {
		for _, field := range float32Fields {
			var reply jsonrpc.UpdateDatacenterReply
			err := svc.UpdateDatacenter(req, &jsonrpc.UpdateDatacenterArgs{HexDatacenterID: fmt.Sprintf("%016x", crypto.HashID("datacenter name")), Field: field, Value: float64(23.32)}, &reply)
			assert.NoError(t, err)
		}
	})
}
