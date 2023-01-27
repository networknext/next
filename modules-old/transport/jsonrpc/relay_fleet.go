package jsonrpc

import (
	"encoding/csv"
	"fmt"
	"net/http"

	"github.com/networknext/backend/modules/core"
)

const (
	DevDatabaseBinGCPBucketName     = "dev_database_bin"
	StagingDatabaseBinGCPBucketName = "staging_database_bin"
	ProdDatabaseBinGCPBucketName    = "prod_database_bin"
	LocalDatabaseBinGCPBucketName   = "happy_path_testing"
)

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

type RelayFleetService struct {
       AnalyticsMIG       string
       AnalyticsPusherURI string
       BillingMIG         string
       PingdomURI         string
       PortalBackendMIG   string
       PortalCruncherURI  string
       RelayForwarderURI  string
       RelayBackendURI    string
       RelayGatewayURI    string
       RelayPusherURI     string
       ServerBackendMIG   string
       Env                string
}

func (rfs *RelayFleetService) RelayFleet(r *http.Request, args *RelayFleetArgs, reply *RelayFleetReply) error {
	authHeader := r.Header.Get("Authorization")

	uri := "http://" + rfs.RelayBackendURI + "/relays"

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
