package jsonrpc

import (
	"cloud.google.com/go/bigquery"
)

// InvoiceRecord contains an invoice for an individual buyer. This
// structure is inferred from the SQL query. There are nulls returned
// via the currrent query.
type InvoiceRecord struct {
	BuyerID                 string
	Date                    bigquery.NullDate
	SliceBillUsd            bigquery.NullFloat64
	TotalBandwidthGb        bigquery.NullFloat64
	AvgUsdPerGb             bigquery.NullFloat64
	AvgImprovementMs        bigquery.NullFloat64
	SessionsOnNetworkNext   int
	PercentageOnNetworkNext bigquery.NullFloat64
	HoursWith20msOrGreater  bigquery.NullFloat64
	HoursOnNetworkNext      bigquery.NullFloat64
	HoursImproved           bigquery.NullFloat64
	HoursDegraded           bigquery.NullFloat64
	hoursNoMeasurement      bigquery.NullFloat64
	hours0To5Ms             bigquery.NullFloat64
	hours5To10Ms            bigquery.NullFloat64
	hours10To15Ms           bigquery.NullFloat64
	hours15To20Ms           bigquery.NullFloat64
	hours20To30Ms           bigquery.NullFloat64
	hours30To50Ms           bigquery.NullFloat64
	hours50To100Ms          bigquery.NullFloat64
	hours100PlusMs          bigquery.NullFloat64
	hoursPacketLossReduced  bigquery.NullFloat64
}
