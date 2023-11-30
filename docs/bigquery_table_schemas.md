<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Bigquery table schemas

Network Next writes data to bigquery by default so you can run data science and analytics queries.
 
For example, once every 10 seconds network performance data such as accelerated and non-accelerated latency (RTT), jitter, packet loss, bandwidth usage and out order packets for sessions are written to bigquery. 

At the end of each session a summary data entry is written, which makes it much faster and cheaper to query data on a per-session basis.

There is also data written each time a client pings near relays at the start of each session, so you can look at direct ping results from clients to nearby relays, and data from each server and relay in your fleet is sent so you can track their performance and uptime.

Schemas for all this data are described below.

## Session Update

Session updates contain network performance data once every 10 seconds for a session. This is the primary network performance data, including everything shown in the portal for a session and much more.

| Field Name | Type | Description |
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

## 2. Session Summary

[Back to main documentation](../README.md)
