package com.networknext.dataflow.util;

import com.networknext.api.router.Router.BillingEntry;
import com.networknext.api.router.Router.RouteRequest;
import com.networknext.api.session.SessionOuterClass.Session;
import com.networknext.api.session.SessionOuterClass.SessionRouting;

public class BillingEntryHelpers {
    public static float improvementRtt(BillingEntry entry) {
        RouteRequest request = entry.getRequest();

        if (!entry.getOnNetworkNext()) {
            return -1;
        }
        if (request.getNextRtt() <= 0) {
            return -1;
        }

        float improvement = request.getDirectRtt() - request.getNextRtt();

        if (improvement < 0) {
            return 0;
        }

        return improvement;
    }

    public static float improvementJitter(BillingEntry entry) {
        RouteRequest request = entry.getRequest();

        if (!entry.getOnNetworkNext()) {
            return -1;
        }
        if (request.getNextRtt() <= 0) {
            return -1;
        }

        float improvement = request.getDirectJitter() - request.getNextJitter();

        if (improvement < 0) {
            return 0;
        }

        return improvement;
    }

    public static float improvementPacketLoss(BillingEntry entry) {
        RouteRequest request = entry.getRequest();

        if (!entry.getOnNetworkNext()) {
            return -1;
        }
        if (request.getNextRtt() <= 0) {
            return -1;
        }

        float improvement = request.getDirectPacketLoss() - request.getNextPacketLoss();

        if (improvement < 0) {
            return 0;
        }

        return improvement;
    }

    public static Session toSession(BillingEntry entry) {
        RouteRequest request = entry.getRequest();

        return Session.newBuilder().setSessionId(request.getSessionId())
                .setBuyerId(request.getBuyerId())
                .setServerAddress(Utils.addressPortToString(request.getServerIpAddress()))
                .setUserHash(request.getUserHash()).setDatacenterId(request.getDatacenterId())
                .setPlatformId(request.getPlatformId())
                .setSessionStartUtc(Utils.toProtobufTimestamp(entry.getTimestampStart()))
                .clearDatacenterName() // admin must look this up later
                .setSessionIdHexadecimal(Utils.hexPrint(request.getSessionId()))
                .setUserIsp(request.getLocation().getIsp())
                .setBandwidthUpBytes(entry.getUsageBytesUp())
                .setBandwidthDownBytes(entry.getUsageBytesDown())
                .setDirectRtt(request.getDirectRtt()).setNextRtt(request.getNextRtt())
                .clearNearRelayName() // we know these fields aren't used in the frontend
                .clearNearRelayRtt() // and they're annoying to compute for now
                .setDirectJitter(request.getDirectJitter()).setNextJitter(request.getNextJitter())
                .setDirectPacketLoss(request.getDirectPacketLoss())
                .setNextPacketLoss(request.getNextPacketLoss())
                .setDirectCost(request.getDirectRtt() + request.getDirectJitter()
                        + request.getDirectPacketLoss())
                .setNextCost(request.getNextRtt() + request.getNextJitter()
                        + request.getNextPacketLoss())
                .setIsLive(true)
                .setRouting(entry.getOnNetworkNext() ? SessionRouting.SESSION_ROUTING_NETWORK_NEXT
                        : SessionRouting.SESSION_ROUTING_DIRECT)
                .setFullVisibility(false)
                .setImprovementRtt(BillingEntryHelpers.improvementRtt(entry))
                .setImprovementJitter(BillingEntryHelpers.improvementJitter(entry))
                .setImprovementPacketLoss(BillingEntryHelpers.improvementPacketLoss(entry)).build();
    }
}
