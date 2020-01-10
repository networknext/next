package transport_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/networknext/backend/transport"
)

func TestIPStackClient(t *testing.T) {
	ipStackClient := transport.IPStackClient{
		Client: NewTestHTTPClient(func(req *http.Request) *http.Response {
			return &http.Response{
				StatusCode: 200,
				Header:     make(http.Header),
				Body: ioutil.NopCloser(bytes.NewBufferString(`{
					"ip": "0.0.0.0",
					"continent_code": "NA",
					"country_code": "US",
					"region_code": "NY",
					"city": "Troy",
					"latitude": 43.05036163330078,
					"longitude": -73.75393676757812,
					"connection": {
						"asn": 11351,
						"isp": "Charter Communications Inc"
					}
				}`)),
			}
		}),
	}

	expected := transport.IPStackResponse{
		IP:            "0.0.0.0",
		ContinentCode: "NA",
		CountryCode:   "US",
		RegionCode:    "NY",
		City:          "Troy",
		Latitude:      43.05036163330078,
		Longitude:     -73.75393676757812,
	}
	expected.Connection.ASN = 11351
	expected.Connection.ISP = "Charter Communications Inc"

	actual, err := ipStackClient.Lookup("0.0.0.0")
	assert.NoError(t, err)
	assert.Equal(t, expected, *actual)
}
