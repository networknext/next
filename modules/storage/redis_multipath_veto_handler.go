package storage

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/networknext/backend/modules/routing"

	"github.com/gomodule/redigo/redis"
)

type RedisMultipathVetoHandler struct {
	getDatabase                func() *routing.DatabaseBinWrapper
	redisPool                  *redis.Pool
	cachedMultipathVetoes      map[string]map[uint64]bool
	cachedMultipathVetoesMutex sync.RWMutex
}

func NewRedisMultipathVetoHandler(redisHost string, redisPassword string, redisMaxIdleConns int, redisMaxActiveConns int, getDatabase func() *routing.DatabaseBinWrapper) (*RedisMultipathVetoHandler, error) {
	redisPool := NewRedisPool(redisHost, redisPassword, redisMaxIdleConns, redisMaxActiveConns)
	conn := redisPool.Get()
	defer conn.Close()

	if _, err := conn.Do("PING"); err != nil {
		return nil, fmt.Errorf("could not ping multipath veto redis instance: %v", err)
	}

	return &RedisMultipathVetoHandler{
		getDatabase:           getDatabase,
		redisPool:             redisPool,
		cachedMultipathVetoes: make(map[string]map[uint64]bool),
	}, nil
}

func (rmvh *RedisMultipathVetoHandler) GetMapCopy(companyCode string) map[uint64]bool {
	rmvh.cachedMultipathVetoesMutex.RLock()
	defer rmvh.cachedMultipathVetoesMutex.RUnlock()

	multipathVetoMap, ok := rmvh.cachedMultipathVetoes[companyCode]
	if !ok {
		return make(map[uint64]bool)
	}

	multipathVetoMapCopy := make(map[uint64]bool)
	for k, v := range multipathVetoMap {
		multipathVetoMapCopy[k] = v
	}

	return multipathVetoMapCopy
}

func (rmvh *RedisMultipathVetoHandler) MultipathVetoUser(companyCode string, userHash uint64) error {
	conn := rmvh.redisPool.Get()
	defer conn.Close()

	result, err := redis.String(conn.Do("SET", companyCode+"-"+fmt.Sprintf("%016x", userHash), "1", "EX", "604800")) // 7 days
	if err != nil {
		return fmt.Errorf("failed setting multipath veto on user %016x for buyer %s: %v", userHash, companyCode, err)
	}

	if result != "OK" {
		return fmt.Errorf("failed setting multipath veto on user %016x for buyer %s: set command returned false", userHash, companyCode)
	}

	rmvh.cachedMultipathVetoesMutex.Lock()
	if _, ok := rmvh.cachedMultipathVetoes[companyCode]; !ok {
		rmvh.cachedMultipathVetoes[companyCode] = map[uint64]bool{}
	}
	rmvh.cachedMultipathVetoes[companyCode][userHash] = true
	rmvh.cachedMultipathVetoesMutex.Unlock()
	return nil
}

func (rmvh *RedisMultipathVetoHandler) Sync() error {
	newMultipathVetoMap := map[string]map[uint64]bool{}

	conn := rmvh.redisPool.Get()
	defer conn.Close()

	binWrapper := rmvh.getDatabase()

	for _, buyer := range binWrapper.BuyerMap {
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

	rmvh.cachedMultipathVetoesMutex.Lock()
	rmvh.cachedMultipathVetoes = newMultipathVetoMap
	rmvh.cachedMultipathVetoesMutex.Unlock()

	return nil
}
