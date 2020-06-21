package com.networknext.dataflow.util.dataflow;

import com.google.api.services.bigquery.model.TableRow;
import org.apache.beam.sdk.transforms.SerializableFunction;
import org.apache.beam.sdk.values.KV;

public class StripKeysFromTableRows
        implements SerializableFunction<KV<String, TableRow>, TableRow> {
    private static final long serialVersionUID = -4919787039912560457L;

    @Override
    public TableRow apply(KV<String, TableRow> entry) {
        return entry.getValue();
    }
}
