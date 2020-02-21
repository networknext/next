package storage

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/routing"
	"google.golang.org/api/iterator"
)

type Firestore struct {
	Client *firestore.Client
	Logger log.Logger

	relays map[uint64]*routing.Relay
	buyers map[uint64]*routing.Buyer
}

type buyer struct {
	ID        int    `firestore:"sdkVersion3PublicKeyId"`
	Name      string `firestore:"name"`
	Active    bool   `firestore:"active"`
	Live      bool   `firestore:"isLiveCustomer"`
	PublicKey []byte `firestore:"sdkVersion3PublicKeyData"`
}

type relay struct {
	Address    string                 `firestore:"publicAddress"`
	PublicKey  []byte                 `firestore:"updateKey"`
	Datacenter *firestore.DocumentRef `firestore:"datacenter"`
}

type datacenter struct {
	Name    string `firestore:"name"`
	Enabled bool   `firestore:"enabled"`
}

type routingRulesSettings struct {
	DisplayName           string  `firestore:"displayName"`
	EnvelopeKbpsUp        int64   `firestore:"envelopeKbpsUp"`
	EnvelopeKbpsDown      int64   `firestore:"envelopeKbpsDown"`
	Mode                  int64   `firestore:"mode"`
	MaxPricePerGBNibblins int64   `firestore:"maxPricePerGBNibblins"`
	AcceptableLatency     float32 `firestore:"acceptableLatency"`
	RttRouteSwitch        float32 `firestore:"rttRouteSwitch"`
	RttThreshold          float32 `firestore:"rttThreshold"`
	RttHysteresis         float32 `firestore:"rttHysteresis"`
	RttVeto               float32 `firestore:"rttVeto"`
	YouOnlyLiveOnce       bool    `firestore:"youOnlyLiveOnce"`
	PacketLossSafety      bool    `firestore:"packetLossSafety"`
	PacketLossMultipath   bool    `firestore:"packetLossMultipath"`
	JitterMultipath       bool    `firestore:"jitterMultipath"`
	RttMultipath          bool    `firestore:"rttMultipath"`
	AbTest                bool    `firestore:"abTest"`
}

func (s *Firestore) Relay(id uint64) (*routing.Relay, bool) {
	b, found := s.relays[id]
	return b, found
}

func (s *Firestore) Buyer(id uint64) (*routing.Buyer, bool) {
	b, found := s.buyers[id]
	return b, found
}

// SyncLoop is a helper method that calls Sync
func (s *Firestore) SyncLoop(ctx context.Context, c <-chan time.Time) {
	if err := s.Sync(ctx); err != nil {
		s.Logger.Log("during", "SyncLoop", "err", err)
	}

	for {
		select {
		case <-c:
			if err := s.Sync(ctx); err != nil {
				s.Logger.Log("during", "SyncLoop", "err", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

// Sync fetches relays and buyers from Firestore and places copies into local caches
func (s *Firestore) Sync(ctx context.Context) error {
	if err := s.syncRelays(ctx); err != nil {
		return fmt.Errorf("failed to sync relays: %v", err)
	}

	if err := s.syncBuyers(ctx); err != nil {
		return fmt.Errorf("failed to sync buyers: %v", err)
	}

	return nil
}

func (s *Firestore) syncRelays(ctx context.Context) error {
	s.relays = make(map[uint64]*routing.Relay)

	rdocs := s.Client.Collection("Relay").Documents(ctx)
	for {
		rdoc, err := rdocs.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("unknown error: %v", err)
		}

		var r relay
		err = rdoc.DataTo(&r)
		if err != nil {
			return fmt.Errorf("failed to marshal document: %v", err)
		}

		rid := crypto.HashID(r.Address)

		host, port, err := net.SplitHostPort(r.Address)
		if err != nil {
			return fmt.Errorf("failed to split host and port: %v", err)
		}
		iport, err := strconv.ParseInt(port, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to convert port to int: %v", err)
		}

		relay := routing.Relay{
			Addr: net.UDPAddr{
				IP:   net.ParseIP(host),
				Port: int(iport),
			},
			PublicKey: []byte(r.PublicKey),
		}

		ddoc, err := r.Datacenter.Get(ctx)
		if err != nil {
			return fmt.Errorf("failed to get document: %v", err)
		}

		var d datacenter
		err = ddoc.DataTo(&d)
		if err != nil {
			return fmt.Errorf("failed to marshal document: %v", err)
		}

		if !d.Enabled {
			continue
		}

		datacenter := routing.Datacenter{
			ID:   crypto.HashID(d.Name),
			Name: d.Name,
		}

		relay.Datacenter = datacenter

		s.relays[rid] = &relay
	}

	level.Debug(s.Logger).Log("during", "syncRelays", "num", len(s.relays))

	return nil
}

func (s *Firestore) syncBuyers(ctx context.Context) error {
	s.buyers = make(map[uint64]*routing.Buyer)

	bdocs := s.Client.Collection("Buyer").Documents(ctx)
	for {
		bdoc, err := bdocs.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("unknown error: %v", err)
		}

		var b buyer
		err = bdoc.DataTo(&b)
		if err != nil {
			return fmt.Errorf("failed to marshal document: %v", err)
		}

		if !b.Active {
			continue
		}

		// Attempt to get routing rules settings for buyer (acceptable to fallback to default settings if none defined)
		rrs, err := s.GetRoutingRulesSettingsForBuyerID(ctx, bdoc.Ref.ID)
		if err != nil {
			level.Debug(s.Logger).Log("msg", fmt.Sprintf("using default route rules for buyer %v", bdoc.Ref.ID), "err", err)
		}

		s.buyers[uint64(b.ID)] = &routing.Buyer{
			ID:                   uint64(b.ID),
			Name:                 b.Name,
			Active:               b.Active,
			Live:                 b.Live,
			PublicKey:            b.PublicKey,
			RoutingRulesSettings: rrs,
		}
	}

	level.Debug(s.Logger).Log("during", "syncBuyers", "num", len(s.buyers))

	return nil
}

func (s *Firestore) GetRoutingRulesSettingsForBuyerID(ctx context.Context, ID string) (routing.RoutingRulesSettings, error) {
	// Comment below taken from old backend, at least attempting to explain why we need to append _0 (no existing entries have suffixes other than _0)
	// "Must be of the form '<buyer key>_<tag id>'. The buyer key can be found by looking at the ID under Buyer; it should be something like 763IMDH693HLsr2LGTJY. The tag ID should be 0 (for default) or the fnv64a hash of the tag the customer is using. Therefore this value should look something like: 763IMDH693HLsr2LGTJY_0. This value can not be changed after the entity is created."
	routeShaderID := ID + "_0"

	// Set up our return value with default settings, which will be used if no settings found for buyer or other errors are encountered
	rrs := routing.GetDefaultRoutingRulesSettings()

	// Attempt to get route shader for buyer (sadly not linked by actual reference in prod so have to fetch it ourselves using buyer ID + "_0" which happens to match)
	rsDoc, err := s.Client.Collection("RouteShader").Doc(routeShaderID).Get(ctx)
	if err != nil {
		return rrs, err
	}

	// Marshal into our firestore struct
	var tempRRS routingRulesSettings
	err = rsDoc.DataTo(&tempRRS)
	if err != nil {
		return rrs, err
	}

	// If successful, convert into routing.Buyer version and return it
	rrs.DisplayName = tempRRS.DisplayName
	rrs.EnvelopeKbpsUp = tempRRS.EnvelopeKbpsUp
	rrs.EnvelopeKbpsDown = tempRRS.EnvelopeKbpsDown
	rrs.Mode = tempRRS.Mode
	rrs.MaxCentsPerGB = tempRRS.MaxPricePerGBNibblins / 1e9 // Note: Nibblins is a made up unit in the old backend presumably to deal with floating point issues. 1000000000 Niblins = $0.01 USD
	rrs.AcceptableLatency = tempRRS.AcceptableLatency
	rrs.RttRouteSwitch = tempRRS.RttRouteSwitch
	rrs.RttThreshold = tempRRS.RttThreshold
	rrs.RttHysteresis = tempRRS.RttHysteresis
	rrs.RttVeto = tempRRS.RttVeto
	rrs.YouOnlyLiveOnce = tempRRS.YouOnlyLiveOnce
	rrs.PacketLossSafety = tempRRS.PacketLossSafety
	rrs.PacketLossMultipath = tempRRS.PacketLossMultipath
	rrs.JitterMultipath = tempRRS.JitterMultipath
	rrs.RttMultipath = tempRRS.RttMultipath
	rrs.AbTest = tempRRS.AbTest

	return rrs, nil
}
