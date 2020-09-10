package transport

import (
	"errors"
	"io"
	"net"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/networknext/backend/billing"
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

func SessionUpdateHandlerFunc4(logger log.Logger, getIPLocator func() routing.IPLocator, getRouteProvider func() RouteProvider, routerPrivateKey []byte, postSessionHandler *PostSessionHandler, metrics *metrics.SessionMetrics) UDPHandlerFunc {
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

		// todo: use 20 seconds if the previous slice was an initial route slice
		usageBytesUp, usageBytesDown := CalculateNextBytesUpAndDown(uint64(packet.KbpsUp), uint64(packet.KbpsDown), billing.BillingSliceSeconds)

		var sessionData SessionData4

		ipLocator := getIPLocator()
		routeMatrix := getRouteProvider()
		location := routing.LocationNullIsland
		var nearRelays []routing.NearRelayData
		route := routing.Route{}
		var err error

		response := SessionResponsePacket4{
			SessionID:   packet.SessionID,
			SliceNumber: packet.SliceNumber,
			RouteType:   routing.RouteTypeDirect,
		}

		if newSession {
			sessionData.Version = SessionDataVersion4
			sessionData.SessionID = packet.SessionID
			sessionData.SliceNumber = uint32(packet.SliceNumber + 1)

			location, err = ipLocator.LocateIP(packet.ClientAddress.IP)
			if err != nil {
				level.Error(logger).Log("msg", "failed to locate IP", "err", err)
				metrics.ErrorMetrics.ClientLocateFailure.Add(1)
				if err := writeSessionResponse4(w, &response, &sessionData); err != nil {
					level.Error(logger).Log("msg", "failed to write session update response", "err", err)
					metrics.ErrorMetrics.WriteResponseFailure.Add(1)
					return
				}

				go PostSessionUpdate4(postSessionHandler, &packet, &location, nearRelays, usageBytesUp, usageBytesDown, metrics)
				return
			}

			nearRelays, err = routeMatrix.GetNearRelays(location.Latitude, location.Longitude, MaxNearRelays)
			if err != nil {
				level.Error(logger).Log("msg", "failed to get near relays", "err", err)
				metrics.ErrorMetrics.NearRelaysLocateFailure.Add(1)
				if err := writeSessionResponse4(w, &response, &sessionData); err != nil {
					level.Error(logger).Log("msg", "failed to write session update response", "err", err)
					metrics.ErrorMetrics.WriteResponseFailure.Add(1)
					return
				}

				go PostSessionUpdate4(postSessionHandler, &packet, &location, nearRelays, usageBytesUp, usageBytesDown, metrics)
				return
			}

			// todo: fix this
			response.NumNearRelays = /*int32(len(nearRelays))*/ 0
			response.NearRelayIDs = make([]uint64, response.NumNearRelays)
			response.NearRelayAddresses = make([]net.UDPAddr, response.NumNearRelays)
			for i := int32(0); i < response.NumNearRelays; i++ {
				response.NearRelayIDs[i] = nearRelays[i].ID
				response.NearRelayAddresses[i] = *nearRelays[i].Addr
			}
		} else {
			if err := UnmarshalSessionData(&sessionData, packet.SessionData[:]); err != nil {
				level.Error(logger).Log("msg", "could not read session data in session update packet", "err", err)
				metrics.ErrorMetrics.ReadPacketFailure.Add(1)
				return
			}

			if sessionData.SessionID != packet.SessionID {
				level.Error(logger).Log("err", "bad session ID in session data")
				metrics.SessionDataMetrics.BadSessionID.Add(1)
				return
			}

			if sessionData.SliceNumber != uint32(packet.SliceNumber) {
				level.Error(logger).Log("err", "bad sequence number in session data")
				metrics.SessionDataMetrics.BadSequenceNumber.Add(1)
				return
			}

			sessionData.SliceNumber = uint32(packet.SliceNumber + 1)

			// todo: we need to store the near relays in the session data so that we can update their client stats

			// for i := range nearRelays {
			// 	for j, clientNearRelayID := range packet.NearRelayIDs {
			// 		if nearRelays[i].ID == clientNearRelayID {
			// 			nearRelays[i].ClientStats.RTT = float64(packet.NearRelayRTT[j])
			// 			nearRelays[i].ClientStats.Jitter = float64(packet.NearRelayJitter[j])
			// 			nearRelays[i].ClientStats.PacketLoss = float64(packet.NearRelayPacketLoss[j])
			// 		}
			// 	}
			// }
		}

		var numRouteTokens int
		var routeTokens []byte

		if response.RouteType != routing.RouteTypeDirect {
			var token routing.Token

			relayTokens := make([]routing.RelayToken, route.NumRelays)
			for i := range relayTokens {
				relayTokens[i] = routing.RelayToken{
					ID:        route.RelayIDs[i],
					Addr:      route.RelayAddrs[i],
					PublicKey: route.RelayPublicKeys[i],
				}
			}

			if route.Equals(sessionData.Route) {
				token = &routing.ContinueRouteToken{
					Expires: uint64(time.Now().Unix() + 15),

					SessionID: packet.SessionID,

					SessionVersion: uint8(sessionData.SessionVersion),

					Client: routing.Client{
						Addr:      packet.ClientAddress,
						PublicKey: packet.ClientRoutePublicKey,
					},

					Server: routing.Server{
						Addr:      packet.ServerAddress,
						PublicKey: packet.ServerRoutePublicKey,
					},

					Relays: relayTokens,
				}
			} else {
				sessionData.SessionVersion++
				token = &routing.NextRouteToken{
					Expires: uint64(time.Now().Unix() + 15),

					SessionID: packet.SessionID,

					SessionVersion: uint8(sessionData.SessionVersion),

					// todo: add back when buyer lookup is implemented in the session update handler

					// KbpsUp:   uint32(buyer.RoutingRulesSettings.EnvelopeKbpsUp),
					// KbpsDown: uint32(buyer.RoutingRulesSettings.EnvelopeKbpsDown),

					Client: routing.Client{
						Addr:      packet.ClientAddress,
						PublicKey: packet.ClientRoutePublicKey,
					},

					Server: routing.Server{
						Addr:      packet.ServerAddress,
						PublicKey: packet.ServerRoutePublicKey,
					},

					Relays: relayTokens,
				}
			}

			routeTokens, numRouteTokens, err = token.Encrypt(routerPrivateKey)
			if err != nil {
				metrics.ErrorMetrics.EncryptionFailure.Add(1)
				return
			}

			response.RouteType = int32(token.Type())
			response.NumTokens = int32(numRouteTokens)
			response.Tokens = routeTokens
		}

		if err := writeSessionResponse4(w, &response, &sessionData); err != nil {
			level.Error(logger).Log("msg", "failed to write session update response", "err", err)
			metrics.ErrorMetrics.WriteResponseFailure.Add(1)
			return
		}

		go PostSessionUpdate4(postSessionHandler, &packet, &location, nearRelays, usageBytesUp, usageBytesDown, metrics)

		level.Debug(logger).Log("msg", "session updated successfully", "source_address", incoming.SourceAddr.String(), "server_address", packet.ServerAddress.String(), "client_address", packet.ClientAddress.String())
	}
}

func PostSessionUpdate4(postSessionHandler *PostSessionHandler, packet *SessionUpdatePacket4, location *routing.Location, nearRelays []routing.NearRelayData, nextBytesUp uint64, nextBytesDown uint64, metrics *metrics.SessionMetrics) {
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
		DatacenterID:              0,
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
			DatacenterName:  "local", // todo: we need the datacenter ID or name in the session update packet
			DatacenterAlias: "local",
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
