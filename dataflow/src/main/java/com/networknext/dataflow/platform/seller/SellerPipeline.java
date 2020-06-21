package com.networknext.dataflow.platform.seller;

import com.google.api.services.bigquery.model.TableRow;
import com.google.api.services.bigquery.model.TimePartitioning;
import com.networknext.api.router.Router.BillingEntry;
import com.networknext.api.router.Router.BillingRouteHop;
import com.networknext.dataflow.platform.PlatformContext;
import com.networknext.dataflow.util.EntityIdHelpers;
import com.networknext.dataflow.util.bigquery.BigQuerySchemaUpdater;
import com.networknext.dataflow.util.dataflow.StripKeysFromTableRows;
import org.apache.beam.sdk.io.gcp.bigquery.BigQueryIO;
import org.apache.beam.sdk.io.gcp.bigquery.InsertRetryPolicy;
import org.apache.beam.sdk.schemas.Schema;
import org.apache.beam.sdk.transforms.Combine;
import org.apache.beam.sdk.transforms.DoFn;
import org.apache.beam.sdk.transforms.ParDo;
import org.apache.beam.sdk.transforms.windowing.FixedWindows;
import org.apache.beam.sdk.transforms.windowing.Window;
import org.apache.beam.sdk.values.KV;
import org.apache.beam.sdk.values.PCollection;
import org.apache.beam.sdk.values.Row;
import org.joda.time.Duration;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class SellerPipeline {

    private static final Logger LOG = LoggerFactory.getLogger(SellerPipeline.class);

    public static void buildSellerPipeline(PlatformContext context) {

        // seller aggregation pipeline

        Schema sellerSchema = Schema.builder().addStringField(SellerBq.SELLER_ID)
                .addStringField(SellerBq.RELAY_ID).addInt64Field(SellerBq.REVENUE_INGRESS)
                .addInt64Field(SellerBq.REVENUE_EGRESS).addInt64Field(SellerBq.TRAFFIC).build();

        PCollection<KV<String, Row>> sellerPipeline = context.pubSubBilling
                .apply("Extract seller data", ParDo.of(new DoFn<BillingEntry, KV<String, Row>>() {
                    private static final long serialVersionUID = -8435407610481132899L;

                    @ProcessElement
                    public void processElement(ProcessContext c) {
                        try {
                            BillingEntry input = c.element();

                            for (int i = 0; i < input.getRouteCount(); i++) {
                                BillingRouteHop hop = input.getRoute(i);
                                String key = String.format("%s#%s",
                                        EntityIdHelpers
                                                .getStringForNetworkNextStorage(hop.getSellerId()),
                                        EntityIdHelpers
                                                .getStringForNetworkNextStorage(hop.getRelayId()));
                                Row value = Row.withSchema(sellerSchema).addValues(
                                        EntityIdHelpers
                                                .getStringForNetworkNextStorage(hop.getSellerId()),
                                        EntityIdHelpers
                                                .getStringForNetworkNextStorage(hop.getRelayId()),
                                        hop.getPriceIngress(), hop.getPriceEgress(),
                                        input.getEnvelopeBytesUp() + input.getEnvelopeBytesDown())
                                        .build();
                                c.output(KV.of(key, value));
                            }
                        } catch (NumberFormatException e) {
                            LOG.error("failed to parse unsigned long ID: ", e);
                        }
                    }
                })).apply("Window supplier data",
                        Window.<KV<String, Row>>into(FixedWindows.of(Duration.standardMinutes(1))));

        // seller BigQuery output pipeline

        BigQuerySchemaUpdater.execute(context.bigQuerySeller, SellerBq.schema);

        sellerPipeline.apply("Sum seller aggregates", Combine.perKey(new SellerAccumFn()))
                .apply("Insert into seller BigQuery Table", BigQueryIO.<KV<String, TableRow>>write()
                        .withFormatFunction(new StripKeysFromTableRows()).to(context.bigQuerySeller)
                        .withSchema(SellerBq.schema)
                        .withTimePartitioning(new TimePartitioning().setRequirePartitionFilter(true)
                                .setType("DAY").setField(SellerBq.TIMESTAMP))
                        .withCreateDisposition(BigQueryIO.Write.CreateDisposition.CREATE_IF_NEEDED)
                        .withWriteDisposition(BigQueryIO.Write.WriteDisposition.WRITE_APPEND)
                        .withMethod(BigQueryIO.Write.Method.STREAMING_INSERTS)
                        .withFailedInsertRetryPolicy(InsertRetryPolicy.retryTransientErrors())
                        .withExtendedErrorInfo());

    }

}
