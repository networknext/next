<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Bigquery table schemas

Network Next writes data to bigquery by default so you can run data science and analytics queries.
 
For example, once every 10 seconds network performance data such as accelerated and non-accelerated latency (RTT), jitter, packet loss, bandwidth usage and out order packets for sessions are written to bigquery. 

At the end of each session a summary data entry is written, which makes it much faster and cheaper to query data on a per-session basis.

There is also data written each time a client pings near relays at the start of each session, so you can look at direct ping results from clients to nearby relays, and data from each server and relay in your fleet is sent so you can track their performance and uptime.

Data in bigquery is retained for 90 days by default to comply with GDPR.

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
| direct_kbps_up | INT64 | Bandwidth in the server to client direction along the direct path (unaccelerated). Kilobits per-second |
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
| client_next_bandwidth_over_limit | BOOL | client_next_bandwidth_over_limitTrue if the client to server next bandwidth went over the envelope limit this slice and was sent over direct. |
| server_next_bandwidth_over_limit | BOOL | True if the client to server next bandwidth went over the envelope limit this slice and was sent over direct. |
| veto | BOOL | True if the routing logic decided that this session should no longer be accelerated for some reason. |
| disabled | BOOL | True if the buyer is disabled. Disabled buyers don't perform any acceleration or analytics on network next. |
| not_selected | BOOL | If the route shader selection % is any value other than 100%, then this is true for sessions that were not selected for acceleration. |
| a | BOOL | This session was part of an AB test, and is in the A group. (potentially accelerated) |
| b | BOOL | This session was part of an AB test, and is in the A group. (never accelerated) |
| latency_worse | BOOL | True if we made significantly latency worse. In this case the session is told to stop acceleration immediately. |
| location_veto | BOOL | True if we could not locate the player using ip2location, eg. lat long is at null island (0,0). |
| mispredict | BOOL | True if we significantly mispredicted the latency reduction we could provide for this session. |
| lack_of_diversity | BOOL | True if route diversity is set in the route shader, and we don't have enough route diversity to accelerate this session. |

## 2. Session Summary

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
| client_address | STRING | Client address and port number |
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
| server_to_client_packets_out_of_order | INT64 | The total number of game packets received out of order from server to client in this session |

  {
    "name": "total_next_envelope_bytes_up",
    "type": "INT64",
    "mode": "REQUIRED",
    "description": "The total number of envelope bytes sent across network next in the client to server direction for this session"
  },
  {
    "name": "total_next_envelope_bytes_down",
    "type": "INT64",
    "mode": "REQUIRED",
    "description": "The total number of envelope bytes sent across netwnork next in the server to client direction for this session"
  },
  {
    "name": "duration_on_next",
    "type": "INT64",
    "mode": "REQUIRED",
    "description": "Total time spent on network next in this session (time accelerated). Seconds"
  },
  {
    "name": "session_duration",
    "type": "INT64",
    "mode": "REQUIRED",
    "description": "Length of this session in seconds"
  },
  {
    "name": "start_timestamp",
    "type": "TIMESTAMP",
    "mode": "REQUIRED",
    "description": "The time when this session started"
  },
  {
    "name": "error",
    "type": "INT64",
    "mode": "REQUIRED",
    "description": "Error flags to diagnose what's happening with a session. Look up SessionError_* in the codebase for a list of errors. 0 if no error has occurred."
  },
  {
    "name": "reported",
    "type": "BOOL",
    "mode": "REQUIRED",
    "description": "True if this session was reported by the player"
  },
  {
    "name": "latency_reduction",
    "type": "BOOL",
    "mode": "REQUIRED",
    "description": "True if this session took network next to reduce latency"
  },
  {
    "name": "packet_loss_reduction",
    "type": "BOOL",
    "mode": "REQUIRED",
    "description": "True if this session took network next to reduce packet loss"
  },
  {
    "name": "force_next",
    "type": "BOOL",
    "mode": "REQUIRED",
    "description": "True if this session took network next because it was forced to"
  },
  {
    "name": "long_session_update",
    "type": "BOOL",
    "mode": "REQUIRED",
    "description": "True if the processing for any slices in this session took a long time. This may indicate that the server backend is overloaded."
  },
  {
    "name": "client_next_bandwidth_over_limit",
    "type": "BOOL",
    "mode": "REQUIRED",
    "description": "True if the client to server next bandwidth went over the envelope limit at some point and was sent over direct."
  },
  {
    "name": "server_next_bandwidth_over_limit",
    "type": "BOOL",
    "mode": "REQUIRED",
    "description": "True if the server to client next bandwidth went over the envelope limit at some point and was sent over direct."
  },
  {
    "name": "veto",
    "type": "BOOL",
    "mode": "REQUIRED",
    "description": "True if the routing logic decided that this session should no longer be accelerated for some reason."
  },
  {
    "name": "disabled",
    "type": "BOOL",
    "mode": "REQUIRED",
    "description": "True if the buyer is disabled. Disabled buyers don't perform any acceleration or analytics on network next."
  },
  {
    "name": "not_selected",
    "type": "BOOL",
    "mode": "REQUIRED",
    "description": "If the route shader selection % is any value other than 100%, then this is true for sessions that were not selected for acceleration."
  },
  {
    "name": "a",
    "type": "BOOL",
    "mode": "REQUIRED",
    "description": "This session was part of an AB test, and is in the A group."
  },
  {
    "name": "b",
    "type": "BOOL",
    "mode": "REQUIRED",
    "description": "This session was part of an AB test, and is in the B group."
  },
  {
    "name": "latency_worse",
    "type": "BOOL",
    "mode": "REQUIRED",
    "description": "True if we made latency worse."
  },
  {
    "name": "location_veto",
    "type": "BOOL",
    "mode": "REQUIRED",
    "description": "True if we could not locate the player, eg. lat long is at null island (0,0)."
  },
  {
    "name": "mispredict",
    "type": "BOOL",
    "mode": "REQUIRED",
    "description": "True if we significantly mispredicted the latency reduction we could provide for this session."
  },
  {
    "name": "lack_of_diversity",
    "type": "BOOL",
    "mode": "REQUIRED",
    "description": "True if route diversity is set in the route shader, and we don't have enough route diversity to accelerate this session."
  }



[Back to main documentation](../README.md)
