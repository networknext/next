package transport

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// IPStackClient wraps an *http.Client to perform IP lookups to ipstack.com with an AccessKey
type IPStackClient struct {
	*http.Client

	AccessKey string
}

// Lookup performs a GET request to api.ipstack.com and decodes the response to an IPStackResponse
func (c *IPStackClient) Lookup(ip string) (*IPStackResponse, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://api.ipstack.com/%s?access_key=%s&output=json", ip, c.AccessKey), nil)
	if err != nil {
		return nil, err
	}

	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	var response IPStackResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

// IPStackResponse is the decoded JSON from api.ipstack.com
type IPStackResponse struct {
	IP            string  `json:"ip"`
	Type          string  `json:"type"`
	ContinentCode string  `json:"continent_code"`
	ContinentName string  `json:"continent_name"`
	CountryCode   string  `json:"country_code"`
	CountryName   string  `json:"country_name"`
	RegionCode    string  `json:"region_code"`
	RegionName    string  `json:"region_name"`
	City          string  `json:"city"`
	ZIP           string  `json:"zip"`
	Latitude      float32 `json:"latitude"`
	Longitude     float32 `json:"longitude"`
	Location      struct {
		GeonameID int    `json:"geoname_id"`
		Capital   string `json:"capital"`
		Languages []struct {
			Code   string `json:"code"`
			Name   string `json:"name"`
			Native string `json:"native"`
		} `json:"languages"`
		CountryFlag             string `json:"country_flag"`
		CountryFlagEmoji        string `json:"country_flag_emoji"`
		CountryFlagEmojiUnicode string `json:"country_flag_emoji_unicode"`
		CallingCode             string `json:"calling_code"`
		IsEU                    bool   `json:"is_eu"`
	} `json:"location"`
	Timezone struct {
		ID               string `json:"id"`
		CurrentTime      string `json:"current_time"`
		GMTOffset        int    `json:"gmt_offset"`
		Code             string `json:"code"`
		IsDaylightSaving bool   `json:"is_daylight_saving"`
	} `json:"time_zone"`
	Currency struct {
		Code         string `json:"code"`
		Name         string `json:"name"`
		Plural       string `json:"plural"`
		Symbol       string `json:"symbol"`
		SymbolNative string `json:"symbol_native"`
	} `json:"currency"`
	Connection struct {
		ASN int    `json:"asn"`
		ISP string `json:"isp"`
	} `json:"connection"`
	Security struct {
		IsProxy     bool    `json:"is_proxy"`
		ProxyType   *string `json:"proxy_type"`
		IsCrawler   bool    `json:"is_crawler"`
		CrawlerName *string `json:"crawler_name"`
		CrawlerType *string `json:"crawler_type"`
		IsTor       bool    `json:"is_tor"`
		ThreatLevel string  `json:"threat_level"`
		ThreatType  *string `json:"threat_type"`
	}
}
