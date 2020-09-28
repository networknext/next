package storage

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/gomodule/redigo/redis"
)

type MultipathVetoHandler struct {
	storer                     Storer
	redisPool                  *redis.Pool
	cachedMultipathVetoes      map[string]map[uint64]bool
	cachedMultipathVetoesMutex sync.RWMutex
}

func NewMultipathVetoHandler(redisHost string, storer Storer) (*MultipathVetoHandler, error) {
	redisPool := NewRedisPool(redisHost, 5, 64)
	conn := redisPool.Get()
	defer conn.Close()

	if _, err := conn.Do("PING"); err != nil {
		return nil, fmt.Errorf("could not ping multipath veto redis instance: %v", err)
	}

	return &MultipathVetoHandler{
		storer:                storer,
		redisPool:             redisPool,
		cachedMultipathVetoes: make(map[string]map[uint64]bool),
	}, nil
}

func (mvh *MultipathVetoHandler) GetMapCopy(companyCode string) map[uint64]bool {
	mvh.cachedMultipathVetoesMutex.RLock()
	defer mvh.cachedMultipathVetoesMutex.RUnlock()

	multipathVetoMap, ok := mvh.cachedMultipathVetoes[companyCode]
	if !ok {
		return make(map[uint64]bool)
	}

	multipathVetoMapCopy := make(map[uint64]bool)
	for k, v := range multipathVetoMap {
		multipathVetoMapCopy[k] = v
	}

	return multipathVetoMapCopy
}

func (mvh *MultipathVetoHandler) MultipathVetoUser(companyCode string, userHash uint64) error {
	conn := mvh.redisPool.Get()
	defer conn.Close()

	result, err := redis.String(conn.Do("SET", companyCode+"-"+fmt.Sprintf("%016x", userHash), "1", "EX", "604800")) // 7 days
	if err != nil {
		return fmt.Errorf("failed setting multipath veto on user %016x for buyer %s: %v", userHash, companyCode, err)
	}

	if result != "OK" {
		return fmt.Errorf("failed setting multipath veto on user %016x for buyer %s: set command returned false", userHash, companyCode)
	}

	mvh.cachedMultipathVetoesMutex.Lock()
	if _, ok := mvh.cachedMultipathVetoes[companyCode]; !ok {
		mvh.cachedMultipathVetoes[companyCode] = map[uint64]bool{}
	}
	mvh.cachedMultipathVetoes[companyCode][userHash] = true
	mvh.cachedMultipathVetoesMutex.Unlock()
	return nil
}

func (mvh *MultipathVetoHandler) Sync() error {
	newMultipathVetoMap := map[string]map[uint64]bool{}

	conn := mvh.redisPool.Get()
	defer conn.Close()

	buyers := mvh.storer.Buyers()
	for _, buyer := range buyers {
		scanMatch := buyer.CompanyCode + "-*"
		var scanCursor int64

		for {
			values, err := redis.Values(conn.Do("scan", scanCursor, "match", scanMatch, "count", 1000))
			if err != nil {
				return fmt.Errorf("failed to get customer keys from multipath veto redis: %v", err)
			}

			var results []string
			_, err = redis.Scan(values, &scanCursor, &results)
			if err != nil {
				return fmt.Errorf("failed to scan cusor and user hashes: %v", err)
			}

			for _, result := range results {
				userHashString := strings.Split(result, "-")[1]
				userHash, err := strconv.ParseUint(userHashString, 16, 64)
				if err != nil {
					return fmt.Errorf("failed to parse user hash %s: %v", userHashString, err)
				}

				if _, ok := newMultipathVetoMap[buyer.CompanyCode]; !ok {
					newMultipathVetoMap[buyer.CompanyCode] = map[uint64]bool{}
				}

				newMultipathVetoMap[buyer.CompanyCode][userHash] = true
			}

			if scanCursor == 0 {
				break
			}
		}
	}

	mvh.cachedMultipathVetoesMutex.Lock()
	mvh.cachedMultipathVetoes = newMultipathVetoMap
	mvh.cachedMultipathVetoesMutex.Unlock()

	return nil
}
