package com.networknext.dataflow.platform.session;

import java.io.Serializable;
import com.networknext.api.admin.Admin.SessionMapDataPoint;
import com.networknext.api.router.Router.BillingEntry;
import com.networknext.api.session.SessionOuterClass.SessionCountResult;
import org.apache.beam.sdk.transforms.Combine.AccumulatingCombineFn;

public class MapDataPointAccumFn extends
        AccumulatingCombineFn<SessionMapDataPoint, MapDataPointAccumFn.Accum, SessionCountResult> {
    private static final long serialVersionUID = 1L;

    public MapDataPointAccumFn.Accum createAccumulator() {
        return new Accum();
    }

    public static class Accum implements
            AccumulatingCombineFn.Accumulator<SessionMapDataPoint, MapDataPointAccumFn.Accum, SessionCountResult>,
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

        public void addInput(SessionMapDataPoint value) {
            total += value.getCount();
            if (value.getOnNetworkNext()) {
                onNetworkNext += value.getCount();
            }
        }

        public void mergeAccumulator(MapDataPointAccumFn.Accum other) {
            total += other.total;
            onNetworkNext += other.onNetworkNext;
        }

        public SessionCountResult extractOutput() {
            return SessionCountResult.newBuilder().setSessionsTotal(total)
                    .setSessionsOnNetworkNext(onNetworkNext).build();
        }
    }
}
