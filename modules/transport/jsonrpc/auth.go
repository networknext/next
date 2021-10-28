package jsonrpc

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-kit/kit/log"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/storage"
	"github.com/networknext/backend/modules/transport/middleware"
	"github.com/networknext/backend/modules/transport/notifications"
	"gopkg.in/auth0.v4/management"
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
	Logger               log.Logger
	LookerSecret         string
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
	var totalUsers []*management.User
	ctx := r.Context()

	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OwnerRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		s.Logger.Log("err", fmt.Errorf("AllAccounts(): %v", err.Error()))
		return &err
	}

	reply.UserAccounts = make([]account, 0)
	keepSearching := true
	page := 0

	for keepSearching {
		accountList, err := s.UserManager.List(management.PerPage(100), management.Page(page))
		if err != nil {
			s.Logger.Log("err", fmt.Errorf("AllAccounts(): %v: Failed to fetch user list", err.Error()))
			err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
			return &err
		}

		totalUsers = append(totalUsers, accountList.Users...)
		if len(totalUsers)%100 != 0 {
			keepSearching = false
		}
		page++
	}

	requestUser := middleware.RequestUserInformation(ctx)
	if requestUser == nil {
		err := JSONRPCErrorCodes[int(ERROR_JWT_PARSE_FAILURE)]
		s.Logger.Log("err", fmt.Errorf("AllAccounts(): %v: Failed to parse user", err.Error()))
		return &err
	}

	requestCompany := middleware.RequestUserCustomerCode(ctx)
	if requestCompany == "" {
		err := JSONRPCErrorCodes[int(ERROR_USER_IS_NOT_ASSIGNED)]
		s.Logger.Log("err", fmt.Errorf("AllAccounts(): %v", err.Error()))
		return &err
	}

	for _, a := range totalUsers {
		companyCode, ok := a.AppMetadata["company_code"].(string)
		if !ok || requestCompany != companyCode {
			continue
		}
		userRoles, err := s.UserManager.Roles(*a.ID)
		if err != nil {
			s.Logger.Log("err", fmt.Errorf("AllAccounts(): %v: Failed to get user roles", err.Error()))
			err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
			return &err
		}

		buyer, _ := s.Storage.BuyerWithCompanyCode(r.Context(), companyCode)
		seller, _ := s.Storage.SellerWithCompanyCode(r.Context(), companyCode)
		company, err := s.Storage.Customer(r.Context(), companyCode)
		if err != nil {
			s.Logger.Log("err", fmt.Errorf("AllAccounts(): %v", err.Error()))
			err := JSONRPCErrorCodes[int(ERROR_STORAGE_FAILURE)]
			return &err
		}

		reply.UserAccounts = append(reply.UserAccounts, newAccount(a, userRoles.Roles, buyer, company.Name, company.Code, seller.Name != ""))
	}

	return nil
}

func (s *AuthService) UserAccount(r *http.Request, args *AccountArgs, reply *AccountReply) error {
	if args.UserID == "" {
		err := JSONRPCErrorCodes[int(ERROR_MISSING_FIELD)]
		err.Data.(*JSONRPCErrorData).MissingField = "UserID"
		s.Logger.Log("err", fmt.Errorf("UserAccount(): %v: UserID is required", err.Error()))
		return &err
	}

	user := r.Context().Value(middleware.Keys.UserKey)
	if user == nil {
		err := JSONRPCErrorCodes[int(ERROR_JWT_PARSE_FAILURE)]
		s.Logger.Log("err", fmt.Errorf("UserAccount(): %v", err.Error()))
		return &err
	}

	claims := user.(*jwt.Token).Claims.(jwt.MapClaims)
	requestID, ok := claims["sub"].(string)
	if !ok {
		err := JSONRPCErrorCodes[int(ERROR_JWT_PARSE_FAILURE)]
		s.Logger.Log("err", fmt.Errorf("UserAccount(): %v: Failed to parse user ID", err.Error()))
		return &err
	}
	if requestID != args.UserID && !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OwnerRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		s.Logger.Log("err", fmt.Errorf("UserAccount(): %v", err.Error()))
		return &err
	}

	userAccount, err := s.UserManager.Read(args.UserID)
	if err != nil {
		s.Logger.Log("err", fmt.Errorf("UserAccount(): %v: Failed to get user account details", err.Error()))
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
			s.Logger.Log("err", fmt.Errorf("UserAccount(): %v: Could not find customer account for customer code: %v", err.Error(), companyCode))
			err := JSONRPCErrorCodes[int(ERROR_STORAGE_FAILURE)]
			return &err
		}
	}
	buyer, _ := s.Storage.BuyerWithCompanyCode(r.Context(), companyCode)
	seller, _ := s.Storage.SellerWithCompanyCode(r.Context(), companyCode)

	userRoles, err := s.UserManager.Roles(*userAccount.ID)
	if err != nil {
		s.Logger.Log("err", fmt.Errorf("UserAccount(): %v: Failed to get user account roles", err.Error()))
		err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
		return &err
	}

	if middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OwnerRole) && requestID == args.UserID {
		reply.Domains = strings.Split(company.AutomaticSignInDomains, ",")
	}

	reply.UserAccount = newAccount(userAccount, userRoles.Roles, buyer, company.Name, company.Code, seller.Name != "")

	return nil
}

func (s *AuthService) DeleteUserAccount(r *http.Request, args *AccountArgs, reply *AccountReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OwnerRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		s.Logger.Log("err", fmt.Errorf("DeleteUserAccount(): %v", err.Error()))
		return &err
	}

	if args.UserID == "" {
		err := JSONRPCErrorCodes[int(ERROR_MISSING_FIELD)]
		err.Data.(*JSONRPCErrorData).MissingField = "UserID"
		s.Logger.Log("err", fmt.Errorf("DeleteUserAccount(): %v", err.Error()))
		return &err
	}
	user, err := s.UserManager.Read(args.UserID)
	if err != nil {
		s.Logger.Log("err", fmt.Errorf("DeleteUserAccount(): %v: Failed to read user account", err.Error()))
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
		s.Logger.Log("err", fmt.Errorf("DeleteUserAccount(): %v", err.Error()))
		return &err
	}

	if err := s.UserManager.Update(args.UserID, &management.User{
		AppMetadata: map[string]interface{}{
			"company_code": "",
		},
	}); err != nil {
		s.Logger.Log("err", fmt.Errorf("DeleteUserAccount(): %v: Failed to update deleted user company code", err.Error()))
		err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
		return &err
	}
	return nil
}

func (s *AuthService) AddUserAccount(r *http.Request, args *AccountsArgs, reply *AccountsReply) error {
	var adminString string = "Admin"
	var accounts []account

	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OwnerRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		s.Logger.Log("err", fmt.Errorf("AddUserAccount(): %v", err.Error()))
		return &err
	}

	// Check if non admin is assigning admin role
	for _, role := range args.Roles {
		if role.Name == &adminString && !middleware.VerifyAllRoles(r, middleware.AdminRole) {
			err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
			s.Logger.Log("err", fmt.Errorf("AddUserAccount(): %v", err.Error()))
			return &err
		}
	}

	// Gather request user information
	userCompanyCode, ok := r.Context().Value(middleware.Keys.CustomerKey).(string)
	if !ok || userCompanyCode == "" {
		err := JSONRPCErrorCodes[int(ERROR_USER_IS_NOT_ASSIGNED)]
		s.Logger.Log("err", fmt.Errorf("AddUserAccount(): %v", err.Error()))
		return &err
	}

	connectionID := "Username-Password-Authentication"
	emails := args.Emails
	falseValue := false

	buyer, _ := s.Storage.BuyerWithCompanyCode(r.Context(), userCompanyCode)
	seller, _ := s.Storage.SellerWithCompanyCode(r.Context(), userCompanyCode)

	registered := make(map[string]*management.User)

	var totalUsers []*management.User
	keepSearching := true
	page := 0

	for keepSearching {
		accountList, err := s.UserManager.List(management.PerPage(100), management.Page(page))
		if err != nil {
			s.Logger.Log("err", fmt.Errorf("AllAccounts(): %v: Failed to fetch user list", err.Error()))
			err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
			return &err
		}

		totalUsers = append(totalUsers, accountList.Users...)
		if len(totalUsers)%100 != 0 {
			keepSearching = false
		}
		page++
	}

	emailString := strings.Join(emails, ",")

	for _, a := range totalUsers {
		if strings.Contains(emailString, *a.Email) {
			registered[*a.Email] = a
		}
	}

	emptyName := ""

	// Create an account for each new email
	var newUser *management.User
	for _, e := range emails {
		user, ok := registered[e]
		if !ok {
			pw, err := GenerateRandomString(32)
			if err != nil {
				err := JSONRPCErrorCodes[int(ERROR_PASSWORD_GENERATION_FAILURE)]
				s.Logger.Log("err", fmt.Errorf("AddUserAccount(): %v", err.Error()))
				return &err
			}
			newUser = &management.User{
				Connection:    &connectionID,
				Email:         &e,
				Name:          &emptyName,
				FamilyName:    &emptyName,
				EmailVerified: &falseValue,
				VerifyEmail:   &falseValue,
				Password:      &pw,
				AppMetadata: map[string]interface{}{
					"company_code": userCompanyCode,
				},
			}
			if err = s.UserManager.Create(newUser); err != nil {
				s.Logger.Log("err", fmt.Errorf("AddUserAccount(): %v: Failed to create new user account", err.Error()))
				err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
				return &err
			}
			if len(args.Roles) > 0 {
				if err = s.UserManager.AssignRoles(*newUser.ID, args.Roles...); err != nil {
					s.Logger.Log("err", fmt.Errorf("AddUserAccount(): %v: Failed to assign new user roles", err.Error()))
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
				s.Logger.Log("err", fmt.Errorf("AddUserAccount(): %v: Failed to update user account", err.Error()))
				err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
				return &err
			}
			roles := []*management.Role{
				s.RoleCache["Viewer"],
				s.RoleCache["Owner"],
				s.RoleCache["Admin"],
			}
			if err := s.UserManager.RemoveRoles(user.GetID(), roles...); err != nil {
				s.Logger.Log("err", fmt.Errorf("AddUserAccount(): %v: Failed to remove exist roles from user account", err.Error()))
				err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
				return &err
			}
			if len(args.Roles) > 0 {
				if err := s.UserManager.AssignRoles(user.GetID(), args.Roles...); err != nil {
					s.Logger.Log("err", fmt.Errorf("AddUserAccount(): %v: Failed to assign new roles to user account", err.Error()))
					err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
					return &err
				}
			}
		}

		company, err := s.Storage.Customer(r.Context(), userCompanyCode)
		if err != nil {
			s.Logger.Log("err", fmt.Errorf("AddUserAccount(): %v", err.Error()))
			err := JSONRPCErrorCodes[int(ERROR_STORAGE_FAILURE)]
			return &err
		}
		accounts = append(accounts, newAccount(newUser, args.Roles, buyer, company.Name, company.Code, seller.Name != ""))
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

func newAccount(u *management.User, r []*management.Role, buyer routing.Buyer, companyName string, companyCode string, isSeller bool) account {
	buyerID := ""
	if buyer.ID != 0 {
		buyerID = fmt.Sprintf("%016x", buyer.ID)
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
		Roles:       r,
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
		s.Logger.Log("err", fmt.Errorf("AllRoles(): %v", err.Error()))
		return &err
	}

	var totalUsers []*management.User
	keepSearching := true
	page := 0

	for keepSearching {
		accountList, err := s.UserManager.List(management.PerPage(100), management.Page(page))
		if err != nil {
			s.Logger.Log("err", fmt.Errorf("AllAccounts(): %v: Failed to fetch user list", err.Error()))
			err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
			return &err
		}

		totalUsers = append(totalUsers, accountList.Users...)
		if len(totalUsers)%100 != 0 {
			keepSearching = false
		}
		page++
	}

	for _, account := range totalUsers {
		var companyCode string
		companyCode, ok := account.AppMetadata["company_code"].(string)
		if !ok || (companyCode == "" || companyCode == "next") {
			continue
		}

		userRoles, err := s.UserManager.Roles(account.GetID())
		if err != nil {
			s.Logger.Log("err", fmt.Errorf("AllAccounts(): %v: Failed to get user roles", err.Error()))
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

	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OwnerRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		s.Logger.Log("err", fmt.Errorf("AllRoles(): %v", err.Error()))
		return &err
	}

	if middleware.VerifyAllRoles(r, middleware.AdminRole) {
		reply.Roles = []*management.Role{
			s.RoleCache["Viewer"],
			s.RoleCache["Owner"],
			s.RoleCache["Admin"],
		}
	} else {
		reply.Roles = []*management.Role{
			s.RoleCache["Viewer"],
			s.RoleCache["Owner"],
		}
	}

	return nil
}

func (s *AuthService) UserRoles(r *http.Request, args *RolesArgs, reply *RolesReply) error {
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OwnerRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		s.Logger.Log("err", fmt.Errorf("UserRoles(): %v", err.Error()))
		return &err
	}

	if args.UserID == "" {
		err := JSONRPCErrorCodes[int(ERROR_MISSING_FIELD)]
		err.Data.(*JSONRPCErrorData).MissingField = "UserID"
		s.Logger.Log("err", fmt.Errorf("UserRoles(): %v", err.Error()))
		return &err
	}

	userRoles, err := s.UserManager.Roles(args.UserID)
	if err != nil {
		s.Logger.Log("err", fmt.Errorf("UserRoles(): %v: Failed to fetch user roles", err.Error()))
		err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
		return &err
	}

	reply.Roles = userRoles.Roles

	return nil
}

func (s *AuthService) UpdateUserRoles(r *http.Request, args *RolesArgs, reply *RolesReply) error {
	var err error
	if !middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OwnerRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		s.Logger.Log("err", fmt.Errorf("UpdateUserRoles(): %v", err.Error()))
		return &err
	}

	if args.UserID == "" {
		err := JSONRPCErrorCodes[int(ERROR_MISSING_FIELD)]
		err.Data.(*JSONRPCErrorData).MissingField = "UserID"
		s.Logger.Log("err", fmt.Errorf("UpdateUserRoles(): %v: missing UserID", err.Error()))
		return &err
	}

	userRoles, err := s.UserManager.Roles(args.UserID)
	if err != nil {
		s.Logger.Log("err", fmt.Errorf("UpdateUserRoles(): %v: failed to fetch user roles", err.Error()))
		err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
		return &err
	}

	removeRoles := []*management.Role{
		s.RoleCache["Viewer"],
		s.RoleCache["Owner"],
	}

	if len(userRoles.Roles) > 0 {
		if middleware.VerifyAllRoles(r, middleware.AdminRole) {
			err = s.UserManager.RemoveRoles(args.UserID, removeRoles...)
			if err != nil {
				s.Logger.Log("err", fmt.Errorf("UpdateUserRoles(): %v: Failed to remove old user roles", err.Error()))
				err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
				return &err
			}
		} else {
			err = s.UserManager.RemoveRoles(args.UserID, userRoles.Roles...)
			if err != nil {
				s.Logger.Log("err", fmt.Errorf("UpdateUserRoles(): %v: Failed to remove old user roles", err.Error()))
				err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
				return &err
			}
		}
	}

	if len(args.Roles) == 0 {
		reply.Roles = make([]*management.Role, 0)
		return nil
	}

	// Make sure someone who isn't admin isn't assigning admin
	for _, role := range args.Roles {
		if *role.Name == "Admin" && !middleware.VerifyAllRoles(r, middleware.AdminRole) {
			err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
			s.Logger.Log("err", fmt.Errorf("UpdateUserRoles(): %v", err.Error()))
			return &err
		}
	}

	err = s.UserManager.AssignRoles(args.UserID, args.Roles...)
	if err != nil {
		s.Logger.Log("err", fmt.Errorf("UpdateUserRoles(): %v: Failed to assign user roles", err.Error()))
		err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
		return &err
	}

	reply.Roles = args.Roles
	return nil
}

type SetupCompanyAccountArgs struct {
	CompanyName string `json:"company_name"`
	CompanyCode string `json:"company_code"`
}

type SetupCompanyAccountReply struct {
	CompanyName string `json:"company_name"`
	CompanyCode string `json:"company_code"`
}

func (s *AuthService) SetupCompanyAccount(r *http.Request, args *SetupCompanyAccountArgs, reply *SetupCompanyAccountReply) error {
	if middleware.VerifyAnyRole(r, middleware.AnonymousRole, middleware.UnverifiedRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		s.Logger.Log("err", fmt.Errorf("SetupCompanyAccount(): %v", err.Error()))
		return &err
	}

	if args.CompanyCode == "" {
		err := JSONRPCErrorCodes[int(ERROR_MISSING_FIELD)]
		err.Data.(*JSONRPCErrorData).MissingField = "CompanyCode"
		s.Logger.Log("err", fmt.Errorf("SetupCompanyAccount(): %v: missing CompanyCode", err.Error()))
		return &err
	}

	assignedCompanyCode, ok := r.Context().Value(middleware.Keys.CustomerKey).(string)
	if ok && assignedCompanyCode != "" {
		err := JSONRPCErrorCodes[int(ERROR_ILLEGAL_OPERATION)]
		s.Logger.Log("err", fmt.Errorf("SetupCompanyAccount(): %v: User is already assigned to a company. Please reach out to support for further assistance.", err.Error()))
		return &err
	}

	// grab request user information
	requestUser := r.Context().Value(middleware.Keys.UserKey)
	if requestUser == nil {
		err := JSONRPCErrorCodes[int(ERROR_JWT_PARSE_FAILURE)]
		s.Logger.Log("err", fmt.Errorf("SetupCompanyAccount(): %v", err.Error()))
		return &err
	}

	// get request user ID for role assignment
	requestID, ok := requestUser.(*jwt.Token).Claims.(jwt.MapClaims)["sub"].(string)
	if !ok {
		err := JSONRPCErrorCodes[int(ERROR_JWT_PARSE_FAILURE)]
		s.Logger.Log("err", fmt.Errorf("SetupCompanyAccount(): %v: Failed to parse user ID", err.Error()))
		return &err
	}

	ctx := r.Context()
	roles := []*management.Role{
		s.RoleCache["Viewer"],
	}

	// Check if customer account exists already
	customer, err := s.Storage.Customer(ctx, args.CompanyCode)
	if err == nil {
		// exists, Check if the users domain matches the automatic signup domains
		email := requestUser.(*jwt.Token).Claims.(jwt.MapClaims)["email"].(string)
		domain := strings.Split(email, "@")
		if !strings.Contains(customer.AutomaticSignInDomains, domain[1]) {
			err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
			s.Logger.Log("err", fmt.Errorf("SetupCompanyAccount(): %v: User's email domain is not listed in accepted domains", err.Error()))
			return &err
		}
	} else {
		// Add the new customer account
		if args.CompanyName == "" {
			err := JSONRPCErrorCodes[int(ERROR_MISSING_FIELD)]
			err.Data.(*JSONRPCErrorData).MissingField = "CompanyName"
			s.Logger.Log("err", fmt.Errorf("SetupCompanyAccount(): %v: Missing company name field", err.Error()))
			return &err
		}

		if err := s.Storage.AddCustomer(ctx, routing.Customer{
			Code: args.CompanyCode,
			Name: args.CompanyName,
		}); err != nil {
			s.Logger.Log("err", fmt.Errorf("SetupCompanyAccount(): %v: Failed to add new customer entry", err.Error()))
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
		s.Logger.Log("err", fmt.Errorf("SetupCompanyAccount() failed to update user company code: %v", err))
		err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
		return &err
	}

	if !middleware.VerifyAllRoles(r, middleware.AdminRole) {
		// Remove existing roles if there are any
		if err := s.UserManager.RemoveRoles(requestID, []*management.Role{
			s.RoleCache["Viewer"],
			s.RoleCache["Owner"],
		}...); err != nil {
			s.Logger.Log("err", fmt.Errorf("SetupCompanyAccount() failed to remove roles: %v", err))
			err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
			return &err
		}

		// Assign new roles (owner and viewer)
		if err := s.UserManager.AssignRoles(requestID, roles...); err != nil {
			s.Logger.Log("err", fmt.Errorf("SetupCompanyAccount() failed to assign user roles: %v", err))
			err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
			return &err
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
		s.Logger.Log("err", fmt.Errorf("UpdateAccountDetails(): %v", err.Error()))
		return &err
	}

	requestUser := r.Context().Value(middleware.Keys.UserKey)
	if requestUser == nil {
		err := JSONRPCErrorCodes[int(ERROR_JWT_PARSE_FAILURE)]
		s.Logger.Log("err", fmt.Errorf("UpdateAccountDetails(): %v", err.Error()))
		return &err
	}

	requestID, ok := requestUser.(*jwt.Token).Claims.(jwt.MapClaims)["sub"].(string)
	if !ok {
		err := JSONRPCErrorCodes[int(ERROR_JWT_PARSE_FAILURE)]
		s.Logger.Log("err", fmt.Errorf("UpdateAccountDetails(): %v: Failed to parse user ID", err.Error()))
		return &err
	}

	userAccount, err := s.UserManager.Read(requestID)
	if err != nil {
		s.Logger.Log("err", fmt.Errorf("UpdateAccountDetails(): %v: Failed to read user account", err.Error()))
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
		s.Logger.Log("err", fmt.Errorf("UpdateAccountDetails(): %v: Failed to update user account", err.Error()))
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
		s.Logger.Log("err", fmt.Errorf("ResetPasswordEmail(): %v: Email is required", err.Error()))
		return &err
	}

	userAccounts, err := s.UserManager.List(management.Query(fmt.Sprintf(`email:"%s"`, args.Email)))
	if err != nil || len(userAccounts.Users) != 1 {
		err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
		s.Logger.Log("err", fmt.Errorf("ResetPasswordEmail(): Failed to look up user account: %s", err.Error()))
		return &err
	}

	if err = s.AuthenticationClient.SendChangePasswordEmail(args.Email); err != nil {
		err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
		s.Logger.Log("err", fmt.Errorf("ResetPasswordEmail(): %v: Failed to send reset password email", err.Error()))
		return &err
	}

	return nil
}

type VerifyEmailArgs struct {
	UserID string `json:"user_id"`
}

type VerifyEmailReply struct{}

func (s *AuthService) ResendVerificationEmail(r *http.Request, args *VerifyEmailArgs, reply *VerifyEmailReply) error {
	if middleware.VerifyAllRoles(r, middleware.UnverifiedRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		s.Logger.Log("err", fmt.Errorf("VerifyEmailUrl(): %v:", err.Error()))
		return &err
	}

	job := &management.Job{
		UserID: &args.UserID,
	}

	err := s.JobManager.VerifyEmail(job)
	if err != nil {
		s.Logger.Log("err", fmt.Errorf("VerifyEmailUrl(): %v: Failed to generate verification email", err.Error()))
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
		s.Logger.Log("err", fmt.Errorf("UpdateAutoSignupDomains(): %v", err.Error()))
		return &err
	}
	customerCode, ok := r.Context().Value(middleware.Keys.CustomerKey).(string)
	if !ok {
		err := JSONRPCErrorCodes[int(ERROR_JWT_PARSE_FAILURE)]
		s.Logger.Log("err", fmt.Errorf("UpdateAutoSignupDomains(): %v: Failed to parse customer code", err.Error()))
		return &err
	}
	if customerCode == "" {
		err := JSONRPCErrorCodes[int(ERROR_JWT_PARSE_FAILURE)]
		s.Logger.Log("err", fmt.Errorf("UpdateAutoSignupDomains(): %v: Failed to parse customer code", err.Error()))
		return &err
	}
	ctx := context.Background()

	company, err := s.Storage.Customer(r.Context(), customerCode)
	if err != nil {
		s.Logger.Log("err", fmt.Errorf("UpdateAutoSignupDomains(): %v", err.Error()))
		err := JSONRPCErrorCodes[int(ERROR_STORAGE_FAILURE)]
		return &err
	}

	company.AutomaticSignInDomains = strings.Join(args.Domains, ", ")

	err = s.Storage.SetCustomer(ctx, company)
	if err != nil {
		s.Logger.Log("err", fmt.Errorf("UpdateAutoSignupDomains(): %v", err.Error()))
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
		s.Logger.Log("err", fmt.Errorf("CustomerSignedUpSlackNotification(): %v", err.Error()))
		return &err
	}

	if args.Email == "" {
		err := JSONRPCErrorCodes[int(ERROR_MISSING_FIELD)]
		err.Data.(*JSONRPCErrorData).MissingField = "Email"
		s.Logger.Log("err", fmt.Errorf("UserAccount(): %v: Email is required", err.Error()))
		return &err
	}

	message := fmt.Sprintf("%s signed up on the Portal! :tada:", args.Email)

	if err := s.SlackClient.SendInfo(message); err != nil {
		err := JSONRPCErrorCodes[int(ERROR_SLACK_FAILURE)]
		s.Logger.Log("err", fmt.Errorf("CustomerSignedUpSlackNotification(): %v: Email is required", err.Error()))
		return &err
	}
	return nil
}

func (s *AuthService) CustomerViewedTheDocsSlackNotification(r *http.Request, args *CustomerSlackNotification, reply *GenericSlackNotificationReply) error {
	if middleware.VerifyAllRoles(r, middleware.AnonymousRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		s.Logger.Log("err", fmt.Errorf("CustomerViewedTheDocsSlackNotification(): %v", err.Error()))
		return &err
	}

	if args.Email == "" {
		err := JSONRPCErrorCodes[int(ERROR_MISSING_FIELD)]
		err.Data.(*JSONRPCErrorData).MissingField = "Email"
		s.Logger.Log("err", fmt.Errorf("UserAccount(): %v: Email is required", err.Error()))
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
		s.Logger.Log("err", fmt.Errorf("CustomerViewedTheDocsSlackNotification(): %v: Email is required", err.Error()))
		return &err
	}
	return nil
}

func (s *AuthService) CustomerDownloadedSDKSlackNotification(r *http.Request, args *CustomerSlackNotification, reply *GenericSlackNotificationReply) error {
	if middleware.VerifyAllRoles(r, middleware.AnonymousRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		s.Logger.Log("err", fmt.Errorf("CustomerDownloadedSDKSlackNotification(): %v", err.Error()))
		return &err
	}

	if args.Email == "" {
		err := JSONRPCErrorCodes[int(ERROR_MISSING_FIELD)]
		err.Data.(*JSONRPCErrorData).MissingField = "Email"
		s.Logger.Log("err", fmt.Errorf("UserAccount(): %v: Email is required", err.Error()))
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
		s.Logger.Log("err", fmt.Errorf("CustomerDownloadedSDKSlackNotification(): %v: Email is required", err.Error()))
		return &err
	}
	return nil
}

func (s *AuthService) CustomerEnteredPublicKeySlackNotification(r *http.Request, args *CustomerSlackNotification, reply *GenericSlackNotificationReply) error {
	if middleware.VerifyAllRoles(r, middleware.AnonymousRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		s.Logger.Log("err", fmt.Errorf("CustomerEnteredPublicKeySlackNotification(): %v", err.Error()))
		return &err
	}

	if args.Email == "" {
		err := JSONRPCErrorCodes[int(ERROR_MISSING_FIELD)]
		err.Data.(*JSONRPCErrorData).MissingField = "Email"
		s.Logger.Log("err", fmt.Errorf("UserAccount(): %v: Email is required", err.Error()))
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
		s.Logger.Log("err", fmt.Errorf("CustomerEnteredPublicKeySlackNotification(): %v: Email is required", err.Error()))
		return &err
	}
	return nil
}

func (s *AuthService) CustomerDownloadedUE4PluginNotifications(r *http.Request, args *CustomerSlackNotification, reply *GenericSlackNotificationReply) error {
	if middleware.VerifyAnyRole(r, middleware.AnonymousRole, middleware.UnverifiedRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		s.Logger.Log("err", fmt.Errorf("CustomerDownloadedUE4PluginNotifications(): %v", err.Error()))
		return &err
	}

	if args.Email == "" {
		err := JSONRPCErrorCodes[int(ERROR_MISSING_FIELD)]
		err.Data.(*JSONRPCErrorData).MissingField = "Email"
		s.Logger.Log("err", fmt.Errorf("CustomerDownloadedUE4PluginNotifications(): %v", err.Error()))
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
		s.Logger.Log("err", fmt.Errorf("CustomerDownloadedUE4PluginNotifications(): %v", err.Error()))
		return &err
	}
	return nil
}

func (s *AuthService) SlackNotification(r *http.Request, args *GenericSlackNotificationArgs, reply *GenericSlackNotificationReply) error {
	if !middleware.VerifyAllRoles(r, middleware.AdminRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		s.Logger.Log("err", fmt.Errorf("SlackNotification(): %v", err.Error()))
		return &err
	}

	if args.Message == "" {
		err := JSONRPCErrorCodes[int(ERROR_MISSING_FIELD)]
		err.Data.(*JSONRPCErrorData).MissingField = "Message"
		s.Logger.Log("err", fmt.Errorf("SlackNotification(): %v: Message is required", err.Error()))
		return &err
	}

	if args.Type == "" {
		err := JSONRPCErrorCodes[int(ERROR_MISSING_FIELD)]
		err.Data.(*JSONRPCErrorData).MissingField = "Type"
		s.Logger.Log("err", fmt.Errorf("SlackNotification(): %v: Type is required", err.Error()))
		return &err
	}

	switch args.Type {
	case "info":
		if err := s.SlackClient.SendInfo(args.Message); err != nil {
			err := JSONRPCErrorCodes[int(ERROR_SLACK_FAILURE)]
			s.Logger.Log("err", fmt.Errorf("CustomerSignedUpSlackNotification(): %v: Failed to send info Slack notification: %s", args.Message, err.Error()))
			return &err
		}
	case "warning":
		if err := s.SlackClient.SendWarning(args.Message); err != nil {
			err := JSONRPCErrorCodes[int(ERROR_SLACK_FAILURE)]
			s.Logger.Log("err", fmt.Errorf("CustomerSignedUpSlackNotification(): %v: Failed to send warning Slack notification: %s", args.Message, err.Error()))
			return &err
		}
	case "error":
		if err := s.SlackClient.SendError(args.Message); err != nil {
			err := JSONRPCErrorCodes[int(ERROR_SLACK_FAILURE)]
			s.Logger.Log("err", fmt.Errorf("CustomerSignedUpSlackNotification(): %v: Failed to send error Slack notification: %s", args.Message, err.Error()))
			return &err
		}
	default:
		err := JSONRPCErrorCodes[int(ERROR_ILLEGAL_OPERATION)]
		s.Logger.Log("err", fmt.Errorf("CustomerSignedUpSlackNotification(): %v: Slack notification type not supported: %s", args.Type, err.Error()))
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
	message := fmt.Sprintf("%s signed up on the Portal! :tada:\nCompany Name: %s\nCompany Website: %s", args.Email, args.CompanyName, args.CompanyWebsite)

	// If we error here we don't worry about it. Setting up everything is more important and the error shouldn't hinder the rest of the sign up process on the portal side
	if err := s.SlackClient.SendInfo(message); err != nil {
		s.Logger.Log("err", fmt.Errorf("ProcessNewSignup(): %v: Failed to send slack notification", err.Error()))
	}

	if s.HubSpotClient.APIKey != "" {
		companies, err := s.HubSpotClient.CompanyEntrySearch(args.CompanyName, args.CompanyWebsite)
		if err != nil {
			s.Logger.Log("err", fmt.Errorf("ProcessNewSignup(): %v: Failed to fetch company entries from hubspot", err.Error()))
		}

		foundCompanyID := ""
		if len(companies) != 0 {
			foundCompanyID = companies[0].ID
		}

		contacts, err := s.HubSpotClient.ContactEntrySearch(args.FirstName, args.LastName, args.Email)
		if err != nil {
			s.Logger.Log("err", fmt.Errorf("ProcessNewSignup(): %v: Failed to fetch contact entries from hubspot", err.Error()))
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
				s.Logger.Log("err", fmt.Errorf("ProcessNewSignup(): %v: Failed to add contact entry to hubspot", err.Error()))
			} else {
				if err := s.HubSpotClient.AssociateCompanyToContact(notifications.NewFunnelCoID, newContactID); err != nil {
					s.Logger.Log("err", fmt.Errorf("ProcessNewSignup(): %v: Failed to associate NewFunnelCo to new contact", err.Error()))
				}
				if err := s.HubSpotClient.AssociateContactToCompany(newContactID, notifications.NewFunnelCoID); err != nil {
					s.Logger.Log("err", fmt.Errorf("ProcessNewSignup(): %v: Failed to associate new contact to NewFunnelCo", err.Error()))
				}
				if err := s.HubSpotClient.CreateCompanyNote(message, notifications.NewFunnelCoID); err != nil {
					s.Logger.Log("err", fmt.Errorf("ProcessNewSignup(): %v: Failed to NewFunnelCo note", err.Error()))
				}
				if err := s.HubSpotClient.CreateContactNote(message, newContactID); err != nil {
					s.Logger.Log("err", fmt.Errorf("ProcessNewSignup(): %v: Failed to create new contact note", err.Error()))
				}
			}
		}

		// Company exists but contact does not
		if foundCompanyID != "" && foundContactID == "" {
			// Create contact and associate it to NewFunnelCo with notes about sign up
			newContactID, err := s.HubSpotClient.CreateNewContactEntry(args.FirstName, args.LastName, args.Email, args.CompanyName, args.CompanyWebsite)
			if err != nil {
				s.Logger.Log("err", fmt.Errorf("ProcessNewSignup(): %v: Failed to add contact entry to hubspot", err.Error()))
			} else {
				if err := s.HubSpotClient.AssociateCompanyToContact(foundCompanyID, newContactID); err != nil {
					s.Logger.Log("err", fmt.Errorf("ProcessNewSignup(): %v: Failed to associate found company to new contact", err.Error()))
				}
				if err := s.HubSpotClient.AssociateContactToCompany(newContactID, foundCompanyID); err != nil {
					s.Logger.Log("err", fmt.Errorf("ProcessNewSignup(): %v: Failed to associate new contact to found company", err.Error()))
				}
				if err := s.HubSpotClient.CreateCompanyNote(message, foundCompanyID); err != nil {
					s.Logger.Log("err", fmt.Errorf("ProcessNewSignup(): %v: Failed to create found company note", err.Error()))
				}
				if err := s.HubSpotClient.CreateContactNote(message, newContactID); err != nil {
					s.Logger.Log("err", fmt.Errorf("ProcessNewSignup(): %v: Failed to create new contact note", err.Error()))
				}
			}
		}

		// Company doesn't exist but contact does
		if foundCompanyID == "" && foundContactID != "" {
			// Associate contact with NewFunnelCo with note about signup with company name and website
			if err := s.HubSpotClient.AssociateCompanyToContact(notifications.NewFunnelCoID, foundContactID); err != nil {
				s.Logger.Log("err", fmt.Errorf("ProcessNewSignup(): %v: Failed to associate NewFunnelCo to found contact", err.Error()))
			}
			if err := s.HubSpotClient.AssociateContactToCompany(foundContactID, notifications.NewFunnelCoID); err != nil {
				s.Logger.Log("err", fmt.Errorf("ProcessNewSignup(): %v: Failed to associate found contact to NewFunnelCo", err.Error()))
			}
			if err := s.HubSpotClient.CreateCompanyNote(message, notifications.NewFunnelCoID); err != nil {
				s.Logger.Log("err", fmt.Errorf("ProcessNewSignup(): %v: Failed to NewFunnelCo note", err.Error()))
			}
			if err := s.HubSpotClient.CreateContactNote(message, foundContactID); err != nil {
				s.Logger.Log("err", fmt.Errorf("ProcessNewSignup(): %v: Failed to create found contact note", err.Error()))
			}
		}

		// Company and contact exist
		if foundCompanyID != "" && foundContactID != "" {
			// Associate the company and contact and make notes on both entries about the sign up
			if err := s.HubSpotClient.AssociateCompanyToContact(foundCompanyID, foundContactID); err != nil {
				s.Logger.Log("err", fmt.Errorf("ProcessNewSignup(): %v: Failed to associate found company to found contact", err.Error()))
			}
			if err := s.HubSpotClient.AssociateContactToCompany(foundContactID, foundCompanyID); err != nil {
				s.Logger.Log("err", fmt.Errorf("ProcessNewSignup(): %v: Failed to associate found contact to found company", err.Error()))
			}
			if err := s.HubSpotClient.CreateCompanyNote(message, foundCompanyID); err != nil {
				s.Logger.Log("err", fmt.Errorf("ProcessNewSignup(): %v: Failed to create found company note", err.Error()))
			}
			if err := s.HubSpotClient.CreateContactNote(message, foundContactID); err != nil {
				s.Logger.Log("err", fmt.Errorf("ProcessNewSignup(): %v: Failed to create found contact note", err.Error()))
			}
		}
	}

	userAccounts, err := s.UserManager.List(management.Query(fmt.Sprintf(`email:"%s"`, args.Email)))
	if err != nil || len(userAccounts.Users) != 1 {
		err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
		s.Logger.Log("err", fmt.Errorf("ProcessNewSignup(): Failed to look up user account: %s", err.Error()))
		return &err
	}

	user := userAccounts.Users[0]

	updateUser := &management.User{
		FamilyName: &args.LastName,
		GivenName:  &args.FirstName,
	}

	if err := s.UserManager.Update(user.GetID(), updateUser); err != nil {
		s.Logger.Log("err", fmt.Errorf("ProcessNewSignup(): %v: Failed to update new user's first and last name", err.Error()))
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
		s.Logger.Log("err", fmt.Errorf("RefreshAuthRolesCache(): Failed to refresh role caches: %s", err.Error()))
		return err
	}

	roles := roleList.Roles

	s.RoleCache = make(map[string]*management.Role)

	for _, r := range roles {
		s.RoleCache[r.GetName()] = r
	}

	return nil
}
