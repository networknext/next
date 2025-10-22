<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Bigquery table schemas

Network Next writes data to bigquery so you can run data science and analytics queries.
 
For example, once every 10 seconds network performance data such as accelerated and non-accelerated latency (RTT), jitter, packet loss, bandwidth usage and out order packets for every session is written to bigquery. 

At the end of each session a summary data entry is written, which makes it much faster and cheaper to query data on a per-session basis.

There is also data written each time a client pings relays at the start of each session, so you can look at direct ping results from clients to relays, and data from each server and relay in your fleet is sent so you can track their performance and uptime.

Data in bigquery is retained for 90 days by default to comply with GDPR. Within this dataset the "user hash" is considered pseudonymized data within the GDPR, and is personal data only if you have the ability to identify the player from the user hash plus other data that you have for this player. The client IP address and port are also considered personal data.

Schemas for all this data are described below.

## Session Update

Session updates contain network performance data once every 10 seconds for a session. This is the primary network performance data, including everything shown in the portal for a session and much more.

| Field | Type | Description |
| ------------- | ------------- | ------------- |
| timestamp | TIMESTAMP | The timestamp when the session update occurred |
| session_id | INT64 | Unique identifier for this session |
| slice_number | INT64 | Slices are 10 second periods starting from slice number 0 at the start of the session |
| real_packet_loss | FLOAT64 | Packet loss between the client and the server measured from game packets (%) |
| real_jitter | FLOAT64 | Jitter between the client and the server measured from game packets (milliseconds) |
| real_out_of_order | FLOAT64 | Percentage of packets that arrive out of order between the client and the server (%) |
| session_events | INT64 | Custom set of 64bit event flags. Optional. NULL if no flags are set |
| internal_events | INT64 | Internal SDK event flags. Optional. NULL if no flags are set |
| direct_rtt | FLOAT64 | Latency between client and server as measured by direct pings (unaccelerated path). Milliseconds. IMPORTANT: Will be 0.0 on slice 0 always because it is not known yet |
| direct_jitter | FLOAT64 | Jitter between client and server as measured by direct pings (unaccelerated path). Milliseconds. IMPORTANT: Will be 0.0 on slice 0 always because it is not known yet |
| direct_packet_loss | FLOAT64 | Packet loss between client and server as measured by direct pings (unaccelerated path). Percent. Generally this is inaccurate and higher that real value, because direct pings are sent infrequently, any packet loss results in an outsized % of packet loss in a 10 second period. |
| direct_kbps_up | INT64 | Bandwidth in the client to server direction along the direct path (unaccelerated). Kilobits per-second |
| direct_kbps_down | INT64 | Bandwidth in the server to client direction along the direct path (unaccelerated). Kilobits per-second |
| next | BOOL | True if this slice is being accelerated over network next |
| next_rtt | FLOAT64 | Latency between client and server as measured by next pings (accelerated path). Milliseconds. NULL if not on network next |
| next_jitter | FLOAT64 | Jitter between client and server as measured by next pings (accelerated path). Milliseconds. NULL if not on network next |
| next_packet_loss | FLOAT64 | Packet loss between client and server as measured by next pings (accelerated path). Percent. NULL if not on network next. Generally inaccurate and higher than real packet loss due to infrequent sending of ping packets |
| next_kbps_up | FLOAT64 | Bandwidth in the client to server direction along the next path (accelerated). Kilobits per-second |
| next_kbps_down | FLOAT64 | Bandwidth in the server to client direction along the next path (accelerated). Kilobits per-second |
| next_predicted_rtt | FLOAT64 | Conservative predicted latency between client and server from the control plane. Milliseconds. NULL if not on network next |
| next_route_relays | []INT64 | Array of relay ids for the network next path (accelerated). NULL if not on network next |
| fallback_to_direct | BOOL | True if the SDK has encountered a fatal error and cannot continue acceleration. Typically this only happens when the system is misconfigured or overloaded. |
| reported | BOOL | True if this session was reported by the player |
| latency_reduction | BOOL | True if this session took network next this slice to reduce latency |
| packet_loss_reduction | BOOL | True if this session took network next this slice to reduce packet loss |
| force_next | BOOL | True if this session took network next this slice because it was forced to |
| long_session_update | BOOL | True if the processing for this slice on the server backend took a long time. This may indicate that the server backend is overloaded. |
| veto | BOOL | True if the routing logic decided that this session should no longer be accelerated for some reason. |
| disabled | BOOL | True if the buyer is disabled. Disabled buyers don't perform any acceleration or analytics on network next. |
| not_selected | BOOL | If the route shader selection % is any value other than 100%, then this is true for sessions that were not selected for acceleration. |
| a | BOOL | This session was part of an AB test, and is in the A group. (potentially accelerated) |
| b | BOOL | This session was part of an AB test, and is in the B group. (never accelerated) |

## Session Summary

A session summary is written at the end of each session, with the intent that if you want per-session data you can query it here, instead of needing to process all the 10 second slices belonging to that session to get the data you want.

| Field | Type | Description |
| ------------- | ------------- | ------------- |
| timestamp | TIMESTAMP | The timestamp when the session summary was generated (at the end of the session). |
| session_id | INT64 | Unique identifier for this session |
| datacenter_id | INT64 | The datacenter the server is in |
| buyer_id | INT64 | The buyer this session belongs to |
| user_hash | INT64 | Pseudonymized hash of a unique user id passed up from the SDK |
| latitude | FLOAT64 | Approximate latitude of the player from ip2location |
| longitude | FLOAT64 | Approximate longitude of the player from ip2location |
| client_address | STRING | Anonymized client address |
| server_address | STRING | Server address and port number |
| connection_type | INT64 | Connection type: 0 = unknown, 1 = wired, 2 = wifi, 3 = cellular |
| platform_type | INT64 | Platform type: 0 = unknown, 1 = windows, 2 = mac, 3 = linux, 4 = switch, 5 = ps4, 6 = ios, 7 = xbox one, 8 = xbox series x, 9 = ps5 |
| sdk_version_major | INT64 | The major SDK version on the server |
| sdk_version_minor | INT64 | The minor SDK version on the server |
| sdk_version_patch | INT64 | The patch SDK version on the server |
| client_to_server_packets_sent | INT64 | The total number of game packets sent from client to server in this session |
| server_to_client_packets_sent | INT64 | The total number of game packets sent from server to client in this session |
| client_to_server_packets_lost | INT64 | The total number of game packets lost from client to server in this session |
| server_to_client_packets_lost | INT64 | The total number of game packets lost from server to client in this session |
| client_to_server_packets_out_of_order | INT64 | The total number of game packets received out of order from client to server in this session |
| server_to_client_packets_out_of_order | INT64 | The total number of game packets received out of order from server to client in this session | INT64 | The total number of envelope bytes sent across network next in the client to server direction for this session |
| total_next_envelope_bytes_up | INT64 | The total number of envelope bytes sent across network next in the client to server direction for this session |
| total_next_envelope_bytes_down | INT64 | The total number of envelope bytes sent across network next in the server to client direction for this session |
| duration_on_next | INT64 | Total time spent on network next in this session (time accelerated). Seconds |
| session_duration | INT64 | Length of this session in seconds |
| start_timestamp | TIMESTAMP | The time when this session started |
| error | INT64 | Error flags to diagnose what's happening with a session. Look up SessionError_* in the codebase for a list of errors. 0 if no error has occurred. |
| reported | BOOL | True if this session was reported by the player |
| latency_reduction | BOOL | True if this session took network next to reduce latency |
| packet_loss_reduction | BOOL | True if this session took network next to reduce packet loss |
| force_next | BOOL | True if this session took network next because it was forced to |
| long_session_update | BOOL | True if the processing for any slices in this session took a long time. This may indicate that the server backend is overloaded. |
| veto | BOOL | True if the routing logic decided that this session should no longer be accelerated for some reason. |
| disabled | BOOL | True if the buyer is disabled. Disabled buyers don't perform any acceleration or analytics on network next. |
| not_selected | BOOL | If the route shader selection % is any value other than 100%, then this is true for sessions that were not selected for acceleration. |
| a | BOOL | This session was part of an AB test, and is in the A group (potentially accelerated) |
| b | BOOL | This session was part of an AB test, and is in the B group (never accelerated) |

## Server Init

Server init entries are added once when the server first connects with the Network Next backend. They are intended to let you quickly look up the datacenter name for servers, which is not included in the regular update.

| Field | Type | Description |
| ------------- | ------------- | ------------- |
| timestamp | TIMESTAMP | The timestamp when the server init occurred |
| sdk_version_major | INT64 | The major SDK version number on the server |
| sdk_version_minor | INT64 | The minor SDK version number on the server |
| sdk_version_patch | INT64 | The patch SDK version number on the server |
| buyer_id | INT64 | The buyer this server belongs to |
| datacenter_id | INT64 | The datacenter this server is in |
| datacenter_name | STRING | The name of the datacenter, for example: 'google.iowa.1' |
| server_address | STRING | The address and port of the server, for example: '123.254.10.5:40000' |

## Server Update

This entry is updated once every 10 seconds while a server is running. The only value that changes is the number of sessions connected to the server, which you can use to track server utilization.

| Field | Type | Description |
| ------------- | ------------- | ------------- |
| timestamp | TIMESTAMP | The timestamp when the server init occurred |
| sdk_version_major | INT64 | The major SDK version number on the server |
| sdk_version_minor | INT64 | The minor SDK version number on the server |
| sdk_version_patch | INT64 | The patch SDK version number on the server |
| buyer_id | INT64 | The buyer this server belongs to |
| datacenter_id | INT64 | The datacenter this server is in |
| server_address | STRING | The address and port of the server, for example: '123.254.10.5:40000' |
| num_sessions | INT64 | The number of client sessions currently connected to the server |

## Relay Update

This data is updated once every 10 seconds per-relay. It is useful for tracking the activity of your relay fleet, and identifying high and low performing relays.

| Field | Type | Description |
| ------------- | ------------- | ------------- |
| timestamp | TIMESTAMP | The timestamp when the relay update occurred |
| relay_id | INT64 | Unique relay id. The fnv1a hash of the relay address + port as a string |
| session_count | INT64 | The number of sessions currently going through this relay |
| max_sessions | INT64 | The maximum number of sessions allowed through this relay (optional: NULL if not specified) |
| envelope_bandwidth_up_kbps | INT64 | The current amount of envelope bandwidth in the client to server direction through this relay |
| envelope_bandwidth_down_kbps | INT64 | The current amount of envelope bandwidth in the server to client direction through this relay |
| packets_sent_per_second | FLOAT64 | The number of packets sent per-second by this relay |
| packets_received_per_second | FLOAT64 | The number of packets received per-second by this relay |
| bandwidth_sent_kbps | FLOAT64 | The amount of bandwidth sent by this relay in kilobits per-second |
| bandwidth_received_kbps | FLOAT64 | The amount of bandwidth received by this relay in kilobits per-second |
| client_pings_per_second | FLOAT64 | The number of client relay pings received by this relay per-second |
| server_pings_per_second | FLOAT64 | The number of server relay pings received by this relay per-second |
| relay_pings_per_second | FLOAT64 | The number of relay pings sent from other relays received by this relay per-second |
| relay_flags | INT64 | The current value of the relay flags. See RelayFlags_* in the source code |
| num_routable | INT64 | The number of other relays this relay can route to |
| num_unroutable | INT64 | The number of other relays this relay cannot route to |
| start_time | INT64 | The start time of the relay as a unix timestamp according to the clock on the relay |
| current_time | INT64 | The start time of the relay as a unix timestamp according to the clock on the relay. Together with start_time and timestamp this can be used to determine relay uptime, and clock desynchronization between the relay and the backend. |
| relay_counters | []INT64 | Array of counters used to diagnose what is going on with a relay. Search for RELAY_COUNTER_ in the codebase for counter names |

## Client Relay Ping

These entries are written to bigquery at the start of each session when relays are pinged by the client.

| Field | Type | Description |
| ------------- | ------------- | ------------- |
| timestamp | TIMESTAMP | The timestamp when the client relay ping occurred |
| buyer_id | INT64 | The buyer this player belongs to |
| session_id | INT64 | Unique id for the session |
| user_hash | INT64 | Pseudonymized hash of a user id passed up from the SDK |
| latitude | FLOAT64 | Approximate latitude of the player from ip2location |
| longitude | FLOAT64 | Approximate longitude of the player from ip2location |
| client_address | STRING | Anonymized client address |
| connection_type | INT64 | Connection type: 0 = unknown, 1 = wired, 2 = wifi, 3 = cellular |
| platform_type | INT64 | Platform type: 0 = unknown, 1 = windows, 2 = mac, 3 = linux, 4 = switch, 5 = ps4, 6 = ios, 7 = xbox one, 8 = xbox series x, 9 = ps5 |
| client_relay_id | INT64 | Relay id being pinged by the client |
| client_relay_rtt | INT64 | Round trip time ping between the client and the relay (milliseconds) |
| client_relay_jitter | INT64 | Jitter between the client and the relay (milliseconds) |
| client_relay_packet_loss | FLOAT64 | Packet loss between the client and the relay (%). Generally inaccurate and higher than true value because client relay pings are sent infrequently. |

## Server Relay Ping

These entries are written to bigquery at the start of each session contining pings between the server and destination relays in the same datacenter.

| Field | Type | Description |
| ------------- | ------------- | ------------- |
| timestamp | TIMESTAMP | The timestamp when the server relay ping occurred |
| buyer_id | INT64 | The buyer this player belongs to |
| session_id | INT64 | Unique id for the session |
| datacenter_id | INT64 | Unique id for the datacenter |
| server_address | STRING | Server address and port number |
| server_relay_id | INT64 | Relay id being pinged by the server |
| server_relay_rtt | INT64 | Round trip time ping between the server and the relay (milliseconds) |
| server_relay_jitter | INT64 | Jitter between the server and the relay (milliseconds) |
| server_relay_packet_loss | FLOAT64 | Packet loss between the server and the relay (%). Generally inaccurate and higher than true value because server relay pings are sent infrequently. |

## Route Matrix update

Updated once per-second with each route matrix updated. The route matrix is the core data structure used for route planning across relays. You can use this data to track the results of route optimization over time.

| Field | Type | Description |
| ------------- | ------------- | ------------- |
| timestamp | TIMESTAMP | The timestamp when the route matrix update occurred |
| cost_matrix_size | INT64 | The size of the cost matrix in bytes |
| route_matrix_size | INT64 | The size of the route matrix in bytes |
| optimize_time | INT64 | Time it took produce this route matrix from the cost matrix (milliseconds) |
| num_relays | INT64 | The number of relays in the route matrix |
| num_active_relays | INT64 | The number of active relays in the relay manager |
| num_dest_relays | INT64 | The number of destination relays in the route matrix |
| num_datacenters | INT64 | The number of datacenters in the route matrix |
| total_routes | INT64 | The total number of routes in the route matrix |
| average_num_routes | FLOAT64 | The average number of routes between any two relays |
| average_route_length | FLOAT64 | The average number of relays per-route |
| no_route_percent | FLOAT64 | The percent of relay pairs that have no route between them |
| one_route_percent | FLOAT64 | The percent of relay pairs with only one route between them |
| no_direct_route_percent | FLOAT64 | The percent of relay pairs with no direct route between them |
| rtt_bucket_no_improvement | FLOAT64 | The percent of relay pairs with no improvement |
| rtt_bucket_0_5ms | FLOAT64 | The percent of relay pairs with 0-5ms reduction in latency |
| rtt_bucket_5_10ms | FLOAT64 | The percent of relay pairs with 5-10ms reduction in latency |
| rtt_bucket_10_15ms | FLOAT64 | The percent of relay pairs with 10-15ms reduction in latency |
| rtt_bucket_15_20ms | FLOAT64 | The percent of relay pairs with 15-20ms reduction in latency |
| rtt_bucket_20_25ms | FLOAT64 | The percent of relay pairs with 20-25ms reduction in latency |
| rtt_bucket_25_30ms | FLOAT64 | The percent of relay pairs with 25-30ms reduction in latency |
| rtt_bucket_30_35ms | FLOAT64 | The percent of relay pairs with 30-35ms reduction in latency |
| rtt_bucket_35_40ms | FLOAT64 | The percent of relay pairs with 35-40ms reduction in latency |
| rtt_bucket_40_45ms | FLOAT64 | The percent of relay pairs with 40-45ms reduction in latency |
| rtt_bucket_45_50ms | FLOAT64 | The percent of relay pairs with 45-50ms reduction in latency |
| rtt_bucket_50ms_plus | FLOAT64 | The percent of relay pairs with 50ms+ reduction in latency |

## Relay to Relay Ping

This data is not uploaded by default because at 1000 relays the number of rows inserted is ~1 million per-second. However, in smaller relay fleets or dev the system can be modified to upload this data, which provides good visibility into intra-relay performance that you can analyze via bigquery data.

| Field | Type | Description |
| ------------- | ------------- | ------------- |
| timestamp | TIMESTAMP | The timestamp when the relay ping occurred |
| source_relay_id | INT64 | The id of the source relay |
| destination_relay_id | INT64 | The id of the destination relay |
| rtt | INT64 | Round trip latency between the two relays (milliseconds) |
| jitter | INT64 | Time variance in latency between the two relays (milliseconds) |
| packet_loss | FLOAT64 | The packet loss between the two relays (%) |

[Back to main documentation](../README.md)
