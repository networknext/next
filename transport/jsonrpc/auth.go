package jsonrpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	"github.com/networknext/backend/storage"
	"gopkg.in/auth0.v4/management"
)

type AuthService struct {
	Auth0 storage.Auth0
}

type AccountsArgs struct {
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
	UserID string             `json:"user_id"`
	Name   string             `json:"name"`
	Email  string             `json:"email"`
	Roles  []*management.Role `json:"roles"`
}

func (s *AuthService) AllAccounts(r *http.Request, args *AccountsArgs, reply *AccountsReply) error {
	accountList, err := s.Auth0.Manager.User.List()
	if err != nil {
		return fmt.Errorf("failed to fetch user list: %w", err)
	}

	for _, a := range accountList.Users {
		userRoles, err := s.Auth0.Manager.User.Roles(*a.ID)
		if err != nil {
			return fmt.Errorf("failed to fetch user roles: %w", err)
		}
		reply.UserAccounts = append(reply.UserAccounts, UnMarshalUser(a, userRoles))
	}
	return nil
}

func (s *AuthService) UserAccount(r *http.Request, args *AccountArgs, reply *AccountReply) error {
	if args.UserID == "" {
		return fmt.Errorf("user_id is required")
	}

	userAccount, err := s.Auth0.Manager.User.Read(args.UserID)
	if err != nil {
		return fmt.Errorf("failed to fetch user account: %w", err)
	}

	userRoles, err := s.Auth0.Manager.User.Roles(args.UserID)

	if err != nil {
		return fmt.Errorf("failed to get user roles: %w", err)
	}

	reply.UserAccount = UnMarshalUser(userAccount, userRoles)

	return nil
}

func (s *AuthService) DeleteUserAccount(r *http.Request, args *AccountArgs, reply *AccountReply) error {
	if args.UserID == "" {
		return fmt.Errorf("user_id is required")
	}

	if err := s.Auth0.Manager.User.Delete(args.UserID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

func UnMarshalUser(u *management.User, r *management.RoleList) account {
	account := account{
		UserID: *u.Identities[0].UserID,
		Name:   *u.Name,
		Email:  *u.Email,
		Roles:  r.Roles,
	}

	return account
}

func MarshalUser(a account) *management.User {
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
	roleList, err := s.Auth0.Manager.Role.List()
	if err != nil {
		fmt.Errorf("failed to fetch role list: %w", err)
	}

	reply.Roles = roleList.Roles

	return nil
}

func (s *AuthService) UserRoles(r *http.Request, args *RolesArgs, reply *RolesReply) error {
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
	if args.UserID == "" {
		return fmt.Errorf("user_id is required")
	}

	userRoles, err := s.Auth0.Manager.User.Roles(args.UserID)

	if err != nil {
		return fmt.Errorf("failed to fetch user roles: %w", err)
	}

	for _, r := range userRoles.Roles {
		err := s.Auth0.Manager.User.RemoveRoles(args.UserID, r)
		if err != nil {
			return fmt.Errorf("failed to remove current user role: %w", err)
		}
	}

	if len(args.Roles) == 0 {
		return nil
	}

	for _, r := range args.Roles {
		err := s.Auth0.Manager.User.AssignRoles(args.UserID, r)
		if err != nil {
			return fmt.Errorf("failed to assign role: %w", err)
		}
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
			// Verify 'aud' claim
			checkAud := token.Claims.(jwt.MapClaims).VerifyAudience(audience, false)
			if !checkAud {
				return token, errors.New("Invalid audience.")
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
		SigningMethod: jwt.SigningMethodRS256,
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
