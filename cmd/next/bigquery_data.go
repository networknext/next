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

// BigQueryBilling2Entry contains 1 row of the BQ billing2 table
type BigQueryBilling2Entry struct {

	// always
	Timestamp        time.Time
	SessionID        int64
	SliceNumber      int64
	DirectRTT        int64 // This is needed for reading only
	DirectMinRTT     int64
	DirectMaxRTT     bigquery.NullInt64
	DirectPrimeRTT   bigquery.NullInt64
	DirectJitter     int64
	DirectPacketLoss int64
	RealPacketLoss   float64
	RealJitter       int64
	Next             bigquery.NullBool
	Flagged          bigquery.NullBool
	Summary          bigquery.NullBool
	Debug            bigquery.NullString
	RouteDiversity   bigquery.NullInt64

	// first slice and summary slice only

	DatacenterID      bigquery.NullInt64
	DatacenterString  bigquery.NullString
	BuyerID           bigquery.NullInt64
	UserHash          bigquery.NullInt64
	EnvelopeBytesUp   bigquery.NullInt64
	EnvelopeBytesDown bigquery.NullInt64
	Latitude          bigquery.NullFloat64
	Longitude         bigquery.NullFloat64
	ClientAddress     bigquery.NullString
	ISP               bigquery.NullString
	ConnectionType    bigquery.NullInt64
	PlatformType      bigquery.NullInt64
	SDKVersion        bigquery.NullString
	Tags              []bigquery.NullInt64
	ABTest            bigquery.NullBool
	Pro               bigquery.NullBool

	// summary slice only

	ClientToServerPacketsSent       bigquery.NullInt64
	ServerToClientPacketsSent       bigquery.NullInt64
	ClientToServerPacketsLost       bigquery.NullInt64
	ServerToClientPacketsLost       bigquery.NullInt64
	ClientToServerPacketsOutOfOrder bigquery.NullInt64
	ServerToClientPacketsOutOfOrder bigquery.NullInt64
	NearRelayIDs                    []bigquery.NullInt64
	NearRelayRTTs                   []bigquery.NullInt64
	NearRelayJitters                []bigquery.NullInt64
	NearRelayPacketLosses           []bigquery.NullInt64
	EverOnNext                      bigquery.NullBool
	SessionDuration                 bigquery.NullInt64
	TotalPriceSum                   bigquery.NullInt64
	EnvelopeBytesUpSum              bigquery.NullInt64
	EnvelopeBytesDownSum            bigquery.NullInt64
	DurationOnNext                  bigquery.NullInt64
	StartTimestamp                  bigquery.NullTimestamp

	// network next only

	NextRTT             bigquery.NullInt64
	NextJitter          bigquery.NullInt64
	NextPacketLoss      bigquery.NullInt64
	PredictedNextRTT    bigquery.NullInt64
	NearRelayRTT        bigquery.NullInt64
	NextRelay0          bigquery.NullInt64
	NextRelay1          bigquery.NullInt64
	NextRelay2          bigquery.NullInt64
	NextRelay3          bigquery.NullInt64
	NextRelay4          bigquery.NullInt64
	NextRelaysPrice     []bigquery.NullInt64
	NextRelaysStrings   string
	TotalPrice          bigquery.NullInt64
	Uncommitted         bigquery.NullBool
	Multipath           bigquery.NullBool
	RTTReduction        bigquery.NullBool
	PacketLossReduction bigquery.NullBool
	RouteChanged        bigquery.NullBool
	NextBytesUp         bigquery.NullInt64
	NextBytesDown       bigquery.NullInt64

	// error state only

	FallbackToDirect     bigquery.NullBool
	MultipathVetoed      bigquery.NullBool
	Mispredicted         bigquery.NullBool
	Vetoed               bigquery.NullBool
	LatencyWorse         bigquery.NullBool
	NoRoute              bigquery.NullBool
	NextLatencyTooHigh   bigquery.NullBool
	CommitVeto           bigquery.NullBool
	UnknownDatacenter    bigquery.NullBool
	DatacenterNotEnabled bigquery.NullBool
	BuyerNotLive         bigquery.NullBool
	StaleRouteMatrix     bigquery.NullBool
}

// BigQueryBilling2EntrySummary contains 1 row of the BQ billing2_session_summary table
type BigQueryBilling2EntrySummary struct {
	SessionID                       int64
	BuyerID                         bigquery.NullInt64
	UserHash                        bigquery.NullInt64
	DatacenterID                    bigquery.NullInt64
	DatacenterString                bigquery.NullString
	StartTimestamp                  bigquery.NullTimestamp
	Latitude                        bigquery.NullFloat64
	Longitude                       bigquery.NullFloat64
	ISP                             bigquery.NullString
	ConnectionType                  bigquery.NullInt64
	PlatformType                    bigquery.NullInt64
	Tags                            []bigquery.NullInt64
	ABTest                          bigquery.NullBool
	Pro                             bigquery.NullBool
	SDKVersion                      bigquery.NullString
	EnvelopeBytesUp                 bigquery.NullInt64
	EnvelopeBytesDown               bigquery.NullInt64
	ClientToServerPacketsSent       bigquery.NullInt64
	ServerToClientPacketsSent       bigquery.NullInt64
	ClientToServerPacketsLost       bigquery.NullInt64
	ServerToClientPacketsLost       bigquery.NullInt64
	ClientToServerPacketsOutOfOrder bigquery.NullInt64
	ServerToClientPacketsOutOfOrder bigquery.NullInt64
	NearRelayIDs                    []bigquery.NullInt64
	NearRelayRTTs                   []bigquery.NullInt64
	NearRelayJitters                []bigquery.NullInt64
	NearRelayPacketLosses           []bigquery.NullInt64
	EverOnNext                      bigquery.NullBool
	SessionDuration                 bigquery.NullInt64
	TotalPriceSum                   bigquery.NullInt64
	EnvelopeBytesUpSum              bigquery.NullInt64
	EnvelopeBytesDownSum            bigquery.NullInt64
	DurationOnNext                  bigquery.NullInt64
	ClientAddress                   bigquery.NullString
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
	InstanceID string    `bigquery:"instance_id"`
	Debug      bool      `bigquery:"debug"`
}
