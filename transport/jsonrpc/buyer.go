package jsonrpc

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	fnv "hash/fnv"
	"net/http"
	"sort"
	"strconv"
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
	TopSessionsSize = 1000
)

type BuyersService struct {
	mu                    sync.Mutex
	mapPointsCache        json.RawMessage
	mapPointsCompactCache json.RawMessage

	RedisClient redis.Cmdable
	Storage     storage.Storer
	Logger      log.Logger
}

type FlushSessionsArgs struct{}

type FlushSessionsReply struct{}

func (s *BuyersService) FlushSessions(r *http.Request, args *FlushSessionsArgs, reply *FlushSessionsReply) error {
	if !CheckIsOps(r) {
		return ErrInsufficientPrivileges
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
	userhash := args.UserHash

	var sessionIDs []string

	var isAdmin bool = false
	var isSameBuyer bool = false
	var isAnon bool = true

	isAnon = IsAnonymous(r) || IsAnonymousPlus(r)

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

			if !isAnon {
				isAdmin, err = CheckRoles(r, "Admin")
				if err != nil {
					err = fmt.Errorf("UserSessions() CheckRoles error: %v", err)
					s.Logger.Log("err", err)
					return err
				}

				if !isAdmin {
					isSameBuyer, err = s.IsSameBuyer(r, meta.BuyerID)
					if err != nil {
						err = fmt.Errorf("UserSessions() IsSameBuyer error: %v", err)
						s.Logger.Log("err", err)
						return err
					}
				} else {
					isSameBuyer = true
				}
			}

			if isAnon || (!isSameBuyer && !isAdmin) {
				meta.Anonymise()
			}

			reply.Sessions = append(reply.Sessions, meta)
		}
	}

	sremcmds, err := sremtx.Exec()
	if err != nil && err != redis.Nil {
		err = fmt.Errorf("UserSessions() redit.Pipeliner error: %v", err)
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
	direct, err := s.RedisClient.SCard("total-direct").Result()
	if err != nil {
		return err
	}

	next, err := s.RedisClient.SCard("total-next").Result()
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
	var result []redis.Z

	var isAdmin bool = false
	var isSameBuyer bool = false
	var isAnon bool = true
	var isOps bool = false

	isAnon = IsAnonymous(r) || IsAnonymousPlus(r)
	isOps = CheckIsOps(r)

	if !isAnon && !isOps {
		isAdmin, err = CheckRoles(r, "Admin")
		if err != nil {
			err = fmt.Errorf("TopSessions() CheckRoles error: %v", err)
			s.Logger.Log("err", err)
			return err
		}
	}

	if !isAdmin && !isOps {
		isSameBuyer, err = s.IsSameBuyer(r, args.BuyerID)
		if err != nil {
			err = fmt.Errorf("TopSessions() IsSameBuyer error: %v", err)
			s.Logger.Log("err", err)
			return err
		}
	} else {
		isSameBuyer = true
	}

	reply.Sessions = make([]routing.SessionMeta, 0)

	// get the top session IDs globally or for a buyer from the sorted set
	switch args.BuyerID {
	case "":
		result, err = s.RedisClient.ZRangeWithScores("top-global", 0, TopSessionsSize*2).Result()
		if err != nil {
			err = fmt.Errorf("TopSessions() failed getting top global sessions: %v", err)
			s.Logger.Log("err", err)
			return err
		}
	default:
		if !isSameBuyer && !isAdmin && !isOps {
			err = fmt.Errorf("TopSessions() insufficient privileges")
			s.Logger.Log("err", err)
			return err
		}
		result, err = s.RedisClient.ZRangeWithScores(fmt.Sprintf("top-buyer-%s", args.BuyerID), 0, TopSessionsSize*2).Result()
		if err != nil {
			err = fmt.Errorf("TopSessions() failed getting top buyer sessions: %v", err)
			s.Logger.Log("err", err)
			return err
		}
	}

	// build a single transaction to get the sessions details for the session IDs
	var getCmds []*redis.StringCmd
	{
		gettx := s.RedisClient.TxPipeline()
		for _, member := range result {
			getCmds = append(getCmds, gettx.Get(fmt.Sprintf("session-%s-meta", member.Member.(string))))
		}
		_, err = gettx.Exec()
		if err != nil && err != redis.Nil {
			err = fmt.Errorf("TopSessions() failed getting top sessions meta: %v", err)
			s.Logger.Log("err", err)
			return err
		}
	}

	buyers := s.Storage.Buyers()

	// build a single transaction to remove any session ID from the sorted set if the
	// session-*-meta key is missing or expired
	zremtx := s.RedisClient.TxPipeline()
	{
		var meta routing.SessionMeta
		for _, cmd := range getCmds {
			// scan the data from Redis into its SessionMeta struct
			err = cmd.Scan(&meta)

			// if there was an error then the session-*-point key expired
			// so add ZREM commands to remove it from the key sets globally
			// and for each buyer
			if err != nil {
				args := cmd.Args()
				key := args[1].(string)
				keyparts := strings.Split(key, "-")

				zremtx.ZRem("top-global", keyparts[1])
				for _, buyer := range buyers {
					zremtx.ZRem(fmt.Sprintf("top-buyer-%016x", buyer.ID), keyparts[1])
				}
				continue
			}

			if isAnon || (!isSameBuyer && !isAdmin) {
				meta.Anonymise()
			}

			reply.Sessions = append(reply.Sessions, meta)
		}
	}

	// execute the transaction to remove the sessions IDs from the sorted key sets
	zremcmds, err := zremtx.Exec()
	if err != nil && err != redis.Nil {
		err = fmt.Errorf("TopSessions() redit.Pipeliner error: %v", err)
		s.Logger.Log("err", err)
		return err
	}

	level.Info(s.Logger).Log("key", "top-global", "key", "top-buyer-*", "removed", len(zremcmds))

	// sort the sessions
	sort.Slice(reply.Sessions, func(i int, j int) bool {
		// if comparing NN to Direct
		if reply.Sessions[i].OnNetworkNext && !reply.Sessions[j].OnNetworkNext {
			return true
		}
		// if comparing NN to NN
		if reply.Sessions[i].OnNetworkNext && reply.Sessions[j].OnNetworkNext {
			return reply.Sessions[i].DeltaRTT > reply.Sessions[j].DeltaRTT
		}
		// if comparing Direct to NN or Direct to Direct
		return false
	})

	// crop the list to TopSessionsSize if needed
	if len(reply.Sessions) > TopSessionsSize {
		reply.Sessions = reply.Sessions[:TopSessionsSize]
	}

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
	var isAdmin bool = false
	var isSameBuyer bool = false
	var isAnon bool = true
	var isOps bool = false

	isAnon = IsAnonymous(r) || IsAnonymousPlus(r)
	isOps = CheckIsOps(r)

	if !isAnon && !isOps {
		isAdmin, err = CheckRoles(r, "Admin")
		if err != nil {
			err = fmt.Errorf("SessionDetails() CheckRoles error: %v", err)
			s.Logger.Log("err", err)
			return err
		}
	}

	if !isAdmin && !isOps {
		isSameBuyer, err = s.IsSameBuyer(r, reply.Meta.BuyerID)
		if err != nil {
			err = fmt.Errorf("SessionDetails() IsSameBuyer error: %v", err)
			s.Logger.Log("err", err)
			return err
		}
	} else {
		isSameBuyer = true
	}

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

	// We fill in the name on the portal side so we don't spend time doing it while serving sessions
	if isAdmin {
		for idx, relay := range reply.Meta.NearbyRelays {
			r, err := s.Storage.Relay(relay.ID)
			if err != nil {
				continue
			}

			reply.Meta.NearbyRelays[idx].Name = r.Name
		}
	}

	if isAnon || (!isSameBuyer && !isAdmin && !isOps) {
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
func (s *BuyersService) GenerateMapPoints() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// get all the session IDs from the map-points-global key set
	sessionIDs, err := s.RedisClient.SMembers("map-points-global").Result()
	if err != nil {
		err = fmt.Errorf("SessionMapPoints() failed getting global map points: %v", err)
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

	// slice to hold all the final map points
	mappoints := make([]routing.SessionMapPoint, 0)
	mappointscompact := make([][]interface{}, 0)

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
				sremtx.SRem("map-points-global", keyparts[1])
				sremtx.SRem("total-next", keyparts[1])
				sremtx.SRem("total-direct", keyparts[1])
				continue
			}

			// if there was no error then add the SessionMapPoint to the slice
			mappoints = append(mappoints, point)

			var onNN uint
			if point.OnNetworkNext {
				onNN = 1
			}
			mappointscompact = append(mappointscompact, []interface{}{point.Longitude, point.Latitude, onNN})
		}
	}

	// execute the transaction to remove the sessions IDs from the key sets
	sremcmds, err := sremtx.Exec()
	if err != nil {
		return err
	}

	level.Info(s.Logger).Log("key", "map-points-global", "removed", len(sremcmds))

	// marshal the map points slice to local cache
	s.mapPointsCache, err = json.Marshal(mappoints)
	if err != nil {
		return err
	}

	s.mapPointsCompactCache, err = json.Marshal(mappointscompact)
	if err != nil {
		return err
	}

	return nil
}

// SessionMapPoints returns the locally cached JSON from GenerateSessionMapPoints
func (s *BuyersService) SessionMapPoints(r *http.Request, args *MapPointsArgs, reply *MapPointsReply) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// pull the local cache and reply with it
	reply.Points = s.mapPointsCache

	return nil
}

func (s *BuyersService) SessionMap(r *http.Request, args *MapPointsArgs, reply *MapPointsReply) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// pull the local cache and reply with it
	reply.Points = s.mapPointsCompactCache

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

	if IsAnonymous(r) || IsAnonymousPlus(r) {
		err = fmt.Errorf("GameConfiguration() insufficient privileges")
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
	if IsAnonymous(r) || IsAnonymousPlus(r) {
		return fmt.Errorf("UpdateGameConfiguration() insufficient privileges")
	}

	var err error
	var buyerID uint64
	var buyer routing.Buyer

	isAdmin, err := CheckRoles(r, "Admin")
	if err != nil {
		err = fmt.Errorf("UpdateGameConfiguration() CheckRoles error: %v", err)
		return err
	}

	isOwner, err := CheckRoles(r, "Owner")
	if err != nil {
		err = fmt.Errorf("UpdateGameConfiguration() CheckRoles error: %v", err)
		return err
	}

	if !isAdmin && !isOwner {
		err = fmt.Errorf("UpdateGameConfiguration() CheckRoles error: %v", ErrInsufficientPrivileges)
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
			Active:    false,
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
	if IsAnonymous(r) || IsAnonymousPlus(r) {
		return nil
	}

	for _, b := range s.Storage.Buyers() {
		id := strconv.FormatUint(b.ID, 10)
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

func (s *BuyersService) IsSameBuyer(r *http.Request, buyerID string) (bool, error) {
	if buyerID == "" {
		return false, nil
	}

	requestUser := r.Context().Value("user")
	if requestUser == nil {
		err := fmt.Errorf("Buyers() unable to parse user from token")
		s.Logger.Log("err", err)
		return false, err
	}

	requestEmail, ok := requestUser.(*jwt.Token).Claims.(jwt.MapClaims)["email"].(string)
	if !ok {
		err := fmt.Errorf("Buyers() unable to parse email from token")
		s.Logger.Log("err", err)
		return false, err
	}
	requestEmailParts := strings.Split(requestEmail, "@")
	requestDomain := requestEmailParts[len(requestEmailParts)-1] // Domain is the last entry of the split since an email as only one @ sign
	buyer, err := s.Storage.BuyerWithDomain(requestDomain)
	if err != nil {
		err = fmt.Errorf("Buyers() BuyerWithDomain error: %v", err)
		s.Logger.Log("err", err)
		return false, err
	}

	return buyerID != fmt.Sprintf("%016x", buyer.ID), nil
}
