package jsonrpc_test

import (
	"context"
	"encoding/base64"
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

/*
// 1 customer with a buyer and a seller ID
func TestCustomersSingle(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}
	storer.AddBuyer(context.Background(), routing.Buyer{ID: 1, Name: "Fred Scuttle"})
	storer.AddSeller(context.Background(), routing.Seller{ID: "some seller", Name: "Fred Scuttle"})

	logger := log.NewNopLogger()
	svc := jsonrpc.OpsService{
		Storage: &storer,
		Logger:  logger,
	}

	t.Run("single customer", func(t *testing.T) {
		var reply jsonrpc.CustomersReply
		err := svc.Customers(nil, &jsonrpc.CustomersArgs{}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, "1", reply.Customers[0].BuyerID)
		assert.Equal(t, "some seller", reply.Customers[0].SellerID)
		assert.Equal(t, "Fred Scuttle", reply.Customers[0].Name)
	})
}

// Multiple customers with different names (2 records)
func TestCustomersMultiple(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}
	storer.AddBuyer(context.Background(), routing.Buyer{ID: 1, Name: "Fred Scuttle"})
	storer.AddSeller(context.Background(), routing.Seller{ID: "some seller", Name: "Bull Winkle"})

	logger := log.NewNopLogger()
	svc := jsonrpc.OpsService{
		Storage: &storer,
		Logger:  logger,
	}

	t.Run("multiple customers", func(t *testing.T) {
		var reply jsonrpc.CustomersReply
		err := svc.Customers(nil, &jsonrpc.CustomersArgs{}, &reply)
		assert.NoError(t, err)

		// sorted alphabetically by name
		assert.Equal(t, "", reply.Customers[0].BuyerID)
		assert.Equal(t, "some seller", reply.Customers[0].SellerID)
		assert.Equal(t, "Bull Winkle", reply.Customers[0].Name)

		assert.Equal(t, "1", reply.Customers[1].BuyerID)
		assert.Equal(t, "", reply.Customers[1].SellerID)
		assert.Equal(t, "Fred Scuttle", reply.Customers[1].Name)
	})
}

func TestAddBuyer(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}

	logger := log.NewNopLogger()
	svc := jsonrpc.OpsService{
		Storage: &storer,
		Logger:  logger,
	}

	publicKey := make([]byte, crypto.KeySize)
	_, err := rand.Read(publicKey)
	assert.NoError(t, err)

	expected := routing.Buyer{
		ID:                   1,
		Name:                 "local buyer",
		Active:               true,
		Live:                 false,
		PublicKey:            publicKey,
		RoutingRulesSettings: routing.DefaultRoutingRulesSettings,
	}

	t.Run("add", func(t *testing.T) {
		var reply jsonrpc.AddBuyerReply
		err := svc.AddBuyer(nil, &jsonrpc.AddBuyerArgs{Buyer: expected}, &reply)
		assert.NoError(t, err)

		var buyersReply jsonrpc.BuyersReply
		err = svc.Buyers(nil, &jsonrpc.BuyersArgs{}, &buyersReply)
		assert.NoError(t, err)

		assert.Len(t, buyersReply.Buyers, 1)
		assert.Equal(t, buyersReply.Buyers[0].ID, fmt.Sprintf("%x", expected.ID))
		assert.Equal(t, buyersReply.Buyers[0].Name, expected.Name)
	})

	t.Run("exists", func(t *testing.T) {
		var reply jsonrpc.AddBuyerReply

		err = svc.AddBuyer(nil, &jsonrpc.AddBuyerArgs{Buyer: expected}, &reply)
		assert.EqualError(t, err, "buyer with reference 1 already exists")
	})
}

func TestRemoveBuyer(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}

	logger := log.NewNopLogger()
	svc := jsonrpc.OpsService{
		Storage: &storer,
		Logger:  logger,
	}

	publicKey := make([]byte, crypto.KeySize)
	_, err := rand.Read(publicKey)
	assert.NoError(t, err)

	expected := routing.Buyer{
		ID:                   1,
		Name:                 "local buyer",
		Active:               true,
		Live:                 false,
		PublicKey:            publicKey,
		RoutingRulesSettings: routing.DefaultRoutingRulesSettings,
	}

	t.Run("doesn't exist", func(t *testing.T) {
		var reply jsonrpc.RemoveBuyerReply

		err = svc.RemoveBuyer(nil, &jsonrpc.RemoveBuyerArgs{ID: fmt.Sprintf("%x", expected.ID)}, &reply)
		assert.EqualError(t, err, "buyer with reference 1 not found")
	})

	t.Run("remove", func(t *testing.T) {
		var addReply jsonrpc.AddBuyerReply
		err := svc.AddBuyer(nil, &jsonrpc.AddBuyerArgs{Buyer: expected}, &addReply)
		assert.NoError(t, err)

		var reply jsonrpc.RemoveBuyerReply
		err = svc.RemoveBuyer(nil, &jsonrpc.RemoveBuyerArgs{ID: fmt.Sprintf("%x", expected.ID)}, &reply)
		assert.NoError(t, err)

		var buyersReply jsonrpc.BuyersReply
		err = svc.Buyers(nil, &jsonrpc.BuyersArgs{}, &buyersReply)
		assert.NoError(t, err)

		assert.Len(t, buyersReply.Buyers, 0)
	})
}

func TestRoutingRulesSettings(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}

	logger := log.NewNopLogger()
	svc := jsonrpc.OpsService{
		Storage: &storer,
		Logger:  logger,
	}

	t.Run("doesn't exist", func(t *testing.T) {
		var reply jsonrpc.RoutingRulesSettingsReply

		err := svc.RoutingRulesSettings(nil, &jsonrpc.RoutingRulesSettingsArgs{BuyerID: "0"}, &reply)
		assert.EqualError(t, err, "buyer with reference 0 not found")
	})

	t.Run("list", func(t *testing.T) {
		storer.AddBuyer(context.Background(), routing.Buyer{ID: 0, Name: "local.local.1", RoutingRulesSettings: routing.DefaultRoutingRulesSettings})

		var reply jsonrpc.RoutingRulesSettingsReply
		err := svc.RoutingRulesSettings(nil, &jsonrpc.RoutingRulesSettingsArgs{BuyerID: "0"}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, reply.RoutingRuleSettings[0].EnvelopeKbpsUp, routing.DefaultRoutingRulesSettings.EnvelopeKbpsUp)
		assert.Equal(t, reply.RoutingRuleSettings[0].EnvelopeKbpsDown, routing.DefaultRoutingRulesSettings.EnvelopeKbpsDown)
		assert.Equal(t, reply.RoutingRuleSettings[0].Mode, routing.DefaultRoutingRulesSettings.Mode)
		assert.Equal(t, reply.RoutingRuleSettings[0].MaxCentsPerGB, routing.DefaultRoutingRulesSettings.MaxCentsPerGB)
		assert.Equal(t, reply.RoutingRuleSettings[0].RTTEpsilon, routing.DefaultRoutingRulesSettings.RTTEpsilon)
		assert.Equal(t, reply.RoutingRuleSettings[0].RTTThreshold, routing.DefaultRoutingRulesSettings.RTTThreshold)
		assert.Equal(t, reply.RoutingRuleSettings[0].RTTHysteresis, routing.DefaultRoutingRulesSettings.RTTHysteresis)
		assert.Equal(t, reply.RoutingRuleSettings[0].RTTVeto, routing.DefaultRoutingRulesSettings.RTTVeto)
		assert.Equal(t, reply.RoutingRuleSettings[0].EnableYouOnlyLiveOnce, routing.DefaultRoutingRulesSettings.EnableYouOnlyLiveOnce)
		assert.Equal(t, reply.RoutingRuleSettings[0].EnablePacketLossSafety, routing.DefaultRoutingRulesSettings.EnablePacketLossSafety)
		assert.Equal(t, reply.RoutingRuleSettings[0].EnableMultipathForPacketLoss, routing.DefaultRoutingRulesSettings.EnableMultipathForPacketLoss)
		assert.Equal(t, reply.RoutingRuleSettings[0].EnableMultipathForJitter, routing.DefaultRoutingRulesSettings.EnableMultipathForJitter)
		assert.Equal(t, reply.RoutingRuleSettings[0].EnableMultipathForRTT, routing.DefaultRoutingRulesSettings.EnableMultipathForRTT)
		assert.Equal(t, reply.RoutingRuleSettings[0].EnableABTest, routing.DefaultRoutingRulesSettings.EnableABTest)
		assert.Equal(t, reply.RoutingRuleSettings[0].EnableTryBeforeYouBuy, routing.DefaultRoutingRulesSettings.EnableTryBeforeYouBuy)
		assert.Equal(t, reply.RoutingRuleSettings[0].TryBeforeYouBuyMaxSlices, routing.DefaultRoutingRulesSettings.TryBeforeYouBuyMaxSlices)
	})
}

func TestSetRoutingRulesSettings(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}

	logger := log.NewNopLogger()
	svc := jsonrpc.OpsService{
		Storage: &storer,
		Logger:  logger,
	}

	t.Run("doesn't exist", func(t *testing.T) {
		var reply jsonrpc.SetRoutingRulesSettingsReply

		err := svc.SetRoutingRulesSettings(nil, &jsonrpc.SetRoutingRulesSettingsArgs{BuyerID: "0", RoutingRulesSettings: routing.LocalRoutingRulesSettings}, &reply)
		assert.EqualError(t, err, "SetRoutingRulesSettings() Storage.Buyer error: buyer with reference 0 not found")
	})

	t.Run("set", func(t *testing.T) {
		storer.AddBuyer(context.Background(), routing.Buyer{ID: 1, Name: "local.local.1", RoutingRulesSettings: routing.DefaultRoutingRulesSettings})

		var reply jsonrpc.SetRoutingRulesSettingsReply
		err := svc.SetRoutingRulesSettings(nil, &jsonrpc.SetRoutingRulesSettingsArgs{BuyerID: "1", RoutingRulesSettings: routing.LocalRoutingRulesSettings}, &reply)
		assert.NoError(t, err)

		var rrsReply jsonrpc.RoutingRulesSettingsReply
		err = svc.RoutingRulesSettings(nil, &jsonrpc.RoutingRulesSettingsArgs{BuyerID: "1"}, &rrsReply)
		assert.NoError(t, err)

		assert.Equal(t, rrsReply.RoutingRuleSettings[0].EnvelopeKbpsUp, routing.LocalRoutingRulesSettings.EnvelopeKbpsUp)
		assert.Equal(t, rrsReply.RoutingRuleSettings[0].EnvelopeKbpsDown, routing.LocalRoutingRulesSettings.EnvelopeKbpsDown)
		assert.Equal(t, rrsReply.RoutingRuleSettings[0].Mode, routing.LocalRoutingRulesSettings.Mode)
		assert.Equal(t, rrsReply.RoutingRuleSettings[0].MaxCentsPerGB, routing.LocalRoutingRulesSettings.MaxCentsPerGB)
		assert.Equal(t, rrsReply.RoutingRuleSettings[0].RTTEpsilon, routing.LocalRoutingRulesSettings.RTTEpsilon)
		assert.Equal(t, rrsReply.RoutingRuleSettings[0].RTTThreshold, routing.LocalRoutingRulesSettings.RTTThreshold)
		assert.Equal(t, rrsReply.RoutingRuleSettings[0].RTTHysteresis, routing.LocalRoutingRulesSettings.RTTHysteresis)
		assert.Equal(t, rrsReply.RoutingRuleSettings[0].RTTVeto, routing.LocalRoutingRulesSettings.RTTVeto)
		assert.Equal(t, rrsReply.RoutingRuleSettings[0].EnableYouOnlyLiveOnce, routing.LocalRoutingRulesSettings.EnableYouOnlyLiveOnce)
		assert.Equal(t, rrsReply.RoutingRuleSettings[0].EnablePacketLossSafety, routing.LocalRoutingRulesSettings.EnablePacketLossSafety)
		assert.Equal(t, rrsReply.RoutingRuleSettings[0].EnableMultipathForPacketLoss, routing.LocalRoutingRulesSettings.EnableMultipathForPacketLoss)
		assert.Equal(t, rrsReply.RoutingRuleSettings[0].EnableMultipathForJitter, routing.LocalRoutingRulesSettings.EnableMultipathForJitter)
		assert.Equal(t, rrsReply.RoutingRuleSettings[0].EnableMultipathForRTT, routing.LocalRoutingRulesSettings.EnableMultipathForRTT)
		assert.Equal(t, rrsReply.RoutingRuleSettings[0].EnableABTest, routing.LocalRoutingRulesSettings.EnableABTest)
		assert.Equal(t, rrsReply.RoutingRuleSettings[0].EnableTryBeforeYouBuy, routing.DefaultRoutingRulesSettings.EnableTryBeforeYouBuy)
	})
}

func TestSellers(t *testing.T) {
	t.Parallel()

	expected := routing.Seller{
		ID:                "1",
		Name:              "local.local.1",
		IngressPriceCents: 10,
		EgressPriceCents:  20,
	}

	storer := storage.InMemory{}
	storer.AddSeller(context.Background(), expected)

	logger := log.NewNopLogger()
	svc := jsonrpc.OpsService{
		Storage: &storer,
		Logger:  logger,
	}

	t.Run("list", func(t *testing.T) {
		var reply jsonrpc.SellersReply
		err := svc.Sellers(nil, &jsonrpc.SellersArgs{}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, reply.Sellers[0].ID, expected.ID)
		assert.Equal(t, reply.Sellers[0].Name, expected.Name)
		assert.Equal(t, reply.Sellers[0].IngressPriceCents, expected.IngressPriceCents)
		assert.Equal(t, reply.Sellers[0].EgressPriceCents, expected.EgressPriceCents)
	})
}

func TestAddSeller(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}

	logger := log.NewNopLogger()
	svc := jsonrpc.OpsService{
		Storage: &storer,
		Logger:  logger,
	}

	expected := routing.Seller{
		ID:                "id",
		Name:              "local seller",
		IngressPriceCents: 10,
		EgressPriceCents:  20,
	}

	t.Run("add", func(t *testing.T) {
		var reply jsonrpc.AddSellerReply
		err := svc.AddSeller(nil, &jsonrpc.AddSellerArgs{Seller: expected}, &reply)
		assert.NoError(t, err)

		var sellersReply jsonrpc.SellersReply
		err = svc.Sellers(nil, &jsonrpc.SellersArgs{}, &sellersReply)
		assert.NoError(t, err)

		assert.Len(t, sellersReply.Sellers, 1)
		assert.Equal(t, sellersReply.Sellers[0].ID, expected.ID)
		assert.Equal(t, sellersReply.Sellers[0].Name, expected.Name)
		assert.Equal(t, sellersReply.Sellers[0].IngressPriceCents, expected.IngressPriceCents)
		assert.Equal(t, sellersReply.Sellers[0].EgressPriceCents, expected.EgressPriceCents)
	})

	t.Run("exists", func(t *testing.T) {
		var reply jsonrpc.AddSellerReply

		err := svc.AddSeller(nil, &jsonrpc.AddSellerArgs{Seller: expected}, &reply)
		assert.EqualError(t, err, "AddSeller() error: seller with reference id already exists")
	})
}

func TestRemoveSeller(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}

	logger := log.NewNopLogger()
	svc := jsonrpc.OpsService{
		Storage: &storer,
		Logger:  logger,
	}

	expected := routing.Seller{
		ID:                "1",
		Name:              "local seller",
		IngressPriceCents: 10,
		EgressPriceCents:  20,
	}

	t.Run("doesn't exist", func(t *testing.T) {
		var reply jsonrpc.RemoveSellerReply

		err := svc.RemoveSeller(nil, &jsonrpc.RemoveSellerArgs{ID: expected.ID}, &reply)
		assert.EqualError(t, err, "RemoveSeller() error: seller with reference 1 not found")
	})

	t.Run("remove", func(t *testing.T) {
		var addReply jsonrpc.AddSellerReply
		err := svc.AddSeller(nil, &jsonrpc.AddSellerArgs{Seller: expected}, &addReply)
		assert.NoError(t, err)

		var reply jsonrpc.RemoveSellerReply
		err = svc.RemoveSeller(nil, &jsonrpc.RemoveSellerArgs{ID: expected.ID}, &reply)
		assert.NoError(t, err)

		var sellersReply jsonrpc.SellersReply
		err = svc.Sellers(nil, &jsonrpc.SellersArgs{}, &sellersReply)
		assert.NoError(t, err)

		assert.Len(t, sellersReply.Sellers, 0)
	})
}

func TestRelays(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}

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
		Name:       "local.local.23",
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

	redisServer, err := miniredis.Run()
	assert.NoError(t, err)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	logger := log.NewNopLogger()
	svc := jsonrpc.OpsService{
		Storage:     &storer,
		RedisClient: redisClient,
		Logger:      logger,
	}

	t.Run("list", func(t *testing.T) {
		var reply jsonrpc.RelaysReply
		err := svc.Relays(nil, &jsonrpc.RelaysArgs{}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, reply.Relays[0].ID, uint64(1))
		assert.Equal(t, reply.Relays[0].Name, "local.local.1")
		assert.Equal(t, reply.Relays[1].ID, uint64(2))
		assert.Equal(t, reply.Relays[1].Name, "local.local.2")
		assert.Equal(t, reply.Relays[2].ID, uint64(3))
		assert.Equal(t, reply.Relays[2].Name, "local.local.23")
	})

	t.Run("exact match", func(t *testing.T) {
		var reply jsonrpc.RelaysReply
		err := svc.Relays(nil, &jsonrpc.RelaysArgs{Regex: "local.local.2"}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, len(reply.Relays), 1)
		assert.Equal(t, reply.Relays[0].ID, uint64(2))
		assert.Equal(t, reply.Relays[0].Name, "local.local.2")

		var empty jsonrpc.RelaysReply
		err = svc.Relays(nil, &jsonrpc.RelaysArgs{Regex: "not.found"}, &empty)
		assert.NoError(t, err)

		assert.Equal(t, len(empty.Relays), 0)
	})

	t.Run("filter", func(t *testing.T) {
		var reply jsonrpc.RelaysReply
		err := svc.Relays(nil, &jsonrpc.RelaysArgs{Regex: "local.1"}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, len(reply.Relays), 1)
		assert.Equal(t, reply.Relays[0].ID, uint64(1))
		assert.Equal(t, reply.Relays[0].Name, "local.local.1")

		var empty jsonrpc.RelaysReply
		err = svc.Relays(nil, &jsonrpc.RelaysArgs{Regex: "not.found"}, &empty)
		assert.NoError(t, err)

		assert.Equal(t, len(empty.Relays), 0)
	})

	t.Run("filter by seller", func(t *testing.T) {
		var reply jsonrpc.RelaysReply
		err := svc.Relays(nil, &jsonrpc.RelaysArgs{Regex: "seller name"}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, len(reply.Relays), 3)
		assert.Equal(t, reply.Relays[0].ID, uint64(1))
		assert.Equal(t, reply.Relays[0].Name, "local.local.1")
		assert.Equal(t, reply.Relays[1].ID, uint64(2))
		assert.Equal(t, reply.Relays[1].Name, "local.local.2")
		assert.Equal(t, reply.Relays[2].ID, uint64(3))
		assert.Equal(t, reply.Relays[2].Name, "local.local.23")

		var empty jsonrpc.RelaysReply
		err = svc.Relays(nil, &jsonrpc.RelaysArgs{Regex: "not.found"}, &empty)
		assert.NoError(t, err)

		assert.Equal(t, len(empty.Relays), 0)
	})
}

func TestAddRelay(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}

	redisServer, err := miniredis.Run()
	assert.NoError(t, err)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	logger := log.NewNopLogger()
	svc := jsonrpc.OpsService{
		Storage:     &storer,
		RedisClient: redisClient,
		Logger:      logger,
	}

	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
	assert.NoError(t, err)

	expected := routing.Relay{
		ID:   crypto.HashID(addr.String()),
		Name: "local relay",
		Addr: *addr,
	}

	t.Run("seller doesn't exist", func(t *testing.T) {
		var reply jsonrpc.AddRelayReply
		err := svc.AddRelay(nil, &jsonrpc.AddRelayArgs{Relay: expected}, &reply)
		assert.EqualError(t, err, "AddRelay() error: seller with reference  not found")
	})

	t.Run("datacenter doesn't exist", func(t *testing.T) {
		expected.Seller = routing.Seller{
			ID:                "sellerID",
			Name:              "seller name",
			IngressPriceCents: 10,
			EgressPriceCents:  20,
		}

		var sellerReply jsonrpc.AddSellerReply
		err := svc.AddSeller(nil, &jsonrpc.AddSellerArgs{Seller: expected.Seller}, &sellerReply)
		assert.NoError(t, err)

		var reply jsonrpc.AddRelayReply
		err = svc.AddRelay(nil, &jsonrpc.AddRelayArgs{Relay: expected}, &reply)
		assert.EqualError(t, err, "AddRelay() error: datacenter with reference 0 not found")
	})

	t.Run("add", func(t *testing.T) {
		expected.Datacenter = routing.Datacenter{
			ID:       crypto.HashID("datacenter name"),
			Name:     "datacenter name",
			Enabled:  true,
			Location: routing.LocationNullIsland,
		}

		var datacenterReply jsonrpc.AddDatacenterReply
		err := svc.AddDatacenter(nil, &jsonrpc.AddDatacenterArgs{Datacenter: expected.Datacenter}, &datacenterReply)
		assert.NoError(t, err)

		var reply jsonrpc.AddRelayReply
		err = svc.AddRelay(nil, &jsonrpc.AddRelayArgs{Relay: expected}, &reply)
		assert.NoError(t, err)

		var relaysReply jsonrpc.RelaysReply
		err = svc.Relays(nil, &jsonrpc.RelaysArgs{}, &relaysReply)
		assert.NoError(t, err)

		assert.Len(t, relaysReply.Relays, 1)
		assert.Equal(t, relaysReply.Relays[0].ID, expected.ID)
		assert.Equal(t, relaysReply.Relays[0].Name, expected.Name)
		assert.Equal(t, relaysReply.Relays[0].Addr, expected.Addr.String())
	})

	t.Run("exists", func(t *testing.T) {
		var reply jsonrpc.AddRelayReply

		err = svc.AddRelay(nil, &jsonrpc.AddRelayArgs{Relay: expected}, &reply)
		assert.EqualError(t, err, fmt.Sprintf("AddRelay() error: relay with reference %d already exists", expected.ID))
	})
}

func TestRemoveRelay(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}

	redisServer, err := miniredis.Run()
	assert.NoError(t, err)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	logger := log.NewNopLogger()
	svc := jsonrpc.OpsService{
		Storage:     &storer,
		RedisClient: redisClient,
		Logger:      logger,
	}

	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
	assert.NoError(t, err)

	seller := routing.Seller{
		ID:   "sellerID",
		Name: "seller name",
	}

	datacenter := routing.Datacenter{
		ID:   crypto.HashID("datacenter name"),
		Name: "datacenter name",
	}

	expected := routing.Relay{
		ID:         crypto.HashID(addr.String()),
		Name:       "local relay",
		Addr:       *addr,
		Seller:     seller,
		Datacenter: datacenter,
	}

	svc.AddSeller(nil, &jsonrpc.AddSellerArgs{Seller: seller}, &jsonrpc.AddSellerReply{})
	svc.AddDatacenter(nil, &jsonrpc.AddDatacenterArgs{Datacenter: datacenter}, &jsonrpc.AddDatacenterReply{})

	t.Run("doesn't exist", func(t *testing.T) {
		var reply jsonrpc.RemoveRelayReply

		err = svc.RemoveRelay(nil, &jsonrpc.RemoveRelayArgs{RelayID: expected.ID}, &reply)
		assert.EqualError(t, err, fmt.Sprintf("RemoveRelay() Storage.Relay error: relay with reference %d not found", expected.ID))
	})

	t.Run("remove", func(t *testing.T) {
		var addReply jsonrpc.AddRelayReply
		err := svc.AddRelay(nil, &jsonrpc.AddRelayArgs{Relay: expected}, &addReply)
		assert.NoError(t, err)

		var reply jsonrpc.RemoveRelayReply
		err = svc.RemoveRelay(nil, &jsonrpc.RemoveRelayArgs{RelayID: expected.ID}, &reply)
		assert.NoError(t, err)

		var relaysReply jsonrpc.RelaysReply
		err = svc.Relays(nil, &jsonrpc.RelaysArgs{}, &relaysReply)
		assert.NoError(t, err)

		// Remove shouldn't actually remove it anymore, just set the state to decommissioned
		assert.Len(t, relaysReply.Relays, 1)
		assert.Equal(t, relaysReply.Relays[0].ID, expected.ID)
		assert.Equal(t, relaysReply.Relays[0].State, routing.RelayStateDecommissioned.String())
	})
}

func TestRelayStateUpdate(t *testing.T) {
	t.Parallel()

	logger := log.NewNopLogger()
	makeSvc := func() *jsonrpc.OpsService {
		var storer storage.InMemory

		seller := routing.Seller{
			ID:   "sellerID",
			Name: "seller name",
		}

		datacenter := routing.Datacenter{
			ID:   crypto.HashID("datacenter name"),
			Name: "datacenter name",
		}

		storer.AddSeller(context.Background(), seller)
		storer.AddDatacenter(context.Background(), datacenter)

		err := storer.AddRelay(context.Background(), routing.Relay{
			ID:         1,
			State:      0,
			Seller:     seller,
			Datacenter: datacenter,
		})
		assert.NoError(t, err)
		err = storer.AddRelay(context.Background(), routing.Relay{
			ID:         2,
			State:      123456,
			Seller:     seller,
			Datacenter: datacenter,
		})
		assert.NoError(t, err)

		return &jsonrpc.OpsService{
			Storage: &storer,
			Logger:  logger,
		}
	}

	t.Run("found", func(t *testing.T) {
		svc := makeSvc()
		err := svc.RelayStateUpdate(nil, &jsonrpc.RelayStateUpdateArgs{
			RelayID:    1,
			RelayState: routing.RelayStateDisabled,
		}, &jsonrpc.RelayStateUpdateReply{})
		assert.NoError(t, err)

		relay, err := svc.Storage.Relay(1)
		assert.NoError(t, err)
		assert.Equal(t, routing.RelayStateDisabled, relay.State)

		relay, err = svc.Storage.Relay(2)
		assert.NoError(t, err)
		assert.Equal(t, routing.RelayState(123456), relay.State)
	})

	t.Run("not found", func(t *testing.T) {
		svc := makeSvc()
		err := svc.RelayStateUpdate(nil, &jsonrpc.RelayStateUpdateArgs{
			RelayID:    987654321,
			RelayState: routing.RelayStateDisabled,
		}, &jsonrpc.RelayStateUpdateReply{})
		assert.Error(t, err)

		relay, err := svc.Storage.Relay(1)
		assert.NoError(t, err)
		assert.Equal(t, routing.RelayState(0), relay.State)

		relay, err = svc.Storage.Relay(2)
		assert.NoError(t, err)
		assert.Equal(t, routing.RelayState(123456), relay.State)
	})
}

func TestRelayPublicKeyUpdate(t *testing.T) {
	t.Parallel()

	logger := log.NewNopLogger()

	makeSvc := func() *jsonrpc.OpsService {
		var storer storage.InMemory

		seller := routing.Seller{
			ID:   "sellerID",
			Name: "seller name",
		}

		datacenter := routing.Datacenter{
			ID:   crypto.HashID("datacenter name"),
			Name: "datacenter name",
		}

		storer.AddSeller(context.Background(), seller)
		storer.AddDatacenter(context.Background(), datacenter)

		err := storer.AddRelay(context.Background(), routing.Relay{
			ID:         1,
			PublicKey:  []byte("oldpublickey"),
			Seller:     seller,
			Datacenter: datacenter,
		})
		assert.NoError(t, err)
		err = storer.AddRelay(context.Background(), routing.Relay{
			ID:         2,
			PublicKey:  []byte("oldpublickey"),
			Seller:     seller,
			Datacenter: datacenter,
		})
		assert.NoError(t, err)

		return &jsonrpc.OpsService{
			Storage: &storer,
			Logger:  logger,
		}
	}

	t.Run("found", func(t *testing.T) {
		svc := makeSvc()
		err := svc.RelayPublicKeyUpdate(nil, &jsonrpc.RelayPublicKeyUpdateArgs{
			RelayID:        1,
			RelayPublicKey: "newpublickey",
		}, &jsonrpc.RelayPublicKeyUpdateReply{})
		assert.NoError(t, err)

		relay, err := svc.Storage.Relay(1)
		assert.NoError(t, err)
		assert.Equal(t, "newpublickey", base64.StdEncoding.EncodeToString(relay.PublicKey))

		relay, err = svc.Storage.Relay(2)
		assert.NoError(t, err)
		assert.Equal(t, []byte("oldpublickey"), relay.PublicKey)
	})

	t.Run("not found", func(t *testing.T) {
		svc := makeSvc()
		err := svc.RelayPublicKeyUpdate(nil, &jsonrpc.RelayPublicKeyUpdateArgs{
			RelayID:        987654321,
			RelayPublicKey: "newpublickey",
		}, &jsonrpc.RelayPublicKeyUpdateReply{})
		assert.Error(t, err)

		relay, err := svc.Storage.Relay(1)
		assert.NoError(t, err)
		assert.Equal(t, []byte("oldpublickey"), relay.PublicKey)

		relay, err = svc.Storage.Relay(2)
		assert.NoError(t, err)
		assert.Equal(t, []byte("oldpublickey"), relay.PublicKey)
	})
}

func TestRelayNICSpeedUpdate(t *testing.T) {
	t.Parallel()

	logger := log.NewNopLogger()

	makeSvc := func() *jsonrpc.OpsService {
		var storer storage.InMemory

		seller := routing.Seller{
			ID:   "sellerID",
			Name: "seller name",
		}

		datacenter := routing.Datacenter{
			ID:   crypto.HashID("datacenter name"),
			Name: "datacenter name",
		}

		storer.AddSeller(context.Background(), seller)
		storer.AddDatacenter(context.Background(), datacenter)

		err := storer.AddRelay(context.Background(), routing.Relay{
			ID:           1,
			NICSpeedMbps: 1000,
			Seller:       seller,
			Datacenter:   datacenter,
		})
		assert.NoError(t, err)
		err = storer.AddRelay(context.Background(), routing.Relay{
			ID:           2,
			NICSpeedMbps: 2000,
			Seller:       seller,
			Datacenter:   datacenter,
		})
		assert.NoError(t, err)

		return &jsonrpc.OpsService{
			Storage: &storer,
			Logger:  logger,
		}
	}

	t.Run("found", func(t *testing.T) {
		svc := makeSvc()
		err := svc.RelayNICSpeedUpdate(nil, &jsonrpc.RelayNICSpeedUpdateArgs{
			RelayID:       1,
			RelayNICSpeed: 10000,
		}, &jsonrpc.RelayNICSpeedUpdateReply{})
		assert.NoError(t, err)

		relay, err := svc.Storage.Relay(1)
		assert.NoError(t, err)
		assert.Equal(t, uint64(10000), relay.NICSpeedMbps)

		relay, err = svc.Storage.Relay(2)
		assert.NoError(t, err)
		assert.Equal(t, uint64(2000), relay.NICSpeedMbps)
	})

	t.Run("not found", func(t *testing.T) {
		svc := makeSvc()
		err := svc.RelayNICSpeedUpdate(nil, &jsonrpc.RelayNICSpeedUpdateArgs{
			RelayID:       987654321,
			RelayNICSpeed: 10000,
		}, &jsonrpc.RelayNICSpeedUpdateReply{})
		assert.Error(t, err)

		relay, err := svc.Storage.Relay(1)
		assert.NoError(t, err)
		assert.Equal(t, uint64(1000), relay.NICSpeedMbps)

		relay, err = svc.Storage.Relay(2)
		assert.NoError(t, err)
		assert.Equal(t, uint64(2000), relay.NICSpeedMbps)
	})
}

func TestDatacenters(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}
	storer.AddDatacenter(context.Background(), routing.Datacenter{ID: 1, Name: "local.local.1"})
	storer.AddDatacenter(context.Background(), routing.Datacenter{ID: 2, Name: "local.local.2"})

	logger := log.NewNopLogger()
	svc := jsonrpc.OpsService{
		Storage: &storer,
		Logger:  logger,
	}

	t.Run("list", func(t *testing.T) {
		var reply jsonrpc.DatacentersReply
		err := svc.Datacenters(nil, &jsonrpc.DatacentersArgs{}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, reply.Datacenters[0].Name, "local.local.1")
		assert.Equal(t, reply.Datacenters[1].Name, "local.local.2")
	})

	t.Run("filter", func(t *testing.T) {
		var reply jsonrpc.DatacentersReply
		err := svc.Datacenters(nil, &jsonrpc.DatacentersArgs{Name: "local.1"}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, len(reply.Datacenters), 1)
		assert.Equal(t, reply.Datacenters[0].Name, "local.local.1")

		var empty jsonrpc.DatacentersReply
		err = svc.Datacenters(nil, &jsonrpc.DatacentersArgs{Name: "not.found"}, &empty)
		assert.NoError(t, err)

		assert.Equal(t, len(empty.Datacenters), 0)
	})
}

func TestAddDatacenter(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}

	logger := log.NewNopLogger()
	svc := jsonrpc.OpsService{
		Storage: &storer,
		Logger:  logger,
	}

	expected := routing.Datacenter{
		ID:      1,
		Name:    "local datacenter",
		Enabled: false,
		Location: routing.Location{
			Latitude:  70.5,
			Longitude: 120.5,
		},
	}

	t.Run("add", func(t *testing.T) {
		var reply jsonrpc.AddDatacenterReply
		err := svc.AddDatacenter(nil, &jsonrpc.AddDatacenterArgs{Datacenter: expected}, &reply)
		assert.NoError(t, err)

		var datacentersReply jsonrpc.DatacentersReply
		err = svc.Datacenters(nil, &jsonrpc.DatacentersArgs{}, &datacentersReply)
		assert.NoError(t, err)

		assert.Len(t, datacentersReply.Datacenters, 1)
		assert.Equal(t, datacentersReply.Datacenters[0].Name, expected.Name)
		assert.Equal(t, datacentersReply.Datacenters[0].Latitude, expected.Location.Latitude)
		assert.Equal(t, datacentersReply.Datacenters[0].Longitude, expected.Location.Longitude)
		assert.Equal(t, datacentersReply.Datacenters[0].Enabled, expected.Enabled)
	})

	t.Run("exists", func(t *testing.T) {
		var reply jsonrpc.AddDatacenterReply

		err := svc.AddDatacenter(nil, &jsonrpc.AddDatacenterArgs{Datacenter: expected}, &reply)
		assert.EqualError(t, err, "AddDatacenter() error: datacenter with reference 1 already exists")
	})
}

func TestRemoveDatacenter(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}

	logger := log.NewNopLogger()
	svc := jsonrpc.OpsService{
		Storage: &storer,
		Logger:  logger,
	}

	expected := routing.Datacenter{
		ID:      crypto.HashID("local datacenter"),
		Name:    "local datacenter",
		Enabled: false,
		Location: routing.Location{
			Latitude:  70.5,
			Longitude: 120.5,
		},
	}

	t.Run("doesn't exist", func(t *testing.T) {
		var reply jsonrpc.RemoveDatacenterReply

		err := svc.RemoveDatacenter(nil, &jsonrpc.RemoveDatacenterArgs{Name: expected.Name}, &reply)
		assert.EqualError(t, err, fmt.Sprintf("RemoveDatacenter() error: datacenter with reference %d not found", expected.ID))
	})

	t.Run("remove", func(t *testing.T) {
		var addReply jsonrpc.AddDatacenterReply
		err := svc.AddDatacenter(nil, &jsonrpc.AddDatacenterArgs{Datacenter: expected}, &addReply)
		assert.NoError(t, err)

		var reply jsonrpc.RemoveDatacenterReply
		err = svc.RemoveDatacenter(nil, &jsonrpc.RemoveDatacenterArgs{Name: expected.Name}, &reply)
		assert.NoError(t, err)

		var datacentersReply jsonrpc.DatacentersReply
		err = svc.Datacenters(nil, &jsonrpc.DatacentersArgs{}, &datacentersReply)
		assert.NoError(t, err)

		assert.Len(t, datacentersReply.Datacenters, 0)
	})
}
*/
