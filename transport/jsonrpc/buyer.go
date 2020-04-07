package jsonrpc

import (
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/networknext/backend/transport"
)

type BuyersService struct {
	RedisClient redis.Cmdable
}

type MapArgs struct {
	BuyerID uint64 `json:"buyer_id"`
}

type MapReply struct {
	Clusters []cluster `json:"clusters"`
}

type cluster struct {
	Country   string  `json:"country"`
	Region    string  `json:"region"`
	City      string  `json:"city"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Count     int     `json:"count"`
}

func (s *BuyersService) SessionsMap(r *http.Request, args *MapArgs, reply *MapReply) error {
	reply.Clusters = []cluster{
		{Country: "United States", Region: "NY", City: "Troy", Latitude: 42.7273, Longitude: -73.6696, Count: 10},
		{Country: "United States", Region: "NY", City: "Saratoga Springs", Latitude: 43.0034, Longitude: -73.842, Count: 5},
		{Country: "United States", Region: "NY", City: "Albany", Latitude: 42.6701, Longitude: -73.7754, Count: 200},
	}

	return nil
}

type SessionsArgs struct {
	BuyerID   uint64 `json:"buyer_id"`
	SessionID uint64 `json:"session_id"`
}

type SessionsReply struct {
	Sessions []session `json:"sessions"`
}

type session struct {
	SessionID     uint64    `json:"session_id"`
	UserHash      uint64    `json:"user_hash"`
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

	if args.BuyerID == 0 {
		return fmt.Errorf("buyer_id is required")
	}

	if args.SessionID > 0 {
		getCmd := s.RedisClient.Get(fmt.Sprintf("SESSION-%d-%d", args.BuyerID, args.SessionID))
		if cacheEntryData, err = getCmd.Bytes(); err != nil {
			return fmt.Errorf("failed to get session %d: %w", args.SessionID, err)
		}

		if err := cacheEntry.UnmarshalBinary(cacheEntryData); err != nil {
			return fmt.Errorf("failed to unmarshal session %d: %w", args.SessionID, err)
		}

		reply.Sessions = append(reply.Sessions, session{
			SessionID:     cacheEntry.SessionID,
			UserHash:      cacheEntry.UserHash,
			DirectRTT:     cacheEntry.DirectRTT,
			NextRTT:       cacheEntry.NextRTT,
			ChangeRTT:     cacheEntry.NextRTT - cacheEntry.DirectRTT,
			OnNetworkNext: cacheEntry.RouteDecision.OnNetworkNext,
			ExpiresAt:     cacheEntry.TimestampExpire,
		})

		return nil
	}

	iter := s.RedisClient.Scan(0, fmt.Sprintf("SESSION-%d-*", args.BuyerID), 1000).Iterator()
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
			SessionID:     cacheEntry.SessionID,
			UserHash:      cacheEntry.UserHash,
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
