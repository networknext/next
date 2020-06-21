package com.networknext.dataflow.export.sellerexport;

import java.net.InetAddress;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import com.google.protobuf.ByteString;
import com.google.protobuf.Descriptors.FieldDescriptor;
import com.networknext.api.near_data.NearData.NearRelay;
import com.networknext.api.router.Router.BillingEntry;
import com.networknext.api.seller_near_export.SellerNearExport.ExportedSession;
import com.networknext.api.seller_near_export.SellerNearExport.ExportedSessionConnectionType;
import com.networknext.api.seller_near_export.SellerNearExport.ExportedSessionPlatformType;
import com.networknext.dataflow.export.ExportContext;
import com.networknext.dataflow.util.EntityIdHelpers;
import com.networknext.dataflow.util.pubsub.PubSubTopicAutoCreate;
import org.apache.avro.Schema;
import org.apache.avro.SchemaBuilder;
import org.apache.avro.generic.GenericData;
import org.apache.avro.generic.GenericRecord;
import org.apache.avro.generic.GenericRecordBuilder;
import org.apache.beam.sdk.coders.AvroCoder;
import org.apache.beam.sdk.io.FileIO;
import org.apache.beam.sdk.io.gcp.pubsub.PubsubIO;
import org.apache.beam.sdk.io.parquet.ParquetIO;
import org.apache.beam.sdk.transforms.DoFn;
import org.apache.beam.sdk.transforms.ParDo;
import org.apache.beam.sdk.transforms.windowing.FixedWindows;
import org.apache.beam.sdk.transforms.windowing.Window;
import org.apache.beam.sdk.values.PCollection;
import org.joda.time.Duration;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class SellerExportPipeline {

    private static final Logger LOG = LoggerFactory.getLogger(SellerExportPipeline.class);

    private static class Mapping {
        String name;
        String pubSub;
        String parquet;
    }

    public static void buildSellerExportPipeline(ExportContext context) {

        // read config first

        List<Mapping> mappings = new ArrayList<Mapping>();
        String config = context.sellerExportConfig;
        if (config != null) {
            String[] entries = config.split(";");
            for (String entry : entries) {
                Mapping mapping = new Mapping();
                String[] settings = entry.split(",");
                for (String setting : settings) {
                    String[] kv = setting.split("=", 2);
                    if (kv.length == 2) {
                        String key = kv[0];
                        String value = kv[1];

                        if ("name".equals(key)) {
                            mapping.name = value;
                        }
                        if ("pubsub".equals(key)) {
                            mapping.pubSub = value;
                        }
                        if ("parquet".equals(key)) {
                            mapping.parquet = value;
                        }
                    }
                }
                if (mapping.name != null && mapping.name.trim().equals("")) {
                    mapping.name = null;
                }
                if (mapping.pubSub != null && mapping.pubSub.trim().equals("")) {
                    mapping.pubSub = null;
                }
                if (mapping.parquet != null && mapping.parquet.trim().equals("")) {
                    mapping.parquet = null;
                }
                if (mapping.name != null) {
                    if (mapping.pubSub != null || mapping.parquet != null) {
                        mappings.add(mapping);
                    }
                } else {
                    LOG.error("Mapping entry does not specify name=, skipping.");
                }
            }
        }

        if (mappings.size() == 0) {
            // no sellers to export to, skip this pipeline entirely
            return;
        }

        // seller export pipeline

        PCollection<ExportedSession> exportPipeline =
                context.pubSubBilling.apply("Get exportable slice data from Pub/Sub",
                        ParDo.of(new DoFn<BillingEntry, ExportedSession>() {
                            private static final long serialVersionUID = 6922898617241332266L;

                            @ProcessElement
                            public void processElement(ProcessContext c) throws Exception {

                                BillingEntry entry = c.element();

                                ExportedSessionConnectionType connectionType =
                                        ExportedSessionConnectionType.EXPORTED_SESSION_CONNECTION_TYPE_UNKNOWN;
                                switch (entry.getRequest().getConnectionType()) {
                                    case SESSION_CONNECTION_TYPE_WIRED:
                                        connectionType =
                                                ExportedSessionConnectionType.EXPORTED_SESSION_CONNECTION_TYPE_WIRED;
                                        break;
                                    case SESSION_CONNECTION_TYPE_WIFI:
                                        connectionType =
                                                ExportedSessionConnectionType.EXPORTED_SESSION_CONNECTION_TYPE_WIFI;
                                        break;
                                    case SESSION_CONNECTION_TYPE_CELLULAR:
                                        connectionType =
                                                ExportedSessionConnectionType.EXPORTED_SESSION_CONNECTION_TYPE_CELLULAR;
                                        break;
                                    case SESSION_CONNECTION_TYPE_UNKNOWN:
                                    default:
                                        connectionType =
                                                ExportedSessionConnectionType.EXPORTED_SESSION_CONNECTION_TYPE_UNKNOWN;
                                        break;
                                }

                                ExportedSessionPlatformType platformType =
                                        ExportedSessionPlatformType.EXPORTED_SESSION_PLATFORM_TYPE_UNKNOWN;
                                switch ((int) (entry.getRequest().getPlatformId() & 0xFF)) {
                                    case 0:
                                        platformType =
                                                ExportedSessionPlatformType.EXPORTED_SESSION_PLATFORM_TYPE_UNKNOWN;
                                        break;
                                    case 1:
                                        platformType =
                                                ExportedSessionPlatformType.EXPORTED_SESSION_PLATFORM_TYPE_WINDOWS;
                                        break;
                                    case 2:
                                        platformType =
                                                ExportedSessionPlatformType.EXPORTED_SESSION_PLATFORM_TYPE_MAC;
                                        break;
                                    case 3:
                                        platformType =
                                                ExportedSessionPlatformType.EXPORTED_SESSION_PLATFORM_TYPE_LINUX;
                                        break;
                                    case 4:
                                        platformType =
                                                ExportedSessionPlatformType.EXPORTED_SESSION_PLATFORM_TYPE_SWITCH;
                                        break;
                                    case 5:
                                        platformType =
                                                ExportedSessionPlatformType.EXPORTED_SESSION_PLATFORM_TYPE_PS4;
                                        break;
                                    case 6:
                                        platformType =
                                                ExportedSessionPlatformType.EXPORTED_SESSION_PLATFORM_TYPE_IOS;
                                        break;
                                    case 7:
                                        platformType =
                                                ExportedSessionPlatformType.EXPORTED_SESSION_PLATFORM_TYPE_XBOX_ONE;
                                        break;
                                }

                                // IP addresses are now anonymized at the system level; we don't
                                // store the
                                // original IP address at all, therefore we don't need to do any
                                // anonymization
                                // here.

                                ExportedSession.Builder builder = ExportedSession.newBuilder()
                                        .setTimestamp(com.google.protobuf.Timestamp.newBuilder()
                                                .setSeconds(entry.getTimestamp()).build())
                                        .setSessionId(entry.getRequest().getSessionId())
                                        .setIpAddress(ByteString.copyFrom(entry.getRequest()
                                                .getClientIpAddress().getIp().toByteArray()))
                                        .setLatitude((float) entry.getRequest().getLocation()
                                                .getLatitude())
                                        .setLongitude((float) entry.getRequest().getLocation()
                                                .getLongitude())
                                        .clearCity().clearCountry()
                                        .setIsp(entry.getRequest().getLocation().getIsp())
                                        .clearIspAsn().setPlatformType(platformType)
                                        .setConnectionType(connectionType)
                                        .setNumNearRelays(entry.getRequest().getNearRelaysCount());
                                for (NearRelay relay : entry.getRequest().getNearRelaysList()) {
                                    String nearRelayIdHash = EntityIdHelpers
                                            .getStringForThirdPartyStorage(relay.getRelayId());

                                    builder.addNearRelayIdHash(nearRelayIdHash);
                                    builder.addNearRelayRtt((float) relay.getRtt());
                                    builder.addNearRelayJitter((float) relay.getJitter());
                                    builder.addNearRelayPacketLoss((float) relay.getPacketLoss());
                                }

                                ExportedSession session = builder.build();

                                c.output(session);
                            }
                        }));

        boolean needsParquetExport = false;
        for (Mapping mapping : mappings) {
            if (mapping.parquet != null) {
                needsParquetExport = true;
            }
        }

        PCollection<GenericRecord> parquetPipeline = null;
        org.apache.avro.Schema parquetSchema = null;
        if (needsParquetExport) {
            parquetSchema = org.apache.avro.SchemaBuilder.record("session").fields()
                    .requiredLong("timestampUtc").requiredLong("sessionId")
                    .requiredString("ipAddress").requiredFloat("latitude")
                    .requiredFloat("longitude").optionalString("city").optionalString("country")
                    .requiredString("isp").optionalLong("ispAsn").requiredString("connectionType")
                    .requiredString("platformType").requiredLong("numNearRelays")
                    .name("nearRelaysIdHash").type().array().items().stringType().noDefault()
                    .name("nearRelaysRtt").type().array().items().floatType().noDefault()
                    .name("nearRelaysJitter").type().array().items().floatType().noDefault()
                    .name("nearRelaysPacketLoss").type().array().items().floatType().noDefault()
                    .endRecord();

            parquetPipeline = exportPipeline
                    .apply("Window hourly",
                            Window.<ExportedSession>into(
                                    FixedWindows.of(Duration.standardHours(1))))
                    .apply("Convert to Parquet",
                            ParDo.of(new DoFn<ExportedSession, GenericRecord>() {
                                private static final long serialVersionUID = 6922898617241332366L;

                                private Schema parquetSchema;

                                @Setup
                                public void setup() {
                                    // not serializable, so we have to define it in both places :/
                                    this.parquetSchema = org.apache.avro.SchemaBuilder
                                            .record("session").fields().requiredLong("timestampUtc")
                                            .requiredLong("sessionId").requiredString("ipAddress")
                                            .requiredFloat("latitude").requiredFloat("longitude")
                                            .optionalString("city").optionalString("country")
                                            .requiredString("isp").optionalLong("ispAsn")
                                            .requiredString("connectionType")
                                            .requiredString("platformType")
                                            .requiredLong("numNearRelays").name("nearRelaysIdHash")
                                            .type().array().items().stringType().noDefault()
                                            .name("nearRelaysRtt").type().array().items()
                                            .floatType().noDefault().name("nearRelaysJitter").type()
                                            .array().items().floatType().noDefault()
                                            .name("nearRelaysPacketLoss").type().array().items()
                                            .floatType().noDefault().endRecord();
                                }

                                @ProcessElement
                                public void processElement(ProcessContext c) throws Exception {
                                    ExportedSession session = c.element();

                                    byte[] rawAddr = session.getIpAddress().toByteArray();
                                    if (rawAddr.length != 4 && rawAddr.length != 16) {
                                        // invalid address, ignore this slice (this should only
                                        // occur in v2).
                                        return;
                                    }

                                    InetAddress addr = InetAddress.getByAddress(rawAddr);

                                    String connectionType = "unknown";
                                    String platformType = "unknown";
                                    switch (session.getConnectionType()) {
                                        case EXPORTED_SESSION_CONNECTION_TYPE_UNKNOWN:
                                            connectionType = "unknown";
                                            break;
                                        case EXPORTED_SESSION_CONNECTION_TYPE_WIRED:
                                            connectionType = "wired";
                                            break;
                                        case EXPORTED_SESSION_CONNECTION_TYPE_WIFI:
                                            connectionType = "wifi";
                                            break;
                                        case EXPORTED_SESSION_CONNECTION_TYPE_CELLULAR:
                                            connectionType = "cellular";
                                            break;
                                        default:
                                            connectionType = "unknown";
                                            break;
                                    }
                                    switch (session.getPlatformType()) {
                                        case EXPORTED_SESSION_PLATFORM_TYPE_UNKNOWN:
                                            platformType = "unknown";
                                            break;
                                        case EXPORTED_SESSION_PLATFORM_TYPE_WINDOWS:
                                            platformType = "windows";
                                            break;
                                        case EXPORTED_SESSION_PLATFORM_TYPE_MAC:
                                            platformType = "mac";
                                            break;
                                        case EXPORTED_SESSION_PLATFORM_TYPE_LINUX:
                                            platformType = "linux";
                                            break;
                                        case EXPORTED_SESSION_PLATFORM_TYPE_SWITCH:
                                            platformType = "switch";
                                            break;
                                        case EXPORTED_SESSION_PLATFORM_TYPE_PS4:
                                            platformType = "ps4";
                                            break;
                                        case EXPORTED_SESSION_PLATFORM_TYPE_IOS:
                                            platformType = "ios";
                                            break;
                                        case EXPORTED_SESSION_PLATFORM_TYPE_XBOX_ONE:
                                            platformType = "xboxone";
                                            break;
                                        default:
                                            platformType = "unknown";
                                            break;
                                    }

                                    List<String> nearRelaysIdHash = new ArrayList<String>();
                                    List<Float> nearRelaysRtt = new ArrayList<Float>();
                                    List<Float> nearRelaysJitter = new ArrayList<Float>();
                                    List<Float> nearRelaysPacketLoss = new ArrayList<Float>();
                                    for (int i = 0; i < (int) session.getNumNearRelays(); i++) {
                                        nearRelaysIdHash.add(session.getNearRelayIdHash(i));
                                        nearRelaysRtt.add(session.getNearRelayRtt(i));
                                        nearRelaysJitter.add(session.getNearRelayJitter(i));
                                        nearRelaysPacketLoss.add(session.getNearRelayPacketLoss(i));
                                    }

                                    GenericRecordBuilder recordBuilder =
                                            new GenericRecordBuilder(parquetSchema)
                                                    .set("timestampUtc",
                                                            session.getTimestamp().getSeconds())
                                                    .set("sessionId", session.getSessionId())
                                                    .set("ipAddress", addr.getHostAddress())
                                                    .set("latitude", session.getLatitude())
                                                    .set("longitude", session.getLongitude())
                                                    .set("isp", session.getIsp())
                                                    .set("connectionType", connectionType)
                                                    .set("platformType", platformType)
                                                    .set("numNearRelays",
                                                            (long) session.getNumNearRelays())
                                                    .set("nearRelaysIdHash",
                                                            new GenericData.Array<String>(
                                                                    SchemaBuilder.array().items()
                                                                            .stringType(),
                                                                    nearRelaysIdHash))
                                                    .set("nearRelaysRtt",
                                                            new GenericData.Array<Float>(
                                                                    SchemaBuilder.array().items()
                                                                            .floatType(),
                                                                    nearRelaysRtt))
                                                    .set("nearRelaysJitter",
                                                            new GenericData.Array<Float>(
                                                                    SchemaBuilder.array().items()
                                                                            .floatType(),
                                                                    nearRelaysJitter))
                                                    .set("nearRelaysPacketLoss",
                                                            new GenericData.Array<Float>(
                                                                    SchemaBuilder.array().items()
                                                                            .floatType(),
                                                                    nearRelaysPacketLoss));

                                    if (SellerExportPipeline.hasField(session,
                                            ExportedSession.CITY_FIELD_NUMBER)) {
                                        recordBuilder.set("city", session.getCity());
                                    } else {
                                        recordBuilder.set("city", null);
                                    }
                                    if (SellerExportPipeline.hasField(session,
                                            ExportedSession.COUNTRY_FIELD_NUMBER)) {
                                        recordBuilder.set("country", session.getCountry());
                                    } else {
                                        recordBuilder.set("country", null);
                                    }
                                    if (SellerExportPipeline.hasField(session,
                                            ExportedSession.ISPASN_FIELD_NUMBER)) {
                                        recordBuilder.set("ispAsn", session.getIspAsn());
                                    } else {
                                        recordBuilder.set("ispAsn", null);
                                    }

                                    GenericRecord record = recordBuilder.build();

                                    c.output(record);
                                }
                            }))
                    .setCoder(AvroCoder.of(parquetSchema));
        }

        for (Mapping mapping : mappings) {
            if (mapping.pubSub != null) {
                // Create Pub/Sub topic (required for local testing)
                try {
                    PubSubTopicAutoCreate.execute(null, mapping.pubSub);
                } catch (Exception e) {
                    // Ignore
                }

                exportPipeline.apply("Forward Pub/Sub to " + mapping.name,
                        PubsubIO.writeProtos(ExportedSession.class).to(mapping.pubSub));
            }

            if (mapping.parquet != null && parquetPipeline != null) {
                parquetPipeline.apply("Forward Parquet to " + mapping.name,
                        FileIO.<GenericRecord>write().withNumShards(10)
                                .via(ParquetIO.sink(parquetSchema)).to(mapping.parquet));
            }
        }

    }

    public static boolean hasField(com.google.protobuf.GeneratedMessageV3 proto, int field) {
        Map<FieldDescriptor, Object> modifiedFields = proto.getAllFields();

        for (FieldDescriptor fieldDescriptor : modifiedFields.keySet()) {
            if (fieldDescriptor.toProto().getNumber() == field) {
                return true;
            }
        }

        return false;
    }


}
