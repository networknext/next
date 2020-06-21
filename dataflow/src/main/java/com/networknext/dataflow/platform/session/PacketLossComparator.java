package com.networknext.dataflow.platform.session;

import java.io.Serializable;
import java.util.Comparator;
import com.networknext.api.router.Router.BillingEntry;
import com.networknext.dataflow.util.BillingEntryHelpers;

public class PacketLossComparator implements Comparator<BillingEntry>, Serializable {
    private static final long serialVersionUID = 1L;

    @Override
    public int compare(BillingEntry o1, BillingEntry o2) {

        float i1 = BillingEntryHelpers.improvementPacketLoss(o1);
        float i2 = BillingEntryHelpers.improvementPacketLoss(o2);

        if (i1 == i2) {
            return 0;
        }

        if (i1 < i2) {
            return -1;
        }

        return 1;
    }

}
