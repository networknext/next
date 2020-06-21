package com.networknext.dataflow.export;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.networknext.api.router.Router.BillingEntry;
import org.apache.beam.sdk.values.PCollection;

public class ExportContext {

    public ObjectMapper json;

    public PCollection<BillingEntry> pubSubBilling;

    public String sellerExportConfig;

}
