package jsonrpc

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	fnv "hash/fnv"
	"net/http"
	"sort"
	"strings"
	"sync"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/go-redis/redis/v7"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
)

const (
	TopSessionsSize     = 1000
	TopNextSessionsSize = 1200
)

type BuyersService struct {
	mu                    sync.Mutex
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
	Sessions []routing.SessionMeta `json:"sessions"`
}

func (s *BuyersService) UserSessions(r *http.Request, args *UserSessionsArgs, reply *UserSessionsReply) error {
	var sessionIDs []string

	userhash := args.UserHash
	reply.Sessions = make([]routing.SessionMeta, 0)

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
		hashedID := fmt.Sprintf("%x", hash.Sum64())

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
		var meta routing.SessionMeta
		for _, cmd := range getCmds {
			err = cmd.Scan(&meta)
			if err != nil {
				args := cmd.Args()
				key := args[1].(string)
				keyparts := strings.Split(key, "-")

				sremtx.SRem(fmt.Sprintf("user-%s-sessions", userhash), keyparts[1])
				continue
			}

			if VerifyAnyRole(r, AnonymousRole, UnverifiedRole) || !VerifyAllRoles(r, s.SameBuyerRole(meta.BuyerID)) {
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
	direct, err := s.RedisClient.ZCard("total-direct").Result()
	if err != nil {
		return err
	}

	next, err := s.RedisClient.ZCard("total-next").Result()
	if err != nil {
		return err
	}

	reply.Direct = int(direct)
	reply.Next = int(next)

	return nil
}

type TopSessionsArgs struct {
	BuyerID string `json:"buyer_id"`
}

type TopSessionsReply struct {
	Sessions []routing.SessionMeta `json:"sessions"`
}

// TopSessions generates the top sessions sorted by improved RTT
func (s *BuyersService) TopSessions(r *http.Request, args *TopSessionsArgs, reply *TopSessionsReply) error {
	var err error
	var topnext []string
	var topdirect []string

	reply.Sessions = make([]routing.SessionMeta, 0)

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
	var nextSessions []routing.SessionMeta
	zremtx := s.RedisClient.TxPipeline()
	{
		var meta routing.SessionMeta
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

	var directSessions []routing.SessionMeta
	zremtx = s.RedisClient.TxPipeline()
	{
		var meta routing.SessionMeta
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
	directMap := make(map[string]*routing.SessionMeta)
	for i := range directSessions {
		directMap[directSessions[i].ID] = &directSessions[i]
	}
	for i := range nextSessions {
		delete(directMap, nextSessions[i].ID)
	}
	cleanDirectSessions := make([]routing.SessionMeta, 0)
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
	Meta   routing.SessionMeta    `json:"meta"`
	Slices []routing.SessionSlice `json:"slices"`
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

	if !VerifyAllRoles(r, s.SameBuyerRole(reply.Meta.BuyerID)) {
		reply.Meta.Anonymise()
	}

	reply.Slices = make([]routing.SessionSlice, 0)

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

// GenerateMapPoints warms a local cache of JSON to be used by SessionMapPoints
func (s *BuyersService) GenerateMapPointsPerBuyer() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var sessionIDs []string
	var err error

	buyers := s.Storage.Buyers()

	// slice to hold all the final map points
	mapPointsBuyers := make(map[string][]routing.SessionMapPoint, 0)
	mapPointsBuyerscompact := make(map[string][][]interface{}, 0)

	mapPointsGlobal := make([]routing.SessionMapPoint, 0)
	mapPointsGlobalcompact := make([][]interface{}, 0)

	s.mapPointsBuyerCache = make(map[string]json.RawMessage, 0)
	s.mapPointsCompactBuyerCache = make(map[string]json.RawMessage, 0)

	for _, buyer := range buyers { // get all the session IDs from the map-points-global key set
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
		// removed from map-points-global, total-next, and total-direct key sets
		sremtx := s.RedisClient.TxPipeline()
		{
			var point routing.SessionMapPoint
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
					sremtx.ZRem("total-direct", keyparts[1])
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
					mapPointsBuyerscompact[stringID] = append(mapPointsBuyerscompact[stringID], []interface{}{point.Longitude, point.Latitude, onNN})
					mapPointsGlobalcompact = append(mapPointsGlobalcompact, []interface{}{point.Longitude, point.Latitude, onNN})
				}
			}
		}

		// execute the transaction to remove the sessions IDs from the key sets
		sremcmds, err := sremtx.Exec()
		if err != nil {
			return err
		}

		level.Info(s.Logger).Log("key", "map-points-global", "removed", len(sremcmds))

		// marshal the map points slice to local cache
		s.mapPointsBuyerCache[stringID], err = json.Marshal(mapPointsBuyers[stringID])
		if err != nil {
			return err
		}

		s.mapPointsCompactBuyerCache[stringID], err = json.Marshal(mapPointsBuyerscompact[stringID])
		if err != nil {
			return err
		}
	}

	// marshal the map points slice to local cache
	s.mapPointsCache, err = json.Marshal(mapPointsGlobal)
	if err != nil {
		return err
	}

	s.mapPointsCompactCache, err = json.Marshal(mapPointsGlobalcompact)
	if err != nil {
		return err
	}

	return nil
}

// SessionMapPoints returns the locally cached JSON from GenerateSessionMapPoints
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

	// Buyer not found
	if buyer.ID == 0 {
		byteKey := []byte(args.NewPublicKey)

		buyerID = binary.LittleEndian.Uint64(byteKey[0:8])
		err := s.Storage.AddBuyer(ctx, routing.Buyer{
			ID:        buyerID,
			Name:      args.Name,
			Domain:    args.Domain,
			Active:    true,
			Live:      false,
			PublicKey: byteKey,
		})

		if err != nil {
			err = fmt.Errorf("UpdateGameConfiguration() failed to add buyer")
			s.Logger.Log("err", err)
			return err
		}

		if buyer, err = s.Storage.Buyer(buyerID); err != nil {
			return nil
		}
	}

	if err = buyer.DecodedPublicKey(args.NewPublicKey); err != nil {
		err = fmt.Errorf("UpdateGameConfiguration() failed to decode public key")
		s.Logger.Log("err", err)
		return err
	}

	if err = s.Storage.SetBuyer(ctx, buyer); err != nil {
		err = fmt.Errorf("UpdateGameConfiguration() failed to update buyer public key")
		s.Logger.Log("err", err)
		return err
	}

	reply.GameConfiguration.PublicKey = buyer.EncodedPublicKey()
	reply.GameConfiguration.Company = buyer.Name

	return nil
}

type BuyerListArgs struct{}

type BuyerListReply struct {
	Buyers []buyerAccount
}

type buyerAccount struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (s *BuyersService) Buyers(r *http.Request, args *BuyerListArgs, reply *BuyerListReply) error {
	reply.Buyers = make([]buyerAccount, 0)
	if VerifyAllRoles(r, AnonymousRole) {
		return nil
	}

	for _, b := range s.Storage.Buyers() {
		id := fmt.Sprintf("%016x", b.ID)
		account := buyerAccount{
			ID:   id,
			Name: b.Name,
		}
		reply.Buyers = append(reply.Buyers, account)
	}

	sort.Slice(reply.Buyers, func(i int, j int) bool {
		return reply.Buyers[i].Name < reply.Buyers[j].Name
	})

	return nil
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

		return buyerID != fmt.Sprintf("%016x", buyer.ID), nil
	}
}
