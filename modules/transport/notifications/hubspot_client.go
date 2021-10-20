package notifications

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

const DefaultHubSpotTimeout = 10 * time.Second

type HubSpotClient struct {
	APIKey  string
	TimeOut time.Duration
}

func NewHubSpotClient(APIKey string, timeout time.Duration) (*HubSpotClient, error) {
	if APIKey == "" || timeout == 0 {
		return nil, fmt.Errorf("API key and timeout are required")
	}

	return &HubSpotClient{
		APIKey:  APIKey,
		TimeOut: timeout,
	}, nil
}

type FetchAllCompanyEntriesResponse struct {
	Results []HubspotCompanyResults `json:"results,omitempty"`
}

type HubspotCompanyResults struct {
	ID         string                   `json:"id,omitempty"`
	Properties HubspotCompanyProperties `json:"properties,omitempty"`
	CreatedAt  string                   `json:"createdAt,omitempty"`
	UpdatedAt  string                   `json:"updatedAt,omitempty"`
	Archived   bool                     `json:"archived,omitempty"`
}

type HubspotCompanyProperties struct {
	City                    string `json:"city,omitempty"`
	CreateDate              string `json:"createdate,omitempty"`
	Domain                  string `json:"domain,omitempty"`
	HubspotLastModifiedDate string `json:"hs_lastmodifieddate,omitempty"`
	Industry                string `json:"industry,omitempty"`
	Name                    string `json:"name,omitempty"`
	Phone                   string `json:"phone,omitempty"`
	State                   string `json:"state,omitempty"`
}

func (hsc *HubSpotClient) FetchAllCompanyEntries() ([]HubspotCompanyProperties, error) {
	companies := make([]HubspotCompanyProperties, 0)

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://api.hubapi.com/crm/v3/objects/companies?limit=10&archived=false&hapikey=%s", hsc.APIKey), nil)
	if err != nil {
		return companies, err
	}
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{Timeout: hsc.TimeOut}
	resp, err := client.Do(req)
	if err != nil {
		return companies, err
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return companies, err
	}

	var response FetchAllCompanyEntriesResponse

	if err = json.Unmarshal(buf.Bytes(), &response); err != nil {
		return companies, err
	}

	for _, result := range response.Results {
		companies = append(companies, result.Properties)
	}

	// Check if the status code is outside the range of possible 200 level responses
	if resp.StatusCode-200 > 26 {
		return companies, errors.New("HubSpot returned non-200 status code")
	}

	return companies, nil
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

func (hsc *HubSpotClient) CreateHubSpotDealEntry(action string) error {
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
