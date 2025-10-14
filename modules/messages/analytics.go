package messages

// ----------------------------------------------------------------------------------------

type AnalyticsSessionUpdateMessage struct {

	// always

	Timestamp         int64   `avro:"timestamp"`
	SessionId         int64   `avro:"session_id"`
	ServerId          int64   `avro:"server_id"`
	SliceNumber       int32   `avro:"slice_number"`
	RealPacketLoss    float32 `avro:"real_packet_loss"`
	RealJitter        float32 `avro:"real_jitter"`
	RealOutOfOrder    float32 `avro:"real_out_of_order"`
	SessionEvents     int64   `avro:"session_events"`
	InternalEvents    int64   `avro:"internal_events"`
	DirectRTT         float32 `avro:"direct_rtt"`
	DirectJitter      float32 `avro:"direct_jitter"`
	DirectPacketLoss  float32 `avro:"direct_packet_loss"`
	BandwidthKbpsUp   int32   `avro:"bandwidth_kbps_up"`
	BandwidthKbpsDown int32   `avro:"bandwidth_kbps_down"`
	DeltaTimeMin      float32 `avro:"delta_time_min"`
	DeltaTimeMax      float32 `avro:"delta_time_max"`
	DeltaTimeAvg      float32 `avro:"delta_time_avg"`
	GameRTT           float32 `avro:"game_rtt"`
	GameJitter        float32 `avro:"game_jitter"`
	GamePacketLoss    float32 `avro:"game_packet_loss"`

	// next only

	NextRTT          float32 `avro:"next_rtt"`
	NextJitter       float32 `avro:"next_jitter"`
	NextPacketLoss   float32 `avro:"next_packet_loss"`
	NextPredictedRTT float32 `avro:"next_predicted_rtt"`
	NextRouteRelays  []int64 `avro:"next_route_relays"`

	// flags

	Next                bool  `avro:"next"`
	FallbackToDirect    bool  `avro:"fallback_to_direct"`
	Reported            bool  `avro:"reported"`
	LatencyReduction    bool  `avro:"latency_reduction"`
	PacketLossReduction bool  `avro:"packet_loss_reduction"`
	ForceNext           bool  `avro:"force_next"`
	LongSessionUpdate   bool  `avro:"long_session_update"`
	Veto                bool  `avro:"veto"`
	Disabled            bool  `avro:"disabled"`
	NotSelected         bool  `avro:"not_selected"`
	A                   bool  `avro:"a"`
	B                   bool  `avro:"b"`
	LatencyWorse        bool  `avro:"latency_worse"`
	Mispredict          bool  `avro:"mispredict"`
	LackOfDiversity     bool  `avro:"lack_of_diversity"`
	Flags               int64 `avro:"flags"`
}

// ----------------------------------------------------------------------------------------

type AnalyticsSessionSummaryMessage struct {

	// summary data

	Timestamp                       int64   `avro:"timestamp"`
	SessionId                       int64   `avro:"session_id"`
	MatchId                         int64   `avro:"match_id"`
	DatacenterId                    int64   `avro:"datacenter_id"`
	BuyerId                         int64   `avro:"buyer_id"`
	UserHash                        int64   `avro:"user_hash"`
	Latitude                        float32 `avro:"latitude"`
	Longitude                       float32 `avro:"longitude"`
	ClientAddress                   string  `avro:"client_address"`
	ServerAddress                   string  `avro:"server_address"`
	ConnectionType                  int32   `avro:"connection_type"`
	PlatformType                    int32   `avro:"platform_type"`
	SDKVersion_Major                int32   `avro:"sdk_version_major"`
	SDKVersion_Minor                int32   `avro:"sdk_version_minor"`
	SDKVersion_Patch                int32   `avro:"sdk_version_patch"`
	ClientToServerPacketsSent       int64   `avro:"client_to_server_packets_sent"`
	ServerToClientPacketsSent       int64   `avro:"server_to_client_packets_sent"`
	ClientToServerPacketsLost       int64   `avro:"client_to_server_packets_lost"`
	ServerToClientPacketsLost       int64   `avro:"server_to_client_packets_lost"`
	ClientToServerPacketsOutOfOrder int64   `avro:"client_to_server_packets_out_of_order"`
	ServerToClientPacketsOutOfOrder int64   `avro:"server_to_client_packets_out_of_order"`
	TotalNextEnvelopeBytesUp        int64   `avro:"total_next_envelope_bytes_up"`
	TotalNextEnvelopeBytesDown      int64   `avro:"total_next_envelope_bytes_down"`
	DurationOnNext                  int32   `avro:"duration_on_next"`
	SessionDuration                 int32   `avro:"session_duration"`
	StartTimestamp                  int64   `avro:"start_timestamp"`
	Error                           int64   `avro:"error"`
	ISP                             string  `avro:"isp"`
	Country                         string  `avro:"country"`
	BestLatencyReduction            int64   `avro:"best_latency_reduction"`
	FallbackToDirect                bool    `avro:"fallback_to_direct"`
	NextLatencyTooHigh              bool    `avro:"next_latency_too_high"`
	LikelyVPNOrCrossRegion          bool    `avro:"likely_vpn_or_cross_region"`
	NoClientRelays                  bool    `avro:"no_client_relays"`
	NoServerRelays                  bool    `avro:"no_server_relays"`
	AllClientRelaysAreZero          bool    `avro:"all_client_relays_are_zero"`

	// flags

	Reported            bool  `avro:"reported"`
	LatencyReduction    bool  `avro:"latency_reduction"`
	PacketLossReduction bool  `avro:"packet_loss_reduction"`
	ForceNext           bool  `avro:"force_next"`
	LongSessionUpdate   bool  `avro:"long_session_update"`
	Veto                bool  `avro:"veto"`
	Disabled            bool  `avro:"disabled"`
	NotSelected         bool  `avro:"not_selected"`
	A                   bool  `avro:"a"`
	B                   bool  `avro:"b"`
	LatencyWorse        bool  `avro:"latency_worse"`
	Mispredict          bool  `avro:"mispredict"`
	LackOfDiversity     bool  `avro:"lack_of_diversity"`
	Flags               int64 `avro:"flags"`
}

// ----------------------------------------------------------------------------------------

type AnalyticsServerInitMessage struct {
	Timestamp        int64  `avro:"timestamp"`
	SDKVersion_Major int32  `avro:"sdk_version_major"`
	SDKVersion_Minor int32  `avro:"sdk_version_minor"`
	SDKVersion_Patch int32  `avro:"sdk_version_patch"`
	BuyerId          int64  `avro:"buyer_id"`
	MatchId          int64  `avro:"match_id"`
	DatacenterId     int64  `avro:"datacenter_id"`
	DatacenterName   string `avro:"datacenter_name"`
	ServerId         int64  `avro:"server_id"`
	ServerAddress    string `avro:"server_address"`
}

// ----------------------------------------------------------------------------------------

type AnalyticsServerUpdateMessage struct {
	Timestamp        int64   `avro:"timestamp"`
	SDKVersion_Major int32   `avro:"sdk_version_major"`
	SDKVersion_Minor int32   `avro:"sdk_version_minor"`
	SDKVersion_Patch int32   `avro:"sdk_version_patch"`
	BuyerId          int64   `avro:"buyer_id"`
	MatchId          int64   `avro:"match_id"`
	DatacenterId     int64   `avro:"datacenter_id"`
	NumSessions      int32   `avro:"num_sessions"`
	ServerId         int64   `avro:"server_id"`
	ServerAddress    string  `avro:"server_address"`
	DeltaTimeMin     float32 `avro:"delta_time_min"`
	DeltaTimeMax     float32 `avro:"delta_time_max"`
	DeltaTimeAvg     float32 `avro:"delta_time_avg"`
}

// ----------------------------------------------------------------------------------------

type AnalyticsRelayUpdateMessage struct {
	Timestamp                 int64   `avro:"timestamp"`
	RelayId                   int64   `avro:"relay_id"`
	SessionCount              int32   `avro:"session_count"`
	MaxSessions               int32   `avro:"max_sessions"`
	EnvelopeBandwidthUpKbps   int64   `avro:"envelope_bandwidth_up_kbps"`
	EnvelopeBandwidthDownKbps int64   `avro:"envelope_bandwidth_down_kbps"`
	PacketsSentPerSecond      float32 `avro:"packets_sent_per_second"`
	PacketsReceivedPerSecond  float32 `avro:"packets_received_per_second"`
	BandwidthSentKbps         float32 `avro:"bandwidth_sent_kbps"`
	BandwidthReceivedKbps     float32 `avro:"bandwidth_received_kbps"`
	ClientPingsPerSecond      float32 `avro:"client_pings_per_second"`
	ServerPingsPerSecond      float32 `avro:"server_pings_per_second"`
	RelayPingsPerSecond       float32 `avro:"relay_pings_per_second"`
	RelayFlags                int64   `avro:"relay_flags"`
	NumRoutable               int32   `avro:"num_routable"`
	NumUnroutable             int32   `avro:"num_unroutable"`
	StartTime                 int64   `avro:"start_time"`
	CurrentTime               int64   `avro:"current_time"`
	RelayCounters             []int64 `avro:"relay_counters"`
}

// ----------------------------------------------------------------------------------------

type AnalyticsRouteMatrixUpdateMessage struct {
	Timestamp               int64   `avro:"timestamp"`
	NumRelays               int32   `avro:"num_relays"`
	NumActiveRelays         int32   `avro:"num_active_relays"`
	NumDestRelays           int32   `avro:"num_dest_relays"`
	NumDatacenters          int32   `avro:"num_datacenters"`
	TotalRoutes             int32   `avro:"total_routes"`
	AverageNumRoutes        float32 `avro:"average_num_routes"`
	AverageRouteLength      float32 `avro:"average_route_length"`
	NoRoutePercent          float32 `avro:"no_route_percent"`
	OneRoutePercent         float32 `avro:"one_route_percent"`
	NoDirectRoutePercent    float32 `avro:"no_direct_route_percent"`
	RTTBucket_NoImprovement float32 `avro:"rtt_bucket_no_improvement"`
	RTTBucket_0_5ms         float32 `avro:"rtt_bucket_0_5ms"`
	RTTBucket_5_10ms        float32 `avro:"rtt_bucket_5_10ms"`
	RTTBucket_10_15ms       float32 `avro:"rtt_bucket_10_15ms"`
	RTTBucket_15_20ms       float32 `avro:"rtt_bucket_15_20ms"`
	RTTBucket_20_25ms       float32 `avro:"rtt_bucket_20_25ms"`
	RTTBucket_25_30ms       float32 `avro:"rtt_bucket_25_30ms"`
	RTTBucket_30_35ms       float32 `avro:"rtt_bucket_30_35ms"`
	RTTBucket_35_40ms       float32 `avro:"rtt_bucket_35_40ms"`
	RTTBucket_40_45ms       float32 `avro:"rtt_bucket_40_45ms"`
	RTTBucket_45_50ms       float32 `avro:"rtt_bucket_45_50ms"`
	RTTBucket_50ms_Plus     float32 `avro:"rtt_bucket_50ms_plus"`
	CostMatrixSize          int32   `avro:"cost_matrix_size"`
	RouteMatrixSize         int32   `avro:"route_matrix_size"`
	DatabaseSize            int32   `avro:"database_size"`
	OptimizeTime            int32   `avro:"optimize_time"`
}

// ----------------------------------------------------------------------------------------

type AnalyticsClientRelayPingMessage struct {
	Timestamp             int64   `avro:"timestamp"`
	BuyerId               int64   `avro:"buyer_id"`
	SessionId             int64   `avro:"session_id"`
	UserHash              int64   `avro:"user_hash"`
	Latitude              float32 `avro:"latitude"`
	Longitude             float32 `avro:"longitude"`
	ClientAddress         string  `avro:"client_address"`
	ConnectionType        int32   `avro:"connection_type"`
	PlatformType          int32   `avro:"platform_type"`
	ClientRelayId         int64   `avro:"client_relay_id"`
	ClientRelayRTT        int32   `avro:"client_relay_rtt"`
	ClientRelayJitter     int32   `avro:"client_relay_jitter"`
	ClientRelayPacketLoss float32 `avro:"client_relay_packet_loss"`
}

// ----------------------------------------------------------------------------------------

type AnalyticsServerRelayPingMessage struct {
	Timestamp             int64   `avro:"timestamp"`
	BuyerId               int64   `avro:"buyer_id"`
	SessionId             int64   `avro:"session_id"`
	DatacenterId          int64   `avro:"datacenter_id"`
	ServerAddress         string  `avro:"server_address"`
	ServerRelayId         int64   `avro:"server_relay_id"`
	ServerRelayRTT        int32   `avro:"server_relay_rtt"`
	ServerRelayJitter     int32   `avro:"server_relay_jitter"`
	ServerRelayPacketLoss float32 `avro:"server_relay_packet_loss"`
}

// ----------------------------------------------------------------------------------------

type AnalyticsRelayToRelayPingMessage struct {
	Timestamp          int64   `avro:"timestamp"`
	SourceRelayId      int64   `avro:"source_relay_id"`
	DestinationRelayId int64   `avro:"destination_relay_id"`
	RTT                int32   `avro:"rtt"`
	Jitter             int32   `avro:"jitter"`
	PacketLoss         float32 `avro:"packet_loss"`
}

// ----------------------------------------------------------------------------------------
