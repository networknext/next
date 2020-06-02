package jsonrpc

import (
	"time"
)

// InvoiceRecord contains an invoice for an individual buyer. This
// structure is inferred from the SQL query.
type InvoiceRecord struct {
	BuyerID                 string
	Date                    time.Time
	SliceBillUsd            float32
	TotalBandwidthGb        int
	AvgUsdPerGb             float32
	AvgImprovementMs        float32
	SessionsOnNetworkNext   int
	PercentageOnNetworkNext float32
	HoursWith20msOrGreater  int
	HoursOnNetworkNext      int
	HoursImproved           int
	HoursDegraded           int
	hoursNoMeasurement      int
	hours0To5Ms             int
	hours5To10Ms            int
	hours10To15Ms           int
	hours15To20Ms           int
	hours20To30Ms           int
	hours30To50Ms           int
	hours50To100Ms          int
	hours100PlusMs          int
	hoursPacketLossReduced  int
}
