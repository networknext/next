package pingdom

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/metrics"

	"cloud.google.com/go/bigquery"
	gopingdom "github.com/russellcardullo/go-pingdom/pingdom"
	"google.golang.org/api/iterator"
)

const (
	SecondsInHour = 3600
)

type PingdomUptime struct {
	Timestamp                     time.Time
	PortalUptimePercentage        float64
	ServerBackendUptimePercentage float64
}

func (entry *PingdomUptime) Save() (map[string]bigquery.Value, string, error) {

	e := make(map[string]bigquery.Value)

	e["timestamp"] = int(entry.Timestamp.Unix())
	e["portal_uptime_percentage"] = entry.PortalUptimePercentage
	e["server_backend_uptime_percentage"] = entry.ServerBackendUptimePercentage

	return e, "", nil
}

type PingdomClient struct {
	client         *gopingdom.Client
	pingdomMetrics *metrics.PingdomMetrics
	uptimeChan     chan *PingdomUptime

	bqClient      *bigquery.Client
	tableInserter *bigquery.Inserter
	gcpProjectID  string
	datasetName   string
	tableName     string
}

func NewPingdomClient(apiToken string, pingdomMetrics *metrics.PingdomMetrics, bqClient *bigquery.Client, gcpProjectID string, datasetName string, tableName string, chanSize int) (*PingdomClient, error) {
	client, err := gopingdom.NewClientWithConfig(gopingdom.ClientConfig{
		APIToken: apiToken,
	})

	if err != nil {
		return &PingdomClient{}, fmt.Errorf("NewPingdomClient(): failed to create pingdom client: %v", err)
	}

	if chanSize <= 0 {
		return &PingdomClient{}, fmt.Errorf("NewPingdomClient(): cannot have channel size <= 0")
	}

	uptimeChan := make(chan *PingdomUptime, chanSize)
	tableInserter := bqClient.Dataset(datasetName).Table(tableName).Inserter()

	return &PingdomClient{
		client:         client,
		pingdomMetrics: pingdomMetrics,
		uptimeChan:     uptimeChan,
		bqClient:       bqClient,
		tableInserter:  tableInserter,
		gcpProjectID:   gcpProjectID,
		datasetName:    datasetName,
		tableName:      tableName,
	}, nil
}

// Gets the pingdom check ID for the provided hostname
func (pc *PingdomClient) GetIDForHostname(hostname string) (int, error) {
	checks, err := pc.client.Checks.List()
	if err != nil {
		pc.pingdomMetrics.ErrorMetrics.PingdomAPICallFailure.Add(1)
		return -1, fmt.Errorf("GetIDForHostname(): failed to get pingdom checks: %v", err)
	}

	for _, check := range checks {
		if check.Hostname == hostname {
			return check.ID, nil
		}
	}

	return -1, fmt.Errorf("GetIDForHostname(): could not find ID for hostname %s", hostname)
}

// Read from BigQuery to get the latest timestamp
// Useful since we want to pick up from where we left off last time the service wrote data
func (pc *PingdomClient) GetLatestTimestamp(ctx context.Context) (int64, error) {
	latestTimestampQuery := fmt.Sprintf(`SELECT timestamp FROM %s.%s.%s ORDER BY timestamp DESC LIMIT 1`, pc.gcpProjectID, pc.datasetName, pc.tableName)

	q := pc.bqClient.Query(latestTimestampQuery)
	it, err := q.Read(ctx)
	if err != nil {
		pc.pingdomMetrics.ErrorMetrics.BigQueryReadFailure.Add(1)
		return 0, fmt.Errorf("GetLatestTimestamp(): failed to read from BigQuery using query %s: %v", latestTimestampQuery, err)
	}

	var rows []PingdomUptime
	// Iterate through rows response
	for {
		var rec PingdomUptime
		err := it.Next(&rec)

		if err == iterator.Done {
			break
		}
		if err != nil {
			return 0, fmt.Errorf("GetLatestTimestamp(): BigQuery iterator error: %v", err)
		}
		rows = append(rows, rec)
	}

	// No rows in BigQuery
	if len(rows) == 0 {
		return 0, nil
	}

	latestTimestamp := rows[0].Timestamp.Unix()
	return latestTimestamp, nil
}

// Creates a PingdomUptime struct to be written to BigQuery
func (pc *PingdomClient) GetUptimeForIDs(ctx context.Context, portalID int, serverBackendID int, pingFrequency time.Duration, wg *sync.WaitGroup, errChan chan error) {
	defer wg.Done()

	// Get the last timestmap written to BigQuery
	latestTimestamp, err := pc.GetLatestTimestamp(ctx)
	if err != nil {
		errChan <- err
		return
	}

	core.Debug("latest timestamp written to BigQuery: %d", latestTimestamp)

	// Occassionally ping the API for summary uptime performance
	ticker := time.NewTicker(pingFrequency)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:

			portalSummary, err := pc.getSummary(portalID, latestTimestamp)
			if err != nil {
				errChan <- err
				return
			}

			var latestPortalTimestamp int64
			var portalUptimePercentage float64
			var completeHour bool

			// Get the latest portal timestamp
			for _, day := range portalSummary {
				if int64(day.StartTime) <= latestTimestamp {
					continue
				}

				if day.Uptime+day.Downtime+day.Unmonitored == SecondsInHour {
					latestPortalTimestamp = int64(day.StartTime)
					portalUptimePercentage = float64(day.Uptime+day.Unmonitored-day.Downtime) / SecondsInHour

					completeHour = true
					break
				}
			}

			// Early out if we have already recorded this hour
			if !completeHour {
				continue
			}

			serverBackendSummary, err := pc.getSummary(serverBackendID, latestTimestamp)
			if err != nil {
				errChan <- err
				return
			}

			var latestServerBackendTimestamp int64
			var serverBackendUptimePercentage float64
			completeHour = false

			// Get the latest server backend timestamp
			for _, day := range serverBackendSummary {
				if int64(day.StartTime) <= latestTimestamp {
					continue
				}

				if day.Uptime+day.Downtime+day.Unmonitored == SecondsInHour {
					latestServerBackendTimestamp = int64(day.StartTime)
					serverBackendUptimePercentage = float64(day.Uptime+day.Unmonitored-day.Downtime) / SecondsInHour

					completeHour = true
					break
				}
			}

			// Early out if we have already recorded this hour
			if !completeHour {
				continue
			}

			// Portal and Server Backend timestamps should be the same
			if latestPortalTimestamp != latestServerBackendTimestamp {
				continue
			}

			// Update the latest timestamp
			latestTimestamp = latestPortalTimestamp

			// Create the uptime struct and insert it into the channel
			uptime := &PingdomUptime{
				Timestamp:                     time.Unix(latestTimestamp, 0),
				PortalUptimePercentage:        portalUptimePercentage,
				ServerBackendUptimePercentage: serverBackendUptimePercentage,
			}

			pc.uptimeChan <- uptime
			pc.pingdomMetrics.CreatePingdomUptime.Add(1)
		}
	}
}

// Inserts data to BigQuery
func (pc *PingdomClient) WriteLoop(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	var uptimeBuffer []*PingdomUptime

	for {
		select {
		case <-ctx.Done():
			return
		case uptime := <-pc.uptimeChan:
			core.Debug("pingdom uptime struct received: %+v", uptime)

			uptimeBuffer = append(uptimeBuffer, uptime)

			if err := pc.tableInserter.Put(context.Background(), uptimeBuffer); err != nil {
				pc.pingdomMetrics.ErrorMetrics.BigQueryWriteFailure.Add(float64(len(uptimeBuffer)))
				core.Error("failed to write pingdom uptime to BigQuery: %v", err)
				continue
			}

			pc.pingdomMetrics.BigQueryWriteSuccess.Add(float64(len(uptimeBuffer)))
			uptimeBuffer = uptimeBuffer[:0]
		}
	}
}

// Makes the API call to Pingdom for summary performance
// Note that we do not use SummaryPerformanceRequest because the gopingdom library
// disregards the from and to times in the request
func (pc *PingdomClient) getSummary(id int, latestTimestamp int64) ([]gopingdom.SummaryPerformanceSummary, error) {
	var fromTime int
	if latestTimestamp == 0 {
		// The from time should be the first hour of the month in UTC
		// This way we can backfill BigQuery in case it is empty
		year, month, _ := time.Now().UTC().Date()
		loc := time.FixedZone("UTC", 0)

		fromTime = int(time.Date(year, month, 1, 0, 0, 0, 0, loc).UTC().Unix())
	} else {
		// The from time is 1 hour before the latest timestamp
		mostRecent := time.Unix(latestTimestamp, 0)
		fromTime = int(mostRecent.Add(-1 * time.Hour).UTC().Unix())
	}

	toTime := int(time.Now().Unix())

	perfRequest := make(map[string]string)
	perfRequest["from"] = strconv.Itoa(fromTime)
	perfRequest["includeuptime"] = "true"
	perfRequest["order"] = "asc"
	perfRequest["resolution"] = "hour"
	perfRequest["to"] = strconv.Itoa(toTime)

	req, err := pc.client.NewRequest("GET", "/summary.performance/"+strconv.Itoa(id), perfRequest)
	if err != nil {
		pc.pingdomMetrics.ErrorMetrics.BadSummaryPerformanceRequest.Add(1)
		err = fmt.Errorf("getSummary(): invalid summary performance request: %v", err)
		return []gopingdom.SummaryPerformanceSummary{}, err
	}

	summary := &gopingdom.SummaryPerformanceResponse{}

	_, err = pc.client.Do(req, summary)
	if err != nil {
		pc.pingdomMetrics.ErrorMetrics.PingdomAPICallFailure.Add(1)
		err = fmt.Errorf("getSummary(): failed to get summary performance map: %v", err)
		return []gopingdom.SummaryPerformanceSummary{}, err
	}

	return summary.Summary.Hours, nil
}
