package jsonrpc

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/networknext/backend/crypto"
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

type AddBuyerArgs struct {
	Buyer routing.Buyer
}

type AddBuyerReply struct{}

func (s *OpsService) AddBuyer(r *http.Request, args *AddBuyerArgs, reply *AddBuyerReply) error {
	ctx, cancelFunc := context.WithDeadline(context.Background(), time.Now().Add(10*time.Second))
	defer cancelFunc()

	if err := s.Storage.AddBuyer(ctx, args.Buyer); err != nil {
		return err
	}

	return nil
}

type RemoveBuyerArgs struct {
	ID uint64
}

type RemoveBuyerReply struct{}

func (s *OpsService) RemoveBuyer(r *http.Request, args *RemoveBuyerArgs, reply *RemoveBuyerReply) error {
	ctx, cancelFunc := context.WithDeadline(context.Background(), time.Now().Add(10*time.Second))
	defer cancelFunc()

	if err := s.Storage.RemoveBuyer(ctx, args.ID); err != nil {
		return err
	}

	return nil
}

type SellersArgs struct{}

type SellersReply struct {
	Sellers []seller
}

type seller struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	IngressPriceCents uint64 `json:"ingressPriceCents"`
	EgressPriceCents  uint64 `json:"egressPriceCents"`
}

func (s *OpsService) Sellers(r *http.Request, args *SellersArgs, reply *SellersReply) error {
	for _, s := range s.Storage.Sellers() {
		reply.Sellers = append(reply.Sellers, seller{
			ID:                s.ID,
			Name:              s.Name,
			IngressPriceCents: s.IngressPriceCents,
			EgressPriceCents:  s.EgressPriceCents,
		})
	}

	sort.Slice(reply.Sellers, func(i int, j int) bool {
		return reply.Sellers[i].Name < reply.Sellers[j].Name
	})

	return nil
}

type AddSellerArgs struct {
	Seller routing.Seller
}

type AddSellerReply struct{}

func (s *OpsService) AddSeller(r *http.Request, args *AddSellerArgs, reply *AddSellerReply) error {
	ctx, cancelFunc := context.WithDeadline(context.Background(), time.Now().Add(10*time.Second))
	defer cancelFunc()

	if err := s.Storage.AddSeller(ctx, args.Seller); err != nil {
		return err
	}

	return nil
}

type RemoveSellerArgs struct {
	ID string
}

type RemoveSellerReply struct{}

func (s *OpsService) RemoveSeller(r *http.Request, args *RemoveSellerArgs, reply *RemoveSellerReply) error {
	ctx, cancelFunc := context.WithDeadline(context.Background(), time.Now().Add(10*time.Second))
	defer cancelFunc()

	if err := s.Storage.RemoveSeller(ctx, args.ID); err != nil {
		return err
	}

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

type RelayStateUpdateArgs struct {
	RelayID    uint64             `json:"relay_id"`
	RelayState routing.RelayState `json:"relay_state"`
}

type RelayStateUpdateReply struct {
}

func (s *OpsService) RelayStateUpdate(r *http.Request, args *RelayStateUpdateArgs, reply *RelayStateUpdateReply) error {
	relay, err := s.Storage.Relay(args.RelayID)

	if err != nil {
		return err
	}

	relay.State = args.RelayState

	if err := s.Storage.SetRelay(context.Background(), relay); err != nil {
		return err
	}

	return nil
}

type RelayPublicKeyUpdateArgs struct {
	RelayID        uint64 `json:"relay_id"`
	RelayPublicKey string `json:"relay_public_key"`
}

type RelayPublicKeyUpdateReply struct {
}

func (s *OpsService) RelayPublicKeyUpdate(r *http.Request, args *RelayPublicKeyUpdateArgs, reply *RelayPublicKeyUpdateReply) error {
	relay, err := s.Storage.Relay(args.RelayID)

	if err != nil {
		return err
	}

	relay.PublicKey, err = base64.StdEncoding.DecodeString(args.RelayPublicKey)

	if err != nil {
		return fmt.Errorf("could not decode relay public key: %v", err)
	}

	if err := s.Storage.SetRelay(context.Background(), relay); err != nil {
		return err
	}

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

type AddDatacenterArgs struct {
	Datacenter routing.Datacenter
}

type AddDatacenterReply struct{}

func (s *OpsService) AddDatacenter(r *http.Request, args *AddDatacenterArgs, reply *AddDatacenterReply) error {
	ctx, cancelFunc := context.WithDeadline(context.Background(), time.Now().Add(10*time.Second))
	defer cancelFunc()

	if err := s.Storage.AddDatacenter(ctx, args.Datacenter); err != nil {
		return err
	}

	return nil
}

type RemoveDatacenterArgs struct {
	Name string
}

type RemoveDatacenterReply struct{}

func (s *OpsService) RemoveDatacenter(r *http.Request, args *RemoveDatacenterArgs, reply *RemoveDatacenterReply) error {
	ctx, cancelFunc := context.WithDeadline(context.Background(), time.Now().Add(10*time.Second))
	defer cancelFunc()

	id := crypto.HashID(args.Name)

	if err := s.Storage.RemoveDatacenter(ctx, id); err != nil {
		return err
	}

	return nil
}
