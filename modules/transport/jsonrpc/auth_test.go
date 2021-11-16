package jsonrpc_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-kit/kit/log"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/storage"
	"github.com/networknext/backend/modules/transport/jsonrpc"
	"github.com/networknext/backend/modules/transport/middleware"
	"github.com/stretchr/testify/assert"
	"gopkg.in/auth0.v4/management"
)

func TestAllAccounts(t *testing.T) {
	t.Parallel()
	var userManager = storage.NewLocalUserManager()
	var jobManager = storage.LocalJobManager{}
	var storer = storage.InMemory{}

	logger := log.NewNopLogger()

	svc := jsonrpc.AuthService{
		UserManager: userManager,
		JobManager:  &jobManager,
		Storage:     &storer,
		Logger:      logger,
	}

	IDs := []string{
		"123",
		"456",
		"789",
	}

	emails := []string{
		"test@test.com",
		"test@test1.com",
		"test@test2.com",
	}

	names := []string{
		"Frank",
		"George",
		"Lenny",
	}

	roleIDs := []string{
		"rol_ScQpWhLvmTKRlqLU",
		"rol_8r0281hf2oC4cvrD",
		"rol_YfFrtom32or4vH89",
	}
	roleNames := []string{
		"Explorer",
		"Owner",
		"Admin",
	}
	roleDescriptions := []string{
		"Can access the explore tab",
		"Can access and manage everything in an account.",
		"Can manage the Network Next system, including access to configstore.",
	}

	currentTime := time.Now()

	userManager.Create(&management.User{
		ID:    &IDs[0],
		Email: &emails[0],
		AppMetadata: map[string]interface{}{
			"company_code": "test",
		},
		Identities: []*management.UserIdentity{
			{
				UserID: &IDs[0],
			},
		},
		Name:       &emails[0],
		GivenName:  &names[0],
		FamilyName: &names[0],
		CreatedAt:  &currentTime,
	})

	userManager.Create(&management.User{
		ID:    &IDs[1],
		Email: &emails[1],
		AppMetadata: map[string]interface{}{
			"company_code": "test-test",
		},
		Identities: []*management.UserIdentity{
			{
				UserID: &IDs[1],
			},
		},
		Name:       &emails[1],
		GivenName:  &names[1],
		FamilyName: &names[1],
		CreatedAt:  &currentTime,
	})

	userManager.AssignRoles(IDs[1], []*management.Role{
		{
			ID:          &roleIDs[0],
			Name:        &roleNames[0],
			Description: &roleDescriptions[0],
		},
	}...)

	userManager.Create(&management.User{
		ID:    &IDs[2],
		Email: &emails[2],
		Identities: []*management.UserIdentity{
			{
				UserID: &IDs[2],
			},
		},
		Name:       &emails[2],
		GivenName:  &names[2],
		FamilyName: &names[2],
		CreatedAt:  &currentTime,
	})

	storer.AddCustomer(context.Background(), routing.Customer{Code: "test", Name: "Test"})
	storer.AddBuyer(context.Background(), routing.Buyer{CompanyCode: "test", ID: 123})
	storer.AddCustomer(context.Background(), routing.Customer{Code: "test-test", Name: "Test Test"})
	storer.AddBuyer(context.Background(), routing.Buyer{CompanyCode: "test-test", ID: 456})
	storer.AddCustomer(context.Background(), routing.Customer{Code: "test-test-test", Name: "Test Test Test"})
	storer.AddBuyer(context.Background(), routing.Buyer{CompanyCode: "test-test-test", ID: 789})

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("all - failure", func(t *testing.T) {
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, middleware.Keys.CustomerKey, "test")
		req = req.WithContext(reqContext)
		var reply jsonrpc.AccountsReply
		err := svc.AllAccounts(req, &jsonrpc.AccountsArgs{}, &reply)
		assert.Error(t, err)
	})

	userManager.AssignRoles(IDs[0], []*management.Role{
		{
			ID:          &roleIDs[0],
			Name:        &roleNames[0],
			Description: &roleDescriptions[0],
		},
	}...)

	t.Run("all - success - no users in company", func(t *testing.T) {
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, middleware.Keys.UserKey, &jwt.Token{
			Claims: jwt.MapClaims{
				"email": "test@test.com",
			},
		})
		reqContext = context.WithValue(reqContext, middleware.Keys.CustomerKey, "test")
		reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
			"Admin",
		})
		req = req.WithContext(reqContext)
		var reply jsonrpc.AccountsReply
		err := svc.AllAccounts(req, &jsonrpc.AccountsArgs{}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(reply.UserAccounts))
		assert.Equal(t, names[0], reply.UserAccounts[0].FirstName)
		assert.Equal(t, emails[0], reply.UserAccounts[0].Email)
		assert.Equal(t, IDs[0], reply.UserAccounts[0].UserID)
		assert.Equal(t, fmt.Sprintf("%016x", 123), reply.UserAccounts[0].BuyerID)
	})

	t.Run("all - success", func(t *testing.T) {
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, middleware.Keys.UserKey, &jwt.Token{
			Claims: jwt.MapClaims{
				"email": "test@test.com",
			},
		})
		reqContext = context.WithValue(reqContext, middleware.Keys.CustomerKey, "test-test")
		reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
			"Admin",
		})
		req = req.WithContext(reqContext)
		var reply jsonrpc.AccountsReply
		err := svc.AllAccounts(req, &jsonrpc.AccountsArgs{}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(reply.UserAccounts))
		assert.Equal(t, "George", reply.UserAccounts[0].FirstName)
		assert.Equal(t, "test@test1.com", reply.UserAccounts[0].Email)
		assert.Equal(t, "test-test", reply.UserAccounts[0].CompanyCode)
		assert.Equal(t, "Test Test", reply.UserAccounts[0].CompanyName)
		assert.Equal(t, IDs[1], reply.UserAccounts[0].UserID)
		assert.Equal(t, fmt.Sprintf("%016x", 456), reply.UserAccounts[0].BuyerID)
		assert.Equal(t, roleNames[0], *reply.UserAccounts[0].Roles[0].Name)
		assert.Equal(t, roleIDs[0], *reply.UserAccounts[0].Roles[0].ID)
		assert.Equal(t, roleDescriptions[0], *reply.UserAccounts[0].Roles[0].Description)
	})
}

func TestUserAccount(t *testing.T) {
	t.Parallel()
	var userManager = storage.NewLocalUserManager()
	var jobManager = storage.LocalJobManager{}
	var storer = storage.InMemory{}

	logger := log.NewNopLogger()

	svc := jsonrpc.AuthService{
		UserManager: userManager,
		JobManager:  &jobManager,
		Storage:     &storer,
		Logger:      logger,
	}

	IDs := []string{
		"123",
		"456",
		"789",
	}

	emails := []string{
		"test@test.com",
		"test@test1.com",
		"test@test2.com",
	}

	names := []string{
		"Frank",
		"George",
		"Lenny",
	}

	roleIDs := []string{
		"rol_ScQpWhLvmTKRlqLU",
		"rol_8r0281hf2oC4cvrD",
		"rol_YfFrtom32or4vH89",
	}
	roleNames := []string{
		"Explorer",
		"Owner",
		"Admin",
	}
	roleDescriptions := []string{
		"Can access the explore tab",
		"Can access and manage everything in an account.",
		"Can manage the Network Next system, including access to configstore.",
	}

	currentTime := time.Now()

	userManager.Create(&management.User{
		ID:    &IDs[0],
		Email: &emails[0],
		AppMetadata: map[string]interface{}{
			"company_code": "test",
		},
		Identities: []*management.UserIdentity{
			{
				UserID: &IDs[0],
			},
		},
		Name:       &emails[0],
		GivenName:  &names[0],
		FamilyName: &names[0],
		CreatedAt:  &currentTime,
	})

	userManager.AssignRoles(IDs[0], []*management.Role{
		{
			ID:          &roleIDs[2],
			Name:        &roleNames[2],
			Description: &roleDescriptions[2],
		},
	}...)

	userManager.Create(&management.User{
		ID:    &IDs[1],
		Email: &emails[1],
		AppMetadata: map[string]interface{}{
			"company_code": "test-test",
		},
		Identities: []*management.UserIdentity{
			{
				UserID: &IDs[1],
			},
		},
		Name:       &emails[1],
		GivenName:  &names[1],
		FamilyName: &names[1],
		CreatedAt:  &currentTime,
	})

	userManager.AssignRoles(IDs[1], []*management.Role{
		{
			ID:          &roleIDs[0],
			Name:        &roleNames[0],
			Description: &roleDescriptions[0],
		},
	}...)

	userManager.Create(&management.User{
		ID:    &IDs[2],
		Email: &emails[2],
		Identities: []*management.UserIdentity{
			{
				UserID: &IDs[2],
			},
		},
		Name:       &emails[2],
		GivenName:  &names[2],
		FamilyName: &names[2],
		CreatedAt:  &currentTime,
	})

	userManager.AssignRoles(IDs[2], []*management.Role{
		{
			ID:          &roleIDs[0],
			Name:        &roleNames[0],
			Description: &roleDescriptions[0],
		},
	}...)

	storer.AddCustomer(context.Background(), routing.Customer{Code: "test", Name: "Test", AutomaticSignInDomains: "google.com,test.com"})
	storer.AddBuyer(context.Background(), routing.Buyer{CompanyCode: "test", ID: 123})
	storer.AddCustomer(context.Background(), routing.Customer{Code: "test-test", Name: "Test Test"})
	storer.AddBuyer(context.Background(), routing.Buyer{CompanyCode: "test-test", ID: 456})

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("failure - no id", func(t *testing.T) {
		var reply jsonrpc.AccountReply
		err := svc.UserAccount(req, &jsonrpc.AccountArgs{}, &reply)
		assert.Error(t, err)
	})

	t.Run("failure - no token", func(t *testing.T) {
		var reply jsonrpc.AccountReply
		err := svc.UserAccount(req, &jsonrpc.AccountArgs{UserID: "123"}, &reply)
		assert.Error(t, err)
	})

	t.Run("success - request user - !admin", func(t *testing.T) {
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, middleware.Keys.UserKey, &jwt.Token{
			Claims: jwt.MapClaims{
				"email": "test@test.com",
				"sub":   "123",
			},
		})
		reqContext = context.WithValue(reqContext, middleware.Keys.CustomerKey, "test")

		req = req.WithContext(reqContext)
		var reply jsonrpc.AccountReply
		err := svc.UserAccount(req, &jsonrpc.AccountArgs{UserID: "123"}, &reply)
		assert.NoError(t, err)

		assert.Nil(t, reply.Domains)
		assert.Equal(t, "123", reply.UserAccount.UserID)
		assert.Equal(t, fmt.Sprintf("%016x", 123), reply.UserAccount.BuyerID)
		assert.Equal(t, "test", reply.UserAccount.CompanyCode)
		assert.Equal(t, "Test", reply.UserAccount.CompanyName)
		assert.Equal(t, "Frank", reply.UserAccount.FirstName)
		assert.Equal(t, "test@test.com", reply.UserAccount.Email)
	})

	t.Run("success - request user - admin", func(t *testing.T) {
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, middleware.Keys.UserKey, &jwt.Token{
			Claims: jwt.MapClaims{
				"email": "test@test.com",
				"sub":   "123",
			},
		})
		reqContext = context.WithValue(reqContext, middleware.Keys.CustomerKey, "test")
		reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
			"Admin",
		})

		req = req.WithContext(reqContext)
		var reply jsonrpc.AccountReply
		err := svc.UserAccount(req, &jsonrpc.AccountArgs{UserID: "123"}, &reply)
		assert.NoError(t, err)

		assert.NotNil(t, reply.Domains)
		assert.Equal(t, 2, len(reply.Domains))
		assert.Equal(t, "google.com", reply.Domains[0])
		assert.Equal(t, "test.com", reply.Domains[1])
	})

	t.Run("success - random user - !admin", func(t *testing.T) {
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, middleware.Keys.UserKey, &jwt.Token{
			Claims: jwt.MapClaims{
				"email": "test@test.com",
				"sub":   "123",
			},
		})
		reqContext = context.WithValue(reqContext, middleware.Keys.CustomerKey, "test")
		reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{})
		req = req.WithContext(reqContext)
		var reply jsonrpc.AccountReply
		err := svc.UserAccount(req, &jsonrpc.AccountArgs{UserID: "456"}, &reply)
		assert.Error(t, err)
	})

	t.Run("success - random user - admin", func(t *testing.T) {
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, middleware.Keys.UserKey, &jwt.Token{
			Claims: jwt.MapClaims{
				"email": "test@test.com",
				"sub":   "123",
			},
		})
		reqContext = context.WithValue(reqContext, middleware.Keys.CustomerKey, "test")
		reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
			"Admin",
		})
		req = req.WithContext(reqContext)
		var reply jsonrpc.AccountReply
		err := svc.UserAccount(req, &jsonrpc.AccountArgs{UserID: "456"}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, 0, len(reply.Domains))
		assert.Equal(t, "456", reply.UserAccount.UserID)
		assert.Equal(t, fmt.Sprintf("%016x", 456), reply.UserAccount.BuyerID)
		assert.Equal(t, "test-test", reply.UserAccount.CompanyCode)
		assert.Equal(t, "Test Test", reply.UserAccount.CompanyName)
		assert.Equal(t, "George", reply.UserAccount.FirstName)
		assert.Equal(t, "test@test1.com", reply.UserAccount.Email)
	})

	t.Run("success - random user - admin - no company", func(t *testing.T) {
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, middleware.Keys.UserKey, &jwt.Token{
			Claims: jwt.MapClaims{
				"email": "test@test.com",
				"sub":   "123",
			},
		})
		reqContext = context.WithValue(reqContext, middleware.Keys.CustomerKey, "test")
		reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
			"Admin",
		})
		req = req.WithContext(reqContext)
		var reply jsonrpc.AccountReply
		err := svc.UserAccount(req, &jsonrpc.AccountArgs{UserID: "789"}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, 0, len(reply.Domains))
		assert.Equal(t, "789", reply.UserAccount.UserID)
		assert.Equal(t, "", reply.UserAccount.BuyerID)
		assert.Equal(t, "", reply.UserAccount.CompanyCode)
		assert.Equal(t, "", reply.UserAccount.CompanyName)
		assert.Equal(t, "Lenny", reply.UserAccount.FirstName)
		assert.Equal(t, "test@test2.com", reply.UserAccount.Email)
	})
}

func TestDeleteAccount(t *testing.T) {
	t.Parallel()
	var userManager = storage.NewLocalUserManager()
	var jobManager = storage.LocalJobManager{}
	var storer = storage.InMemory{}

	logger := log.NewNopLogger()

	svc := jsonrpc.AuthService{
		UserManager: userManager,
		JobManager:  &jobManager,
		Storage:     &storer,
		Logger:      logger,
	}

	IDs := []string{
		"123",
		"456",
		"789",
	}

	emails := []string{
		"test@test.com",
		"test@test1.com",
		"test@test2.com",
	}

	names := []string{
		"Frank",
		"George",
		"Lenny",
	}

	roleIDs := []string{
		"rol_ScQpWhLvmTKRlqLU",
		"rol_8r0281hf2oC4cvrD",
		"rol_YfFrtom32or4vH89",
	}
	roleNames := []string{
		"Explorer",
		"Owner",
		"Admin",
	}
	roleDescriptions := []string{
		"Can access the explore tab",
		"Can access and manage everything in an account.",
		"Can manage the Network Next system, including access to configstore.",
	}

	userManager.Create(&management.User{
		ID:    &IDs[0],
		Email: &emails[0],
		AppMetadata: map[string]interface{}{
			"company_code": "test",
		},
		Identities: []*management.UserIdentity{
			{
				UserID: &IDs[0],
			},
		},
		Name: &names[0],
	})

	userManager.AssignRoles(IDs[0], []*management.Role{
		{
			ID:          &roleIDs[2],
			Name:        &roleNames[2],
			Description: &roleDescriptions[2],
		},
	}...)

	userManager.Create(&management.User{
		ID:    &IDs[1],
		Email: &emails[1],
		AppMetadata: map[string]interface{}{
			"company_code": "test",
		},
		Identities: []*management.UserIdentity{
			{
				UserID: &IDs[1],
			},
		},
		Name: &names[1],
	})

	userManager.AssignRoles(IDs[1], []*management.Role{
		{
			ID:          &roleIDs[0],
			Name:        &roleNames[0],
			Description: &roleDescriptions[0],
		},
	}...)

	userManager.Create(&management.User{
		ID:    &IDs[2],
		Email: &emails[2],
		Identities: []*management.UserIdentity{
			{
				UserID: &IDs[2],
			},
		},
		AppMetadata: map[string]interface{}{
			"company_code": "test-test",
		},
		Name: &names[2],
	})

	userManager.AssignRoles(IDs[2], []*management.Role{
		{
			ID:          &roleIDs[0],
			Name:        &roleNames[0],
			Description: &roleDescriptions[0],
		},
	}...)

	storer.AddCustomer(context.Background(), routing.Customer{Code: "test", Name: "Test", AutomaticSignInDomains: "google.com,test.com"})
	storer.AddBuyer(context.Background(), routing.Buyer{CompanyCode: "test", ID: 123})
	storer.AddCustomer(context.Background(), routing.Customer{Code: "test-test", Name: "Test Test"})
	storer.AddBuyer(context.Background(), routing.Buyer{CompanyCode: "test-test", ID: 456})

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("failure - insufficient privileges", func(t *testing.T) {
		var reply jsonrpc.AccountReply
		err := svc.DeleteUserAccount(req, &jsonrpc.AccountArgs{}, &reply)
		assert.Error(t, err)
	})

	t.Run("failure - no id", func(t *testing.T) {
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
			"Owner",
		})
		req = req.WithContext(reqContext)
		var reply jsonrpc.AccountReply
		err := svc.DeleteUserAccount(req, &jsonrpc.AccountArgs{}, &reply)
		assert.Error(t, err)
	})

	t.Run("success - same company", func(t *testing.T) {
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, middleware.Keys.CustomerKey, "test")
		req = req.WithContext(reqContext)
		var reply jsonrpc.AccountReply
		err := svc.DeleteUserAccount(req, &jsonrpc.AccountArgs{UserID: "456"}, &reply)
		assert.NoError(t, err)
		users, err := userManager.List()
		found := false
		for _, u := range users.Users {
			if *u.ID == "456" && u.AppMetadata["company_code"] != "" {
				found = true
			}
		}
		assert.False(t, found)
	})

	t.Run("failure - !same company - !admin", func(t *testing.T) {
		var reply jsonrpc.AccountReply
		err := svc.DeleteUserAccount(req, &jsonrpc.AccountArgs{UserID: "789"}, &reply)
		assert.Error(t, err)
	})

	t.Run("failure - !same company - admin", func(t *testing.T) {
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
			"Admin",
		})
		req = req.WithContext(reqContext)
		var reply jsonrpc.AccountReply
		err := svc.DeleteUserAccount(req, &jsonrpc.AccountArgs{UserID: "789"}, &reply)
		assert.NoError(t, err)
		users, err := userManager.List()
		found := false
		for _, u := range users.Users {
			if *u.ID == "789" && u.AppMetadata["company_code"] != "" {
				found = true
			}
		}
		assert.False(t, found)
	})
}

func TestAddUserAccount(t *testing.T) {
	t.Parallel()
	var jobManager = storage.LocalJobManager{}
	var userManager = storage.NewLocalUserManager()
	var roleManager = storage.NewLocalRoleManager()
	var storer = storage.InMemory{}

	logger := log.NewNopLogger()

	svc := jsonrpc.AuthService{
		JobManager:  &jobManager,
		Logger:      logger,
		RoleCache:   make(map[string]*management.Role),
		RoleManager: roleManager,
		Storage:     &storer,
		UserManager: userManager,
	}

	roleIDs := []string{
		"rol_ScQpWhLvmTKRlqLU",
		"rol_8r0281hf2oC4cvrD",
		"rol_YfFrtom32or4vH89",
	}
	roleNames := []string{
		"Explorer",
		"Owner",
		"Admin",
	}
	roleDescriptions := []string{
		"Can access the explore tab",
		"Can access and manage everything in an account.",
		"Can manage the Network Next system, including access to configstore.",
	}

	svc.RoleCache["Admin"] = &management.Role{
		Description: &roleDescriptions[2],
		ID:          &roleIDs[2],
		Name:        &roleNames[2],
	}

	svc.RoleCache["Owner"] = &management.Role{
		Description: &roleDescriptions[1],
		ID:          &roleIDs[1],
		Name:        &roleNames[1],
	}

	svc.RoleCache["Explorer"] = &management.Role{
		Description: &roleDescriptions[0],
		ID:          &roleIDs[0],
		Name:        &roleNames[0],
	}

	IDs := []string{
		"123",
		"456",
		"789",
	}

	emails := []string{
		"test@test.com",
		"test@test1.com",
		"test@test2.com",
	}

	names := []string{
		"Frank",
		"George",
		"Lenny",
	}

	currentTime := time.Now()

	userManager.Create(&management.User{
		ID:    &IDs[1],
		Email: &emails[1],
		AppMetadata: map[string]interface{}{
			"company_code": "test",
		},
		Identities: []*management.UserIdentity{
			{
				UserID: &IDs[1],
			},
		},
		Name:      &names[1],
		CreatedAt: &currentTime,
	})

	userManager.AssignRoles(IDs[1], []*management.Role{
		{
			ID:          &roleIDs[0],
			Name:        &roleNames[0],
			Description: &roleDescriptions[0],
		},
	}...)

	storer.AddCustomer(context.Background(), routing.Customer{Code: "test", Name: "Test", AutomaticSignInDomains: "google.com,test.com"})

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("failure - insufficient privileges", func(t *testing.T) {
		var reply jsonrpc.AccountsReply
		err := svc.AddUserAccount(req, &jsonrpc.AccountsArgs{}, &reply)
		assert.Error(t, err)
	})

	t.Run("failure - no roles", func(t *testing.T) {
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
			"Owner",
		})
		req = req.WithContext(reqContext)
		var reply jsonrpc.AccountsReply
		err := svc.AddUserAccount(req, &jsonrpc.AccountsArgs{}, &reply)
		assert.Error(t, err)
	})

	t.Run("failure - admin role - !admin", func(t *testing.T) {
		var reply jsonrpc.AccountsReply
		err := svc.AddUserAccount(req, &jsonrpc.AccountsArgs{Roles: []*management.Role{
			{
				ID:          &roleIDs[2],
				Name:        &roleNames[2],
				Description: &roleDescriptions[2],
			},
		}}, &reply)
		assert.Error(t, err)
	})

	t.Run("failure - admin role - !admin", func(t *testing.T) {
		var reply jsonrpc.AccountsReply
		err := svc.AddUserAccount(req, &jsonrpc.AccountsArgs{Roles: []*management.Role{
			{
				ID:          &roleIDs[2],
				Name:        &roleNames[2],
				Description: &roleDescriptions[2],
			},
		}}, &reply)
		assert.Error(t, err)
	})

	t.Run("failure - no request company code", func(t *testing.T) {
		var reply jsonrpc.AccountsReply
		err := svc.AddUserAccount(req, &jsonrpc.AccountsArgs{Roles: []*management.Role{
			{
				ID:          &roleIDs[0],
				Name:        &roleNames[0],
				Description: &roleDescriptions[0],
			},
		}}, &reply)
		assert.Error(t, err)
	})

	t.Run("failure - no buyer", func(t *testing.T) {
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, middleware.Keys.CustomerKey, "test")
		req = req.WithContext(reqContext)
		var reply jsonrpc.AccountsReply
		err := svc.AddUserAccount(req, &jsonrpc.AccountsArgs{Roles: []*management.Role{
			{
				ID:          &roleIDs[0],
				Name:        &roleNames[0],
				Description: &roleDescriptions[0],
			},
		}}, &reply)
		assert.NoError(t, err)
	})

	storer.AddBuyer(context.Background(), routing.Buyer{CompanyCode: "test", ID: 123})

	t.Run("success - not registered", func(t *testing.T) {
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, middleware.Keys.CustomerKey, "test")
		req = req.WithContext(reqContext)
		var reply jsonrpc.AccountsReply
		err := svc.AddUserAccount(req, &jsonrpc.AccountsArgs{Roles: []*management.Role{
			{
				ID:          &roleIDs[0],
				Name:        &roleNames[0],
				Description: &roleDescriptions[0],
			},
		}, Emails: []string{"test@test123.com"}}, &reply)
		assert.NoError(t, err)
		users, err := userManager.List()
		found := false
		for _, u := range users.Users {
			if *u.Email == "test@test123.com" {
				found = true
			}
		}
		assert.True(t, found)
	})

	t.Run("success - registered", func(t *testing.T) {
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, middleware.Keys.CustomerKey, "test")
		req = req.WithContext(reqContext)
		var reply jsonrpc.AccountsReply
		err := svc.AddUserAccount(req, &jsonrpc.AccountsArgs{
			Roles: []*management.Role{
				{
					ID:          &roleIDs[0],
					Name:        &roleNames[0],
					Description: &roleDescriptions[0],
				},
			},
			Emails: []string{"test@test1.com"},
		}, &reply)
		assert.NoError(t, err)
	})
}

func TestAllRoles(t *testing.T) {
	t.Parallel()
	var jobManager = storage.LocalJobManager{}
	var roleManager = storage.NewLocalRoleManager()
	var storer = storage.InMemory{}
	var userManager = storage.NewLocalUserManager()

	roleIDs := []string{
		"rol_ScQpWhLvmTKRlqLU",
		"rol_8r0281hf2oC4cvrD",
		"rol_YfFrtom32or4vH89",
	}
	roleNames := []string{
		"Explorer",
		"Owner",
		"Admin",
	}
	roleDescriptions := []string{
		"Can access the explore tab.",
		"Can access and manage everything in an account.",
		"Can manage the Network Next system, including access to configstore.",
	}

	logger := log.NewNopLogger()

	svc := jsonrpc.AuthService{
		JobManager:  &jobManager,
		Logger:      logger,
		RoleCache:   make(map[string]*management.Role),
		RoleManager: roleManager,
		Storage:     &storer,
		UserManager: userManager,
	}

	svc.RoleCache["Admin"] = &management.Role{
		Description: &roleDescriptions[2],
		ID:          &roleIDs[2],
		Name:        &roleNames[2],
	}

	svc.RoleCache["Owner"] = &management.Role{
		Description: &roleDescriptions[1],
		ID:          &roleIDs[1],
		Name:        &roleNames[1],
	}

	svc.RoleCache["Explorer"] = &management.Role{
		Description: &roleDescriptions[0],
		ID:          &roleIDs[0],
		Name:        &roleNames[0],
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("failure - insufficient privileges", func(t *testing.T) {
		var reply jsonrpc.RolesReply
		err := svc.AllRoles(req, &jsonrpc.RolesArgs{}, &reply)
		assert.Error(t, err)
	})

	t.Run("success - owner", func(t *testing.T) {
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
			"Owner",
		})
		req = req.WithContext(reqContext)
		var reply jsonrpc.RolesReply
		err := svc.AllRoles(req, &jsonrpc.RolesArgs{}, &reply)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(reply.Roles))
		assert.Equal(t, svc.RoleCache["Explorer"].GetName(), reply.Roles[0].GetName())
		assert.Equal(t, svc.RoleCache["Explorer"].GetID(), reply.Roles[0].GetID())
		assert.Equal(t, svc.RoleCache["Explorer"].GetDescription(), reply.Roles[0].GetDescription())
		assert.Equal(t, svc.RoleCache["Owner"].GetName(), reply.Roles[1].GetName())
		assert.Equal(t, svc.RoleCache["Owner"].GetID(), reply.Roles[1].GetID())
		assert.Equal(t, svc.RoleCache["Owner"].GetDescription(), reply.Roles[1].GetDescription())
	})

	t.Run("success - admin", func(t *testing.T) {
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
			"Admin",
		})
		req = req.WithContext(reqContext)
		var reply jsonrpc.RolesReply
		err := svc.AllRoles(req, &jsonrpc.RolesArgs{}, &reply)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(reply.Roles))
		fmt.Println(reply.Roles)
		assert.Equal(t, svc.RoleCache["Admin"].GetName(), reply.Roles[0].GetName())
		assert.Equal(t, svc.RoleCache["Admin"].GetID(), reply.Roles[0].GetID())
		assert.Equal(t, svc.RoleCache["Admin"].GetDescription(), reply.Roles[0].GetDescription())
		assert.Equal(t, svc.RoleCache["Explorer"].GetName(), reply.Roles[1].GetName())
		assert.Equal(t, svc.RoleCache["Explorer"].GetID(), reply.Roles[1].GetID())
		assert.Equal(t, svc.RoleCache["Explorer"].GetDescription(), reply.Roles[1].GetDescription())
		assert.Equal(t, svc.RoleCache["Owner"].GetName(), reply.Roles[2].GetName())
		assert.Equal(t, svc.RoleCache["Owner"].GetID(), reply.Roles[2].GetID())
		assert.Equal(t, svc.RoleCache["Owner"].GetDescription(), reply.Roles[2].GetDescription())
	})
}

func TestUserRoles(t *testing.T) {
	t.Parallel()
	var userManager = storage.NewLocalUserManager()
	var jobManager = storage.LocalJobManager{}
	var storer = storage.InMemory{}

	roleIDs := []string{
		"rol_ScQpWhLvmTKRlqLU",
		"rol_8r0281hf2oC4cvrD",
		"rol_YfFrtom32or4vH89",
	}
	roleNames := []string{
		"Explorer",
		"Owner",
		"Admin",
	}
	roleDescriptions := []string{
		"Can access the explore tab",
		"Can access and manage everything in an account.",
		"Can manage the Network Next system, including access to configstore.",
	}

	IDs := []string{
		"123",
		"456",
		"789",
	}

	emails := []string{
		"test@test.com",
		"test@test1.com",
		"test@test2.com",
	}

	names := []string{
		"Frank",
		"George",
		"Lenny",
	}

	logger := log.NewNopLogger()

	svc := jsonrpc.AuthService{
		UserManager: userManager,
		JobManager:  &jobManager,
		Storage:     &storer,
		Logger:      logger,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("failure - insufficient privileges", func(t *testing.T) {
		var reply jsonrpc.RolesReply
		err := svc.UserRoles(req, &jsonrpc.RolesArgs{}, &reply)
		assert.Error(t, err)
	})

	t.Run("failure - no user ID", func(t *testing.T) {
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
			"Owner",
		})
		req = req.WithContext(reqContext)
		var reply jsonrpc.RolesReply
		err := svc.UserRoles(req, &jsonrpc.RolesArgs{}, &reply)
		assert.Error(t, err)
	})

	userManager.Create(&management.User{
		ID:    &IDs[1],
		Email: &emails[1],
		AppMetadata: map[string]interface{}{
			"company_code": "test",
		},
		Identities: []*management.UserIdentity{
			{
				UserID: &IDs[1],
			},
		},
		Name: &names[1],
	})

	userManager.AssignRoles(IDs[1], []*management.Role{
		{
			ID:          &roleIDs[0],
			Name:        &roleNames[0],
			Description: &roleDescriptions[0],
		},
	}...)

	t.Run("success", func(t *testing.T) {
		var reply jsonrpc.RolesReply
		err := svc.UserRoles(req, &jsonrpc.RolesArgs{UserID: "456"}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(reply.Roles))
		assert.Equal(t, roleIDs[0], reply.Roles[0].GetID())
		assert.Equal(t, roleNames[0], reply.Roles[0].GetName())
		assert.Equal(t, roleDescriptions[0], reply.Roles[0].GetDescription())
	})
}

func TestUpdateUserRoles(t *testing.T) {
	t.Parallel()
	var jobManager = storage.LocalJobManager{}
	var roleManager = storage.NewLocalRoleManager()
	var storer = storage.InMemory{}
	var userManager = storage.NewLocalUserManager()

	roleIDs := []string{
		"rol_ScQpWhLvmTKRlqLU",
		"rol_8r0281hf2oC4cvrD",
		"rol_YfFrtom32or4vH89",
	}
	roleNames := []string{
		"Explorer",
		"Owner",
		"Admin",
	}
	roleDescriptions := []string{
		"Can access the explore tab",
		"Can access and manage everything in an account.",
		"Can manage the Network Next system, including access to configstore.",
	}

	IDs := []string{
		"123",
		"456",
		"789",
	}

	emails := []string{
		"test@test.com",
		"test@test1.com",
		"test@test2.com",
	}

	names := []string{
		"Frank",
		"George",
		"Lenny",
	}

	logger := log.NewNopLogger()

	svc := jsonrpc.AuthService{
		JobManager:  &jobManager,
		Logger:      logger,
		RoleCache:   make(map[string]*management.Role),
		RoleManager: roleManager,
		Storage:     &storer,
		UserManager: userManager,
	}

	svc.RoleCache["Admin"] = &management.Role{
		Description: &roleDescriptions[2],
		ID:          &roleIDs[2],
		Name:        &roleNames[2],
	}

	svc.RoleCache["Owner"] = &management.Role{
		Description: &roleDescriptions[1],
		ID:          &roleIDs[1],
		Name:        &roleNames[1],
	}

	svc.RoleCache["Explorer"] = &management.Role{
		Description: &roleDescriptions[0],
		ID:          &roleIDs[0],
		Name:        &roleNames[0],
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("failure - insufficient privileges", func(t *testing.T) {
		var reply jsonrpc.RolesReply
		err := svc.UpdateUserRoles(req, &jsonrpc.RolesArgs{}, &reply)
		assert.Error(t, err)
	})

	t.Run("failure - no user ID", func(t *testing.T) {
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
			"Owner",
		})
		req = req.WithContext(reqContext)
		var reply jsonrpc.RolesReply
		err := svc.UpdateUserRoles(req, &jsonrpc.RolesArgs{}, &reply)
		assert.Error(t, err)
	})

	userManager.Create(&management.User{
		ID:    &IDs[1],
		Email: &emails[1],
		AppMetadata: map[string]interface{}{
			"company_code": "test",
		},
		Identities: []*management.UserIdentity{
			{
				UserID: &IDs[1],
			},
		},
		Name: &names[1],
	})

	userManager.AssignRoles(IDs[1], []*management.Role{
		{
			ID:          &roleIDs[0],
			Name:        &roleNames[0],
			Description: &roleDescriptions[0],
		},
	}...)

	t.Run("failure - !admin assigning admin", func(t *testing.T) {
		var reply jsonrpc.RolesReply
		err := svc.UpdateUserRoles(req, &jsonrpc.RolesArgs{UserID: "456", Roles: []*management.Role{
			{
				ID:          &roleIDs[2],
				Name:        &roleNames[2],
				Description: &roleDescriptions[2],
			},
		}}, &reply)
		assert.Error(t, err)
	})

	t.Run("success - !admin assigning !admin", func(t *testing.T) {
		var reply jsonrpc.RolesReply
		err := svc.UpdateUserRoles(req, &jsonrpc.RolesArgs{UserID: "456", Roles: []*management.Role{
			{
				ID:          &roleIDs[0],
				Name:        &roleNames[0],
				Description: &roleDescriptions[0],
			},
		}}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(reply.Roles))
		assert.Equal(t, roleIDs[0], reply.Roles[0].GetID())
		assert.Equal(t, roleNames[0], reply.Roles[0].GetName())
		assert.Equal(t, roleDescriptions[0], reply.Roles[0].GetDescription())
	})

	t.Run("success - admin assigning admin", func(t *testing.T) {
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
			"Admin",
		})
		req = req.WithContext(reqContext)
		var reply jsonrpc.RolesReply
		err := svc.UpdateUserRoles(req, &jsonrpc.RolesArgs{UserID: "456", Roles: []*management.Role{
			{
				ID:          &roleIDs[2],
				Name:        &roleNames[2],
				Description: &roleDescriptions[2],
			},
		}}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(reply.Roles))
		assert.Equal(t, roleIDs[2], reply.Roles[0].GetID())
		assert.Equal(t, roleNames[2], reply.Roles[0].GetName())
		assert.Equal(t, roleDescriptions[2], reply.Roles[0].GetDescription())
	})
}

func TestSetupCompanyAccount(t *testing.T) {
	t.Parallel()
	var jobManager = storage.LocalJobManager{}
	var roleManager = storage.NewLocalRoleManager()
	var storer = storage.InMemory{}
	var userManager = storage.NewLocalUserManager()

	IDs := []string{
		"123",
		"456",
		"789",
	}

	emails := []string{
		"test@test.com",
		"test@test1.com",
		"test@test2.com",
	}

	names := []string{
		"Frank",
		"George",
		"Lenny",
	}

	roleIDs := []string{
		"rol_ScQpWhLvmTKRlqLU",
		"rol_8r0281hf2oC4cvrD",
		"rol_YfFrtom32or4vH89",
	}
	roleNames := []string{
		"Explorer",
		"Owner",
		"Admin",
	}
	roleDescriptions := []string{
		"Can access the explore tab.",
		"Can access and manage everything in an account.",
		"Can manage the Network Next system, including access to configstore.",
	}

	logger := log.NewNopLogger()

	svc := jsonrpc.AuthService{
		JobManager:  &jobManager,
		Logger:      logger,
		RoleCache:   make(map[string]*management.Role),
		RoleManager: roleManager,
		Storage:     &storer,
		UserManager: userManager,
	}

	svc.RoleCache["Admin"] = &management.Role{
		Description: &roleDescriptions[2],
		ID:          &roleIDs[2],
		Name:        &roleNames[2],
	}

	svc.RoleCache["Owner"] = &management.Role{
		Description: &roleDescriptions[1],
		ID:          &roleIDs[1],
		Name:        &roleNames[1],
	}

	svc.RoleCache["Explorer"] = &management.Role{
		Description: &roleDescriptions[0],
		ID:          &roleIDs[0],
		Name:        &roleNames[0],
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("failure - insufficient privileges", func(t *testing.T) {
		var reply jsonrpc.SetupCompanyAccountReply
		err := svc.SetupCompanyAccount(req, &jsonrpc.SetupCompanyAccountArgs{}, &reply)
		assert.Error(t, err)
	})

	t.Run("failure - no company code", func(t *testing.T) {
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, middleware.Keys.UserKey, &jwt.Token{
			Claims: jwt.MapClaims{
				"email_verified": true,
			},
		})
		req = req.WithContext(reqContext)
		var reply jsonrpc.SetupCompanyAccountReply
		err := svc.SetupCompanyAccount(req, &jsonrpc.SetupCompanyAccountArgs{}, &reply)
		assert.Error(t, err)
	})

	t.Run("failure - malformed user", func(t *testing.T) {
		var reply jsonrpc.SetupCompanyAccountReply
		err := svc.SetupCompanyAccount(req, &jsonrpc.SetupCompanyAccountArgs{CompanyCode: "test", CompanyName: "Test"}, &reply)
		assert.Error(t, err)
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, middleware.Keys.UserKey, &jwt.Token{
			Claims: jwt.MapClaims{},
		})
		req = req.WithContext(reqContext)
		err = svc.SetupCompanyAccount(req, &jsonrpc.SetupCompanyAccountArgs{CompanyCode: "test", CompanyName: "Test"}, &reply)
		assert.Error(t, err)
		reqContext = req.Context()
		reqContext = context.WithValue(reqContext, middleware.Keys.UserKey, &jwt.Token{
			Claims: jwt.MapClaims{
				"sub": "1234",
			},
		})
		req = req.WithContext(reqContext)
		err = svc.SetupCompanyAccount(req, &jsonrpc.SetupCompanyAccountArgs{CompanyCode: "test", CompanyName: "Test"}, &reply)
		assert.Error(t, err)
	})

	t.Run("failure - no company name", func(t *testing.T) {
		var reply jsonrpc.SetupCompanyAccountReply
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, middleware.Keys.UserKey, &jwt.Token{
			Claims: jwt.MapClaims{
				"sub":   "123",
				"email": "test@test.com",
			},
		})
		req = req.WithContext(reqContext)
		err := svc.SetupCompanyAccount(req, &jsonrpc.SetupCompanyAccountArgs{CompanyCode: "test"}, &reply)
		assert.Error(t, err)
	})

	userManager.Create(&management.User{
		ID:    &IDs[0],
		Email: &emails[0],
		AppMetadata: map[string]interface{}{
			"company_code": "test",
		},
		Identities: []*management.UserIdentity{
			{
				UserID: &IDs[0],
			},
		},
		Name: &names[0],
	})

	userManager.AssignRoles("123", []*management.Role{
		{
			ID:          &roleIDs[0],
			Name:        &roleNames[0],
			Description: &roleDescriptions[0],
		},
	}...)

	t.Run("success - unassigned - new company", func(t *testing.T) {
		var reply jsonrpc.SetupCompanyAccountReply
		err := svc.SetupCompanyAccount(req, &jsonrpc.SetupCompanyAccountArgs{CompanyCode: "testing", CompanyName: "Testing"}, &reply)
		assert.NoError(t, err)

		userRoles, err := userManager.Roles("123")
		assert.NoError(t, err)
		assert.Equal(t, 1, len(userRoles.Roles))
		assert.Equal(t, roleNames[1], userRoles.Roles[0].GetName())
		assert.Equal(t, roleIDs[1], userRoles.Roles[0].GetID())
		assert.Equal(t, roleDescriptions[1], userRoles.Roles[0].GetDescription())
		customers := storer.Customers(req.Context())
		assert.Equal(t, 2, len(customers))
		assert.Equal(t, "testing", customers[1].Code)
		assert.Equal(t, "Testing", customers[1].Name)
	})

	storer.AddCustomer(context.Background(), routing.Customer{Code: "test-test", Name: "Test Test", AutomaticSignInDomains: "test2.com"})
	storer.AddBuyer(context.Background(), routing.Buyer{CompanyCode: "test-test", ID: 123})

	t.Run("failure - unassigned - old company - wrong domain", func(t *testing.T) {
		var reply jsonrpc.SetupCompanyAccountReply
		err := svc.SetupCompanyAccount(req, &jsonrpc.SetupCompanyAccountArgs{CompanyCode: "test-test"}, &reply)
		assert.Error(t, err)
	})

	t.Run("failure - assigned", func(t *testing.T) {
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, middleware.Keys.CustomerKey, "test-test")
		req = req.WithContext(reqContext)
		var reply jsonrpc.SetupCompanyAccountReply
		err := svc.SetupCompanyAccount(req, &jsonrpc.SetupCompanyAccountArgs{CompanyCode: "test-test-test"}, &reply)
		assert.Error(t, err)
	})
}

func TestUpdateAccountDetails(t *testing.T) {
	t.Parallel()
	var userManager = storage.NewLocalUserManager()
	var jobManager = storage.LocalJobManager{}
	var storer = storage.InMemory{}

	IDs := []string{
		"123",
		"456",
		"789",
	}

	emails := []string{
		"test@test.com",
		"test@test1.com",
		"test@test2.com",
	}

	names := []string{
		"Frank",
		"George",
		"Lenny",
	}

	userManager.Create(&management.User{
		ID:    &IDs[0],
		Email: &emails[0],
		AppMetadata: map[string]interface{}{
			"company_code": "test",
		},
		Identities: []*management.UserIdentity{
			{
				UserID: &IDs[0],
			},
		},
		Name: &names[0],
	})

	logger := log.NewNopLogger()

	svc := jsonrpc.AuthService{
		UserManager: userManager,
		JobManager:  &jobManager,
		Storage:     &storer,
		Logger:      logger,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("failure - insufficient privileges", func(t *testing.T) {
		var reply jsonrpc.UpdateAccountDetailsReply
		err := svc.UpdateAccountDetails(req, &jsonrpc.UpdateAccountDetailsArgs{}, &reply)
		assert.Error(t, err)
	})

	t.Run("failure - malformed user context", func(t *testing.T) {
		req = middleware.SetIsAnonymous(req, false)
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, middleware.Keys.UserKey, &jwt.Token{
			Claims: jwt.MapClaims{
				"email_verified": true,
			},
		})
		req = req.WithContext(reqContext)

		var reply jsonrpc.UpdateAccountDetailsReply
		err := svc.UpdateAccountDetails(req, &jsonrpc.UpdateAccountDetailsArgs{}, &reply)
		assert.Error(t, err)
	})

	t.Run("success - no newsletter", func(t *testing.T) {
		var reply jsonrpc.UpdateAccountDetailsReply
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, middleware.Keys.UserKey, &jwt.Token{
			Claims: jwt.MapClaims{
				"email": "test@test.com",
				"sub":   "123",
			},
		})
		req = req.WithContext(reqContext)
		err := svc.UpdateAccountDetails(req, &jsonrpc.UpdateAccountDetailsArgs{Newsletter: false}, &reply)
		assert.NoError(t, err)
		userAccount, err := userManager.Read("123")
		assert.NoError(t, err)
		assert.False(t, userAccount.AppMetadata["newsletter"].(bool))
	})

	t.Run("success - yes newsletter", func(t *testing.T) {
		var reply jsonrpc.UpdateAccountDetailsReply
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, middleware.Keys.UserKey, &jwt.Token{
			Claims: jwt.MapClaims{
				"sub": "123",
			},
		})
		req = req.WithContext(reqContext)
		err := svc.UpdateAccountDetails(req, &jsonrpc.UpdateAccountDetailsArgs{Newsletter: true}, &reply)
		assert.NoError(t, err)
		userAccount, err := userManager.Read("123")
		assert.NoError(t, err)
		assert.True(t, userAccount.AppMetadata["newsletter"].(bool))
	})

	t.Run("success - no first name", func(t *testing.T) {
		var reply jsonrpc.UpdateAccountDetailsReply
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, middleware.Keys.UserKey, &jwt.Token{
			Claims: jwt.MapClaims{
				"sub": "123",
			},
		})
		req = req.WithContext(reqContext)
		err := svc.UpdateAccountDetails(req, &jsonrpc.UpdateAccountDetailsArgs{FirstName: names[0]}, &reply)
		assert.NoError(t, err)
		userAccount, err := userManager.Read("123")
		assert.NoError(t, err)
		assert.Equal(t, names[0], userAccount.GetGivenName())
	})

	t.Run("success - same first name", func(t *testing.T) {
		var reply jsonrpc.UpdateAccountDetailsReply
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, middleware.Keys.UserKey, &jwt.Token{
			Claims: jwt.MapClaims{
				"sub": "123",
			},
		})
		req = req.WithContext(reqContext)
		err := svc.UpdateAccountDetails(req, &jsonrpc.UpdateAccountDetailsArgs{FirstName: names[0]}, &reply)
		assert.NoError(t, err)
		userAccount, err := userManager.Read("123")
		assert.NoError(t, err)
		assert.Equal(t, names[0], userAccount.GetGivenName())
	})

	t.Run("success - different first name", func(t *testing.T) {
		var reply jsonrpc.UpdateAccountDetailsReply
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, middleware.Keys.UserKey, &jwt.Token{
			Claims: jwt.MapClaims{
				"sub": "123",
			},
		})
		req = req.WithContext(reqContext)
		err := svc.UpdateAccountDetails(req, &jsonrpc.UpdateAccountDetailsArgs{FirstName: names[1]}, &reply)
		assert.NoError(t, err)
		userAccount, err := userManager.Read("123")
		assert.NoError(t, err)
		assert.Equal(t, names[1], userAccount.GetGivenName())
	})

	t.Run("success - no last name", func(t *testing.T) {
		var reply jsonrpc.UpdateAccountDetailsReply
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, middleware.Keys.UserKey, &jwt.Token{
			Claims: jwt.MapClaims{
				"sub": "123",
			},
		})
		req = req.WithContext(reqContext)
		err := svc.UpdateAccountDetails(req, &jsonrpc.UpdateAccountDetailsArgs{LastName: names[0]}, &reply)
		assert.NoError(t, err)
		userAccount, err := userManager.Read("123")
		assert.NoError(t, err)
		assert.Equal(t, names[0], userAccount.GetFamilyName())
	})

	t.Run("success - same last name", func(t *testing.T) {
		var reply jsonrpc.UpdateAccountDetailsReply
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, middleware.Keys.UserKey, &jwt.Token{
			Claims: jwt.MapClaims{
				"sub": "123",
			},
		})
		req = req.WithContext(reqContext)
		err := svc.UpdateAccountDetails(req, &jsonrpc.UpdateAccountDetailsArgs{LastName: names[0]}, &reply)
		assert.NoError(t, err)
		userAccount, err := userManager.Read("123")
		assert.NoError(t, err)
		assert.Equal(t, names[0], userAccount.GetFamilyName())
	})

	t.Run("success - different last name", func(t *testing.T) {
		var reply jsonrpc.UpdateAccountDetailsReply
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, middleware.Keys.UserKey, &jwt.Token{
			Claims: jwt.MapClaims{
				"sub": "123",
			},
		})
		req = req.WithContext(reqContext)
		err := svc.UpdateAccountDetails(req, &jsonrpc.UpdateAccountDetailsArgs{LastName: names[1]}, &reply)
		assert.NoError(t, err)
		userAccount, err := userManager.Read("123")
		assert.NoError(t, err)
		assert.Equal(t, names[1], userAccount.GetFamilyName())
	})
}

func TestSendVerificationEmail(t *testing.T) {
	t.Parallel()
	var userManager = storage.NewLocalUserManager()
	var jobManager = storage.LocalJobManager{}
	var storer = storage.InMemory{}

	logger := log.NewNopLogger()

	svc := jsonrpc.AuthService{
		UserManager: userManager,
		JobManager:  &jobManager,
		Storage:     &storer,
		Logger:      logger,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("failure - insufficient privileges", func(t *testing.T) {
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, middleware.Keys.VerifiedKey, true)

		req = req.WithContext(reqContext)

		var reply jsonrpc.VerifyEmailReply
		err := svc.ResendVerificationEmail(req, &jsonrpc.VerifyEmailArgs{}, &reply)
		assert.Error(t, err)
	})

	t.Run("failure - no ID", func(t *testing.T) {
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, middleware.Keys.VerifiedKey, false)

		req = req.WithContext(reqContext)

		var reply jsonrpc.VerifyEmailReply
		err := svc.ResendVerificationEmail(req, &jsonrpc.VerifyEmailArgs{}, &reply)
		assert.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, middleware.Keys.VerifiedKey, false)

		req = req.WithContext(reqContext)

		var reply jsonrpc.VerifyEmailReply
		err := svc.ResendVerificationEmail(req, &jsonrpc.VerifyEmailArgs{UserID: "123"}, &reply)
		assert.NoError(t, err)
	})
}

func TestUpdateAutoSignupDomains(t *testing.T) {
	t.Parallel()
	var userManager = storage.NewLocalUserManager()
	var jobManager = storage.LocalJobManager{}
	var storer = storage.InMemory{}

	logger := log.NewNopLogger()

	svc := jsonrpc.AuthService{
		UserManager: userManager,
		JobManager:  &jobManager,
		Storage:     &storer,
		Logger:      logger,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("failure - insufficient privileges", func(t *testing.T) {
		var reply jsonrpc.UpdateDomainsReply
		err := svc.UpdateAutoSignupDomains(req, &jsonrpc.UpdateDomainsArgs{}, &reply)
		assert.Error(t, err)
	})

	t.Run("failure - no company code", func(t *testing.T) {
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
			"Owner",
		})
		req = req.WithContext(reqContext)
		var reply jsonrpc.UpdateDomainsReply
		err := svc.UpdateAutoSignupDomains(req, &jsonrpc.UpdateDomainsArgs{}, &reply)
		assert.Error(t, err)
	})

	t.Run("failure - no company code", func(t *testing.T) {
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, middleware.Keys.CustomerKey, "test")
		req = req.WithContext(reqContext)
		var reply jsonrpc.UpdateDomainsReply
		err := svc.UpdateAutoSignupDomains(req, &jsonrpc.UpdateDomainsArgs{}, &reply)
		assert.Error(t, err)
	})

	storer.AddCustomer(context.Background(), routing.Customer{Code: "test", Name: "Test"})

	t.Run("success", func(t *testing.T) {
		var reply jsonrpc.UpdateDomainsReply
		err := svc.UpdateAutoSignupDomains(req, &jsonrpc.UpdateDomainsArgs{}, &reply)
		assert.NoError(t, err)
	})
}
