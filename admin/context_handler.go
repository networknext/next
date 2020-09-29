package admin

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dgrijalva/jwt-go"

	"github.com/gorilla/rpc/v2"
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

func RPCInterceptHandler(i *rpc.RequestInfo) *http.Request {
	user := i.Request.Context().Value("user")
	if user != nil {
		claims := user.(*jwt.Token).Claims.(jwt.MapClaims)

		if requestData, ok := claims["https://networknext.com/userData"]; ok {
			var userRoles []string
			if roles, ok := requestData.(map[string]interface{})["roles"]; ok {
				rolesInterface := roles.([]interface{})
				userRoles = make([]string, len(rolesInterface))
				for i, v := range rolesInterface {
					userRoles[i] = v.(string)
				}
			}
			var companyCode string
			if companyCodeInterface, ok := requestData.(map[string]interface{})["company_code"]; ok {
				companyCode = companyCodeInterface.(string)
			}
			var newsletterConsent bool
			if consent, ok := requestData.(map[string]interface{})["newsletter"]; ok {
				newsletterConsent = consent.(bool)
			}
			return AddTokenContext(i.Request, userRoles, companyCode, newsletterConsent)
		}
	}
	return SetIsAnonymous(i.Request, i.Request.Header.Get("Authorization") == "")
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
	requestUser, ok := r.Context().Value("user").(*jwt.Token)
	if !ok || requestUser == nil {
		return nil, fmt.Errorf("RequestEmail() unable to parse request user from token")
	}
	return requestUser, nil
}

func RequestEmail(requestUser *jwt.Token) (string, error) {
	requestEmail, ok := requestUser.Claims.(jwt.MapClaims)["email"].(string)
	if !ok {
		return "", fmt.Errorf("RequestEmail() unable to parse email from token")
	}
	return requestEmail, nil
}

func RequestCompany(r *http.Request) (string, error) {
	requestCompany, ok := r.Context().Value(Keys.CompanyKey).(string)
	if !ok || requestCompany == "" {
		return "", fmt.Errorf("RequestCompany(): failed to get company from context")
	}
	return requestCompany, nil
}
