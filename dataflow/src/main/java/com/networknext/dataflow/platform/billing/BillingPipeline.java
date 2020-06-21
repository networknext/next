package com.networknext.dataflow.platform.billing;

import java.net.InetAddress;
import java.net.InetSocketAddress;
import java.net.UnknownHostException;
import java.nio.charset.StandardCharsets;
import java.util.ArrayList;

import com.google.api.services.bigquery.model.TableRow;
import com.google.api.services.bigquery.model.TimePartitioning;
import com.google.cloud.bigquery.storage.v1beta1.BigQueryStorageSettings;
import com.networknext.api.address.AddressOuterClass.Address;
import com.networknext.api.ip2location.Ip2Location;
import com.networknext.api.near_data.NearData;
import com.networknext.api.near_data.NearData.NearRelay;
import com.networknext.api.router.Router.BillingEntry;
import com.networknext.api.router.Router.BillingRoute;
import com.networknext.api.router.Router.BillingRouteHop;
import com.networknext.api.router.Router.RouteRequest;
import com.networknext.dataflow.platform.PlatformContext;
import com.networknext.dataflow.util.EntityIdHelpers;
import com.networknext.dataflow.util.Utils;
import com.networknext.dataflow.util.bigquery.BigQuerySchemaUpdater;

import org.apache.beam.sdk.io.gcp.bigquery.BigQueryIO;
import org.apache.beam.sdk.io.gcp.bigquery.InsertRetryPolicy;
import org.apache.beam.sdk.transforms.DoFn;
import org.apache.beam.sdk.transforms.ParDo;
import org.apache.beam.sdk.transforms.SerializableFunction;
import org.apache.hadoop.hbase.client.Mutation;
import org.apache.hadoop.hbase.client.Put;

public class BillingPipeline {

    public static void buildBillingPipeline(PlatformContext context) throws Exception {

        // export to BigQuery

        BigQuerySchemaUpdater.execute(context.bigQueryBilling, BillingBq.schema);

        context.pubSubBilling.apply("Export to BigQuery", BigQueryIO.<BillingEntry>write()
                .withFormatFunction(new SerializableFunction<BillingEntry, TableRow>() {
                    private static final long serialVersionUID = -4919787039912560456L;

                    @Override
                    public TableRow apply(BillingEntry entry) {
                        RouteRequest request = entry.getRequest();
                        Ip2Location.Location ip2location = request.getLocation();
                        ArrayList<TableRow> nearRelays =
                                new ArrayList<TableRow>(request.getNearRelaysCount());
                        for (NearRelay relay : request.getNearRelaysList()) {
                            nearRelays.add(new TableRow()
                                    .set(BillingBq.NEAR_RELAYS_ID,
                                            EntityIdHelpers.getStringForNetworkNextStorage(
                                                    relay.getRelayId()))
                                    .set(BillingBq.NEAR_RELAYS_RTT, relay.getRtt())
                                    .set(BillingBq.NEAR_RELAYS_JITTER, relay.getJitter())
                                    .set(BillingBq.NEAR_RELAYS_PACKET_LOSS, relay.getPacketLoss()));
                        }
                        ArrayList<TableRow> selectedRoute =
                                new ArrayList<TableRow>(entry.getRouteCount());
                        for (BillingRouteHop hop : entry.getRouteList()) {
                            selectedRoute.add(new TableRow()
                                    .set(BillingBq.ROUTE_ID,
                                            EntityIdHelpers.getStringForNetworkNextStorage(
                                                    hop.getRelayId()))
                                    .set(BillingBq.ROUTE_SELLER_ID,
                                            EntityIdHelpers.getStringForNetworkNextStorage(
                                                    hop.getSellerId()))
                                    .set(BillingBq.ROUTE_PRICE_INGRESS, hop.getPriceIngress())
                                    .set(BillingBq.ROUTE_PRICE_EGRESS, hop.getPriceEgress()));
                        }
                        ArrayList<TableRow> consideredRoutes =
                                new ArrayList<TableRow>(entry.getConsideredRoutesCount());
                        for (BillingRoute route : entry.getConsideredRoutesList()) {
                            ArrayList<TableRow> consideredRoute =
                                    new ArrayList<TableRow>(entry.getRouteCount());
                            for (BillingRouteHop hop : route.getRouteList()) {
                                consideredRoute.add(new TableRow()
                                        .set(BillingBq.CONSIDERED_ROUTES_ROUTE_ID,
                                                EntityIdHelpers.getStringForNetworkNextStorage(
                                                        hop.getRelayId()))
                                        .set(BillingBq.CONSIDERED_ROUTES_ROUTE_SELLER_ID,
                                                EntityIdHelpers.getStringForNetworkNextStorage(
                                                        hop.getSellerId()))
                                        .set(BillingBq.CONSIDERED_ROUTES_ROUTE_PRICE_INGRESS,
                                                hop.getPriceIngress())
                                        .set(BillingBq.CONSIDERED_ROUTES_ROUTE_PRICE_EGRESS,
                                                hop.getPriceEgress()));
                            }
                            consideredRoutes.add(new TableRow()
                                    .set(BillingBq.CONSIDERED_ROUTES_ROUTE, consideredRoute));
                        }
                        ArrayList<TableRow> acceptableRoutes =
                                new ArrayList<TableRow>(entry.getAcceptableRoutesCount());
                        for (BillingRoute route : entry.getAcceptableRoutesList()) {
                            ArrayList<TableRow> acceptableRoute =
                                    new ArrayList<TableRow>(entry.getRouteCount());
                            for (BillingRouteHop hop : route.getRouteList()) {
                                acceptableRoute.add(new TableRow()
                                        .set(BillingBq.ACCEPTABLE_ROUTES_ROUTE_ID,
                                                EntityIdHelpers.getStringForNetworkNextStorage(
                                                        hop.getRelayId()))
                                        .set(BillingBq.ACCEPTABLE_ROUTES_ROUTE_SELLER_ID,
                                                EntityIdHelpers.getStringForNetworkNextStorage(
                                                        hop.getSellerId()))
                                        .set(BillingBq.ACCEPTABLE_ROUTES_ROUTE_PRICE_INGRESS,
                                                hop.getPriceIngress())
                                        .set(BillingBq.ACCEPTABLE_ROUTES_ROUTE_PRICE_EGRESS,
                                                hop.getPriceEgress()));
                            }
                            acceptableRoutes.add(new TableRow()
                                    .set(BillingBq.ACCEPTABLE_ROUTES_ROUTE, acceptableRoute));
                        }
                        ArrayList<TableRow> issuedNearRelays = new ArrayList<TableRow>(
                                entry.getRequest().getIssuedNearRelaysCount());
                        for (NearData.IssuedNearRelay issuedNearRelay : entry.getRequest()
                                .getIssuedNearRelaysList()) {
                            issuedNearRelays.add(new TableRow()
                                    .set(BillingBq.ISSUED_NEAR_RELAYS_INDEX,
                                            issuedNearRelay.getIndex())
                                    .set(BillingBq.ISSUED_NEAR_RELAYS_ID,
                                            EntityIdHelpers.getStringForNetworkNextStorage(
                                                    issuedNearRelay.getRelayId()))
                                    .set(BillingBq.ISSUED_NEAR_RELAYS_IP_ADDRESS,
                                            Utils.addressPortToString(
                                                    issuedNearRelay.getRelayIpAddress())));
                        }
                        return new TableRow()
                                .set(BillingBq.BUYER_ID,
                                        EntityIdHelpers.getStringForNetworkNextStorage(
                                                request.getBuyerId()))
                                .set(BillingBq.SESSION_ID, request.getSessionId())
                                .set(BillingBq.USER_ID, request.getUserHash())
                                .set(BillingBq.PLATFORM_ID, request.getPlatformId())
                                .set(BillingBq.DIRECT_RTT, request.getDirectRtt())
                                .set(BillingBq.DIRECT_JITTER, request.getDirectJitter())
                                .set(BillingBq.DIRECT_PACKET_LOSS, request.getDirectPacketLoss())
                                .set(BillingBq.NEXT_RTT, request.getNextRtt())
                                .set(BillingBq.NEXT_JITTER, request.getNextJitter())
                                .set(BillingBq.NEXT_PACKET_LOSS, request.getNextPacketLoss())
                                .set(BillingBq.CLIENT_IP_ADDRESS,
                                        Utils.addressToString(request.getClientIpAddress()))
                                .set(BillingBq.SERVER_IP_ADDRESS,
                                        Utils.addressToString(request.getServerIpAddress()))
                                .set(BillingBq.SERVER_PRIVATE_IP_ADDRESS,
                                        Utils.addressToString(request.getServerPrivateIpAddress()))
                                .set(BillingBq.TAG, request.getTag())
                                .set(BillingBq.NEAR_RELAYS, nearRelays)
                                .set(BillingBq.ISSUED_NEAR_RELAYS, issuedNearRelays)
                                .set(BillingBq.CONNECTION_TYPE,
                                        request.getConnectionType().getNumber())
                                .set(BillingBq.DATACENTER_ID,
                                        EntityIdHelpers.getStringForNetworkNextStorage(
                                                request.getDatacenterId()))
                                .set(BillingBq.SEQUENCE_NUMBER, request.getSequenceNumber())
                                .set(BillingBq.FALLBACK_TO_DIRECT, request.getFallbackToDirect())
                                .set(BillingBq.VERSION_MAJOR, request.getVersionMajor())
                                .set(BillingBq.VERSION_MINOR, request.getVersionMinor())
                                .set(BillingBq.VERSION_PATCH, request.getVersionPatch())
                                .set(BillingBq.KBPS_UP, request.getUsageKbpsUp())
                                .set(BillingBq.KBPS_DOWN, request.getUsageKbpsDown())
                                .set(BillingBq.COUNTRY_CODE, ip2location.getCountryCode())
                                .set(BillingBq.COUNTRY, ip2location.getCountry())
                                .set(BillingBq.REGION, ip2location.getRegion())
                                .set(BillingBq.CITY, ip2location.getCity())
                                .set(BillingBq.LATITUDE, ip2location.getLatitude())
                                .set(BillingBq.LONGITUDE, ip2location.getLongitude())
                                .set(BillingBq.ISP, ip2location.getIsp())
                                .set(BillingBq.ROUTE, selectedRoute)
                                .set(BillingBq.ROUTE_DECISION, entry.getRouteDecision())
                                .set(BillingBq.DURATION, entry.getDuration())
                                .set(BillingBq.BYTES_UP, entry.getEnvelopeBytesUp())
                                .set(BillingBq.BYTES_DOWN, entry.getEnvelopeBytesDown())
                                .set(BillingBq.TIMESTAMP, entry.getTimestamp())
                                .set(BillingBq.TIMESTAMP_START, entry.getTimestampStart())
                                .set(BillingBq.PREDICTED_RTT, entry.getPredictedRtt())
                                .set(BillingBq.PREDICTED_JITTER, entry.getPredictedJitter())
                                .set(BillingBq.PREDICTED_PACKET_LOSS,
                                        entry.getPredictedPacketLoss())
                                .set(BillingBq.ROUTE_CHANGED, entry.getRouteChanged())
                                .set(BillingBq.NETWORK_NEXT, entry.getNetworkNextAvailable())
                                .set(BillingBq.INITIAL, entry.getInitial())
                                .set(BillingBq.FLAGGED, request.getFlagged())
                                .set(BillingBq.TRY_BEFORE_YOU_BUY, request.getTryBeforeYouBuy())
                                .set(BillingBq.PACKETS_LOST_CLIENT_TO_SERVER,
                                        request.getPacketsLostClientToServer())
                                .set(BillingBq.PACKETS_LOST_SERVER_TO_CLIENT,
                                        request.getPacketsLostServerToClient())
                                .set(BillingBq.CONSIDERED_ROUTES, consideredRoutes)
                                .set(BillingBq.ACCEPTABLE_ROUTES, acceptableRoutes)
                                .set(BillingBq.SAME_ROUTE, entry.getSameRoute());
                    }

                }).to(context.bigQueryBilling).withTableDescription(BillingBq.TABLE_DESCRIPTION)
                .withSchema(BillingBq.schema).optimizedWrites()
                .withTimePartitioning(new TimePartitioning().setRequirePartitionFilter(true)
                        .setType("DAY").setField(BillingBq.TIMESTAMP_START))
                .withCreateDisposition(BigQueryIO.Write.CreateDisposition.CREATE_IF_NEEDED)
                .withWriteDisposition(BigQueryIO.Write.WriteDisposition.WRITE_APPEND)
                .withMethod(BigQueryIO.Write.Method.STREAMING_INSERTS)
                .withFailedInsertRetryPolicy(InsertRetryPolicy.retryTransientErrors())
                .withExtendedErrorInfo());
    }
}
