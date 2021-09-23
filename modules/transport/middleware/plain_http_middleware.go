package middleware

import (
	"encoding/json"
	"errors"
	"net/http"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
)

func PlainHttpAuthMiddleware(keys JWKS, audience string, next http.Handler, allowedOrigins []string) http.Handler {

	mw := jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			claims := token.Claims.(jwt.MapClaims)

			if _, ok := claims["scope"]; !ok {
				if !claims.VerifyAudience(audience, false) {
					return nil, errors.New("invalid audience")
				}
			}

			iss := "https://networknext.auth0.com/"
			checkIss := token.Claims.(jwt.MapClaims).VerifyIssuer(iss, false)
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

func FetchAuth0Cert() (JWKS, error) {
	resp, err := http.Get("https://networknext.auth0.com/.well-known/jwks.json")
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
