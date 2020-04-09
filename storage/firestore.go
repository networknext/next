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

	datacenters map[uint64]routing.Datacenter
	relays      map[uint64]routing.Relay
	buyers      map[uint64]routing.Buyer

	datacenterMutex sync.RWMutex
	relayMutex      sync.RWMutex
	buyerMutex      sync.RWMutex
}

type buyer struct {
	ID        uint64 `firestore:"sdkVersion3PublicKeyId"`
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
	NICSpeedMbps       int                    `firestore:"nicSpeedMbps"`
	IncludedBandwithGB int                    `firestore:"includedBandwidthGB"`
	Datacenter         *firestore.DocumentRef `firestore:"datacenter"`
	Seller             *firestore.DocumentRef `firestore:"seller"`
	ManagementAddress  string                 `firestore:"managementAddress"`
	SSHUser            string                 `firestore:"sshUser"`
	SSHPort            int64                  `firestore:"sshPort"`
	State              routing.RelayState     `firestore:"state"`
	StateUpdateTime    time.Time              `firestore:"stateUpdateTime"`
}

type datacenter struct {
	Name      string  `firestore:"name"`
	Enabled   bool    `firestore:"enabled"`
	Latitude  float64 `firestore:"latitude"`
	Longitude float64 `firestore:"longitude"`
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

func (fs *Firestore) Buyer(id uint64) (routing.Buyer, error) {
	fs.buyerMutex.RLock()
	defer fs.buyerMutex.RUnlock()

	b, found := fs.buyers[id]
	if !found {
		return routing.Buyer{}, fmt.Errorf("buyer with id %d not found in firestore", id)
	}

	return b, nil
}

func (fs *Firestore) Buyers() []routing.Buyer {
	fs.buyerMutex.RLock()
	defer fs.buyerMutex.RUnlock()

	var buyers []routing.Buyer
	for _, buyer := range fs.buyers {
		buyers = append(buyers, buyer)
	}

	return buyers
}

func (fs *Firestore) AddBuyer(ctx context.Context, b routing.Buyer) error {
	newBuyerData := buyer{
		ID:        b.ID,
		Name:      b.Name,
		Active:    b.Active,
		Live:      b.Live,
		PublicKey: b.PublicKey,
	}

	// Add the buyer in remote storage
	ref, _, err := fs.Client.Collection("Buyer").Add(ctx, newBuyerData)
	if err != nil {
		return err
	}

	// Add the buyer's routing rules settings to re4mote storage
	fs.createRouteRulesSerttingsForBuyerID(ctx, ref.ID, b.Name, b.RoutingRulesSettings)

	// Add the buyer in cached storage
	fs.buyerMutex.Lock()
	fs.buyers[b.ID] = b
	fs.buyerMutex.Unlock()

	return nil
}

func (fs *Firestore) RemoveBuyer(ctx context.Context, id uint64) error {
	// Check if the buyer exists
	fs.buyerMutex.RLock()
	_, ok := fs.buyers[id]
	fs.buyerMutex.RUnlock()

	if !ok {
		return fmt.Errorf("buyer with ID %d doesn't exist", id)
	}

	bdocs := fs.Client.Collection("Buyer").Documents(ctx)
	for bdoc, err := bdocs.Next(); err != iterator.Done; {
		if err != nil {
			return fmt.Errorf("unknown error: %v", err)
		}

		// Unmarshal the buyer in firestore to see if it's the buyer we want to delete
		var buyerInRemoteStorage buyer
		err = bdoc.DataTo(&buyerInRemoteStorage)
		if err != nil {
			return fmt.Errorf("failed to unmarshal document: %v", err)
		}

		if buyerInRemoteStorage.ID == id {
			// Delete the buyer in remote storage
			if _, err := bdoc.Ref.Delete(ctx); err != nil {
				return err
			}

			// Delete the buyer in cached storage
			fs.buyerMutex.Lock()
			delete(fs.buyers, id)
			fs.buyerMutex.Unlock()
			return nil
		}
	}

	return fmt.Errorf("could not remove buyer with id %d in firestore", id)
}

func (fs *Firestore) SetBuyer(ctx context.Context, b routing.Buyer) error {
	// Get a copy of the buyer in cached storage
	fs.buyerMutex.RLock()
	buyerInCachedStorage, ok := fs.buyers[b.ID]
	fs.buyerMutex.RUnlock()

	if !ok {
		return fmt.Errorf("buyer with ID %d doesn't exist", b.ID)
	}

	// Set the data to update the buyer with
	newBuyerData := buyer{
		Name:      b.Name,
		Active:    b.Active,
		Live:      b.Live,
		PublicKey: b.PublicKey,
	}

	// Loop through all buyers in firestore
	bdocs := fs.Client.Collection("Buyer").Documents(ctx)
	for bdoc, err := bdocs.Next(); err != iterator.Done; {
		if err != nil {
			return fmt.Errorf("unknown error: %v", err)
		}

		// Unmarshal the buyer in firestore to see if it's the buyer we want to update
		var buyerInRemoteStorage buyer
		err = bdoc.DataTo(&buyerInRemoteStorage)
		if err != nil {
			return fmt.Errorf("failed to unmarshal document: %v", err)
		}

		// If the buyer is the one we want to update, update it with the new data
		if buyerInRemoteStorage.ID == b.ID {
			// Update the buyer in firestore
			if _, err := bdoc.Ref.Set(ctx, newBuyerData, firestore.MergeAll); err != nil {
				return err
			}

			// Update the buyer's routing rules settings in firestore
			if err := fs.setRoutingRulesSettingsForBuyerID(ctx, bdoc.Ref.ID, b.Name, b.RoutingRulesSettings); err != nil {
				return err
			}

			// Update the cached version
			buyerInCachedStorage = b

			fs.buyerMutex.Lock()
			fs.buyers[b.ID] = buyerInCachedStorage
			fs.buyerMutex.Unlock()

			return nil
		}
	}

	return fmt.Errorf("could not update buyer with id %d in firestore", b.ID)
}

func (fs *Firestore) Relay(id uint64) (routing.Relay, error) {
	fs.relayMutex.RLock()
	defer fs.relayMutex.RUnlock()

	relay, found := fs.relays[id]
	if !found {
		return routing.Relay{}, fmt.Errorf("relay with id %d not found in firestore", id)
	}

	return relay, nil

}

func (fs *Firestore) Relays() []routing.Relay {
	fs.relayMutex.RLock()
	defer fs.relayMutex.RUnlock()

	var relays []routing.Relay
	for _, relay := range fs.relays {
		relays = append(relays, relay)
	}

	return relays
}

func (fs *Firestore) AddRelay(ctx context.Context, r routing.Relay) error {
	var sellerRef *firestore.DocumentRef
	var datacenterRef *firestore.DocumentRef

	// Loop through all sellers in firestore
	sdocs := fs.Client.Collection("Seller").Documents(ctx)
	for sdoc, err := sdocs.Next(); err != iterator.Done; {
		if err != nil {
			return fmt.Errorf("unknown error: %v", err)
		}

		// If the seller is the one associated with this relay, set the relay's seller reference
		if sdoc.Ref.ID == r.Seller.ID {
			sellerRef = sdoc.Ref
			break
		}
	}

	if sellerRef == nil {
		return fmt.Errorf("unknown seller with ID %s - be sure to create the seller in firestore first", r.Seller.ID)
	}

	// Loop through all datacenters in firestore
	ddocs := fs.Client.Collection("Datacenter").Documents(ctx)
	for ddoc, err := ddocs.Next(); err != iterator.Done; {
		if err != nil {
			return fmt.Errorf("unknown error: %v", err)
		}

		// Unmarshal the datacenter so we can check if the ID matches the datacenter associated with this relay
		var d datacenter
		err = ddoc.DataTo(&d)
		if err != nil {
			return fmt.Errorf("failed to unmarshal document: %v", err)
		}

		// If the datacenter is the one associated with this relay, set the relay's datacenter reference
		if crypto.HashID(d.Name) == r.Datacenter.ID {
			datacenterRef = ddoc.Ref
			break
		}
	}

	if datacenterRef == nil {
		return fmt.Errorf("unknown datacenter with ID %d - be sure to create the datacenter in firestore first", r.Datacenter.ID)
	}

	newRelayData := relay{
		Name:               r.Name,
		Address:            r.Addr.String(),
		PublicKey:          r.PublicKey,
		UpdateKey:          r.PublicKey,
		NICSpeedMbps:       r.NICSpeedMbps,
		IncludedBandwithGB: r.IncludedBandwidthGB,
		Datacenter:         datacenterRef,
		Seller:             sellerRef,
		ManagementAddress:  r.ManagementAddr,
		SSHUser:            r.SSHUser,
		SSHPort:            r.SSHPort,
		State:              r.State,
		StateUpdateTime:    r.LastUpdateTime,
	}

	// Add the relay in remote storage
	if _, _, err := fs.Client.Collection("Relay").Add(ctx, newRelayData); err != nil {
		return err
	}

	// Add the relay in cached storage
	fs.relayMutex.Lock()
	fs.relays[r.ID] = r
	fs.relayMutex.Unlock()

	return nil
}

func (fs *Firestore) RemoveRelay(ctx context.Context, id uint64) error {
	// Check if the relay exists
	fs.relayMutex.RLock()
	_, ok := fs.relays[id]
	fs.relayMutex.RUnlock()

	if !ok {
		return fmt.Errorf("relay with ID %d doesn't exist", id)
	}

	rdocs := fs.Client.Collection("Relay").Documents(ctx)
	for rdoc, err := rdocs.Next(); err != iterator.Done; {
		if err != nil {
			return fmt.Errorf("unknown error: %v", err)
		}

		// Unmarshal the relay in firestore to see if it's the relay we want to delete
		var relayInRemoteStorage relay
		err = rdoc.DataTo(&relayInRemoteStorage)
		if err != nil {
			return fmt.Errorf("failed to unmarshal document: %v", err)
		}

		rid := crypto.HashID(relayInRemoteStorage.Address)
		if rid == id {
			// Delete the relay in remote storage
			if _, err := rdoc.Ref.Delete(ctx); err != nil {
				return err
			}

			// Delete the relay in cached storage
			fs.relayMutex.Lock()
			delete(fs.relays, id)
			fs.relayMutex.Unlock()
			return nil
		}
	}

	return fmt.Errorf("could not remove relay with id %d in firestore", id)
}

// Only relay state is updated in firestore for now
func (fs *Firestore) SetRelay(ctx context.Context, r routing.Relay) error {
	// Get a copy of the relay in cached storage
	fs.relayMutex.RLock()
	relayInCachedStorage, ok := fs.relays[r.ID]
	fs.relayMutex.RUnlock()

	if !ok {
		return fmt.Errorf("relay with ID %d doesn't exist", r.ID)
	}

	// Set the data to update the relay with
	stateUpdateTime := time.Now()
	newRelayData := relay{
		State:           r.State,
		StateUpdateTime: stateUpdateTime,
	}

	// Loop through all relays in firestore
	rdocs := fs.Client.Collection("Relay").Documents(ctx)
	for rdoc, err := rdocs.Next(); err != iterator.Done; {
		if err != nil {
			return fmt.Errorf("unknown error: %v", err)
		}

		// Unmarshal the relay in firestore to see if it's the relay we want to update
		var relayInRemoteStorage relay
		err = rdoc.DataTo(&relayInRemoteStorage)
		if err != nil {
			return fmt.Errorf("failed to unmarshal document: %v", err)
		}

		// If the relay is the one we want to update, update it with the new data
		rid := crypto.HashID(relayInRemoteStorage.Address)
		if rid == r.ID {
			// Update the relay in firestore
			if _, err := rdoc.Ref.Set(ctx, newRelayData, firestore.MergeAll); err != nil {
				return err
			}

			// Update the cached version
			relayInCachedStorage.State = newRelayData.State
			relayInCachedStorage.LastUpdateTime = newRelayData.StateUpdateTime

			fs.relayMutex.Lock()
			fs.relays[r.ID] = relayInCachedStorage
			fs.relayMutex.Unlock()

			return nil
		}
	}

	return fmt.Errorf("could not update relay with id %d in firestore", r.ID)
}

func (fs *Firestore) Datacenter(id uint64) (routing.Datacenter, error) {
	fs.datacenterMutex.RLock()
	defer fs.datacenterMutex.RUnlock()

	d, found := fs.datacenters[id]
	if !found {
		return routing.Datacenter{}, fmt.Errorf("datacenter with id %d not found in firestore", id)
	}

	return d, nil
}

func (fs *Firestore) Datacenters() []routing.Datacenter {
	fs.datacenterMutex.RLock()
	defer fs.datacenterMutex.RUnlock()

	var datacenters []routing.Datacenter
	for _, datacenter := range fs.datacenters {
		datacenters = append(datacenters, datacenter)
	}

	return datacenters
}

func (fs *Firestore) AddDatacenter(ctx context.Context, d routing.Datacenter) error {
	newDatacenterData := datacenter{
		Name:      d.Name,
		Enabled:   d.Enabled,
		Latitude:  d.Location.Latitude,
		Longitude: d.Location.Longitude,
	}

	// Add the datacenter in remote storage
	_, _, err := fs.Client.Collection("Datacenter").Add(ctx, newDatacenterData)
	if err != nil {
		return err
	}

	// Add the datacenter in cached storage
	fs.datacenterMutex.Lock()
	fs.datacenters[d.ID] = d
	fs.datacenterMutex.Unlock()

	return nil
}

func (fs *Firestore) RemoveDatacenter(ctx context.Context, id uint64) error {
	// Check if the datacenter exists
	fs.datacenterMutex.RLock()
	_, ok := fs.datacenters[id]
	fs.datacenterMutex.RUnlock()

	if !ok {
		return fmt.Errorf("datacenter with ID %d doesn't exist", id)
	}

	ddocs := fs.Client.Collection("Datacenter").Documents(ctx)
	for ddoc, err := ddocs.Next(); err != iterator.Done; {
		if err != nil {
			return fmt.Errorf("unknown error: %v", err)
		}

		// Unmarshal the datanceter in firestore to see if it's the datacenter we want to delete
		var datacenterInRemoteStorage datacenter
		err = ddoc.DataTo(&datacenterInRemoteStorage)
		if err != nil {
			return fmt.Errorf("failed to unmarshal document: %v", err)
		}

		if crypto.HashID(datacenterInRemoteStorage.Name) == id {
			// Delete the datacenter in remote storage
			if _, err := ddoc.Ref.Delete(ctx); err != nil {
				return err
			}

			// Delete the datacenter in cached storage
			fs.datacenterMutex.Lock()
			delete(fs.datacenters, id)
			fs.datacenterMutex.Unlock()
			return nil
		}
	}

	return fmt.Errorf("could not remove datacenter with id %d in firestore", id)
}

func (fs *Firestore) SetDatacenter(ctx context.Context, d routing.Datacenter) error {
	// Get a copy of the datacenter in cached storage
	fs.datacenterMutex.RLock()
	datacenterInCachedStorage, ok := fs.datacenters[d.ID]
	fs.datacenterMutex.RUnlock()

	if !ok {
		return fmt.Errorf("datacenter with ID %d doesn't exist", d.ID)
	}

	// Set the data to update the datacenter with
	newDatacenterData := datacenter{
		Name:      d.Name,
		Enabled:   d.Enabled,
		Latitude:  d.Location.Latitude,
		Longitude: d.Location.Longitude,
	}

	// Loop through all datacenters in firestore
	bdocs := fs.Client.Collection("Datacenter").Documents(ctx)
	for bdoc, err := bdocs.Next(); err != iterator.Done; {
		if err != nil {
			return fmt.Errorf("unknown error: %v", err)
		}

		// Unmarshal the datacenter in firestore to see if it's the datacenter we want to update
		var datacenterInRemoteStorage datacenter
		err = bdoc.DataTo(&datacenterInRemoteStorage)
		if err != nil {
			return fmt.Errorf("failed to unmarshal document: %v", err)
		}

		// If the datacenter is the one we want to update, update it with the new data
		if crypto.HashID(datacenterInRemoteStorage.Name) == d.ID {
			// Update the datacenter in firestore
			if _, err := bdoc.Ref.Set(ctx, newDatacenterData, firestore.MergeAll); err != nil {
				return err
			}

			// Update the cached version
			datacenterInCachedStorage = d

			fs.datacenterMutex.Lock()
			fs.datacenters[d.ID] = datacenterInCachedStorage
			fs.datacenterMutex.Unlock()

			return nil
		}
	}

	return fmt.Errorf("could not update datacenter with id %d in firestore", d.ID)
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
	datacenters := make(map[uint64]routing.Datacenter)

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
			return fmt.Errorf("failed to unmarshal document: %v", err)
		}

		datacenters[crypto.HashID(d.Name)] = routing.Datacenter{
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
	relays := make(map[uint64]routing.Relay)

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
			return fmt.Errorf("failed to unmarshal document: %v", err)
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
			NICSpeedMbps:        r.NICSpeedMbps,
			IncludedBandwidthGB: r.IncludedBandwithGB,
			ManagementAddr:      r.ManagementAddress,
			SSHUser:             r.SSHUser,
			SSHPort:             r.SSHPort,
			State:               r.State,
			LastUpdateTime:      r.StateUpdateTime,
		}

		// Get datacenter
		ddoc, err := r.Datacenter.Get(ctx)
		if err != nil {
			return fmt.Errorf("failed to get document: %v", err)
		}

		var d datacenter
		err = ddoc.DataTo(&d)
		if err != nil {
			return fmt.Errorf("failed to unmarshal document: %v", err)
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
			return fmt.Errorf("failed to unmarshal document: %v", err)
		}

		seller := routing.Seller{
			ID:                sdoc.Ref.ID,
			Name:              s.Name,
			IngressPriceCents: convertNibblinsToCents(s.PricePublicIngressNibblins),
			EgressPriceCents:  convertNibblinsToCents(s.PricePublicEgressNibblins),
		}

		relay.Seller = seller

		// add populated relay to list
		relays[rid] = relay
	}

	fs.relayMutex.Lock()
	fs.relays = relays
	fs.relayMutex.Unlock()

	level.Info(fs.Logger).Log("during", "syncRelays", "num", len(fs.relays))

	return nil
}

func (fs *Firestore) syncBuyers(ctx context.Context) error {
	buyers := make(map[uint64]routing.Buyer)

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
			return fmt.Errorf("failed to unmarshal document: %v", err)
		}

		if !b.Active {
			continue
		}

		// Attempt to get routing rules settings for buyer (acceptable to fallback to default settings if none defined)
		rrs, err := fs.getRoutingRulesSettingsForBuyerID(ctx, bdoc.Ref.ID)
		if err != nil {
			level.Debug(fs.Logger).Log("msg", fmt.Sprintf("using default route rules for buyer %v", bdoc.Ref.ID), "err", err)
		}

		buyers[b.ID] = routing.Buyer{
			ID:                   b.ID,
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

func (fs *Firestore) createRouteRulesSerttingsForBuyerID(ctx context.Context, ID string, name string, rrs routing.RoutingRulesSettings) error {
	// Comment below taken from old backend, at least attempting to explain why we need to append _0 (no existing entries have suffixes other than _0)
	// "Must be of the form '<buyer key>_<tag id>'. The buyer key can be found by looking at the ID under Buyer; it should be something like 763IMDH693HLsr2LGTJY. The tag ID should be 0 (for default) or the fnv64a hash of the tag the customer is using. Therefore this value should look something like: 763IMDH693HLsr2LGTJY_0. This value can not be changed after the entity is created."
	routeShaderID := ID + "_0"

	// Convert RoutingRulesSettings struct to firestore version
	rrsFirestore := routingRulesSettings{
		DisplayName:                  name,
		EnvelopeKbpsUp:               rrs.EnvelopeKbpsUp,
		EnvelopeKbpsDown:             rrs.EnvelopeKbpsDown,
		Mode:                         rrs.Mode,
		MaxPricePerGBNibblins:        convertCentsToNibblins(rrs.MaxCentsPerGB),
		AcceptableLatency:            rrs.AcceptableLatency,
		RTTEpsilon:                   rrs.RTTEpsilon,
		RTTThreshold:                 rrs.RTTThreshold,
		RTTHysteresis:                rrs.RTTHysteresis,
		RTTVeto:                      rrs.RTTVeto,
		EnableYouOnlyLiveOnce:        rrs.EnableYouOnlyLiveOnce,
		EnablePacketLossSafety:       rrs.EnablePacketLossSafety,
		EnableMultipathForPacketLoss: rrs.EnableMultipathForPacketLoss,
		EnableMultipathForJitter:     rrs.EnableMultipathForJitter,
		EnableMultipathForRTT:        rrs.EnableMultipathForRTT,
		EnableABTest:                 rrs.EnableABTest,
	}

	// Attempt to create route shader for buyer
	rsDocRef := fs.Client.Collection("RouteShader").NewDoc()
	rsDocRef.ID = routeShaderID

	_, err := rsDocRef.Create(ctx, rrsFirestore)
	return err
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

	// Unmarshal into our firestore struct
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

func (fs *Firestore) setRoutingRulesSettingsForBuyerID(ctx context.Context, ID string, name string, rrs routing.RoutingRulesSettings) error {
	// Comment below taken from old backend, at least attempting to explain why we need to append _0 (no existing entries have suffixes other than _0)
	// "Must be of the form '<buyer key>_<tag id>'. The buyer key can be found by looking at the ID under Buyer; it should be something like 763IMDH693HLsr2LGTJY. The tag ID should be 0 (for default) or the fnv64a hash of the tag the customer is using. Therefore this value should look something like: 763IMDH693HLsr2LGTJY_0. This value can not be changed after the entity is created."
	routeShaderID := ID + "_0"

	// Convert RoutingRulesSettings struct to firestore version
	rrsFirestore := routingRulesSettings{
		DisplayName:                  name,
		EnvelopeKbpsUp:               rrs.EnvelopeKbpsUp,
		EnvelopeKbpsDown:             rrs.EnvelopeKbpsDown,
		Mode:                         rrs.Mode,
		MaxPricePerGBNibblins:        convertCentsToNibblins(rrs.MaxCentsPerGB),
		AcceptableLatency:            rrs.AcceptableLatency,
		RTTEpsilon:                   rrs.RTTEpsilon,
		RTTThreshold:                 rrs.RTTThreshold,
		RTTHysteresis:                rrs.RTTHysteresis,
		RTTVeto:                      rrs.RTTVeto,
		EnableYouOnlyLiveOnce:        rrs.EnableYouOnlyLiveOnce,
		EnablePacketLossSafety:       rrs.EnablePacketLossSafety,
		EnableMultipathForPacketLoss: rrs.EnableMultipathForPacketLoss,
		EnableMultipathForJitter:     rrs.EnableMultipathForJitter,
		EnableMultipathForRTT:        rrs.EnableMultipathForRTT,
		EnableABTest:                 rrs.EnableABTest,
	}

	// Attempt to set route shader for buyer
	_, err := fs.Client.Collection("RouteShader").Doc(routeShaderID).Set(ctx, rrsFirestore, firestore.MergeAll)
	return err
}

// Note: Nibblins is a made up unit in the old backend presumably to deal with floating point issues. 1000000000 Niblins = $0.01 USD
func convertNibblinsToCents(nibblins int64) uint64 {
	return uint64(nibblins) / 1e9
}

// Note: Nibblins is a made up unit in the old backend presumably to deal with floating point issues. 1000000000 Niblins = $0.01 USD
func convertCentsToNibblins(cents uint64) int64 {
	return int64(cents * 1e9)
}
