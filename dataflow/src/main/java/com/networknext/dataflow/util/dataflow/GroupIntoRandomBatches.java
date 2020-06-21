package com.networknext.dataflow.util.dataflow;

import java.util.Random;

import org.apache.beam.sdk.transforms.ParDo;
import org.apache.beam.sdk.transforms.windowing.FixedWindows;
import org.apache.beam.sdk.transforms.windowing.Window;
import org.apache.beam.sdk.transforms.DoFn;
import org.apache.beam.sdk.transforms.GroupIntoBatches;
import org.apache.beam.sdk.transforms.PTransform;
import org.apache.beam.sdk.values.KV;
import org.apache.beam.sdk.values.PCollection;
import org.joda.time.Duration;

public class GroupIntoRandomBatches<T> extends PTransform<PCollection<T>, PCollection<Iterable<T>>> {

    private static final long serialVersionUID = -6668126391239383930L;

    public static final int BATCH_SIZE = 25000;
    // Define window duration.
    // After window's end - elements are emitted even if there are less then
    // BATCH_SIZE elements
    public static final int WINDOW_DURATION_SECONDS = 10;
    private static final int DEFAULT_SHARDS_NUMBER = 16;
    // Determine possible parallelism level
    private int shardsNumber = DEFAULT_SHARDS_NUMBER;

    public GroupIntoRandomBatches() {
        super();
    }

    public GroupIntoRandomBatches(int shardsNumber) {
        super();
        this.shardsNumber = shardsNumber;
    }

    @Override
    public PCollection<Iterable<T>> expand(PCollection<T> input) {
        return input
                // assign keys, as "GroupIntoBatches" works only with key-value pairs
                .apply(ParDo.of(new AssignRandomKeys(shardsNumber)))
                .apply(Window.into(FixedWindows.of(Duration.standardSeconds(WINDOW_DURATION_SECONDS))))
                .apply(GroupIntoBatches.ofSize(BATCH_SIZE)).apply(ParDo.of(new ExtractValues()));
    }

    private class AssignRandomKeys extends DoFn<T, KV<Integer, T>> {
        private static final long serialVersionUID = 6404147299941916978L;

        private int shardsNumber;
        private Random random;

        AssignRandomKeys(int shardsNumber) {
            super();
            this.shardsNumber = shardsNumber;
        }

        @Setup
        public void setup() {
            random = new Random();
        }

        @ProcessElement
        public void processElement(ProcessContext c) {
            T data = c.element();
            c.output(KV.of(random.nextInt(shardsNumber), data));
        }
    }

    private class ExtractValues extends DoFn<KV<Integer, Iterable<T>>, Iterable<T>> {
        private static final long serialVersionUID = -3878893763143806876L;

        @ProcessElement
        public void processElement(ProcessContext c) {
            KV<Integer, Iterable<T>> kv = c.element();
            c.output(kv.getValue());
        }
    }
}