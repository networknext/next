package admin

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
	ghostarmy "github.com/networknext/backend/ghost_army"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
	"github.com/networknext/backend/transport"
)

func FlushSessions(r *http.Request, topSessionsPool *redis.Pool, sessionMetaPool *redis.Pool, sessionSlicesPool *redis.Pool, sessionMapPool *redis.Pool) error {
	if !VerifyAllRoles(r, OpsRole) {
		return fmt.Errorf("FlushSessions(): %v", ErrInsufficientPrivileges)
	}

	topSessions := topSessionsPool.Get()
	defer topSessions.Close()
	if _, err := topSessions.Do("FLUSHALL", "ASYNC"); err != nil {
		return err
	}

	sessionMeta := sessionMetaPool.Get()
	defer sessionMeta.Close()
	if _, err := sessionMeta.Do("FLUSHALL", "ASYNC"); err != nil {
		return err
	}

	sessionSlices := sessionSlicesPool.Get()
	defer sessionSlices.Close()
	if _, err := sessionSlices.Do("FLUSHALL", "ASYNC"); err != nil {
		return err
	}

	sessionMap := sessionMapPool.Get()
	defer sessionMap.Close()
	if _, err := sessionMap.Do("FLUSHALL", "ASYNC"); err != nil {
		return err
	}

	return nil
}

type TotalSessionsOutput struct {
	Next   int
	Direct int
}

func TotalSessions(r *http.Request, sessionMapPool *redis.Pool, buyer routing.Buyer, buyers []routing.Buyer) (TotalSessionsOutput, error) {
	if r.Body != nil {
		defer r.Body.Close()
	}

	var output TotalSessionsOutput

	redisClient := sessionMapPool.Get()
	defer redisClient.Close()
	minutes := time.Now().Unix() / 60

	ghostArmyBuyerID := ghostarmy.GhostArmyBuyerID(os.Getenv("ENV"))
	var ghostArmyScalar uint64 = 25
	if v, ok := os.LookupEnv("GHOST_ARMY_SCALER"); ok {
		if v, err := strconv.ParseUint(v, 10, 64); err == nil {
			ghostArmyScalar = v
		}
	}

	switch buyer.CompanyCode {
	case "":
		var oldCount int
		var newCount int

		for _, buyer := range buyers {
			stringID := fmt.Sprintf("%016x", buyer.ID)
			redisClient.Send("HLEN", fmt.Sprintf("n-%s-%d", stringID, minutes-1))
			redisClient.Send("HLEN", fmt.Sprintf("n-%s-%d", stringID, minutes))
		}
		redisClient.Flush()

		var ghostArmyNextCount int
		for _, buyer := range buyers {
			firstCount, err := redis.Int(redisClient.Receive())
			if err != nil {
				err = fmt.Errorf("TotalSessions() failed getting total session count next: %v", err)
				return output, err
			}
			oldCount += firstCount

			secondCount, err := redis.Int(redisClient.Receive())
			if err != nil {
				err = fmt.Errorf("TotalSessions() failed getting total session count next: %v", err)
				return output, err
			}
			newCount += secondCount

			if buyer.ID == ghostArmyBuyerID {
				if firstCount > secondCount {
					ghostArmyNextCount = firstCount
				} else {
					ghostArmyNextCount = secondCount
				}
			}
		}

		output.Next = oldCount
		if newCount > oldCount {
			output.Next = newCount
		}

		oldCount = 0
		newCount = 0

		for _, buyer := range buyers {
			stringID := fmt.Sprintf("%016x", buyer.ID)
			redisClient.Send("HLEN", fmt.Sprintf("d-%s-%d", stringID, minutes-1))
			redisClient.Send("HLEN", fmt.Sprintf("d-%s-%d", stringID, minutes))
		}
		redisClient.Flush()

		for _, buyer := range buyers {
			count, err := redis.Int(redisClient.Receive())
			if err != nil {
				err = fmt.Errorf("TotalSessions() failed getting total session count direct: %v", err)
				return output, err
			}
			oldCount += count

			count, err = redis.Int(redisClient.Receive())
			if err != nil {
				err = fmt.Errorf("TotalSessions() failed getting total session count direct: %v", err)
				return output, err
			}

			if buyer.ID == ghostArmyBuyerID {
				// scale by next values because ghost army data contains 0 direct
				// if ghost army is turned off then this number will be 0 and have no effect
				count = ghostArmyNextCount * int(ghostArmyScalar)
			}
			newCount += count
		}

		output.Direct = oldCount
		if newCount > oldCount {
			output.Direct = newCount
		}
	default:
		buyerID := fmt.Sprintf("%016x", buyer.ID)
		if !VerifyAllRoles(r, SameBuyerRole(buyer.CompanyCode)) {
			err := fmt.Errorf("TotalSessions(): %v", ErrInsufficientPrivileges)
			return output, err
		}

		redisClient.Send("HLEN", fmt.Sprintf("d-%s-%d", buyerID, minutes-1))
		redisClient.Send("HLEN", fmt.Sprintf("d-%s-%d", buyerID, minutes))
		redisClient.Send("HLEN", fmt.Sprintf("n-%s-%d", buyerID, minutes-1))
		redisClient.Send("HLEN", fmt.Sprintf("n-%s-%d", buyerID, minutes))
		redisClient.Flush()

		oldDirectCount, err := redis.Int(redisClient.Receive())
		if err != nil {
			err = fmt.Errorf("TotalSessions() failed getting buyer session direct counts: %v", err)
			return output, err
		}

		newDirectCount, err := redis.Int(redisClient.Receive())
		if err != nil {
			err = fmt.Errorf("TotalSessions() failed getting buyer session direct counts: %v", err)
			return output, err
		}

		output.Direct = oldDirectCount
		if newDirectCount > oldDirectCount {
			output.Direct = newDirectCount
		}

		oldNextCount, err := redis.Int(redisClient.Receive())
		if err != nil {
			err = fmt.Errorf("TotalSessions() failed getting buyer session next counts: %v", err)
			return output, err
		}

		newNextCount, err := redis.Int(redisClient.Receive())
		if err != nil {
			err = fmt.Errorf("TotalSessions() failed getting buyer session next counts: %v", err)
			return output, err
		}

		output.Next = oldNextCount
		if newNextCount > oldNextCount {
			output.Next = newNextCount
		}
	}

	return output, nil
}

// TopSessions generates the top sessions sorted by improved RTT
func TopSessions(r *http.Request, topSessionsPool *redis.Pool, sessionMetaPool *redis.Pool, buyer routing.Buyer) ([]transport.SessionMeta, error) {
	var err error
	var topSessionsA []string
	var topSessionsB []string

	sessions := make([]transport.SessionMeta, 0)

	minutes := time.Now().Unix() / 60

	topSessionsClient := topSessionsPool.Get()
	defer topSessionsClient.Close()

	// get the top session IDs globally or for a buyer from the sorted set
	switch buyer.CompanyCode {
	case "":
		// Get top sessions from the past 2 minutes sorted by greatest to least improved RTT
		topSessionsA, err = redis.Strings(topSessionsClient.Do("ZREVRANGE", fmt.Sprintf("s-%d", minutes-1), "0", fmt.Sprintf("%d", TopSessionsSize)))
		if err != nil && err != redis.ErrNil {
			err = fmt.Errorf("TopSessions() failed getting top sessions A: %v", err)
			return sessions, err
		}
		topSessionsB, err = redis.Strings(topSessionsClient.Do("ZREVRANGE", fmt.Sprintf("s-%d", minutes), "0", fmt.Sprintf("%d", TopSessionsSize)))
		if err != nil && err != redis.ErrNil {
			err = fmt.Errorf("TopSessions() failed getting top sessions B: %v", err)
			return sessions, err
		}
	default:
		if err != nil {
			err = fmt.Errorf("TopSessions() failed getting company with code: %v", err)
			return sessions, err
		}
		buyerID := fmt.Sprintf("%x", buyer.ID)
		if !VerifyAllRoles(r, SameBuyerRole(buyer.CompanyCode)) {
			err := fmt.Errorf("TopSessions(): %v", ErrInsufficientPrivileges)
			return sessions, err
		}

		topSessionsA, err = redis.Strings(topSessionsClient.Do("ZREVRANGE", fmt.Sprintf("sc-%s-%d", buyerID, minutes-1), "0", fmt.Sprintf("%d", TopSessionsSize)))
		if err != nil && err != redis.ErrNil {
			err = fmt.Errorf("TopSessions() failed getting top sessions A for buyer ID %016x: %v", buyerID, err)
			return sessions, err
		}
		topSessionsB, err = redis.Strings(topSessionsClient.Do("ZREVRANGE", fmt.Sprintf("sc-%s-%d", buyerID, minutes), "0", fmt.Sprintf("%d", TopSessionsSize)))
		if err != nil && err != redis.ErrNil {
			err = fmt.Errorf("TopSessions() failed getting top sessions B for buyer ID %016x: %v", buyerID, err)
			return sessions, err
		}
	}

	sessionMetaClient := sessionMetaPool.Get()
	defer sessionMetaClient.Close()

	sessionIDsRetreivedMap := make(map[string]bool)
	for _, sessionID := range topSessionsA {
		sessionMetaClient.Send("GET", fmt.Sprintf("sm-%s", sessionID))
		sessionIDsRetreivedMap[sessionID] = true
	}
	for _, sessionID := range topSessionsB {
		if _, ok := sessionIDsRetreivedMap[sessionID]; !ok {
			sessionMetaClient.Send("GET", fmt.Sprintf("sm-%s", sessionID))
			sessionIDsRetreivedMap[sessionID] = true
		}
	}
	sessionMetaClient.Flush()

	var sessionMetasNext []transport.SessionMeta
	var sessionMetasDirect []transport.SessionMeta
	var meta transport.SessionMeta
	for i := 0; i < len(sessionIDsRetreivedMap); i++ {
		metaString, err := redis.String(sessionMetaClient.Receive())
		if err != nil && err != redis.ErrNil {
			err = fmt.Errorf("TopSessions() failed getting top sessions meta: %v", err)
			return sessions, err
		}

		splitMetaStrings := strings.Split(metaString, "|")
		if err := meta.ParseRedisString(splitMetaStrings); err != nil {
			err = fmt.Errorf("TopSessions() failed to parse redis string into meta: %v", err)
			continue
		}

		if !VerifyAllRoles(r, SameBuyerRole(buyer.CompanyCode)) {
			meta.Anonymise()
		}

		// Split the sessions metas into two slices so we can sort them separately.
		// This is necessary because if we were to force sessions next, then sorting
		// by improvement won't always put next sessions on top.
		if meta.OnNetworkNext {
			sessionMetasNext = append(sessionMetasNext, meta)
		} else {
			sessionMetasDirect = append(sessionMetasDirect, meta)
		}
	}

	// These sorts are necessary because we are combining two ZREVRANGEs from two separate minute buckets.
	sort.Slice(sessionMetasNext, func(i, j int) bool {
		return sessionMetasNext[i].DeltaRTT > sessionMetasNext[j].DeltaRTT
	})

	sort.Slice(sessionMetasDirect, func(i, j int) bool {
		return sessionMetasDirect[i].DirectRTT > sessionMetasDirect[j].DirectRTT
	})

	sessionMetas := append(sessionMetasNext, sessionMetasDirect...)

	if len(sessionMetas) > TopSessionsSize {
		sessions = sessionMetas[:TopSessionsSize]
		return sessions, nil
	}

	return sessions, nil
}

type SessionDetailsOutput struct {
	Meta   transport.SessionMeta
	Slices []transport.SessionSlice
}

func SessionDetails(r *http.Request, storage storage.Storer, sessionMetaPool *redis.Pool, sessionSlicesPool *redis.Pool, sessionID string) (SessionDetailsOutput, error) {
	var err error
	var output SessionDetailsOutput

	sessionMetaClient := sessionMetaPool.Get()
	defer sessionMetaClient.Close()

	metaString, err := redis.String(sessionMetaClient.Do("GET", fmt.Sprintf("sm-%s", sessionID)))
	if err != nil && err != redis.ErrNil {
		err = fmt.Errorf("SessionDetails() failed getting session meta: %v", err)
		return output, err
	}

	metaStringsSplit := strings.Split(metaString, "|")
	if err := output.Meta.ParseRedisString(metaStringsSplit); err != nil {
		err = fmt.Errorf("SessionDetails() SessionMeta unmarshaling error: %v", err)
		return output, err
	}

	buyer, err := storage.Buyer(output.Meta.BuyerID)
	if err != nil {
		err = fmt.Errorf("SessionDetails() failed to fetch buyer: %v", err)
		return output, err
	}

	if !VerifyAllRoles(r, SameBuyerRole(buyer.CompanyCode)) {
		output.Meta.Anonymise()
	}

	output.Slices = make([]transport.SessionSlice, 0)

	sessionSlicesClient := sessionSlicesPool.Get()
	defer sessionSlicesClient.Close()

	slices, err := redis.Strings(sessionSlicesClient.Do("LRANGE", fmt.Sprintf("ss-%s", sessionID), "0", "-1"))
	if err != nil && err != redis.ErrNil {
		err = fmt.Errorf("SessionDetails() failed getting session slices: %v", err)
		return output, err
	}

	for i := 0; i < len(slices); i++ {
		sliceStrings := strings.Split(slices[i], "|")
		var sessionSlice transport.SessionSlice
		if err := sessionSlice.ParseRedisString(sliceStrings); err != nil {
			err = fmt.Errorf("SessionDetails() SessionSlice parsing error: %v", err)
			return output, err
		}

		output.Slices = append(output.Slices, sessionSlice)
	}

	sort.Slice(output.Slices, func(i, j int) bool {
		return output.Slices[i].Timestamp.Before(output.Slices[j].Timestamp)
	})

	return output, nil
}

func UpdateGameConfiguration(r *http.Request, storage storage.Storer, newPublicKey string) (string, error) {
	var err error
	var buyerID uint64

	if !VerifyAnyRole(r, AdminRole, OwnerRole) {
		err = fmt.Errorf("UpdateGameConfiguration(): %v", ErrInsufficientPrivileges)
		return "", err
	}

	companyCode, err := RequestCompany(r)
	if err != nil {
		err := fmt.Errorf("UpdateGameConfiguration(): user is not assigned to a company")
		return "", err
	}
	if companyCode == "" {
		err = fmt.Errorf("UpdateGameConfiguration(): failed to parse company code")
		return "", err
	}

	ctx := context.Background()
	buyer, err := storage.BuyerWithCompanyCode(companyCode)

	byteKey, err := base64.StdEncoding.DecodeString(newPublicKey)
	if err != nil {
		err = fmt.Errorf("UpdateGameConfiguration() could not decode public key string")
		return "", err
	}

	buyerID = binary.LittleEndian.Uint64(byteKey[0:8])

	// Buyer not found
	if buyer.ID == 0 {

		// Create new buyer
		err = storage.AddBuyer(ctx, routing.Buyer{
			CompanyCode: companyCode,
			ID:          buyerID,
			Live:        false,
			PublicKey:   byteKey[8:],
		})

		if err != nil {
			err = fmt.Errorf("UpdateGameConfiguration() failed to add buyer")
			return "", err
		}

		// Check if buyer is associated with the ID and everything worked
		if buyer, err = storage.Buyer(buyerID); err != nil {
			err = fmt.Errorf("UpdateGameConfiguration() buyer creation failed: %v", err)
			return "", err
		}

		// Setup reply
		return buyer.EncodedPublicKey(), nil
	}

	live := buyer.Live

	if err = storage.RemoveBuyer(ctx, buyer.ID); err != nil {
		err = fmt.Errorf("UpdateGameConfiguration() failed to remove buyer")
		return "", err
	}

	err = storage.AddBuyer(ctx, routing.Buyer{
		CompanyCode: companyCode,
		ID:          buyerID,
		Live:        live,
		PublicKey:   byteKey[8:],
	})

	if err != nil {
		err = fmt.Errorf("UpdateGameConfiguration() buyer update failed: %v", err)
		return "", err
	}

	// Check if buyer is associated with the ID and everything worked
	if buyer, err = storage.Buyer(buyerID); err != nil {
		err = fmt.Errorf("UpdateGameConfiguration() buyer update check failed: %v", err)
		return "", err
	}

	return buyer.EncodedPublicKey(), nil
}

type MapPointCaches struct {
	MapPointsBuyerCache        map[string]json.RawMessage
	MapPointsCompactBuyerCache map[string]json.RawMessage
	MapPointsCache             json.RawMessage
	MapPointsCompactCache      json.RawMessage
}

func GenerateMapPointsPerBuyer(sessionMapPool *redis.Pool, buyers []routing.Buyer) (MapPointCaches, error) {
	var err error
	var caches MapPointCaches

	// slices to hold all the final map points
	mapPointsBuyers := make(map[string][]transport.SessionMapPoint, 0)
	mapPointsBuyersCompact := make(map[string][][]interface{}, 0)
	mapPointsGlobal := make([]transport.SessionMapPoint, 0)
	mapPointsGlobalCompact := make([][]interface{}, 0)

	caches.MapPointsBuyerCache = make(map[string]json.RawMessage, 0)
	caches.MapPointsCompactBuyerCache = make(map[string]json.RawMessage, 0)

	for _, buyer := range buyers {
		directPointStrings, nextPointStrings, err := getDirectAndNextMapPointStrings(sessionMapPool, &buyer)
		if err != nil && err != redis.ErrNil {
			err = fmt.Errorf("SessionMapPoints() failed getting map points for buyer %s: %v", buyer.CompanyCode, err)
			return MapPointCaches{}, err
		}

		var point transport.SessionMapPoint
		for _, directPointString := range directPointStrings {
			directSplitStrings := strings.Split(directPointString, "|")
			if err := point.ParseRedisString(directSplitStrings); err != nil {
				err = fmt.Errorf("SessionMapPoints() failed to parse direct map point for buyer %s: %v", buyer.CompanyCode, err)
				return MapPointCaches{}, err
			}

			if point.Latitude != 0 && point.Longitude != 0 {
				mapPointsBuyers[buyer.CompanyCode] = append(mapPointsBuyers[buyer.CompanyCode], point)
				mapPointsGlobal = append(mapPointsGlobal, point)

				mapPointsBuyersCompact[buyer.CompanyCode] = append(mapPointsBuyersCompact[buyer.CompanyCode], []interface{}{point.Longitude, point.Latitude, false})
				mapPointsGlobalCompact = append(mapPointsGlobalCompact, []interface{}{point.Longitude, point.Latitude, false})
			}
		}

		for _, nextPointString := range nextPointStrings {
			nextSplitStrings := strings.Split(nextPointString, "|")
			if err := point.ParseRedisString(nextSplitStrings); err != nil {
				err = fmt.Errorf("SessionMapPoints() failed to next parse map point for buyer %s: %v", buyer.CompanyCode, err)
				return MapPointCaches{}, err
			}

			if point.Latitude != 0 && point.Longitude != 0 {
				mapPointsBuyers[buyer.CompanyCode] = append(mapPointsBuyers[buyer.CompanyCode], point)
				mapPointsGlobal = append(mapPointsGlobal, point)

				mapPointsBuyersCompact[buyer.CompanyCode] = append(mapPointsBuyersCompact[buyer.CompanyCode], []interface{}{point.Longitude, point.Latitude, true})
				mapPointsGlobalCompact = append(mapPointsGlobalCompact, []interface{}{point.Longitude, point.Latitude, true})
			}
		}

		caches.MapPointsBuyerCache[buyer.CompanyCode], err = json.Marshal(mapPointsBuyers[buyer.CompanyCode])
		if err != nil {
			return MapPointCaches{}, err
		}

		caches.MapPointsCompactBuyerCache[buyer.CompanyCode], err = json.Marshal(mapPointsBuyersCompact[buyer.CompanyCode])
		if err != nil {
			return MapPointCaches{}, err
		}
	}

	// marshal the map points slice to local cache
	caches.MapPointsCache, err = json.Marshal(mapPointsGlobal)
	if err != nil {
		return MapPointCaches{}, err
	}

	caches.MapPointsCompactCache, err = json.Marshal(mapPointsGlobalCompact)
	if err != nil {
		return MapPointCaches{}, err
	}
	return caches, nil
}

func getDirectAndNextMapPointStrings(sessionMapPool *redis.Pool, buyer *routing.Buyer) ([]string, []string, error) {
	minutes := time.Now().Unix() / 60

	redisClient := sessionMapPool.Get()
	defer redisClient.Close()

	stringID := fmt.Sprintf("%016x", buyer.ID)
	redisClient.Send("HGETALL", fmt.Sprintf("d-%s-%d", stringID, minutes-1))
	redisClient.Send("HGETALL", fmt.Sprintf("d-%s-%d", stringID, minutes))
	redisClient.Send("HGETALL", fmt.Sprintf("n-%s-%d", stringID, minutes-1))
	redisClient.Send("HGETALL", fmt.Sprintf("n-%s-%d", stringID, minutes))
	redisClient.Flush()

	directMap, err := redis.StringMap(redisClient.Receive())
	if err != nil {
		return nil, nil, err
	}

	directMapB, err := redis.StringMap(redisClient.Receive())
	if err != nil {
		return nil, nil, err
	}

	for k, v := range directMapB {
		directMap[k] = v
	}

	nextMap, err := redis.StringMap(redisClient.Receive())
	if err != nil {
		return nil, nil, err
	}

	nextMapB, err := redis.StringMap(redisClient.Receive())
	if err != nil {
		return nil, nil, err
	}

	for k, v := range nextMapB {
		nextMap[k] = v
	}

	direct := make([]string, len(directMap))
	var index int
	for _, v := range directMap {
		direct[index] = v
		index++
	}

	next := make([]string, len(nextMap))
	index = 0
	for _, v := range nextMap {
		next[index] = v
		index++
	}

	return direct, next, nil
}
