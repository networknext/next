package com.networknext.dataflow.platform.trafficstats;

import com.google.api.services.bigquery.model.TableRow;
import com.google.api.services.bigquery.model.TimePartitioning;
import com.networknext.api.relay_data.RelayData.RelayTrafficStats;
import com.networknext.dataflow.platform.PlatformContext;
import com.networknext.dataflow.util.EntityIdHelpers;
import com.networknext.dataflow.util.bigquery.BigQuerySchemaUpdater;
import org.apache.beam.sdk.io.gcp.bigquery.BigQueryIO;
import org.apache.beam.sdk.io.gcp.bigquery.InsertRetryPolicy;
import org.apache.beam.sdk.transforms.DoFn;
import org.apache.beam.sdk.transforms.ParDo;

public class TrafficStatsPipeline {

    public static void buildTrafficStatsPipeline(PlatformContext context) {

        // traffic stats pipeline

        BigQuerySchemaUpdater.execute(context.bigQueryTrafficStats, TrafficStatsBq.schema);

        context.pubSubTrafficStats
                .apply("Convert to BigQuery", ParDo.of(new DoFn<RelayTrafficStats, TableRow>() {
                    private static final long serialVersionUID = 209374876543453745L;

                    @ProcessElement
                    public void processElement(ProcessContext c) throws Exception {
                        RelayTrafficStats stats = c.element();

                        c.output(new TableRow()
                                .set(TrafficStatsBq.ID,
                                        EntityIdHelpers.getStringForNetworkNextStorage(
                                                stats.getRelayId()))
                                .set(TrafficStatsBq.TIMESTAMP, stats.getTimestamp().getSeconds())
                                .set(TrafficStatsBq.USAGE, stats.getUsage())
                                .set(TrafficStatsBq.TRAFFIC_STATS, new TableRow()
                                        .set(TrafficStatsBq.BYTES_PAID_TX, stats.getBytesPaidTx())
                                        .set(TrafficStatsBq.BYTES_PAID_RX, stats.getBytesPaidRx())
                                        .set(TrafficStatsBq.BYTES_MANAGEMENT_TX,
                                                stats.getBytesManagementTx())
                                        .set(TrafficStatsBq.BYTES_MANAGEMENT_RX,
                                                stats.getBytesManagementRx())
                                        .set(TrafficStatsBq.BYTES_MEASUREMENT_TX,
                                                stats.getBytesMeasurementTx())
                                        .set(TrafficStatsBq.BYTES_MEASUREMENT_RX,
                                                stats.getBytesMeasurementRx())
                                        .set(TrafficStatsBq.BYTES_INVALID_RX,
                                                stats.getBytesInvalidRx())
                                        .set(TrafficStatsBq.SESSION_COUNT,
                                                stats.getSessionCount())));
                    }
                }))
                .apply("Insert into BigQuery", BigQueryIO.writeTableRows()
                        .to(context.bigQueryTrafficStats).withSchema(TrafficStatsBq.schema)
                        .withTimePartitioning(new TimePartitioning().setRequirePartitionFilter(true)
                                .setType("DAY").setField(TrafficStatsBq.TIMESTAMP))
                        .withCreateDisposition(BigQueryIO.Write.CreateDisposition.CREATE_IF_NEEDED)
                        .withWriteDisposition(BigQueryIO.Write.WriteDisposition.WRITE_APPEND)
                        .withMethod(BigQueryIO.Write.Method.STREAMING_INSERTS)
                        .withFailedInsertRetryPolicy(InsertRetryPolicy.retryTransientErrors())
                        .withExtendedErrorInfo());

    }

}
