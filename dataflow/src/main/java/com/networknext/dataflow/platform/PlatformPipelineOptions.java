package com.networknext.dataflow.platform;

import org.apache.beam.runners.dataflow.options.DataflowPipelineOptions;

public interface PlatformPipelineOptions extends DataflowPipelineOptions {

    String getKeyDbSessionListHost();

    void setKeyDbSessionListHost(String hostname);

    int getKeyDbSessionListPort();

    void setKeyDbSessionListPort(int port);

    String getKeyDbSessionToolHost();

    void setKeyDbSessionToolHost(String hostname);

    int getKeyDbSessionToolPort();

    void setKeyDbSessionToolPort(int port);

    String getDisableBigTable();

    void setDisableBigTable(String disableBigTable);

    String getBillingPubsubSubscription();

    void setBillingPubsubSubscription(String subscription);

    String getBillingPubsubTopic();

    void setBillingPubsubTopic(String topic);

    String getTrafficStatsPubsubSubscription();

    void setTrafficStatsPubsubSubscription(String subscription);

    String getTrafficStatsPubsubTopic();

    void setTrafficStatsPubsubTopic(String topic);

    String getBigTableInstanceId();

    void setBigTableInstanceId(String id);

    String getBigTableBillingTableId();

    void setBigTableBillingTableId(String id);

    String getBigTableSessionTableId();

    void setBigTableSessionTableId(String id);

    String getBigTableUserTableId();

    void setBigTableUserTableId(String id);

    String getBigQueryDataset();

    void setBigQueryDataset(String dataset);

    String getBigQueryBillingTableId();

    void setBigQueryBillingTableId(String id); // billing

    String getBigQueryBuyerTableId();

    void setBigQueryBuyerTableId(String id); // buyer

    String getBigQuerySellerTableId();

    void setBigQuerySellerTableId(String id); // seller

    String getBigQueryTrafficStatsTableId();

    void setBigQueryTrafficStatsTableId(String id); // traffic_stats

}
