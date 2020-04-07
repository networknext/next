package jsonrpc

import (
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
)

type OpsService struct {
	RedisClient redis.Cmdable
	Storage     storage.Storer
}

type BuyersArgs struct{}

type BuyersReply struct {
	Buyers []buyer
}

type buyer struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

func (s *OpsService) Buyers(r *http.Request, args *BuyersArgs, reply *BuyersReply) error {
	for _, b := range s.Storage.Buyers() {
		reply.Buyers = append(reply.Buyers, buyer{
			ID:   b.ID,
			Name: b.Name,
		})
	}

	sort.Slice(reply.Buyers, func(i int, j int) bool {
		return reply.Buyers[i].Name < reply.Buyers[j].Name
	})

	return nil
}

type RelaysArgs struct {
	Name string `json:"name"`
}

type RelaysReply struct {
	Relays []relay
}

type relay struct {
	ID                  uint64             `json:"id"`
	Name                string             `json:"name"`
	Addr                string             `json:"addr"`
	Latitude            float64            `json:"latitude"`
	Longitude           float64            `json:"longitude"`
	NICSpeedMbps        int                `json:"nic_speed_mpbs"`
	IncludedBandwidthGB int                `json:"included_bandwidth_gb"`
	State               routing.RelayState `json:"state"`
	StateUpdateTime     time.Time          `json:"stateUpdateTime"`
	ManagementAddr      string             `json:"management_addr"`
	SSHUser             string             `json:"ssh_user"`
	SSHPort             int64              `json:"ssh_port"`
}

func (s *OpsService) Relays(r *http.Request, args *RelaysArgs, reply *RelaysReply) error {
	for _, r := range s.Storage.Relays() {
		reply.Relays = append(reply.Relays, relay{
			ID:                  r.ID,
			Name:                r.Name,
			Addr:                r.Addr.String(),
			Latitude:            r.Latitude,
			Longitude:           r.Longitude,
			NICSpeedMbps:        r.NICSpeedMbps,
			IncludedBandwidthGB: r.IncludedBandwidthGB,
			ManagementAddr:      r.ManagementAddr,
			SSHUser:             r.SSHUser,
			SSHPort:             r.SSHPort,
			State:               r.State,
			StateUpdateTime:     r.LastUpdateTime,
		})
	}

	if args.Name != "" {
		var filtered []relay
		for idx := range reply.Relays {
			if strings.Contains(reply.Relays[idx].Name, args.Name) {
				filtered = append(filtered, reply.Relays[idx])
			}
		}
		reply.Relays = filtered
	}

	sort.Slice(reply.Relays, func(i int, j int) bool {
		return reply.Relays[i].Name < reply.Relays[j].Name
	})

	return nil
}

type DatacentersArgs struct {
	Name string `json:"name"`
}

type DatacentersReply struct {
	Datacenters []datacenter
}

type datacenter struct {
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Enabled   bool    `json:"enabled"`
}

func (s *OpsService) Datacenters(r *http.Request, args *DatacentersArgs, reply *DatacentersReply) error {
	for _, d := range s.Storage.Datacenters() {
		reply.Datacenters = append(reply.Datacenters, datacenter{
			Name:      d.Name,
			Enabled:   d.Enabled,
			Latitude:  d.Location.Latitude,
			Longitude: d.Location.Longitude,
		})
	}

	if args.Name != "" {
		var filtered []datacenter
		for idx := range reply.Datacenters {
			if strings.Contains(reply.Datacenters[idx].Name, args.Name) {
				filtered = append(filtered, reply.Datacenters[idx])
			}
		}
		reply.Datacenters = filtered
	}

	sort.Slice(reply.Datacenters, func(i int, j int) bool {
		return reply.Datacenters[i].Name < reply.Datacenters[j].Name
	})

	return nil
}
