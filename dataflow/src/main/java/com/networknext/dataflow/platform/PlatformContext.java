package com.networknext.dataflow.platform;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.google.api.services.bigquery.model.TableReference;
import com.google.cloud.bigtable.beam.CloudBigtableTableConfiguration;
import com.networknext.api.relay_data.RelayData.RelayTrafficStats;
import com.networknext.api.router.Router.BillingEntry;

import org.apache.beam.sdk.values.PCollection;

public class PlatformContext {

    public ObjectMapper json;

    public PCollection<BillingEntry> pubSubBilling;

    public PCollection<RelayTrafficStats> pubSubTrafficStats;

    public CloudBigtableTableConfiguration bigTableBilling;

    public CloudBigtableTableConfiguration bigTableSession;

    public CloudBigtableTableConfiguration bigTableUser;

    public TableReference bigQueryBilling;

    public TableReference bigQueryBuyer;

    public TableReference bigQuerySeller;

    public TableReference bigQueryTrafficStats;

    public String sellerExportConfig;

    public String keyDbSessionListHost;

    public int keyDbSessionListPort;

    public String keyDbSessionToolHost;

    public int keyDbSessionToolPort;

}
