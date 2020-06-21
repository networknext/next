package com.networknext.dataflow.platform.buyer;

import java.io.Serializable;
import com.google.api.services.bigquery.model.TableRow;
import org.apache.beam.sdk.transforms.Combine.AccumulatingCombineFn;
import org.apache.beam.sdk.values.Row;

public class BuyerAccumFn extends AccumulatingCombineFn<Row, BuyerAccumFn.Accum, TableRow> {
    private static final long serialVersionUID = -1523464748567613946L;

    public BuyerAccumFn.Accum createAccumulator() {
        return new Accum();
    }

    public static class Accum implements
            AccumulatingCombineFn.Accumulator<Row, BuyerAccumFn.Accum, TableRow>, Serializable {

        private static final long serialVersionUID = 5394722990454780658L;

        private String buyerId = "";
        private String datacenterId = "";
        private long revenueSellers = 0;
        private long revenueNetworkNext = 0;
        private long trafficNetworkNextUp = 0;
        private long trafficNetworkNextDown = 0;
        private long trafficDirectUp = 0;
        private long trafficDirectDown = 0;
        private long sessionCount = 0;
        private long secondsNext = 0;
        private long secondsDirect = 0;
        private long secondsMeasured = 0;
        private long seconds0to5cu = 0;
        private long seconds5to10cu = 0;
        private long seconds10to15cu = 0;
        private long seconds15to20cu = 0;
        private long seconds20to30cu = 0;
        private long seconds30pluscu = 0;

        @Override
        public boolean equals(Object other) {
            if (!(other instanceof Accum)) {
                return false;
            }
            Accum b = (Accum) other;
            return b.buyerId.equals(buyerId) && b.datacenterId.equals(datacenterId)
                    && b.revenueSellers == revenueSellers
                    && b.revenueNetworkNext == revenueNetworkNext
                    && b.trafficNetworkNextUp == trafficNetworkNextUp
                    && b.trafficNetworkNextDown == trafficNetworkNextDown
                    && b.trafficDirectUp == trafficDirectUp
                    && b.trafficDirectDown == trafficDirectDown && b.sessionCount == sessionCount
                    && b.secondsNext == secondsNext && b.secondsDirect == secondsDirect
                    && b.secondsMeasured == secondsMeasured && b.seconds0to5cu == seconds0to5cu
                    && b.seconds5to10cu == seconds5to10cu && b.seconds10to15cu == seconds10to15cu
                    && b.seconds15to20cu == seconds15to20cu && b.seconds20to30cu == seconds20to30cu
                    && b.seconds30pluscu == seconds30pluscu;
        }

        public void addInput(Row value) {
            String inputBuyerId = value.<String>getValue(BuyerBq.BUYER_ID);
            String inputDatacenterId = value.<String>getValue(BuyerBq.DATACENTER_ID);
            if (buyerId.equals("") && inputBuyerId != null) {
                buyerId = inputBuyerId;
            }
            if (datacenterId.equals("") && inputDatacenterId != null) {
                datacenterId = inputDatacenterId;
            }
            revenueSellers += value.<Long>getValue(BuyerBq.REVENUE_SELLERS);
            revenueNetworkNext += value.<Long>getValue(BuyerBq.REVENUE_NETWORK_NEXT);
            trafficNetworkNextUp += value.<Long>getValue(BuyerBq.TRAFFIC_NETWORK_NEXT_UP);
            trafficNetworkNextDown += value.<Long>getValue(BuyerBq.TRAFFIC_NETWORK_NEXT_DOWN);
            trafficDirectUp += value.<Long>getValue(BuyerBq.TRAFFIC_DIRECT_UP);
            trafficDirectDown += value.<Long>getValue(BuyerBq.TRAFFIC_DIRECT_DOWN);
            if (value.<Boolean>getValue(BuyerBq.IS_NEW_SESSION)) {
                sessionCount += 1;
            }
            secondsNext += value.<Long>getValue(BuyerBq.SECONDS_NEXT);
            secondsDirect += value.<Long>getValue(BuyerBq.SECONDS_DIRECT);
            Long measured = value.<Long>getValue(BuyerBq.SECONDS_MEASURED);
            secondsMeasured += measured;
            double improvement = value.<Double>getValue(BuyerBq.IMPROVEMENT);
            if (improvement < 5.0) {
                seconds0to5cu += measured;
            } else if (improvement < 10.0) {
                seconds5to10cu += measured;
            } else if (improvement < 15.0) {
                seconds10to15cu += measured;
            } else if (improvement < 20.0) {
                seconds15to20cu += measured;
            } else if (improvement < 30.0) {
                seconds20to30cu += measured;
            } else {
                seconds30pluscu += measured;
            }
        }

        public void mergeAccumulator(BuyerAccumFn.Accum other) {
            if (buyerId.equals("") && other.buyerId != null) {
                buyerId = other.buyerId;
            }
            if (datacenterId.equals("") && other.datacenterId != null) {
                datacenterId = other.datacenterId;
            }
            revenueSellers += other.revenueSellers;
            revenueNetworkNext += other.revenueNetworkNext;
            trafficNetworkNextUp += other.trafficNetworkNextUp;
            trafficNetworkNextDown += other.trafficNetworkNextDown;
            trafficDirectUp += other.trafficDirectUp;
            trafficDirectDown += other.trafficDirectDown;
            sessionCount += other.sessionCount;
            secondsNext += other.secondsNext;
            secondsDirect += other.secondsDirect;
            secondsMeasured += other.secondsMeasured;
            seconds0to5cu += other.seconds0to5cu;
            seconds5to10cu += other.seconds5to10cu;
            seconds10to15cu += other.seconds10to15cu;
            seconds15to20cu += other.seconds15to20cu;
            seconds20to30cu += other.seconds20to30cu;
            seconds30pluscu += other.seconds30pluscu;
        }

        public TableRow extractOutput() {
            return new TableRow().set(BuyerBq.TIMESTAMP, java.time.Instant.now().getEpochSecond())
                    .set(BuyerBq.BUYER_ID, buyerId).set(BuyerBq.DATACENTER_ID, datacenterId)
                    .set(BuyerBq.REVENUE_SELLERS, revenueSellers)
                    .set(BuyerBq.REVENUE_NETWORK_NEXT, revenueNetworkNext)
                    .set(BuyerBq.TRAFFIC_NETWORK_NEXT_UP, trafficNetworkNextUp)
                    .set(BuyerBq.TRAFFIC_NETWORK_NEXT_DOWN, trafficNetworkNextDown)
                    .set(BuyerBq.TRAFFIC_DIRECT_UP, trafficDirectUp)
                    .set(BuyerBq.TRAFFIC_DIRECT_DOWN, trafficDirectDown)
                    .set(BuyerBq.SESSION_COUNT, sessionCount).set(BuyerBq.SECONDS_NEXT, secondsNext)
                    .set(BuyerBq.SECONDS_DIRECT, secondsDirect)
                    .set(BuyerBq.SECONDS_MEASURED, secondsMeasured)
                    .set(BuyerBq.SECONDS_0_TO_5_CU, seconds0to5cu)
                    .set(BuyerBq.SECONDS_5_TO_10_CU, seconds5to10cu)
                    .set(BuyerBq.SECONDS_10_TO_15_CU, seconds10to15cu)
                    .set(BuyerBq.SECONDS_15_TO_20_CU, seconds15to20cu)
                    .set(BuyerBq.SECONDS_20_TO_30_CU, seconds20to30cu)
                    .set(BuyerBq.SECONDS_30_PLUS_CU, seconds30pluscu);
        }
    }
}
