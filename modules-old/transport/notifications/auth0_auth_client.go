package notifications

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	CHANGE_PASSWORD_URL = "/dbconnections/change_password"
)

// Client for talking with the auth0 Authentication API
type Auth0AuthClient struct {
	ClientID string
	Domain   string
}

func NewAuth0AuthClient(clientID string, domain string) (*Auth0AuthClient, error) {
	if clientID == "" || domain == "" {
		return nil, fmt.Errorf("ClientID and Domain are required")
	}

	return &Auth0AuthClient{
		ClientID: clientID,
		Domain:   domain,
	}, nil
}

func (c *Auth0AuthClient) SendChangePasswordEmail(email string) error {
	url := fmt.Sprintf("https://%s/dbconnections/change_password", c.Domain)

	payloadString := fmt.Sprintf("{\"client_id\": \"%s\",\"email\": \"%s\",\"connection\": \"Username-Password-Authentication\"}", c.ClientID, email)
	payload := strings.NewReader(payloadString)

	req, _ := http.NewRequest("POST", url, payload)

	req.Header.Add("content-type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if _, err = ioutil.ReadAll(res.Body); err != nil {
		return err
	}

	return nil
}
