package jsonrpc

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/encoding"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/storage"
	"github.com/networknext/backend/modules/transport/middleware"
)

type RelayVersion struct {
	Major uint8
	Minor uint8
	Patch uint8
}

func (self *RelayVersion) String() string {
	return fmt.Sprintf("%d.%d.%d", self.Major, self.Minor, self.Patch)
}

type RelayData struct {
	SessionCount   uint64
	Version        RelayVersion
	LastUpdateTime time.Time
	CPU            float32
	Mem            float32
	TrafficStats   routing.TrafficStats
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

func (r *RelayStatsMap) ReadAndSwap(data []byte) error {
	index := 0

	var version uint8
	if !encoding.ReadUint8(data, &index, &version) {
		return errors.New("unable to read relay stats version")
	}

	/*
		if version != routing.VersionNumberRelayMap {
			return fmt.Errorf("incorrect relay map version number: %d", version)
		}
	*/

	var count uint64
	if !encoding.ReadUint64(data, &index, &count) {
		return errors.New("unable to read relay stats count")
	}

	m := make(map[uint64]RelayData)

	for i := uint64(0); i < count; i++ {
		var id uint64
		if !encoding.ReadUint64(data, &index, &id) {
			return errors.New("unable to read relay stats id")
		}

		var relay RelayData

		// currently map version & traffic stats match up, but not binding them together in case one changes and the other doesn't
		switch version {
		case 0:
			if err := relay.TrafficStats.ReadFrom(data, &index, 0); err != nil {
				return err
			}
		case 1:
			if err := relay.TrafficStats.ReadFrom(data, &index, 1); err != nil {
				return err
			}
		case 2:
			if err := relay.TrafficStats.ReadFrom(data, &index, 2); err != nil {
				return err
			}
		default:
			return fmt.Errorf("invalid relay map version: %d", version)
		}

		// result of a merge with master, relay.SessionCount was supposed to be removed but the merge put it back in
		// once this code is in prod for compatability, relay.SessionCount can be removed
		if version <= 1 {
			relay.TrafficStats.SessionCount = relay.SessionCount
		} else {
			relay.SessionCount = relay.TrafficStats.SessionCount
		}

		if !encoding.ReadUint8(data, &index, &relay.Version.Major) {
			return errors.New("unable to relay stats major version")
		}

		if !encoding.ReadUint8(data, &index, &relay.Version.Minor) {
			return errors.New("unable to relay stats minor version")
		}

		if !encoding.ReadUint8(data, &index, &relay.Version.Patch) {
			return errors.New("unable to relay stats patch version")
		}

		var unixTime uint64
		if !encoding.ReadUint64(data, &index, &unixTime) {
			return errors.New("unable to read relay stats last update time")
		}
		relay.LastUpdateTime = time.Unix(int64(unixTime), 0)

		if !encoding.ReadFloat32(data, &index, &relay.CPU) {
			return errors.New("unable to read relay stats cpu usage")
		}

		if !encoding.ReadFloat32(data, &index, &relay.Mem) {
			return errors.New("unable to read relay stats memory usage")
		}

		m[id] = relay
	}

	r.Swap(&m)

	return nil
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

	// TODO: remove RelayMay
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
	CompanyName         string `json:"company_name"`
	CompanyCode         string `json:"company_code"`
	ShortName           string `json:"short_name"`
	ID                  uint64 `json:"id"`
	HexID               string `json:"hexID"`
	Live                bool   `json:"live"`
	Debug               bool   `json:"debug"`
	Analytics           bool   `json:"analytics"`
	Billing             bool   `json:"billing"`
	Trial               bool   `json:"trial"`
	ExoticLocationFee   string `json:"exotic_location_fee"`
	StandardLocationFee string `json:"standard_location_fee"`
}

func (s *OpsService) Buyers(r *http.Request, args *BuyersArgs, reply *BuyersReply) error {
	for _, b := range s.Storage.Buyers(r.Context()) {
		c, err := s.Storage.Customer(r.Context(), b.CompanyCode)
		if err != nil {
			err = fmt.Errorf("Buyers() could not find Customer %s for %s: %v", b.CompanyCode, b.String(), err)
			s.Logger.Log("err", err)
			return err
		}

		reply.Buyers = append(reply.Buyers, buyer{
			ID:                  b.ID,
			HexID:               b.HexID,
			CompanyName:         c.Name,
			CompanyCode:         b.CompanyCode,
			ShortName:           b.ShortName,
			Live:                b.Live,
			Debug:               b.Debug,
			Analytics:           b.Analytics,
			Billing:             b.Billing,
			Trial:               b.Trial,
			ExoticLocationFee:   fmt.Sprintf("%f", b.ExoticLocationFee),
			StandardLocationFee: fmt.Sprintf("%f", b.StandardLocationFee),
		})
	}

	sort.Slice(reply.Buyers, func(i int, j int) bool {
		return reply.Buyers[i].CompanyName < reply.Buyers[j].CompanyName
	})

	return nil
}

type JSAddBuyerArgs struct {
	ShortName           string `json:"shortName"`
	Live                bool   `json:"live"`
	Debug               bool   `json:"debug"`
	Analytics           bool   `json:"analytics"`
	Billing             bool   `json:"billing"`
	Trial               bool   `json:"trial"`
	ExoticLocationFee   string `json:"exoticLocationFee"`
	StandardLocationFee string `json:"standardLocationFee"`
	PublicKey           string `json:"publicKey"`
}

type JSAddBuyerReply struct{}

func (s *OpsService) JSAddBuyer(r *http.Request, args *JSAddBuyerArgs, reply *JSAddBuyerReply) error {
	ctx, cancelFunc := context.WithDeadline(r.Context(), time.Now().Add(10*time.Second))
	defer cancelFunc()

	publicKey, err := base64.StdEncoding.DecodeString(args.PublicKey)
	if err != nil {
		s.Logger.Log("err", err)
		return err
	}

	if len(publicKey) != crypto.KeySize+8 {
		s.Logger.Log("err", err)
		return err
	}

	exoticLocationFee, err := strconv.ParseFloat(args.ExoticLocationFee, 64)
	if err != nil {
		s.Logger.Log("err", err)
		return err
	}

	standardLocationFee, err := strconv.ParseFloat(args.StandardLocationFee, 64)
	if err != nil {
		s.Logger.Log("err", err)
		return err
	}

	// slice the public key here instead of in the clients
	buyer := routing.Buyer{
		CompanyCode:         args.ShortName,
		ShortName:           args.ShortName,
		ID:                  binary.LittleEndian.Uint64(publicKey[:8]),
		Live:                args.Live,
		Debug:               args.Debug,
		Analytics:           args.Analytics,
		Billing:             args.Billing,
		Trial:               args.Trial,
		ExoticLocationFee:   exoticLocationFee,
		StandardLocationFee: standardLocationFee,
		PublicKey:           publicKey[8:],
	}

	return s.Storage.AddBuyer(ctx, buyer)
}

type RemoveBuyerArgs struct {
	ID string
}

type RemoveBuyerReply struct{}

func (s *OpsService) RemoveBuyer(r *http.Request, args *RemoveBuyerArgs, reply *RemoveBuyerReply) error {
	ctx, cancelFunc := context.WithDeadline(r.Context(), time.Now().Add(10*time.Second))
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

type SellersArgs struct{}

type SellersReply struct {
	Sellers []seller
}

type seller struct {
	ID                  string          `json:"id"`
	Name                string          `json:"name"`
	EgressPriceNibblins routing.Nibblin `json:"egressPriceNibblins"`
	Secret              bool            `json:"secret"`
}

func (s *OpsService) Sellers(r *http.Request, args *SellersArgs, reply *SellersReply) error {
	for _, localSeller := range s.Storage.Sellers(r.Context()) {
		// this is broken in firestore, customers in general do not exist
		// c, err := s.Storage.Customer(localSeller.CompanyCode)
		// if err != nil {
		// 	err = fmt.Errorf("Sellers() could not find Customer %s: %v", localSeller.CompanyCode, err)
		// 	s.Logger.Log("err", err)
		// 	return err
		// }
		reply.Sellers = append(reply.Sellers, seller{
			ID:                  localSeller.ID,
			Name:                localSeller.Name,
			EgressPriceNibblins: localSeller.EgressPriceNibblinsPerGB,
			Secret:              localSeller.Secret,
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
	Name                   string `json:"name"`
	Code                   string `json:"code"`
	AutomaticSignInDomains string `json:"automaticSigninDomains"`
	BuyerID                string `json:"buyer_id"`
	SellerID               string `json:"seller_id"`
}

func (s *OpsService) Customers(r *http.Request, args *CustomersArgs, reply *CustomersReply) error {

	customers := s.Storage.Customers(r.Context())

	for _, c := range customers {

		buyerID := ""

		// TODO both of these support functions should be
		// removed or modified to check by FK
		buyer, _ := s.Storage.BuyerWithCompanyCode(r.Context(), c.Code)
		seller, _ := s.Storage.SellerWithCompanyCode(r.Context(), c.Code)

		if buyer.ID != 0 {
			buyerID = fmt.Sprintf("%x", buyer.ID)
		}

		reply.Customers = append(reply.Customers, customer{
			Name:                   c.Name,
			Code:                   c.Code,
			AutomaticSignInDomains: c.AutomaticSignInDomains,
			BuyerID:                buyerID,
			SellerID:               seller.ID,
		})
	}

	sort.Slice(reply.Customers, func(i int, j int) bool {
		return reply.Customers[i].Name < reply.Customers[j].Name
	})
	return nil
}

type JSAddCustomerArgs struct {
	Code                   string `json:"code"`
	Name                   string `json:"name"`
	AutomaticSignInDomains string `json:"automaticSignInDomains"`
}

type JSAddCustomerReply struct{}

func (s *OpsService) JSAddCustomer(r *http.Request, args *JSAddCustomerArgs, reply *JSAddCustomerReply) error {
	ctx, cancelFunc := context.WithDeadline(r.Context(), time.Now().Add(10*time.Second))
	defer cancelFunc()

	customer := routing.Customer{
		Name:                   args.Name,
		Code:                   args.Code,
		AutomaticSignInDomains: args.AutomaticSignInDomains,
	}

	if err := s.Storage.AddCustomer(ctx, customer); err != nil {
		err = fmt.Errorf("AddCustomer() error: %w", err)
		s.Logger.Log("err", err)
		return err
	}
	return nil
}

type AddCustomerArgs struct {
	Customer routing.Customer
}

type AddCustomerReply struct{}

func (s *OpsService) AddCustomer(r *http.Request, args *AddCustomerArgs, reply *AddCustomerReply) error {
	ctx, cancelFunc := context.WithDeadline(r.Context(), time.Now().Add(10*time.Second))
	defer cancelFunc()

	if err := s.Storage.AddCustomer(ctx, args.Customer); err != nil {
		err = fmt.Errorf("AddCustomer() error: %w", err)
		s.Logger.Log("err", err)
		return err
	}
	return nil
}

type CustomerArg struct {
	CustomerID string
}

type CustomerReply struct {
	Customer routing.Customer
}

func (s *OpsService) Customer(r *http.Request, arg *CustomerArg, reply *CustomerReply) error {

	var c routing.Customer
	var err error

	if c, err = s.Storage.Customer(r.Context(), arg.CustomerID); err != nil {
		err = fmt.Errorf("Customer() error: %w", err)
		s.Logger.Log("err", err)
		return err
	}
	reply.Customer = c

	return nil
}

type SellerArg struct {
	ID string
}

type SellerReply struct {
	Seller routing.Seller
}

func (s *OpsService) Seller(r *http.Request, arg *SellerArg, reply *SellerReply) error {

	var seller routing.Seller
	var err error
	if seller, err = s.Storage.Seller(r.Context(), arg.ID); err != nil {
		err = fmt.Errorf("Seller() error: %w", err)
		s.Logger.Log("err", err)
		return err
	}

	reply.Seller = seller
	return nil

}

type JSAddSellerArgs struct {
	ShortName    string `json:"shortName"`
	Secret       bool   `json:"secret"`
	IngressPrice int64  `json:"ingressPrice"` // nibblins
	EgressPrice  int64  `json:"egressPrice"`  // nibblins
}

type JSAddSellerReply struct{}

func (s *OpsService) JSAddSeller(r *http.Request, args *JSAddSellerArgs, reply *JSAddSellerReply) error {
	ctx, cancelFunc := context.WithDeadline(r.Context(), time.Now().Add(10*time.Second))
	defer cancelFunc()

	seller := routing.Seller{
		ID:                       args.ShortName,
		ShortName:                args.ShortName,
		CompanyCode:              args.ShortName,
		Secret:                   args.Secret,
		EgressPriceNibblinsPerGB: routing.Nibblin(args.EgressPrice),
	}

	if err := s.Storage.AddSeller(ctx, seller); err != nil {
		err = fmt.Errorf("AddSeller() error: %w", err)
		s.Logger.Log("err", err)
		return err
	}

	return nil
}

type AddSellerArgs struct {
	Seller routing.Seller
}

type AddSellerReply struct{}

func (s *OpsService) AddSeller(r *http.Request, args *AddSellerArgs, reply *AddSellerReply) error {
	ctx, cancelFunc := context.WithDeadline(r.Context(), time.Now().Add(10*time.Second))
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
	ctx, cancelFunc := context.WithDeadline(r.Context(), time.Now().Add(10*time.Second))
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

	ctx, cancelFunc := context.WithDeadline(r.Context(), time.Now().Add(10*time.Second))
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
	HexID               string                `json:"hexID"`
	DatacenterHexID     string                `json:"datacenterHexID"`
	BillingSupplier     string                `json:"billingSupplier"`
	SignedID            int64                 `json:"signed_id"`
	Name                string                `json:"name"`
	Addr                string                `json:"addr"`
	InternalAddr        string                `json:"internalAddr"`
	Latitude            float64               `json:"latitude"`
	Longitude           float64               `json:"longitude"`
	NICSpeedMbps        int32                 `json:"nicSpeedMbps"`
	IncludedBandwidthGB int32                 `json:"includedBandwidthGB"`
	State               string                `json:"state"`
	ManagementAddr      string                `json:"management_addr"`
	SSHUser             string                `json:"ssh_user"`
	SSHPort             int64                 `json:"ssh_port"`
	MaxSessionCount     uint32                `json:"maxSessionCount"`
	PublicKey           string                `json:"public_key"`
	Version             string                `json:"relay_version"`
	SellerName          string                `json:"seller_name"`
	EgressPriceOverride routing.Nibblin       `json:"egressPriceOverride"`
	MRC                 routing.Nibblin       `json:"monthlyRecurringChargeNibblins"`
	Overage             routing.Nibblin       `json:"overage"`
	BWRule              routing.BandWidthRule `json:"bandwidthRule"`
	ContractTerm        int32                 `json:"contractTerm"`
	StartDate           time.Time             `json:"startDate"`
	EndDate             time.Time             `json:"endDate"`
	Type                routing.MachineType   `json:"machineType"`
	Notes               string                `json:"notes"`
	DatabaseID          int64
	DatacenterID        uint64
}

func (s *OpsService) Relays(r *http.Request, args *RelaysArgs, reply *RelaysReply) error {

	for _, r := range s.Storage.Relays(r.Context()) {
		relay := relay{
			ID:                  r.ID,
			HexID:               fmt.Sprintf("%016x", r.ID),
			DatacenterHexID:     fmt.Sprintf("%016x", r.Datacenter.ID),
			BillingSupplier:     r.BillingSupplier,
			SignedID:            r.SignedID,
			Name:                r.Name,
			Addr:                r.Addr.String(),
			Latitude:            float64(r.Datacenter.Location.Latitude),
			Longitude:           float64(r.Datacenter.Location.Longitude),
			NICSpeedMbps:        r.NICSpeedMbps,
			IncludedBandwidthGB: r.IncludedBandwidthGB,
			ManagementAddr:      r.ManagementAddr,
			SSHUser:             r.SSHUser,
			SSHPort:             r.SSHPort,
			State:               r.State.String(),
			PublicKey:           base64.StdEncoding.EncodeToString(r.PublicKey),
			MaxSessionCount:     r.MaxSessions,
			SellerName:          r.Seller.Name,
			EgressPriceOverride: r.EgressPriceOverride,
			MRC:                 r.MRC,
			Overage:             r.Overage,
			BWRule:              r.BWRule,
			ContractTerm:        r.ContractTerm,
			StartDate:           r.StartDate,
			EndDate:             r.EndDate,
			Type:                r.Type,
			Notes:               r.Notes,
			Version:             r.Version,
			DatabaseID:          r.DatabaseID,
		}

		if addrStr := r.InternalAddr.String(); addrStr != ":0" {
			relay.InternalAddr = addrStr
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
	Relay routing.Relay `json:"Relay"`
}

type AddRelayReply struct{}

func (s *OpsService) AddRelay(r *http.Request, args *AddRelayArgs, reply *AddRelayReply) error {
	ctx, cancelFunc := context.WithDeadline(r.Context(), time.Now().Add(10*time.Second))
	defer cancelFunc()

	if err := s.Storage.AddRelay(ctx, args.Relay); err != nil {
		err = fmt.Errorf("AddRelay() error: %w", err)
		s.Logger.Log("err", err)
		return err
	}

	return nil
}

type JSAddRelayArgs struct {
	Name                string `json:"name"`
	Addr                string `json:"addr"`
	InternalAddr        string `json:"internal_addr"`
	PublicKey           string `json:"public_key"`
	SellerID            string `json:"seller"`
	DatacenterID        string `json:"datacenter"`
	NICSpeedMbps        int64  `json:"nicSpeedMbps"`
	IncludedBandwidthGB int64  `json:"includedBandwidthGB"`
	ManagementAddr      string `json:"management_addr"`
	SSHUser             string `json:"ssh_user"`
	SSHPort             int64  `json:"ssh_port"`
	MaxSessions         int64  `json:"max_sessions"`
	EgressPriceOverride int64  `json:"egressPriceOverride"`
	MRC                 int64  `json:"monthlyRecurringChargeNibblins"`
	Overage             int64  `json:"overage"`
	BWRule              int64  `json:"bandwidthRule"`
	ContractTerm        int64  `json:"contractTerm"`
	StartDate           string `json:"startDate"`
	EndDate             string `json:"endDate"`
	Type                int64  `json:"machineType"`
	Notes               string `json:"notes"`
	BillingSupplier     string `json:"billingSupplier"`
	Version             string `json:"relay_version"`
}

type JSAddRelayReply struct{}

func (s *OpsService) JSAddRelay(r *http.Request, args *JSAddRelayArgs, reply *JSAddRelayReply) error {
	ctx, cancelFunc := context.WithDeadline(r.Context(), time.Now().Add(10*time.Second))
	defer cancelFunc()

	addr, err := net.ResolveUDPAddr("udp", args.Addr)
	if err != nil {
		s.Logger.Log("err", err)
		return err
	}

	dcID, err := strconv.ParseUint(args.DatacenterID, 16, 64)
	if err != nil {
		s.Logger.Log("err", err)
		return err
	}

	// seller is not required for SQL Storer AddRelay() method
	var datacenter routing.Datacenter
	if datacenter, err = s.Storage.Datacenter(r.Context(), dcID); err != nil {
		err = fmt.Errorf("Datacenter() error: %w", err)
		s.Logger.Log("err", err)
		return err
	}

	publicKey, err := base64.StdEncoding.DecodeString(args.PublicKey)
	if err != nil {
		err = fmt.Errorf("could not decode base64 public key %s: %v", args.PublicKey, err)
		s.Logger.Log("err", err)
	}

	rid := crypto.HashID(args.Addr)
	relay := routing.Relay{
		ID:                  rid,
		Name:                args.Name,
		Addr:                *addr,
		PublicKey:           publicKey,
		Datacenter:          datacenter,
		NICSpeedMbps:        int32(args.NICSpeedMbps),
		IncludedBandwidthGB: int32(args.IncludedBandwidthGB),
		State:               routing.RelayStateEnabled,
		ManagementAddr:      args.ManagementAddr,
		SSHUser:             args.SSHUser,
		SSHPort:             args.SSHPort,
		MaxSessions:         uint32(args.MaxSessions),
		EgressPriceOverride: routing.Nibblin(args.EgressPriceOverride),
		MRC:                 routing.Nibblin(args.MRC),
		Overage:             routing.Nibblin(args.Overage),
		BWRule:              routing.BandWidthRule(args.BWRule),
		ContractTerm:        int32(args.ContractTerm),
		Type:                routing.MachineType(args.Type),
		Notes:               args.Notes,
		BillingSupplier:     args.BillingSupplier,
		Version:             args.Version,
	}

	var internalAddr *net.UDPAddr
	if args.InternalAddr != "" {
		internalAddr, err = net.ResolveUDPAddr("udp", args.InternalAddr)
		if err != nil {
			s.Logger.Log("err", err)
			return err
		}
		relay.InternalAddr = *internalAddr
	}

	if args.StartDate != "" {
		startDate, err := time.Parse("2006-01-02", args.StartDate)
		if err != nil {
			s.Logger.Log("err", err)
			return err
		}
		relay.StartDate = startDate
	}

	if args.EndDate != "" {
		endDate, err := time.Parse("2006-01-02", args.EndDate)
		if err != nil {
			s.Logger.Log("err", err)
			return err
		}
		relay.EndDate = endDate
	}

	if err := s.Storage.AddRelay(ctx, relay); err != nil {
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
	relay, err := s.Storage.Relay(r.Context(), args.RelayID)
	if err != nil {
		err = fmt.Errorf("RemoveRelay() Storage.Relay error: %w", err)
		s.Logger.Log("err", err)
		return err
	}

	// Rather than actually removing the relay from firestore, just
	// rename it and set it to the decomissioned state
	relay.State = routing.RelayStateDecommissioned

	// want: “$(relayname)-removed-$(date-time-of-removal)”
	shortDate := time.Now().Format("2006-01-02")
	shortTime := time.Now().Format("15:04:05")
	relay.Name = fmt.Sprintf("%s-removed-%s-%s", relay.Name, shortDate, shortTime)

	relay.Addr = net.UDPAddr{} // clear the address to 0 when removed

	if err = s.Storage.SetRelay(r.Context(), relay); err != nil {
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

	relay, err := s.Storage.Relay(r.Context(), args.RelayID)
	if err != nil {
		err = fmt.Errorf("RelayNameUpdate() Storage.Relay error: %w", err)
		s.Logger.Log("err", err)
		return err
	}

	relay.Name = args.RelayName
	if err = s.Storage.SetRelay(r.Context(), relay); err != nil {
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

	relay, err := s.Storage.Relay(r.Context(), args.RelayID)
	if err != nil {
		err = fmt.Errorf("RelayStateUpdate() Storage.Relay error: %w", err)
		s.Logger.Log("err", err)
		return err
	}

	relay.State = args.RelayState
	if err = s.Storage.SetRelay(r.Context(), relay); err != nil {
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

	relay, err := s.Storage.Relay(r.Context(), args.RelayID)
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

	if err = s.Storage.SetRelay(r.Context(), relay); err != nil {
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

	relay, err := s.Storage.Relay(r.Context(), args.RelayID)
	if err != nil {
		err = fmt.Errorf("RelayNICSpeedUpdate() Relay error: %w", err)
		s.Logger.Log("err", err)
		return err
	}

	relay.NICSpeedMbps = int32(args.RelayNICSpeed)
	if err = s.Storage.SetRelay(r.Context(), relay); err != nil {
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
	if datacenter, err = s.Storage.Datacenter(r.Context(), arg.ID); err != nil {
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
	HexID        string  `json:"hexID"`
	ID           uint64  `json:"id"`
	SignedID     int64   `json:"signed_id"`
	Latitude     float32 `json:"latitude"`
	Longitude    float32 `json:"longitude"`
	SupplierName string  `json:"supplierName"`
}

func (s *OpsService) Datacenters(r *http.Request, args *DatacentersArgs, reply *DatacentersReply) error {
	for _, d := range s.Storage.Datacenters(r.Context()) {
		reply.Datacenters = append(reply.Datacenters, datacenter{
			Name:      d.Name,
			HexID:     fmt.Sprintf("%016x", d.ID),
			ID:        d.ID,
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
	ctx, cancelFunc := context.WithDeadline(r.Context(), time.Now().Add(10*time.Second))
	defer cancelFunc()

	if err := s.Storage.AddDatacenter(ctx, args.Datacenter); err != nil {
		err = fmt.Errorf("AddDatacenter() error: %w", err)
		s.Logger.Log("err", err)
		return err
	}

	return nil
}

type JSAddDatacenterArgs struct {
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	SellerID  string  `json:"sellerID"`
}

type JSAddDatacenterReply struct{}

func (s *OpsService) JSAddDatacenter(r *http.Request, args *JSAddDatacenterArgs, reply *JSAddDatacenterReply) error {
	ctx, cancelFunc := context.WithDeadline(r.Context(), time.Now().Add(10*time.Second))
	defer cancelFunc()

	dcID := crypto.HashID(args.Name)

	seller, err := s.Storage.Seller(r.Context(), args.SellerID)
	if err != nil {
		s.Logger.Log("err", err)
		return err
	}

	datacenter := routing.Datacenter{
		Name: args.Name,
		ID:   dcID,
		Location: routing.Location{
			Latitude:  float32(args.Latitude),
			Longitude: float32(args.Longitude),
		},
		SellerID: seller.DatabaseID,
	}

	if err := s.Storage.AddDatacenter(ctx, datacenter); err != nil {
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
	ctx, cancelFunc := context.WithDeadline(r.Context(), time.Now().Add(10*time.Second))
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

	dcm := s.Storage.ListDatacenterMaps(r.Context(), args.DatacenterID)

	var replySlice []DatacenterMapsFull
	for _, dcMap := range dcm {
		buyer, err := s.Storage.Buyer(r.Context(), dcMap.BuyerID)
		if err != nil {
			err = fmt.Errorf("ListDatacenterMaps() could not parse buyer: %w", err)
			s.Logger.Log("err", err)
			return err
		}
		datacenter, err := s.Storage.Datacenter(r.Context(), dcMap.DatacenterID)
		if err != nil {
			err = fmt.Errorf("ListDatacenterMaps() could not parse datacenter: %w", err)
			s.Logger.Log("err", err)
			return err
		}

		company, err := s.Storage.Customer(r.Context(), buyer.CompanyCode)
		if err != nil {
			err = fmt.Errorf("ListDatacenterMaps() failed to find buyer company: %w", err)
			s.Logger.Log("err", err)
			return err
		}

		dcmFull := DatacenterMapsFull{
			DatacenterName: datacenter.Name,
			DatacenterID:   fmt.Sprintf("%016x", dcMap.DatacenterID),
			BuyerName:      company.Name,
			BuyerID:        fmt.Sprintf("%016x", dcMap.BuyerID),
		}

		replySlice = append(replySlice, dcmFull)
	}

	reply.DatacenterMaps = replySlice

	return nil
}

// type RelayMetadataArgs struct {
// 	Relay routing.Relay
// }

// type RelayMetadataReply struct {
// 	Ok           bool
// 	ErrorMessage string
// }

// func (s *OpsService) RelayMetadata(r *http.Request, args *RelayMetadataArgs, reply *RelayMetadataReply) error {

// 	err := s.Storage.SetRelayMetadata(r.Context(), args.Relay)
// 	if err != nil {
// 		return err // TODO detail
// 	}

// 	return nil
// }

type RouteSelectionArgs struct {
	SourceRelays      []string `json:"src_relays"`
	DestinationRelays []string `json:"dest_relays"`
	RTT               float64  `json:"rtt"`
	RouteHash         uint64   `json:"route_hash"`
}

type RouteSelectionReply struct {
	Routes []routing.Route `json:"routes"`
}

type CheckRelayIPAddressArgs struct {
	IpAddress string `json:"ipAddress"`
	HexID     string `json:"hexID"`
}

type CheckRelayIPAddressReply struct {
	Valid bool `json:"valid"`
}

// CheckRelayIPAddress is used by the Admin tool when recommissioning a relay to ensure the
// selected IP address "matches" the HexID (which was derived from its original IP address).
func (s *OpsService) CheckRelayIPAddress(r *http.Request, args *CheckRelayIPAddressArgs, reply *CheckRelayIPAddressReply) error {

	internalIDFromHexID, err := strconv.ParseUint(args.HexID, 16, 64)
	if err != nil {
		reply.Valid = false
		return err
	}

	addr, err := net.ResolveUDPAddr("udp", args.IpAddress)
	if err != nil {
		reply.Valid = false
		return err
	}

	internalIdFromIpAddress := crypto.HashID(addr.String())
	if internalIDFromHexID != internalIdFromIpAddress {
		reply.Valid = false
		return err
	}

	reply.Valid = true
	return nil
}

type UpdateRelayArgs struct {
	RelayID    uint64      `json:"relayID"`    // used by next tool
	HexRelayID string      `json:"hexRelayID"` // used by javascript clients
	Field      string      `json:"field"`
	Value      interface{} `json:"value"`
}

type UpdateRelayReply struct{}

func (s *OpsService) UpdateRelay(r *http.Request, args *UpdateRelayArgs, reply *UpdateRelayReply) error {

	relayID := args.RelayID
	var err error
	if args.HexRelayID != "" {
		relayID, err = strconv.ParseUint(args.HexRelayID, 16, 64)
		if err != nil {
			err = fmt.Errorf("UpdateRelay() failed to parse HexRelayID %s: %w", args.HexRelayID, err)
			s.Logger.Log("err", err)
			return err
		}
	}
	err = s.Storage.UpdateRelay(r.Context(), relayID, args.Field, args.Value)
	if err != nil {
		err = fmt.Errorf("UpdateRelay() failed to modify relay record for field %s with value %v: %w", args.Field, args.Value, err)
		s.Logger.Log("err", err)
		return err
	}
	return nil
}

type GetRelayArgs struct {
	RelayID uint64
}

type GetRelayReply struct {
	Relay relay
}

func (s *OpsService) GetRelay(r *http.Request, args *GetRelayArgs, reply *GetRelayReply) error {
	routingRelay, err := s.Storage.Relay(r.Context(), args.RelayID)
	if err != nil {
		err = fmt.Errorf("Error retrieving relay ID %016x: %v", args.RelayID, err)
		s.Logger.Log("err", err)
		return err
	}

	relay := relay{
		ID:                  routingRelay.ID,
		SignedID:            routingRelay.SignedID,
		Name:                routingRelay.Name,
		Addr:                routingRelay.Addr.String(),
		InternalAddr:        routingRelay.InternalAddr.String(),
		Latitude:            float64(routingRelay.Datacenter.Location.Latitude),
		Longitude:           float64(routingRelay.Datacenter.Location.Longitude),
		NICSpeedMbps:        routingRelay.NICSpeedMbps,
		IncludedBandwidthGB: routingRelay.IncludedBandwidthGB,
		ManagementAddr:      routingRelay.ManagementAddr,
		SSHUser:             routingRelay.SSHUser,
		SSHPort:             routingRelay.SSHPort,
		State:               routingRelay.State.String(),
		PublicKey:           base64.StdEncoding.EncodeToString(routingRelay.PublicKey),
		MaxSessionCount:     routingRelay.MaxSessions,
		SellerName:          routingRelay.Seller.Name,
		EgressPriceOverride: routingRelay.EgressPriceOverride,
		MRC:                 routingRelay.MRC,
		Overage:             routingRelay.Overage,
		BWRule:              routingRelay.BWRule,
		ContractTerm:        routingRelay.ContractTerm,
		StartDate:           routingRelay.StartDate,
		EndDate:             routingRelay.EndDate,
		Type:                routingRelay.Type,
		Notes:               routingRelay.Notes,
		DatabaseID:          routingRelay.DatabaseID,
		DatacenterID:        routingRelay.Datacenter.ID,
		BillingSupplier:     routingRelay.BillingSupplier,
		Version:             routingRelay.Version,
	}

	reply.Relay = relay

	return nil
}

type ModifyRelayFieldArgs struct {
	RelayID uint64
	Field   string
	Value   string
}

type ModifyRelayFieldReply struct{}

func (s *OpsService) ModifyRelayField(r *http.Request, args *ModifyRelayFieldArgs, reply *ModifyRelayFieldReply) error {

	if middleware.VerifyAllRoles(r, middleware.AnonymousRole) {
		return nil
	}

	// sort out the value type here (comes from the next tool and javascript UI as a string)
	switch args.Field {
	// sent to storer as float64
	case "NICSpeedMbps", "IncludedBandwidthGB", "ContractTerm", "SSHPort", "MaxSessions":
		newfloat, err := strconv.ParseFloat(args.Value, 64)
		if err != nil {
			return fmt.Errorf("Value: %v is not a valid numeric type", args.Value)
		}
		err = s.Storage.UpdateRelay(r.Context(), args.RelayID, args.Field, newfloat)
		if err != nil {
			err = fmt.Errorf("UpdateRelay() error updating field for relay %016x: %v", args.RelayID, err)
			level.Error(s.Logger).Log("err", err)
			return err
		}

	// net.UDPAddr, time.Time - all sent to storer as strings
	case "Addr", "InternalAddr", "ManagementAddr", "SSHUser", "StartDate", "EndDate", "BillingSupplier", "Version":
		err := s.Storage.UpdateRelay(r.Context(), args.RelayID, args.Field, args.Value)
		if err != nil {
			err = fmt.Errorf("UpdateRelay() error updating field for relay %016x: %v", args.RelayID, err)
			level.Error(s.Logger).Log("err", err)
			return err
		}

	// relay.PublicKey
	case "PublicKey":
		newPublicKey := string(args.Value)
		err := s.Storage.UpdateRelay(r.Context(), args.RelayID, args.Field, newPublicKey)
		if err != nil {
			err = fmt.Errorf("UpdateRelay() error updating field for relay %016x: %v", args.RelayID, err)
			level.Error(s.Logger).Log("err", err)
			return err
		}

	// routing.RelayState
	case "State":

		state, err := routing.ParseRelayState(args.Value)
		if err != nil {
			err := fmt.Errorf("value '%s' is not a valid relay state", args.Value)
			level.Error(s.Logger).Log("err", err)
			return err
		}
		err = s.Storage.UpdateRelay(r.Context(), args.RelayID, args.Field, float64(state))
		if err != nil {
			err = fmt.Errorf("UpdateRelay() error updating field for relay %016x: %v", args.RelayID, err)
			level.Error(s.Logger).Log("err", err)
			return err
		}

	// nibblins (received as USD, sent to storer as float64)
	case "EgressPriceOverride", "MRC", "Overage":
		newValue, err := strconv.ParseFloat(args.Value, 64)
		if err != nil {
			err = fmt.Errorf("value '%s' is not a valid float64 port number: %v", args.Value, err)
			level.Error(s.Logger).Log("err", err)
			return err
		}
		err = s.Storage.UpdateRelay(r.Context(), args.RelayID, args.Field, newValue)
		if err != nil {
			err = fmt.Errorf("UpdateRelay() error updating field for relay %016x: %v", args.RelayID, err)
			level.Error(s.Logger).Log("err", err)
			return err
		}

	// routing.BandwidthRule
	case "BWRule":

		bwRule, err := routing.ParseBandwidthRule(args.Value)
		if err != nil {
			err := fmt.Errorf("value '%s' is not a valid bandwidth rule", args.Value)
			level.Error(s.Logger).Log("err", err)
			return err
		}
		err = s.Storage.UpdateRelay(r.Context(), args.RelayID, args.Field, float64(bwRule))
		if err != nil {
			err = fmt.Errorf("UpdateRelay() error updating field for relay %016x: %v", args.RelayID, err)
			level.Error(s.Logger).Log("err", err)
			return err
		}

		// routing.MachineType
	case "Type":

		machineType, err := routing.ParseMachineType(args.Value)
		if err != nil {
			err := fmt.Errorf("value '%s' is not a valid machine type", args.Value)
			level.Error(s.Logger).Log("err", err)
			return err
		}
		err = s.Storage.UpdateRelay(r.Context(), args.RelayID, args.Field, float64(machineType))
		if err != nil {
			err = fmt.Errorf("UpdateRelay() error updating field for relay %016x: %v", args.RelayID, err)
			level.Error(s.Logger).Log("err", err)
			return err
		}

	default:
		return fmt.Errorf("Field '%v' does not exist on the Relay type", args.Field)
	}

	return nil
}

type UpdateCustomerArgs struct {
	CustomerID string `json:"customerCode"`
	Field      string `json:"field"`
	Value      string `json:"value"`
}

type UpdateCustomerReply struct{}

func (s *OpsService) UpdateCustomer(r *http.Request, args *UpdateCustomerArgs, reply *UpdateCustomerReply) error {
	if middleware.VerifyAllRoles(r, middleware.AnonymousRole) {
		return nil
	}

	// sort out the value type here (comes from the next tool and javascript UI as a string)
	switch args.Field {
	case "Name", "AutomaticSigninDomains":
		err := s.Storage.UpdateCustomer(r.Context(), args.CustomerID, args.Field, args.Value)
		if err != nil {
			err = fmt.Errorf("UpdateCustomer() error updating record for customer %s: %v", args.CustomerID, err)
			level.Error(s.Logger).Log("err", err)
			return err
		}

	default:
		return fmt.Errorf("Field '%v' does not exist (or is not editable) on the Customer type", args.Field)
	}

	return nil
}

type RemoveCustomerArgs struct {
	CustomerCode string `json:"customerCode"`
}

type RemoveCustomerReply struct{}

func (s *OpsService) RemoveCustomer(r *http.Request, args *RemoveCustomerArgs, reply *RemoveCustomerReply) error {
	ctx, cancelFunc := context.WithDeadline(r.Context(), time.Now().Add(10*time.Second))
	defer cancelFunc()

	return s.Storage.RemoveCustomer(ctx, args.CustomerCode)
}

type UpdateSellerArgs struct {
	SellerID string `json:"shortName"`
	Field    string `json:"field"`
	Value    string `json:"value"`
}

type UpdateSellerReply struct{}

func (s *OpsService) UpdateSeller(r *http.Request, args *UpdateSellerArgs, reply *UpdateSellerReply) error {
	if middleware.VerifyAllRoles(r, middleware.AnonymousRole) {
		return nil
	}

	// sort out the value type here (comes from the next tool and javascript UI as a string)
	switch args.Field {
	case "ShortName":
		err := s.Storage.UpdateSeller(r.Context(), args.SellerID, args.Field, args.Value)
		if err != nil {
			err = fmt.Errorf("UpdateSeller() error updating record for seller %s: %v", args.SellerID, err)
			level.Error(s.Logger).Log("err", err)
			return err
		}
	case "Secret":
		secret, err := strconv.ParseBool(args.Value)
		if err != nil {
			err = fmt.Errorf("UpdateSeller() value '%s' is not a valid Secret/boolean: %v", args.Value, err)
			level.Error(s.Logger).Log("err", err)
			return err
		}
		err = s.Storage.UpdateSeller(r.Context(), args.SellerID, args.Field, secret)
		if err != nil {
			err = fmt.Errorf("UpdateSeller() error updating record for seller %s: %v", args.SellerID, err)
			level.Error(s.Logger).Log("err", err)
			return err
		}
	case "EgressPrice", "IngressPrice":
		newValue, err := strconv.ParseFloat(args.Value, 64)
		if err != nil {
			err = fmt.Errorf("value '%s' is not a valid price: %v", args.Value, err)
			level.Error(s.Logger).Log("err", err)
			return err
		}

		if args.Field == "EgressPrice" {
			args.Field = "EgressPriceNibblinsPerGB"
		} else {
			args.Field = "IngressPriceNibblinsPerGB"
		}
		err = s.Storage.UpdateSeller(r.Context(), args.SellerID, args.Field, newValue)
		if err != nil {
			err = fmt.Errorf("UpdateSeller() error updating field for seller %s: %v", args.SellerID, err)
			level.Error(s.Logger).Log("err", err)
			return err
		}

	default:
		return fmt.Errorf("Field '%v' does not exist (or is not editable) on the Seller type", args.Field)
	}

	return nil
}

type ResetSellerEgressPriceOverrideArgs struct {
	SellerID string `json:"shortName"`
	Field    string `json:"field"`
}

type ResetSellerEgressPriceOverrideReply struct{}

func (s *OpsService) ResetSellerEgressPriceOverride(r *http.Request, args *ResetSellerEgressPriceOverrideArgs, reply *ResetSellerEgressPriceOverrideReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		return nil
	}

	// Iterate through relays and reset egress price override for this seller's relays
	relays := s.Storage.Relays(r.Context())

	for _, relay := range relays {

		switch args.Field {
		case "EgressPriceOverride":
			if relay.Seller.ShortName == args.SellerID {
				err := s.Storage.UpdateRelay(r.Context(), relay.ID, args.Field, float64(0))
				if err != nil {
					err = fmt.Errorf("ResetSellerEgressPriceOverride() error updating %s for seller %s: %v", args.Field, args.SellerID, err)
					level.Error(s.Logger).Log("err", err)
					return err
				}
			}
		default:
			return fmt.Errorf("Field '%s' is not a valid Relay type for resetting seller egress price override", args.Field)
		}
	}

	return nil
}

type UpdateDatacenterArgs struct {
	HexDatacenterID string      `json:"hexDatacenterID"`
	Field           string      `json:"field"`
	Value           interface{} `json:"value"`
}

type UpdateDatacenterReply struct{}

func (s *OpsService) UpdateDatacenter(r *http.Request, args *UpdateDatacenterArgs, reply *UpdateDatacenterReply) error {
	if middleware.VerifyAllRoles(r, middleware.AnonymousRole) {
		return nil
	}

	dcID, err := strconv.ParseUint(args.HexDatacenterID, 16, 64)
	if err != nil {
		level.Error(s.Logger).Log("err", err)
		return err
	}

	switch args.Field {
	case "Latitude", "Longitude":
		newValue := float32(args.Value.(float64))
		err := s.Storage.UpdateDatacenter(r.Context(), dcID, args.Field, newValue)
		if err != nil {
			err = fmt.Errorf("UpdateDatacenter() error updating record for customer %s: %v", args.HexDatacenterID, err)
			level.Error(s.Logger).Log("err", err)
			return err
		}

	default:
		return fmt.Errorf("Field '%v' does not exist (or is not editable) on the Datacenter type", args.Field)
	}

	return nil
}
