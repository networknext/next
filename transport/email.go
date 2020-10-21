package transport

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const (
	APIKEY        = "b4aefe551af3c1167fc6bfa257786874-us20"
	SERVER_PREFIX = "us20"
	LIST_ID       = "553903bc6f"
)

func TagNewSignup(email string) error {
	if email == "" {
		err := fmt.Errorf("TagNewSignup() email can not be empty")
		return err
	}
	emailHash := md5.Sum([]byte(strings.ToLower(email)))

	tags := map[string]interface{}{
		"tags": []map[string]string{
			{
				"name":   "Portal Signups",
				"status": "active",
			},
		},
	}

	jsonValue, err := json.Marshal(tags)
	if err != nil {
		err = fmt.Errorf("TagNewSignup() failed marshal the payload: %v", err)
		return err
	}
	payload := bytes.NewBuffer(jsonValue)

	URL := fmt.Sprintf("https://%s.api.mailchimp.com/3.0/lists/%s/members/%s/tags", SERVER_PREFIX, LIST_ID, fmt.Sprintf("%x", emailHash))

	req, err := http.NewRequest("POST", URL, payload)
	if err != nil {
		err = fmt.Errorf("TagNewSignup() failed to setup request: %v", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth("key", APIKEY)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		err = fmt.Errorf("TagNewSignup() failed to tag contact")
		return err
	}
	defer resp.Body.Close()
	return nil
}

func AddEmailToMailChimp(email string) error {
	if email == "" {
		err := fmt.Errorf("TagNewSignup() email can not be empty")
		return err
	}
	payload := map[string]string{
		"email_address": email,
		"status":        "pending",
	}

	jsonValue, err := json.Marshal(payload)
	if err != nil {
		err = fmt.Errorf("AddSignupToMailChimp() failed marshal the payload: %v", err)
		return err
	}

	URL := fmt.Sprintf("https://%s.api.mailchimp.com/3.0/lists/%s/members", SERVER_PREFIX, LIST_ID)

	req, err := http.NewRequest("POST", URL, bytes.NewBuffer(jsonValue))
	if err != nil {
		err = fmt.Errorf("AddSignupToMailChimp() failed to setup request: %v", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth("key", APIKEY)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		err = fmt.Errorf("AddSignupToMailChimp() failed to send email")
		return err
	}
	defer resp.Body.Close()
	return nil
}
