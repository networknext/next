package jsonrpc

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-redis/redis/v7"
	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
)

type OpsService struct {
	Release   string
	BuildTime string

	RedisClient redis.Cmdable
	Storage     storage.Storer
	RouteMatrix *routing.RouteMatrix

	Logger log.Logger
}

type CurrentReleaseArgs struct{}

type CurrentReleaseReply struct {
	Release   string
	BuildTime string
}

func (s *OpsService) CurrentRelease(r *http.Request, args *CurrentReleaseArgs, reply *CurrentReleaseReply) error {
	reply.Release = s.Release
	reply.BuildTime = s.BuildTime
	return nil
}

type BuyersArgs struct{}

type BuyersReply struct {
	Buyers []buyer
}

type buyer struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (s *OpsService) Buyers(r *http.Request, args *BuyersArgs, reply *BuyersReply) error {
	for _, b := range s.Storage.Buyers() {
		reply.Buyers = append(reply.Buyers, buyer{
			ID:   fmt.Sprintf("%x", b.ID),
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

	return s.Storage.AddBuyer(ctx, args.Buyer)
}

type RemoveBuyerArgs struct {
	ID string
}

type RemoveBuyerReply struct{}

func (s *OpsService) RemoveBuyer(r *http.Request, args *RemoveBuyerArgs, reply *RemoveBuyerReply) error {
	ctx, cancelFunc := context.WithDeadline(context.Background(), time.Now().Add(10*time.Second))
	defer cancelFunc()

	buyerID, err := strconv.ParseUint(args.ID, 16, 64)
	if err != nil {
		err = fmt.Errorf("RemoveBuyer() could not convert buyer ID %s to uint64: %v", args.ID, err)
		s.Logger.Log("err", err)
		return err
	}

	return s.Storage.RemoveBuyer(ctx, buyerID)
}

type RoutingRulesSettingsArgs struct {
	BuyerID string
}

type RoutingRulesSettingsReply struct {
	RoutingRuleSettings []routingRuleSettings
}

type routingRuleSettings struct {
	EnvelopeKbpsUp               int64   `json:"envelopeKbpsUp"`
	EnvelopeKbpsDown             int64   `json:"envelopeKbpsDown"`
	Mode                         int64   `json:"mode"`
	MaxCentsPerGB                uint64  `json:"maxCentsPerGB"`
	RTTEpsilon                   float32 `json:"rttEpsilon"`
	RTTThreshold                 float32 `json:"rttThreshold"`
	RTTHysteresis                float32 `json:"rttHysteresis"`
	RTTVeto                      float32 `json:"rttVeto"`
	EnableYouOnlyLiveOnce        bool    `json:"yolo"`
	EnablePacketLossSafety       bool    `json:"plSafety"`
	EnableMultipathForPacketLoss bool    `json:"plMultipath"`
	EnableMultipathForJitter     bool    `json:"jitterMultipath"`
	EnableMultipathForRTT        bool    `json:"rttMultipath"`
	EnableABTest                 bool    `json:"abTest"`
	EnableTryBeforeYouBuy        bool    `json:"tryBeforeYouBuy"`
	TryBeforeYouBuyMaxSlices     int8    `json:"tryBeforeYouBuyMaxSlices"`
	SelectionPercentage          int64   `json:"selectionPercentage"`
}

func (s *OpsService) RoutingRulesSettings(r *http.Request, args *RoutingRulesSettingsArgs, reply *RoutingRulesSettingsReply) error {
	buyerID, err := strconv.ParseUint(args.BuyerID, 16, 64)
	if err != nil {
		err = fmt.Errorf("RoutingRulesSettings() could not convert buyer ID %s to uint64: %v", args.BuyerID, err)
		s.Logger.Log("err", err)
		return err
	}

	buyer, err := s.Storage.Buyer(buyerID)
	if err != nil {
		return err
	}

	reply.RoutingRuleSettings = []routingRuleSettings{
		{
			EnvelopeKbpsUp:               buyer.RoutingRulesSettings.EnvelopeKbpsUp,
			EnvelopeKbpsDown:             buyer.RoutingRulesSettings.EnvelopeKbpsDown,
			Mode:                         buyer.RoutingRulesSettings.Mode,
			MaxCentsPerGB:                buyer.RoutingRulesSettings.MaxCentsPerGB,
			RTTEpsilon:                   buyer.RoutingRulesSettings.RTTEpsilon,
			RTTThreshold:                 buyer.RoutingRulesSettings.RTTThreshold,
			RTTHysteresis:                buyer.RoutingRulesSettings.RTTHysteresis,
			RTTVeto:                      buyer.RoutingRulesSettings.RTTVeto,
			EnableYouOnlyLiveOnce:        buyer.RoutingRulesSettings.EnableYouOnlyLiveOnce,
			EnablePacketLossSafety:       buyer.RoutingRulesSettings.EnablePacketLossSafety,
			EnableMultipathForPacketLoss: buyer.RoutingRulesSettings.EnableMultipathForPacketLoss,
			EnableMultipathForJitter:     buyer.RoutingRulesSettings.EnableMultipathForJitter,
			EnableMultipathForRTT:        buyer.RoutingRulesSettings.EnableMultipathForRTT,
			EnableABTest:                 buyer.RoutingRulesSettings.EnableABTest,
			EnableTryBeforeYouBuy:        buyer.RoutingRulesSettings.EnableTryBeforeYouBuy,
			TryBeforeYouBuyMaxSlices:     buyer.RoutingRulesSettings.TryBeforeYouBuyMaxSlices,
			SelectionPercentage:          buyer.RoutingRulesSettings.SelectionPercentage,
		},
	}

	return nil
}

type SetRoutingRulesSettingsArgs struct {
	BuyerID              string
	RoutingRulesSettings routing.RoutingRulesSettings
}

func (s *OpsService) SetRoutingRulesSettings(r *http.Request, args *SetRoutingRulesSettingsArgs, reply *SetRoutingRulesSettingsReply) error {
	ctx, cancelFunc := context.WithDeadline(context.Background(), time.Now().Add(10*time.Second))
	defer cancelFunc()

	buyerID, err := strconv.ParseUint(args.BuyerID, 16, 64)
	if err != nil {
		err = fmt.Errorf("SetRoutingRulesSettings() could not convert buyer ID %s to uint64: %v", args.BuyerID, err)
		s.Logger.Log("err", err)
		return err
	}

	buyer, err := s.Storage.Buyer(buyerID)
	if err != nil {
		err = fmt.Errorf("SetRoutingRulesSettings() Storage.Buyer error: %w", err)
		s.Logger.Log("err", err)
		return err
	}

	buyer.RoutingRulesSettings = args.RoutingRulesSettings

	return s.Storage.SetBuyer(ctx, buyer)
}

type SetRoutingRulesSettingsReply struct{}

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

type CustomersArgs struct{}

type CustomersReply struct {
	Customers []customer
}

type customer struct {
	BuyerID  string `json:"buyer_id"`
	SellerID string `json:"seller_id"`
	Name     string `json:"name"`
}

func (s *OpsService) Customers(r *http.Request, args *CustomersArgs, reply *CustomersReply) error {

	customers := map[string]customer{}

	for _, b := range s.Storage.Buyers() {
		customers[b.Name] = customer{
			BuyerID: fmt.Sprintf("%x", b.ID),
			Name:    b.Name,
		}
	}

	for _, s := range s.Storage.Sellers() {
		if _, ok := customers[s.Name]; ok {
			cust := customers[s.Name]
			cust.SellerID = s.ID
			customers[s.Name] = cust
		} else {
			customers[s.Name] = customer{
				SellerID: s.ID,
				Name:     s.Name,
			}
		}
	}

	for _, c := range customers {
		reply.Customers = append(reply.Customers, c)
	}

	sort.Slice(reply.Customers, func(i int, j int) bool {
		return reply.Customers[i].Name < reply.Customers[j].Name
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
		err = fmt.Errorf("AddSeller() error: %w", err)
		s.Logger.Log("err", err)
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
		err = fmt.Errorf("RemoveSeller() error: %w", err)
		s.Logger.Log("err", err)
		return err
	}

	return nil
}

type RelaysArgs struct {
	Regex string `json:"name"`
}

type RelaysReply struct {
	Relays []relay `json:"relays"`
}

type relay struct {
	ID                  uint64    `json:"id"`
	Name                string    `json:"name"`
	Addr                string    `json:"addr"`
	Latitude            float64   `json:"latitude"`
	Longitude           float64   `json:"longitude"`
	NICSpeedMbps        uint64    `json:"nic_speed_mpbs"`
	IncludedBandwidthGB uint64    `json:"included_bandwidth_gb"`
	State               string    `json:"state"`
	LastUpdateTime      time.Time `json:"lastUpdateTime"`
	ManagementAddr      string    `json:"management_addr"`
	SSHUser             string    `json:"ssh_user"`
	SSHPort             int64     `json:"ssh_port"`
	MaxSessionCount     uint32    `json:"maxSessionCount"`
	SessionCount        uint64    `json:"sessionCount"`
	BytesSent           uint64    `json:"bytesTx"`
	BytesReceived       uint64    `json:"bytesRx"`
	PublicKey           string    `json:"public_key"`
	UpdateKey           string    `json:"update_key"`
	FirestoreID         string    `json:"firestore_id"`
	Version             string    `json:"relay_version"`
	SellerID            string    `json:"seller_id"`
	SellerName          string    `json:"seller_name"`
	DatacenterName      string    `json:"datacenter_name"`
}

func (s *OpsService) Relays(r *http.Request, args *RelaysArgs, reply *RelaysReply) error {
	hgetallResult := s.RedisClient.HGetAll(routing.HashKeyAllRelays)
	if hgetallResult.Err() != nil && hgetallResult.Err() != redis.Nil {
		err := fmt.Errorf("failed to get all relays: %v", hgetallResult.Err())
		s.Logger.Log("err", err)
		return err
	}

	relayCacheEntries := hgetallResult.Val()

	for _, r := range s.Storage.Relays() {
		relay := relay{
			ID:                  r.ID,
			Name:                r.Name,
			Addr:                r.Addr.String(),
			Latitude:            r.Datacenter.Location.Latitude,
			Longitude:           r.Datacenter.Location.Longitude,
			NICSpeedMbps:        r.NICSpeedMbps,
			IncludedBandwidthGB: r.IncludedBandwidthGB,
			ManagementAddr:      r.ManagementAddr,
			SSHUser:             r.SSHUser,
			SSHPort:             r.SSHPort,
			State:               r.State.String(),
			LastUpdateTime:      r.LastUpdateTime,
			PublicKey:           base64.StdEncoding.EncodeToString(r.PublicKey),
			UpdateKey:           base64.StdEncoding.EncodeToString(r.UpdateKey),
			FirestoreID:         r.FirestoreID,
			MaxSessionCount:     r.MaxSessions,
			SellerID:            r.Seller.ID,
			SellerName:          r.Seller.Name,
			DatacenterName:      r.Datacenter.Name,
		}

		relayCacheEntry := routing.RelayCacheEntry{
			ID: r.ID,
		}

		// If the relay is in redis, get its traffic stats and last update time
		if relayCacheEntryString, ok := relayCacheEntries[relayCacheEntry.Key()]; ok {
			if err := relayCacheEntry.UnmarshalBinary([]byte(relayCacheEntryString)); err == nil {
				relay.SessionCount = relayCacheEntry.TrafficStats.SessionCount
				relay.BytesSent = relayCacheEntry.TrafficStats.BytesSent
				relay.BytesReceived = relayCacheEntry.TrafficStats.BytesReceived

				relay.LastUpdateTime = relayCacheEntry.LastUpdateTime
				relay.Version = relayCacheEntry.Version
			}
		}

		reply.Relays = append(reply.Relays, relay)
	}

	if args.Regex != "" {
		var filtered []relay
		for idx := range reply.Relays {
			if match, err := regexp.Match(args.Regex, []byte(reply.Relays[idx].Name)); match && err == nil {
				filtered = append(filtered, reply.Relays[idx])
				continue
			} else if err != nil {
				return err
			}

			if match, err := regexp.Match(args.Regex, []byte(reply.Relays[idx].SellerName)); match && err == nil {
				filtered = append(filtered, reply.Relays[idx])
			} else if err != nil {
				return err
			}
		}
		reply.Relays = filtered
	}

	sort.Slice(reply.Relays, func(i int, j int) bool {
		return reply.Relays[i].Name < reply.Relays[j].Name
	})

	return nil
}

type AddRelayArgs struct {
	Relay routing.Relay
}

type AddRelayReply struct{}

func (s *OpsService) AddRelay(r *http.Request, args *AddRelayArgs, reply *AddRelayReply) error {
	ctx, cancelFunc := context.WithDeadline(context.Background(), time.Now().Add(10*time.Second))
	defer cancelFunc()

	if err := s.Storage.AddRelay(ctx, args.Relay); err != nil {
		err = fmt.Errorf("AddRelay() error: %w", err)
		s.Logger.Log("err", err)
		return err
	}

	return nil
}

type RemoveRelayArgs struct {
	RelayID uint64
}

type RemoveRelayReply struct{}

func (s *OpsService) RemoveRelay(r *http.Request, args *RemoveRelayArgs, reply *RemoveRelayReply) error {
	relay, err := s.Storage.Relay(args.RelayID)
	if err != nil {
		err = fmt.Errorf("RemoveRelay() Storage.Relay error: %w", err)
		s.Logger.Log("err", err)
		return err
	}

	// Rather than actually removing the relay from firestore, just set it to the decomissioned state
	relay.State = routing.RelayStateDecommissioned

	if err = s.Storage.SetRelay(context.Background(), relay); err != nil {
		err = fmt.Errorf("RemoveRelay() Storage.SetRelay error: %w", err)
		s.Logger.Log("err", err)
		return err
	}

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
		err = fmt.Errorf("RelayStateUpdate() Storage.Relay error: %w", err)
		s.Logger.Log("err", err)
		return err
	}

	relay.State = args.RelayState
	if err = s.Storage.SetRelay(context.Background(), relay); err != nil {
		err = fmt.Errorf("RelayStateUpdate() Storage.SetRelay error: %w", err)
		s.Logger.Log("err", err)
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
		err = fmt.Errorf("RelayPublicKeyUpdate()")
		return err
	}

	relay.PublicKey, err = base64.StdEncoding.DecodeString(args.RelayPublicKey)

	if err != nil {
		err = fmt.Errorf("RelayPublicKeyUpdate() could not decode relay public key: %v", err)
		s.Logger.Log("err", err)
		return err
	}

	if err = s.Storage.SetRelay(context.Background(), relay); err != nil {
		err = fmt.Errorf("RelayPublicKeyUpdate() SetRelay error: %w", err)
		s.Logger.Log("err", err)
		return err
	}

	return nil
}

type RelayNICSpeedUpdateArgs struct {
	RelayID       uint64 `json:"relay_id"`
	RelayNICSpeed uint64 `json:"relay_nic_speed"`
}

type RelayNICSpeedUpdateReply struct {
}

func (s *OpsService) RelayNICSpeedUpdate(r *http.Request, args *RelayNICSpeedUpdateArgs, reply *RelayNICSpeedUpdateReply) error {

	relay, err := s.Storage.Relay(args.RelayID)
	if err != nil {
		err = fmt.Errorf("RelayNICSpeedUpdate() Relay error: %w", err)
		s.Logger.Log("err", err)
		return err
	}

	relay.NICSpeedMbps = args.RelayNICSpeed
	if err = s.Storage.SetRelay(context.Background(), relay); err != nil {
		err = fmt.Errorf("RelayNICSpeedUpdate() SetRelay error: %w", err)
		s.Logger.Log("err", err)
		return err
	}

	return nil
}

type RelayEditArgs struct {
	RelayID   uint64                 `json:"relay_id"`
	RelayData map[string]interface{} `json:"relay_data"`
}

type RelayEditReply struct {
}

func (s *OpsService) RelayEdit(r *http.Request, args *RelayEditArgs, reply *RelayEditReply) error {
	// Get current relay data
	relay, err := s.Storage.Relay(args.RelayID)
	if err != nil {
		err = fmt.Errorf("RelayEdit() Relay error: %w", err)
		s.Logger.Log("err", err)
		return err
	}

	// Edit data
	// This way if a value isn't provided in the map data argument it won't be updated
	for k, v := range args.RelayData {
		switch k {
		case "Name":
			relay.Name = v.(string)
		case "Addr":
			addr, err := net.ResolveUDPAddr("udp", v.(string))
			if err != nil {
				err = fmt.Errorf("RelayEdit() addr parse error: %w", err)
				s.Logger.Log("err", err)
				return err
			}
			relay.Addr = *addr
		case "PublicKey":
			publicKey, err := base64.StdEncoding.DecodeString(v.(string))
			if err != nil {
				err = fmt.Errorf("RelayEdit() public key parse error: %w", err)
				s.Logger.Log("err", err)
				return err
			}
			relay.PublicKey = publicKey
		case "SellerID":
			relay.Seller = routing.Seller{ID: v.(string)}
		case "DatacenterName":
			relay.Datacenter = routing.Datacenter{ID: crypto.HashID(v.(string))}
		case "NicSpeedMbps":
			relay.NICSpeedMbps = uint64(v.(float64))
		case "IncludedBandwidthGB":
			relay.IncludedBandwidthGB = uint64(v.(float64))
		case "ManagementAddr":
			relay.ManagementAddr = v.(string)
		case "SSHUser":
			relay.SSHUser = v.(string)
		case "SSHPort":
			relay.SSHPort = int64(v.(float64))
		case "MaxSessions":
			relay.MaxSessions = uint32(v.(float64))
		}
	}

	// Save new relay data
	if err = s.Storage.SetRelay(context.Background(), relay); err != nil {
		err = fmt.Errorf("RelayEdit() SetRelay error: %w", err)
		s.Logger.Log("err", err)
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
	ID        string  `json:"id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Enabled   bool    `json:"enabled"`
}

func (s *OpsService) Datacenters(r *http.Request, args *DatacentersArgs, reply *DatacentersReply) error {
	for _, d := range s.Storage.Datacenters() {
		reply.Datacenters = append(reply.Datacenters, datacenter{
			Name:      d.Name,
			ID:        fmt.Sprintf("%x", d.ID),
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
		err = fmt.Errorf("AddDatacenter() error: %w", err)
		s.Logger.Log("err", err)
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
		err = fmt.Errorf("RemoveDatacenter() error: %w", err)
		s.Logger.Log("err", err)
		return err
	}

	return nil
}

type RouteSelectionArgs struct {
	SourceRelays      []string `json:"src_relays"`
	DestinationRelays []string `json:"dest_relays"`
	RTT               float64  `json:"rtt"`
	RouteHash         uint64   `json:"route_hash"`
}

type RouteSelectionReply struct {
	Routes []routing.Route `json:"routes"`
}

func (s *OpsService) RouteSelection(r *http.Request, args *RouteSelectionArgs, reply *RouteSelectionReply) error {
	relays := s.Storage.Relays()

	var srcrelays []routing.Relay
	for _, relay := range relays {
		for _, srcrelay := range args.SourceRelays {
			if relay.Name == srcrelay {
				srcrelays = append(srcrelays, relay)
			}
		}
	}
	if len(srcrelays) == 0 {
		srcrelays = relays
	}

	var destrelays []routing.Relay
	for _, relay := range relays {
		for _, destrelay := range args.DestinationRelays {
			if relay.Name == destrelay {
				destrelays = append(destrelays, relay)
			}
		}
	}
	if len(destrelays) == 0 {
		destrelays = relays
	}

	var selectors []routing.SelectorFunc
	selectors = append(selectors, routing.SelectUnencumberedRoutes(0.8))

	if args.RTT > 0 {
		selectors = append(selectors, routing.SelectAcceptableRoutesFromBestRTT(args.RTT))
	}

	if args.RouteHash > 0 {
		selectors = append(selectors, routing.SelectContainsRouteHash(args.RouteHash))
	}

	routes, err := s.RouteMatrix.Routes(srcrelays, destrelays, selectors...)
	if err != nil {
		err = fmt.Errorf("RouteSelection() Routes error: %w", err)
		s.Logger.Log("err", err)
		return err
	}

	for routeidx := range routes {
		for relayidx := range routes[routeidx].Relays {
			routes[routeidx].Relays[relayidx], err = s.Storage.Relay(routes[routeidx].Relays[relayidx].ID)
			if err != nil {
				err = fmt.Errorf("RouteSelection() Relays error: %w", err)
				s.Logger.Log("err", err)
				return err
			}
		}
	}

	sort.Slice(routes, func(i int, j int) bool {
		return routes[i].Stats.RTT < routes[j].Stats.RTT && routes[i].Relays[0].Name < routes[j].Relays[0].Name
	})

	reply.Routes = routes

	return nil
}
