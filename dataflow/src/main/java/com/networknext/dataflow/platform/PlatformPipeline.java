package com.networknext.dataflow.platform;

import java.io.BufferedReader;
import java.io.FileReader;
import java.io.IOException;
import java.util.ArrayList;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.google.api.services.bigquery.model.TableReference;
import com.networknext.api.router.Router.BillingEntry;
import com.networknext.dataflow.platform.billing.BillingPipeline;
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

        PlatformContext context = new PlatformContext();
        context.json = new ObjectMapper();

        context.bigQueryBilling = new TableReference().setProjectId(options.getProject())
                .setDatasetId(options.getBigQueryDataset())
                .setTableId(options.getBigQueryBillingTableId());
 
        // now start adding all the pipelines to the root pipeline
        BillingPipeline.buildBillingPipeline(context);
 
        // run the pipeline
        pipeline.run();

    }

}
