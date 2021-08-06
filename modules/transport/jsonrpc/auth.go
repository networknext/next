package jsonrpc

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-kit/kit/log"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/storage"
	"github.com/networknext/backend/modules/transport/middleware"
	"github.com/networknext/backend/modules/transport/notifications"
	"gopkg.in/auth0.v4/management"
)

type AuthService struct {
	MailChimpManager notifications.MailChimpHandler
	UserManager      storage.UserManager
	JobManager       storage.JobManager
	SlackClient      notifications.SlackClient
	Storage          storage.Storer
	Logger           log.Logger
	LookerSecret     string
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
	UserID      string             `json:"user_id"`
	ID          string             `json:"id"`
	CompanyName string             `json:"company_name"`
	CompanyCode string             `json:"company_code"`
	Name        string             `json:"name"`
	Email       string             `json:"email"`
	Roles       []*management.Role `json:"roles"`
}

var roleIDs []string = []string{
	"rol_ScQpWhLvmTKRlqLU",
	"rol_8r0281hf2oC4cvrD",
	"rol_YfFrtom32or4vH89",
}
var roleNames []string = []string{
	"Viewer",
	"Owner",
	"Admin",
}
var roleDescriptions []string = []string{
	"Can see current sessions and the map.",
	"Can access and manage everything in an account.",
	"Can manage the Network Next system, including access to configstore.",
}

func (s *AuthService) AllAccounts(r *http.Request, args *AccountsArgs, reply *AccountsReply) error {
	var totalUsers []*management.User

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

	requestUser := r.Context().Value(middleware.Keys.UserKey)
	if requestUser == nil {
		err := JSONRPCErrorCodes[int(ERROR_JWT_PARSE_FAILURE)]
		s.Logger.Log("err", fmt.Errorf("AllAccounts(): %v: Failed to parse user", err.Error()))
		return &err
	}

	requestCompany, ok := r.Context().Value(middleware.Keys.CompanyKey).(string)
	if !ok {
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
		company, err := s.Storage.Customer(r.Context(), companyCode)
		if err != nil {
			s.Logger.Log("err", fmt.Errorf("AllAccounts(): %v", err.Error()))
			err := JSONRPCErrorCodes[int(ERROR_STORAGE_FAILURE)]
			return &err
		}

		reply.UserAccounts = append(reply.UserAccounts, newAccount(a, userRoles.Roles, buyer, company.Name, company.Code))
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
	buyer, err := s.Storage.BuyerWithCompanyCode(r.Context(), companyCode)
	userRoles, err := s.UserManager.Roles(*userAccount.ID)
	if err != nil {
		s.Logger.Log("err", fmt.Errorf("UserAccount(): %v: Failed to get user account roles", err.Error()))
		err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
		return &err
	}

	if middleware.VerifyAnyRole(r, middleware.AdminRole, middleware.OwnerRole) && requestID == args.UserID {
		reply.Domains = strings.Split(company.AutomaticSignInDomains, ",")
	}

	reply.UserAccount = newAccount(userAccount, userRoles.Roles, buyer, company.Name, company.Code)

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
	requestCompanyCode, ok := r.Context().Value(middleware.Keys.CompanyKey).(string)
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
	userCompanyCode, ok := r.Context().Value(middleware.Keys.CompanyKey).(string)
	if !ok || userCompanyCode == "" {
		err := JSONRPCErrorCodes[int(ERROR_USER_IS_NOT_ASSIGNED)]
		s.Logger.Log("err", fmt.Errorf("AddUserAccount(): %v", err.Error()))
		return &err
	}

	connectionID := "Username-Password-Authentication"
	emails := args.Emails
	falseValue := false

	buyer, _ := s.Storage.BuyerWithCompanyCode(r.Context(), userCompanyCode)

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
				AppMetadata: map[string]interface{}{
					"company_code": userCompanyCode,
				},
			}
			if err := s.UserManager.Update(*user.ID, newUser); err != nil {
				s.Logger.Log("err", fmt.Errorf("AddUserAccount(): %v: Failed to update user account", err.Error()))
				err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
				return &err
			}
			roles := []*management.Role{
				{
					ID:          &roleIDs[0],
					Name:        &roleNames[0],
					Description: &roleDescriptions[0],
				},
				{
					ID:          &roleIDs[1],
					Name:        &roleNames[1],
					Description: &roleDescriptions[1],
				},
			}
			if err := s.UserManager.RemoveRoles(*user.ID, roles...); err != nil {
				s.Logger.Log("err", fmt.Errorf("AddUserAccount(): %v: Failed to remove exist roles from user account", err.Error()))
				err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
				return &err
			}
			if len(args.Roles) > 0 {
				if err := s.UserManager.AssignRoles(*user.ID, args.Roles...); err != nil {
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
		accounts = append(accounts, newAccount(newUser, args.Roles, buyer, company.Name, company.Code))
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

func newAccount(u *management.User, r []*management.Role, buyer routing.Buyer, companyName string, companyCode string) account {
	buyerID := ""
	if buyer.ID != 0 {
		buyerID = fmt.Sprintf("%016x", buyer.ID)
	}

	account := account{
		UserID:      *u.Identities[0].UserID,
		ID:          buyerID,
		CompanyCode: companyCode,
		CompanyName: companyName,
		Name:        *u.Name,
		Email:       *u.Email,
		Roles:       r,
	}

	return account
}

type databaseEntry struct {
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

		userRoles, err := s.UserManager.Roles(*account.ID)
		if err != nil {
			s.Logger.Log("err", fmt.Errorf("AllAccounts(): %v: Failed to get user roles", err.Error()))
			err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
			return &err
		}

		isOwner := false
		for _, role := range userRoles.Roles {
			// Check if the role is an owner role
			if *role.ID == roleIDs[1] {
				isOwner = true
				break
			}
		}

		entry := databaseEntry{
			Email:        *account.Email,
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
			{
				ID:          &roleIDs[0],
				Name:        &roleNames[0],
				Description: &roleDescriptions[0],
			},
			{
				ID:          &roleIDs[1],
				Name:        &roleNames[1],
				Description: &roleDescriptions[1],
			},
			{
				ID:          &roleIDs[2],
				Name:        &roleNames[2],
				Description: &roleDescriptions[2],
			},
		}
	} else {
		reply.Roles = []*management.Role{
			{
				ID:          &roleIDs[0],
				Name:        &roleNames[0],
				Description: &roleDescriptions[0],
			},
			{
				ID:          &roleIDs[1],
				Name:        &roleNames[1],
				Description: &roleDescriptions[1],
			},
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
		{
			Name:        &roleNames[0],
			ID:          &roleIDs[0],
			Description: &roleDescriptions[0],
		},
		{
			Name:        &roleNames[1],
			ID:          &roleIDs[1],
			Description: &roleDescriptions[1],
		},
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

type CompanyNameArgs struct {
	CompanyName string `json:"company_name"`
	CompanyCode string `json:"company_code"`
}

type CompanyNameReply struct {
	NewRoles []*management.Role `json:"new_roles"`
}

func (s *AuthService) UpdateCompanyInformation(r *http.Request, args *CompanyNameArgs, reply *CompanyNameReply) error {
	if middleware.VerifyAnyRole(r, middleware.AnonymousRole, middleware.UnverifiedRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		s.Logger.Log("err", fmt.Errorf("UpdateCompanyInformation(): %v", err.Error()))
		return &err
	}

	newCompanyCode := args.CompanyCode

	if newCompanyCode == "" {
		err := JSONRPCErrorCodes[int(ERROR_MISSING_FIELD)]
		err.Data.(*JSONRPCErrorData).MissingField = "CompanyCode"
		s.Logger.Log("err", fmt.Errorf("UpdateCompanyInformation(): %v: missing CompanyCode", err.Error()))
		return &err
	}

	assignedCustomerCode, ok := r.Context().Value(middleware.Keys.CompanyKey).(string)
	if !ok {
		assignedCustomerCode = ""
	}

	if assignedCustomerCode != "" {
		err := JSONRPCErrorCodes[int(ERROR_ILLEGAL_OPERATION)]
		s.Logger.Log("err", fmt.Errorf("UpdateCompanyInformation(): %v: User is already assigned to the customer account: %v", err.Error(), assignedCustomerCode))
		return &err
	}

	// grab request user information
	requestUser := r.Context().Value(middleware.Keys.UserKey)
	if requestUser == nil {
		err := JSONRPCErrorCodes[int(ERROR_JWT_PARSE_FAILURE)]
		s.Logger.Log("err", fmt.Errorf("UpdateCompanyInformation(): %v", err.Error()))
		return &err
	}

	// get request user ID for role assignment
	requestID, ok := requestUser.(*jwt.Token).Claims.(jwt.MapClaims)["sub"].(string)
	if !ok {
		err := JSONRPCErrorCodes[int(ERROR_JWT_PARSE_FAILURE)]
		s.Logger.Log("err", fmt.Errorf("UpdateCompanyInformation(): %v: Failed to parse user ID", err.Error()))
		return &err
	}

	ctx := context.Background()

	company, err := s.Storage.Customer(r.Context(), newCompanyCode)
	roles := []*management.Role{}
	if err != nil {
		// New Company
		if args.CompanyName == "" {
			err := JSONRPCErrorCodes[int(ERROR_MISSING_FIELD)]
			err.Data.(*JSONRPCErrorData).MissingField = "CompanyName"
			s.Logger.Log("err", fmt.Errorf("UpdateCompanyInformation(): %v: Missing company name field", err.Error()))
			return &err
		}
		if err := s.Storage.AddCustomer(ctx, routing.Customer{
			Code: newCompanyCode,
			Name: args.CompanyName,
		}); err != nil {
			s.Logger.Log("err", fmt.Errorf("UpdateCompanyInformation(): %v: Failed to add new customer entry", err.Error()))
			err := JSONRPCErrorCodes[int(ERROR_STORAGE_FAILURE)]
			return &err
		}
		roles = []*management.Role{
			{
				Name:        &roleNames[0],
				ID:          &roleIDs[0],
				Description: &roleDescriptions[0],
			},
			{
				Name:        &roleNames[1],
				ID:          &roleIDs[1],
				Description: &roleDescriptions[1],
			},
		}
	} else {
		// Old Company
		requestEmail, ok := requestUser.(*jwt.Token).Claims.(jwt.MapClaims)["email"].(string)
		if !ok {
			err := JSONRPCErrorCodes[int(ERROR_JWT_PARSE_FAILURE)]
			s.Logger.Log("err", fmt.Errorf("UpdateCompanyInformation(): %v: Failed to parse email", err.Error()))
			return &err
		}
		requestEmailParts := strings.Split(requestEmail, "@")
		requestDomain := requestEmailParts[len(requestEmailParts)-1] // Domain is the last entry of the split since an email as only one @ sign
		autoSigninDomains := company.AutomaticSignInDomains

		// the company exists and the new user is part of the auto signup
		if strings.Contains(autoSigninDomains, requestDomain) {
			roles = []*management.Role{
				{
					Name:        &roleNames[0],
					ID:          &roleIDs[0],
					Description: &roleDescriptions[0],
				},
			}
		} else {
			// the company exists and the new user is not part of the auto signup
			err := JSONRPCErrorCodes[int(ERROR_ILLEGAL_OPERATION)]
			s.Logger.Log("err", fmt.Errorf("UpdateCompanyInformation(): %v: User email was not found in acceptable domains", err.Error()))
			return &err
		}
	}

	if err = s.UserManager.Update(requestID, &management.User{
		AppMetadata: map[string]interface{}{
			"company_code": args.CompanyCode,
		},
	}); err != nil {
		s.Logger.Log("err", fmt.Errorf("UpdateCompanyInformation() failed to update user company code: %v", err))
		err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
		return &err
	}

	if !middleware.VerifyAllRoles(r, middleware.AdminRole) {
		if err = s.UserManager.RemoveRoles(requestID, []*management.Role{
			{
				Name:        &roleNames[0],
				ID:          &roleIDs[0],
				Description: &roleDescriptions[0],
			},
			{
				Name:        &roleNames[1],
				ID:          &roleIDs[1],
				Description: &roleDescriptions[1],
			},
		}...); err != nil {
			s.Logger.Log("err", fmt.Errorf("UpdateCompanyInformation() failed to remove roles: %v", err))
			err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
			return &err
		}
		if err = s.UserManager.AssignRoles(requestID, roles...); err != nil {
			s.Logger.Log("err", fmt.Errorf("UpdateCompanyInformation() failed to assign user roles: %v", err))
			err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
			return &err
		}
		reply.NewRoles = roles
	}
	return nil
}

type AccountSettingsArgs struct {
	Password          string `json:"password"`
	NewsLetterConsent bool   `json:"newsletter"`
}

type AccountSettingsReply struct {
}

func (s *AuthService) UpdateAccountSettings(r *http.Request, args *AccountSettingsArgs, reply *AccountSettingsReply) error {
	if middleware.VerifyAnyRole(r, middleware.AnonymousRole, middleware.UnverifiedRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		s.Logger.Log("err", fmt.Errorf("UpdateAccountSettings(): %v", err.Error()))
		return &err
	}

	requestUser := r.Context().Value(middleware.Keys.UserKey)
	if requestUser == nil {
		err := JSONRPCErrorCodes[int(ERROR_JWT_PARSE_FAILURE)]
		s.Logger.Log("err", fmt.Errorf("UpdateAccountSettings(): %v", err.Error()))
		return &err
	}

	requestID, ok := requestUser.(*jwt.Token).Claims.(jwt.MapClaims)["sub"].(string)
	if !ok {
		err := JSONRPCErrorCodes[int(ERROR_JWT_PARSE_FAILURE)]
		s.Logger.Log("err", fmt.Errorf("UpdateAccountSettings(): %v: Failed to parse user ID", err.Error()))
		return &err
	}

	userAccount, err := s.UserManager.Read(requestID)
	if err != nil {
		s.Logger.Log("err", fmt.Errorf("UpdateAccountSettings(): %v: Failed to read user account", err.Error()))
		err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
		return &err
	}

	if args.Password != "" {
		userAccount.Password = &args.Password
	}

	err = s.UserManager.Update(requestID, &management.User{
		AppMetadata: map[string]interface{}{
			"newsletter": args.NewsLetterConsent,
		},
	})
	if err != nil {
		s.Logger.Log("err", fmt.Errorf("UpdateAccountSettings(): %v: Failed to update user account", err.Error()))
		err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
		return &err
	}

	return nil
}

type VerifyEmailArgs struct {
	UserID string `json:"user_id"`
}

type VerifyEmailReply struct {
	Sent bool `json:"sent"`
}

func (s *AuthService) ResendVerificationEmail(r *http.Request, args *VerifyEmailArgs, reply *VerifyEmailReply) error {
	reply.Sent = false

	if !middleware.VerifyAllRoles(r, middleware.UnverifiedRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		s.Logger.Log("err", fmt.Errorf("VerifyEmailUrl(): %v: Failed to read user account", err.Error()))
		return &err
	}

	job := &management.Job{
		UserID: &args.UserID,
	}

	err := s.JobManager.VerifyEmail(job)
	if err != nil {
		s.Logger.Log("err", fmt.Errorf("VerifyEmailUrl(): %v: Failed to generate verification link", err.Error()))
		err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
		return &err
	}

	reply.Sent = true

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
	customerCode, ok := r.Context().Value(middleware.Keys.CompanyKey).(string)
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
	// TODO: update this in the hubspot PR

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
	// TODO: update this in the hubspot PR

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
	// TODO: update this in the hubspot PR

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

	// TODO: update this in the hubspot PR
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
