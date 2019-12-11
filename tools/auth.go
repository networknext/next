/*
   Network Next. You control the network.
   Copyright © 2017 - 2019 Network Next, Inc. All rights reserved.
*/

package tools

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	gorillaContext "github.com/gorilla/context"
	"github.com/gorilla/mux"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Tokens struct {
	Access  string `json:"access_token"`
	Refresh string `json:"refresh_token"`
}

type Auth struct {
	mutex  sync.RWMutex
	tokens Tokens
	state  string
}

var auth Auth

func AuthTokens() Tokens {
	return auth.tokens
}

func AuthPath() (string, error) {
	return filepath.Abs("./.auth")
}

func AuthClear() error {
	path, err := AuthPath()
	if err != nil {
		return fmt.Errorf("error calculating filepath to save access token: %v", err)
	}
	os.Remove(path)
	return nil
}

func AuthRead() error {
	path, err := AuthPath()
	if err != nil {
		return fmt.Errorf("error calculating filepath to save access token: %v", err)
	}

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("error reading file %s: %v", path, err)
	}

	auth.mutex.Lock()
	err = json.Unmarshal(bytes, &auth.tokens)
	auth.mutex.Unlock()
	if err != nil {
		return fmt.Errorf("error parsing JSON credentials: %v", err)
	}

	return AuthVerifyToken(auth.tokens.Access)
}

func AuthGetToken(ctx context.Context, env string, code_type string, code string) error {

	domain := "networknext.auth0.com"
	clientId := "yjv3Uqy6Iiwh6_xMY6CfPrWMc0hVMnF2"
	clientSecret := "pAiWSsF5NHRWR158PweqBW1ixQ8utenN7P2yHJLcHZIygNddEelqSCR7MozTEdUQ"

	conf := &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  "http://localhost:44100/callback",
		Scopes:       []string{"cli", "offline_access", fmt.Sprintf("env:%s", env)},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://" + domain + "/authorize",
			TokenURL: "https://" + domain + "/oauth/token",
		},
	}

	token, err := conf.Exchange(ctx, code)
	if err != nil {
		return err
	}

	var tokens Tokens
	tokens.Refresh = token.Extra("refresh_token").(string)
	tokens.Access = token.AccessToken

	auth.mutex.RLock()
	// make sure to preserve the refresh token
	if auth.tokens.Refresh != "" {
		tokens.Refresh = auth.tokens.Refresh
	}
	auth.mutex.RUnlock()

	if err := AuthVerifyToken(tokens.Access); err != nil {
		return fmt.Errorf("failed to verify access token: %v", err)
	}

	path, err := AuthPath()
	if err != nil {
		return fmt.Errorf("error calculating filepath to save access token: %v", err)
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("couldn't open '%s' for writing: %v", path, err)
	}
	defer file.Close()

	data, err := json.Marshal(tokens)
	if err != nil {
		return fmt.Errorf("failed to marshal token json: %v", err)
	}

	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("couldn't write access token to '%s': %v", path, err)
	}

	err = file.Sync()
	if err != nil {
		return fmt.Errorf("couldn't commit file '%s' to disk: %v", path, err)
	}

	auth.mutex.Lock()
	auth.tokens = tokens
	auth.mutex.Unlock()

	return nil
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	nonce := base64.StdEncoding.EncodeToString(RandomBytes(8))

	url := fmt.Sprintf("https://networknext.auth0.com/authorize"+
		"?audience=%s"+
		"&scope=cli%%20offline_access%%20system.costmatrix.view%%20system.relays.list%%20system.sessions.list%%20system.invoicing.buyer%%20system.invoicing.seller%%20env:%s"+
		"&response_type=code"+
		"&client_id=yjv3Uqy6Iiwh6_xMY6CfPrWMc0hVMnF2"+
		"&redirect_uri=http://localhost:44100/auth"+
		"&nonce=%s"+
		"&state=%s\n\n",
		url.QueryEscape("https://networknext.com"), Env, url.QueryEscape(nonce), url.QueryEscape(auth.state))
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func AuthHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	var err error

	if query.Get("state") != auth.state {
		err = fmt.Errorf("state mismatch: expected %s, got %s", auth.state, query.Get("state"))
	}

	if err == nil {
		err = AuthGetToken(r.Context(), Env, "code", query.Get("code"))
	}

	var msg string
	var tryAgain string

	if err == nil {
		msg = "Login successful! You can close this window."
	} else {
		msg = fmt.Sprintf("auth error: %v", err)
		tryAgain = "<p><a href=\"./\">Try again</a></p>"
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(
		"<!DOCTYPE html>" +
			"<html>" +
			"<head>" +
			"<title>Network Next</title>" +
			"<style>" +
			"* { box-sizing: border-box; }" +
			"body { background-color: #000; color: #fff; font-family: sans-serif; font-size: x-large; font-weight: bold; }" +
			"img { display: block; margin-left: auto; margin-right: auto; width: 100%; max-width: 720px; padding: 3em; padding-bottom: 0px; }" +
			"p { display: block; text-align: center; padding: 3em; }" +
			"a { color: #1aa0e0; text-decoration: none; }" +
			"a:hover { text-decoration: underline; }" +
			"</style>" +
			"</head>" +
			"<body>" +
			"<img src=\"https://assets.networknext.com/assets/logo.png\" />" +
			"<p id=\"msg\">" + msg + "</p>" +
			tryAgain +
			"</body>" +
			"</html>"))
}

func AuthVerifyToken(tokenString string) error {
	auth.mutex.RLock()
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte("FNL2IOw1nx20LtzFxpcZoumTglDcgOAL"), nil
	})
	auth.mutex.RUnlock()

	// deal with clock skew: if the only validation error is an invalid issued at time field, let it go through
	issuedAtHack := false
	if err != nil {
		if err.(*jwt.ValidationError).Errors == jwt.ValidationErrorIssuedAt {
			fmt.Printf("warning: token is not valid yet (possible clock skew)\n")
			issuedAtHack = true
		} else {
			return fmt.Errorf("failed to validate JWT token: %v", err)
		}
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && (token.Valid || issuedAtHack) {
		if !strings.Contains(claims["scope"].(string), "cli") ||
			!strings.Contains(claims["scope"].(string), "offline_access") ||
			claims["iss"] != "https://networknext.auth0.com/" ||
			claims["aud"] != "https://networknext.com" {
			return fmt.Errorf("invalid claims: %v", claims)
		}
	} else {
		return fmt.Errorf("invalid token")
	}

	return nil
}

var AuthenticationNotPermitted bool

func Auth_main() {
	// used to prevent "next local" from asking for permission EVER
	if AuthenticationNotPermitted {
		return
	}

	auth.mutex.Lock()
	auth.tokens.Access = ""
	auth.tokens.Refresh = ""
	auth.mutex.Unlock()
	auth.state = base64.StdEncoding.EncodeToString(RandomBytes(8))

	router := mux.NewRouter()

	router.HandleFunc("/", IndexHandler)
	router.HandleFunc("/auth", AuthHandler)

	server := &http.Server{
		Addr:         ":44100",
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      gorillaContext.ClearHandler(router),
	}

	OpenBrowser("http://localhost:44100")

	go server.ListenAndServe()

	for {
		auth.mutex.RLock()
		token := auth.tokens.Access
		auth.mutex.RUnlock()
		if token != "" {
			break
		}
		time.Sleep(time.Millisecond * 100)
	}

	if err := server.Shutdown(context.Background()); err != nil {
		fmt.Printf("error: could not gracefully shut down server: %v\n", err)
		os.Exit(1)
	}
}
