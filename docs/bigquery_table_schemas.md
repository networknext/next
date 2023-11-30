<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Bigquery table schemas

Network Next writes data to bigquery by default so you can run data science and analytics queries.
 
For example, once every 10 seconds network performance data such as accelerated and non-accelerated latency (RTT), jitter, packet loss, bandwidth usage and out order packets for sessions are written to bigquery. 

At the end of each session a summary data entry is written, which makes it faster generally to query data on a per-session bases, than on a per-"slice" basis (a 10 second period of time within a session).

There is also data written each time a client pings near relays at the start of each session, so you can look at direct ping results from clients to nearby relays, and data from each relay in your fleet is sent so you can track its performance and uptime.

Schemas for all of this data are described below.

## 1. Session Update

...

## 2. Session Summary

[Back to main documentation](../README.md)
