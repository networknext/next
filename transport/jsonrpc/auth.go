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
	"github.com/go-kit/kit/log"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
	"gopkg.in/auth0.v4/management"
)

var (
	ErrInsufficientPrivileges = errors.New("insufficient privileges")
)

type contextType string

const (
	anonymousCallKey contextType = "anonymous"
	rolesKey         contextType = "roles"
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

	if !VerifyAnyRole(r, AdminRole, OwnerRole) {
		err := fmt.Errorf("AllAccounts() CheckRoles error: %v", ErrInsufficientPrivileges)
		return err
	}

	accountList, err := s.Auth0.Manager.User.List()

	reply.UserAccounts = make([]account, 0)

	if err != nil {
		err := fmt.Errorf("AllAcounts() failed to fetch user list: %v", err)
		s.Logger.Log("err", err)
		return err
	}
	requestUser := r.Context().Value("user")
	if requestUser == nil {
		err = fmt.Errorf("AllAcounts() unable to parse user from token")
		s.Logger.Log("err", err)
		return err
	}

	requestEmail, ok := requestUser.(*jwt.Token).Claims.(jwt.MapClaims)["email"].(string)
	if !ok {
		err = fmt.Errorf("AllAcounts() unable to parse email from token")
		s.Logger.Log("err", err)
		return err
	}
	requestEmailParts := strings.Split(requestEmail, "@")
	requestDomain := requestEmailParts[len(requestEmailParts)-1] // Domain is the last entry of the split since an email as only one @ sign

	for _, a := range accountList.Users {
		emailParts := strings.Split(*a.Email, "@")
		if len(emailParts) <= 0 {
			err = fmt.Errorf("AllAcounts() failed to parse email %s for domain", *a.Email)
			s.Logger.Log("err", err)
			return err
		}
		domain := emailParts[len(emailParts)-1] // Domain is the last entry of the split since an email as only one @ sign
		if requestDomain != domain {
			continue
		}
		userRoles, err := s.Auth0.Manager.User.Roles(*a.ID)
		if err != nil {
			err = fmt.Errorf("AllAcounts() failed to fetch user roles: %v", err)
			s.Logger.Log("err", err)
			return err
		}

		buyer, err := s.Storage.BuyerWithDomain(domain)

		reply.UserAccounts = append(reply.UserAccounts, newAccount(a, userRoles.Roles, buyer))
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
	if ok && requestID != args.UserID {
		if !VerifyAnyRole(r, AdminRole, OwnerRole) {
			err := fmt.Errorf("UserAccount(): %v", ErrInsufficientPrivileges)
			s.Logger.Log("err", err)
			return err
		}
	}

	userAccount, err := s.Auth0.Manager.User.Read(args.UserID)
	if err != nil {
		err := fmt.Errorf("UserAccount() failed to fetch user account: %w", err)
		s.Logger.Log("err", err)
		return err
	}

	emailParts := strings.Split(*userAccount.Email, "@")
	if len(emailParts) <= 0 {
		err := fmt.Errorf("UserAccount() failed to parse email %s for domain", *userAccount.Email)
		s.Logger.Log("err", err)
		return err
	}

	domain := emailParts[len(emailParts)-1] // Domain is the last entry of the split since an email as only one @ sign
	buyer, err := s.Storage.BuyerWithDomain(domain)
	userRoles, err := s.Auth0.Manager.User.Roles(*userAccount.ID)
	if err != nil {
		err := fmt.Errorf("UserAccount() failed to fetch user roles: %w", err)
		s.Logger.Log("err", err)
		return err
	}

	reply.UserAccount = newAccount(userAccount, userRoles.Roles, buyer)

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

	if err := s.Auth0.Manager.User.Delete(args.UserID); err != nil {
		err := fmt.Errorf("DeleteUserAccount() failed to delete user: %w", err)
		s.Logger.Log("err", err)
		return err
	}
	return nil
}

func (s *AuthService) AddUserAccount(req *http.Request, args *AccountsArgs, reply *AccountsReply) error {
	var adminString string = "Admin"
	var accounts []account

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

	connectionID := "Username-Password-Authentication"
	emails := args.Emails
	falseValue := false

	newCustomerIDsMap := make(map[string]interface{})

	// Not an onboard signup
	for _, e := range emails {

		emailParts := strings.Split(e, "@")
		if len(emailParts) <= 0 {
			err := fmt.Errorf("AddUserAccount() failed to parse email %s for domain", e)
			s.Logger.Log("err", err)
			return err
		}
		domain := emailParts[len(emailParts)-1] // Domain is the last entry of the split since an email as only one @ sign

		buyer, err := s.Storage.BuyerWithDomain(domain)

		pw, err := GenerateRandomString(32)
		if err != nil {
			err := fmt.Errorf("AddUserAccount() failed to generate a random password: %w", err)
			s.Logger.Log("err", err)
			return err
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
			err := fmt.Errorf("AddUserAccount() failed to create new user: %w", err)
			s.Logger.Log("err", err)
			return err
		}

		if err = s.Auth0.Manager.User.AssignRoles(*newUser.ID, args.Roles...); err != nil {
			err := fmt.Errorf("AddUserAccount() failed to add user roles: %w", err)
			s.Logger.Log("err", err)
			return err
		}

		accounts = append(accounts, newAccount(newUser, args.Roles, buyer))
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

	if !VerifyAnyRole(r, AdminRole, OwnerRole) {
		err := fmt.Errorf("UserAccount(): %v", ErrInsufficientPrivileges)
		s.Logger.Log("err", err)
		return err
	}

	roleList, err := s.Auth0.Manager.Role.List()
	if err != nil {
		err := fmt.Errorf("AllRoles() failed to fetch role list: %w", err)
		s.Logger.Log("err", err)
		return err
	}

	if !VerifyAllRoles(r, AdminRole) {
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
	if !VerifyAnyRole(r, AdminRole, OwnerRole) {
		err := fmt.Errorf("UserAccount(): %v", ErrInsufficientPrivileges)
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

	removeRoles := []*management.Role{
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

type UpgradeArgs struct {
}

type UpgradeReply struct {
	NewRoles []*management.Role `json:"new_roles"`
}

func (s *AuthService) UpgradeAccount(r *http.Request, args *UpgradeArgs, reply *UpgradeReply) error {
	if VerifyAnyRole(r, AdminRole, OwnerRole) {
		return nil
	}
	var companyUsers []*management.User
	accountList, err := s.Auth0.Manager.User.List()

	if err != nil {
		err = fmt.Errorf("UpgradeAccount() failed to fetch user list: %v", err)
		s.Logger.Log("err", err)
		return err
	}
	requestUser := r.Context().Value("user")
	if requestUser == nil {
		err = fmt.Errorf("UpgradeAccount() unable to parse user from token")
		s.Logger.Log("err", err)
		return err
	}

	requestEmail, ok := requestUser.(*jwt.Token).Claims.(jwt.MapClaims)["email"].(string)
	if !ok {
		err = fmt.Errorf("UpgradeAccount() unable to parse email from token")
		s.Logger.Log("err", err)
		return err
	}
	requestEmailParts := strings.Split(requestEmail, "@")
	requestDomain := requestEmailParts[len(requestEmailParts)-1] // Domain is the last entry of the split since an email as only one @ sign

	for _, a := range accountList.Users {
		emailParts := strings.Split(*a.Email, "@")
		if len(emailParts) <= 0 {
			err = fmt.Errorf("UpgradeAccount() failed to parse email %s for domain", *a.Email)
			s.Logger.Log("err", err)
			return err
		}
		domain := emailParts[len(emailParts)-1] // Domain is the last entry of the split since an email as only one @ sign
		if requestDomain != domain {
			continue
		}
		companyUsers = append(companyUsers, a)
	}
	if len(companyUsers) > 1 {
		return nil
	}
	roleNames := []string{
		"rol_ScQpWhLvmTKRlqLU",
		"rol_8r0281hf2oC4cvrD",
	}
	roleTypes := []string{
		"Viewer",
		"Owner",
	}
	roleDescriptions := []string{
		"Can see current sessions and the map.",
		"Can access and manage everything in an account.",
	}

	requestID, ok := requestUser.(*jwt.Token).Claims.(jwt.MapClaims)["sub"].(string)
	if !ok {
		err = fmt.Errorf("UpgradeAccount() unable to parse id from token")
		s.Logger.Log("err", err)
		return err
	}
	// Upgrade account
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

	err = s.Auth0.Manager.User.AssignRoles(requestID, roles...)

	if err != nil {
		return err
	}

	userAccount, err := s.Auth0.Manager.User.Read(requestID)
	if err != nil {
		err := fmt.Errorf("UpgradeAccount() failed to fetch user account: %w", err)
		s.Logger.Log("err", err)
		return err
	}

	userRoles, err := s.Auth0.Manager.User.Roles(*userAccount.ID)
	if err != nil {
		err := fmt.Errorf("UpgradeAccount() failed to fetch user roles: %w", err)
		s.Logger.Log("err", err)
		return fmt.Errorf("failed to fetch user roles: %w", err)
	}

	reply.NewRoles = userRoles.Roles

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

	err := s.Auth0.Manager.Job.VerifyEmail(job)
	if err != nil {
		err := fmt.Errorf("VerifyEmailUrl() failed to creating verification email link: %s", err)
		s.Logger.Log("err", err)
		return err
	}

	reply.Sent = true

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
	ctx = context.WithValue(ctx, anonymousCallKey, value)
	return r.WithContext(ctx)
}

func IsAnonymous(r *http.Request) bool {
	anon, ok := r.Context().Value(anonymousCallKey).(bool)
	return ok && anon
}

func SetRoles(r *http.Request, roles []string) *http.Request {
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

// RoleFunc defines a function that takes in an http.Request and perform a check on it whether it has a role or not.
type RoleFunc func(req *http.Request) (bool, error)

// Ops checks the request for the appropriate "scope" in the JWT
var OpsRole = func(req *http.Request) (bool, error) {
	user := req.Context().Value("user")

	if user != nil {
		claims := user.(*jwt.Token).Claims.(jwt.MapClaims)

		if _, ok := claims["scope"]; ok {
			return true, nil
		}
	}
	return false, fmt.Errorf("OpsRole(): failed to fetch user from token")
}

var AdminRole = func(req *http.Request) (bool, error) {
	requestRoles := req.Context().Value(rolesKey)

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
	requestRoles := req.Context().Value(rolesKey)

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
	anon, ok := req.Context().Value(anonymousCallKey).(bool)
	return ok && anon, nil
}

// Ops checks the request for the appropriate "scope" in the JWT
var UnverifiedRole = func(req *http.Request) (bool, error) {
	user := req.Context().Value("user")

	if user == nil {
		return false, fmt.Errorf("UnverifiedRole(): failed to fetch user from token")
	}
	claims := user.(*jwt.Token).Claims.(jwt.MapClaims)

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
