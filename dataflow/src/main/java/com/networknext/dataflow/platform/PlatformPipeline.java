package com.networknext.dataflow.platform;

import java.io.BufferedReader;
import java.io.FileReader;
import java.io.IOException;
import java.util.ArrayList;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.google.api.services.bigquery.model.TableReference;
import com.networknext.api.relay_data.RelayData.RelayTrafficStats;
import com.networknext.api.router.Router.BillingEntry;
import com.networknext.dataflow.platform.billing.BillingPipeline;
import com.networknext.dataflow.platform.buyer.BuyerPipeline;
import com.networknext.dataflow.platform.seller.SellerPipeline;
import com.networknext.dataflow.platform.session.SessionPipeline;
import com.networknext.dataflow.platform.trafficstats.TrafficStatsPipeline;
import com.networknext.dataflow.util.bigtable.BigTableConnectionHelper;
import com.networknext.dataflow.util.pubsub.PubSubSubscriptionAutoCreate;

import org.apache.beam.sdk.Pipeline;
import org.apache.beam.sdk.io.gcp.pubsub.PubsubIO;
import org.apache.beam.sdk.io.gcp.pubsub.PubsubMessage;
import org.apache.beam.sdk.options.PipelineOptionsFactory;
import org.apache.beam.sdk.transforms.DoFn;
import org.apache.beam.sdk.transforms.ParDo;
import org.joda.time.Instant;

public class PlatformPipeline {

    private static String[] readFileLineByLine(String path) throws IOException {
        System.out.println("Reading arguments from " + path);
        ArrayList<String> arr = new ArrayList<String>();
        BufferedReader reader = new BufferedReader(new FileReader(path));
        String line = reader.readLine();
        while (line != null) {
            arr.add(line);
            line = reader.readLine();
        }
        reader.close();
        System.out.println("Finished reading arguments from " + path);
        return arr.toArray(new String[0]);
    }

    public static void main(String[] args) throws Exception {

        PipelineOptionsFactory.register(PlatformPipelineOptions.class);

        String argumentsFile = System.getenv("PIPELINE_ARGUMENTS_FILE");
        if (argumentsFile != null) {
            args = readFileLineByLine(argumentsFile);
        }

        for (int i = 0; i < args.length; i++) {
            System.out.println(args[i]);
        }

        PlatformPipelineOptions options = PipelineOptionsFactory.fromArgs(args).withValidation()
                .as(PlatformPipelineOptions.class);

        Pipeline pipeline = Pipeline.create(options);

        PubSubSubscriptionAutoCreate.execute(options, options.getBillingPubsubTopic(),
                options.getBillingPubsubSubscription());
        PubSubSubscriptionAutoCreate.execute(options, options.getTrafficStatsPubsubTopic(),
                options.getTrafficStatsPubsubSubscription());

        PlatformContext context = new PlatformContext();
        context.json = new ObjectMapper();

        // connect to all the external endpoints for all the pipelines
        context.keyDbSessionListHost = options.getKeyDbSessionListHost();
        context.keyDbSessionListPort = options.getKeyDbSessionListPort();
        context.keyDbSessionToolHost = options.getKeyDbSessionToolHost();
        context.keyDbSessionToolPort = options.getKeyDbSessionToolPort();
        context.pubSubBilling = pipeline
                .apply("Read billing Pub/Sub",
                        PubsubIO.readMessagesWithAttributes()
                                .fromSubscription(options.getBillingPubsubSubscription()))
                .apply("Extract billing data from Pub/Sub",
                        ParDo.of(new DoFn<PubsubMessage, BillingEntry>() {
                            private static final long serialVersionUID = 6922898617241332266L;

                            @ProcessElement
                            public void processElement(ProcessContext c) throws Exception {
                                c.output(BillingEntry.parseFrom(c.element().getPayload()));
                            }
                        }));
        context.pubSubTrafficStats = pipeline
                .apply("Read traffic stats Pub/Sub",
                        PubsubIO.readMessagesWithAttributes()
                                .fromSubscription(options.getTrafficStatsPubsubSubscription()))
                .apply("Extract traffic stats from Pub/Sub",
                        ParDo.of(new DoFn<PubsubMessage, RelayTrafficStats>() {
                            private static final long serialVersionUID = 6922898617241332266L;

                            @ProcessElement
                            public void processElement(ProcessContext c) throws Exception {
                                c.output(RelayTrafficStats.parseFrom(c.element().getPayload()));
                            }
                        }));
        if (!options.getDisableBigTable().equals("true")) {
            context.bigTableBilling = BigTableConnectionHelper.connectToBigTable(
                    options.getProject(), options.getBigTableInstanceId(),
                    options.getBigTableBillingTableId(), new String[] {"data"});
            context.bigTableSession = BigTableConnectionHelper.connectToBigTable(
                    options.getProject(), options.getBigTableInstanceId(),
                    options.getBigTableSessionTableId(), new String[] {"flow"});
            context.bigTableUser = BigTableConnectionHelper.connectToBigTable(options.getProject(),
                    options.getBigTableInstanceId(), options.getBigTableUserTableId(),
                    new String[] {"flow"});
        }
        context.bigQueryBilling = new TableReference().setProjectId(options.getProject())
                .setDatasetId(options.getBigQueryDataset())
                .setTableId(options.getBigQueryBillingTableId());
        context.bigQueryBuyer = new TableReference().setProjectId(options.getProject())
                .setDatasetId(options.getBigQueryDataset())
                .setTableId(options.getBigQueryBuyerTableId());
        context.bigQuerySeller = new TableReference().setProjectId(options.getProject())
                .setDatasetId(options.getBigQueryDataset())
                .setTableId(options.getBigQuerySellerTableId());
        context.bigQueryTrafficStats = new TableReference().setProjectId(options.getProject())
                .setDatasetId(options.getBigQueryDataset())
                .setTableId(options.getBigQueryTrafficStatsTableId());

        // now start adding all the pipelines to the root pipeline
        BillingPipeline.buildBillingPipeline(context);
        BuyerPipeline.buildBuyerPipeline(context);
        SellerPipeline.buildSellerPipeline(context);
        TrafficStatsPipeline.buildTrafficStatsPipeline(context);
        SessionPipeline.buildSessionPipeline(context);

        // run the pipeline
        pipeline.run();

    }

}
