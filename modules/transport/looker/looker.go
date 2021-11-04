package looker

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	LOOKER_SESSION_TIMEOUT = 86400
	EMBEDDED_USER_GROUP_ID = 3
	USAGE_DASH_URL         = "/embed/dashboards-next/11"
)

type LookerClient struct {
	HostURL string
	Secret  string
}

func NewLookerClient(hostURL string, secret string) (*LookerClient, error) {
	if hostURL == "" || secret == "" {
		return nil, fmt.Errorf("Host and Secret are required")
	}

	return &LookerClient{
		HostURL: hostURL,
		Secret:  secret,
	}, nil
}

// TODO: Get these from storage at some point
type AnalyticsDashboardCategory struct {
	ID      int64
	Label   string
	Premium bool
}

type AnalyticsDashboard struct {
	ID           int64
	Name         string
	Discovery    bool
	DashboardID  int64
	CustomerCode string
	Category     AnalyticsDashboardCategory
}

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

func (l *LookerClient) GenerateUsageDashboardURL(userID string, customerCode string) (string, error) {
	nonce, err := GenerateRandomString(16)
	if err != nil {
		return "", err
	}

	urlOptions := LookerURLOptions{
		Host:            l.HostURL,
		Secret:          l.Secret,
		ExternalUserId:  fmt.Sprintf("\"%s\"", userID),
		GroupsIds:       []int{EMBEDDED_USER_GROUP_ID},
		ExternalGroupId: "",
		Permissions:     []string{"access_data", "see_looks", "see_user_dashboards"}, // TODO: This may or may not need to change
		Models:          []string{"networknext_prod"},                                // TODO: This may or may not need to change
		AccessFilters:   make(map[string]map[string]interface{}),
		UserAttributes:  make(map[string]interface{}),
		SessionLength:   LOOKER_SESSION_TIMEOUT,
		EmbedURL:        "/login/embed/" + url.QueryEscape(USAGE_DASH_URL),
		ForceLogout:     true,
		Nonce:           fmt.Sprintf("\"%s\"", nonce),
		Time:            time.Now().Unix(),
	}

	urlOptions.UserAttributes["customer_code"] = customerCode

	return BuildLookerURL(urlOptions), nil
}

func (l *LookerClient) GenerateAnalyticsCategories(userID string, customerCode string, showPremium bool) error {
	// TODO: Implement with storer
	return nil
}

func (l *LookerClient) GenerateLookerTrialURL(requestID string) string {
	nonce, err := GenerateRandomString(16)
	if err != nil {
		return ""
	}

	options := LookerURLOptions{
		Host:            l.HostURL,
		Secret:          l.Secret,
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

	return BuildLookerURL(options)
}

func GenerateRandomString(n int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"
	bytes, err := GenerateRandomBytes(n)
	if err != nil {
		// not method, no service logger
		return "", err
	}
	for i, b := range bytes {
		bytes[i] = letters[b%byte(len(letters))]
	}
	return string(bytes), nil
}

func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		// not method, no service logger
		return nil, err
	}

	return b, nil
}
