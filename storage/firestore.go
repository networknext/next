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

		s.buyers[uint64(b.ID)] = &routing.Buyer{
			ID:        uint64(b.ID),
			Name:      b.Name,
			Active:    b.Active,
			Live:      b.Live,
			PublicKey: b.PublicKey,
		}
	}

	level.Debug(s.Logger).Log("during", "syncBuyers", "num", len(s.buyers))

	return nil
}
