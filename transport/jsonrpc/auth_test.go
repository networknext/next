package jsonrpc_test

// All tests listed below depend on test@networknext.com being a user in auth0
/* func TestAuthMiddleware(t *testing.T) {
	// JWT obtained from Portal Login Dev SPA (Auth0)
	// Note: 5 year expiration time (expires on 18 May 2025)
	// test@networknext.com => Delete this in auth0 and these tests will break
	jwtSideload := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6Ik5rWXpOekkwTkVVNVFrSTVNRVF4TURRMk5UWTBNakkzTmpOQlJESkVNa1E0TnpGRFF6QkdRdyJ9.eyJuaWNrbmFtZSI6InRlc3QiLCJuYW1lIjoidGVzdEBuZXR3b3JrbmV4dC5jb20iLCJwaWN0dXJlIjoiaHR0cHM6Ly9zLmdyYXZhdGFyLmNvbS9hdmF0YXIvMmRhNWMwMjU5ZTQ3NmI1MDg0MTBlZWY3ZjI5Zjc1NGE_cz00ODAmcj1wZyZkPWh0dHBzJTNBJTJGJTJGY2RuLmF1dGgwLmNvbSUyRmF2YXRhcnMlMkZ0ZS5wbmciLCJ1cGRhdGVkX2F0IjoiMjAyMC0wNi0yM1QxMzozOToyMS44ODFaIiwiZW1haWwiOiJ0ZXN0QG5ldHdvcmtuZXh0LmNvbSIsImVtYWlsX3ZlcmlmaWVkIjp0cnVlLCJpc3MiOiJodHRwczovL25ldHdvcmtuZXh0LmF1dGgwLmNvbS8iLCJzdWIiOiJhdXRoMHw1Yjk2ZjYxY2YxNjQyNzIxYWQ4NGVlYjYiLCJhdWQiOiJvUUpIM1lQSGR2WkpueENQbzFJcnR6NVVLaTV6cnI2biIsImlhdCI6MTU5MjkxOTU2NSwiZXhwIjoxNzUwNzA0MzI1LCJub25jZSI6ImRHZFNUWEpRTnpkdE5GcHNjR0Z1YVg1dlQxVlNhVFZXUjJoK2VHdG1hMnB2TkcweFZuNTFZalJJZmc9PSJ9.BvMe5fWJcheGzKmt3nCIeLjMD-C5426cpjtJiR55i7lmbT0k4h8Z2X6rynZ_aKR-gaCTY7FG5gI-Ty9ZY1zboWcIkxaTi0VKQzdMUTYVMXVEK2cQ1NVbph7_RSJhLfgO5y7PkmuMZXJEFdrI_2PkO4b3tOU-vpUHFUPtTsESV79a81kXn2C5j_KkKzCOPZ4zol1aEU3WliaaJNT38iSz3NX9URshrrdCE39JRClx6wbUgrfCGnVtfens-Sg7atijivaOx8IlUGOxLMEciYwBL2aY5EXaa7tp7c8ZvoEEj7uZH2R35fV7eUzACwShU-JLR9oOsNEhS4XO1AzTMtNHQA"

	noopHandler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	t.Run("skip auth", func(t *testing.T) {
		authMiddleware := jsonrpc.AuthMiddleware("", http.HandlerFunc(noopHandler))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		res := httptest.NewRecorder()

		authMiddleware.ServeHTTP(res, req)
		assert.Equal(t, http.StatusOK, res.Code)
	})

	t.Run("check auth claims", func(t *testing.T) {
		authMiddleware := jsonrpc.AuthMiddleware("oQJH3YPHdvZJnxCPo1Irtz5UKi5zrr6n", http.HandlerFunc(noopHandler))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Add("Authorization", "Bearer "+jwtSideload)
		res := httptest.NewRecorder()

		authMiddleware.ServeHTTP(res, req)
		assert.Equal(t, http.StatusOK, res.Code)
	})

	t.Run("anonymous auth", func(t *testing.T) {
		authMiddleware := jsonrpc.AuthMiddleware("oQJH3YPHdvZJnxCPo1Irtz5UKi5zrr6n", http.HandlerFunc(noopHandler))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		res := httptest.NewRecorder()

		authMiddleware.ServeHTTP(res, req)
		assert.Equal(t, http.StatusOK, res.Code)
	})
}

func TestAuthClient(t *testing.T) {
	t.Skip()
	// test@networknext.com => Delete this in auth0 and these tests will break
	jwtSideload := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6Ik5rWXpOekkwTkVVNVFrSTVNRVF4TURRMk5UWTBNakkzTmpOQlJESkVNa1E0TnpGRFF6QkdRdyJ9.eyJuaWNrbmFtZSI6InRlc3QiLCJuYW1lIjoidGVzdEBuZXR3b3JrbmV4dC5jb20iLCJwaWN0dXJlIjoiaHR0cHM6Ly9zLmdyYXZhdGFyLmNvbS9hdmF0YXIvMmRhNWMwMjU5ZTQ3NmI1MDg0MTBlZWY3ZjI5Zjc1NGE_cz00ODAmcj1wZyZkPWh0dHBzJTNBJTJGJTJGY2RuLmF1dGgwLmNvbSUyRmF2YXRhcnMlMkZ0ZS5wbmciLCJ1cGRhdGVkX2F0IjoiMjAyMC0wNi0yM1QxMzozOToyMS44ODFaIiwiZW1haWwiOiJ0ZXN0QG5ldHdvcmtuZXh0LmNvbSIsImVtYWlsX3ZlcmlmaWVkIjp0cnVlLCJpc3MiOiJodHRwczovL25ldHdvcmtuZXh0LmF1dGgwLmNvbS8iLCJzdWIiOiJhdXRoMHw1Yjk2ZjYxY2YxNjQyNzIxYWQ4NGVlYjYiLCJhdWQiOiJvUUpIM1lQSGR2WkpueENQbzFJcnR6NVVLaTV6cnI2biIsImlhdCI6MTU5MjkxOTU2NSwiZXhwIjoxNzUwNzA0MzI1LCJub25jZSI6ImRHZFNUWEpRTnpkdE5GcHNjR0Z1YVg1dlQxVlNhVFZXUjJoK2VHdG1hMnB2TkcweFZuNTFZalJJZmc9PSJ9.BvMe5fWJcheGzKmt3nCIeLjMD-C5426cpjtJiR55i7lmbT0k4h8Z2X6rynZ_aKR-gaCTY7FG5gI-Ty9ZY1zboWcIkxaTi0VKQzdMUTYVMXVEK2cQ1NVbph7_RSJhLfgO5y7PkmuMZXJEFdrI_2PkO4b3tOU-vpUHFUPtTsESV79a81kXn2C5j_KkKzCOPZ4zol1aEU3WliaaJNT38iSz3NX9URshrrdCE39JRClx6wbUgrfCGnVtfens-Sg7atijivaOx8IlUGOxLMEciYwBL2aY5EXaa7tp7c8ZvoEEj7uZH2R35fV7eUzACwShU-JLR9oOsNEhS4XO1AzTMtNHQA"
	noopHandler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
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
		Logger:  logger,
	}
	authMiddleware := jsonrpc.AuthMiddleware("oQJH3YPHdvZJnxCPo1Irtz5UKi5zrr6n", http.HandlerFunc(noopHandler))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Add("Authorization", "Bearer "+jwtSideload)
	res := httptest.NewRecorder()

	authMiddleware.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	user := req.Context().Value("user")
	assert.NotEqual(t, user, nil)
	claims := user.(*jwt.Token).Claims.(jwt.MapClaims)

	requestID, ok := claims["sub"]

	assert.True(t, ok)
	assert.Equal(t, "auth0|5b96f61cf1642721ad84eeb6", requestID)

	roles, err := auth0Client.Manager.User.Roles(requestID.(string))

	assert.NoError(t, err)
	req = jsonrpc.SetRoles(req, *roles)

	t.Run("fetch all auth0 accounts", func(t *testing.T) {
		var reply jsonrpc.AccountsReply

		err = svc.AllAccounts(req, &jsonrpc.AccountsArgs{}, &reply)
		assert.NoError(t, err)
	})

	t.Run("fetch user no user id", func(t *testing.T) {
		var reply jsonrpc.AccountReply

		err := svc.UserAccount(req, &jsonrpc.AccountArgs{}, &reply)
		assert.Error(t, err)
	})

	t.Run("fetch user auth0 account", func(t *testing.T) {
		var reply jsonrpc.AccountReply

		err := svc.UserAccount(req, &jsonrpc.AccountArgs{UserID: "auth0|5b96f61cf1642721ad84eeb6"}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, reply.UserAccount.Name, "test@networknext.com")
		assert.Equal(t, reply.UserAccount.Email, "test@networknext.com")
		assert.Equal(t, reply.UserAccount.UserID, "5b96f61cf1642721ad84eeb6")
		assert.Equal(t, reply.UserAccount.ID, "111")
		assert.Equal(t, reply.UserAccount.CompanyName, "")
	})

	t.Run("fetch user roles no user id", func(t *testing.T) {
		var reply jsonrpc.RolesReply

		err := svc.UserRoles(req, &jsonrpc.RolesArgs{}, &reply)
		assert.Error(t, err)
	})

	t.Run("fetch all auth0 roles", func(t *testing.T) {
		var reply jsonrpc.RolesReply

		err := svc.AllRoles(req, &jsonrpc.RolesArgs{}, &reply)
		assert.NoError(t, err)

		assert.NotEqual(t, len(reply.Roles), 0)
	})

	t.Run("Remove all auth0 roles", func(t *testing.T) {
		var reply jsonrpc.RolesReply

		id := "rol_YfFrtom32or4vH89"
		name := "Admin"
		description := "Can manage the Network Next system, including access to configstore."

		// Need to keep Admin role as a minimum to not break further tests
		roles := []*management.Role{
			{
				ID:          &id,
				Name:        &name,
				Description: &description,
			},
		}

		err := svc.UpdateUserRoles(req, &jsonrpc.RolesArgs{UserID: "auth0|5b96f61cf1642721ad84eeb6", Roles: roles}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, len(reply.Roles), 1)
	})

	t.Run("add all auth0 roles", func(t *testing.T) {
		var reply jsonrpc.RolesReply

		roleNames := []string{
			"rol_8r0281hf2oC4cvrD",
			"rol_ScQpWhLvmTKRlqLU",
		}
		roleTypes := []string{
			"Owner",
			"Viewer",
		}
		roleDescriptions := []string{
			"Can access and manage everything in an account.",
			"Can see current sessions and the map.",
		}

		// Need to keep Admin role as a minimum to not break further tests
		roles := []*management.Role{
			{
				ID:          &roleNames[0],
				Name:        &roleTypes[0],
				Description: &roleDescriptions[0],
			},
			{
				ID:          &roleNames[1],
				Name:        &roleTypes[1],
				Description: &roleDescriptions[1],
			},
		}

		err := svc.UpdateUserRoles(req, &jsonrpc.RolesArgs{UserID: "auth0|5b96f61cf1642721ad84eeb6", Roles: roles}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, 2, len(reply.Roles))
	})
}
*/
