package storage

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"sync"
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

	datacenters map[uint64]*routing.Datacenter
	relays      map[uint64]*routing.Relay
	buyers      map[uint64]*routing.Buyer

	datacenterMutex sync.Mutex
	relayMutex      sync.Mutex
	buyerMutex      sync.Mutex
}

type buyer struct {
	ID        int    `firestore:"sdkVersion3PublicKeyId"`
	Name      string `firestore:"name"`
	Active    bool   `firestore:"active"`
	Live      bool   `firestore:"isLiveCustomer"`
	PublicKey []byte `firestore:"sdkVersion3PublicKeyData"`
}

type seller struct {
	Name                       string `firestore:"name"`
	PricePublicIngressNibblins int64  `firestore:"pricePublicIngressNibblins"`
	PricePublicEgressNibblins  int64  `firestore:"pricePublicEgressNibblins"`
}

type relay struct {
	Name               string                 `firestore:"displayName"`
	Address            string                 `firestore:"publicAddress"`
	PublicKey          []byte                 `firestore:"publicKey"`
	UpdateKey          []byte                 `firestore:"updateKey"`
	NICSpeedMbps       float32                `firestore:"nicSpeedMbps"`
	IncludedBandwithGB float32                `firestore:"includedBandwidthGB"`
	Datacenter         *firestore.DocumentRef `firestore:"datacenter"`
	Seller             *firestore.DocumentRef `firestore:"seller"`
}

type datacenter struct {
	Name      string  `firestore:"name"`
	Enabled   bool    `firestore:"enabled"`
	Latitude  float32 `firestore:"latitude"`
	Longitude float32 `firestore:"longitude"`
}

type routingRulesSettings struct {
	DisplayName                  string  `firestore:"displayName"`
	EnvelopeKbpsUp               int64   `firestore:"envelopeKbpsUp"`
	EnvelopeKbpsDown             int64   `firestore:"envelopeKbpsDown"`
	Mode                         int64   `firestore:"mode"`
	MaxPricePerGBNibblins        int64   `firestore:"maxPricePerGBNibblins"`
	AcceptableLatency            float32 `firestore:"acceptableLatency"`
	RTTEpsilon                   float32 `firestore:"rttRouteSwitch"`
	RTTThreshold                 float32 `firestore:"rttThreshold"`
	RTTHysteresis                float32 `firestore:"rttHysteresis"`
	RTTVeto                      float32 `firestore:"rttVeto"`
	EnableYouOnlyLiveOnce        bool    `firestore:"youOnlyLiveOnce"`
	EnablePacketLossSafety       bool    `firestore:"packetLossSafety"`
	EnableMultipathForPacketLoss bool    `firestore:"packetLossMultipath"`
	EnableMultipathForJitter     bool    `firestore:"jitterMultipath"`
	EnableMultipathForRTT        bool    `firestore:"rttMultipath"`
	EnableABTest                 bool    `firestore:"abTest"`
}

func (fs *Firestore) Relay(id uint64) (*routing.Relay, bool) {
	fs.relayMutex.Lock()
	defer fs.relayMutex.Unlock()

	b, found := fs.relays[id]
	return b, found
}

func (fs *Firestore) Relays() []routing.Relay {
	fs.relayMutex.Lock()
	defer fs.relayMutex.Unlock()

	var relays []routing.Relay
	for _, relay := range fs.relays {
		relays = append(relays, *relay)
	}

	return relays
}

func (fs *Firestore) Datacenters() []routing.Datacenter {
	fs.datacenterMutex.Lock()
	defer fs.datacenterMutex.Unlock()

	var datacenters []routing.Datacenter
	for _, datacenter := range fs.datacenters {
		datacenters = append(datacenters, *datacenter)
	}

	return datacenters
}

func (fs *Firestore) Buyer(id uint64) (*routing.Buyer, bool) {
	fs.buyerMutex.Lock()
	defer fs.buyerMutex.Unlock()

	b, found := fs.buyers[id]
	return b, found
}

// SyncLoop is a helper method that calls Sync
func (fs *Firestore) SyncLoop(ctx context.Context, c <-chan time.Time) {
	if err := fs.Sync(ctx); err != nil {
		level.Error(fs.Logger).Log("during", "SyncLoop", "err", err)
	}

	for {
		select {
		case <-c:
			if err := fs.Sync(ctx); err != nil {
				level.Error(fs.Logger).Log("during", "SyncLoop", "err", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

// Sync fetches relays and buyers from Firestore and places copies into local caches
func (fs *Firestore) Sync(ctx context.Context) error {
	var outerErr error
	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		if err := fs.syncRelays(ctx); err != nil {
			outerErr = fmt.Errorf("failed to sync relays: %v", err)
		}
		wg.Done()
	}()

	go func() {
		if err := fs.syncBuyers(ctx); err != nil {
			outerErr = fmt.Errorf("failed to sync buyers: %v", err)
		}
		wg.Done()
	}()

	go func() {
		if err := fs.syncDatacenters(ctx); err != nil {
			outerErr = fmt.Errorf("failed to sync buyers: %v", err)
		}
		wg.Done()
	}()

	wg.Wait()

	return outerErr
}

func (fs *Firestore) syncDatacenters(ctx context.Context) error {
	datacenters := make(map[uint64]*routing.Datacenter)

	ddocs := fs.Client.Collection("Datacenter").Documents(ctx)
	for {
		ddoc, err := ddocs.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("unknown error: %v", err)
		}

		var d datacenter
		err = ddoc.DataTo(&d)
		if err != nil {
			return fmt.Errorf("failed to marshal document: %v", err)
		}

		datacenters[crypto.HashID(d.Name)] = &routing.Datacenter{
			Name:    d.Name,
			Enabled: d.Enabled,
			Location: routing.Location{
				Latitude:  float64(d.Latitude),
				Longitude: float64(d.Longitude),
			},
		}
	}

	fs.datacenterMutex.Lock()
	fs.datacenters = datacenters
	fs.datacenterMutex.Unlock()

	level.Info(fs.Logger).Log("during", "syncDatacenters", "num", len(fs.datacenters))

	return nil
}

func (fs *Firestore) syncRelays(ctx context.Context) error {
	relays := make(map[uint64]*routing.Relay)

	rdocs := fs.Client.Collection("Relay").Documents(ctx)
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

		// Default to the relay public key, but if that isn't in firestore
		// then use the old update key for compatibility
		publicKey := r.PublicKey
		if publicKey == nil {
			publicKey = r.UpdateKey
		}

		relay := routing.Relay{
			ID:   rid,
			Name: r.Name,
			Addr: net.UDPAddr{
				IP:   net.ParseIP(host),
				Port: int(iport),
			},
			PublicKey:           publicKey,
			NICSpeedMbps:        int(r.NICSpeedMbps),
			IncludedBandwidthGB: int(r.IncludedBandwithGB),
		}

		// Get datacenter
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
			ID:      crypto.HashID(d.Name),
			Name:    d.Name,
			Enabled: d.Enabled,
		}

		relay.Datacenter = datacenter
		relay.Latitude = float64(d.Latitude)
		relay.Longitude = float64(d.Longitude)

		// Get seller
		sdoc, err := r.Seller.Get(ctx)
		if err != nil {
			return fmt.Errorf("failed to get document: %v", err)
		}

		var s seller
		err = sdoc.DataTo(&s)
		if err != nil {
			return fmt.Errorf("failed to marshal document: %v", err)
		}

		seller := routing.Seller{
			ID:                sdoc.Ref.ID,
			Name:              s.Name,
			IngressPriceCents: convertNibblinsToCents(s.PricePublicIngressNibblins),
			EgressPriceCents:  convertNibblinsToCents(s.PricePublicEgressNibblins),
		}

		relay.Seller = seller

		// add populated relay to list
		relays[rid] = &relay
	}

	fs.relayMutex.Lock()
	fs.relays = relays
	fs.relayMutex.Unlock()

	level.Info(fs.Logger).Log("during", "syncRelays", "num", len(fs.relays))

	return nil
}

func (fs *Firestore) syncBuyers(ctx context.Context) error {
	buyers := make(map[uint64]*routing.Buyer)

	bdocs := fs.Client.Collection("Buyer").Documents(ctx)
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
		rrs, err := fs.getRoutingRulesSettingsForBuyerID(ctx, bdoc.Ref.ID)
		if err != nil {
			level.Debug(fs.Logger).Log("msg", fmt.Sprintf("using default route rules for buyer %v", bdoc.Ref.ID), "err", err)
		}

		buyers[uint64(b.ID)] = &routing.Buyer{
			ID:                   uint64(b.ID),
			Name:                 b.Name,
			Active:               b.Active,
			Live:                 b.Live,
			PublicKey:            b.PublicKey,
			RoutingRulesSettings: rrs,
		}
	}

	fs.buyerMutex.Lock()
	fs.buyers = buyers
	fs.buyerMutex.Unlock()

	level.Info(fs.Logger).Log("during", "syncBuyers", "num", len(fs.buyers))

	return nil
}

func (fs *Firestore) getRoutingRulesSettingsForBuyerID(ctx context.Context, ID string) (routing.RoutingRulesSettings, error) {
	// Comment below taken from old backend, at least attempting to explain why we need to append _0 (no existing entries have suffixes other than _0)
	// "Must be of the form '<buyer key>_<tag id>'. The buyer key can be found by looking at the ID under Buyer; it should be something like 763IMDH693HLsr2LGTJY. The tag ID should be 0 (for default) or the fnv64a hash of the tag the customer is using. Therefore this value should look something like: 763IMDH693HLsr2LGTJY_0. This value can not be changed after the entity is created."
	routeShaderID := ID + "_0"

	// Set up our return value with default settings, which will be used if no settings found for buyer or other errors are encountered
	rrs := routing.DefaultRoutingRulesSettings

	// Attempt to get route shader for buyer (sadly not linked by actual reference in prod so have to fetch it ourselves using buyer ID + "_0" which happens to match)
	rsDoc, err := fs.Client.Collection("RouteShader").Doc(routeShaderID).Get(ctx)
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
	rrs.EnvelopeKbpsUp = tempRRS.EnvelopeKbpsUp
	rrs.EnvelopeKbpsDown = tempRRS.EnvelopeKbpsDown
	rrs.Mode = tempRRS.Mode
	rrs.MaxCentsPerGB = convertNibblinsToCents(tempRRS.MaxPricePerGBNibblins)
	rrs.AcceptableLatency = tempRRS.AcceptableLatency
	rrs.RTTEpsilon = tempRRS.RTTEpsilon
	rrs.RTTThreshold = tempRRS.RTTThreshold
	rrs.RTTHysteresis = tempRRS.RTTHysteresis
	rrs.RTTVeto = tempRRS.RTTVeto
	rrs.EnableYouOnlyLiveOnce = tempRRS.EnableYouOnlyLiveOnce
	rrs.EnablePacketLossSafety = tempRRS.EnablePacketLossSafety
	rrs.EnableMultipathForPacketLoss = tempRRS.EnableMultipathForPacketLoss
	rrs.EnableMultipathForJitter = tempRRS.EnableMultipathForJitter
	rrs.EnableMultipathForRTT = tempRRS.EnableMultipathForRTT
	rrs.EnableABTest = tempRRS.EnableABTest

	return rrs, nil
}

// Note: Nibblins is a made up unit in the old backend presumably to deal with floating point issues. 1000000000 Niblins = $0.01 USD
func convertNibblinsToCents(nibblins int64) uint64 {
	return uint64(nibblins) / 1e9
}
