package jsonrpc

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-kit/kit/log"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/storage"
	"github.com/networknext/backend/modules/transport"
	"github.com/networknext/backend/modules/transport/middleware"
	"gopkg.in/auth0.v4/management"
)

var (
	ErrInsufficientPrivileges = errors.New("insufficient privileges")
)

type contextKeys struct {
	AnonymousCallKey     string
	RolesKey             string
	CompanyKey           string
	NewsletterConsentKey string
	UserKey              string
}

var Keys contextKeys = contextKeys{
	AnonymousCallKey:     "anonymous",
	RolesKey:             "roles",
	CompanyKey:           "company",
	NewsletterConsentKey: "newsletter",
	UserKey:              "user",
}

type AuthService struct {
	MailChimpManager transport.MailChimpHandler
	UserManager      storage.UserManager
	JobManager       storage.JobManager
	Storage          storage.Storer
	Logger           log.Logger
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
}

type account struct {
	UserID      string             `json:"user_id"`
	ID          string             `json:"id"`
	CompanyName string             `json:"company_name"`
	CompanyCode string             `json:"company_code"`
	Name        string             `json:"name"`
	Email       string             `json:"email"`
	Roles       []*management.Role `json:"roles"`
	CreatedAt   time.Time          `json:"created_at"`
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
	var accountList *management.UserList

	if !VerifyAnyRole(r, AdminRole, OwnerRole) {
		err := JSONRPCErrorCodes[ERROR_INSUFFICIENT_PRIVILEGES]
		s.Logger.Log("err", fmt.Errorf("AllAccounts(): %v", err.Error()))
		return &err
	}

	reply.UserAccounts = make([]account, 0)
	accountList, err := s.UserManager.List()
	if err != nil {
		err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
		s.Logger.Log("err", fmt.Errorf("AllAccounts(): %v: Failed to fetch user list", err.Error()))
		return &err
	}

	requestUser := r.Context().Value(Keys.UserKey)
	if requestUser == nil {
		err := JSONRPCErrorCodes[int(ERROR_JWT_PARSE_FAILURE)]
		s.Logger.Log("err", fmt.Errorf("AllAccounts(): %v: Failed to parse user", err.Error()))
		return &err
	}

	requestCompany, ok := r.Context().Value(Keys.CompanyKey).(string)
	if !ok {
		err := JSONRPCErrorCodes[int(ERROR_USER_IS_NOT_ASSIGNED)]
		s.Logger.Log("err", fmt.Errorf("AllAccounts(): %v", err.Error()))
		return &err
	}

	for _, a := range accountList.Users {
		companyCode, ok := a.AppMetadata["company_code"].(string)
		if !ok || requestCompany != companyCode {
			continue
		}
		userRoles, err := s.UserManager.Roles(*a.ID)
		if err != nil {
			err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
			s.Logger.Log("err", fmt.Errorf("AllAccounts(): %v: Failed to get user roles", err.Error()))
			return &err
		}

		buyer, _ := s.Storage.BuyerWithCompanyCode(companyCode)
		company, err := s.Storage.Customer(companyCode)
		if err != nil {
			err := JSONRPCErrorCodes[int(ERROR_STORAGE_FAILURE)]
			s.Logger.Log("err", fmt.Errorf("AllAccounts(): %v", err.Error()))
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

	user := r.Context().Value(Keys.UserKey)
	if user == nil {
		err := JSONRPCErrorCodes[int(ERROR_JWT_PARSE_FAILURE)]
		s.Logger.Log("err", fmt.Errorf("UserAccount(): %v", err.Error()))
		return &err
	}

	claims := user.(*jwt.Token).Claims.(jwt.MapClaims)
	requestID, ok := claims["sub"]
	if !ok {
		err := JSONRPCErrorCodes[int(ERROR_JWT_PARSE_FAILURE)]
		s.Logger.Log("err", fmt.Errorf("UserAccount(): %v: Failed to parse user ID", err.Error()))
		return &err
	}
	if requestID != args.UserID && !VerifyAnyRole(r, AdminRole, OwnerRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		s.Logger.Log("err", fmt.Errorf("UserAccount(): %v", err.Error()))
		return &err
	}

	userAccount, err := s.UserManager.Read(args.UserID)
	if err != nil {
		err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
		s.Logger.Log("err", fmt.Errorf("UserAccount(): %v: Failed to get user account details", err.Error()))
		return &err
	}
	companyCode, ok := userAccount.AppMetadata["company_code"].(string)
	if !ok {
		companyCode = ""
	}
	var company routing.Customer
	if companyCode != "" {
		company, err = s.Storage.Customer(companyCode)
		if err != nil {
			err := JSONRPCErrorCodes[int(ERROR_STORAGE_FAILURE)]
			s.Logger.Log("err", fmt.Errorf("UserAccount(): %v: Could not find customer account for customer code: %v", err.Error(), companyCode))
			return &err
		}
	}
	buyer, err := s.Storage.BuyerWithCompanyCode(companyCode)
	userRoles, err := s.UserManager.Roles(*userAccount.ID)
	if err != nil {
		err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
		s.Logger.Log("err", fmt.Errorf("UserAccount(): %v: Failed to get user account roles", err.Error()))
		return &err
	}

	if VerifyAnyRole(r, AdminRole, OwnerRole) && requestID == args.UserID {
		reply.Domains = strings.Split(company.AutomaticSignInDomains, ",")
	}

	reply.UserAccount = newAccount(userAccount, userRoles.Roles, buyer, company.Name, company.Code)

	return nil
}

func (s *AuthService) DeleteUserAccount(r *http.Request, args *AccountArgs, reply *AccountReply) error {
	if !VerifyAnyRole(r, AdminRole, OwnerRole) {
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
		err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
		s.Logger.Log("err", fmt.Errorf("DeleteUserAccount(): %v: Failed to read user account", err.Error()))
		return &err
	}

	userCompanyCode, ok := user.AppMetadata["company_code"].(string)
	if !ok || userCompanyCode == "" {
		return nil
	}

	// Non admin trying to delete user from another company
	requestCompanyCode, ok := r.Context().Value(Keys.CompanyKey).(string)
	if (!ok || requestCompanyCode != userCompanyCode) && !VerifyAllRoles(r, AdminRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		s.Logger.Log("err", fmt.Errorf("DeleteUserAccount(): %v", err.Error()))
		return &err
	}

	if err := s.UserManager.Update(args.UserID, &management.User{
		AppMetadata: map[string]interface{}{
			"company_code": "",
		},
	}); err != nil {
		err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
		s.Logger.Log("err", fmt.Errorf("DeleteUserAccount(): %v: Failed to update deleted user company code", err.Error()))
		return &err
	}
	return nil
}

func (s *AuthService) AddUserAccount(req *http.Request, args *AccountsArgs, reply *AccountsReply) error {
	var adminString string = "Admin"
	var accounts []account

	if !VerifyAnyRole(req, AdminRole, OwnerRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		s.Logger.Log("err", fmt.Errorf("AddUserAccount(): %v", err.Error()))
		return &err
	}

	// Check if non admin is assigning admin role
	for _, r := range args.Roles {
		if r.Name == &adminString && !VerifyAllRoles(req, AdminRole) {
			err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
			s.Logger.Log("err", fmt.Errorf("AddUserAccount(): %v", err.Error()))
			return &err
		}
	}

	// Gather request user information
	userCompanyCode, ok := req.Context().Value(Keys.CompanyKey).(string)
	if !ok || userCompanyCode == "" {
		err := JSONRPCErrorCodes[int(ERROR_USER_IS_NOT_ASSIGNED)]
		s.Logger.Log("err", fmt.Errorf("AddUserAccount(): %v", err.Error()))
		return &err
	}

	connectionID := "Username-Password-Authentication"
	emails := args.Emails
	falseValue := false

	buyer, _ := s.Storage.BuyerWithCompanyCode(userCompanyCode)

	registered := make(map[string]*management.User)
	accountList, err := s.UserManager.List()
	if err != nil {
		err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
		s.Logger.Log("err", fmt.Errorf("AddUserAccount(): %v: Failed to get user list", err.Error()))
		return &err
	}
	emailString := strings.Join(emails, ",")

	for _, a := range accountList.Users {
		if strings.Contains(emailString, *a.Email) {
			registered[*a.Email] = a
		}
	}
	currentTime := time.Now()

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
				CreatedAt: &currentTime,
			}
			if err = s.UserManager.Create(newUser); err != nil {
				err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
				s.Logger.Log("err", fmt.Errorf("AddUserAccount(): %v: Failed to create new user account", err.Error()))
				return &err
			}
			if len(args.Roles) > 0 {
				if err = s.UserManager.AssignRoles(*newUser.ID, args.Roles...); err != nil {
					err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
					s.Logger.Log("err", fmt.Errorf("AddUserAccount(): %v: Failed to assign new user roles", err.Error()))
					return &err
				}
			}
		} else {
			newUser = &management.User{
				AppMetadata: map[string]interface{}{
					"company_code": userCompanyCode,
				},
			}
			if err = s.UserManager.Update(*user.ID, newUser); err != nil {
				err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
				s.Logger.Log("err", fmt.Errorf("AddUserAccount(): %v: Failed to update user account", err.Error()))
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
			if err = s.UserManager.RemoveRoles(*user.ID, roles...); err != nil {
				err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
				s.Logger.Log("err", fmt.Errorf("AddUserAccount(): %v: Failed to remove exist roles from user account", err.Error()))
				return &err
			}
			if len(args.Roles) > 0 {
				if err = s.UserManager.AssignRoles(*user.ID, args.Roles...); err != nil {
					err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
					s.Logger.Log("err", fmt.Errorf("AddUserAccount(): %v: Failed to assign new roles to user account", err.Error()))
					return &err
				}
			}
		}

		company, err := s.Storage.Customer(userCompanyCode)
		if err != nil {
			err := JSONRPCErrorCodes[int(ERROR_STORAGE_FAILURE)]
			s.Logger.Log("err", fmt.Errorf("AddUserAccount(): %v", err.Error()))
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
		CreatedAt:   *u.CreatedAt,
	}

	return account
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

	if !VerifyAnyRole(r, AdminRole, OwnerRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		s.Logger.Log("err", fmt.Errorf("AllRoles(): %v", err.Error()))
		return &err
	}

	if VerifyAllRoles(r, AdminRole) {
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
	if !VerifyAnyRole(r, AdminRole, OwnerRole) {
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
		err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
		s.Logger.Log("err", fmt.Errorf("UserRoles(): %v: Failed to fetch user roles", err.Error()))
		return &err
	}

	reply.Roles = userRoles.Roles

	return nil
}

func (s *AuthService) UpdateUserRoles(r *http.Request, args *RolesArgs, reply *RolesReply) error {
	var err error
	if !VerifyAnyRole(r, AdminRole, OwnerRole) {
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
		err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
		s.Logger.Log("err", fmt.Errorf("UpdateUserRoles(): %v: failed to fetch user roles", err.Error()))
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
		if VerifyAllRoles(r, AdminRole) {
			err = s.UserManager.RemoveRoles(args.UserID, removeRoles...)
			if err != nil {
				err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
				s.Logger.Log("err", fmt.Errorf("UpdateUserRoles(): %v: Failed to remove old user roles", err.Error()))
				return &err
			}
		} else {
			err = s.UserManager.RemoveRoles(args.UserID, userRoles.Roles...)
			if err != nil {
				err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
				s.Logger.Log("err", fmt.Errorf("UpdateUserRoles(): %v: Failed to remove old user roles", err.Error()))
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
		if *role.Name == "Admin" && !VerifyAllRoles(r, AdminRole) {
			err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
			s.Logger.Log("err", fmt.Errorf("UpdateUserRoles(): %v", err.Error()))
			return &err
		}
	}

	err = s.UserManager.AssignRoles(args.UserID, args.Roles...)
	if err != nil {
		err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
		s.Logger.Log("err", fmt.Errorf("UpdateUserRoles(): %v: Failed to assign user roles", err.Error()))
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
	if VerifyAnyRole(r, AnonymousRole, UnverifiedRole) {
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

	assignedCustomerCode, ok := r.Context().Value(Keys.CompanyKey).(string)
	if !ok {
		assignedCustomerCode = ""
	}

	if assignedCustomerCode != "" {
		err := JSONRPCErrorCodes[int(ERROR_ILLEGAL_OPERATION)]
		s.Logger.Log("err", fmt.Errorf("UpdateCompanyInformation(): %v: User is already assigned to the customer account: %v", err.Error(), assignedCustomerCode))
		return &err
	}

	// grab request user information
	requestUser := r.Context().Value(Keys.UserKey)
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

	company, err := s.Storage.Customer(newCompanyCode)
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
			err := JSONRPCErrorCodes[int(ERROR_STORAGE_FAILURE)]
			s.Logger.Log("err", fmt.Errorf("UpdateCompanyInformation(): %v: Failed to add new customer entry", err.Error()))
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
		err = fmt.Errorf("UpdateCompanyInformation() failed to update user company code: %v", err)
		s.Logger.Log("err", err)
		return err
	}

	if !VerifyAllRoles(r, AdminRole) {
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
			err := fmt.Errorf("UpdateCompanyInformation() failed to remove roles: %w", err)
			s.Logger.Log("err", err)
			return err
		}
		if err = s.UserManager.AssignRoles(requestID, roles...); err != nil {
			err := fmt.Errorf("UpdateCompanyInformation() failed to assign user roles: %w", err)
			s.Logger.Log("err", err)
			return err
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
	if VerifyAnyRole(r, AnonymousRole, UnverifiedRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		s.Logger.Log("err", fmt.Errorf("UpdateAccountSettings(): %v", err.Error()))
		return &err
	}

	requestUser := r.Context().Value(Keys.UserKey)
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
		err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
		s.Logger.Log("err", fmt.Errorf("UpdateAccountSettings(): %v: Failed to read user account", err.Error()))
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
		err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
		s.Logger.Log("err", fmt.Errorf("UpdateAccountSettings(): %v: Failed to update user account", err.Error()))
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

	if !VerifyAllRoles(r, UnverifiedRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		s.Logger.Log("err", fmt.Errorf("VerifyEmailUrl(): %v: Failed to read user account", err.Error()))
		return &err
	}

	job := &management.Job{
		UserID: &args.UserID,
	}

	err := s.JobManager.VerifyEmail(job)
	if err != nil {
		err := JSONRPCErrorCodes[int(ERROR_AUTH0_FAILURE)]
		s.Logger.Log("err", fmt.Errorf("VerifyEmailUrl(): %v: Failed to generate verification link", err.Error()))
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
	if !VerifyAnyRole(r, AdminRole, OwnerRole) {
		err := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		s.Logger.Log("err", fmt.Errorf("UpdateAutoSignupDomains(): %v", err.Error()))
		return &err
	}

	customerCode, ok := r.Context().Value(Keys.CompanyKey).(string)
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

	company, err := s.Storage.Customer(customerCode)
	if err != nil {
		err := JSONRPCErrorCodes[int(ERROR_STORAGE_FAILURE)]
		s.Logger.Log("err", fmt.Errorf("UpdateAutoSignupDomains(): %v", err.Error()))
		return &err
	}

	company.AutomaticSignInDomains = strings.Join(args.Domains, ", ")

	err = s.Storage.SetCustomer(ctx, company)
	if err != nil {
		err := JSONRPCErrorCodes[int(ERROR_STORAGE_FAILURE)]
		s.Logger.Log("err", fmt.Errorf("UpdateAutoSignupDomains(): %v", err.Error()))
		return &err
	}

	return nil
}

type response struct {
	Message string `json:"message"`
}

type jwks struct {
	Keys []struct {
		Kty string   `json:"kty"`
		Kid string   `json:"kid"`
		Use string   `json:"use"`
		N   string   `json:"n"`
		E   string   `json:"e"`
		X5c []string `json:"x5c"`
	} `json:"keys"`
}

func AuthMiddleware(audience string, next http.Handler, allowCORS bool, allowedOrigins []string) http.Handler {
	if audience == "" {
		return next
	}

	mw := jwtmiddleware.New(jwtmiddleware.Options{
		UserProperty: Keys.UserKey,
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			// Check if OpsService token
			claims := token.Claims.(jwt.MapClaims)

			if _, ok := claims["scope"]; !ok {
				if !claims.VerifyAudience(audience, false) {
					return token, errors.New("Invalid audience.")
				}
			}
			// Verify 'iss' claim
			iss := "https://networknext.auth0.com/"
			checkIss := token.Claims.(jwt.MapClaims).VerifyIssuer(iss, false)
			if !checkIss {
				return token, errors.New("Invalid issuer.")
			}

			cert, err := getPemCert(token)
			if err != nil {
				return nil, err
			}

			return jwt.ParseRSAPublicKeyFromPEM([]byte(cert))
		},
		SigningMethod:       jwt.SigningMethodRS256,
		CredentialsOptional: true,
	})

	return middleware.CORSControlHandler(allowCORS, allowedOrigins, mw.Handler(next))
}

func getPemCert(token *jwt.Token) (string, error) {
	cert := ""
	resp, err := http.Get("https://networknext.auth0.com/.well-known/jwks.json")

	if err != nil {
		return cert, err
	}
	defer resp.Body.Close()

	var keys = jwks{}
	err = json.NewDecoder(resp.Body).Decode(&keys)

	if err != nil {
		return cert, err
	}

	for k := range keys.Keys {
		if token.Header["kid"] == keys.Keys[k].Kid {
			cert = "-----BEGIN CERTIFICATE-----\n" + keys.Keys[k].X5c[0] + "\n-----END CERTIFICATE-----"
		}
	}

	if cert == "" {
		err := errors.New("Unable to find appropriate key.")
		return cert, err
	}

	return cert, nil
}

func SetIsAnonymous(r *http.Request, value bool) *http.Request {
	ctx := r.Context()
	ctx = context.WithValue(ctx, Keys.AnonymousCallKey, value)
	return r.WithContext(ctx)
}

func IsAnonymous(r *http.Request) bool {
	anon, ok := r.Context().Value(Keys.AnonymousCallKey).(bool)
	return ok && anon
}

func AddTokenContext(r *http.Request, roles []string, companyCode string, newsletterConsent bool) *http.Request {
	ctx := r.Context()
	if len(roles) > 0 {
		ctx = context.WithValue(ctx, Keys.RolesKey, roles)
	}
	if companyCode != "" {
		ctx = context.WithValue(ctx, Keys.CompanyKey, companyCode)
	}
	ctx = context.WithValue(ctx, Keys.NewsletterConsentKey, newsletterConsent)
	return r.WithContext(ctx)
}

// RoleFunc defines a function that takes in an http.Request and perform a check on it whether it has a role or not.
type RoleFunc func(req *http.Request) (bool, error)

// Ops checks the request for the appropriate "scope" in the JWT
var OpsRole = func(req *http.Request) (bool, error) {
	user := req.Context().Value(Keys.UserKey)

	if user != nil {
		claims := user.(*jwt.Token).Claims.(jwt.MapClaims)

		if _, ok := claims["scope"]; ok {
			return true, nil
		}
	}
	return false, fmt.Errorf("OpsRole(): failed to fetch user from token")
}

var AdminRole = func(req *http.Request) (bool, error) {
	requestRoles := req.Context().Value(Keys.RolesKey)
	if requestRoles == nil {
		return false, fmt.Errorf("AdminRole(): failed to get roles from context")
	}

	found := false

	for _, role := range requestRoles.([]string) {
		if found {
			continue
		}
		if role == "Admin" {
			found = true
		}
	}
	return found, nil
}

var OwnerRole = func(req *http.Request) (bool, error) {
	requestRoles := req.Context().Value(Keys.RolesKey)

	if requestRoles == nil {
		return false, fmt.Errorf("OwnerRole(): failed to get roles from context")
	}

	found := false

	for _, role := range requestRoles.([]string) {
		if found {
			continue
		}
		if role == "Owner" {
			found = true
		}
	}
	return found, nil
}

// Ops checks the request for the appropriate "scope" in the JWT
var AnonymousRole = func(req *http.Request) (bool, error) {
	anon, ok := req.Context().Value(Keys.AnonymousCallKey).(bool)
	return ok && anon, nil
}

// Ops checks the request for the appropriate "scope" in the JWT
var UnverifiedRole = func(req *http.Request) (bool, error) {
	user := req.Context().Value(Keys.UserKey)

	if user == nil {
		return false, fmt.Errorf("UnverifiedRole(): failed to fetch user from token")
	}
	claims, ok := user.(*jwt.Token).Claims.(jwt.MapClaims)

	if !ok {
		return false, fmt.Errorf("UnverifiedRole(): failed to fetch verified claim")
	}

	if verified, ok := claims["email_verified"]; ok && !verified.(bool) {
		return true, nil
	}
	return false, nil
}

func VerifyAllRoles(req *http.Request, roleFuncs ...RoleFunc) bool {
	for _, f := range roleFuncs {
		authorized, err := f(req)
		if !authorized || err != nil {
			return false
		}
	}
	return true
}

func VerifyAnyRole(req *http.Request, roleFuncs ...RoleFunc) bool {
	for _, f := range roleFuncs {
		authorized, err := f(req)
		if authorized && err == nil {
			return true
		}
	}
	return false
}
