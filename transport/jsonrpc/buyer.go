package jsonrpc

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-kit/kit/log"
	"github.com/gomodule/redigo/redis"
	"github.com/networknext/backend/encoding"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
	"github.com/networknext/backend/transport"
)

const (
	TopSessionsSize          = 1000
	MapPointByteCacheVersion = uint8(1)
)

type BuyersService struct {
	mu                      sync.Mutex
	mapPointsByteCache      []byte
	mapPointsBuyerByteCache map[string][]byte

	mapPointsCache        json.RawMessage
	mapPointsCompactCache json.RawMessage

	mapPointsBuyerCache        map[string]json.RawMessage
	mapPointsCompactBuyerCache map[string]json.RawMessage

	RedisPoolTopSessions   *redis.Pool
	RedisPoolSessionMeta   *redis.Pool
	RedisPoolSessionSlices *redis.Pool
	RedisPoolSessionMap    *redis.Pool
	Storage                storage.Storer
	Logger                 log.Logger
}

type FlushSessionsArgs struct{}

type FlushSessionsReply struct{}

func (s *BuyersService) FlushSessions(r *http.Request, args *FlushSessionsArgs, reply *FlushSessionsReply) error {
	if !VerifyAllRoles(r, OpsRole) {
		return fmt.Errorf("FlushSessions(): %v", ErrInsufficientPrivileges)
	}

	topSessions := s.RedisPoolTopSessions.Get()
	defer topSessions.Close()

	topSessions.Send("FLUSHALL", "ASYNC")
	topSessions.Flush()

	if _, err := topSessions.Receive(); err != nil {
		return err
	}

	sessionMeta := s.RedisPoolSessionMeta.Get()
	defer sessionMeta.Close()

	sessionMeta.Send("FLUSHALL", "ASYNC")
	sessionMeta.Flush()

	if _, err := sessionMeta.Receive(); err != nil {
		return err
	}

	sessionSlices := s.RedisPoolSessionSlices.Get()
	defer sessionSlices.Close()

	sessionSlices.Send("FLUSHALL", "ASYNC")
	sessionSlices.Flush()

	if _, err := sessionSlices.Receive(); err != nil {
		return err
	}

	sessionMap := s.RedisPoolSessionMap.Get()
	defer sessionMap.Close()

	sessionMap.Send("FLUSHALL", "ASYNC")
	sessionMap.Flush()

	if _, err := sessionMap.Receive(); err != nil {
		return err
	}

	return nil
}

type UserSessionsArgs struct {
	UserHash string `json:"user_hash"`
}

type UserSessionsReply struct {
	Sessions []transport.SessionMeta `json:"sessions"`
}

func (s *BuyersService) UserSessions(r *http.Request, args *UserSessionsArgs, reply *UserSessionsReply) error {
	// var sessionIDs []string

	// userhash := args.UserHash
	// reply.Sessions = make([]transport.SessionMeta, 0)

	// err := s.Sli.SMembers(fmt.Sprintf("user-%s-sessions", userhash)).ScanSlice(&sessionIDs)
	// if err != nil {
	// 	err = fmt.Errorf("UserSessions() failed getting user sessions: %v", err)
	// 	s.Logger.Log("err", err)
	// 	return err
	// }

	// if len(sessionIDs) == 0 {
	// 	hash := fnv.New64a()
	// 	_, err := hash.Write([]byte(userhash))
	// 	if err != nil {
	// 		err = fmt.Errorf("UserSessions() error writing 64a hash: %v", err)
	// 		s.Logger.Log("err", err)
	// 		return err
	// 	}
	// 	hashedID := fmt.Sprintf("%016x", hash.Sum64())

	// 	err = s.RedisClient.SMembers(fmt.Sprintf("user-%s-sessions", hashedID)).ScanSlice(&sessionIDs)
	// 	if err != nil {
	// 		err = fmt.Errorf("UserSessions() failed getting user sessions: %v", err)
	// 		s.Logger.Log("err", err)
	// 		return err
	// 	}
	// }

	// if len(sessionIDs) == 0 {
	// 	return nil
	// }

	// var getCmds []*redis.StringCmd
	// {
	// 	gettx := s.RedisClient.TxPipeline()
	// 	for _, sessionID := range sessionIDs {
	// 		getCmds = append(getCmds, gettx.Get(fmt.Sprintf("session-%s-meta", sessionID)))
	// 	}
	// 	_, err = gettx.Exec()
	// 	if err != nil && err != redis.Nil {
	// 		err = fmt.Errorf("UserSessions() redis.Pipeliner error: %v", err)
	// 		s.Logger.Log("err", err)
	// 		return err
	// 	}
	// }

	// sremtx := s.RedisClient.TxPipeline()
	// {
	// 	var meta transport.SessionMeta
	// 	for _, cmd := range getCmds {
	// 		err = cmd.Scan(&meta)
	// 		if err != nil {
	// 			args := cmd.Args()
	// 			key := args[1].(string)
	// 			keyparts := strings.Split(key, "-")

	// 			sremtx.SRem(fmt.Sprintf("user-%s-sessions", userhash), keyparts[1])
	// 			continue
	// 		}

	// 		if VerifyAnyRole(r, AnonymousRole, UnverifiedRole) || !VerifyAllRoles(r, s.SameBuyerRole(fmt.Sprintf("%016x", meta.BuyerID))) {
	// 			meta.Anonymise()
	// 		}

	// 		reply.Sessions = append(reply.Sessions, meta)
	// 	}
	// }

	// sremcmds, err := sremtx.Exec()
	// if err != nil && err != redis.Nil {
	// 	err = fmt.Errorf("UserSessions() redis.Pipeliner error: %v", err)
	// 	s.Logger.Log("err", err)
	// 	return err
	// }

	// level.Info(s.Logger).Log("key", "user-*-sessions", "removed", len(sremcmds))

	// sort.Slice(reply.Sessions, func(i int, j int) bool {
	// 	return reply.Sessions[i].ID < reply.Sessions[j].ID
	// })

	// return nil

	reply.Sessions = make([]transport.SessionMeta, 0)
	return nil
}

type TotalSessionsArgs struct {
	BuyerID string `json:"buyer_id"`
}

type TotalSessionsReply struct {
	Direct int `json:"direct"`
	Next   int `json:"next"`
}

func (s *BuyersService) TotalSessions(r *http.Request, args *TotalSessionsArgs, reply *TotalSessionsReply) error {
	// if r.Body != nil {
	// 	defer r.Body.Close()
	// }

	// // get the top session IDs globally or for a buyer from the sorted set
	// switch args.BuyerID {
	// case "":
	// 	// Get top Direct sessions sorted by least to greatest direct RTT
	// 	directTotals, err := s.RedisClient.HGetAll("session-count-total-direct").Result()
	// 	if err != nil {
	// 		err = fmt.Errorf("TotalSessions() failed getting session-count-total-direct: %v", err)
	// 		s.Logger.Log("err", err)
	// 		return err
	// 	}

	// 	// Get top Next sessions
	// 	nextTotals, err := s.RedisClient.HGetAll("session-count-total-next").Result()
	// 	if err != nil {
	// 		err = fmt.Errorf("TotalSessions() failed getting session-count-total-next: %v", err)
	// 		s.Logger.Log("err", err)
	// 		return err
	// 	}

	// 	var totalDirect int
	// 	for _, directString := range directTotals {
	// 		direct, err := strconv.Atoi(directString)
	// 		if err != nil {
	// 			err = fmt.Errorf("TotalSessions() failed to parse direct session count (%s): %v", directString, err)
	// 			s.Logger.Log("err", err)
	// 			return err
	// 		}

	// 		totalDirect += direct
	// 	}

	// 	var totalNext int
	// 	for _, nextString := range nextTotals {
	// 		next, err := strconv.Atoi(nextString)
	// 		if err != nil {
	// 			err = fmt.Errorf("TotalSessions() failed to parse next session count (%s): %v", nextString, err)
	// 			s.Logger.Log("err", err)
	// 			return err
	// 		}
	// 		totalNext += next
	// 	}

	// 	reply.Direct = totalDirect
	// 	reply.Next = totalNext
	// default:
	// 	if !VerifyAllRoles(r, s.SameBuyerRole(args.BuyerID)) {
	// 		err := fmt.Errorf("TotalSessions(): %v", ErrInsufficientPrivileges)
	// 		s.Logger.Log("err", err)
	// 		return err
	// 	}

	// 	buyerDirectTotals, err := s.RedisClient.HGetAll(fmt.Sprintf("session-count-direct-buyer-%s", args.BuyerID)).Result()
	// 	if err != nil {
	// 		err = fmt.Errorf("TotalSessions() failed getting session-count-direct-buyer-%s: %v", args.BuyerID, err)
	// 		s.Logger.Log("err", err)
	// 		return err
	// 	}

	// 	buyerNextTotals, err := s.RedisClient.HGetAll(fmt.Sprintf("session-count-next-buyer-%s", args.BuyerID)).Result()
	// 	if err != nil {
	// 		err = fmt.Errorf("TotalSessions() failed getting session-count-next-buyer-%s: %v", args.BuyerID, err)
	// 		s.Logger.Log("err", err)
	// 		return err
	// 	}

	// 	var buyerDirectTotal int
	// 	for _, directString := range buyerDirectTotals {
	// 		direct, err := strconv.Atoi(directString)
	// 		if err != nil {
	// 			err = fmt.Errorf("TotalSessions() failed to parse buyer direct session count (%s): %v", directString, err)
	// 			s.Logger.Log("err", err)
	// 			return err
	// 		}

	// 		buyerDirectTotal += direct
	// 	}

	// 	var buyerNextTotal int
	// 	for _, nextString := range buyerNextTotals {
	// 		next, err := strconv.Atoi(nextString)
	// 		if err != nil {
	// 			err = fmt.Errorf("TotalSessions() failed to parse buyer next session count (%s): %v", nextString, err)
	// 			s.Logger.Log("err", err)
	// 			return err
	// 		}

	// 		buyerNextTotal += next
	// 	}

	// 	reply.Direct = buyerDirectTotal
	// 	reply.Next = buyerNextTotal
	// }

	// return nil

	reply.Direct = 0
	reply.Next = 0
	return nil
}

type TopSessionsArgs struct {
	BuyerID string `json:"buyer_id"`
}

type TopSessionsReply struct {
	Sessions []transport.SessionMeta `json:"sessions"`
}

// TopSessions generates the top sessions sorted by improved RTT
func (s *BuyersService) TopSessions(r *http.Request, args *TopSessionsArgs, reply *TopSessionsReply) error {
	var err error
	var topSessionsA []string
	var topSessionsB []string

	reply.Sessions = make([]transport.SessionMeta, 0)

	minutes := time.Now().Unix() / 60

	topSessionsClient := s.RedisPoolTopSessions.Get()
	defer topSessionsClient.Close()

	// get the top session IDs globally or for a buyer from the sorted set
	switch args.BuyerID {
	case "":
		// Get top sessions from the past 2 minutes sorted by greatest to least improved RTT
		topSessionsA, err = redis.Strings(topSessionsClient.Do("ZREVRANGE", fmt.Sprintf("s-%d", minutes-1), "0", fmt.Sprintf("%d", TopSessionsSize)))
		if err != nil && err != redis.ErrNil {
			err = fmt.Errorf("TopSessions() failed getting top sessions A: %v", err)
			s.Logger.Log("err", err)
			return err
		}
		topSessionsB, err = redis.Strings(topSessionsClient.Do("ZREVRANGE", fmt.Sprintf("s-%d", minutes), "0", fmt.Sprintf("%d", TopSessionsSize)))
		if err != nil && err != redis.ErrNil {
			err = fmt.Errorf("TopSessions() failed getting top sessions B: %v", err)
			s.Logger.Log("err", err)
			return err
		}
	default:
		if !VerifyAllRoles(r, s.SameBuyerRole(args.BuyerID)) {
			err = fmt.Errorf("TopSessions(): %v", ErrInsufficientPrivileges)
			s.Logger.Log("err", err)
			return err
		}

		topSessionsA, err = redis.Strings(topSessionsClient.Do("ZREVRANGE", fmt.Sprintf("sc-%s-%d", args.BuyerID, minutes), "0", fmt.Sprintf("%d", TopSessionsSize)))
		if err != nil && err != redis.ErrNil {
			err = fmt.Errorf("TopSessions() failed getting top sessions A for buyer ID %016x: %v", args.BuyerID, err)
			s.Logger.Log("err", err)
			return err
		}
		topSessionsB, err = redis.Strings(topSessionsClient.Do("ZREVRANGE", fmt.Sprintf("sc-%s-%d", args.BuyerID, minutes), "0", fmt.Sprintf("%d", TopSessionsSize)))
		if err != nil && err != redis.ErrNil {
			err = fmt.Errorf("TopSessions() failed getting top sessions B for buyer ID %016x: %v", args.BuyerID, err)
			s.Logger.Log("err", err)
			return err
		}
	}

	// build a single transaction to get the sessions details for the session IDs
	sessionMetaClient := s.RedisPoolSessionMeta.Get()
	defer sessionMetaClient.Close()

	sessionMetaClient.Send("MULTI")
	sessionIDsRetreived := make(map[string]bool)
	for _, sessionID := range topSessionsA {
		sessionMetaClient.Send("GET", fmt.Sprintf("session-%s-meta", sessionID))
		sessionIDsRetreived[sessionID] = true
	}
	for _, sessionID := range topSessionsB {
		if _, ok := sessionIDsRetreived[sessionID]; !ok {
			sessionMetaClient.Send("GET", fmt.Sprintf("session-%s-meta", sessionID))
			sessionIDsRetreived[sessionID] = true
		}
	}
	sessionMetaClient.Send("EXEC")
	sessionMetaClient.Flush()

	metas, err := redis.Strings(sessionMetaClient.Receive())
	if err != nil && err != redis.ErrNil {
		err = fmt.Errorf("TopSessions() failed getting top sessions meta: %v", err)
		s.Logger.Log("err", err)
		return err
	}

	var sessionMetas []transport.SessionMeta
	{
		var meta transport.SessionMeta
		for _, cmd := range metas {
			// scan the data from Redis into its SessionMeta struct
			err = cmd.Scan(&meta)

			if !VerifyAllRoles(r, s.SameBuyerRole(args.BuyerID)) {
				meta.Anonymise()
			}

			sessionMetas = append(sessionMetas, meta)
		}
	}

	sort.Slice(sessionMetas, func(i int, j int) bool {
		return sessionMetas[i].DeltaRTT > sessionMetas[j].DeltaRTT
	})

	reply.Sessions = sessionMetas[:TopSessionsSize]
	return nil
}

type SessionDetailsArgs struct {
	SessionID string `json:"session_id"`
}

type SessionDetailsReply struct {
	Meta   transport.SessionMeta    `json:"meta"`
	Slices []transport.SessionSlice `json:"slices"`
}

func (s *BuyersService) SessionDetails(r *http.Request, args *SessionDetailsArgs, reply *SessionDetailsReply) error {
	var err error

	sessionMetaClient := s.RedisPoolSessionMeta.Get()
	defer sessionMetaClient.Close()

	data, err := redis.Bytes(sessionMetaClient.Do("GET", fmt.Sprintf("session-%s-meta", args.SessionID)))
	if err != nil {
		err = fmt.Errorf("SessionDetails() failed getting session meta: %v", err)
		s.Logger.Log("err", err)
		return err
	}

	if err := reply.Meta.UnmarshalBinary(data); err != nil {
		err = fmt.Errorf("SessionDetails() SessionMeta unmarshaling error: %v", err)
		s.Logger.Log("err", err)
		return err
	}

	if !VerifyAllRoles(r, s.SameBuyerRole(fmt.Sprintf("%016x", reply.Meta.BuyerID))) {
		reply.Meta.Anonymise()
	}

	reply.Slices = make([]transport.SessionSlice, 0)

	sessionSlicesClient := s.RedisPoolSessionMeta.Get()
	defer sessionSlicesClient.Close()

	sessionSlicesClient.Send("LRANGE", fmt.Sprintf("ss-%s", args.SessionID), "0", "-1")
	sessionSlicesClient.Flush()

	slices, err := redis.Strings(sessionSlicesClient.Receive())
	if err != nil {
		err = fmt.Errorf("SessionDetails() failed getting session slices: %v", err)
		s.Logger.Log("err", err)
		return err
	}

	for i := 0; i < len(slices); i++ {
		var sessionSlice transport.SessionSlice
		if err := sessionSlice.UnmarshalBinary(slices[i]); err != nil {
			err = fmt.Errorf("SessionDetails() SessionSlice unmarshaling error: %v", err)
			s.Logger.Log("err", err)
			return err
		}

		reply.Slices = append(reply.Slices, sessionSlice)
	}

	// Shouldn't be necessary
	sort.Slice(reply.Slices, func(i int, j int) bool {
		return reply.Slices[i].Timestamp.Before(reply.Slices[j].Timestamp)
	})

	return nil
}

type MapPointsArgs struct {
	BuyerID string `json:"buyer_id"`
}

type MapPointsReply struct {
	Points json.RawMessage `json:"map_points"`
}

type MapPointsByteReply struct {
	Points mapPointsByte `json:"map_points"`
}

type mapPointsByte struct {
	Version     uint8   `json:"version"`
	GreenPoints []point `json:"green_points"`
	BluePoints  []point `json:"blue_points"`
}

type point struct {
	Latitude      float32 `json:"latitude"`
	Longitude     float32 `json:"longitude"`
	OnNetworkNext bool    `json:"on_network_next"`
}

func (s *BuyersService) GenerateMapPointsPerBuyer() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var err error

	buyers := s.Storage.Buyers()

	// slices to hold all the final map points
	mapPointsBuyers := make(map[string][]transport.SessionMapPoint, 0)
	mapPointsBuyersCompact := make(map[string][][]interface{}, 0)
	mapPointsGlobal := make([]transport.SessionMapPoint, 0)
	mapPointsGlobalCompact := make([][]interface{}, 0)

	s.mapPointsBuyerCache = make(map[string]json.RawMessage, 0)
	s.mapPointsCompactBuyerCache = make(map[string]json.RawMessage, 0)

	for _, buyer := range buyers {
		stringID := fmt.Sprintf("%016x", buyer.ID)

		directPointStrings, nextPointStrings, err := s.getDirectAndNextMapPointStrings(&buyer)
		if err != nil {
			err = fmt.Errorf("SessionMapPoints() failed getting map points for buyer %s: %v", stringID, err)
			s.Logger.Log("err", err)
			return err
		}

		var point transport.SessionMapPoint
		for _, directPointString := range directPointStrings {
			// scan the data from Redis into its SessionMapPoint struct
			err := cmd.Scan(&directPointString)

			if point.Latitude != 0 && point.Longitude != 0 {
				mapPointsBuyers[stringID] = append(mapPointsBuyers[stringID], point)
				mapPointsGlobal = append(mapPointsGlobal, point)

				var onNN uint
				if point.OnNetworkNext {
					onNN = 1
				}
				mapPointsBuyersCompact[stringID] = append(mapPointsBuyersCompact[stringID], []interface{}{point.Longitude, point.Latitude, onNN})
				mapPointsGlobalCompact = append(mapPointsGlobalCompact, []interface{}{point.Longitude, point.Latitude, onNN})
			}
		}

		for _, nextPointString := range nextPointStrings {
			// scan the data from Redis into its SessionMapPoint struct
			err := cmd.Scan(&nextPointString)

			if point.Latitude != 0 && point.Longitude != 0 {
				mapPointsBuyers[stringID] = append(mapPointsBuyers[stringID], point)
				mapPointsGlobal = append(mapPointsGlobal, point)

				var onNN uint
				if point.OnNetworkNext {
					onNN = 1
				}
				mapPointsBuyersCompact[stringID] = append(mapPointsBuyersCompact[stringID], []interface{}{point.Longitude, point.Latitude, onNN})
				mapPointsGlobalCompact = append(mapPointsGlobalCompact, []interface{}{point.Longitude, point.Latitude, onNN})
			}
		}

		s.mapPointsBuyerCache[stringID], err = json.Marshal(mapPointsBuyers[stringID])
		if err != nil {
			return err
		}

		s.mapPointsCompactBuyerCache[stringID], err = json.Marshal(mapPointsBuyersCompact[stringID])
		if err != nil {
			return err
		}
	}

	// marshal the map points slice to local cache
	s.mapPointsCache, err = json.Marshal(mapPointsGlobal)
	if err != nil {
		return err
	}

	s.mapPointsCompactCache, err = json.Marshal(mapPointsGlobalCompact)
	if err != nil {
		return err
	}
	return nil
}

func (s *BuyersService) GenerateMapPointsPerBuyerBytes() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var mapPointsGlobal mapPointsByte
	var mapPointsBuyer mapPointsByte

	s.mapPointsBuyerByteCache = make(map[string][]byte, 0)
	s.mapPointsByteCache = make([]byte, 0)

	buyers := s.Storage.Buyers()

	for _, buyer := range buyers {
		stringID := fmt.Sprintf("%016x", buyer.ID)

		mapPointsBuyer.GreenPoints = make([]point, 0)
		mapPointsBuyer.BluePoints = make([]point, 0)

		directPointStrings, nextPointStrings, err := s.getDirectAndNextMapPointStrings(&buyer)
		if err != nil {
			err = fmt.Errorf("SessionMapPoints() failed getting map points for buyer %s: %v", stringID, err)
			s.Logger.Log("err", err)
			return err
		}

		var currentPoint transport.SessionMapPoint
		for _, directPointString := range directPointStrings {
			// scan the data from Redis into its SessionMapPoint struct
			err := cmd.Scan(&directPointString)

			if currentPoint.Latitude != 0 && currentPoint.Longitude != 0 {
				bytePoint := point{
					Latitude:      float32(currentPoint.Latitude),
					Longitude:     float32(currentPoint.Longitude),
					OnNetworkNext: false,
				}

				mapPointsGlobal.BluePoints = append(mapPointsGlobal.BluePoints, bytePoint)
				mapPointsBuyer.BluePoints = append(mapPointsGlobal.BluePoints, bytePoint)
			}
		}

		for _, nextPointString := range nextPointStrings {
			// scan the data from Redis into its SessionMapPoint struct
			err := cmd.Scan(&nextPointString)

			if currentPoint.Latitude != 0 && currentPoint.Longitude != 0 {
				bytePoint := point{
					Latitude:      float32(currentPoint.Latitude),
					Longitude:     float32(currentPoint.Longitude),
					OnNetworkNext: false,
				}

				bytePoint.OnNetworkNext = true
				mapPointsGlobal.GreenPoints = append(mapPointsGlobal.GreenPoints, bytePoint)
				mapPointsBuyer.GreenPoints = append(mapPointsGlobal.GreenPoints, bytePoint)
			}
		}

		// Write entries to byte cache
		s.mapPointsBuyerByteCache[stringID] = WriteMapPointCache(&mapPointsBuyer)
	}
	s.mapPointsByteCache = WriteMapPointCache(&mapPointsGlobal)

	return nil
}

func (s *BuyersService) getDirectAndNextMapPointStrings(buyer *routing.Buyer) ([]string, []string, error) {
	minutes := time.Now().Unix() / 60

	redisClient := s.RedisPoolSessionMap.Get()
	defer redisClient.Close()

	stringID := fmt.Sprintf("%016x", buyer.ID)
	redisClient.Send("HGETALL", fmt.Sprintf("d-%s-%d", stringID, minutes-1))
	redisClient.Send("HGETALL", fmt.Sprintf("d-%s-%d", stringID, minutes))
	redisClient.Send("HGETALL", fmt.Sprintf("n-%s-%d", stringID, minutes-1))
	redisClient.Send("HGETALL", fmt.Sprintf("n-%s-%d", stringID, minutes))
	redisClient.Flush()

	direct, err := redis.StringMap(redisClient.Receive())
	if err != nil {
		return nil, nil, err
	}

	directB, err := redis.StringMap(redisClient.Receive())
	if err != nil {
		return nil, nil, err
	}

	for k, v := range directB {
		direct[k] = v
	}

	next, err := redis.StringMap(redisClient.Receive())
	if err != nil {
		return nil, nil, err
	}

	nextB, err := redis.StringMap(redisClient.Receive())
	if err != nil {
		return nil, nil, err
	}

	for k, v := range nextB {
		next[k] = v
	}

	for _, directSessionID := range direct {
		redisClient.Send("HGET", fmt.Sprintf("n-%s-%d", stringID, minutes), fmt.Sprintf("%s", directSessionID))
	}

	redisClient.Flush()
	directPointStrings, err := redis.Strings(redisClient.Receive())

	for _, nextSessionID := range next {
		redisClient.Send("HGET", fmt.Sprintf("n-%s-%d", stringID, minutes), fmt.Sprintf("%s", nextSessionID))
	}

	redisClient.Flush()
	nextPointStrings, err := redis.Strings(redisClient.Receive())

	return directPointStrings, nextPointStrings, nil
}

func WriteMapPointCache(points *mapPointsByte) []byte {
	var length uint32
	numGreenPoints := uint32(len(points.GreenPoints))
	numBluePoints := uint32(len(points.BluePoints))

	length = 1 + 4 + 4 + (1+4+4)*numGreenPoints + (1+4+4)*numBluePoints

	data := make([]byte, length)
	index := 0
	encoding.WriteUint8(data, &index, MapPointByteCacheVersion)

	encoding.WriteUint32(data, &index, numGreenPoints)

	for _, point := range points.GreenPoints {
		encoding.WriteFloat32(data, &index, point.Latitude)
		encoding.WriteFloat32(data, &index, point.Longitude)
		encoding.WriteBool(data, &index, point.OnNetworkNext)
	}

	encoding.WriteUint32(data, &index, numBluePoints)

	for _, point := range points.BluePoints {
		encoding.WriteFloat32(data, &index, point.Latitude)
		encoding.WriteFloat32(data, &index, point.Longitude)
		encoding.WriteBool(data, &index, point.OnNetworkNext)
	}
	return data
}

func ReadMapPointsCache(points *mapPointsByte, data []byte) bool {
	var numGreenPoints uint32
	var numBluePoints uint32

	index := 0
	if !encoding.ReadUint8(data, &index, &points.Version) {
		return false
	}
	if points.Version != 1 {
		return false
	}
	if !encoding.ReadUint32(data, &index, &numGreenPoints) {
		return false
	}

	var latitude float32
	var longitude float32
	var onNetworkNext bool
	points.GreenPoints = make([]point, numGreenPoints)

	for i := 0; i < int(numGreenPoints); i++ {
		if !encoding.ReadFloat32(data, &index, &latitude) {
			return false
		}
		if !encoding.ReadFloat32(data, &index, &longitude) {
			return false
		}
		if !encoding.ReadBool(data, &index, &onNetworkNext) {
			return false
		}
		points.GreenPoints[i] = point{
			Latitude:      latitude,
			Longitude:     longitude,
			OnNetworkNext: onNetworkNext,
		}
	}

	if !encoding.ReadUint32(data, &index, &numBluePoints) {
		return false
	}

	points.BluePoints = make([]point, numBluePoints)

	for i := 0; i < int(numBluePoints); i++ {
		if !encoding.ReadFloat32(data, &index, &latitude) {
			return false
		}
		if !encoding.ReadFloat32(data, &index, &longitude) {
			return false
		}
		if !encoding.ReadBool(data, &index, &onNetworkNext) {
			return false
		}
		points.BluePoints[i] = point{
			Latitude:      latitude,
			Longitude:     longitude,
			OnNetworkNext: onNetworkNext,
		}
	}

	return true
}

func (s *BuyersService) SessionMap(r *http.Request, args *MapPointsArgs, reply *MapPointsReply) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	switch args.BuyerID {
	case "":
		// pull the local cache and reply with it
		reply.Points = s.mapPointsCompactCache
	default:
		if !VerifyAllRoles(r, s.SameBuyerRole(args.BuyerID)) {
			err := fmt.Errorf("SessionMap(): %v", ErrInsufficientPrivileges)
			s.Logger.Log("err", err)
			return err
		}
		reply.Points = s.mapPointsCompactBuyerCache[args.BuyerID]
	}

	return nil
}

func (s *BuyersService) SessionMapPoints(r *http.Request, args *MapPointsArgs, reply *MapPointsReply) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	switch args.BuyerID {
	case "":
		// pull the local cache and reply with it
		reply.Points = s.mapPointsCache
	default:
		if !VerifyAllRoles(r, s.SameBuyerRole(args.BuyerID)) {
			err := fmt.Errorf("SessionMap(): %v", ErrInsufficientPrivileges)
			s.Logger.Log("err", err)
			return err
		}
		reply.Points = s.mapPointsBuyerCache[args.BuyerID]
	}

	return nil
}

// SessionMapPoints returns the locally cached JSON from GenerateSessionMapPoints
func (s *BuyersService) SessionMapPointsByte(r *http.Request, args *MapPointsArgs, reply *MapPointsByteReply) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch args.BuyerID {
	case "":
		// pull the local cache and reply with it
		ReadMapPointsCache(&reply.Points, s.mapPointsCache)
	default:
		if !VerifyAllRoles(r, s.SameBuyerRole(args.BuyerID)) {
			err := fmt.Errorf("SessionMap(): %v", ErrInsufficientPrivileges)
			s.Logger.Log("err", err)
			return err
		}
		ReadMapPointsCache(&reply.Points, s.mapPointsBuyerCache[args.BuyerID])
	}

	return nil
}

func (s *BuyersService) SessionMapByte(r *http.Request, args *MapPointsArgs, reply *MapPointsByteReply) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch args.BuyerID {
	case "":
		// pull the local cache and reply with it
		ReadMapPointsCache(&reply.Points, s.mapPointsByteCache)
	default:
		if !VerifyAllRoles(r, s.SameBuyerRole(args.BuyerID)) {
			err := fmt.Errorf("SessionMap(): %v", ErrInsufficientPrivileges)
			s.Logger.Log("err", err)
			return err
		}
		ReadMapPointsCache(&reply.Points, s.mapPointsBuyerByteCache[args.BuyerID])
	}

	return nil
}

type GameConfigurationArgs struct {
	Domain       string `json:"domain"`
	Name         string `json:"name"`
	NewPublicKey string `json:"new_public_key"`
}

type GameConfigurationReply struct {
	GameConfiguration gameConfiguration `json:"game_config"`
}

type gameConfiguration struct {
	Company   string `json:"company"`
	PublicKey string `json:"public_key"`
}

func (s *BuyersService) GameConfiguration(r *http.Request, args *GameConfigurationArgs, reply *GameConfigurationReply) error {
	var err error
	var buyer routing.Buyer

	if VerifyAnyRole(r, AnonymousRole, UnverifiedRole) {
		err = fmt.Errorf("GameConfiguration(): %v", ErrInsufficientPrivileges)
		s.Logger.Log("err", err)
		return err
	}

	reply.GameConfiguration.PublicKey = ""
	reply.GameConfiguration.Company = ""

	buyer, err = s.Storage.BuyerWithDomain(args.Domain)
	// Buyer not found
	if err != nil {
		return nil
	}

	reply.GameConfiguration.PublicKey = buyer.EncodedPublicKey()
	reply.GameConfiguration.Company = buyer.Name

	return nil
}

func (s *BuyersService) UpdateGameConfiguration(r *http.Request, args *GameConfigurationArgs, reply *GameConfigurationReply) error {
	var err error
	var buyerID uint64
	var buyer routing.Buyer

	if !VerifyAnyRole(r, AdminRole, OwnerRole) {
		err = fmt.Errorf("UpdateGameConfiguration(): %v", ErrInsufficientPrivileges)
		s.Logger.Log("err", err)
		return err
	}

	ctx := context.Background()

	if args.Domain == "" {
		err = fmt.Errorf("UpdateGameConfiguration() domain is required")
		s.Logger.Log("err", err)
		return err
	}

	if args.Name == "" {
		err = fmt.Errorf("UpdateGameConfiguration() company name is required")
		s.Logger.Log("err", err)
		return err
	}

	if args.NewPublicKey == "" {
		err = fmt.Errorf("UpdateGameConfiguration() new public key is required")
		s.Logger.Log("err", err)
		return err
	}

	buyer, err = s.Storage.BuyerWithDomain(args.Domain)

	byteKey, err := base64.StdEncoding.DecodeString(args.NewPublicKey)
	if err != nil {
		err = fmt.Errorf("UpdateGameConfiguration() could not decode public key string")
		s.Logger.Log("err", err)
		return err
	}

	buyerID = binary.LittleEndian.Uint64(byteKey[0:8])

	// Buyer not found
	if buyer.ID == 0 {

		// Create new buyer
		err = s.Storage.AddBuyer(ctx, routing.Buyer{
			ID:        buyerID,
			Name:      args.Name,
			Domain:    args.Domain,
			Active:    true,
			Live:      false,
			PublicKey: byteKey[8:],
		})

		if err != nil {
			err = fmt.Errorf("UpdateGameConfiguration() failed to add buyer")
			s.Logger.Log("err", err)
			return err
		}

		// Check if buyer is associated with the ID and everything worked
		if buyer, err = s.Storage.Buyer(buyerID); err != nil {
			err = fmt.Errorf("UpdateGameConfiguration() buyer creation failed: %v", err)
			s.Logger.Log("err", err)
			return err
		}

		// Setup reply
		reply.GameConfiguration.PublicKey = buyer.EncodedPublicKey()
		reply.GameConfiguration.Company = buyer.Name

		return nil
	}

	live := buyer.Live
	active := buyer.Active

	if err = s.Storage.RemoveBuyer(ctx, buyer.ID); err != nil {
		err = fmt.Errorf("UpdateGameConfiguration() failed to remove buyer")
		s.Logger.Log("err", err)
		return err
	}

	err = s.Storage.AddBuyer(ctx, routing.Buyer{
		ID:        buyerID,
		Name:      args.Name,
		Domain:    args.Domain,
		Active:    active,
		Live:      live,
		PublicKey: byteKey[8:],
	})

	if err != nil {
		err = fmt.Errorf("UpdateGameConfiguration() buyer update failed: %v", err)
		s.Logger.Log("err", err)
		return err
	}

	// Check if buyer is associated with the ID and everything worked
	if buyer, err = s.Storage.Buyer(buyerID); err != nil {
		err = fmt.Errorf("UpdateGameConfiguration() buyer update check failed: %v", err)
		s.Logger.Log("err", err)
		return err
	}

	// Set reply
	reply.GameConfiguration.PublicKey = buyer.EncodedPublicKey()
	reply.GameConfiguration.Company = buyer.Name

	return nil
}

type BuyerListArgs struct{}

type BuyerListReply struct {
	Buyers []buyerAccount `json:"buyers"`
}

type buyerAccount struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	IsLive bool   `json:"is_live"`
}

func (s *BuyersService) Buyers(r *http.Request, args *BuyerListArgs, reply *BuyerListReply) error {
	reply.Buyers = make([]buyerAccount, 0)
	if VerifyAllRoles(r, AnonymousRole) {
		return nil
	}

	for _, b := range s.Storage.Buyers() {
		id := fmt.Sprintf("%016x", b.ID)
		account := buyerAccount{
			ID:     id,
			Name:   b.Name,
			IsLive: b.Live,
		}
		if VerifyAllRoles(r, s.SameBuyerRole(id)) {
			reply.Buyers = append(reply.Buyers, account)
		}
	}

	sort.Slice(reply.Buyers, func(i int, j int) bool {
		return reply.Buyers[i].Name < reply.Buyers[j].Name
	})

	return nil
}

type DatacenterMapsArgs struct {
	ID uint64 `json:"buyer_id"`
}

type DatacenterMapsFull struct {
	Alias          string
	DatacenterName string
	DatacenterID   string
	BuyerName      string
	BuyerID        string
	SupplierName   string
}

type DatacenterMapsReply struct {
	DatacenterMaps []DatacenterMapsFull
}

func (s *BuyersService) DatacenterMapsForBuyer(r *http.Request, args *DatacenterMapsArgs, reply *DatacenterMapsReply) error {
	if VerifyAllRoles(r, AnonymousRole) {
		return nil
	}

	var dcm map[uint64]routing.DatacenterMap

	dcm = s.Storage.GetDatacenterMapsForBuyer(args.ID)

	var replySlice []DatacenterMapsFull
	for _, dcMap := range dcm {
		buyer, err := s.Storage.Buyer(dcMap.BuyerID)
		if err != nil {
			err = fmt.Errorf("DatacenterMapsForBuyer() could not parse buyer")
			s.Logger.Log("err", err)
			return err
		}
		datacenter, err := s.Storage.Datacenter(dcMap.Datacenter)
		if err != nil {
			err = fmt.Errorf("DatacenterMapsForBuyer() could not parse datacenter")
			s.Logger.Log("err", err)
			return err
		}

		dcmFull := DatacenterMapsFull{
			Alias:          dcMap.Alias,
			DatacenterName: datacenter.Name,
			DatacenterID:   fmt.Sprintf("%016x", dcMap.Datacenter),
			BuyerName:      buyer.Name,
			BuyerID:        fmt.Sprintf("%016x", dcMap.BuyerID),
			SupplierName:   datacenter.SupplierName,
		}

		replySlice = append(replySlice, dcmFull)
	}

	reply.DatacenterMaps = replySlice
	return nil

}

type RemoveDatacenterMapArgs struct {
	DatacenterMap routing.DatacenterMap
}

type RemoveDatacenterMapReply struct{}

func (s *BuyersService) RemoveDatacenterMap(r *http.Request, args *RemoveDatacenterMapArgs, reply *RemoveDatacenterMapReply) error {
	ctx, cancelFunc := context.WithDeadline(context.Background(), time.Now().Add(10*time.Second))
	defer cancelFunc()

	return s.Storage.RemoveDatacenterMap(ctx, args.DatacenterMap)

}

type AddDatacenterMapArgs struct {
	DatacenterMap routing.DatacenterMap
}

type AddDatacenterMapReply struct{}

func (s *BuyersService) AddDatacenterMap(r *http.Request, args *AddDatacenterMapArgs, reply *AddDatacenterMapReply) error {

	ctx, cancelFunc := context.WithDeadline(context.Background(), time.Now().Add(10*time.Second))
	defer cancelFunc()

	return s.Storage.AddDatacenterMap(ctx, args.DatacenterMap)

}

// SameBuyerRole checks the JWT for the correct passed in buyerID
func (s *BuyersService) SameBuyerRole(buyerID string) RoleFunc {
	return func(req *http.Request) (bool, error) {
		if VerifyAnyRole(req, AdminRole, OpsRole) {
			return true, nil
		}
		if VerifyAllRoles(req, AnonymousRole) {
			return false, nil
		}

		if buyerID == "" {
			return false, fmt.Errorf("SameBuyerRole(): buyerID is required")
		}

		requestUser := req.Context().Value("user")
		if requestUser == nil {
			return false, fmt.Errorf("SameBuyerRole(): unable to parse user from token")
		}

		requestEmail, ok := requestUser.(*jwt.Token).Claims.(jwt.MapClaims)["email"].(string)
		if !ok {
			return false, fmt.Errorf("SameBuyerRole(): unable to parse email from token")
		}
		requestEmailParts := strings.Split(requestEmail, "@")
		requestDomain := requestEmailParts[len(requestEmailParts)-1] // Domain is the last entry of the split since an email as only one @ sign
		buyer, err := s.Storage.BuyerWithDomain(requestDomain)
		if err != nil {
			return false, fmt.Errorf("SameBuyerRole(): BuyerWithDomain error: %v", err)
		}

		return buyerID == fmt.Sprintf("%016x", buyer.ID), nil
	}
}
