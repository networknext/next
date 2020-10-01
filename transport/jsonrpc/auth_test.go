package jsonrpc_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
	"github.com/networknext/backend/transport/jsonrpc"
	"github.com/stretchr/testify/assert"
)

// All tests listed below depend on test@networknext.com being a user in auth0
func TestAuthMiddleware(t *testing.T) {
	// JWT obtained from Portal Login Dev SPA (Auth0)
	// Note: 5 year expiration time (expires on 18 May 2025)
	// test@networknext.com => Delete this in auth0 and these tests will break
	jwtSideload := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6Ik5rWXpOekkwTkVVNVFrSTVNRVF4TURRMk5UWTBNakkzTmpOQlJESkVNa1E0TnpGRFF6QkdRdyJ9.eyJuaWNrbmFtZSI6InRlc3QiLCJuYW1lIjoidGVzdEBuZXR3b3JrbmV4dC5jb20iLCJwaWN0dXJlIjoiaHR0cHM6Ly9zLmdyYXZhdGFyLmNvbS9hdmF0YXIvMmRhNWMwMjU5ZTQ3NmI1MDg0MTBlZWY3ZjI5Zjc1NGE_cz00ODAmcj1wZyZkPWh0dHBzJTNBJTJGJTJGY2RuLmF1dGgwLmNvbSUyRmF2YXRhcnMlMkZ0ZS5wbmciLCJ1cGRhdGVkX2F0IjoiMjAyMC0wNi0yM1QxMzozOToyMS44ODFaIiwiZW1haWwiOiJ0ZXN0QG5ldHdvcmtuZXh0LmNvbSIsImVtYWlsX3ZlcmlmaWVkIjp0cnVlLCJpc3MiOiJodHRwczovL25ldHdvcmtuZXh0LmF1dGgwLmNvbS8iLCJzdWIiOiJhdXRoMHw1Yjk2ZjYxY2YxNjQyNzIxYWQ4NGVlYjYiLCJhdWQiOiJvUUpIM1lQSGR2WkpueENQbzFJcnR6NVVLaTV6cnI2biIsImlhdCI6MTU5MjkxOTU2NSwiZXhwIjoxNzUwNzA0MzI1LCJub25jZSI6ImRHZFNUWEpRTnpkdE5GcHNjR0Z1YVg1dlQxVlNhVFZXUjJoK2VHdG1hMnB2TkcweFZuNTFZalJJZmc9PSJ9.BvMe5fWJcheGzKmt3nCIeLjMD-C5426cpjtJiR55i7lmbT0k4h8Z2X6rynZ_aKR-gaCTY7FG5gI-Ty9ZY1zboWcIkxaTi0VKQzdMUTYVMXVEK2cQ1NVbph7_RSJhLfgO5y7PkmuMZXJEFdrI_2PkO4b3tOU-vpUHFUPtTsESV79a81kXn2C5j_KkKzCOPZ4zol1aEU3WliaaJNT38iSz3NX9URshrrdCE39JRClx6wbUgrfCGnVtfens-Sg7atijivaOx8IlUGOxLMEciYwBL2aY5EXaa7tp7c8ZvoEEj7uZH2R35fV7eUzACwShU-JLR9oOsNEhS4XO1AzTMtNHQA"

	noopHandler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	t.Run("skip auth", func(t *testing.T) {
		authMiddleware := jsonrpc.AuthMiddleware("", http.HandlerFunc(noopHandler), false)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		res := httptest.NewRecorder()

		authMiddleware.ServeHTTP(res, req)
		assert.Equal(t, http.StatusOK, res.Code)
	})

	t.Run("check auth claims", func(t *testing.T) {
		authMiddleware := jsonrpc.AuthMiddleware("oQJH3YPHdvZJnxCPo1Irtz5UKi5zrr6n", http.HandlerFunc(noopHandler), false)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Add("Authorization", "Bearer "+jwtSideload)
		res := httptest.NewRecorder()

		authMiddleware.ServeHTTP(res, req)
		assert.Equal(t, http.StatusOK, res.Code)
	})

	t.Run("anonymous auth", func(t *testing.T) {
		authMiddleware := jsonrpc.AuthMiddleware("oQJH3YPHdvZJnxCPo1Irtz5UKi5zrr6n", http.HandlerFunc(noopHandler), false)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		res := httptest.NewRecorder()

		authMiddleware.ServeHTTP(res, req)
		assert.Equal(t, http.StatusOK, res.Code)
	})
}

func TestRoleVerification(t *testing.T) {
	db := storage.InMemory{}
	db.AddCustomer(context.Background(), routing.Customer{Code: "local", Name: "Local"})
	db.AddBuyer(context.Background(), routing.Buyer{ID: 111, CompanyCode: "local"})

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, jsonrpc.Keys.CompanyKey, "local")
	reqContext = context.WithValue(reqContext, jsonrpc.Keys.RolesKey, []string{
		"Admin",
	})
	req = req.WithContext(reqContext)

	t.Run("admin role function", func(t *testing.T) {
		verified, err := jsonrpc.AdminRole(req)
		assert.NoError(t, err)
		assert.True(t, verified)
	})

	reqContext = req.Context()
	reqContext = context.WithValue(reqContext, jsonrpc.Keys.RolesKey, []string{
		"Owner",
	})
	req = req.WithContext(reqContext)

	t.Run("owner role function", func(t *testing.T) {
		verified, err := jsonrpc.OwnerRole(req)
		assert.NoError(t, err)
		assert.True(t, verified)
	})

	reqContext = req.Context()
	reqContext = context.WithValue(reqContext, jsonrpc.Keys.RolesKey, []string{
		"Admin",
		"Owner",
	})
	req = req.WithContext(reqContext)

	t.Run("verify all roles function", func(t *testing.T) {
		verified := jsonrpc.VerifyAllRoles(req, jsonrpc.AdminRole, jsonrpc.OwnerRole)
		assert.True(t, verified)
		verified = jsonrpc.VerifyAllRoles(req, jsonrpc.AdminRole, jsonrpc.AnonymousRole)
		assert.False(t, verified)
	})

	reqContext = req.Context()
	reqContext = context.WithValue(reqContext, jsonrpc.Keys.RolesKey, []string{
		"Admin",
	})
	req = req.WithContext(reqContext)

	t.Run("verify any role function", func(t *testing.T) {
		verified := jsonrpc.VerifyAnyRole(req, jsonrpc.AdminRole, jsonrpc.OwnerRole)
		assert.True(t, verified)
		verified = jsonrpc.VerifyAnyRole(req, jsonrpc.AdminRole, jsonrpc.AnonymousRole)
		assert.True(t, verified)
	})
}
