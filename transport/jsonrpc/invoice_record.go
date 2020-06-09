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
	HoursNoMeasurement      bigquery.NullFloat64 `json:"hoursNoMeasurement"`
	Hours0To5Ms             bigquery.NullFloat64 `json:"hours0To5Ms"`
	Hours5To10Ms            bigquery.NullFloat64 `json:"hours5To10Ms"`
	Hours10To15Ms           bigquery.NullFloat64 `json:"hours10To15Ms"`
	Hours15To20Ms           bigquery.NullFloat64 `json:"hours15To20Ms"`
	Hours20To30Ms           bigquery.NullFloat64 `json:"hours20To30Ms"`
	Hours30To50Ms           bigquery.NullFloat64 `json:"hours30To50Ms"`
	Hours50To100Ms          bigquery.NullFloat64 `json:"hours50To100Ms"`
	Hours100PlusMs          bigquery.NullFloat64 `json:"hours100PlusMs"`
	HoursPacketLossReduced  bigquery.NullFloat64 `json:"hoursPacketLossReduced"`
}
