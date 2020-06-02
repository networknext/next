package jsonrpc

import (
	"cloud.google.com/go/bigquery"
)

// InvoiceRecord contains an invoice for an individual buyer. This
// structure is inferred from the SQL query. There are nulls returned
// via the current query.
type InvoiceRecord struct {
	BuyerID                 string               `json:"buyerId`
	Date                    bigquery.NullDate    `json:"date"`
	SliceBillUsd            bigquery.NullFloat64 `json:"sliceBillUsd"`
	TotalBandwidthGb        bigquery.NullFloat64 `json:"totalBandwidthGb"`
	AvgUsdPerGb             bigquery.NullFloat64 `json:"avgUsdPerGb"`
	AvgImprovementMs        bigquery.NullFloat64 `json:"avgImprovementMs"`
	SessionsOnNetworkNext   int                  `json:"sessionsOnNetworkNext"`
	PercentageOnNetworkNext bigquery.NullFloat64 `json:"percentageOnNetworkNext"`
	HoursWith20msOrGreater  bigquery.NullFloat64 `json:"hoursWith20msOrGreater"`
	HoursOnNetworkNext      bigquery.NullFloat64 `json:"hoursOnNetworkNext"`
	HoursImproved           bigquery.NullFloat64 `json:"hoursImproved"`
	HoursDegraded           bigquery.NullFloat64 `json:"hoursDegraded"`
	hoursNoMeasurement      bigquery.NullFloat64 `json:"hoursNoMeasurement"`
	hours0To5Ms             bigquery.NullFloat64 `json:"hours0To5Ms"`
	hours5To10Ms            bigquery.NullFloat64 `json:"hours5To10Ms"`
	hours10To15Ms           bigquery.NullFloat64 `json:"hours10To15Ms"`
	hours15To20Ms           bigquery.NullFloat64 `json:"hours15To20Ms"`
	hours20To30Ms           bigquery.NullFloat64 `json:"hours20To30Ms"`
	hours30To50Ms           bigquery.NullFloat64 `json:"hours30To50Ms"`
	hours50To100Ms          bigquery.NullFloat64 `json:"hours50To100Ms"`
	hours100PlusMs          bigquery.NullFloat64 `json:"hours100PlusMs"`
	hoursPacketLossReduced  bigquery.NullFloat64 `json:"hoursPacketLossReduced"`
}
