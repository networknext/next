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

// InvoiceService brings in the Google storage service and BigQuery client.
// Both Storage and BqClient *must* be initialized by the caller before use (requires context)
type InvoiceService struct {
	Storage  *storage.Client
	BqClient *bigquery.Client
	Invoices InvoiceGetter
}

// InvoiceArgs maintains the start and end dates for the invoice query
type InvoiceArgs struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

// InvoiceReply contains the JSON reply string
type InvoiceReply struct {
	// Invoices []InvoiceRecord `json:"invoice_list"`
	Invoices string `json:"invoice_list"`
}

// InvoiceGetter is a utility function that allows the actual GCS
// call to be mocked out in unit test.
type InvoiceGetter func(s *InvoiceService, cachename string) ([]byte, error)

// Getter provides an implementation of InvoiceGetter
type Getter struct {
	getInvoices InvoiceGetter
}

// NewGetter mounts the download function attached to the service
func NewGetter(ig InvoiceGetter) *Getter {
	return &Getter{getInvoices: ig}
}

func (d *Getter) download(s *InvoiceService, cacheName string) ([]byte, error) {
	return d.getInvoices(s, cacheName)
}

// GetInvoices is the real Google call. It is replaced by a mocked version in test
func GetInvoices(s *InvoiceService, cacheName string) ([]byte, error) {
	ctx := context.Background()
	rc, err := s.Storage.Bucket("network-next-invoice-cache").Object(cacheName).NewReader(ctx)
	if err != nil {
		return nil, err
	}

	existingData, err := ioutil.ReadAll(rc)
	if err != nil {
		return nil, err
	}

	return existingData, nil
}

// InvoiceAllBuyers issues a BigQuery to generate invoices for all buyers within
// a provided data range and return them in a single JSON reply.
func (s *InvoiceService) InvoiceAllBuyers(r *http.Request, args *InvoiceArgs, reply *InvoiceReply) error {

	ctx := context.Background()

	// needs to check for future date?
	startDate := args.StartDate.Format("2006-01-02")
	endDate := args.EndDate.Format("2006-01-02")

	// fmt.Printf("Checking for cached result...\n")
	cacheName := "cache-" + startDate + "-to-" + endDate + ".json"
	fmt.Printf("cacheName: %s\n", cacheName)
	d := NewGetter(s.Invoices)
	rc, err := d.download(s, cacheName)

	if err == nil {
		// cache file found, skip BQ query
		reply.Invoices = string(rc)
		return nil
	}

	if err == storage.ErrBucketNotExist {
		// cache file not found so query db
		fmt.Printf("Reading query from 'query.sql'...\n")
		queryText, err := ioutil.ReadFile("query.sql")
		if err != nil {
			return err
		}

		// fmt.Printf("Preparing query...\n")
		q := s.BqClient.Query(string(queryText))
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
		it, err := job.Read(ctx)
		if err != nil {
			return err
		}

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
		if err != nil {
			return err
		}
		reply.Invoices = string(data)

		return nil

	}

	// fall through
	return err
}
