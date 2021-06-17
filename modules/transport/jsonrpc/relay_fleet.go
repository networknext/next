package jsonrpc

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-kit/kit/log"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/storage"
	"github.com/networknext/backend/modules/transport/middleware"
)

// RelayFleetService provides access to real-time data provided by the endpoints
// mounted on the relay_frontend (/relays, /cost_matrix (tbd), etc.).
type RelayFleetService struct {
	RelayFrontendURI  string
	RelayGatewayURI   string
	RelayForwarderURI string
	Logger            log.Logger
	Storage           storage.Storer
	Env               string
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

// type RelayDashboardAnalysisJsonReply struct {
// 	Analysis jsonAnalysisResponse `json:"fleetAnalysis"`
// }

// type RelayDashboardAnalysisJsonArgs struct{}

type jsonAnalysisResponse struct {
	Analysis routing.JsonMatrixAnalysis
}

// func (rfs *RelayFleetService) RelayDashboardAnalysisJson(r *http.Request, args *RelayDashboardAnalysisJsonArgs, reply *RelayDashboardAnalysisJsonReply) error {

// 	var analysis jsonAnalysisResponse
// 	authHeader := r.Header.Get("Authorization")

// 	uri := rfs.RelayFrontendURI + "/relay_dashboard_analysis"

// 	client := &http.Client{}
// 	req, _ := http.NewRequest("GET", uri, nil)
// 	req.Header.Set("Authorization", authHeader)

// 	response, err := client.Do(req)
// 	if err != nil {
// 		err = fmt.Errorf("RelayDashboardAnalysisJson() error getting fleet relay json: %w", err)
// 		rfs.Logger.Log("err", err)
// 		return err
// 	}
// 	defer response.Body.Close()

// 	byteValue, err := ioutil.ReadAll(response.Body)
// 	if err != nil {
// 		err = fmt.Errorf("RelayDashboardJson() error getting reading HTTP response body: %w", err)
// 		rfs.Logger.Log("err", err)
// 		return err
// 	}

// 	json.Unmarshal(byteValue, &analysis)
// 	reply.Analysis = analysis

// 	return nil
// }

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
		err = fmt.Errorf("RelayDashboardJson() error getting fleet relay json: %w", err)
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
		err := fmt.Errorf("RelayDashboardJson() relay backend returned an empty dashboard file")
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

// GetServiceURI provides a lookup table for service status URIs,
// each of which are defined by environment variables at run time.
//
// Note that as new status endpoints are added they must be attached at the
// service mount point in portal.go.
func (rfs *RelayFleetService) GetServiceURI(serviceName string) (string, error) {

	var serviceURI string
	var err error
	switch serviceName {
	case "RelayGateway":
		serviceURI = rfs.RelayGatewayURI + "/status"
	case "RelayFrontEnd":
		serviceURI = rfs.RelayFrontendURI + "/status"
	case "RelayBackEnd":
		serviceURI = rfs.RelayFrontendURI + "/master_status"
	case "RelayForwarder":
		if rfs.RelayForwarderURI != "" {
			serviceURI = rfs.RelayForwarderURI + "/status"
		}
	case "RelayPusher":
	case "ServerBackend":
	case "Billing":
	case "Analytics":
	case "Api":
	case "PortalCruncher":
	case "Portal":
	case "Vanity":
	default:
		err = fmt.Errorf("service %s does not exist", serviceName)
	}

	return serviceURI, err
}

type AdminFrontPageArgs struct {
	ServiceName string `json:"serviceName"`
}

var ServiceStatusList = []string{
	"RelayGateway",
	"RelayFrontEnd",
	"RelayBackEnd",
	"RelayForwarder",
	"RelayPusher",
	"ServerBackend",
	"Billing",
	"Analytics",
	"Api",
	"PortalCruncher",
	"Portal",
	"Vanity",
	"RelayDashboardAnalysis",
}

type AdminFrontPageReply struct {
	BinFileCreationTime time.Time `json:"binFileCreationTime"`
	BinFileAuthor       string    `json:"binFileAuthor"`
	ServiceStatusText   []string  `json:"serviceStatusText"`
	ServiceNameList     []string  `json:"serviceNameList"`
	SelectedService     string    `json:"selectedService"`
}

// AdminFrontPage returns the current database.bin file metadata status
// as well as the status text provided by the provided service. It returns
// the Analysis section of the RouteMatrix by default as well as the
// list of service names.
func (rfs *RelayFleetService) AdminFrontPage(r *http.Request, args *AdminFrontPageArgs, reply *AdminFrontPageReply) error {

	authHeader := r.Header.Get("Authorization")
	if args.ServiceName == "" || args.ServiceName == "RelayDashboardAnalysis" {
		var analysis routing.JsonMatrixAnalysis

		uri := rfs.RelayFrontendURI + "/relay_dashboard_analysis"

		client := &http.Client{}
		req, _ := http.NewRequest("GET", uri, nil)
		req.Header.Set("Authorization", authHeader)

		response, err := client.Do(req)
		if err != nil {
			err = fmt.Errorf("AdminFrontPage() error getting fleet relay dashboard analysis: %w", err)
			rfs.Logger.Log("err", err)
			return err
		}
		defer response.Body.Close()

		byteValue, err := ioutil.ReadAll(response.Body)
		if err != nil {
			err = fmt.Errorf("AdminFrontPage() error reading /relay_dashboard_analysis HTTP response body: %w", err)
			rfs.Logger.Log("err", err)
			return err
		}

		json.Unmarshal(byteValue, &analysis)

		reply.ServiceStatusText = strings.Split(analysis.String(), "\n")
		reply.SelectedService = "RelayDashboardAnalysis"

	} else {
		serviceURI, err := rfs.GetServiceURI(args.ServiceName)
		if err != nil {
			err = fmt.Errorf("AdminFrontPage() error getting service status URI: %w", err)
			rfs.Logger.Log("err", err)
			return err
		} else if serviceURI == "" {
			reply.ServiceStatusText = []string{fmt.Sprintf("%s has no status endpoint defined", args.ServiceName)}
			reply.SelectedService = args.ServiceName
		} else {
			client := &http.Client{}
			req, _ := http.NewRequest("GET", serviceURI, nil)
			req.Header.Set("Authorization", authHeader)

			response, err := client.Do(req)
			if err != nil {
				err = fmt.Errorf("AdminFrontPage() error getting relay_frontend/status: %w", err)
				rfs.Logger.Log("err", err)
				return err
			}
			defer response.Body.Close()

			b, err := ioutil.ReadAll(response.Body)
			if err != nil {
				fmt.Println(err)
				err := fmt.Errorf("AdminFrontPage() error parsing relay_frontend/status: %v", err)
				rfs.Logger.Log("err", err)
				return err
			}

			serviceStatusText := strings.Split(string(b), "\n")
			// remove the first line, it contains the service name
			reply.ServiceStatusText = append(reply.ServiceStatusText, serviceStatusText[1:len(serviceStatusText)-1]...)
			reply.SelectedService = args.ServiceName

		}

	}

	binFileMetaData, err := rfs.Storage.GetDatabaseBinFileMetaData()
	if err != nil {
		reply.BinFileAuthor = "Arthur Dent"
		reply.BinFileCreationTime = time.Now()
	} else {
		reply.BinFileAuthor = binFileMetaData.DatabaseBinFileAuthor
		reply.BinFileCreationTime = binFileMetaData.DatabaseBinFileCreationTime.UTC()
	}

	reply.ServiceNameList = ServiceStatusList
	return nil
}

type AdminBinFileHandlerArgs struct{}

type AdminBinFileHandlerReply struct {
	Message string `json:"message"`
}

// AdminBinFileHandler generates and commits a database.bin file
// for the Admin UI tool. The Admin UI (js) can not commit or
// otherwise work with the bin file.
func (rfs *RelayFleetService) AdminBinFileHandler(
	r *http.Request,
	args *AdminBinFileHandlerArgs,
	reply *AdminBinFileHandlerReply,
) error {

	requestUser := r.Context().Value(middleware.Keys.UserKey)
	if requestUser == nil {
		errCode := JSONRPCErrorCodes[int(ERROR_INSUFFICIENT_PRIVILEGES)]
		err := fmt.Errorf("AdminBinFileHandler() error getting userid: %v", errCode)
		rfs.Logger.Log("err", err)
		return err
	}

	requestEmail, ok := requestUser.(*jwt.Token).Claims.(jwt.MapClaims)["name"].(string)
	if !ok {
		err := JSONRPCErrorCodes[int(ERROR_JWT_PARSE_FAILURE)]
		rfs.Logger.Log("err", fmt.Errorf("AdminBinFileHandler(): %v: Failed to parse user ID", err.Error()))
		return &err
	}

	var buffer bytes.Buffer

	dbWrapper, err := rfs.BinFileGenerator(requestEmail)
	if err != nil {
		err := fmt.Errorf("AdminBinFileHandler() error generating database.bin file: %v", err)
		rfs.Logger.Log("err", err)
		reply.Message = err.Error()
		return err
	}

	encoder := gob.NewEncoder(&buffer)
	encoder.Encode(dbWrapper)

	tempFile, err := ioutil.TempFile("", "database.bin")
	if err != nil {
		err := fmt.Errorf("AdminBinFileHandler() error writing database.bin to temporary file: %v", err)
		rfs.Logger.Log("err", err)
		reply.Message = err.Error()
		return err
	}
	defer os.Remove(tempFile.Name())

	_, err = tempFile.Write(buffer.Bytes())
	if err != nil {
		err := fmt.Errorf("AdminBinFileHandler() error writing database.bin to filesystem: %v", err)
		rfs.Logger.Log("err", err)
		reply.Message = err.Error()
		return err
	}

	// should come from env var?
	bucketName := "gs://"
	switch rfs.Env {
	case "dev":
		bucketName += "dev_database_bin"
	case "prod":
		bucketName += "prod_database_bin"
	case "staging":
		bucketName += "staging_database_bin"
	case "local":
		bucketName += "happy_path_testing"
	}

	// enforce target file name, copy in /tmp has random numbers appended
	bucketName += "/database.bin"

	// gsutil cp /tmp/database.bin84756774 gs://${bucketName}
	gsutilCpCommand := exec.Command("gsutil", "cp", tempFile.Name(), bucketName)

	err = gsutilCpCommand.Run()
	if err != nil {
		err := fmt.Errorf("AdminBinFileHandler() error copying database.bin to %s: %v", bucketName, err)
		rfs.Logger.Log("err", err)
		reply.Message = err.Error()
	} else {
		reply.Message = "success!"
	}

	metaData := routing.DatabaseBinFileMetaData{
		DatabaseBinFileAuthor:       requestEmail,
		DatabaseBinFileCreationTime: time.Now(),
	}

	err = rfs.Storage.UpdateDatabaseBinFileMetaData(context.Background(), metaData)
	if err != nil {
		err := fmt.Errorf("AdminBinFileHandler() error writing bin file metadata to db: %v", err)
		rfs.Logger.Log("err", err)
		reply.Message = err.Error()
	}

	return nil
}

type NextBinFileHandlerArgs struct{}

type NextBinFileHandlerReply struct {
	DBWrapper routing.DatabaseBinWrapper `json:"dbWrapper"`
}

// NextBinFileHandler generates and returns a DatabaseBinWrapper struct
// for the next CLI tool. The next tool handles the commit.
func (rfs *RelayFleetService) NextBinFileHandler(
	r *http.Request,
	args *NextBinFileHandlerArgs,
	reply *NextBinFileHandlerReply,
) error {

	dbWrapper, err := rfs.BinFileGenerator("next")
	if err != nil {
		err := fmt.Errorf("BinFileHandler() error generating database.bin file: %v", err)
		rfs.Logger.Log("err", err)
		return err
	}

	reply.DBWrapper = dbWrapper
	return nil
}

func (rfs *RelayFleetService) BinFileGenerator(userEmail string) (routing.DatabaseBinWrapper, error) {

	var dbWrapper routing.DatabaseBinWrapper
	var enabledRelays []routing.Relay
	relayMap := make(map[uint64]routing.Relay)
	buyerMap := make(map[uint64]routing.Buyer)
	sellerMap := make(map[string]routing.Seller)
	datacenterMap := make(map[uint64]routing.Datacenter)
	datacenterMaps := make(map[uint64]map[uint64]routing.DatacenterMap)

	buyers := rfs.Storage.Buyers()
	for _, buyer := range buyers {
		buyerMap[buyer.ID] = buyer
		dcMapsForBuyer := rfs.Storage.GetDatacenterMapsForBuyer(buyer.ID)
		datacenterMaps[buyer.ID] = dcMapsForBuyer
	}

	for _, seller := range rfs.Storage.Sellers() {
		sellerMap[seller.ShortName] = seller
	}

	for _, datacenter := range rfs.Storage.Datacenters() {
		datacenterMap[datacenter.ID] = datacenter
	}

	for _, localRelay := range rfs.Storage.Relays() {
		if localRelay.State == routing.RelayStateEnabled {
			enabledRelays = append(enabledRelays, localRelay)
			relayMap[localRelay.ID] = localRelay
		}
	}

	dbWrapper.Relays = enabledRelays
	dbWrapper.RelayMap = relayMap
	dbWrapper.BuyerMap = buyerMap
	dbWrapper.SellerMap = sellerMap
	dbWrapper.DatacenterMap = datacenterMap
	dbWrapper.DatacenterMaps = datacenterMaps

	loc, err := time.LoadLocation("UTC")
	if err != nil {
		return routing.DatabaseBinWrapper{}, err
	}
	now := time.Now().In(loc)

	timeStamp := fmt.Sprintf("%s %d, %d %02d:%02d UTC\n", now.Month(), now.Day(), now.Year(), now.Hour(), now.Minute())
	dbWrapper.CreationTime = timeStamp
	dbWrapper.Creator = userEmail

	return dbWrapper, nil
}
