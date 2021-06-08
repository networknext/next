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

func (env *TestEnvironment) GetDatabaseWrapper() *routing.DatabaseBinWrapper {
	return env.DatabaseWrapper
}
