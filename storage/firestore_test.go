package storage_test

import (
	//"context"
	"os"
	"testing"

	//"cloud.google.com/go/firestore"
	//"github.com/go-kit/kit/log"
	//"github.com/networknext/backend/crypto"
	//"github.com/networknext/backend/routing"
	//"github.com/networknext/backend/storage"
	"github.com/stretchr/testify/assert"
)

func TestFirestoreGetBuyer(t *testing.T) {
	err := os.Setenv("FIRESTORE_EMULATOR_HOST", "::1:8037")
	assert.NoError(t, err)

	// ctx := context.Background()
	// client, err := firestore.NewClient(ctx, "*detect-project-id*")
	// assert.NoError(t, err)

	// fs := storage.Firestore{
	// 	Client: client,
	// 	Logger: log.NewNopLogger(),
	// }

	// expected := routing.Buyer{
	// 	ID: 1,
	// 	Name: "local",
	// 	Active: true,
	// 	Live: false,
	// 	PublicKey: make([]byte, crypto.KeySize),
	// 	RoutingRulesSettings: routing.DefaultRoutingRulesSettings,
	// }

	// err = fs.AddBuyer(ctx, expected)
	// assert.NoError(t, err)

	// actual, err := fs.Buyer(expected.ID)
	// assert.NoError(t, err)

	// assert.Equal(t, expected, actual)
}

func TestFirestoreGetBuyers(t *testing.T) {
}

func TestFirestoreAddBuyer(t *testing.T) {
}

func TestFirestoreRemoveBuyer(t *testing.T) {
}

func TestFirestoreSetBuyer(t *testing.T) {
}

func TestFirestoreGetRelay(t *testing.T) {
}

func TestFirestoreGetRelays(t *testing.T) {
}

func TestFirestoreAddRelay(t *testing.T) {
}

func TestFirestoreRemoveRelay(t *testing.T) {
}

func TestFirestoreSetRelay(t *testing.T) {
}

func TestFirestoreGetDatacenter(t *testing.T) {
}

func TestFirestoreGetDatacenters(t *testing.T) {
}

func TestFirestoreAddDatacenter(t *testing.T) {
}

func TestFirestoreRemoveDatacenter(t *testing.T) {
}

func TestFirestoreSetDatacenter(t *testing.T) {
}