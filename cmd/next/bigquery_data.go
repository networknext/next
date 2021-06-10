package main

import (
	"time"

	"cloud.google.com/go/bigquery"
)

// BigQueryBillingEntry contains 1 row of the BQ billing table
type BigQueryBillingEntry struct {
	ABTest                          bigquery.NullBool
	BuyerID                         int64
	BuyerString                     string
	ClientFlags                     bigquery.NullInt64
	ClientToServerPacketsLost       bigquery.NullInt64
	ClientToServerPacketsSent       bigquery.NullInt64
	Committed                       bigquery.NullBool
	CommitVeto                      bigquery.NullBool
	ConnectionType                  bigquery.NullInt64
	DatacenterID                    bigquery.NullInt64
	DatacenterString                bigquery.NullString
	Debug                           bigquery.NullString
	DirectJitter                    float64
	DirectPacketLoss                float64
	DirectRTT                       float64
	EnvelopeBytesDown               bigquery.NullInt64
	EnvelopeBytesUp                 bigquery.NullInt64
	FallbackToDirect                bigquery.NullBool
	Flagged                         bigquery.NullBool
	ISP                             bigquery.NullString
	JitterClientToServer            bigquery.NullFloat64
	JitterServerToClient            bigquery.NullFloat64
	LackOfDiversity                 bigquery.NullBool
	LatencyWorse                    bigquery.NullBool
	Latitude                        bigquery.NullFloat64
	Longitude                       bigquery.NullFloat64
	Mispredicted                    bigquery.NullBool
	Multipath                       bigquery.NullBool
	MultipathRestricted             bigquery.NullBool
	MultipathVetoed                 bigquery.NullBool
	NearRelayIDs                    []int64
	NearRelayJitters                []float64
	NearRelayPacketLosses           []float64
	NearRelayRTT                    bigquery.NullFloat64
	NearRelayRTTs                   []float64
	Next                            bigquery.NullBool
	Initial                         bool
	NextBytesDown                   bigquery.NullInt64
	NextBytesUp                     bigquery.NullInt64
	NextJitter                      bigquery.NullFloat64
	NextLatencyTooHigh              bigquery.NullBool
	NextPacketLoss                  bigquery.NullFloat64
	NextRelay0                      bigquery.NullInt64
	NextRelay1                      bigquery.NullInt64
	NextRelay2                      bigquery.NullInt64
	NextRelay3                      bigquery.NullInt64
	NextRelay4                      bigquery.NullInt64
	NextRelaysPrice                 []int64
	NextRelaysStrings               string
	NextRTT                         bigquery.NullFloat64
	NoRoute                         bigquery.NullBool
	NumNearRelays                   bigquery.NullInt64
	PacketLoss                      bigquery.NullFloat64
	PacketLossReduction             bigquery.NullBool
	PacketsOutOfOrderClientToServer bigquery.NullInt64
	PacketsOutOfOrderServerToClient bigquery.NullInt64
	PlatformType                    bigquery.NullInt64
	PredictedNextRTT                bigquery.NullFloat64
	Pro                             bigquery.NullBool
	RelayWentAway                   bigquery.NullBool
	RouteChanged                    bigquery.NullBool
	RouteDiversity                  bigquery.NullInt64
	RouteLost                       bigquery.NullBool
	RttReduction                    bigquery.NullBool
	SdkVersion                      bigquery.NullString
	ServerToClientPacketsLost       bigquery.NullInt64
	ServerToClientPacketsSent       bigquery.NullInt64
	SessionID                       int64
	SliceNumber                     int64
	Tags                            []int64
	Timestamp                       time.Time
	TotalPrice                      bigquery.NullInt64
	UserFlags                       bigquery.NullInt64
	UserHash                        bigquery.NullInt64
	Vetoed                          bigquery.NullBool
	UnknownDatacenter               bigquery.NullBool
	DatacenterNotEnabled            bigquery.NullBool
	BuyerNotLive                    bigquery.NullBool
	StaleRouteMatrix                bigquery.NullBool
}

// BigQueryRelayPingsEntry contains 1 row of the BQ relay_pings table
type BigQueryRelayPingsEntry struct {
	Timestamp  time.Time `bigquery:"timestamp"`
	RelayA     int64     `bigquery:"relay_a"`
	RelayB     int64     `bigquery:"relay_b"`
	RTT        float64   `bigquery:"rtt"`
	Jitter     float64   `bigquery:"jitter"`
	PacketLoss float64   `bigquery:"packet_loss"`
	Routable   bool      `bigquery:"routable"`
}
