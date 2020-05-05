package jsonrpc

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

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
	err := s.RedisClient.SMembers(fmt.Sprintf("user-%s-sessions", args.UserHash)).ScanSlice(&sessionIDs)
	if err != nil {
		return err
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

	switch args.BuyerID {
	case "":
		result, err = s.RedisClient.ZRangeWithScores("top-global", 0, 1000).Result()
		if err != nil {
			return err
		}
	default:
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

			reply.Sessions = append(reply.Sessions, meta)
		}
	}
	_, err = exptx.Exec()
	if err != nil && err != redis.Nil {
		return err
	}

	sort.Slice(reply.Sessions, func(i int, j int) bool {
		return reply.Sessions[i].DeltaRTT < reply.Sessions[j].DeltaRTT
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
	data, err := s.RedisClient.Get(fmt.Sprintf("session-%s-meta", args.SessionID)).Bytes()
	if err != nil {
		return err
	}
	err = reply.Meta.UnmarshalBinary(data)
	if err != nil {
		return err
	}

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

	switch args.BuyerID {
	case "":
		sessionIDs, err = s.RedisClient.SMembers("map-points-global").Result()
		if err != nil {
			return err
		}
	default:
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
	BuyerID      string `json:"buyer_id"`
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
	var buyerID uint64
	var buyer routing.Buyer

	if args.BuyerID == "" {
		return fmt.Errorf("buyer_id is required")
	}

	if buyerID, err = strconv.ParseUint(args.BuyerID, 10, 64); err != nil {
		return fmt.Errorf("failed to convert BuyerID to uint64")
	}

	if buyer, err = s.Storage.Buyer(buyerID); err != nil {
		return fmt.Errorf("failed to fetch buyer info from Storer")
	}

	reply.GameConfiguration.PublicKey = buyer.EncodedPublicKey()

	return nil
}

func (s *BuyersService) UpdateGameConfiguration(r *http.Request, args *GameConfigurationArgs, reply *GameConfigurationReply) error {
	var err error
	var buyerID uint64
	var buyer routing.Buyer

	ctx := context.Background()

	if args.BuyerID == "" {
		return fmt.Errorf("buyer_id is required")
	}

	if buyerID, err = strconv.ParseUint(args.BuyerID, 10, 64); err != nil {
		return fmt.Errorf("failed to convert BuyerID to uint64")
	}

	if buyer, err = s.Storage.Buyer(buyerID); err != nil {
		return fmt.Errorf("failed to fetch buyer info from Storer")
	}

	if err = buyer.DecodedPublicKey(args.NewPublicKey); err != nil {
		return fmt.Errorf("failed to decode public key")
	}

	if err = s.Storage.SetBuyer(ctx, buyer); err != nil {
		return fmt.Errorf("failed to update buyer public key")
	}

	reply.GameConfiguration.PublicKey = buyer.EncodedPublicKey()

	return nil
}
