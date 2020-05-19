package jsonrpc_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
	"github.com/networknext/backend/transport/jsonrpc"
	"github.com/stretchr/testify/assert"
	"gopkg.in/auth0.v4/management"
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
		assert.Equal(t, http.StatusOK, res.Code)
	})

	t.Run("check auth claims", func(t *testing.T) {
		t.Skip("Skipping auth claims testing with time-sensitive JWT")

		// JWT from john@nn, obtained from Localhost Test network response viewer
		jwtSideload := ".eyJuaWNrbmFtZSI6ImpvaG4iLCJuYW1lIjoiam9obkBuZXR3b3JrbmV4dC5jb20iLCJwaWN0dXJlIjoiaHR0cHM6Ly9zLmdyYXZhdGFyLmNvbS9hdmF0YXIvMGIzZTgwMDFjYTJkN2NlM2I2ZmZlMTU2ZTczODRlZTU_cz00ODAmcj1wZyZkPWh0dHBzJTNBJTJGJTJGY2RuLmF1dGgwLmNvbSUyRmF2YXRhcnMlMkZqby5wbmciLCJ1cGRhdGVkX2F0IjoiMjAyMC0wNS0xOFQxOTowMzozMy41MzNaIiwiZW1haWwiOiJqb2huQG5ldHdvcmtuZXh0LmNvbSIsImVtYWlsX3ZlcmlmaWVkIjp0cnVlLCJpc3MiOiJodHRwczovL25ldHdvcmtuZXh0LmF1dGgwLmNvbS8iLCJzdWIiOiJhdXRoMHw1ZWJhYzhiMjA3ZWU4YjFjMTliNGMwZTIiLCJhdWQiOiJxUjYwY2pvNkZjZW9UV1BZTE5aU1FtTmlQMldJQVFyOSIsImlhdCI6MTU4OTgyODYyMywiZXhwIjoxNTg5ODY0NjIzLCJub25jZSI6ImEzNWxTRGx5YkhWTGZsazBhbFprZm1kK00wWm9Za3R0VW1WaGJWcGhTWGR6VlVkME9XdFNNR1Y0YWc9PSJ9.EyTBNZSEZc6CA8fyhbnpDj8JjJ3e-srKOs2hPaSWgrnlVtjIZi-L6BZ6Xb1kluDuRlSSBYVLXDic3dh_g2MgozdMhPoq3VqFtkg1Zc-fVhKFlieXDgb6j01cIcp8fFWiv_rif8ui50-H9QQuQm-GTC0A-YFLOUDdXrE8cjSKLV6meZ-hXQaH8jgtCuNBO-QJrG_lkEfNdhtvpl8Yhf5D1KekM3Qg_KfQqxumST1QzBD1D5yLveLhtIwBB5S1mtMJoOM6k6dpaLffm5sc4NJVtN4XoJvsiRXMqwjfFGWY6Ali1XJ2bXQ3uynYNmuvkv06u4TP6Z7bvkVC5rURvaEOxQ"

		noopHandler := func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}

		authMiddleware := jsonrpc.AuthMiddleware("qR60cjo6FceoTWPYLNZSQmNiP2WIAQr9", http.HandlerFunc(noopHandler))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Add("Authorization", "Bearer "+jwtSideload)
		res := httptest.NewRecorder()

		authMiddleware.ServeHTTP(res, req)
		assert.Equal(t, http.StatusOK, res.Code)
	})

}

func TestAuthClient(t *testing.T) {
	logger := log.NewNopLogger()

	t.Run("create auth0Client", func(t *testing.T) {
		manager, err := management.New(
			"networknext.auth0.com",
			"0Hn8oZfUwy5UPo6bUk0hYCQ2hMJnwQYg",
			"l2namTU5jKVAkuCwV3votIPcP87jcOuJREtscx07aLgo8EykReX69StUVBfJOzx5",
		)
		assert.NoError(t, err)
		auth0 := storage.Auth0{
			Manager: manager,
			Logger:  logger,
		}
		assert.NotEmpty(t, auth0)
	})

	manager, err := management.New(
		"networknext.auth0.com",
		"0Hn8oZfUwy5UPo6bUk0hYCQ2hMJnwQYg",
		"l2namTU5jKVAkuCwV3votIPcP87jcOuJREtscx07aLgo8EykReX69StUVBfJOzx5",
	)
	assert.NoError(t, err)

	auth0Client := storage.Auth0{
		Manager: manager,
		Logger:  logger,
	}
	db := storage.InMemory{}
	db.AddBuyer(context.Background(), routing.Buyer{ID: 111, Domain: "networknext.com"})

	svc := jsonrpc.AuthService{
		Auth0:   auth0Client,
		Storage: &db,
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
		assert.Equal(t, reply.UserAccount.ID, "111")
		assert.Equal(t, reply.UserAccount.CompanyName, "")
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

	t.Run("Remove all auth0 roles", func(t *testing.T) {
		var reply jsonrpc.RolesReply

		roles := []*management.Role{}

		// The user ID here is linked to baumbachandrew@gmail.com => Delete the user and this will not pass
		err := svc.UpdateUserRoles(nil, &jsonrpc.RolesArgs{UserID: "auth0|5eb41e3195054819ac206076", Roles: roles}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, len(reply.Roles), 0)
	})

	t.Run("Update auth0 roles", func(t *testing.T) {
		var reply jsonrpc.RolesReply

		id := "rol_YfFrtom32or4vH89"
		name := "Admin"
		description := "Can manage the Network Next system, including access to configstore."

		roles := []*management.Role{
			{ID: &id, Name: &name, Description: &description},
		}

		// The user ID here is linked to baumbachandrew@gmail.com => Delete the user and this will not pass
		err := svc.UpdateUserRoles(nil, &jsonrpc.RolesArgs{UserID: "auth0|5eb41e3195054819ac206076", Roles: roles}, &reply)
		assert.NoError(t, err)

		assert.NotEqual(t, len(reply.Roles), 0)
		assert.Equal(t, len(reply.Roles), 1)
		assert.Equal(t, reply.Roles[0].ID, &id)
		assert.Equal(t, reply.Roles[0].Name, &name)
		assert.Equal(t, reply.Roles[0].Description, &description)
	})
}
