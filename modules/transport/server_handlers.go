package transport

import (
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"sort"
	"strings"
	"time"

	"github.com/networknext/backend/modules/envvar"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/networknext/backend/modules/billing"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/storage"
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
		return fmt.Errorf("session data of %d exceeds limit of %d bytes", len(sessionDataBuffer), MaxSessionDataSize)
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

type ErrDatacenterNotFound struct {
	buyer          uint64
	datacenter     uint64
	datacenterName string
}

func (e ErrDatacenterNotFound) Error() string {
	if e.datacenterName != "" {
		return fmt.Sprintf("datacenter %s for buyer %016x not found", e.datacenterName, e.buyer)
	}

	return fmt.Sprintf("datacenter %016x for buyer %016x not found", e.datacenter, e.buyer)
}

type ErrDatacenterMapMisconfigured struct {
	buyer         uint64
	datacenterMap routing.DatacenterMap
}

func (e ErrDatacenterMapMisconfigured) Error() string {
	return fmt.Sprintf("datacenter alias %s misconfigured for buyer %016x: mapped to datacenter \"%016x\" which doesn't exist", e.datacenterMap.Alias, e.buyer, e.datacenterMap.DatacenterID)
}

type ErrDatacenterNotAllowed struct {
	buyer          uint64
	datacenter     uint64
	datacenterName string
}

func (e ErrDatacenterNotAllowed) Error() string {
	if e.datacenterName != "" {
		return fmt.Sprintf("buyer %016x tried to use datacenter %s when they are not configured to do so", e.buyer, e.datacenterName)
	}

	return fmt.Sprintf("buyer %016x tried to use datacenter %016x when they are not configured to do so", e.buyer, e.datacenter)
}

func getDatacenter(storer storage.Storer, buyerID uint64, datacenterID uint64, datacenterName string) (routing.Datacenter, error) {
	// We should always support the "local" datacenter, even without a datacenter mapping
	if crypto.HashID("local") == datacenterID {
		return routing.Datacenter{
			ID:   crypto.HashID("local"),
			Name: "local",
		}, nil
	}

	// enforce that whatever datacenter the server says it's in, we have a mapping for it
	datacenterAliases := storer.GetDatacenterMapsForBuyer(buyerID)
	for _, dcMap := range datacenterAliases {
		if datacenterID == dcMap.DatacenterID {
			// We found the datacenter
			datacenter, err := storer.Datacenter(datacenterID)
			if err != nil {
				// The datacenter map is misconfigured in our database
				fmt.Printf("Datacenter map misconfigured: BuyerID: %016x, DatacenterMap: %s\n", buyerID, dcMap.String())
				return routing.UnknownDatacenter, ErrDatacenterMapMisconfigured{buyerID, dcMap}
			}

			return datacenter, nil
		}

		if datacenterID == crypto.HashID(dcMap.Alias) {
			// We found the datacenter from the mapped alias
			datacenter, err := storer.Datacenter(dcMap.DatacenterID)
			if err != nil {
				// The datacenter map is misconfigured in our database
				fmt.Printf("Datacenter map misconfigured: BuyerID: %016x, DatacenterMap: %s\n", buyerID, dcMap.String())
				return routing.UnknownDatacenter, ErrDatacenterMapMisconfigured{buyerID, dcMap}
			}

			datacenter.AliasName = dcMap.Alias
			return datacenter, nil
		}
	}

	// We couldn't find the datacenter, check if it is a datacenter that we have in our database
	_, err := storer.Datacenter(datacenterID)
	if err != nil {
		// This isn't a datacenter we know about. It's either brand new and not configured yet
		// or there is a typo in the server's integration of the SDK
		fmt.Printf("Datacenter not found: DatacenterID: %016x, BuyerID: %016x, DatacenterName: %s\n", datacenterID, buyerID, datacenterName)
		return routing.UnknownDatacenter, ErrDatacenterNotFound{buyerID, datacenterID, datacenterName}
	}

	// This is a datacenter we know about, but the buyer isn't configured to use it
	fmt.Printf("Datacenter use not allowed: DatacenterID: %016x, BuyerID: %016x, DatacenterName: %s\n", datacenterID, buyerID, datacenterName)
	return routing.UnknownDatacenter, ErrDatacenterNotAllowed{buyerID, datacenterID, datacenterName}
}

type nearRelayGroup struct {
	Count        int32
	IDs          []uint64
	Addrs        []net.UDPAddr
	Names        []string
	RTTs         []int32
	Jitters      []int32
	PacketLosses []int32
}

func newNearRelayGroup(count int32) nearRelayGroup {
	return nearRelayGroup{
		Count:        count,
		IDs:          make([]uint64, count),
		Addrs:        make([]net.UDPAddr, count),
		Names:        make([]string, count),
		RTTs:         make([]int32, count),
		Jitters:      make([]int32, count),
		PacketLosses: make([]int32, count),
	}
}

func (n nearRelayGroup) Copy(other *nearRelayGroup) {
	other.Count = n.Count
	other.IDs = make([]uint64, n.Count)
	other.Addrs = make([]net.UDPAddr, n.Count)
	other.Names = make([]string, n.Count)
	other.RTTs = make([]int32, n.Count)
	other.Jitters = make([]int32, n.Count)
	other.PacketLosses = make([]int32, n.Count)

	copy(other.IDs, n.IDs)
	copy(other.Addrs, n.Addrs)
	copy(other.Names, n.Names)
	copy(other.RTTs, n.RTTs)
	copy(other.Jitters, n.Jitters)
	copy(other.PacketLosses, n.PacketLosses)
}

func handleNearAndDestRelays(
	sliceNumber int32,
	routeMatrix *routing.RouteMatrix,
	incomingNearRelays nearRelayGroup,
	routeShader *core.RouteShader,
	routeState *core.RouteState,
	newSession bool,
	clientLat float32,
	clientLong float32,
	serverLat float32,
	serverLong float32,
	maxNearRelays int,
	directLatency int32,
	directJitter int32,
	directPacketLoss int32,
	nextPacketLoss int32,
	firstRouteRelayID uint64,
	destRelayIDs []uint64,
	debug *string,
) (bool, nearRelayGroup, []int32, error) {
	if newSession {
		nearRelayIDs := routeMatrix.GetNearRelays(float32(directLatency), clientLat, clientLong, serverLat, serverLong, maxNearRelays)
		if len(nearRelayIDs) == 0 {
			return false, nearRelayGroup{}, nil, errors.New("no near relays")
		}

		nearRelays := newNearRelayGroup(int32(len(nearRelayIDs)))
		for i := int32(0); i < nearRelays.Count; i++ {
			relayIndex, ok := routeMatrix.RelayIDsToIndices[nearRelayIDs[i]]
			if !ok {
				continue
			}

			nearRelays.IDs[i] = nearRelayIDs[i]
			nearRelays.Addrs[i] = routeMatrix.RelayAddresses[relayIndex]
			nearRelays.Names[i] = routeMatrix.RelayNames[relayIndex]
		}

		routeState.NumNearRelays = nearRelays.Count
		return true, nearRelays, nil, nil
	}

	var nearRelays nearRelayGroup
	incomingNearRelays.Copy(&nearRelays)

	if nearRelays.Count != routeState.NumNearRelays {
		return false, nearRelayGroup{}, nil, fmt.Errorf("near relays changed from %d to %d", routeState.NumNearRelays, nearRelays.Count)
	}

	var numDestRelays int32
	reframedDestRelays := make([]int32, len(destRelayIDs))

	core.ReframeRelays(routeShader, routeState, routeMatrix.RelayIDsToIndices, directLatency, directJitter, directPacketLoss, nextPacketLoss, firstRouteRelayID, sliceNumber, incomingNearRelays.IDs, incomingNearRelays.RTTs, incomingNearRelays.Jitters, incomingNearRelays.PacketLosses, destRelayIDs, nearRelays.RTTs, nearRelays.Jitters, &numDestRelays, reframedDestRelays)

	return false, nearRelays, reframedDestRelays[:numDestRelays], nil
}

func ServerInitHandlerFunc(logger log.Logger, storer storage.Storer, metrics *metrics.ServerInitMetrics) UDPHandlerFunc {
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

		if !crypto.VerifyPacket(buyer.PublicKey, incoming.Data) {
			level.Error(logger).Log("err", "signature check failed", "customerID", packet.CustomerID)
			metrics.SignatureCheckFailed.Add(1)

			if err := writeServerInitResponse(w, &packet, InitResponseSignatureCheckFailed); err != nil {
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

		if _, err := getDatacenter(storer, packet.CustomerID, packet.DatacenterID, packet.DatacenterName); err != nil {
			level.Error(logger).Log("handler", "server_init", "err", err)

			switch err.(type) {
			case ErrDatacenterNotFound:
				metrics.DatacenterNotFound.Add(1)

			case ErrDatacenterMapMisconfigured:
				metrics.MisconfiguredDatacenterAlias.Add(1)

			case ErrDatacenterNotAllowed:
				metrics.DatacenterNotAllowed.Add(1)
			}

			if err := writeServerInitResponse(w, &packet, InitResponseUnknownDatacenter); err != nil {
				level.Error(logger).Log("msg", "failed to write server init response", "err", err)
				metrics.WriteResponseFailure.Add(1)
			}
			return
		}

		if err := writeServerInitResponse(w, &packet, InitResponseOK); err != nil {
			level.Error(logger).Log("msg", "failed to write server init response", "err", err)
			metrics.WriteResponseFailure.Add(1)
			return
		}

		level.Debug(logger).Log("msg", "server initialized successfully", "source_address", incoming.SourceAddr.String())
	}
}

func ServerUpdateHandlerFunc(logger log.Logger, storer storage.Storer, postSessionHandler *PostSessionHandler, metrics *metrics.ServerUpdateMetrics) UDPHandlerFunc {
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

		if !crypto.VerifyPacket(buyer.PublicKey, incoming.Data) {
			level.Error(logger).Log("err", "signature check failed", "customerID", packet.CustomerID)
			metrics.SignatureCheckFailed.Add(1)
			return
		}

		if !packet.Version.AtLeast(SDKVersion{4, 0, 0}) && !buyer.Debug {
			level.Error(logger).Log("err", "sdk too old", "version", packet.Version.String())
			metrics.SDKTooOld.Add(1)
			return
		}

		if _, err := getDatacenter(storer, packet.CustomerID, packet.DatacenterID, ""); err != nil {
			level.Error(logger).Log("handler", "server_update", "err", err)

			switch err.(type) {
			case ErrDatacenterNotFound:
				metrics.DatacenterNotFound.Add(1)

			case ErrDatacenterMapMisconfigured:
				metrics.MisconfiguredDatacenterAlias.Add(1)

			case ErrDatacenterNotAllowed:
				metrics.DatacenterNotAllowed.Add(1)
			}

			return
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

func SessionUpdateHandlerFunc(
	logger log.Logger,
	getIPLocator func(sessionID uint64) routing.IPLocator,
	getRouteMatrix func() *routing.RouteMatrix,
	multipathVetoHandler *storage.MultipathVetoHandler,
	storer storage.Storer,
	maxNearRelays int,
	routerPrivateKey [crypto.KeySize]byte,
	postSessionHandler *PostSessionHandler,
	metrics *metrics.SessionUpdateMetrics,
) UDPHandlerFunc {
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

		newSession := packet.SliceNumber == 0

		var sessionData SessionData
		var prevSessionData SessionData
		var routeDiversity int32

		ipLocator := getIPLocator(packet.SessionID)
		routeMatrix := getRouteMatrix()
		buyer := routing.Buyer{}
		datacenter := routing.UnknownDatacenter

		response := SessionResponsePacket{
			Version:     packet.Version,
			SessionID:   packet.SessionID,
			SliceNumber: packet.SliceNumber,
			RouteType:   routing.RouteTypeDirect,
		}

		var slicePacketLossClientToServer float32
		var slicePacketLossServerToClient float32
		var slicePacketLoss float32

		var debug *string

		// If we've gotten this far, use a deferred function so that we always at least return a direct response
		// and run the post session update logic
		defer func() {
			if response.RouteType != routing.RouteTypeDirect {
				metrics.NextSlices.Add(1)
				sessionData.EverOnNext = true
			} else {
				metrics.DirectSlices.Add(1)
			}

			packet.ClientAddress = AnonymizeAddr(packet.ClientAddress) // Make sure to always anonymize the client's IP address

			// Store the packets sent and lost in the session data to calculate the next slice's delta
			sessionData.PrevPacketsSentClientToServer = packet.PacketsSentClientToServer
			sessionData.PrevPacketsSentServerToClient = packet.PacketsSentServerToClient
			sessionData.PrevPacketsLostClientToServer = packet.PacketsLostClientToServer
			sessionData.PrevPacketsLostServerToClient = packet.PacketsLostServerToClient

			if err := writeSessionResponse(w, &response, &sessionData); err != nil {
				level.Error(logger).Log("msg", "failed to write session update response", "err", err)
				metrics.WriteResponseFailure.Add(1)
				return
			}

			// Rebuild the arrays of route relay names and sellers from the previous session data
			routeRelayNames := [core.MaxRelaysPerRoute]string{}
			routeRelaySellers := [core.MaxRelaysPerRoute]routing.Seller{}
			for i := int32(0); i < prevSessionData.RouteNumRelays; i++ {
				relay, err := storer.Relay(prevSessionData.RouteRelayIDs[i])
				if err != nil {
					continue
				}

				routeRelayNames[i] = relay.Name
				routeRelaySellers[i] = relay.Seller
			}

			// Rebuild the near relays from the previous session data
			var nearRelays nearRelayGroup

			// Make sure we only rebuild the previous near relays if we haven't gotten out of sync somehow
			if prevSessionData.RouteState.NumNearRelays == packet.NumNearRelays {
				nearRelays = newNearRelayGroup(prevSessionData.RouteState.NumNearRelays)
			}

			for i := int32(0); i < nearRelays.Count; i++ {

				// Since we now guarantee that the near relay IDs reported up from the SDK each slice don't change,
				// we can use the packet's near relay IDs here instead of storing the near relay IDs in the session data
				relayID := packet.NearRelayIDs[i]

				// Make sure to check if the relay exists in case the near relays are gone
				// this slice compared to the previous slice
				relayIndex, ok := routeMatrix.RelayIDsToIndices[relayID]
				if !ok {
					continue
				}

				nearRelays.IDs[i] = relayID
				nearRelays.Names[i] = routeMatrix.RelayNames[relayIndex]
				nearRelays.Addrs[i] = routeMatrix.RelayAddresses[relayIndex]
				nearRelays.RTTs[i] = prevSessionData.RouteState.NearRelayRTT[i]
				nearRelays.Jitters[i] = prevSessionData.RouteState.NearRelayJitter[i]

				// We don't actually store the packet loss in the session data, so just use the
				// values from the session update packet (no max history)
				if nearRelays.RTTs[i] >= 255 {
					nearRelays.PacketLosses[i] = 100
				} else {
					nearRelays.PacketLosses[i] = packet.NearRelayPacketLoss[i]
				}
			}

			if !packet.ClientPingTimedOut {
				go PostSessionUpdate(postSessionHandler, &packet, &prevSessionData, &buyer, multipathVetoHandler, routeRelayNames, routeRelaySellers, nearRelays, &datacenter, routeDiversity, slicePacketLossClientToServer, slicePacketLossServerToClient, debug)
			}
		}()

		if packet.ClientPingTimedOut {
			metrics.ClientPingTimedOut.Add(1)
			return
		}

		buyer, err := storer.Buyer(packet.CustomerID)
		if err != nil {
			level.Error(logger).Log("msg", "buyer not found", "err", err)
			metrics.BuyerNotFound.Add(1)
			return
		}

		if !crypto.VerifyPacket(buyer.PublicKey, incoming.Data) {
			level.Error(logger).Log("err", "signature check failed", "customerID", packet.CustomerID)
			metrics.SignatureCheckFailed.Add(1)
			return
		}

		if buyer.Debug {
			debug = new(string)
		}

		// If a player has the "pro" tag, set pro mode in the route shader
		if packet.Version.AtLeast(SDKVersion{4, 0, 3}) {
			for i := int32(0); i < packet.NumTags; i++ {
				if packet.Tags[i] == crypto.HashID("pro") {
					buyer.RouteShader.ProMode = true
					break
				}
			}
			// Case for older SDK versions where there was only 1 tag
		} else if len(packet.Tags) > 0 && packet.Tags[0] == crypto.HashID("pro") {
			buyer.RouteShader.ProMode = true
		}

		if datacenter, err = getDatacenter(storer, packet.CustomerID, packet.DatacenterID, ""); err != nil {
			level.Error(logger).Log("handler", "session_update", "err", err)

			switch err.(type) {
			case ErrDatacenterNotFound:
				metrics.DatacenterNotFound.Add(1)

			case ErrDatacenterMapMisconfigured:
				metrics.MisconfiguredDatacenterAlias.Add(1)

			case ErrDatacenterNotAllowed:
				metrics.DatacenterNotAllowed.Add(1)
			}

			return
		}

		if newSession {
			sessionData.Version = SessionDataVersion
			sessionData.SessionID = packet.SessionID
			sessionData.SliceNumber = packet.SliceNumber + 1
			sessionData.ExpireTimestamp = uint64(time.Now().Unix()) + billing.BillingSliceSeconds
			sessionData.RouteState.UserID = packet.UserHash
			sessionData.Location, err = ipLocator.LocateIP(packet.ClientAddress.IP)

			// Set the AB test field manually on the first slice only, so that
			// existing sessions don't start or stop running the AB test
			sessionData.RouteState.ABTest = buyer.RouteShader.ABTest

			// Save constant session data in the prev session data so that they
			// are displayed in the portal and billing correctly
			prevSessionData.Location = sessionData.Location
			prevSessionData.RouteState.ABTest = sessionData.RouteState.ABTest

			if err != nil {
				level.Error(logger).Log("msg", "failed to locate IP", "err", err)
				metrics.ClientLocateFailure.Add(1)
				return
			}

		} else {
			err := UnmarshalSessionData(&sessionData, packet.SessionData[:])
			prevSessionData = sessionData // Have an extra copy of the session data so we can use the unmodified one in the post session

			if err != nil {
				level.Error(logger).Log("msg", "could not read session data in session update packet", "err", err)
				metrics.ReadSessionDataFailure.Add(1)
				return
			}

			if sessionData.SessionID != packet.SessionID {
				level.Error(logger).Log("err", "bad session ID in session data")
				metrics.BadSessionID.Add(1)
				return
			}

			if sessionData.SliceNumber != packet.SliceNumber {
				level.Error(logger).Log("err", "bad slice number in session data", "packet_slice_number", packet.SliceNumber, "session_data_slice_number", sessionData.SliceNumber,
					"retry_count", packet.RetryNumber, "packet_next", packet.Next, "session_data_next", sessionData.RouteState.Next, "ever_on_next", sessionData.EverOnNext)
				metrics.BadSliceNumber.Add(1)
				return
			}

			sessionData.SliceNumber = packet.SliceNumber + 1
			sessionData.ExpireTimestamp += billing.BillingSliceSeconds

			slicePacketsSentClientToServer := packet.PacketsSentClientToServer - sessionData.PrevPacketsSentClientToServer
			slicePacketsSentServerToClient := packet.PacketsSentServerToClient - sessionData.PrevPacketsSentServerToClient

			slicePacketsLostClientToServer := packet.PacketsLostClientToServer - sessionData.PrevPacketsLostClientToServer
			slicePacketsLostServerToClient := packet.PacketsLostServerToClient - sessionData.PrevPacketsLostServerToClient

			if slicePacketsSentClientToServer == uint64(0) {
				slicePacketLossClientToServer = float32(0)
			} else {
				slicePacketLossClientToServer = float32(float64(slicePacketsLostClientToServer) / float64(slicePacketsSentClientToServer))
			}

			if slicePacketsSentServerToClient == uint64(0) {
				slicePacketLossServerToClient = float32(0)
			} else {
				slicePacketLossServerToClient = float32(float64(slicePacketsLostServerToClient) / float64(slicePacketsSentServerToClient))
			}

			slicePacketLoss = slicePacketLossClientToServer
			if slicePacketLossServerToClient > slicePacketLossClientToServer {
				slicePacketLoss = slicePacketLossServerToClient
			}
		}

		// Don't accelerate any sessions if the buyer is not yet live
		if !buyer.Live {
			metrics.BuyerNotLive.Add(1)
			return
		}

		if packet.FallbackToDirect {
			if !sessionData.FellBackToDirect {
				sessionData.FellBackToDirect = true

				switch packet.Flags {
				case FallbackFlagsBadRouteToken:
					metrics.FallbackToDirectBadRouteToken.Add(1)
				case FallbackFlagsNoNextRouteToContinue:
					metrics.FallbackToDirectNoNextRouteToContinue.Add(1)
				case FallbackFlagsPreviousUpdateStillPending:
					metrics.FallbackToDirectPreviousUpdateStillPending.Add(1)
				case FallbackFlagsBadContinueToken:
					metrics.FallbackToDirectBadContinueToken.Add(1)
				case FallbackFlagsRouteExpired:
					metrics.FallbackToDirectRouteExpired.Add(1)
				case FallbackFlagsRouteRequestTimedOut:
					metrics.FallbackToDirectRouteRequestTimedOut.Add(1)
				case FallbackFlagsContinueRequestTimedOut:
					metrics.FallbackToDirectContinueRequestTimedOut.Add(1)
				case FallbackFlagsClientTimedOut:
					metrics.FallbackToDirectClientTimedOut.Add(1)
				case FallbackFlagsUpgradeResponseTimedOut:
					metrics.FallbackToDirectUpgradeResponseTimedOut.Add(1)
				case FallbackFlagsRouteUpdateTimedOut:
					metrics.FallbackToDirectRouteUpdateTimedOut.Add(1)
				case FallbackFlagsDirectPongTimedOut:
					metrics.FallbackToDirectDirectPongTimedOut.Add(1)
				case FallbackFlagsNextPongTimedOut:
					metrics.FallbackToDirectNextPongTimedOut.Add(1)
				default:
					metrics.FallbackToDirectUnknownReason.Add(1)
				}
			}
			return
		}

		destRelayIDs := routeMatrix.GetDatacenterRelayIDs(datacenter.ID)
		if len(destRelayIDs) == 0 {
			level.Error(logger).Log("msg", "failed to get dest relays")
			metrics.NoRelaysInDatacenter.Add(1)
			return
		}

		incomingNearRelays := newNearRelayGroup(packet.NumNearRelays)
		for i := int32(0); i < incomingNearRelays.Count; i++ {
			incomingNearRelays.IDs[i] = packet.NearRelayIDs[i]

			incomingNearRelays.RTTs[i] = packet.NearRelayRTT[i]
			incomingNearRelays.Jitters[i] = packet.NearRelayJitter[i]
			incomingNearRelays.PacketLosses[i] = packet.NearRelayPacketLoss[i]

			// The SDK doesn't send up the relay name or relay address, so we have to get those from the route matrix
			relayIndex, ok := routeMatrix.RelayIDsToIndices[packet.NearRelayIDs[i]]
			if !ok {
				continue
			}

			incomingNearRelays.Addrs[i] = routeMatrix.RelayAddresses[relayIndex]
			incomingNearRelays.Names[i] = routeMatrix.RelayNames[relayIndex]
		}

		nearRelaysChanged, nearRelays, reframedDestRelays, err := handleNearAndDestRelays(
			int32(packet.SliceNumber),
			routeMatrix,
			incomingNearRelays,
			&buyer.RouteShader,
			&sessionData.RouteState,
			newSession,
			sessionData.Location.Latitude,
			sessionData.Location.Longitude,
			datacenter.Location.Latitude,
			datacenter.Location.Longitude,
			maxNearRelays,
			int32(math.Ceil(float64(packet.DirectRTT))),
			int32(math.Ceil(float64(packet.DirectJitter))),
			int32(slicePacketLoss),
			int32(math.Floor(float64(packet.NextPacketLoss)+0.5)),
			sessionData.RouteRelayIDs[0],
			destRelayIDs,
			debug,
		)

		response.NumNearRelays = nearRelays.Count
		response.NearRelayIDs = nearRelays.IDs
		response.NearRelayAddresses = nearRelays.Addrs
		response.NearRelaysChanged = nearRelaysChanged
		response.HighFrequencyPings = buyer.InternalConfig.HighFrequencyPings

		if err != nil {
			if strings.HasPrefix(err.Error(), "near relays changed") {
				level.Error(logger).Log("msg", "near relays changed", "err", err)
				metrics.NearRelaysChanged.Add(1)
			} else {
				level.Error(logger).Log("msg", "failed to get near relays", "err", err)
				metrics.NearRelaysLocateFailure.Add(1)
			}

			return
		}

		// First slice always direct
		if newSession {
			level.Debug(logger).Log("msg", "session updated successfully", "source_address", incoming.SourceAddr.String(), "server_address", packet.ServerAddress.String(), "client_address", packet.ClientAddress.String())
			return
		}

		var routeCost int32
		routeRelays := [core.MaxRelaysPerRoute]int32{}

		sessionData.Initial = false

		multipathVetoMap := multipathVetoHandler.GetMapCopy(buyer.CompanyCode)

		level.Debug(logger).Log("buyer", buyer.CompanyCode,
			"acceptable_latency", buyer.RouteShader.AcceptableLatency,
			"rtt_threshold", buyer.RouteShader.LatencyThreshold,
			"selection_percent", buyer.RouteShader.SelectionPercent,
			"route_switch_threshold", buyer.InternalConfig.RouteSwitchThreshold)

		nearRelayIndices := make([]int32, nearRelays.Count)
		nearRelayCosts := make([]int32, nearRelays.Count)
		for i := int32(0); i < nearRelays.Count; i++ {
			nearRelayIndex, ok := routeMatrix.RelayIDsToIndices[nearRelays.IDs[i]]
			if !ok {
				continue
			}

			nearRelayIndices[i] = nearRelayIndex
			nearRelayCosts[i] = nearRelays.RTTs[i]
		}

		var routeNumRelays int32

		var nextRouteSwitched bool

		if !sessionData.RouteState.Next || sessionData.RouteNumRelays == 0 {
			sessionData.RouteState.Next = false
			if core.MakeRouteDecision_TakeNetworkNext(routeMatrix.RouteEntries, &buyer.RouteShader, &sessionData.RouteState, multipathVetoMap, &buyer.InternalConfig, int32(packet.DirectRTT), slicePacketLoss, nearRelayIndices, nearRelayCosts, reframedDestRelays, &routeCost, &routeNumRelays, routeRelays[:], &routeDiversity, debug) {
				HandleNextToken(&sessionData, storer, &buyer, &packet, routeNumRelays, routeRelays[:], routeMatrix.RelayIDs, routerPrivateKey, &response)
			}
		} else {
			if !core.ReframeRoute(&sessionData.RouteState, routeMatrix.RelayIDsToIndices, sessionData.RouteRelayIDs[:sessionData.RouteNumRelays], &routeRelays) {

				routeRelays = [core.MaxRelaysPerRoute]int32{}

				level.Warn(logger).Log("warn", "one or more relays in the route no longer exist. Clearing route.")
				metrics.RouteDoesNotExist.Add(1)
			}

			// The SDK sent up "next = false" but didn't fall back to direct - the SDK "aborted" this session
			if !packet.Next {
				sessionData.RouteState.Next = false
				sessionData.RouteState.Veto = true

				level.Warn(logger).Log("warn", "SDK aborted session")
				metrics.SDKAborted.Add(1)
			} else {
				var stay bool
				if stay, nextRouteSwitched = core.MakeRouteDecision_StayOnNetworkNext(routeMatrix.RouteEntries, routeMatrix.RelayNames, &buyer.RouteShader, &sessionData.RouteState, &buyer.InternalConfig, int32(packet.DirectRTT), int32(packet.NextRTT), sessionData.RouteCost, slicePacketLoss, packet.NextPacketLoss, sessionData.RouteNumRelays, routeRelays, nearRelayIndices, nearRelayCosts, reframedDestRelays, &routeCost, &routeNumRelays, routeRelays[:], debug); stay {

					// Continue token

					// Check if the route has changed
					if nextRouteSwitched {
						metrics.RouteSwitched.Add(1)

						// Create a next token here rather than a continue token since the route has switched
						HandleNextToken(&sessionData, storer, &buyer, &packet, routeNumRelays, routeRelays[:], routeMatrix.RelayIDs, routerPrivateKey, &response)
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

					if sessionData.RouteState.Mispredict {
						level.Warn(logger).Log("warn", "we mispredicted too many times")
						metrics.MispredictVeto.Add(1)
					}

					if sessionData.RouteState.LatencyWorse {
						level.Warn(logger).Log("warn", "this route makes latency worse")
						metrics.LatencyWorse.Add(1)
					}
				}
			}
		}

		if routeCost > routing.InvalidRouteValue {
			routeCost = routing.InvalidRouteValue
		}

		response.Committed = sessionData.RouteState.Committed
		response.Multipath = sessionData.RouteState.Multipath

		// Store the route back into the session data
		sessionData.RouteNumRelays = routeNumRelays
		sessionData.RouteCost = routeCost
		sessionData.RouteChanged = nextRouteSwitched

		for i := int32(0); i < routeNumRelays; i++ {
			relayID := routeMatrix.RelayIDs[routeRelays[i]]
			sessionData.RouteRelayIDs[i] = relayID
		}

		if debug != nil {
			response.Debug = *debug
			if response.Debug != "" {
				response.HasDebug = true
			}
		}

		level.Debug(logger).Log("msg", "session updated successfully", "source_address", incoming.SourceAddr.String(), "server_address", packet.ServerAddress.String(), "client_address", packet.ClientAddress.String())
	}
}

func HandleNextToken(
	sessionData *SessionData,
	storer storage.Storer,
	buyer *routing.Buyer,
	packet *SessionUpdatePacket,
	routeNumRelays int32,
	routeRelays []int32,
	allRelayIDs []uint64,
	routerPrivateKey [crypto.KeySize]byte,
	response *SessionResponsePacket,
) {
	// Add another 10 seconds to the slice and increment the session version
	sessionData.Initial = true
	sessionData.ExpireTimestamp += billing.BillingSliceSeconds
	sessionData.SessionVersion++

	numTokens := routeNumRelays + 2 // relays + client + server
	routeAddresses, routePublicKeys := GetRouteAddressesAndPublicKeys(&packet.ClientAddress, packet.ClientRoutePublicKey, &packet.ServerAddress, packet.ServerRoutePublicKey, numTokens, routeRelays, allRelayIDs, storer)
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

func HandleContinueToken(
	sessionData *SessionData,
	storer storage.Storer,
	buyer *routing.Buyer,
	packet *SessionUpdatePacket,
	routeNumRelays int32,
	routeRelays []int32,
	allRelayIDs []uint64,
	routerPrivateKey [crypto.KeySize]byte,
	response *SessionResponsePacket,
) {
	numTokens := routeNumRelays + 2 // relays + client + server
	// empty string array b/c don't care for internal ips here
	routeAddresses, routePublicKeys := GetRouteAddressesAndPublicKeys(&packet.ClientAddress, packet.ClientRoutePublicKey, &packet.ServerAddress, packet.ServerRoutePublicKey, numTokens, routeRelays, allRelayIDs, storer)
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

func GetRouteAddressesAndPublicKeys(
	clientAddress *net.UDPAddr,
	clientPublicKey []byte,
	serverAddress *net.UDPAddr,
	serverPublicKey []byte,
	numTokens int32,
	routeRelays []int32,
	allRelayIDs []uint64,
	storer storage.Storer,
) ([]*net.UDPAddr, [][]byte) {
	routeAddresses := make([]*net.UDPAddr, numTokens)
	routePublicKeys := make([][]byte, numTokens)

	routeAddresses[0] = clientAddress
	routePublicKeys[0] = clientPublicKey
	routeAddresses[numTokens-1] = serverAddress
	routePublicKeys[numTokens-1] = serverPublicKey

	totalNumRelays := int32(len(allRelayIDs))
	foundRelayCount := int32(0)

	enableInternalIPs, _ := envvar.GetBool("FEATURE_ENABLE_INTERNAL_IPS", false)

	for i := int32(0); i < numTokens-2; i++ {
		relayIndex := routeRelays[i]
		if relayIndex < totalNumRelays {
			relayID := allRelayIDs[relayIndex]
			relay, err := storer.Relay(relayID)
			if err != nil {
				continue
			}

			routeAddresses = AddAddress(enableInternalIPs, i, relay, allRelayIDs, storer, routeRelays, routeAddresses)

			routePublicKeys[i+1] = relay.PublicKey
			foundRelayCount++
		}
	}

	if foundRelayCount != numTokens-2 {
		return nil, nil
	}

	return routeAddresses, routePublicKeys
}

func AddAddress(enableInternalIPs bool, index int32, relay routing.Relay, allRelayIDs []uint64, storer storage.Storer, routeRelays []int32, routeAddresses []*net.UDPAddr) []*net.UDPAddr {
	totalNumRelays := int32(len(allRelayIDs))
	routeAddresses[index+1] = &relay.Addr
	if enableInternalIPs {
		// check if the previous relay is the same seller
		if index >= 1 {
			prevRelayIndex := routeRelays[index-1]
			if prevRelayIndex < totalNumRelays {
				prevID := allRelayIDs[prevRelayIndex]
				prev, err := storer.Relay(prevID)
				if err == nil && prev.Seller.ID == relay.Seller.ID && prev.InternalAddr.String() != ":0" && relay.InternalAddr.String() != ":0" {
					routeAddresses[index+1] = &relay.InternalAddr
				}
			}
		}
	}

	return routeAddresses
}

func PostSessionUpdate(
	postSessionHandler *PostSessionHandler,
	packet *SessionUpdatePacket,
	sessionData *SessionData,
	buyer *routing.Buyer,
	multipathVetoHandler *storage.MultipathVetoHandler,
	routeRelayNames [core.MaxRelaysPerRoute]string,
	routeRelaySellers [core.MaxRelaysPerRoute]routing.Seller,
	nearRelays nearRelayGroup,
	datacenter *routing.Datacenter,
	routeDiversity int32,
	slicePacketLossClientToServer float32,
	slicePacketLossServerToClient float32,
	debug *string,
) {
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

	nextRelaysPrice := [core.MaxRelaysPerRoute]uint64{}
	for i := 0; i < core.MaxRelaysPerRoute; i++ {
		nextRelaysPrice[i] = uint64(routeRelayPrices[i])
	}

	var routeCost int32 = sessionData.RouteCost
	if sessionData.RouteCost == math.MaxInt32 {
		routeCost = 0
	}

	var nearRelayRTT float32
	if sessionData.RouteNumRelays > 0 {
		for i, nearRelayID := range nearRelays.IDs {
			if nearRelayID == sessionData.RouteRelayIDs[0] {
				nearRelayRTT = float32(nearRelays.RTTs[i])
				break
			}
		}
	}

	debugString := ""
	if debug != nil {
		debugString = *debug
	}

	var numNearRelays uint8
	nearRelayIDs := [billing.BillingEntryMaxNearRelays]uint64{}
	nearRelayRTTs := [billing.BillingEntryMaxNearRelays]float32{}
	nearRelayJitters := [billing.BillingEntryMaxNearRelays]float32{}
	nearRelayPacketLosses := [billing.BillingEntryMaxNearRelays]float32{}

	if buyer.Debug {
		numNearRelays = uint8(nearRelays.Count)
		for i := uint8(0); i < numNearRelays; i++ {
			nearRelayIDs[i] = nearRelays.IDs[i]
			nearRelayRTTs[i] = float32(nearRelays.RTTs[i])
			nearRelayJitters[i] = float32(nearRelays.Jitters[i])
			nearRelayPacketLosses[i] = float32(nearRelays.PacketLosses[i])
		}
	}

	slicePacketLoss := slicePacketLossClientToServer
	if slicePacketLossServerToClient > slicePacketLossClientToServer {
		slicePacketLoss = slicePacketLossServerToClient
	}

	billingEntry := &billing.BillingEntry{
		Timestamp:                       uint64(time.Now().Unix()),
		BuyerID:                         packet.CustomerID,
		UserHash:                        packet.UserHash,
		SessionID:                       packet.SessionID,
		SliceNumber:                     packet.SliceNumber,
		DirectRTT:                       packet.DirectRTT,
		DirectJitter:                    packet.DirectJitter,
		DirectPacketLoss:                packet.DirectPacketLoss,
		Next:                            packet.Next,
		NextRTT:                         packet.NextRTT,
		NextJitter:                      packet.NextJitter,
		NextPacketLoss:                  packet.NextPacketLoss,
		NumNextRelays:                   uint8(sessionData.RouteNumRelays),
		NextRelays:                      sessionData.RouteRelayIDs,
		TotalPrice:                      uint64(totalPrice),
		ClientToServerPacketsLost:       packet.PacketsLostClientToServer,
		ServerToClientPacketsLost:       packet.PacketsLostServerToClient,
		Committed:                       packet.Committed,
		Flagged:                         packet.Reported,
		Multipath:                       sessionData.RouteState.Multipath,
		Initial:                         sessionData.Initial,
		NextBytesUp:                     nextBytesUp,
		NextBytesDown:                   nextBytesDown,
		EnvelopeBytesUp:                 nextEnvelopeBytesUp,
		EnvelopeBytesDown:               nextEnvelopeBytesDown,
		DatacenterID:                    datacenter.ID,
		RTTReduction:                    sessionData.RouteState.ReduceLatency,
		PacketLossReduction:             sessionData.RouteState.ReducePacketLoss,
		NextRelaysPrice:                 nextRelaysPrice,
		Latitude:                        float32(sessionData.Location.Latitude),
		Longitude:                       float32(sessionData.Location.Longitude),
		ISP:                             sessionData.Location.ISP,
		ABTest:                          sessionData.RouteState.ABTest,
		RouteDecision:                   0,
		ConnectionType:                  uint8(packet.ConnectionType),
		PlatformType:                    uint8(packet.PlatformType),
		SDKVersion:                      packet.Version.String(),
		PacketLoss:                      slicePacketLoss,
		PredictedNextRTT:                float32(routeCost),
		MultipathVetoed:                 sessionData.RouteState.MultipathOverload,
		UseDebug:                        buyer.Debug,
		Debug:                           debugString,
		FallbackToDirect:                packet.FallbackToDirect,
		ClientFlags:                     packet.Flags,
		UserFlags:                       packet.UserFlags,
		NearRelayRTT:                    nearRelayRTT,
		PacketsOutOfOrderClientToServer: packet.PacketsOutOfOrderClientToServer,
		PacketsOutOfOrderServerToClient: packet.PacketsOutOfOrderServerToClient,
		JitterClientToServer:            packet.JitterClientToServer,
		JitterServerToClient:            packet.JitterServerToClient,
		NumNearRelays:                   numNearRelays,
		NearRelayIDs:                    nearRelayIDs,
		NearRelayRTTs:                   nearRelayRTTs,
		NearRelayJitters:                nearRelayJitters,
		NearRelayPacketLosses:           nearRelayPacketLosses,
		RelayWentAway:                   sessionData.RouteState.RelayWentAway,
		RouteLost:                       sessionData.RouteState.RouteLost,
		NumTags:                         uint8(packet.NumTags),
		Tags:                            packet.Tags,
		Mispredicted:                    sessionData.RouteState.Mispredict,
		Vetoed:                          sessionData.RouteState.Veto,
		LatencyWorse:                    sessionData.RouteState.LatencyWorse,
		NoRoute:                         sessionData.RouteState.NoRoute,
		NextLatencyTooHigh:              sessionData.RouteState.NextLatencyTooHigh,
		RouteChanged:                    sessionData.RouteChanged,
		CommitVeto:                      sessionData.RouteState.CommitVeto,
		RouteDiversity:                  uint32(routeDiversity),
		LackOfDiversity:                 sessionData.RouteState.LackOfDiversity,
		Pro:                             buyer.RouteShader.ProMode && !sessionData.RouteState.MultipathRestricted,
		MultipathRestricted:             sessionData.RouteState.MultipathRestricted,
		ClientToServerPacketsSent:       packet.PacketsSentClientToServer,
		ServerToClientPacketsSent:       packet.PacketsSentServerToClient,
	}

	postSessionHandler.SendBillingEntry(billingEntry)

	if postSessionHandler.useVanityMetrics {
		postSessionHandler.SendVanityMetric(billingEntry)
	}

	hops := make([]RelayHop, sessionData.RouteNumRelays)
	for i := int32(0); i < sessionData.RouteNumRelays; i++ {
		hops[i] = RelayHop{
			ID:   sessionData.RouteRelayIDs[i],
			Name: routeRelayNames[i],
		}
	}

	nearRelayPortalData := make([]NearRelayPortalData, nearRelays.Count)
	for i := range nearRelayPortalData {
		nearRelayPortalData[i] = NearRelayPortalData{
			ID:   nearRelays.IDs[i],
			Name: nearRelays.Names[i],
			ClientStats: routing.Stats{
				RTT:        float64(nearRelays.RTTs[i]),
				Jitter:     float64(nearRelays.Jitters[i]),
				PacketLoss: float64(nearRelays.PacketLosses[i]),
			},
		}
	}

	// Sort the near relays for display purposes
	sort.Slice(nearRelayPortalData, func(i, j int) bool {
		return nearRelayPortalData[i].Name < nearRelayPortalData[j].Name
	})

	var deltaRTT float32
	if packet.Next && packet.NextRTT != 0 && packet.DirectRTT >= packet.NextRTT {
		deltaRTT = packet.DirectRTT - packet.NextRTT
	}

	predictedRTT := float64(sessionData.RouteCost)
	if sessionData.RouteCost >= routing.InvalidRouteValue {
		predictedRTT = 0
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
			Predicted: routing.Stats{
				RTT: predictedRTT,
			},
			ClientToServerStats: routing.Stats{
				Jitter:     float64(packet.JitterClientToServer),
				PacketLoss: float64(slicePacketLossClientToServer),
			},
			ServerToClientStats: routing.Stats{
				Jitter:     float64(packet.JitterServerToClient),
				PacketLoss: float64(slicePacketLossServerToClient),
			},
			RouteDiversity: uint32(routeDiversity),
			Envelope: routing.Envelope{
				Up:   int64(packet.NextKbpsUp),
				Down: int64(packet.NextKbpsDown),
			},
			IsMultiPath:       sessionData.RouteState.Multipath,
			IsTryBeforeYouBuy: !sessionData.RouteState.Committed,
			OnNetworkNext:     packet.Next,
		},
		Point: SessionMapPoint{
			Latitude:  float64(sessionData.Location.Latitude),
			Longitude: float64(sessionData.Location.Longitude),
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

func CalculateTotalPriceNibblins(routeNumRelays int, relaySellers [core.MaxRelaysPerRoute]routing.Seller, envelopeBytesUp uint64, envelopeBytesDown uint64) routing.Nibblin {

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

func CalculateRouteRelaysPrice(routeNumRelays int, relaySellers [core.MaxRelaysPerRoute]routing.Seller, envelopeBytesUp uint64, envelopeBytesDown uint64) [core.MaxRelaysPerRoute]routing.Nibblin {
	relayPrices := [core.MaxRelaysPerRoute]routing.Nibblin{}

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
