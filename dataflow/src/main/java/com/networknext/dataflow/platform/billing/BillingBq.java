package com.networknext.dataflow.platform.billing;

import com.google.api.services.bigquery.model.TableFieldSchema;
import com.google.api.services.bigquery.model.TableSchema;
import com.google.common.collect.ImmutableList;

public class BillingBq {

    public static final String BUYER_ID = "buyerId";
    public static final String SESSION_ID = "sessionId";
    public static final String USER_ID = "userId";
    public static final String PLATFORM_ID = "platformId";
    public static final String DIRECT_RTT = "directRtt";
    public static final String DIRECT_JITTER = "directJitter";
    public static final String DIRECT_PACKET_LOSS = "directPacketLoss";
    public static final String NEXT_RTT = "nextRtt";
    public static final String NEXT_JITTER = "nextJitter";
    public static final String NEXT_PACKET_LOSS = "nextPacketLoss";
    public static final String CLIENT_IP_ADDRESS = "clientIpAddress";
    public static final String SERVER_IP_ADDRESS = "serverIpAddress";
    public static final String SERVER_PRIVATE_IP_ADDRESS = "serverPrivateIpAddress";
    public static final String TAG = "tag";
    public static final String NEAR_RELAYS = "nearRelays";
    public static final String NEAR_RELAYS_ID = "id";
    public static final String NEAR_RELAYS_RTT = "rtt";
    public static final String NEAR_RELAYS_JITTER = "jitter";
    public static final String NEAR_RELAYS_PACKET_LOSS = "packetLoss";
    public static final String ISSUED_NEAR_RELAYS = "issuedNearRelays";
    public static final String ISSUED_NEAR_RELAYS_INDEX = "index";
    public static final String ISSUED_NEAR_RELAYS_ID = "id";
    public static final String ISSUED_NEAR_RELAYS_IP_ADDRESS = "ipAddress";
    public static final String CONNECTION_TYPE = "connectionType";
    public static final String DATACENTER_ID = "datacenterId";
    public static final String SEQUENCE_NUMBER = "sequenceNumber";
    public static final String FALLBACK_TO_DIRECT = "fallbackToDirect";
    public static final String VERSION_MAJOR = "versionMajor";
    public static final String VERSION_MINOR = "versionMinor";
    public static final String VERSION_PATCH = "versionPatch";
    public static final String KBPS_UP = "kbpsUp";
    public static final String KBPS_DOWN = "kbpsDown";
    public static final String COUNTRY_CODE = "countryCode";
    public static final String COUNTRY = "country";
    public static final String REGION = "region";
    public static final String CITY = "city";
    public static final String LATITUDE = "latitude";
    public static final String LONGITUDE = "longitude";
    public static final String ISP = "isp";
    public static final String ROUTE = "route";
    public static final String ROUTE_ID = "id";
    public static final String ROUTE_SELLER_ID = "sellerId";
    public static final String ROUTE_PRICE_INGRESS = "priceIngress";
    public static final String ROUTE_PRICE_EGRESS = "priceEgress";
    public static final String ROUTE_DECISION = "routeDecision";
    public static final String DURATION = "duration";
    public static final String BYTES_UP = "bytesUp";
    public static final String BYTES_DOWN = "bytesDown";
    public static final String TIMESTAMP = "timestamp";
    public static final String TIMESTAMP_START = "timestampStart";
    public static final String PREDICTED_RTT = "predictedRtt";
    public static final String PREDICTED_JITTER = "predictedJitter";
    public static final String PREDICTED_PACKET_LOSS = "predictedPacketLoss";
    public static final String ROUTE_CHANGED = "routeChanged";
    public static final String NETWORK_NEXT = "networkNext";
    public static final String INITIAL = "initial";
    public static final String FLAGGED = "flagged";
    public static final String TRY_BEFORE_YOU_BUY = "tryBeforeYouBuy";
    public static final String PACKETS_LOST_CLIENT_TO_SERVER = "packetsLostClientToServer";
    public static final String PACKETS_LOST_SERVER_TO_CLIENT = "packetsLostServerToClient";
    public static final String CONSIDERED_ROUTES = "consideredRoutes";
    public static final String CONSIDERED_ROUTES_ROUTE = "route";
    public static final String CONSIDERED_ROUTES_ROUTE_ID = "id";
    public static final String CONSIDERED_ROUTES_ROUTE_SELLER_ID = "sellerId";
    public static final String CONSIDERED_ROUTES_ROUTE_PRICE_INGRESS = "priceIngress";
    public static final String CONSIDERED_ROUTES_ROUTE_PRICE_EGRESS = "priceEgress";
    public static final String ACCEPTABLE_ROUTES = "acceptableRoutes";
    public static final String ACCEPTABLE_ROUTES_ROUTE = "route";
    public static final String ACCEPTABLE_ROUTES_ROUTE_ID = "id";
    public static final String ACCEPTABLE_ROUTES_ROUTE_SELLER_ID = "sellerId";
    public static final String ACCEPTABLE_ROUTES_ROUTE_PRICE_INGRESS = "priceIngress";
    public static final String ACCEPTABLE_ROUTES_ROUTE_PRICE_EGRESS = "priceEgress";
    public static final String SAME_ROUTE = "sameRoute";

    public static final String TABLE_DESCRIPTION =
            "Network Next works by selling 10 second \"slices\" of network bandwidth. This table contains all the slices issued by Network Next, whether they route over Network Next or the public Internet.\n\nEach record in this table contains information about the slice that is being issued to the client, as well as information about the previous slice (which Network Next uses to make routing decisions).";
    public static final String BUYER_ID_DESCRIPTION =
            "The ID of the buyer. This represents who is buying the traffic, e.g. Psyonix.";
    public static final String SESSION_ID_DESCRIPTION =
            "The ID of the session. Each session is a connection to a game server, and there are multiple 10 second slices per session (this table contains slices). If you want statistics per session, you need to group by (sessionId, timestampStart).";
    public static final String USER_ID_DESCRIPTION =
            "A hash of the true user ID, which is often PS4/XB1/Steam user ID. The user ID is hashed before it ever gets to Network Next. If you want statistics per player, you need to group by (buyerId, platformId, userId).";
    public static final String PLATFORM_ID_DESCRIPTION =
            "The platform ID; represents whether the player is on Windows (1), Mac (2), Linux (3), Nintendo Switch (4), PS4 (5), iOS (6) or Xbox One (7).";
    public static final String DIRECT_RTT_DESCRIPTION =
            "The RTT of the player's connection over the public Internet, during the previous slice.";
    public static final String DIRECT_JITTER_DESCRIPTION =
            "The jitter of the player's connection over the public Internet, during the previous slice.";
    public static final String DIRECT_PACKET_LOSS_DESCRIPTION =
            "The packet loss of the player's connection over the public Internet, during the previous slice.";
    public static final String NEXT_RTT_DESCRIPTION =
            "If the session was on Network Next during the previous slice, this is the RTT over Network Next as measured by the client.";
    public static final String NEXT_JITTER_DESCRIPTION =
            "If the session was on Network Next during the previous slice, this is the jitter over Network Next as measured by the client.";
    public static final String NEXT_PACKET_LOSS_DESCRIPTION =
            "If the session was on Network Next during the previous slice, this is the packet loss over Network Next as measured by the client.";
    public static final String CLIENT_IP_ADDRESS_DESCRIPTION =
            "The anonymized IP address of the client (player). In older data, this field may not be anonymized.";
    public static final String SERVER_IP_ADDRESS_DESCRIPTION = "The IP address of the game server.";
    public static final String SERVER_PRIVATE_IP_ADDRESS_DESCRIPTION =
            "The IP address of the game server on LAN (if specified).";
    public static final String TAG_DESCRIPTION =
            "The route shader tag used. This tells Network Next what routing profile to use.";
    public static final String NEAR_RELAYS_DESCRIPTION =
            "A list of near relays issued in the previous slice, and their performance over the previous 10 seconds.";
    public static final String NEAR_RELAYS_ID_DESCRIPTION = "The ID of the relay.";
    public static final String NEAR_RELAYS_RTT_DESCRIPTION =
            "The RTT from the client to the nearby relay.";
    public static final String NEAR_RELAYS_JITTER_DESCRIPTION =
            "The jitter from the client to the nearby relay.";
    public static final String NEAR_RELAYS_PACKET_LOSS_DESCRIPTION =
            "The packet loss from the client to the nearby relay.";
    public static final String ISSUED_NEAR_RELAYS_DESCRIPTION =
            "A list of near relays that were issued for this slice (the next 10 seconds).";
    public static final String ISSUED_NEAR_RELAYS_INDEX_DESCRIPTION =
            "The numeric index of the near relay in the list sent to the client.";
    public static final String ISSUED_NEAR_RELAYS_ID_DESCRIPTION = "The ID of the relay.";
    public static final String ISSUED_NEAR_RELAYS_IP_ADDRESS_DESCRIPTION =
            "The IP address of the relay.";
    public static final String CONNECTION_TYPE_DESCRIPTION =
            "The type of network connection the player has. Either unknown (0), wired Ethernet (1), wireless Wi-Fi (2) or cellular 4G (3).";
    public static final String DATACENTER_ID_DESCRIPTION =
            "The ID of the datacenter the game server is located in.";
    public static final String SEQUENCE_NUMBER_DESCRIPTION =
            "The sequence number of the route request packet. This is an internal value.";
    public static final String FALLBACK_TO_DIRECT_DESCRIPTION =
            "If true, the client fell back to using a direct connection over the public Internet, and will not route over Network Next again.";
    public static final String VERSION_MAJOR_DESCRIPTION = "The major version number of the SDK.";
    public static final String VERSION_MINOR_DESCRIPTION = "The minor version number of the SDK.";
    public static final String VERSION_PATCH_DESCRIPTION = "The patch version number of the SDK.";
    public static final String KBPS_UP_DESCRIPTION =
            "The average KBPS used by the client in sending data to the game server, for the previous slice. This is informational only, and is not what Network Next bills on.";
    public static final String KBPS_DOWN_DESCRIPTION =
            "The average KBPS used by the client in receiving data from the game server, for the previous slice. This is informational only, and is not what Network Next bills on.";
    public static final String COUNTRY_CODE_DESCRIPTION =
            "The country code of the location of the player (guessed from the original client IP address).";
    public static final String COUNTRY_DESCRIPTION =
            "The country that the player is located in (guessed from the original client IP address).";
    public static final String REGION_DESCRIPTION =
            "The region that the player is located in (guessed from the original client IP address).";
    public static final String CITY_DESCRIPTION =
            "The city that the player is located in (guessed from the original client IP address).";
    public static final String LATITUDE_DESCRIPTION =
            "The latitude of the player's location (guessed from the original client IP address).";
    public static final String LONGITUDE_DESCRIPTION =
            "The longitude of the player's location (guessed from the original client IP address).";
    public static final String ISP_DESCRIPTION =
            "The ISP of the player (based on IP block allocation).";
    public static final String ROUTE_DESCRIPTION =
            "The route over Network Next that traffic will take for this slice. If this list is empty, the session will go over the public Internet.";
    public static final String ROUTE_ID_DESCRIPTION = "The ID of the relay.";
    public static final String ROUTE_SELLER_ID_DESCRIPTION =
            "The ID of the seller that owns the relay.";
    public static final String ROUTE_PRICE_INGRESS_DESCRIPTION =
            "The amount billed for ingress traffic (in billionths of a cent).";
    public static final String ROUTE_PRICE_EGRESS_DESCRIPTION =
            "The amount billed for egress traffic (in billionths of a cent).";
    public static final String ROUTE_DECISION_DESCRIPTION =
            "The route decision flags representing why Network Next made the routing decision it did.";
    public static final String DURATION_DESCRIPTION = "The duration, in seconds, of this slice.";
    public static final String BYTES_UP_DESCRIPTION =
            "The amount of bandwidth up (client to server) reserved by this slice. This is the amount of bandwidth (plus bytesDown) that is billed on.";
    public static final String BYTES_DOWN_DESCRIPTION =
            "The amount of bandwidth down (server to client) reserved by this slice. This is the amount of bandwidth (plus bytesUp) that is billed on.";
    public static final String TIMESTAMP_DESCRIPTION =
            "The UTC timestamp for this slice in the session.";
    public static final String TIMESTAMP_START_DESCRIPTION =
            "The UTC timestamp at which the session started.";
    public static final String PREDICTED_RTT_DESCRIPTION =
            "The RTT that we predict we'll get on Network Next for this slice (the next 10 seconds).";
    public static final String PREDICTED_JITTER_DESCRIPTION =
            "The jitter that we predict we'll get on Network Next for this slice (the next 10 seconds).";
    public static final String PREDICTED_PACKET_LOSS_DESCRIPTION =
            "The packet loss that we predict we'll get on Network Next for this slice (the next 10 seconds).";
    public static final String ROUTE_CHANGED_DESCRIPTION =
            "If true, the route has changed since the previous slice.";
    public static final String NETWORK_NEXT_DESCRIPTION =
            "If true, this slice will be on Network Next.";
    public static final String INITIAL_DESCRIPTION =
            "If true, this is the first slice in a session.";
    public static final String FLAGGED_DESCRIPTION =
            "If true, the player reported this session. If this is the first slice in a session with this set to true, then the player reported the session during the previous slice.";
    public static final String TRY_BEFORE_YOU_BUY_DESCRIPTION =
            "If true, the previous slice was 'try-before-you-buy'.";
    public static final String PACKETS_LOST_CLIENT_TO_SERVER_DESCRIPTION =
            "The total number of packets lost from the client to the server up to the previous slice.";
    public static final String PACKETS_LOST_SERVER_TO_CLIENT_DESCRIPTION =
            "The total number of packets lost from the server to the client up to the previous slice.";
    public static final String CONSIDERED_ROUTES_DESCRIPTION =
            "The full list of routes considered by the router when making route decisions. This information is not present by default due the large volume of data it generates.";
    public static final String CONSIDERED_ROUTES_ROUTE_DESCRIPTION =
            "A route that was considered by the router.";
    public static final String CONSIDERED_ROUTES_ROUTE_ID_DESCRIPTION = "The ID of the relay.";
    public static final String CONSIDERED_ROUTES_ROUTE_SELLER_ID_DESCRIPTION =
            "The ID of the seller that owns the relay.";
    public static final String CONSIDERED_ROUTES_ROUTE_PRICE_INGRESS_DESCRIPTION =
            "The ingress amount we would have billed if this route was chosen, in billionths of a cent.";
    public static final String CONSIDERED_ROUTES_ROUTE_PRICE_EGRESS_DESCRIPTION =
            "The egress amount we would have billed if this route was chosen, in billionths of a cent.";
    public static final String ACCEPTABLE_ROUTES_DESCRIPTION =
            "Of the full list of routes considered, these are the routes that were deemed acceptable by the router. This information is not present by default due the large volume of data it generates.";
    public static final String ACCEPTABLE_ROUTES_ROUTE_DESCRIPTION =
            "A route that was deemed acceptable by the router.";
    public static final String ACCEPTABLE_ROUTES_ROUTE_ID_DESCRIPTION = "The ID of the relay.";
    public static final String ACCEPTABLE_ROUTES_ROUTE_SELLER_ID_DESCRIPTION =
            "The ID of the seller that owns the relay.";
    public static final String ACCEPTABLE_ROUTES_ROUTE_PRICE_INGRESS_DESCRIPTION =
            "The ingress amount we would have billed if this route was chosen, in billionths of a cent.";
    public static final String ACCEPTABLE_ROUTES_ROUTE_PRICE_EGRESS_DESCRIPTION =
            "The egress amount we would have billed if this route was chosen, in billionths of a cent.";
    public static final String SAME_ROUTE_DESCRIPTION =
            "If true, the route issued for this slice is the same route as the previous slice.";

    public static TableSchema schema;

    static {
        schema = new TableSchema().setFields(ImmutableList.of(
                new TableFieldSchema().setName(BUYER_ID).setDescription(BUYER_ID_DESCRIPTION)
                        .setType("STRING").setMode("REQUIRED"),
                new TableFieldSchema().setName(SESSION_ID).setDescription(SESSION_ID_DESCRIPTION)
                        .setType("INT64").setMode("REQUIRED"),
                new TableFieldSchema().setName(USER_ID).setDescription(USER_ID_DESCRIPTION)
                        .setType("STRING").setMode("REQUIRED"),
                new TableFieldSchema().setName(PLATFORM_ID).setDescription(PLATFORM_ID_DESCRIPTION)
                        .setType("INT64").setMode("REQUIRED"),
                new TableFieldSchema().setName(DIRECT_RTT).setDescription(DIRECT_RTT_DESCRIPTION)
                        .setType("FLOAT64").setMode("REQUIRED"),
                new TableFieldSchema().setName(DIRECT_JITTER)
                        .setDescription(DIRECT_JITTER_DESCRIPTION).setType("FLOAT64")
                        .setMode("REQUIRED"),
                new TableFieldSchema().setName(DIRECT_PACKET_LOSS)
                        .setDescription(DIRECT_PACKET_LOSS_DESCRIPTION).setType("FLOAT64")
                        .setMode("REQUIRED"),
                new TableFieldSchema().setName(NEXT_RTT).setDescription(NEXT_RTT_DESCRIPTION)
                        .setType("FLOAT64").setMode("REQUIRED"),
                new TableFieldSchema().setName(NEXT_JITTER).setDescription(NEXT_JITTER_DESCRIPTION)
                        .setType("FLOAT64").setMode("REQUIRED"),
                new TableFieldSchema().setName(NEXT_PACKET_LOSS)
                        .setDescription(NEXT_PACKET_LOSS_DESCRIPTION).setType("FLOAT64")
                        .setMode("REQUIRED"),
                new TableFieldSchema().setName(CLIENT_IP_ADDRESS)
                        .setDescription(CLIENT_IP_ADDRESS_DESCRIPTION).setType("STRING")
                        .setMode("REQUIRED"),
                new TableFieldSchema().setName(SERVER_IP_ADDRESS)
                        .setDescription(SERVER_IP_ADDRESS_DESCRIPTION).setType("STRING")
                        .setMode("REQUIRED"),
                new TableFieldSchema().setName(SERVER_PRIVATE_IP_ADDRESS)
                        .setDescription(SERVER_PRIVATE_IP_ADDRESS_DESCRIPTION).setType("STRING")
                        .setMode("NULLABLE"),
                new TableFieldSchema().setName(TAG).setDescription(TAG_DESCRIPTION).setType("INT64")
                        .setMode("REQUIRED"),
                new TableFieldSchema().setName(NEAR_RELAYS).setDescription(NEAR_RELAYS_DESCRIPTION)
                        .setType("RECORD").setMode("REPEATED")
                        .setFields(ImmutableList.of(
                                new TableFieldSchema().setName(NEAR_RELAYS_ID)
                                        .setDescription(NEAR_RELAYS_ID_DESCRIPTION)
                                        .setType("STRING").setMode("REQUIRED"),
                                new TableFieldSchema().setName(NEAR_RELAYS_RTT)
                                        .setDescription(NEAR_RELAYS_RTT_DESCRIPTION)
                                        .setType("FLOAT64").setMode("REQUIRED"),
                                new TableFieldSchema().setName(NEAR_RELAYS_JITTER)
                                        .setDescription(NEAR_RELAYS_JITTER_DESCRIPTION)
                                        .setType("FLOAT64").setMode("REQUIRED"),
                                new TableFieldSchema().setName(NEAR_RELAYS_PACKET_LOSS)
                                        .setDescription(NEAR_RELAYS_PACKET_LOSS_DESCRIPTION)
                                        .setType("FLOAT64").setMode("REQUIRED"))),
                new TableFieldSchema().setName(ISSUED_NEAR_RELAYS)
                        .setDescription(ISSUED_NEAR_RELAYS_DESCRIPTION).setType("RECORD")
                        .setMode("REPEATED").setFields(
                                ImmutableList.of(
                                        new TableFieldSchema().setName(ISSUED_NEAR_RELAYS_INDEX)
                                                .setDescription(
                                                        ISSUED_NEAR_RELAYS_INDEX_DESCRIPTION)
                                                .setType("INTEGER").setMode("NULLABLE"),
                                        new TableFieldSchema()
                                                .setName(ISSUED_NEAR_RELAYS_ID)
                                                .setDescription(ISSUED_NEAR_RELAYS_ID_DESCRIPTION)
                                                .setType("STRING").setMode("NULLABLE"),
                                        new TableFieldSchema()
                                                .setName(ISSUED_NEAR_RELAYS_IP_ADDRESS)
                                                .setDescription(
                                                        ISSUED_NEAR_RELAYS_IP_ADDRESS_DESCRIPTION)
                                                .setType("STRING").setMode("NULLABLE"))),
                new TableFieldSchema().setName(CONNECTION_TYPE)
                        .setDescription(CONNECTION_TYPE_DESCRIPTION).setType("INT64")
                        .setMode("REQUIRED"),
                new TableFieldSchema().setName(DATACENTER_ID)
                        .setDescription(DATACENTER_ID_DESCRIPTION).setType("STRING")
                        .setMode("REQUIRED"),
                new TableFieldSchema().setName(SEQUENCE_NUMBER)
                        .setDescription(SEQUENCE_NUMBER_DESCRIPTION).setType("INT64")
                        .setMode("REQUIRED"),
                new TableFieldSchema().setName(FALLBACK_TO_DIRECT)
                        .setDescription(FALLBACK_TO_DIRECT_DESCRIPTION).setType("BOOL")
                        .setMode("REQUIRED"),
                new TableFieldSchema().setName(VERSION_MAJOR)
                        .setDescription(VERSION_MAJOR_DESCRIPTION).setType("INT64")
                        .setMode("REQUIRED"),
                new TableFieldSchema().setName(VERSION_MINOR)
                        .setDescription(VERSION_MINOR_DESCRIPTION).setType("INT64")
                        .setMode("REQUIRED"),
                new TableFieldSchema().setName(VERSION_PATCH)
                        .setDescription(VERSION_PATCH_DESCRIPTION).setType("INT64")
                        .setMode("REQUIRED"),
                new TableFieldSchema().setName(KBPS_UP).setDescription(KBPS_UP_DESCRIPTION)
                        .setType("INT64").setMode("REQUIRED"),
                new TableFieldSchema().setName(KBPS_DOWN).setDescription(KBPS_DOWN_DESCRIPTION)
                        .setType("INT64").setMode("REQUIRED"),
                new TableFieldSchema().setName(COUNTRY_CODE)
                        .setDescription(COUNTRY_CODE_DESCRIPTION).setType("STRING")
                        .setMode("NULLABLE"),
                new TableFieldSchema().setName(COUNTRY).setDescription(COUNTRY_DESCRIPTION)
                        .setType("STRING").setMode("NULLABLE"),
                new TableFieldSchema().setName(REGION).setDescription(REGION_DESCRIPTION)
                        .setType("STRING").setMode("NULLABLE"),
                new TableFieldSchema().setName(CITY).setDescription(CITY_DESCRIPTION)
                        .setType("STRING").setMode("NULLABLE"),
                new TableFieldSchema().setName(LATITUDE).setDescription(LATITUDE_DESCRIPTION)
                        .setType("FLOAT64").setMode("NULLABLE"),
                new TableFieldSchema().setName(LONGITUDE).setDescription(LONGITUDE_DESCRIPTION)
                        .setType("FLOAT64").setMode("NULLABLE"),
                new TableFieldSchema().setName(ISP).setDescription(ISP_DESCRIPTION)
                        .setType("STRING").setMode("NULLABLE"),
                new TableFieldSchema().setName(ROUTE).setDescription(ROUTE_DESCRIPTION)
                        .setType("RECORD").setMode("REPEATED")
                        .setFields(ImmutableList.of(
                                new TableFieldSchema().setName(ROUTE_ID)
                                        .setDescription(ROUTE_ID_DESCRIPTION).setType("STRING")
                                        .setMode("REQUIRED"),
                                new TableFieldSchema().setName(ROUTE_SELLER_ID)
                                        .setDescription(ROUTE_SELLER_ID_DESCRIPTION)
                                        .setType("STRING").setMode("REQUIRED"),
                                new TableFieldSchema().setName(ROUTE_PRICE_INGRESS)
                                        .setDescription(ROUTE_PRICE_INGRESS_DESCRIPTION)
                                        .setType("INT64").setMode("REQUIRED"),
                                new TableFieldSchema().setName(ROUTE_PRICE_EGRESS)
                                        .setDescription(ROUTE_PRICE_EGRESS_DESCRIPTION)
                                        .setType("INT64").setMode("REQUIRED"))),
                new TableFieldSchema().setName(ROUTE_DECISION)
                        .setDescription(ROUTE_DECISION_DESCRIPTION).setType("INT64")
                        .setMode("REQUIRED"),
                new TableFieldSchema().setName(DURATION).setDescription(DURATION_DESCRIPTION)
                        .setType("INT64").setMode("REQUIRED"),
                new TableFieldSchema().setName(BYTES_UP).setDescription(BYTES_UP_DESCRIPTION)
                        .setType("INT64").setMode("REQUIRED"),
                new TableFieldSchema().setName(BYTES_DOWN).setDescription(BYTES_DOWN_DESCRIPTION)
                        .setType("INT64").setMode("REQUIRED"),
                new TableFieldSchema().setName(TIMESTAMP).setDescription(TIMESTAMP_DESCRIPTION)
                        .setType("TIMESTAMP").setMode("REQUIRED"),
                new TableFieldSchema().setName(TIMESTAMP_START)
                        .setDescription(TIMESTAMP_START_DESCRIPTION).setType("TIMESTAMP")
                        .setMode("REQUIRED"),
                new TableFieldSchema().setName(PREDICTED_RTT)
                        .setDescription(PREDICTED_RTT_DESCRIPTION).setType("FLOAT64")
                        .setMode("REQUIRED"),
                new TableFieldSchema().setName(PREDICTED_JITTER)
                        .setDescription(PREDICTED_JITTER_DESCRIPTION).setType("FLOAT64")
                        .setMode("REQUIRED"),
                new TableFieldSchema().setName(PREDICTED_PACKET_LOSS)
                        .setDescription(PREDICTED_PACKET_LOSS_DESCRIPTION).setType("FLOAT64")
                        .setMode("REQUIRED"),
                new TableFieldSchema().setName(ROUTE_CHANGED)
                        .setDescription(ROUTE_CHANGED_DESCRIPTION).setType("BOOL")
                        .setMode("REQUIRED"),
                new TableFieldSchema().setName(NETWORK_NEXT)
                        .setDescription(NETWORK_NEXT_DESCRIPTION).setType("BOOL")
                        .setMode("REQUIRED"),
                new TableFieldSchema().setName(INITIAL).setDescription(INITIAL_DESCRIPTION)
                        .setType("BOOL").setMode("REQUIRED"),
                new TableFieldSchema().setName(FLAGGED).setDescription(FLAGGED_DESCRIPTION)
                        .setType("BOOL").setMode("REQUIRED"),
                new TableFieldSchema().setName(TRY_BEFORE_YOU_BUY)
                        .setDescription(TRY_BEFORE_YOU_BUY_DESCRIPTION).setType("BOOL")
                        .setMode("REQUIRED"),
                new TableFieldSchema().setName(PACKETS_LOST_CLIENT_TO_SERVER)
                        .setDescription(PACKETS_LOST_CLIENT_TO_SERVER_DESCRIPTION).setType("INT64")
                        .setMode("REQUIRED"),
                new TableFieldSchema().setName(PACKETS_LOST_SERVER_TO_CLIENT)
                        .setDescription(PACKETS_LOST_SERVER_TO_CLIENT_DESCRIPTION).setType("INT64")
                        .setMode("REQUIRED"),
                new TableFieldSchema().setName(CONSIDERED_ROUTES)
                        .setDescription(CONSIDERED_ROUTES_DESCRIPTION).setType("RECORD")
                        .setMode("REPEATED")
                        .setFields(ImmutableList.of(new TableFieldSchema()
                                .setName(CONSIDERED_ROUTES_ROUTE)
                                .setDescription(CONSIDERED_ROUTES_ROUTE_DESCRIPTION)
                                .setType("RECORD").setMode("REPEATED")
                                .setFields(ImmutableList.of(
                                        new TableFieldSchema().setName(CONSIDERED_ROUTES_ROUTE_ID)
                                                .setDescription(
                                                        CONSIDERED_ROUTES_ROUTE_ID_DESCRIPTION)
                                                .setType("STRING").setMode("NULLABLE"),
                                        new TableFieldSchema()
                                                .setName(CONSIDERED_ROUTES_ROUTE_SELLER_ID)
                                                .setDescription(
                                                        CONSIDERED_ROUTES_ROUTE_SELLER_ID_DESCRIPTION)
                                                .setType("STRING").setMode("NULLABLE"),
                                        new TableFieldSchema()
                                                .setName(CONSIDERED_ROUTES_ROUTE_PRICE_INGRESS)
                                                .setDescription(
                                                        CONSIDERED_ROUTES_ROUTE_PRICE_INGRESS_DESCRIPTION)
                                                .setType("INT64").setMode("NULLABLE"),
                                        new TableFieldSchema()
                                                .setName(CONSIDERED_ROUTES_ROUTE_PRICE_EGRESS)
                                                .setDescription(
                                                        CONSIDERED_ROUTES_ROUTE_PRICE_EGRESS_DESCRIPTION)
                                                .setType("INT64").setMode("NULLABLE"))))),
                new TableFieldSchema().setName(ACCEPTABLE_ROUTES)
                        .setDescription(ACCEPTABLE_ROUTES_DESCRIPTION).setType("RECORD")
                        .setMode("REPEATED")
                        .setFields(ImmutableList.of(new TableFieldSchema()
                                .setName(ACCEPTABLE_ROUTES_ROUTE)
                                .setDescription(ACCEPTABLE_ROUTES_ROUTE_DESCRIPTION)
                                .setType("RECORD").setMode("REPEATED")
                                .setFields(ImmutableList.of(
                                        new TableFieldSchema().setName(ACCEPTABLE_ROUTES_ROUTE_ID)
                                                .setDescription(
                                                        ACCEPTABLE_ROUTES_ROUTE_ID_DESCRIPTION)
                                                .setType("STRING").setMode("NULLABLE"),
                                        new TableFieldSchema()
                                                .setName(ACCEPTABLE_ROUTES_ROUTE_SELLER_ID)
                                                .setDescription(
                                                        ACCEPTABLE_ROUTES_ROUTE_SELLER_ID_DESCRIPTION)
                                                .setType("STRING").setMode("NULLABLE"),
                                        new TableFieldSchema()
                                                .setName(ACCEPTABLE_ROUTES_ROUTE_PRICE_INGRESS)
                                                .setDescription(
                                                        ACCEPTABLE_ROUTES_ROUTE_PRICE_INGRESS_DESCRIPTION)
                                                .setType("INT64").setMode("NULLABLE"),
                                        new TableFieldSchema()
                                                .setName(ACCEPTABLE_ROUTES_ROUTE_PRICE_EGRESS)
                                                .setDescription(
                                                        ACCEPTABLE_ROUTES_ROUTE_PRICE_EGRESS_DESCRIPTION)
                                                .setType("INT64").setMode("NULLABLE"))))),
                new TableFieldSchema().setName(SAME_ROUTE).setDescription(SAME_ROUTE_DESCRIPTION)
                        .setType("BOOL").setMode("NULLABLE")));
    }

}
