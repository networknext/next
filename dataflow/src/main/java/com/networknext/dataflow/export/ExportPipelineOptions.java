package com.networknext.dataflow.export;

import org.apache.beam.runners.dataflow.options.DataflowPipelineOptions;

public interface ExportPipelineOptions extends DataflowPipelineOptions {

    String getBillingPubsubSubscription();

    void setBillingPubsubSubscription(String subscription);

    String getBillingPubsubTopic();

    void setBillingPubsubTopic(String topic);

    String getSellerExportConfig();

    void setSellerExportConfig(String config); // pubsub=...,parquet=...;pubsub=...,parquet=...;...

}
