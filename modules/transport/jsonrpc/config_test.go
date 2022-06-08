package jsonrpc_test

// todo: maybe convert to functional test

/*
import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/networknext/backend/modules/storage"
	"github.com/networknext/backend/modules/transport/jsonrpc"
	"github.com/stretchr/testify/assert"
)

func TestFlagList(t *testing.T) {
	t.Parallel()
	var storer = storage.InMemory{}

	svc := jsonrpc.ConfigService{
		Storage: &storer,
	}

	t.Run("list - empty", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		var reply jsonrpc.FeatureFlagReply
		err := svc.AllFeatureFlags(req, &jsonrpc.FeatureFlagArgs{}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, 0, len(reply.Flags))
	})
}
*/