package jsonrpc_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-kit/kit/log"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/storage"
	"github.com/networknext/backend/modules/transport/jsonrpc"
	"github.com/stretchr/testify/assert"
	"gopkg.in/auth0.v4/management"
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
		authMiddleware := jsonrpc.AuthMiddleware("", http.HandlerFunc(noopHandler), false, []string{})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		res := httptest.NewRecorder()

		authMiddleware.ServeHTTP(res, req)
		assert.Equal(t, http.StatusOK, res.Code)
	})

	t.Run("check auth claims", func(t *testing.T) {
		authMiddleware := jsonrpc.AuthMiddleware("oQJH3YPHdvZJnxCPo1Irtz5UKi5zrr6n", http.HandlerFunc(noopHandler), false, []string{})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Add("Authorization", "Bearer "+jwtSideload)
		res := httptest.NewRecorder()

		authMiddleware.ServeHTTP(res, req)
		assert.Equal(t, http.StatusOK, res.Code)
	})

	t.Run("anonymous auth", func(t *testing.T) {
		authMiddleware := jsonrpc.AuthMiddleware("oQJH3YPHdvZJnxCPo1Irtz5UKi5zrr6n", http.HandlerFunc(noopHandler), false, []string{})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		res := httptest.NewRecorder()

		authMiddleware.ServeHTTP(res, req)
		assert.Equal(t, http.StatusOK, res.Code)
	})
}

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
		"Viewer",
		"Owner",
		"Admin",
	}
	roleDescriptions := []string{
		"Can see current sessions and the map.",
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
		reqContext = context.WithValue(reqContext, jsonrpc.Keys.CompanyKey, "test")
		req = req.WithContext(reqContext)
		var reply jsonrpc.AccountsReply
		err := svc.AllAccounts(req, &jsonrpc.AccountsArgs{}, &reply)
		assert.Error(t, err)
	})

	t.Run("all - success - no users in company", func(t *testing.T) {
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, jsonrpc.Keys.UserKey, &jwt.Token{
			Claims: jwt.MapClaims{
				"email": "test@test.com",
			},
		})
		reqContext = context.WithValue(reqContext, jsonrpc.Keys.CompanyKey, "test")
		reqContext = context.WithValue(reqContext, jsonrpc.Keys.RolesKey, []string{
			"Admin",
		})
		req = req.WithContext(reqContext)
		var reply jsonrpc.AccountsReply
		err := svc.AllAccounts(req, &jsonrpc.AccountsArgs{}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, 0, len(reply.UserAccounts))
	})

	t.Run("all - success", func(t *testing.T) {
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, jsonrpc.Keys.UserKey, &jwt.Token{
			Claims: jwt.MapClaims{
				"email": "test@test.com",
			},
		})
		reqContext = context.WithValue(reqContext, jsonrpc.Keys.CompanyKey, "test-test")
		reqContext = context.WithValue(reqContext, jsonrpc.Keys.RolesKey, []string{
			"Admin",
		})
		req = req.WithContext(reqContext)
		var reply jsonrpc.AccountsReply
		err := svc.AllAccounts(req, &jsonrpc.AccountsArgs{}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(reply.UserAccounts))
		assert.Equal(t, "George", reply.UserAccounts[0].Name)
		assert.Equal(t, "test@test1.com", reply.UserAccounts[0].Email)
		assert.Equal(t, "test-test", reply.UserAccounts[0].CompanyCode)
		assert.Equal(t, "Test Test", reply.UserAccounts[0].CompanyName)
		assert.Equal(t, IDs[1], reply.UserAccounts[0].UserID)
		assert.Equal(t, fmt.Sprintf("%016x", 456), reply.UserAccounts[0].ID)
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
		"Viewer",
		"Owner",
		"Admin",
	}
	roleDescriptions := []string{
		"Can see current sessions and the map.",
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
			"company_code": "test-test",
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
		reqContext = context.WithValue(reqContext, jsonrpc.Keys.UserKey, &jwt.Token{
			Claims: jwt.MapClaims{
				"email": "test@test.com",
				"sub":   "123",
			},
		})
		reqContext = context.WithValue(reqContext, jsonrpc.Keys.CompanyKey, "test")

		req = req.WithContext(reqContext)
		var reply jsonrpc.AccountReply
		err := svc.UserAccount(req, &jsonrpc.AccountArgs{UserID: "123"}, &reply)
		assert.NoError(t, err)

		assert.Nil(t, reply.Domains)
		assert.Equal(t, "123", reply.UserAccount.UserID)
		assert.Equal(t, fmt.Sprintf("%016x", 123), reply.UserAccount.ID)
		assert.Equal(t, "test", reply.UserAccount.CompanyCode)
		assert.Equal(t, "Test", reply.UserAccount.CompanyName)
		assert.Equal(t, "Frank", reply.UserAccount.Name)
		assert.Equal(t, "test@test.com", reply.UserAccount.Email)
	})

	t.Run("success - request user - admin", func(t *testing.T) {
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, jsonrpc.Keys.UserKey, &jwt.Token{
			Claims: jwt.MapClaims{
				"email": "test@test.com",
				"sub":   "123",
			},
		})
		reqContext = context.WithValue(reqContext, jsonrpc.Keys.CompanyKey, "test")
		reqContext = context.WithValue(reqContext, jsonrpc.Keys.RolesKey, []string{
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
		reqContext = context.WithValue(reqContext, jsonrpc.Keys.UserKey, &jwt.Token{
			Claims: jwt.MapClaims{
				"email": "test@test.com",
				"sub":   "123",
			},
		})
		reqContext = context.WithValue(reqContext, jsonrpc.Keys.CompanyKey, "test")
		reqContext = context.WithValue(reqContext, jsonrpc.Keys.RolesKey, []string{})
		req = req.WithContext(reqContext)
		var reply jsonrpc.AccountReply
		err := svc.UserAccount(req, &jsonrpc.AccountArgs{UserID: "456"}, &reply)
		assert.Error(t, err)
	})

	t.Run("success - random user - admin", func(t *testing.T) {
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, jsonrpc.Keys.UserKey, &jwt.Token{
			Claims: jwt.MapClaims{
				"email": "test@test.com",
				"sub":   "123",
			},
		})
		reqContext = context.WithValue(reqContext, jsonrpc.Keys.CompanyKey, "test")
		reqContext = context.WithValue(reqContext, jsonrpc.Keys.RolesKey, []string{
			"Admin",
		})
		req = req.WithContext(reqContext)
		var reply jsonrpc.AccountReply
		err := svc.UserAccount(req, &jsonrpc.AccountArgs{UserID: "456"}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, 0, len(reply.Domains))
		assert.Equal(t, "456", reply.UserAccount.UserID)
		assert.Equal(t, fmt.Sprintf("%016x", 456), reply.UserAccount.ID)
		assert.Equal(t, "test-test", reply.UserAccount.CompanyCode)
		assert.Equal(t, "Test Test", reply.UserAccount.CompanyName)
		assert.Equal(t, "George", reply.UserAccount.Name)
		assert.Equal(t, "test@test1.com", reply.UserAccount.Email)
	})

	t.Run("success - random user - admin - no company", func(t *testing.T) {
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, jsonrpc.Keys.UserKey, &jwt.Token{
			Claims: jwt.MapClaims{
				"email": "test@test.com",
				"sub":   "123",
			},
		})
		reqContext = context.WithValue(reqContext, jsonrpc.Keys.CompanyKey, "test")
		reqContext = context.WithValue(reqContext, jsonrpc.Keys.RolesKey, []string{
			"Admin",
		})
		req = req.WithContext(reqContext)
		var reply jsonrpc.AccountReply
		err := svc.UserAccount(req, &jsonrpc.AccountArgs{UserID: "789"}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, 0, len(reply.Domains))
		assert.Equal(t, "789", reply.UserAccount.UserID)
		assert.Equal(t, "", reply.UserAccount.ID)
		assert.Equal(t, "", reply.UserAccount.CompanyCode)
		assert.Equal(t, "", reply.UserAccount.CompanyName)
		assert.Equal(t, "Lenny", reply.UserAccount.Name)
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
		"Viewer",
		"Owner",
		"Admin",
	}
	roleDescriptions := []string{
		"Can see current sessions and the map.",
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
		reqContext = context.WithValue(reqContext, jsonrpc.Keys.RolesKey, []string{
			"Owner",
		})
		req = req.WithContext(reqContext)
		var reply jsonrpc.AccountReply
		err := svc.DeleteUserAccount(req, &jsonrpc.AccountArgs{}, &reply)
		assert.Error(t, err)
	})

	t.Run("success - same company", func(t *testing.T) {
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, jsonrpc.Keys.CompanyKey, "test")
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
		reqContext = context.WithValue(reqContext, jsonrpc.Keys.RolesKey, []string{
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

	roleIDs := []string{
		"rol_ScQpWhLvmTKRlqLU",
		"rol_8r0281hf2oC4cvrD",
		"rol_YfFrtom32or4vH89",
	}
	roleNames := []string{
		"Viewer",
		"Owner",
		"Admin",
	}
	roleDescriptions := []string{
		"Can see current sessions and the map.",
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

	storer.AddCustomer(context.Background(), routing.Customer{Code: "test", Name: "Test", AutomaticSignInDomains: "google.com,test.com"})

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("failure - insufficient privileges", func(t *testing.T) {
		var reply jsonrpc.AccountsReply
		err := svc.AddUserAccount(req, &jsonrpc.AccountsArgs{}, &reply)
		assert.Error(t, err)
	})

	t.Run("failure - no roles", func(t *testing.T) {
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, jsonrpc.Keys.RolesKey, []string{
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
		reqContext = context.WithValue(reqContext, jsonrpc.Keys.CompanyKey, "test")
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
		reqContext = context.WithValue(reqContext, jsonrpc.Keys.CompanyKey, "test")
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
		reqContext = context.WithValue(reqContext, jsonrpc.Keys.CompanyKey, "test")
		req = req.WithContext(reqContext)
		var reply jsonrpc.AccountsReply
		err := svc.AddUserAccount(req, &jsonrpc.AccountsArgs{Roles: []*management.Role{
			{
				ID:          &roleIDs[0],
				Name:        &roleNames[0],
				Description: &roleDescriptions[0],
			},
		}, Emails: []string{"test@test1.com"}}, &reply)
		assert.NoError(t, err)
	})
}

func TestAllRoles(t *testing.T) {
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
		"Viewer",
		"Owner",
		"Admin",
	}
	roleDescriptions := []string{
		"Can see current sessions and the map.",
		"Can access and manage everything in an account.",
		"Can manage the Network Next system, including access to configstore.",
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
		err := svc.AllRoles(req, &jsonrpc.RolesArgs{}, &reply)
		assert.Error(t, err)
	})

	t.Run("success - owner", func(t *testing.T) {
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, jsonrpc.Keys.RolesKey, []string{
			"Owner",
		})
		req = req.WithContext(reqContext)
		var reply jsonrpc.RolesReply
		err := svc.AllRoles(req, &jsonrpc.RolesArgs{}, &reply)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(reply.Roles))
		assert.Equal(t, roleNames[0], *reply.Roles[0].Name)
		assert.Equal(t, roleIDs[0], *reply.Roles[0].ID)
		assert.Equal(t, roleDescriptions[0], *reply.Roles[0].Description)
		assert.Equal(t, roleNames[1], *reply.Roles[1].Name)
		assert.Equal(t, roleIDs[1], *reply.Roles[1].ID)
		assert.Equal(t, roleDescriptions[1], *reply.Roles[1].Description)
	})

	t.Run("success - admin", func(t *testing.T) {
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, jsonrpc.Keys.RolesKey, []string{
			"Admin",
		})
		req = req.WithContext(reqContext)
		var reply jsonrpc.RolesReply
		err := svc.AllRoles(req, &jsonrpc.RolesArgs{}, &reply)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(reply.Roles))
		assert.Equal(t, roleNames[0], *reply.Roles[0].Name)
		assert.Equal(t, roleIDs[0], *reply.Roles[0].ID)
		assert.Equal(t, roleDescriptions[0], *reply.Roles[0].Description)
		assert.Equal(t, roleNames[1], *reply.Roles[1].Name)
		assert.Equal(t, roleIDs[1], *reply.Roles[1].ID)
		assert.Equal(t, roleDescriptions[1], *reply.Roles[1].Description)
		assert.Equal(t, roleNames[2], *reply.Roles[2].Name)
		assert.Equal(t, roleIDs[2], *reply.Roles[2].ID)
		assert.Equal(t, roleDescriptions[2], *reply.Roles[2].Description)
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
		"Viewer",
		"Owner",
		"Admin",
	}
	roleDescriptions := []string{
		"Can see current sessions and the map.",
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
		reqContext = context.WithValue(reqContext, jsonrpc.Keys.RolesKey, []string{
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
		assert.Equal(t, roleIDs[0], *reply.Roles[0].ID)
		assert.Equal(t, roleNames[0], *reply.Roles[0].Name)
		assert.Equal(t, roleDescriptions[0], *reply.Roles[0].Description)
	})
}

func TestUpdateUserRoles(t *testing.T) {
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
		"Viewer",
		"Owner",
		"Admin",
	}
	roleDescriptions := []string{
		"Can see current sessions and the map.",
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
		err := svc.UpdateUserRoles(req, &jsonrpc.RolesArgs{}, &reply)
		assert.Error(t, err)
	})

	t.Run("failure - no user ID", func(t *testing.T) {
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, jsonrpc.Keys.RolesKey, []string{
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
		assert.Equal(t, roleIDs[0], *reply.Roles[0].ID)
		assert.Equal(t, roleNames[0], *reply.Roles[0].Name)
		assert.Equal(t, roleDescriptions[0], *reply.Roles[0].Description)
	})

	t.Run("success - admin assigning admin", func(t *testing.T) {
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, jsonrpc.Keys.RolesKey, []string{
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
		assert.Equal(t, roleIDs[2], *reply.Roles[0].ID)
		assert.Equal(t, roleNames[2], *reply.Roles[0].Name)
		assert.Equal(t, roleDescriptions[2], *reply.Roles[0].Description)
	})
}

func TestUpdateCompanyInformation(t *testing.T) {
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

	roleIDs := []string{
		"rol_ScQpWhLvmTKRlqLU",
		"rol_8r0281hf2oC4cvrD",
		"rol_YfFrtom32or4vH89",
	}
	roleNames := []string{
		"Viewer",
		"Owner",
		"Admin",
	}
	roleDescriptions := []string{
		"Can see current sessions and the map.",
		"Can access and manage everything in an account.",
		"Can manage the Network Next system, including access to configstore.",
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
		var reply jsonrpc.CompanyNameReply
		err := svc.UpdateCompanyInformation(req, &jsonrpc.CompanyNameArgs{}, &reply)
		assert.Error(t, err)
	})

	t.Run("failure - no company code", func(t *testing.T) {
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, jsonrpc.Keys.UserKey, &jwt.Token{
			Claims: jwt.MapClaims{
				"email_verified": true,
			},
		})
		req = req.WithContext(reqContext)
		var reply jsonrpc.CompanyNameReply
		err := svc.UpdateCompanyInformation(req, &jsonrpc.CompanyNameArgs{}, &reply)
		assert.Error(t, err)
	})

	t.Run("failure - malformed user", func(t *testing.T) {
		var reply jsonrpc.CompanyNameReply
		err := svc.UpdateCompanyInformation(req, &jsonrpc.CompanyNameArgs{CompanyCode: "test", CompanyName: "Test"}, &reply)
		assert.Error(t, err)
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, jsonrpc.Keys.UserKey, &jwt.Token{
			Claims: jwt.MapClaims{},
		})
		req = req.WithContext(reqContext)
		err = svc.UpdateCompanyInformation(req, &jsonrpc.CompanyNameArgs{CompanyCode: "test", CompanyName: "Test"}, &reply)
		assert.Error(t, err)
		reqContext = req.Context()
		reqContext = context.WithValue(reqContext, jsonrpc.Keys.UserKey, &jwt.Token{
			Claims: jwt.MapClaims{
				"sub": "1234",
			},
		})
		req = req.WithContext(reqContext)
		err = svc.UpdateCompanyInformation(req, &jsonrpc.CompanyNameArgs{CompanyCode: "test", CompanyName: "Test"}, &reply)
		assert.Error(t, err)
	})

	t.Run("failure - no company name", func(t *testing.T) {
		var reply jsonrpc.CompanyNameReply
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, jsonrpc.Keys.UserKey, &jwt.Token{
			Claims: jwt.MapClaims{
				"sub":   "123",
				"email": "test@test.com",
			},
		})
		req = req.WithContext(reqContext)
		err := svc.UpdateCompanyInformation(req, &jsonrpc.CompanyNameArgs{CompanyCode: "test"}, &reply)
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
		var reply jsonrpc.CompanyNameReply
		err := svc.UpdateCompanyInformation(req, &jsonrpc.CompanyNameArgs{CompanyCode: "test", CompanyName: "Test"}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, 2, len(reply.NewRoles))
		assert.Equal(t, roleNames[0], *reply.NewRoles[0].Name)
		assert.Equal(t, roleIDs[0], *reply.NewRoles[0].ID)
		assert.Equal(t, roleDescriptions[0], *reply.NewRoles[0].Description)
		assert.Equal(t, roleNames[1], *reply.NewRoles[1].Name)
		assert.Equal(t, roleIDs[1], *reply.NewRoles[1].ID)
		assert.Equal(t, roleDescriptions[1], *reply.NewRoles[1].Description)
		userRoles, err := userManager.Roles("123")
		assert.NoError(t, err)
		assert.Equal(t, 2, len(userRoles.Roles))
		assert.Equal(t, roleNames[0], *userRoles.Roles[0].Name)
		assert.Equal(t, roleIDs[0], *userRoles.Roles[0].ID)
		assert.Equal(t, roleDescriptions[0], *userRoles.Roles[0].Description)
		assert.Equal(t, roleNames[1], *userRoles.Roles[1].Name)
		assert.Equal(t, roleIDs[1], *userRoles.Roles[1].ID)
		assert.Equal(t, roleDescriptions[1], *userRoles.Roles[1].Description)
		customers := storer.Customers()
		assert.Equal(t, 1, len(customers))
		assert.Equal(t, "test", customers[0].Code)
		assert.Equal(t, "Test", customers[0].Name)
	})

	storer.AddCustomer(context.Background(), routing.Customer{Code: "test-test", Name: "Test Test", AutomaticSignInDomains: "test2.com"})
	storer.AddBuyer(context.Background(), routing.Buyer{CompanyCode: "test-test", ID: 123})

	t.Run("failure - unassigned - old company - wrong domain", func(t *testing.T) {
		var reply jsonrpc.CompanyNameReply
		err := svc.UpdateCompanyInformation(req, &jsonrpc.CompanyNameArgs{CompanyCode: "test-test"}, &reply)
		assert.Error(t, err)
	})

	t.Run("failure - assigned - different code - insufficient privileges", func(t *testing.T) {
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, jsonrpc.Keys.CompanyKey, "test-test")
		req = req.WithContext(reqContext)
		var reply jsonrpc.CompanyNameReply
		err := svc.UpdateCompanyInformation(req, &jsonrpc.CompanyNameArgs{CompanyCode: "test-test-test"}, &reply)
		assert.Error(t, err)
	})

	t.Run("failure - assigned - different code - company exists", func(t *testing.T) {
		var reply jsonrpc.CompanyNameReply
		err := svc.UpdateCompanyInformation(req, &jsonrpc.CompanyNameArgs{CompanyCode: "test"}, &reply)
		assert.Error(t, err)
	})

	t.Run("failure - assigned - different code - no company name", func(t *testing.T) {
		var reply jsonrpc.CompanyNameReply
		err := svc.UpdateCompanyInformation(req, &jsonrpc.CompanyNameArgs{CompanyCode: "test-test-test"}, &reply)
		assert.Error(t, err)
	})

	t.Run("success - assigned - different code", func(t *testing.T) {
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, jsonrpc.Keys.RolesKey, []string{
			"Owner",
		})
		req = req.WithContext(reqContext)
		var reply jsonrpc.CompanyNameReply
		err := svc.UpdateCompanyInformation(req, &jsonrpc.CompanyNameArgs{CompanyCode: "test-test-test", CompanyName: "Test Test Test"}, &reply)
		assert.NoError(t, err)

		customers := storer.Customers()
		assert.Equal(t, 2, len(customers))
		assert.Equal(t, "test-test-test", customers[1].Code)
		assert.Equal(t, "Test Test Test", customers[1].Name)
		buyer, err := storer.BuyerWithCompanyCode("test-test-test")
		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("%016x", 123), fmt.Sprintf("%016x", buyer.ID))
		assert.Equal(t, "test-test-test", buyer.CompanyCode)
		userAccount, err := userManager.Read("123")
		assert.NoError(t, err)
		assert.Equal(t, "test-test-test", userAccount.AppMetadata["company_code"])
	})

	t.Run("failure - assigned - same code - insufficient privileges", func(t *testing.T) {
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, jsonrpc.Keys.RolesKey, []string{})
		req = req.WithContext(reqContext)
		var reply jsonrpc.CompanyNameReply
		err := svc.UpdateCompanyInformation(req, &jsonrpc.CompanyNameArgs{CompanyCode: "test-test-test"}, &reply)
		assert.Error(t, err)
	})

	t.Run("failure - assigned - same code - no name", func(t *testing.T) {
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, jsonrpc.Keys.RolesKey, []string{
			"Owner",
		})
		req = req.WithContext(reqContext)
		var reply jsonrpc.CompanyNameReply
		err := svc.UpdateCompanyInformation(req, &jsonrpc.CompanyNameArgs{CompanyCode: "test-test-test"}, &reply)
		assert.Error(t, err)
	})

	t.Run("success - assigned - same code", func(t *testing.T) {
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, jsonrpc.Keys.CompanyKey, "test-test-test")
		req = req.WithContext(reqContext)
		var reply jsonrpc.CompanyNameReply
		err := svc.UpdateCompanyInformation(req, &jsonrpc.CompanyNameArgs{CompanyCode: "test-test-test", CompanyName: "Test 3"}, &reply)
		assert.NoError(t, err)

		customer, err := storer.Customer("test-test-test")
		assert.NoError(t, err)
		assert.Equal(t, "test-test-test", customer.Code)
		assert.Equal(t, "Test 3", customer.Name)
		buyer, err := storer.BuyerWithCompanyCode("test-test-test")
		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("%016x", 123), fmt.Sprintf("%016x", buyer.ID))
		assert.Equal(t, "test-test-test", buyer.CompanyCode)
		userAccount, err := userManager.Read("123")
		assert.NoError(t, err)
		assert.Equal(t, "test-test-test", userAccount.AppMetadata["company_code"])
	})
}

func TestUpdateAccountInformation(t *testing.T) {
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
		var reply jsonrpc.AccountSettingsReply
		err := svc.UpdateAccountSettings(req, &jsonrpc.AccountSettingsArgs{}, &reply)
		assert.Error(t, err)
	})

	t.Run("failure - malformed user context", func(t *testing.T) {
		req = jsonrpc.SetIsAnonymous(req, false)
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, jsonrpc.Keys.UserKey, &jwt.Token{
			Claims: jwt.MapClaims{
				"email_verified": true,
			},
		})
		req = req.WithContext(reqContext)

		var reply jsonrpc.AccountSettingsReply
		err := svc.UpdateAccountSettings(req, &jsonrpc.AccountSettingsArgs{}, &reply)
		assert.Error(t, err)
	})

	t.Run("success - no newsletter", func(t *testing.T) {
		var reply jsonrpc.AccountSettingsReply
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, jsonrpc.Keys.UserKey, &jwt.Token{
			Claims: jwt.MapClaims{
				"email": "test@test.com",
				"sub":   "123",
			},
		})
		req = req.WithContext(reqContext)
		err := svc.UpdateAccountSettings(req, &jsonrpc.AccountSettingsArgs{NewsLetterConsent: false}, &reply)
		assert.NoError(t, err)
		userAccount, err := userManager.Read("123")
		assert.NoError(t, err)
		assert.False(t, userAccount.AppMetadata["newsletter"].(bool))
	})

	t.Run("success - yes newsletter", func(t *testing.T) {
		var reply jsonrpc.AccountSettingsReply
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, jsonrpc.Keys.UserKey, &jwt.Token{
			Claims: jwt.MapClaims{
				"sub": "123",
			},
		})
		req = req.WithContext(reqContext)
		err := svc.UpdateAccountSettings(req, &jsonrpc.AccountSettingsArgs{NewsLetterConsent: true}, &reply)
		assert.NoError(t, err)
		userAccount, err := userManager.Read("123")
		assert.NoError(t, err)
		assert.True(t, userAccount.AppMetadata["newsletter"].(bool))
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
		reqContext = context.WithValue(reqContext, jsonrpc.Keys.UserKey, &jwt.Token{
			Claims: jwt.MapClaims{
				"email_verified": true,
			},
		})
		req = req.WithContext(reqContext)
		var reply jsonrpc.VerifyEmailReply
		err := svc.ResendVerificationEmail(req, &jsonrpc.VerifyEmailArgs{}, &reply)
		assert.Error(t, err)
	})

	t.Run("failure - no ID", func(t *testing.T) {
		var reply jsonrpc.VerifyEmailReply
		err := svc.ResendVerificationEmail(req, &jsonrpc.VerifyEmailArgs{}, &reply)
		assert.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, jsonrpc.Keys.UserKey, &jwt.Token{
			Claims: jwt.MapClaims{
				"email_verified": false,
			},
		})
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
		reqContext = context.WithValue(reqContext, jsonrpc.Keys.RolesKey, []string{
			"Owner",
		})
		req = req.WithContext(reqContext)
		var reply jsonrpc.UpdateDomainsReply
		err := svc.UpdateAutoSignupDomains(req, &jsonrpc.UpdateDomainsArgs{}, &reply)
		assert.Error(t, err)
	})

	t.Run("failure - no company code", func(t *testing.T) {
		reqContext := req.Context()
		reqContext = context.WithValue(reqContext, jsonrpc.Keys.CompanyKey, "test")
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
