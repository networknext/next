package com.networknext.dataflow.platform.buyer;

import com.google.api.services.bigquery.model.TableFieldSchema;
import com.google.api.services.bigquery.model.TableSchema;
import com.google.common.collect.ImmutableList;

public class BuyerBq {

        public static final String TIMESTAMP = "timestamp";
        public static final String BUYER_ID = "buyerId";
        public static final String DATACENTER_ID = "datacenterId";
        public static final String REVENUE_SELLERS = "revenueSellers";
        public static final String REVENUE_NETWORK_NEXT = "revenueNetworkNext";
        public static final String TRAFFIC_NETWORK_NEXT_UP = "trafficNetworkNextUp";
        public static final String TRAFFIC_NETWORK_NEXT_DOWN = "trafficNetworkNextDown";
        public static final String TRAFFIC_DIRECT_UP = "trafficDirectUp";
        public static final String TRAFFIC_DIRECT_DOWN = "trafficDirectDown";
        public static final String SESSION_COUNT = "sessionCount";
        public static final String SECONDS_NEXT = "secondsNext";
        public static final String SECONDS_DIRECT = "secondsDirect";
        public static final String SECONDS_MEASURED = "secondsMeasured";
        public static final String SECONDS_0_TO_5_CU = "seconds0to5cu";
        public static final String SECONDS_5_TO_10_CU = "seconds5to10cu";
        public static final String SECONDS_10_TO_15_CU = "seconds10to15cu";
        public static final String SECONDS_15_TO_20_CU = "seconds15to20cu";
        public static final String SECONDS_20_TO_30_CU = "seconds20to30cu";
        public static final String SECONDS_30_PLUS_CU = "seconds30pluscu";

        public static final String IS_NEW_SESSION = "isNewSession";
        public static final String IMPROVEMENT = "improvement";

        public static TableSchema schema;

        static {
                schema = new TableSchema().setFields(ImmutableList.of(
                                new TableFieldSchema().setName(TIMESTAMP).setType("TIMESTAMP").setMode("REQUIRED"),
                                new TableFieldSchema().setName(BUYER_ID).setType("STRING").setMode("REQUIRED"),
                                new TableFieldSchema().setName(DATACENTER_ID).setType("STRING").setMode("REQUIRED"),
                                new TableFieldSchema().setName(REVENUE_SELLERS).setType("INT64").setMode("REQUIRED"),
                                new TableFieldSchema().setName(REVENUE_NETWORK_NEXT).setType("INT64")
                                                .setMode("REQUIRED"),
                                new TableFieldSchema().setName(TRAFFIC_NETWORK_NEXT_UP).setType("INT64")
                                                .setMode("REQUIRED"),
                                new TableFieldSchema().setName(TRAFFIC_NETWORK_NEXT_DOWN).setType("INT64")
                                                .setMode("REQUIRED"),
                                new TableFieldSchema().setName(TRAFFIC_DIRECT_UP).setType("INT64").setMode("REQUIRED"),
                                new TableFieldSchema().setName(TRAFFIC_DIRECT_DOWN).setType("INT64")
                                                .setMode("REQUIRED"),
                                new TableFieldSchema().setName(SESSION_COUNT).setType("INT64").setMode("REQUIRED"),
                                new TableFieldSchema().setName(SECONDS_NEXT).setType("INT64").setMode("REQUIRED"),
                                new TableFieldSchema().setName(SECONDS_DIRECT).setType("INT64").setMode("REQUIRED"),
                                new TableFieldSchema().setName(SECONDS_MEASURED).setType("INT64").setMode("REQUIRED"),
                                new TableFieldSchema().setName(SECONDS_0_TO_5_CU).setType("INT64").setMode("REQUIRED"),
                                new TableFieldSchema().setName(SECONDS_5_TO_10_CU).setType("INT64").setMode("REQUIRED"),
                                new TableFieldSchema().setName(SECONDS_10_TO_15_CU).setType("INT64")
                                                .setMode("REQUIRED"),
                                new TableFieldSchema().setName(SECONDS_15_TO_20_CU).setType("INT64")
                                                .setMode("REQUIRED"),
                                new TableFieldSchema().setName(SECONDS_20_TO_30_CU).setType("INT64")
                                                .setMode("REQUIRED"),
                                new TableFieldSchema().setName(SECONDS_30_PLUS_CU).setType("INT64")
                                                .setMode("REQUIRED")));
        }

}