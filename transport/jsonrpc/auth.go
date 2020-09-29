package jsonrpc

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/networknext/backend/admin"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
	"gopkg.in/auth0.v4/management"
)

type AuthService struct {
	Auth0   storage.Auth0
	Storage storage.Storer
	Logger  log.Logger
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

func (s *AuthService) AllAccounts(r *http.Request, args *AccountsArgs, reply *AccountsReply) error {
	var accountList *management.UserList

	if !admin.VerifyAnyRole(r, admin.AdminRole, admin.OwnerRole) {
		err := fmt.Errorf("AllAccounts() CheckRoles error: %v", admin.ErrInsufficientPrivileges)
		return err
	}

	reply.UserAccounts = make([]account, 0)
	accountList, err := s.Auth0.Manager.User.List()
	if err != nil {
		err := fmt.Errorf("AllAccounts() failed to fetch user list: %v", err)
		s.Logger.Log("err", err)
		return err
	}

	requestUser, err := admin.RequestUser(r)
	if requestUser == nil || err != nil {
		err = fmt.Errorf("AllAccounts() unable to parse user from token")
		s.Logger.Log("err", err)
		return err
	}

	requestEmail, err := admin.RequestEmail(requestUser)
	if err != nil || requestEmail == "" {
		err := fmt.Errorf("AllAccounts() unable to parse email from token")
		s.Logger.Log("err", err)
		return err
	}

	requestCompany, err := admin.RequestCompany(r)
	if err == nil || requestCompany == "" {
		err := fmt.Errorf("AllAccounts(): failed to get company from context")
		s.Logger.Log("err", err)
		return err
	}

	for _, a := range accountList.Users {
		if requestEmail == *a.Email {
			continue
		}
		companyCode, ok := a.AppMetadata["company_code"].(string)
		if !ok || requestCompany != companyCode {
			continue
		}
		userRoles, err := s.Auth0.Manager.User.Roles(*a.ID)
		if err != nil {
			err = fmt.Errorf("AllAccounts() failed to fetch user roles: %v", err)
			s.Logger.Log("err", err)
			return err
		}

		buyer, _ := s.Storage.BuyerWithCompanyCode(companyCode)
		company, err := s.Storage.Customer(companyCode)
		if err != nil {
			err = fmt.Errorf("AllAccounts() failed to fetch company: %v", err)
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

	user := r.Context().Value("user")
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
	if requestID != args.UserID && !admin.VerifyAnyRole(r, admin.AdminRole, admin.OwnerRole) {
		err := fmt.Errorf("UserAccount(): %v", admin.ErrInsufficientPrivileges)
		s.Logger.Log("err", err)
		return err
	}

	userAccount, err := s.Auth0.Manager.User.Read(args.UserID)
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
	userRoles, err := s.Auth0.Manager.User.Roles(*userAccount.ID)
	if err != nil {
		err := fmt.Errorf("UserAccount() failed to fetch user roles: %w", err)
		s.Logger.Log("err", err)
		return err
	}

	if admin.VerifyAnyRole(r, admin.AdminRole, admin.OwnerRole) {
		reply.Domains = strings.Split(company.AutomaticSignInDomains, ",")
	}

	reply.UserAccount = newAccount(userAccount, userRoles.Roles, buyer, company.Name)

	return nil
}

func (s *AuthService) DeleteUserAccount(r *http.Request, args *AccountArgs, reply *AccountReply) error {
	if !admin.VerifyAnyRole(r, admin.AdminRole, admin.OwnerRole) {
		err := fmt.Errorf("UserAccount(): %v", admin.ErrInsufficientPrivileges)
		s.Logger.Log("err", err)
		return err
	}

	if args.UserID == "" {
		err := fmt.Errorf("DeleteUserAccount() user_id is required")
		s.Logger.Log("err", err)
		return err
	}
	if err := s.Auth0.Manager.User.Update(args.UserID, &management.User{
		AppMetadata: map[string]interface{}{
			"company_code": "",
		},
	}); err != nil {
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

	if !admin.VerifyAnyRole(req, admin.AdminRole, admin.OwnerRole) {
		err := fmt.Errorf("UserAccount(): %v", admin.ErrInsufficientPrivileges)
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
		if r.Name == &adminString && !admin.VerifyAllRoles(req, admin.AdminRole) {
			err := fmt.Errorf("AddUserAccount() insufficient privileges")
			s.Logger.Log("err", err)
			return err
		}
	}

	// Gather request user information
	requestUser := req.Context().Value("user")
	if requestUser == nil {
		err := fmt.Errorf("AddUserAccount() unable to parse user from token")
		s.Logger.Log("err", err)
		return err
	}

	requestID, ok := requestUser.(*jwt.Token).Claims.(jwt.MapClaims)["sub"].(string)
	if !ok {
		err := fmt.Errorf("AddUserAccount() unable to parse id from token")
		s.Logger.Log("err", err)
		return err
	}

	userAccount, err := s.Auth0.Manager.User.Read(requestID)
	if err != nil {
		err := fmt.Errorf("AddUserAccount() failed to fetch user account: %w", err)
		s.Logger.Log("err", err)
		return err
	}

	userCompanyCode, ok := userAccount.AppMetadata["company_code"].(string)
	if !ok {
		err := fmt.Errorf("AddUserAccount() user is not assigned to a company: %w", err)
		s.Logger.Log("err", err)
		return err
	}

	connectionID := "Username-Password-Authentication"
	emails := args.Emails
	falseValue := false

	for _, b := range s.Storage.Buyers() {
		if b.CompanyCode == userCompanyCode {
			buyer = b
		}
	}

	registered := make(map[string]*management.User)
	accountList, err := s.Auth0.Manager.User.List()
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
			if err = s.Auth0.Manager.User.Create(newUser); err != nil {
				err := fmt.Errorf("AddUserAccount() failed to create new user: %w", err)
				s.Logger.Log("err", err)
				return err
			}
			accountList, err := s.Auth0.Manager.User.List()
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
			if err = s.Auth0.Manager.User.AssignRoles(userID, args.Roles...); err != nil {
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
			}
			if err = s.Auth0.Manager.User.Update(*user.ID, newUser); err != nil {
				err := fmt.Errorf("AddUserAccount() failed to update user: %w", err)
				s.Logger.Log("err", err)
				return err
			}
			roles := []*management.Role{
				{
					ID:          &admin.RoleNames[0],
					Name:        &admin.RoleTypes[0],
					Description: &admin.RoleDescriptions[0],
				},
				{
					ID:          &admin.RoleNames[1],
					Name:        &admin.RoleTypes[1],
					Description: &admin.RoleDescriptions[1],
				},
			}
			if err = s.Auth0.Manager.User.RemoveRoles(*user.ID, roles...); err != nil {
				err := fmt.Errorf("UpdateUserRoles() failed to remove current user role: %w", err)
				s.Logger.Log("err", err)
				return err
			}
			if err = s.Auth0.Manager.User.AssignRoles(*user.ID, args.Roles...); err != nil {
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

	if !admin.VerifyAnyRole(r, admin.AdminRole, admin.OwnerRole) {
		err := fmt.Errorf("UserAccount(): %v", admin.ErrInsufficientPrivileges)
		s.Logger.Log("err", err)
		return err
	}

	if admin.VerifyAllRoles(r, admin.AdminRole) {
		reply.Roles = []*management.Role{
			{
				ID:          &admin.RoleNames[0],
				Name:        &admin.RoleTypes[0],
				Description: &admin.RoleDescriptions[0],
			},
			{
				ID:          &admin.RoleNames[1],
				Name:        &admin.RoleTypes[1],
				Description: &admin.RoleDescriptions[1],
			},
			{
				ID:          &admin.RoleNames[2],
				Name:        &admin.RoleTypes[2],
				Description: &admin.RoleDescriptions[2],
			},
		}
	} else {
		reply.Roles = []*management.Role{
			{
				ID:          &admin.RoleNames[0],
				Name:        &admin.RoleTypes[0],
				Description: &admin.RoleDescriptions[0],
			},
			{
				ID:          &admin.RoleNames[1],
				Name:        &admin.RoleTypes[1],
				Description: &admin.RoleDescriptions[1],
			},
		}
	}

	return nil
}

func (s *AuthService) UserRoles(r *http.Request, args *RolesArgs, reply *RolesReply) error {
	if !admin.VerifyAnyRole(r, admin.AdminRole, admin.OwnerRole) {
		err := fmt.Errorf("UserAccount(): %v", admin.ErrInsufficientPrivileges)
		s.Logger.Log("err", err)
		return err
	}

	if args.UserID == "" {
		err := fmt.Errorf("UserRoles() user_id is required")
		s.Logger.Log("err", err)
		return err
	}

	userRoles, err := s.Auth0.Manager.User.Roles(args.UserID)

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
	if !admin.VerifyAnyRole(r, admin.AdminRole, admin.OwnerRole) {
		err := fmt.Errorf("UserAccount(): %v", admin.ErrInsufficientPrivileges)
		s.Logger.Log("err", err)
		return err
	}

	if args.UserID == "" {
		err := fmt.Errorf("UpdateUserRoles() user_id is required")
		s.Logger.Log("err", err)
		return err
	}

	userRoles, err := s.Auth0.Manager.User.Roles(args.UserID)
	if err != nil {
		err := fmt.Errorf("UpdateUserRoles() failed to fetch user roles: %w", err)
		s.Logger.Log("err", err)
		return err
	}

	removeRoles := []*management.Role{
		{
			ID:          &admin.RoleNames[0],
			Name:        &admin.RoleTypes[0],
			Description: &admin.RoleDescriptions[0],
		},
		{
			ID:          &admin.RoleNames[1],
			Name:        &admin.RoleTypes[1],
			Description: &admin.RoleDescriptions[1],
		},
	}

	// Need all this for admins that accidently delete admin role and for tests
	found := false

	for _, role := range userRoles.Roles {
		if found {
			continue
		}
		if *role.Name == "Admin" {
			found = true
		}
	}

	if found {
		err = s.Auth0.Manager.User.RemoveRoles(args.UserID, removeRoles...)
		if err != nil {
			err := fmt.Errorf("UpdateUserRoles() failed to remove current user role: %w", err)
			s.Logger.Log("err", err)
			return err
		}
	} else {
		err = s.Auth0.Manager.User.RemoveRoles(args.UserID, userRoles.Roles...)
		if err != nil {
			err := fmt.Errorf("UpdateUserRoles() failed to remove current user role: %w", err)
			s.Logger.Log("err", err)
			return err
		}
	}

	if len(args.Roles) == 0 {
		reply.Roles = make([]*management.Role, 0)
		return nil
	}

	err = s.Auth0.Manager.User.AssignRoles(args.UserID, args.Roles...)
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
	if admin.VerifyAnyRole(r, admin.AnonymousRole, admin.UnverifiedRole) {
		return nil
	}

	newCompanyCode := args.CompanyCode

	if newCompanyCode == "" {
		err := fmt.Errorf("UpdateCompanyInformation() new company code is required")
		s.Logger.Log("err", err)
		return err
	}

	oldCompanyCode, ok := r.Context().Value(admin.Keys.CompanyKey).(string)
	if !ok {
		oldCompanyCode = ""
	}

	// grab request user information
	requestUser := r.Context().Value("user")
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
					ID:          &admin.RoleNames[0],
					Name:        &admin.RoleTypes[0],
					Description: &admin.RoleDescriptions[0],
				},
				{
					ID:          &admin.RoleNames[1],
					Name:        &admin.RoleTypes[1],
					Description: &admin.RoleDescriptions[1],
				},
			}
		} else {
			// Old Company
			autoSigninDomains := company.AutomaticSignInDomains
			// the company exists and the new user is part of the auto signup
			if strings.Contains(autoSigninDomains, requestDomain) {
				roles = []*management.Role{
					{
						ID:          &admin.RoleNames[0],
						Name:        &admin.RoleTypes[0],
						Description: &admin.RoleDescriptions[0],
					},
				}
			} else {
				// the company exists and the new user is not part of the auto signup
				err = fmt.Errorf("UpdateCompanyInformation() email domain is not part of auto signup for this company: %v", err)
				s.Logger.Log("err", err)
				return err
			}
		}
		if err = s.Auth0.Manager.User.Update(requestID, &management.User{
			AppMetadata: map[string]interface{}{
				"company_code": args.CompanyCode,
			},
		}); err != nil {
			err = fmt.Errorf("UpdateCompanyInformation() failed to update user company name: %v", err)
			s.Logger.Log("err", err)
			return err
		}
		if !admin.VerifyAllRoles(r, admin.AdminRole) {
			if err = s.Auth0.Manager.User.AssignRoles(requestID, roles...); err != nil {
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
		if !admin.VerifyAnyRole(r, admin.AdminRole, admin.OwnerRole) {
			err := fmt.Errorf("UpdateCompanyInformation(): %v", admin.ErrInsufficientPrivileges)
			s.Logger.Log("err", err)
			return err
		}

		_, err := s.Storage.Customer(newCompanyCode)
		if err == nil {
			err = fmt.Errorf("UpdateCompanyInformation() company already exists: %v", err)
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
		if err = s.Auth0.Manager.User.Update(requestID, &management.User{
			AppMetadata: map[string]interface{}{
				"company_code": args.CompanyCode,
			},
		}); err != nil {
			err = fmt.Errorf("UpdateCompanyInformation() failed to update user company name: %v", err)
			s.Logger.Log("err", err)
			return err
		}
		return nil
	}

	if oldCompanyCode == newCompanyCode {
		// Assigned and code is the same
		if !admin.VerifyAnyRole(r, admin.AdminRole, admin.OwnerRole) {
			err := fmt.Errorf("UpdateCompanyInformation(): %v", admin.ErrInsufficientPrivileges)
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
	if admin.VerifyAnyRole(r, admin.AnonymousRole, admin.UnverifiedRole) {
		return nil
	}

	var updateUser management.User

	requestUser := r.Context().Value("user")
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

	if args.Password != "" {
		updateUser.Password = &args.Password
	}

	updateUser.AppMetadata = map[string]interface{}{
		"newsletter": args.NewsLetterConsent,
	}

	err := s.Auth0.Manager.User.Update(requestID, &updateUser)
	if err != nil {
		err = fmt.Errorf("UpdateAccountSettings() failed to update user password: %v", err)
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

	if !admin.VerifyAllRoles(r, admin.UnverifiedRole) {
		err := fmt.Errorf("VerifyEmailUrl() failed to creating verification email link: %v", admin.ErrInsufficientPrivileges)
		s.Logger.Log("err", err)
		return err
	}

	job := &management.Job{
		UserID: &args.UserID,
	}

	err := s.Auth0.Manager.Job.VerifyEmail(job)
	if err != nil {
		err := fmt.Errorf("VerifyEmailUrl() failed to creating verification email link: %s", err)
		s.Logger.Log("err", err)
		return err
	}

	reply.Sent = true

	return nil
}

type UpdateDomainsArgs struct {
	Domains []string `json:"domains"`
}

type UpdateDomainsReply struct {
}

func (s *AuthService) UpdateAutoSignupDomains(r *http.Request, args *UpdateDomainsArgs, reply *UpdateDomainsReply) error {
	if !admin.VerifyAnyRole(r, admin.AdminRole, admin.OwnerRole) {
		err := fmt.Errorf("UpdateAutoSignupDomains(): %v", admin.ErrInsufficientPrivileges)
		s.Logger.Log("err", err)
		return err
	}

	companyCode, ok := r.Context().Value(admin.Keys.CompanyKey).(string)
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
