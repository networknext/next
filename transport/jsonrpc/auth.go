package jsonrpc

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
	"github.com/networknext/backend/transport"
	"github.com/rs/cors"
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
		err := fmt.Errorf("AllAccounts() CheckRoles error: %v", ErrInsufficientPrivileges)
		return err
	}

	reply.UserAccounts = make([]account, 0)
	accountList, err := s.UserManager.List()
	if err != nil {
		err := fmt.Errorf("AllAcounts() failed to fetch user list: %v", err)
		s.Logger.Log("err", err)
		return err
	}

	requestUser := r.Context().Value(Keys.UserKey)
	if requestUser == nil {
		err = fmt.Errorf("AllAcounts() unable to parse user from token")
		s.Logger.Log("err", err)
		return err
	}

	requestEmail, ok := requestUser.(*jwt.Token).Claims.(jwt.MapClaims)["email"].(string)
	if !ok {
		err := fmt.Errorf("AllAcounts() unable to parse email from token")
		s.Logger.Log("err", err)
		return err
	}

	requestCompany, ok := r.Context().Value(Keys.CompanyKey).(string)
	if !ok {
		return fmt.Errorf("AllAcounts(): user is not assigned to a company")
	}

	for _, a := range accountList.Users {
		if requestEmail == *a.Email {
			continue
		}
		companyCode, ok := a.AppMetadata["company_code"].(string)
		if !ok || requestCompany != companyCode {
			continue
		}
		userRoles, err := s.UserManager.Roles(*a.ID)
		if err != nil {
			err = fmt.Errorf("AllAcounts() failed to fetch user roles: %v", err)
			s.Logger.Log("err", err)
			return err
		}

		buyer, _ := s.Storage.BuyerWithCompanyCode(companyCode)
		company, err := s.Storage.Customer(companyCode)
		if err != nil {
			err = fmt.Errorf("AllAcounts() failed to fetch company: %v", err)
			s.Logger.Log("err", err)
			return err
		}

		reply.UserAccounts = append(reply.UserAccounts, newAccount(a, userRoles.Roles, buyer, company.Name))
	}

	return nil
}

func (s *AuthService) UserAccount(r *http.Request, args *AccountArgs, reply *AccountReply) error {
	if args.UserID == "" {
		err := fmt.Errorf("UserAccount() user_id is required")
		s.Logger.Log("err", err)
		return err
	}

	// Check if this is for authed user profile or other users

	user := r.Context().Value(Keys.UserKey)
	if user == nil {
		err := fmt.Errorf("UserAccount() failed to fetch calling user from token")
		s.Logger.Log("err", err)
		return err
	}

	claims := user.(*jwt.Token).Claims.(jwt.MapClaims)
	requestID, ok := claims["sub"]
	if !ok {
		err := fmt.Errorf("UserAccount(): failed to parse user id from token")
		s.Logger.Log("err", err)
		return err
	}
	if requestID != args.UserID && !VerifyAnyRole(r, AdminRole, OwnerRole) {
		err := fmt.Errorf("UserAccount(): %v", ErrInsufficientPrivileges)
		s.Logger.Log("err", err)
		return err
	}

	userAccount, err := s.UserManager.Read(args.UserID)
	if err != nil {
		err := fmt.Errorf("UserAccount() failed to fetch user account: %w", err)
		s.Logger.Log("err", err)
		return err
	}
	companyCode, ok := userAccount.AppMetadata["company_code"].(string)
	if !ok {
		companyCode = ""
	}
	var company routing.Customer
	if companyCode != "" {
		company, err = s.Storage.Customer(companyCode)
		if err != nil {
			err := fmt.Errorf("UserAccount() failed to fetch user company: %w", err)
			s.Logger.Log("err", err)
			return err
		}
	}
	buyer, err := s.Storage.BuyerWithCompanyCode(companyCode)
	userRoles, err := s.UserManager.Roles(*userAccount.ID)
	if err != nil {
		err := fmt.Errorf("UserAccount() failed to fetch user roles: %w", err)
		s.Logger.Log("err", err)
		return err
	}

	if VerifyAnyRole(r, AdminRole, OwnerRole) && requestID == args.UserID {
		reply.Domains = strings.Split(company.AutomaticSignInDomains, ",")
	}

	reply.UserAccount = newAccount(userAccount, userRoles.Roles, buyer, company.Name)

	return nil
}

func (s *AuthService) DeleteUserAccount(r *http.Request, args *AccountArgs, reply *AccountReply) error {
	if !VerifyAnyRole(r, AdminRole, OwnerRole) {
		err := fmt.Errorf("UserAccount(): %v", ErrInsufficientPrivileges)
		s.Logger.Log("err", err)
		return err
	}

	if args.UserID == "" {
		err := fmt.Errorf("DeleteUserAccount() user_id is required")
		s.Logger.Log("err", err)
		return err
	}
	user, err := s.UserManager.Read(args.UserID)
	if err != nil {
		err = fmt.Errorf("DeleteUserAccount() failed to read user account: %v", err)
		s.Logger.Log("err", err)
		return err
	}

	userCompanyCode, ok := user.AppMetadata["company_code"].(string)
	if !ok || userCompanyCode == "" {
		return nil
	}

	// Non admin trying to delete user from another company
	requestCompanyCode, ok := r.Context().Value(Keys.CompanyKey).(string)
	if (!ok || requestCompanyCode != userCompanyCode) && !VerifyAllRoles(r, AdminRole) {
		err := fmt.Errorf("UserAccount(): %v", ErrInsufficientPrivileges)
		s.Logger.Log("err", err)
		return err
	}

	user.AppMetadata["company_code"] = ""
	if err := s.UserManager.Update(args.UserID, user); err != nil {
		err = fmt.Errorf("DeleteUserAccount() failed to update user company code: %v", err)
		s.Logger.Log("err", err)
		return err
	}
	return nil
}

func (s *AuthService) AddUserAccount(req *http.Request, args *AccountsArgs, reply *AccountsReply) error {
	var adminString string = "Admin"
	var accounts []account
	var buyer routing.Buyer

	if !VerifyAnyRole(req, AdminRole, OwnerRole) {
		err := fmt.Errorf("UserAccount(): %v", ErrInsufficientPrivileges)
		s.Logger.Log("err", err)
		return err
	}

	if len(args.Roles) == 0 {
		err := fmt.Errorf("UserAccount(): roles are required")
		s.Logger.Log("err", err)
		return err
	}

	// Check if non admin is assigning admin role
	for _, r := range args.Roles {
		if r.Name == &adminString && !VerifyAllRoles(req, AdminRole) {
			err := fmt.Errorf("AddUserAccount() insufficient privileges")
			s.Logger.Log("err", err)
			return err
		}
	}

	// Gather request user information
	userCompanyCode, ok := req.Context().Value(Keys.CompanyKey).(string)
	if !ok || userCompanyCode == "" {
		err := fmt.Errorf("AddUserAccount() user is not assigned to a company")
		s.Logger.Log("err", err)
		return err
	}

	connectionID := "Username-Password-Authentication"
	emails := args.Emails
	falseValue := false

	buyer, err := s.Storage.BuyerWithCompanyCode(userCompanyCode)
	if err != nil {
		err := fmt.Errorf("AddUserAccount() failed to fetch request buyer: %v", err)
		s.Logger.Log("err", err)
		return err
	}

	registered := make(map[string]*management.User)
	accountList, err := s.UserManager.List()
	if err != nil {
		err := fmt.Errorf("AddUserAccount() failed to get auth0 account list: %v", err)
		s.Logger.Log("err", err)
		return err
	}
	emailString := strings.Join(emails, ",")

	for _, a := range accountList.Users {
		if strings.Contains(emailString, *a.Email) {
			registered[*a.Email] = a
		}
	}

	// Create an account for each new email
	var newUser *management.User
	var userID string
	for _, e := range emails {
		user, ok := registered[e]
		if !ok {
			pw, err := GenerateRandomString(32)
			if err != nil {
				err := fmt.Errorf("AddUserAccount() failed to generate a random password: %w", err)
				s.Logger.Log("err", err)
				return err
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
				err := fmt.Errorf("AddUserAccount() failed to create new user: %w", err)
				s.Logger.Log("err", err)
				return err
			}
			accountList, err := s.UserManager.List()
			if err != nil {
				err := fmt.Errorf("AddUserAccount() failed to get auth0 account list: %v", err)
				s.Logger.Log("err", err)
				return err
			}
			for _, a := range accountList.Users {
				if *a.Email == e {
					userID = *a.ID
					break
				}
			}
			if err = s.UserManager.AssignRoles(userID, args.Roles...); err != nil {
				err := fmt.Errorf("AddUserAccount() failed to add user roles: %w", err)
				s.Logger.Log("err", err)
				return err
			}
		} else {
			newUser = &management.User{
				Connection:    &connectionID,
				Email:         user.Email,
				EmailVerified: user.EmailVerified,
				VerifyEmail:   user.VerifyEmail,
				Password:      user.Password,
				AppMetadata: map[string]interface{}{
					"company_code": userCompanyCode,
				},
				Identities: user.Identities,
				Name:       user.Name,
			}
			if err = s.UserManager.Update(*user.ID, newUser); err != nil {
				err := fmt.Errorf("AddUserAccount() failed to update user: %w", err)
				s.Logger.Log("err", err)
				return err
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
				err := fmt.Errorf("AddUserAccount() failed to remove current user role: %w", err)
				s.Logger.Log("err", err)
				return err
			}
			if err = s.UserManager.AssignRoles(*user.ID, args.Roles...); err != nil {
				err := fmt.Errorf("AddUserAccount() failed to add user roles: %w", err)
				s.Logger.Log("err", err)
				return err
			}
		}

		company, err := s.Storage.Customer(userCompanyCode)
		if err != nil {
			err := fmt.Errorf("AddUserAccount() failed to fetch user company: %w", err)
			s.Logger.Log("err", err)
			return err
		}
		accounts = append(accounts, newAccount(newUser, args.Roles, buyer, company.Name))
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

func newAccount(u *management.User, r []*management.Role, buyer routing.Buyer, companyName string) account {
	buyerID := ""
	if buyer.ID != 0 {
		buyerID = fmt.Sprintf("%016x", buyer.ID)
	}

	account := account{
		UserID:      *u.Identities[0].UserID,
		ID:          buyerID,
		CompanyCode: buyer.CompanyCode,
		CompanyName: companyName,
		Name:        *u.Name,
		Email:       *u.Email,
		Roles:       r,
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
		err := fmt.Errorf("UserAccount(): %v", ErrInsufficientPrivileges)
		s.Logger.Log("err", err)
		return err
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
		err := fmt.Errorf("UserAccount(): %v", ErrInsufficientPrivileges)
		s.Logger.Log("err", err)
		return err
	}

	if args.UserID == "" {
		err := fmt.Errorf("UserRoles() user_id is required")
		s.Logger.Log("err", err)
		return err
	}

	userRoles, err := s.UserManager.Roles(args.UserID)
	if err != nil {
		err := fmt.Errorf("UserRoles() failed to get user roles: %w", err)
		s.Logger.Log("err", err)
		return err
	}

	reply.Roles = userRoles.Roles

	return nil
}

func (s *AuthService) UpdateUserRoles(r *http.Request, args *RolesArgs, reply *RolesReply) error {
	var err error
	if !VerifyAnyRole(r, AdminRole, OwnerRole) {
		err := fmt.Errorf("UpdateUserRoles(): %v", ErrInsufficientPrivileges)
		s.Logger.Log("err", err)
		return err
	}

	if args.UserID == "" {
		err := fmt.Errorf("UpdateUserRoles() user_id is required")
		s.Logger.Log("err", err)
		return err
	}

	userRoles, err := s.UserManager.Roles(args.UserID)
	if err != nil {
		err := fmt.Errorf("UpdateUserRoles() failed to fetch user roles: %w", err)
		s.Logger.Log("err", err)
		return err
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
				err := fmt.Errorf("UpdateUserRoles() failed to remove current user roles: %w", err)
				s.Logger.Log("err", err)
				return err
			}
		} else {
			err = s.UserManager.RemoveRoles(args.UserID, userRoles.Roles...)
			if err != nil {
				err := fmt.Errorf("UpdateUserRoles() failed to remove current user roles: %w", err)
				s.Logger.Log("err", err)
				return err
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
			err := fmt.Errorf("UpdateUserRoles(): %v", ErrInsufficientPrivileges)
			s.Logger.Log("err", err)
			return err
		}
	}

	err = s.UserManager.AssignRoles(args.UserID, args.Roles...)
	if err != nil {
		err := fmt.Errorf("UpdateUserRoles() failed to assign role: %w", err)
		s.Logger.Log("err", err)
		return err
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
		err := fmt.Errorf("UpdateCompanyInformation(): %v", ErrInsufficientPrivileges)
		s.Logger.Log("err", err)
		return err
	}

	newCompanyCode := args.CompanyCode

	if newCompanyCode == "" {
		err := fmt.Errorf("UpdateCompanyInformation() new company code is required")
		s.Logger.Log("err", err)
		return err
	}

	oldCompanyCode, ok := r.Context().Value(Keys.CompanyKey).(string)
	if !ok {
		oldCompanyCode = ""
	}

	// grab request user information
	requestUser := r.Context().Value(Keys.UserKey)
	if requestUser == nil {
		err := fmt.Errorf("UpdateCompanyInformation() unable to parse user from token")
		s.Logger.Log("err", err)
		return err
	}

	// get request user ID for role assignment
	requestID, ok := requestUser.(*jwt.Token).Claims.(jwt.MapClaims)["sub"].(string)
	if !ok {
		err := fmt.Errorf("UpdateCompanyInformation() unable to parse id from token")
		s.Logger.Log("err", err)
		return err
	}

	// parse request email for domain
	requestEmail, ok := requestUser.(*jwt.Token).Claims.(jwt.MapClaims)["email"].(string)
	if !ok {
		err := fmt.Errorf("UpdateCompanyInformation() unable to parse email from token")
		s.Logger.Log("err", err)
		return err
	}
	requestEmailParts := strings.Split(requestEmail, "@")
	requestDomain := requestEmailParts[len(requestEmailParts)-1] // Domain is the last entry of the split since an email as only one @ sign

	ctx := context.Background()

	if oldCompanyCode == "" {
		// Unassigned
		company, err := s.Storage.Customer(newCompanyCode)
		roles := []*management.Role{}
		if err != nil {
			// New Company
			if args.CompanyName == "" {
				err := fmt.Errorf("UpdateCompanyInformation() new company name is required")
				s.Logger.Log("err", err)
				return err
			}
			if err := s.Storage.AddCustomer(ctx, routing.Customer{
				Code: newCompanyCode,
				Name: args.CompanyName,
			}); err != nil {
				err = fmt.Errorf("UpdateCompanyInformation() failed to create new company: %v", err)
				s.Logger.Log("err", err)
				return err
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
				err = fmt.Errorf("UpdateCompanyInformation() email domain is not part of auto signup for this company: %v", err)
				s.Logger.Log("err", err)
				return err
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

	if oldCompanyCode != newCompanyCode {
		// Assigned and code is different
		if !VerifyAnyRole(r, AdminRole, OwnerRole) {
			// new company
			if args.CompanyName == "" {
				err := fmt.Errorf("UpdateCompanyInformation() new company name is required")
				s.Logger.Log("err", err)
				return err
			}
			if err := s.Storage.AddCustomer(ctx, routing.Customer{
				Code: newCompanyCode,
				Name: args.CompanyName,
			}); err != nil {
				err = fmt.Errorf("UpdateCompanyInformation() failed to create new company: %v", err)
				s.Logger.Log("err", err)
				return err
			}
			roles := []*management.Role{
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
			if err := s.UserManager.Update(requestID, &management.User{
				AppMetadata: map[string]interface{}{
					"company_code": args.CompanyCode,
				},
			}); err != nil {
				err = fmt.Errorf("UpdateCompanyInformation() failed to update user company code: %v", err)
				s.Logger.Log("err", err)
				return err
			}

			if !VerifyAllRoles(r, AdminRole) {
				if err := s.UserManager.RemoveRoles(requestID, []*management.Role{
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
				if err := s.UserManager.AssignRoles(requestID, roles...); err != nil {
					err := fmt.Errorf("UpdateCompanyInformation() failed to assign user roles: %w", err)
					s.Logger.Log("err", err)
					return err
				}
				reply.NewRoles = roles
			}
			return nil
		}

		_, err := s.Storage.Customer(newCompanyCode)
		if err == nil {
			err = fmt.Errorf("UpdateCompanyInformation() company already exists")
			s.Logger.Log("err", err)
			return err
		}

		oldCompany, err := s.Storage.Customer(oldCompanyCode)
		if err != nil {
			err = fmt.Errorf("UpdateCompanyInformation() failed to fetch company: %v", err)
			s.Logger.Log("err", err)
			return err
		}

		companyName := args.CompanyName
		if companyName == "" {
			companyName = oldCompany.Name
		}

		newCompany := routing.Customer{
			Code:                   newCompanyCode,
			Name:                   companyName,
			BuyerRef:               oldCompany.BuyerRef,
			SellerRef:              oldCompany.SellerRef,
			AutomaticSignInDomains: oldCompany.AutomaticSignInDomains,
			Active:                 oldCompany.Active,
		}

		buyer, err := s.Storage.BuyerWithCompanyCode(oldCompanyCode)
		if err == nil {
			buyer.CompanyCode = newCompanyCode
			if err := s.Storage.SetBuyer(ctx, buyer); err != nil {
				err = fmt.Errorf("UpdateCompanyInformation() failed to update buyer: %v", err)
				s.Logger.Log("err", err)
				return err
			}
		}
		seller, err := s.Storage.SellerWithCompanyCode(oldCompanyCode)
		if err == nil {
			seller.CompanyCode = newCompanyCode
			if err := s.Storage.SetSeller(ctx, seller); err != nil {
				err = fmt.Errorf("UpdateCompanyInformation() failed to update seller: %v", err)
				s.Logger.Log("err", err)
				return err
			}
		}
		if err := s.Storage.RemoveCustomer(ctx, oldCompanyCode); err != nil {
			err = fmt.Errorf("UpdateCompanyInformation() failed to remove old customer: %v", err)
			s.Logger.Log("err", err)
			return err
		}

		if err := s.Storage.AddCustomer(ctx, newCompany); err != nil {
			err = fmt.Errorf("UpdateCompanyInformation() failed to add new customer: %v", err)
			s.Logger.Log("err", err)
			return err
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
		return nil
	}

	if oldCompanyCode == newCompanyCode {
		// Assigned and code is the same
		if !VerifyAnyRole(r, AdminRole, OwnerRole) {
			err := fmt.Errorf("UpdateCompanyInformation(): %v", ErrInsufficientPrivileges)
			s.Logger.Log("err", err)
			return err
		}

		if args.CompanyName == "" {
			err := fmt.Errorf("UpdateCompanyInformation() new company code is required")
			s.Logger.Log("err", err)
			return err
		}

		company, err := s.Storage.Customer(oldCompanyCode)
		if err != nil {
			err = fmt.Errorf("UpdateCompanyInformation() failed to fetch company: %v", err)
			s.Logger.Log("err", err)
			return err
		}

		company.Name = args.CompanyName

		if err := s.Storage.SetCustomer(ctx, company); err != nil {
			err = fmt.Errorf("UpdateCompanyInformation() failed to update company: %v", err)
			s.Logger.Log("err", err)
			return err
		}
		return nil
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
		err := fmt.Errorf("UpdateCompanyInformation(): %v", ErrInsufficientPrivileges)
		s.Logger.Log("err", err)
		return err
	}

	requestUser := r.Context().Value(Keys.UserKey)
	if requestUser == nil {
		err := fmt.Errorf("UpdateAccountSettings() unable to parse user from token")
		s.Logger.Log("err", err)
		return err
	}

	requestID, ok := requestUser.(*jwt.Token).Claims.(jwt.MapClaims)["sub"].(string)
	if !ok {
		err := fmt.Errorf("UpdateAccountSettings() unable to parse id from token")
		s.Logger.Log("err", err)
		return err
	}

	userAccount, err := s.UserManager.Read(requestID)
	if err != nil {
		err := fmt.Errorf("UpdateAccountSettings() failed to fetch user account")
		s.Logger.Log("err", err)
		return err
	}

	if args.Password != "" {
		userAccount.Password = &args.Password
	}

	userAccount.AppMetadata = map[string]interface{}{
		"newsletter": args.NewsLetterConsent,
	}

	err = s.UserManager.Update(requestID, userAccount)
	if err != nil {
		err = fmt.Errorf("UpdateAccountSettings() failed to update user: %v", err)
		s.Logger.Log("err", err)
		return err
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
		err := fmt.Errorf("VerifyEmailUrl() failed to creating verification email link: %v", ErrInsufficientPrivileges)
		s.Logger.Log("err", err)
		return err
	}

	job := &management.Job{
		UserID: &args.UserID,
	}

	err := s.JobManager.VerifyEmail(job)
	if err != nil {
		err := fmt.Errorf("VerifyEmailUrl() failed to creating verification email link: %s", err)
		s.Logger.Log("err", err)
		return err
	}

	reply.Sent = true

	return nil
}

type AddContactArgs struct {
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
}

type UpdateDomainsArgs struct {
	Domains []string `json:"domains"`
}

type UpdateDomainsReply struct {
}

func (s *AuthService) UpdateAutoSignupDomains(r *http.Request, args *UpdateDomainsArgs, reply *UpdateDomainsReply) error {
	if !VerifyAnyRole(r, AdminRole, OwnerRole) {
		err := fmt.Errorf("UpdateAutoSignupDomains(): %v", ErrInsufficientPrivileges)
		s.Logger.Log("err", err)
		return err
	}

	companyCode, ok := r.Context().Value(Keys.CompanyKey).(string)
	if !ok {
		err := fmt.Errorf("UpdateAutoSignupDomains(): user is not assigned to a company")
		level.Error(s.Logger).Log("err", err)
		return err
	}
	if companyCode == "" {
		err := fmt.Errorf("UpdateAutoSignupDomains(): failed to parse company code")
		level.Error(s.Logger).Log("err", err)
		return err
	}
	ctx := context.Background()

	company, err := s.Storage.Customer(companyCode)
	if err != nil {
		err := fmt.Errorf("UpdateAutoSignupDomains(): failed to get request company")
		level.Error(s.Logger).Log("err", err)
		return err
	}

	company.AutomaticSignInDomains = strings.Join(args.Domains, ", ")

	err = s.Storage.SetCustomer(ctx, company)
	if err != nil {
		err := fmt.Errorf("UpdateAutoSignupDomains(): failed to update company")
		level.Error(s.Logger).Log("err", err)
		return err
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

func AuthMiddleware(audience string, next http.Handler, allowCORS bool) http.Handler {
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

	if !allowCORS {
		allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
		return cors.New(cors.Options{
			AllowedOrigins:   strings.Split(allowedOrigins, ","),
			AllowCredentials: true,
			AllowedHeaders:   []string{"Authorization", "Content-Type"},
			AllowedMethods:   []string{"POST", "GET", "OPTION"},
		}).Handler(mw.Handler(next))
	} else {
		return mw.Handler(next)
	}
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
