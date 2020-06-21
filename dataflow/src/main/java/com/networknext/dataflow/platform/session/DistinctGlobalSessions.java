package com.networknext.dataflow.platform.session;

import com.networknext.api.router.Router.BillingEntry;
import org.apache.beam.sdk.transforms.SerializableFunction;

public class DistinctGlobalSessions implements SerializableFunction<BillingEntry, String> {
    private static final long serialVersionUID = 1L;

    @Override
    public String apply(BillingEntry input) {
        return input.getRequest().getSessionId() + "_" + input.getTimestampStart();
    }
}
