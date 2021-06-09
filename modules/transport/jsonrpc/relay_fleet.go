package jsonrpc

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/networknext/backend/modules/routing"
)

// RelayFleetService provides access to real-time data provided by the endpoints
// mounted on the relay_frontend (/relays, /cost_matrix (tbd), etc.).
type RelayFleetService struct {
	RelayFrontendURI  string
	RelayGatewayURI   string
	RelayForwarderURI string
	Env               string
	Logger            log.Logger
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
	// Dashboard string `json:"relay_dashboard"`
	Dashboard jsonResponse `json:"relay_dashboard"`
}

type RelayDashboardJsonArgs struct {
	XAxisRelayFilter string `json:"xAxisFilters"`
	YAxisRelayFilter string `json:"yAxisFilters"`
}

type jsonRelay struct {
	Name string
	Addr string
}

type jsonResponse struct {
	Analysis routing.JsonMatrixAnalysis
	Relays   []jsonRelay
	Stats    map[string]map[string]routing.Stats
}

// RelayDashboardJson retrieves the JSON representation of the current relay dashboard
// provided by relay_backend/relay_dashboard_data
func (rfs *RelayFleetService) RelayDashboardJson(r *http.Request, args *RelayDashboardJsonArgs, reply *RelayDashboardJsonReply) error {

	if args.XAxisRelayFilter == "" || args.YAxisRelayFilter == "" {
		err := fmt.Errorf("a filter must be supplied for each axis")
		rfs.Logger.Log("err", err)
		return err
	}

	var fullDashboard, filteredDashboard jsonResponse
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

	byteValue, err := ioutil.ReadAll(response.Body)
	if err != nil {
		err = fmt.Errorf("RelayDashboardJson() error getting reading HTTP response body: %w", err)
		rfs.Logger.Log("err", err)
		return err
	}

	json.Unmarshal(byteValue, &fullDashboard)
	if len(fullDashboard.Relays) == 0 {
		err := fmt.Errorf("relay backend returned an empty dashboard file")
		rfs.Logger.Log("err", err)
		return err
	}

	filteredDashboard.Analysis = fullDashboard.Analysis
	filteredDashboard.Stats = make(map[string]map[string]routing.Stats)

	// x-axis relays first
	xFilters := strings.Split(args.XAxisRelayFilter, ",")
	for _, xFilter := range xFilters {
		xAxisRegex := regexp.MustCompile("(?i)" + strings.TrimSpace(xFilter))
		for _, relayEntry := range fullDashboard.Relays {
			if xAxisRegex.MatchString(relayEntry.Name) {
				filteredDashboard.Relays = append(filteredDashboard.Relays, relayEntry)
				continue
			}
		}
	}

	if len(filteredDashboard.Relays) == 0 {
		err := fmt.Errorf("no matches found for x-axis query string")
		rfs.Logger.Log("err", err)
		return err
	}

	// then the y-axis
	yFilters := strings.Split(args.YAxisRelayFilter, ",")
	for _, yFilter := range yFilters {
		yAxisRegex := regexp.MustCompile("(?i)" + strings.TrimSpace(yFilter))
		for yAxisRelayName, relayStatsLine := range fullDashboard.Stats {
			if yAxisRegex.MatchString(yAxisRelayName) {
				filteredDashboard.Stats[yAxisRelayName] = make(map[string]routing.Stats)
				for relayName, statsLineEntry := range relayStatsLine {
					for _, xFilter := range xFilters {
						xAxisRegex := regexp.MustCompile("(?i)" + strings.TrimSpace(xFilter))
						if xAxisRegex.MatchString(relayName) {
							filteredDashboard.Stats[yAxisRelayName][relayName] = statsLineEntry
						}
					}
				}
			}
		}
	}

	if len(filteredDashboard.Stats) == 0 {
		err := fmt.Errorf("no matches found for y-axis query string")
		rfs.Logger.Log("err", err)
		return err
	}

	reply.Dashboard = filteredDashboard

	return nil
}

type AdminFrontPageArgs struct{}

type AdminFrontPageReply struct {
	BinFileCreationTime  time.Time `json:"binFileCreationTime"`
	BinFileCreator       string    `json:"binFileCreator"`
	RelayGatewayStatus   []string  `json:"relayGatewayStatus"`
	RelayFrontEndStatus  []string  `json:"relayFrontEndStatus"`
	RelayBackEndStatus   []string  `json:"relayBackEndStatus"`
	RelayForwarderStatus []string  `json:"relayForwarderStatus"`
}

// RelayDashboardJson retrieves the JSON representation of the current relay dashboard
// provided by relay_backend/relay_dashboard_data
func (rfs *RelayFleetService) AdminFrontPage(r *http.Request, args *AdminFrontPageArgs, reply *AdminFrontPageReply) error {

	authHeader := r.Header.Get("Authorization")

	// relay_frontend/status
	frontEndURI := rfs.RelayFrontendURI + "/status"
	client := &http.Client{}
	req, _ := http.NewRequest("GET", frontEndURI, nil)
	req.Header.Set("Authorization", authHeader)

	response, err := client.Do(req)
	if err != nil {
		err = fmt.Errorf("RelayFleet() error getting relay_frontend/status: %w", err)
		rfs.Logger.Log("err", err)
		return err
	}
	defer response.Body.Close()

	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
		err := fmt.Errorf("error parsing relay_frontend/status: %v", err)
		rfs.Logger.Log("err", err)
		return err
	}

	frontEndText := strings.Split(string(b), "\n")
	reply.RelayFrontEndStatus = append(reply.RelayFrontEndStatus, frontEndText...)

	// relay_frontend/master_status
	backEndMasterURI := rfs.RelayFrontendURI + "/master_status"
	client = &http.Client{}
	req, _ = http.NewRequest("GET", backEndMasterURI, nil)
	req.Header.Set("Authorization", authHeader)

	response, err = client.Do(req)
	if err != nil {
		err = fmt.Errorf("RelayFleet() error getting relay_frontend/master_status: %w", err)
		rfs.Logger.Log("err", err)
		return err
	}
	defer response.Body.Close()

	b, err = ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
		err := fmt.Errorf("error parsing relay_frontend/master_status: %v", err)
		rfs.Logger.Log("err", err)
		return err
	}

	backEndText := strings.Split(string(b), "\n")
	reply.RelayBackEndStatus = append(reply.RelayBackEndStatus, backEndText...)

	// relay_gateway/status
	gatewayURI := rfs.RelayGatewayURI + "/status"
	client = &http.Client{}
	req, _ = http.NewRequest("GET", gatewayURI, nil)
	req.Header.Set("Authorization", authHeader)

	response, err = client.Do(req)
	if err != nil {
		err = fmt.Errorf("RelayFleet() error getting relay_gateway/status: %w", err)
		rfs.Logger.Log("err", err)
		return err
	}
	defer response.Body.Close()

	b, err = ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
		err := fmt.Errorf("error parsing relay_gateway/status: %v", err)
		rfs.Logger.Log("err", err)
		return err
	}

	gatewayText := strings.Split(string(b), "\n")
	reply.RelayGatewayStatus = append(reply.RelayGatewayStatus, gatewayText...)

	// relay_forwarder/status
	fmt.Printf("'%s'\n", rfs.RelayForwarderURI)
	if rfs.RelayForwarderURI != "" {
		gatewayURI := rfs.RelayForwarderURI + "/status"
		client = &http.Client{}
		req, _ = http.NewRequest("GET", gatewayURI, nil)
		req.Header.Set("Authorization", authHeader)

		response, err = client.Do(req)
		if err != nil {
			err = fmt.Errorf("RelayFleet() error getting relay_forwarder/status: %w", err)
			rfs.Logger.Log("err", err)
			return err
		}
		defer response.Body.Close()

		b, err = ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Println(err)
			err := fmt.Errorf("error parsing relay_forwarder/status: %v", err)
			rfs.Logger.Log("err", err)
			return err
		}

		forwaderText := strings.Split(string(b), "\n")
		reply.RelayForwarderStatus = append(reply.RelayForwarderStatus, forwaderText...)
	} else {
		reply.RelayForwarderStatus = []string{"relay_forwarder dne in dev/local"}
	}

	reply.BinFileCreator = "Arthur Dent"
	reply.BinFileCreationTime = time.Now()

	return nil
}
