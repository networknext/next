package notifications

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type NotificationPriorty int32
type NotificationType int32

const (
	LOOKER_HOST             = "networknextexternal.cloud.looker.com"
	ANALYTICS_TRIAL_TITLE   = "Data analytics trial opportunity"
	ANALYTICS_TRIAL_MESSAGE = "Super awesome marketing message that explains this is a no strings attached trial. Clicking the link will give you free trial access to our analytics service"

	// TODO: Move these somewhere else like the jsonrpc error codes and use them for something
	DEFAULT_PRIORITY NotificationPriorty = 0
	INFO_PRIORITY    NotificationPriorty = 1
	WARNING_PRIORITY NotificationPriorty = 2
	URGENT_PRIORITY  NotificationPriorty = 3
	// TODO: Move these somewhere else like the jsonrpc error codes and actually use them for something
	NOTIFICATION_SYSTEM        NotificationType = 0
	NOTIFICATION_RELEASE_NOTES NotificationType = 1
	NOTIFICATION_ANALYTICS     NotificationType = 2
	NOTIFICATION_INVOICE       NotificationType = 3
)

type LookerURLOptions struct {
	Secret          string                            //required
	Host            string                            //required
	EmbedURL        string                            //required
	Nonce           string                            //required
	Time            int64                             //required
	SessionLength   int                               //required
	ExternalUserId  string                            //required
	Permissions     []string                          //required
	Models          []string                          //required
	ForceLogout     bool                              //required
	GroupsIds       []int                             //optional
	ExternalGroupId string                            //optional
	UserAttributes  map[string]interface{}            //optional
	AccessFilters   map[string]map[string]interface{} //required
	FirstName       string                            //optional
	LastName        string                            //optional
}

func BuildLookerURL(urlOptions LookerURLOptions) string {
	// TODO: Verify logic below, this came from here: https://github.com/looker/looker_embed_sso_examples/pull/36 and is NOT an official implementation. That being said, be careful changing it because it works :P
	jsonPerms, _ := json.Marshal(urlOptions.Permissions)
	jsonModels, _ := json.Marshal(urlOptions.Models)
	jsonUserAttrs, _ := json.Marshal(urlOptions.UserAttributes)
	jsonFilters, _ := json.Marshal(urlOptions.AccessFilters)
	jsonGroupIds, _ := json.Marshal(urlOptions.GroupsIds)
	strTime := strconv.Itoa(int(urlOptions.Time))
	strSessionLen := strconv.Itoa(urlOptions.SessionLength)
	strForceLogin := strconv.FormatBool(urlOptions.ForceLogout)

	strToSign := strings.Join([]string{urlOptions.Host,
		urlOptions.EmbedURL,
		urlOptions.Nonce,
		strTime,
		strSessionLen,
		urlOptions.ExternalUserId,
		string(jsonPerms),
		string(jsonModels)}, "\n")

	strToSign = strToSign + "\n"

	if len(urlOptions.GroupsIds) > 0 {
		strToSign = strToSign + string(jsonGroupIds) + "\n"
	}

	if urlOptions.ExternalGroupId != "" {
		strToSign = strToSign + urlOptions.ExternalGroupId + "\n"
	}

	if len(urlOptions.UserAttributes) > 0 {
		strToSign = strToSign + string(jsonUserAttrs) + "\n"
	}

	strToSign = strToSign + string(jsonFilters)

	h := hmac.New(sha1.New, []byte(urlOptions.Secret))
	h.Write([]byte(strToSign))
	encoded := base64.StdEncoding.EncodeToString(h.Sum(nil))

	query := url.Values{}
	query.Add("nonce", urlOptions.Nonce)
	query.Add("time", strTime)
	query.Add("session_length", strSessionLen)
	query.Add("external_user_id", urlOptions.ExternalUserId)
	query.Add("permissions", string(jsonPerms))
	query.Add("models", string(jsonModels))
	query.Add("access_filters", string(jsonFilters))
	query.Add("first_name", urlOptions.FirstName)
	query.Add("last_name", urlOptions.LastName)
	query.Add("force_logout_login", strForceLogin)
	query.Add("signature", encoded)

	if len(urlOptions.GroupsIds) > 0 {
		query.Add("group_ids", string(jsonGroupIds))
	}

	if urlOptions.ExternalGroupId != "" {
		query.Add("external_group_id", urlOptions.ExternalGroupId)
	}

	if len(urlOptions.UserAttributes) > 0 {
		query.Add("user_attributes", string(jsonUserAttrs))
	}

	finalUrl := fmt.Sprintf("https://%s%s?%s", urlOptions.Host, urlOptions.EmbedURL, query.Encode())

	return finalUrl
}

type GistEmbed struct {
	CSSURL    string
	EmbedHTML string
}

type ReleaseNotesNotification struct {
	Type      NotificationType    `json:"type"`
	Title     string              `json:"title"`
	Priority  NotificationPriorty `json:"priority"`
	CSSURL    string              `json:"css_url"`
	EmbedHTML string              `json:"embed_html"`
}

func NewReleaseNotesNotification() ReleaseNotesNotification {
	return ReleaseNotesNotification{
		Type:     NOTIFICATION_RELEASE_NOTES,
		Priority: DEFAULT_PRIORITY,
	}
}

type SystemNotification struct {
	Type     NotificationType    `json:"type"`
	Title    string              `json:"title"`
	Message  string              `json:"message"`
	Priority NotificationPriorty `json:"priority"`
}

func NewSystemNotifications() SystemNotification {
	return SystemNotification{
		Type:     NOTIFICATION_SYSTEM,
		Priority: DEFAULT_PRIORITY,
	}
}

type AnalyticsNotification struct {
	Type      NotificationType    `json:"type"`
	Title     string              `json:"title"`
	Message   string              `json:"message"`
	Priority  NotificationPriorty `json:"priority"`
	LookerURL string              `json:"looker_url"`
}

func NewAnalyticsNotification() AnalyticsNotification {
	return AnalyticsNotification{
		Type:     NOTIFICATION_ANALYTICS,
		Priority: DEFAULT_PRIORITY,
	}
}

func NewTrialAnalyticsNotification(lookerSecret string, nonce string, requestID string) AnalyticsNotification {
	notification := NewAnalyticsNotification()

	options := LookerURLOptions{
		Host:            LOOKER_HOST,
		Secret:          lookerSecret,
		ExternalUserId:  fmt.Sprintf("\"%s\"", requestID),
		FirstName:       "",
		LastName:        "",
		GroupsIds:       make([]int, 0),
		ExternalGroupId: "",
		Permissions:     []string{"access_data", "see_looks", "see_user_dashboards"}, // TODO: Verify these are final
		Models:          []string{"networknext_pbl"},                                 // TODO: Verify these are final
		AccessFilters:   make(map[string]map[string]interface{}),
		UserAttributes:  make(map[string]interface{}),
		SessionLength:   3600,
		EmbedURL:        "/login/embed/" + url.QueryEscape("/embed/dashboards-next/?"), // TODO: Replace the ? with the correct dash ID
		ForceLogout:     true,
		Nonce:           fmt.Sprintf("\"%s\"", nonce),
		Time:            time.Now().Unix(),
	}

	notification.Title = ANALYTICS_TRIAL_TITLE
	notification.Message = ANALYTICS_TRIAL_MESSAGE
	notification.LookerURL = BuildLookerURL(options)

	return notification
}

type InvoiceNotification struct {
	Type      NotificationType    `json:"type"`
	Title     string              `json:"title"`
	Message   string              `json:"message"`
	Priority  NotificationPriorty `json:"priority"`
	InvoiceID string              `json:"invoice_id"`
}

func NewInvoiceNotification() InvoiceNotification {
	return InvoiceNotification{
		Title:    fmt.Sprintf("Invoice for the month of %s", time.Now().Month()),
		Message:  "Your invoice is ready and can be viewed in the \"Invoicing\" tab",
		Type:     NOTIFICATION_INVOICE,
		Priority: DEFAULT_PRIORITY,
	}
}
