package storage

import (
	"sync"

	"github.com/networknext/backend/modules/routing"
)

type LocalMultipathVetoHandler struct {
	getDatabase                func() *routing.DatabaseBinWrapper
	cachedMultipathVetoes      map[string]map[uint64]bool
	cachedMultipathVetoesMutex sync.RWMutex
}

func NewLocalMultipathVetoHandler(redisHost string, getDatabase func() *routing.DatabaseBinWrapper) (*LocalMultipathVetoHandler, error) {
	return &LocalMultipathVetoHandler{
		getDatabase:           getDatabase,
		cachedMultipathVetoes: make(map[string]map[uint64]bool),
	}, nil
}

func (lmvh *LocalMultipathVetoHandler) GetMapCopy(companyCode string) map[uint64]bool {
	lmvh.cachedMultipathVetoesMutex.RLock()
	defer lmvh.cachedMultipathVetoesMutex.RUnlock()

	multipathVetoMap, ok := lmvh.cachedMultipathVetoes[companyCode]
	if !ok {
		return make(map[uint64]bool)
	}

	multipathVetoMapCopy := make(map[uint64]bool)
	for k, v := range multipathVetoMap {
		multipathVetoMapCopy[k] = v
	}

	return multipathVetoMapCopy
}

func (lmvh *LocalMultipathVetoHandler) MultipathVetoUser(companyCode string, userHash uint64) error {
	lmvh.cachedMultipathVetoesMutex.Lock()
	if _, ok := lmvh.cachedMultipathVetoes[companyCode]; !ok {
		lmvh.cachedMultipathVetoes[companyCode] = map[uint64]bool{}
	}
	lmvh.cachedMultipathVetoes[companyCode][userHash] = true
	lmvh.cachedMultipathVetoesMutex.Unlock()
	return nil
}

func (lmvh *LocalMultipathVetoHandler) Sync() error {
	// Local handler doesn't need to sync
	return nil
}
