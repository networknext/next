package jsonrpc

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	fnv "hash/fnv"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/go-redis/redis/v7"
	"github.com/networknext/backend/encoding"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
	"github.com/networknext/backend/transport"
)

const (
	TopSessionsSize          = 1000
	TopNextSessionsSize      = 1200
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

	RedisClient redis.Cmdable
	Storage     storage.Storer
	Logger      log.Logger
}

type FlushSessionsArgs struct{}

type FlushSessionsReply struct{}

func (s *BuyersService) FlushSessions(r *http.Request, args *FlushSessionsArgs, reply *FlushSessionsReply) error {
	if !VerifyAllRoles(r, OpsRole) {
		return fmt.Errorf("FlushSessions(): %v", ErrInsufficientPrivileges)
	}

	return s.RedisClient.FlushAllAsync().Err()
}

type UserSessionsArgs struct {
	UserHash string `json:"user_hash"`
}

type UserSessionsReply struct {
	Sessions []transport.SessionMeta `json:"sessions"`
}

func (s *BuyersService) UserSessions(r *http.Request, args *UserSessionsArgs, reply *UserSessionsReply) error {
	var sessionIDs []string

	userhash := args.UserHash
	reply.Sessions = make([]transport.SessionMeta, 0)

	err := s.RedisClient.SMembers(fmt.Sprintf("user-%s-sessions", userhash)).ScanSlice(&sessionIDs)
	if err != nil {
		err = fmt.Errorf("UserSessions() failed getting user sessions: %v", err)
		s.Logger.Log("err", err)
		return err
	}

	if len(sessionIDs) == 0 {
		hash := fnv.New64a()
		_, err := hash.Write([]byte(userhash))
		if err != nil {
			err = fmt.Errorf("UserSessions() error writing 64a hash: %v", err)
			s.Logger.Log("err", err)
			return err
		}
		hashedID := fmt.Sprintf("%016x", hash.Sum64())

		err = s.RedisClient.SMembers(fmt.Sprintf("user-%s-sessions", hashedID)).ScanSlice(&sessionIDs)
		if err != nil {
			err = fmt.Errorf("UserSessions() failed getting user sessions: %v", err)
			s.Logger.Log("err", err)
			return err
		}
	}

	if len(sessionIDs) == 0 {
		return nil
	}

	var getCmds []*redis.StringCmd
	{
		gettx := s.RedisClient.TxPipeline()
		for _, sessionID := range sessionIDs {
			getCmds = append(getCmds, gettx.Get(fmt.Sprintf("session-%s-meta", sessionID)))
		}
		_, err = gettx.Exec()
		if err != nil && err != redis.Nil {
			err = fmt.Errorf("UserSessions() redis.Pipeliner error: %v", err)
			s.Logger.Log("err", err)
			return err
		}
	}

	sremtx := s.RedisClient.TxPipeline()
	{
		var meta transport.SessionMeta
		for _, cmd := range getCmds {
			err = cmd.Scan(&meta)
			if err != nil {
				args := cmd.Args()
				key := args[1].(string)
				keyparts := strings.Split(key, "-")

				sremtx.SRem(fmt.Sprintf("user-%s-sessions", userhash), keyparts[1])
				continue
			}

			if VerifyAnyRole(r, AnonymousRole, UnverifiedRole) || !VerifyAllRoles(r, s.SameBuyerRole(fmt.Sprintf("%016x", meta.BuyerID))) {
				meta.Anonymise()
			}

			reply.Sessions = append(reply.Sessions, meta)
		}
	}

	sremcmds, err := sremtx.Exec()
	if err != nil && err != redis.Nil {
		err = fmt.Errorf("UserSessions() redis.Pipeliner error: %v", err)
		s.Logger.Log("err", err)
		return err
	}

	level.Info(s.Logger).Log("key", "user-*-sessions", "removed", len(sremcmds))

	sort.Slice(reply.Sessions, func(i int, j int) bool {
		return reply.Sessions[i].ID < reply.Sessions[j].ID
	})

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
	if r.Body != nil {
		defer r.Body.Close()
	}

	// get the top session IDs globally or for a buyer from the sorted set
	switch args.BuyerID {
	case "":
		// Get top Direct sessions sorted by least to greatest direct RTT
		directTotals, err := s.RedisClient.HGetAll("session-count-total-direct").Result()
		if err != nil {
			err = fmt.Errorf("TotalSessions() failed getting session-count-total-direct: %v", err)
			s.Logger.Log("err", err)
			return err
		}

		// Get top Next sessions
		nextTotals, err := s.RedisClient.HGetAll("session-count-total-next").Result()
		if err != nil {
			err = fmt.Errorf("TotalSessions() failed getting session-count-total-next: %v", err)
			s.Logger.Log("err", err)
			return err
		}

		var totalDirect int
		for _, directString := range directTotals {
			direct, err := strconv.Atoi(directString)
			if err != nil {
				err = fmt.Errorf("TotalSessions() failed to parse direct session count (%s): %v", directString, err)
				s.Logger.Log("err", err)
				return err
			}

			totalDirect += direct
		}

		var totalNext int
		for _, nextString := range nextTotals {
			next, err := strconv.Atoi(nextString)
			if err != nil {
				err = fmt.Errorf("TotalSessions() failed to parse next session count (%s): %v", nextString, err)
				s.Logger.Log("err", err)
				return err
			}
			totalNext += next
		}

		reply.Direct = totalDirect
		reply.Next = totalNext
	default:
		if !VerifyAllRoles(r, s.SameBuyerRole(args.BuyerID)) {
			err := fmt.Errorf("TotalSessions(): %v", ErrInsufficientPrivileges)
			s.Logger.Log("err", err)
			return err
		}

		buyerDirectTotals, err := s.RedisClient.HGetAll(fmt.Sprintf("session-count-direct-buyer-%s", args.BuyerID)).Result()
		if err != nil {
			err = fmt.Errorf("TotalSessions() failed getting session-count-direct-buyer-%s: %v", args.BuyerID, err)
			s.Logger.Log("err", err)
			return err
		}

		buyerNextTotals, err := s.RedisClient.HGetAll(fmt.Sprintf("session-count-next-buyer-%s", args.BuyerID)).Result()
		if err != nil {
			err = fmt.Errorf("TotalSessions() failed getting session-count-next-buyer-%s: %v", args.BuyerID, err)
			s.Logger.Log("err", err)
			return err
		}

		var buyerDirectTotal int
		for _, directString := range buyerDirectTotals {
			direct, err := strconv.Atoi(directString)
			if err != nil {
				err = fmt.Errorf("TotalSessions() failed to parse buyer direct session count (%s): %v", directString, err)
				s.Logger.Log("err", err)
				return err
			}

			buyerDirectTotal += direct
		}

		var buyerNextTotal int
		for _, nextString := range buyerNextTotals {
			next, err := strconv.Atoi(nextString)
			if err != nil {
				err = fmt.Errorf("TotalSessions() failed to parse buyer next session count (%s): %v", nextString, err)
				s.Logger.Log("err", err)
				return err
			}

			buyerNextTotal += next
		}

		reply.Direct = buyerDirectTotal
		reply.Next = buyerNextTotal
	}

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
	var topnext []string
	var topdirect []string

	reply.Sessions = make([]transport.SessionMeta, 0)

	buyers := s.Storage.Buyers()

	// get the top session IDs globally or for a buyer from the sorted set
	switch args.BuyerID {
	case "":
		// Get top Next sessions sorted by greatest to least improved RTT
		topnext, err = s.RedisClient.ZRevRange("total-next", 0, TopNextSessionsSize).Result()
		if err != nil {
			err = fmt.Errorf("TopSessions() failed getting total-next sessions: %v", err)
			s.Logger.Log("err", err)
			return err
		}

		// Get top Direct sessions sorted by least to greatest direct RTT
		topdirect, err = s.RedisClient.ZRange("total-direct", 0, TopSessionsSize).Result()
		if err != nil {
			err = fmt.Errorf("TopSessions() failed getting total-next sessions: %v", err)
			s.Logger.Log("err", err)
			return err
		}
	default:
		if !VerifyAllRoles(r, s.SameBuyerRole(args.BuyerID)) {
			err = fmt.Errorf("TopSessions(): %v", ErrInsufficientPrivileges)
			s.Logger.Log("err", err)
			return err
		}
		topnext, err = s.RedisClient.ZRevRange(fmt.Sprintf("total-next-buyer-%s", args.BuyerID), 0, TopNextSessionsSize).Result()
		if err != nil {
			err = fmt.Errorf("TopSessions() failed getting total-next sessions: %v", err)
			s.Logger.Log("err", err)
			return err
		}
		topdirect, err = s.RedisClient.ZRange(fmt.Sprintf("total-direct-buyer-%s", args.BuyerID), 0, TopSessionsSize).Result()
		if err != nil {
			err = fmt.Errorf("TopSessions() failed getting total-next sessions: %v", err)
			s.Logger.Log("err", err)
			return err
		}
	}

	// build a single transaction to get the sessions details for the session IDs
	var getNextCmds []*redis.StringCmd
	{
		gettx := s.RedisClient.TxPipeline()
		for _, sessionID := range topnext {
			getNextCmds = append(getNextCmds, gettx.Get(fmt.Sprintf("session-%s-meta", sessionID)))
		}
		_, err = gettx.Exec()
		if err != nil && err != redis.Nil {
			err = fmt.Errorf("TopSessions() failed getting top sessions meta: %v", err)
			s.Logger.Log("err", err)
			return err
		}
	}

	// build a single transaction to remove any session ID from the sorted set if the
	// session-*-meta key is missing or expired
	var nextSessions []transport.SessionMeta
	zremtx := s.RedisClient.TxPipeline()
	{
		var meta transport.SessionMeta
		for _, cmd := range getNextCmds {
			// scan the data from Redis into its SessionMeta struct
			err = cmd.Scan(&meta)

			// if there was an error then the session-*-point key expired
			// so add ZREM commands to remove it from the key sets globally
			// and for each buyer
			if err != nil {
				args := cmd.Args()
				key := args[1].(string)
				keyparts := strings.Split(key, "-")

				zremtx.ZRem("total-next", keyparts[1])
				for _, buyer := range buyers {
					zremtx.ZRem(fmt.Sprintf("total-next-buyer-%016x", buyer.ID), keyparts[1])
				}
				continue
			}

			if !VerifyAllRoles(r, s.SameBuyerRole(args.BuyerID)) {
				meta.Anonymise()
			}

			nextSessions = append(nextSessions, meta)
		}
	}

	_, err = zremtx.Exec()
	if err != nil && err != redis.Nil {
		err = fmt.Errorf("TopSessions() redis.Pipeliner error: %v", err)
		s.Logger.Log("err", err)
		return err
	}

	sort.Slice(nextSessions, func(i int, j int) bool {
		return nextSessions[i].DeltaRTT > nextSessions[j].DeltaRTT
	})

	// If the result of nextSessions fills the page then early out and do not fill with direct
	// This will skip talking to redis to get meta details for sessions we do not need to get
	if len(nextSessions) >= TopSessionsSize {
		reply.Sessions = nextSessions[:TopSessionsSize]
		return nil
	}

	// If we get here then there are not enough Next sessions to fill the list
	// so we continue to get the top Direct sessions to fill it out more

	var getDirectCmds []*redis.StringCmd
	{
		gettx := s.RedisClient.TxPipeline()
		for _, sessionID := range topdirect {
			getDirectCmds = append(getDirectCmds, gettx.Get(fmt.Sprintf("session-%s-meta", sessionID)))
		}
		_, err = gettx.Exec()
		if err != nil && err != redis.Nil {
			err = fmt.Errorf("TopSessions() failed getting top sessions meta: %v", err)
			s.Logger.Log("err", err)
			return err
		}
	}

	var directSessions []transport.SessionMeta
	zremtx = s.RedisClient.TxPipeline()
	{
		var meta transport.SessionMeta
		for _, cmd := range getDirectCmds {
			// scan the data from Redis into its SessionMeta struct
			err = cmd.Scan(&meta)

			// if there was an error then the session-*-point key expired
			// so add ZREM commands to remove it from the key sets globally
			// and for each buyer
			if err != nil {
				args := cmd.Args()
				key := args[1].(string)
				keyparts := strings.Split(key, "-")

				zremtx.ZRem("total-direct", keyparts[1])
				for _, buyer := range buyers {
					zremtx.ZRem(fmt.Sprintf("total-direct-buyer-%016x", buyer.ID), keyparts[1])
				}
				continue
			}

			if !VerifyAllRoles(r, s.SameBuyerRole(args.BuyerID)) {
				meta.Anonymise()
			}

			directSessions = append(directSessions, meta)
		}
	}

	// execute the transaction to remove the sessions IDs from the sorted key sets
	_, err = zremtx.Exec()
	if err != nil && err != redis.Nil {
		err = fmt.Errorf("TopSessions() redis.Pipeliner error: %v", err)
		s.Logger.Log("err", err)
		return err
	}

	// IMPORTANT: Clean direct sessions to remove any that are also in the next set
	directMap := make(map[uint64]*transport.SessionMeta)
	for i := range directSessions {
		directMap[directSessions[i].ID] = &directSessions[i]
	}
	for i := range nextSessions {
		delete(directMap, nextSessions[i].ID)
	}
	cleanDirectSessions := make([]transport.SessionMeta, 0)
	for _, v := range directMap {
		cleanDirectSessions = append(cleanDirectSessions, *v)
	}

	// Sort cleaned direct slices in order of least to greatest direct RTT
	sort.Slice(cleanDirectSessions, func(i int, j int) bool {
		return cleanDirectSessions[i].DirectRTT < cleanDirectSessions[j].DirectRTT
	})

	// Append the two sets. next sessions first, followed by direct sessions that are not included in the next set.
	reply.Sessions = append(reply.Sessions, nextSessions...)
	reply.Sessions = append(reply.Sessions, cleanDirectSessions...)

	if len(reply.Sessions) > TopSessionsSize {
		reply.Sessions = reply.Sessions[:TopSessionsSize]
	}

	allDirectSessions := len(nextSessions) == 0

	sort.SliceStable(reply.Sessions, func(i int, j int) bool {
		firstSession := reply.Sessions[i]
		secondSession := reply.Sessions[j]
		if allDirectSessions {
			return firstSession.DirectRTT > secondSession.DirectRTT
		}
		if firstSession.OnNetworkNext && secondSession.OnNetworkNext {
			return firstSession.DeltaRTT > secondSession.DeltaRTT
		}
		if firstSession.OnNetworkNext && !secondSession.OnNetworkNext {
			return true
		}
		if !firstSession.OnNetworkNext && secondSession.OnNetworkNext {
			return false
		}
		return firstSession.DirectRTT < secondSession.DirectRTT
	})

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

	data, err := s.RedisClient.Get(fmt.Sprintf("session-%s-meta", args.SessionID)).Bytes()
	if err != nil {
		err = fmt.Errorf("SessionDetails() failed getting session meta: %v", err)
		s.Logger.Log("err", err)
		return err
	}
	err = reply.Meta.UnmarshalBinary(data)
	if err != nil {
		err = fmt.Errorf("SessionDetails() SessionMeta unmarshaling error: %v", err)
		s.Logger.Log("err", err)
		return err
	}

	if !VerifyAllRoles(r, s.SameBuyerRole(fmt.Sprintf("%016x", reply.Meta.BuyerID))) {
		reply.Meta.Anonymise()
	}

	reply.Slices = make([]transport.SessionSlice, 0)

	err = s.RedisClient.SMembers(fmt.Sprintf("session-%s-slices", args.SessionID)).ScanSlice(&reply.Slices)
	if err != nil {
		err = fmt.Errorf("SessionDetails() failed getting session slices: %v", err)
		s.Logger.Log("err", err)
		return err
	}

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

	var sessionIDs []string
	var err error

	buyers := s.Storage.Buyers()

	// slice to hold all the final map points
	mapPointsBuyers := make(map[string][]transport.SessionMapPoint, 0)
	mapPointsBuyersCompact := make(map[string][][]interface{}, 0)
	mapPointsGlobal := make([]transport.SessionMapPoint, 0)
	mapPointsGlobalCompact := make([][]interface{}, 0)

	s.mapPointsBuyerCache = make(map[string]json.RawMessage, 0)
	s.mapPointsCompactBuyerCache = make(map[string]json.RawMessage, 0)

	for _, buyer := range buyers {
		stringID := fmt.Sprintf("%016x", buyer.ID)
		sessionIDs, err = s.RedisClient.SMembers(fmt.Sprintf("map-points-%016x-buyer", buyer.ID)).Result()
		if err != nil {
			err = fmt.Errorf("SessionMapPoints() failed getting map points for buyer %016x: %v", buyer.ID, err)
			s.Logger.Log("err", err)
			return err
		}

		// build a single transaction of gets for each session ID
		var getCmds []*redis.StringCmd
		{
			gettx := s.RedisClient.TxPipeline()
			for _, sessionID := range sessionIDs {
				getCmds = append(getCmds, gettx.Get(fmt.Sprintf("session-%s-point", sessionID)))
			}
			_, err = gettx.Exec()
			if err != nil && err != redis.Nil {
				err = fmt.Errorf("SessionMapPoints() failed getting session points: %v", err)
				s.Logger.Log("err", err)
				return err
			}
		}

		// build a single transaction for any missing session-*-point keys to be
		// removed from map-points-buyer, total-next, and total-direct key sets
		sremtx := s.RedisClient.TxPipeline()
		{
			var point transport.SessionMapPoint
			for _, cmd := range getCmds {
				// scan the data from Redis into its SessionMapPoint struct
				err = cmd.Scan(&point)

				// if there was an error then the session-*-point key expired
				// so add SREM commands to remove it from the key sets
				if err != nil {
					key := cmd.Args()[1].(string)
					keyparts := strings.Split(key, "-")
					sremtx.SRem(fmt.Sprintf("map-points-%016x-buyer", buyer.ID), keyparts[1])
					sremtx.ZRem("total-next", keyparts[1])
					sremtx.ZRem(fmt.Sprintf("total-next-buyer-%016x", buyer.ID), keyparts[1])
					sremtx.ZRem("total-direct", keyparts[1])
					sremtx.ZRem(fmt.Sprintf("total-direct-buyer-%016x", buyer.ID), keyparts[1])
					continue
				}

				// if there was no error then add the SessionMapPoint to the slice
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
		}

		// execute the transaction to remove the sessions IDs from the key sets
		_, err := sremtx.Exec()
		if err != nil {
			return err
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

	var sessionIDs []string
	var stringID string
	var err error
	var mapPointsGlobal mapPointsByte
	var mapPointsBuyer mapPointsByte

	s.mapPointsBuyerByteCache = make(map[string][]byte, 0)
	s.mapPointsByteCache = make([]byte, 0)

	buyers := s.Storage.Buyers()

	for _, buyer := range buyers {
		mapPointsBuyer.GreenPoints = make([]point, 0)
		mapPointsBuyer.BluePoints = make([]point, 0)

		stringID = fmt.Sprintf("%016x", buyer.ID)
		sessionIDs, err = s.RedisClient.SMembers(fmt.Sprintf("map-points-%016x-buyer", buyer.ID)).Result()
		if err != nil {
			err = fmt.Errorf("SessionMapPoints() failed getting map points for buyer %016x: %v", buyer.ID, err)
			s.Logger.Log("err", err)
			return err
		}

		// build a single transaction of gets for each session ID
		var getCmds []*redis.StringCmd
		{
			gettx := s.RedisClient.TxPipeline()
			for _, sessionID := range sessionIDs {
				getCmds = append(getCmds, gettx.Get(fmt.Sprintf("session-%s-point", sessionID)))
			}
			_, err = gettx.Exec()
			if err != nil && err != redis.Nil {
				err = fmt.Errorf("SessionMapPoints() failed getting session points: %v", err)
				s.Logger.Log("err", err)
				return err
			}
		}

		// build a single transaction for any missing session-*-point keys to be
		// removed from map-points-buyer, total-next, and total-direct key sets
		sremtx := s.RedisClient.TxPipeline()
		{
			var currentPoint transport.SessionMapPoint
			for _, cmd := range getCmds {
				// scan the data from Redis into its SessionMapPoint struct
				err = cmd.Scan(&currentPoint)

				// if there was an error then the session-*-point key expired
				// so add SREM commands to remove it from the key sets
				if err != nil {
					key := cmd.Args()[1].(string)
					keyparts := strings.Split(key, "-")
					sremtx.SRem(fmt.Sprintf("map-points-%016x-buyer", buyer.ID), keyparts[1])
					sremtx.ZRem("total-next", keyparts[1])
					sremtx.ZRem(fmt.Sprintf("total-next-buyer-%016x", buyer.ID), keyparts[1])
					sremtx.ZRem("total-direct", keyparts[1])
					sremtx.ZRem(fmt.Sprintf("total-direct-buyer-%016x", buyer.ID), keyparts[1])
					continue
				}

				// if there was no error then add the SessionMapPoint to the slice
				if currentPoint.Latitude != 0 && currentPoint.Longitude != 0 {
					bytePoint := point{
						Latitude:      float32(currentPoint.Latitude),
						Longitude:     float32(currentPoint.Longitude),
						OnNetworkNext: false,
					}

					if currentPoint.OnNetworkNext {
						bytePoint.OnNetworkNext = true
						mapPointsGlobal.GreenPoints = append(mapPointsGlobal.GreenPoints, bytePoint)
						mapPointsBuyer.GreenPoints = append(mapPointsGlobal.GreenPoints, bytePoint)
					} else {
						mapPointsGlobal.BluePoints = append(mapPointsGlobal.BluePoints, bytePoint)
						mapPointsBuyer.BluePoints = append(mapPointsGlobal.BluePoints, bytePoint)
					}
				}
			}
		}

		// execute the transaction to remove the sessions IDs from the key sets
		_, err := sremtx.Exec()
		if err != nil {
			return err
		}

		// Write entries to byte cache
		s.mapPointsBuyerByteCache[stringID] = WriteMapPointCache(&mapPointsBuyer)
	}
	s.mapPointsByteCache = WriteMapPointCache(&mapPointsGlobal)

	return nil
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
	GameConfiguration   gameConfiguration   `json:"game_config"`
	CustomerRouteShader CustomerRouteShader `json:"customer_route_shader"`
}

type gameConfiguration struct {
	BuyerID   string `json:"buyer_id"`
	Company   string `json:"company"`
	PublicKey string `json:"public_key"`
}

type CustomerRouteShader struct {
	EnableNetworkNext   bool   `json:"enable_nn"`
	EnableRoundTripTime bool   `json:"enable_rtt"`
	EnablePacketLoss    bool   `json:"enable_pl"`
	EnableABTest        bool   `json:"enable_ab"`
	EnableMultiPath     bool   `json:"enable_mp"`
	AcceptableLatency   string `json:"acceptable_latency"`
	PacketLossThreshold string `json:"pl_threshold"`
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
		reply.CustomerRouteShader = CustomerRouteShader{
			EnableNetworkNext:   true,
			EnableRoundTripTime: true,
			EnableMultiPath:     false,
			EnableABTest:        false,
			EnablePacketLoss:    false,
			AcceptableLatency:   "20",
			PacketLossThreshold: "1",
		}
		return nil
	}

	reply.GameConfiguration.PublicKey = buyer.EncodedPublicKey()
	reply.GameConfiguration.Company = buyer.Name
	reply.CustomerRouteShader = CustomerRouteShader{
		EnableNetworkNext:   buyer.CustomerRouteShader.EnableNetworkNext,
		EnableRoundTripTime: buyer.CustomerRouteShader.EnableRoundTripTime,
		EnableABTest:        buyer.CustomerRouteShader.EnableABTest,
		EnableMultiPath:     buyer.CustomerRouteShader.EnableMultiPath,
		EnablePacketLoss:    buyer.CustomerRouteShader.EnablePacketLoss,
		AcceptableLatency:   fmt.Sprintf("%v", buyer.CustomerRouteShader.AcceptableLatency),
		PacketLossThreshold: fmt.Sprintf("%v", buyer.CustomerRouteShader.PacketLossThreshold),
	}

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
			ID:                  buyerID,
			Name:                args.Name,
			Domain:              args.Domain,
			Active:              true,
			Live:                false,
			PublicKey:           byteKey[8:],
			CustomerRouteShader: routing.DefaultCustomerRouteShader,
			RouteShader:         routing.DefaultRouteShader,
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
		reply.GameConfiguration.BuyerID = fmt.Sprintf("%016x", buyer.ID)
		reply.GameConfiguration.PublicKey = buyer.EncodedPublicKey()
		reply.GameConfiguration.Company = buyer.Name
		reply.CustomerRouteShader = CustomerRouteShader{
			EnableNetworkNext:   buyer.CustomerRouteShader.EnableNetworkNext,
			EnableRoundTripTime: buyer.CustomerRouteShader.EnableRoundTripTime,
			EnableABTest:        buyer.CustomerRouteShader.EnableABTest,
			EnableMultiPath:     buyer.CustomerRouteShader.EnableMultiPath,
			EnablePacketLoss:    buyer.CustomerRouteShader.EnablePacketLoss,
			AcceptableLatency:   fmt.Sprintf("%v", buyer.CustomerRouteShader.AcceptableLatency),
			PacketLossThreshold: fmt.Sprintf("%v", buyer.CustomerRouteShader.PacketLossThreshold),
		}
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
		ID:                  buyerID,
		Name:                args.Name,
		Domain:              args.Domain,
		Active:              active,
		Live:                live,
		PublicKey:           byteKey[8:],
		CustomerRouteShader: buyer.CustomerRouteShader,
		RouteShader:         buyer.RouteShader,
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
	reply.CustomerRouteShader = CustomerRouteShader{
		EnableNetworkNext:   buyer.CustomerRouteShader.EnableNetworkNext,
		EnableRoundTripTime: buyer.CustomerRouteShader.EnableRoundTripTime,
		EnableABTest:        buyer.CustomerRouteShader.EnableABTest,
		EnableMultiPath:     buyer.CustomerRouteShader.EnableMultiPath,
		EnablePacketLoss:    buyer.CustomerRouteShader.EnablePacketLoss,
		AcceptableLatency:   fmt.Sprintf("%v", buyer.CustomerRouteShader.AcceptableLatency),
		PacketLossThreshold: fmt.Sprintf("%v", buyer.CustomerRouteShader.PacketLossThreshold),
	}
	return nil
}

type RouteShaderUpdateArgs struct {
	EnableNN          bool   `json:"enable_nn"`
	EnableRTT         bool   `json:"enable_rtt"`
	EnablePL          bool   `json:"enable_pl"`
	EnableAB          bool   `json:"enable_ab"`
	EnableMP          bool   `json:"enable_mp"`
	AcceptableLatency string `json:"acceptable_latency"`
	PLThreshold       string `json:"pl_threshold"`
}

type RouteShaderUpdateReply struct {
	CustomerRouteShader CustomerRouteShader `json:"customer_route_shader"`
}

func (s *BuyersService) UpdateRouteShader(req *http.Request, args *RouteShaderUpdateArgs, reply *RouteShaderUpdateReply) error {
	if !VerifyAnyRole(req, AdminRole, OwnerRole) {
		err := fmt.Errorf("UpdateRouteShader(): %v", ErrInsufficientPrivileges)
		s.Logger.Log("err", err)
		return err
	}

	ctx := context.Background()

	requestUser := req.Context().Value("user")
	if requestUser == nil {
		return fmt.Errorf("UpdateRouteShader(): unable to parse user from token")
	}

	requestEmail, ok := requestUser.(*jwt.Token).Claims.(jwt.MapClaims)["email"].(string)
	if !ok {
		return fmt.Errorf("UpdateRouteShader(): unable to parse email from token")
	}
	requestEmailParts := strings.Split(requestEmail, "@")
	requestDomain := requestEmailParts[len(requestEmailParts)-1] // Domain is the last entry of the split since an email as only one @ sign
	buyer, err := s.Storage.BuyerWithDomain(requestDomain)
	if err != nil {
		return fmt.Errorf("UpdateRouteShader(): BuyerWithDomain error: %v", err)
	}

	acceptableLatency, err := strconv.Atoi(args.AcceptableLatency)
	if err != nil {
		return fmt.Errorf("UpdateRouteShader(): Failed to parse acceptable latency value: %v", err)
	}

	packetLossThreshold, err := strconv.Atoi(args.PLThreshold)
	if err != nil {
		return fmt.Errorf("UpdateRouteShader(): Failed to parse packet loss threshold value: %v", err)
	}

	buyer.CustomerRouteShader = routing.CustomerRouteShader{
		EnableNetworkNext:   args.EnableNN,
		EnableRoundTripTime: args.EnableRTT,
		EnablePacketLoss:    args.EnablePL,
		EnableMultiPath:     args.EnableMP,
		EnableABTest:        args.EnableAB,
		AcceptableLatency:   int64(acceptableLatency),
		PacketLossThreshold: int64(packetLossThreshold),
	}

	if err := s.Storage.SetBuyer(ctx, buyer); err != nil {
		return fmt.Errorf("UpdateRouteShader(): Failed to update buyer: %v", err)
	}

	reply.CustomerRouteShader = CustomerRouteShader{
		EnableNetworkNext:   buyer.CustomerRouteShader.EnableNetworkNext,
		EnableRoundTripTime: buyer.CustomerRouteShader.EnableRoundTripTime,
		EnableABTest:        buyer.CustomerRouteShader.EnableABTest,
		EnableMultiPath:     buyer.CustomerRouteShader.EnableMultiPath,
		EnablePacketLoss:    buyer.CustomerRouteShader.EnablePacketLoss,
		AcceptableLatency:   fmt.Sprintf("%v", buyer.CustomerRouteShader.AcceptableLatency),
		PacketLossThreshold: fmt.Sprintf("%v", buyer.CustomerRouteShader.PacketLossThreshold),
	}

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
