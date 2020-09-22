package jsonrpc

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
)

type RelayData struct {
	SessionCount   uint64
	Tx             uint64
	Rx             uint64
	Version        string
	LastUpdateTime time.Time
	CPU            float32
	Mem            float32
}

type RelayStatsMap struct {
	Internal *map[uint64]RelayData
	mu       sync.RWMutex
}

func NewRelayStatsMap() RelayStatsMap {
	m := make(map[uint64]RelayData)
	return RelayStatsMap{
		Internal: &m,
	}
}

func (r *RelayStatsMap) Get(id uint64) (RelayData, bool) {
	r.mu.RLock()
	relay, ok := (*r.Internal)[id]
	r.mu.RUnlock()
	return relay, ok
}

func (r *RelayStatsMap) Swap(m *map[uint64]RelayData) {
	r.mu.Lock()
	r.Internal = m
	r.mu.Unlock()
}

type OpsService struct {
	Release   string
	BuildTime string

	Storage storage.Storer
	// RouteMatrix *routing.RouteMatrix

	Logger log.Logger

	RelayMap *RelayStatsMap
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
	CompanyName string `json:"company_name"`
	CompanyCode string `json:"company_code"`
	ID          uint64 `json:"id"`
}

func (s *OpsService) Buyers(r *http.Request, args *BuyersArgs, reply *BuyersReply) error {
	for _, b := range s.Storage.Buyers() {
		company, err := s.Storage.Customer(b.CompanyCode)
		if err != nil {
			err = fmt.Errorf("Buyers() failed to find company: %v", err)
			s.Logger.Log("err", err)
			return err
		}
		reply.Buyers = append(reply.Buyers, buyer{
			ID:          b.ID,
			CompanyName: company.Name,
			CompanyCode: company.Code,
		})
	}

	sort.Slice(reply.Buyers, func(i int, j int) bool {
		return reply.Buyers[i].CompanyName < reply.Buyers[j].CompanyName
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
	EnvelopeKbpsUp               int64           `json:"envelopeKbpsUp"`
	EnvelopeKbpsDown             int64           `json:"envelopeKbpsDown"`
	Mode                         int64           `json:"mode"`
	MaxNibblinsPerGB             routing.Nibblin `json:"maxNibblinsPerGB"`
	RTTEpsilon                   float32         `json:"rttEpsilon"`
	RTTThreshold                 float32         `json:"rttThreshold"`
	RTTHysteresis                float32         `json:"rttHysteresis"`
	RTTVeto                      float32         `json:"rttVeto"`
	EnableYouOnlyLiveOnce        bool            `json:"yolo"`
	EnablePacketLossSafety       bool            `json:"plSafety"`
	EnableMultipathForPacketLoss bool            `json:"plMultipath"`
	EnableMultipathForJitter     bool            `json:"jitterMultipath"`
	EnableMultipathForRTT        bool            `json:"rttMultipath"`
	EnableABTest                 bool            `json:"abTest"`
	EnableTryBeforeYouBuy        bool            `json:"tryBeforeYouBuy"`
	TryBeforeYouBuyMaxSlices     int8            `json:"tryBeforeYouBuyMaxSlices"`
	SelectionPercentage          int64           `json:"selectionPercentage"`
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
			MaxNibblinsPerGB:             buyer.RoutingRulesSettings.MaxNibblinsPerGB,
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
	ID                   string          `json:"id"`
	Name                 string          `json:"name"`
	IngressPriceNibblins routing.Nibblin `json:"ingressPriceNibblins"`
	EgressPriceNibblins  routing.Nibblin `json:"egressPriceNibblins"`
}

func (s *OpsService) Sellers(r *http.Request, args *SellersArgs, reply *SellersReply) error {
	for _, s := range s.Storage.Sellers() {
		reply.Sellers = append(reply.Sellers, seller{
			ID:                   s.ID,
			Name:                 s.Name,
			IngressPriceNibblins: s.IngressPriceNibblinsPerGB,
			EgressPriceNibblins:  s.EgressPriceNibblinsPerGB,
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
	Name     string `json:"name"`
	Code     string `json:"code"`
	BuyerID  string `json:"buyer_id"`
	SellerID string `json:"seller_id"`
}

func (s *OpsService) Customers(r *http.Request, args *CustomersArgs, reply *CustomersReply) error {
	var buyerID string

	customers := s.Storage.Customers()

	for _, c := range customers {
		buyer, _ := s.Storage.BuyerWithCompanyCode(c.Code)
		seller, _ := s.Storage.SellerWithCompanyCode(c.Code)
		if buyer.ID == 0 {
			buyerID = ""
		} else {
			buyerID = fmt.Sprintf("%x", buyer.ID)
		}
		reply.Customers = append(reply.Customers, customer{
			Name:     c.Name,
			Code:     c.Code,
			BuyerID:  buyerID,
			SellerID: seller.ID,
		})
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

type SetCustomerLinkArgs struct {
	CustomerName string
	BuyerID      uint64
	SellerID     string
}

type SetCustomerLinkReply struct{}

func (s *OpsService) SetCustomerLink(r *http.Request, args *SetCustomerLinkArgs, reply *SetCustomerLinkReply) error {
	if args.CustomerName == "" {
		err := errors.New("SetCustomerLink() error: customer name empty")
		s.Logger.Log("err", err)
		return err
	}

	if args.BuyerID == 0 && args.SellerID == "" {
		err := errors.New("SetCustomerLink() error: invalid paramters - both buyer ID and seller ID are empty")
		s.Logger.Log("err", err)
		return err
	}

	if args.BuyerID != 0 && args.SellerID != "" {
		err := errors.New("SetCustomerLink() error: invalid paramters - both buyer ID and seller ID are given which is not allowed")
		s.Logger.Log("err", err)
		return err
	}

	ctx, cancelFunc := context.WithDeadline(context.Background(), time.Now().Add(10*time.Second))
	defer cancelFunc()

	buyerID := args.BuyerID
	sellerID := args.SellerID

	if buyerID != 0 {
		// We're trying to update the link to the buyer ID, so get the existing seller ID so it doesn't change
		var err error
		sellerID, err = s.Storage.SellerIDFromCustomerName(ctx, args.CustomerName)
		if err != nil {
			err = fmt.Errorf("SetCustomerLink() error: %w", err)
			s.Logger.Log("err", err)
			return err
		}
	}

	if sellerID != "" {
		// We're trying to update the link to the seller ID, so get the existing buyer ID so it doesn't change
		var err error
		buyerID, err = s.Storage.BuyerIDFromCustomerName(ctx, args.CustomerName)
		if err != nil {
			err = fmt.Errorf("SetCustomerLink() error: %w", err)
			s.Logger.Log("err", err)
			return err
		}
	}

	if err := s.Storage.SetCustomerLink(ctx, args.CustomerName, buyerID, sellerID); err != nil {
		err = fmt.Errorf("SetCustomerLink() error: %w", err)
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
	ID                  uint64                `json:"id"`
	SignedID            int64                 `json:"signed_id"`
	Name                string                `json:"name"`
	Addr                string                `json:"addr"`
	Latitude            float64               `json:"latitude"`
	Longitude           float64               `json:"longitude"`
	NICSpeedMbps        int32                 `json:"nicSpeedMbps"`
	IncludedBandwidthGB int32                 `json:"includedBandwidthGB"`
	State               string                `json:"state"`
	LastUpdateTime      time.Time             `json:"lastUpdateTime"`
	ManagementAddr      string                `json:"management_addr"`
	SSHUser             string                `json:"ssh_user"`
	SSHPort             int64                 `json:"ssh_port"`
	MaxSessionCount     uint32                `json:"maxSessionCount"`
	SessionCount        uint64                `json:"sessionCount"`
	BytesSent           uint64                `json:"bytesTx"`
	BytesReceived       uint64                `json:"bytesRx"`
	PublicKey           string                `json:"public_key"`
	UpdateKey           string                `json:"update_key"`
	FirestoreID         string                `json:"firestore_id"`
	Version             string                `json:"relay_version"`
	SellerName          string                `json:"seller_name"`
	MRC                 routing.Nibblin       `json:"monthlyRecurringChargeNibblins"`
	Overage             routing.Nibblin       `json:"overage"`
	BWRule              routing.BandWidthRule `json:"bandwidthRule"`
	ContractTerm        int32                 `json:"contractTerm"`
	StartDate           time.Time             `json:"startDate"`
	EndDate             time.Time             `json:"endDate"`
	Type                routing.MachineType   `json:"machineType"`
	CPUUsage            float32               `json:"cpu_usage"`
	MemUsage            float32               `json:"mem_usage"`
}

func (s *OpsService) Relays(r *http.Request, args *RelaysArgs, reply *RelaysReply) error {
	for _, r := range s.Storage.Relays() {
		relay := relay{
			ID:                  r.ID,
			SignedID:            r.SignedID,
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
			SellerName:          r.Seller.Name,
			MRC:                 r.MRC,
			Overage:             r.Overage,
			BWRule:              r.BWRule,
			ContractTerm:        r.ContractTerm,
			StartDate:           r.StartDate,
			EndDate:             r.EndDate,
			Type:                r.Type,
		}

		if relayData, ok := s.RelayMap.Get(r.ID); ok {
			relay.SessionCount = relayData.SessionCount
			relay.BytesSent = relayData.Tx
			relay.BytesReceived = relayData.Rx
			relay.Version = relayData.Version
			relay.LastUpdateTime = relayData.LastUpdateTime
			relay.CPUUsage = relayData.CPU
			relay.MemUsage = relayData.Mem
		}

		reply.Relays = append(reply.Relays, relay)
	}

	if args.Regex != "" {
		var filtered []relay

		// first check for an exact match
		for idx := range reply.Relays {
			relay := &reply.Relays[idx]
			if relay.Name == args.Regex {
				filtered = append(filtered, *relay)
				break
			}
		}

		// if no relay found, attemt to see if the query matches any seller names
		if len(filtered) == 0 {
			for idx := range reply.Relays {
				relay := &reply.Relays[idx]
				if args.Regex == relay.SellerName {
					filtered = append(filtered, *relay)
				}
			}
		}

		// if still no matches are found, match by regex
		if len(filtered) == 0 {
			for idx := range reply.Relays {
				relay := &reply.Relays[idx]
				if match, err := regexp.Match(args.Regex, []byte(relay.Name)); match && err == nil {
					filtered = append(filtered, *relay)
					continue
				} else if err != nil {
					return err
				}
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

	// Rather than actually removing the relay from firestore, just
	// rename it and set it to the decomissioned state
	relay.State = routing.RelayStateDecommissioned

	shortDate := time.Now().Format("2006-01-02")
	shortTime := time.Now().Format("15:04:05")
	relay.Name = fmt.Sprintf("%s-%s-%s", relay.Name, shortDate, shortTime)

	if err = s.Storage.SetRelay(context.Background(), relay); err != nil {
		err = fmt.Errorf("RemoveRelay() Storage.SetRelay error: %w", err)
		s.Logger.Log("err", err)
		return err
	}

	return nil
}

type RelayNameUpdateArgs struct {
	RelayID   uint64 `json:"relay_id"`
	RelayName string `json:"relay_name"`
}

type RelayNameUpdateReply struct {
}

func (s *OpsService) RelayNameUpdate(r *http.Request, args *RelayNameUpdateArgs, reply *RelayNameUpdateReply) error {

	relay, err := s.Storage.Relay(args.RelayID)
	if err != nil {
		err = fmt.Errorf("RelayNameUpdate() Storage.Relay error: %w", err)
		s.Logger.Log("err", err)
		return err
	}

	relay.Name = args.RelayName
	if err = s.Storage.SetRelay(context.Background(), relay); err != nil {
		err = fmt.Errorf("Storage.SetRelay error: %w", err)
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

// TODO This endpoint has been deprecated by SetRelayMetadata()?
func (s *OpsService) RelayNICSpeedUpdate(r *http.Request, args *RelayNICSpeedUpdateArgs, reply *RelayNICSpeedUpdateReply) error {

	relay, err := s.Storage.Relay(args.RelayID)
	if err != nil {
		err = fmt.Errorf("RelayNICSpeedUpdate() Relay error: %w", err)
		s.Logger.Log("err", err)
		return err
	}

	relay.NICSpeedMbps = int32(args.RelayNICSpeed)
	if err = s.Storage.SetRelay(context.Background(), relay); err != nil {
		err = fmt.Errorf("RelayNICSpeedUpdate() SetRelay error: %w", err)
		s.Logger.Log("err", err)
		return err
	}

	return nil
}

type DatacenterArg struct {
	ID uint64
}

type DatacenterReply struct {
	Datacenter routing.Datacenter
}

func (s *OpsService) Datacenter(r *http.Request, arg *DatacenterArg, reply *DatacenterReply) error {

	var datacenter routing.Datacenter
	var err error
	if datacenter, err = s.Storage.Datacenter(arg.ID); err != nil {
		err = fmt.Errorf("Datacenter() error: %w", err)
		s.Logger.Log("err", err)
		return err
	}

	reply.Datacenter = datacenter
	return nil

}

type DatacentersArgs struct {
	Name string `json:"name"`
}

type DatacentersReply struct {
	Datacenters []datacenter
}

type datacenter struct {
	Name         string  `json:"name"`
	ID           uint64  `json:"id"`
	SignedID     int64   `json:"signed_id"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	Enabled      bool    `json:"enabled"`
	SupplierName string  `json:"supplierName"`
}

func (s *OpsService) Datacenters(r *http.Request, args *DatacentersArgs, reply *DatacentersReply) error {
	for _, d := range s.Storage.Datacenters() {
		reply.Datacenters = append(reply.Datacenters, datacenter{
			Name:         d.Name,
			ID:           d.ID,
			SignedID:     d.SignedID,
			Enabled:      d.Enabled,
			Latitude:     d.Location.Latitude,
			Longitude:    d.Location.Longitude,
			SupplierName: d.SupplierName,
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

type ListDatacenterMapsArgs struct {
	DatacenterID uint64
}

type ListDatacenterMapsReply struct {
	DatacenterMaps []DatacenterMapsFull
}

// A zero DatacenterID returns a list of all maps.
func (s *OpsService) ListDatacenterMaps(r *http.Request, args *ListDatacenterMapsArgs, reply *ListDatacenterMapsReply) error {

	var dcm map[uint64]routing.DatacenterMap
	dcm = s.Storage.ListDatacenterMaps(args.DatacenterID)

	var replySlice []DatacenterMapsFull
	for _, dcMap := range dcm {
		buyer, err := s.Storage.Buyer(dcMap.BuyerID)
		if err != nil {
			err = fmt.Errorf("ListDatacenterMaps() could not parse buyer: %w", err)
			s.Logger.Log("err", err)
			return err
		}
		datacenter, err := s.Storage.Datacenter(dcMap.Datacenter)
		if err != nil {
			err = fmt.Errorf("ListDatacenterMaps() could not parse datacenter: %w", err)
			s.Logger.Log("err", err)
			return err
		}

		company, err := s.Storage.Customer(buyer.CompanyCode)
		if err != nil {
			err = fmt.Errorf("ListDatacenterMaps() failed to find buyer company: %w", err)
			s.Logger.Log("err", err)
			return err
		}

		dcmFull := DatacenterMapsFull{
			Alias:          dcMap.Alias,
			DatacenterName: datacenter.Name,
			DatacenterID:   fmt.Sprintf("%016x", dcMap.Datacenter),
			BuyerName:      company.Name,
			BuyerID:        fmt.Sprintf("%016x", dcMap.BuyerID),
		}

		replySlice = append(replySlice, dcmFull)
	}

	reply.DatacenterMaps = replySlice

	return nil
}

type RelayMetadataArgs struct {
	Relay routing.Relay
}

type RelayMetadataReply struct {
	Ok           bool
	ErrorMessage string
}

func (s *OpsService) RelayMetadata(r *http.Request, args *RelayMetadataArgs, reply *RelayMetadataReply) error {

	err := s.Storage.SetRelayMetadata(context.Background(), args.Relay)
	if err != nil {
		return err // TODO detail
	}

	return nil
}

// used in routes.go

type RouteSelectionArgs struct {
	SourceRelays      []string `json:"src_relays"`
	DestinationRelays []string `json:"dest_relays"`
	RTT               float64  `json:"rtt"`
	RouteHash         uint64   `json:"route_hash"`
}

type RouteSelectionReply struct {
	Routes []routing.Route `json:"routes"`
}

// func (s *OpsService) RouteSelection(r *http.Request, args *RouteSelectionArgs, reply *RouteSelectionReply) error {
// 	relays := s.Storage.Relays()

// 	var srcrelays []routing.Relay
// 	for _, relay := range relays {
// 		for _, srcrelay := range args.SourceRelays {
// 			if relay.Name == srcrelay {
// 				srcrelays = append(srcrelays, relay)
// 			}
// 		}
// 	}
// 	if len(srcrelays) == 0 {
// 		srcrelays = relays
// 	}

// 	var destrelays []routing.Relay
// 	for _, relay := range relays {
// 		for _, destrelay := range args.DestinationRelays {
// 			if relay.Name == destrelay {
// 				destrelays = append(destrelays, relay)
// 			}
// 		}
// 	}
// 	if len(destrelays) == 0 {
// 		destrelays = relays
// 	}

// 	var selectors []routing.SelectorFunc
// 	selectors = append(selectors, routing.SelectUnencumberedRoutes(0.8))

// 	if args.RTT > 0 {
// 		selectors = append(selectors, routing.SelectAcceptableRoutesFromBestRTT(args.RTT))
// 	}

// 	if args.RouteHash > 0 {
// 		selectors = append(selectors, routing.SelectContainsRouteHash(args.RouteHash))
// 	}

// 	// todo: fill in source relay costs here
// 	sourceRelayCosts := make([]int, len(srcrelays))

// 	routes, err := s.RouteMatrix.Routes(srcrelays, sourceRelayCosts, destrelays, selectors...)
// 	if err != nil {
// 		err = fmt.Errorf("RouteSelection() Routes error: %w", err)
// 		s.Logger.Log("err", err)
// 		return err
// 	}

// 	for routeidx := range routes {
// 		for relayidx := range routes[routeidx].Relays {
// 			routes[routeidx].Relays[relayidx], err = s.Storage.Relay(routes[routeidx].Relays[relayidx].ID)
// 			if err != nil {
// 				err = fmt.Errorf("RouteSelection() Relays error: %w", err)
// 				s.Logger.Log("err", err)
// 				return err
// 			}
// 		}
// 	}

// 	sort.Slice(routes, func(i int, j int) bool {
// 		return routes[i].Stats.RTT < routes[j].Stats.RTT && routes[i].Relays[0].Name < routes[j].Relays[0].Name
// 	})

// 	reply.Routes = routes

// 	return nil
// }
