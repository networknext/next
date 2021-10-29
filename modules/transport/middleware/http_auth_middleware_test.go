package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/storage"
	"github.com/networknext/backend/modules/transport/middleware"
	"github.com/stretchr/testify/assert"
)

func TestRoleVerification(t *testing.T) {
	db := storage.InMemory{}
	db.AddCustomer(context.Background(), routing.Customer{Code: "local", Name: "Local"})
	db.AddBuyer(context.Background(), routing.Buyer{ID: 111, CompanyCode: "local"})

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.CustomerKey, "local")
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
		"Admin",
	})
	req = req.WithContext(reqContext)

	t.Run("admin role function", func(t *testing.T) {
		verified, err := middleware.AdminRole(req)
		assert.NoError(t, err)
		assert.True(t, verified)
	})

	reqContext = req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
		"Owner",
	})
	req = req.WithContext(reqContext)

	t.Run("owner role function", func(t *testing.T) {
		verified, err := middleware.OwnerRole(req)
		assert.NoError(t, err)
		assert.True(t, verified)
	})

	reqContext = req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
		"Admin",
		"Owner",
	})
	req = req.WithContext(reqContext)

	t.Run("verify all roles function", func(t *testing.T) {
		verified := middleware.VerifyAllRoles(req, middleware.AdminRole, middleware.OwnerRole)
		assert.True(t, verified)
		verified = middleware.VerifyAllRoles(req, middleware.AdminRole, middleware.AnonymousRole)
		assert.False(t, verified)
	})

	reqContext = req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
		"Admin",
	})
	req = req.WithContext(reqContext)

	t.Run("verify any role function", func(t *testing.T) {
		verified := middleware.VerifyAnyRole(req, middleware.AdminRole, middleware.OwnerRole)
		assert.True(t, verified)
		verified = middleware.VerifyAnyRole(req, middleware.AdminRole, middleware.AnonymousRole)
		assert.True(t, verified)
	})
}
