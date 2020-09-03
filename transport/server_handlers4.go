package transport

import (
	"io"
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

func writeSessionResponse4(w io.Writer, packet *SessionUpdatePacket4) error {
	responsePacket := SessionResponsePacket4{
		Sequence:             packet.Sequence,
		SessionID:            packet.SessionID,
		NumNearRelays:        0,
		NearRelayIDs:         nil,
		NearRelayAddresses:   nil,
		RouteType:            routing.RouteTypeDirect,
		Multipath:            false,
		Committed:            false,
		NumTokens:            0,
		Tokens:               nil,
		ServerRoutePublicKey: packet.ServerRoutePublicKey,
		SessionDataBytes:     0,
		SessionData:          [MaxSessionDataSize]byte{},
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

		level.Debug(logger).Log("msg", "server updated successfully", "server_address", packet.ServerAddress.String(), "sequence", packet.Sequence)
	}
}

func SessionUpdateHandlerFunc4(logger log.Logger, storer storage.Storer, metrics *metrics.SessionMetrics) UDPHandlerFunc {
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

		// For now, only send back direct routes
		writeSessionResponse4(w, &packet)

		level.Debug(logger).Log("msg", "successfully sent direct route")
	}
}
