package jsonrpc

import (
	"fmt"
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
	reply.SessionPoints = []point{
		{Coordinates: []float64{-73.6, 42.7}, OnNetworkNext: true},
		{Coordinates: []float64{-73.7, 42.6}, OnNetworkNext: false},
		{Coordinates: []float64{-73.8, 43.0}, OnNetworkNext: true},
		{Coordinates: []float64{-73.8, 43.0}, OnNetworkNext: false},
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
	var cacheKeys []string
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

	iter := s.RedisClient.Scan(0, fmt.Sprintf("SESSION-%d-*", buyerID), 1000).Iterator()
	for iter.Next() {
		cacheKeys = append(cacheKeys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan redis: %w", err)
	}

	res, err := s.RedisClient.MGet(cacheKeys...).Result()
	if err != nil {
		return fmt.Errorf("failed to multi-get redis: %w", err)
	}

	for _, val := range res {
		if err := cacheEntry.UnmarshalBinary([]byte(val.(string))); err != nil {
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

type GameConfigurationArgs struct {
	BuyerID string `json:"buyer_id"`
}

type GameConfigurationReply struct {
	PublicKey string `json:"public_key"`
}

type gameConfiguration struct {
	SessionID     string    `json:"session_id"`
	UserHash      string    `json:"user_hash"`
	DirectRTT     float64   `json:"direct_rtt"`
	NextRTT       float64   `json:"next_rtt"`
	ChangeRTT     float64   `json:"change_rtt"`
	OnNetworkNext bool      `json:"on_network_next"`
	ExpiresAt     time.Time `json:"expires_at"`
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

	reply.PublicKey = buyer.EncodedPublicKey()

	return nil
}
