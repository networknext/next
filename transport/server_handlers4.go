package transport

import (
	"errors"
	"io"
	"net"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/metrics"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
)

func writeServerInitResponse4(w io.Writer, packet *ServerInitRequestPacket4, response uint8) error {
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

func writeSessionResponse4(w io.Writer, packet *SessionUpdatePacket4, routeType int32, nearRelays []routing.NearRelayData, sessionData *SessionData4) error {
	sessionData.Version = SessionDataVersion4
	sessionData.SessionID = packet.SessionID
	sessionData.SliceNumber = uint32(packet.SliceNumber + 1)

	sessionDataBuffer, err := MarshalSessionData(sessionData)
	if err != nil {
		return err
	}

	if len(sessionDataBuffer) > MaxSessionDataSize {
		return errors.New("session data too large")
	}

	numNearRelays := int32(len(nearRelays))
	nearRelayIDs := make([]uint64, numNearRelays)
	nearRelayAddrs := make([]net.UDPAddr, numNearRelays)
	for i := int32(0); i < numNearRelays; i++ {
		nearRelayIDs[i] = nearRelays[i].ID
		nearRelayAddrs[i] = *nearRelays[i].Addr
	}

	responsePacket := SessionResponsePacket4{
		SessionID:          packet.SessionID,
		SliceNumber:        packet.SliceNumber,
		SessionDataBytes:   int32(len(sessionDataBuffer)),
		RouteType:          routeType,
		NumNearRelays:      numNearRelays,
		NearRelayIds:       nearRelayIDs,
		NearRelayAddresses: nearRelayAddrs,
		NumTokens:          0,
		Tokens:             nil,
		Multipath:          false,
		Committed:          false,
	}
	copy(responsePacket.SessionData[:], sessionDataBuffer)

	responseData, err := MarshalPacket(&responsePacket)
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

func SessionUpdateHandlerFunc4(logger log.Logger, getIPLocator func() routing.IPLocator, getRouteProvider func() RouteProvider, metrics *metrics.SessionMetrics) UDPHandlerFunc {
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

		var sessionData SessionData4
		location := routing.LocationNullIsland
		ipLocator := getIPLocator()
		routeMatrix := getRouteProvider()
		var nearRelays []routing.NearRelayData
		var err error

		newSession := packet.SliceNumber == 0
		if newSession {
			location, err = ipLocator.LocateIP(incoming.SourceAddr.IP)
			if err != nil {
				level.Error(logger).Log("msg", "failed to locate IP", "err", err)
				metrics.ErrorMetrics.ClientLocateFailure.Add(1)
				writeSessionResponse4(w, &packet, routing.RouteTypeDirect, nearRelays, &sessionData)
				return
			}

			nearRelays, err = routeMatrix.GetNearRelays(location.Latitude, location.Longitude, MaxNearRelays)
			if err != nil {
				level.Error(logger).Log("msg", "failed to get near relays", "err", err)
				metrics.ErrorMetrics.NearRelaysLocateFailure.Add(1)
				writeSessionResponse4(w, &packet, routing.RouteTypeDirect, nearRelays, &sessionData)
				return
			}
		}

		if !newSession {
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

			// todo: apply near relay stats from client -> relay

			// for i, nearRelay := range nearRelays {
			// 	for j, clientNearRelayID := range packet. {
			// 		if nearRelay.ID == clientNearRelayID {
			// 			nearRelayData[i].ClientStats.RTT = float64(packet.NearRelayMinRTT[j])
			// 			nearRelayData[i].ClientStats.Jitter = float64(packet.NearRelayJitter[j])
			// 			nearRelayData[i].ClientStats.PacketLoss = float64(packet.NearRelayPacketLoss[j])
			// 		}
			// 	}
			// }
		}

		// For now, only send back direct routes
		writeSessionResponse4(w, &packet, routing.RouteTypeDirect, nearRelays, &sessionData)

		level.Debug(logger).Log("msg", "successfully sent direct route")
	}
}
