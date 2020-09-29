package admin

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dgrijalva/jwt-go"
)

type contextType string

type contextKeys struct {
	AnonymousCallKey     contextType
	RolesKey             contextType
	CompanyKey           contextType
	NewsletterConsentKey contextType
}

var Keys contextKeys = contextKeys{
	AnonymousCallKey:     "anonymous",
	RolesKey:             "roles",
	CompanyKey:           "company",
	NewsletterConsentKey: "newsletter",
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

func RequestUser(r *http.Request) (*jwt.Token, error) {
	requestUser := r.Context().Value("user")
	if requestUser == nil {
		return nil, err
	}
}

func RequestEmail(requestUser *jwt.Token) (string, error) {
	requestEmail, ok := requestUser.Claims.(jwt.MapClaims)["email"].(string)
	if !ok {
		return "", fmt.Errorf("RequestEmail() unable to parse email from token")
	}
	return requestEmail, nil
}

func RequestCompany(r *http.Request) (string, error) {
	requestCompany := r.Context().Value(Keys.CompanyKey)
	if requestCompany == nil {
		return "", fmt.Errorf("RequestCompany(): failed to get company from context")
	}
}
