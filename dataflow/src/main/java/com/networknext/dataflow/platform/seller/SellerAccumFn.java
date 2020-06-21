package com.networknext.dataflow.platform.seller;

import java.io.Serializable;
import com.google.api.services.bigquery.model.TableRow;
import org.apache.beam.sdk.transforms.Combine.AccumulatingCombineFn;
import org.apache.beam.sdk.values.Row;

public class SellerAccumFn extends AccumulatingCombineFn<Row, SellerAccumFn.Accum, TableRow> {
    private static final long serialVersionUID = -1523464748567613946L;

    public SellerAccumFn.Accum createAccumulator() {
        return new Accum();
    }

    public static class Accum implements
            AccumulatingCombineFn.Accumulator<Row, SellerAccumFn.Accum, TableRow>, Serializable {

        private static final long serialVersionUID = -4532410852004718199L;

        private String sellerId = "";
        private String relayId = "";
        private long revenueIngress = 0;
        private long revenueEgress = 0;
        private long traffic = 0;

        @Override
        public boolean equals(Object other) {
            if (!(other instanceof Accum)) {
                return false;
            }
            Accum b = (Accum) other;
            return b.sellerId.equals(sellerId) && b.relayId.equals(relayId)
                    && b.revenueIngress == revenueIngress && b.revenueEgress == revenueEgress
                    && b.traffic == traffic;
        }

        public void addInput(Row value) {
            String inputSellerId = value.<String>getValue("sellerId");
            String inputRelayId = value.<String>getValue("relayId");
            if (sellerId.equals("") && inputSellerId != null) {
                sellerId = inputSellerId;
            }
            if (relayId.equals("") && inputRelayId != null) {
                relayId = inputRelayId;
            }
            revenueIngress += value.<Long>getValue("revenueIngress");
            revenueEgress += value.<Long>getValue("revenueEgress");
            traffic += value.<Long>getValue("traffic");
        }

        public void mergeAccumulator(SellerAccumFn.Accum other) {
            if (sellerId.equals("") && other.sellerId != null) {
                sellerId = other.sellerId;
            }
            if (relayId.equals("") && other.relayId != null) {
                relayId = other.relayId;
            }
            revenueIngress += other.revenueIngress;
            revenueEgress += other.revenueEgress;
            traffic += other.traffic;
        }

        public TableRow extractOutput() {
            return new TableRow().set("timestamp", java.time.Instant.now().getEpochSecond())
                    .set("sellerId", sellerId).set("relayId", relayId)
                    .set("revenueIngress", revenueIngress).set("revenueEgress", revenueEgress)
                    .set("traffic", traffic);
        }
    }
}
