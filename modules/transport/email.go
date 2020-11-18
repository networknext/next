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
	APIKEY = "b4aefe551af3c1167fc6bfa257786874-us20"
)

type MailChimpHandler struct {
	HTTPHandler http.Client
	MembersURI  string
}

func (h *MailChimpHandler) TagNewSignup(email string) error {
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

	URL := fmt.Sprintf("%s/%s/tags", h.MembersURI, fmt.Sprintf("%x", emailHash))

	req, err := http.NewRequest("POST", URL, payload)
	if err != nil {
		err = fmt.Errorf("TagNewSignup() failed to setup request: %v", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth("key", APIKEY)

	resp, err := h.HTTPHandler.Do(req)
	if err != nil || !(resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusOK) {
		err = fmt.Errorf("TagNewSignup() failed to tag contact: error: %v, status: %v", err, resp.StatusCode)
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (h *MailChimpHandler) AddEmailToMailChimp(email string) error {
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

	req, err := http.NewRequest("POST", h.MembersURI, bytes.NewBuffer(jsonValue))
	if err != nil {
		err = fmt.Errorf("AddSignupToMailChimp() failed to setup request: %v", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth("key", APIKEY)

	resp, err := h.HTTPHandler.Do(req)
	if err != nil || !(resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusOK) {
		err = fmt.Errorf("TagNewSignup() failed to tag contact: error: %v, status: %v", err, resp.StatusCode)
		return err
	}
	defer resp.Body.Close()
	return nil
}
