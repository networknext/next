package jsonrpc

import (
	"errors"
	"net/http"
	"sort"
	"strings"

	"github.com/go-redis/redis/v7"
	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/storage"
)

type OpsService struct {
	RedisClient redis.Cmdable
	Storage     storage.Storer
}

type RelaysArgs struct {
	Addr       string `json:"addr"`
	Datacenter string `json:"datacenter"`
}

type RelaysReply struct {
	Relays []relay
}

type relay struct {
	ID                  uint64  `json:"id"`
	Addr                string  `json:"addr"`
	PublicKey           string  `json:"public_key"`
	Datacenter          string  `json:"datacenter"`
	Seller              string  `json:"seller"`
	Latitude            float64 `json:"latitude"`
	Longitude           float64 `json:"longitude"`
	NICSpeedMbps        int     `json:"nic_speed_mpbs"`
	IncludedBandwidthGB int     `json:"included_bandwidth_gb"`
}

func (s *OpsService) Relays(r *http.Request, args *RelaysArgs, reply *RelaysReply) error {
	if args.Addr != "" {
		r, ok := s.Storage.Relay(crypto.HashID(args.Addr))
		if !ok {
			return errors.New("not found")
		}

		reply.Relays = []relay{
			{
				ID:                  r.ID,
				Addr:                r.Addr.String(),
				PublicKey:           r.EncodedPublicKey(),
				Datacenter:          r.Datacenter.Name,
				Seller:              r.Seller.Name,
				Latitude:            r.Latitude,
				Longitude:           r.Longitude,
				NICSpeedMbps:        r.NICSpeedMbps,
				IncludedBandwidthGB: r.IncludedBandwidthGB,
			},
		}

		return nil
	}

	for _, r := range s.Storage.Relays() {
		reply.Relays = append(reply.Relays, relay{
			ID:                  r.ID,
			Addr:                r.Addr.String(),
			PublicKey:           r.EncodedPublicKey(),
			Datacenter:          r.Datacenter.Name,
			Seller:              r.Seller.Name,
			Latitude:            r.Latitude,
			Longitude:           r.Longitude,
			NICSpeedMbps:        r.NICSpeedMbps,
			IncludedBandwidthGB: r.IncludedBandwidthGB,
		})
	}

	if args.Datacenter != "" {
		var filtered []relay
		for idx := range reply.Relays {
			if strings.Contains(reply.Relays[idx].Datacenter, args.Datacenter) {
				filtered = append(filtered, reply.Relays[idx])
			}
		}
		reply.Relays = filtered
	}

	sort.Slice(reply.Relays, func(i int, j int) bool {
		return reply.Relays[i].Datacenter < reply.Relays[j].Datacenter
	})

	return nil
}
