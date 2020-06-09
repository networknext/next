package jsonrpc_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/networknext/backend/transport/jsonrpc"
	"github.com/stretchr/testify/assert"
)

func mockGetInvoices(s *jsonrpc.InvoiceService, cacheName string) ([]byte, error) {

	existingData := []byte("Fred Scuttle")

	return existingData, nil
}

func TestInvoiceService(t *testing.T) {

	args := jsonrpc.InvoiceArgs{
		StartDate: time.Now(),
		EndDate:   time.Now(),
		BuyerID:   "some string",
	}

	svc := &jsonrpc.InvoiceService{}
	svc.Storage = nil
	svc.BqClient = nil

	t.Run("basic call", func(t *testing.T) {
		var reply jsonrpc.InvoiceReply

		svc.Invoices = mockGetInvoices

		err := svc.InvoiceAllBuyers(nil, &args, &reply)
		assert.NoError(t, err)

		assert.Equal(t, "Fred Scuttle", string(reply.Invoices))

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
