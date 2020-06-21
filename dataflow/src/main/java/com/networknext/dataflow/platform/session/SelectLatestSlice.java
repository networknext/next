package com.networknext.dataflow.platform.session;

import com.networknext.api.router.Router.BillingEntry;
import org.apache.beam.sdk.transforms.Combine;
import org.apache.beam.sdk.transforms.DoFn;
import org.apache.beam.sdk.transforms.PTransform;
import org.apache.beam.sdk.transforms.ParDo;
import org.apache.beam.sdk.transforms.SerializableFunction;
import org.apache.beam.sdk.transforms.WithKeys;
import org.apache.beam.sdk.values.KV;
import org.apache.beam.sdk.values.PCollection;

/**
 * Returns the latest slice (by timestamp) for each session in the window. When this transform is
 * complete, the resulting set is a distinct set of sessions (session ID + timestamp start), with
 * the billing entry information being the latest update we have so far.
 */
public class SelectLatestSlice
        extends PTransform<PCollection<BillingEntry>, PCollection<BillingEntry>> {

    private static final long serialVersionUID = 1L;

    @Override
    public PCollection<BillingEntry> expand(PCollection<BillingEntry> input) {
        return input
                .apply("Group by session ID and timestamp start",
                        WithKeys.of(new DistinctBillingEntry()))
                .apply("Select latest slice in session",
                        Combine.perKey(new Combine.BinaryCombineFn<BillingEntry>() {
                            private static final long serialVersionUID = 1L;

                            @Override
                            public BillingEntry apply(BillingEntry left, BillingEntry right) {
                                if (left.getTimestamp() > right.getTimestamp()) {
                                    return left;
                                }
                                return right;
                            }
                        }))
                .apply("Discard keys", ParDo.of(new DoFn<KV<String, BillingEntry>, BillingEntry>() {
                    private static final long serialVersionUID = 1L;

                    @ProcessElement
                    public void processElement(ProcessContext c) throws Exception {
                        c.output(c.element().getValue());
                    }
                }));
    }

    private class DistinctBillingEntry implements SerializableFunction<BillingEntry, String> {

        private static final long serialVersionUID = 1L;

        @Override
        public String apply(BillingEntry input) {
            return "" + input.getRequest().getSessionId();// + "/" + input.getTimestampStart();
        }

    }

}
