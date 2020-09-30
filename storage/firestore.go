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
	"github.com/networknext/backend/core"
	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/routing"
	"google.golang.org/api/iterator"
)

type Firestore struct {
	Client *firestore.Client
	Logger log.Logger

	datacenters    map[uint64]routing.Datacenter
	relays         map[uint64]routing.Relay
	customers      map[string]routing.Customer
	buyers         map[uint64]routing.Buyer
	sellers        map[string]routing.Seller
	datacenterMaps map[uint64]routing.DatacenterMap

	syncSequenceNumber int64

	datacenterMutex     sync.RWMutex
	relayMutex          sync.RWMutex
	customerMutex       sync.RWMutex
	buyerMutex          sync.RWMutex
	sellerMutex         sync.RWMutex
	datacenterMapMutex  sync.RWMutex
	sequenceNumberMutex sync.RWMutex
}

type customer struct {
	Code                   string                 `firestore:"code"`
	Name                   string                 `firestore:"name"`
	AutomaticSignInDomains string                 `firestore:"automaticSigninDomains"`
	Active                 bool                   `firestore:"active"`
	BuyerRef               *firestore.DocumentRef `firestore:"buyerRef"`
	SellerRef              *firestore.DocumentRef `firestore:"sellerRef"`
}

type buyer struct {
	ID          int64  `firestore:"sdkVersion3PublicKeyId"`
	CompanyCode string `firestore:"companyCode"`
	Live        bool   `firestore:"isLiveCustomer"`
	Debug       bool   `firestore:"isDebug"`
	PublicKey   []byte `firestore:"sdkVersion3PublicKeyData"`
}

type seller struct {
	ID                        string `firestore:"id"`
	Name                      string `firestore:"name"`
	CompanyCode               string `firestore:"companyCode"`
	IngressPriceNibblinsPerGB int64  `firestore:"pricePublicIngressNibblins"`
	EgressPriceNibblinsPerGB  int64  `firestore:"pricePublicEgressNibblins"`
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
	MRC                int64                  `firestore:"monthlyRecurringChargeNibblins"`
	Overage            int64                  `firestore:"overage"`
	BWRule             int32                  `firestore:"bandwidthRule"`
	ContractTerm       int32                  `firestore:"contractTerm"`
	StartDate          time.Time              `firestore:"startDate"`
	EndDate            time.Time              `firestore:"endDate"`
	Type               string                 `firestore:"machineType"`
}

type datacenter struct {
	Name         string  `firestore:"name"`
	Enabled      bool    `firestore:"enabled"`
	Latitude     float64 `firestore:"latitude"`
	Longitude    float64 `firestore:"longitude"`
	SupplierName string  `firestore:"supplierName"`
}

type datacenterMap struct {
	Alias      string `firestore:"Alias"`
	Datacenter string `firestore:"Datacenter"`
	Buyer      string `firestore:"Buyer"`
}

type routingRulesSettings struct {
	DisplayName                  string          `firestore:"displayName"`
	EnvelopeKbpsUp               int64           `firestore:"envelopeKbpsUp"`
	EnvelopeKbpsDown             int64           `firestore:"envelopeKbpsDown"`
	Mode                         int64           `firestore:"mode"`
	MaxPricePerGBNibblins        int64           `firestore:"maxPricePerGBNibblins"`
	AcceptableLatency            float32         `firestore:"acceptableLatency"`
	RTTEpsilon                   float32         `firestore:"rttRouteSwitch"`
	RTTThreshold                 float32         `firestore:"rttThreshold"`
	RTTHysteresis                float32         `firestore:"rttHysteresis"`
	RTTVeto                      float32         `firestore:"rttVeto"`
	EnableYouOnlyLiveOnce        bool            `firestore:"youOnlyLiveOnce"`
	EnablePacketLossSafety       bool            `firestore:"packetLossSafety"`
	EnableMultipathForPacketLoss bool            `firestore:"packetLossMultipath"`
	MultipathPacketLossThreshold float32         `firestore:"multipathPacketLossThreshold"`
	EnableMultipathForJitter     bool            `firestore:"jitterMultipath"`
	EnableMultipathForRTT        bool            `firestore:"rttMultipath"`
	EnableABTest                 bool            `firestore:"abTest"`
	EnableTryBeforeYouBuy        bool            `firestore:"tryBeforeYouBuy"`
	TryBeforeYouBuyMaxSlices     int8            `firestore:"tryBeforeYouBuyMaxSlices"`
	SelectionPercentage          int64           `firestore:"selectionPercentage"`
	ExcludedUserHashes           map[string]bool `firestore:"excludedUserHashes"`
}

type routeShader struct {
	DisableNetworkNext        bool           `firestore:"disableNetworkNext"`
	SelectionPercent          int            `firestore:"selectionPercent"`
	ABTest                    bool           `firestore:"abTest"`
	ProMode                   bool           `firestore:"proMode"`
	ReduceLatency             bool           `firestore:"reduceLatency"`
	ReducePacketLoss          bool           `firestore:"reducePacketLoss"`
	Multipath                 bool           `firestore:"multipath"`
	AcceptableLatency         int32          `firestore:"acceptableLatency"`
	LatencyThreshold          int32          `firestore:"latencyThreshold"`
	AcceptablePacketLoss      float32        `firestore:"acceptablePacketLoss"`
	BandwidthEnvelopeUpKbps   int32          `firestore:"bandwidthEnvelopeUpKbps"`
	BandwidthEnvelopeDownKbps int32          `firestore:"bandwidthEnvelopeDownKbps"`
	BannedUsers               map[int64]bool `firestore:"bannedUsers"`
}

type internalConfig struct {
	RouteSwitchThreshold       int32 `firestore:"routeSwitchThreshold"`
	MaxLatencyTradeOff         int32 `firestore:"maxLatencyTradeOff"`
	RTTVeto_Default            int32 `firestore:"rttVeto_default"`
	RTTVeto_PacketLoss         int32 `firestore:"rttVeto_packetLoss"`
	RTTVeto_Multipath          int32 `firestore:"rttVeto_multipath"`
	MultipathOverloadThreshold int32 `firestore:"multipathOverloadThreshold"`
	TryBeforeYouBuy            bool  `firestore:"tryBeforeYouBuy"`
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
		Client:             client,
		Logger:             logger,
		datacenters:        make(map[uint64]routing.Datacenter),
		datacenterMaps:     make(map[uint64]routing.DatacenterMap),
		relays:             make(map[uint64]routing.Relay),
		customers:          make(map[string]routing.Customer),
		buyers:             make(map[uint64]routing.Buyer),
		sellers:            make(map[string]routing.Seller),
		syncSequenceNumber: -1,
	}, nil

}

// SetSequenceNumber is required for testing with the Firestore emulator
func (fs *Firestore) SetSequenceNumber(ctx context.Context, value int64) error {
	seqDocs := fs.Client.Collection("MetaData")
	seq := seqDocs.Doc("SyncSequenceNumber")

	var num struct {
		Value int64 `firestore:"value"`
	}
	num.Value = value

	if _, err := seq.Set(ctx, num); err != nil {
		return &FirestoreError{err: err}
	}

	return nil

}

// IncrementSequenceNumber is called by all CRUD operations defined in the Storage interface. It only
// increments the remote seq number. When the sync() functions call CheckSequenceNumber(), if the
// local and remote numbers are not the same, the data will be sync'd from Firestore, and the local
// sequence numbers updated.
func (fs *Firestore) IncrementSequenceNumber(ctx context.Context) error {

	seqDocs := fs.Client.Collection("MetaData")
	seq, err := seqDocs.Doc("SyncSequenceNumber").Get(ctx)
	if err != nil {
		return &DoesNotExistError{resourceType: "sequence number", resourceRef: ""}
	}

	var num struct {
		Value int64 `firestore:"value"`
	}

	err = seq.DataTo(&num)
	if err != nil {
		return &UnmarshalError{err: err}
	}

	num.Value++
	if _, err = seq.Ref.Set(ctx, num); err != nil {
		return &FirestoreError{err: err}
	}

	return nil
}

// CheckSequenceNumber is called in the Firestore sync*() operations to see if a sync is required.
// Returns true if the remote number != the local number which forces the caller to sync from
// Firestore and updates the local sequence number. Returns false (no need to sync) and does
// not modify the local number, otherwise.
func (fs *Firestore) CheckSequenceNumber(ctx context.Context) (bool, error) {

	var num struct {
		Value int64 `firestore:"value"`
	}

	seqDocs := fs.Client.Collection("MetaData")
	seq, err := seqDocs.Doc("SyncSequenceNumber").Get(ctx)
	if err != nil {
		return false, &DoesNotExistError{resourceType: "sequence number", resourceRef: ""}
	}

	err = seq.DataTo(&num)
	if err != nil {
		return false, &UnmarshalError{err: err}
	}

	fs.sequenceNumberMutex.RLock()
	localSeqNum := fs.syncSequenceNumber
	fs.sequenceNumberMutex.RUnlock()

	if localSeqNum != num.Value {
		fs.sequenceNumberMutex.Lock()
		fs.syncSequenceNumber = num.Value
		fs.sequenceNumberMutex.Unlock()
		return true, nil
	}

	return false, nil
}

func (fs *Firestore) Customer(code string) (routing.Customer, error) {
	fs.customerMutex.RLock()
	defer fs.customerMutex.RUnlock()

	c, found := fs.customers[code]
	if !found {
		return routing.Customer{}, &DoesNotExistError{resourceType: "customer", resourceRef: fmt.Sprintf("%s", code)}
	}

	return c, nil
}

func (fs *Firestore) Customers() []routing.Customer {
	var customers []routing.Customer
	for _, customer := range fs.customers {
		customers = append(customers, customer)
	}

	sort.Slice(customers, func(i int, j int) bool { return customers[i].Name < customers[j].Name })
	return customers
}

func (fs *Firestore) CustomerWithName(name string) (routing.Customer, error) {
	fs.customerMutex.RLock()
	defer fs.customerMutex.RUnlock()

	for _, customer := range fs.customers {
		if customer.Name == name {
			return customer, nil
		}
	}

	return routing.Customer{}, &DoesNotExistError{resourceType: "buyer", resourceRef: name}
}

func (fs *Firestore) AddCustomer(ctx context.Context, c routing.Customer) error {
	fs.customerMutex.RLock()
	_, ok := fs.customers[c.Code]
	fs.customerMutex.RUnlock()

	if ok {
		return &AlreadyExistsError{resourceType: "customer", resourceRef: c.Code}
	}

	newCustomerData := customer{
		Code:                   c.Code,
		Name:                   c.Name,
		AutomaticSignInDomains: "",
		Active:                 false,
	}

	// Add the buyer in remote storage
	_, _, err := fs.Client.Collection("Customer").Add(ctx, newCustomerData)
	if err != nil {
		return &FirestoreError{err: err}
	}

	// Add the buyer in cached storage
	fs.customerMutex.Lock()
	fs.customers[c.Code] = c
	fs.customerMutex.Unlock()

	fs.IncrementSequenceNumber(ctx)

	return nil
}

func (fs *Firestore) RemoveCustomer(ctx context.Context, code string) error {
	// Check if the buyer exists
	fs.customerMutex.RLock()
	_, ok := fs.customers[code]
	fs.customerMutex.RUnlock()

	if !ok {
		return &DoesNotExistError{resourceType: "customer", resourceRef: code}
	}

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

		// Unmarshal the buyer in firestore to see if it's the buyer we want to delete
		var customerInRemoteStorage customer
		err = cdoc.DataTo(&customerInRemoteStorage)
		if err != nil {
			level.Error(fs.Logger).Log("err", &UnmarshalError{err: err})
			continue
		}

		if customerInRemoteStorage.Code == code {
			companyBuyer, err := fs.BuyerWithCompanyCode(customerInRemoteStorage.Code)
			if err == nil {
				fs.RemoveBuyer(ctx, companyBuyer.ID)
			}

			companySeller, err := fs.SellerWithCompanyCode(customerInRemoteStorage.Code)
			if err == nil {
				fs.RemoveSeller(ctx, companySeller.ID)
			}

			// Delete the customer in remote storage
			if _, err := cdoc.Ref.Delete(ctx); err != nil {
				return &FirestoreError{err: err}
			}

			// Delete the customer in cached storage
			fs.customerMutex.Lock()
			delete(fs.customers, code)
			fs.customerMutex.Unlock()

			fs.IncrementSequenceNumber(ctx)

			return nil
		}
	}

	return &DoesNotExistError{resourceType: "customer", resourceRef: code}
}

func (fs *Firestore) SetCustomer(ctx context.Context, c routing.Customer) error {
	// Get a copy of the customer in cached storage
	fs.customerMutex.RLock()
	customerInCachedStorage, ok := fs.customers[c.Code]
	fs.customerMutex.RUnlock()

	if !ok {
		return &DoesNotExistError{resourceType: "customer", resourceRef: c.Code}
	}

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

		// Unmarshal the customer in firestore to see if it's the customer we want to update
		var customerInRemoteStorage customer
		err = cdoc.DataTo(&customerInRemoteStorage)
		if err != nil {
			return &UnmarshalError{err: err}
		}

		// If the customer is the one we want to update, update it with the new data
		if customerInRemoteStorage.Code == c.Code {
			// Update the customer in firestore
			newCustomerData := map[string]interface{}{
				"name":                   c.Name,
				"code":                   c.Code,
				"buyerRef":               c.BuyerRef,
				"sellerRef":              c.SellerRef,
				"automaticSigninDomains": c.AutomaticSignInDomains,
				"active":                 c.Active,
			}

			if _, err := cdoc.Ref.Set(ctx, newCustomerData, firestore.MergeAll); err != nil {
				return &FirestoreError{err: err}
			}

			// Update the cached version
			customerInCachedStorage = c

			fs.customerMutex.Lock()
			fs.customers[c.Code] = customerInCachedStorage
			fs.customerMutex.Unlock()

			fs.IncrementSequenceNumber(ctx)

			return nil
		}
	}

	return &DoesNotExistError{resourceType: "customer", resourceRef: c.Code}
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

func (fs *Firestore) BuyerWithCompanyCode(code string) (routing.Buyer, error) {
	fs.buyerMutex.RLock()
	defer fs.buyerMutex.RUnlock()

	for _, buyer := range fs.buyers {
		if buyer.CompanyCode == code {
			return buyer, nil
		}
	}
	return routing.Buyer{}, &DoesNotExistError{resourceType: "buyer", resourceRef: code}
}

func (fs *Firestore) AddBuyer(ctx context.Context, b routing.Buyer) error {
	// Check if there is a company with the new buyer already / if there is a customer with the same code
	// Check if the buyer exists
	fs.buyerMutex.RLock()
	_, ok := fs.buyers[b.ID]
	fs.buyerMutex.RUnlock()

	var company routing.Customer
	for _, customer := range fs.customers {
		if customer.Code == b.CompanyCode {
			company = customer
		}
	}

	// If there is no company with that code error out
	if company.Code == "" {
		return &DoesNotExistError{resourceType: "customer", resourceRef: b.CompanyCode}
	}

	if ok {
		return &AlreadyExistsError{resourceType: "buyer", resourceRef: b.ID}
	}

	newBuyerData := buyer{
		CompanyCode: company.Code,
		ID:          int64(b.ID),
		Live:        b.Live,
		Debug:       b.Debug,
		PublicKey:   b.PublicKey,
	}

	// Add the buyer in remote storage
	ref, _, err := fs.Client.Collection("Buyer").Add(ctx, newBuyerData)
	if err != nil {
		return &FirestoreError{err: err}
	}

	// Add the buyer's routing rules settings to remote storage
	if err := fs.setRoutingRulesSettingsForBuyerID(ctx, ref.ID, company.Name, b.RoutingRulesSettings); err != nil {
		return &FirestoreError{err: err}
	}

	// Add the buyer's route shader to remote storage
	if err := fs.setRouteShaderForBuyerID(ctx, ref.ID, company.Name, b.RouteShader); err != nil {
		return &FirestoreError{err: err}
	}

	// Add the buyer's internal config to remote storage
	if err := fs.setInternalConfigForBuyerID(ctx, ref.ID, company.Name, b.InternalConfig); err != nil {
		return &FirestoreError{err: err}
	}

	company.BuyerRef = ref

	if err = fs.SetCustomer(ctx, company); err != nil {
		err = fmt.Errorf("AddBuyer() failed to update customer")
		return err
	}

	// Add the buyer in cached storage
	fs.buyerMutex.Lock()
	fs.buyers[b.ID] = b
	fs.buyerMutex.Unlock()

	fs.IncrementSequenceNumber(ctx)

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
			level.Error(fs.Logger).Log("err", &UnmarshalError{err: err})
			continue
		}

		if uint64(buyerInRemoteStorage.ID) == id {
			// Delete the buyer's routing rules settings in remote storage
			if err := fs.deleteRouteRulesSettingsForBuyerID(ctx, bdoc.Ref.ID); err != nil {
				return &FirestoreError{err: err}
			}

			// Delete the buyer's route shader in remote storage
			if err := fs.deleteRouteShaderForBuyerID(ctx, bdoc.Ref.ID); err != nil {
				return &FirestoreError{err: err}
			}

			// Delete the buyer's internal config in remote storage
			if err := fs.deleteInternalConfigForBuyerID(ctx, bdoc.Ref.ID); err != nil {
				return &FirestoreError{err: err}
			}

			// Delete the buyer in remote storage
			if _, err := bdoc.Ref.Delete(ctx); err != nil {
				return &FirestoreError{err: err}
			}
			associatedCustomer, err := fs.Customer(buyerInRemoteStorage.CompanyCode)
			if err != nil {
				err = fmt.Errorf("RemoveBuyer() failed to fetch customer")
				return err
			}

			associatedCustomer.BuyerRef = nil

			err = fs.SetCustomer(ctx, associatedCustomer)
			if err != nil {
				err = fmt.Errorf("RemoveBuyer() failed to update customer")
				return err
			}

			// Delete the buyer in cached storage
			fs.buyerMutex.Lock()
			delete(fs.buyers, id)
			fs.buyerMutex.Unlock()

			fs.IncrementSequenceNumber(ctx)

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
			level.Error(fs.Logger).Log("err", &UnmarshalError{err: err})
			continue
		}

		// If the buyer is the one we want to update, update it with the new data
		if uint64(buyerInRemoteStorage.ID) == b.ID {
			// Update the buyer in firestore
			newBuyerData := map[string]interface{}{
				"sdkVersion3PublicKeyId":   int64(b.ID),
				"companyCode":              b.CompanyCode,
				"isLiveCustomer":           b.Live,
				"isDebug":                  b.Debug,
				"sdkVersion3PublicKeyData": b.PublicKey,
			}

			if _, err := bdoc.Ref.Set(ctx, newBuyerData, firestore.MergeAll); err != nil {
				return &FirestoreError{err: err}
			}

			var company routing.Customer
			for _, customer := range fs.customers {
				if customer.BuyerRef != nil && customer.BuyerRef.ID == bdoc.Ref.ID {
					company = customer
				}
			}

			// Update the buyer's routing rules settings in firestore
			if err := fs.setRoutingRulesSettingsForBuyerID(ctx, bdoc.Ref.ID, company.Name, b.RoutingRulesSettings); err != nil {
				return &FirestoreError{err: err}
			}

			// Update the buyer's route shader in firestore
			if err := fs.setRouteShaderForBuyerID(ctx, bdoc.Ref.ID, company.Name, b.RouteShader); err != nil {
				return &FirestoreError{err: err}
			}

			// Update the buyer's internal config in firestore
			if err := fs.setInternalConfigForBuyerID(ctx, bdoc.Ref.ID, company.Name, b.InternalConfig); err != nil {
				return &FirestoreError{err: err}
			}

			// Update the cached version
			buyerInCachedStorage = b

			fs.buyerMutex.Lock()
			fs.buyers[b.ID] = buyerInCachedStorage
			fs.buyerMutex.Unlock()

			fs.IncrementSequenceNumber(ctx)

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

func (fs *Firestore) SellerWithCompanyCode(code string) (routing.Seller, error) {
	fs.sellerMutex.RLock()
	defer fs.sellerMutex.RUnlock()

	for _, seller := range fs.sellers {
		if seller.CompanyCode == code {
			return seller, nil
		}
	}
	return routing.Seller{}, &DoesNotExistError{resourceType: "seller", resourceRef: code}
}

func (fs *Firestore) AddSeller(ctx context.Context, s routing.Seller) error {
	// Check if the seller exists
	fs.sellerMutex.RLock()
	_, ok := fs.sellers[s.ID]
	fs.sellerMutex.RUnlock()

	var company routing.Customer
	for _, customer := range fs.customers {
		if customer.Code == s.CompanyCode {
			company = customer
		}
	}

	// If there is no company with that code error out
	if company.Code == "" {
		return &DoesNotExistError{resourceType: "customer", resourceRef: s.CompanyCode}
	}

	if ok {
		return &AlreadyExistsError{resourceType: "seller", resourceRef: s.ID}
	}

	newSellerData := seller{
		CompanyCode:               s.CompanyCode,
		ID:                        s.ID,
		Name:                      s.Name,
		IngressPriceNibblinsPerGB: int64(s.IngressPriceNibblinsPerGB),
		EgressPriceNibblinsPerGB:  int64(s.EgressPriceNibblinsPerGB),
	}

	// Add the seller in remote storage
	ref := fs.Client.Collection("Seller").Doc(s.ID)
	_, err := ref.Set(ctx, newSellerData)
	if err != nil {
		return &FirestoreError{err: err}
	}

	company.SellerRef = ref

	if err = fs.SetCustomer(ctx, company); err != nil {
		err = fmt.Errorf("AddSeller() failed to update customer")
		return err
	}

	// Add the seller in cached storage
	fs.sellerMutex.Lock()
	fs.sellers[s.ID] = s
	fs.sellerMutex.Unlock()

	fs.IncrementSequenceNumber(ctx)

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

		// Unmarshal the seller in firestore to see if it's the seller we want to delete
		var sellerInRemoteStorage seller
		err = sdoc.DataTo(&sellerInRemoteStorage)
		if err != nil {
			level.Error(fs.Logger).Log("err", &UnmarshalError{err: err})
			continue
		}

		if sellerInRemoteStorage.ID == id {
			associatedCustomer, err := fs.Customer(sellerInRemoteStorage.CompanyCode)
			if err != nil {
				err = fmt.Errorf("RemoveSeller() failed to fetch customer")
				return err
			}

			associatedCustomer.SellerRef = nil

			err = fs.SetCustomer(ctx, associatedCustomer)
			if err != nil {
				err = fmt.Errorf("RemoveSeller() failed to update customer")
				return err
			}

			// Delete the seller in remote storage
			if _, err := sdoc.Ref.Delete(ctx); err != nil {
				return &FirestoreError{err: err}
			}

			// Delete the seller in cached storage
			fs.sellerMutex.Lock()
			delete(fs.sellers, id)
			fs.sellerMutex.Unlock()

			fs.IncrementSequenceNumber(ctx)

			return nil
		}
	}

	return &DoesNotExistError{resourceType: "seller", resourceRef: id}
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
		"pricePublicIngressNibblins": int64(seller.IngressPriceNibblinsPerGB),
		"pricePublicEgressNibblins":  int64(seller.EgressPriceNibblinsPerGB),
	}

	if _, err := fs.Client.Collection("Seller").Doc(seller.ID).Set(ctx, newSellerData, firestore.MergeAll); err != nil {
		return &FirestoreError{err: err}
	}

	// Update the cached version
	sellerInCachedStorage = seller

	fs.sellerMutex.Lock()
	fs.sellers[seller.ID] = sellerInCachedStorage
	fs.sellerMutex.Unlock()

	fs.IncrementSequenceNumber(ctx)

	return nil
}

func (fs *Firestore) SetCustomerLink(ctx context.Context, companyCode string, buyerID uint64, sellerID string) error {
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
			level.Error(fs.Logger).Log("err", &UnmarshalError{err: err})
			continue
		}

		if c.Code == companyCode {
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
					level.Error(fs.Logger).Log("err", &UnmarshalError{err: err})
					continue
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
			c.BuyerRef = buyerRef
			c.SellerRef = sellerRef

			if _, err := cdoc.Ref.Set(ctx, c); err != nil {
				return &FirestoreError{err: err}
			}

			fs.IncrementSequenceNumber(ctx)

			return nil
		}
	}

	return &DoesNotExistError{resourceType: "customer", resourceRef: companyCode}
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
			level.Error(fs.Logger).Log("err", &UnmarshalError{err: err})
			continue
		}

		if c.Name == customerName {
			bdoc, err := c.BuyerRef.Get(ctx)
			if err != nil {
				return 0, &DoesNotExistError{resourceType: "buyer", resourceRef: c.BuyerRef.ID}
			}

			var b buyer
			if err := bdoc.DataTo(&b); err != nil {
				level.Error(fs.Logger).Log("err", &UnmarshalError{err: err})
				continue
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
			level.Error(fs.Logger).Log("err", &UnmarshalError{err: err})
			continue
		}
		if c.Name == customerName {
			sdoc, err := c.SellerRef.Get(ctx)
			if err != nil {
				return "", &DoesNotExistError{resourceType: "seller", resourceRef: c.SellerRef.ID}
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
			level.Error(fs.Logger).Log("err", &UnmarshalError{err: err})
			continue
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

	// Firestore docs save machineType as a string, currently
	var serverType string
	switch r.Type {
	case routing.BareMetal:
		serverType = "bare-metal"
	case routing.VirtualMachine:
		serverType = "vm"
	case routing.NoneSpecified:
		serverType = "n/a"
	default:
		serverType = "n/a"
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
		MRC:                int64(r.MRC),
		Overage:            int64(r.Overage),
		BWRule:             int32(r.BWRule),
		ContractTerm:       r.ContractTerm,
		StartDate:          r.StartDate.UTC(),
		EndDate:            r.EndDate.UTC(),
		Type:               serverType,
	}

	// Add the relay in remote storage
	if _, _, err := fs.Client.Collection("Relay").Add(ctx, newRelayData); err != nil {
		return &FirestoreError{err: err}
	}

	// Add the relay in cached storage
	fs.relayMutex.Lock()
	fs.relays[r.ID] = r
	fs.relayMutex.Unlock()

	fs.IncrementSequenceNumber(ctx)

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
			level.Error(fs.Logger).Log("err", &UnmarshalError{err: err})
			continue
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

			fs.IncrementSequenceNumber(ctx)

			return nil
		}
	}

	return &DoesNotExistError{resourceType: "relay", resourceRef: fmt.Sprintf("%x", id)}
}

// Only relay name, state, public key, and NIC speed are updated in firestore for now
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
			level.Error(fs.Logger).Log("err", &UnmarshalError{err: err})
			continue
		}

		// If the relay is the one we want to update, update it with the new data
		rid := crypto.HashID(relayInRemoteStorage.Address)
		if rid == r.ID {
			// Set the data to update the relay with
			newRelayData := map[string]interface{}{
				"name":            r.Name,
				"state":           r.State,
				"lastUpdateTime":  r.LastUpdateTime,
				"stateUpdateTime": time.Now(),
				"publicKey":       r.PublicKey,
				"nicSpeedMbps":    int64(r.NICSpeedMbps),
				"bandwidthRule":   int64(r.BWRule),
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

			fs.IncrementSequenceNumber(ctx)

			return nil
		}
	}

	return &DoesNotExistError{resourceType: "relay", resourceRef: fmt.Sprintf("%x", r.ID)}
}

func (fs *Firestore) GetDatacenterMapsForBuyer(buyerID uint64) map[uint64]routing.DatacenterMap {
	fs.datacenterMapMutex.RLock()
	defer fs.datacenterMapMutex.RUnlock()

	// buyer can have multiple dc aliases
	var dcs = make(map[uint64]routing.DatacenterMap)
	for _, dc := range fs.datacenterMaps {
		if dc.BuyerID == buyerID {
			id := crypto.HashID(dc.Alias + fmt.Sprintf("%x", dc.BuyerID) + fmt.Sprintf("%x", dc.Datacenter))
			dcs[id] = dc
		}
	}

	return dcs
}

func (fs *Firestore) AddDatacenterMap(ctx context.Context, dcMap routing.DatacenterMap) error {

	// ToDo: make sure buyer and datacenter exist?
	bID := dcMap.BuyerID

	dcID := dcMap.Datacenter

	if _, ok := fs.buyers[bID]; !ok {
		return &DoesNotExistError{resourceType: "BuyerID", resourceRef: dcMap.BuyerID}
	}

	if _, ok := fs.datacenters[dcID]; !ok {
		return &DoesNotExistError{resourceType: "Datacenter", resourceRef: dcMap.Datacenter}
	}

	dcMaps := fs.GetDatacenterMapsForBuyer(dcMap.BuyerID)
	if len(dcMaps) != 0 {
		for _, dc := range dcMaps {
			if dc.Alias == dcMap.Alias && dc.Datacenter == dcMap.Datacenter {
				return &AlreadyExistsError{resourceType: "datacenterMap", resourceRef: dcMap.Alias}
			}
		}
	}

	var dcMapInt64 datacenterMap
	dcMapInt64.Alias = dcMap.Alias
	dcMapInt64.Buyer = fmt.Sprintf("%016x", dcMap.BuyerID)
	dcMapInt64.Datacenter = fmt.Sprintf("%016x", dcMap.Datacenter)

	_, _, err := fs.Client.Collection("DatacenterMaps").Add(ctx, dcMapInt64)
	if err != nil {
		return &FirestoreError{err: err}
	}

	// update local store
	fs.datacenterMapMutex.Lock()
	id := crypto.HashID(dcMap.Alias + fmt.Sprintf("%x", dcMap.BuyerID) + fmt.Sprintf("%x", dcMap.Datacenter))
	fs.datacenterMaps[id] = dcMap
	fs.datacenterMapMutex.Unlock()

	fs.IncrementSequenceNumber(ctx)

	return nil

}

func (fs *Firestore) ListDatacenterMaps(dcID uint64) map[uint64]routing.DatacenterMap {
	fs.datacenterMapMutex.RLock()
	defer fs.datacenterMapMutex.RUnlock()

	var dcs = make(map[uint64]routing.DatacenterMap)
	for _, dc := range fs.datacenterMaps {
		if dc.Datacenter == dcID || dcID == 0 {
			id := crypto.HashID(dc.Alias + fmt.Sprintf("%x", dc.BuyerID) + fmt.Sprintf("%x", dc.Datacenter))
			dcs[id] = dc
		}
	}

	return dcs
}

func (fs *Firestore) RemoveDatacenterMap(ctx context.Context, dcMap routing.DatacenterMap) error {
	dmdocs := fs.Client.Collection("DatacenterMaps").Documents(ctx)
	defer dmdocs.Stop()

	// Firestore is the source of truth
	var dcm datacenterMap
	for {
		dmdoc, err := dmdocs.Next()
		ref := dmdoc.Ref
		if err == iterator.Done {
			break
		}

		if err != nil {
			return &FirestoreError{err: err}
		}

		err = dmdoc.DataTo(&dcm)
		if err != nil {
			level.Error(fs.Logger).Log("err", &UnmarshalError{err: err})
			continue
		}

		// all components must match (one-to-many)
		buyerID, err := strconv.ParseUint(dcm.Buyer, 16, 64)
		if err != nil {
			return &HexStringConversionError{hexString: dcm.Buyer}
		}
		datacenter, err := strconv.ParseUint(dcm.Datacenter, 16, 64)
		if err != nil {
			return &HexStringConversionError{hexString: dcm.Datacenter}
		}

		if dcMap.Alias == dcm.Alias && dcMap.BuyerID == buyerID && dcMap.Datacenter == datacenter {
			_, err := ref.Delete(ctx)
			if err != nil {
				return &FirestoreError{err: err}
			}

			// delete local copy as well
			fs.datacenterMapMutex.Lock()
			id := crypto.HashID(dcMap.Alias + fmt.Sprintf("%x", dcMap.BuyerID) + fmt.Sprintf("%x", dcMap.Datacenter))
			delete(fs.datacenterMaps, id)
			fs.datacenterMapMutex.Unlock()

			fs.IncrementSequenceNumber(ctx)

			return nil
		}
	}

	return &DoesNotExistError{resourceType: "datacenterMap", resourceRef: fmt.Sprintf("%v", dcMap)}
}

// UpdateRelay updates only the specified fields in the provided relay. The inputs are sanitized
// by the caller.
func (fs *Firestore) UpdateRelay(ctx context.Context, modifiedRelay routing.Relay, dirtyFields map[string]interface{}) error {

	query := fs.Client.Collection("Relay").Where("displayName", "==", modifiedRelay.Name)
	docs, err := query.Documents(ctx).GetAll()
	if err != nil {
		return &DatabaseError{dbErr: err, resourceType: "relay", resourceRef: fmt.Sprintf("%x", modifiedRelay.ID)}
	}
	if len(docs) > 1 {
		return &MultipleDBEntriesError{resourceType: "relay", resourceRef: fmt.Sprintf("%x", modifiedRelay.ID)}
	}

	// docs is a slice of length 1
	counter := 1
	for _, doc := range docs {
		fmt.Printf("found relay: %v\n", doc.Ref)
		fmt.Printf("counter    : %d\n", counter)
		// for key, value := range dirtyFields {
		// 	fmt.Printf("key: %s, value: %v\n", key, value)
		// 	_, err = doc.Ref.Update(ctx, []firestore.Update{{Path: "state", Value: routing.RelayStateDisabled}})
		// 	if err != nil {
		// 		return &DatabaseError{dbErr: err, resourceType: "relay", resourceRef: fmt.Sprintf("%x", modifiedRelay.ID)}
		// 	}
		// }
		if _, err := (*doc).Ref.Set(ctx, dirtyFields, firestore.MergeAll); err != nil {
			return &FirestoreError{err: err}
		}
		counter += 1
	}

	fs.relayMutex.Lock()
	fs.relays[modifiedRelay.ID] = modifiedRelay
	fs.relayMutex.Unlock()

	fs.IncrementSequenceNumber(ctx)

	return nil

}

func (fs *Firestore) Datacenter(id uint64) (routing.Datacenter, error) {
	fs.datacenterMutex.RLock()
	defer fs.datacenterMutex.RUnlock()

	d, found := fs.datacenters[id]
	if !found {
		return routing.Datacenter{}, &DoesNotExistError{resourceType: "datacenter", resourceRef: id}
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
		Name:         d.Name,
		Enabled:      d.Enabled,
		Latitude:     d.Location.Latitude,
		Longitude:    d.Location.Longitude,
		SupplierName: d.SupplierName,
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

	fs.IncrementSequenceNumber(ctx)

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
			level.Error(fs.Logger).Log("err", &UnmarshalError{err: err})
			continue
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

			fs.IncrementSequenceNumber(ctx)

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
			level.Error(fs.Logger).Log("err", &UnmarshalError{err: err})
			continue
		}

		// If the datacenter is the one we want to update, update it with the new data
		if crypto.HashID(datacenterInRemoteStorage.Name) == d.ID {
			// Set the data to update the datacenter with
			newDatacenterData := map[string]interface{}{
				"name":         d.Name,
				"enabled":      d.Enabled,
				"latitude":     d.Location.Latitude,
				"longitude":    d.Location.Longitude,
				"supplierName": d.SupplierName,
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

			fs.IncrementSequenceNumber(ctx)

			return nil
		}
	}

	return &DoesNotExistError{resourceType: "datacenter", resourceRef: fmt.Sprintf("%x", d.ID)}
}

// SyncLoop is a helper method that calls Sync
// func (fs *Firestore) SyncLoop(ctx context.Context, c <-chan time.Time) {
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

	seqNumberNotInSync, err := fs.CheckSequenceNumber(ctx)
	if err != nil {
		return err
	}
	if !seqNumberNotInSync {
		return nil
	}

	var outerErr error
	var wg sync.WaitGroup
	wg.Add(6)

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
		if err := fs.syncBuyers(ctx); err != nil {
			outerErr = fmt.Errorf("failed to sync buyers: %v", err)
		}
		wg.Done()
	}()

	go func() {
		if err := fs.syncSellers(ctx); err != nil {
			outerErr = fmt.Errorf("failed to sync sellers: %v", err)
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
			ID:      did,
			Name:    d.Name,
			Enabled: d.Enabled,
			Location: routing.Location{
				Latitude:  float64(d.Latitude),
				Longitude: float64(d.Longitude),
			},
			SupplierName: d.SupplierName,
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

		var bwRule routing.BandWidthRule
		switch r.BWRule {
		case 3:
			bwRule = routing.BWRulePool
		case 2:
			bwRule = routing.BWRuleBurst
		case 1:
			bwRule = routing.BWRuleFlat
		case 0:
			bwRule = routing.BWRuleNone
		default:
			bwRule = routing.BWRuleNone
		}

		var serverType routing.MachineType
		switch r.Type {
		case "vm":
			serverType = routing.VirtualMachine
		case "bare-metal":
			serverType = routing.BareMetal
		case "n/a":
			serverType = routing.NoneSpecified
		default:
			serverType = routing.NoneSpecified
		}

		relay := routing.Relay{
			ID:   rid,
			Name: r.Name,
			Addr: net.UDPAddr{
				IP:   net.ParseIP(host),
				Port: int(iport),
			},
			PublicKey:           publicKey,
			NICSpeedMbps:        int32(r.NICSpeedMbps),
			IncludedBandwidthGB: int32(r.IncludedBandwithGB),
			ManagementAddr:      r.ManagementAddress,
			SSHUser:             r.SSHUser,
			SSHPort:             r.SSHPort,
			State:               r.State,
			LastUpdateTime:      r.LastUpdateTime,
			MaxSessions:         uint32(r.MaxSessions),
			UpdateKey:           r.UpdateKey,
			FirestoreID:         rdoc.Ref.ID,
			MRC:                 routing.Nibblin(r.MRC),
			Overage:             routing.Nibblin(r.Overage),
			BWRule:              bwRule,
			ContractTerm:        r.ContractTerm,
			StartDate:           r.StartDate.UTC(),
			EndDate:             r.EndDate.UTC(),
			Type:                serverType,
		}

		// Set a default max session count of 3000 if the value isn't set in firestore
		if relay.MaxSessions == 0 {
			relay.MaxSessions = 3000
		}

		// Get datacenter
		ddoc, err := r.Datacenter.Get(ctx)
		if err != nil {
			level.Warn(fs.Logger).Log("msg", fmt.Sprintf("failed to get datacenter reference %s on relay %016x", r.Datacenter.ID, rid), "err", err)
			continue
		}

		var d datacenter
		err = ddoc.DataTo(&d)
		if err != nil {
			level.Warn(fs.Logger).Log("msg", fmt.Sprintf("failed to unmarshal datacenter %v", ddoc.Ref.ID), "err", err)
			continue
		}

		datacenter := routing.Datacenter{
			ID:      crypto.HashID(d.Name),
			Name:    d.Name,
			Enabled: d.Enabled,
			Location: routing.Location{
				Latitude:  d.Latitude,
				Longitude: d.Longitude,
			},
			SupplierName: d.SupplierName,
		}

		relay.Datacenter = datacenter

		// Get seller
		sdoc, err := r.Seller.Get(ctx)
		if err != nil {
			level.Warn(fs.Logger).Log("msg", fmt.Sprintf("failed to get seller reference %s on relay %016x", r.Datacenter.ID, rid), "err", err)
			continue
		}

		var s seller
		err = sdoc.DataTo(&s)
		if err != nil {
			level.Warn(fs.Logger).Log("msg", fmt.Sprintf("failed to unmarshal seller %v", sdoc.Ref.ID), "err", err)
			continue
		}

		seller := routing.Seller{
			ID:                        sdoc.Ref.ID,
			Name:                      s.Name,
			CompanyCode:               s.CompanyCode,
			IngressPriceNibblinsPerGB: routing.Nibblin(s.IngressPriceNibblinsPerGB),
			EgressPriceNibblinsPerGB:  routing.Nibblin(s.EgressPriceNibblinsPerGB),
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

	buyerDocs := fs.Client.Collection("Buyer").Documents(ctx)
	defer buyerDocs.Stop()

	for {
		buyerDoc, err := buyerDocs.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return &FirestoreError{err: err}
		}

		var b buyer
		err = buyerDoc.DataTo(&b)
		if err != nil {
			level.Warn(fs.Logger).Log("msg", fmt.Sprintf("failed to unmarshal buyer %v", buyerDoc.Ref.ID), "err", err)
			continue
		}
		rrs, err := fs.getRoutingRulesSettingsForBuyerID(ctx, buyerDoc.Ref.ID)
		if err != nil {
			level.Warn(fs.Logger).Log("msg", fmt.Sprintf("failed to completely read route shader for buyer %v, some fields will have default values", buyerDoc.Ref.ID), "err", err)
		}
		rs, err := fs.getRouteShaderForBuyerID(ctx, buyerDoc.Ref.ID)
		if err != nil {
			level.Warn(fs.Logger).Log("msg", fmt.Sprintf("failed to completely read route shader for buyer %v, some fields will have default values", buyerDoc.Ref.ID), "err", err)
		}

		ic, err := fs.getInternalConfigForBuyerID(ctx, buyerDoc.Ref.ID)
		if err != nil {
			level.Warn(fs.Logger).Log("msg", fmt.Sprintf("failed to completely read internal config for buyer %v, some fields will have default values", buyerDoc.Ref.ID), "err", err)
		}

		buyer := routing.Buyer{
			ID:                   uint64(b.ID),
			Live:                 b.Live,
			Debug:                b.Debug,
			CompanyCode:          b.CompanyCode,
			PublicKey:            b.PublicKey,
			RoutingRulesSettings: rrs,
			RouteShader:          rs,
			InternalConfig:       ic,
		}

		buyers[buyer.ID] = buyer
	}

	fs.buyerMutex.Lock()
	fs.buyers = buyers
	fs.buyerMutex.Unlock()

	level.Info(fs.Logger).Log("during", "syncBuyers", "num", len(fs.buyers))

	return nil

}

func (fs *Firestore) syncSellers(ctx context.Context) error {
	sellers := make(map[string]routing.Seller)

	sellerDocs := fs.Client.Collection("Seller").Documents(ctx)
	defer sellerDocs.Stop()

	for {
		sellerDoc, err := sellerDocs.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return &FirestoreError{err: err}
		}

		var s seller
		err = sellerDoc.DataTo(&s)
		if err != nil {
			level.Warn(fs.Logger).Log("msg", fmt.Sprintf("failed to unmarshal seller %v", sellerDoc.Ref.ID), "err", err)
			continue
		}

		seller := routing.Seller{
			ID:                        sellerDoc.Ref.ID,
			CompanyCode:               s.CompanyCode,
			Name:                      s.Name,
			IngressPriceNibblinsPerGB: routing.Nibblin(s.IngressPriceNibblinsPerGB),
			EgressPriceNibblinsPerGB:  routing.Nibblin(s.EgressPriceNibblinsPerGB),
		}

		sellers[sellerDoc.Ref.ID] = seller
	}

	fs.sellerMutex.Lock()
	fs.sellers = sellers
	fs.sellerMutex.Unlock()

	level.Info(fs.Logger).Log("during", "syncSellers", "num", len(fs.sellers))

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
		var dcMapInt64 datacenterMap
		err = dcdoc.DataTo(&dcMapInt64)
		if err != nil {
			level.Warn(fs.Logger).Log("msg", fmt.Sprintf("failed to unmarshal datacenterMap %v", dcdoc.Ref.ID), "err", err)
			continue
		}

		buyerID, err := strconv.ParseUint(dcMapInt64.Buyer, 16, 64)
		if err != nil {
			level.Error(fs.Logger).Log("msg", "could not parse buyerID on datacenter map", "buyerID", dcMapInt64.Buyer, "err", err)
			continue
		}

		datacenterID, err := strconv.ParseUint(dcMapInt64.Datacenter, 16, 64)
		if err != nil {
			level.Error(fs.Logger).Log("msg", "could not parse datacenterID on datacenter map", "datacenterID", dcMapInt64.Datacenter, "err", err)
			continue
		}

		dcMap.Alias = dcMapInt64.Alias
		dcMap.BuyerID = buyerID
		dcMap.Datacenter = datacenterID

		id := crypto.HashID(dcMap.Alias + fmt.Sprintf("%x", dcMap.BuyerID) + fmt.Sprintf("%x", dcMap.Datacenter))
		dcMaps[id] = dcMap
	}

	fs.datacenterMapMutex.Lock()
	fs.datacenterMaps = dcMaps
	fs.datacenterMapMutex.Unlock()
	return nil

}

func (fs *Firestore) syncCustomers(ctx context.Context) error {

	customers := make(map[string]routing.Customer)

	customerDocs := fs.Client.Collection("Customer").Documents(ctx)
	defer customerDocs.Stop()

	for {
		customerDoc, err := customerDocs.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return &FirestoreError{err: err}
		}

		var c customer
		err = customerDoc.DataTo(&c)
		if err != nil {
			level.Warn(fs.Logger).Log("msg", fmt.Sprintf("failed to unmarshal customer %v", customerDoc.Ref.ID), "err", err)
			continue
		}

		customer := routing.Customer{
			Code:                   c.Code,
			Name:                   c.Name,
			AutomaticSignInDomains: c.AutomaticSignInDomains,
			Active:                 c.Active,
			BuyerRef:               c.BuyerRef,
			SellerRef:              c.SellerRef,
		}

		customers[customer.Code] = customer
	}

	fs.customerMutex.Lock()
	fs.customers = customers
	fs.customerMutex.Unlock()

	level.Info(fs.Logger).Log("during", "syncCustomers", "num", len(fs.customers))

	return nil
}

func (fs *Firestore) deleteRouteRulesSettingsForBuyerID(ctx context.Context, ID string) error {
	// Comment below taken from old backend, at least attempting to explain why we need to append _0 (no existing entries have suffixes other than _0)
	// "Must be of the form '<buyer key>_<tag id>'. The buyer key can be found by looking at the ID under Buyer; it should be something like 763IMDH693HLsr2LGTJY. The tag ID should be 0 (for default) or the fnv64a hash of the tag the customer is using. Therefore this value should look something like: 763IMDH693HLsr2LGTJY_0. This value can not be changed after the entity is created."
	routeShaderID := ID + "_0"

	// Attempt to delete route shader for buyer
	_, err := fs.Client.Collection("RouteShader").Doc(routeShaderID).Delete(ctx)
	return err
}

func (fs *Firestore) deleteRouteShaderForBuyerID(ctx context.Context, ID string) error {
	routeShaderID := ID + "_0"

	// Attempt to delete route shader for buyer
	_, err := fs.Client.Collection("RouteShader4").Doc(routeShaderID).Delete(ctx)
	return err
}

func (fs *Firestore) deleteInternalConfigForBuyerID(ctx context.Context, ID string) error {
	internalConfigID := ID + "_0"

	// Attempt to delete route shader for buyer
	_, err := fs.Client.Collection("InternalConfig").Doc(internalConfigID).Delete(ctx)
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
	rrs.MaxNibblinsPerGB = routing.Nibblin(tempRRS.MaxPricePerGBNibblins)
	rrs.AcceptableLatency = tempRRS.AcceptableLatency
	rrs.RTTEpsilon = tempRRS.RTTEpsilon
	rrs.RTTThreshold = tempRRS.RTTThreshold
	rrs.RTTHysteresis = tempRRS.RTTHysteresis
	rrs.RTTVeto = tempRRS.RTTVeto
	rrs.EnableYouOnlyLiveOnce = tempRRS.EnableYouOnlyLiveOnce
	rrs.EnablePacketLossSafety = tempRRS.EnablePacketLossSafety
	rrs.EnableMultipathForPacketLoss = tempRRS.EnableMultipathForPacketLoss
	rrs.MultipathPacketLossThreshold = tempRRS.MultipathPacketLossThreshold
	rrs.EnableMultipathForJitter = tempRRS.EnableMultipathForJitter
	rrs.EnableMultipathForRTT = tempRRS.EnableMultipathForRTT
	rrs.EnableABTest = tempRRS.EnableABTest
	rrs.EnableTryBeforeYouBuy = tempRRS.EnableTryBeforeYouBuy
	rrs.TryBeforeYouBuyMaxSlices = tempRRS.TryBeforeYouBuyMaxSlices
	rrs.SelectionPercentage = tempRRS.SelectionPercentage

	rrs.ExcludedUserHashes = map[uint64]bool{}
	for userHashString := range tempRRS.ExcludedUserHashes {
		userHash, err := strconv.ParseUint(userHashString, 16, 64)
		if err != nil {
			return rrs, err
		}

		rrs.ExcludedUserHashes[userHash] = true
	}

	return rrs, nil
}

func (fs *Firestore) setRoutingRulesSettingsForBuyerID(ctx context.Context, ID string, name string, rrs routing.RoutingRulesSettings) error {
	// Comment below taken from old backend, at least attempting to explain why we need to append _0 (no existing entries have suffixes other than _0)
	// "Must be of the form '<buyer key>_<tag id>'. The buyer key can be found by looking at the ID under Buyer; it should be something like 763IMDH693HLsr2LGTJY. The tag ID should be 0 (for default) or the fnv64a hash of the tag the customer is using. Therefore this value should look something like: 763IMDH693HLsr2LGTJY_0. This value can not be changed after the entity is created."
	routeShaderID := ID + "_0"

	// Convert the excluded user hashes to strings
	excludedUserHashes := map[string]bool{}
	for userHash := range rrs.ExcludedUserHashes {
		excludedUserHashes[fmt.Sprintf("%016x", userHash)] = true
	}

	// Convert RoutingRulesSettings struct to firestore map
	rrsFirestore := map[string]interface{}{
		"displayName":                  name,
		"envelopeKbpsUp":               rrs.EnvelopeKbpsUp,
		"envelopeKbpsDown":             rrs.EnvelopeKbpsDown,
		"mode":                         rrs.Mode,
		"maxPricePerGBNibblins":        int64(rrs.MaxNibblinsPerGB),
		"acceptableLatency":            rrs.AcceptableLatency,
		"rttRouteSwitch":               rrs.RTTEpsilon,
		"rttThreshold":                 rrs.RTTThreshold,
		"rttHysteresis":                rrs.RTTHysteresis,
		"rttVeto":                      rrs.RTTVeto,
		"youOnlyLiveOnce":              rrs.EnableYouOnlyLiveOnce,
		"packetLossSafety":             rrs.EnablePacketLossSafety,
		"packetLossMultipath":          rrs.EnableMultipathForPacketLoss,
		"multipathPacketLossThreshold": rrs.MultipathPacketLossThreshold,
		"jitterMultipath":              rrs.EnableMultipathForJitter,
		"rttMultipath":                 rrs.EnableMultipathForRTT,
		"abTest":                       rrs.EnableABTest,
		"tryBeforeYouBuy":              rrs.EnableTryBeforeYouBuy,
		"tryBeforeYouBuyMaxSlices":     rrs.TryBeforeYouBuyMaxSlices,
		"selectionPercentage":          rrs.SelectionPercentage,
		"excludedUserHashes":           excludedUserHashes,
	}

	// Attempt to set route shader for buyer
	_, err := fs.Client.Collection("RouteShader").Doc(routeShaderID).Set(ctx, rrsFirestore, firestore.MergeAll)
	return err
}

func (fs *Firestore) getRouteShaderForBuyerID(ctx context.Context, buyerID string) (core.RouteShader, error) {
	routeShaderID := buyerID + "_0"
	rs := core.NewRouteShader()

	rsDoc, err := fs.Client.Collection("RouteShader4").Doc(routeShaderID).Get(ctx)
	if err != nil {
		return rs, err
	}

	var tempRS routeShader
	err = rsDoc.DataTo(&tempRS)
	if err != nil {
		return rs, err
	}

	rs.DisableNetworkNext = tempRS.DisableNetworkNext
	rs.SelectionPercent = tempRS.SelectionPercent
	rs.ABTest = tempRS.ABTest
	rs.ProMode = tempRS.ProMode
	rs.ReduceLatency = tempRS.ReduceLatency
	rs.ReducePacketLoss = tempRS.ReducePacketLoss
	rs.Multipath = tempRS.Multipath
	rs.AcceptableLatency = tempRS.AcceptableLatency
	rs.LatencyThreshold = tempRS.LatencyThreshold
	rs.AcceptablePacketLoss = tempRS.AcceptablePacketLoss
	rs.BandwidthEnvelopeUpKbps = tempRS.BandwidthEnvelopeUpKbps
	rs.BandwidthEnvelopeDownKbps = tempRS.BandwidthEnvelopeDownKbps

	// Convert user IDs from int64 to uint64
	bannedUsers := make(map[uint64]bool)
	for k, v := range tempRS.BannedUsers {
		rs.BannedUsers[uint64(k)] = v
	}
	rs.BannedUsers = bannedUsers

	return rs, nil
}

func (fs *Firestore) setRouteShaderForBuyerID(ctx context.Context, buyerID string, name string, routeShader core.RouteShader) error {
	routeShaderID := buyerID + "_0"

	// Convert user IDs from uint64 to int64
	bannedUsers := make(map[int64]bool)
	for k, v := range routeShader.BannedUsers {
		bannedUsers[int64(k)] = v
	}

	rsFirestore := map[string]interface{}{
		"displayName":               name,
		"disableNetworkNext":        routeShader.DisableNetworkNext,
		"selectionPercent":          routeShader.SelectionPercent,
		"abTest":                    routeShader.ABTest,
		"proMode":                   routeShader.ProMode,
		"reduceLatency":             routeShader.ReduceLatency,
		"reducePacketLoss":          routeShader.ReducePacketLoss,
		"multipath":                 routeShader.Multipath,
		"acceptableLatency":         routeShader.AcceptableLatency,
		"latencyThreshold":          routeShader.LatencyThreshold,
		"acceptablePacketLoss":      routeShader.AcceptablePacketLoss,
		"bandwidthEnvelopeUpKbps":   routeShader.BandwidthEnvelopeUpKbps,
		"bandwidthEnvelopeDownKbps": routeShader.BandwidthEnvelopeDownKbps,
		"bannedUsers":               bannedUsers,
	}

	_, err := fs.Client.Collection("RouteShader4").Doc(routeShaderID).Set(ctx, rsFirestore, firestore.MergeAll)
	return err
}

func (fs *Firestore) getInternalConfigForBuyerID(ctx context.Context, buyerID string) (core.InternalConfig, error) {
	internalConfigID := buyerID + "_0"
	ic := core.NewInternalConfig()

	icDoc, err := fs.Client.Collection("InternalConfig").Doc(internalConfigID).Get(ctx)
	if err != nil {
		return ic, err
	}

	var tempIC internalConfig
	err = icDoc.DataTo(&tempIC)
	if err != nil {
		return ic, err
	}

	ic.RouteSwitchThreshold = tempIC.RouteSwitchThreshold
	ic.MaxLatencyTradeOff = tempIC.MaxLatencyTradeOff
	ic.RTTVeto_Default = tempIC.RTTVeto_Default
	ic.RTTVeto_PacketLoss = tempIC.RTTVeto_PacketLoss
	ic.RTTVeto_Multipath = tempIC.RTTVeto_Multipath
	ic.MultipathOverloadThreshold = tempIC.MultipathOverloadThreshold
	ic.TryBeforeYouBuy = tempIC.TryBeforeYouBuy

	return ic, nil
}

func (fs *Firestore) setInternalConfigForBuyerID(ctx context.Context, buyerID string, name string, internalConfig core.InternalConfig) error {
	internalConfigID := buyerID + "_0"

	icFirestore := map[string]interface{}{
		"displayName":                name,
		"routeSwitchThreshold":       internalConfig.RouteSwitchThreshold,
		"maxLatencyTradeOff":         internalConfig.MaxLatencyTradeOff,
		"rttVeto_default":            internalConfig.RTTVeto_Default,
		"rttVeto_packetLoss":         internalConfig.RTTVeto_PacketLoss,
		"rttVeto_multipath":          internalConfig.RTTVeto_Multipath,
		"multipathOverloadThreshold": internalConfig.MultipathOverloadThreshold,
		"tryBeforeYouBuy":            internalConfig.TryBeforeYouBuy,
	}

	_, err := fs.Client.Collection("InternalConfig").Doc(internalConfigID).Set(ctx, icFirestore, firestore.MergeAll)
	return err
}
