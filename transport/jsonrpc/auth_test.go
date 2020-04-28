package jsonrpc_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/networknext/backend/storage"
	"github.com/networknext/backend/transport/jsonrpc"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
	t.Run("skip auth", func(t *testing.T) {
		noopHandler := func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}

		authMiddleware := jsonrpc.AuthMiddleware("", http.HandlerFunc(noopHandler))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		res := httptest.NewRecorder()

		authMiddleware.ServeHTTP(res, req)
		assert.Equal(t, res.Code, http.StatusOK)
	})
}

func TestAuthClient(t *testing.T) {
	logger := log.NewNopLogger()
	t.Run("create auth0Client", func(t *testing.T) {
		_, err := storage.NewAuth0(context.Background(), logger)
		assert.NoError(t, err)
	})

	auth0Client, _ := storage.NewAuth0(context.Background(), logger)
	svc := jsonrpc.AuthService{
		Auth0: *auth0Client,
	}

	t.Run("fetch all auth0 accounts", func(t *testing.T) {
		var reply jsonrpc.AccountsReply

		err := svc.AllAccounts(nil, &jsonrpc.AccountsArgs{}, &reply)
		assert.NoError(t, err)
	})

	t.Run("fetch user no user id", func(t *testing.T) {
		var reply jsonrpc.AccountReply

		err := svc.UserAccount(nil, &jsonrpc.AccountArgs{}, &reply)
		assert.Error(t, err)
	})

	t.Run("fetch user auth0 account", func(t *testing.T) {
		var reply jsonrpc.AccountReply

		err := svc.UserAccount(nil, &jsonrpc.AccountArgs{UserID: "auth0|5e823e827e97a90cf402109e"}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, reply.UserAccount.Name, "andrew@networknext.com")
		assert.Equal(t, reply.UserAccount.Email, "andrew@networknext.com")
		assert.Equal(t, reply.UserAccount.UserID, "5e823e827e97a90cf402109e")
	})

	t.Run("fetch user roles no user id", func(t *testing.T) {
		var reply jsonrpc.RolesReply

		err := svc.UserRoles(nil, &jsonrpc.RolesArgs{}, &reply)
		assert.Error(t, err)
	})

	t.Run("fetch user auth0 roles", func(t *testing.T) {
		var reply jsonrpc.RolesReply

		err := svc.UserRoles(nil, &jsonrpc.RolesArgs{UserID: "auth0|5e823e827e97a90cf402109e"}, &reply)
		assert.NoError(t, err)

		assert.NotEqual(t, len(reply.Roles), 0)
	})

	t.Run("fetch all auth0 roles", func(t *testing.T) {
		var reply jsonrpc.RolesReply

		err := svc.AllRoles(nil, &jsonrpc.RolesArgs{}, &reply)
		assert.NoError(t, err)

		assert.NotEqual(t, len(reply.Roles), 0)
	})
}
