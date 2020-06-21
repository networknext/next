package com.networknext.dataflow.platform.session;

import java.nio.ByteBuffer;
import java.nio.charset.StandardCharsets;
import java.util.List;
import com.networknext.api.admin.Admin.KeyDBSessionDebugToolUserEntry;
import com.networknext.api.admin.Admin.SessionMapDataPoint;
import com.networknext.api.ip2location.Ip2Location.Location;
import com.networknext.api.router.Router.BillingEntry;
import com.networknext.api.router.Router.RouteRequest;
import com.networknext.api.session.SessionOuterClass.SessionCountResult;
import com.networknext.api.session.SessionOuterClass.SessionsList;
import com.networknext.dataflow.platform.PlatformContext;
import com.networknext.dataflow.util.BillingEntryHelpers;
import com.networknext.dataflow.util.Utils;
import org.apache.beam.sdk.io.redis.RedisIO;
import org.apache.beam.sdk.io.redis.RedisIO.Write.Method;
import org.apache.beam.sdk.transforms.Combine;
import org.apache.beam.sdk.transforms.Distinct;
import org.apache.beam.sdk.transforms.DoFn;
import org.apache.beam.sdk.transforms.ParDo;
import org.apache.beam.sdk.transforms.ParDo.SingleOutput;
import org.apache.beam.sdk.transforms.Top;
import org.apache.beam.sdk.transforms.windowing.AfterProcessingTime;
import org.apache.beam.sdk.transforms.windowing.SlidingWindows;
import org.apache.beam.sdk.transforms.windowing.Window;
import org.apache.beam.sdk.values.KV;
import org.apache.beam.sdk.values.PCollection;
import org.joda.time.Duration;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.threeten.bp.Instant;

public class SessionPipeline {

    private static final Logger LOG = LoggerFactory.getLogger(SessionPipeline.class);

    private static Long THIRTY_DAYS_MS = 1000L * 60 * 60 * 24 * 30;
    private static Long THIRTY_MINUTES_MS = 1000L * 60 * 30;

    public static void buildSessionPipeline(PlatformContext context) throws Exception {

        context.pubSubBilling.apply("Prepare KeyDB billing",
                ParDo.of(new DoFn<BillingEntry, KV<byte[], byte[]>>() {
                    private static final long serialVersionUID = 1L;

                    @ProcessElement
                    public void processElement(ProcessContext c) throws Exception {

                        BillingEntry entry = c.element();

                        LOG.info("processed billing entry");

                        if (entry.getRequest().getBuyerId().getName() == "dtqb5J5pAGMS7IizFfWD" && !entry.getOnNetworkNext()) {
                                return;
                        }

                        byte[] key = String
                                .format("sdt-billing#%s#%s#%d",
                                        entry.getRequest().getBuyerId().getName(),
                                        Utils.hexPrint(entry.getRequest().getSessionId()),
                                        entry.getTimestamp())
                                .getBytes(StandardCharsets.UTF_8);

                        c.output(KV.of(key, entry.toByteArray()));
                    }
                }))
                .apply("Write KeyDB billing", RedisIO.write()
                        .withEndpoint(context.keyDbSessionToolHost, context.keyDbSessionToolPort)
                        .withMethod(Method.SET).withExpireTime(THIRTY_MINUTES_MS));

        // For Network Next admins, we don't know the buyer of the session, so write a session ID ->
        // buyer name
        // lookup table that the admin service can use to then resolve the actual KeyDB key for a
        // session's slice data.

        context.pubSubBilling.apply("Prepare KeyDB global session lookup",
                ParDo.of(new DoFn<BillingEntry, KV<byte[], byte[]>>() {
                    private static final long serialVersionUID = 1L;

                    @ProcessElement
                    public void processElement(ProcessContext c) throws Exception {

                        BillingEntry entry = c.element();

                        byte[] key = String
                                .format("sdt-global-session-lookup#%s",
                                        Utils.hexPrint(entry.getRequest().getSessionId()))
                                .getBytes(StandardCharsets.UTF_8);
                        byte[] value = entry.getRequest().getBuyerId().getName()
                                .getBytes(StandardCharsets.UTF_8);

                        c.output(KV.of(key, value));
                    }
                }))
                .apply("Write KeyDB global session lookup", RedisIO.write()
                        .withEndpoint(context.keyDbSessionToolHost, context.keyDbSessionToolPort)
                        .withMethod(Method.SET).withExpireTime(THIRTY_MINUTES_MS));

        context.pubSubBilling
                .apply("Prepare KeyDB user", ParDo.of(new DoFn<BillingEntry, KV<byte[], byte[]>>() {
                    private static final long serialVersionUID = 1L;

                    @ProcessElement
                    public void processElement(ProcessContext c) throws Exception {

                        BillingEntry entry = c.element();

                        if (entry.getRequest().getBuyerId().getName() == "dtqb5J5pAGMS7IizFfWD" && !entry.getOnNetworkNext()) {
                                return;
                        }

                        if (entry.getInitial()) {
                            int version = entry.getRequest().getVersionMajor() * 100
                                    + entry.getRequest().getVersionMinor() * 10
                                    + entry.getRequest().getVersionPatch();

                            KeyDBSessionDebugToolUserEntry userEntry =
                                    KeyDBSessionDebugToolUserEntry.newBuilder()
                                            .setSessionId(entry.getRequest().getSessionId())
                                            .setBuyerEntityId(
                                                    entry.getRequest().getBuyerId().getName())
                                            .setTimestampStart(entry.getTimestampStart())
                                            .setUserHash(entry.getRequest().getUserHash())
                                            .setDatacenterEntityId(
                                                    entry.getRequest().getDatacenterId().getName())
                                            .setServerAddress(Utils.addressToString(
                                                    entry.getRequest().getServerIpAddress()))
                                            .setPlatform(entry.getRequest().getPlatformId())
                                            .setVersion(version)
                                            .setConnectionType(entry.getRequest()
                                                    .getConnectionType().getNumber())
                                            .setUserIp(Utils.addressToString(
                                                    entry.getRequest().getClientIpAddress()))
                                            .setUserIsp(entry.getRequest().getLocation().getIsp())
                                            .setUserLatitude((double) entry.getRequest()
                                                    .getLocation().getLatitude())
                                            .setUserLongitude((double) entry.getRequest()
                                                    .getLocation().getLongitude())
                                            .build();

                            byte[] key = String
                                    .format("sdt-user#%s#%s#%d",
                                            entry.getRequest().getBuyerId().getName(),
                                            Utils.hexPrint(entry.getRequest().getUserHash()),
                                            entry.getTimestamp())
                                    .getBytes(StandardCharsets.UTF_8);

                            c.output(KV.of(key, userEntry.toByteArray()));
                        }
                    }
                }))
                .apply("Write KeyDB user", RedisIO.write()
                        .withEndpoint(context.keyDbSessionToolHost, context.keyDbSessionToolPort)
                        .withMethod(Method.SET).withExpireTime(THIRTY_MINUTES_MS));

        // For session lists and the map, we need to bucket the incoming slices into 10 second
        // buckets. At the end of each 10 seconds, the KeyDB values will be emitted. Here we trade
        // exactness for realtime-ness ("this list is exactly the latest information for all
        // sessions in a 10 second window" for "we always emit a list every 10 seconds that's
        // mostly up to date"). This is because Pub/Sub is not a perfectly realtime stream, and we
        // don't want to delay updating the viewer on 99% of sessions just because information about
        // one session is late.

        /*
         * 
         * 
         * 
         * 
         * new GlobalWindows())
         * .triggering(Repeatedly.forever(AfterProcessingTime.pastFirstElementInPane()
         * .plusDelayOf(Duration.standardSeconds(10))))
         * .withAllowedLateness(Duration.ZERO).discardingFiredPanes()
         */

        PCollection<BillingEntry> windowedSlices = context.pubSubBilling
                .apply("KeyDB Window slices",
                        Window.<BillingEntry>into(SlidingWindows.of(Duration.standardSeconds(60))
                                .every(Duration.standardSeconds(30)))
                                .triggering(AfterProcessingTime.pastFirstElementInPane()
                                        .plusDelayOf(Duration.standardSeconds(60)))
                                .withAllowedLateness(Duration.ZERO).discardingFiredPanes())
                .apply("KeyDB Latest slices", new SelectLatestSlice());

        PCollection<KV<String, BillingEntry>> perBuyerWindowedSlices = windowedSlices.apply(
                "KeyDB Key by buyer", ParDo.of(new DoFn<BillingEntry, KV<String, BillingEntry>>() {
                    private static final long serialVersionUID = 1L;

                    @ProcessElement
                    public void processElement(ProcessContext c) throws Exception {
                        BillingEntry element = c.element();
                        c.output(KV.of(element.getRequest().getBuyerId().getName(), element));
                    }
                }));

        // Compute sessions lists for buyers and globally, sorting by most improved RTT.

        perBuyerWindowedSlices
                .apply("KeyDB Compute top 1000 RTT by buyer", Top.perKey(1000, new RttComparator()))
                .apply("KeyDB Prepare top 1000 RTT by buyer",
                        PrepareKeyedSessionsToList("sl-top1000-rtt#b"))
                .apply("KeyDB Write top 1000 RTT by buyer", RedisIO.write()
                        .withEndpoint(context.keyDbSessionListHost, context.keyDbSessionListPort)
                        .withMethod(Method.SET).withExpireTime(THIRTY_DAYS_MS));


        windowedSlices.apply("KeyDB Compute top 1000 RTT globally",
                // todo: this should be asSingletonView but I can't figure out how to get the value
                // out.
                Top.of(1000, new RttComparator()).withoutDefaults())
                .apply("KeyDB Prepare top 1000 RTT globally",
                        PrepareGlobalSessionsToList("sl-top1000-rtt#g"))
                .apply("KeyDB Write top 1000 RTT globally", RedisIO.write()
                        .withEndpoint(context.keyDbSessionListHost, context.keyDbSessionListPort)
                        .withMethod(Method.SET).withExpireTime(THIRTY_DAYS_MS));

        // Compute sessions lists for buyers and globally, sorting by most improved packet loss.

        perBuyerWindowedSlices
                .apply("KeyDB Compute top 1000 PL by buyer",
                        Top.perKey(1000, new PacketLossComparator()))
                .apply("KeyDB Prepare top 1000 PL by buyer",
                        PrepareKeyedSessionsToList("sl-top1000-pl#b"))
                .apply("KeyDB Write top 1000 PL by buyer", RedisIO.write()
                        .withEndpoint(context.keyDbSessionListHost, context.keyDbSessionListPort)
                        .withMethod(Method.SET).withExpireTime(THIRTY_DAYS_MS));

        windowedSlices
                .apply("KeyDB Compute top 1000 PL globally",
                        Top.of(1000, new PacketLossComparator()).withoutDefaults())
                .apply("KeyDB Prepare top 1000 PL globally",
                        PrepareGlobalSessionsToList("sl-top1000-pl#g"))
                .apply("KeyDB Write top 1000 PL globally", RedisIO.write()
                        .withEndpoint(context.keyDbSessionListHost, context.keyDbSessionListPort)
                        .withMethod(Method.SET).withExpireTime(THIRTY_DAYS_MS));

        // Compute counts ("total sessions" and "sessions on network next") for buyers and global.

        perBuyerWindowedSlices
                .apply("KeyDB Count sessions by buyer", Combine.perKey(new SessionCountAccumFn()))
                .apply("KeyDB Prepare counts by buyer",
                        ParDo.of(new DoFn<KV<String, SessionCountResult>, KV<byte[], byte[]>>() {
                            private static final long serialVersionUID = 1L;

                            @ProcessElement
                            public void processElement(ProcessContext c) throws Exception {
                                byte[] key = String.format("sl-count#b_%s", c.element().getKey())
                                        .getBytes(StandardCharsets.UTF_8);
                                byte[] value = c.element().getValue().toByteArray();

                                c.output(KV.of(key, value));
                            }
                        }))
                .apply("KeyDB Write counts by buyer", RedisIO.write()
                        .withEndpoint(context.keyDbSessionListHost, context.keyDbSessionListPort)
                        .withMethod(Method.SET).withExpireTime(THIRTY_DAYS_MS));

        windowedSlices
                .apply("KeyDB Count sessions globally",
                        Combine.globally(new SessionCountAccumFn()).withoutDefaults())
                .apply("KeyDB Prepare counts globally",
                        ParDo.of(new DoFn<SessionCountResult, KV<byte[], byte[]>>() {
                            private static final long serialVersionUID = 1L;

                            @ProcessElement
                            public void processElement(ProcessContext c) throws Exception {
                                byte[] key = String.format("sl-count#g")
                                        .getBytes(StandardCharsets.UTF_8);
                                byte[] value = c.element().toByteArray();

                                c.output(KV.of(key, value));
                            }
                        }))
                .apply("KeyDB Write counts globally", RedisIO.write()
                        .withEndpoint(context.keyDbSessionListHost, context.keyDbSessionListPort)
                        .withMethod(Method.SET).withExpireTime(THIRTY_DAYS_MS));

        // Compute map data globally.

        PCollection<SessionMapDataPoint> globalMapEntries =
                windowedSlices.apply("KeyDB Convert to map points",
                        ParDo.of(new DoFn<BillingEntry, SessionMapDataPoint>() {
                            private static final long serialVersionUID = 1L;

                            @ProcessElement
                            public void processElement(ProcessContext c) throws Exception {
                                BillingEntry entry = c.element();
                                RouteRequest request = entry.getRequest();
                                Location location = request.getLocation();

                                String countryCode = "XX";
                                long latitudeMinute = 0;
                                long longitudeMinute = 0;
                                if (location != null) {
                                    countryCode = location.getCountryCode();
                                    latitudeMinute = Math.round(location.getLatitude() * 60);
                                    longitudeMinute = Math.round(location.getLongitude() * 60);
                                }

                                LOG.info("got map point");

                                c.output(SessionMapDataPoint.newBuilder()
                                        .setCountryCode(countryCode)
                                        .setLatitudeMinute(latitudeMinute)
                                        .setLongitudeMinute(longitudeMinute)
                                        .setOnNetworkNext(entry.getOnNetworkNext()).setCount(1)
                                        .setTimestamp(entry.getTimestamp()).build());
                            }
                        }));

        // We need to do two seperate aggregations - the first is finding all the distinct
        // points on the map that we need to render, per country (this makes the list of points
        // as small as possible for sending to the browser). The second is the total and on
        // Network Next counts per country for display at the top.

        globalMapEntries
                .apply("KeyDB Map DP Distinct lat/longs", Distinct.<SessionMapDataPoint>create())
                .apply("KeyDB Map DP Key on country code",
                        ParDo.of(new DoFn<SessionMapDataPoint, KV<byte[], byte[]>>() {
                            private static final long serialVersionUID = 1L;

                            @ProcessElement
                            public void processElement(ProcessContext c) throws Exception {
                                SessionMapDataPoint dataPoint = c.element();

                                long nearestTenSecondBucket =
                                        (dataPoint.getTimestamp() / 10L) * 10L;

                                LOG.info("emit country code " + dataPoint.getCountryCode() + " at "
                                        + nearestTenSecondBucket);

                                byte[] key = String
                                        .format("map-points-%d:%d", dataPoint.getLatitudeMinute(), dataPoint.getLongitudeMinute())
                                        .getBytes(StandardCharsets.UTF_8);
                                byte[] value = dataPoint.toByteArray();

                                c.output(KV.of(key, value));
                            }
                        }))
                .apply("KeyDB Map DP Write global", RedisIO.write()
                        .withEndpoint(context.keyDbSessionListHost, context.keyDbSessionListPort)
                        .withMethod(Method.SET).withExpireTime(THIRTY_MINUTES_MS));

        globalMapEntries.apply("KeyDB Map Counts Global group by country",
                ParDo.of(new DoFn<SessionMapDataPoint, KV<String, SessionMapDataPoint>>() {
                    private static final long serialVersionUID = 1L;

                    @ProcessElement
                    public void processElement(ProcessContext c) throws Exception {
                        c.output(KV.of(String.format("map-counts#%d#g#c#%s",
                                (c.element().getTimestamp() / 10L) * 10L,
                                c.element().getCountryCode()), c.element()));
                    }
                }))
                .apply("KeyDB Map Counts Global total by country",
                        Combine.perKey(new MapDataPointAccumFn()))
                .apply("KeyDB Map Counts Global prepare by country",
                        ParDo.of(new DoFn<KV<String, SessionCountResult>, KV<byte[], byte[]>>() {
                            private static final long serialVersionUID = 1L;

                            @ProcessElement
                            public void processElement(ProcessContext c) throws Exception {
                                byte[] key = c.element().getKey().getBytes(StandardCharsets.UTF_8);
                                byte[] value = c.element().getValue().toByteArray();

                                c.output(KV.of(key, value));
                            }
                        }))
                .apply("KeyDB Map Counts Global write by country", RedisIO.write()
                        .withEndpoint(context.keyDbSessionListHost, context.keyDbSessionListPort)
                        .withMethod(Method.SET).withExpireTime(THIRTY_MINUTES_MS));

        globalMapEntries
                .apply("KeyDB Map Get timestamps", ParDo.of(new DoFn<SessionMapDataPoint, Long>() {
                    private static final long serialVersionUID = 1L;

                    @ProcessElement
                    public void processElement(ProcessContext c) throws Exception {
                        c.output((c.element().getTimestamp() / 10L) * 10L);
                    }
                }))
                .apply("KeyDB Map Compute latest timestamp",
                        Combine.globally(Top.largestLongsFn(1)).withoutDefaults())
                .apply("KeyDB Map Prepare latest timestamp",
                        ParDo.of(new DoFn<Iterable<Long>, KV<byte[], byte[]>>() {
                            private static final long serialVersionUID = 1L;

                            @ProcessElement
                            public void processElement(ProcessContext c) throws Exception {
                                Long v = 0L;
                                for (Long vv : c.element()) {
                                    v = vv;
                                    break;
                                }

                                if (v > 0L) {
                                    byte[] key = "map-ts".getBytes(StandardCharsets.UTF_8);
                                    ByteBuffer buffer = ByteBuffer.allocate(Long.BYTES);
                                    buffer.putLong(v);
                                    byte[] value = buffer.array();

                                    LOG.info("updating map timestamp to " + v);

                                    c.output(KV.of(key, value));
                                }
                            }
                        }))
                .apply("KeyDB Map Write latest timestamp", RedisIO.write()
                        .withEndpoint(context.keyDbSessionListHost, context.keyDbSessionListPort)
                        .withMethod(Method.SET).withExpireTime(THIRTY_DAYS_MS));
    }

    private static SingleOutput<KV<String, List<BillingEntry>>, KV<byte[], byte[]>> PrepareKeyedSessionsToList(
            String prefix) {
        return ParDo.of(new DoFn<KV<String, List<BillingEntry>>, KV<byte[], byte[]>>() {
            private static final long serialVersionUID = 1L;

            @ProcessElement
            public void processElement(ProcessContext c) throws Exception {

                String buyerId = c.element().getKey();
                List<BillingEntry> entriesList = c.element().getValue();

                SessionsList.Builder sessionsListBuilder = SessionsList.newBuilder();
                for (int i = 0; i < entriesList.size(); i++) {
                    BillingEntry entry = entriesList.get(i);
                    sessionsListBuilder.addSessions(BillingEntryHelpers.toSession(entry));
                }

                Instant now = Instant.now();
                sessionsListBuilder.setDataTimestampUtc(com.google.protobuf.Timestamp.newBuilder()
                        .setSeconds(now.getEpochSecond()).build());

                byte[] key =
                        String.format("%s_%s", prefix, buyerId).getBytes(StandardCharsets.UTF_8);
                byte[] value = sessionsListBuilder.build().toByteArray();

                c.output(KV.of(key, value));
            }
        });
    }

    private static SingleOutput<List<BillingEntry>, KV<byte[], byte[]>> PrepareGlobalSessionsToList(
            String redisKey) {
        return ParDo.of(new DoFn<List<BillingEntry>, KV<byte[], byte[]>>() {
            private static final long serialVersionUID = 1L;

            @ProcessElement
            public void processElement(ProcessContext c) throws Exception {

                List<BillingEntry> entriesList = c.element();

                SessionsList.Builder sessionsListBuilder = SessionsList.newBuilder();
                for (int i = 0; i < entriesList.size(); i++) {
                    BillingEntry entry = entriesList.get(i);
                    sessionsListBuilder.addSessions(BillingEntryHelpers.toSession(entry));
                }

                Instant now = Instant.now();
                sessionsListBuilder.setDataTimestampUtc(com.google.protobuf.Timestamp.newBuilder()
                        .setSeconds(now.getEpochSecond()).build());

                byte[] key = redisKey.getBytes(StandardCharsets.UTF_8);
                byte[] value = sessionsListBuilder.build().toByteArray();

                c.output(KV.of(key, value));
            }
        });
    }
}
