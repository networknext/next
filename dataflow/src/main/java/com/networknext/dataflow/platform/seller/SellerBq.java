package com.networknext.dataflow.platform.seller;

import com.google.api.services.bigquery.model.TableFieldSchema;
import com.google.api.services.bigquery.model.TableSchema;
import com.google.common.collect.ImmutableList;

public class SellerBq {

    public static final String TIMESTAMP = "timestamp";
    public static final String SELLER_ID = "sellerId";
    public static final String RELAY_ID = "relayId";
    public static final String REVENUE_INGRESS = "revenueIngress";
    public static final String REVENUE_EGRESS = "revenueEgress";
    public static final String TRAFFIC = "traffic";

    public static TableSchema schema;

    static {
        schema = new TableSchema().setFields(ImmutableList.of(
                new TableFieldSchema().setName(TIMESTAMP).setType("TIMESTAMP").setMode("REQUIRED"),
                new TableFieldSchema().setName(SELLER_ID).setType("STRING").setMode("REQUIRED"),
                new TableFieldSchema().setName(RELAY_ID).setType("STRING").setMode("REQUIRED"),
                new TableFieldSchema().setName(REVENUE_INGRESS).setType("INT64")
                        .setMode("REQUIRED"),
                new TableFieldSchema().setName(REVENUE_EGRESS).setType("INT64").setMode("REQUIRED"),
                new TableFieldSchema().setName(TRAFFIC).setType("INT64").setMode("REQUIRED")));
    }

}
