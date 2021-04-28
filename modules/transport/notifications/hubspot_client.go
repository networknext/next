package notifications

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

const DefaultHubSpotTimeout = 5 * time.Second

type HubSpotClient struct {
	APIKey  string
	TimeOut time.Duration
}

type CreateDealMessage struct {
	Amount    string    `json:"amount,omitempty"`
	CloseDate time.Time `json:"closedate,omitempty"`
	DealName  string    `json:"dealname,omitempty"`
	DealStage string    `json:"dealstage,omitempty"`
	Pipeline  string    `json:"pipeline,omitempty"`
}

func (hsc HubSpotClient) CreateHubSpotDealEntry(customerName string, action string) error {
	message := CreateDealMessage{
		Amount:    "",                                 // To be filled in later
		CloseDate: time.Now().Add(7 * 24 * time.Hour), // Two week default close date
		DealName:  fmt.Sprintf("%s %s", customerName, action),
		DealStage: "Self Serve Lead Stages Catch All",
		Pipeline:  "default",
	}
	hubspotBody, _ := json.Marshal(message)
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", hsc, "https://api.hubapi.com/crm/v3/objects/deals"), bytes.NewBuffer(hubspotBody))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	if hsc.TimeOut == 0 {
		hsc.TimeOut = DefaultHubSpotTimeout
	}
	client := &http.Client{Timeout: hsc.TimeOut}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return err
	}
	if buf.String() != "ok" {
		return errors.New("Non-ok response returned from Hubspot")
	}
	return nil
}
