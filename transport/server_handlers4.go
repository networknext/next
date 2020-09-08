package transport

import (
	"errors"
	"io"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
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
			}

			return
		}

		if !packet.Version.AtLeast(SDKVersion{4, 0, 0}) && !buyer.Debug {
			level.Error(logger).Log("err", "sdk too old", "version", packet.Version.String())
			metrics.ErrorMetrics.SDKTooOld.Add(1)

			writeServerInitResponse4(w, &packet, InitResponseOldSDKVersion)
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

						writeServerInitResponse4(w, &packet, InitResponseUnknownDatacenter)
						return
					}

					datacenter.AliasName = dcMap.Alias
					break
				}
			}
		}

		writeServerInitResponse4(w, &packet, InitResponseOK)
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

func SessionUpdateHandlerFunc4(logger log.Logger, getIPLocator func() routing.IPLocator, getRouteProvider func() RouteProvider, routerPrivateKey []byte, metrics *metrics.SessionMetrics) UDPHandlerFunc {
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
		routeMatrix := getRouteProvider()
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

			location, err := ipLocator.LocateIP(incoming.SourceAddr.IP)
			if err != nil {
				level.Error(logger).Log("msg", "failed to locate IP", "err", err)
				metrics.ErrorMetrics.ClientLocateFailure.Add(1)
				writeSessionResponse4(w, &response, &sessionData)
				return
			}

			nearRelays, err := routeMatrix.GetNearRelays(location.Latitude, location.Longitude, MaxNearRelays)
			if err != nil {
				level.Error(logger).Log("msg", "failed to get near relays", "err", err)
				metrics.ErrorMetrics.NearRelaysLocateFailure.Add(1)
				writeSessionResponse4(w, &response, &sessionData)
				return
			}

			response.NumNearRelays = int32(len(nearRelays))
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

		writeSessionResponse4(w, &response, &sessionData)
	}
}
