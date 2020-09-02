package storage_test

import (
	"context"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/go-kit/kit/log"
	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/metrics"
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

func TestSequenceNumbers(t *testing.T) {

	checkFirestoreEmulator(t)
	ctx := context.Background()

	t.Run("Sync", func(t *testing.T) {
		fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger(), "test")
		assert.NoError(t, err)

		fs.SyncMetrics = metrics.EmptyFirestoreSyncMetrics

		defer func() {
			err = cleanFireStore(ctx, fs.Client)
			assert.NoError(t, err)
		}()

		err = fs.SetSequenceNumber(ctx, -1)
		assert.NoError(t, err)

		err = fs.IncrementSequenceNumber(ctx)
		assert.NoError(t, err)

		// CheckSequenceNumber() should return true as the remote seq value
		// has been incremented, but the local value is still zero from above
		// (true -> sync from Firestore)
		same, err := fs.CheckSequenceNumber(ctx)
		assert.Equal(t, true, same)
		assert.NoError(t, err)

	})

	t.Run("Do Not Sync", func(t *testing.T) {
		fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger(), "test")
		assert.NoError(t, err)

		fs.SyncMetrics = metrics.EmptyFirestoreSyncMetrics

		defer func() {
			err = cleanFireStore(ctx, fs.Client)
			assert.NoError(t, err)
		}()

		err = fs.SetSequenceNumber(ctx, -1)
		assert.NoError(t, err)

		// CheckSequenceNumber() should return false as the remote seq value
		// has not been incremented and the local value is the initial defautl (-1)
		same, err := fs.CheckSequenceNumber(ctx)
		assert.Equal(t, false, same)
		assert.NoError(t, err)

	})

}

func TestFirestore(t *testing.T) {
	t.Parallel()

	checkFirestoreEmulator(t)
	ctx := context.Background()

	t.Run("NewFirestore", func(t *testing.T) {
		// This test returns an error on Semaphore but does not return an error
		// on a local emulator as of 9/2/20, possible due to different emulator
		// code levels.
		// See: https://github.com/googleapis/google-cloud-go/issues/1751
		// t.Run("firestore client failure", func(t *testing.T) {
		// 	_, err := storage.NewFirestore(ctx, "*detect-project-id*", log.NewNopLogger(), "test")
		// 	assert.Error(t, err)
		// })

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

			actual, err := storage.NewFirestore(ctx, projectID, logger, "test")
			assert.NoError(t, err)

			assert.Equal(t, expected.Logger, actual.Logger)
		})
	})

	t.Run("Buyer", func(t *testing.T) {
		t.Run("buyer not found", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger(), "test")
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
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger(), "test")
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
		fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger(), "test")
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
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger(), "test")
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
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger(), "test")
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

		t.Run("validate only 1 route shader", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger(), "test")
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

			err = fs.AddBuyer(ctx, expected)
			assert.NoError(t, err)

			actual, err := fs.Buyer(expected.ID)
			assert.NoError(t, err)

			assert.Equal(t, expected, actual)

			// Check that only 1 route shader exists in firestore at this point
			rsdocs := fs.Client.Collection("RouteShader").Documents(ctx)
			rsSnapshot, err := rsdocs.GetAll()
			assert.NoError(t, err)
			assert.Len(t, rsSnapshot, 1)
		})
	})

	t.Run("RemoveBuyer", func(t *testing.T) {
		t.Run("buyer not found", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger(), "test")
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			err = fs.RemoveBuyer(ctx, 0)
			assert.EqualError(t, err, "buyer with reference 0 not found")
		})

		t.Run("success - update existing customer", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger(), "test")
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
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger(), "test")
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
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger(), "test")
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
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger(), "test")
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
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger(), "test")
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
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger(), "test")
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
		fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger(), "test")
		assert.NoError(t, err)

		defer func() {
			err := cleanFireStore(ctx, fs.Client)
			assert.NoError(t, err)
		}()

		expected := []routing.Seller{
			{
				ID:                        "id1",
				Name:                      "local",
				IngressPriceNibblinsPerGB: 10,
				EgressPriceNibblinsPerGB:  20,
			},
			{
				ID:                        "id2",
				Name:                      "local",
				IngressPriceNibblinsPerGB: 10,
				EgressPriceNibblinsPerGB:  20,
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
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger(), "test")
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			expected := routing.Seller{
				ID:                        "id",
				Name:                      "local",
				IngressPriceNibblinsPerGB: 10,
				EgressPriceNibblinsPerGB:  20,
			}

			err = fs.AddSeller(ctx, expected)
			assert.NoError(t, err)

			err = fs.AddSeller(ctx, expected)
			assert.EqualError(t, err, "seller with reference id already exists")
		})

		t.Run("success - add new customer", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger(), "test")
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			expected := routing.Seller{
				ID:                        "id",
				Name:                      "local",
				IngressPriceNibblinsPerGB: 10,
				EgressPriceNibblinsPerGB:  20,
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
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger(), "test")
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			expected := routing.Seller{
				ID:                        "id",
				Name:                      "local",
				IngressPriceNibblinsPerGB: 10,
				EgressPriceNibblinsPerGB:  20,
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
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger(), "test")
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			err = fs.RemoveSeller(ctx, "id")
			assert.EqualError(t, err, "seller with reference id not found")
		})

		t.Run("success - removed customer", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger(), "test")
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			seller := routing.Seller{
				ID:                        "id",
				Name:                      "local",
				IngressPriceNibblinsPerGB: 10,
				EgressPriceNibblinsPerGB:  20,
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
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger(), "test")
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			seller := routing.Seller{
				ID:                        "id",
				Name:                      "local",
				IngressPriceNibblinsPerGB: 10,
				EgressPriceNibblinsPerGB:  20,
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
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger(), "test")
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			seller := routing.Seller{
				ID:                        "id",
				Name:                      "local",
				IngressPriceNibblinsPerGB: 10,
				EgressPriceNibblinsPerGB:  20,
			}

			err = fs.SetSeller(ctx, seller)
			assert.EqualError(t, err, "seller with reference id not found")
		})

		t.Run("success", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger(), "test")
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			expected := routing.Seller{
				ID:                        "id",
				Name:                      "local",
				IngressPriceNibblinsPerGB: 10,
				EgressPriceNibblinsPerGB:  20,
			}

			err = fs.AddSeller(ctx, expected)
			assert.NoError(t, err)

			actual := expected
			actual.IngressPriceNibblinsPerGB = 20
			actual.EgressPriceNibblinsPerGB = 10

			err = fs.SetSeller(ctx, actual)
			assert.NoError(t, err)

			actual, err = fs.Seller(expected.ID)
			assert.NoError(t, err)

			assert.NotEqual(t, expected, actual)
			actual.IngressPriceNibblinsPerGB = 10
			actual.EgressPriceNibblinsPerGB = 20
			assert.Equal(t, expected, actual)
		})
	})

	t.Run("SetCustomerLink", func(t *testing.T) {
		t.Run("customer not found", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger(), "test")
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			err = fs.SetCustomerLink(ctx, "not found", 0, "")
			assert.EqualError(t, err, "customer with reference not found not found")
		})

		t.Run("clear buyer", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger(), "test")
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			buyer := routing.Buyer{
				ID:   1,
				Name: "customer name",
			}

			err = fs.AddBuyer(ctx, buyer)
			assert.NoError(t, err)

			err = fs.SetCustomerLink(ctx, "customer name", 0, "")
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

		t.Run("clear seller", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger(), "test")
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			seller := routing.Seller{
				ID:   "sellerID",
				Name: "customer name",
			}

			err = fs.AddSeller(ctx, seller)
			assert.NoError(t, err)

			err = fs.SetCustomerLink(ctx, "customer name", 0, "")
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

		t.Run("change buyer", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger(), "test")
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			oldBuyer := routing.Buyer{
				ID:   1,
				Name: "customer name",
			}

			newBuyer := routing.Buyer{
				ID:   2,
				Name: "different customer name",
			}

			err = fs.AddBuyer(ctx, oldBuyer)
			assert.NoError(t, err)

			err = fs.AddBuyer(ctx, newBuyer)
			assert.NoError(t, err)

			err = fs.SetCustomerLink(ctx, "customer name", newBuyer.ID, "")
			assert.NoError(t, err)

			// Check that the customer was updated successfully
			cdocs := fs.Client.Collection("Customer").Documents(ctx)

			cdoc, err := cdocs.Next()
			assert.NoError(t, err)

			var customerInRemoteStorage customer
			err = cdoc.DataTo(&customerInRemoteStorage)
			assert.NoError(t, err)

			assert.NotNil(t, customerInRemoteStorage.Buyer)

			// Verify that the buyer was set to the new buyer
			var b buyer
			bdoc, err := customerInRemoteStorage.Buyer.Get(ctx)
			assert.NoError(t, err)

			err = bdoc.DataTo(&b)
			assert.NoError(t, err)

			assert.Equal(t, newBuyer.ID, uint64(b.ID))
		})

		t.Run("change seller", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger(), "test")
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			oldSeller := routing.Seller{
				ID:   "oldSellerID",
				Name: "customer name",
			}

			newSeller := routing.Seller{
				ID:   "newSellerID",
				Name: "different customer name",
			}

			err = fs.AddSeller(ctx, oldSeller)
			assert.NoError(t, err)

			err = fs.AddSeller(ctx, newSeller)
			assert.NoError(t, err)

			err = fs.SetCustomerLink(ctx, "customer name", 0, newSeller.ID)
			assert.NoError(t, err)

			// Check that the customer was updated successfully
			cdocs := fs.Client.Collection("Customer").Documents(ctx)

			cdoc, err := cdocs.Next()
			assert.NoError(t, err)

			var customerInRemoteStorage customer
			err = cdoc.DataTo(&customerInRemoteStorage)
			assert.NoError(t, err)

			assert.NotNil(t, customerInRemoteStorage.Seller)

			// Verify that the seller was set to the new seller
			sdoc, err := customerInRemoteStorage.Seller.Get(ctx)
			assert.NoError(t, err)

			assert.Equal(t, newSeller.ID, sdoc.Ref.ID)
		})
	})

	t.Run("BuyerIDFromCustomerName", func(t *testing.T) {
		t.Run("customer not found", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger(), "test")
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			buyerID, err := fs.BuyerIDFromCustomerName(ctx, "not found")
			assert.Empty(t, buyerID)
			assert.EqualError(t, err, "customer with reference not found not found")
		})

		t.Run("success", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger(), "test")
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			buyer := routing.Buyer{
				ID:   1,
				Name: "customer",
			}

			err = fs.AddBuyer(ctx, buyer)
			assert.NoError(t, err)

			buyerID, err := fs.BuyerIDFromCustomerName(ctx, "customer")
			assert.NoError(t, err)
			assert.Equal(t, buyer.ID, buyerID)
		})
	})

	t.Run("SellerIDFromCustomerName", func(t *testing.T) {
		t.Run("customer not found", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger(), "test")
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			buyerID, err := fs.SellerIDFromCustomerName(ctx, "not found")
			assert.Empty(t, buyerID)
			assert.EqualError(t, err, "customer with reference not found not found")
		})

		t.Run("success", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger(), "test")
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			seller := routing.Seller{
				ID:   "sellerID",
				Name: "customer",
			}

			err = fs.AddSeller(ctx, seller)
			assert.NoError(t, err)

			sellerID, err := fs.SellerIDFromCustomerName(ctx, "customer")
			assert.NoError(t, err)
			assert.Equal(t, seller.ID, sellerID)
		})
	})

	t.Run("Relay", func(t *testing.T) {
		t.Run("relay not found", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger(), "test")
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
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger(), "test")
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
			assert.NoError(t, err)

			seller := routing.Seller{
				ID:                        "seller ID",
				Name:                      "seller name",
				IngressPriceNibblinsPerGB: 10,
				EgressPriceNibblinsPerGB:  20,
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
				ID:           1,
				Name:         "local",
				Addr:         *addr,
				PublicKey:    make([]byte, crypto.KeySize),
				Seller:       seller,
				Datacenter:   datacenter,
				MRC:          19700000000000,
				Overage:      26000000000000,
				BWRule:       routing.BWRuleBurst,
				ContractTerm: 12,
				StartDate:    time.Now(),
				EndDate:      time.Now(),
				Type:         routing.BareMetal,
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
		fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger(), "test")
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
			ID:                        "seller ID",
			Name:                      "seller name",
			IngressPriceNibblinsPerGB: 10,
			EgressPriceNibblinsPerGB:  20,
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
				ID:           1,
				Name:         "local",
				Addr:         *addr1,
				PublicKey:    make([]byte, crypto.KeySize),
				Seller:       seller,
				Datacenter:   datacenter,
				MRC:          19700000000000,
				Overage:      26000000000000,
				BWRule:       routing.BWRuleBurst,
				ContractTerm: 12,
				StartDate:    time.Now(),
				EndDate:      time.Now(),
				Type:         routing.BareMetal,
			},
			{
				ID:           2,
				Name:         "local",
				Addr:         *addr2,
				PublicKey:    make([]byte, crypto.KeySize),
				Seller:       seller,
				Datacenter:   datacenter,
				MRC:          19700000000000,
				Overage:      26000000000000,
				BWRule:       routing.BWRuleBurst,
				ContractTerm: 12,
				StartDate:    time.Now(),
				EndDate:      time.Now(),
				Type:         routing.VirtualMachine,
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
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger(), "test")
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
			assert.NoError(t, err)

			expected := routing.Relay{
				ID:           1,
				Name:         "local",
				Addr:         *addr,
				PublicKey:    make([]byte, crypto.KeySize),
				MRC:          19700000000000,
				Overage:      26000000000000,
				BWRule:       routing.BWRuleBurst,
				ContractTerm: 12,
				StartDate:    time.Now(),
				EndDate:      time.Now(),
				Type:         routing.BareMetal,
			}

			err = fs.AddRelay(ctx, expected)
			assert.EqualError(t, err, "seller with reference  not found")
		})

		t.Run("datacenter not found", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger(), "test")
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
			assert.NoError(t, err)

			seller := routing.Seller{
				ID:                        "seller ID",
				Name:                      "seller name",
				IngressPriceNibblinsPerGB: 10,
				EgressPriceNibblinsPerGB:  20,
			}

			expected := routing.Relay{
				ID:           1,
				Name:         "local",
				Addr:         *addr,
				PublicKey:    make([]byte, crypto.KeySize),
				Seller:       seller,
				MRC:          19700000000000,
				Overage:      26000000000000,
				BWRule:       routing.BWRuleBurst,
				ContractTerm: 12,
				StartDate:    time.Now(),
				EndDate:      time.Now(),
				Type:         routing.BareMetal,
			}

			err = fs.AddSeller(ctx, seller)
			assert.NoError(t, err)

			err = fs.AddRelay(ctx, expected)
			assert.EqualError(t, err, "datacenter with reference 0 not found")
		})

		t.Run("success", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger(), "test")
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
			assert.NoError(t, err)

			seller := routing.Seller{
				ID:                        "seller ID",
				Name:                      "seller name",
				IngressPriceNibblinsPerGB: 10,
				EgressPriceNibblinsPerGB:  20,
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
				ID:           1,
				Name:         "local",
				Addr:         *addr,
				PublicKey:    make([]byte, crypto.KeySize),
				Seller:       seller,
				Datacenter:   datacenter,
				MRC:          19700000000000,
				Overage:      26000000000000,
				BWRule:       routing.BWRuleBurst,
				ContractTerm: 12,
				StartDate:    time.Now(),
				EndDate:      time.Now(),
				Type:         routing.BareMetal,
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
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger(), "test")
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			err = fs.RemoveRelay(ctx, 0)
			assert.EqualError(t, err, "relay with reference 0 not found")
		})

		t.Run("success", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger(), "test")
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
			assert.NoError(t, err)

			seller := routing.Seller{
				ID:                        "seller ID",
				Name:                      "seller name",
				IngressPriceNibblinsPerGB: 10,
				EgressPriceNibblinsPerGB:  20,
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
				ID:           crypto.HashID(addr.String()),
				Name:         "local",
				Addr:         *addr,
				PublicKey:    make([]byte, crypto.KeySize),
				Seller:       seller,
				Datacenter:   datacenter,
				MRC:          19700000000000,
				Overage:      26000000000000,
				BWRule:       routing.BWRuleBurst,
				ContractTerm: 12,
				StartDate:    time.Now(),
				EndDate:      time.Now(),
				Type:         routing.BareMetal,
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
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger(), "test")
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

		t.Run("success", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger(), "test")
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
			assert.NoError(t, err)

			seller := routing.Seller{
				ID:                        "seller ID",
				Name:                      "seller name",
				IngressPriceNibblinsPerGB: 10,
				EgressPriceNibblinsPerGB:  20,
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
				ID:           crypto.HashID(addr.String()),
				Name:         "local",
				Addr:         *addr,
				PublicKey:    make([]byte, crypto.KeySize),
				Seller:       seller,
				Datacenter:   datacenter,
				State:        routing.RelayStateEnabled,
				MRC:          19700000000000,
				Overage:      26000000000000,
				BWRule:       routing.BWRuleBurst,
				ContractTerm: 12,
				StartDate:    time.Now(),
				EndDate:      time.Now(),
				Type:         routing.BareMetal,
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

			assert.Equal(t, expected.ID, actual.ID)
			assert.Equal(t, expected.Name, actual.Name)
			assert.Equal(t, expected.Addr, actual.Addr)
			assert.Equal(t, expected.PublicKey, actual.PublicKey)
			assert.Equal(t, expected.Seller, actual.Seller)
			assert.Equal(t, expected.Datacenter, actual.Datacenter)
		})
	})

	t.Run("Datacenter", func(t *testing.T) {
		t.Run("datacenter not found", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger(), "test")
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
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger(), "test")
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
		fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger(), "test")
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
		fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger(), "test")
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
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger(), "test")
			assert.NoError(t, err)

			defer func() {
				err := cleanFireStore(ctx, fs.Client)
				assert.NoError(t, err)
			}()

			err = fs.RemoveDatacenter(ctx, 0)
			assert.EqualError(t, err, "datacenter with reference 0 not found")
		})

		t.Run("success", func(t *testing.T) {
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger(), "test")
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
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger(), "test")
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
			fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger(), "test")
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

	t.Run("Add and Get DatacenterMap", func(t *testing.T) {
		fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger(), "test")
		assert.NoError(t, err)

		defer func() {
			err := cleanFireStore(ctx, fs.Client)
			assert.NoError(t, err)
		}()

		buyer := routing.Buyer{
			ID: 11,
		}

		expected := routing.DatacenterMap{
			BuyerID:    11,
			Datacenter: 1,
			Alias:      "local",
		}

		id := crypto.HashID(expected.Alias + fmt.Sprintf("%x", expected.BuyerID) + fmt.Sprintf("%x", expected.Datacenter))

		datacenter := routing.Datacenter{
			ID:      1,
			Name:    "local",
			Enabled: true,
			Location: routing.Location{
				Latitude:  70.5,
				Longitude: 120.5,
			},
		}

		err = fs.AddBuyer(ctx, buyer)
		assert.NoError(t, err)

		err = fs.AddDatacenter(ctx, datacenter)
		assert.NoError(t, err)

		err = fs.AddDatacenterMap(ctx, expected)
		assert.NoError(t, err)

		actual := fs.GetDatacenterMapsForBuyer(buyer.ID)
		assert.Equal(t, expected, actual[id])
	})

	t.Run("Add two and get the list", func(t *testing.T) {
		fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger(), "test")
		assert.NoError(t, err)

		defer func() {
			err := cleanFireStore(ctx, fs.Client)
			assert.NoError(t, err)
		}()

		buyer1 := routing.Buyer{
			ID: 11,
		}

		buyer2 := routing.Buyer{
			ID: 22,
		}

		expected1 := routing.DatacenterMap{
			BuyerID:    11,
			Datacenter: 1,
			Alias:      "local.alias",
		}

		expected2 := routing.DatacenterMap{
			BuyerID:    22,
			Datacenter: 1,
			Alias:      "other.local.alias",
		}

		datacenter := routing.Datacenter{
			ID:      1,
			Name:    "local",
			Enabled: true,
			Location: routing.Location{
				Latitude:  70.5,
				Longitude: 120.5,
			},
		}

		id1 := crypto.HashID(expected1.Alias + fmt.Sprintf("%x", expected1.BuyerID) + fmt.Sprintf("%x", expected1.Datacenter))
		id2 := crypto.HashID(expected2.Alias + fmt.Sprintf("%x", expected2.BuyerID) + fmt.Sprintf("%x", expected2.Datacenter))

		err = fs.AddBuyer(ctx, buyer1)
		assert.NoError(t, err)

		err = fs.AddBuyer(ctx, buyer2)
		assert.NoError(t, err)

		err = fs.AddDatacenter(ctx, datacenter)
		assert.NoError(t, err)

		err = fs.AddDatacenterMap(ctx, expected1)
		assert.NoError(t, err)

		err = fs.AddDatacenterMap(ctx, expected2)
		assert.NoError(t, err)

		actual := fs.ListDatacenterMaps(0)
		assert.Equal(t, expected1, actual[id1])
		assert.Equal(t, expected2, actual[id2])
	})

	t.Run("Add and Remove DatacenterMap", func(t *testing.T) {
		fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger(), "test")
		assert.NoError(t, err)

		defer func() {
			err := cleanFireStore(ctx, fs.Client)
			assert.NoError(t, err)
		}()

		buyer := routing.Buyer{
			ID: 11,
		}

		dcMap := routing.DatacenterMap{
			BuyerID:    11,
			Datacenter: 1,
			Alias:      "local",
		}

		datacenter := routing.Datacenter{
			ID:      1,
			Name:    "local",
			Enabled: true,
			Location: routing.Location{
				Latitude:  70.5,
				Longitude: 120.5,
			},
		}

		err = fs.AddBuyer(ctx, buyer)
		assert.NoError(t, err)

		err = fs.AddDatacenter(ctx, datacenter)
		assert.NoError(t, err)

		err = fs.AddDatacenterMap(ctx, dcMap)
		assert.NoError(t, err)

		err = fs.RemoveDatacenterMap(ctx, dcMap)
		assert.NoError(t, err)

		var dcMapsEmpty = make(map[uint64]routing.DatacenterMap)
		dcMapsEmpty = fs.GetDatacenterMapsForBuyer(buyer.ID)
		assert.Equal(t, 0, len(dcMapsEmpty))

	})
	t.Run("Sync", func(t *testing.T) {
		fs, err := storage.NewFirestore(ctx, "default", log.NewNopLogger(), "test")
		assert.NoError(t, err)

		fs.SyncMetrics = metrics.EmptyFirestoreSyncMetrics

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
			ID:                        "id",
			Name:                      "local",
			IngressPriceNibblinsPerGB: 10,
			EgressPriceNibblinsPerGB:  20,
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

		startDate, _ := time.Parse("January 2, 2006", "January 2, 2006")
		endDate, _ := time.Parse("January 2, 2006", "January 2, 2007")

		expectedRelay := routing.Relay{
			ID:           crypto.HashID(addr.String()),
			Name:         "local",
			Addr:         *addr,
			PublicKey:    make([]byte, crypto.KeySize),
			Seller:       expectedSeller,
			Datacenter:   expectedDatacenter,
			MaxSessions:  3000,
			UpdateKey:    make([]byte, 32),
			MRC:          19700000000000,
			Overage:      26000000000000,
			BWRule:       routing.BWRuleBurst,
			ContractTerm: 12,
			StartDate:    startDate,
			EndDate:      endDate,
			Type:         routing.BareMetal,
		}

		err = fs.SetSequenceNumber(ctx, -1)
		assert.NoError(t, err)

		err = fs.AddBuyer(ctx, expectedBuyer)
		assert.NoError(t, err)

		err = fs.AddSeller(ctx, expectedSeller)
		assert.NoError(t, err)

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
