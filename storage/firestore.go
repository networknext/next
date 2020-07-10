package storage

import (
	"context"
	"fmt"
	"net"
	"sort"
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

	datacenters    map[uint64]routing.Datacenter
	relays         map[uint64]routing.Relay
	buyers         map[uint64]routing.Buyer
	sellers        map[string]routing.Seller
	datacenterMaps map[uint64]routing.DatacenterMap

	datacenterMutex    sync.RWMutex
	relayMutex         sync.RWMutex
	buyerMutex         sync.RWMutex
	sellerMutex        sync.RWMutex
	datacenterMapMutex sync.RWMutex
}

type customer struct {
	Name   string                 `firestore:"name"`
	Domain string                 `firestore:"automaticSigninDomain"`
	Active bool                   `firestore:"active"`
	Buyer  *firestore.DocumentRef `firestore:"buyer"`
	Seller *firestore.DocumentRef `firestore:"seller"`
}

type buyer struct {
	ID        int64  `firestore:"sdkVersion3PublicKeyId"`
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
	NICSpeedMbps       int64                  `firestore:"nicSpeedMbps"`
	IncludedBandwithGB int64                  `firestore:"includedBandwidthGB"`
	Datacenter         *firestore.DocumentRef `firestore:"datacenter"`
	Seller             *firestore.DocumentRef `firestore:"seller"`
	ManagementAddress  string                 `firestore:"managementAddress"`
	SSHUser            string                 `firestore:"sshUser"`
	SSHPort            int64                  `firestore:"sshPort"`
	State              routing.RelayState     `firestore:"state"`
	LastUpdateTime     time.Time              `firestore:"lastUpdateTime"`
	MaxSessions        int32                  `firestore:"maxSessions"`
}

type datacenter struct {
	Name      string  `firestore:"name"`
	AliasName string  `firestore:"name_alias"`
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
	EnableTryBeforeYouBuy        bool    `firestore:"tryBeforeYouBuy"`
	TryBeforeYouBuyMaxSlices     int8    `firestore:"tryBeforeYouBuyMaxSlices"`
	SelectionPercentage          int64   `firestore:"selectionPercentage"`
}

type FirestoreError struct {
	err error
}

func (e *FirestoreError) Error() string {
	return fmt.Sprintf("unknown Firestore error: %v", e.err)
}

func NewFirestore(ctx context.Context, gcpProjectID string, logger log.Logger) (*Firestore, error) {
	client, err := firestore.NewClient(ctx, gcpProjectID)
	if err != nil {
		return nil, err
	}

	return &Firestore{
		Client:      client,
		Logger:      logger,
		datacenters: make(map[uint64]routing.Datacenter),
		relays:      make(map[uint64]routing.Relay),
		buyers:      make(map[uint64]routing.Buyer),
		sellers:     make(map[string]routing.Seller),
	}, nil
}

func (fs *Firestore) Buyer(id uint64) (routing.Buyer, error) {
	fs.buyerMutex.RLock()
	defer fs.buyerMutex.RUnlock()

	b, found := fs.buyers[id]
	if !found {
		return routing.Buyer{}, &DoesNotExistError{resourceType: "buyer", resourceRef: fmt.Sprintf("%x", id)}
	}

	return b, nil
}

func (fs *Firestore) BuyerWithDomain(domain string) (routing.Buyer, error) {
	fs.buyerMutex.RLock()
	defer fs.buyerMutex.RUnlock()

	for _, buyer := range fs.buyers {
		if buyer.Domain == domain {
			return buyer, nil
		}
	}

	return routing.Buyer{}, &DoesNotExistError{resourceType: "buyer", resourceRef: domain}
}

func (fs *Firestore) Buyers() []routing.Buyer {
	fs.buyerMutex.RLock()
	defer fs.buyerMutex.RUnlock()

	var buyers []routing.Buyer
	for _, buyer := range fs.buyers {
		buyers = append(buyers, buyer)
	}

	sort.Slice(buyers, func(i int, j int) bool { return buyers[i].ID < buyers[j].ID })
	return buyers
}

func (fs *Firestore) AddBuyer(ctx context.Context, b routing.Buyer) error {
	newBuyerData := buyer{
		ID:        int64(b.ID),
		Name:      b.Name,
		Active:    b.Active,
		Live:      b.Live,
		PublicKey: b.PublicKey,
	}

	// Add the buyer in remote storage
	ref, _, err := fs.Client.Collection("Buyer").Add(ctx, newBuyerData)
	if err != nil {
		return &FirestoreError{err: err}
	}

	// Add the buyer's routing rules settings to remote storage
	if err := fs.createRouteRulesSettingsForBuyerID(ctx, ref.ID, b.Name, b.RoutingRulesSettings); err != nil {
		return &FirestoreError{err: err}
	}

	// Check if a customer already exists for this buyer
	var customerFound bool

	cdocs := fs.Client.Collection("Customer").Documents(ctx)
	defer cdocs.Stop()
	for {
		cdoc, err := cdocs.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return &FirestoreError{err: err}
		}

		// Unmarshal the customer in firestore to see if it's the customer we want to add the buyer to
		var customerInRemoteStorage customer
		err = cdoc.DataTo(&customerInRemoteStorage)
		if err != nil {
			return &UnmarshalError{err: err}
		}

		if customerInRemoteStorage.Name == b.Name && customerInRemoteStorage.Buyer == nil {
			customerFound = true

			customerInRemoteStorage.Buyer = ref

			// Update the customer references
			if _, err := cdoc.Ref.Set(ctx, customerInRemoteStorage); err != nil {
				return &FirestoreError{err: err}
			}

			break
		}
	}

	if !customerFound {
		// Customer was not found, so make a new one
		newCustomerData := customer{
			Name:   b.Name,
			Active: b.Active,
			Domain: b.Domain,
			Buyer:  ref,
		}

		// Create the customer object in remote storage
		if _, _, err = fs.Client.Collection("Customer").Add(ctx, newCustomerData); err != nil {
			return &FirestoreError{err: err}
		}
	}

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
		return &DoesNotExistError{resourceType: "buyer", resourceRef: fmt.Sprintf("%x", id)}
	}

	bdocs := fs.Client.Collection("Buyer").Documents(ctx)
	defer bdocs.Stop()
	for {
		bdoc, err := bdocs.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return &FirestoreError{err: err}
		}

		// Unmarshal the buyer in firestore to see if it's the buyer we want to delete
		var buyerInRemoteStorage buyer
		err = bdoc.DataTo(&buyerInRemoteStorage)
		if err != nil {
			return &UnmarshalError{err: err}
		}

		if uint64(buyerInRemoteStorage.ID) == id {
			// Delete the buyer's routing rules settings in remote storage
			if err := fs.deleteRouteRulesSettingsForBuyerID(ctx, bdoc.Ref.ID); err != nil {
				return &FirestoreError{err: err}
			}

			// Delete the buyer in remote storage
			if _, err := bdoc.Ref.Delete(ctx); err != nil {
				return &FirestoreError{err: err}
			}

			// Find the associated customer, remove the link to the buyer, and check if we should remove the customer
			cdocs := fs.Client.Collection("Customer").Documents(ctx)
			defer cdocs.Stop()
			for {
				cdoc, err := cdocs.Next()
				if err == iterator.Done {
					break
				}

				if err != nil {
					return &FirestoreError{err: err}
				}

				// Unmarshal the customer in firestore to see if it's the customer we need
				var customerInRemoteStorage customer
				err = cdoc.DataTo(&customerInRemoteStorage)
				if err != nil {
					return &UnmarshalError{err: err}
				}

				if customerInRemoteStorage.Buyer != nil && customerInRemoteStorage.Buyer.ID == bdoc.Ref.ID {
					customerInRemoteStorage.Buyer = nil

					if customerInRemoteStorage.Buyer == nil && customerInRemoteStorage.Seller == nil {
						// Remove the customer
						if _, err := cdoc.Ref.Delete(ctx); err != nil {
							return &FirestoreError{err: err}
						}

						break
					}

					// Customer is still needed, but update the references
					if _, err := cdoc.Ref.Set(ctx, customerInRemoteStorage); err != nil {
						return &FirestoreError{err: err}
					}

					break
				}
			}

			// Delete the buyer in cached storage
			fs.buyerMutex.Lock()
			delete(fs.buyers, id)
			fs.buyerMutex.Unlock()
			return nil
		}
	}

	return &DoesNotExistError{resourceType: "buyer", resourceRef: fmt.Sprintf("%x", id)}
}

func (fs *Firestore) SetBuyer(ctx context.Context, b routing.Buyer) error {
	// Get a copy of the buyer in cached storage
	fs.buyerMutex.RLock()
	buyerInCachedStorage, ok := fs.buyers[b.ID]
	fs.buyerMutex.RUnlock()

	if !ok {
		return &DoesNotExistError{resourceType: "buyer", resourceRef: fmt.Sprintf("%x", b.ID)}
	}

	// Loop through all buyers in firestore
	bdocs := fs.Client.Collection("Buyer").Documents(ctx)
	defer bdocs.Stop()
	for {
		bdoc, err := bdocs.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return &FirestoreError{err: err}
		}

		// Unmarshal the buyer in firestore to see if it's the buyer we want to update
		var buyerInRemoteStorage buyer
		err = bdoc.DataTo(&buyerInRemoteStorage)
		if err != nil {
			return &UnmarshalError{err: err}
		}

		// If the buyer is the one we want to update, update it with the new data
		if uint64(buyerInRemoteStorage.ID) == b.ID {
			// Update the buyer in firestore
			newBuyerData := map[string]interface{}{
				"name":                     b.Name,
				"active":                   b.Active,
				"isLiveCustomer":           b.Live,
				"sdkVersion3PublicKeyData": b.PublicKey,
			}

			if _, err := bdoc.Ref.Set(ctx, newBuyerData, firestore.MergeAll); err != nil {
				return &FirestoreError{err: err}
			}

			// Update the buyer's routing rules settings in firestore
			if err := fs.setRoutingRulesSettingsForBuyerID(ctx, bdoc.Ref.ID, b.Name, b.RoutingRulesSettings); err != nil {
				return &FirestoreError{err: err}
			}

			// Update the cached version
			buyerInCachedStorage = b

			fs.buyerMutex.Lock()
			fs.buyers[b.ID] = buyerInCachedStorage
			fs.buyerMutex.Unlock()

			return nil
		}
	}

	return &DoesNotExistError{resourceType: "buyer", resourceRef: fmt.Sprintf("%x", b.ID)}
}

func (fs *Firestore) Seller(id string) (routing.Seller, error) {
	fs.sellerMutex.RLock()
	defer fs.sellerMutex.RUnlock()

	s, found := fs.sellers[id]
	if !found {
		return routing.Seller{}, &DoesNotExistError{resourceType: "seller", resourceRef: id}
	}

	return s, nil
}

func (fs *Firestore) Sellers() []routing.Seller {
	fs.sellerMutex.RLock()
	defer fs.sellerMutex.RUnlock()

	var sellers []routing.Seller
	for _, seller := range fs.sellers {
		sellers = append(sellers, seller)
	}

	sort.Slice(sellers, func(i int, j int) bool { return sellers[i].ID < sellers[j].ID })
	return sellers
}

func (fs *Firestore) AddSeller(ctx context.Context, s routing.Seller) error {
	// Check if the seller exists
	fs.sellerMutex.RLock()
	_, ok := fs.sellers[s.ID]
	fs.sellerMutex.RUnlock()

	if ok {
		return &AlreadyExistsError{resourceType: "seller", resourceRef: s.ID}
	}

	newSellerData := seller{
		Name:                       s.Name,
		PricePublicIngressNibblins: convertCentsToNibblins(s.IngressPriceCents),
		PricePublicEgressNibblins:  convertCentsToNibblins(s.EgressPriceCents),
	}

	// Add the seller in remote storage
	ref := fs.Client.Collection("Seller").Doc(s.ID)
	_, err := ref.Set(ctx, newSellerData)
	if err != nil {
		return &FirestoreError{err: err}
	}

	// Check if a customer already exists for this seller
	var customerFound bool

	cdocs := fs.Client.Collection("Customer").Documents(ctx)
	defer cdocs.Stop()
	for {
		cdoc, err := cdocs.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return &FirestoreError{err: err}
		}

		// Unmarshal the customer in firestore to see if it's the customer we want to add the seller to
		var customerInRemoteStorage customer
		err = cdoc.DataTo(&customerInRemoteStorage)
		if err != nil {
			return &UnmarshalError{err: err}
		}

		if customerInRemoteStorage.Name == s.Name && customerInRemoteStorage.Seller == nil {
			customerFound = true

			customerInRemoteStorage.Seller = ref

			// Update the customer references
			if _, err := cdoc.Ref.Set(ctx, customerInRemoteStorage); err != nil {
				return &FirestoreError{err: err}
			}

			break
		}
	}

	if !customerFound {
		// Customer was not found, so make a new one
		newCustomerData := customer{
			Name:   s.Name,
			Active: true,
			Seller: ref,
		}

		// Create the customer object in remote storage
		if _, _, err = fs.Client.Collection("Customer").Add(ctx, newCustomerData); err != nil {
			return &FirestoreError{err: err}
		}
	}

	// Add the seller in cached storage
	fs.sellerMutex.Lock()
	fs.sellers[s.ID] = s
	fs.sellerMutex.Unlock()

	return nil
}

func (fs *Firestore) RemoveSeller(ctx context.Context, id string) error {
	// Check if the seller exists
	fs.sellerMutex.RLock()
	_, ok := fs.sellers[id]
	fs.sellerMutex.RUnlock()

	if !ok {
		return &DoesNotExistError{resourceType: "seller", resourceRef: id}
	}

	// Delete the seller in remote storage
	sdoc := fs.Client.Collection("Seller").Doc(id)
	if _, err := sdoc.Delete(ctx); err != nil {
		return &FirestoreError{err: err}
	}

	// Find the associated customer, remove the link to the seller, and check if we should remove the customer
	cdocs := fs.Client.Collection("Customer").Documents(ctx)
	defer cdocs.Stop()
	for {
		cdoc, err := cdocs.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return &FirestoreError{err: err}
		}

		// Unmarshal the customer in firestore to see if it's the customer we need
		var customerInRemoteStorage customer
		err = cdoc.DataTo(&customerInRemoteStorage)
		if err != nil {
			return &UnmarshalError{err: err}
		}

		if customerInRemoteStorage.Seller != nil && customerInRemoteStorage.Seller.ID == sdoc.ID {
			customerInRemoteStorage.Seller = nil

			if customerInRemoteStorage.Buyer == nil && customerInRemoteStorage.Seller == nil {
				// Remove the customer
				if _, err := cdoc.Ref.Delete(ctx); err != nil {
					return &FirestoreError{err: err}
				}

				break
			}

			// Customer is still needed, but update the references
			if _, err := cdoc.Ref.Set(ctx, customerInRemoteStorage); err != nil {
				return &FirestoreError{err: err}
			}

			break
		}
	}

	// Delete the seller in cached storage
	fs.sellerMutex.Lock()
	delete(fs.sellers, id)
	fs.sellerMutex.Unlock()
	return nil
}

func (fs *Firestore) SetSeller(ctx context.Context, seller routing.Seller) error {
	// Get a copy of the seller in cached storage
	fs.sellerMutex.RLock()
	sellerInCachedStorage, ok := fs.sellers[seller.ID]
	fs.sellerMutex.RUnlock()

	if !ok {
		return &DoesNotExistError{resourceType: "seller", resourceRef: seller.ID}
	}

	// Update the seller in firestore
	newSellerData := map[string]interface{}{
		"name":                       seller.Name,
		"pricePublicIngressNibblins": convertCentsToNibblins(seller.IngressPriceCents),
		"pricePublicEgressNibblins":  convertCentsToNibblins(seller.EgressPriceCents),
	}

	if _, err := fs.Client.Collection("Seller").Doc(seller.ID).Set(ctx, newSellerData, firestore.MergeAll); err != nil {
		return &FirestoreError{err: err}
	}

	// Update the cached version
	sellerInCachedStorage = seller

	fs.sellerMutex.Lock()
	fs.sellers[seller.ID] = sellerInCachedStorage
	fs.sellerMutex.Unlock()

	return nil
}

func (fs *Firestore) SetCustomerLink(ctx context.Context, customerName string, buyerID uint64, sellerID string) error {
	// Loop through all customers in firestore
	cdocs := fs.Client.Collection("Customer").Documents(ctx)
	defer cdocs.Stop()
	for {
		cdoc, err := cdocs.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return &FirestoreError{err: err}
		}

		// Unmarshal the customer so we can check if this is the customer we want to edit
		var c customer
		err = cdoc.DataTo(&c)
		if err != nil {
			return &UnmarshalError{err: err}
		}

		if c.Name == customerName {
			// Customer was found, now find the associated buyer and seller we want to update the customer's references to
			var buyerRef *firestore.DocumentRef
			var sellerRef *firestore.DocumentRef

			// Find the buyer
			bdocs := fs.Client.Collection("Buyer").Documents(ctx)
			defer bdocs.Stop()
			for {
				bdoc, err := bdocs.Next()
				if err == iterator.Done {
					break
				}

				if err != nil {
					return &FirestoreError{err: err}
				}

				// Unmarshal the buyer so we can check if this is the buyer we want to link the customer to
				var b buyer
				err = bdoc.DataTo(&b)
				if err != nil {
					return &UnmarshalError{err: err}
				}

				if uint64(b.ID) == buyerID {
					buyerRef = bdoc.Ref
					break
				}
			}

			// Find the seller
			sdocs := fs.Client.Collection("Seller").Documents(ctx)
			defer sdocs.Stop()
			for {
				sdoc, err := sdocs.Next()
				if err == iterator.Done {
					break
				}

				if err != nil {
					return &FirestoreError{err: err}
				}

				if sdoc.Ref.ID == sellerID {
					sellerRef = sdoc.Ref
					break
				}
			}

			// Assign the references and restore the customer
			c.Buyer = buyerRef
			c.Seller = sellerRef

			if _, err := cdoc.Ref.Set(ctx, c); err != nil {
				return &FirestoreError{err: err}
			}

			return nil
		}
	}

	return &DoesNotExistError{resourceType: "customer", resourceRef: customerName}
}

func (fs *Firestore) BuyerIDFromCustomerName(ctx context.Context, customerName string) (uint64, error) {
	// Loop through all customers in firestore
	cdocs := fs.Client.Collection("Customer").Documents(ctx)
	defer cdocs.Stop()
	for {
		cdoc, err := cdocs.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return 0, &FirestoreError{err: err}
		}

		// Unmarshal the customer so we can check if this is the customer we want to edit
		var c customer
		err = cdoc.DataTo(&c)
		if err != nil {
			return 0, &UnmarshalError{err: err}
		}

		if c.Name == customerName {
			bdoc, err := c.Buyer.Get(ctx)
			if err != nil {
				return 0, &DoesNotExistError{resourceType: "buyer", resourceRef: c.Buyer.ID}
			}

			var b buyer
			if err := bdoc.DataTo(&b); err != nil {
				return 0, &UnmarshalError{err: err}
			}

			return uint64(b.ID), nil
		}
	}

	return 0, &DoesNotExistError{resourceType: "customer", resourceRef: customerName}
}

func (fs *Firestore) SellerIDFromCustomerName(ctx context.Context, customerName string) (string, error) {
	// Loop through all customers in firestore
	cdocs := fs.Client.Collection("Customer").Documents(ctx)
	defer cdocs.Stop()
	for {
		cdoc, err := cdocs.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return "", &FirestoreError{err: err}
		}

		// Unmarshal the customer so we can check if this is the customer we want to edit
		var c customer
		err = cdoc.DataTo(&c)
		if err != nil {
			return "", &UnmarshalError{err: err}
		}

		if c.Name == customerName {
			sdoc, err := c.Seller.Get(ctx)
			if err != nil {
				return "", &DoesNotExistError{resourceType: "seller", resourceRef: c.Seller.ID}
			}

			return sdoc.Ref.ID, nil
		}
	}

	return "", &DoesNotExistError{resourceType: "customer", resourceRef: customerName}
}

func (fs *Firestore) Relay(id uint64) (routing.Relay, error) {
	fs.relayMutex.RLock()
	defer fs.relayMutex.RUnlock()

	relay, found := fs.relays[id]
	if !found {
		return routing.Relay{}, &DoesNotExistError{resourceType: "relay", resourceRef: fmt.Sprintf("%x", id)}
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

	sort.Slice(relays, func(i int, j int) bool { return relays[i].ID < relays[j].ID })
	return relays
}

func (fs *Firestore) AddRelay(ctx context.Context, r routing.Relay) error {
	var sellerRef *firestore.DocumentRef
	var datacenterRef *firestore.DocumentRef

	// Loop through all sellers in firestore
	sdocs := fs.Client.Collection("Seller").Documents(ctx)
	defer sdocs.Stop()
	for {
		sdoc, err := sdocs.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return &FirestoreError{err: err}
		}

		// If the seller is the one associated with this relay, set the relay's seller reference
		if sdoc.Ref.ID == r.Seller.ID {
			sellerRef = sdoc.Ref
			break
		}
	}

	if sellerRef == nil {
		return &DoesNotExistError{resourceType: "seller", resourceRef: r.Seller.ID}
	}

	// Loop through all datacenters in firestore
	ddocs := fs.Client.Collection("Datacenter").Documents(ctx)
	defer ddocs.Stop()
	for {
		ddoc, err := ddocs.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return &FirestoreError{err: err}
		}

		// Unmarshal the datacenter so we can check if the ID matches the datacenter associated with this relay
		var d datacenter
		err = ddoc.DataTo(&d)
		if err != nil {
			return &UnmarshalError{err: err}
		}

		// If the datacenter is the one associated with this relay, set the relay's datacenter reference
		if crypto.HashID(d.Name) == r.Datacenter.ID {
			datacenterRef = ddoc.Ref
			break
		}
	}

	if datacenterRef == nil {
		return &DoesNotExistError{resourceType: "datacenter", resourceRef: fmt.Sprintf("%x", r.Datacenter.ID)}
	}

	newRelayData := relay{
		Name:               r.Name,
		Address:            r.Addr.String(),
		PublicKey:          r.PublicKey,
		UpdateKey:          r.PublicKey,
		NICSpeedMbps:       int64(r.NICSpeedMbps),
		IncludedBandwithGB: int64(r.IncludedBandwidthGB),
		Datacenter:         datacenterRef,
		Seller:             sellerRef,
		ManagementAddress:  r.ManagementAddr,
		SSHUser:            r.SSHUser,
		SSHPort:            r.SSHPort,
		State:              r.State,
		LastUpdateTime:     r.LastUpdateTime,
	}

	// Add the relay in remote storage
	if _, _, err := fs.Client.Collection("Relay").Add(ctx, newRelayData); err != nil {
		return &FirestoreError{err: err}
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
		return &DoesNotExistError{resourceType: "relay", resourceRef: fmt.Sprintf("%x", id)}
	}

	rdocs := fs.Client.Collection("Relay").Documents(ctx)
	defer rdocs.Stop()
	for {
		rdoc, err := rdocs.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return &FirestoreError{err: err}
		}

		// Unmarshal the relay in firestore to see if it's the relay we want to delete
		var relayInRemoteStorage relay
		err = rdoc.DataTo(&relayInRemoteStorage)
		if err != nil {
			return &UnmarshalError{err: err}
		}

		rid := crypto.HashID(relayInRemoteStorage.Address)
		if rid == id {
			// Delete the relay in remote storage
			if _, err := rdoc.Ref.Delete(ctx); err != nil {
				return &FirestoreError{err: err}
			}

			// Delete the relay in cached storage
			fs.relayMutex.Lock()
			delete(fs.relays, id)
			fs.relayMutex.Unlock()
			return nil
		}
	}

	return &DoesNotExistError{resourceType: "relay", resourceRef: fmt.Sprintf("%x", id)}
}

// Only relay state, public key, and NIC speed are updated in firestore for now
func (fs *Firestore) SetRelay(ctx context.Context, r routing.Relay) error {
	// Get a copy of the relay in cached storage
	fs.relayMutex.RLock()
	relayInCachedStorage, ok := fs.relays[r.ID]
	fs.relayMutex.RUnlock()

	if !ok {
		return &DoesNotExistError{resourceType: "relay", resourceRef: fmt.Sprintf("%x", r.ID)}
	}

	// Loop through all relays in firestore
	rdocs := fs.Client.Collection("Relay").Documents(ctx)
	defer rdocs.Stop()
	for {
		rdoc, err := rdocs.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return &FirestoreError{err: err}
		}

		// Unmarshal the relay in firestore to see if it's the relay we want to update
		var relayInRemoteStorage relay
		err = rdoc.DataTo(&relayInRemoteStorage)
		if err != nil {
			return &UnmarshalError{err: err}
		}

		// If the relay is the one we want to update, update it with the new data
		rid := crypto.HashID(relayInRemoteStorage.Address)
		if rid == r.ID {
			// Set the data to update the relay with
			newRelayData := map[string]interface{}{
				"state":           r.State,
				"lastUpdateTime":  r.LastUpdateTime,
				"stateUpdateTime": time.Now(),
				"publicKey":       r.PublicKey,
				"nicSpeedMbps":    int64(r.NICSpeedMbps),
			}

			// Update the relay in firestore
			if _, err := rdoc.Ref.Set(ctx, newRelayData, firestore.MergeAll); err != nil {
				return &FirestoreError{err: err}
			}

			// Update the cached version
			relayInCachedStorage.State = r.State
			relayInCachedStorage.LastUpdateTime = r.LastUpdateTime

			fs.relayMutex.Lock()
			fs.relays[r.ID] = relayInCachedStorage
			fs.relayMutex.Unlock()

			return nil
		}
	}

	return &DoesNotExistError{resourceType: "relay", resourceRef: fmt.Sprintf("%x", r.ID)}
}

func (fs *Firestore) DatacenterMapsForBuyer(buyerID string) map[uint64]routing.DatacenterMap {
	fs.datacenterMapMutex.RLock()
	defer fs.datacenterMapMutex.RUnlock()

	// buyer can have multiple dc aliases
	var dcs = make(map[uint64]routing.DatacenterMap)
	for _, dc := range fs.datacenterMaps {
		if dc.BuyerID == buyerID {
			id := crypto.HashID(dc.Alias + dc.BuyerID + dc.Datacenter)
			dcs[id] = dc
		}
	}

	return dcs
}

func (fs *Firestore) AddDatacenterMap(ctx context.Context, dcMap routing.DatacenterMap) error {

	// ToDo: make sure buyer and datacenter exist?

	dcMaps := fs.DatacenterMapsForBuyer(dcMap.BuyerID)
	if len(dcMaps) != 0 {
		for _, dc := range dcMaps {
			if dc.Alias == dcMap.Alias && dc.Datacenter == dcMap.Datacenter {
				return &AlreadyExistsError{resourceType: "datacenterMap", resourceRef: dcMap.Alias}
			}
		}
	}
	_, _, err := fs.Client.Collection("DatacenterMaps").Add(ctx, dcMap)
	if err != nil {
		return &FirestoreError{err: err}
	}

	// update local store
	fs.datacenterMapMutex.Lock()
	id := crypto.HashID(dcMap.Alias + dcMap.BuyerID + dcMap.Datacenter)
	fs.datacenterMaps[id] = dcMap
	fs.datacenterMapMutex.Unlock()

	return nil

}

func (fs *Firestore) ListDatacenterMaps(dcID string) map[uint64]routing.DatacenterMap {
	fs.datacenterMapMutex.RLock()
	defer fs.datacenterMapMutex.RUnlock()

	if dcID == "" {
		return fs.datacenterMaps
	}

	var dcs = make(map[uint64]routing.DatacenterMap)
	for _, dc := range fs.datacenterMaps {
		if dc.Datacenter == dcID {
			id := crypto.HashID(dc.Alias + dc.BuyerID + dc.Datacenter)
			dcs[id] = dc
		}
	}

	return dcs
}

func (fs *Firestore) ModifyDatacenterMap(ctx context.Context, dcMap routing.DatacenterMap) error {

	return nil
}

func (fs *Firestore) RemoveDatacenterMap(ctx context.Context, dcMap routing.DatacenterMap) error {
	dmdocs := fs.Client.Collection("DatacenterMaps").Documents(ctx)
	defer dmdocs.Stop()

	// Firestore is the source of truth
	found := false
	for {
		dmdoc, err := dmdocs.Next()
		ref := dmdoc.Ref
		if err == iterator.Done {
			break
		}

		if err != nil {
			return &FirestoreError{err: err}
		}

		var dcm routing.DatacenterMap
		err = dmdoc.DataTo(&dcm)
		if err != nil {
			return &UnmarshalError{err: err}
		}

		// all components must match (one-to-many)
		// could use cmp?
		if dcMap.Alias == dcm.Alias && dcMap.BuyerID == dcm.BuyerID && dcMap.Datacenter == dcm.Datacenter {
			_, err := ref.Delete(ctx)
			if err != nil {
				return &FirestoreError{err: err}
			}
			found = true
		}
	}

	if found {
		fs.datacenterMapMutex.RLock()
		id := crypto.HashID(dcMap.Alias + dcMap.BuyerID + dcMap.Datacenter)

		delete(fs.datacenterMaps, id)

		fs.datacenterMapMutex.RUnlock()
		return nil
	}

	return &DoesNotExistError{resourceType: "datacenterMap", resourceRef: fmt.Sprintf("%v", dcMap)}
}

func (fs *Firestore) Datacenter(id uint64) (routing.Datacenter, error) {
	fs.datacenterMutex.RLock()
	defer fs.datacenterMutex.RUnlock()

	d, found := fs.datacenters[id]
	if !found {
		// Check if there is a datacenter with this alias
		for _, datacenter := range fs.datacenters {
			if id == crypto.HashID(datacenter.AliasName) {
				return datacenter, nil
			}
		}

		return routing.Datacenter{}, &DoesNotExistError{resourceType: "datacenter", resourceRef: fmt.Sprintf("%x", id)}
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

	sort.Slice(datacenters, func(i int, j int) bool { return datacenters[i].ID < datacenters[j].ID })
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
		return &FirestoreError{err: err}
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
		return &DoesNotExistError{resourceType: "datacenter", resourceRef: fmt.Sprintf("%x", id)}
	}

	ddocs := fs.Client.Collection("Datacenter").Documents(ctx)
	defer ddocs.Stop()
	for {
		ddoc, err := ddocs.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return &FirestoreError{err: err}
		}

		// Unmarshal the datanceter in firestore to see if it's the datacenter we want to delete
		var datacenterInRemoteStorage datacenter
		err = ddoc.DataTo(&datacenterInRemoteStorage)
		if err != nil {
			return &UnmarshalError{err: err}
		}

		if crypto.HashID(datacenterInRemoteStorage.Name) == id {
			// Delete the datacenter in remote storage
			if _, err := ddoc.Ref.Delete(ctx); err != nil {
				return &FirestoreError{err: err}
			}

			// Delete the datacenter in cached storage
			fs.datacenterMutex.Lock()
			delete(fs.datacenters, id)
			fs.datacenterMutex.Unlock()
			return nil
		}
	}

	return &DoesNotExistError{resourceType: "datacenter", resourceRef: fmt.Sprintf("%x", id)}
}

func (fs *Firestore) SetDatacenter(ctx context.Context, d routing.Datacenter) error {
	// Get a copy of the datacenter in cached storage
	fs.datacenterMutex.RLock()
	datacenterInCachedStorage, ok := fs.datacenters[d.ID]
	fs.datacenterMutex.RUnlock()

	if !ok {
		return &DoesNotExistError{resourceType: "datacenter", resourceRef: fmt.Sprintf("%x", d.ID)}
	}

	// Loop through all datacenters in firestore
	ddocs := fs.Client.Collection("Datacenter").Documents(ctx)
	defer ddocs.Stop()
	for {
		ddoc, err := ddocs.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return &FirestoreError{err: err}
		}

		// Unmarshal the datacenter in firestore to see if it's the datacenter we want to update
		var datacenterInRemoteStorage datacenter
		err = ddoc.DataTo(&datacenterInRemoteStorage)
		if err != nil {
			return &UnmarshalError{err: err}
		}

		// If the datacenter is the one we want to update, update it with the new data
		if crypto.HashID(datacenterInRemoteStorage.Name) == d.ID {
			// Set the data to update the datacenter with
			newDatacenterData := map[string]interface{}{
				"name":      d.Name,
				"enabled":   d.Enabled,
				"latitude":  d.Location.Latitude,
				"longitude": d.Location.Longitude,
			}

			// Update the datacenter in firestore
			if _, err := ddoc.Ref.Set(ctx, newDatacenterData, firestore.MergeAll); err != nil {
				return &FirestoreError{err: err}
			}

			// Update the cached version
			datacenterInCachedStorage = d

			fs.datacenterMutex.Lock()
			fs.datacenters[d.ID] = datacenterInCachedStorage
			fs.datacenterMutex.Unlock()

			return nil
		}
	}

	return &DoesNotExistError{resourceType: "datacenter", resourceRef: fmt.Sprintf("%x", d.ID)}
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
		if err := fs.syncCustomers(ctx); err != nil {
			outerErr = fmt.Errorf("failed to sync customers: %v", err)
		}
		wg.Done()
	}()

	go func() {
		if err := fs.syncDatacenters(ctx); err != nil {
			outerErr = fmt.Errorf("failed to sync datacenters: %v", err)
		}
		wg.Done()
	}()

	go func() {
		if err := fs.syncDatacenterMaps(ctx); err != nil {
			outerErr = fmt.Errorf("failed to sync datacenterMaps: %v", err)
		}
		wg.Done()
	}()

	wg.Wait()

	return outerErr
}

func (fs *Firestore) syncDatacenters(ctx context.Context) error {
	datacenters := make(map[uint64]routing.Datacenter)

	ddocs := fs.Client.Collection("Datacenter").Documents(ctx)
	defer ddocs.Stop()
	for {
		ddoc, err := ddocs.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return &FirestoreError{err: err}
		}

		var d datacenter
		err = ddoc.DataTo(&d)
		if err != nil {
			level.Warn(fs.Logger).Log("msg", fmt.Sprintf("failed to unmarshal datacenter %v", ddoc.Ref.ID), "err", err)
			continue
		}

		did := crypto.HashID(d.Name)
		datacenters[did] = routing.Datacenter{
			ID:        did,
			Name:      d.Name,
			AliasName: d.AliasName,
			Enabled:   d.Enabled,
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
	defer rdocs.Stop()
	for {
		rdoc, err := rdocs.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return &FirestoreError{err: err}
		}

		var r relay
		err = rdoc.DataTo(&r)
		if err != nil {
			level.Warn(fs.Logger).Log("msg", fmt.Sprintf("failed to unmarshal relay %v", rdoc.Ref.ID), "err", err)
			continue
		}

		rid := crypto.HashID(r.Address)

		host, port, err := net.SplitHostPort(r.Address)
		if err != nil {
			return &UnmarshalError{err: fmt.Errorf("failed to split host and port: %v", err)}
		}
		iport, err := strconv.ParseInt(port, 10, 64)
		if err != nil {
			return &UnmarshalError{err: fmt.Errorf("failed to convert port to int: %v", err)}
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
			NICSpeedMbps:        uint64(r.NICSpeedMbps),
			IncludedBandwidthGB: uint64(r.IncludedBandwithGB),
			ManagementAddr:      r.ManagementAddress,
			SSHUser:             r.SSHUser,
			SSHPort:             r.SSHPort,
			State:               r.State,
			LastUpdateTime:      r.LastUpdateTime,
			MaxSessions:         uint32(r.MaxSessions),
			UpdateKey:           r.UpdateKey,
			FirestoreID:         rdoc.Ref.ID,
		}

		// Set a default max session count of 3000 if the value isn't set in firestore
		if relay.MaxSessions == 0 {
			relay.MaxSessions = 3000
		}

		// Get datacenter
		ddoc, err := r.Datacenter.Get(ctx)
		if err != nil {
			return &FirestoreError{err: err}
		}

		var d datacenter
		err = ddoc.DataTo(&d)
		if err != nil {
			level.Warn(fs.Logger).Log("msg", fmt.Sprintf("failed to unmarshal datacenter %v", ddoc.Ref.ID), "err", err)
			continue
		}

		datacenter := routing.Datacenter{
			ID:        crypto.HashID(d.Name),
			Name:      d.Name,
			AliasName: d.AliasName,
			Enabled:   d.Enabled,
			Location: routing.Location{
				Latitude:  d.Latitude,
				Longitude: d.Longitude,
			},
		}

		relay.Datacenter = datacenter

		// Get seller
		sdoc, err := r.Seller.Get(ctx)
		if err != nil {
			return &FirestoreError{err: err}
		}

		var s seller
		err = sdoc.DataTo(&s)
		if err != nil {
			level.Warn(fs.Logger).Log("msg", fmt.Sprintf("failed to unmarshal seller %v", sdoc.Ref.ID), "err", err)
			continue
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

func (fs *Firestore) syncDatacenterMaps(ctx context.Context) error {
	dcMaps := make(map[uint64]routing.DatacenterMap)

	dcdocs := fs.Client.Collection("DatacenterMaps").Documents(ctx)
	defer dcdocs.Stop()
	for {
		dcdoc, err := dcdocs.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return &FirestoreError{err: err}
		}

		var dcMap routing.DatacenterMap
		err = dcdoc.DataTo(&dcMap)
		if err != nil {
			level.Warn(fs.Logger).Log("msg", fmt.Sprintf("failed to unmarshal datacenterMap %v", dcdoc.Ref.ID), "err", err)
			continue
		}

		id := crypto.HashID(dcMap.Alias + dcMap.BuyerID + dcMap.Datacenter)
		dcMaps[id] = dcMap
	}

	fs.datacenterMaps = dcMaps
	return nil

}

func (fs *Firestore) syncCustomers(ctx context.Context) error {
	buyers := make(map[uint64]routing.Buyer)
	sellers := make(map[string]routing.Seller)

	cdocs := fs.Client.Collection("Customer").Documents(ctx)
	defer cdocs.Stop()
	for {
		cdoc, err := cdocs.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return &FirestoreError{err: err}
		}

		var c customer
		err = cdoc.DataTo(&c)
		if err != nil {
			level.Warn(fs.Logger).Log("msg", fmt.Sprintf("failed to unmarshal customer %v", cdoc.Ref.ID), "err", err)
			continue
		}

		if !c.Active {
			continue
		}

		// Get the associated buyer for the customer
		if c.Buyer != nil {
			bdoc, err := c.Buyer.Get(ctx)
			if err != nil {
				level.Warn(fs.Logger).Log("msg", fmt.Sprintf("failed to get buyer %v", c.Buyer.ID), "err", err)
				continue
			}
			var b buyer
			err = bdoc.DataTo(&b)
			if err != nil {
				level.Warn(fs.Logger).Log("msg", fmt.Sprintf("failed to unmarshal seller %v", bdoc.Ref.ID), "err", err)
				continue
			}
			rrs, err := fs.getRoutingRulesSettingsForBuyerID(ctx, bdoc.Ref.ID)
			if err != nil {
				level.Warn(fs.Logger).Log("msg", fmt.Sprintf("using default route rules for buyer %v", bdoc.Ref.ID), "err", err)
			}
			buyers[uint64(b.ID)] = routing.Buyer{
				ID:                   uint64(b.ID),
				Name:                 b.Name,
				Domain:               c.Domain,
				Active:               b.Active,
				Live:                 b.Live,
				PublicKey:            b.PublicKey,
				RoutingRulesSettings: rrs,
			}
		}

		// Get the associated seller for the customer
		if c.Seller != nil {
			sdoc, err := c.Seller.Get(ctx)
			if err != nil {
				level.Warn(fs.Logger).Log("msg", fmt.Sprintf("failed to get seller %v", c.Seller.ID), "err", err)
				continue
			}
			var s seller
			err = sdoc.DataTo(&s)
			if err != nil {
				level.Warn(fs.Logger).Log("msg", fmt.Sprintf("failed to unmarshal seller %v", sdoc.Ref.ID), "err", err)
				continue
			}

			sellers[sdoc.Ref.ID] = routing.Seller{
				ID:                sdoc.Ref.ID,
				Name:              s.Name,
				IngressPriceCents: convertNibblinsToCents(s.PricePublicIngressNibblins),
				EgressPriceCents:  convertNibblinsToCents(s.PricePublicEgressNibblins),
			}
		}
	}

	fs.buyerMutex.Lock()
	fs.buyers = buyers
	fs.buyerMutex.Unlock()

	fs.sellerMutex.Lock()
	fs.sellers = sellers
	fs.sellerMutex.Unlock()

	level.Info(fs.Logger).Log("during", "syncBuyers", "num", len(fs.buyers))
	level.Info(fs.Logger).Log("during", "syncSellers", "num", len(fs.sellers))

	return nil
}

func (fs *Firestore) createRouteRulesSettingsForBuyerID(ctx context.Context, ID string, name string, rrs routing.RoutingRulesSettings) error {
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
		EnableTryBeforeYouBuy:        rrs.EnableTryBeforeYouBuy,
		TryBeforeYouBuyMaxSlices:     rrs.TryBeforeYouBuyMaxSlices,
		SelectionPercentage:          rrs.SelectionPercentage,
	}

	// Attempt to create route shader for buyer
	rsDocRef := fs.Client.Collection("RouteShader").NewDoc()
	rsDocRef.ID = routeShaderID

	_, err := rsDocRef.Create(ctx, rrsFirestore)
	return err
}

func (fs *Firestore) deleteRouteRulesSettingsForBuyerID(ctx context.Context, ID string) error {
	// Comment below taken from old backend, at least attempting to explain why we need to append _0 (no existing entries have suffixes other than _0)
	// "Must be of the form '<buyer key>_<tag id>'. The buyer key can be found by looking at the ID under Buyer; it should be something like 763IMDH693HLsr2LGTJY. The tag ID should be 0 (for default) or the fnv64a hash of the tag the customer is using. Therefore this value should look something like: 763IMDH693HLsr2LGTJY_0. This value can not be changed after the entity is created."
	routeShaderID := ID + "_0"

	// Attempt to delete route shader for buyer
	_, err := fs.Client.Collection("RouteShader").Doc(routeShaderID).Delete(ctx)
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
	rrs.EnableTryBeforeYouBuy = tempRRS.EnableTryBeforeYouBuy
	rrs.TryBeforeYouBuyMaxSlices = tempRRS.TryBeforeYouBuyMaxSlices
	rrs.SelectionPercentage = tempRRS.SelectionPercentage

	return rrs, nil
}

func (fs *Firestore) setRoutingRulesSettingsForBuyerID(ctx context.Context, ID string, name string, rrs routing.RoutingRulesSettings) error {
	// Comment below taken from old backend, at least attempting to explain why we need to append _0 (no existing entries have suffixes other than _0)
	// "Must be of the form '<buyer key>_<tag id>'. The buyer key can be found by looking at the ID under Buyer; it should be something like 763IMDH693HLsr2LGTJY. The tag ID should be 0 (for default) or the fnv64a hash of the tag the customer is using. Therefore this value should look something like: 763IMDH693HLsr2LGTJY_0. This value can not be changed after the entity is created."
	routeShaderID := ID + "_0"

	// Convert RoutingRulesSettings struct to firestore map
	rrsFirestore := map[string]interface{}{
		"displayName":              name,
		"envelopeKbpsUp":           rrs.EnvelopeKbpsUp,
		"envelopeKbpsDown":         rrs.EnvelopeKbpsDown,
		"mode":                     rrs.Mode,
		"maxPricePerGBNibblins":    convertCentsToNibblins(rrs.MaxCentsPerGB),
		"acceptableLatency":        rrs.AcceptableLatency,
		"rttRouteSwitch":           rrs.RTTEpsilon,
		"rttThreshold":             rrs.RTTThreshold,
		"rttHysteresis":            rrs.RTTHysteresis,
		"rttVeto":                  rrs.RTTVeto,
		"youOnlyLiveOnce":          rrs.EnableYouOnlyLiveOnce,
		"packetLossSafety":         rrs.EnablePacketLossSafety,
		"packetLossMultipath":      rrs.EnableMultipathForPacketLoss,
		"jitterMultipath":          rrs.EnableMultipathForJitter,
		"rttMultipath":             rrs.EnableMultipathForRTT,
		"abTest":                   rrs.EnableABTest,
		"tryBeforeYouBuy":          rrs.EnableTryBeforeYouBuy,
		"tryBeforeYouBuyMaxSlices": rrs.TryBeforeYouBuyMaxSlices,
		"selectionPercentage":      rrs.SelectionPercentage,
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
