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
	Properties CreateDealProperties `json:"properties,omitempty"`
}

type CreateDealProperties struct {
	Amount    string `json:"amount,omitempty"`
	CloseDate string `json:"closedate,omitempty"`
	DealName  string `json:"dealname,omitempty"`
	DealStage string `json:"dealstage,omitempty"`
	Pipeline  string `json:"pipeline,omitempty"`
}

func (hsc HubSpotClient) CreateHubSpotDealEntry(action string) error {
	properties := CreateDealProperties{
		Amount:    "", // To be filled in later
		DealName:  action,
		DealStage: "13770918",
		Pipeline:  "default",
	}

	message := CreateDealMessage{
		Properties: properties,
	}
	hubspotBody, _ := json.Marshal(message)
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("https://api.hubapi.com/crm/v3/objects/deals?hapikey=%s", hsc.APIKey), bytes.NewBuffer(hubspotBody))
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

	// Check if the status code is outside the range of possible 200 level responses
	if resp.StatusCode-200 > 26 {
		return errors.New("HubSpot returned non-200 status code")
	}
	return nil
}
