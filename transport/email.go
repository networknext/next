package transport

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const (
	APIKEY        = "b4aefe551af3c1167fc6bfa257786874-us20"
	SERVER_PREFIX = "us20"
)

type Contact struct {
	address string `json:"email_address"`
	status  string `json:"status"`
}

type Tag struct {
	name   string `json:"name"`
	status string `json:"status"`
}

type TagUpdate struct {
	tags []Tag `json:"tags"`
}

func TagNewSignup(email string) error {
	/* 	payload := EmailPayload{
		key: APIKEY,
		message: EmailMessage{
			from:    "",
			subject: "Welcome to Network Next!",
			text:    "",
			to: []Email{
				{
					emailType: "to",
					address:   toAddress,
				},
			},
		},
	} */

	emailHash := md5.Sum([]byte(strings.ToLower(email)))

	tags := TagUpdate{
		tags: []Tag{
			{
				name: "Portal Signups",
				status: "active",
			},
		},
	}

	bytes, err := json.Marshal(jsonObject)
	if err != nil {
		err = fmt.Errorf("TagNewSignup() failed marshal the payload: %v", err)
		return err
	}
	payload := strings.NewReader(string(bytes))

	URL := "https://" + SERVER_PREFIX + ".api.mailchimp.com/3.0/lists/$list_id/members/" + emailHash + "/tags"
	URL := fmt.Sprintf("https://%s.api.mailchimp.com/3.0/lists/%s/members/%s/tags", SERVER_PREFIX, , emailemailHash)

	req, err := http.NewRequest("POST", URL, payload)
	if err != nil {
		err = fmt.Errorf("TagNewSignup() failed to setup request: %v", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		err = fmt.Errorf("TagNewSignup() failed to tag contact: %v", err)
		return err
	}
	defer resp.Body.Close()
	return nil
}

func AddSignupToMailChimp(email string) error {
	/* 	payload := EmailPayload{
		key: APIKEY,
		message: EmailMessage{
			from:    "",
			subject: "Welcome to Network Next!",
			text:    "",
			to: []Email{
				{
					emailType: "to",
					address:   toAddress,
				},
			},
		},
	} */

	jsonObject := Contact{
		address: email,
		status:  "pending",
	}

	bytes, err := json.Marshal(jsonObject)
	if err != nil {
		err = fmt.Errorf("AddSignupToMailChimp() failed marshal the payload: %v", err)
		return err
	}
	payload := strings.NewReader(string(bytes))

	URL := "https://" + SERVER_PREFIX + ".api.mailchimp.com/3.0/lists/$list_id/members"

	req, err := http.NewRequest("POST", URL, payload)
	if err != nil {
		err = fmt.Errorf("AddSignupToMailChimp() failed to setup request: %v", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		err = fmt.Errorf("AddSignupToMailChimp() failed to send email: %v", err)
		return err
	}
	defer resp.Body.Close()
	return nil
}
