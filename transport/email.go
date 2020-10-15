package transport

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const (
	APIKEY        = ""
	SERVER_PREFIX = ""
)

type Contact struct {
	address string `json:"email_address"`
	status  string `json:"status"`
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
		status:  "not subscribed",
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
