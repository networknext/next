package test

import (
	"net"
	"testing"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/routing"
)

type TestEnvironment struct {
	TestContext     *testing.T
	relayArray      []*TestRelayData
	relays          map[string]*TestRelayData
	cost            [][]int32
	DatabaseWrapper *routing.DatabaseBinWrapper
	MetricsHandler  metrics.LocalHandler
}

func NewTestEnvironment(t *testing.T) *TestEnvironment {
	env := &TestEnvironment{
		TestContext:     t,
		relays:          make(map[string]*TestRelayData),
		DatabaseWrapper: routing.CreateEmptyDatabaseBinWrapper(),
		MetricsHandler:  metrics.LocalHandler{},
	}
	return env
}

func (env *TestEnvironment) GetDatabaseWrapper() *routing.DatabaseBinWrapper {
	return env.DatabaseWrapper
}

func (env *TestEnvironment) GetMagicValues() ([8]byte, [8]byte, [8]byte) {
	var upcoming [8]byte
	var current [8]byte
	var previous [8]byte

	return upcoming, current, previous
}

func (env *TestEnvironment) GetBackendLoadBalancerIP() *net.UDPAddr {
	return core.ParseAddress("127.0.0.1:40000")
}
