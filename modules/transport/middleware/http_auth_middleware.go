package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
)

type contextKeys struct {
	AnonymousCallKey     string
	RolesKey             string
	CustomerKey          string
	NewsletterConsentKey string
	UserKey              string
}

var Keys contextKeys = contextKeys{
	AnonymousCallKey:     "anonymous",
	RolesKey:             "roles",
	CustomerKey:          "customer",
	NewsletterConsentKey: "newsletter",
	UserKey:              "user",
}

// Standard Auth0 HTTP authentication middleware. If the endpoint being secured by this middleware is an RPC endpoint, set "useJSONRPC to true"
func HTTPAuthMiddleware(keys JWKS, audiences []string, next http.Handler, allowedOrigins []string, issuer string, useJSONRPC bool) http.Handler {
	middlewareOptions := jwtmiddleware.Options{}

	if useJSONRPC {
		middlewareOptions.UserProperty = Keys.UserKey
	}

	middlewareOptions.SigningMethod = jwt.SigningMethodRS256
	middlewareOptions.CredentialsOptional = useJSONRPC

	mw := jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			claims := token.Claims.(jwt.MapClaims)

			if _, ok := claims["scope"]; !ok {
				valid := false
				for _, audience := range audiences {
					valid = !claims.VerifyAudience(audience, false)
					if valid {
						break
					}
				}
				if !valid {
					return token, errors.New("Invalid audience.")
				}
			}

			checkIss := token.Claims.(jwt.MapClaims).VerifyIssuer(issuer, false)
			if !checkIss {
				return nil, errors.New("invalid issuer")
			}

			cert, err := getPemCert(keys, token)
			if err != nil {
				return nil, err
			}

			return jwt.ParseRSAPublicKeyFromPEM([]byte(cert))
		},
		SigningMethod:       jwt.SigningMethodRS256,
		CredentialsOptional: false,
	})

	return CORSControlHandler(allowedOrigins, mw.Handler(next))
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

func AddTokenContext(r *http.Request, roles []string, customerCode string, newsletterConsent bool) *http.Request {
	ctx := r.Context()
	if len(roles) > 0 {
		ctx = context.WithValue(ctx, Keys.RolesKey, roles)
	}
	if customerCode != "" {
		ctx = context.WithValue(ctx, Keys.CustomerKey, customerCode)
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

var AssignedToCompanyRole = func(req *http.Request) (bool, error) {
	requestCustomerCode, ok := req.Context().Value(Keys.CustomerKey).(string)
	if !ok || requestCustomerCode == "" {
		return false, nil
	}
	return true, nil
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

func RequestUserInformation(ctx context.Context) interface{} {
	return ctx.Value(Keys.UserKey)
}

func RequestUserCustomerCode(ctx context.Context) string {
	customerCode, ok := ctx.Value(Keys.CustomerKey).(string)
	if !ok {
		customerCode = ""
	}
	return customerCode
}

type JWKS struct {
	Keys []struct {
		Kty string   `json:"kty"`
		Kid string   `json:"kid"`
		Use string   `json:"use"`
		N   string   `json:"n"`
		E   string   `json:"e"`
		X5c []string `json:"x5c"`
	} `json:"keys"`
}

func FetchAuth0Cert(domain string) (JWKS, error) {
	resp, err := http.Get(fmt.Sprintf("https://%s/.well-known/jwks.json", domain))
	if err != nil {
		return JWKS{}, err
	}
	defer resp.Body.Close()

	keys := JWKS{}
	err = json.NewDecoder(resp.Body).Decode(&keys)
	if err != nil {
		return JWKS{}, err
	}

	return keys, nil
}

func getPemCert(keys JWKS, token *jwt.Token) (string, error) {
	cert := ""
	for k := range keys.Keys {
		if token.Header["kid"] == keys.Keys[k].Kid {
			cert = "-----BEGIN CERTIFICATE-----\n" + keys.Keys[k].X5c[0] + "\n-----END CERTIFICATE-----"
		}
	}

	if cert == "" {
		err := errors.New("unable to find appropriate key")
		return cert, err
	}

	return cert, nil
}
