package jsonrpc

import (
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/go-kit/kit/log"
	"github.com/networknext/backend/modules/routing"
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

type RelayFleetArgs struct{}

type RelayFleetReply struct {
	RelayFleet []RelayFleetEntry `json:"relay_fleet"`
}

// RelayFleet retrieves the CSV file from relay_frontend/relays, converts it to
// json and puts it on the wire.
func (rfs *RelayFleetService) RelayFleet(r *http.Request, args *RelayFleetArgs, reply *RelayFleetReply) error {
	authHeader := r.Header.Get("Authorization")

	uri := rfs.RelayFrontendURI + "/relays"

	client := &http.Client{}
	req, _ := http.NewRequest("GET", uri, nil)
	req.Header.Set("Authorization", authHeader)

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

type RelayDashboardJsonReply struct {
	Dashboard string `json:"relay_dashboard"`
}

type RelayDashboardJsonArgs struct{}

// RelayDashboardJson retrieves the JSON representation of the current relay dashboard
// provided by relay_backend/relay_dashboard_data
func (rfs *RelayFleetService) RelayDashboardJson(r *http.Request, args *RelayDashboardJsonArgs, reply *RelayDashboardJsonReply) error {

	type jsonRelay struct {
		Name string
		Addr string
	}

	type jsonResponse struct {
		Analysis string
		Relays   []jsonRelay
		Stats    map[string]map[string]routing.Stats
	}

	authHeader := r.Header.Get("Authorization")

	uri := rfs.RelayFrontendURI + "/relay_dashboard_data"

	client := &http.Client{}
	req, _ := http.NewRequest("GET", uri, nil)
	req.Header.Set("Authorization", authHeader)

	response, err := client.Do(req)
	if err != nil {
		err = fmt.Errorf("RelayDashboardJson() error getting relays.csv: %w", err)
		rfs.Logger.Log("err", err)
		return err
	}
	defer response.Body.Close()

	dashboardData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		err = fmt.Errorf("RelayDashboardJson() error getting reading HTTP response body: %w", err)
		rfs.Logger.Log("err", err)
		return err
	}

	reply.Dashboard = string(dashboardData)

	return nil
}
