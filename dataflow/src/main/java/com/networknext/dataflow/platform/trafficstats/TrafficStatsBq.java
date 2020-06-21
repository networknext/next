package com.networknext.dataflow.platform.trafficstats;

import com.google.api.services.bigquery.model.TableFieldSchema;
import com.google.api.services.bigquery.model.TableSchema;
import com.google.common.collect.ImmutableList;

public class TrafficStatsBq {

    public static final String ID = "id";
    public static final String TIMESTAMP = "timestamp";
    public static final String USAGE = "usage";
    public static final String TRAFFIC_STATS = "trafficStats";
    public static final String BYTES_PAID_TX = "bytesPaidTx";
    public static final String BYTES_PAID_RX = "bytesPaidRx";
    public static final String BYTES_MANAGEMENT_TX = "bytesManagementTx";
    public static final String BYTES_MANAGEMENT_RX = "bytesManagementRx";
    public static final String BYTES_MEASUREMENT_TX = "bytesMeasurementTx";
    public static final String BYTES_MEASUREMENT_RX = "bytesMeasurementRx";
    public static final String BYTES_INVALID_RX = "bytesInvalidRx";
    public static final String SESSION_COUNT = "sessionCount";

    public static TableSchema schema;

    static {
        schema = new TableSchema().setFields(ImmutableList.of(
                new TableFieldSchema().setName(ID).setType("STRING").setMode("NULLABLE"),
                new TableFieldSchema().setName(TIMESTAMP).setType("TIMESTAMP").setMode("NULLABLE"),
                new TableFieldSchema().setName(USAGE).setType("FLOAT").setMode("NULLABLE"),
                new TableFieldSchema().setName(TRAFFIC_STATS).setType("RECORD").setMode("NULLABLE")
                        .setFields(ImmutableList.of(
                                new TableFieldSchema().setName(BYTES_PAID_TX).setType("INTEGER")
                                        .setMode("NULLABLE"),
                                new TableFieldSchema().setName(BYTES_PAID_RX).setType("INTEGER")
                                        .setMode("NULLABLE"),
                                new TableFieldSchema().setName(BYTES_MANAGEMENT_TX)
                                        .setType("INTEGER").setMode("NULLABLE"),
                                new TableFieldSchema().setName(BYTES_MANAGEMENT_RX)
                                        .setType("INTEGER").setMode("NULLABLE"),
                                new TableFieldSchema().setName(BYTES_MEASUREMENT_TX)
                                        .setType("INTEGER").setMode("NULLABLE"),
                                new TableFieldSchema().setName(BYTES_MEASUREMENT_RX)
                                        .setType("INTEGER").setMode("NULLABLE"),
                                new TableFieldSchema().setName(BYTES_INVALID_RX).setType("INTEGER")
                                        .setMode("NULLABLE"),
                                new TableFieldSchema().setName(SESSION_COUNT).setType("INTEGER")
                                        .setMode("NULLABLE")))));
    }

}
