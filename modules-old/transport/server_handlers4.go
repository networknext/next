package transport

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/networknext/backend/modules/core"

	"github.com/networknext/backend/modules-old/crypto_old"
	"github.com/networknext/backend/modules-old/metrics"
	"github.com/networknext/backend/modules-old/routing"
	"github.com/networknext/backend/modules-old/storage"
)

const (
	UDPIPPacketHeaderSize = 28 // IP: 20, UDP: 8
)

type UDPPacket struct {
	From net.UDPAddr
	Data []byte
}

type UDPHandlerFunc func(io.Writer, *UDPPacket)

func datacenterExists(database *routing.DatabaseBinWrapper, datacenterID uint64) bool {
	_, exists := database.DatacenterMap[datacenterID]
	return exists
}

func datacenterEnabled(database *routing.DatabaseBinWrapper, buyerID uint64, datacenterID uint64) bool {
	// Get all datacenters for buyer
	datacenterAliases, ok := database.DatacenterMaps[buyerID]
	if !ok {
		return false
	}

	// Check if the datacenter in question is linked to the buyer
	_, ok = datacenterAliases[datacenterID]

	return ok
}

func accelerateDatacenter(database *routing.DatabaseBinWrapper, buyerID uint64, datacenterID uint64) bool {
	// Get all datacenters for buyer
	datacenterAliases, ok := database.DatacenterMaps[buyerID]
	if !ok {
		return false
	}

	// Check if the datacenter in question is linked to the buyer
	dcMap, ok := datacenterAliases[datacenterID]
	if !ok {
		return false
	}

	return dcMap.EnableAcceleration
}

func getDatacenter(database *routing.DatabaseBinWrapper, datacenterID uint64) routing.Datacenter {
	value, exists := database.DatacenterMap[datacenterID]

	if !exists {
		return routing.UnknownDatacenter
	}

	return value
}

func writeServerInitResponse(w io.Writer, packet *ServerInitRequestPacket, response uint32) error {
	responsePacket := ServerInitResponsePacket{
		RequestID: packet.RequestID,
		Response:  response,
	}
	responsePacketData, err := MarshalPacket(&responsePacket)
	if err != nil {
		return err
	}
	packetHeader := append([]byte{PacketTypeServerInitResponse}, make([]byte, crypto_old.PacketHashSize)...)
	responseData := append(packetHeader, responsePacketData...)
	if _, err := w.Write(responseData); err != nil {
		return err
	}
	return nil
}

func writeMatchDataResponse(w io.Writer, packet *MatchDataRequestPacket, response uint32) error {
	responsePacket := MatchDataResponsePacket{
		SessionID: packet.SessionID,
		Response:  response,
	}
	responsePacketData, err := MarshalPacket(&responsePacket)
	if err != nil {
		return err
	}
	packetHeader := append([]byte{PacketTypeMatchDataResponse}, make([]byte, crypto_old.PacketHashSize)...)
	responseData := append(packetHeader, responsePacketData...)
	if _, err := w.Write(responseData); err != nil {
		return err
	}
	return nil
}

// ----------------------------------------------------------------------------

func ServerInitHandlerFunc(getDatabase func() *routing.DatabaseBinWrapper, ServerTracker *storage.ServerTracker, metrics *metrics.ServerInitMetrics) UDPHandlerFunc {

	return func(w io.Writer, incoming *UDPPacket) {

		core.Debug("-----------------------------------------")
		core.Debug("server init packet from %s", incoming.From.String())

		metrics.HandlerMetrics.Invocations.Add(1)

		timeStart := time.Now()
		defer func() {
			milliseconds := float64(time.Since(timeStart).Milliseconds())
			metrics.HandlerMetrics.Duration.Set(milliseconds)
			if milliseconds > 100 {
				metrics.HandlerMetrics.LongDuration.Add(1)
			}
			core.Debug("server init duration: %fms\n-----------------------------------------", milliseconds)
		}()

		var packet ServerInitRequestPacket
		if err := UnmarshalPacket(&packet, incoming.Data); err != nil {
			core.Debug("could not read server init packet:\n\n%v\n", err)
			metrics.ReadPacketFailure.Add(1)
			return
		}

		core.Debug("server buyer id is %x", packet.BuyerID)

		database := getDatabase()

		responseType := InitResponseOK

		defer func() {
			if err := writeServerInitResponse(w, &packet, uint32(responseType)); err != nil {
				core.Debug("failed to write server init response: %s", err)
				metrics.WriteResponseFailure.Add(1)
			}
		}()

		buyer, exists := database.BuyerMap[packet.BuyerID]
		if !exists {
			core.Debug("unknown buyer")
			metrics.BuyerNotFound.Add(1)
			responseType = InitResponseUnknownBuyer
			return
		}

		if !buyer.Live {
			core.Debug("buyer not active")
			metrics.BuyerNotActive.Add(1)
			responseType = InitResponseBuyerNotActive
			return
		}

		if !crypto_old.VerifyPacket(buyer.PublicKey, incoming.Data) {
			core.Debug("signature check failed")
			metrics.SignatureCheckFailed.Add(1)
			responseType = InitResponseSignatureCheckFailed
			return
		}

		if !packet.Version.AtLeast(SDKVersion{4, 0, 0}) {
			core.Debug("sdk version is too old: %s", packet.Version.String())
			metrics.SDKTooOld.Add(1)
			responseType = InitResponseOldSDKVersion
			return
		}

		// Track which servers are initing
		// This is where we get the datacenter name
		if strings.TrimSpace(packet.DatacenterName) == "" {
			ServerTracker.AddServer(packet.BuyerID, packet.DatacenterID, incoming.From, "unknown_init")
		} else {
			ServerTracker.AddServer(packet.BuyerID, packet.DatacenterID, incoming.From, packet.DatacenterName)
		}

		/*
		   IMPORTANT: When the datacenter doesn't exist, we intentionally let the server init succeed anyway
		   and just log here, so we can map the datacenter name to the datacenter id, when we are tracking it down.
		*/

		if !datacenterExists(database, packet.DatacenterID) {
			// core.Error("unknown datacenter %s [%016x, %s] for buyer id %016x", packet.DatacenterName, packet.DatacenterID, incoming.From.String(), packet.BuyerID)
			metrics.DatacenterNotFound.Add(1)
			return
		}

		core.Debug("server is in datacenter \"%s\" [%x]", packet.DatacenterName, packet.DatacenterID)

		core.Debug("server initialized successfully")
	}
}

// ----------------------------------------------------------------------------

func ServerTrackerHandlerFunc(serverTracker *storage.ServerTracker) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		w.Header().Set("Content-Type", "application/json")

		serverTracker.TrackerMutex.RLock()
		defer serverTracker.TrackerMutex.RUnlock()
		json.NewEncoder(w).Encode(serverTracker.Tracker)
	}
}

// ----------------------------------------------------------------------------
