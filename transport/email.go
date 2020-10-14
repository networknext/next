package transport

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const (
	APIKEY = ""
)

type Email struct {
	address   string `json:"email"`
	emailType string `json:"type"`
}

type EmailMessage struct {
	from    string  `json:"from_email"`
	subject string  `json:"subject"`
	text    string  `json:"text"`
	to      []Email `json:"to"`
}

type EmailPayload struct {
	key     string       `json:"key"`
	message EmailMessage `json:"message"`
}

func SendEmail(bytePayload []byte) error {
	payload := strings.NewReader(string(bytePayload))

	req, _ := http.NewRequest("POST", "https://mandrillapp.com/api/1.0/messages/send", payload)
	if err != nil {
		err = fmt.Errorf("SendEmail() failed to setup request: %v", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		err = fmt.Errorf("SendEmail() failed to send email: %v", err)
		return err
	}
	defer resp.Body.Close()
}

func SendSignupEmail(toAddress string) error {
	payload := EmailPayload{
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
	}

	bytes, err := json.Marshal(payload)
	if err != nil {
		err = fmt.Errorf("SendSignupEmail() failed marshal the email payload: %v", err)
		return err
	}

	return SendEmail(bytes)
}
