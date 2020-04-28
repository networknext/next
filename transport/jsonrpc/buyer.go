package jsonrpc

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
	"github.com/networknext/backend/transport"
)

type BuyersService struct {
	RedisClient redis.Cmdable
	Storage     storage.Storer
}

type MapArgs struct {
	BuyerID string `json:"buyer_id"`
}

type MapReply struct {
	SessionPoints []point `json:"sess_points"`
}

type point struct {
	Coordinates   []float64 `json:"COORDINATES"`
	OnNetworkNext bool      `json:"on_network_next"`
}

func (s *BuyersService) SessionsMap(r *http.Request, args *MapArgs, reply *MapReply) error {
	var err error
	var cacheEntry transport.SessionCacheEntry
	var cacheEntryData []byte
	var buyerID uint64

	if args.BuyerID == "" {
		return fmt.Errorf("buyer_id is required")
	}

	if buyerID, err = strconv.ParseUint(args.BuyerID, 10, 64); err != nil {
		return fmt.Errorf("failed to convert BuyerID to uint64")
	}

	var getCmds []*redis.StringCmd
	{
		iter := s.RedisClient.Scan(0, fmt.Sprintf("SESSION-%d-*", buyerID), 10000).Iterator()
		tx := s.RedisClient.TxPipeline()
		for iter.Next() {
			getCmds = append(getCmds, tx.Get(iter.Val()))
		}
		if err := iter.Err(); err != nil {
			return fmt.Errorf("failed to scan redis: %w", err)
		}
		_, err = tx.Exec()
		if err != nil {
			return fmt.Errorf("failed to multi-get redis: %w", err)
		}
	}

	for _, cmd := range getCmds {
		cacheEntryData, err = cmd.Bytes()
		if err != nil {
			continue
		}

		if err := cacheEntry.UnmarshalBinary(cacheEntryData); err != nil {
			continue
		}

		lat := cacheEntry.Location.Latitude
		lng := cacheEntry.Location.Longitude
		if lat == 0 && lng == 0 {
			lat = (-123) + rand.Float64()*((-70)-(-123))
			lng = 28 + rand.Float64()*(48-28)
		}

		reply.SessionPoints = append(reply.SessionPoints, point{
			Coordinates:   []float64{lat, lng},
			OnNetworkNext: cacheEntry.RouteDecision.OnNetworkNext,
		})
	}

	return nil
}

type SessionsArgs struct {
	BuyerID   string `json:"buyer_id"`
	SessionID string `json:"session_id"`
}

type SessionsReply struct {
	Sessions []session `json:"sessions"`
}

type session struct {
	SessionID     string    `json:"session_id"`
	UserHash      string    `json:"user_hash"`
	DirectRTT     float64   `json:"direct_rtt"`
	NextRTT       float64   `json:"next_rtt"`
	ChangeRTT     float64   `json:"change_rtt"`
	OnNetworkNext bool      `json:"on_network_next"`
	ExpiresAt     time.Time `json:"expires_at"`
}

func (s *BuyersService) Sessions(r *http.Request, args *SessionsArgs, reply *SessionsReply) error {
	var err error
	var cacheEntry transport.SessionCacheEntry
	var cacheEntryData []byte
	var buyerID uint64
	var sessionID uint64

	if args.BuyerID == "" {
		return fmt.Errorf("buyer_id is required")
	}

	if buyerID, err = strconv.ParseUint(args.BuyerID, 10, 64); err != nil {
		return fmt.Errorf("failed to convert BuyerID to uint64")
	}

	if args.SessionID != "" {
		if sessionID, err = strconv.ParseUint(args.SessionID, 10, 64); err != nil {
			return fmt.Errorf("failed to convert SessionID to uint64")
		}

		getCmd := s.RedisClient.Get(fmt.Sprintf("SESSION-%d-%d", buyerID, sessionID))

		if cacheEntryData, err = getCmd.Bytes(); err != nil {
			return fmt.Errorf("failed to get session %d: %w", sessionID, err)
		}

		if err := cacheEntry.UnmarshalBinary(cacheEntryData); err != nil {
			return fmt.Errorf("failed to unmarshal session %d: %w", sessionID, err)
		}

		reply.Sessions = append(reply.Sessions, session{
			SessionID:     strconv.FormatUint(cacheEntry.SessionID, 10),
			UserHash:      strconv.FormatUint(cacheEntry.UserHash, 10),
			DirectRTT:     cacheEntry.DirectRTT,
			NextRTT:       cacheEntry.NextRTT,
			ChangeRTT:     cacheEntry.NextRTT - cacheEntry.DirectRTT,
			OnNetworkNext: cacheEntry.RouteDecision.OnNetworkNext,
			ExpiresAt:     cacheEntry.TimestampExpire,
		})

		return nil
	}

	var getCmds []*redis.StringCmd
	{
		iter := s.RedisClient.Scan(0, fmt.Sprintf("SESSION-%d-*", buyerID), 10000).Iterator()
		tx := s.RedisClient.TxPipeline()
		for iter.Next() {
			getCmds = append(getCmds, tx.Get(iter.Val()))
		}
		if err := iter.Err(); err != nil {
			return fmt.Errorf("failed to scan redis: %w", err)
		}
		_, err = tx.Exec()
		if err != nil {
			return fmt.Errorf("failed to multi-get redis: %w", err)
		}
	}

	for _, cmd := range getCmds {
		cacheEntryData, err = cmd.Bytes()
		if err != nil {
			continue
		}

		if err := cacheEntry.UnmarshalBinary(cacheEntryData); err != nil {
			continue
		}

		reply.Sessions = append(reply.Sessions, session{
			SessionID:     strconv.FormatUint(cacheEntry.SessionID, 10),
			UserHash:      strconv.FormatUint(cacheEntry.UserHash, 10),
			DirectRTT:     cacheEntry.DirectRTT,
			NextRTT:       cacheEntry.NextRTT,
			ChangeRTT:     cacheEntry.NextRTT - cacheEntry.DirectRTT,
			OnNetworkNext: cacheEntry.RouteDecision.OnNetworkNext,
			ExpiresAt:     cacheEntry.TimestampExpire,
		})
	}

	sort.Slice(reply.Sessions, func(i int, j int) bool {
		return reply.Sessions[i].ChangeRTT < reply.Sessions[j].ChangeRTT
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
