package jsonrpc

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/networknext/backend/storage"

	"google.golang.org/api/iterator"
)

// InvoiceService brings in the storage types.
type InvoiceService struct {
	Auth0   storage.Auth0
	Storage storage.Storer
}

// InvoiceArgs maintains the start and end dates for the invoice query
type InvoiceArgs struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

// InvoiceReply contains the JSON reply string
type InvoiceReply struct {
	Invoices string `json:"invoice_list"`
}

// AllBuyers issues a BigQuery to generate invoices for all buyers within
// a provided data range and return them in a single JSON reply.
func (s *InvoiceService) AllBuyers(r *http.Request, args *InvoiceArgs, reply *InvoiceReply) error {

	startDate := args.StartDate.Format("2006-01-02")
	endDate := args.EndDate.Format("2006-01-02")

	reply.Invoices = ""

	storageSrv, err := storage.NewClient(context.Background())
	if err != nil {
		return err
	}

	// fmt.Printf("Checking for cached result...\n")
	cacheName := "cache-" + startDate + "-to-" + endDate + ".cache"
	rc, err := storageSrv.Bucket("network-next-bill-cache").Object(cacheName).NewReader(context.Background())
	if err == nil {
		defer rc.Close()

		existingData, err := ioutil.ReadAll(rc)
		if err != nil {
			return err
		}

		var existingRows [][]bigquery.Value
		err = json.Unmarshal(existingData, &existingRows)
		if err != nil {
			return err
		}
		// fmt.Printf("Found cached result, using that instead.\n")
		data, err := json.MarshalIndent(existingRows, "", " ")
		reply.Invoices = string(data)
	}

	// fmt.Printf("Reading query from 'query.sql'...\n")
	queryText, err := ioutil.ReadFile("query.sql")
	if err != nil {
		return err
	}

	// fmt.Printf("Connecting to BigQuery...\n")
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "network-next-v3-prod")
	if err != nil {
		return err
	}

	// fmt.Printf("Preparing query...\n")
	q := client.Query(string(queryText))
	q.Parameters = []bigquery.QueryParameter{
		{
			Name:  "start",
			Value: startDate,
		},
		{
			Name:  "end",
			Value: endDate,
		},
	}

	// fmt.Printf("Starting query...\n")
	job, err := q.Run(ctx)
	if err != nil {
		return err
	}

	// fmt.Printf("Waiting for query to complete...\n")
	status, err := job.Wait(ctx)
	if err != nil {
		return err
	}
	if err := status.Err(); err != nil {
		return err
	}

	// billed := status.Statistics.Details.(*bigquery.QueryStatistics).TotalBytesBilled
	// fmt.Printf("Query billed us %d GB.\n", billed/1024/1024/1024)

	// fmt.Printf("Reading result rows...\n")
	var rows [][]bigquery.Value
	it, _ := job.Read(ctx)
	for {
		var row []bigquery.Value
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		rows = append(rows, row)
	}

	// fmt.Printf("Writing result to output...\n")
	data, err := json.MarshalIndent(rows, "", " ")
	reply.Invoices = string(data)

	return nil
}
