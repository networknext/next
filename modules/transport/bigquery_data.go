package transport

import (
	"time"

	"cloud.google.com/go/bigquery"
)

// BigQueryBillingEntry contains 1 row of the BQ billing table
type BigQueryBillingEntry struct {
	Timestamp                 time.Time
	BuyerID                   int64
	SessionID                 int64
	SliceNumber               int64
	Next                      bigquery.NullBool
	DirectRTT                 float64
	DirectJitter              float64
	DirectPacketLoss          float64
	NextRTT                   bigquery.NullFloat64
	NextJitter                bigquery.NullFloat64
	NextPacketLoss            bigquery.NullFloat64
	NextRelays                []int64
	TotalPrice                int64
	ClientToServerPacketsLost bigquery.NullInt64
	ServerToClientPacketsLost bigquery.NullInt64
	Committed                 bigquery.NullBool
	Flagged                   bigquery.NullBool
	Multipath                 bigquery.NullBool
	NextBytesUp               bigquery.NullInt64
	NextBytesDown             bigquery.NullInt64
	Initial                   bigquery.NullBool
	DatacenterID              bigquery.NullInt64
	RttReduction              bigquery.NullBool
	PacketLossReduction       bigquery.NullBool
	NextRelaysPrice           []int64
	UserHash                  bigquery.NullInt64
	Latitude                  bigquery.NullFloat64
	Longitude                 bigquery.NullFloat64
	ISP                       bigquery.NullString
	ABTest                    bigquery.NullBool
	RouteDecision             bigquery.NullInt64
	ConnectionType            bigquery.NullInt64
	PlatformType              bigquery.NullInt64
	SdkVersion                bigquery.NullString
	PacketLoss                bigquery.NullFloat64
	EnvelopeBytesUp           bigquery.NullInt64
	EnvelopeBytesDown         bigquery.NullInt64
	PredictedNextRTT          bigquery.NullFloat64
	MultipathVetoed           bigquery.NullBool
}
