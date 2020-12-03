package transport

import (
	"time"

	"cloud.google.com/go/bigquery"
)

// BigQueryBillingEntry contains 1 row of the BQ billing table
type BigQueryBillingEntry struct {
	Timestamp                       time.Time
	BuyerID                         int64
	BuyerString                     string
	SessionID                       int64
	SliceNumber                     int64
	Next                            bigquery.NullBool
	DirectRTT                       float64
	DirectJitter                    float64
	DirectPacketLoss                float64
	NextRTT                         bigquery.NullFloat64
	NextJitter                      bigquery.NullFloat64
	NextPacketLoss                  bigquery.NullFloat64
	NextRelays                      []int64
	NextRelaysStrings               []string
	TotalPrice                      int64
	ClientToServerPacketsLost       bigquery.NullInt64
	ServerToClientPacketsLost       bigquery.NullInt64
	Committed                       bigquery.NullBool
	Flagged                         bigquery.NullBool
	Multipath                       bigquery.NullBool
	NextBytesUp                     bigquery.NullInt64
	NextBytesDown                   bigquery.NullInt64
	Initial                         bigquery.NullBool
	DatacenterID                    bigquery.NullInt64
	DatacenterString                bigquery.NullString
	RttReduction                    bigquery.NullBool
	PacketLossReduction             bigquery.NullBool
	NextRelaysPrice                 []int64
	UserHash                        bigquery.NullInt64
	Latitude                        bigquery.NullFloat64
	Longitude                       bigquery.NullFloat64
	ISP                             bigquery.NullString
	ABTest                          bigquery.NullBool
	ConnectionType                  bigquery.NullInt64
	PlatformType                    bigquery.NullInt64
	SdkVersion                      bigquery.NullString
	PacketLoss                      bigquery.NullFloat64
	EnvelopeBytesUp                 bigquery.NullInt64
	EnvelopeBytesDown               bigquery.NullInt64
	PredictedNextRTT                bigquery.NullFloat64
	MultipathVetoed                 bigquery.NullBool
	Debug                           bigquery.NullString
	FallbackToDirect                bigquery.NullBool
	ClientFlags                     bigquery.NullInt64
	UserFlags                       bigquery.NullInt64
	NearRelayRTT                    bigquery.NullFloat64
	PacketsOutOfOrderClientToServer bigquery.NullInt64
	PacketsOutOfOrderServerToClient bigquery.NullInt64
	JitterClientToServer            bigquery.NullFloat64
	JitterServerToClient            bigquery.NullFloat64
	NumNearRelays                   bigquery.NullInt64
	NearRelayIDs                    []int64
	NearRelayRTTs                   []float64
	NearRelayJitters                []float64
	NearRelayPacketLosses           []float64
}
