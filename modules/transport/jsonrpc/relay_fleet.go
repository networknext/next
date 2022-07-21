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
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/storage"
	"github.com/networknext/backend/modules/transport/middleware"
)

const (
	DevDatabaseBinGCPBucketName     = "dev_database_bin"
	StagingDatabaseBinGCPBucketName = "staging_database_bin"
	ProdDatabaseBinGCPBucketName    = "prod_database_bin"
	LocalDatabaseBinGCPBucketName   = "happy_path_testing"
)

// RelayFleetService provides access to real-time data provided by the endpoints
// mounted on the relay_frontend (/relays, /cost_matrix (tbd), etc.).
type RelayFleetService struct {
	AnalyticsMIG       string
	AnalyticsPusherURI string
	BillingMIG         string
	PingdomURI         string
	PortalBackendMIG   string
	PortalCruncherURI  string
	RelayForwarderURI  string
	RelayFrontendURI   string
	RelayGatewayURI    string
	RelayPusherURI     string
	ServerBackendMIG   string
	Storage            storage.Storer
	Env                string
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

	uri := "http://" + rfs.RelayFrontendURI + "/relays"

	client := &http.Client{}
	req, _ := http.NewRequest("GET", uri, nil)
	req.Header.Set("Authorization", authHeader)

	response, err := client.Do(req)
	if err != nil {
		err = fmt.Errorf("RelayFleet() error getting relays.csv: %w", err)
		core.Error("%v", err)
		return err
	}
	defer response.Body.Close()

	reader := csv.NewReader(response.Body)
	relayData, err := reader.ReadAll()
	if err != nil {
		err = fmt.Errorf("RelayFleet() could not parse relays csv file from %s: %v", uri, err)
		core.Error("%v", err)
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
		core.Error("%v", err)
		return err
	}

	var fullDashboard, filteredDashboard jsonResponse
	authHeader := r.Header.Get("Authorization")

	uri := "http://" + rfs.RelayFrontendURI + "/relay_dashboard_data"

	client := &http.Client{}
	req, _ := http.NewRequest("GET", uri, nil)
	req.Header.Set("Authorization", authHeader)

	response, err := client.Do(req)
	if err != nil {
		err = fmt.Errorf("RelayDashboardJson() error getting fleet relay json: %w", err)
		core.Error("%v", err)
		return err
	}
	defer response.Body.Close()

	byteValue, err := ioutil.ReadAll(response.Body)
	if err != nil {
		err = fmt.Errorf("RelayDashboardJson() error getting reading HTTP response body: %w", err)
		core.Error("%v", err)
		return err
	}

	json.Unmarshal(byteValue, &fullDashboard)
	if len(fullDashboard.Relays) == 0 {
		err := fmt.Errorf("RelayDashboardJson() relay backend returned an empty dashboard file")
		core.Error("%v", err)
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
		core.Error("%v", err)
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
		core.Error("%v", err)
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
	case "Analytics":
		healthyInstanceName, err := rfs.GetHealthyInstanceInMIG(rfs.AnalyticsMIG)
		if err != nil {
			return serviceURI, err
		}
		instanceInternalIP, err := rfs.GetIPAddressForInstanceName(healthyInstanceName)
		if err != nil {
			return serviceURI, err
		}
		serviceURI = fmt.Sprintf("http://%s/status", instanceInternalIP)
	case "AnalyticsPusher":
		serviceURI = fmt.Sprintf("http://%s/status", rfs.AnalyticsPusherURI)
	case "Billing":
		healthyInstanceName, err := rfs.GetHealthyInstanceInMIG(rfs.BillingMIG)
		if err != nil {
			return serviceURI, err
		}
		instanceInternalIP, err := rfs.GetIPAddressForInstanceName(healthyInstanceName)
		if err != nil {
			return serviceURI, err
		}
		serviceURI = fmt.Sprintf("http://%s/status", instanceInternalIP)
	case "Pingdom":
		serviceURI = fmt.Sprintf("http://%s/status", rfs.PingdomURI)
	case "PortalBackend":
		healthyInstanceName, err := rfs.GetHealthyInstanceInMIG(rfs.PortalBackendMIG)
		if err != nil {
			return serviceURI, err
		}
		instanceInternalIP, err := rfs.GetIPAddressForInstanceName(healthyInstanceName)
		if err != nil {
			return serviceURI, err
		}
		serviceURI = fmt.Sprintf("http://%s/status", instanceInternalIP)
	case "PortalCruncher":
		serviceURI = fmt.Sprintf("http://%s/status", rfs.PortalCruncherURI)
	case "RelayBackend":
		serviceURI = fmt.Sprintf("http://%s/master_status", rfs.RelayFrontendURI)
	case "RelayForwarder":
		if rfs.RelayForwarderURI != "" {
			serviceURI = fmt.Sprintf("http://%s/status", rfs.RelayForwarderURI)
		}
	case "RelayFrontend":
		serviceURI = fmt.Sprintf("http://%s/status", rfs.RelayFrontendURI)
	case "RelayGateway":
		serviceURI = fmt.Sprintf("http://%s/status", rfs.RelayGatewayURI)
	case "RelayPusher":
		serviceURI = fmt.Sprintf("http://%s/status", rfs.RelayPusherURI)
	case "ServerBackend":
		healthyInstanceName, err := rfs.GetHealthyInstanceInMIG(rfs.ServerBackendMIG)
		if err != nil {
			return serviceURI, err
		}
		instanceInternalIP, err := rfs.GetIPAddressForInstanceName(healthyInstanceName)
		if err != nil {
			return serviceURI, err
		}
		serviceURI = fmt.Sprintf("http://%s/status", instanceInternalIP)
	default:
		err = fmt.Errorf("service %s does not exist", serviceName)
	}

	return serviceURI, err
}

// Gets the first healthy instance in a MIG in GCP
func (rfs *RelayFleetService) GetHealthyInstanceInMIG(migName string) (string, error) {
	var gcpProjectID string
	switch rfs.Env {
	case "prod":
		gcpProjectID = "network-next-v3-prod"
	case "staging":
		gcpProjectID = "network-next-v3-staging"
	case "dev":
		gcpProjectID = "network-next-v3-dev"
	case "local":
		// For local env, mig name is the local IP address
		return migName, nil
	default:
		err := fmt.Errorf("GetHealthyInstanceInMIG(): env %s not supported", rfs.Env)
		return "", err
	}

	cmd := exec.Command("gcloud", "compute", "instance-groups", "managed", "list-instances", migName, "--format", "value(instance)", "--filter", "healthy", "--project", gcpProjectID, "--zone", "us-central1-a")
	buffer, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	migInstanceNames := strings.Split(string(buffer), "\n")

	// Using the method above causes an empty string to be added at the end of the slice - remove it
	if len(migInstanceNames) > 0 {
		migInstanceNames = migInstanceNames[:len(migInstanceNames)-1]
	}

	if len(migInstanceNames) == 0 {
		return "", fmt.Errorf("no healthy instances in %s MIG", migName)
	}

	return migInstanceNames[0], nil
}

// Gets the internal IP address for an instance in GCP
func (rfs *RelayFleetService) GetIPAddressForInstanceName(instanceName string) (string, error) {
	var gcpProjectID string
	switch rfs.Env {
	case "prod":
		gcpProjectID = "network-next-v3-prod"
	case "staging":
		gcpProjectID = "network-next-v3-staging"
	case "dev":
		gcpProjectID = "network-next-v3-dev"
	case "local":
		// For local env, instance name is the local IP address
		return instanceName, nil
	default:
		err := fmt.Errorf("GetIPAddressForInstanceName(): env %s not supported", rfs.Env)
		return "", err
	}

	cmd := exec.Command("gcloud", "compute", "instances", "describe", instanceName, "--format", "get(networkInterfaces[0].networkIP)", "--project", gcpProjectID, "--zone", "us-central1-a")
	buffer, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	// Remove the \n at the end of the buffer
	instanceInternalIP := strings.Split(string(buffer), "\n")[0]

	return instanceInternalIP, nil
}

type AdminFrontPageArgs struct {
	ServiceName string `json:"serviceName"`
}

var ServiceStatusList = []string{
	"Analytics",
	"AnalyticsPusher",
	"Billing",
	"Pingdom",
	"PortalBackend",
	"PortalCruncher",
	"RelayBackend",
	"RelayFrontend",
	"RelayGateway",
	"RelayPusher",
	"ServerBackend",
	"RelayDashboardAnalysis",
}

type AdminFrontPageReply struct {
	BinFileCreationTime time.Time `json:"binFileCreationTime"`
	BinFileAuthor       string    `json:"binFileAuthor"`
	ServiceStatusText   []string  `json:"serviceStatusText"`
	ServiceNameList     []string  `json:"serviceNameList"`
	SelectedService     string    `json:"selectedService"`
	MondayApiKey        string    `json:"mondayApiKey"`
}

// AdminFrontPage returns the current database.bin file metadata status
// as well as the status text provided by the provided service. It returns
// the Analysis section of the RouteMatrix by default as well as the
// list of service names.
func (rfs *RelayFleetService) AdminFrontPage(r *http.Request, args *AdminFrontPageArgs, reply *AdminFrontPageReply) error {

	authHeader := r.Header.Get("Authorization")
	if args.ServiceName == "" || args.ServiceName == "RelayDashboardAnalysis" {

		uri := "http://" + rfs.RelayFrontendURI + "/relay_dashboard_analysis"

		client := &http.Client{}
		req, err := http.NewRequest("GET", uri, nil)
		if err != nil {
			err = fmt.Errorf("AdminFrontPage() error setting up NewRequest(): %w", err)
			core.Error("%v", err)
			return err
		}
		req.Header.Set("Authorization", authHeader)

		response, err := client.Do(req)
		if err != nil {
			err = fmt.Errorf("AdminFrontPage() error getting fleet relay dashboard analysis: %w", err)
			core.Error("%v", err)
			return err
		}
		defer response.Body.Close()

		byteValue, err := ioutil.ReadAll(response.Body)
		if err != nil {
			err = fmt.Errorf("AdminFrontPage() error reading /relay_dashboard_analysis HTTP response body: %w", err)
			core.Error("%v", err)
			return err
		}

		type jsonIncoming struct {
			Analysis routing.JsonMatrixAnalysis `json:"analysis"`
		}

		var incoming jsonIncoming

		json.Unmarshal(byteValue, &incoming)

		reply.ServiceStatusText = strings.Split(incoming.Analysis.String(), "\n")
		reply.SelectedService = "RelayDashboardAnalysis"

	} else {
		serviceURI, err := rfs.GetServiceURI(args.ServiceName)
		if err != nil {
			err = fmt.Errorf("AdminFrontPage() error getting service status URI: %w", err)
			core.Error("%v", err)
			return err
		} else if serviceURI == "" {
			reply.ServiceStatusText = []string{fmt.Sprintf("%s has no status endpoint defined", args.ServiceName)}
			reply.SelectedService = args.ServiceName
		} else {
			client := &http.Client{}
			req, err := http.NewRequest("GET", serviceURI, nil)
			if err != nil {
				err = fmt.Errorf("AdminFrontPage() error getting status for service: %w", err)
				core.Error("%v", err)
				return err
			}
			req.Header.Set("Authorization", authHeader)

			response, err := client.Do(req)
			if err != nil {
				err = fmt.Errorf("AdminFrontPage() error getting status for service %s (%s): %v", args.ServiceName, serviceURI, err)
				core.Error("%v", err)
				return err
			}
			defer response.Body.Close()

			b, err := ioutil.ReadAll(response.Body)
			if err != nil {
				err := fmt.Errorf("AdminFrontPage() error parsing status for service %s (%s): %v", args.ServiceName, serviceURI, err)
				core.Error("%v", err)
				return err
			}

			var fields reflect.Type
			var values reflect.Value

			// Unmarshal the status into the corresponding service's status struct
			switch args.ServiceName {
			case "Analytics":
				var status metrics.AnalyticsStatus
				json.Unmarshal(b, &status)
				fields = reflect.TypeOf(status)
				values = reflect.ValueOf(status)

			case "AnalyticsPusher":
				var status metrics.AnalyticsPusherStatus
				json.Unmarshal(b, &status)
				fields = reflect.TypeOf(status)
				values = reflect.ValueOf(status)
			case "Billing":
				var status metrics.BillingStatus
				json.Unmarshal(b, &status)
				fields = reflect.TypeOf(status)
				values = reflect.ValueOf(status)
			case "Pingdom":
				var status metrics.PingdomStatus
				json.Unmarshal(b, &status)
				fields = reflect.TypeOf(status)
				values = reflect.ValueOf(status)
			case "PortalBackend":
				var status metrics.PortalStatus
				json.Unmarshal(b, &status)
				fields = reflect.TypeOf(status)
				values = reflect.ValueOf(status)
			case "PortalCruncher":
				var status metrics.PortalCruncherStatus
				json.Unmarshal(b, &status)
				fields = reflect.TypeOf(status)
				values = reflect.ValueOf(status)
			case "RelayBackend":
				var status metrics.RelayBackendStatus
				json.Unmarshal(b, &status)
				fields = reflect.TypeOf(status)
				values = reflect.ValueOf(status)
			case "RelayForwarder":
				var status metrics.RelayForwarderStatus
				json.Unmarshal(b, &status)
				fields = reflect.TypeOf(status)
				values = reflect.ValueOf(status)
			case "RelayFrontend":
				var status metrics.RelayFrontendStatus
				json.Unmarshal(b, &status)
				fields = reflect.TypeOf(status)
				values = reflect.ValueOf(status)
			case "RelayGateway":
				var status metrics.RelayGatewayStatus
				json.Unmarshal(b, &status)
				fields = reflect.TypeOf(status)
				values = reflect.ValueOf(status)
			case "RelayPusher":
				var status metrics.RelayPusherStatus
				json.Unmarshal(b, &status)
				fields = reflect.TypeOf(status)
				values = reflect.ValueOf(status)
			case "ServerBackend":
				var status metrics.ServerBackendStatus
				json.Unmarshal(b, &status)
				fields = reflect.TypeOf(status)
				values = reflect.ValueOf(status)
			default:
				err := fmt.Errorf("AdminFrontPage() service %s does not have status", args.ServiceName)
				core.Error("%v", err)
				return err
			}

			for i := 0; i < fields.NumField(); i++ {
				reply.ServiceStatusText = append(reply.ServiceStatusText, fmt.Sprintf("%s: %v", fields.Field(i).Name, values.Field(i)))
			}

			reply.SelectedService = args.ServiceName
		}

	}

	binFileMetaData, err := rfs.Storage.GetDatabaseBinFileMetaData(r.Context())
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
		core.Error("%v", err)
		return err
	}

	requestEmail, ok := requestUser.(*jwt.Token).Claims.(jwt.MapClaims)["name"].(string)
	if !ok {
		err := JSONRPCErrorCodes[int(ERROR_JWT_PARSE_FAILURE)]

		core.Error("AdminBinFileHandler(): %v: Failed to parse user ID", err.Error())
		return &err
	}

	var buffer bytes.Buffer

	dbWrapper, err := rfs.BinFileGenerator(r.Context(), requestEmail)
	if err != nil {
		err := fmt.Errorf("AdminBinFileHandler() error generating database.bin file: %v", err)
		core.Error("%v", err)
		reply.Message = err.Error()
		return err
	}

	encoder := gob.NewEncoder(&buffer)
	encoder.Encode(dbWrapper)

	tempFile, err := ioutil.TempFile("", "database.bin")
	if err != nil {
		err := fmt.Errorf("AdminBinFileHandler() error writing database.bin to temporary file: %v", err)
		core.Error("%v", err)
		reply.Message = err.Error()
		return err
	}
	defer os.Remove(tempFile.Name())

	_, err = tempFile.Write(buffer.Bytes())
	if err != nil {
		err := fmt.Errorf("AdminBinFileHandler() error writing database.bin to filesystem: %v", err)
		core.Error("%v", err)
		reply.Message = err.Error()
		return err
	}

	bucketName := "gs://"
	switch rfs.Env {
	case "dev4", "dev5":
		bucketName += DevDatabaseBinGCPBucketName
	case "staging":
		bucketName += StagingDatabaseBinGCPBucketName
	case "prod":
		bucketName += ProdDatabaseBinGCPBucketName
	case "local":
		bucketName += LocalDatabaseBinGCPBucketName
	}

	// enforce target file name, copy in /tmp has random numbers appended
	bucketName += "/database.bin"

	// gsutil cp /tmp/database.bin84756774 gs://${bucketName}
	gsutilCpCommand := exec.Command("gsutil", "cp", tempFile.Name(), bucketName)

	err = gsutilCpCommand.Run()
	if err != nil {
		err := fmt.Errorf("AdminBinFileHandler() error copying database.bin to %s: %v", bucketName, err)
		core.Error("%v", err)
		reply.Message = err.Error()
	} else {
		reply.Message = "success!"
	}

	metaData := routing.DatabaseBinFileMetaData{
		DatabaseBinFileAuthor:       requestEmail,
		DatabaseBinFileCreationTime: time.Now(),
	}

	err = rfs.Storage.UpdateDatabaseBinFileMetaData(r.Context(), metaData)
	if err != nil {
		err := fmt.Errorf("AdminBinFileHandler() error writing bin file metadata to db: %v", err)
		core.Error("%v", err)
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

	dbWrapper, err := rfs.BinFileGenerator(r.Context(), "next")
	if err != nil {
		err := fmt.Errorf("BinFileHandler() error generating database.bin file: %v", err)
		core.Error("%v", err)
		return err
	}

	reply.DBWrapper = dbWrapper
	return nil
}

type NextBinFileCommitTimeStampArgs struct{}

type NextBinFileCommitTimeStampReply struct{}

func (rfs *RelayFleetService) NextBinFileCommitTimeStamp(
	r *http.Request,
	args *NextBinFileCommitTimeStampArgs,
	reply *NextBinFileCommitTimeStampReply,
) error {

	metaData := routing.DatabaseBinFileMetaData{
		DatabaseBinFileAuthor:       "next cli",
		DatabaseBinFileCreationTime: time.Now(),
	}

	err := rfs.Storage.UpdateDatabaseBinFileMetaData(r.Context(), metaData)
	if err != nil {
		err := fmt.Errorf("NextBinFileCommitTimeStamp() error writing bin file metadata to db: %v", err)
		core.Error("%v", err)
		return err
	}

	return nil

}

func (rfs *RelayFleetService) BinFileGenerator(ctx context.Context, userEmail string) (routing.DatabaseBinWrapper, error) {

	var dbWrapper routing.DatabaseBinWrapper
	var enabledRelays []routing.Relay
	relayMap := make(map[uint64]routing.Relay)
	buyerMap := make(map[uint64]routing.Buyer)
	sellerMap := make(map[string]routing.Seller)
	datacenterMap := make(map[uint64]routing.Datacenter)
	datacenterMaps := make(map[uint64]map[uint64]routing.DatacenterMap)

	buyers := rfs.Storage.Buyers(ctx)
	for _, buyer := range buyers {
		buyerMap[buyer.ID] = buyer
		dcMapsForBuyer := rfs.Storage.GetDatacenterMapsForBuyer(ctx, buyer.ID)
		datacenterMaps[buyer.ID] = dcMapsForBuyer
	}

	for _, seller := range rfs.Storage.Sellers(ctx) {
		sellerMap[seller.ShortName] = seller
	}

	for _, datacenter := range rfs.Storage.Datacenters(ctx) {
		datacenterMap[datacenter.ID] = datacenter
	}

	for _, localRelay := range rfs.Storage.Relays(ctx) {
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

type NextCostMatrixHandlerArgs struct{}

type NextCostMatrixHandlerReply struct {
	CostMatrix []byte `json:"costMatrix"`
}

// NextCostMatrixHandler gets the []byte cost matrix from
// relay_frontend/cost_matrix and returns it
func (rfs *RelayFleetService) NextCostMatrixHandler(
	r *http.Request,
	args *NextCostMatrixHandlerArgs,
	reply *NextCostMatrixHandlerReply,
) error {

	authHeader := r.Header.Get("Authorization")

	uri := "http://" + rfs.RelayFrontendURI + "/cost_matrix"

	client := &http.Client{}
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		err = fmt.Errorf("NextCostMatrixHandler() error creating new request: %w", err)
		core.Error("%v", err)
		return err
	}
	req.Header.Set("Authorization", authHeader)

	response, err := client.Do(req)
	if err != nil {
		err = fmt.Errorf("NextCostMatrixHandler() error getting cost matrix: %w", err)
		core.Error("%v", err)
		return err
	}
	defer response.Body.Close()

	byteValue, err := ioutil.ReadAll(response.Body)
	if err != nil {
		err = fmt.Errorf("NextCostMatrixHandler() error reading /cost_matrix HTTP response body: %w", err)
		core.Error("%v", err)
		return err
	}

	reply.CostMatrix = byteValue

	return nil
}
