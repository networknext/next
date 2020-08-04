package transport

import (
	"context"
	"fmt"

	"github.com/go-kit/kit/log/level"
	"github.com/networknext/backend/billing"
	"github.com/networknext/backend/routing"
)

type PostSessionData struct {
	Params          *PostSessionUpdateParams
	PortalData      *SessionPortalData
	PortalCountData *SessionCountData
	BillingEntry    *billing.BillingEntry
}

func (post *PostSessionData) ProcessBillingEntry() {
	// Send billing specific data to the billing service via google pubsub
	// The billing service subscribes to this topic, and writes the billing data to bigquery.
	// We tried writing to bigquery directly here, but it didn't work because bigquery would stall out.
	// BigQuery really doesn't make performance guarantees on how fast it is to load data, so we need
	// pubsub to act as a queue to smooth that out. Pubsub can buffer billing data for up to 7 days.

	isMultipath := routing.IsMultipath(post.Params.prevRouteDecision)

	nextRelays := [billing.BillingEntryMaxRelays]uint64{}
	for i := 0; i < len(post.Params.routeRelays) && i < len(nextRelays); i++ {
		nextRelays[i] = post.Params.routeRelays[i].ID
	}

	nextRelaysPriceArray := [billing.BillingEntryMaxRelays]uint64{}
	for i := 0; i < len(nextRelaysPriceArray) && i < len(post.Params.nextRelaysPrice); i++ {
		nextRelaysPriceArray[i] = uint64(post.Params.nextRelaysPrice[i])
	}

	billingEntry := billing.BillingEntry{
		BuyerID:                   post.Params.packet.CustomerID,
		UserHash:                  post.Params.packet.UserHash,
		SessionID:                 post.Params.packet.SessionID,
		SliceNumber:               uint32(post.Params.packet.Sequence),
		DirectRTT:                 float32(post.Params.lastDirectStats.RTT),
		DirectJitter:              float32(post.Params.lastDirectStats.Jitter),
		DirectPacketLoss:          float32(post.Params.lastDirectStats.PacketLoss),
		Next:                      post.Params.packet.OnNetworkNext,
		NextRTT:                   float32(post.Params.lastNextStats.RTT),
		NextJitter:                float32(post.Params.lastNextStats.Jitter),
		NextPacketLoss:            float32(post.Params.lastNextStats.PacketLoss),
		NumNextRelays:             uint8(len(post.Params.routeRelays)),
		NextRelays:                nextRelays,
		TotalPrice:                uint64(post.Params.totalPriceNibblins),
		ClientToServerPacketsLost: post.Params.packet.PacketsLostClientToServer,
		ServerToClientPacketsLost: post.Params.packet.PacketsLostServerToClient,
		Committed:                 post.Params.packet.Committed,
		Flagged:                   post.Params.packet.Flagged,
		Multipath:                 isMultipath,
		Initial:                   post.Params.prevInitial,
		NextBytesUp:               post.Params.nextBytesUp,
		NextBytesDown:             post.Params.nextBytesDown,
		DatacenterID:              post.Params.serverDataReadOnly.Datacenter.ID,
		RTTReduction:              post.Params.prevRouteDecision.Reason&routing.DecisionRTTReduction != 0 || post.Params.prevRouteDecision.Reason&routing.DecisionRTTReductionMultipath != 0,
		PacketLossReduction:       post.Params.prevRouteDecision.Reason&routing.DecisionHighPacketLossMultipath != 0,
		NextRelaysPrice:           nextRelaysPriceArray,
	}

	if err := post.Params.sessionUpdateParams.Biller.Bill(context.Background(), &billingEntry); err != nil {
		level.Error(post.Params.sessionUpdateParams.Logger).Log("msg", "could not submit billing entry", "err", err)
		post.Params.sessionUpdateParams.Metrics.ErrorMetrics.BillingFailure.Add(1)
	}
}

func (post *PostSessionData) ProcessPortalData() {
	// IMPORTANT: we actually need to display the true datacenter name in the demo and demo plus views,
	// while in the customer view of the portal, we need to display the alias. this is because aliases will
	// shortly become per-customer, thus there is really no global concept of "multiplay.losangeles", for example.

	datacenterName := post.Params.serverDataReadOnly.Datacenter.Name
	datacenterAlias := post.Params.serverDataReadOnly.Datacenter.AliasName

	// Send a massive amount of data to the portal via ZeroMQ to the portal cruncher.
	// This drives all the stuff you see in the portal, including the map and top sessions list.
	// We send it via ZeroMQ to the portal cruncher because google pubsub is not able to deliver data quickly enough,
	// and writing all to redis would stall the session update.

	isMultipath := routing.IsMultipath(post.Params.prevRouteDecision)

	sessionCountData := SessionCountData{
		InstanceID:                post.Params.sessionUpdateParams.InstanceID,
		TotalNumDirectSessions:    post.Params.sessionUpdateParams.SessionMap.GetDirectSessionCount(),
		TotalNumNextSessions:      post.Params.sessionUpdateParams.SessionMap.GetNextSessionCount(),
		NumDirectSessionsPerBuyer: post.Params.sessionUpdateParams.SessionMap.GetDirectSessionCountPerBuyer(),
		NumNextSessionsPerBuyer:   post.Params.sessionUpdateParams.SessionMap.GetNextSessionCountPerBuyer(),
	}

	hops := make([]RelayHop, len(post.Params.routeRelays))
	for i := range hops {
		hops[i] = RelayHop{
			ID:   post.Params.routeRelays[i].ID,
			Name: post.Params.routeRelays[i].Name,
		}
	}

	nearRelayData := make([]NearRelayPortalData, len(post.Params.nearRelays))
	for i := range nearRelayData {
		nearRelayData[i] = NearRelayPortalData{
			ID:          post.Params.nearRelays[i].ID,
			Name:        post.Params.nearRelays[i].Name,
			ClientStats: post.Params.nearRelays[i].ClientStats,
		}
	}

	portalDataBytes, err := updatePortalData(post.Params.sessionUpdateParams.PortalPublisher, post.Params.packet, post.Params.lastNextStats, post.Params.lastDirectStats, hops,
		post.Params.packet.OnNetworkNext, datacenterName, post.Params.location, nearRelayData, post.Params.timeNow, isMultipath, datacenterAlias, &sessionCountData)
	if err != nil {
		level.Error(post.Params.sessionUpdateParams.Logger).Log("msg", "could not update portal data", "err", err)
		post.Params.sessionUpdateParams.Metrics.ErrorMetrics.UpdatePortalFailure.Add(1)
	}

	level.Debug(post.Params.sessionUpdateParams.Logger).Log("msg", fmt.Sprintf("published %d bytes to portal cruncher", portalDataBytes))
}

type PostSessionHandler struct {
	numGoroutines      int
	postSessionChannel chan *PostSessionData
}

func NewPostSessionHandler(numGoroutines int) *PostSessionHandler {
	return &PostSessionHandler{
		numGoroutines:      numGoroutines,
		postSessionChannel: make(chan *PostSessionData, 1000000),
	}
}

func (post *PostSessionHandler) StartProcessing(ctx context.Context) {
	for i := 0; i < post.numGoroutines; i++ {
		go func() {
			for {
				select {
				case postSessionData := <-post.postSessionChannel:
					postSessionData.ProcessPortalData()
					postSessionData.ProcessBillingEntry()
				case <-ctx.Done():
					return
				}
			}
		}()
	}
}

func (post *PostSessionHandler) Send(postSessionData *PostSessionData) {
	post.postSessionChannel <- postSessionData
}
