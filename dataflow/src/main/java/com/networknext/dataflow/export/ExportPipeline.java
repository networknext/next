package com.networknext.dataflow.export;

import java.io.BufferedReader;
import java.io.FileReader;
import java.io.IOException;
import java.util.ArrayList;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.networknext.api.router.Router.BillingEntry;
import com.networknext.dataflow.export.sellerexport.SellerExportPipeline;
import com.networknext.dataflow.util.pubsub.PubSubSubscriptionAutoCreate;
import org.apache.beam.sdk.Pipeline;
import org.apache.beam.sdk.io.gcp.pubsub.PubsubIO;
import org.apache.beam.sdk.io.gcp.pubsub.PubsubMessage;
import org.apache.beam.sdk.options.PipelineOptionsFactory;
import org.apache.beam.sdk.transforms.DoFn;
import org.apache.beam.sdk.transforms.ParDo;

public class ExportPipeline {

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

        PipelineOptionsFactory.register(ExportPipelineOptions.class);

        String argumentsFile = System.getenv("PIPELINE_ARGUMENTS_FILE");
        if (argumentsFile != null) {
            args = readFileLineByLine(argumentsFile);
        }

        for (int i = 0; i < args.length; i++) {
            System.out.println(args[i]);
        }

        ExportPipelineOptions options = PipelineOptionsFactory.fromArgs(args).withValidation()
                .as(ExportPipelineOptions.class);

        Pipeline pipeline = Pipeline.create(options);

        PubSubSubscriptionAutoCreate.execute(options, options.getBillingPubsubTopic(),
                options.getBillingPubsubSubscription());

        ExportContext context = new ExportContext();
        context.json = new ObjectMapper();

        // connect to all the external endpoints for all the pipelines
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
        context.sellerExportConfig = options.getSellerExportConfig();

        // now start adding all the pipelines to the root pipeline
        SellerExportPipeline.buildSellerExportPipeline(context);

        // run the pipeline
        pipeline.run();

    }

}
