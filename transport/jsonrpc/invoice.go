package jsonrpc

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/storage"

	"google.golang.org/api/iterator"
)

// InvoiceService brings in the Google storage service. InvoiceService.Storage
// *must* be initialized by the caller before use (requires context)
type InvoiceService struct {
	Storage *storage.Client
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

// InvoiceAllBuyers issues a BigQuery to generate invoices for all buyers within
// a provided data range and return them in a single JSON reply.
func (s InvoiceService) InvoiceAllBuyers(r *http.Request, args *InvoiceArgs, reply *InvoiceReply) error {

	// needs to check for future date?
	startDate := args.StartDate.Format("2006-01-02")
	endDate := args.EndDate.Format("2006-01-02")

	fmt.Printf("Checking for cached result...\n")
	cacheName := "cache-" + startDate + "-to-" + endDate + ".json"
	rc, err := s.Storage.Bucket("network-next-invoice-cache").Object(cacheName).NewReader(context.Background())

	if err == nil {
		// Cache file found, skip BQ query
		existingData, err := ioutil.ReadAll(rc)
		if err != nil {
			return err
		}

		var existingRows [][]bigquery.Value
		err = json.Unmarshal(existingData, &existingRows)
		if err != nil {
			return err
		}
		fmt.Printf("Found cached result, using that instead.\n")
		data, err := json.Marshal(existingRows)
		reply.Invoices = string(data)
		return nil

	}

	if err == storage.ErrBucketNotExist {
		// cache file not found so query db
		fmt.Printf("Reading query from 'query.sql'...\n")
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
		defer client.Close()

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
		data, err := json.Marshal(rows)
		reply.Invoices = string(data)

		return nil

	}

	// fall through
	return err
}
