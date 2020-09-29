package admin

import (
	"fmt"
	"net/http"

	"github.com/dgrijalva/jwt-go"
)

var RoleNames []string = []string{
	"rol_ScQpWhLvmTKRlqLU",
	"rol_8r0281hf2oC4cvrD",
	"rol_YfFrtom32or4vH89",
}
var RoleTypes []string = []string{
	"Viewer",
	"Owner",
	"Admin",
}
var RoleDescriptions []string = []string{
	"Can see current sessions and the map.",
	"Can access and manage everything in an account.",
	"Can manage the Network Next system, including access to configstore.",
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

var SameBuyerRole = func(companyCode string) RoleFunc {
	return func(req *http.Request) (bool, error) {
		if VerifyAnyRole(req, AdminRole, OpsRole) {
			return true, nil
		}
		if VerifyAllRoles(req, AnonymousRole) {
			return false, nil
		}
		if companyCode == "" {
			return false, fmt.Errorf("SameBuyerRole(): buyerID is required")
		}
		requestCompanyCode, ok := req.Context().Value(Keys.CompanyKey).(string)
		if !ok {
			err := fmt.Errorf("SameBuyerRole(): user is not assigned to a company")
			return false, err
		}
		if requestCompanyCode == "" {
			err := fmt.Errorf("SameBuyerRole(): failed to parse company code")
			return false, err
		}

		return companyCode == requestCompanyCode, nil
	}
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
