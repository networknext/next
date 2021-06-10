package test

import (
	"encoding/binary"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/routing"
	"github.com/stretchr/testify/assert"
)

func (env *TestEnvironment) AddBuyer(companyCode string, live bool) (uint64, []byte, []byte) {
	buyer := routing.Buyer{
		RouteShader:    core.NewRouteShader(),
		InternalConfig: core.NewInternalConfig(),
	}

	publicKey, privateKey, err := crypto.GenerateCustomerKeyPair()
	assert.NoError(env.TestContext, err)

	buyerID := binary.LittleEndian.Uint64(publicKey[:8])
	publicKey = publicKey[8:]
	privateKey = privateKey[8:]

	buyer.PublicKey = publicKey
	buyer.ID = buyerID
	buyer.CompanyCode = companyCode
	buyer.Live = live

	env.DatabaseWrapper.BuyerMap[buyerID] = buyer
	env.DatabaseWrapper.DatacenterMaps[buyerID] = make(map[uint64]routing.DatacenterMap, 0)

	return buyerID, publicKey, privateKey
}

func (env *TestEnvironment) AddDCMap(buyerID uint64, datacenterID uint64, datacenterName string) {
	env.DatabaseWrapper.DatacenterMaps[buyerID][datacenterID] = routing.DatacenterMap{
		BuyerID:      buyerID,
		DatacenterID: datacenterID,
	}
}
