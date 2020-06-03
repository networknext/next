package jsonrpc

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
	"gopkg.in/auth0.v4/management"
)

type contextType string

const (
	anonymousCallKey contextType = "anonymous"
	rolesKey         contextType = "roles"
)

type AuthService struct {
	Auth0   storage.Auth0
	Storage storage.Storer
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
	UserAccount account `json:"account"`
}

type account struct {
	UserID      string             `json:"user_id"`
	ID          string             `json:"id"`
	CompanyName string             `json:"company_name"`
	Name        string             `json:"name"`
	Email       string             `json:"email"`
	Roles       []*management.Role `json:"roles"`
}

func (s *AuthService) AllAccounts(r *http.Request, args *AccountsArgs, reply *AccountsReply) error {
	var accountList *management.UserList
	if IsAnonymous(r) {
		return fmt.Errorf("insufficient privileges")
	}
	_, err := CheckRoles(r, "Admin")
	if err != nil {
		if _, err := CheckRoles(r, "Owner"); err != nil {
			return err
		}
	}

	accountList, err = s.Auth0.Manager.User.List()

	reply.UserAccounts = make([]account, 0)

	if err != nil {
		return fmt.Errorf("failed to fetch user list: %w", err)
	}
	requestUser := r.Context().Value("user")
	if _, ok := requestUser.(string); !ok {
		return fmt.Errorf("unable to parse user from token")
	}

	requestEmail, ok := requestUser.(*jwt.Token).Claims.(jwt.MapClaims)["email"].(string)
	if !ok {
		return fmt.Errorf("unable to parse email from token")
	}
	requestEmailParts := strings.Split(requestEmail, "@")
	requestDomain := requestEmailParts[len(requestEmailParts)-1] // Domain is the last entry of the split since an email as only one @ sign

	for _, a := range accountList.Users {
		emailParts := strings.Split(*a.Email, "@")
		if len(emailParts) <= 0 {
			return fmt.Errorf("failed to parse email %s for domain", *a.Email)
		}
		domain := emailParts[len(emailParts)-1] // Domain is the last entry of the split since an email as only one @ sign
		if requestDomain != domain {
			continue
		}
		userRoles, err := s.Auth0.Manager.User.Roles(*a.ID)
		if err != nil {
			return fmt.Errorf("failed to fetch user roles: %w", err)
		}

		buyer, err := s.Storage.BuyerWithDomain(domain)

		reply.UserAccounts = append(reply.UserAccounts, newAccount(a, userRoles.Roles, buyer))
	}

	return nil
}

func (s *AuthService) UserAccount(r *http.Request, args *AccountArgs, reply *AccountReply) error {
	if IsAnonymous(r) {
		return fmt.Errorf("insufficient privileges")
	}

	if args.UserID == "" {
		return fmt.Errorf("user_id is required")
	}

	// Check if this is for authed user profile or other users

	user := r.Context().Value("user")
	if user == nil {
		return fmt.Errorf("failed to fetch calling user from token")
	}

	claims := user.(*jwt.Token).Claims.(jwt.MapClaims)
	if requestID, ok := claims["sub"]; ok && requestID != args.UserID {
		// If they want a different user profile they need perms
		if _, err := CheckRoles(r, "Admin"); err != nil {
			if _, err := CheckRoles(r, "Owner"); err != nil {
				return err
			}
		}
	}

	userAccount, err := s.Auth0.Manager.User.Read(args.UserID)
	if err != nil {
		return fmt.Errorf("failed to fetch user account: %w", err)
	}

	emailParts := strings.Split(*userAccount.Email, "@")
	if len(emailParts) <= 0 {
		return fmt.Errorf("failed to parse email %s for domain", *userAccount.Email)
	}

	domain := emailParts[len(emailParts)-1] // Domain is the last entry of the split since an email as only one @ sign
	buyer, err := s.Storage.BuyerWithDomain(domain)
	userRoles, err := s.Auth0.Manager.User.Roles(*userAccount.ID)
	if err != nil {
		return fmt.Errorf("failed to fetch user roles: %w", err)
	}

	reply.UserAccount = newAccount(userAccount, userRoles.Roles, buyer)

	return nil
}

func (s *AuthService) DeleteUserAccount(r *http.Request, args *AccountArgs, reply *AccountReply) error {
	if IsAnonymous(r) {
		return fmt.Errorf("insufficient privileges")
	}

	if _, err := CheckRoles(r, "Admin"); err != nil {
		if _, err := CheckRoles(r, "Owner"); err != nil {
			return err
		}
	}

	if args.UserID == "" {
		return fmt.Errorf("user_id is required")
	}

	if err := s.Auth0.Manager.User.Delete(args.UserID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

func (s *AuthService) AddUserAccount(r *http.Request, args *AccountsArgs, reply *AccountsReply) error {
	if IsAnonymous(r) {
		return fmt.Errorf("insufficient privileges")
	}

	if _, err := CheckRoles(r, "Admin"); err != nil {
		if _, err := CheckRoles(r, "Owner"); err != nil {
			return err
		}
	}

	connectionID := "Username-Password-Authentication"
	emails := args.Emails
	roles := args.Roles
	falseValue := false

	var accounts []account

	newCustomerIDsMap := make(map[string]interface{})

	for _, e := range emails {

		emailParts := strings.Split(e, "@")
		if len(emailParts) <= 0 {
			return fmt.Errorf("failed to parse email %s for domain", e)
		}
		domain := emailParts[len(emailParts)-1] // Domain is the last entry of the split since an email as only one @ sign

		buyer, err := s.Storage.BuyerWithDomain(domain)

		pw, err := GenerateRandomString(32)
		if err != nil {
			return fmt.Errorf("failed to generate a random password: %w", err)
		}
		newUser := &management.User{
			Connection:    &connectionID,
			Email:         &e,
			EmailVerified: &falseValue,
			VerifyEmail:   &falseValue,
			Password:      &pw,
			AppMetadata: map[string]interface{}{
				"customer": newCustomerIDsMap,
			},
		}
		if err = s.Auth0.Manager.User.Create(newUser); err != nil {
			return fmt.Errorf("failed to create new user: %w", err)
		}

		if err = s.Auth0.Manager.User.AssignRoles(*newUser.ID, args.Roles...); err != nil {
			return fmt.Errorf("failed to add user roles: %w", err)
		}

		accounts = append(accounts, newAccount(newUser, roles, buyer))
	}
	reply.UserAccounts = accounts
	return nil
}

func GenerateRandomString(n int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"
	bytes, err := GenerateRandomBytes(n)
	if err != nil {
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
		return nil, err
	}

	return b, nil
}

func newAccount(u *management.User, r []*management.Role, buyer routing.Buyer) account {
	id := strconv.FormatUint(buyer.ID, 10)

	if id == "0" {
		id = ""
	}

	account := account{
		UserID:      *u.Identities[0].UserID,
		ID:          id,
		CompanyName: buyer.Name,
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
	if IsAnonymous(r) {
		return fmt.Errorf("insufficient privileges")
	}
	isAdmin, err := CheckRoles(r, "Admin")
	if err != nil {
		if _, err := CheckRoles(r, "Owner"); err != nil {
			return err
		}
	}

	roleList, err := s.Auth0.Manager.Role.List()
	if err != nil {
		fmt.Errorf("failed to fetch role list: %w", err)
	}

	if !isAdmin {
		for _, role := range roleList.Roles {
			if *role.Name != "Admin" {
				reply.Roles = append(reply.Roles, role)
			}
		}
	} else {
		reply.Roles = roleList.Roles
	}

	return nil
}

func (s *AuthService) UserRoles(r *http.Request, args *RolesArgs, reply *RolesReply) error {
	if IsAnonymous(r) {
		return fmt.Errorf("insufficient privileges")
	}

	if _, err := CheckRoles(r, "Admin"); err != nil {
		if _, err := CheckRoles(r, "Owner"); err != nil {
			return err
		}
	}

	if args.UserID == "" {
		return fmt.Errorf("user_id is required")
	}

	userRoles, err := s.Auth0.Manager.User.Roles(args.UserID)

	if err != nil {
		return fmt.Errorf("failed to get user roles: %w", err)
	}

	reply.Roles = userRoles.Roles

	return nil
}

func (s *AuthService) UpdateUserRoles(r *http.Request, args *RolesArgs, reply *RolesReply) error {
	var err error
	if IsAnonymous(r) {
		return fmt.Errorf("insufficient privileges")
	}

	if _, err := CheckRoles(r, "Admin"); err != nil {
		if _, err := CheckRoles(r, "Owner"); err != nil {
			return err
		}
	}

	if args.UserID == "" {
		return fmt.Errorf("user_id is required")
	}

	userRoles, err := s.Auth0.Manager.User.Roles(args.UserID)

	if err != nil {
		return fmt.Errorf("failed to fetch user roles: %w", err)
	}

	if len(userRoles.Roles) > 0 {
		err = s.Auth0.Manager.User.RemoveRoles(args.UserID, userRoles.Roles...)
	}
	if err != nil {
		return fmt.Errorf("failed to remove current user role: %w", err)
	}

	if len(args.Roles) == 0 {
		reply.Roles = make([]*management.Role, 0)
		return nil
	}

	err = s.Auth0.Manager.User.AssignRoles(args.UserID, args.Roles...)
	if err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	reply.Roles = args.Roles
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

func AuthMiddleware(audience string, next http.Handler) http.Handler {
	if audience == "" {
		return next
	}

	mw := jwtmiddleware.New(jwtmiddleware.Options{
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

	return mw.Handler(next)
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

	for k, _ := range keys.Keys {
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

func CheckRoles(r *http.Request, requiredRole string) (bool, error) {
	requestRoles := r.Context().Value(rolesKey)

	if requestRoles == nil {
		return false, fmt.Errorf("failed to get roles from context")
	}

	found := false

	for _, role := range requestRoles.(management.RoleList).Roles {
		if found {
			continue
		}
		if *role.Name == requiredRole {
			found = true
		}
	}
	if found {
		return true, nil
	}
	return false, fmt.Errorf("insufficient privileges")
}

func SetIsAnonymous(r *http.Request, value bool) *http.Request {
	ctx := r.Context()
	ctx = context.WithValue(ctx, anonymousCallKey, value)
	return r.WithContext(ctx)
}

func IsAnonymous(r *http.Request) bool {
	anon, ok := r.Context().Value(anonymousCallKey).(bool)
	return ok && anon
}

func SetRoles(r *http.Request, roles management.RoleList) *http.Request {
	ctx := r.Context()
	ctx = context.WithValue(ctx, rolesKey, roles)
	return r.WithContext(ctx)
}

func RequestRoles(r *http.Request) management.RoleList {
	roles := r.Context().Value(rolesKey)

	if roles != nil {
		return roles.(management.RoleList)
	}
	return management.RoleList{}
}
