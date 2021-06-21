package test

import (
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/routing"
)

func (env *TestEnvironment) AddDatacenter(datacenterName string) routing.Datacenter {
	datacenter := routing.Datacenter{
		ID:   crypto.HashID(datacenterName),
		Name: datacenterName,
	}

	env.DatabaseWrapper.DatacenterMap[datacenter.ID] = datacenter

	return datacenter
}
