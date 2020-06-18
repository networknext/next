package storage_test

import (
	"context"
	"fmt"
	"net"
	"os"
	"testing"

	"cloud.google.com/go/firestore"
	"github.com/go-kit/kit/log"
	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
	"github.com/stretchr/testify/assert"
	"google.golang.org/api/iterator"
)

type customer struct {
	Name   string                 `firestore:"name"`
	Domain string                 `firestore:"automaticSigninDomain"`
	Active bool                   `firestore:"active"`
	Buyer  *firestore.DocumentRef `firestore:"buyer"`
	Seller *firestore.DocumentRef `firestore:"seller"`
}

func checkFirestoreEmulator(t *testing.T) {
	firestoreEmulatorHost := os.Getenv("FIRESTORE_EMULATOR_HOST")
	if firestoreEmulatorHost == "" {
		t.Skip("Firestore emulator not set up, skipping firestore test")
	}
}

func cleanFireStore(ctx context.Context, client *firestore.Client) error {
	collections := client.Collections(ctx)
	for {
		collection, err := collections.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return err
		}

		documents := collection.DocumentRefs(ctx)
		for {
			document, err := documents.Next()
			if err == iterator.Done {
				break
			}

			if err != nil {
				return err
			}

			if _, err = document.Delete(ctx); err != nil {
				return err
			}
		}
	}

	return nil
}

func TestFirestore(t *testing.T) {
	checkFirestoreEmulator(t)
	ctx := context.Background()

	t.Run("NewFirestore", func(t *testing.T) {
		t.Run("firestore client failure", func(t *testing.T) {
			_, err := storage.NewFirestore(ctx, "*detect-project-id*", log.NewNopLogger())
			assert.Error(t, err)
		})

		t.Run("success", func(t *testing.T) {
			projectID := "default"
			client, err := firestore.NewClient(ctx, projectID)
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, client)
				assert.NoError(t, err)
			}()

			logger := log.NewNopLogger()

			expected := storage.Firestore{
				Client: client,
				Logger: logger,
			}

			actual, err := storage.NewFirestore(ctx, projectID, logger)
			assert.NoError(t, err)

			assert.Equal(t, expected.Logger, actual.Logger)
		})
	})

	t.Run("Buyer", func(t *testing.T) {
		t.Run("buyer not found", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger())
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			buyer, err := fs.Buyer(0)
			assert.Empty(t, buyer)
			assert.EqualError(t, err, "buyer with reference 0 not found")
		})

		t.Run("success", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger())
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			expected := routing.Buyer{
				ID:                   1,
				Name:                 "local",
				Active:               true,
				Live:                 false,
				PublicKey:            make([]byte, crypto.KeySize),
				RoutingRulesSettings: routing.DefaultRoutingRulesSettings,
			}

			err = fs.AddBuyer(ctx, expected)
			assert.NoError(t, err)

			actual, err := fs.Buyer(expected.ID)
			assert.NoError(t, err)

			assert.Equal(t, expected, actual)
		})
	})

	t.Run("Buyers", func(t *testing.T) {
		fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger())
		assert.NoError(t, err)

		defer func() {
			err := cleanFireStore(ctx, fs.Client)
			assert.NoError(t, err)
		}()

		expected := []routing.Buyer{
			{
				ID:                   1,
				Name:                 "local",
				Active:               true,
				Live:                 false,
				PublicKey:            make([]byte, crypto.KeySize),
				RoutingRulesSettings: routing.DefaultRoutingRulesSettings,
			},
			{
				ID:                   2,
				Name:                 "local",
				Active:               false,
				Live:                 true,
				PublicKey:            make([]byte, crypto.KeySize),
				RoutingRulesSettings: routing.LocalRoutingRulesSettings,
			},
		}

		for i := 0; i < len(expected); i++ {
			err = fs.AddBuyer(ctx, expected[i])
			assert.NoError(t, err)
		}

		actual := fs.Buyers()
		assert.Equal(t, expected, actual)
	})

	t.Run("AddBuyer", func(t *testing.T) {
		t.Run("new customer", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger())
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			expected := routing.Buyer{
				ID:                   1,
				Name:                 "local",
				Domain:               "example.com",
				Active:               true,
				Live:                 false,
				PublicKey:            make([]byte, crypto.KeySize),
				RoutingRulesSettings: routing.DefaultRoutingRulesSettings,
			}

			expectedCustomer := customer{
				Name:   expected.Name,
				Domain: expected.Domain,
				Active: expected.Active,
			}

			err = fs.AddBuyer(ctx, expected)
			assert.NoError(t, err)

			actual, err := fs.Buyer(expected.ID)
			assert.NoError(t, err)

			assert.Equal(t, expected, actual)

			// Check that the customer exists and is properly linked to the buyer

			// Grab the customer
			cdocs := fs.Client.Collection("Customer").Documents(ctx)

			cdoc, err := cdocs.Next()
			assert.NoError(t, err)

			var customerInRemoteStorage customer
			err = cdoc.DataTo(&customerInRemoteStorage)
			assert.NoError(t, err)

			cdocs.Stop()

			// Grab the buyer to compare the reference on the customer
			bdocs := fs.Client.Collection("Buyer").Documents(ctx)

			bdoc, err := bdocs.Next()
			assert.NoError(t, err)

			bdocs.Stop()

			expectedCustomer.Buyer = bdoc.Ref

			assert.Equal(t, expectedCustomer, customerInRemoteStorage)
		})

		t.Run("existing customer", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger())
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			expected := routing.Buyer{
				ID:                   1,
				Name:                 "local",
				Domain:               "example.com",
				Active:               true,
				Live:                 false,
				PublicKey:            make([]byte, crypto.KeySize),
				RoutingRulesSettings: routing.DefaultRoutingRulesSettings,
			}

			expectedCustomer := customer{
				Name:   expected.Name,
				Domain: expected.Domain,
				Active: expected.Active,
			}

			// Add the preexisting customer
			_, _, err = fs.Client.Collection("Customer").Add(ctx, expectedCustomer)
			assert.NoError(t, err)

			err = fs.AddBuyer(ctx, expected)
			assert.NoError(t, err)

			actual, err := fs.Buyer(expected.ID)
			assert.NoError(t, err)

			assert.Equal(t, expected, actual)

			// Check that the customer exists and is properly linked to the buyer

			// Grab the customer
			cdocs := fs.Client.Collection("Customer").Documents(ctx)

			cdoc, err := cdocs.Next()
			assert.NoError(t, err)

			var customerInRemoteStorage customer
			err = cdoc.DataTo(&customerInRemoteStorage)
			assert.NoError(t, err)

			// Check to make sure this is the only customer
			cdoc, err = cdocs.Next()
			assert.Error(t, err)

			cdocs.Stop()

			// Grab the buyer to compare the reference on the customer
			bdocs := fs.Client.Collection("Buyer").Documents(ctx)

			bdoc, err := bdocs.Next()
			assert.NoError(t, err)

			bdocs.Stop()

			expectedCustomer.Buyer = bdoc.Ref

			assert.Equal(t, expectedCustomer, customerInRemoteStorage)
		})
	})

	t.Run("RemoveBuyer", func(t *testing.T) {
		t.Run("buyer not found", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger())
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			err = fs.RemoveBuyer(ctx, 0)
			assert.EqualError(t, err, "buyer with reference 0 not found")
		})

		t.Run("success - update existing customer", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger())
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			buyer := routing.Buyer{
				ID:                   1,
				Name:                 "local",
				Active:               true,
				Live:                 false,
				PublicKey:            make([]byte, crypto.KeySize),
				RoutingRulesSettings: routing.DefaultRoutingRulesSettings,
			}

			err = fs.AddBuyer(ctx, buyer)
			assert.NoError(t, err)

			// Add a seller so that the customer isn't removed
			seller := routing.Seller{
				ID:   "sellerID",
				Name: "seller name",
			}

			err = fs.AddSeller(ctx, seller)
			assert.NoError(t, err)

			err = fs.RemoveBuyer(ctx, buyer.ID)
			assert.NoError(t, err)

			// Check that the customer was updated successfully
			cdocs := fs.Client.Collection("Customer").Documents(ctx)

			cdoc, err := cdocs.Next()
			assert.NoError(t, err)

			var customerInRemoteStorage customer
			err = cdoc.DataTo(&customerInRemoteStorage)
			assert.NoError(t, err)

			assert.Nil(t, customerInRemoteStorage.Buyer)
		})

		t.Run("success - removed customer", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger())
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			buyer := routing.Buyer{
				ID:                   1,
				Name:                 "local",
				Active:               true,
				Live:                 false,
				PublicKey:            make([]byte, crypto.KeySize),
				RoutingRulesSettings: routing.DefaultRoutingRulesSettings,
			}

			err = fs.AddBuyer(ctx, buyer)
			assert.NoError(t, err)

			err = fs.RemoveBuyer(ctx, buyer.ID)
			assert.NoError(t, err)

			// Check that the customer was removed successfully
			cdocs := fs.Client.Collection("Customer").Documents(ctx)

			_, err = cdocs.Next()
			assert.Error(t, err)
		})
	})

	t.Run("SetBuyer", func(t *testing.T) {
		t.Run("buyer not found", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger())
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			buyer := routing.Buyer{
				ID:                   1,
				Name:                 "local",
				Active:               true,
				Live:                 false,
				PublicKey:            make([]byte, crypto.KeySize),
				RoutingRulesSettings: routing.DefaultRoutingRulesSettings,
			}

			err = fs.SetBuyer(ctx, buyer)
			assert.EqualError(t, err, "buyer with reference 1 not found")
		})

		t.Run("success", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger())
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			expected := routing.Buyer{
				ID:                   1,
				Name:                 "local",
				Active:               true,
				Live:                 false,
				PublicKey:            make([]byte, crypto.KeySize),
				RoutingRulesSettings: routing.DefaultRoutingRulesSettings,
			}

			err = fs.AddBuyer(ctx, expected)
			assert.NoError(t, err)

			actual := expected
			actual.Active = false
			actual.Live = true

			err = fs.SetBuyer(ctx, actual)
			assert.NoError(t, err)

			actual, err = fs.Buyer(expected.ID)
			assert.NoError(t, err)

			assert.NotEqual(t, expected, actual)
			actual.Active = true
			actual.Live = false
			assert.Equal(t, expected, actual)
		})
	})

	t.Run("Seller", func(t *testing.T) {
		t.Run("seller not found", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger())
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			seller, err := fs.Seller("id")
			assert.Empty(t, seller)
			assert.EqualError(t, err, "seller with reference id not found")
		})

		t.Run("success", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger())
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			expected := routing.Seller{
				ID:   "id",
				Name: "local",
			}

			err = fs.AddSeller(ctx, expected)
			assert.NoError(t, err)

			actual, err := fs.Seller(expected.ID)
			assert.NoError(t, err)

			assert.Equal(t, expected, actual)
		})
	})

	t.Run("Sellers", func(t *testing.T) {
		fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger())
		assert.NoError(t, err)

		defer func() {
			err := cleanFireStore(ctx, fs.Client)
			assert.NoError(t, err)
		}()

		expected := []routing.Seller{
			{
				ID:                "id1",
				Name:              "local",
				IngressPriceCents: 10,
				EgressPriceCents:  20,
			},
			{
				ID:                "id2",
				Name:              "local",
				IngressPriceCents: 10,
				EgressPriceCents:  20,
			},
		}

		for i := 0; i < len(expected); i++ {
			err = fs.AddSeller(ctx, expected[i])
			assert.NoError(t, err)
		}

		actual := fs.Sellers()
		assert.Equal(t, expected, actual)
	})

	t.Run("AddSeller", func(t *testing.T) {
		t.Run("seller already exists", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger())
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			expected := routing.Seller{
				ID:                "id",
				Name:              "local",
				IngressPriceCents: 10,
				EgressPriceCents:  20,
			}

			err = fs.AddSeller(ctx, expected)
			assert.NoError(t, err)

			err = fs.AddSeller(ctx, expected)
			assert.EqualError(t, err, "seller with reference id already exists")
		})

		t.Run("success - add new customer", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger())
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			expected := routing.Seller{
				ID:                "id",
				Name:              "local",
				IngressPriceCents: 10,
				EgressPriceCents:  20,
			}

			expectedCustomer := customer{
				Name:   expected.Name,
				Active: true,
			}

			err = fs.AddSeller(ctx, expected)
			assert.NoError(t, err)

			actual, err := fs.Seller(expected.ID)
			assert.NoError(t, err)

			assert.Equal(t, expected, actual)

			// Check that the customer exists and is properly linked to the seller

			// Grab the customer
			cdocs := fs.Client.Collection("Customer").Documents(ctx)

			cdoc, err := cdocs.Next()
			assert.NoError(t, err)

			var customerInRemoteStorage customer
			err = cdoc.DataTo(&customerInRemoteStorage)
			assert.NoError(t, err)

			cdocs.Stop()

			// Grab the seller to compare the reference on the customer
			sdocs := fs.Client.Collection("Seller").Documents(ctx)

			sdoc, err := sdocs.Next()
			assert.NoError(t, err)

			sdocs.Stop()

			expectedCustomer.Seller = sdoc.Ref

			assert.Equal(t, expectedCustomer, customerInRemoteStorage)
		})

		t.Run("success - update existing customer", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger())
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			expected := routing.Seller{
				ID:                "id",
				Name:              "local",
				IngressPriceCents: 10,
				EgressPriceCents:  20,
			}

			expectedCustomer := customer{
				Name:   expected.Name,
				Active: true,
			}

			// Add the preexisting customer
			_, _, err = fs.Client.Collection("Customer").Add(ctx, expectedCustomer)
			assert.NoError(t, err)

			err = fs.AddSeller(ctx, expected)
			assert.NoError(t, err)

			actual, err := fs.Seller(expected.ID)
			assert.NoError(t, err)

			assert.Equal(t, expected, actual)

			// Check that the customer exists and is properly linked to the seller

			// Grab the customer
			cdocs := fs.Client.Collection("Customer").Documents(ctx)

			cdoc, err := cdocs.Next()
			assert.NoError(t, err)

			var customerInRemoteStorage customer
			err = cdoc.DataTo(&customerInRemoteStorage)
			assert.NoError(t, err)

			// Check to make sure this is the only customer
			cdoc, err = cdocs.Next()
			assert.Error(t, err)

			cdocs.Stop()

			// Grab the seller to compare the reference on the customer
			sdocs := fs.Client.Collection("Seller").Documents(ctx)

			sdoc, err := sdocs.Next()
			assert.NoError(t, err)

			sdocs.Stop()

			expectedCustomer.Seller = sdoc.Ref

			assert.Equal(t, expectedCustomer, customerInRemoteStorage)
		})
	})

	t.Run("RemoveSeller", func(t *testing.T) {
		t.Run("seller not found", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger())
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			err = fs.RemoveSeller(ctx, "id")
			assert.EqualError(t, err, "seller with reference id not found")
		})

		t.Run("success - removed customer", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger())
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			seller := routing.Seller{
				ID:                "id",
				Name:              "local",
				IngressPriceCents: 10,
				EgressPriceCents:  20,
			}

			err = fs.AddSeller(ctx, seller)
			assert.NoError(t, err)

			err = fs.RemoveSeller(ctx, seller.ID)
			assert.NoError(t, err)

			// Check that the customer was removed successfully
			cdocs := fs.Client.Collection("Customer").Documents(ctx)

			_, err = cdocs.Next()
			assert.Error(t, err)
		})

		t.Run("success - update existing customer", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger())
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			seller := routing.Seller{
				ID:                "id",
				Name:              "local",
				IngressPriceCents: 10,
				EgressPriceCents:  20,
			}

			err = fs.AddSeller(ctx, seller)
			assert.NoError(t, err)

			// Add a buyer so that the customer isn't removed
			buyer := routing.Buyer{
				ID:   1,
				Name: "buyer name",
			}

			err = fs.AddBuyer(ctx, buyer)
			assert.NoError(t, err)

			err = fs.RemoveSeller(ctx, seller.ID)
			assert.NoError(t, err)

			// Check that the customer was updated successfully
			cdocs := fs.Client.Collection("Customer").Documents(ctx)

			cdoc, err := cdocs.Next()
			assert.NoError(t, err)

			var customerInRemoteStorage customer
			err = cdoc.DataTo(&customerInRemoteStorage)
			assert.NoError(t, err)

			assert.Nil(t, customerInRemoteStorage.Seller)
		})
	})

	t.Run("SetSeller", func(t *testing.T) {
		t.Run("seller not found", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger())
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			seller := routing.Seller{
				ID:                "id",
				Name:              "local",
				IngressPriceCents: 10,
				EgressPriceCents:  20,
			}

			err = fs.SetSeller(ctx, seller)
			assert.EqualError(t, err, "seller with reference id not found")
		})

		t.Run("success", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger())
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			expected := routing.Seller{
				ID:                "id",
				Name:              "local",
				IngressPriceCents: 10,
				EgressPriceCents:  20,
			}

			err = fs.AddSeller(ctx, expected)
			assert.NoError(t, err)

			actual := expected
			actual.IngressPriceCents = 20
			actual.EgressPriceCents = 10

			err = fs.SetSeller(ctx, actual)
			assert.NoError(t, err)

			actual, err = fs.Seller(expected.ID)
			assert.NoError(t, err)

			assert.NotEqual(t, expected, actual)
			actual.IngressPriceCents = 10
			actual.EgressPriceCents = 20
			assert.Equal(t, expected, actual)
		})
	})

	t.Run("Relay", func(t *testing.T) {
		t.Run("relay not found", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger())
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			relay, err := fs.Relay(0)
			assert.Empty(t, relay)
			assert.EqualError(t, err, "relay with reference 0 not found")
		})

		t.Run("success", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger())
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
			assert.NoError(t, err)

			seller := routing.Seller{
				ID:                "seller ID",
				Name:              "seller name",
				IngressPriceCents: 10,
				EgressPriceCents:  20,
			}

			datacenter := routing.Datacenter{
				ID:      crypto.HashID("datacenter name"),
				Name:    "datacenter name",
				Enabled: true,
				Location: routing.Location{
					Latitude:  70.5,
					Longitude: 120.5,
				},
			}

			expected := routing.Relay{
				ID:         1,
				Name:       "local",
				Addr:       *addr,
				PublicKey:  make([]byte, crypto.KeySize),
				Seller:     seller,
				Datacenter: datacenter,
			}

			err = fs.AddSeller(ctx, seller)
			assert.NoError(t, err)

			err = fs.AddDatacenter(ctx, datacenter)
			assert.NoError(t, err)

			err = fs.AddRelay(ctx, expected)
			assert.NoError(t, err)

			actual, err := fs.Relay(expected.ID)
			assert.NoError(t, err)

			assert.Equal(t, expected, actual)
		})
	})

	t.Run("Relays", func(t *testing.T) {
		fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger())
		assert.NoError(t, err)

		defer func() {
			err := cleanFireStore(ctx, fs.Client)
			assert.NoError(t, err)
		}()

		addr1, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
		assert.NoError(t, err)

		addr2, err := net.ResolveUDPAddr("udp", "127.0.0.2:40000")
		assert.NoError(t, err)

		seller := routing.Seller{
			ID:                "seller ID",
			Name:              "seller name",
			IngressPriceCents: 10,
			EgressPriceCents:  20,
		}

		datacenter := routing.Datacenter{
			ID:      crypto.HashID("datacenter name"),
			Name:    "datacenter name",
			Enabled: true,
			Location: routing.Location{
				Latitude:  70.5,
				Longitude: 120.5,
			},
		}

		expected := []routing.Relay{
			{
				ID:         1,
				Name:       "local",
				Addr:       *addr1,
				PublicKey:  make([]byte, crypto.KeySize),
				Seller:     seller,
				Datacenter: datacenter,
			},
			{
				ID:         2,
				Name:       "local",
				Addr:       *addr2,
				PublicKey:  make([]byte, crypto.KeySize),
				Seller:     seller,
				Datacenter: datacenter,
			},
		}

		err = fs.AddSeller(ctx, seller)
		assert.NoError(t, err)

		err = fs.AddDatacenter(ctx, datacenter)
		assert.NoError(t, err)

		for i := 0; i < len(expected); i++ {
			err = fs.AddRelay(ctx, expected[i])
			assert.NoError(t, err)
		}

		actual := fs.Relays()
		assert.Equal(t, expected, actual)
	})

	t.Run("AddRelay", func(t *testing.T) {
		t.Run("seller not found", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger())
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
			assert.NoError(t, err)

			expected := routing.Relay{
				ID:        1,
				Name:      "local",
				Addr:      *addr,
				PublicKey: make([]byte, crypto.KeySize),
			}

			err = fs.AddRelay(ctx, expected)
			assert.EqualError(t, err, "seller with reference  not found")
		})

		t.Run("datacenter not found", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger())
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
			assert.NoError(t, err)

			seller := routing.Seller{
				ID:                "seller ID",
				Name:              "seller name",
				IngressPriceCents: 10,
				EgressPriceCents:  20,
			}

			expected := routing.Relay{
				ID:        1,
				Name:      "local",
				Addr:      *addr,
				PublicKey: make([]byte, crypto.KeySize),
				Seller:    seller,
			}

			err = fs.AddSeller(ctx, seller)
			assert.NoError(t, err)

			err = fs.AddRelay(ctx, expected)
			assert.EqualError(t, err, "datacenter with reference 0 not found")
		})

		t.Run("success", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger())
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
			assert.NoError(t, err)

			seller := routing.Seller{
				ID:                "seller ID",
				Name:              "seller name",
				IngressPriceCents: 10,
				EgressPriceCents:  20,
			}

			datacenter := routing.Datacenter{
				ID:      crypto.HashID("datacenter name"),
				Name:    "datacenter name",
				Enabled: true,
				Location: routing.Location{
					Latitude:  70.5,
					Longitude: 120.5,
				},
			}

			expected := routing.Relay{
				ID:         1,
				Name:       "local",
				Addr:       *addr,
				PublicKey:  make([]byte, crypto.KeySize),
				Seller:     seller,
				Datacenter: datacenter,
			}

			err = fs.AddSeller(ctx, seller)
			assert.NoError(t, err)

			err = fs.AddDatacenter(ctx, datacenter)
			assert.NoError(t, err)

			err = fs.AddRelay(ctx, expected)
			assert.NoError(t, err)

			actual, err := fs.Relay(expected.ID)
			assert.NoError(t, err)

			assert.Equal(t, expected, actual)
		})
	})

	t.Run("RemoveRelay", func(t *testing.T) {
		t.Run("relay not found", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger())
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			err = fs.RemoveRelay(ctx, 0)
			assert.EqualError(t, err, "relay with reference 0 not found")
		})

		t.Run("success", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger())
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
			assert.NoError(t, err)

			seller := routing.Seller{
				ID:                "seller ID",
				Name:              "seller name",
				IngressPriceCents: 10,
				EgressPriceCents:  20,
			}

			datacenter := routing.Datacenter{
				ID:      crypto.HashID("datacenter name"),
				Name:    "datacenter name",
				Enabled: true,
				Location: routing.Location{
					Latitude:  70.5,
					Longitude: 120.5,
				},
			}

			relay := routing.Relay{
				ID:         crypto.HashID(addr.String()),
				Name:       "local",
				Addr:       *addr,
				PublicKey:  make([]byte, crypto.KeySize),
				Seller:     seller,
				Datacenter: datacenter,
			}

			err = fs.AddSeller(ctx, seller)
			assert.NoError(t, err)

			err = fs.AddDatacenter(ctx, datacenter)
			assert.NoError(t, err)

			err = fs.AddRelay(ctx, relay)
			assert.NoError(t, err)

			err = fs.RemoveRelay(ctx, relay.ID)
			assert.NoError(t, err)
		})
	})

	t.Run("SetRelay", func(t *testing.T) {
		t.Run("relay not found", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger())
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
			assert.NoError(t, err)

			relay := routing.Relay{
				ID:        1,
				Name:      "local",
				Addr:      *addr,
				PublicKey: make([]byte, crypto.KeySize),
			}

			err = fs.SetRelay(ctx, relay)
			assert.EqualError(t, err, "relay with reference 1 not found")
		})

		t.Run("seller not found", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger())
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
			assert.NoError(t, err)

			seller := routing.Seller{
				ID:                "seller ID",
				Name:              "seller name",
				IngressPriceCents: 10,
				EgressPriceCents:  20,
			}

			datacenter := routing.Datacenter{
				ID:      crypto.HashID("datacenter name"),
				Name:    "datacenter name",
				Enabled: true,
				Location: routing.Location{
					Latitude:  70.5,
					Longitude: 120.5,
				},
			}

			expected := routing.Relay{
				ID:         crypto.HashID(addr.String()),
				Name:       "local",
				Addr:       *addr,
				PublicKey:  make([]byte, crypto.KeySize),
				Seller:     seller,
				Datacenter: datacenter,
				State:      routing.RelayStateEnabled,
			}

			err = fs.AddSeller(ctx, seller)
			assert.NoError(t, err)

			err = fs.AddDatacenter(ctx, datacenter)
			assert.NoError(t, err)

			err = fs.AddRelay(ctx, expected)
			assert.NoError(t, err)

			actual := expected
			actual.Seller = routing.Seller{}

			err = fs.SetRelay(ctx, actual)
			assert.EqualError(t, err, "seller with reference  not found")
		})

		t.Run("datacenter not found", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger())
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
			assert.NoError(t, err)

			seller := routing.Seller{
				ID:                "seller ID",
				Name:              "seller name",
				IngressPriceCents: 10,
				EgressPriceCents:  20,
			}

			datacenter := routing.Datacenter{
				ID:      crypto.HashID("datacenter name"),
				Name:    "datacenter name",
				Enabled: true,
				Location: routing.Location{
					Latitude:  70.5,
					Longitude: 120.5,
				},
			}

			expected := routing.Relay{
				ID:         crypto.HashID(addr.String()),
				Name:       "local",
				Addr:       *addr,
				PublicKey:  make([]byte, crypto.KeySize),
				Seller:     seller,
				Datacenter: datacenter,
				State:      routing.RelayStateEnabled,
			}

			err = fs.AddSeller(ctx, seller)
			assert.NoError(t, err)

			err = fs.AddDatacenter(ctx, datacenter)
			assert.NoError(t, err)

			err = fs.AddRelay(ctx, expected)
			assert.NoError(t, err)

			actual := expected
			actual.Datacenter = routing.Datacenter{}

			err = fs.SetRelay(ctx, actual)
			assert.EqualError(t, err, "datacenter with reference 0 not found")
		})

		t.Run("relay with new ID already exists", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger())
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			oldAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
			assert.NoError(t, err)

			newAddr, err := net.ResolveUDPAddr("udp", "127.0.0.2:40000")
			assert.NoError(t, err)

			seller := routing.Seller{
				ID:                "seller ID",
				Name:              "seller name",
				IngressPriceCents: 10,
				EgressPriceCents:  20,
			}

			datacenter := routing.Datacenter{
				ID:      crypto.HashID("datacenter name"),
				Name:    "datacenter name",
				Enabled: true,
				Location: routing.Location{
					Latitude:  70.5,
					Longitude: 120.5,
				},
			}

			oldRelay := routing.Relay{
				ID:         crypto.HashID(oldAddr.String()),
				Name:       "local",
				Addr:       *oldAddr,
				PublicKey:  make([]byte, crypto.KeySize),
				Seller:     seller,
				Datacenter: datacenter,
				State:      routing.RelayStateEnabled,
			}

			existingRelay := routing.Relay{
				ID:         crypto.HashID(newAddr.String()),
				Name:       "local",
				Addr:       *newAddr,
				PublicKey:  make([]byte, crypto.KeySize),
				Seller:     seller,
				Datacenter: datacenter,
				State:      routing.RelayStateEnabled,
			}

			err = fs.AddSeller(ctx, seller)
			assert.NoError(t, err)

			err = fs.AddDatacenter(ctx, datacenter)
			assert.NoError(t, err)

			err = fs.AddRelay(ctx, oldRelay)
			assert.NoError(t, err)

			err = fs.AddRelay(ctx, existingRelay)
			assert.NoError(t, err)

			newRelay := oldRelay
			newRelay.Addr = *newAddr

			err = fs.SetRelay(ctx, newRelay)
			assert.EqualError(t, err, fmt.Sprintf("relay with reference %x already exists", existingRelay.ID))
		})

		t.Run("success", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger())
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
			assert.NoError(t, err)

			seller := routing.Seller{
				ID:                "seller ID",
				Name:              "seller name",
				IngressPriceCents: 10,
				EgressPriceCents:  20,
			}

			datacenter := routing.Datacenter{
				ID:      crypto.HashID("datacenter name"),
				Name:    "datacenter name",
				Enabled: true,
				Location: routing.Location{
					Latitude:  70.5,
					Longitude: 120.5,
				},
			}

			expected := routing.Relay{
				ID:         crypto.HashID(addr.String()),
				Name:       "local",
				Addr:       *addr,
				PublicKey:  make([]byte, crypto.KeySize),
				Seller:     seller,
				Datacenter: datacenter,
				State:      routing.RelayStateEnabled,
			}

			err = fs.AddSeller(ctx, seller)
			assert.NoError(t, err)

			err = fs.AddDatacenter(ctx, datacenter)
			assert.NoError(t, err)

			err = fs.AddRelay(ctx, expected)
			assert.NoError(t, err)

			actual := expected
			actual.State = routing.RelayStateDisabled

			err = fs.SetRelay(ctx, actual)
			assert.NoError(t, err)

			actual, err = fs.Relay(expected.ID)
			assert.NoError(t, err)

			assert.NotEqual(t, expected, actual)
			actual.State = routing.RelayStateEnabled

			assert.Equal(t, expected, actual)
		})
	})

	t.Run("Datacenter", func(t *testing.T) {
		t.Run("datacenter not found", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger())
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			datacenter, err := fs.Datacenter(0)
			assert.Empty(t, datacenter)
			assert.EqualError(t, err, "datacenter with reference 0 not found")
		})

		t.Run("success", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger())
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			expected := routing.Datacenter{
				ID:      1,
				Name:    "local",
				Enabled: true,
				Location: routing.Location{
					Latitude:  70.5,
					Longitude: 120.5,
				},
			}

			err = fs.AddDatacenter(ctx, expected)
			assert.NoError(t, err)

			actual, err := fs.Datacenter(expected.ID)
			assert.NoError(t, err)

			assert.Equal(t, expected, actual)
		})
	})

	t.Run("Datacenters", func(t *testing.T) {
		fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger())
		assert.NoError(t, err)

		defer func() {
			err := cleanFireStore(ctx, fs.Client)
			assert.NoError(t, err)
		}()

		expected := []routing.Datacenter{
			{
				ID:      1,
				Name:    "local",
				Enabled: true,
				Location: routing.Location{
					Latitude:  70.5,
					Longitude: 120.5,
				},
			},
			{
				ID:      2,
				Name:    "local",
				Enabled: false,
				Location: routing.Location{
					Latitude:  72.5,
					Longitude: 122.5,
				},
			},
		}

		for i := 0; i < len(expected); i++ {
			err = fs.AddDatacenter(ctx, expected[i])
			assert.NoError(t, err)
		}

		actual := fs.Datacenters()
		assert.Equal(t, expected, actual)
	})

	t.Run("AddDatacenter", func(t *testing.T) {
		fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger())
		assert.NoError(t, err)

		defer func() {
			err := cleanFireStore(ctx, fs.Client)
			assert.NoError(t, err)
		}()

		expected := routing.Datacenter{
			ID:      1,
			Name:    "local",
			Enabled: true,
			Location: routing.Location{
				Latitude:  70.5,
				Longitude: 120.5,
			},
		}

		err = fs.AddDatacenter(ctx, expected)
		assert.NoError(t, err)

		actual, err := fs.Datacenter(expected.ID)
		assert.NoError(t, err)

		assert.Equal(t, expected, actual)
	})

	t.Run("RemoveDatacenter", func(t *testing.T) {
		t.Run("datacenter not found", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger())
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			err = fs.RemoveDatacenter(ctx, 0)
			assert.EqualError(t, err, "datacenter with reference 0 not found")
		})

		t.Run("success", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger())
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			datacenter := routing.Datacenter{
				ID:      crypto.HashID("local"),
				Name:    "local",
				Enabled: true,
				Location: routing.Location{
					Latitude:  70.5,
					Longitude: 120.5,
				},
			}

			err = fs.AddDatacenter(ctx, datacenter)
			assert.NoError(t, err)

			err = fs.RemoveDatacenter(ctx, datacenter.ID)
			assert.NoError(t, err)
		})
	})

	t.Run("SetDatacenter", func(t *testing.T) {
		t.Run("datacenter not found", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger())
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			datacenter := routing.Datacenter{
				ID:      1,
				Name:    "local",
				Enabled: true,
				Location: routing.Location{
					Latitude:  70.5,
					Longitude: 120.5,
				},
			}

			err = fs.SetDatacenter(ctx, datacenter)
			assert.EqualError(t, err, "datacenter with reference 1 not found")
		})

		t.Run("success", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger())
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			expected := routing.Datacenter{
				ID:      crypto.HashID("local"),
				Name:    "local",
				Enabled: true,
				Location: routing.Location{
					Latitude:  70.5,
					Longitude: 120.5,
				},
			}

			err = fs.AddDatacenter(ctx, expected)
			assert.NoError(t, err)

			actual := expected
			actual.Enabled = false

			err = fs.SetDatacenter(ctx, actual)
			assert.NoError(t, err)

			actual, err = fs.Datacenter(expected.ID)
			assert.NoError(t, err)

			assert.NotEqual(t, expected, actual)
			actual.Enabled = true
			assert.Equal(t, expected, actual)
		})
	})

	t.Run("Sync", func(t *testing.T) {
		fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger())
		assert.NoError(t, err)

		defer func() {
			err := cleanFireStore(ctx, fs.Client)
			assert.NoError(t, err)
		}()

		expectedBuyer := routing.Buyer{
			ID:                   1,
			Name:                 "local",
			Domain:               "example.com",
			Active:               true,
			Live:                 false,
			PublicKey:            make([]byte, crypto.KeySize),
			RoutingRulesSettings: routing.DefaultRoutingRulesSettings,
		}

		expectedSeller := routing.Seller{
			ID:                "id",
			Name:              "local",
			IngressPriceCents: 10,
			EgressPriceCents:  20,
		}

		expectedDatacenter := routing.Datacenter{
			ID:      crypto.HashID("local"),
			Name:    "local",
			Enabled: true,
			Location: routing.Location{
				Latitude:  70.5,
				Longitude: 120.5,
			},
		}

		addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
		assert.NoError(t, err)

		expectedRelay := routing.Relay{
			ID:          crypto.HashID(addr.String()),
			Name:        "local",
			Addr:        *addr,
			PublicKey:   make([]byte, crypto.KeySize),
			Seller:      expectedSeller,
			Datacenter:  expectedDatacenter,
			MaxSessions: 3000,
			UpdateKey:   make([]byte, 32),
		}

		err = fs.AddBuyer(ctx, expectedBuyer)
		assert.NoError(t, err)

		err = fs.AddSeller(ctx, expectedSeller)
		assert.NoError(t, err)

		// Need to inject a Customer above the Buyer and Seller since Sync requires the Customer exist to get the Domain
		// This saves us from adding Customer(), AddCustomer(), RemoveCustomer() for now to get the portal working requiring
		// the Buyer.ID based on Customer.Domain
		{
			snap, err := fs.Client.Collection("Buyer").Where("name", "==", "local").Documents(ctx).Next()
			assert.NoError(t, err)
			bdoc := snap.Ref

			snap, err = fs.Client.Collection("Seller").Where("name", "==", "local").Documents(ctx).Next()
			assert.NoError(t, err)
			sdoc := snap.Ref

			_, err = fs.Client.Collection("Customer").NewDoc().Create(ctx, map[string]interface{}{
				"buyer":                 bdoc,
				"seller":                sdoc,
				"automaticSigninDomain": "example.com",
				"name":                  "local",
				"active":                true,
			})
			assert.NoError(t, err)
		}

		err = fs.AddDatacenter(ctx, expectedDatacenter)
		assert.NoError(t, err)

		err = fs.AddRelay(ctx, expectedRelay)
		assert.NoError(t, err)

		err = fs.Sync(ctx)
		assert.NoError(t, err)

		actualBuyer, err := fs.Buyer(expectedBuyer.ID)
		assert.NoError(t, err)

		actualSeller, err := fs.Seller(expectedSeller.ID)
		assert.NoError(t, err)

		actualDatacenter, err := fs.Datacenter(expectedDatacenter.ID)
		assert.NoError(t, err)

		actualRelay, err := fs.Relay(expectedRelay.ID)
		assert.NoError(t, err)

		assert.Equal(t, expectedBuyer, actualBuyer)
		assert.Equal(t, expectedSeller, actualSeller)
		assert.Equal(t, expectedDatacenter, actualDatacenter)

		// this is random, no easy way to test so just assert it is present
		expectedRelay.FirestoreID = actualRelay.FirestoreID
		assert.NotEmpty(t, expectedRelay.FirestoreID)

		assert.Equal(t, expectedRelay, actualRelay)
	})
}
