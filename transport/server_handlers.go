package transport

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/go-redis/redis/v7"
	"github.com/networknext/backend/billing"
	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/metrics"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"

	fnv "hash/fnv"

	jsoniter "github.com/json-iterator/go"
)

type UDPPacket struct {
	SourceAddr *net.UDPAddr
	Data       []byte
}

// UDPHandlerFunc acts the same way http.HandlerFunc does, but for UDP packets and address
type UDPHandlerFunc func(io.Writer, *UDPPacket)

// ServerIngress is a simple UDP router for specific packets and runs each UDPHandlerFunc based on the incoming packet type
type UDPServerMux struct {
	Conn          *net.UDPConn
	MaxPacketSize int

	ServerInitHandlerFunc    UDPHandlerFunc
	ServerUpdateHandlerFunc  UDPHandlerFunc
	SessionUpdateHandlerFunc UDPHandlerFunc
}

// Start begins accepting UDP packets from the UDP connection and will block
func (m *UDPServerMux) Start(ctx context.Context) error {
	if m.Conn == nil {
		return errors.New("udp connection cannot be nil")
	}

	for i := 0; i < runtime.NumCPU(); i++ {
		go m.handler(ctx, i)
	}

	<-ctx.Done()

	return nil
}

func (m *UDPServerMux) handler(ctx context.Context, id int) {

	for {

		data := make([]byte, m.MaxPacketSize)

		size, addr, _ := m.Conn.ReadFromUDP(data)
		if size <= 0 {
			continue
		}

		data = data[:size]

		go func(packet_data []byte, packet_size int, from *net.UDPAddr) {

			// Check the packet hash is legit and remove the hash from the beginning of the packet
			// to continue processing the packet as normal
			hashedPacket := crypto.Check(crypto.PacketHashKey, packet_data)
			switch hashedPacket {
			case true:
				packet_data = packet_data[crypto.PacketHashSize:packet_size]
			default:
				// todo: once everybody has upgraded to SDK 3.4.5 or greater, this is an error. ignore packet.
				packet_data = packet_data[:packet_size]
			}

			packet := UDPPacket{SourceAddr: from, Data: packet_data}

			var buf bytes.Buffer

			switch packet.Data[0] {
			case PacketTypeServerInitRequest:
				m.ServerInitHandlerFunc(&buf, &packet)
			case PacketTypeServerUpdate:
				m.ServerUpdateHandlerFunc(&buf, &packet)
			case PacketTypeSessionUpdate:
				m.SessionUpdateHandlerFunc(&buf, &packet)
			}

			if buf.Len() > 0 {
				res := buf.Bytes()

				// If the hash checks out above then hash the response to the sender
				if hashedPacket {
					res = crypto.Hash(crypto.PacketHashKey, res)
				}

				m.Conn.WriteToUDP(res, packet.SourceAddr)
			}

		}(data, size, addr)
	}
}

// ==========================================================================================

// todo: I would prefer a single counters struct, with more descriptive names: eg: "NumServerInitPackets", "LongServerInits".
// having generic names like you do below this makes it difficult to search throughout the code and find all instances where "Packets" and "LongDuration"
// are used, because it picks up the counters for server update and session update as well. please prefer fully descriptive, not generic names within structs.
// in other words, avoid using structs as namespace.

type ServerInitCounters struct {
	Packets      uint64
	LongDuration uint64
}

type ServerInitParams struct {
	ServerPrivateKey []byte
	Storer           storage.Storer
	Metrics          *metrics.ServerInitMetrics
	Logger           log.Logger
	Counters         *ServerInitCounters
}

func writeServerInitResponse(params *ServerInitParams, w io.Writer, packet *ServerInitRequestPacket, response uint32) {
	responsePacket := ServerInitResponsePacket{
		RequestID: packet.RequestID,
		Response:  response,
		Version:   packet.Version,
	}
	if err := writeInitResponse(w, responsePacket, params.ServerPrivateKey); err != nil {
		params.Metrics.ErrorMetrics.WriteResponseFailure.Add(1)
		return
	}
}

func ServerInitHandlerFunc(params *ServerInitParams) UDPHandlerFunc {

	return func(w io.Writer, incoming *UDPPacket) {

		// Server init is called when the server first starts up.
		// Its purpose is to give feedback to people integrating our SDK into their game server when something is not setup correctly.
		// For example, if they have not setup the datacenter name, or the datacenter name does not exist, it will tell them that.

		// IMPORTANT: Server init is a new concept that only exists in SDK 3.4.5 and greater.

		// Psyonix is currently on an older SDK version, so server inits don't show up for them.

		start := time.Now()
		defer func() {
			if time.Since(start).Seconds() > 0.1 {
				level.Debug(params.Logger).Log("msg", "long server init")
				atomic.AddUint64(&params.Counters.LongDuration, 1)
				params.Metrics.LongDuration.Add(1)
			}
		}()

		params.Metrics.Invocations.Add(1)

		atomic.AddUint64(&params.Counters.Packets, 1)

		// Read the server init packet. We can do this all at once because the server init packet includes the SDK version.

		var packet ServerInitRequestPacket
		if err := packet.UnmarshalBinary(incoming.Data); err != nil {
			// fmt.Printf("could not read server init packet\n")
			params.Metrics.ErrorMetrics.ReadPacketFailure.Add(1)
			return
		}

		// todo: ryan. in the old code we checked if this buyer had the "internal" flag set, and then only in that case we
		// allowed 0.0.0 version. this is a MUCH better approach than checking source ip address for loopback. please fix.
		if !incoming.SourceAddr.IP.IsLoopback() && !packet.Version.AtLeast(SDKVersionMin) {
			// fmt.Printf("sdk too old: %s\n", packet.Version.String())
			params.Metrics.ErrorMetrics.SDKTooOld.Add(1)
			writeServerInitResponse(params, w, &packet, InitResponseOldSDKVersion)
			return
		}

		// We need to look up the buyer from the customer id included in the packet.
		// If the buyer does not exist, then the user has probably not setup their customer private/public keypair correctly.

		buyer, err := params.Storer.Buyer(packet.CustomerID)
		if err != nil {
			// fmt.Printf("unknown customer: %x\n", packet.CustomerID)
			params.Metrics.ErrorMetrics.BuyerNotFound.Add(1)
			writeServerInitResponse(params, w, &packet, InitResponseUnknownCustomer)
			return
		}

		// Now that we have the buyer, we know the public key that corresponds to this customer's private key.
		// Only the customer knows their private key, but we can use their public key to cryptographically check
		// that this server init packet was signed by somebody with the private key. This is how we ensure that
		// only real customer servers are allowed on our system.

		if !crypto.Verify(buyer.PublicKey, packet.GetSignData(), packet.Signature) {
			// fmt.Printf("signature check failed\n")
			params.Metrics.ErrorMetrics.VerificationFailure.Add(1)
			writeServerInitResponse(params, w, &packet, InitResponseSignatureCheckFailed)
			return
		}

		// If neither the datacenter nor a relevent alias exists, the user has probably not set the
		// datacenter string correctly on their server instance, or the datacenter name they are
		// passing in does not exist (yet).

		// IMPORTANT: In the future, we will extend the SDK to pass in the datacenter name as a string
		// because it's really difficult to debug what the incorrectly datacenter string is, when we only
		// see the hash :(

		datacenter, _ := params.Storer.Datacenter(packet.DatacenterID)
		if datacenter == routing.UnknownDatacenter {
			// search the list of aliases created by/for this buyer
			datacenterAliases := params.Storer.GetDatacenterMapsForBuyer(fmt.Sprintf("%x", packet.CustomerID))
			if len(datacenterAliases) == 0 {
				params.Metrics.ErrorMetrics.DatacenterNotFound.Add(1)
				writeServerInitResponse(params, w, &packet, InitResponseUnknownDatacenter)
			}
			for _, dcMap := range datacenterAliases {
				if packet.DatacenterID == crypto.HashID(dcMap.Alias) {
					datacenter, err = params.Storer.Datacenter(packet.DatacenterID)
					if err != nil {
						params.Metrics.ErrorMetrics.DatacenterNotFound.Add(1)
						writeServerInitResponse(params, w, &packet, InitResponseUnknownDatacenter)
					}
				}
			}
		}

		// If we get down here, all checks have passed and this server is OK to init.
		// Once a server inits, it goes into a mode where it can potentially monitor and accelerate sessions.
		// After 10 seconds, if the server fails to init, it will fall back to direct and never monitor or accelerate sessions until it is restarted.

		// IMPORTANT: In a future SDK version, it is probably important that we extend the server code to retry initialization,
		// since right now it only re-initializes if that server is restarted, and we can't rely on all our customers to regularly
		// restart their servers (although Psyonix does do this).

		writeServerInitResponse(params, w, &packet, InitResponseOK)
	}
}

// ==========================================================================================

type ServerCacheEntry struct {
	Sequence   uint64
	Server     routing.Server
	Datacenter routing.Datacenter
	SDKVersion SDKVersion
}

func (e *ServerCacheEntry) UnmarshalBinary(data []byte) error {
	return jsoniter.Unmarshal(data, e)
}

func (e ServerCacheEntry) MarshalBinary() ([]byte, error) {
	return jsoniter.Marshal(e)
}

type ServerUpdateCounters struct {
	Packets      uint64
	LongDuration uint64
}

type ServerUpdateParams struct {
	Storer    storage.Storer
	Metrics   *metrics.ServerUpdateMetrics
	Logger    log.Logger
	ServerMap *ServerMap
	Counters  *ServerUpdateCounters
}

// =============================================================================

func ServerUpdateHandlerFunc(params *ServerUpdateParams) UDPHandlerFunc {

	return func(w io.Writer, incoming *UDPPacket) {

		start := time.Now()
		defer func() {
			if time.Since(start).Seconds() > 0.1 {
				// level.Debug(params.Logger).Log("msg", "long server update")
				atomic.AddUint64(&params.Counters.LongDuration, 1)
				params.Metrics.LongDuration.Add(1)
			}
		}()

		params.Metrics.Invocations.Add(1)

		atomic.AddUint64(&params.Counters.Packets, 1)

		// Read the entire server update packet. We can do this all at once because the packet contains the SDK version in it.

		var packet ServerUpdatePacket
		if err := packet.UnmarshalBinary(incoming.Data); err != nil {
			// level.Error(params.Logger).Log("msg", "could not read server update packet", "err", err)
			params.Metrics.ErrorMetrics.UnserviceableUpdate.Add(1)
			params.Metrics.ErrorMetrics.ReadPacketFailure.Add(1)
			return
		}

		// todo: in the old code we checked if we were running on a buyer account with "internal" set, and allowed 0.0.0 there only.
		// this is much better than checking the loopback address here. please fix ryan

		// Check if the sdk version is too old
		if !incoming.SourceAddr.IP.IsLoopback() && !packet.Version.AtLeast(SDKVersionMin) {
			// level.Error(params.Logger).Log("msg", "ignoring old sdk version", "version", packet.Version.String())
			params.Metrics.ErrorMetrics.UnserviceableUpdate.Add(1)
			params.Metrics.ErrorMetrics.SDKTooOld.Add(1)
			return
		}

		// Get the buyer information for the customer id in the packet.
		// If the buyer does not exist, this is not a server we care about. Don't even waste bandwidth to respond.

		buyer, err := params.Storer.Buyer(packet.CustomerID)
		if err != nil {
			// level.Error(locallogger).Log("msg", "failed to get buyer from storage", "err", err, "customer_id", packet.CustomerID)
			params.Metrics.ErrorMetrics.UnserviceableUpdate.Add(1)
			params.Metrics.ErrorMetrics.BuyerNotFound.Add(1)
			return
		}

		// Check the server update is signed by the private key of the buyer.
		// If the signature does not match, this is not a server we care about. Don't even waste bandwidth to respond.

		if !crypto.Verify(buyer.PublicKey, packet.GetSignData(), packet.Signature) {
			// level.Error(locallogger).Log("msg", "signature verification failed")
			params.Metrics.ErrorMetrics.UnserviceableUpdate.Add(1)
			params.Metrics.ErrorMetrics.VerificationFailure.Add(1)
			return
		}

		// Look up the datacenter by id and make sure it exists.
		// Sometimes the customer has datacenter aliases, eg: "multiplay.newyork" -> "inap.newyork".
		// To support this, when we can't find a datacenter directly by id, we look it up by alias instead.

		datacenter, err := params.Storer.Datacenter(packet.DatacenterID) // todo: ryan, profiling indicates this is slow. please investigate
		if datacenter == routing.UnknownDatacenter {
			// search the list of aliases created by/for this buyer
			datacenterAliases := params.Storer.GetDatacenterMapsForBuyer(fmt.Sprintf("%x", packet.CustomerID))
			for _, dcMap := range datacenterAliases {
				if packet.DatacenterID == crypto.HashID(dcMap.Alias) {
					datacenter, err = params.Storer.Datacenter(packet.DatacenterID)
					if err != nil {
						params.Metrics.ErrorMetrics.UnserviceableUpdate.Add(1)
						params.Metrics.ErrorMetrics.VerificationFailure.Add(1)
						return

					}
				}

			}
		}

		// UDP packets may arrive out of order. So that we don't have stale server update packets arriving late and
		// ruining our server map with stale information, we must check the server update sequence number, and discard
		// any server updates that are the same sequence number or older than the current server entry in the server map.
		var sequence uint64

		serverAddress := packet.ServerAddress.String()

		// IMPORTANT: The server data *must* be treated as read only or it is not threadsafe!
		serverDataReadOnly := params.ServerMap.GetServerData(serverAddress)
		if serverDataReadOnly != nil {
			sequence = serverDataReadOnly.sequence
		}

		// Drop the packet if the sequence number is older than the previously cached sequence number
		if packet.Sequence < sequence {
			// level.Error(params.Logger).Log("handler", "server", "msg", "packet too old", "packet sequence", packet.Sequence, "lastest sequence", serverDataReadOnly.sequence)
			params.Metrics.ErrorMetrics.UnserviceableUpdate.Add(1)
			params.Metrics.ErrorMetrics.PacketSequenceTooOld.Add(1)
			return
		}

		// Each one of our customer's servers reports to us with this server update packet every 10 seconds.
		// Therefore we must update the server data each time we receive an update, to keep this server entry live in our server map.
		// When we don't receive an update for a server for a certain period of time (for example 30 seconds), that server entry times out.

		server := ServerData{
			timestamp:      time.Now().Unix(),
			routePublicKey: packet.ServerRoutePublicKey,
			version:        packet.Version,
			datacenter:     datacenter,
			sequence:       packet.Sequence,
		}

		serverMutexStart := time.Now()
		params.ServerMap.UpdateServerData(serverAddress, &server)
		if time.Since(serverMutexStart).Seconds() > 0.1 {
			level.Debug(params.Logger).Log("msg", "long server mutex in server update")
		}
	}
}

// =============================================================================

type SessionCacheEntry struct {
	CustomerID                 uint64
	SessionID                  uint64
	UserHash                   uint64
	Sequence                   uint64
	RouteHash                  uint64
	RouteDecision              routing.Decision
	OnNNSliceCounter           uint64
	CommitPending              bool
	CommitObservedSliceCounter uint8
	Committed                  bool
	TimestampStart             time.Time
	TimestampExpire            time.Time
	Version                    uint8
	DirectRTT                  float64
	NextRTT                    float64
	Location                   routing.Location
	Response                   []byte
}

func (e *SessionCacheEntry) UnmarshalBinary(data []byte) error {
	return jsoniter.Unmarshal(data, e)
}

func (e SessionCacheEntry) MarshalBinary() ([]byte, error) {
	return jsoniter.Marshal(e)
}

type VetoCacheEntry struct {
	VetoTimestamp time.Time
	Reason        routing.DecisionReason
}

func (e *VetoCacheEntry) UnmarshalBinary(data []byte) error {
	return jsoniter.Unmarshal(data, e)
}

func (e VetoCacheEntry) MarshalBinary() ([]byte, error) {
	return jsoniter.Marshal(e)
}

type RouteProvider interface {
	ResolveRelay(id uint64) (routing.Relay, error)
	GetDatacenterRelays(datacenter routing.Datacenter) []routing.Relay
	GetRoutes(near []routing.Relay, dest []routing.Relay) ([]routing.Route, error)
	GetNearRelays(latitude float64, longitude float64, maxNearRelays int) ([]routing.Relay, error)
}

type SessionUpdateCounters struct {
	Packets      uint64
	LongDuration uint64
}

type SessionUpdateParams struct {
	ServerPrivateKey     []byte
	RouterPrivateKey     []byte
	GetRouteProvider     func() RouteProvider
	IPLoc                routing.IPLocator
	Storer               storage.Storer
	RedisClientPortal    redis.Cmdable
	RedisClientPortalExp time.Duration
	Biller               billing.Biller
	Metrics              *metrics.SessionMetrics
	Logger               log.Logger
	VetoMap              *VetoMap
	ServerMap            *ServerMap
	SessionMap           *SessionMap
	Counters             *SessionUpdateCounters
}

// =========================================================================================================

func SessionUpdateHandlerFunc(params *SessionUpdateParams) UDPHandlerFunc {

	return func(w io.Writer, incoming *UDPPacket) {

		start := time.Now()
		defer func() {
			if time.Since(start).Seconds() > 0.1 {
				level.Debug(params.Logger).Log("msg", "long session update")
				atomic.AddUint64(&params.Counters.LongDuration, 1)
				params.Metrics.LongDuration.Add(1)
			}
		}()

		params.Metrics.Invocations.Add(1)

		atomic.AddUint64(&params.Counters.Packets, 1)

		// First, read the session update packet header.
		// We have to read only the header first, because the rest of the session update packet depends on SDK version
		// and we don't know the version yet, since that's stored in the server data for this session, not in the packet.

		var header SessionUpdatePacketHeader
		if err := header.UnmarshalBinary(incoming.Data); err != nil {
			// level.Error(params.Logger).Log("msg", "could not read session update packet header", "err", err)
			params.Metrics.ErrorMetrics.UnserviceableUpdate.Add(1)
			params.Metrics.ErrorMetrics.ReadPacketHeaderFailure.Add(1)
			return
		}

		// Grab the server data corresponding to the server this session is talking to.
		// The server data is necessary for us to read the rest of the session update packet.

		// IMPORTANT: The server data *must* be treated as read only or it is not threadsafe!
		serverMutexStart := time.Now()
		serverDataReadOnly := params.ServerMap.GetServerData(header.ServerAddress.String())
		if serverDataReadOnly == nil {
			params.Metrics.ErrorMetrics.UnserviceableUpdate.Add(1)
			params.Metrics.ErrorMetrics.ServerDataMissing.Add(1)
			return
		}
		if time.Since(serverMutexStart).Seconds() > 0.1 {
			level.Debug(params.Logger).Log("msg", "long server mutex in session update")
		}

		// Now that we have the server data, we know the SDK version, so we can read the rest of the session update packet.

		var packet SessionUpdatePacket
		packet.Version = serverDataReadOnly.version
		if err := packet.UnmarshalBinary(incoming.Data); err != nil {
			// level.Error(params.Logger).Log("msg", "could not read session update packet", "err", err)
			params.Metrics.ErrorMetrics.UnserviceableUpdate.Add(1)
			params.Metrics.ErrorMetrics.ReadPacketFailure.Add(1)
			return
		}

		// IMPORTANT: The session data *must* be treated as read only or it is not threadsafe!
		sessionMutexStart := time.Now()
		sessionDataReadOnly := params.SessionMap.GetSessionData(header.SessionID)
		if time.Since(sessionMutexStart).Seconds() > 0.1 {
			level.Debug(params.Logger).Log("msg", "long session mutex in session update")
		}
		if sessionDataReadOnly == nil {
			sessionDataReadOnly = &SessionData{}
		}

		// Check the packet sequence number vs. the most recent sequence number in redis.
		// The packet sequence number must be at least as old as the current session sequence #
		// otherwise this is a stale session update packet from an older slice so we ignore it!

		if packet.Sequence < sessionDataReadOnly.sequence {
			// level.Error(params.Logger).Log("handler", "session", "msg", "packet too old", "packet sequence", packet.Sequence, "lastest sequence", sessionDataReadOnly.sequence)
			params.Metrics.ErrorMetrics.UnserviceableUpdate.Add(1)
			params.Metrics.ErrorMetrics.OldSequence.Add(1)
			return
		}

		// Look up the buyer entry by the customer id. At this point if we can't find it, just ignore the session and don't respond.
		// If somebody is sending us a session update with an invalid customer id, we don't need to waste any bandwidth responding to it.

		buyer, err := params.Storer.Buyer(packet.CustomerID)
		if err != nil {
			// level.Error(locallogger).Log("msg", "failed to get buyer from storage", "err", err, "customer_id", packet.CustomerID)
			params.Metrics.ErrorMetrics.UnserviceableUpdate.Add(1)
			params.Metrics.ErrorMetrics.BuyerNotFound.Add(1)
			return
		}

		// Check the session update packet is properly signed with the customer private key.
		// Any session update not signed is invalid, so we don't waste bandwidth responding to it.

		if !crypto.Verify(buyer.PublicKey, packet.GetSignData(), packet.Signature) {
			// level.Error(locallogger).Log("err", "could not verify session update packet", "customer_id", packet.CustomerID)
			params.Metrics.ErrorMetrics.VerifyFailure.Add(1)
			return
		}

		// When multiple session updates are in flight, especially under a retry storm, there can be simultaneous calls
		// to this handler for the same session and slice. It is *extremely important* that we don't generate multiple route
		// responses in this case, otherwise we'll bill our customers multiple times for the same slice!. Instead, we implement
		// a locking system here, such that if the same slices is already being processed in another handler, we block until
		// the other handler completes, then send down the cached session response.

		// IMPORTANT: This ensures we bill our customers only once per-slice!

		// todo: ryan. fun work below... :)

		// todo: acquire session lock for current slice. lock should be keyed on session id *and* current slice # (eg. the "sequence" in the packet)

		// todo: if we can't lock, somebody else has it... block until the lock is released.

		// todo: look for a cached session response for this session and slice sequence #

		// todo: if the cached session response exists, write the cached response back to the SDK and return without any further processing.

		// todo: if we did not acquire the lock, but no cached session response exists, something went wrong. increment an error counter and return.

		// todo: otherwise, carry on with regular processing below. add a defer to make sure we unlock. (we must have acquired the lock to get here)

		// Create the default response packet with a direct route and same SDK version as the server data.
		// This makes sure that we respond to the session update with the packet version the SDK expects.

		response := SessionResponsePacket{
			Version:              serverDataReadOnly.version,
			Sequence:             header.Sequence,
			SessionID:            header.SessionID,
			RouteType:            int32(routing.RouteTypeDirect),
			ServerRoutePublicKey: serverDataReadOnly.routePublicKey,
		}

		directRoute := routing.Route{}

		// The SDK uploads the result of pings to us for the previous 10 seconds (aka. "a slice")
		// These ping values are uploaded to the portal for visibility, and are used when we plan a route,
		// both to determine the baseline cost across the default public internet route (direct),
		// and to see how we have been doing so far if we served up a network next route for the previous slice (next).

		// IMPORTANT: We use the *minimum* RTT values instead of mean because these are stable even under significant jitter caused by wifi.

		lastNextStats := routing.Stats{
			RTT:        float64(packet.NextMinRTT),
			Jitter:     float64(packet.NextJitter),
			PacketLoss: float64(packet.NextPacketLoss),
		}

		lastDirectStats := routing.Stats{
			RTT:        float64(packet.DirectMinRTT),
			Jitter:     float64(packet.DirectJitter),
			PacketLoss: float64(packet.DirectPacketLoss),
		}

		// Keep track if this is the first slice in a new session
		// We need to check for this because we only run certain code paths on the first slice
		// ex. always send a direct route first slice, cache the near relays on the first slice, etc.
		newSession := packet.Sequence == 1

		// Retrieve the slice of near relays to the client.
		// Because the session data should be read only and we need to update the near relay slice,
		// we have to make a copy. We can then update the session data with that copy later.
		nearRelays := make([]routing.Relay, len(sessionDataReadOnly.nearRelays))
		copy(nearRelays, sessionDataReadOnly.nearRelays)

		// Pull some data from the session data that we may need to modify
		// Since it's not thread safe to modify the session data directly, we modify local copies and update it back later in the PostSessionUpdate
		routeExpireTimestamp := sessionDataReadOnly.routeExpireTimestamp
		location := sessionDataReadOnly.location
		routeDecision := sessionDataReadOnly.routeDecision
		onNNSliceCounter := sessionDataReadOnly.onNNSliceCounter
		committedData := sessionDataReadOnly.committedData
		committedData.Committed = !buyer.RoutingRulesSettings.EnableTryBeforeYouBuy // default state is based on the buyer's route shader. Will be overwritten later in the route decision if necessary.

		// Grab the veto reason so that we can use it later to keep off vetoed sessions and check if a session was vetoed this slice.
		vetoReason := params.VetoMap.GetVeto(header.SessionID)

		// Purchase 20 seconds ahead for new sessions and 10 seconds ahead for existing ones
		// This way we always have a 10 second buffer
		var sliceDuration uint64
		if newSession {
			sliceDuration = billing.BillingSliceSeconds * 2
			routeExpireTimestamp = start.Unix() + int64(sliceDuration)
		} else {
			sliceDuration = billing.BillingSliceSeconds
			routeExpireTimestamp += int64(sliceDuration)
		}

		// Run IP2Location on the session IP address.
		// IMPORTANT: Immediately after ip2location we *must* anonymize the IP address so there is no chance we accidentally
		// use or store the non-anonymized IP address past this point. This is an important business requirement because IP addresses
		// are considered private identifiable information according to the GDRP and CCPA. We must *never* collect or store non-anonymized IP addresses!

		if location.IsZero() {
			var err error
			location, err = params.IPLoc.LocateIP(packet.ClientAddress.IP)

			if err != nil {
				routeDecision = routing.Decision{
					OnNetworkNext: false,
					Reason:        routing.DecisionNoLocation,
				}

				if buyer.RoutingRulesSettings.EnableYouOnlyLiveOnce {
					// If we can't locate the client then make sure to veto the session when yolo is enabled,
					// since we can't serve them network next routes anyway
					routeDecision.Reason |= routing.DecisionVetoYOLO
				}

				params.Metrics.ErrorMetrics.ClientLocateFailure.Add(1)
				// IMPORTANT: We send a direct route response here because we want to see the session in our total session count, even if ip2loc fails.
				// Context: As soon as we don't respond to a session update, the SDK "falls back to direct" and stops sending session update packets.
				sendRouteResponse(w, &directRoute, params, &packet, &response, serverDataReadOnly, &buyer, &lastNextStats, &lastDirectStats, &location, nearRelays, routeDecision, vetoReason, onNNSliceCounter,
					committedData, sessionDataReadOnly.routeHash, sessionDataReadOnly.routeDecision.OnNetworkNext, start, routeExpireTimestamp, sessionDataReadOnly.tokenVersion, sliceDuration, params.RouterPrivateKey)
				return
			}
		}

		// Has this session been vetoed? A vetoed session must always go direct, because for some reason we have found a problem
		// when they are going across network next. Perhaps we made packet loss or latency worse for this player? To make sure
		// this doesn't happen repeatedly, the session is vetoed from taking network next until they connect to a new server.

		if vetoReason != routing.DecisionNoReason {
			routeDecision = routing.Decision{
				OnNetworkNext: false,
				Reason:        vetoReason,
			}

			sendRouteResponse(w, &directRoute, params, &packet, &response, serverDataReadOnly, &buyer, &lastNextStats, &lastDirectStats, &location, nearRelays, routeDecision, vetoReason, onNNSliceCounter,
				committedData, sessionDataReadOnly.routeHash, sessionDataReadOnly.routeDecision.OnNetworkNext, start, routeExpireTimestamp, sessionDataReadOnly.tokenVersion, sliceDuration, params.RouterPrivateKey)
			return
		}

		// We no longer need the full IP address, so immediately anonymize it
		packet.ClientAddress = AnonymizeAddr(packet.ClientAddress)

		if packet.ClientAddress.IP == nil {
			// If we can't anonymize the IP, then we somehow have a bad IP address, so just veto the session
			routeDecision = routing.Decision{
				OnNetworkNext: false,
				Reason:        routing.DecisionVetoNoRoute,
			}

			if buyer.RoutingRulesSettings.EnableYouOnlyLiveOnce {
				routeDecision.Reason |= routing.DecisionVetoYOLO
			}

			params.Metrics.ErrorMetrics.ClientIPAnonymizeFailure.Add(1)

			sendRouteResponse(w, &directRoute, params, &packet, &response, serverDataReadOnly, &buyer, &lastNextStats, &lastDirectStats, &location, nearRelays, routeDecision, vetoReason, onNNSliceCounter,
				committedData, sessionDataReadOnly.routeHash, sessionDataReadOnly.routeDecision.OnNetworkNext, start, routeExpireTimestamp, sessionDataReadOnly.tokenVersion, sliceDuration, params.RouterPrivateKey)
			return
		}

		// Use the route matrix to get a list of relays closest to the lat/long of the client.
		// These near relays are returned back down to the SDK for this slice. The SDK then pings these relays,
		// and reports the results back up to us in the next session update. We use the near relay pings to know
		// the cost of the first hop, from the client to the first relay in their route.

		routeMatrix := params.GetRouteProvider()

		if newSession {
			// If this is a new session, get the near relays from the route matrix to send down to the client.
			// The client will then report the ping stats in the next slice so we can properly serve them a route.
			// Because this is an expensive function, we only want to do this on the first slice, then just update
			// the relays with the client stats every subsequent slice.
			if nearRelays, err = routeMatrix.GetNearRelays(location.Latitude, location.Longitude, MaxNearRelays); err != nil {
				routeDecision = routing.Decision{
					OnNetworkNext: false,
					Reason:        routing.DecisionNoNearRelays,
				}

				if buyer.RoutingRulesSettings.EnableYouOnlyLiveOnce {
					routeDecision.Reason |= routing.DecisionVetoYOLO
				}

				params.Metrics.ErrorMetrics.NearRelaysLocateFailure.Add(1)
				sendRouteResponse(w, &directRoute, params, &packet, &response, serverDataReadOnly, &buyer, &lastNextStats, &lastDirectStats, &location, nearRelays, routeDecision, vetoReason, onNNSliceCounter,
					committedData, sessionDataReadOnly.routeHash, sessionDataReadOnly.routeDecision.OnNetworkNext, start, routeExpireTimestamp, sessionDataReadOnly.tokenVersion, sliceDuration, params.RouterPrivateKey)
				return
			}
		} else {
			// If this is not a new session, then just update the near relay list with the reported client stats.
			// We need to keep these stats updated so that if relays fluctuate we can recalculate the best route to serve.
			for i, nearRelay := range nearRelays {
				for j, clientNearRelayID := range packet.NearRelayIDs {
					if nearRelay.ID == clientNearRelayID {
						nearRelays[i].ClientStats.RTT = float64(packet.NearRelayMinRTT[j])
					}
				}
			}
		}

		// Fill out the near relay response data to send down to the client
		response.NumNearRelays = int32(len(nearRelays))
		response.NearRelayIDs = make([]uint64, len(nearRelays))
		response.NearRelayAddresses = make([]net.UDPAddr, len(nearRelays))
		for i := range nearRelays {
			response.NearRelayIDs[i] = nearRelays[i].ID
			response.NearRelayAddresses[i] = nearRelays[i].Addr
		}

		// If the session has fallen back to direct, just give them a direct route.
		if packet.FallbackToDirect {
			routeDecision = routing.Decision{
				OnNetworkNext: false,
				Reason:        routing.DecisionFallbackToDirect,
			}

			sendRouteResponse(w, &directRoute, params, &packet, &response, serverDataReadOnly, &buyer, &lastNextStats, &lastDirectStats, &location, nearRelays, routeDecision, vetoReason, onNNSliceCounter,
				committedData, sessionDataReadOnly.routeHash, sessionDataReadOnly.routeDecision.OnNetworkNext, start, routeExpireTimestamp, sessionDataReadOnly.tokenVersion, sliceDuration, params.RouterPrivateKey)
			return
		}

		// If this is a new session, send a direct response back no matter what
		// This is necessary because the SDK needs to ping near relays and send back those
		// stats for us to serve a network next route to that session.
		if newSession {
			routeDecision = routing.Decision{
				OnNetworkNext: false,
				Reason:        routing.DecisionInitialSlice,
			}

			sendRouteResponse(w, &directRoute, params, &packet, &response, serverDataReadOnly, &buyer, &lastNextStats, &lastDirectStats, &location, nearRelays, routeDecision, vetoReason, onNNSliceCounter,
				committedData, sessionDataReadOnly.routeHash, sessionDataReadOnly.routeDecision.OnNetworkNext, start, routeExpireTimestamp, sessionDataReadOnly.tokenVersion, sliceDuration, params.RouterPrivateKey)
			return
		}

		// If the buyer's route shader is set to force direct, we always send a direct route.
		// If we modulo the session ID and it's greater than or equal to the selection percentage, we send a direct route.
		// This selection percentage is useful for controling what percentage of session should be considered for a network next route.
		if buyer.RoutingRulesSettings.Mode == routing.ModeForceDirect || header.SessionID%100 >= uint64(buyer.RoutingRulesSettings.SelectionPercentage) {
			routeDecision = routing.Decision{
				OnNetworkNext: false,
				Reason:        routing.DecisionForceDirect,
			}

			sendRouteResponse(w, &directRoute, params, &packet, &response, serverDataReadOnly, &buyer, &lastNextStats, &lastDirectStats, &location, nearRelays, routeDecision, vetoReason, onNNSliceCounter,
				committedData, sessionDataReadOnly.routeHash, sessionDataReadOnly.routeDecision.OnNetworkNext, start, routeExpireTimestamp, sessionDataReadOnly.tokenVersion, sliceDuration, params.RouterPrivateKey)
			return
		}

		// If the buyer's route shader has the AB test enabled, send all odd numbered sessions direct.
		// This is a useful way to send roughly half of the sessions direct and the other half considering network next
		// to show customers that network next really does improve sessions.
		if buyer.RoutingRulesSettings.EnableABTest && header.SessionID%2 == 1 {
			routeDecision = routing.Decision{
				OnNetworkNext: false,
				Reason:        routing.DecisionABTestDirect,
			}

			sendRouteResponse(w, &directRoute, params, &packet, &response, serverDataReadOnly, &buyer, &lastNextStats, &lastDirectStats, &location, nearRelays, routeDecision, vetoReason, onNNSliceCounter,
				committedData, sessionDataReadOnly.routeHash, sessionDataReadOnly.routeDecision.OnNetworkNext, start, routeExpireTimestamp, sessionDataReadOnly.tokenVersion, sliceDuration, params.RouterPrivateKey)
			return
		}

		// Retrieve all relays within the game server's datacenter.
		// This way we can find all of the routes between the client's near relays and the
		// relays in the same datacenter as the server (effectively 0 RTT from datacenter relay -> game server)
		datacenterRelays := routeMatrix.GetDatacenterRelays(serverDataReadOnly.datacenter)
		if len(datacenterRelays) == 0 {
			routeDecision = routing.Decision{
				OnNetworkNext: false,
				Reason:        routing.DecisionDatacenterHasNoRelays,
			}

			if buyer.RoutingRulesSettings.EnableYouOnlyLiveOnce {
				routeDecision.Reason |= routing.DecisionVetoYOLO
			}

			params.Metrics.ErrorMetrics.NoRelaysInDatacenter.Add(1)
			sendRouteResponse(w, &directRoute, params, &packet, &response, serverDataReadOnly, &buyer, &lastNextStats, &lastDirectStats, &location, nearRelays, routeDecision, vetoReason, onNNSliceCounter,
				committedData, sessionDataReadOnly.routeHash, sessionDataReadOnly.routeDecision.OnNetworkNext, start, routeExpireTimestamp, sessionDataReadOnly.tokenVersion, sliceDuration, params.RouterPrivateKey)
			return
		}

		// Now that we have a slice of all routes a client can take, we need to get the best route.
		// This could either be a network next route or a direct route.
		var bestRoute *routing.Route
		bestRoute, routeDecision = GetBestRoute(routeMatrix, nearRelays, datacenterRelays, &params.Metrics.ErrorMetrics, &buyer,
			sessionDataReadOnly.routeHash, sessionDataReadOnly.routeDecision, &lastNextStats, &lastDirectStats, onNNSliceCounter, &committedData, &directRoute)

		if routeDecision.OnNetworkNext {
			onNNSliceCounter++
		}

		// Send a session update response back to the SDK.

		// IMPORTANT: If the SDK does not receive a session update response quickly, it will resend the session update packet
		// to us 10X every second. This is extremely aggressive resend behavior, and in hindsight was probably a mistake on our part.
		// Even so, we want to avoid it if at all possible, because it greatly increases the number of packets we have to process,
		// and the cost to us to service each session.

		// IMPORTANT: In future SDK versions we are much less aggressive with session update packet retries. eg. 3.4.5 and later.

		sendRouteResponse(w, bestRoute, params, &packet, &response, serverDataReadOnly, &buyer, &lastNextStats, &lastDirectStats, &location, nearRelays, routeDecision, vetoReason, onNNSliceCounter,
			committedData, sessionDataReadOnly.routeHash, sessionDataReadOnly.routeDecision.OnNetworkNext, start, routeExpireTimestamp, sessionDataReadOnly.tokenVersion, sliceDuration, params.RouterPrivateKey)
	}
}

// GetBestRoute returns the best route that a session can take for this slice. If we can't serve a network next route, the returned route will be the passed in direct route.
// This function can either return a network next route or a direct route, and it also returns a reason as to why the route was chosen.
func GetBestRoute(routeMatrix RouteProvider, nearRelays []routing.Relay, datacenterRelays []routing.Relay, errorMetrics *metrics.SessionErrorMetrics,
	buyer *routing.Buyer, prevRouteHash uint64, prevRouteDecision routing.Decision, lastNextStats *routing.Stats, lastDirectStats *routing.Stats,
	onNNSliceCounter uint64, committedData *routing.CommittedData, directRoute *routing.Route) (*routing.Route, routing.Decision) {
	// We need to get a next route to compare against direct
	nextRoute := GetNextRoute(routeMatrix, nearRelays, datacenterRelays, errorMetrics, buyer, prevRouteHash)
	if nextRoute == nil {
		// We couldn't find a network next route at all. This may happen if something goes wrong with the route matrix or if relays are flickering.
		decision := routing.Decision{OnNetworkNext: false, Reason: routing.DecisionNoNextRoute}

		if buyer.RoutingRulesSettings.EnableYouOnlyLiveOnce {
			decision.Reason = routing.DecisionVetoNoRoute | routing.DecisionVetoYOLO
		}

		return directRoute, decision
	}

	// If the buyer's route shader is set to force next, don't bother running the decision logic,
	// just send back the route we've selected.
	// Make sure to set the committed flag to true so the SDK always commits to the route.
	if buyer.RoutingRulesSettings.Mode == routing.ModeForceNext {
		committedData.Pending = false
		committedData.ObservedSliceCounter = 0
		committedData.Committed = true

		return nextRoute, routing.Decision{OnNetworkNext: true, Reason: routing.DecisionForceNext}
	}

	// Now that we have a next route, we have to decide if the route is worth taking over direct.
	// This process can vary based on the customer's route shader.

	// The logic is as follows:
	//	1. Decide if we should accelerate a session (direct -> next). If a session is already on network next, this decision is skipped.
	//	2. Decide if we should bring a session back to direct (next -> direct). If a session is already on direct, this decision is skipped.
	//	3. Decide if we should veto a session (next -> direct permanently). If a session is already on direct, this decision is skipped.
	// 	4. Decide if we should consider multipath. If multipath is enabled, then the decision process is reset and only multipath logic is considered.
	//	5. Decide if we should run the committed logic. This is only run if the buyer has "try before you buy" enabled in the route shader.
	// More information on how each decision is made can be found in their respective decision functions.
	deciderFuncs := []routing.DecisionFunc{
		routing.DecideUpgradeRTT(float64(buyer.RoutingRulesSettings.RTTThreshold)),
		routing.DecideDowngradeRTT(float64(buyer.RoutingRulesSettings.RTTHysteresis), buyer.RoutingRulesSettings.EnableYouOnlyLiveOnce),
		routing.DecideVeto(onNNSliceCounter, float64(buyer.RoutingRulesSettings.RTTVeto), buyer.RoutingRulesSettings.EnablePacketLossSafety, buyer.RoutingRulesSettings.EnableYouOnlyLiveOnce),
		routing.DecideMultipath(buyer.RoutingRulesSettings.EnableMultipathForRTT, buyer.RoutingRulesSettings.EnableMultipathForJitter, buyer.RoutingRulesSettings.EnableMultipathForPacketLoss, float64(buyer.RoutingRulesSettings.RTTThreshold)),
	}

	if buyer.RoutingRulesSettings.EnableTryBeforeYouBuy {
		deciderFuncs = append(deciderFuncs,
			routing.DecideCommitted(prevRouteDecision.OnNetworkNext, uint8(buyer.RoutingRulesSettings.TryBeforeYouBuyMaxSlices), buyer.RoutingRulesSettings.EnableYouOnlyLiveOnce, committedData))
	} else {
		// If we aren't using the try before you buy logic, then we always want to commit to routes
		committedData.Pending = false
		committedData.ObservedSliceCounter = 0
		committedData.Committed = true
	}

	routeDecision := nextRoute.Decide(prevRouteDecision, lastNextStats, lastDirectStats, deciderFuncs...)

	// As a safety measure, if the route decision goes from on network next to direct with yolo enabled for any reason, veto the session with yolo reason
	if buyer.RoutingRulesSettings.EnableYouOnlyLiveOnce && prevRouteDecision.OnNetworkNext && !routeDecision.OnNetworkNext {
		if routeDecision.Reason&routing.DecisionVetoYOLO == 0 {
			routeDecision.Reason |= routing.DecisionVetoYOLO
		}
	}

	if routeDecision.OnNetworkNext {
		return nextRoute, routeDecision
	}

	return directRoute, routeDecision
}

// GetNextRoute returns the best network next route a session can take for this slice, or nil if a route couldn't be found.
func GetNextRoute(routeMatrix RouteProvider, nearRelays []routing.Relay, datacenterRelays []routing.Relay, errorMetrics *metrics.SessionErrorMetrics,
	buyer *routing.Buyer, prevRouteHash uint64) *routing.Route {
	// We need to get all of the routes from the route matrix that connect any of the client's near relays and any of the game server's datacenter relays
	routes, err := routeMatrix.GetRoutes(nearRelays, datacenterRelays)
	if err != nil {
		errorMetrics.RouteFailure.Add(1)
		return nil
	}

	// Now that we have all of the routes a client could take, we can start filtering the slice down to determine the best route among them.

	// The logic is as follows:
	//	1. Only select routes whose relays have session counts of less than 80% of their maximum allowed session counts (this is to avoid overloading a relay).
	// 	2. Find the route with the lowest RTT, and return all routes whose RTT is with the given epsilon value. These are "acceptable routes".
	// 	3. If the route the session is already taking is within the set of acceptable routes, choose that one. If it's not, continue to step 4.
	// 	4. Choose a random destination relay (since all destination relays are in the same datacenter and have effectively the same RTT from relay -> game server)
	//		and only select routes with that destination relay
	//	5. If we still don't only have 1 route, choose a random one.
	selectorFuncs := []routing.SelectorFunc{
		routing.SelectUnencumberedRoutes(0.8),
		routing.SelectAcceptableRoutesFromBestRTT(float64(buyer.RoutingRulesSettings.RTTEpsilon)),
		routing.SelectContainsRouteHash(prevRouteHash),
		routing.SelectRoutesByRandomDestRelay(rand.NewSource(rand.Int63())),
		routing.SelectRandomRoute(rand.NewSource(rand.Int63())),
	}

	for _, selectorFunc := range selectorFuncs {
		routes = selectorFunc(routes)

		if len(routes) == 0 {
			break
		}
	}

	if len(routes) == 0 {
		errorMetrics.RouteSelectFailure.Add(1)
		return nil
	}

	return &routes[0]
}

func PostSessionUpdate(params *SessionUpdateParams, packet *SessionUpdatePacket, response *SessionResponsePacket, serverDataReadOnly *ServerData,
	chosenRoute *routing.Route, lastNextStats *routing.Stats, lastDirectStats *routing.Stats, location *routing.Location, nearRelays []routing.Relay,
	routeDecision routing.Decision, timeNow time.Time, envelopeBytesUp uint64, envelopeBytesDown uint64) {

	// IMPORTANT: we actually need to display the true datacenter name in the demo and demo plus views,
	// while in the customer view of the portal, we need to display the alias. this is because aliases will
	// shortly become per-customer, thus there is really no global concept of "multiplay.losangeles", for example.

	// todo: ryan, please make it so. you'll probably have to send both datacenter names down to the portal
	// and let the portal select which one to display, depending on context.

	datacenterName := serverDataReadOnly.datacenter.Name
	datacenterAlias := serverDataReadOnly.datacenter.AliasName

	// Send a massive amount of data to the portal via redis.
	// This drives all the stuff you see in the portal, including the map and top sessions list.
	// We send it via redis because google pubsub is not able to deliver data quickly enough.

	// IMPORTANT: We could possibly offload some work from here by sending to another service
	// via redis pubsub (this is different to google pubsub).

	if err := updatePortalData(params.RedisClientPortal, params.RedisClientPortalExp, packet, lastNextStats, lastDirectStats, chosenRoute.Relays,
		packet.OnNetworkNext, datacenterName, location, nearRelays, timeNow, routing.IsMultipath(routeDecision), datacenterAlias); err != nil {
		// level.Error(params.Logger).Log("msg", "could not update portal data", "err", err)
		params.Metrics.ErrorMetrics.UpdatePortalFailure.Add(1)
	}

	// Send billing specific data to the billing service via google pubsub
	// The billing service subscribes to this topic, and writes the billing data to bigquery.
	// We tried writing to bigquery directly here, but it didn't work because bigquery would stall out.
	// BigQuery really doesn't make performance guarantees on how fast it is to load data, so we need
	// pubsub to act as a queue to smooth that out. Pubsub can buffer billing data for up to 7 days.

	nextRelays := [billing.BillingEntryMaxRelays]uint64{}
	for i := 0; i < len(chosenRoute.Relays) && i < len(nextRelays); i++ {
		nextRelays[i] = chosenRoute.Relays[i].ID
	}

	onNetworkNext := len(chosenRoute.Relays) > 0

	bytes := envelopeBytesDown + envelopeBytesUp
	totalPriceCents := 1000000000 * bytes // 1 cent per GB
	if !onNetworkNext {
		// no revenue on direct
		totalPriceCents = 0
	}

	billingEntry := billing.BillingEntry{
		BuyerID:                   packet.CustomerID,
		SessionID:                 packet.SessionID,
		SliceNumber:               uint32(packet.Sequence),
		DirectRTT:                 float32(lastDirectStats.RTT),
		DirectJitter:              float32(lastDirectStats.Jitter),
		DirectPacketLoss:          float32(lastDirectStats.PacketLoss),
		Next:                      onNetworkNext,
		NextRTT:                   float32(chosenRoute.Stats.RTT),
		NextJitter:                float32(chosenRoute.Stats.Jitter),
		NextPacketLoss:            float32(chosenRoute.Stats.PacketLoss),
		NumNextRelays:             uint8(len(chosenRoute.Relays)),
		NextRelays:                nextRelays,
		TotalPrice:                totalPriceCents,
		ClientToServerPacketsLost: packet.PacketsLostClientToServer,
		ServerToClientPacketsLost: packet.PacketsLostServerToClient,
	}

	if err := params.Biller.Bill(context.Background(), &billingEntry); err != nil {
		// level.Error(params.Logger).Log("msg", "could not submit billing entry", "err", err)
		params.Metrics.ErrorMetrics.BillingFailure.Add(1)
	}
}

// =========================================================================================================

// todo: cut down
/*
func SessionUpdateHandlerFunc(logger log.Logger, redisClientCache redis.Cmdable, redisClientPortal redis.Cmdable, redisClientPortalExp time.Duration, storer storage.Storer, rp RouteProvider, iploc routing.IPLocator, geoClient *routing.GeoClient, metrics *metrics.SessionMetrics, biller billing.Biller, serverPrivateKey []byte, routerPrivateKey []byte) UDPHandlerFunc {

	logger = log.With(logger, "handler", "session")

	return func(w io.Writer, incoming *UDPPacket) {
		durationStart := time.Now()
		defer func() {
			durationSince := time.Since(durationStart)
			level.Info(logger).Log("duration", durationSince.Milliseconds())
			metrics.Invocations.Add(1)
		}()

		timestampNow := durationStart

		// Whether or not we should make a route selection/decision on a network next route, or serve a direct route
		shouldSelect := true
		shouldDecide := true

		// Flag to check if this session is a new session
		newSession := false

		// Deserialize the Session packet header
		var header SessionUpdatePacketHeader
		if err := header.UnmarshalBinary(incoming.Data); err != nil {
			level.Error(logger).Log("msg", "could not read packet header", "err", err)
			metrics.ErrorMetrics.UnserviceableUpdate.Add(1)
			metrics.ErrorMetrics.ReadPacketFailure.Add(1)
			return
		}

		serverCacheKey := fmt.Sprintf("SERVER-%d-%s", header.CustomerID, header.ServerAddress.String())
		sessionCacheKey := fmt.Sprintf("SESSION-%d-%d", header.CustomerID, header.SessionID)
		vetoCacheKey := fmt.Sprintf("VETO-%d-%d", header.CustomerID, header.SessionID)

		locallogger := log.With(logger, "src_addr", incoming.SourceAddr.String(), "server_addr", header.ServerAddress.String(), "session_id", header.SessionID)

		var serverCacheEntry ServerCacheEntry
		var sessionCacheEntry SessionCacheEntry
		var vetoCacheEntry VetoCacheEntry

		// Start building session response packet, defaulting to a direct route
		response := SessionResponsePacket{
			Sequence:  header.Sequence,
			SessionID: header.SessionID,
			RouteType: int32(routing.RouteTypeDirect),
		}

		// Build a redis transaction to make a single network call
		tx := redisClientCache.TxPipeline()
		{
			serverCacheCmd := tx.Get(serverCacheKey)
			sessionCacheCmd := tx.Get(sessionCacheKey)
			vetoCacheCmd := tx.Get(vetoCacheKey)
			if _, err := tx.Exec(); err != nil && err != redis.Nil {

				level.Error(locallogger).Log("msg", "failed to execute redis pipeline", "err", err)
				metrics.ErrorMetrics.UnserviceableUpdate.Add(1)
				metrics.ErrorMetrics.PipelineExecFailure.Add(1)
				return
			}

			// Note that if we fail to retrieve the server data, we don't bother responding since server will ignore response without ServerRoutePublicKey set
			// See next_server_internal_process_packet in next.cpp for full requirements of response packet
			if sessionCacheCmd.Err() != redis.Nil {
				serverCacheData, err := serverCacheCmd.Bytes()
				if err != nil {
					// This error case should never happen, can't produce it in test cases, but leaving it in anyway

					level.Error(locallogger).Log("msg", "failed to get server bytes", "err", err)
					metrics.ErrorMetrics.UnserviceableUpdate.Add(1)
					metrics.ErrorMetrics.GetServerDataFailure.Add(1)
					return
				}
				if err := serverCacheEntry.UnmarshalBinary(serverCacheData); err != nil {

					level.Error(locallogger).Log("msg", "failed to unmarshal server bytes", "err", err)
					metrics.ErrorMetrics.UnserviceableUpdate.Add(1)
					metrics.ErrorMetrics.UnmarshalServerDataFailure.Add(1)
					return
				}

				// Set public key on response as soon as we get it
				response.ServerRoutePublicKey = serverCacheEntry.Server.PublicKey
			} else {
				level.Error(locallogger).Log("msg", "server data missing")
				metrics.ErrorMetrics.UnserviceableUpdate.Add(1)
				metrics.ErrorMetrics.ServerDataMissing.Add(1)
				return
			}

			if sessionCacheCmd.Err() != redis.Nil {
				sessionCacheData, err := sessionCacheCmd.Bytes()
				if err != nil {
					// This error case should never happen, can't produce it in test cases, but leaving it in anyway

					level.Error(locallogger).Log("msg", "failed to get session bytes", "err", err)
					if _, err := writeSessionErrorResponse(w, response, serverPrivateKey, metrics.DirectSessions, metrics.ErrorMetrics.UnserviceableUpdate, metrics.ErrorMetrics.GetSessionDataFailure); err != nil {

						level.Error(locallogger).Log("msg", "failed to write session error response", "err", err)
						metrics.ErrorMetrics.WriteResponseFailure.Add(1)
					}
					return
				}

				if len(sessionCacheData) != 0 {
					if err := sessionCacheEntry.UnmarshalBinary(sessionCacheData); err != nil {

						level.Error(locallogger).Log("msg", "failed to unmarshal session bytes", "err", err)
						if _, err := writeSessionErrorResponse(w, response, serverPrivateKey, metrics.DirectSessions, metrics.ErrorMetrics.UnserviceableUpdate, metrics.ErrorMetrics.UnmarshalSessionDataFailure); err != nil {

							level.Error(locallogger).Log("msg", "failed to write session error response", "err", err)
							metrics.ErrorMetrics.WriteResponseFailure.Add(1)
						}
						return
					}
				}
			} else {
				// Session not cached yet, mark it as a new session
				sessionCacheEntry.RouteDecision = routing.Decision{
					OnNetworkNext: false,
					Reason:        routing.DecisionInitialSlice,
				}
				shouldSelect = false
				newSession = true
			}

			if vetoCacheCmd.Err() != redis.Nil {
				vetoCacheData, err := vetoCacheCmd.Bytes()
				if err != nil {
					// This error case should never happen, can't produce it in test cases, but leaving it in anyway

					level.Error(locallogger).Log("msg", "failed to get session veto bytes", "err", err)
					if _, err := writeSessionErrorResponse(w, response, serverPrivateKey, metrics.DirectSessions, metrics.ErrorMetrics.UnserviceableUpdate, metrics.ErrorMetrics.GetVetoDataFailure); err != nil {

						level.Error(locallogger).Log("msg", "failed to write session veto error response", "err", err)
						metrics.ErrorMetrics.WriteResponseFailure.Add(1)
					}
					return
				}

				if len(vetoCacheData) != 0 {
					if err := vetoCacheEntry.UnmarshalBinary(vetoCacheData); err != nil {

						level.Error(locallogger).Log("msg", "failed to unmarshal session veto bytes", "err", err)
						if _, err := writeSessionErrorResponse(w, response, serverPrivateKey, metrics.DirectSessions, metrics.ErrorMetrics.UnserviceableUpdate, metrics.ErrorMetrics.UnmarshalVetoDataFailure); err != nil {

							level.Error(locallogger).Log("msg", "failed to write session veto error response", "err", err)
							metrics.ErrorMetrics.WriteResponseFailure.Add(1)
						}
						return
					}

					if !vetoCacheEntry.VetoTimestamp.Before(timestampNow) {
						// If we have a veto cache entry and the session is still being vetoed, set the route decision as default-vetoed
						sessionCacheEntry.RouteDecision = routing.Decision{
							OnNetworkNext: false,
							Reason:        vetoCacheEntry.Reason,
						}
					}
				}
			}
		}

		// Deserialize the Session packet now that we have the version
		var packet SessionUpdatePacket
		packet.Version = serverCacheEntry.SDKVersion
		response.Version = serverCacheEntry.SDKVersion
		if err := packet.UnmarshalBinary(incoming.Data); err != nil {
			level.Error(logger).Log("msg", "could not read packet", "err", err)
			metrics.ErrorMetrics.UnserviceableUpdate.Add(1)
			metrics.ErrorMetrics.ReadPacketFailure.Add(1)
			return
		}

		locallogger = log.With(locallogger, "client_addr", packet.ClientAddress.String(), "datacenter_id", serverCacheEntry.Datacenter.ID)

		buyer, err := storer.Buyer(packet.CustomerID)
		if err != nil {
			level.Error(locallogger).Log("msg", "failed to get buyer from storage", "err", err, "customer_id", packet.CustomerID)
			if _, err := writeSessionErrorResponse(w, response, serverPrivateKey, metrics.DirectSessions, metrics.ErrorMetrics.UnserviceableUpdate, metrics.ErrorMetrics.BuyerNotFound); err != nil {

				level.Error(locallogger).Log("msg", "failed to write session error response", "err", err)
				metrics.ErrorMetrics.WriteResponseFailure.Add(1)
			}
			return
		}

		locallogger = log.With(locallogger, "customer_id", packet.CustomerID)

		if !crypto.Verify(buyer.PublicKey, packet.GetSignData(), packet.Signature) {
			err := errors.New("failed to verify packet signature with buyer public key")
			level.Error(locallogger).Log("err", err)
			if _, err := writeSessionErrorResponse(w, response, serverPrivateKey, metrics.DirectSessions, metrics.ErrorMetrics.UnserviceableUpdate, metrics.ErrorMetrics.VerifyFailure); err != nil {

				level.Error(locallogger).Log("msg", "failed to write session error response", "err", err)
				metrics.ErrorMetrics.WriteResponseFailure.Add(1)
			}
			return
		}

		switch seq := packet.Sequence; {
		case seq < sessionCacheEntry.Sequence:
			err := fmt.Errorf("packet sequence too old. current_sequence %v, previous sequence %v", packet.Sequence, sessionCacheEntry.Sequence)
			level.Error(locallogger).Log("err", err)
			if _, err := writeSessionErrorResponse(w, response, serverPrivateKey, metrics.DirectSessions, metrics.ErrorMetrics.UnserviceableUpdate, metrics.ErrorMetrics.OldSequence); err != nil {

				level.Error(locallogger).Log("msg", "failed to write session error response", "err", err)
				metrics.ErrorMetrics.WriteResponseFailure.Add(1)
			}
			return
		case seq == sessionCacheEntry.Sequence:
			if _, err := w.Write(sessionCacheEntry.Response); err != nil {

				level.Error(locallogger).Log("err", err)
				metrics.ErrorMetrics.UnserviceableUpdate.Add(1)
				metrics.ErrorMetrics.WriteCachedResponseFailure.Add(1)
			}
			return
		}

		// Set initial route decision values
		nnStats := routing.Stats{
			RTT:        float64(packet.NextMinRTT),
			Jitter:     float64(packet.NextJitter),
			PacketLoss: float64(packet.NextPacketLoss),
		}

		directStats := routing.Stats{
			RTT:        float64(packet.DirectMinRTT),
			Jitter:     float64(packet.DirectJitter),
			PacketLoss: float64(packet.DirectPacketLoss),
		}

		chosenRoute := routing.Route{
			Stats:  directStats,
			Relays: make([]routing.Relay, 0),
		}

		routeDecision := sessionCacheEntry.RouteDecision

		// Purchase 20 seconds ahead for new sessions and 10 seconds ahead for existing ones
		// This way we always have a 10 second buffer
		timestampStart := sessionCacheEntry.TimestampStart
		timestampExpire := sessionCacheEntry.TimestampExpire
		var sliceDuration uint64
		if newSession {
			sliceDuration = billing.BillingSliceSeconds * 2
			timestampStart = timestampNow
			timestampExpire = timestampNow.Add(time.Duration(sliceDuration) * time.Second)
		} else {
			sliceDuration = billing.BillingSliceSeconds
			timestampExpire = timestampExpire.Add(time.Duration(sliceDuration) * time.Second)
		}

		// Check if the client is falling back to direct
		if packet.FallbackToDirect {
			// if we are about to issue the second slice, and the client has already fallen
			// back to direct, then this is an early fallback. the first slice we issue is
			// always a direct route, so the client can never have been on a Network Next route
			// at this point, and thus they shouldn't be falling back from anything.
			if timestampNow.Sub(timestampStart) < (billing.BillingSliceSeconds*1.5*time.Second) &&
				timestampNow.Sub(timestampStart) > (billing.BillingSliceSeconds*0.5*time.Second) {
				level.Error(locallogger).Log("err", "early fallback to direct", "flag", FallbackFlagText(packet.Flags))
				if _, err := writeSessionErrorResponse(w, response, serverPrivateKey, metrics.DirectSessions, metrics.ErrorMetrics.UnserviceableUpdate, metrics.ErrorMetrics.EarlyFallbackToDirect); err != nil {

					level.Error(locallogger).Log("msg", "failed to write session error response", "err", err)
					metrics.ErrorMetrics.WriteResponseFailure.Add(1)
				}
				return
			}

			level.Error(locallogger).Log("err", "fallback to direct", "flag", FallbackFlagText(packet.Flags))

			responseData, err := writeSessionResponse(w, response, serverPrivateKey)
			if err != nil {

				level.Error(locallogger).Log("msg", "failed to write session response", "err", err)
				metrics.ErrorMetrics.WriteResponseFailure.Add(1)
				return
			}

			metrics.DirectSessions.Add(1)

			routeDecision = routing.Decision{OnNetworkNext: false, Reason: routing.DecisionFallbackToDirect}
			addRouteDecisionMetric(routeDecision, metrics)

			if err := updateCacheEntries(redisClientCache, sessionCacheKey, vetoCacheKey, sessionCacheEntry, vetoCacheEntry, packet, chosenRoute.Hash64(), routeDecision, timestampStart, timestampExpire,
				responseData, directStats.RTT, nnStats.RTT, sessionCacheEntry.Location); err != nil {

				level.Error(locallogger).Log("msg", "failed to update session", "err", err)
				metrics.ErrorMetrics.UpdateCacheFailure.Add(1)
			}

			if err := updatePortalData(redisClientPortal, redisClientPortalExp, packet, nnStats, directStats, chosenRoute.Relays, sessionCacheEntry.RouteDecision.OnNetworkNext, serverCacheEntry.Datacenter.Name, sessionCacheEntry.Location, timestampNow, response.Multipath); err != nil {

				level.Error(locallogger).Log("msg", "failed to update portal data", "err", err)
			}

			if err := submitBillingEntry(biller, serverCacheEntry, sessionCacheEntry, packet, response, buyer, chosenRoute, sessionCacheEntry.Location, storer, nil,
				routeDecision, sliceDuration, timestampStart, timestampNow, newSession); err != nil {

				level.Error(locallogger).Log("msg", "billing failed", "err", err)
				metrics.ErrorMetrics.BillingFailure.Add(1)
			}

			return
		}

		// Look up the client's IP if the previous SessionCacheEntry is the zero value
		if sessionCacheEntry.Location.IsZero() {
			location, err := iploc.LocateIP(packet.ClientAddress.IP)
			if err != nil {

				level.Error(locallogger).Log("msg", "failed to locate client", "err", err)
				responseData, err := writeSessionErrorResponse(w, response, serverPrivateKey, metrics.DirectSessions, metrics.ErrorMetrics.UnserviceableUpdate, metrics.ErrorMetrics.ClientLocateFailure)
				if err != nil {

					level.Error(locallogger).Log("msg", "failed to write session error response", "err", err)
					metrics.ErrorMetrics.WriteResponseFailure.Add(1)
					return
				}

				if err := updateCacheEntries(redisClientCache, sessionCacheKey, vetoCacheKey, sessionCacheEntry, vetoCacheEntry, packet, chosenRoute.Hash64(), routeDecision, timestampStart, timestampExpire,
					responseData, directStats.RTT, nnStats.RTT, location); err != nil {

					level.Error(locallogger).Log("msg", "failed to update session", "err", err)
					metrics.ErrorMetrics.UpdateCacheFailure.Add(1)
				}

				if err := updatePortalData(redisClientPortal, redisClientPortalExp, packet, nnStats, directStats, chosenRoute.Relays, sessionCacheEntry.RouteDecision.OnNetworkNext, serverCacheEntry.Datacenter.Name, location, timestampNow, response.Multipath); err != nil {

					level.Error(locallogger).Log("msg", "failed to update portal data", "err", err)
				}

				if err := submitBillingEntry(biller, serverCacheEntry, sessionCacheEntry, packet, response, buyer, chosenRoute, location, storer, nil,
					routing.Decision{OnNetworkNext: false, Reason: routing.DecisionNoNearRelays}, sliceDuration, timestampStart, timestampNow, newSession); err != nil {

					level.Error(locallogger).Log("msg", "billing failed", "err", err)
					metrics.ErrorMetrics.BillingFailure.Add(1)
				}

				return
			}

			sessionCacheEntry.Location = location
		}
		level.Debug(locallogger).Log("client_ip", packet.ClientAddress.IP.String(), "lat", sessionCacheEntry.Location.Latitude, "long", sessionCacheEntry.Location.Longitude)

		clientRelays, err := geoClient.RelaysWithin(sessionCacheEntry.Location.Latitude, sessionCacheEntry.Location.Longitude, 2500, "mi")

		if len(clientRelays) == 0 || err != nil {

			level.Error(locallogger).Log("msg", "failed to locate relays near client", "err", err)
			responseData, err := writeSessionErrorResponse(w, response, serverPrivateKey, metrics.DirectSessions, metrics.ErrorMetrics.UnserviceableUpdate, metrics.ErrorMetrics.NearRelaysLocateFailure)
			if err != nil {

				level.Error(locallogger).Log("msg", "failed to write session error response", "err", err)
				metrics.ErrorMetrics.WriteResponseFailure.Add(1)
				return
			}

			if err := updateCacheEntries(redisClientCache, sessionCacheKey, vetoCacheKey, sessionCacheEntry, vetoCacheEntry, packet, chosenRoute.Hash64(), routeDecision, timestampStart, timestampExpire,
				responseData, directStats.RTT, nnStats.RTT, sessionCacheEntry.Location); err != nil {

				level.Error(locallogger).Log("msg", "failed to update session", "err", err)
				metrics.ErrorMetrics.UpdateCacheFailure.Add(1)
			}

			if err := updatePortalData(redisClientPortal, redisClientPortalExp, packet, nnStats, directStats, chosenRoute.Relays, sessionCacheEntry.RouteDecision.OnNetworkNext, serverCacheEntry.Datacenter.Name, sessionCacheEntry.Location, timestampNow, response.Multipath); err != nil {

				level.Error(locallogger).Log("msg", "failed to update portal data", "err", err)
			}

			if err := submitBillingEntry(biller, serverCacheEntry, sessionCacheEntry, packet, response, buyer, chosenRoute, sessionCacheEntry.Location, storer, clientRelays,
				routing.Decision{OnNetworkNext: false, Reason: routing.DecisionNoNearRelays}, sliceDuration, timestampStart, timestampNow, newSession); err != nil {

				level.Error(locallogger).Log("msg", "billing failed", "err", err)
				metrics.ErrorMetrics.BillingFailure.Add(1)
			}

			return
		}

		// Clamp relay count to max
		if len(clientRelays) > int(MaxNearRelays) {
			clientRelays = clientRelays[:MaxNearRelays]
		}

		// We need to do this because RelaysWithin only has the ID of the relay and we need the Addr and PublicKey too
		// Maybe we consider a nicer way to do this in the future
		for idx := range clientRelays {
			clientRelays[idx], _ = rp.ResolveRelay(clientRelays[idx].ID)
		}

		// Fill in the near relays
		response.NumNearRelays = int32(len(clientRelays))
		response.NearRelayIDs = make([]uint64, len(clientRelays))
		response.NearRelayAddresses = make([]net.UDPAddr, len(clientRelays))
		for idx, relay := range clientRelays {
			response.NearRelayIDs[idx] = relay.ID
			response.NearRelayAddresses[idx] = relay.Addr
		}

		if !serverCacheEntry.Datacenter.Enabled {
			// datacenter is disabled, so next routes can't be made
			level.Error(locallogger).Log("msg", "datacenter is disabled", "datacenter", serverCacheEntry.Datacenter.Name)
			responseData, err := writeSessionErrorResponse(w, response, serverPrivateKey, metrics.DirectSessions, metrics.ErrorMetrics.UnserviceableUpdate, metrics.ErrorMetrics.DatacenterDisabled)
			if err != nil {

				level.Error(locallogger).Log("msg", "failed to write session error response", "err", err)
				metrics.ErrorMetrics.WriteResponseFailure.Add(1)
				return
			}

			if err := updateCacheEntries(redisClientCache, sessionCacheKey, vetoCacheKey, sessionCacheEntry, vetoCacheEntry, packet, chosenRoute.Hash64(), routeDecision, timestampStart, timestampExpire,
				responseData, directStats.RTT, nnStats.RTT, sessionCacheEntry.Location); err != nil {

				level.Error(locallogger).Log("msg", "failed to update session", "err", err)
				metrics.ErrorMetrics.UpdateCacheFailure.Add(1)
			}

			if err := updatePortalData(redisClientPortal, redisClientPortalExp, packet, nnStats, directStats, chosenRoute.Relays, sessionCacheEntry.RouteDecision.OnNetworkNext, serverCacheEntry.Datacenter.Name, sessionCacheEntry.Location, timestampNow, response.Multipath); err != nil {

				level.Error(locallogger).Log("msg", "failed to update portal data", "err", err)
			}

			if err := submitBillingEntry(biller, serverCacheEntry, sessionCacheEntry, packet, response, buyer, chosenRoute, sessionCacheEntry.Location, storer, clientRelays,
				routing.Decision{OnNetworkNext: false, Reason: routing.DecisionDatacenterDisabled}, sliceDuration, timestampStart, timestampNow, newSession); err != nil {

				level.Error(locallogger).Log("msg", "billing failed", "err", err)
				metrics.ErrorMetrics.BillingFailure.Add(1)
			}

			return
		}

		dsRelays := rp.RelaysIn(serverCacheEntry.Datacenter)

		level.Debug(locallogger).Log("datacenter_relays", routing.RelayAddrs(dsRelays), "client_relays", routing.RelayAddrs(clientRelays))

		if len(dsRelays) == 0 {
			level.Error(locallogger).Log("msg", "no relays in datacenter", "datacenter", serverCacheEntry.Datacenter.Name)
			responseData, err := writeSessionErrorResponse(w, response, serverPrivateKey, metrics.DirectSessions, metrics.ErrorMetrics.UnserviceableUpdate, metrics.ErrorMetrics.NoRelaysInDatacenter)
			if err != nil {

				level.Error(locallogger).Log("msg", "failed to write session error response", "err", err)
				metrics.ErrorMetrics.WriteResponseFailure.Add(1)
				return
			}

			if err := updateCacheEntries(redisClientCache, sessionCacheKey, vetoCacheKey, sessionCacheEntry, vetoCacheEntry, packet, chosenRoute.Hash64(), routeDecision, timestampStart, timestampExpire,
				responseData, directStats.RTT, nnStats.RTT, sessionCacheEntry.Location); err != nil {

				level.Error(locallogger).Log("msg", "failed to update session", "err", err)
				metrics.ErrorMetrics.UpdateCacheFailure.Add(1)
			}

			if err := updatePortalData(redisClientPortal, redisClientPortalExp, packet, nnStats, directStats, chosenRoute.Relays, sessionCacheEntry.RouteDecision.OnNetworkNext, serverCacheEntry.Datacenter.Name, sessionCacheEntry.Location, timestampNow, response.Multipath); err != nil {

				level.Error(locallogger).Log("msg", "failed to update portal data", "err", err)
			}

			if err := submitBillingEntry(biller, serverCacheEntry, sessionCacheEntry, packet, response, buyer, chosenRoute, sessionCacheEntry.Location, storer, clientRelays,
				routing.Decision{OnNetworkNext: false, Reason: routing.DecisionDatacenterHasNoRelays}, sliceDuration, timestampStart, timestampNow, newSession); err != nil {

				level.Error(locallogger).Log("msg", "billing failed", "err", err)
				metrics.ErrorMetrics.BillingFailure.Add(1)
			}

			return
		}

		if routing.IsVetoed(routeDecision) {
			shouldSelect = false // Don't allow vetoed sessions to get next routes

			if vetoCacheEntry.VetoTimestamp.Before(timestampNow) {
				// Don't allow sessions vetoed with YOLO to come back on
				if routeDecision.Reason&routing.DecisionVetoYOLO == 0 {
					// Veto expired, bring the session back on with an initial slice
					routeDecision = routing.Decision{
						OnNetworkNext: false,
						Reason:        routing.DecisionInitialSlice,
					}
					newSession = true
				}
			}
		}

		if buyer.RoutingRulesSettings.Mode == routing.ModeForceDirect || int64(packet.SessionID%100) >= buyer.RoutingRulesSettings.SelectionPercentage {
			shouldSelect = false
			routeDecision = routing.Decision{
				OnNetworkNext: false,
				Reason:        routing.DecisionForceDirect,
			}

		} else if buyer.RoutingRulesSettings.Mode == routing.ModeForceNext {
			shouldDecide = false
			routeDecision = routing.Decision{
				OnNetworkNext: true,
				Reason:        routing.DecisionForceNext,
			}

			// Since the route mode is forced next, always commit to next routes
			sessionCacheEntry.CommitPending = false
			sessionCacheEntry.CommitObservedSliceCounter = 0
			sessionCacheEntry.Committed = true
		} else if buyer.RoutingRulesSettings.EnableABTest && packet.SessionID%2 == 1 {
			shouldSelect = false
			routeDecision = routing.Decision{
				OnNetworkNext: false,
				Reason:        routing.DecisionABTestDirect,
			}
		}

		if shouldSelect { // Only select a route if we should, early out for initial slice and force direct mode
			level.Debug(locallogger).Log("buyer_rtt_epsilon", buyer.RoutingRulesSettings.RTTEpsilon, "cached_route_hash", sessionCacheEntry.RouteHash)

			// hackfix: fill in client relay costs
			clientRelayCosts := make([]int, len(clientRelays))
			for i := range clientRelays {
				clientRelayCosts[i] = 10000
				for j := 0; j < int(packet.NumNearRelays); j++ {
					if packet.NearRelayIDs[j] == clientRelays[i].ID {
						clientRelayCosts[i] = int(math.Ceil(float64(packet.NearRelayMinRTT[j])))
					}
				}
			}

			// Get a set of possible routes from the RouteProvider and on error ensure sends back a direct route
			routes, err := rp.Routes(clientRelays, clientRelayCosts, dsRelays,
				routing.SelectLogger(log.With(locallogger, "step", "start")),
				routing.SelectUnencumberedRoutes(0.8),
				routing.SelectLogger(log.With(locallogger, "step", "unencumbered-routes")),
				routing.SelectAcceptableRoutesFromBestRTT(float64(buyer.RoutingRulesSettings.RTTEpsilon)),
				routing.SelectLogger(log.With(locallogger, "step", "best-rtt", "rtt-epsilon", buyer.RoutingRulesSettings.RTTEpsilon)),
				routing.SelectContainsRouteHash(sessionCacheEntry.RouteHash),
				routing.SelectLogger(log.With(locallogger, "step", "route-hash", "hash", sessionCacheEntry.RouteHash)),
				routing.SelectRoutesByRandomDestRelay(rand.NewSource(rand.Int63())),
				routing.SelectLogger(log.With(locallogger, "step", "rand-destination-relay")),
				routing.SelectRandomRoute(rand.NewSource(rand.Int63())),
				routing.SelectLogger(log.With(locallogger, "step", "rand-route")),
			)

			if err != nil {
				level.Error(locallogger).Log("err", err)
				responseData, err := writeSessionErrorResponse(w, response, serverPrivateKey, metrics.DirectSessions, metrics.ErrorMetrics.UnserviceableUpdate, metrics.ErrorMetrics.RouteFailure)
				if err != nil {

					level.Error(locallogger).Log("msg", "failed to write session error response", "err", err)
					metrics.ErrorMetrics.WriteResponseFailure.Add(1)
					return
				}

				reason := routing.DecisionNoNextRoute

				if packet.OnNetworkNext {
					// Session was on NN but now we can't find a route, probably due to relays flickering or the route being unstable
					//  If this happens, veto the session
					reason = routing.DecisionVetoNoRoute

					// YOLO reason if enabled
					if buyer.RoutingRulesSettings.EnableYouOnlyLiveOnce {
						reason |= routing.DecisionVetoYOLO
					}

					vetoCacheEntry.VetoTimestamp = timestampNow.Add(time.Hour)
					vetoCacheEntry.Reason = reason
				}

				routeDecision = routing.Decision{OnNetworkNext: false, Reason: reason}
				addRouteDecisionMetric(routeDecision, metrics)

				if err := updateCacheEntries(redisClientCache, sessionCacheKey, vetoCacheKey, sessionCacheEntry, vetoCacheEntry, packet, chosenRoute.Hash64(), routeDecision, timestampStart, timestampExpire,
					responseData, directStats.RTT, nnStats.RTT, sessionCacheEntry.Location); err != nil {

					level.Error(locallogger).Log("msg", "failed to update session", "err", err)
					metrics.ErrorMetrics.UpdateCacheFailure.Add(1)
				}

				if err := updatePortalData(redisClientPortal, redisClientPortalExp, packet, nnStats, directStats, chosenRoute.Relays, sessionCacheEntry.RouteDecision.OnNetworkNext, serverCacheEntry.Datacenter.Name, sessionCacheEntry.Location, timestampNow, response.Multipath); err != nil {

					level.Error(locallogger).Log("msg", "failed to update portal data", "err", err)
				}

				if err := submitBillingEntry(biller, serverCacheEntry, sessionCacheEntry, packet, response, buyer, chosenRoute, sessionCacheEntry.Location, storer, clientRelays,
					routeDecision, sliceDuration, timestampStart, timestampNow, newSession); err != nil {

					level.Error(locallogger).Log("msg", "billing failed", "err", err)
					metrics.ErrorMetrics.BillingFailure.Add(1)
				}

				return
			}

			if len(routes) == 0 {
				level.Error(locallogger).Log("msg", "no routes from the route matrix could be selected")
				responseData, err := writeSessionErrorResponse(w, response, serverPrivateKey, metrics.DirectSessions, metrics.ErrorMetrics.UnserviceableUpdate, metrics.ErrorMetrics.RouteFailure)
				if err != nil {

					level.Error(locallogger).Log("msg", "failed to write session error response", "err", err)
					metrics.ErrorMetrics.WriteResponseFailure.Add(1)
					return
				}

				reason := routing.DecisionNoNextRoute

				if packet.OnNetworkNext {
					// Session was on NN but now we can't find a route, probably due to relays flickering or the route being unstable
					//  If this happens, veto the session
					reason = routing.DecisionVetoNoRoute

					// YOLO reason if enabled
					if buyer.RoutingRulesSettings.EnableYouOnlyLiveOnce {
						reason |= routing.DecisionVetoYOLO
					}

					vetoCacheEntry.VetoTimestamp = timestampNow.Add(time.Hour)
					vetoCacheEntry.Reason = reason
				}

				routeDecision = routing.Decision{OnNetworkNext: false, Reason: reason}
				addRouteDecisionMetric(routeDecision, metrics)

				if err := updateCacheEntries(redisClientCache, sessionCacheKey, vetoCacheKey, sessionCacheEntry, vetoCacheEntry, packet, chosenRoute.Hash64(), routeDecision, timestampStart, timestampExpire,
					responseData, directStats.RTT, nnStats.RTT, sessionCacheEntry.Location); err != nil {

					level.Error(locallogger).Log("msg", "failed to update session", "err", err)
					metrics.ErrorMetrics.UpdateCacheFailure.Add(1)
				}

				if err := updatePortalData(redisClientPortal, redisClientPortalExp, packet, nnStats, directStats, chosenRoute.Relays, sessionCacheEntry.RouteDecision.OnNetworkNext, serverCacheEntry.Datacenter.Name, sessionCacheEntry.Location, timestampNow, response.Multipath); err != nil {

					level.Error(locallogger).Log("msg", "failed to update portal data", "err", err)
				}

				if err := submitBillingEntry(biller, serverCacheEntry, sessionCacheEntry, packet, response, buyer, chosenRoute, sessionCacheEntry.Location, storer, clientRelays,
					routeDecision, sliceDuration, timestampStart, timestampNow, newSession); err != nil {

					level.Error(locallogger).Log("msg", "billing failed", "err", err)
					metrics.ErrorMetrics.BillingFailure.Add(1)
				}

				return
			}

			nextRoute := routes[0]

			level.Debug(locallogger).Log(
				"relays", routing.RelayAddrs(nextRoute.Relays),
				"selected_next_route_stats", nextRoute.Stats.String(),
				"packet_next_stats", nnStats.String(),
				"packet_direct_stats", directStats.String(),
				"buyer_rtt_threshold", buyer.RoutingRulesSettings.RTTThreshold,
				"buyer_rtt_hysteresis", buyer.RoutingRulesSettings.RTTHysteresis,
				"buyer_rtt_veto", buyer.RoutingRulesSettings.RTTVeto,
				"buyer_packet_loss_safety", buyer.RoutingRulesSettings.EnablePacketLossSafety,
				"buyer_yolo", buyer.RoutingRulesSettings.EnableYouOnlyLiveOnce,
			)

			if shouldDecide { // Only decide on a route if we should, early out for force next mode
				deciderFuncs := []routing.DecisionFunc{
					routing.DecideUpgradeRTT(float64(buyer.RoutingRulesSettings.RTTThreshold)),
					routing.DecideDowngradeRTT(float64(buyer.RoutingRulesSettings.RTTHysteresis), buyer.RoutingRulesSettings.EnableYouOnlyLiveOnce),
					routing.DecideVeto(sessionCacheEntry.OnNNSliceCounter, float64(buyer.RoutingRulesSettings.RTTVeto), buyer.RoutingRulesSettings.EnablePacketLossSafety, buyer.RoutingRulesSettings.EnableYouOnlyLiveOnce),
					routing.DecideMultipath(buyer.RoutingRulesSettings.EnableMultipathForRTT, buyer.RoutingRulesSettings.EnableMultipathForJitter, buyer.RoutingRulesSettings.EnableMultipathForPacketLoss, float64(buyer.RoutingRulesSettings.RTTThreshold)),
				}

				if buyer.RoutingRulesSettings.EnableTryBeforeYouBuy {
					deciderFuncs = append(deciderFuncs,
						routing.DecideCommitted(sessionCacheEntry.RouteDecision.OnNetworkNext, uint8(buyer.RoutingRulesSettings.TryBeforeYouBuyMaxSlices), buyer.RoutingRulesSettings.EnableYouOnlyLiveOnce,
							&sessionCacheEntry.CommitPending, &sessionCacheEntry.CommitObservedSliceCounter, &sessionCacheEntry.Committed))
				} else {
					sessionCacheEntry.CommitPending = false
					sessionCacheEntry.CommitObservedSliceCounter = 0
					sessionCacheEntry.Committed = true
				}

				routeDecision = nextRoute.Decide(sessionCacheEntry.RouteDecision, nnStats, directStats, deciderFuncs...)

				if !routing.IsVetoed(sessionCacheEntry.RouteDecision) && routing.IsVetoed(routeDecision) {
					// Session was vetoed this update, so set the veto timeout
					vetoCacheEntry.VetoTimestamp = timestampNow.Add(time.Hour)
					vetoCacheEntry.Reason = routeDecision.Reason
				}
			}

			if sessionCacheEntry.Committed {
				// If the session is committed, set the committed flag in the response
				response.Committed = true
			}

			// If the route decision logic has decided to use multipath, then set the multipath flag in the response
			if routing.IsMultipath(routeDecision) {
				response.Multipath = true
			}

			if routeDecision.OnNetworkNext {
				// Increment the on NN slice counter if we are on NN
				sessionCacheEntry.OnNNSliceCounter++
			} else if sessionCacheEntry.RouteDecision.OnNetworkNext {
				// Reset the counter if we are going off NN this slice
				sessionCacheEntry.OnNNSliceCounter = 0
			}

			level.Debug(locallogger).Log(
				"prev_on_network_next", sessionCacheEntry.RouteDecision.OnNetworkNext,
				"prev_decision_reason", sessionCacheEntry.RouteDecision.Reason.String(),
				"on_network_next", routeDecision.OnNetworkNext,
				"decision_reason", routeDecision.Reason.String(),
				"on_NN_slice_counter", sessionCacheEntry.OnNNSliceCounter,
			)

			if routeDecision.OnNetworkNext {
				// Route decision logic decided to serve a next route

				chosenRoute = nextRoute
				var token routing.Token
				if nextRoute.Hash64() == sessionCacheEntry.RouteHash {
					token = &routing.ContinueRouteToken{
						Expires: uint64(timestampExpire.Unix()),

						SessionID: packet.SessionID,

						SessionVersion: sessionCacheEntry.Version,
						SessionFlags:   0, // Haven't figured out what this is for

						Client: routing.Client{
							Addr:      packet.ClientAddress,
							PublicKey: packet.ClientRoutePublicKey,
						},

						Server: routing.Server{
							Addr:      packet.ServerAddress,
							PublicKey: serverCacheEntry.Server.PublicKey,
						},

						Relays: nextRoute.Relays,
					}
				} else {
					sessionCacheEntry.Version++

					token = &routing.NextRouteToken{
						Expires: uint64(timestampExpire.Unix()),

						SessionID: packet.SessionID,

						SessionVersion: sessionCacheEntry.Version,
						SessionFlags:   0, // Haven't figured out what this is for

						KbpsUp:   uint32(buyer.RoutingRulesSettings.EnvelopeKbpsUp),
						KbpsDown: uint32(buyer.RoutingRulesSettings.EnvelopeKbpsDown),

						Client: routing.Client{
							Addr:      packet.ClientAddress,
							PublicKey: packet.ClientRoutePublicKey,
						},

						Server: routing.Server{
							Addr:      packet.ServerAddress,
							PublicKey: serverCacheEntry.Server.PublicKey,
						},

						Relays: nextRoute.Relays,
					}
				}

				tokens, numtokens, err := token.Encrypt(routerPrivateKey)
				if err != nil {

					level.Error(locallogger).Log("msg", "failed to encrypt route token", "err", err)
					if _, err := writeSessionErrorResponse(w, response, serverPrivateKey, metrics.DirectSessions, metrics.ErrorMetrics.UnserviceableUpdate, metrics.ErrorMetrics.EncryptionFailure); err != nil {

						level.Error(locallogger).Log("msg", "failed to write session error response", "err", err)
						metrics.ErrorMetrics.WriteResponseFailure.Add(1)
					}
					return
				}

				level.Debug(locallogger).Log("token_type", token.Type(), "current_route_hash", chosenRoute.Hash64(), "previous_route_hash", sessionCacheEntry.RouteHash)

				// Add token info to the Session Response
				response.RouteType = int32(token.Type())
				response.NumTokens = int32(numtokens) // Num of relays + client + server
				response.Tokens = tokens

				level.Debug(locallogger).Log("msg", "session served network next route")
			}
		}

		addRouteDecisionMetric(routeDecision, metrics)

		// Send the Session Response back to the server
		var responseData []byte
		if responseData, err = writeSessionResponse(w, response, serverPrivateKey); err != nil {
			level.Error(locallogger).Log("msg", "failed to write session response", "err", err)
			metrics.ErrorMetrics.WriteResponseFailure.Add(1)
			return
		}

		// If we managed to send the response, update metrics based on route type
		if response.RouteType == routing.RouteTypeDirect {
			metrics.DirectSessions.Add(1)
		} else {
			metrics.NextSessions.Add(1)
		}

		// Cache the needed information for the next session update
		level.Debug(locallogger).Log("msg", "caching session data")
		if err := updateCacheEntries(redisClientCache, sessionCacheKey, vetoCacheKey, sessionCacheEntry, vetoCacheEntry, packet, chosenRoute.Hash64(), routeDecision, timestampStart, timestampExpire,
			responseData, directStats.RTT, nnStats.RTT, sessionCacheEntry.Location); err != nil {
			level.Error(locallogger).Log("msg", "failed to update cache entries", "err", err)
			metrics.ErrorMetrics.UpdateCacheFailure.Add(1)
		}

		// Set portal data
		if err := updatePortalData(redisClientPortal, redisClientPortalExp, packet, nnStats, directStats, chosenRoute.Relays, sessionCacheEntry.RouteDecision.OnNetworkNext, serverCacheEntry.Datacenter.Name, sessionCacheEntry.Location, timestampNow, response.Multipath); err != nil {
			level.Error(locallogger).Log("msg", "failed to update portal data", "err", err)
		}

		// Submit a new billing entry
		if err := submitBillingEntry(biller, serverCacheEntry, sessionCacheEntry, packet, response, buyer, chosenRoute, sessionCacheEntry.Location, storer, clientRelays,
			routeDecision, sliceDuration, timestampStart, timestampNow, newSession); err != nil {
			level.Error(locallogger).Log("msg", "billing failed", "err", err)
			metrics.ErrorMetrics.BillingFailure.Add(1)
		}
	}
}
*/

func updatePortalData(redisClientPortal redis.Cmdable, redisClientPortalExp time.Duration, packet *SessionUpdatePacket, lastNNStats *routing.Stats, lastDirectStats *routing.Stats,
	relayHops []routing.Relay, onNetworkNext bool, datacenterName string, location *routing.Location, nearRelays []routing.Relay, sessionTime time.Time, isMultiPath bool, datacenterAlias string) error {

	if (lastNNStats.RTT == 0 && lastDirectStats.RTT == 0) || (onNetworkNext && lastNNStats.RTT == 0) {
		return nil
	}

	var hashedID string
	if !packet.Version.IsInternal() && packet.Version.Compare(SDKVersion{3, 4, 5}) == SDKVersionOlder {
		hash := fnv.New64a()
		byteArray := make([]byte, 8)
		binary.LittleEndian.PutUint64(byteArray, packet.UserHash)
		hash.Write(byteArray)
		hashedID = fmt.Sprintf("%016x", hash.Sum64())
	} else {
		hashedID = fmt.Sprintf("%016x", packet.UserHash)
	}

	var deltaRTT float64
	if !onNetworkNext {
		deltaRTT = 0
	} else {
		deltaRTT = lastDirectStats.RTT - lastNNStats.RTT
	}

	meta := routing.SessionMeta{
		ID:              fmt.Sprintf("%016x", packet.SessionID),
		UserHash:        hashedID,
		DatacenterName:  datacenterName,
		DatacenterAlias: datacenterAlias,
		OnNetworkNext:   onNetworkNext,
		NextRTT:         lastNNStats.RTT,
		DirectRTT:       lastDirectStats.RTT,
		DeltaRTT:        deltaRTT,
		Location:        *location,
		ClientAddr:      packet.ClientAddress.String(),
		ServerAddr:      packet.ServerAddress.String(),
		Hops:            relayHops,
		SDK:             packet.Version.String(),
		Connection:      ConnectionTypeText(packet.ConnectionType),
		NearbyRelays:    nearRelays,
		Platform:        PlatformTypeText(packet.PlatformID),
		BuyerID:         fmt.Sprintf("%016x", packet.CustomerID),
	}

	slice := routing.SessionSlice{
		Timestamp: sessionTime,
		Next:      *lastNNStats,
		Direct:    *lastDirectStats,
		Envelope: routing.Envelope{
			Up:   int64(packet.KbpsUp),
			Down: int64(packet.KbpsDown),
		},
		IsMultiPath:       isMultiPath,
		IsTryBeforeYouBuy: packet.TryBeforeYouBuy || !packet.Committed,
		OnNetworkNext:     onNetworkNext,
	}
	point := routing.SessionMapPoint{
		Latitude:      location.Latitude,
		Longitude:     location.Longitude,
		OnNetworkNext: onNetworkNext,
	}

	tx := redisClientPortal.TxPipeline()

	// set total session counts with expiration on the entire key set for safety
	switch meta.OnNetworkNext {
	case true:
		// Remove the session from the direct set if it exists
		tx.ZRem("total-direct", meta.ID)
		tx.ZRem(fmt.Sprintf("total-direct-buyer-%016x", packet.CustomerID), meta.ID)

		tx.ZAdd("total-next", &redis.Z{Score: meta.DeltaRTT, Member: meta.ID})
		tx.Expire("total-next", redisClientPortalExp)
		tx.ZAdd(fmt.Sprintf("total-next-buyer-%016x", packet.CustomerID), &redis.Z{Score: meta.DeltaRTT, Member: meta.ID})
		tx.Expire(fmt.Sprintf("total-next-buyer-%016x", packet.CustomerID), redisClientPortalExp)
	case false:
		// Remove the session from the next set if it exists
		tx.ZRem("total-next", meta.ID)
		tx.ZRem(fmt.Sprintf("total-next-buyer-%016x", packet.CustomerID), meta.ID)

		tx.ZAdd("total-direct", &redis.Z{Score: -meta.DirectRTT, Member: meta.ID})
		tx.Expire("total-direct", redisClientPortalExp)
		tx.ZAdd(fmt.Sprintf("total-direct-buyer-%016x", packet.CustomerID), &redis.Z{Score: -meta.DirectRTT, Member: meta.ID})
		tx.Expire(fmt.Sprintf("total-direct-buyer-%016x", packet.CustomerID), redisClientPortalExp)
	}

	// set session and slice information with expiration on the entire key set for safety
	tx.Set(fmt.Sprintf("session-%016x-meta", packet.SessionID), meta, redisClientPortalExp)
	tx.SAdd(fmt.Sprintf("session-%016x-slices", packet.SessionID), slice)
	tx.Expire(fmt.Sprintf("session-%016x-slices", packet.SessionID), redisClientPortalExp)

	// set the user session reverse lookup sets with expiration on the entire key set for safety
	tx.SAdd(fmt.Sprintf("user-%s-sessions", hashedID), meta.ID)
	tx.Expire(fmt.Sprintf("user-%s-sessions", hashedID), redisClientPortalExp)

	// set the map point key and buyer sessions with expiration on the entire key set for safety
	tx.Set(fmt.Sprintf("session-%016x-point", packet.SessionID), point, redisClientPortalExp)
	tx.SAdd(fmt.Sprintf("map-points-%016x-buyer", packet.CustomerID), meta.ID)
	tx.Expire(fmt.Sprintf("map-points-%016x-buyer", packet.CustomerID), redisClientPortalExp)

	if _, err := tx.Exec(); err != nil {
		return err
	}

	return nil
}

/*
func submitBillingEntry(biller billing.Biller, serverCacheEntry *ServerCacheEntry, prevRouteHash uint64, request *SessionUpdatePacket, response *SessionResponsePacket,
	buyer *routing.Buyer, chosenRoute *routing.Route, location *routing.Location, storer storage.Storer, clientRelays []routing.Relay, routeDecision routing.Decision,
	sliceDuration uint64, timestampStart time.Time, timestampNow time.Time, newSession bool) error {

	sameRoute := chosenRoute.Hash64() == prevRouteHash
	routeRequest := NewRouteRequest(request, buyer, serverCacheEntry, location, storer, clientRelays)
	billingEntry := NewBillingEntry(routeRequest, chosenRoute, int(response.RouteType), sameRoute, &buyer.RoutingRulesSettings, routeDecision, request, sliceDuration, timestampStart, timestampNow, newSession)
	return biller.Bill(context.Background(), request.SessionID, billingEntry)
}
*/

// todo: disabled
/*
func updateCacheEntries(redisClient redis.Cmdable, sessionCacheKey string, vetoCacheKey string, sessionCacheEntry *SessionCacheEntry, vetoCacheEntry *VetoCacheEntry, packet *SessionUpdatePacket, chosenRouteHash uint64,
	routeDecision routing.Decision, timestampStart time.Time, timestampExpire time.Time, responseData []byte, directRTT float64, nextRTT float64, location *routing.Location) error {
	updatedSessionCacheEntry := SessionCacheEntry{
		CustomerID:                 packet.CustomerID,
		SessionID:                  packet.SessionID,
		UserHash:                   packet.UserHash,
		Sequence:                   packet.Sequence,
		RouteHash:                  chosenRouteHash,
		RouteDecision:              routeDecision,
		OnNNSliceCounter:           sessionCacheEntry.OnNNSliceCounter,
		CommitPending:              sessionCacheEntry.CommitPending,
		CommitObservedSliceCounter: sessionCacheEntry.CommitObservedSliceCounter,
		Committed:                  sessionCacheEntry.Committed,
		TimestampStart:             timestampStart,
		TimestampExpire:            timestampExpire,
		Version:                    sessionCacheEntry.Version, //This was already incremented for the route tokens
		Response:                   responseData,
		DirectRTT:                  directRTT,
		NextRTT:                    nextRTT,
		Location:                   *location,
	}

	updatedVetoCacheEntry := VetoCacheEntry{
		VetoTimestamp: vetoCacheEntry.VetoTimestamp,
		Reason:        vetoCacheEntry.Reason,
	}

	tx := redisClient.TxPipeline()
	{
		tx.Set(sessionCacheKey, updatedSessionCacheEntry, 5*time.Minute)
		tx.Set(vetoCacheKey, updatedVetoCacheEntry, 1*time.Hour)

		if _, err := tx.Exec(); err != nil {
			return fmt.Errorf("failed to execute update cache tx pipeline: %v", err)
		}
	}

	return nil
}
*/

func addRouteDecisionMetric(d routing.Decision, m *metrics.SessionMetrics) {
	switch d.Reason {
	case routing.DecisionNoReason:
		m.DecisionMetrics.NoReason.Add(1)
	case routing.DecisionForceDirect:
		m.DecisionMetrics.ForceDirect.Add(1)
	case routing.DecisionForceNext:
		m.DecisionMetrics.ForceNext.Add(1)
	case routing.DecisionNoNextRoute:
		m.DecisionMetrics.NoNextRoute.Add(1)
	case routing.DecisionABTestDirect:
		m.DecisionMetrics.ABTestDirect.Add(1)
	case routing.DecisionRTTReduction:
		m.DecisionMetrics.RTTReduction.Add(1)
	case routing.DecisionHighPacketLossMultipath:
		m.DecisionMetrics.PacketLossMultipath.Add(1)
	case routing.DecisionHighJitterMultipath:
		m.DecisionMetrics.JitterMultipath.Add(1)
	case routing.DecisionVetoRTT:
		m.DecisionMetrics.VetoRTT.Add(1)
	case routing.DecisionRTTReductionMultipath:
		m.DecisionMetrics.RTTMultipath.Add(1)
	case routing.DecisionVetoPacketLoss:
		m.DecisionMetrics.VetoPacketLoss.Add(1)
	case routing.DecisionFallbackToDirect:
		m.DecisionMetrics.FallbackToDirect.Add(1)
	case routing.DecisionVetoYOLO:
		m.DecisionMetrics.VetoYOLO.Add(1)
	case routing.DecisionInitialSlice:
		m.DecisionMetrics.InitialSlice.Add(1)
	case routing.DecisionVetoRTT | routing.DecisionVetoYOLO:
		m.DecisionMetrics.VetoRTTYOLO.Add(1)
	case routing.DecisionVetoPacketLoss | routing.DecisionVetoYOLO:
		m.DecisionMetrics.VetoPacketLossYOLO.Add(1)
	case routing.DecisionRTTHysteresis:
		m.DecisionMetrics.RTTHysteresis.Add(1)
	case routing.DecisionVetoCommit:
		m.DecisionMetrics.VetoCommit.Add(1)
	}
}

func writeInitResponse(w io.Writer, response ServerInitResponsePacket, privateKey []byte) error {
	// Sign the response
	response.Signature = crypto.Sign(privateKey, response.GetSignData())

	// Marshal the packet
	responseData, err := response.MarshalBinary()
	if err != nil {
		return err
	}

	// Send the Session Response back to the server
	if _, err := w.Write(responseData); err != nil {
		return err
	}

	return nil
}

func writeSessionResponse(w io.Writer, response *SessionResponsePacket, privateKey []byte) ([]byte, error) {
	// Sign the response
	response.Signature = crypto.Sign(privateKey, response.GetSignData())

	// Marshal the packet
	responseData, err := response.MarshalBinary()
	if err != nil {
		return nil, err
	}

	// Send the Session Response back to the server
	if _, err := w.Write(responseData); err != nil {
		return nil, err
	}

	return responseData, nil
}

func sendRouteResponse(w io.Writer, route *routing.Route, params *SessionUpdateParams, packet *SessionUpdatePacket, response *SessionResponsePacket, serverDataReadOnly *ServerData,
	buyer *routing.Buyer, lastNextStats *routing.Stats, lastDirectStats *routing.Stats, location *routing.Location, nearRelays []routing.Relay, routeDecision routing.Decision, vetoReason routing.DecisionReason,
	onNNSliceCounter uint64, committedData routing.CommittedData, prevRouteHash uint64, prevOnNetworkNext bool, timeNow time.Time, routeExpireTimestamp int64, tokenVersion uint8, sliceDuration uint64, routerPrivateKey []byte) {
	// Update response data
	{
		if committedData.Committed {
			response.Committed = true
		}

		if routing.IsMultipath(routeDecision) {
			response.Multipath = true
		}
	}

	// Tokenize the route
	if routeDecision.OnNetworkNext {
		var token routing.Token
		if route.Hash64() == prevRouteHash {
			token = &routing.ContinueRouteToken{
				Expires: uint64(routeExpireTimestamp),

				SessionID: packet.SessionID,

				SessionVersion: tokenVersion,

				Client: routing.Client{
					Addr:      packet.ClientAddress,
					PublicKey: packet.ClientRoutePublicKey,
				},

				Server: routing.Server{
					Addr:      packet.ServerAddress,
					PublicKey: serverDataReadOnly.routePublicKey,
				},

				Relays: route.Relays,
			}
		} else {
			tokenVersion++

			token = &routing.NextRouteToken{
				Expires: uint64(routeExpireTimestamp),

				SessionID: packet.SessionID,

				SessionVersion: tokenVersion,

				KbpsUp:   uint32(buyer.RoutingRulesSettings.EnvelopeKbpsUp),
				KbpsDown: uint32(buyer.RoutingRulesSettings.EnvelopeKbpsDown),

				Client: routing.Client{
					Addr:      packet.ClientAddress,
					PublicKey: packet.ClientRoutePublicKey,
				},

				Server: routing.Server{
					Addr:      packet.ServerAddress,
					PublicKey: serverDataReadOnly.routePublicKey,
				},

				Relays: route.Relays,
			}
		}

		tokens, numtokens, err := token.Encrypt(routerPrivateKey)
		if err != nil {
			params.Metrics.ErrorMetrics.UnserviceableUpdate.Add(1)
			params.Metrics.ErrorMetrics.EncryptionFailure.Add(1)
			return
		}

		// Add token info to the Session Response
		response.RouteType = int32(token.Type())
		response.NumTokens = int32(numtokens) // Num of relays + client + server
		response.Tokens = tokens
	}

	// Update the session data
	session := SessionData{
		timestamp:            timeNow.Unix(),
		location:             *location,
		sequence:             packet.Sequence,
		nearRelays:           nearRelays,
		routeHash:            route.Hash64(),
		routeDecision:        routeDecision,
		onNNSliceCounter:     onNNSliceCounter,
		committedData:        committedData,
		routeExpireTimestamp: routeExpireTimestamp,
		tokenVersion:         tokenVersion,
	}
	sessionMutexStart := time.Now()
	params.SessionMap.UpdateSessionData(packet.SessionID, &session)
	if time.Since(sessionMutexStart).Seconds() > 0.1 {
		level.Debug(params.Logger).Log("msg", "long session mutex in send route response")
	}

	// If the session was vetoed this slice, update the veto data
	if routing.IsVetoed(routeDecision) && vetoReason == routing.DecisionNoReason {
		params.VetoMap.SetVeto(packet.SessionID, routeDecision.Reason)
	}

	addRouteDecisionMetric(routeDecision, params.Metrics)

	// The envelope values are averages, so need to multiply by slice duration
	envelopeBytesUp := ((1000 * uint64(buyer.RoutingRulesSettings.EnvelopeKbpsUp)) / 8) * sliceDuration
	envelopeBytesDown := ((1000 * uint64(buyer.RoutingRulesSettings.EnvelopeKbpsDown)) / 8) * sliceDuration

	// IMPORTANT: run post in parallel so it doesn't block the response
	go PostSessionUpdate(params, packet, response, serverDataReadOnly, route, lastNextStats, lastDirectStats, location, nearRelays, routeDecision, timeNow, envelopeBytesUp, envelopeBytesDown)

	if _, err := writeSessionResponse(w, response, params.ServerPrivateKey); err != nil {
		// level.Error(params.Logger).Log("msg", "could not write session update response packet", "err", err)
		params.Metrics.ErrorMetrics.WriteResponseFailure.Add(1)
		return
	}
}
