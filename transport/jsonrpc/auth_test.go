// package jsonrpc_test

// import (
// 	"context"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"

// 	"github.com/dgrijalva/jwt-go"
// 	"github.com/go-kit/kit/log"
// 	"github.com/networknext/backend/routing"
// 	"github.com/networknext/backend/storage"
// 	"github.com/networknext/backend/transport/jsonrpc"
// 	"github.com/stretchr/testify/assert"
// 	"gopkg.in/auth0.v4/management"
// )

// // All tests listed below depend on test@networknext.com being a user in auth0
// func TestAuthMiddleware(t *testing.T) {
// 	// JWT obtained from Portal Login Dev SPA (Auth0)
// 	// Note: 5 year expiration time (expires on 18 May 2025)
// 	// test@networknext.com => Delete this in auth0 and these tests will break
// 	jwtSideload := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6Ik5rWXpOekkwTkVVNVFrSTVNRVF4TURRMk5UWTBNakkzTmpOQlJESkVNa1E0TnpGRFF6QkdRdyJ9.eyJuaWNrbmFtZSI6ImpvaG4iLCJuYW1lIjoiam9obkBuZXR3b3JrbmV4dC5jb20iLCJwaWN0dXJlIjoiaHR0cHM6Ly9zLmdyYXZhdGFyLmNvbS9hdmF0YXIvMGIzZTgwMDFjYTJkN2NlM2I2ZmZlMTU2ZTczODRlZTU_cz00ODAmcj1wZyZkPWh0dHBzJTNBJTJGJTJGY2RuLmF1dGgwLmNvbSUyRmF2YXRhcnMlMkZqby5wbmciLCJ1cGRhdGVkX2F0IjoiMjAyMC0wNS0xOVQxOTo1MDoyMC44NjNaIiwiZW1haWwiOiJqb2huQG5ldHdvcmtuZXh0LmNvbSIsImVtYWlsX3ZlcmlmaWVkIjp0cnVlLCJpc3MiOiJodHRwczovL25ldHdvcmtuZXh0LmF1dGgwLmNvbS8iLCJzdWIiOiJhdXRoMHw1ZWJhYzhiMjA3ZWU4YjFjMTliNGMwZTIiLCJhdWQiOiJvUUpIM1lQSGR2WkpueENQbzFJcnR6NVVLaTV6cnI2biIsImlhdCI6MTU4OTkxNzgyMiwiZXhwIjoxNzQ3NTk3ODIyLCJub25jZSI6IlJuRjFaVzlYYW1aS2VUTkVPRFJMTFhreVVVNVRielJSWVdjdVRGWjFUVlpFZDFVellYNDBXR05sTUE9PSJ9.Va2WRHDUj7XoXzvSkUDfx819RDpewyHMxyv0CIBfsWfVOCB80jRPBvQo7oImRM0FPMYyCl5r4i8-rU5jyg8fZUC3vSABVPALqxX4ViNy3qB4Zgn1RidXoUGKuAUTfi40fS_xDSDBoErRjkxzZuMby_9xNhBw5WwL6sKDGzGL-nayBWHf7LTf0wPwrhZPI4YtHdrJEzYUkwdMCJnMsuSZsgpwvfzvpLgg9NJ4me-VhTQAKJjxXIAsHD_QiI7EEPK1tcd58T11J_xsTktSmDVxuG0-QIs2ioWs0DJSepjcld4tLTlDDZObHIjo_edXd5Wk9zalxfAE7sPWUexFZPQMDA"

// 	noopHandler := func(w http.ResponseWriter, _ *http.Request) {
// 		w.WriteHeader(http.StatusOK)
// 	}

// 	t.Run("skip auth", func(t *testing.T) {
// 		authMiddleware := jsonrpc.AuthMiddleware("", http.HandlerFunc(noopHandler))

// 		req := httptest.NewRequest(http.MethodGet, "/", nil)
// 		res := httptest.NewRecorder()

// 		authMiddleware.ServeHTTP(res, req)
// 		assert.Equal(t, http.StatusOK, res.Code)
// 	})

// 	t.Run("check auth claims", func(t *testing.T) {
// 		authMiddleware := jsonrpc.AuthMiddleware("oQJH3YPHdvZJnxCPo1Irtz5UKi5zrr6n", http.HandlerFunc(noopHandler))

// 		req := httptest.NewRequest(http.MethodGet, "/", nil)
// 		req.Header.Add("Authorization", "Bearer "+jwtSideload)
// 		res := httptest.NewRecorder()

// 		authMiddleware.ServeHTTP(res, req)
// 		assert.Equal(t, http.StatusOK, res.Code)
// 	})

// 	t.Run("anonymous auth", func(t *testing.T) {
// 		authMiddleware := jsonrpc.AuthMiddleware("oQJH3YPHdvZJnxCPo1Irtz5UKi5zrr6n", http.HandlerFunc(noopHandler))

// 		req := httptest.NewRequest(http.MethodGet, "/", nil)
// 		res := httptest.NewRecorder()

// 		authMiddleware.ServeHTTP(res, req)
// 		assert.Equal(t, http.StatusOK, res.Code)
// 	})
// }

// func TestAuthClient(t *testing.T) {
// 	// test@networknext.com => Delete this in auth0 and these tests will break
// 	jwtSideload := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6Ik5rWXpOekkwTkVVNVFrSTVNRVF4TURRMk5UWTBNakkzTmpOQlJESkVNa1E0TnpGRFF6QkdRdyJ9.eyJuaWNrbmFtZSI6ImpvaG4iLCJuYW1lIjoiam9obkBuZXR3b3JrbmV4dC5jb20iLCJwaWN0dXJlIjoiaHR0cHM6Ly9zLmdyYXZhdGFyLmNvbS9hdmF0YXIvMGIzZTgwMDFjYTJkN2NlM2I2ZmZlMTU2ZTczODRlZTU_cz00ODAmcj1wZyZkPWh0dHBzJTNBJTJGJTJGY2RuLmF1dGgwLmNvbSUyRmF2YXRhcnMlMkZqby5wbmciLCJ1cGRhdGVkX2F0IjoiMjAyMC0wNS0xOVQxOTo1MDoyMC44NjNaIiwiZW1haWwiOiJqb2huQG5ldHdvcmtuZXh0LmNvbSIsImVtYWlsX3ZlcmlmaWVkIjp0cnVlLCJpc3MiOiJodHRwczovL25ldHdvcmtuZXh0LmF1dGgwLmNvbS8iLCJzdWIiOiJhdXRoMHw1ZWJhYzhiMjA3ZWU4YjFjMTliNGMwZTIiLCJhdWQiOiJvUUpIM1lQSGR2WkpueENQbzFJcnR6NVVLaTV6cnI2biIsImlhdCI6MTU4OTkxNzgyMiwiZXhwIjoxNzQ3NTk3ODIyLCJub25jZSI6IlJuRjFaVzlYYW1aS2VUTkVPRFJMTFhreVVVNVRielJSWVdjdVRGWjFUVlpFZDFVellYNDBXR05sTUE9PSJ9.Va2WRHDUj7XoXzvSkUDfx819RDpewyHMxyv0CIBfsWfVOCB80jRPBvQo7oImRM0FPMYyCl5r4i8-rU5jyg8fZUC3vSABVPALqxX4ViNy3qB4Zgn1RidXoUGKuAUTfi40fS_xDSDBoErRjkxzZuMby_9xNhBw5WwL6sKDGzGL-nayBWHf7LTf0wPwrhZPI4YtHdrJEzYUkwdMCJnMsuSZsgpwvfzvpLgg9NJ4me-VhTQAKJjxXIAsHD_QiI7EEPK1tcd58T11J_xsTktSmDVxuG0-QIs2ioWs0DJSepjcld4tLTlDDZObHIjo_edXd5Wk9zalxfAE7sPWUexFZPQMDA"
// 	noopHandler := func(w http.ResponseWriter, _ *http.Request) {
// 		w.WriteHeader(http.StatusOK)
// 	}
// 	logger := log.NewNopLogger()

// 	t.Run("create auth0Client", func(t *testing.T) {
// 		manager, err := management.New(
// 			"networknext.auth0.com",
// 			"0Hn8oZfUwy5UPo6bUk0hYCQ2hMJnwQYg",
// 			"l2namTU5jKVAkuCwV3votIPcP87jcOuJREtscx07aLgo8EykReX69StUVBfJOzx5",
// 		)
// 		assert.NoError(t, err)
// 		auth0 := storage.Auth0{
// 			Manager: manager,
// 			Logger:  logger,
// 		}
// 		assert.NotEmpty(t, auth0)
// 	})

// 	manager, err := management.New(
// 		"networknext.auth0.com",
// 		"0Hn8oZfUwy5UPo6bUk0hYCQ2hMJnwQYg",
// 		"l2namTU5jKVAkuCwV3votIPcP87jcOuJREtscx07aLgo8EykReX69StUVBfJOzx5",
// 	)
// 	assert.NoError(t, err)

// 	auth0Client := storage.Auth0{
// 		Manager: manager,
// 		Logger:  logger,
// 	}
// 	db := storage.InMemory{}
// 	db.AddBuyer(context.Background(), routing.Buyer{ID: 111, Domain: "networknext.com"})

// 	svc := jsonrpc.AuthService{
// 		Auth0:   auth0Client,
// 		Storage: &db,
// 		Logger:  logger,
// 	}
// 	authMiddleware := jsonrpc.AuthMiddleware("oQJH3YPHdvZJnxCPo1Irtz5UKi5zrr6n", http.HandlerFunc(noopHandler))

// 	req := httptest.NewRequest(http.MethodGet, "/", nil)
// 	req.Header.Add("Authorization", "Bearer "+jwtSideload)
// 	res := httptest.NewRecorder()

// 	authMiddleware.ServeHTTP(res, req)
// 	assert.Equal(t, http.StatusOK, res.Code)

// 	t.Run("roles from token user", func(t *testing.T) {
// 		user := req.Context().Value("user")
// 		assert.NotEqual(t, user, nil)
// 		claims := user.(*jwt.Token).Claims.(jwt.MapClaims)

// 		requestID, ok := claims["sub"]

// 		assert.True(t, ok)
// 		assert.Equal(t, "auth0|5ebac8b207ee8b1c19b4c0e2", requestID)

// 		roles, err := auth0Client.Manager.User.Roles(requestID.(string))

// 		assert.NoError(t, err)
// 		req = jsonrpc.SetRoles(req, *roles)

// 		userRoles := jsonrpc.RequestRoles(req)

// 		assert.NotEmpty(t, userRoles)

// 		assert.Equal(t, 0, userRoles.Length)
// 		assert.Equal(t, 3, userRoles.Total)
// 		assert.Equal(t, 50, userRoles.Limit)
// 		assert.Equal(t, 0, userRoles.Start)

// 		id := "rol_YfFrtom32or4vH89"
// 		name := "Admin"
// 		description := "Can manage the Network Next system, including access to configstore."

// 		assert.Equal(t, &management.Role{
// 			ID:          &id,
// 			Name:        &name,
// 			Description: &description,
// 		}, userRoles.Roles[0])

// 		id = "rol_8r0281hf2oC4cvrD"
// 		name = "Owner"
// 		description = "Can access and manage everything in an account."

// 		assert.Equal(t, &management.Role{
// 			ID:          &id,
// 			Name:        &name,
// 			Description: &description,
// 		}, userRoles.Roles[1])

// 		id = "rol_ScQpWhLvmTKRlqLU"
// 		name = "Viewer"
// 		description = "Can see current sessions and the map."

// 		assert.Equal(t, &management.Role{
// 			ID:          &id,
// 			Name:        &name,
// 			Description: &description,
// 		}, userRoles.Roles[2])
// 	})

// 	user := req.Context().Value("user")
// 	assert.NotEqual(t, user, nil)
// 	claims := user.(*jwt.Token).Claims.(jwt.MapClaims)

// 	requestID, ok := claims["sub"]

// 	assert.True(t, ok)
// 	assert.Equal(t, "auth0|5ebac8b207ee8b1c19b4c0e2", requestID)

// 	roles, err := auth0Client.Manager.User.Roles(requestID.(string))

// 	assert.NoError(t, err)
// 	req = jsonrpc.SetRoles(req, *roles)

// 	t.Run("fetch all auth0 accounts", func(t *testing.T) {
// 		var reply jsonrpc.AccountsReply

// 		err = svc.AllAccounts(req, &jsonrpc.AccountsArgs{}, &reply)
// 		assert.NoError(t, err)
// 	})

// 	t.Run("fetch user no user id", func(t *testing.T) {
// 		var reply jsonrpc.AccountReply

// 		err := svc.UserAccount(req, &jsonrpc.AccountArgs{}, &reply)
// 		assert.Error(t, err)
// 	})

// 	t.Run("fetch user auth0 account", func(t *testing.T) {
// 		var reply jsonrpc.AccountReply

// 		err := svc.UserAccount(req, &jsonrpc.AccountArgs{UserID: "auth0|5b96f61cf1642721ad84eeb6"}, &reply)
// 		assert.NoError(t, err)

// 		assert.Equal(t, reply.UserAccount.Name, "test@networknext.com")
// 		assert.Equal(t, reply.UserAccount.Email, "test@networknext.com")
// 		assert.Equal(t, reply.UserAccount.UserID, "5b96f61cf1642721ad84eeb6")
// 		assert.Equal(t, reply.UserAccount.ID, "111")
// 		assert.Equal(t, reply.UserAccount.CompanyName, "")
// 	})

// 	t.Run("fetch user roles no user id", func(t *testing.T) {
// 		var reply jsonrpc.RolesReply

// 		err := svc.UserRoles(req, &jsonrpc.RolesArgs{}, &reply)
// 		assert.Error(t, err)
// 	})

// 	t.Run("fetch all auth0 roles", func(t *testing.T) {
// 		var reply jsonrpc.RolesReply

// 		err := svc.AllRoles(req, &jsonrpc.RolesArgs{}, &reply)
// 		assert.NoError(t, err)

// 		assert.NotEqual(t, len(reply.Roles), 0)
// 	})

// 	t.Run("Remove all auth0 roles", func(t *testing.T) {
// 		var reply jsonrpc.RolesReply

// 		roles := []*management.Role{}

// 		err := svc.UpdateUserRoles(req, &jsonrpc.RolesArgs{UserID: "auth0|5b96f61cf1642721ad84eeb6", Roles: roles}, &reply)
// 		assert.NoError(t, err)

// 		assert.Equal(t, len(reply.Roles), 0)
// 	})
// }
