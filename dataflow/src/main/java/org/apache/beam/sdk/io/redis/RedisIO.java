/*
 * Licensed to the Apache Software Foundation (ASF) under one or more contributor license
 * agreements. See the NOTICE file distributed with this work for additional information regarding
 * copyright ownership. The ASF licenses this file to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance with the License. You may obtain a
 * copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 */
package org.apache.beam.sdk.io.redis;

import static org.apache.beam.vendor.guava.v26_0_jre.com.google.common.base.Preconditions.checkArgument;

import com.google.auto.value.AutoValue;
import java.io.UnsupportedEncodingException;
import java.util.List;
import java.util.concurrent.atomic.AtomicInteger;
import javax.annotation.Nullable;
import org.apache.beam.sdk.annotations.Experimental;
import org.apache.beam.sdk.coders.ByteArrayCoder;
import org.apache.beam.sdk.coders.KvCoder;
import org.apache.beam.sdk.coders.StringUtf8Coder;
import org.apache.beam.sdk.transforms.Create;
import org.apache.beam.sdk.transforms.DoFn;
import org.apache.beam.sdk.transforms.Filter;
import org.apache.beam.sdk.transforms.PTransform;
import org.apache.beam.sdk.transforms.ParDo;
import org.apache.beam.sdk.transforms.Reshuffle;
import org.apache.beam.sdk.transforms.SerializableFunctions;
import org.apache.beam.sdk.transforms.View;
import org.apache.beam.sdk.transforms.display.DisplayData;
import org.apache.beam.sdk.transforms.windowing.BoundedWindow;
import org.apache.beam.sdk.values.KV;
import org.apache.beam.sdk.values.PBegin;
import org.apache.beam.sdk.values.PCollection;
import org.apache.beam.sdk.values.PCollectionView;
import org.apache.beam.sdk.values.PDone;
import org.apache.beam.vendor.guava.v26_0_jre.com.google.common.collect.ArrayListMultimap;
import org.apache.beam.vendor.guava.v26_0_jre.com.google.common.collect.Multimap;
import redis.clients.jedis.BinaryJedis;
import redis.clients.jedis.Jedis;
import redis.clients.jedis.Pipeline;
import redis.clients.jedis.ScanParams;
import redis.clients.jedis.ScanResult;

/**
 * An IO to manipulate Redis key/value database.
 *
 * <h3>Reading Redis key/value pairs</h3>
 *
 * <p>
 * {@link #read()} provides a source which returns a bounded {@link PCollection} containing
 * key/value pairs as {@code KV<byte[], byte[]>}.
 *
 * <p>
 * To configure a Redis source, you have to provide Redis server hostname and port number.
 * Optionally, you can provide a key pattern (to filter the keys). The following example illustrates
 * how to configure a source:
 *
 * <pre>
 * {@code
 * pipeline.apply(RedisIO.read()
 *   .withEndpoint("::1", 6379)
 *   .withKeyPattern("foo*"))
 *
 * }
 * </pre>
 *
 * <p>
 * It's also possible to specify Redis authentication and connection timeout with the corresponding
 * methods:
 *
 * <pre>
 * {@code
 * pipeline.apply(RedisIO.read()
 *   .withEndpoint("::1", 6379)
 *   .withAuth("authPassword")
 *   .withTimeout(60000)
 *   .withKeyPattern("foo*"))
 *
 * }
 * </pre>
 *
 * <p>
 * {@link #readAll()} can be used to request Redis server using input PCollection elements as key
 * pattern (as String).
 *
 * <pre>
 * {@code
 * pipeline.apply(...)
 *    // here we have a PCollection<String> with the key patterns
 *    .apply(RedisIO.readAll().withEndpoint("::1", 6379))
 *   // here we have a PCollection<KV<String,String>>
 *
 * }
 * </pre>
 *
 * <h3>Writing Redis key/value pairs</h3>
 *
 * <p>
 * {@link #write()} provides a sink to write key/value pairs represented as {@link KV} from an
 * incoming {@link PCollection}.
 *
 * <p>
 * To configure the target Redis server, you have to provide Redis server hostname and port number.
 * The following example illustrates how to configure a sink:
 *
 * <pre>
 * {@code
 * pipeline.apply(...)
 *   // here we a have a PCollection<String, String> with key/value pairs
 *   .apply(RedisIO.write().withEndpoint("::1", 6379))
 *
 * }
 * </pre>
 */
@Experimental(Experimental.Kind.SOURCE_SINK)
public class RedisIO {

    /** Read data from a Redis server. */
    public static Read read() {
        return new AutoValue_RedisIO_Read.Builder()
                .setConnectionConfiguration(RedisConnectionConfiguration.create())
                .setKeyPattern("*".getBytes()).setBatchSize(1000).build();
    }

    /**
     * Like {@link #read()} but executes multiple instances of the Redis query substituting each
     * element of a {@link PCollection} as key pattern.
     */
    public static ReadAll readAll() {
        return new AutoValue_RedisIO_ReadAll.Builder()
                .setConnectionConfiguration(RedisConnectionConfiguration.create())
                .setBatchSize(1000).build();
    }

    /** Write data to a Redis server. */
    public static Write write() {
        return new AutoValue_RedisIO_Write.Builder()
                .setConnectionConfiguration(RedisConnectionConfiguration.create())
                .setMethod(Write.Method.APPEND).build();
    }

    private RedisIO() {
    }

    /** Implementation of {@link #read()}. */
    @AutoValue
    public abstract static class Read extends PTransform<PBegin, PCollection<KV<byte[], byte[]>>> {

        @Nullable
        abstract RedisConnectionConfiguration connectionConfiguration();

        @Nullable
        abstract byte[] keyPattern();

        abstract int batchSize();

        abstract Builder builder();

        @AutoValue.Builder
        abstract static class Builder {
            @Nullable
            abstract Builder setConnectionConfiguration(RedisConnectionConfiguration connection);

            @Nullable
            abstract Builder setKeyPattern(byte[] keyPattern);

            abstract Builder setBatchSize(int batchSize);

            abstract Read build();
        }

        public Read withEndpoint(String host, int port) {
            checkArgument(host != null, "host can not be null");
            checkArgument(0 < port && port < 65536,
                    "port must be a positive integer less than 65536");
            return builder().setConnectionConfiguration(
                    connectionConfiguration().withHost(host).withPort(port)).build();
        }

        public Read withAuth(String auth) {
            checkArgument(auth != null, "auth can not be null");
            return builder().setConnectionConfiguration(connectionConfiguration().withAuth(auth))
                    .build();
        }

        public Read withTimeout(int timeout) {
            checkArgument(timeout >= 0, "timeout can not be negative");
            return builder()
                    .setConnectionConfiguration(connectionConfiguration().withTimeout(timeout))
                    .build();
        }

        public Read withKeyPattern(byte[] keyPattern) {
            checkArgument(keyPattern != null, "keyPattern can not be null");
            return builder().setKeyPattern(keyPattern).build();
        }

        public Read withConnectionConfiguration(RedisConnectionConfiguration connection) {
            checkArgument(connection != null, "connection can not be null");
            return builder().setConnectionConfiguration(connection).build();
        }

        public Read withBatchSize(int batchSize) {
            return builder().setBatchSize(batchSize).build();
        }

        @Override
        public void populateDisplayData(DisplayData.Builder builder) {
            connectionConfiguration().populateDisplayData(builder);
        }

        @Override
        public PCollection<KV<byte[], byte[]>> expand(PBegin input) {
            checkArgument(connectionConfiguration() != null,
                    "withConnectionConfiguration() is required");

            return input.apply(Create.of(keyPattern()))
                    .apply(ParDo.of(new ReadKeysWithPattern(connectionConfiguration())))
                    .apply(RedisIO.readAll().withConnectionConfiguration(connectionConfiguration())
                            .withBatchSize(batchSize()));
        }
    }

    /** Implementation of {@link #readAll()}. */
    @AutoValue
    public abstract static class ReadAll
            extends PTransform<PCollection<byte[]>, PCollection<KV<byte[], byte[]>>> {

        @Nullable
        abstract RedisConnectionConfiguration connectionConfiguration();

        abstract int batchSize();

        abstract ReadAll.Builder builder();

        @AutoValue.Builder
        abstract static class Builder {
            @Nullable
            abstract ReadAll.Builder setConnectionConfiguration(
                    RedisConnectionConfiguration connection);

            abstract ReadAll.Builder setBatchSize(int batchSize);

            abstract ReadAll build();
        }

        public ReadAll withEndpoint(String host, int port) {
            checkArgument(host != null, "host can not be null");
            checkArgument(port > 0, "port can not be negative or 0");
            return builder().setConnectionConfiguration(
                    connectionConfiguration().withHost(host).withPort(port)).build();
        }

        public ReadAll withAuth(String auth) {
            checkArgument(auth != null, "auth can not be null");
            return builder().setConnectionConfiguration(connectionConfiguration().withAuth(auth))
                    .build();
        }

        public ReadAll withTimeout(int timeout) {
            checkArgument(timeout >= 0, "timeout can not be negative");
            return builder()
                    .setConnectionConfiguration(connectionConfiguration().withTimeout(timeout))
                    .build();
        }

        public ReadAll withConnectionConfiguration(RedisConnectionConfiguration connection) {
            checkArgument(connection != null, "connection can not be null");
            return builder().setConnectionConfiguration(connection).build();
        }

        public ReadAll withBatchSize(int batchSize) {
            return builder().setBatchSize(batchSize).build();
        }

        @Override
        public PCollection<KV<byte[], byte[]>> expand(PCollection<byte[]> input) {
            checkArgument(connectionConfiguration() != null,
                    "withConnectionConfiguration() is required");

            return input.apply(ParDo.of(new ReadFn(connectionConfiguration(), batchSize())))
                    .setCoder(KvCoder.of(ByteArrayCoder.of(), ByteArrayCoder.of()))
                    .apply(new Reparallelize());
        }
    }

    abstract static class BaseReadFn<T> extends DoFn<byte[], T> {
        protected final RedisConnectionConfiguration connectionConfiguration;

        transient BinaryJedis jedis;

        BaseReadFn(RedisConnectionConfiguration connectionConfiguration) {
            this.connectionConfiguration = connectionConfiguration;
        }

        @Setup
        public void setup() {
            jedis = connectionConfiguration.connect();
        }

        @Teardown
        public void teardown() {
            jedis.close();
        }
    }

    private static class ReadKeysWithPattern extends BaseReadFn<byte[]> {

        ReadKeysWithPattern(RedisConnectionConfiguration connectionConfiguration) {
            super(connectionConfiguration);
        }

        @ProcessElement
        public void processElement(ProcessContext processContext) throws Exception {
            ScanParams scanParams = new ScanParams();
            scanParams.match(processContext.element());

            byte[] cursor = ScanParams.SCAN_POINTER_START_BINARY;
            boolean finished = false;
            while (!finished) {
                ScanResult<byte[]> scanResult = jedis.scan(cursor, scanParams);
                List<byte[]> keys = scanResult.getResult();
                for (byte[] k : keys) {
                    processContext.output(k);
                }
                cursor = scanResult.getCursorAsBytes();
                if (cursor.equals(ScanParams.SCAN_POINTER_START_BINARY)) {
                    finished = true;
                }
            }
        }
    }
    /** A {@link DoFn} requesting Redis server to get key/value pairs. */
    private static class ReadFn extends BaseReadFn<KV<byte[], byte[]>> {
        @Nullable
        transient Multimap<BoundedWindow, byte[]> bundles = null;
        @Nullable
        AtomicInteger batchCount = null;
        private final int batchSize;

        @StartBundle
        public void startBundle(StartBundleContext context) {
            bundles = ArrayListMultimap.create();
            batchCount = new AtomicInteger();
        }

        ReadFn(RedisConnectionConfiguration connectionConfiguration, int batchSize) {
            super(connectionConfiguration);
            this.batchSize = batchSize;
        }

        private int getBatchSize() {
            return batchSize;
        }

        @ProcessElement
        public void processElement(ProcessContext processContext, BoundedWindow window)
                throws Exception {
            byte[] key = processContext.element();
            bundles.put(window, key);
            if (batchCount.incrementAndGet() > getBatchSize()) {
                Multimap<BoundedWindow, KV<byte[], byte[]>> kvs = fetchAndFlush();
                for (BoundedWindow w : kvs.keySet()) {
                    for (KV<byte[], byte[]> kv : kvs.get(w)) {
                        processContext.output(kv);
                    }
                }
            }
        }

        private Multimap<BoundedWindow, KV<byte[], byte[]>> fetchAndFlush() {
            Multimap<BoundedWindow, KV<byte[], byte[]>> kvs = ArrayListMultimap.create();
            for (BoundedWindow w : bundles.keySet()) {
                byte[][] keys = new byte[bundles.get(w).size()][];
                bundles.get(w).toArray(keys);
                List<byte[]> results = jedis.mget(keys);
                for (int i = 0; i < results.size(); i++) {
                    if (results.get(i) != null) {
                        kvs.put(w, KV.of(keys[i], results.get(i)));
                    }
                }
            }
            bundles = ArrayListMultimap.create();
            batchCount.set(0);
            return kvs;
        }

        @FinishBundle
        public void finishBundle(FinishBundleContext context) throws Exception {
            Multimap<BoundedWindow, KV<byte[], byte[]>> kvs = fetchAndFlush();
            for (BoundedWindow w : kvs.keySet()) {
                for (KV<byte[], byte[]> kv : kvs.get(w)) {
                    context.output(kv, w.maxTimestamp(), w);
                }
            }
        }
    }

    private static class Reparallelize
            extends PTransform<PCollection<KV<byte[], byte[]>>, PCollection<KV<byte[], byte[]>>> {

        @Override
        public PCollection<KV<byte[], byte[]>> expand(PCollection<KV<byte[], byte[]>> input) {
            // reparallelize mimics the same behavior as in JdbcIO
            // breaking fusion
            PCollectionView<Iterable<KV<byte[], byte[]>>> empty =
                    input.apply("Consume", Filter.by(SerializableFunctions.constant(false)))
                            .apply(View.asIterable());
            PCollection<KV<byte[], byte[]>> materialized = input.apply("Identity",
                    ParDo.of(new DoFn<KV<byte[], byte[]>, KV<byte[], byte[]>>() {
                        @ProcessElement
                        public void processElement(ProcessContext context) {
                            context.output(context.element());
                        }
                    }).withSideInputs(empty));
            return materialized.apply(Reshuffle.viaRandomKey());
        }
    }

    /** A {@link PTransform} to write to a Redis server. */
    @AutoValue
    public abstract static class Write extends PTransform<PCollection<KV<byte[], byte[]>>, PDone> {

        /** Determines the method used to insert data in Redis. */
        public enum Method {

            /**
             * Use APPEND command. If key already exists and is a string, this command appends the
             * value at the end of the string.
             */
            APPEND,

            /** Use SET command. If key already holds a value, it is overwritten. */
            SET,

            /**
             * Use LPUSH command. Insert value at the head of the list stored at key. If key does
             * not exist, it is created as empty list before performing the push operations. When
             * key holds a value that is not a list, an error is returned.
             */
            LPUSH,

            /**
             * Use RPUSH command. Insert value at the tail of the list stored at key. If key does
             * not exist, it is created as empty list before performing the push operations. When
             * key holds a value that is not a list, an error is returned.
             */
            RPUSH,

            /** Use SADD command. Insert value in a set. Duplicated values are ignored. */
            SADD,

            /**
             * Use PFADD command. Insert value in a HLL structure. Create key if it doesn't exist
             */
            PFADD,

            /** Use INCBY command. Increment counter value of a key by a given value. */
            INCRBY,

            /** Use DECRBY command. Decrement counter value of a key by given value. */
            DECRBY,
        }

        @Nullable
        abstract RedisConnectionConfiguration connectionConfiguration();

        @Nullable
        abstract Method method();

        @Nullable
        abstract Long expireTime();

        abstract Builder builder();

        @AutoValue.Builder
        abstract static class Builder {

            abstract Builder setConnectionConfiguration(
                    RedisConnectionConfiguration connectionConfiguration);

            abstract Builder setMethod(Method method);

            abstract Builder setExpireTime(Long expireTimeMillis);

            abstract Write build();
        }

        public Write withEndpoint(String host, int port) {
            checkArgument(host != null, "host can not be null");
            checkArgument(port > 0, "port can not be negative or 0");
            return builder().setConnectionConfiguration(
                    connectionConfiguration().withHost(host).withPort(port)).build();
        }

        public Write withAuth(String auth) {
            checkArgument(auth != null, "auth can not be null");
            return builder().setConnectionConfiguration(connectionConfiguration().withAuth(auth))
                    .build();
        }

        public Write withTimeout(int timeout) {
            checkArgument(timeout >= 0, "timeout can not be negative");
            return builder()
                    .setConnectionConfiguration(connectionConfiguration().withTimeout(timeout))
                    .build();
        }

        public Write withConnectionConfiguration(RedisConnectionConfiguration connection) {
            checkArgument(connection != null, "connection can not be null");
            return builder().setConnectionConfiguration(connection).build();
        }

        public Write withMethod(Method method) {
            checkArgument(method != null, "method can not be null");
            return builder().setMethod(method).build();
        }

        public Write withExpireTime(Long expireTimeMillis) {
            checkArgument(expireTimeMillis != null, "expireTimeMillis can not be null");
            checkArgument(expireTimeMillis > 0, "expireTimeMillis can not be negative or 0");
            return builder().setExpireTime(expireTimeMillis).build();
        }

        @Override
        public PDone expand(PCollection<KV<byte[], byte[]>> input) {
            checkArgument(connectionConfiguration() != null,
                    "withConnectionConfiguration() is required");

            input.apply(ParDo.of(new WriteFn(this)));
            return PDone.in(input.getPipeline());
        }

        private static class WriteFn extends DoFn<KV<byte[], byte[]>, Void> {

            private static final int DEFAULT_BATCH_SIZE = 1000;

            private final Write spec;

            private transient BinaryJedis jedis;
            private transient Pipeline pipeline;

            private int batchCount;

            public WriteFn(Write spec) {
                this.spec = spec;
            }

            @Setup
            public void setup() {
                jedis = spec.connectionConfiguration().connect();
            }

            @StartBundle
            public void startBundle() {
                pipeline = jedis.pipelined();
                pipeline.multi();
                batchCount = 0;
            }

            @ProcessElement
            public void processElement(ProcessContext processContext) {
                KV<byte[], byte[]> record = processContext.element();

                writeRecord(record);

                batchCount++;

                if (batchCount >= DEFAULT_BATCH_SIZE) {
                    pipeline.exec();
                    pipeline.sync();
                    pipeline.multi();
                    batchCount = 0;
                }
            }

            private void writeRecord(KV<byte[], byte[]> record) {
                Method method = spec.method();
                Long expireTime = spec.expireTime();

                if (Method.APPEND == method) {
                    writeUsingAppendCommand(record, expireTime);
                } else if (Method.SET == method) {
                    writeUsingSetCommand(record, expireTime);
                } else if (Method.LPUSH == method || Method.RPUSH == method) {
                    writeUsingListCommand(record, method, expireTime);
                } else if (Method.SADD == method) {
                    writeUsingSaddCommand(record, expireTime);
                } else if (Method.PFADD == method) {
                    writeUsingHLLCommand(record, expireTime);
                } else if (Method.INCRBY == method) {
                    writeUsingIncrBy(record);
                } else if (Method.DECRBY == method) {
                    writeUsingDecrBy(record);
                }
            }

            private void writeUsingAppendCommand(KV<byte[], byte[]> record, Long expireTime) {
                byte[] key = record.getKey();
                byte[] value = record.getValue();

                pipeline.append(key, value);

                setExpireTimeWhenRequired(key, expireTime);
            }

            private void writeUsingSetCommand(KV<byte[], byte[]> record, Long expireTime) {
                byte[] key = record.getKey();
                byte[] value = record.getValue();

                if (expireTime != null) {
                    pipeline.psetex(key, expireTime, value);
                } else {
                    pipeline.set(key, value);
                }
            }

            private void writeUsingListCommand(KV<byte[], byte[]> record, Method method,
                    Long expireTime) {

                byte[] key = record.getKey();
                byte[] value = record.getValue();

                if (Method.LPUSH == method) {
                    pipeline.lpush(key, value);
                } else if (Method.RPUSH == method) {
                    pipeline.rpush(key, value);
                }

                setExpireTimeWhenRequired(key, expireTime);
            }

            private void writeUsingSaddCommand(KV<byte[], byte[]> record, Long expireTime) {
                byte[] key = record.getKey();
                byte[] value = record.getValue();

                pipeline.sadd(key, value);
            }

            private void writeUsingHLLCommand(KV<byte[], byte[]> record, Long expireTime) {
                byte[] key = record.getKey();
                byte[] value = record.getValue();

                pipeline.pfadd(key, value);
            }

            private void writeUsingIncrBy(KV<byte[], byte[]> record) {
                byte[] key = record.getKey();
                byte[] value = record.getValue();
                long inc;
                try {
                    inc = Long.parseLong(new String(value, "UTF-8"));
                } catch (UnsupportedEncodingException e) {
                    throw new NumberFormatException();
                }
                pipeline.incrBy(key, inc);
            }

            private void writeUsingDecrBy(KV<byte[], byte[]> record) {
                byte[] key = record.getKey();
                byte[] value = record.getValue();
                long decr;
                try {
                    decr = Long.parseLong(new String(value, "UTF-8"));
                } catch (UnsupportedEncodingException e) {
                    throw new NumberFormatException();
                }
                pipeline.decrBy(key, decr);
            }

            private void setExpireTimeWhenRequired(byte[] key, Long expireTime) {
                if (expireTime != null) {
                    pipeline.pexpire(key, expireTime);
                }
            }

            @FinishBundle
            public void finishBundle() {
                if (pipeline.isInMulti()) {
                    pipeline.exec();
                    pipeline.sync();
                }
                batchCount = 0;
            }

            @Teardown
            public void teardown() {
                jedis.close();
            }
        }
    }
}
