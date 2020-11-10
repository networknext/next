package jsonrpc_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/networknext/backend/storage"
	"github.com/networknext/backend/transport/jsonrpc"
	"github.com/stretchr/testify/assert"
)

func TestFlagList(t *testing.T) {
	t.Parallel()
	var storer = storage.InMemory{}

	logger := log.NewNopLogger()

	svc := jsonrpc.ConfigService{
		Storage: &storer,
		Logger:  logger,
	}

	t.Run("list - empty", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		var reply jsonrpc.FeatureFlagReply
		err := svc.AllFeatureFlags(req, &jsonrpc.FeatureFlagArgs{}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(reply.Flags))
	})
}
