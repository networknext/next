package com.networknext.dataflow.util.bigquery;

import java.util.ArrayList;
import java.util.List;
import com.google.api.services.bigquery.model.TableFieldSchema;
import com.google.api.services.bigquery.model.TableReference;
import com.google.api.services.bigquery.model.TableSchema;
import com.google.cloud.bigquery.BigQuery;
import com.google.cloud.bigquery.BigQueryOptions;
import com.google.cloud.bigquery.Field;
import com.google.cloud.bigquery.LegacySQLTypeName;
import com.google.cloud.bigquery.Schema;
import com.google.cloud.bigquery.StandardSQLTypeName;
import com.google.cloud.bigquery.StandardTableDefinition;
import com.google.cloud.bigquery.Table;
import com.google.cloud.bigquery.Field.Mode;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class BigQuerySchemaUpdater {
    private static final Logger LOG = LoggerFactory.getLogger(BigQuerySchemaUpdater.class);

    public static void execute(TableReference tableReference, TableSchema tableSchema) {
        LOG.info("Updating schema for BigQuery table '" + tableReference.getTableId() + "'...");

        BigQuery bigquery = BigQueryOptions.newBuilder().setProjectId(tableReference.getProjectId())
                .build().getService();

        Table table = bigquery.getTable(tableReference.getDatasetId(), tableReference.getTableId());
        if (table == null) {
            LOG.info("BigQuery table '" + tableReference.getTableId()
                    + "' does not yet exist, letting Dataflow create it.");
            return;
        }

        StandardTableDefinition definition = table.<StandardTableDefinition>getDefinition()
                .toBuilder().setSchema(convertSchema(tableSchema)).build();

        try {
            table.toBuilder().setDefinition(definition).build().update();
            LOG.info("Successfully updated schema for BigQuery table '"
                    + tableReference.getTableId() + "'.");
        } catch (Exception e) {
            LOG.warn("Unable to update schema for BigQuery table '" + tableReference.getTableId()
                    + "': " + e.toString());
        }
    }

    private static Schema convertSchema(TableSchema schema) {
        List<Field> fields = new ArrayList<Field>();
        for (TableFieldSchema field : schema.getFields()) {
            fields.add(convertField(field));
        }
        return Schema.of(fields);
    }

    private static Field convertField(TableFieldSchema field) {
        List<Field> subfields = new ArrayList<Field>();
        if (field.getFields() != null) {
            for (TableFieldSchema subfield : field.getFields()) {
                subfields.add(convertField(subfield));
            }
        }

        StandardSQLTypeName typeName;
        if (field.getType().equals("RECORD")) {
            typeName = StandardSQLTypeName.STRUCT;
        } else if (field.getType().equals("INTEGER")) {
            typeName = StandardSQLTypeName.INT64;
        } else if (field.getType().equals("FLOAT")) {
            typeName = StandardSQLTypeName.FLOAT64;
        } else {
            typeName = StandardSQLTypeName.valueOf(field.getType());
        }

        return Field.newBuilder(field.getName(), typeName, subfields.toArray(new Field[0]))
                .setMode(Mode.valueOf(field.getMode())).setDescription(field.getDescription())
                .build();
    }
}
