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

func TestInvoiceAllBuyers(t *testing.T) {

	var err error

	args := jsonrpc.InvoiceArgs{
		StartDate: time.Date(2020, 4, 1, 0, 0, 0, 0, time.Local),
		EndDate:   time.Date(2020, 4, 30, 0, 0, 0, 0, time.Local),
	}

	svc := jsonrpc.InvoiceService{}

	t.Run("generate April, 2020 invoices", func(t *testing.T) {
		var reply jsonrpc.InvoiceReply

		ctx := context.Background()

		svc.Storage, err = storage.NewClient(ctx)
		assert.NoError(t, err)
		defer svc.Storage.Close()

		svc.BqClient, err = bigquery.NewClient(ctx, "network-next-v3-prod")
		assert.NoError(t, err)
		defer svc.BqClient.Close()

		err := svc.InvoiceAllBuyers(nil, &args, &reply)
		assert.NoError(t, err)

		// asserting/checking the data returned TBD
		fmt.Printf("data:\n%s\n", reply.Invoices)

	})
}

// func mustParse(d string) time.Time {
// 	const timeFmt string = "2006-01-02"
// 	a1, err := time.Parse(timeFmt, d)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return a1
// }

// func assertDate(t *testing.T, v time.Time, year int, month time.Month, day int) {
// 	y, m, d := v.Date()
// 	assert.Equal(t, y, year)
// 	assert.Equal(t, m, month)
// 	assert.Equal(t, d, day)
// }

// func startOfLastMonth(now time.Time) time.Time {
// 	y, m, _ := now.Date()
// 	return time.Date(y, m-1, 1, 0, 0, 0, 0, now.Location())
// }

// func endOfLastMonth(now time.Time) time.Time {
// 	y, m, _ := now.Date()
// 	return time.Date(y, m, 1, 0, 0, 0, 0, now.Location()).Add(-time.Nanosecond)
// }

// func TestStartOfLastMonth(t *testing.T) {
// 	assertDate(t, startOfLastMonth(mustParse("2018-01-30")), 2017, 12, 1)
// 	assertDate(t, startOfLastMonth(mustParse("2018-01-01")), 2017, 12, 1)
// 	assertDate(t, startOfLastMonth(mustParse("2018-01-15")), 2017, 12, 1)
// 	assertDate(t, startOfLastMonth(mustParse("2020-01-30")), 2019, 12, 1)
// 	assertDate(t, startOfLastMonth(mustParse("2020-01-01")), 2019, 12, 1)
// 	assertDate(t, startOfLastMonth(mustParse("2020-01-15")), 2019, 12, 1)
// 	assertDate(t, startOfLastMonth(mustParse("2020-02-29")), 2020, 1, 1)
// 	assertDate(t, startOfLastMonth(mustParse("2020-02-01")), 2020, 1, 1)
// 	assertDate(t, startOfLastMonth(mustParse("2020-02-15")), 2020, 1, 1)
// }

// func TestEndOfLastMonth(t *testing.T) {
// 	assertDate(t, endOfLastMonth(mustParse("2018-01-30")), 2017, 12, 31)
// 	assertDate(t, endOfLastMonth(mustParse("2018-01-01")), 2017, 12, 31)
// 	assertDate(t, endOfLastMonth(mustParse("2018-01-15")), 2017, 12, 31)
// 	assertDate(t, endOfLastMonth(mustParse("2020-01-30")), 2019, 12, 31)
// 	assertDate(t, endOfLastMonth(mustParse("2020-01-01")), 2019, 12, 31)
// 	assertDate(t, endOfLastMonth(mustParse("2020-01-15")), 2019, 12, 31)
// 	assertDate(t, endOfLastMonth(mustParse("2020-02-29")), 2020, 1, 31)
// 	assertDate(t, endOfLastMonth(mustParse("2020-02-01")), 2020, 1, 31)
// 	assertDate(t, endOfLastMonth(mustParse("2020-02-15")), 2020, 1, 31)
// 	assertDate(t, endOfLastMonth(mustParse("2020-12-31")), 2020, 11, 30)
// 	assertDate(t, endOfLastMonth(mustParse("2020-12-01")), 2020, 11, 30)
// 	assertDate(t, endOfLastMonth(mustParse("2020-12-15")), 2020, 11, 30)
// 	assertDate(t, endOfLastMonth(mustParse("2020-03-15")), 2020, 2, 29)
// }
