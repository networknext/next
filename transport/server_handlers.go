package transport

import (
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/networknext/backend/modules/billing"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
)

type UDPPacket struct {
	SourceAddr net.UDPAddr
	Data       []byte
}

// UDPHandlerFunc acts the same way http.HandlerFunc does, but for UDP packets and address
type UDPHandlerFunc func(io.Writer, *UDPPacket)

func writeServerInitResponse(w io.Writer, packet *ServerInitRequestPacket, response uint32) error {
	responsePacket := ServerInitResponsePacket{
		RequestID: packet.RequestID,
		Response:  response,
	}

	responsePacketData, err := MarshalPacket(&responsePacket)
	if err != nil {
		return err
	}

	packetHeader := append([]byte{PacketTypeServerInitResponse}, make([]byte, crypto.PacketHashSize)...)
	responseData := append(packetHeader, responsePacketData...)
	if _, err := w.Write(responseData); err != nil {
		return err
	}

	return nil
}

func writeSessionResponse(w io.Writer, response *SessionResponsePacket, sessionData *SessionData) error {
	sessionDataBuffer, err := MarshalSessionData(sessionData)
	if err != nil {
		return err
	}

	if len(sessionDataBuffer) > MaxSessionDataSize {
		return errors.New("session data too large")
	}

	response.SessionDataBytes = int32(len(sessionDataBuffer))
	copy(response.SessionData[:], sessionDataBuffer)

	responsePacketData, err := MarshalPacket(response)
	if err != nil {
		return err
	}

	packetHeader := append([]byte{PacketTypeSessionResponse}, make([]byte, crypto.PacketHashSize)...)
	responseData := append(packetHeader, responsePacketData...)
	if _, err := w.Write(responseData); err != nil {
		return err
	}

	return nil
}

func ServerInitHandlerFunc(logger log.Logger, storer storage.Storer, datacenterTracker *DatacenterTracker, metrics *metrics.ServerInitMetrics) UDPHandlerFunc {
	return func(w io.Writer, incoming *UDPPacket) {
		metrics.HandlerMetrics.Invocations.Add(1)

		timeStart := time.Now()
		defer func() {
			milliseconds := float64(time.Since(timeStart).Milliseconds())
			metrics.HandlerMetrics.Duration.Set(milliseconds)

			if milliseconds > 100 {
				metrics.HandlerMetrics.LongDuration.Add(1)
			}
		}()

		var packet ServerInitRequestPacket
		if err := UnmarshalPacket(&packet, incoming.Data); err != nil {
			level.Error(logger).Log("msg", "could not read server init packet", "err", err)
			metrics.ReadPacketFailure.Add(1)
			return
		}

		buyer, err := storer.Buyer(packet.CustomerID)
		if err != nil {
			level.Error(logger).Log("err", "unknown customer", "customerID", packet.CustomerID)
			metrics.BuyerNotFound.Add(1)

			if err := writeServerInitResponse(w, &packet, InitResponseUnknownCustomer); err != nil {
				level.Error(logger).Log("msg", "failed to write server init response", "err", err)
				metrics.WriteResponseFailure.Add(1)
			}

			return
		}

		if !packet.Version.AtLeast(SDKVersion{4, 0, 0}) && !buyer.Debug {
			level.Error(logger).Log("err", "sdk too old", "version", packet.Version.String())
			metrics.SDKTooOld.Add(1)

			if err := writeServerInitResponse(w, &packet, InitResponseOldSDKVersion); err != nil {
				level.Error(logger).Log("msg", "failed to write server init response", "err", err)
				metrics.WriteResponseFailure.Add(1)
			}

			return
		}

		datacenter, err := storer.Datacenter(packet.DatacenterID)

		// If we can't find a datacenter or alias for this customer, send an OK response
		// and track the datacenter so we can work with them and add it to our database.

		defer func() {
			if datacenter.ID == routing.UnknownDatacenter.ID {
				level.Warn(logger).Log("err", "received server init request with unknown datacenter", "datacenter", packet.DatacenterName)
				metrics.DatacenterNotFound.Add(1)

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

						if err := writeServerInitResponse(w, &packet, InitResponseUnknownDatacenter); err != nil {
							level.Error(logger).Log("msg", "failed to write server init response", "err", err)
							metrics.WriteResponseFailure.Add(1)
						}

						return
					}

					datacenter.AliasName = dcMap.Alias
					break
				}
			}
		}

		if err := writeServerInitResponse(w, &packet, InitResponseOK); err != nil {
			level.Error(logger).Log("msg", "failed to write server init response", "err", err)
			metrics.WriteResponseFailure.Add(1)
			return
		}

		level.Debug(logger).Log("msg", "server initialized successfully", "source_address", incoming.SourceAddr.String())
	}
}

func ServerUpdateHandlerFunc(logger log.Logger, storer storage.Storer, datacenterTracker *DatacenterTracker, postSessionHandler *PostSessionHandler, metrics *metrics.ServerUpdateMetrics) UDPHandlerFunc {
	return func(w io.Writer, incoming *UDPPacket) {
		metrics.HandlerMetrics.Invocations.Add(1)

		timeStart := time.Now()
		defer func() {
			milliseconds := float64(time.Since(timeStart).Milliseconds())
			metrics.HandlerMetrics.Duration.Set(milliseconds)

			if milliseconds > 100 {
				metrics.HandlerMetrics.LongDuration.Add(1)
			}
		}()

		var packet ServerUpdatePacket
		if err := UnmarshalPacket(&packet, incoming.Data); err != nil {
			level.Error(logger).Log("msg", "could not read server update packet", "err", err)
			metrics.ReadPacketFailure.Add(1)
			return
		}

		buyer, err := storer.Buyer(packet.CustomerID)
		if err != nil {
			level.Error(logger).Log("err", "unknown customer", "customerID", packet.CustomerID)
			metrics.BuyerNotFound.Add(1)
			return
		}

		if !packet.Version.AtLeast(SDKVersion{4, 0, 0}) && !buyer.Debug {
			level.Error(logger).Log("err", "sdk too old", "version", packet.Version.String())
			metrics.SDKTooOld.Add(1)
			return
		}

		datacenter, err := storer.Datacenter(packet.DatacenterID)

		// If we can't find a datacenter or alias for this customer,
		// track the datacenter so we can work with them and add it to our database.

		defer func() {
			if datacenter.ID == routing.UnknownDatacenter.ID {
				level.Warn(logger).Log("err", "received server update request with unknown datacenter", "datacenter", packet.DatacenterID)
				metrics.DatacenterNotFound.Add(1)

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

		// Send the number of sessions on the server to the portal cruncher
		countData := &SessionCountData{
			ServerID:    crypto.HashID(packet.ServerAddress.String()),
			BuyerID:     buyer.ID,
			NumSessions: packet.NumSessions,
		}
		postSessionHandler.SendPortalCounts(countData)

		level.Debug(logger).Log("msg", "server updated successfully", "source_address", incoming.SourceAddr.String(), "server_address", packet.ServerAddress.String())
	}
}

func SessionUpdateHandlerFunc(logger log.Logger, getIPLocator func(sessionID uint64) routing.IPLocator, getRouteMatrix func() *routing.RouteMatrix, multipathVetoHandler *storage.MultipathVetoHandler, storer storage.Storer, maxNearRelays int, routerPrivateKey [crypto.KeySize]byte, postSessionHandler *PostSessionHandler, metrics *metrics.SessionUpdateMetrics, internalIPSellers []string, enableInternalIPs bool) UDPHandlerFunc {
	return func(w io.Writer, incoming *UDPPacket) {
		metrics.HandlerMetrics.Invocations.Add(1)

		timeStart := time.Now()
		defer func() {
			milliseconds := float64(time.Since(timeStart).Milliseconds())
			metrics.HandlerMetrics.Duration.Set(milliseconds)

			if milliseconds > 100 {
				metrics.HandlerMetrics.LongDuration.Add(1)
			}
		}()

		var packet SessionUpdatePacket
		if err := UnmarshalPacket(&packet, incoming.Data); err != nil {
			level.Error(logger).Log("msg", "could not read session update packet", "err", err)
			metrics.ReadPacketFailure.Add(1)
			return
		}

		buyer, err := storer.Buyer(packet.CustomerID)
		if err != nil {
			level.Error(logger).Log("msg", "buyer not found", "err", err)
			metrics.BuyerNotFound.Add(1)
			return
		}

		newSession := packet.SliceNumber == 0

		var sessionData SessionData

		ipLocator := getIPLocator(packet.SessionID)
		routeMatrix := getRouteMatrix()
		nearRelays := []routing.NearRelayData{}
		datacenter := routing.UnknownDatacenter

		response := SessionResponsePacket{
			SessionID:   packet.SessionID,
			SliceNumber: packet.SliceNumber,
			RouteType:   routing.RouteTypeDirect,
		}

		var routeNumRelays int32
		var routeRelayNames [routing.MaxRelays]string
		var routeRelaySellers [routing.MaxRelays]routing.Seller

		// If we've gotten this far, use a deferred function so that we always at least return a direct response
		// and run the post session update logic
		defer func() {
			if err := writeSessionResponse(w, &response, &sessionData); err != nil {
				level.Error(logger).Log("msg", "failed to write session update response", "err", err)
				metrics.WriteResponseFailure.Add(1)
				return
			}

			packet.ClientAddress = AnonymizeAddr(packet.ClientAddress) // Make sure to always anonymize the client's IP address

			if sessionData.RouteState.Next {
				metrics.NextSlices.Add(1)
				sessionData.EverOnNext = true
			} else {
				metrics.DirectSlices.Add(1)
			}

			go PostSessionUpdate(postSessionHandler, &packet, &sessionData, &buyer, multipathVetoHandler, routeRelayNames, routeRelaySellers, nearRelays, &datacenter)
		}()

		if newSession {
			sessionData.Version = SessionDataVersion
			sessionData.SessionID = packet.SessionID
			sessionData.SliceNumber = uint32(packet.SliceNumber + 1)
			sessionData.ExpireTimestamp = uint64(time.Now().Unix()) + billing.BillingSliceSeconds
			sessionData.RouteState.UserID = packet.UserHash
			sessionData.Location, err = ipLocator.LocateIP(packet.ClientAddress.IP)

			if err != nil {
				level.Error(logger).Log("msg", "failed to locate IP", "err", err)
				metrics.ClientLocateFailure.Add(1)
				return
			}
		} else {
			if err := UnmarshalSessionData(&sessionData, packet.SessionData[:]); err != nil {
				level.Error(logger).Log("msg", "could not read session data in session update packet", "err", err)
				metrics.ReadSessionDataFailure.Add(1)
				return
			}

			if sessionData.SessionID != packet.SessionID {
				level.Error(logger).Log("err", "bad session ID in session data")
				metrics.BadSessionID.Add(1)
				return
			}

			if sessionData.SliceNumber != uint32(packet.SliceNumber) {
				level.Error(logger).Log("err", "bad sequence number in session data")
				metrics.BadSliceNumber.Add(1)
				return
			}

			sessionData.SliceNumber = uint32(packet.SliceNumber + 1)
			sessionData.ExpireTimestamp += billing.BillingSliceSeconds
		}

		// Don't accelerate any sessions if the buyer is not yet live
		if !buyer.Live {
			metrics.BuyerNotLive.Add(1)
			return
		}

		if packet.FallbackToDirect {
			if !sessionData.FellBackToDirect {
				metrics.FallbackToDirect.Add(1)
				sessionData.FellBackToDirect = true
			}
			return
		}

		datacenter, err = storer.Datacenter(packet.DatacenterID)
		if err != nil {
			aliasFound := false

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
					aliasFound = true
					break
				}
			}

			if !aliasFound {
				level.Error(logger).Log("msg", "datacenter not found", "err", err)
				metrics.DatacenterNotFound.Add(1)
				return
			}
		}

		nearRelays, err = routeMatrix.GetNearRelays(sessionData.Location.Latitude, sessionData.Location.Longitude, maxNearRelays)
		if err != nil {
			level.Error(logger).Log("msg", "failed to get near relays", "err", err)
			metrics.NearRelaysLocateFailure.Add(1)
			return
		}

		if !newSession {
			for i := range nearRelays {
				for j, clientNearRelayID := range packet.NearRelayIDs {
					if nearRelays[i].ID == clientNearRelayID {
						nearRelays[i].ClientStats.RTT = math.Ceil(float64(packet.NearRelayRTT[j]))
						nearRelays[i].ClientStats.Jitter = math.Ceil(float64(packet.NearRelayJitter[j]))
						nearRelays[i].ClientStats.PacketLoss = math.Ceil(float64(packet.NearRelayPacketLoss[j]))
					}
				}
			}
		}

		numNearRelays := int32(len(nearRelays))
		nearRelayIDs := make([]uint64, numNearRelays)
		nearRelayAddresses := make([]net.UDPAddr, numNearRelays)
		nearRelayCosts := make([]int32, numNearRelays)
		nearRelayPacketLoss := make([]float32, numNearRelays)

		for i := int32(0); i < numNearRelays; i++ {
			nearRelay := &nearRelays[i]

			nearRelayIDs[i] = nearRelay.ID
			nearRelayAddresses[i] = *nearRelay.Addr
			nearRelayCosts[i] = int32(nearRelay.ClientStats.RTT)
			nearRelayPacketLoss[i] = float32(nearRelay.ClientStats.PacketLoss)
		}

		response.NumNearRelays = numNearRelays
		response.NearRelayIDs = nearRelayIDs
		response.NearRelayAddresses = nearRelayAddresses

		// First slice always direct
		if newSession {
			level.Debug(logger).Log("msg", "session updated successfully", "source_address", incoming.SourceAddr.String(), "server_address", packet.ServerAddress.String(), "client_address", packet.ClientAddress.String())
			return
		}

		destRelayIDs := routeMatrix.GetDatacenterRelayIDs(datacenter.ID)
		if len(destRelayIDs) == 0 {
			level.Error(logger).Log("msg", "failed to get dest relays")
			metrics.NoRelaysInDatacenter.Add(1)
			return
		}

		reframedNearRelays := make([]int32, numNearRelays)
		var numDestRelays int32
		reframedDestRelays := make([]int32, len(destRelayIDs))
		core.ReframeRelays(routeMatrix.RelayIDsToIndices, nearRelayIDs, nearRelayCosts, nearRelayPacketLoss, destRelayIDs, &numNearRelays, reframedNearRelays, &numDestRelays, reframedDestRelays)

		reframedNearRelays = reframedNearRelays[:numNearRelays]
		reframedDestRelays = reframedDestRelays[:numDestRelays]

		var routeCost int32
		routeRelays := [routing.MaxRelays]int32{}

		sessionData.Initial = false

		multipathVetoMap := multipathVetoHandler.GetMapCopy(buyer.CompanyCode)

		level.Debug(logger).Log("buyer", buyer.CompanyCode,
			"acceptable_latency", buyer.RouteShader.AcceptableLatency,
			"rtt_threshold", buyer.RouteShader.LatencyThreshold,
			"selection_percent", buyer.RouteShader.SelectionPercent,
			"route_switch_threshold", buyer.InternalConfig.RouteSwitchThreshold)

		if !sessionData.RouteState.Next || sessionData.RouteNumRelays == 0 {
			sessionData.RouteState.Next = false
			if core.MakeRouteDecision_TakeNetworkNext(routeMatrix.RouteEntries, &buyer.RouteShader, &sessionData.RouteState, multipathVetoMap, &buyer.InternalConfig, int32(packet.DirectRTT), packet.DirectPacketLoss, reframedNearRelays, nearRelayCosts, reframedDestRelays, &routeCost, &routeNumRelays, routeRelays[:]) {
				HandleNextToken(&sessionData, storer, &buyer, &packet, routeNumRelays, routeRelays[:], routeMatrix.RelayIDs, routerPrivateKey, &response, internalIPSellers, enableInternalIPs)
			}
		} else {
			if !core.ReframeRoute(routeMatrix.RelayIDsToIndices, sessionData.RouteRelayIDs[:sessionData.RouteNumRelays], &routeRelays) {
				level.Warn(logger).Log("warn", "one or more relays in the route no longer exist, finding new route.")

				if core.MakeRouteDecision_TakeNetworkNext(routeMatrix.RouteEntries, &buyer.RouteShader, &sessionData.RouteState, multipathVetoMap, &buyer.InternalConfig, int32(packet.DirectRTT), packet.DirectPacketLoss, reframedNearRelays, nearRelayCosts, reframedDestRelays, &routeCost, &routeNumRelays, routeRelays[:]) {
					HandleNextToken(&sessionData, storer, &buyer, &packet, routeNumRelays, routeRelays[:], routeMatrix.RelayIDs, routerPrivateKey, &response, internalIPSellers, enableInternalIPs)
				}
			} else {
				if stay, nextRouteSwitched := core.MakeRouteDecision_StayOnNetworkNext(routeMatrix.RouteEntries, &buyer.RouteShader, &sessionData.RouteState, &buyer.InternalConfig, int32(packet.DirectRTT), int32(packet.NextRTT), packet.DirectPacketLoss, packet.NextPacketLoss, sessionData.RouteNumRelays, routeRelays, reframedNearRelays, nearRelayCosts, reframedDestRelays, &routeCost, &routeNumRelays, routeRelays[:]); stay {
					// Continue token

					// Check if the route has changed
					if nextRouteSwitched {
						// Create a next token here rather than a continue token since the route has switched
						HandleNextToken(&sessionData, storer, &buyer, &packet, routeNumRelays, routeRelays[:], routeMatrix.RelayIDs, routerPrivateKey, &response, internalIPSellers, enableInternalIPs)
					} else {
						HandleContinueToken(&sessionData, storer, &buyer, &packet, routeNumRelays, routeRelays[:], routeMatrix.RelayIDs, routerPrivateKey, &response)
					}
				} else {
					// Route was vetoed - check to see why
					if sessionData.RouteState.NoRoute {
						level.Warn(logger).Log("warn", "route no longer exists")
						metrics.NoRoute.Add(1)
					}

					if sessionData.RouteState.MultipathOverload {
						level.Warn(logger).Log("warn", "multipath overloaded this user's connection", "user_hash", fmt.Sprintf("%016x", sessionData.RouteState.UserID))
						metrics.MultipathOverload.Add(1)

						// We will handle updating the multipath veto redis in the post session update
						// to avoid blocking the routing response
					}

					if sessionData.RouteState.LatencyWorse {
						level.Warn(logger).Log("warn", "this route makes latency worse")
						metrics.LatencyWorse.Add(1)
					}
				}
			}
		}

		response.Committed = sessionData.RouteState.Committed

		// Store the route back into the session data
		sessionData.RouteNumRelays = routeNumRelays
		sessionData.RouteCost = routeCost

		for i := int32(0); i < routeNumRelays; i++ {
			relayID := routeMatrix.RelayIDs[routeRelays[i]]
			sessionData.RouteRelayIDs[i] = relayID

			// Get all of the necessary relay information for the post session update
			relay, err := storer.Relay(relayID)
			if err != nil {
				continue
			}

			routeRelayNames[i] = relay.Name
			routeRelaySellers[i] = relay.Seller
		}

		level.Debug(logger).Log("msg", "session updated successfully", "source_address", incoming.SourceAddr.String(), "server_address", packet.ServerAddress.String(), "client_address", packet.ClientAddress.String())
	}
}

func HandleNextToken(sessionData *SessionData, storer storage.Storer, buyer *routing.Buyer, packet *SessionUpdatePacket, routeNumRelays int32, routeRelays []int32, allRelayIDs []uint64, routerPrivateKey [crypto.KeySize]byte, response *SessionResponsePacket, internalIPSellers []string, enableInternalIPs bool) {
	// Add another 10 seconds to the slice and increment the session version
	sessionData.Initial = true
	sessionData.ExpireTimestamp += billing.BillingSliceSeconds
	sessionData.SessionVersion++

	numTokens := routeNumRelays + 2 // relays + client + server
	routeAddresses, routePublicKeys := GetRouteAddressesAndPublicKeys(&packet.ClientAddress, packet.ClientRoutePublicKey, &packet.ServerAddress, packet.ServerRoutePublicKey, numTokens, routeRelays, allRelayIDs, storer, internalIPSellers, enableInternalIPs)
	if routeAddresses == nil || routePublicKeys == nil {
		response.RouteType = routing.RouteTypeDirect
		response.NumTokens = 0
		response.Tokens = nil
		return
	}

	tokenData := make([]byte, numTokens*routing.EncryptedNextRouteTokenSize)
	core.WriteRouteTokens(tokenData, sessionData.ExpireTimestamp, sessionData.SessionID, uint8(sessionData.SessionVersion), uint32(buyer.RouteShader.BandwidthEnvelopeUpKbps), uint32(buyer.RouteShader.BandwidthEnvelopeDownKbps), int(numTokens), routeAddresses, routePublicKeys, routerPrivateKey)
	response.RouteType = routing.RouteTypeNew
	response.NumTokens = numTokens
	response.Tokens = tokenData
}

func HandleContinueToken(sessionData *SessionData, storer storage.Storer, buyer *routing.Buyer, packet *SessionUpdatePacket, routeNumRelays int32, routeRelays []int32, allRelayIDs []uint64, routerPrivateKey [crypto.KeySize]byte, response *SessionResponsePacket) {
	numTokens := routeNumRelays + 2 // relays + client + server
	// empty string array b/c don't care for internal ips here
	routeAddresses, routePublicKeys := GetRouteAddressesAndPublicKeys(&packet.ClientAddress, packet.ClientRoutePublicKey, &packet.ServerAddress, packet.ServerRoutePublicKey, numTokens, routeRelays, allRelayIDs, storer, []string{}, false)
	if routeAddresses == nil || routePublicKeys == nil {
		response.RouteType = routing.RouteTypeDirect
		response.NumTokens = 0
		response.Tokens = nil
		return
	}

	tokenData := make([]byte, numTokens*routing.EncryptedContinueRouteTokenSize)
	core.WriteContinueTokens(tokenData, sessionData.ExpireTimestamp, sessionData.SessionID, uint8(sessionData.SessionVersion), int(numTokens), routePublicKeys, routerPrivateKey)
	response.RouteType = routing.RouteTypeContinue
	response.NumTokens = numTokens
	response.Tokens = tokenData
}

func GetRouteAddressesAndPublicKeys(clientAddress *net.UDPAddr, clientPublicKey []byte, serverAddress *net.UDPAddr, serverPublicKey []byte, numTokens int32, routeRelays []int32, allRelayIDs []uint64, storer storage.Storer, internalIPSellers []string, enableInternalIPs bool) ([]*net.UDPAddr, [][]byte) {
	routeAddresses := make([]*net.UDPAddr, numTokens)
	routePublicKeys := make([][]byte, numTokens)

	routeAddresses[0] = clientAddress
	routePublicKeys[0] = clientPublicKey
	routeAddresses[numTokens-1] = serverAddress
	routePublicKeys[numTokens-1] = serverPublicKey

	totalNumRelays := int32(len(allRelayIDs))
	foundRelayCount := int32(0)

	for i := int32(0); i < numTokens-2; i++ {
		relayIndex := routeRelays[i]

		for j := int32(0); j < totalNumRelays; j++ {

			if j == relayIndex {
				relayID := allRelayIDs[relayIndex]
				relay, err := storer.Relay(relayID)
				if err != nil {
					continue
				}

				if enableInternalIPs {
					shouldTryUseInternalIPs := false
					for i := range internalIPSellers {
						if internalIPSellers[i] == relay.Seller.Name {
							shouldTryUseInternalIPs = true
						}
					}

					routeAddresses[i+1] = &relay.Addr
					// check if the previous relay is the same seller
					if shouldTryUseInternalIPs && i >= 1 {
						prevIndex := routeRelays[i-1]
						for jprev := int32(0); jprev < totalNumRelays; jprev++ {
							if jprev == prevIndex {
								prevID := allRelayIDs[prevIndex]
								prev, err := storer.Relay(prevID)
								if err == nil && prev.Seller.ID == relay.Seller.ID {
									routeAddresses[i+1] = &relay.InternalAddr
								}
								break
							}
						}
					}
				}

				routePublicKeys[i+1] = relay.PublicKey
				foundRelayCount++
				break
			}
		}
	}

	if foundRelayCount != numTokens-2 {
		return nil, nil
	}

	return routeAddresses, routePublicKeys
}

func PostSessionUpdate(postSessionHandler *PostSessionHandler, packet *SessionUpdatePacket, sessionData *SessionData, buyer *routing.Buyer, multipathVetoHandler *storage.MultipathVetoHandler, routeRelayNames [routing.MaxRelays]string, routeRelaySellers [routing.MaxRelays]routing.Seller, nearRelays []routing.NearRelayData, datacenter *routing.Datacenter) {
	sliceDuration := uint64(billing.BillingSliceSeconds)
	if sessionData.Initial {
		sliceDuration *= 2
	}
	nextBytesUp, nextBytesDown := CalculateNextBytesUpAndDown(uint64(packet.NextKbpsUp), uint64(packet.NextKbpsDown), sliceDuration)
	nextEnvelopeBytesUp, nextEnvelopeBytesDown := CalculateNextBytesUpAndDown(uint64(buyer.RouteShader.BandwidthEnvelopeUpKbps), uint64(buyer.RouteShader.BandwidthEnvelopeDownKbps), sliceDuration)
	totalPrice := CalculateTotalPriceNibblins(int(sessionData.RouteNumRelays), routeRelaySellers, nextEnvelopeBytesUp, nextEnvelopeBytesDown)
	routeRelayPrices := CalculateRouteRelaysPrice(int(sessionData.RouteNumRelays), routeRelaySellers, nextEnvelopeBytesUp, nextEnvelopeBytesDown)

	// Check if we should multipath veto the user
	if packet.Next && sessionData.RouteState.MultipathOverload {
		if err := multipathVetoHandler.MultipathVetoUser(buyer.CompanyCode, packet.UserHash); err != nil {
			level.Error(postSessionHandler.logger).Log("err", err)
		}
	}

	nextRelaysPrice := [routing.MaxRelays]uint64{}
	for i := 0; i < routing.MaxRelays; i++ {
		nextRelaysPrice[i] = uint64(routeRelayPrices[i])
	}

	packetLossClientToServer := float32(packet.PacketsLostClientToServer) / float32(packet.PacketsSentClientToServer) * 100.0
	packetLossServerToClient := float32(packet.PacketsLostServerToClient) / float32(packet.PacketsSentServerToClient) * 100.0

	// Take the max of packet loss client -> server or server -> client
	inGamePacketLoss := packetLossClientToServer
	if inGamePacketLoss < packetLossServerToClient {
		inGamePacketLoss = packetLossServerToClient
	}

	multipathVetoMap := multipathVetoHandler.GetMapCopy(buyer.CompanyCode)
	var multipathVetoed bool
	if _, ok := multipathVetoMap[packet.UserHash]; ok {
		multipathVetoed = true
	}

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
		NumNextRelays:             uint8(sessionData.RouteNumRelays),
		NextRelays:                sessionData.RouteRelayIDs,
		TotalPrice:                uint64(totalPrice),
		ClientToServerPacketsLost: packet.PacketsLostClientToServer,
		ServerToClientPacketsLost: packet.PacketsLostServerToClient,
		Committed:                 packet.Committed,
		Flagged:                   packet.Reported,
		Multipath:                 sessionData.RouteState.Multipath,
		Initial:                   sessionData.Initial,
		NextBytesUp:               nextBytesUp,
		NextBytesDown:             nextBytesDown,
		EnvelopeBytesUp:           nextEnvelopeBytesUp,
		EnvelopeBytesDown:         nextEnvelopeBytesDown,
		DatacenterID:              datacenter.ID,
		RTTReduction:              sessionData.RouteState.ReduceLatency,
		PacketLossReduction:       sessionData.RouteState.ReducePacketLoss,
		NextRelaysPrice:           nextRelaysPrice,
		Latitude:                  float32(sessionData.Location.Latitude),
		Longitude:                 float32(sessionData.Location.Longitude),
		ISP:                       sessionData.Location.ISP,
		ABTest:                    sessionData.RouteState.ABTest,
		RouteDecision:             0,
		ConnectionType:            uint8(packet.ConnectionType),
		PlatformType:              uint8(packet.PlatformType),
		SDKVersion:                packet.Version.String(),
		PacketLoss:                inGamePacketLoss,
		PredictedNextRTT:          float32(sessionData.RouteCost),
		MultipathVetoed:           multipathVetoed,
	}

	postSessionHandler.SendBillingEntry(billingEntry)

	hops := make([]RelayHop, sessionData.RouteNumRelays)
	for i := int32(0); i < sessionData.RouteNumRelays; i++ {
		hops[i] = RelayHop{
			ID:   sessionData.RouteRelayIDs[i],
			Name: routeRelayNames[i],
		}
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
			Location:        sessionData.Location,
			ClientAddr:      packet.ClientAddress.String(),
			ServerAddr:      packet.ServerAddress.String(),
			Hops:            hops,
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
				Up:   int64(packet.NextKbpsUp),
				Down: int64(packet.NextKbpsDown),
			},
			IsMultiPath:       sessionData.RouteState.Multipath,
			IsTryBeforeYouBuy: !sessionData.RouteState.Committed,
			OnNetworkNext:     packet.Next,
		},
		Point: SessionMapPoint{
			Latitude:  sessionData.Location.Latitude,
			Longitude: sessionData.Location.Longitude,
		},
		LargeCustomer: buyer.InternalConfig.LargeCustomer,
		EverOnNext:    sessionData.EverOnNext,
	}

	if portalData.Meta.NextRTT != 0 || portalData.Meta.DirectRTT != 0 {
		postSessionHandler.SendPortalData(portalData)
	}
}

func CalculateNextBytesUpAndDown(kbpsUp uint64, kbpsDown uint64, sliceDuration uint64) (uint64, uint64) {
	bytesUp := (((1000 * kbpsUp) / 8) * sliceDuration)
	bytesDown := (((1000 * kbpsDown) / 8) * sliceDuration)
	return bytesUp, bytesDown
}

func CalculateTotalPriceNibblins(routeNumRelays int, relaySellers [routing.MaxRelays]routing.Seller, envelopeBytesUp uint64, envelopeBytesDown uint64) routing.Nibblin {

	if routeNumRelays == 0 {
		return 0
	}

	envelopeUpGB := float64(envelopeBytesUp) / 1000000000.0
	envelopeDownGB := float64(envelopeBytesDown) / 1000000000.0

	sellerPriceNibblinsPerGB := routing.Nibblin(0)
	for _, seller := range relaySellers {
		sellerPriceNibblinsPerGB += seller.EgressPriceNibblinsPerGB
	}

	nextPriceNibblinsPerGB := routing.Nibblin(1e9)
	totalPriceNibblins := float64(sellerPriceNibblinsPerGB+nextPriceNibblinsPerGB) * (envelopeUpGB + envelopeDownGB)

	return routing.Nibblin(totalPriceNibblins)
}

func CalculateRouteRelaysPrice(routeNumRelays int, relaySellers [routing.MaxRelays]routing.Seller, envelopeBytesUp uint64, envelopeBytesDown uint64) [routing.MaxRelays]routing.Nibblin {
	relayPrices := [routing.MaxRelays]routing.Nibblin{}

	if routeNumRelays == 0 {
		return relayPrices
	}

	envelopeUpGB := float64(envelopeBytesUp) / 1000000000.0
	envelopeDownGB := float64(envelopeBytesDown) / 1000000000.0

	for i := 0; i < len(relayPrices); i++ {
		relayPriceNibblins := float64(relaySellers[i].EgressPriceNibblinsPerGB) * (envelopeUpGB + envelopeDownGB)
		relayPrices[i] = routing.Nibblin(relayPriceNibblins)
	}

	return relayPrices
}
