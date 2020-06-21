package com.networknext.dataflow.platform.buyer;

import com.google.api.services.bigquery.model.TableRow;
import com.google.api.services.bigquery.model.TimePartitioning;
import com.networknext.api.router.Router.BillingEntry;
import com.networknext.api.router.Router.BillingRouteHop;
import com.networknext.dataflow.platform.PlatformConstants;
import com.networknext.dataflow.platform.PlatformContext;
import com.networknext.dataflow.util.EntityIdHelpers;
import com.networknext.dataflow.util.bigquery.BigQuerySchemaUpdater;
import com.networknext.dataflow.util.dataflow.StripKeysFromTableRows;
import org.apache.beam.sdk.io.gcp.bigquery.BigQueryIO;
import org.apache.beam.sdk.io.gcp.bigquery.InsertRetryPolicy;
import org.apache.beam.sdk.schemas.Schema;
import org.apache.beam.sdk.transforms.Combine;
import org.apache.beam.sdk.transforms.DoFn;
import org.apache.beam.sdk.transforms.GroupByKey;
import org.apache.beam.sdk.transforms.ParDo;
import org.apache.beam.sdk.transforms.windowing.FixedWindows;
import org.apache.beam.sdk.transforms.windowing.Window;
import org.apache.beam.sdk.values.KV;
import org.apache.beam.sdk.values.PCollection;
import org.apache.beam.sdk.values.Row;
import org.joda.time.Duration;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class BuyerPipeline {

    private static final Logger LOG = LoggerFactory.getLogger(BuyerPipeline.class);

    public static void buildBuyerPipeline(PlatformContext context) {

        // buyer aggregation pipeline

        Schema buyerSchema = Schema.builder().addStringField(BuyerBq.BUYER_ID)
                .addStringField(BuyerBq.DATACENTER_ID).addInt64Field(BuyerBq.REVENUE_SELLERS)
                .addInt64Field(BuyerBq.REVENUE_NETWORK_NEXT)
                .addInt64Field(BuyerBq.TRAFFIC_NETWORK_NEXT_UP)
                .addInt64Field(BuyerBq.TRAFFIC_NETWORK_NEXT_DOWN)
                .addInt64Field(BuyerBq.TRAFFIC_DIRECT_UP).addInt64Field(BuyerBq.TRAFFIC_DIRECT_DOWN)
                .addBooleanField(BuyerBq.IS_NEW_SESSION).addInt64Field(BuyerBq.SECONDS_NEXT)
                .addInt64Field(BuyerBq.SECONDS_DIRECT).addInt64Field(BuyerBq.SECONDS_MEASURED)
                .addDoubleField(BuyerBq.IMPROVEMENT).build();

        PCollection<KV<String, Row>> buyerPipeline = context.pubSubBilling
                .apply("Extract buyer data", ParDo.of(new DoFn<BillingEntry, KV<String, Row>>() {
                    private static final long serialVersionUID = -5151010713434959555L;

                    @ProcessElement
                    public void processElement(ProcessContext c) {
                        try {
                            BillingEntry input = c.element();

                            long revenueSellers = 0;
                            long revenueNetworkNext = 0;
                            long trafficNetworkNextUp = 0;
                            long trafficNetworkNextDown = 0;
                            long trafficDirectUp = 0;
                            long trafficDirectDown = 0;
                            long secondsNext = 0;
                            long secondsDirect = 0;
                            long secondsMeasured = 0;
                            double improvement = 0;

                            if (input.getRouteCount() == 0) {
                                trafficDirectUp = input.getUsageBytesUp();
                                trafficDirectDown = input.getUsageBytesDown();
                                secondsDirect += input.getDuration();
                            } else {
                                secondsNext += input.getDuration();
                                trafficNetworkNextUp = input.getUsageBytesUp();
                                trafficNetworkNextDown = input.getUsageBytesDown();

                                long networkNextPriceIngress =
                                        (PlatformConstants.NetworkNextCutNibblinsPerGB
                                                * input.getEnvelopeBytesUp())
                                                / PlatformConstants.BytesPerGB;
                                long networkNextPriceEgress =
                                        (PlatformConstants.NetworkNextCutNibblinsPerGB
                                                * input.getEnvelopeBytesDown())
                                                / PlatformConstants.BytesPerGB;

                                revenueNetworkNext =
                                        networkNextPriceIngress + networkNextPriceEgress;

                                for (int i = 0; i < input.getRouteCount(); i++) {
                                    BillingRouteHop hop = input.getRoute(i);
                                    revenueSellers += hop.getPriceIngress() + hop.getPriceEgress();
                                }
                            }

                            if (input.getRequest().getNextRtt() > 0.0
                                    && input.getRequest().getDirectRtt() > 0.0) {
                                secondsMeasured += PlatformConstants.BillingSliceSeconds;
                                improvement = input.getRequest().getDirectRtt()
                                        - input.getRequest().getNextRtt();
                            }

                            String key = String.format("%s#%s",
                                    EntityIdHelpers.getStringForNetworkNextStorage(
                                            input.getRequest().getBuyerId()),
                                    EntityIdHelpers.getStringForNetworkNextStorage(
                                            input.getRequest().getDatacenterId()));

                            Row value = Row.withSchema(buyerSchema).addValues(
                                    EntityIdHelpers.getStringForNetworkNextStorage(
                                            input.getRequest().getBuyerId()),
                                    EntityIdHelpers.getStringForNetworkNextStorage(
                                            input.getRequest().getDatacenterId()),
                                    revenueSellers, revenueNetworkNext, trafficNetworkNextUp,
                                    trafficNetworkNextDown, trafficDirectUp, trafficDirectDown,
                                    input.getInitial(), secondsNext, secondsDirect, secondsMeasured,
                                    improvement).build();

                            c.output(KV.of(key, value));
                        } catch (NumberFormatException e) {
                            LOG.error("failed to parse unsigned long ID: ", e);
                        }
                    }
                })).apply("Window buyer data",
                        Window.<KV<String, Row>>into(FixedWindows.of(Duration.standardMinutes(1))));

        // buyer BigQuery output pipeline

        BigQuerySchemaUpdater.execute(context.bigQueryBuyer, BuyerBq.schema);

        buyerPipeline.apply("Sum buyer aggregates", Combine.perKey(new BuyerAccumFn()))
                .apply("Insert into buyer BigQuery Table", BigQueryIO.<KV<String, TableRow>>write()
                        .withFormatFunction(new StripKeysFromTableRows()).to(context.bigQueryBuyer)
                        .withSchema(BuyerBq.schema)
                        .withTimePartitioning(new TimePartitioning().setRequirePartitionFilter(true)
                                .setType("DAY").setField(BuyerBq.TIMESTAMP))
                        .withCreateDisposition(BigQueryIO.Write.CreateDisposition.CREATE_IF_NEEDED)
                        .withWriteDisposition(BigQueryIO.Write.WriteDisposition.WRITE_APPEND)
                        .withMethod(BigQueryIO.Write.Method.STREAMING_INSERTS)
                        .withFailedInsertRetryPolicy(InsertRetryPolicy.retryTransientErrors())
                        .withExtendedErrorInfo());

    }

}
