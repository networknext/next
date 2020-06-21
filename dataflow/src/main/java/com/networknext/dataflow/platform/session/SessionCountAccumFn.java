package com.networknext.dataflow.platform.session;

import java.io.Serializable;
import com.networknext.api.router.Router.BillingEntry;
import com.networknext.api.session.SessionOuterClass.SessionCountResult;
import org.apache.beam.sdk.transforms.Combine.AccumulatingCombineFn;

public class SessionCountAccumFn
        extends AccumulatingCombineFn<BillingEntry, SessionCountAccumFn.Accum, SessionCountResult> {
    private static final long serialVersionUID = -1523464748567613946L;

    public SessionCountAccumFn.Accum createAccumulator() {
        return new Accum();
    }

    public static class Accum implements
            AccumulatingCombineFn.Accumulator<BillingEntry, SessionCountAccumFn.Accum, SessionCountResult>,
            Serializable {

        private static final long serialVersionUID = 1L;

        private long total = 0;
        private long onNetworkNext = 0;

        @Override
        public boolean equals(Object other) {
            if (!(other instanceof Accum)) {
                return false;
            }
            Accum b = (Accum) other;
            return b.total == total && b.onNetworkNext == onNetworkNext;
        }

        public void addInput(BillingEntry value) {
            total += 1;
            if (value.getOnNetworkNext()) {
                onNetworkNext += 1;
            }
        }

        public void mergeAccumulator(SessionCountAccumFn.Accum other) {
            total += other.total;
            onNetworkNext += other.onNetworkNext;
        }

        public SessionCountResult extractOutput() {
            return SessionCountResult.newBuilder().setSessionsTotal(total)
                    .setSessionsOnNetworkNext(onNetworkNext).build();
        }
    }
}
