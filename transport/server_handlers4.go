package transport

import (
	"errors"
	"io"
	"net"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/networknext/backend/billing"
	"github.com/networknext/backend/core"
	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/metrics"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
)

func writeServerInitResponse4(w io.Writer, packet *ServerInitRequestPacket4, response uint32) error {
	responsePacket := ServerInitResponsePacket4{
		RequestID: packet.RequestID,
		Response:  response,
	}

	responseData, err := MarshalPacket(&responsePacket)
	if err != nil {
		return err
	}

	if _, err := w.Write(responseData); err != nil {
		return err
	}

	return nil
}

func writeSessionResponse4(w io.Writer, response *SessionResponsePacket4, sessionData *SessionData4) error {
	sessionDataBuffer, err := MarshalSessionData(sessionData)
	if err != nil {
		return err
	}

	if len(sessionDataBuffer) > MaxSessionDataSize {
		return errors.New("session data too large")
	}

	response.SessionDataBytes = int32(len(sessionDataBuffer))
	copy(response.SessionData[:], sessionDataBuffer)

	responseData, err := MarshalPacket(response)
	if err != nil {
		return err
	}

	if _, err := w.Write(responseData); err != nil {
		return err
	}

	return nil
}

func ServerInitHandlerFunc4(logger log.Logger, storer storage.Storer, datacenterTracker *DatacenterTracker, metrics *metrics.ServerInitMetrics) UDPHandlerFunc {
	return func(w io.Writer, incoming *UDPPacket) {
		metrics.Invocations.Add(1)

		timeStart := time.Now()
		defer func() {
			milliseconds := float64(time.Since(timeStart).Milliseconds())
			metrics.DurationGauge.Set(milliseconds)

			if milliseconds > 100 {
				metrics.LongDuration.Add(1)
			}
		}()

		var packet ServerInitRequestPacket4
		if err := UnmarshalPacket(&packet, incoming.Data); err != nil {
			level.Error(logger).Log("msg", "could not read server init packet", "err", err)
			metrics.ErrorMetrics.ReadPacketFailure.Add(1)
			return
		}

		buyer, err := storer.Buyer(packet.CustomerID)
		if err != nil {
			level.Error(logger).Log("err", "unknown customer", "customerID", packet.CustomerID)
			metrics.ErrorMetrics.BuyerNotFound.Add(1)

			if err := writeServerInitResponse4(w, &packet, InitResponseUnknownCustomer); err != nil {
				level.Error(logger).Log("msg", "failed to write server init response", "err", err)
				metrics.ErrorMetrics.WriteResponseFailure.Add(1)
			}

			return
		}

		if !packet.Version.AtLeast(SDKVersion{4, 0, 0}) && !buyer.Debug {
			level.Error(logger).Log("err", "sdk too old", "version", packet.Version.String())
			metrics.ErrorMetrics.SDKTooOld.Add(1)

			if err := writeServerInitResponse4(w, &packet, InitResponseOldSDKVersion); err != nil {
				level.Error(logger).Log("msg", "failed to write server init response", "err", err)
				metrics.ErrorMetrics.WriteResponseFailure.Add(1)
			}

			return
		}

		datacenter, err := storer.Datacenter(packet.DatacenterID)

		// If we can't find a datacenter or alias for this customer, send an OK response
		// and track the datacenter so we can work with them and add it to our database.

		defer func() {
			if datacenter.ID == routing.UnknownDatacenter.ID {
				level.Warn(logger).Log("err", "received server init request with unknown datacenter", "datacenter", packet.DatacenterName)
				metrics.ErrorMetrics.DatacenterNotFound.Add(1)

				datacenterTracker.AddUnknownDatacenterName(packet.DatacenterName)
			}
		}()

		if err != nil {
			// search the list of aliases created by/for this buyer
			datacenterAliases := storer.GetDatacenterMapsForBuyer(packet.CustomerID)
			for _, dcMap := range datacenterAliases {
				if packet.DatacenterID == crypto.HashID(dcMap.Alias) {
					datacenter, err = storer.Datacenter(dcMap.Datacenter)

					// If the customer does have a datacenter alias set up but its misconfigured
					// in our database, then send an unknown datacenter response back.

					if err != nil {
						level.Error(logger).Log("msg", "customer has a misconfigured datacenter alias", "err", "datacenter not in database", "datacenter", packet.DatacenterName)

						if err := writeServerInitResponse4(w, &packet, InitResponseUnknownDatacenter); err != nil {
							level.Error(logger).Log("msg", "failed to write server init response", "err", err)
							metrics.ErrorMetrics.WriteResponseFailure.Add(1)
						}

						return
					}

					datacenter.AliasName = dcMap.Alias
					break
				}
			}
		}

		if err := writeServerInitResponse4(w, &packet, InitResponseOK); err != nil {
			level.Error(logger).Log("msg", "failed to write server init response", "err", err)
			metrics.ErrorMetrics.WriteResponseFailure.Add(1)
			return
		}

		level.Debug(logger).Log("msg", "server initialized successfully", "source_address", incoming.SourceAddr.String())
	}
}

func ServerUpdateHandlerFunc4(logger log.Logger, storer storage.Storer, datacenterTracker *DatacenterTracker, metrics *metrics.ServerUpdateMetrics) UDPHandlerFunc {
	return func(w io.Writer, incoming *UDPPacket) {
		metrics.Invocations.Add(1)

		timeStart := time.Now()
		defer func() {
			milliseconds := float64(time.Since(timeStart).Milliseconds())
			metrics.DurationGauge.Set(milliseconds)

			if milliseconds > 100 {
				metrics.LongDuration.Add(1)
			}
		}()

		var packet ServerUpdatePacket4
		if err := UnmarshalPacket(&packet, incoming.Data); err != nil {
			level.Error(logger).Log("msg", "could not read server update packet", "err", err)
			metrics.ErrorMetrics.ReadPacketFailure.Add(1)
			return
		}

		buyer, err := storer.Buyer(packet.CustomerID)
		if err != nil {
			level.Error(logger).Log("err", "unknown customer", "customerID", packet.CustomerID)
			metrics.ErrorMetrics.BuyerNotFound.Add(1)
			return
		}

		if !packet.Version.AtLeast(SDKVersion{4, 0, 0}) && !buyer.Debug {
			level.Error(logger).Log("err", "sdk too old", "version", packet.Version.String())
			metrics.ErrorMetrics.SDKTooOld.Add(1)
			return
		}

		datacenter, err := storer.Datacenter(packet.DatacenterID)

		// If we can't find a datacenter or alias for this customer,
		// track the datacenter so we can work with them and add it to our database.

		defer func() {
			if datacenter.ID == routing.UnknownDatacenter.ID {
				level.Warn(logger).Log("err", "received server update request with unknown datacenter", "datacenter", packet.DatacenterID)
				metrics.ErrorMetrics.DatacenterNotFound.Add(1)

				datacenterTracker.AddUnknownDatacenter(packet.DatacenterID)
			}
		}()

		if err != nil {
			// search the list of aliases created by/for this buyer
			datacenterAliases := storer.GetDatacenterMapsForBuyer(packet.CustomerID)
			for _, dcMap := range datacenterAliases {
				if packet.DatacenterID == crypto.HashID(dcMap.Alias) {
					datacenter, err = storer.Datacenter(dcMap.Datacenter)
					if err != nil {
						level.Error(logger).Log("msg", "customer has a misconfigured datacenter alias", "err", "datacenter not in database", "datacenter", packet.DatacenterID)
						return
					}

					datacenter.AliasName = dcMap.Alias
					break
				}
			}
		}

		level.Debug(logger).Log("msg", "server updated successfully", "server_address", packet.ServerAddress.String())
	}
}

func SessionUpdateHandlerFunc4(logger log.Logger, getIPLocator func() routing.IPLocator, getRouteMatrix4 func() *routing.RouteMatrix4, storer storage.Storer, routerPrivateKey [crypto.KeySize]byte, postSessionHandler *PostSessionHandler, metrics *metrics.SessionMetrics) UDPHandlerFunc {
	return func(w io.Writer, incoming *UDPPacket) {
		metrics.Invocations.Add(1)

		timeStart := time.Now()
		defer func() {
			milliseconds := float64(time.Since(timeStart).Milliseconds())
			metrics.DurationGauge.Set(milliseconds)

			if milliseconds > 100 {
				metrics.LongDuration.Add(1)
			}
		}()

		var packet SessionUpdatePacket4
		if err := UnmarshalPacket(&packet, incoming.Data); err != nil {
			level.Error(logger).Log("msg", "could not read session update packet", "err", err)
			metrics.ErrorMetrics.ReadPacketFailure.Add(1)
			return
		}

		newSession := packet.SliceNumber == 0

		var sessionData SessionData4

		ipLocator := getIPLocator()
		routeMatrix := getRouteMatrix4()
		nearRelays := []routing.NearRelayData{}
		destRelayIDs := []uint64{}
		datacenter := routing.UnknownDatacenter
		routeShader := core.NewRouteShader()       // todo: sync this up to firestore
		customerConfig := core.NewCustomerConfig() // todo: sync this up to firestore
		internalConfig := core.NewInternalConfig() // todo: sync this up to firestore
		var err error

		response := SessionResponsePacket4{
			SessionID:   packet.SessionID,
			SliceNumber: packet.SliceNumber,
			RouteType:   routing.RouteTypeDirect,
		}

		var usageBytesUp uint64
		var usageBytesDown uint64

		// If we've gotten this far, use a deferred function so that we always at least return a direct response
		// and run the post session update logic
		defer func() {
			if err := writeSessionResponse4(w, &response, &sessionData); err != nil {
				level.Error(logger).Log("msg", "failed to write session update response", "err", err)
				metrics.ErrorMetrics.WriteResponseFailure.Add(1)
				return
			}

			go PostSessionUpdate4(postSessionHandler, &packet, &sessionData.Location, nearRelays, usageBytesUp, usageBytesDown, &datacenter, metrics)
		}()

		if newSession {
			sessionData.Version = SessionDataVersion4
			sessionData.SessionID = packet.SessionID
			sessionData.SliceNumber = uint32(packet.SliceNumber + 1)
			sessionData.ExpireTimestamp = uint64(time.Now().Unix())

			sessionData.Location, err = ipLocator.LocateIP(packet.ClientAddress.IP)
			if err != nil {
				level.Error(logger).Log("msg", "failed to locate IP", "err", err)
				metrics.ErrorMetrics.ClientLocateFailure.Add(1)
				return
			}
		} else {
			if err := UnmarshalSessionData(&sessionData, packet.SessionData[:]); err != nil {
				level.Error(logger).Log("msg", "could not read session data in session update packet", "err", err)
				metrics.SessionDataMetrics.ReadSessionDataFailure.Add(1)
				return
			}

			if sessionData.SessionID != packet.SessionID {
				level.Error(logger).Log("err", "bad session ID in session data")
				metrics.SessionDataMetrics.BadSessionID.Add(1)
				return
			}

			if sessionData.SliceNumber != uint32(packet.SliceNumber) {
				level.Error(logger).Log("err", "bad sequence number in session data")
				metrics.SessionDataMetrics.BadSliceNumber.Add(1)
				return
			}

			sessionData.SliceNumber = uint32(packet.SliceNumber + 1)
			sessionData.ExpireTimestamp += billing.BillingSliceSeconds
		}

		sliceDuration := uint64(billing.BillingSliceSeconds)
		if sessionData.Initial {
			sliceDuration *= 2
		}
		usageBytesUp, usageBytesDown = CalculateNextBytesUpAndDown(uint64(packet.KbpsUp), uint64(packet.KbpsDown), sliceDuration)

		datacenter, err = storer.Datacenter(packet.DatacenterID)
		if err != nil {
			// todo: handle error here
			return
		}

		nearRelays, err = routeMatrix.GetNearRelays(sessionData.Location.Latitude, sessionData.Location.Longitude, MaxNearRelays)
		if err != nil {
			level.Error(logger).Log("msg", "failed to get near relays", "err", err)
			metrics.ErrorMetrics.NearRelaysLocateFailure.Add(1)
			return
		}

		numNearRelays := int32(len(nearRelays))
		response.NumNearRelays = numNearRelays
		response.NearRelayIDs = make([]uint64, numNearRelays)
		response.NearRelayAddresses = make([]net.UDPAddr, numNearRelays)
		for i := int32(0); i < numNearRelays; i++ {
			response.NearRelayIDs[i] = nearRelays[i].ID
			response.NearRelayAddresses[i] = *nearRelays[i].Addr
		}

		if !newSession {
			for i := range nearRelays {
				for j, clientNearRelayID := range packet.NearRelayIDs {
					if nearRelays[i].ID == clientNearRelayID {
						nearRelays[i].ClientStats.RTT = float64(packet.NearRelayRTT[j])
						nearRelays[i].ClientStats.Jitter = float64(packet.NearRelayJitter[j])
						nearRelays[i].ClientStats.PacketLoss = float64(packet.NearRelayPacketLoss[j])
					}
				}
			}
		}

		destRelayIDs = routeMatrix.GetDatacenterRelayIDs(packet.DatacenterID)
		if len(destRelayIDs) == 0 {
			level.Error(logger).Log("msg", "failed to get dest relays", "err", err)
			metrics.ErrorMetrics.NoRelaysInDatacenter.Add(1)
			return
		}

		// todo: move this to route matrix 4
		relayIDsToIndices := make(map[uint64]int32)
		for i, relayID := range routeMatrix.RelayIDs {
			relayIDsToIndices[relayID] = int32(i)
		}

		nearRelayIDs := make([]uint64, numNearRelays)
		nearRelayCosts := make([]int32, numNearRelays)
		nearRelayPacketLoss := make([]float32, numNearRelays)

		for i := int32(0); i < numNearRelays; i++ {
			nearRelay := &nearRelays[i]

			nearRelayIDs[i] = nearRelay.ID
			nearRelayCosts[i] = int32(nearRelay.ClientStats.RTT)
			nearRelayPacketLoss[i] = float32(nearRelay.ClientStats.PacketLoss)
		}

		reframedNearRelays := make([]int32, numNearRelays)
		var numDestRelays int32
		reframedDestRelays := make([]int32, len(destRelayIDs))
		core.ReframeRelays(relayIDsToIndices, nearRelayIDs, nearRelayCosts, nearRelayPacketLoss, destRelayIDs, &numNearRelays, reframedNearRelays, &numDestRelays, reframedDestRelays)

		reframedNearRelays = reframedNearRelays[:numNearRelays]
		reframedDestRelays = reframedDestRelays[:numDestRelays]

		var routeCost int32
		var routeNumRelays int32
		routeRelays := make([]int32, routing.MaxRelays)

		// todo: use in local route shader
		routeShader.LatencyThreshold = -1
		routeShader.AcceptableLatency = -1

		sessionData.Initial = false
		if !sessionData.RouteState.Next {
			if core.MakeRouteDecision_TakeNetworkNext(routeMatrix.RouteEntries, &routeShader, &sessionData.RouteState, &customerConfig, &internalConfig, int32(packet.DirectRTT), packet.DirectPacketLoss, reframedNearRelays, nearRelayCosts, reframedDestRelays, &routeCost, &routeNumRelays, routeRelays) {
				// Next token

				// Add another 10 seconds to the slice and increment the session version
				sessionData.Initial = true
				sessionData.ExpireTimestamp += billing.BillingSliceSeconds
				sessionData.SessionVersion++

				numTokens := routeNumRelays + 2 // relays + client + server
				routeAddresses, routePublicKeys := GetRouteAddressesAndPublicKeys(&packet.ClientAddress, packet.ClientRoutePublicKey, &packet.ServerAddress, packet.ServerRoutePublicKey, numTokens, routeRelays, routeMatrix.RelayIDs, storer)
				tokenData := make([]byte, numTokens*routing.EncryptedNextRouteTokenSize4)
				core.WriteRouteTokens(tokenData, sessionData.ExpireTimestamp, sessionData.SessionID, uint8(sessionData.SessionVersion), uint32(routeShader.BandwidthEnvelopeUpKbps), uint32(routeShader.BandwidthEnvelopeDownKbps), int(numTokens), routeAddresses, routePublicKeys, routerPrivateKey)
				response.RouteType = routing.RouteTypeNew
				response.NumTokens = numTokens
				response.Tokens = tokenData
			}
		} else {
			reframedRouteRelays := [routing.MaxRelays]int32{}
			if !core.ReframeRoute(relayIDsToIndices, sessionData.Route.RelayIDs[:], &reframedRouteRelays) {
				// todo: handle error - route relay is no longer in route matrix
			}

			if core.MakeRouteDecision_StayOnNetworkNext(routeMatrix.RouteEntries, &routeShader, &sessionData.RouteState, &customerConfig, &internalConfig, int32(packet.DirectRTT), int32(sessionData.Route.Stats.RTT), int32(sessionData.Route.NumRelays), reframedRouteRelays, reframedNearRelays, nearRelayCosts, reframedDestRelays, &routeCost, &routeNumRelays, routeRelays) {
				// Continue token

				numTokens := routeNumRelays + 2 // relays + client + server
				_, routePublicKeys := GetRouteAddressesAndPublicKeys(&packet.ClientAddress, packet.ClientRoutePublicKey, &packet.ServerAddress, packet.ServerRoutePublicKey, numTokens, routeRelays, routeMatrix.RelayIDs, storer)
				tokenData := make([]byte, routing.EncryptedContinueRouteTokenSize4)
				core.WriteContinueTokens(tokenData, sessionData.ExpireTimestamp, sessionData.SessionID, uint8(sessionData.SessionVersion), int(numTokens), routePublicKeys, routerPrivateKey)
				response.RouteType = routing.RouteTypeContinue
				response.NumTokens = numTokens
				response.Tokens = tokenData
			}
		}

		level.Debug(logger).Log("msg", "session updated successfully", "source_address", incoming.SourceAddr.String(), "server_address", packet.ServerAddress.String(), "client_address", packet.ClientAddress.String())
	}
}

func GetRouteAddressesAndPublicKeys(clientAddress *net.UDPAddr, clientPublicKey []byte, serverAddress *net.UDPAddr, serverPublicKey []byte, numTokens int32, routeRelays []int32, allRelayIDs []uint64, storer storage.Storer) ([]*net.UDPAddr, [][]byte) {
	routeAddresses := make([]*net.UDPAddr, numTokens)
	routePublicKeys := make([][]byte, numTokens)

	routeAddresses[0] = clientAddress
	routePublicKeys[0] = clientPublicKey
	routeAddresses[numTokens-1] = serverAddress
	routePublicKeys[numTokens-1] = serverPublicKey

	totalNumRelays := int32(len(allRelayIDs))
	for i := int32(1); i < numTokens-1; i++ {
		for j := int32(0); j < totalNumRelays; j++ {
			relayIndex := routeRelays[i]
			if j == relayIndex {
				relayID := allRelayIDs[relayIndex]
				relay, err := storer.Relay(relayID)
				if err != nil {
					// todo: handle error?
					continue
				}

				routeAddresses[i] = &relay.Addr
				routePublicKeys[i] = relay.PublicKey
				break
			}
		}
	}

	return routeAddresses, routePublicKeys
}

func PostSessionUpdate4(postSessionHandler *PostSessionHandler, packet *SessionUpdatePacket4, location *routing.Location, nearRelays []routing.NearRelayData, nextBytesUp uint64, nextBytesDown uint64, datacenter *routing.Datacenter, metrics *metrics.SessionMetrics) {
	billingEntry := &billing.BillingEntry{
		BuyerID:                   packet.CustomerID,
		UserHash:                  packet.UserHash,
		SessionID:                 packet.SessionID,
		SliceNumber:               packet.SliceNumber,
		DirectRTT:                 packet.DirectRTT,
		DirectJitter:              packet.DirectJitter,
		DirectPacketLoss:          packet.DirectPacketLoss,
		Next:                      packet.Next,
		NextRTT:                   packet.NextRTT,
		NextJitter:                packet.NextJitter,
		NextPacketLoss:            packet.NextPacketLoss,
		NumNextRelays:             0,
		NextRelays:                [routing.MaxRelays]uint64{},
		TotalPrice:                0,
		ClientToServerPacketsLost: packet.PacketsLostClientToServer,
		ServerToClientPacketsLost: packet.PacketsLostServerToClient,
		Committed:                 packet.Committed,
		Flagged:                   packet.Reported,
		Multipath:                 false,
		Initial:                   false,
		NextBytesUp:               nextBytesUp,
		NextBytesDown:             nextBytesDown,
		DatacenterID:              datacenter.ID,
		RTTReduction:              false,
		PacketLossReduction:       false,
		NextRelaysPrice:           [routing.MaxRelays]uint64{},
		Latitude:                  float32(location.Latitude),
		Longitude:                 float32(location.Longitude),
		ISP:                       location.ISP,
		ABTest:                    false,
		RouteDecision:             0,
		ConnectionType:            uint8(packet.ConnectionType),
		PlatformType:              uint8(packet.PlatformType),
		SDKVersion:                packet.Version.String(),
	}

	if !postSessionHandler.IsBillingBufferFull() {
		postSessionHandler.SendBillingEntry(billingEntry)
		metrics.PostSessionBillingEntriesSent.Add(1)
	} else {
		metrics.PostSessionBillingBufferFull.Add(1)
	}

	nearRelayPortalData := make([]NearRelayPortalData, len(nearRelays))
	for i := range nearRelayPortalData {
		nearRelayPortalData[i] = NearRelayPortalData{
			ID:          nearRelays[i].ID,
			Name:        nearRelays[i].Name,
			ClientStats: nearRelays[i].ClientStats,
		}
	}

	var deltaRTT float32
	if packet.Next && packet.NextRTT != 0 && packet.DirectRTT >= packet.NextRTT {
		deltaRTT = packet.DirectRTT - packet.NextRTT
	}

	portalData := &SessionPortalData{
		Meta: SessionMeta{
			ID:              packet.SessionID,
			UserHash:        packet.UserHash,
			DatacenterName:  datacenter.Name,
			DatacenterAlias: datacenter.AliasName,
			OnNetworkNext:   packet.Next,
			NextRTT:         float64(packet.NextRTT),
			DirectRTT:       float64(packet.DirectRTT),
			DeltaRTT:        float64(deltaRTT),
			Location:        *location,
			ClientAddr:      packet.ClientAddress.String(),
			ServerAddr:      packet.ServerAddress.String(),
			Hops:            nil,
			SDK:             packet.Version.String(),
			Connection:      uint8(packet.ConnectionType),
			NearbyRelays:    nearRelayPortalData,
			Platform:        uint8(packet.PlatformType),
			BuyerID:         packet.CustomerID,
		},
		Slice: SessionSlice{
			Timestamp: time.Now(),
			Next: routing.Stats{
				RTT:        float64(packet.NextRTT),
				Jitter:     float64(packet.NextJitter),
				PacketLoss: float64(packet.NextPacketLoss),
			},
			Direct: routing.Stats{
				RTT:        float64(packet.DirectRTT),
				Jitter:     float64(packet.DirectJitter),
				PacketLoss: float64(packet.DirectPacketLoss),
			},
			Envelope: routing.Envelope{
				Up:   int64(packet.KbpsUp),
				Down: int64(packet.KbpsDown),
			},
			IsMultiPath:       false,
			IsTryBeforeYouBuy: false,
			OnNetworkNext:     packet.Next,
		},
		Point: SessionMapPoint{
			Latitude:  location.Latitude,
			Longitude: location.Longitude,
		},
	}

	if !postSessionHandler.IsPortalBufferFull() {
		if portalData.Meta.NextRTT != 0 || portalData.Meta.DirectRTT != 0 {
			postSessionHandler.SendPortalData(portalData)
			metrics.PostSessionPortalEntriesSent.Add(1)
		}
	} else {
		metrics.PostSessionPortalBufferFull.Add(1)
	}
}
