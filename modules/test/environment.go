package test

import (
	"testing"

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

func (env *TestEnvironment) Clear() {
	numRelays := len(env.relays)
	env.cost = make([][]int32, numRelays)
	for i := 0; i < numRelays; i++ {
		env.cost[i] = make([]int32, numRelays)
		for j := 0; j < numRelays; j++ {
			env.cost[i][j] = -1
		}
	}
}

func (env *TestEnvironment) GetDatabaseWrapper() *routing.DatabaseBinWrapper {
	return env.DatabaseWrapper
}
