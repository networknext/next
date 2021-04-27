package jsonrpc

import (
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/tidwall/gjson"
)

// RelayFleetService provides access to real-time data provided by the endpoints
// mounted on the relay_frontend (/relays, /cost_matrix (tbd), etc.).
type RelayFleetService struct {
	RelayFrontendURI string
	Logger           log.Logger
}

// RelayFleetEntry represents a line in the CSV file provided
// by relay_frontend/relays (all strings)
type RelayFleetEntry struct {
	Name     string `json:"name"`
	Address  string `json:"address"`
	Id       string `json:"hex_id"`
	Status   string `json:"status"`
	Sessions string `json:"sessions"`
	Version  string `json:"version"`
}

type FleetArgs struct{}

type FleetReply struct {
	RelayFleet []RelayFleetEntry `json:"relay_fleet"`
}

// RelayFleet retrieves the CSV file from relay_frontend/relays, converts it to
// json and puts it on the wire.
func (rfs *RelayFleetService) RelayFleet(r *http.Request, args *FleetArgs, reply *FleetReply) error {

	authToken, err := GetOpsToken()
	if err != nil {
		err = fmt.Errorf("RelayFleet() error getting auth token: %w", err)
		rfs.Logger.Log("err", err)
		return err
	}

	uri := rfs.RelayFrontendURI + "/relays"
	fmt.Println("uri:", uri)

	client := &http.Client{}
	req, _ := http.NewRequest("GET", uri, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

	response, err := client.Do(req)
	if err != nil {
		err = fmt.Errorf("RelayFleet() error getting relays.csv: %w", err)
		rfs.Logger.Log("err", err)
		return err
	}
	defer response.Body.Close()

	reader := csv.NewReader(response.Body)
	relayData, err := reader.ReadAll()
	if err != nil {
		err = fmt.Errorf("RelayFleet() could not parse relays csv file from %s: %v", uri, err)
		rfs.Logger.Log("err", err)
		return err
	}

	// drop headings row
	relayData = append(relayData[:0], relayData[1:]...)

	var returnFleetObject []RelayFleetEntry

	for _, relayDataEntry := range relayData {

		relayFleetEntry := RelayFleetEntry{
			Name:     relayDataEntry[0],
			Address:  relayDataEntry[1],
			Id:       relayDataEntry[2],
			Status:   relayDataEntry[3],
			Sessions: relayDataEntry[4],
			Version:  relayDataEntry[5],
		}
		returnFleetObject = append(returnFleetObject, relayFleetEntry)
	}

	reply.RelayFleet = returnFleetObject

	return nil
}

// GetOpsToken is a hack to get a usable token since we can't get the
// Authorization header from the request.
//
// TODO: This function can be removed once relay_frontend/relays
// has been moved to an internal IP address.
//
// See issue #3030: https://github.com/networknext/backend/issues/3030
func GetOpsToken() (string, error) {
	req, err := http.NewRequest(
		http.MethodPost,
		"https://networknext.auth0.com/oauth/token",
		strings.NewReader(`{
				"client_id":"6W6PCgPc6yj6tzO9PtW6IopmZAWmltgb",
				"client_secret":"EPZEHccNbjqh_Zwlc5cSFxvxFQHXZ990yjo6RlADjYWBz47XZMf-_JjVxcMW-XDj",
				"audience":"https://portal.networknext.com",
				"grant_type":"client_credentials"
			}`),
	)
	if err != nil {
		return "", err
	}

	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("auth0 returned code %d", res.StatusCode)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	token := gjson.ParseBytes(body).Get("access_token").String()

	return token, nil

}
