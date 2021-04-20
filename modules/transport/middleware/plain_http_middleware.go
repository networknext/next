package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

func HttpGetMiddleware(audience string, next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {

			var claims jwt.MapClaims

			authHeader := strings.Split(r.Header.Get("Authorization"), "Bearer ")
			if len(authHeader) != 2 {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("Malformed Token"))
			} else {
				jwtToken := authHeader[1]
				token, err := jwt.Parse(jwtToken, func(token *jwt.Token) (interface{}, error) {
					if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
						return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
					}

					claims = token.Claims.(jwt.MapClaims)
					if _, ok := claims["scope"]; !ok {
						if !claims.VerifyAudience(audience, false) {
							return token, errors.New("invalid audience")
						}
					}

					cert, err := getPemCert(token)
					if err != nil {
						return nil, err
					}

					return jwt.ParseRSAPublicKeyFromPEM([]byte(cert))
				})

				if token.Valid {
					// TODO: send the auth0 claims to get the "author" from the sub
					//       when we start rolling database.bin files from the admin tool
					// ctx := context.WithValue(r.Context(), "props", claims)
					ctx := context.Background()
					next.ServeHTTP(w, r.WithContext(ctx))
				} else {
					fmt.Println(err)
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte("Unauthorized"))
				}
			}
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
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
		err := errors.New("unable to find appropriate key")
		return cert, err
	}

	return cert, nil
}
