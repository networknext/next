package com.networknext.dataflow.platform;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.google.api.services.bigquery.model.TableReference;
import com.networknext.api.relay_data.RelayData.RelayTrafficStats;
import com.networknext.api.router.Router.BillingEntry;

import org.apache.beam.sdk.values.PCollection;

public class PlatformContext {

    public ObjectMapper json;

    public PCollection<BillingEntry> pubSubBilling;

    public TableReference bigQueryBilling;

    public String sellerExportConfig;

}
