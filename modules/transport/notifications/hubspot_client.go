package notifications

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// TODO: clean this up and make things a bit more generic for easier use and expansion in the future
const (
	DefaultHubSpotTimeout = 10 * time.Second
	// Const Company ID for new sign up catch all company in hubspot
	NewFunnelCoID = "7152708198"
	// Const Association Types from hubspot API docs: https://legacydocs.hubspot.com/docs/methods/crm-associations/crm-associations-overview
	ContactToCompany            = "contact_to_company"
	CompanyToContact            = "company_to_contact"
	DealToContact               = "deal_to_contact"
	ContactToDeal               = "contact_to_deal"
	DealToCompany               = "deal_to_company"
	CompanyToDeal               = "company_to_deal"
	CompanyToEngagement         = "company_to_engagement"
	EngagementToCompany         = "engagement_to_company"
	ContactToEngagement         = "contact_to_engagement"
	EngagementToContact         = "engagement_to_contact"
	DealToEngagement            = "deal_to_engagement"
	EngagementToDeal            = "parent_company_to_child_company"
	ParentCompanyToChildCompany = "parent_company_to_child_company"
	ChildCompanyToParentCompany = "child_company_to_parent_company"
	ContactToTicket             = "contact_to_ticket"
	TicketToContact             = "ticket_to_contact"
	TicketToEngagement          = "ticket_to_engagement"
	EngagementToTicket          = "engagement_to_ticket"
	DealToLineItem              = "deal_to_line_item"
	LineItemToDeal              = "line_item_to_deal"
	CompanyToTicket             = "company_to_ticket"
	TicketToCompany             = "ticket_to_company"
	DealToTicket                = "deal_to_ticket"
	TicketToDeal                = "ticket_to_deal"
)

type HubSpotClient struct {
	APIKey  string
	TimeOut time.Duration
}

func NewHubSpotClient(APIKey string, timeout time.Duration) (*HubSpotClient, error) {
	return &HubSpotClient{
		APIKey:  APIKey,
		TimeOut: timeout,
	}, nil
}

type SearchProperties struct {
	FilterGroups []FilterGroup `json:"filterGroups,omitempty"`
	Sorts        []string      `json:"sorts,omitempty"`
	Query        string        `json:"query,omitempty"`
	Properties   []string      `json:"properties,omitempty"`
	Limit        int           `json:"limit,omitempty"`
	After        int           `json:"after,omitempty"`
}
type FilterGroup struct {
	Filters []Filter `json:"filters,omitempty"`
}

type Filter struct {
	Value       string `json:"value,omitempty"`
	PropertName string `json:"propertyName,omitempty"`
	Operator    string `json:"operator,omitempty"`
}

type CompanyProperties struct {
	City                    string `json:"city,omitempty"`
	CreateDate              string `json:"createdate,omitempty"`
	Domain                  string `json:"domain,omitempty"`
	HubspotLastModifiedDate string `json:"hs_lastmodifieddate,omitempty"`
	Industry                string `json:"industry,omitempty"`
	Name                    string `json:"name,omitempty"`
	Phone                   string `json:"phone,omitempty"`
	State                   string `json:"state,omitempty"`
}

type CompanyResults struct {
	ID         string            `json:"id,omitempty"`
	Properties CompanyProperties `json:"properties,omitempty"`
	CreatedAt  string            `json:"createdAt,omitempty"`
	UpdatedAt  string            `json:"updatedAt,omitempty"`
	Archived   bool              `json:"archived,omitempty"`
}

type CompanyEntrySearchResponse struct {
	Results []CompanyResults `json:"results,omitempty"`
}

func (hsc *HubSpotClient) CompanyEntrySearch(companyName string, companyWebsite string) ([]CompanyResults, error) {
	noCompanies := make([]CompanyResults, 0)

	// Hubspot isn't enabled for this env
	if hsc.APIKey == "" {
		return noCompanies, nil
	}

	searchProperties := SearchProperties{
		FilterGroups: []FilterGroup{
			{
				Filters: []Filter{
					{
						Value:       companyName,
						PropertName: "name",
						Operator:    "EQ",
					},
					{
						Value:       companyWebsite,
						PropertName: "domain",
						Operator:    "EQ",
					},
				},
			},
		},
	}

	hubspotBody, _ := json.Marshal(searchProperties)
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("https://api.hubapi.com/crm/v3/objects/companies/search?hapikey=%s", hsc.APIKey), bytes.NewBuffer(hubspotBody))
	if err != nil {
		return noCompanies, err
	}
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{Timeout: hsc.TimeOut}
	resp, err := client.Do(req)
	if err != nil {
		return noCompanies, err
	}

	// Check if the status code is outside the range of possible 200 level responses
	if resp.StatusCode-200 > 26 {
		return noCompanies, errors.New("HubSpot returned non-200 status code")
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return noCompanies, err
	}

	var response CompanyEntrySearchResponse
	if err = json.Unmarshal(buf.Bytes(), &response); err != nil {
		return noCompanies, err
	}

	return response.Results, nil
}

type CreateCompanyMessage struct {
	Properties CompanyProperties `json:"properties,omitempty"`
}

type CreateCompanyEntriesResponse struct {
	Results CompanyResults `json:"results,omitempty"`
}

func (hsc *HubSpotClient) CreateNewCompanyEntry(companyName string, companyWebsite string) (string, error) {
	// Hubspot isn't enabled for this env
	if hsc.APIKey == "" {
		return "", nil
	}

	properties := CompanyProperties{
		Domain: companyWebsite,
		Name:   companyName,
	}

	message := CreateCompanyMessage{
		Properties: properties,
	}
	hubspotBody, _ := json.Marshal(message)
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("https://api.hubapi.com/crm/v3/objects/companies?hapikey=%s", hsc.APIKey), bytes.NewBuffer(hubspotBody))
	if err != nil {
		return "", err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{Timeout: hsc.TimeOut}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	// Check if the status code is outside the range of possible 200 level responses
	if resp.StatusCode-200 > 26 {
		return "", errors.New("HubSpot returned non-200 status code")
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return "", err
	}

	var response CreateCompanyEntriesResponse
	if err = json.Unmarshal(buf.Bytes(), &response); err != nil {
		return "", err
	}

	return response.Results.ID, nil
}

type ContactProperties struct {
	Company   string `json:"company,omitempty"`
	FirstName string `json:"firstname,omitempty"`
	LastName  string `json:"lastname,omitempty"`
	Email     string `json:"email,omitempty"`
	Website   string `json:"website,omitempty"`
}

type ContactEntrySearchResponse struct {
	Total   int              `json:"total,omitempty"`
	Results []ContactResults `json:"results,omitempty"`
}

type ContactResults struct {
	ID         string            `json:"id,omitempty"`
	Properties ContactProperties `json:"properties,omitempty"`
	CreatedAt  string            `json:"createdAt,omitempty"`
	UpdatedAt  string            `json:"updatedAt,omitempty"`
	Archived   bool              `json:"archived,omitempty"`
}

func (hsc *HubSpotClient) ContactEntrySearch(firstName string, lastName string, email string) ([]ContactResults, error) {
	noContacts := make([]ContactResults, 0)

	// Hubspot isn't enabled for this env
	if hsc.APIKey == "" {
		return noContacts, nil
	}

	searchProperties := SearchProperties{
		FilterGroups: []FilterGroup{
			{
				Filters: []Filter{
					{
						Value:       firstName,
						PropertName: "firstname",
						Operator:    "EQ",
					},
					{
						Value:       lastName,
						PropertName: "lastname",
						Operator:    "EQ",
					},
					{
						Value:       email,
						PropertName: "email",
						Operator:    "EQ",
					},
				},
			},
		},
	}

	hubspotBody, _ := json.Marshal(searchProperties)
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("https://api.hubapi.com/crm/v3/objects/contacts/search?hapikey=%s", hsc.APIKey), bytes.NewBuffer(hubspotBody))
	if err != nil {
		return noContacts, err
	}
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{Timeout: hsc.TimeOut}
	resp, err := client.Do(req)
	if err != nil {
		return noContacts, err
	}

	// Check if the status code is outside the range of possible 200 level responses
	if resp.StatusCode-200 > 26 {
		return noContacts, errors.New("HubSpot returned non-200 status code")
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return noContacts, err
	}

	var response ContactEntrySearchResponse
	if err = json.Unmarshal(buf.Bytes(), &response); err != nil {
		return noContacts, err
	}

	return response.Results, nil
}

type CreateContactMessage struct {
	Properties ContactProperties `json:"properties,omitempty"`
}

func (hsc *HubSpotClient) CreateNewContactEntry(firstName string, lastName string, email string, companyName string, companyWebsite string) (string, error) {
	// Hubspot isn't enabled for this env
	if hsc.APIKey == "" {
		return "", nil
	}

	properties := ContactProperties{
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		Company:   companyName,
	}

	message := CreateContactMessage{
		Properties: properties,
	}
	hubspotBody, _ := json.Marshal(message)
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("https://api.hubapi.com/crm/v3/objects/contacts?hapikey=%s", hsc.APIKey), bytes.NewBuffer(hubspotBody))
	if err != nil {
		return "", err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{Timeout: hsc.TimeOut}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	// Check if the status code is outside the range of possible 200 level responses
	if resp.StatusCode-200 > 26 {
		return "", errors.New("HubSpot returned non-200 status code")
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return "", err
	}

	var response ContactResults
	if err = json.Unmarshal(buf.Bytes(), &response); err != nil {
		return "", err
	}

	return response.ID, nil
}

type AssociateCompanyAndContactMessage struct {
	Properties AssociateCompanyAndContactProperties `json:"properties,omitempty"`
}

type AssociateCompanyAndContactProperties struct {
	Inputs []AssociationInput `json:"inputs,omitempty"`
}

func (hsc *HubSpotClient) AssociateCompanyToContact(companyID string, contactID string) error {
	// Hubspot isn't enabled for this env
	if hsc.APIKey == "" {
		return nil
	}

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("https://api.hubapi.com/crm/v3/objects/companies/%s/associations/contacts/%s/%s?hapikey=%s", companyID, contactID, CompanyToContact, hsc.APIKey), nil)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{Timeout: hsc.TimeOut}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	// Check if the status code is outside the range of possible 200 level responses
	if resp.StatusCode-200 > 26 {
		return errors.New("HubSpot returned non-200 status code")
	}

	// TODO: Catch response and check if there was a success that way rather than through response code

	return nil
}

func (hsc *HubSpotClient) AssociateContactToCompany(contactID string, companyID string) error {
	// Hubspot isn't enabled for this env
	if hsc.APIKey == "" {
		return nil
	}

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("https://api.hubapi.com/crm/v3/objects/contacts/%s/associations/companies/%s/%s?hapikey=%s", contactID, companyID, ContactToCompany, hsc.APIKey), nil)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{Timeout: hsc.TimeOut}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	// Check if the status code is outside the range of possible 200 level responses
	if resp.StatusCode-200 > 26 {
		return errors.New("HubSpot returned non-200 status code")
	}

	// TODO: Catch response and check if there was a success that way rather than through response code

	return nil
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
	// Hubspot isn't enabled for this env
	if hsc.APIKey == "" {
		return nil
	}

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

	// Check if the status code is outside the range of possible 200 level responses
	if resp.StatusCode-200 > 26 {
		return errors.New("HubSpot returned non-200 status code")
	}

	// TODO: Catch response and check if there was a success that way rather than through response code

	return nil
}

type EngagementProperties struct {
	Active  bool   `json:"active,omitempty"`
	OwnerID string `json:"ownerId,omitempty"`
	Type    string `json:"type,omitempty"`
}

type EngagementMetaData struct {
	Body string `json:"body,omitempty"`
}

type AssociationProperties struct {
	ContactIDs []string `json:"contactIds,omitempty"`
	CompanyIDs []string `json:"companyIds,omitempty"`
	DealIDs    []string `json:"dealIds,omitempty"`
	OwnerIDs   []string `json:"ownerIds,omitempty"`
	TicketIDs  []string `json:"ticketIds,omitempty"`
}

type AssociationInput struct {
	From Association `json:"from,omitempty"`
	To   Association `json:"to,omitempty"`
	Type string      `json:"type,omitempty"`
}

type Association struct {
	ID string `json:"id,omitempty"`
}
type CreateEngagementMessage struct {
	Engagement  EngagementProperties  `json:"engagement,omitempty"`
	Association AssociationProperties `json:"associations,omitempty"`
	MetaData    EngagementMetaData    `json:"metadata,omitempty"`
}

func (hsc *HubSpotClient) CreateCompanyNote(message string, companyID string) error {
	// Hubspot isn't enabled for this env
	if hsc.APIKey == "" {
		return nil
	}

	engagementPayload := CreateEngagementMessage{
		Engagement: EngagementProperties{
			Active: true,
			Type:   "NOTE",
		},
		Association: AssociationProperties{
			CompanyIDs: []string{
				companyID,
			},
		},
		MetaData: EngagementMetaData{
			Body: message,
		},
	}

	hubspotBody, _ := json.Marshal(engagementPayload)
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("https://api.hubapi.com/engagements/v1/engagements?hapikey=%s", hsc.APIKey), bytes.NewBuffer(hubspotBody))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{Timeout: hsc.TimeOut}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	// Check if the status code is outside the range of possible 200 level responses
	if resp.StatusCode-200 > 26 {
		return errors.New("HubSpot returned non-200 status code")
	}

	// TODO: Catch response and check if there was a success that way rather than through response code

	return nil
}

func (hsc *HubSpotClient) CreateContactNote(message string, contactID string) error {
	// Hubspot isn't enabled for this env
	if hsc.APIKey == "" {
		return nil
	}

	engagementPayload := CreateEngagementMessage{
		Engagement: EngagementProperties{
			Active: true,
			Type:   "NOTE",
		},
		Association: AssociationProperties{
			ContactIDs: []string{
				contactID,
			},
		},
		MetaData: EngagementMetaData{
			Body: message,
		},
	}

	hubspotBody, _ := json.Marshal(engagementPayload)
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("https://api.hubapi.com/engagements/v1/engagements?hapikey=%s", hsc.APIKey), bytes.NewBuffer(hubspotBody))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{Timeout: hsc.TimeOut}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	// Check if the status code is outside the range of possible 200 level responses
	if resp.StatusCode-200 > 26 {
		return errors.New("HubSpot returned non-200 status code")
	}

	// TODO: Catch response and check if there was a success that way rather than through response code

	return nil
}
