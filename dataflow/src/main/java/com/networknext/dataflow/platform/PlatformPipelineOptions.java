package com.networknext.dataflow.platform;

import org.apache.beam.runners.dataflow.options.DataflowPipelineOptions;

public interface PlatformPipelineOptions extends DataflowPipelineOptions {

    String getBillingPubsubSubscription();

    void setBillingPubsubSubscription(String subscription);

    String getBillingPubsubTopic();

    void setBillingPubsubTopic(String topic);

    String getBigQueryDataset();

    void setBigQueryDataset(String dataset);

    String getBigQueryBillingTableId();

    void setBigQueryBillingTableId(String id);

}
