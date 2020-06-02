package jsonrpc_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/storage"

	"github.com/networknext/backend/transport/jsonrpc"
	"github.com/stretchr/testify/assert"
)

func mockGetInvoices(s *jsonrpc.InvoiceService, cacheName string) ([]byte, error) {

	existingData := []byte("Fred Scuttle")

	return existingData, nil
}

func TestInvoiceService(t *testing.T) {

	var err error

	args := jsonrpc.InvoiceArgs{
		StartDate: time.Date(2020, 4, 1, 0, 0, 0, 0, time.Local),
		EndDate:   time.Date(2020, 4, 30, 0, 0, 0, 0, time.Local),
	}

	svc := &jsonrpc.InvoiceService{}

	ctx := context.Background()

	svc.Storage, err = storage.NewClient(ctx)
	assert.NoError(t, err)
	defer svc.Storage.Close()

	svc.BqClient, err = bigquery.NewClient(ctx, "network-next-v3-prod")
	assert.NoError(t, err)
	defer svc.BqClient.Close()

	t.Run("generate April, 2020 invoices", func(t *testing.T) {
		var reply jsonrpc.InvoiceReply

		// Use real GCS fetch
		// svc.Invoices = jsonrpc.GetInvoices
		// Mock fetch
		svc.Invoices = mockGetInvoices

		err := svc.InvoiceAllBuyers(nil, &args, &reply)
		assert.NoError(t, err)

		// asserting/checking the data returned TBD
		fmt.Printf("data:\n%s\n", reply.Invoices)

	})

	t.Run("test GetInvoices() directly", func(t *testing.T) {
		t.Skip() // makes a network call so skip by default

		svc.Invoices = mockGetInvoices

		data, err := jsonrpc.GetInvoices(svc, "cache-2020-04-01-to-2020-04-30.json")
		assert.NoError(t, err)

		// assert.Equal(t, dataCheck, data)
		fmt.Printf("data: %v", data)

	})
}
