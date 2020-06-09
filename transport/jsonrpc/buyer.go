package jsonrpc

import (
	"context"
	"encoding/binary"
	"fmt"
	fnv "hash/fnv"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-redis/redis/v7"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
)

type BuyersService struct {
	RedisClient redis.Cmdable
	Storage     storage.Storer
}

type UserSessionsArgs struct {
	UserHash string `json:"user_hash"`
}

type UserSessionsReply struct {
	Sessions []routing.SessionMeta `json:"sessions"`
}

func (s *BuyersService) UserSessions(r *http.Request, args *UserSessionsArgs, reply *UserSessionsReply) error {
	var sessionIDs []string

	var isAdmin bool = false
	var isSameBuyer bool = false
	var isAnon bool = true

	isAnon = IsAnonymous(r)

	reply.Sessions = make([]routing.SessionMeta, 0)

	err := s.RedisClient.SMembers(fmt.Sprintf("user-%s-sessions", args.UserHash)).ScanSlice(&sessionIDs)
	if err != nil {
		return err
	}

	if len(sessionIDs) == 0 {
		hash := fnv.New64a()
		_, err := hash.Write([]byte(args.UserHash))
		if err != nil {
			return err
		}
		hashedID := fmt.Sprintf("%x", hash.Sum64())

		err = s.RedisClient.SMembers(fmt.Sprintf("user-%s-sessions", hashedID)).ScanSlice(&sessionIDs)
		if err != nil {
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
			return err
		}
	}

	exptx := s.RedisClient.TxPipeline()
	{
		var meta routing.SessionMeta
		for _, cmd := range getCmds {
			err = cmd.Scan(&meta)
			if err != nil {
				// If we get here then there is a session in the top lists
				// but has missing meta information so we set an empty meta
				// key to expire in 5 seconds for the main cleanup routine
				// to clear the reference in the top lists
				args := cmd.Args()
				exptx.Set(args[1].(string), "", 5*time.Second)
				continue
			}

			if !isAnon {
				isAdmin, err = CheckRoles(r, "Admin")
				if err != nil {
					return err
				}

				if !isAdmin {
					isSameBuyer, err = s.IsSameBuyer(r, meta.CustomerID)
					if err != nil {
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
	_, err = exptx.Exec()
	if err != nil && err != redis.Nil {
		return err
	}

	sort.Slice(reply.Sessions, func(i int, j int) bool {
		return reply.Sessions[i].ID < reply.Sessions[j].ID
	})

	return nil
}

type TopSessionsArgs struct {
	BuyerID string `json:"buyer_id"`
}

type TopSessionsReply struct {
	Sessions []routing.SessionMeta `json:"sessions"`
}

func (s *BuyersService) TopSessions(r *http.Request, args *TopSessionsArgs, reply *TopSessionsReply) error {
	var err error
	var result []redis.Z

	var isAdmin bool = false
	var isSameBuyer bool = false
	var isAnon bool = true
	var isOps bool = false

	isAnon = IsAnonymous(r)
	isOps = CheckIsOps(r)

	if !isAnon && !isOps {
		isAdmin, err = CheckRoles(r, "Admin")
		if err != nil {
			return err
		}
	}

	if !isAdmin && !isOps {
		isSameBuyer, err = s.IsSameBuyer(r, args.BuyerID)
		if err != nil {
			return err
		}
	} else {
		isSameBuyer = true
	}

	reply.Sessions = make([]routing.SessionMeta, 0)

	switch args.BuyerID {
	case "":
		result, err = s.RedisClient.ZRangeWithScores("top-global", 0, 1000).Result()
		if err != nil {
			return err
		}
	default:
		if !isSameBuyer && !isAdmin && !isOps {
			return fmt.Errorf("insufficient privileges")
		}
		result, err = s.RedisClient.ZRangeWithScores(fmt.Sprintf("top-buyer-%s", args.BuyerID), 0, 1000).Result()
		if err != nil {
			return err
		}
	}

	var getCmds []*redis.StringCmd
	{
		gettx := s.RedisClient.TxPipeline()
		for _, member := range result {
			getCmds = append(getCmds, gettx.Get(fmt.Sprintf("session-%s-meta", member.Member.(string))))
		}
		_, err = gettx.Exec()
		if err != nil && err != redis.Nil {
			return err
		}
	}

	exptx := s.RedisClient.TxPipeline()
	{
		var meta routing.SessionMeta
		for _, cmd := range getCmds {
			err = cmd.Scan(&meta)
			if err != nil {
				// If we get here then there is a session in the top lists
				// but has missing meta information so we set an empty meta
				// key to expire in 5 seconds for the main cleanup routine
				// to clear the reference in the top lists
				args := cmd.Args()
				exptx.Set(args[1].(string), "", 5*time.Second)
				continue
			}

			if isAnon || (!isSameBuyer && !isAdmin) {
				meta.Anonymise()
			}

			reply.Sessions = append(reply.Sessions, meta)
		}
	}
	_, err = exptx.Exec()
	if err != nil && err != redis.Nil {
		return err
	}

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

	isAnon = IsAnonymous(r)
	isOps = CheckIsOps(r)

	if !isAnon && !isOps {
		isAdmin, err = CheckRoles(r, "Admin")
		if err != nil {
			return err
		}
	}

	if !isAdmin && !isOps {
		isSameBuyer, err = s.IsSameBuyer(r, reply.Meta.CustomerID)
		if err != nil {
			return err
		}
	} else {
		isSameBuyer = true
	}

	data, err := s.RedisClient.Get(fmt.Sprintf("session-%s-meta", args.SessionID)).Bytes()
	if err != nil {
		return err
	}
	err = reply.Meta.UnmarshalBinary(data)
	if err != nil {
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
	Points []routing.SessionMapPoint `json:"map_points"`
}

func (s *BuyersService) SessionMapPoints(r *http.Request, args *MapPointsArgs, reply *MapPointsReply) error {
	var err error
	var sessionIDs []string

	var isAdmin bool = false
	var isSameBuyer bool = false
	var isAnon bool = true

	isAnon = IsAnonymous(r)
	if !isAnon {
		isAdmin, err = CheckRoles(r, "Admin")
		if err != nil {
			return err
		}
	}

	if !isAdmin {
		isSameBuyer, err = s.IsSameBuyer(r, args.BuyerID)
		if err != nil {
			return err
		}
	} else {
		isSameBuyer = true
	}

	reply.Points = make([]routing.SessionMapPoint, 0)

	switch args.BuyerID {
	case "":
		sessionIDs, err = s.RedisClient.SMembers("map-points-global").Result()
		if err != nil {
			return err
		}
	default:
		if !isSameBuyer && !isAdmin {
			return fmt.Errorf("insufficient privileges")
		}
		sessionIDs, err = s.RedisClient.SMembers(fmt.Sprintf("map-points-buyer-%s", args.BuyerID)).Result()
		if err != nil {
			return err
		}
	}

	var getCmds []*redis.StringCmd
	{
		gettx := s.RedisClient.TxPipeline()
		for _, sessionID := range sessionIDs {
			getCmds = append(getCmds, gettx.Get(fmt.Sprintf("session-%s-point", sessionID)))
		}
		_, err = gettx.Exec()
		if err != nil && err != redis.Nil {
			return err
		}
	}

	exptx := s.RedisClient.TxPipeline()
	{
		var point routing.SessionMapPoint
		for _, cmd := range getCmds {
			err = cmd.Scan(&point)
			if err != nil {
				// If we get here then there is a point in the point map (map as in key -> val)
				// but the actual point data is missing so we set an empty meta
				// key to expire in 5 seconds for the main cleanup routine
				// to clear the reference in the point map
				args := cmd.Args()
				exptx.Set(args[1].(string), "", 5*time.Second)
				continue
			}

			reply.Points = append(reply.Points, point)
		}
	}
	_, err = exptx.Exec()
	if err != nil && err != redis.Nil {
		return err
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

	if IsAnonymous(r) {
		return fmt.Errorf("insufficient privileges")
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
	if IsAnonymous(r) {
		return fmt.Errorf("insufficient privileges")
	}
	var err error
	var buyerID uint64
	var buyer routing.Buyer

	ctx := context.Background()

	if args.Domain == "" {
		return fmt.Errorf("domain is required")
	}

	if args.Name == "" {
		return fmt.Errorf("company name is required")
	}

	if args.NewPublicKey == "" {
		return fmt.Errorf("new public key is required")
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
			return fmt.Errorf("failed to add buyer")
		}

		if buyer, err = s.Storage.Buyer(buyerID); err != nil {
			return nil
		}
	}

	if err = buyer.DecodedPublicKey(args.NewPublicKey); err != nil {
		return fmt.Errorf("failed to decode public key")
	}

	if err = s.Storage.SetBuyer(ctx, buyer); err != nil {
		return fmt.Errorf("failed to update buyer public key")
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
	if IsAnonymous(r) {
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
		return false, fmt.Errorf("unable to parse user from token")
	}

	requestEmail, ok := requestUser.(*jwt.Token).Claims.(jwt.MapClaims)["email"].(string)
	if !ok {
		return false, fmt.Errorf("unable to parse email from token")
	}
	requestEmailParts := strings.Split(requestEmail, "@")
	requestDomain := requestEmailParts[len(requestEmailParts)-1] // Domain is the last entry of the split since an email as only one @ sign
	buyer, err := s.Storage.BuyerWithDomain(requestDomain)
	if err != nil {
		return false, err
	}

	return buyerID != strconv.FormatUint(buyer.ID, 10), nil
}
