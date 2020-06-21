package com.networknext.dataflow.util.bigtable;

import java.nio.ByteBuffer;
import java.nio.charset.StandardCharsets;

import org.apache.hadoop.hbase.client.Put;

public class BigTableOp {

    public static Put SetBytes(Put mutation, String family, String qualifier, byte[] value) {
        return mutation.addColumn(family.getBytes(StandardCharsets.UTF_8), qualifier.getBytes(StandardCharsets.UTF_8),
                value);
    }

    public static Put SetUint64(Put mutation, String family, String qualifier, long unsignedValue) {
        return SetBytes(mutation, family, qualifier, ByteBuffer.allocate(8).putLong(unsignedValue).array());
    }

    public static Put SetFloat64(Put mutation, String family, String qualifier, Double value) {
        return SetUint64(mutation, family, qualifier, Double.doubleToLongBits(value));
    }
}