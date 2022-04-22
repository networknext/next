package jsonrpc

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/storage"
	"github.com/networknext/backend/modules/transport/looker"
	"github.com/networknext/backend/modules/transport/middleware"
	"github.com/networknext/backend/modules/transport/notifications"
	"gopkg.in/auth0.v4/management"
)

const (
	MAX_USER_LOOKUP_PAGES = 10
)

type AuthService struct {
	AuthenticationClient *notifications.Auth0AuthClient
	HubSpotClient        *notifications.HubSpotClient
	mu                   sync.Mutex
	RoleCache            map[string]*management.Role
	MailChimpManager     notifications.MailChimpHandler
	JobManager           storage.JobManager
	RoleManager          storage.RoleManager
	UserManager          storage.UserManager
	SlackClient          notifications.SlackClient
	Storage              storage.Storer
	LookerClient         *looker.LookerClient
}

type AccountsArgs struct {
	Emails []string           `json:"emails"`
	Roles  []*management.Role `json:"roles"`
}

type AccountsReply struct {
	UserAccounts []account `json:"accounts"`
}

type AccountArgs struct {
	UserID string `json:"user_id"`
}

type AccountReply struct {
	UserAccount account  `json:"account"`
	Domains     []string `json:"domains"`
	LookerURL   string   `json:"looker_url"`
}

type account struct {
	Avatar      string             `json:"avatar"`
	UserID      string             `json:"user_id"`
	BuyerID     string             `json:"buyer_id"`
	Seller      bool               `json:"seller"`
	CompanyName string             `json:"company_name"`
	CompanyCode string             `json:"company_code"`
	FirstName   string             `json:"first_name"`
	LastName    string             `json:"last_name"`
	Email       string             `json:"email"`
	Roles       []*management.Role `json:"roles"`
	Analytics   bool               `json:"analytics"`
	Billing     bool               `json:"billing"`
	Trial       bool               `json:"trial"`
	Verified    bool               `json:"verified"`
}

func (s *AuthService) AllAccounts(r *http.Request, args *AccountsArgs, reply *AccountsReply) error {
	ctx := r.Context()
	reply.UserAccounts = make([]account, 0)

	isAdmin := middleware.VerifyAllRoles(r, middleware.AdminRole)

	if !isAdmin && !middleware.VerifyAllRoles(r, middleware.OwnerRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("AllAccounts(): %v", err.Error())
		return &err
	}

	requestCustomerCode := middleware.RequestUserCustomerCode(ctx)
	if requestCustomerCode == "" {
		err := JSONRPCErrorCodes[int(ERROR_USER_IS_NOT_ASSIGNED)]
		core.Error("AllAccounts(): %v", err.Error())
		return &err
	}

	customer, err := s.Storage.Customer(ctx, requestCustomerCode)
	if err != nil {
		core.Error("AllAccounts(): %v", err.Error())
		err := JSONRPCErrorCodes[int(ERROR_STORAGE_FAILURE)]
		return &err
	}

	// We don't care about the error here due to buyer and seller accounts not being guaranteed
	buyer, _ := s.Storage.BuyerWithCompanyCode(ctx, requestCustomerCode)
	seller, _ := s.Storage.SellerWithCompanyCode(ctx, requestCustomerCode)

	totalUsers, err := s.FetchAllAccountsFromAuth0()
	if err != nil {
		core.Error("AllAccounts(): %v", err.Error())
		err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
		return &err
	}

	// Find all users associated with the customer account
	for _, a := range totalUsers {
		customerCode, ok := a.AppMetadata["company_code"].(string)
		if !ok || requestCustomerCode != customerCode {
			continue
		}
		userRoles, err := s.UserManager.Roles(*a.ID)
		if err != nil {
			core.Error("AllAccounts(): %v: Failed to get user roles", err.Error())
			err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
			return &err
		}

		reply.UserAccounts = append(reply.UserAccounts, newAccount(a, userRoles.Roles, buyer, customer.Name, customer.Code, seller.Name != "", isAdmin))
	}

	return nil
}

func (s *AuthService) FetchAllAccountsFromAuth0() ([]*management.User, error) {
	var totalUsers []*management.User
	keepSearching := true
	page := 0

	for keepSearching && page < MAX_USER_LOOKUP_PAGES { // MAX_USER_LOOKUP_PAGES is a kill switch for unforseen infinite loops
		accountList, err := s.UserManager.List(management.PerPage(100), management.Page(page))
		if err != nil {
			core.Error("AllAccounts(): %v: Failed to fetch user list", err.Error())
			err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
			return totalUsers, &err
		}

		totalUsers = append(totalUsers, accountList.Users...)
		if len(totalUsers)%100 != 0 {
			keepSearching = false
		}
		page = page + 1
	}

	return totalUsers, nil
}

func (s *AuthService) UserAccount(r *http.Request, args *AccountArgs, reply *AccountReply) error {
	isAdmin := middleware.VerifyAllRoles(r, middleware.AdminRole)

	if args.UserID == "" {
		err := JSONRPCErrorCodes[int(ERROR_MISSING_FIELD)]
		err.Data.(*JSONRPCErrorData).MissingField = "UserID"
		core.Error("UserAccount(): %v: UserID is required", err.Error())
		return &err
	}

	user := r.Context().Value(middleware.Keys.UserKey)
	if user == nil {
		err := JSONRPCErrorCodes[int(ERROR_JWT_PARSE_FAILURE)]
		core.Error("UserAccount(): %v", err.Error())
		return &err
	}

	claims := user.(*jwt.Token).Claims.(jwt.MapClaims)
	requestID, ok := claims["sub"].(string)
	if !ok {
		err := JSONRPCErrorCodes[int(ERROR_JWT_PARSE_FAILURE)]
		core.Error("UserAccount(): %v: Failed to parse user ID", err.Error())
		return &err
	}
	if requestID != args.UserID && !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OwnerRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("UserAccount(): %v", err.Error())
		return &err
	}

	userAccount, err := s.UserManager.Read(args.UserID)
	if err != nil {
		core.Error("UserAccount(): %v: Failed to get user account details", err.Error())
		err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
		return &err
	}
	companyCode, ok := userAccount.AppMetadata["company_code"].(string)
	if !ok {
		companyCode = ""
	}
	var company routing.Customer
	if companyCode != "" {
		company, err = s.Storage.Customer(r.Context(), companyCode)
		if err != nil {
			core.Error("UserAccount(): %v: Could not find customer account for customer code: %v", err.Error(), companyCode)
			err := JSONRPCErrorCodes[int(ERROR_STORAGE_FAILURE)]
			return &err
		}
	}
	buyer, _ := s.Storage.BuyerWithCompanyCode(r.Context(), companyCode)
	seller, _ := s.Storage.SellerWithCompanyCode(r.Context(), companyCode)

	userRoles, err := s.UserManager.Roles(*userAccount.ID)
	if err != nil {
		core.Error("UserAccount(): %v: Failed to get user account roles", err.Error())
		err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
		return &err
	}

	if middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OwnerRole) && requestID == args.UserID {
		reply.Domains = strings.Split(company.AutomaticSignInDomains, ",")
	}

	reply.UserAccount = newAccount(userAccount, userRoles.Roles, buyer, company.Name, company.Code, seller.Name != "", isAdmin)

	return nil
}

func (s *AuthService) DeleteUserAccount(r *http.Request, args *AccountArgs, reply *AccountReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OwnerRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("DeleteUserAccount(): %v", err.Error())
		return &err
	}

	if args.UserID == "" {
		err := JSONRPCErrorCodes[int(ERROR_MISSING_FIELD)]
		err.Data.(*JSONRPCErrorData).MissingField = "UserID"
		core.Error("DeleteUserAccount(): %v", err.Error())
		return &err
	}
	user, err := s.UserManager.Read(args.UserID)
	if err != nil {
		core.Error("DeleteUserAccount(): %v: Failed to read user account", err.Error())
		err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
		return &err
	}

	userCompanyCode, ok := user.AppMetadata["company_code"].(string)
	if !ok || userCompanyCode == "" {
		return nil
	}

	// Non admin trying to delete user from another company
	requestCompanyCode, ok := r.Context().Value(middleware.Keys.CustomerKey).(string)
	if (!ok || requestCompanyCode != userCompanyCode) && !middleware.VerifyAllRoles(r, middleware.AdminRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("DeleteUserAccount(): %v", err.Error())
		return &err
	}

	if err := s.UserManager.Update(args.UserID, &management.User{
		AppMetadata: map[string]interface{}{
			"company_code": "",
		},
	}); err != nil {
		core.Error("DeleteUserAccount(): %v: Failed to update deleted user company code", err.Error())
		err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
		return &err
	}

	allRoles := s.RoleCacheToArray(true)

	err = s.UserManager.RemoveRoles(args.UserID, allRoles...)
	if err != nil {
		core.Error("UpdateUserRoles(): %v: Failed to remove old user roles", err.Error())
		err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
		return &err
	}

	return nil
}

func (s *AuthService) AddUserAccount(r *http.Request, args *AccountsArgs, reply *AccountsReply) error {
	var adminString string = "Admin"
	var accounts []account

	isAdmin := middleware.VerifyAllRoles(r, middleware.AdminRole)

	if !isAdmin && !middleware.VerifyAllRoles(r, middleware.OwnerRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("AddUserAccount(): %v", err.Error())
		return &err
	}

	// Check if non admin is assigning admin role
	for _, role := range args.Roles {
		if role.Name == &adminString && !middleware.VerifyAllRoles(r, middleware.AdminRole) {
			err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
			core.Error("AddUserAccount(): %v", err.Error())
			return &err
		}
	}

	// Gather request user information
	userCompanyCode, ok := r.Context().Value(middleware.Keys.CustomerKey).(string)
	if !ok || userCompanyCode == "" {
		err := JSONRPCErrorCodes[int(ERROR_USER_IS_NOT_ASSIGNED)]
		core.Error("AddUserAccount(): %v", err.Error())
		return &err
	}

	connectionID := "Username-Password-Authentication"
	emails := args.Emails
	falseValue := false

	buyer, _ := s.Storage.BuyerWithCompanyCode(r.Context(), userCompanyCode)
	seller, _ := s.Storage.SellerWithCompanyCode(r.Context(), userCompanyCode)

	registered := make(map[string]*management.User)

	totalUsers, err := s.FetchAllAccountsFromAuth0()
	if err != nil {
		err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
		core.Error("AddUserAccount(): %v", err.Error())
		return &err
	}

	emailString := strings.Join(emails, ",")

	for _, a := range totalUsers {
		if strings.Contains(emailString, *a.Email) {
			registered[*a.Email] = a
		}
	}

	// Create an account for each new email
	var newUser *management.User
	for _, e := range emails {
		user, ok := registered[e]
		if !ok {
			pw, err := GenerateRandomString(32)
			if err != nil {
				err := JSONRPCErrorCodes[int(ERROR_PASSWORD_GENERATION_FAILURE)]
				core.Error("AddUserAccount(): %v", err.Error())
				return &err
			}
			newUser = &management.User{
				Connection:    &connectionID,
				Email:         &e,
				Name:          &e,
				EmailVerified: &falseValue,
				VerifyEmail:   &falseValue,
				Password:      &pw,
				AppMetadata: map[string]interface{}{
					"company_code": userCompanyCode,
				},
			}
			if err = s.UserManager.Create(newUser); err != nil {
				core.Error("AddUserAccount(): %v: Failed to create new user account", err.Error())
				err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
				return &err
			}
			if len(args.Roles) > 0 {
				if err = s.UserManager.AssignRoles(*newUser.ID, args.Roles...); err != nil {
					core.Error("AddUserAccount(): %v: Failed to assign new user roles", err.Error())
					err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
					return &err
				}
			}
		} else {
			newUser = &management.User{
				GivenName:  user.GivenName,
				FamilyName: user.FamilyName,
				AppMetadata: map[string]interface{}{
					"company_code": userCompanyCode,
				},
			}
			if err := s.UserManager.Update(user.GetID(), newUser); err != nil {
				core.Error("AddUserAccount(): %v: Failed to update user account", err.Error())
				err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
				return &err
			}

			allRoles := []*management.Role{}

			// Convert map to array
			for _, role := range s.RoleCache {
				allRoles = append(allRoles, role)
			}

			if err := s.UserManager.RemoveRoles(user.GetID(), allRoles...); err != nil {
				core.Error("AddUserAccount(): %v: Failed to remove exist roles from user account", err.Error())
				err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
				return &err
			}
			if len(args.Roles) > 0 {
				if err := s.UserManager.AssignRoles(user.GetID(), args.Roles...); err != nil {
					core.Error("AddUserAccount(): %v: Failed to assign new roles to user account", err.Error())
					err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
					return &err
				}
			}
		}

		company, err := s.Storage.Customer(r.Context(), userCompanyCode)
		if err != nil {
			core.Error("AddUserAccount(): %v", err.Error())
			err := JSONRPCErrorCodes[int(ERROR_STORAGE_FAILURE)]
			return &err
		}
		accounts = append(accounts, newAccount(newUser, args.Roles, buyer, company.Name, company.Code, seller.Name != "", isAdmin))
	}
	reply.UserAccounts = accounts
	return nil
}

func GenerateRandomString(n int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"
	bytes, err := GenerateRandomBytes(n)
	if err != nil {
		// not method, no service logger
		return "", err
	}
	for i, b := range bytes {
		bytes[i] = letters[b%byte(len(letters))]
	}
	return string(bytes), nil
}

func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		// not method, no service logger
		return nil, err
	}

	return b, nil
}

func newAccount(u *management.User, r []*management.Role, buyer routing.Buyer, companyName string, companyCode string, isSeller bool, isAdmin bool) account {
	buyerID := ""
	if buyer.ID != 0 {
		buyerID = fmt.Sprintf("%016x", buyer.ID)
	}

	roles := make([]*management.Role, 0)

	if !isAdmin {
		for _, role := range r {
			if role.GetName() != "Admin" {
				roles = append(roles, role)
			}
		}
	} else {
		roles = r
	}

	account := account{
		Avatar:      u.GetPicture(),
		UserID:      *u.Identities[0].UserID,
		BuyerID:     buyerID,
		Seller:      isSeller,
		CompanyCode: companyCode,
		CompanyName: companyName,
		FirstName:   u.GetGivenName(),
		LastName:    u.GetFamilyName(),
		Email:       u.GetEmail(),
		Roles:       roles,
		Analytics:   buyer.Analytics,
		Billing:     buyer.Billing,
		Trial:       buyer.Trial,
		Verified:    u.GetEmailVerified(),
	}

	return account
}

type databaseEntry struct {
	FirstName    string
	LastName     string
	Email        string
	CompanyCode  string
	BuyerID      string
	IsOwner      bool
	CreationTime string
}

type UserDatabaseArgs struct{}

type UserDatabaseReply struct {
	Entries []databaseEntry
}

func (s *AuthService) UserDatabase(r *http.Request, args *UserDatabaseArgs, reply *UserDatabaseReply) error {
	reply.Entries = make([]databaseEntry, 0)

	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OpsRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("UserDatabase(): %v", err.Error())
		return &err
	}

	totalUsers, err := s.FetchAllAccountsFromAuth0()
	if err != nil {
		err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
		core.Error("UserDatabase(): %v", err.Error())
		return &err
	}

	for _, account := range totalUsers {
		var companyCode string
		companyCode, ok := account.AppMetadata["company_code"].(string)
		if !ok || (companyCode == "" || companyCode == "next") {
			continue
		}

		userRoles, err := s.UserManager.Roles(account.GetID())
		if err != nil {
			core.Error("UserDatabase(): %v: Failed to get user roles", err.Error())
			err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
			return &err
		}

		isOwner := false
		for _, role := range userRoles.Roles {
			// Check if the role is an owner role
			if role.GetID() == s.RoleCache["Owner"].GetID() {
				isOwner = true
				break
			}
		}

		entry := databaseEntry{
			FirstName:    account.GetGivenName(),
			LastName:     account.GetFamilyName(),
			Email:        account.GetEmail(),
			CompanyCode:  companyCode,
			IsOwner:      isOwner,
			CreationTime: account.CreatedAt.String(),
		}

		buyer, _ := s.Storage.BuyerWithCompanyCode(r.Context(), companyCode)

		if buyer.ID != 0 {
			entry.BuyerID = fmt.Sprintf("%016x", buyer.ID)
		}

		reply.Entries = append(reply.Entries, entry)
	}

	return nil
}

type RolesArgs struct {
	UserID string             `json:"user_id"`
	Roles  []*management.Role `json:"roles"`
}

type RolesReply struct {
	Roles []*management.Role `json:"roles"`
}

func (s *AuthService) AllRoles(r *http.Request, args *RolesArgs, reply *RolesReply) error {
	reply.Roles = make([]*management.Role, 0)
	isAdmin := middleware.VerifyAllRoles(r, middleware.AdminRole)

	if !isAdmin && !middleware.VerifyAllRoles(r, middleware.OwnerRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("AllRoles(): %v", err.Error())
		return &err
	}

	if isAdmin {
		reply.Roles = append(reply.Roles, s.RoleCache["Admin"])
	}

	requestCustomerCode, ok := r.Context().Value(middleware.Keys.CustomerKey).(string)
	if !ok {
		err := JSONRPCErrorCodes[int(ERROR_USER_IS_NOT_ASSIGNED)]
		core.Error("AllRoles(): %v", err.Error())
		return &err
	}

	buyer, err := s.Storage.BuyerWithCompanyCode(r.Context(), requestCustomerCode)
	if err != nil {
		// Buyer account doesn't exist - this could be due to the customer not entering a public key yet so return Owner role only
		reply.Roles = append(reply.Roles, s.RoleCache["Owner"])
		return nil
	}

	// If valid buyer account, grab all users to determine Looker usage
	totalUsers, err := s.FetchAllAccountsFromAuth0()
	if err != nil {
		core.Error("AllRoles(): %v", err.Error())
		err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
		return &err
	}

	seatsTaken := int64(0)
	for _, a := range totalUsers {
		userCustomerCode, ok := a.AppMetadata["company_code"].(string)
		if !ok || requestCustomerCode != userCustomerCode {
			continue
		}
		userRoles, err := s.UserManager.Roles(*a.ID)
		if err != nil {
			core.Error("AllRoles(): %v: Failed to get user roles", err.Error())
			err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
			return &err
		}

		for _, role := range userRoles.Roles {
			if role.GetName() == s.RoleCache["Explorer"].GetName() {
				seatsTaken = seatsTaken + 1
			}
		}
	}

	for name, role := range s.RoleCache {
		// Skip the admin role, it was taken care of earlier, and skip Explorer if all seats have been used
		if name == "Admin" || (!isAdmin && name == "Explorer" && seatsTaken >= buyer.LookerSeats) {
			continue
		}
		reply.Roles = append(reply.Roles, role)
	}

	sort.Slice(reply.Roles, func(i, j int) bool {
		return reply.Roles[i].GetName() < reply.Roles[j].GetName()
	})
	return nil
}

func (s *AuthService) UserRoles(r *http.Request, args *RolesArgs, reply *RolesReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OwnerRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("UserRoles(): %v", err.Error())
		return &err
	}

	if args.UserID == "" {
		err := JSONRPCErrorCodes[int(ERROR_MISSING_FIELD)]
		err.Data.(*JSONRPCErrorData).MissingField = "UserID"
		core.Error("UserRoles(): %v", err.Error())
		return &err
	}

	userRoles, err := s.UserManager.Roles(args.UserID)
	if err != nil {
		core.Error("UserRoles(): %v: Failed to fetch user roles", err.Error())
		err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
		return &err
	}

	reply.Roles = userRoles.Roles

	return nil
}

func (s *AuthService) UpdateUserRoles(r *http.Request, args *RolesArgs, reply *RolesReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OwnerRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("UpdateUserRoles(): %v", err.Error())
		return &err
	}

	if args.UserID == "" {
		err := JSONRPCErrorCodes[int(ERROR_MISSING_FIELD)]
		err.Data.(*JSONRPCErrorData).MissingField = "UserID"
		core.Error("UpdateUserRoles(): %v: missing UserID", err.Error())
		return &err
	}

	ctx := r.Context()

	allRoles := s.RoleCacheToArray(true)

	err := s.UserManager.RemoveRoles(args.UserID, allRoles...)
	if err != nil {
		core.Error("UpdateUserRoles(): %v: Failed to remove old user roles", err.Error())
		err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
		return &err
	}

	if len(args.Roles) == 0 {
		reply.Roles = make([]*management.Role, 0)
		return nil
	}

	allowedLookerSeats := 0

	requestCustomerCode, ok := ctx.Value(middleware.Keys.CustomerKey).(string)
	if ok {
		buyer, err := s.Storage.BuyerWithCompanyCode(ctx, requestCustomerCode)
		if err == nil {
			allowedLookerSeats = int(buyer.LookerSeats)
		}
	}

	seatsTaken := 0
	if allowedLookerSeats > 0 {
		// If valid buyer account, grab all users to determine Looker usage
		totalUsers, err := s.FetchAllAccountsFromAuth0()
		if err != nil {
			core.Error("UpdateUserRoles(): %v", err.Error())
			err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
			return &err
		}

		for _, a := range totalUsers {
			userCustomerCode, ok := a.AppMetadata["company_code"].(string)
			if !ok || requestCustomerCode != userCustomerCode {
				continue
			}
			userRoles, err := s.UserManager.Roles(*a.ID)
			if err != nil {
				core.Error("UpdateUserRoles(): %v: Failed to get user roles", err.Error())
				err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
				return &err
			}

			for _, role := range userRoles.Roles {
				if role.GetName() == s.RoleCache["Explorer"].GetName() {
					seatsTaken = seatsTaken + 1
				}
			}
		}
	}

	allowedRoles := make([]*management.Role, 0)

	// Make sure someone who isn't admin isn't assigning admin
	for _, role := range args.Roles {
		if role.GetName() == s.RoleCache["Admin"].GetName() && !middleware.VerifyAllRoles(r, middleware.AdminRole) {
			err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
			core.Error("UpdateUserRoles(): %v", err.Error())
			return &err
		}

		if role.GetName() == s.RoleCache["Explorer"].GetName() && seatsTaken >= allowedLookerSeats {
			continue
		}

		allowedRoles = append(allowedRoles, s.RoleCache[role.GetName()])
	}

	if len(allowedRoles) == 0 {
		reply.Roles = allowedRoles
		return nil
	}

	err = s.UserManager.AssignRoles(args.UserID, allowedRoles...)
	if err != nil {
		core.Error("UpdateUserRoles(): %v: Failed to assign user roles", err.Error())
		err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
		return &err
	}

	reply.Roles = allowedRoles

	sort.Slice(reply.Roles, func(i, j int) bool {
		return reply.Roles[i].GetName() < reply.Roles[j].GetName()
	})
	return nil
}

type SetupCompanyAccountArgs struct {
	CompanyName string `json:"company_name"`
	CompanyCode string `json:"company_code"`
	Email       string `json:"email"`
}

type SetupCompanyAccountReply struct{}

func (s *AuthService) SetupCompanyAccount(r *http.Request, args *SetupCompanyAccountArgs, reply *SetupCompanyAccountReply) error {
	if middleware.VerifyAnyRole(r, middleware.AnonymousRole, middleware.UnverifiedRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("SetupCompanyAccount(): %v", err.Error())
		return &err
	}

	if args.CompanyCode == "" {
		err := JSONRPCErrorCodes[int(ERROR_MISSING_FIELD)]
		err.Data.(*JSONRPCErrorData).MissingField = "CompanyCode"
		core.Error("SetupCompanyAccount(): %v: missing CompanyCode", err.Error())
		return &err
	}

	if args.Email == "" {
		err := JSONRPCErrorCodes[int(ERROR_MISSING_FIELD)]
		err.Data.(*JSONRPCErrorData).MissingField = "Email"
		core.Error("SetupCompanyAccount(): %v: missing Email", err.Error())
		return &err
	}

	assignedCompanyCode, ok := r.Context().Value(middleware.Keys.CustomerKey).(string)
	if ok && assignedCompanyCode != "" {
		err := JSONRPCErrorCodes[int(ERROR_ILLEGAL_OPERATION)]
		core.Error("SetupCompanyAccount(): %v: User is already assigned to a company. Please reach out to support for further assistance.", err.Error())
		return &err
	}

	// grab request user information
	requestUser := r.Context().Value(middleware.Keys.UserKey)
	if requestUser == nil {
		err := JSONRPCErrorCodes[int(ERROR_JWT_PARSE_FAILURE)]
		core.Error("SetupCompanyAccount(): %v", err.Error())
		return &err
	}

	// get request user ID for role assignment
	requestID, ok := requestUser.(*jwt.Token).Claims.(jwt.MapClaims)["sub"].(string)
	if !ok {
		err := JSONRPCErrorCodes[int(ERROR_JWT_PARSE_FAILURE)]
		core.Error("SetupCompanyAccount(): %v: Failed to parse user ID", err.Error())
		return &err
	}

	ctx := r.Context()
	roles := []*management.Role{}

	// Check if customer account exists already
	customer, err := s.Storage.Customer(ctx, args.CompanyCode)
	if err == nil {
		// exists, Check if the users domain matches the automatic signup domains
		domain := strings.Split(args.Email, "@")
		if !strings.Contains(customer.AutomaticSignInDomains, domain[1]) {
			err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
			core.Error("SetupCompanyAccount(): %v: User's email domain is not listed in accepted domains", err.Error())
			return &err
		}
	} else {
		// Add the new customer account
		if args.CompanyName == "" {
			err := JSONRPCErrorCodes[int(ERROR_MISSING_FIELD)]
			err.Data.(*JSONRPCErrorData).MissingField = "CompanyName"
			core.Error("SetupCompanyAccount(): %v: Missing company name field", err.Error())
			return &err
		}

		if err := s.Storage.AddCustomer(ctx, routing.Customer{
			Code: args.CompanyCode,
			Name: args.CompanyName,
		}); err != nil {
			core.Error("SetupCompanyAccount(): %v: Failed to add new customer entry", err.Error())
			err := JSONRPCErrorCodes[int(ERROR_STORAGE_FAILURE)]
			return &err
		}

		roles = append(roles, s.RoleCache["Owner"])
	}

	// Assign the customer code to the user token
	if err := s.UserManager.Update(requestID, &management.User{
		AppMetadata: map[string]interface{}{
			"company_code": args.CompanyCode,
		},
	}); err != nil {
		core.Error("SetupCompanyAccount() failed to update user company code: %v", err)
		err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
		return &err
	}

	if !middleware.VerifyAllRoles(r, middleware.AdminRole) {
		allRoles := []*management.Role{}

		// Convert map to array
		for _, role := range s.RoleCache {
			allRoles = append(allRoles, role)
		}

		// Remove existing roles if there are any
		if err := s.UserManager.RemoveRoles(requestID, allRoles...); err != nil {
			core.Error("SetupCompanyAccount() failed to remove roles: %v", err)
			err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
			return &err
		}

		// Assign new roles (owner)
		if len(roles) > 0 {
			if err := s.UserManager.AssignRoles(requestID, roles...); err != nil {
				core.Error("SetupCompanyAccount() failed to assign user roles: %v", err)
				err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
				return &err
			}
		}
	}
	return nil
}

type UpdateAccountDetailsArgs struct {
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Newsletter bool   `json:"newsletter"`
}

type UpdateAccountDetailsReply struct{}

func (s *AuthService) UpdateAccountDetails(r *http.Request, args *UpdateAccountDetailsArgs, reply *UpdateAccountDetailsReply) error {
	if middleware.VerifyAnyRole(r, middleware.AnonymousRole, middleware.UnverifiedRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("UpdateAccountDetails(): %v", err.Error())
		return &err
	}

	requestUser := r.Context().Value(middleware.Keys.UserKey)
	if requestUser == nil {
		err := JSONRPCErrorCodes[int(ERROR_JWT_PARSE_FAILURE)]
		core.Error("UpdateAccountDetails(): %v", err.Error())
		return &err
	}

	requestID, ok := requestUser.(*jwt.Token).Claims.(jwt.MapClaims)["sub"].(string)
	if !ok {
		err := JSONRPCErrorCodes[int(ERROR_JWT_PARSE_FAILURE)]
		core.Error("UpdateAccountDetails(): %v: Failed to parse user ID", err.Error())
		return &err
	}

	userAccount, err := s.UserManager.Read(requestID)
	if err != nil {
		core.Error("UpdateAccountDetails(): %v: Failed to read user account", err.Error())
		err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
		return &err
	}

	updateUser := &management.User{
		AppMetadata: map[string]interface{}{
			"newsletter": false,
		},
	}

	// Check current newsletter value, if different, update it
	currentNewsletter, ok := userAccount.AppMetadata["newsletter"]
	if !ok || currentNewsletter != args.Newsletter {
		updateUser.AppMetadata["newsletter"] = args.Newsletter
	}

	// Check first and last name, if different than what is passed in, update it
	if args.FirstName != "" && args.FirstName != userAccount.GetGivenName() {
		updateUser.GivenName = &args.FirstName
	}

	if args.LastName != "" && args.LastName != userAccount.GetFamilyName() {
		updateUser.FamilyName = &args.LastName
	}

	err = s.UserManager.Update(requestID, updateUser)
	if err != nil {
		core.Error("UpdateAccountDetails(): %v: Failed to update user account", err.Error())
		err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
		return &err
	}
	return nil
}

type ResetPasswordEmailArgs struct {
	Email string `json:"email"`
}

type ResetPasswordEmailReply struct{}

func (s *AuthService) ResetPasswordEmail(r *http.Request, args *ResetPasswordEmailArgs, reply *ResetPasswordEmailReply) error {
	if args.Email == "" {
		err := JSONRPCErrorCodes[int(ERROR_MISSING_FIELD)]
		err.Data.(*JSONRPCErrorData).MissingField = "Email"
		core.Error("ResetPasswordEmail(): %v: Email is required", err.Error())
		return &err
	}

	userAccounts, err := s.UserManager.List(management.Query(fmt.Sprintf(`email:"%s"`, args.Email)))
	if err != nil || len(userAccounts.Users) != 1 {
		err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
		core.Error("ResetPasswordEmail(): Failed to look up user account: %s", err.Error())
		return &err
	}

	if err = s.AuthenticationClient.SendChangePasswordEmail(args.Email); err != nil {
		err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
		core.Error("ResetPasswordEmail(): %v: Failed to send reset password email", err.Error())
		return &err
	}

	return nil
}

type VerifyEmailArgs struct {
	UserID string `json:"user_id"`
}

type VerifyEmailReply struct{}

func (s *AuthService) ResendVerificationEmail(r *http.Request, args *VerifyEmailArgs, reply *VerifyEmailReply) error {
	if !middleware.VerifyAnyRole(r, middleware.UnverifiedRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("ResendVerificationEmail(): %v", err.Error())
		return &err
	}

	job := &management.Job{
		UserID: &args.UserID,
	}

	err := s.JobManager.VerifyEmail(job)
	if err != nil {
		core.Error("VerifyEmailUrl(): %v: Failed to generate verification email", err.Error())
		err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
		return &err
	}

	return nil
}

// This is currently not in use but could be in the future

/* type AddContactArgs struct {
	Email string `json:"email"`
}

type AddContactReply struct {
}

func (s *AuthService) AddMailChimpContact(r *http.Request, args *AddContactArgs, reply *AddContactReply) error {
	if args.Email == "" {
		err := fmt.Errorf("AddMailChimpContact() email is required")
		s.Logger.Log("err", err)
		return err
	}

	if err := s.MailChimpManager.AddEmailToMailChimp(args.Email); err != nil {
		err := fmt.Errorf("AddMailChimpContact() failed to add signup: %s", err)
		s.Logger.Log("err", err)
	}

	if err := s.MailChimpManager.TagNewSignup(args.Email); err != nil {
		err := fmt.Errorf("AddMailChimpContact() failed to tag signup: %s", err)
		s.Logger.Log("err", err)
		return err
	}
	return nil
} */

type UpdateDomainsArgs struct {
	Domains []string `json:"domains"`
}

type UpdateDomainsReply struct {
}

func (s *AuthService) UpdateAutoSignupDomains(r *http.Request, args *UpdateDomainsArgs, reply *UpdateDomainsReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OwnerRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("UpdateAutoSignupDomains(): %v", err.Error())
		return &err
	}
	customerCode, ok := r.Context().Value(middleware.Keys.CustomerKey).(string)
	if !ok {
		err := JSONRPCErrorCodes[int(ERROR_JWT_PARSE_FAILURE)]
		core.Error("UpdateAutoSignupDomains(): %v: Failed to parse customer code", err.Error())
		return &err
	}
	if customerCode == "" {
		err := JSONRPCErrorCodes[int(ERROR_JWT_PARSE_FAILURE)]
		core.Error("UpdateAutoSignupDomains(): %v: Failed to parse customer code", err.Error())
		return &err
	}
	ctx := context.Background()

	company, err := s.Storage.Customer(r.Context(), customerCode)
	if err != nil {
		core.Error("UpdateAutoSignupDomains(): %v", err.Error())
		err := JSONRPCErrorCodes[int(ERROR_STORAGE_FAILURE)]
		return &err
	}

	company.AutomaticSignInDomains = strings.Join(args.Domains, ", ")

	err = s.Storage.SetCustomer(ctx, company)
	if err != nil {
		core.Error("UpdateAutoSignupDomains(): %v", err.Error())
		err := JSONRPCErrorCodes[int(ERROR_STORAGE_FAILURE)]
		return &err
	}

	return nil
}

type CustomerSlackNotification struct {
	Email        string `json:"email"`
	CustomerName string `json:"customer_name"`
	CustomerCode string `json:"customer_code"`
}

type GenericSlackNotificationArgs struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

type GenericSlackNotificationReply struct {
}

func (s *AuthService) CustomerSignedUpSlackNotification(r *http.Request, args *CustomerSlackNotification, reply *GenericSlackNotificationReply) error {
	if middleware.VerifyAllRoles(r, middleware.AnonymousRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("CustomerSignedUpSlackNotification(): %v", err.Error())
		return &err
	}

	if args.Email == "" {
		err := JSONRPCErrorCodes[int(ERROR_MISSING_FIELD)]
		err.Data.(*JSONRPCErrorData).MissingField = "Email"
		core.Error("UserAccount(): %v: Email is required", err.Error())
		return &err
	}

	message := fmt.Sprintf("%s signed up on the Portal! :tada:", args.Email)

	if err := s.SlackClient.SendInfo(message); err != nil {
		err := JSONRPCErrorCodes[int(ERROR_SLACK_FAILURE)]
		core.Error("CustomerSignedUpSlackNotification(): %v: Email is required", err.Error())
		return &err
	}
	return nil
}

func (s *AuthService) CustomerViewedTheDocsSlackNotification(r *http.Request, args *CustomerSlackNotification, reply *GenericSlackNotificationReply) error {
	if middleware.VerifyAllRoles(r, middleware.AnonymousRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("CustomerViewedTheDocsSlackNotification(): %v", err.Error())
		return &err
	}

	if args.Email == "" {
		err := JSONRPCErrorCodes[int(ERROR_MISSING_FIELD)]
		err.Data.(*JSONRPCErrorData).MissingField = "Email"
		core.Error("UserAccount(): %v: Email is required", err.Error())
		return &err
	}

	message := fmt.Sprintf("%s Viewed documentation", args.Email)

	if args.CustomerName != "" {
		message = fmt.Sprintf("%s from %s Viewed documentation", args.Email, args.CustomerName)
	}

	if args.CustomerCode != "" {
		message += fmt.Sprintf(" - Company Code: %s", args.CustomerCode)
	}

	message += " :muscle:"

	if err := s.SlackClient.SendInfo(message); err != nil {
		err := JSONRPCErrorCodes[int(ERROR_SLACK_FAILURE)]
		core.Error("CustomerViewedTheDocsSlackNotification(): %v: Email is required", err.Error())
		return &err
	}
	return nil
}

func (s *AuthService) CustomerDownloadedSDKSlackNotification(r *http.Request, args *CustomerSlackNotification, reply *GenericSlackNotificationReply) error {
	if middleware.VerifyAllRoles(r, middleware.AnonymousRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("CustomerDownloadedSDKSlackNotification(): %v", err.Error())
		return &err
	}

	if args.Email == "" {
		err := JSONRPCErrorCodes[int(ERROR_MISSING_FIELD)]
		err.Data.(*JSONRPCErrorData).MissingField = "Email"
		core.Error("UserAccount(): %v: Email is required", err.Error())
		return &err
	}

	message := fmt.Sprintf("%s downloaded the SDK", args.Email)

	if args.CustomerName != "" {
		message = fmt.Sprintf("%s from %s downloaded the SDK", args.Email, args.CustomerName)
	}

	if args.CustomerCode != "" {
		message += fmt.Sprintf(" - Company Code: %s", args.CustomerCode)
	}

	message += " :party_parrot:"

	if err := s.SlackClient.SendInfo(message); err != nil {
		err := JSONRPCErrorCodes[int(ERROR_SLACK_FAILURE)]
		core.Error("CustomerDownloadedSDKSlackNotification(): %v: Email is required", err.Error())
		return &err
	}
	return nil
}

func (s *AuthService) CustomerViewedSDKSourceNotification(r *http.Request, args *CustomerSlackNotification, reply *GenericSlackNotificationReply) error {
	if middleware.VerifyAnyRole(r, middleware.AnonymousRole, middleware.UnverifiedRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("CustomerViewedSDKSourceNotification(): %v", err.Error())
		return &err
	}

	if args.Email == "" {
		err := JSONRPCErrorCodes[int(ERROR_MISSING_FIELD)]
		err.Data.(*JSONRPCErrorData).MissingField = "Email"
		core.Error("CustomerViewedSDKSourceNotification(): %v", err.Error())
		return &err
	}

	message := fmt.Sprintf("%s viewed the SDK source", args.Email)

	if args.CustomerName != "" {
		message = fmt.Sprintf("%s from %s viewed the SDK source", args.Email, args.CustomerName)
	}

	if args.CustomerCode != "" {
		message += fmt.Sprintf(" - Company Code: %s", args.CustomerCode)
	}

	message += " :technologist:"

	if err := s.SlackClient.SendInfo(message); err != nil {
		err := JSONRPCErrorCodes[int(ERROR_SLACK_FAILURE)]
		core.Error("CustomerViewedSDKSourceNotification(): %v", err.Error())
		return &err
	}
	return nil
}

func (s *AuthService) CustomerEnteredPublicKeySlackNotification(r *http.Request, args *CustomerSlackNotification, reply *GenericSlackNotificationReply) error {
	if middleware.VerifyAllRoles(r, middleware.AnonymousRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("CustomerEnteredPublicKeySlackNotification(): %v", err.Error())
		return &err
	}

	if args.Email == "" {
		err := JSONRPCErrorCodes[int(ERROR_MISSING_FIELD)]
		err.Data.(*JSONRPCErrorData).MissingField = "Email"
		core.Error("UserAccount(): %v: Email is required", err.Error())
		return &err
	}

	message := fmt.Sprintf("%s entered a public key", args.Email)

	if args.CustomerName != "" {
		message = fmt.Sprintf("%s from %s entered a public key", args.Email, args.CustomerName)
	}

	if args.CustomerCode != "" {
		message += fmt.Sprintf(" - Company Code: %s", args.CustomerCode)
	}

	message += " :rocket:"

	if err := s.SlackClient.SendInfo(message); err != nil {
		err := JSONRPCErrorCodes[int(ERROR_SLACK_FAILURE)]
		core.Error("CustomerEnteredPublicKeySlackNotification(): %v: Email is required", err.Error())
		return &err
	}
	return nil
}

func (s *AuthService) CustomerDownloadedUE4PluginNotification(r *http.Request, args *CustomerSlackNotification, reply *GenericSlackNotificationReply) error {
	if middleware.VerifyAnyRole(r, middleware.AnonymousRole, middleware.UnverifiedRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("CustomerDownloadedUE4PluginNotification(): %v", err.Error())
		return &err
	}

	if args.Email == "" {
		err := JSONRPCErrorCodes[int(ERROR_MISSING_FIELD)]
		err.Data.(*JSONRPCErrorData).MissingField = "Email"
		core.Error("CustomerDownloadedUE4PluginNotification(): %v", err.Error())
		return &err
	}

	message := fmt.Sprintf("%s downloaded the UE4 plugin", args.Email)

	if args.CustomerName != "" {
		message = fmt.Sprintf("%s from %s downloaded the UE4 plugin", args.Email, args.CustomerName)
	}

	if args.CustomerCode != "" {
		message += fmt.Sprintf(" - Company Code: %s", args.CustomerCode)
	}

	message += " :the_horns:"

	if err := s.SlackClient.SendInfo(message); err != nil {
		err := JSONRPCErrorCodes[int(ERROR_SLACK_FAILURE)]
		core.Error("CustomerDownloadedUE4PluginNotification(): %v", err.Error())
		return &err
	}
	return nil
}

func (s *AuthService) CustomerViewedUE4SourceNotification(r *http.Request, args *CustomerSlackNotification, reply *GenericSlackNotificationReply) error {
	if middleware.VerifyAnyRole(r, middleware.AnonymousRole, middleware.UnverifiedRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("CustomerViewedUE4SourceNotification(): %v", err.Error())
		return &err
	}

	if args.Email == "" {
		err := JSONRPCErrorCodes[int(ERROR_MISSING_FIELD)]
		err.Data.(*JSONRPCErrorData).MissingField = "Email"
		core.Error("CustomerViewedUE4SourceNotification(): %v", err.Error())
		return &err
	}

	message := fmt.Sprintf("%s viewed the UE4 source", args.Email)

	if args.CustomerName != "" {
		message = fmt.Sprintf("%s from %s viewed the UE4 source", args.Email, args.CustomerName)
	}

	if args.CustomerCode != "" {
		message += fmt.Sprintf(" - Company Code: %s", args.CustomerCode)
	}

	message += " :information_source:"

	if err := s.SlackClient.SendInfo(message); err != nil {
		err := JSONRPCErrorCodes[int(ERROR_SLACK_FAILURE)]
		core.Error("CustomerViewedUE4SourceNotification(): %v", err.Error())
		return &err
	}
	return nil
}

func (s *AuthService) CustomerDownloadedUnityPluginNotification(r *http.Request, args *CustomerSlackNotification, reply *GenericSlackNotificationReply) error {
	if middleware.VerifyAnyRole(r, middleware.AnonymousRole, middleware.UnverifiedRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("CustomerDownloadedUnityPluginNotification(): %v", err.Error())
		return &err
	}

	if args.Email == "" {
		err := JSONRPCErrorCodes[int(ERROR_MISSING_FIELD)]
		err.Data.(*JSONRPCErrorData).MissingField = "Email"
		core.Error("CustomerDownloadedUnityPluginNotification(): %v", err.Error())
		return &err
	}

	message := fmt.Sprintf("%s downloaded the Unity plugin", args.Email)

	if args.CustomerName != "" {
		message = fmt.Sprintf("%s from %s downloaded the Unity plugin", args.Email, args.CustomerName)
	}

	if args.CustomerCode != "" {
		message += fmt.Sprintf(" - Company Code: %s", args.CustomerCode)
	}

	message += " :electric_plug:"

	if err := s.SlackClient.SendInfo(message); err != nil {
		err := JSONRPCErrorCodes[int(ERROR_SLACK_FAILURE)]
		core.Error("CustomerDownloadedUnityPluginNotification(): %v", err.Error())
		return &err
	}
	return nil
}

func (s *AuthService) CustomerViewedUnitySourceNotification(r *http.Request, args *CustomerSlackNotification, reply *GenericSlackNotificationReply) error {
	if middleware.VerifyAnyRole(r, middleware.AnonymousRole, middleware.UnverifiedRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("CustomerViewedUnitySourceNotification(): %v", err.Error())
		return &err
	}

	if args.Email == "" {
		err := JSONRPCErrorCodes[int(ERROR_MISSING_FIELD)]
		err.Data.(*JSONRPCErrorData).MissingField = "Email"
		core.Error("CustomerViewedUnitySourceNotification(): %v", err.Error())
		return &err
	}

	message := fmt.Sprintf("%s viewed the Unity source", args.Email)

	if args.CustomerName != "" {
		message = fmt.Sprintf("%s from %s viewed the Unity source", args.Email, args.CustomerName)
	}

	if args.CustomerCode != "" {
		message += fmt.Sprintf(" - Company Code: %s", args.CustomerCode)
	}

	message += " :information_source:"

	if err := s.SlackClient.SendInfo(message); err != nil {
		err := JSONRPCErrorCodes[int(ERROR_SLACK_FAILURE)]
		core.Error("CustomerViewedUnitySourceNotification(): %v", err.Error())
		return &err
	}
	return nil
}

func (s *AuthService) CustomerDownloaded2022WhitePaperNotification(r *http.Request, args *CustomerSlackNotification, reply *GenericSlackNotificationReply) error {
	if middleware.VerifyAnyRole(r, middleware.AnonymousRole, middleware.UnverifiedRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("CustomerDownloadedWhitePaperNotification(): %v", err.Error())
		return &err
	}

	if args.Email == "" {
		err := JSONRPCErrorCodes[int(ERROR_MISSING_FIELD)]
		err.Data.(*JSONRPCErrorData).MissingField = "Email"
		core.Error("CustomerDownloadedWhitePaperNotification(): %v", err.Error())
		return &err
	}

	message := fmt.Sprintf("%s downloaded the 2022 white paper", args.Email)

	if args.CustomerName != "" {
		message = fmt.Sprintf("%s from %s downloaded the 2022 white paper", args.Email, args.CustomerName)
	}

	if args.CustomerCode != "" {
		message += fmt.Sprintf(" - Company Code: %s", args.CustomerCode)
	}

	message += " :microscope:"

	if err := s.SlackClient.SendInfo(message); err != nil {
		err := JSONRPCErrorCodes[int(ERROR_SLACK_FAILURE)]
		core.Error("CustomerDownloadedWhitePaperNotification(): %v", err.Error())
		return &err
	}
	return nil
}

func (s *AuthService) CustomerDownloadedENetDownloadNotification(r *http.Request, args *CustomerSlackNotification, reply *GenericSlackNotificationReply) error {
	if middleware.VerifyAnyRole(r, middleware.AnonymousRole, middleware.UnverifiedRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("CustomerDownloadedENetDownloadNotification(): %v", err.Error())
		return &err
	}

	if args.Email == "" {
		err := JSONRPCErrorCodes[int(ERROR_MISSING_FIELD)]
		err.Data.(*JSONRPCErrorData).MissingField = "Email"
		core.Error("CustomerDownloadedENetDownloadNotification(): %v", err.Error())
		return &err
	}

	message := fmt.Sprintf("%s downloaded ENet", args.Email)

	if args.CustomerName != "" {
		message = fmt.Sprintf("%s from %s downloaded ENet", args.Email, args.CustomerName)
	}

	if args.CustomerCode != "" {
		message += fmt.Sprintf(" - Company Code: %s", args.CustomerCode)
	}

	message += " :signal_strength:"

	if err := s.SlackClient.SendInfo(message); err != nil {
		err := JSONRPCErrorCodes[int(ERROR_SLACK_FAILURE)]
		core.Error("CustomerDownloadedENetDownloadNotification(): %v", err.Error())
		return &err
	}
	return nil
}

func (s *AuthService) CustomerViewedENetSourceNotification(r *http.Request, args *CustomerSlackNotification, reply *GenericSlackNotificationReply) error {
	if middleware.VerifyAnyRole(r, middleware.AnonymousRole, middleware.UnverifiedRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("CustomerViewedENetSourceNotification(): %v", err.Error())
		return &err
	}

	if args.Email == "" {
		err := JSONRPCErrorCodes[int(ERROR_MISSING_FIELD)]
		err.Data.(*JSONRPCErrorData).MissingField = "Email"
		core.Error("CustomerViewedENetSourceNotification(): %v", err.Error())
		return &err
	}

	message := fmt.Sprintf("%s viewed the ENet source", args.Email)

	if args.CustomerName != "" {
		message = fmt.Sprintf("%s from %s viewed the ENet source", args.Email, args.CustomerName)
	}

	if args.CustomerCode != "" {
		message += fmt.Sprintf(" - Company Code: %s", args.CustomerCode)
	}

	message += " :information_source:"

	if err := s.SlackClient.SendInfo(message); err != nil {
		err := JSONRPCErrorCodes[int(ERROR_SLACK_FAILURE)]
		core.Error("CustomerViewedENetSourceNotification(): %v", err.Error())
		return &err
	}
	return nil
}

func (s *AuthService) SlackNotification(r *http.Request, args *GenericSlackNotificationArgs, reply *GenericSlackNotificationReply) error {
	if !middleware.VerifyAllRoles(r, middleware.AdminRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		core.Error("SlackNotification(): %v", err.Error())
		return &err
	}

	if args.Message == "" {
		err := JSONRPCErrorCodes[int(ERROR_MISSING_FIELD)]
		err.Data.(*JSONRPCErrorData).MissingField = "Message"
		core.Error("SlackNotification(): %v: Message is required", err.Error())
		return &err
	}

	if args.Type == "" {
		err := JSONRPCErrorCodes[int(ERROR_MISSING_FIELD)]
		err.Data.(*JSONRPCErrorData).MissingField = "Type"
		core.Error("SlackNotification(): %v: Type is required", err.Error())
		return &err
	}

	switch args.Type {
	case "info":
		if err := s.SlackClient.SendInfo(args.Message); err != nil {
			err := JSONRPCErrorCodes[int(ERROR_SLACK_FAILURE)]
			core.Error("CustomerSignedUpSlackNotification(): %v: Failed to send info Slack notification: %s", args.Message, err.Error())
			return &err
		}
	case "warning":
		if err := s.SlackClient.SendWarning(args.Message); err != nil {
			err := JSONRPCErrorCodes[int(ERROR_SLACK_FAILURE)]
			core.Error("CustomerSignedUpSlackNotification(): %v: Failed to send warning Slack notification: %s", args.Message, err.Error())
			return &err
		}
	case "error":
		if err := s.SlackClient.SendError(args.Message); err != nil {
			err := JSONRPCErrorCodes[int(ERROR_SLACK_FAILURE)]
			core.Error("CustomerSignedUpSlackNotification(): %v: Failed to send error Slack notification: %s", args.Message, err.Error())
			return &err
		}
	default:
		err := JSONRPCErrorCodes[int(ERROR_ILLEGAL_OPERATION)]
		core.Error("CustomerSignedUpSlackNotification(): %v: Slack notification type not supported: %s", args.Type, err.Error())
		return &err
	}
	return nil
}

type ProcessNewSignupArgs struct {
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	Email          string `json:"email"`
	CompanyName    string `json:"company_name"`
	CompanyWebsite string `json:"company_website"`
}

type ProcessNewSignupReply struct {
}

func (s *AuthService) ProcessNewSignup(r *http.Request, args *ProcessNewSignupArgs, reply *ProcessNewSignupReply) error {
	message := fmt.Sprintf("%s %s signed up on the Portal! :tada:\nEmail: %s\nCompany Name: %s\nCompany Website: %s", args.FirstName, args.LastName, args.Email, args.CompanyName, args.CompanyWebsite)

	// If we error here we don't worry about it. Setting up everything is more important and the error shouldn't hinder the rest of the sign up process on the portal side
	if err := s.SlackClient.SendInfo(message); err != nil {
		core.Error("ProcessNewSignup(): %v: Failed to send slack notification", err.Error())
	}

	if s.HubSpotClient.APIKey != "" {
		companies, err := s.HubSpotClient.CompanyEntrySearch(args.CompanyName, args.CompanyWebsite)
		if err != nil {
			core.Error("ProcessNewSignup(): %v: Failed to fetch company entries from hubspot", err.Error())
		}

		foundCompanyID := ""
		if len(companies) != 0 {
			foundCompanyID = companies[0].ID
		}

		contacts, err := s.HubSpotClient.ContactEntrySearch(args.FirstName, args.LastName, args.Email)
		if err != nil {
			core.Error("ProcessNewSignup(): %v: Failed to fetch contact entries from hubspot", err.Error())
		}

		foundContactID := ""
		if len(contacts) != 0 {
			foundContactID = contacts[0].ID
		}

		message := fmt.Sprintf("First name: %s - Last name: %s - Email: %s - Company name: %s - Website: %s", args.FirstName, args.LastName, args.Email, args.CompanyName, args.CompanyWebsite)

		// Company and contact do not exist
		if foundCompanyID == "" && foundContactID == "" {
			// Create contact and associate it with NewFunnelCo with notes about sign up
			newContactID, err := s.HubSpotClient.CreateNewContactEntry(args.FirstName, args.LastName, args.Email, args.CompanyName, args.CompanyWebsite)
			if err != nil {
				core.Error("ProcessNewSignup(): %v: Failed to add contact entry to hubspot", err.Error())
			} else {
				if err := s.HubSpotClient.AssociateCompanyToContact(notifications.NewFunnelCoID, newContactID); err != nil {
					core.Error("ProcessNewSignup(): %v: Failed to associate NewFunnelCo to new contact", err.Error())
				}
				if err := s.HubSpotClient.AssociateContactToCompany(newContactID, notifications.NewFunnelCoID); err != nil {
					core.Error("ProcessNewSignup(): %v: Failed to associate new contact to NewFunnelCo", err.Error())
				}
				if err := s.HubSpotClient.CreateCompanyNote(message, notifications.NewFunnelCoID); err != nil {
					core.Error("ProcessNewSignup(): %v: Failed to NewFunnelCo note", err.Error())
				}
				if err := s.HubSpotClient.CreateContactNote(message, newContactID); err != nil {
					core.Error("ProcessNewSignup(): %v: Failed to create new contact note", err.Error())
				}
			}
		}

		// Company exists but contact does not
		if foundCompanyID != "" && foundContactID == "" {
			// Create contact and associate it to NewFunnelCo with notes about sign up
			newContactID, err := s.HubSpotClient.CreateNewContactEntry(args.FirstName, args.LastName, args.Email, args.CompanyName, args.CompanyWebsite)
			if err != nil {
				core.Error("ProcessNewSignup(): %v: Failed to add contact entry to hubspot", err.Error())
			} else {
				if err := s.HubSpotClient.AssociateCompanyToContact(foundCompanyID, newContactID); err != nil {
					core.Error("ProcessNewSignup(): %v: Failed to associate found company to new contact", err.Error())
				}
				if err := s.HubSpotClient.AssociateContactToCompany(newContactID, foundCompanyID); err != nil {
					core.Error("ProcessNewSignup(): %v: Failed to associate new contact to found company", err.Error())
				}
				if err := s.HubSpotClient.CreateCompanyNote(message, foundCompanyID); err != nil {
					core.Error("ProcessNewSignup(): %v: Failed to create found company note", err.Error())
				}
				if err := s.HubSpotClient.CreateContactNote(message, newContactID); err != nil {
					core.Error("ProcessNewSignup(): %v: Failed to create new contact note", err.Error())
				}
			}
		}

		// Company doesn't exist but contact does
		if foundCompanyID == "" && foundContactID != "" {
			// Associate contact with NewFunnelCo with note about signup with company name and website
			if err := s.HubSpotClient.AssociateCompanyToContact(notifications.NewFunnelCoID, foundContactID); err != nil {
				core.Error("ProcessNewSignup(): %v: Failed to associate NewFunnelCo to found contact", err.Error())
			}
			if err := s.HubSpotClient.AssociateContactToCompany(foundContactID, notifications.NewFunnelCoID); err != nil {
				core.Error("ProcessNewSignup(): %v: Failed to associate found contact to NewFunnelCo", err.Error())
			}
			if err := s.HubSpotClient.CreateCompanyNote(message, notifications.NewFunnelCoID); err != nil {
				core.Error("ProcessNewSignup(): %v: Failed to NewFunnelCo note", err.Error())
			}
			if err := s.HubSpotClient.CreateContactNote(message, foundContactID); err != nil {
				core.Error("ProcessNewSignup(): %v: Failed to create found contact note", err.Error())
			}
		}

		// Company and contact exist
		if foundCompanyID != "" && foundContactID != "" {
			// Associate the company and contact and make notes on both entries about the sign up
			if err := s.HubSpotClient.AssociateCompanyToContact(foundCompanyID, foundContactID); err != nil {
				core.Error("ProcessNewSignup(): %v: Failed to associate found company to found contact", err.Error())
			}
			if err := s.HubSpotClient.AssociateContactToCompany(foundContactID, foundCompanyID); err != nil {
				core.Error("ProcessNewSignup(): %v: Failed to associate found contact to found company", err.Error())
			}
			if err := s.HubSpotClient.CreateCompanyNote(message, foundCompanyID); err != nil {
				core.Error("ProcessNewSignup(): %v: Failed to create found company note", err.Error())
			}
			if err := s.HubSpotClient.CreateContactNote(message, foundContactID); err != nil {
				core.Error("ProcessNewSignup(): %v: Failed to create found contact note", err.Error())
			}
		}
	}

	userAccounts, err := s.UserManager.List(management.Query(fmt.Sprintf(`email:"%s"`, args.Email)))
	if err != nil || len(userAccounts.Users) != 1 {
		err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
		core.Error("ProcessNewSignup(): Failed to look up user account: %s", err.Error())
		return &err
	}

	user := userAccounts.Users[0]

	updateUser := &management.User{
		FamilyName: &args.LastName,
		GivenName:  &args.FirstName,
	}

	if err := s.UserManager.Update(user.GetID(), updateUser); err != nil {
		core.Error("ProcessNewSignup(): %v: Failed to update new user's first and last name", err.Error())
		err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
		return &err
	}

	return nil
}

func (s *AuthService) RefreshAuthRolesCache() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	roleList, err := s.RoleManager.List()
	if err != nil {
		core.Error("RefreshAuthRolesCache(): Failed to refresh role caches: %s", err.Error())
		return err
	}

	roles := roleList.Roles

	s.RoleCache = make(map[string]*management.Role)

	for _, r := range roles {
		s.RoleCache[r.GetName()] = r
	}

	return nil
}

func (s *AuthService) RoleCacheToArray(removeAdmin bool) []*management.Role {
	allRoles := []*management.Role{}
	// Convert map to array
	for _, role := range s.RoleCache {
		// Don't remove admin role from admin. They can be removed from a company account but should retain admin privileges
		if removeAdmin && role.GetName() == s.RoleCache["Admin"].GetName() {
			continue
		}
		allRoles = append(allRoles, role)
	}

	return allRoles
}

func (s *AuthService) CleanUpExplorerRoles(ctx context.Context) error {
	allUserAccounts, err := s.FetchAllAccountsFromAuth0()
	if err != nil {
		core.Error("CleanUpExplorerRoles(): %v: Failed to fetch user list", err.Error())
		return err
	}

	currentTime := time.Now().UTC()

	removedUsers := make([]string, 0)
	for _, a := range allUserAccounts {
		// If the account hasn't logged in in 30 days or more
		if currentTime.Sub(a.GetLastLogin()) >= (time.Hour * 24 * 30) {
			customerCode, ok := a.AppMetadata["company_code"].(string)
			if !ok || customerCode == "" {
				continue
			}

			buyerAccount, err := s.Storage.BuyerWithCompanyCode(ctx, customerCode)
			if err != nil {
				core.Error("CleanUpExplorerRoles(): %v: Failed to look up buyer account", err.Error())
				err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
				return &err
			}

			// If the buyer is live, ignore them
			if buyerAccount.Live {
				continue
			}

			// Get all of the roles assigned to the user
			userRoles, err := s.UserManager.Roles(*a.ID)
			if err != nil {
				core.Error("CleanUpExplorerRoles(): %v: Failed to get user roles", err.Error())
				err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
				return &err
			}

			for _, role := range userRoles.Roles {
				explorerRole := s.RoleCache["Explorer"]
				if role.GetName() == explorerRole.GetName() {
					err = s.UserManager.RemoveRoles(*a.ID, explorerRole)
					if err != nil {
						core.Error("CleanUpExplorerRoles(): %v: Failed to remove explorer role", err.Error())
						return err
					}

					removedUsers = append(removedUsers, a.GetID())
				}
			}
		}
	}

	// Loop through removed user IDs and delete them from our Looker account
	for _, userID := range removedUsers {
		err := s.LookerClient.RemoveLookerUserByAuth0ID(userID)
		if err != nil {
			return err
		}
	}

	return nil
}
