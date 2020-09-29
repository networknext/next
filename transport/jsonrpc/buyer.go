package jsonrpc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gomodule/redigo/redis"
	"github.com/networknext/backend/admin"
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
	if err := admin.FlushSessions(r, s.RedisPoolTopSessions, s.RedisPoolSessionMeta, s.RedisPoolSessionSlices, s.RedisPoolSessionMap); err != nil {
		err = fmt.Errorf("FlushSessions() failed to flush sessions: %v", err)
		level.Error(s.Logger).Log("err", err)
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
	// 	level.Error(s.Logger).Log("err", err)
	// 	return err
	// }

	// if len(sessionIDs) == 0 {
	// 	hash := fnv.New64a()
	// 	_, err := hash.Write([]byte(userhash))
	// 	if err != nil {
	// 		err = fmt.Errorf("UserSessions() error writing 64a hash: %v", err)
	// 		level.Error(s.Logger).Log("err", err)
	// 		return err
	// 	}
	// 	hashedID := fmt.Sprintf("%016x", hash.Sum64())

	// 	err = s.RedisClient.SMembers(fmt.Sprintf("user-%s-sessions", hashedID)).ScanSlice(&sessionIDs)
	// 	if err != nil {
	// 		err = fmt.Errorf("UserSessions() failed getting user sessions: %v", err)
	// 		level.Error(s.Logger).Log("err", err)
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
	// 		level.Error(s.Logger).Log("err", err)
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

	// 		if admin.admin.VerifyAnyRole(r, admin.AnonymousRole, admin.UnverifiedRole) || !admin.VerifyAllRoles(r, admin.VerifyAnyRole(fmt.Sprintf("%016x", meta.BuyerID))) {
	// 			meta.Anonymise()
	// 		}

	// 		reply.Sessions = append(reply.Sessions, meta)
	// 	}
	// }

	// sremcmds, err := sremtx.Exec()
	// if err != nil && err != redis.Nil {
	// 	err = fmt.Errorf("UserSessions() redis.Pipeliner error: %v", err)
	// 	level.Error(s.Logger).Log("err", err)
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
	CompanyCode string `json:"company_code"`
}

type TotalSessionsReply struct {
	Direct int `json:"direct"`
	Next   int `json:"next"`
}

func (s *BuyersService) TotalSessions(r *http.Request, args *TotalSessionsArgs, reply *TotalSessionsReply) error {
	buyer, _ := s.Storage.BuyerWithCompanyCode(args.CompanyCode)
	output, err := admin.TotalSessions(r, s.RedisPoolSessionMap, buyer, s.Storage.Buyers())
	if err != nil {
		err = fmt.Errorf("TotalSessions() failed to gather total sessions: %v", err)
		level.Error(s.Logger).Log("err", err)
		return err
	}
	reply.Next = output.Next
	reply.Direct = output.Direct
	return nil
}

type TopSessionsArgs struct {
	CompanyCode string `json:"company_code"`
}

type TopSessionsReply struct {
	Sessions []transport.SessionMeta `json:"sessions"`
}

// TopSessions generates the top sessions sorted by improved RTT
func (s *BuyersService) TopSessions(r *http.Request, args *TopSessionsArgs, reply *TopSessionsReply) error {
	buyer, _ := s.Storage.BuyerWithCompanyCode(args.CompanyCode)
	sessions, err := admin.TopSessions(r, s.RedisPoolTopSessions, s.RedisPoolSessionMeta, buyer)
	if err != nil {
		err = fmt.Errorf("TopSessions() failed to gather top sessions: %v", err)
		level.Error(s.Logger).Log("err", err)
		return err
	}
	reply.Sessions = sessions
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
	output, err := admin.SessionDetails(r, s.Storage, s.RedisPoolSessionMeta, s.RedisPoolSessionSlices, args.SessionID)
	if err != nil {
		err = fmt.Errorf("TopSessions() failed to gather top sessions: %v", err)
		level.Error(s.Logger).Log("err", err)
		return err
	}
	reply.Meta = output.Meta
	reply.Slices = output.Slices
	return nil
}

type MapPointsArgs struct {
	CompanyCode string `json:"company_code"`
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

	caches, err := admin.GenerateMapPointsPerBuyer(s.RedisPoolSessionMap, s.Storage.Buyers())
	if err != nil {
		err = fmt.Errorf("GenerateMapPointsPerBuyer() failed to generate map points: %v", err)
		level.Error(s.Logger).Log("err", err)
		return err
	}
	s.mapPointsBuyerCache = caches.MapPointsBuyerCache
	s.mapPointsCompactBuyerCache = caches.MapPointsCompactBuyerCache
	s.mapPointsCache = caches.MapPointsCache
	s.mapPointsCompactCache = caches.MapPointsCompactCache
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
			level.Error(s.Logger).Log("err", err)
			return err
		}

		var currentPoint transport.SessionMapPoint
		for _, directPointString := range directPointStrings {
			directSplitStrings := strings.Split(directPointString, "|")
			if err := currentPoint.ParseRedisString(directSplitStrings); err != nil {
				err = fmt.Errorf("SessionMapPoints() failed to parse direct map point for buyer %s: %v", stringID, err)
				level.Error(s.Logger).Log("err", err)
				return err
			}

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
			nextSplitStrings := strings.Split(nextPointString, "|")
			if err := currentPoint.ParseRedisString(nextSplitStrings); err != nil {
				err = fmt.Errorf("SessionMapPoints() failed to next parse map point for buyer %s: %v", stringID, err)
				level.Error(s.Logger).Log("err", err)
				return err
			}

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
	switch args.CompanyCode {
	case "":
		// pull the local cache and reply with it
		reply.Points = s.mapPointsCompactCache
	default:
		if !admin.VerifyAllRoles(r, admin.SameBuyerRole(args.CompanyCode)) {
			err := fmt.Errorf("SessionMap(): %v", admin.ErrInsufficientPrivileges)
			level.Error(s.Logger).Log("err", err)
			return err
		}
		reply.Points = s.mapPointsCompactBuyerCache[args.CompanyCode]
	}

	return nil
}

func (s *BuyersService) SessionMapPoints(r *http.Request, args *MapPointsArgs, reply *MapPointsReply) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch args.CompanyCode {
	case "":
		// pull the local cache and reply with it
		reply.Points = s.mapPointsCache
	default:
		if !admin.VerifyAllRoles(r, admin.SameBuyerRole(args.CompanyCode)) {
			err := fmt.Errorf("SessionMap(): %v", admin.ErrInsufficientPrivileges)
			level.Error(s.Logger).Log("err", err)
			return err
		}
		reply.Points = s.mapPointsBuyerCache[args.CompanyCode]
	}

	return nil
}

// SessionMapPoints returns the locally cached JSON from GenerateSessionMapPoints
func (s *BuyersService) SessionMapPointsByte(r *http.Request, args *MapPointsArgs, reply *MapPointsByteReply) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch args.CompanyCode {
	case "":
		// pull the local cache and reply with it
		ReadMapPointsCache(&reply.Points, s.mapPointsCache)
	default:
		if !admin.VerifyAllRoles(r, admin.SameBuyerRole(args.CompanyCode)) {
			err := fmt.Errorf("SessionMap(): %v", admin.ErrInsufficientPrivileges)
			level.Error(s.Logger).Log("err", err)
			return err
		}
		ReadMapPointsCache(&reply.Points, s.mapPointsBuyerCache[args.CompanyCode])
	}

	return nil
}

func (s *BuyersService) SessionMapByte(r *http.Request, args *MapPointsArgs, reply *MapPointsByteReply) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch args.CompanyCode {
	case "":
		// pull the local cache and reply with it
		ReadMapPointsCache(&reply.Points, s.mapPointsByteCache)
	default:
		if !admin.VerifyAllRoles(r, admin.SameBuyerRole(args.CompanyCode)) {
			err := fmt.Errorf("SessionMap(): %v", admin.ErrInsufficientPrivileges)
			level.Error(s.Logger).Log("err", err)
			return err
		}
		ReadMapPointsCache(&reply.Points, s.mapPointsBuyerByteCache[args.CompanyCode])
	}

	return nil
}

type GameConfigurationArgs struct {
	NewPublicKey string `json:"new_public_key"`
}

type GameConfigurationReply struct {
	GameConfiguration gameConfiguration `json:"game_config"`
}

type gameConfiguration struct {
	PublicKey string `json:"public_key"`
}

func (s *BuyersService) GameConfiguration(r *http.Request, args *GameConfigurationArgs, reply *GameConfigurationReply) error {
	var err error
	var buyer routing.Buyer

	companyCode, err := admin.RequestCompany(r)
	if err != nil {
		err := fmt.Errorf("GameConfiguration(): user is not assigned to a company")
		level.Error(s.Logger).Log("err", err)
		return err
	}

	if companyCode == "" {
		err = fmt.Errorf("GameConfiguration(): failed to parse company code")
		level.Error(s.Logger).Log("err", err)
		return err
	}

	if admin.VerifyAnyRole(r, admin.AnonymousRole, admin.UnverifiedRole) {
		err = fmt.Errorf("GameConfiguration(): %v", admin.ErrInsufficientPrivileges)
		level.Error(s.Logger).Log("err", err)
		return err
	}

	reply.GameConfiguration.PublicKey = ""

	buyer, err = s.Storage.BuyerWithCompanyCode(companyCode)
	// Buyer not found
	if err != nil {
		return nil
	}

	reply.GameConfiguration.PublicKey = buyer.EncodedPublicKey()

	return nil
}

func (s *BuyersService) UpdateGameConfiguration(r *http.Request, args *GameConfigurationArgs, reply *GameConfigurationReply) error {
	if args.NewPublicKey == "" {
		err := fmt.Errorf("UpdateGameConfiguration() new public key is required")
		level.Error(s.Logger).Log("err", err)
		return err
	}

	key, err := admin.UpdateGameConfiguration(r, s.Storage, args.NewPublicKey)
	if err != nil {
		err := fmt.Errorf("UpdateGameConfiguration() new public key is required")
		level.Error(s.Logger).Log("err", err)
		return err
	}

	reply.GameConfiguration.PublicKey = key
	return nil
}

type BuyerListArgs struct{}

type BuyerListReply struct {
	Buyers []buyerAccount `json:"buyers"`
}

type buyerAccount struct {
	CompanyName string `json:"company_name"`
	CompanyCode string `json:"company_code"`
	ID          string `json:"id"`
	IsLive      bool   `json:"is_live"`
}

func (s *BuyersService) Buyers(r *http.Request, args *BuyerListArgs, reply *BuyerListReply) error {
	reply.Buyers = make([]buyerAccount, 0)
	if admin.VerifyAllRoles(r, admin.AnonymousRole) {
		return nil
	}

	for _, b := range s.Storage.Buyers() {
		id := fmt.Sprintf("%016x", b.ID)
		customer, err := s.Storage.Customer(b.CompanyCode)
		if err != nil {
			continue
		}
		account := buyerAccount{
			CompanyName: customer.Name,
			CompanyCode: b.CompanyCode,
			ID:          id,
			IsLive:      b.Live,
		}
		if admin.VerifyAllRoles(r, admin.SameBuyerRole(b.CompanyCode)) {
			reply.Buyers = append(reply.Buyers, account)
		}
	}

	sort.Slice(reply.Buyers, func(i int, j int) bool {
		return reply.Buyers[i].CompanyName < reply.Buyers[j].CompanyName
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
	if admin.VerifyAllRoles(r, admin.AnonymousRole) {
		return nil
	}

	var dcm map[uint64]routing.DatacenterMap

	dcm = s.Storage.GetDatacenterMapsForBuyer(args.ID)

	var replySlice []DatacenterMapsFull
	for _, dcMap := range dcm {
		buyer, err := s.Storage.Buyer(dcMap.BuyerID)
		if err != nil {
			err = fmt.Errorf("DatacenterMapsForBuyer() could not parse buyer")
			level.Error(s.Logger).Log("err", err)
			return err
		}
		datacenter, err := s.Storage.Datacenter(dcMap.Datacenter)
		if err != nil {
			err = fmt.Errorf("DatacenterMapsForBuyer() could not parse datacenter")
			level.Error(s.Logger).Log("err", err)
			return err
		}

		customer, err := s.Storage.Customer(buyer.CompanyCode)
		if err != nil {
			continue
		}

		dcmFull := DatacenterMapsFull{
			Alias:          dcMap.Alias,
			DatacenterName: datacenter.Name,
			DatacenterID:   fmt.Sprintf("%016x", dcMap.Datacenter),
			BuyerName:      customer.Name,
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
