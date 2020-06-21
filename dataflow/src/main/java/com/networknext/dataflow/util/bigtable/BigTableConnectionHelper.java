package com.networknext.dataflow.util.bigtable;

import java.io.IOException;

import com.google.cloud.bigtable.hbase.BigtableConfiguration;

import org.apache.hadoop.hbase.HColumnDescriptor;
import org.apache.hadoop.hbase.HTableDescriptor;
import org.apache.hadoop.hbase.TableExistsException;
import org.apache.hadoop.hbase.TableName;
import org.apache.hadoop.hbase.client.Admin;
import org.apache.hadoop.hbase.client.Connection;

import com.google.cloud.bigtable.beam.CloudBigtableTableConfiguration;

public class BigTableConnectionHelper {

    private static CloudBigtableTableConfiguration.Builder addBigtableEmulatorConfiguration(
            CloudBigtableTableConfiguration.Builder builder, String host, String port) {
        return builder.withConfiguration("google.bigtable.instance.admin.endpoint.host", host)
                .withConfiguration("google.bigtable.admin.endpoint.host", host)
                .withConfiguration("google.bigtable.endpoint.host", host)
                .withConfiguration("google.bigtable.endpoint.port", port)
                .withConfiguration("google.bigtable.use.plaintext.negotiation", "true")
                .withConfiguration("google.bigtable.grpc.read.partial.row.timeout.ms", "5000");
    }

    private static void createTable(Admin admin, String tableId, String[] columnFamilies) throws IOException {
        HTableDescriptor tableDesc = new HTableDescriptor(TableName.valueOf(tableId));
        for (String family : columnFamilies) {
            tableDesc.addFamily(new HColumnDescriptor(family));
        }
        try {
            System.out.println(String.format("creating BigTable table '%s'...", tableId));
            admin.createTable(tableDesc);
        } catch (TableExistsException e) {
            System.out.println(String.format("BigTable table '%s' exists already", tableId));
        }
    }

    public static CloudBigtableTableConfiguration connectToBigTable(String project, String instanceId, String tableId,
            String[] columnFamilies) throws Exception {

        CloudBigtableTableConfiguration.Builder configBuilder = new CloudBigtableTableConfiguration.Builder()
                .withProjectId(project).withInstanceId(instanceId).withTableId(tableId);

        String bigtableEmulatorHost = System.getenv("BIGTABLE_EMULATOR_HOST");
        if (bigtableEmulatorHost != null) {
            String[] parts = bigtableEmulatorHost.split(":");
            String host = parts[0];
            String port = parts[1];
            configBuilder = addBigtableEmulatorConfiguration(configBuilder, host, port);
        }

        try (Connection connection = BigtableConfiguration.connect(project, instanceId)) {
            Admin admin = connection.getAdmin();
            createTable(admin, tableId, columnFamilies);
        }

        return configBuilder.build();

    }

}