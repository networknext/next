package transport

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"runtime"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	jsoniter "github.com/json-iterator/go"

	"github.com/go-redis/redis/v7"
	"github.com/networknext/backend/billing"
	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/metrics"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
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
	data := make([]byte, m.MaxPacketSize)

	for {
		numbytes, addr, _ := m.Conn.ReadFromUDP(data)
		if numbytes <= 0 {
			continue
		}

		packet := UDPPacket{SourceAddr: addr, Data: data[:numbytes]}

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
			m.Conn.WriteToUDP(buf.Bytes(), packet.SourceAddr)
		}
	}
}

func ServerInitHandlerFunc(logger log.Logger, storer storage.Storer, metrics *metrics.ServerInitMetrics, serverPrivateKey []byte) UDPHandlerFunc {
	logger = log.With(logger, "handler", "init")

	return func(w io.Writer, incoming *UDPPacket) {
		durationStart := time.Now()
		defer func() {
			durationSince := time.Since(durationStart)
			level.Info(logger).Log("duration", durationSince.Milliseconds())
			metrics.Invocations.Add(1)
		}()

		var packet ServerInitRequestPacket
		if err := packet.UnmarshalBinary(incoming.Data); err != nil {
			sentry.CaptureException(err)
			level.Error(logger).Log("msg", "could not read packet", "err", err)
			metrics.ErrorMetrics.UnmarshalFailure.Add(1)
			return
		}

		locallogger := log.With(
			logger,
			"src_addr", incoming.SourceAddr.String(),
			"request_id", packet.RequestID,
			"customer_id", packet.CustomerID,
			"datacenter_id", packet.DatacenterID,
			"sdk", packet.Version.String(),
		)

		response := ServerInitResponsePacket{
			RequestID: packet.RequestID,
			Response:  InitResponseOK,
		}
		defer func() {
			if err := writeInitResponse(w, response, serverPrivateKey); err != nil {
				level.Error(locallogger).Log("msg", "failed to write init response", "err", err)
			}
		}()

		if !incoming.SourceAddr.IP.IsLoopback() && !packet.Version.AtLeast(SDKVersionMin) {
			sentry.CaptureMessage("sdk version is too old")
			level.Error(locallogger).Log("msg", "sdk version is too old")
			response.Response = InitResponseOldSDKVersion
			metrics.ErrorMetrics.SDKTooOld.Add(1)
			return
		}

		_, err := storer.Datacenter(packet.DatacenterID)
		if err != nil {
			sentry.CaptureException(err)
			level.Error(locallogger).Log("msg", "failed to get datacenter from storage", "err", err)
			response.Response = InitResponseUnknownDatacenter
			metrics.ErrorMetrics.DatacenterNotFound.Add(1)
			return
		}

		buyer, err := storer.Buyer(packet.CustomerID)
		if err != nil {
			sentry.CaptureException(err)
			level.Error(locallogger).Log("msg", "failed to get buyer from storage", "err", err)
			response.Response = InitResponseUnknownCustomer
			metrics.ErrorMetrics.BuyerNotFound.Add(1)
			return
		}

		if !crypto.Verify(buyer.PublicKey, packet.GetSignData(), packet.Signature) {
			sentry.CaptureMessage("signature verification failed")
			level.Error(locallogger).Log("msg", "signature verification failed")
			response.Response = InitResponseSignatureCheckFailed
			metrics.ErrorMetrics.VerificationFailure.Add(1)
			return
		}
	}
}

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

// ServerUpdateHandlerFunc ...
func ServerUpdateHandlerFunc(logger log.Logger, redisClient redis.Cmdable, storer storage.Storer, metrics *metrics.ServerUpdateMetrics) UDPHandlerFunc {
	logger = log.With(logger, "handler", "server")

	return func(w io.Writer, incoming *UDPPacket) {
		durationStart := time.Now()
		defer func() {
			durationSince := time.Since(durationStart)
			level.Info(logger).Log("duration", durationSince.Milliseconds())
			metrics.Invocations.Add(1)
		}()

		var packet ServerUpdatePacket
		if err := packet.UnmarshalBinary(incoming.Data); err != nil {
			sentry.CaptureException(err)
			level.Error(logger).Log("msg", "could not read packet", "err", err)
			metrics.ErrorMetrics.UnmarshalFailure.Add(1)
			return
		}

		serverCacheKey := fmt.Sprintf("SERVER-%d-%s", packet.CustomerID, packet.ServerAddress.String())

		locallogger := log.With(logger, "src_addr", incoming.SourceAddr.String(), "server_addr", packet.ServerAddress.String())

		// Drop the packet if version is older that the minimun sdk version
		if !incoming.SourceAddr.IP.IsLoopback() && !packet.Version.AtLeast(SDKVersionMin) {
			sentry.CaptureMessage(fmt.Sprintf("sdk version is too old: %s", packet.Version.String()))
			level.Error(locallogger).Log("msg", "sdk version is too old", "sdk", packet.Version.String())
			metrics.ErrorMetrics.SDKTooOld.Add(1)
			return
		}

		locallogger = log.With(locallogger, "sdk", packet.Version.String())

		// Get the buyer information for the id in the packet
		buyer, err := storer.Buyer(packet.CustomerID)
		if err != nil {
			sentry.CaptureException(err)
			level.Error(locallogger).Log("msg", "failed to get buyer from storage", "err", err, "customer_id", packet.CustomerID)
			metrics.ErrorMetrics.BuyerNotFound.Add(1)
			return
		}

		datacenter, err := storer.Datacenter(packet.DatacenterID)
		if err != nil {
			level.Error(locallogger).Log("msg", "failed to get datacenter from storage", "err", err, "customer_id", packet.CustomerID)
			metrics.ErrorMetrics.DatacenterNotFound.Add(1)
			return
		}

		locallogger = log.With(locallogger, "customer_id", packet.CustomerID)

		// Drop the packet if the signed packet data cannot be verified with the buyers public key
		if !crypto.Verify(buyer.PublicKey, packet.GetSignData(), packet.Signature) {
			sentry.CaptureMessage("signature verification failed")
			level.Error(locallogger).Log("msg", "signature verification failed")
			metrics.ErrorMetrics.VerificationFailure.Add(1)
			return
		}

		// Get the the old ServerCacheEntry if it exists, otherwise serverentry is in zero value state
		var serverentry ServerCacheEntry
		{
			result := redisClient.Get(serverCacheKey)
			if result.Err() != nil && result.Err() != redis.Nil {
				level.Error(locallogger).Log("msg", "failed to get server", "err", result.Err())
				return
			}
			serverdata, err := result.Bytes()
			if err != nil && result.Err() != redis.Nil {
				level.Error(locallogger).Log("msg", "failed to get server bytes", "err", err)
				return
			}
			if serverdata != nil {
				if err := serverentry.UnmarshalBinary(serverdata); err != nil {
					level.Error(locallogger).Log("msg", "failed to unmarshal server bytes", "err", err)
				}
			}
		}

		// Drop the packet if the sequence number is older than the previously cached sequence number
		if packet.Sequence < serverentry.Sequence {
			sentry.CaptureMessage(fmt.Sprintf("packet too old: seq = %d | latest sequence = %d", packet.Sequence, serverentry.Sequence))
			level.Error(locallogger).Log("msg", "packet too old", "packet sequence", packet.Sequence, "lastest sequence", serverentry.Sequence)
			metrics.ErrorMetrics.PacketSequenceTooOld.Add(1)
			return
		}

		// Save some of the packet information to be used in SessionUpdateHandlerFunc
		serverentry = ServerCacheEntry{
			Sequence:   packet.Sequence,
			Server:     routing.Server{Addr: packet.ServerPrivateAddress, PublicKey: packet.ServerRoutePublicKey},
			Datacenter: datacenter,
			SDKVersion: packet.Version,
		}
		result := redisClient.Set(serverCacheKey, serverentry, 5*time.Minute)
		if result.Err() != nil {
			level.Error(locallogger).Log("msg", "failed to update server", "err", result.Err())
			return
		}

		level.Debug(locallogger).Log("msg", "updated server")
	}
}

type SessionCacheEntry struct {
	CustomerID                 uint64
	SessionID                  uint64
	UserHash                   uint64
	Sequence                   uint64
	RouteHash                  uint64
	RouteDecision              routing.Decision
	CommitPending              bool
	CommitObservedSliceCounter uint8
	Committed                  bool
	TimestampStart             time.Time
	TimestampExpire            time.Time
	VetoTimestamp              time.Time
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

type RouteProvider interface {
	ResolveRelay(uint64) (routing.Relay, error)
	RelaysIn(routing.Datacenter) []routing.Relay
	Routes([]routing.Relay, []routing.Relay, ...routing.SelectorFunc) ([]routing.Route, error)
}

// SessionUpdateHandlerFunc ...
func SessionUpdateHandlerFunc(logger log.Logger, redisClientCache redis.Cmdable, redisClientPortal redis.Cmdable, storer storage.Storer, rp RouteProvider, iploc routing.IPLocator, geoClient *routing.GeoClient, metrics *metrics.SessionMetrics, biller billing.Biller, serverPrivateKey []byte, routerPrivateKey []byte) UDPHandlerFunc {
	logger = log.With(logger, "handler", "session")

	return func(w io.Writer, incoming *UDPPacket) {
		durationStart := time.Now()
		defer func() {
			durationSince := time.Since(durationStart)
			level.Info(logger).Log("duration", durationSince.Milliseconds())
			metrics.Invocations.Add(1)
		}()

		timestampNow := time.Now()

		// Whether or not we should make a route selection/decision on a network next route, or serve a direct route
		shouldSelect := true
		shouldDecide := true

		// Flag to check if this session is a new session
		newSession := false

		// Deserialize the Session packet
		var packet SessionUpdatePacket
		if err := packet.UnmarshalBinary(incoming.Data); err != nil {
			sentry.CaptureException(err)
			level.Error(logger).Log("msg", "could not read packet", "err", err)
			metrics.ErrorMetrics.ReadPacketFailure.Add(1)
			return
		}

		serverCacheKey := fmt.Sprintf("SERVER-%d-%s", packet.CustomerID, packet.ServerAddress.String())
		sessionCacheKey := fmt.Sprintf("SESSION-%d-%d", packet.CustomerID, packet.SessionID)

		locallogger := log.With(logger, "src_addr", incoming.SourceAddr.String(), "server_addr", packet.ServerAddress.String(), "client_addr", packet.ClientAddress.String(), "session_id", packet.SessionID)

		var serverCacheEntry ServerCacheEntry
		var sessionCacheEntry SessionCacheEntry

		// Start building session response packet, defaulting to a direct route
		response := SessionResponsePacket{
			Sequence:  packet.Sequence,
			SessionID: packet.SessionID,
			RouteType: int32(routing.RouteTypeDirect),
		}

		// Build a redis transaction to make a single network call
		tx := redisClientCache.TxPipeline()
		{
			serverCacheCmd := tx.Get(serverCacheKey)
			sessionCacheCmd := tx.Get(sessionCacheKey)
			if _, err := tx.Exec(); err != nil && err != redis.Nil {
				sentry.CaptureException(err)
				level.Error(locallogger).Log("msg", "failed to execute redis pipeline", "err", err)
				metrics.ErrorMetrics.PipelineExecFailure.Add(1)
				return
			}

			// Note that if we fail to retrieve the server data, we don't bother responding since server will ignore response without ServerRoutePublicKey set
			// See next_server_internal_process_packet in next.cpp for full requirements of response packet
			serverCacheData, err := serverCacheCmd.Bytes()
			if err != nil {
				sentry.CaptureException(err)
				level.Error(locallogger).Log("msg", "failed to get server bytes", "err", err)
				metrics.ErrorMetrics.GetServerDataFailure.Add(1)
				return
			}
			if err := serverCacheEntry.UnmarshalBinary(serverCacheData); err != nil {
				sentry.CaptureException(err)
				level.Error(locallogger).Log("msg", "failed to unmarshal server bytes", "err", err)
				metrics.ErrorMetrics.UnmarshalServerDataFailure.Add(1)
				return
			}

			// Set public key on response as soon as we get it
			response.ServerRoutePublicKey = serverCacheEntry.Server.PublicKey

			if sessionCacheCmd.Err() != redis.Nil {
				sessionCacheData, err := sessionCacheCmd.Bytes()
				if err != nil {
					// This error case should never happen, can't produce it in test cases, but leaving it in anyway
					sentry.CaptureException(err)
					level.Error(locallogger).Log("msg", "failed to get session bytes", "err", err)
					writeSessionErrorResponse(w, response, serverPrivateKey, metrics.DirectSessions, metrics.ErrorMetrics.WriteResponseFailure, metrics.ErrorMetrics.GetSessionDataFailure)
					return
				}

				if len(sessionCacheData) != 0 {
					if err := sessionCacheEntry.UnmarshalBinary(sessionCacheData); err != nil {
						sentry.CaptureException(err)
						level.Error(locallogger).Log("msg", "failed to unmarshal session bytes", "err", err)
						writeSessionErrorResponse(w, response, serverPrivateKey, metrics.DirectSessions, metrics.ErrorMetrics.WriteResponseFailure, metrics.ErrorMetrics.UnmarshalSessionDataFailure)
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
		}

		locallogger = log.With(locallogger, "datacenter_id", serverCacheEntry.Datacenter.ID)

		buyer, err := storer.Buyer(packet.CustomerID)
		if err != nil {
			sentry.CaptureException(err)
			level.Error(locallogger).Log("msg", "failed to get buyer from storage", "err", err, "customer_id", packet.CustomerID)
			writeSessionErrorResponse(w, response, serverPrivateKey, metrics.DirectSessions, metrics.ErrorMetrics.WriteResponseFailure, metrics.ErrorMetrics.BuyerNotFound)
			return
		}

		locallogger = log.With(locallogger, "customer_id", packet.CustomerID)

		if !crypto.Verify(buyer.PublicKey, packet.GetSignData(), packet.Signature) {
			sentry.CaptureException(err)
			err := errors.New("failed to verify packet signature with buyer public key")
			level.Error(locallogger).Log("err", err)
			writeSessionErrorResponse(w, response, serverPrivateKey, metrics.DirectSessions, metrics.ErrorMetrics.WriteResponseFailure, metrics.ErrorMetrics.VerifyFailure)
			return
		}

		switch seq := packet.Sequence; {
		case seq < sessionCacheEntry.Sequence:
			err := fmt.Errorf("packet sequence too old. current_sequence %v, previous sequence %v", packet.Sequence, sessionCacheEntry.Sequence)
			sentry.CaptureException(err)
			level.Error(locallogger).Log("err", err)
			writeSessionErrorResponse(w, response, serverPrivateKey, metrics.DirectSessions, metrics.ErrorMetrics.WriteResponseFailure, metrics.ErrorMetrics.OldSequence)
			return
		case seq == sessionCacheEntry.Sequence:
			if _, err := w.Write(sessionCacheEntry.Response); err != nil {
				sentry.CaptureException(err)
				level.Error(locallogger).Log("err", err)
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
			Stats: directStats,
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
			level.Error(logger).Log("err", "fallback to direct")
			writeSessionErrorResponse(w, response, serverPrivateKey, metrics.DirectSessions, metrics.ErrorMetrics.WriteResponseFailure, metrics.ErrorMetrics.FallbackToDirect)

			if err := updatePortalData(redisClientPortal, packet, nnStats, directStats, len(chosenRoute.Relays), routeDecision.OnNetworkNext, serverCacheEntry.Datacenter.Name, routing.LocationNullIsland); err != nil {
				sentry.CaptureException(err)
				level.Error(locallogger).Log("msg", "failed to update portal data", "err", err)
			}

			if err := submitBillingEntry(biller, serverCacheEntry, sessionCacheEntry, packet, response, buyer, chosenRoute, routing.LocationNullIsland, storer, nil,
				routing.Decision{OnNetworkNext: false, Reason: routing.DecisionFallbackToDirect}, sliceDuration, timestampStart, timestampNow, newSession); err != nil {
				sentry.CaptureException(err)
				level.Error(locallogger).Log("msg", "billing failed", "err", err)
				metrics.ErrorMetrics.BillingFailure.Add(1)
			}

			return
		}

		// Get relays near the client
		location, err := iploc.LocateIP(packet.ClientAddress.IP)
		if err != nil {
			sentry.CaptureException(err)
			level.Error(locallogger).Log("msg", "failed to locate client", "err", err)
			writeSessionErrorResponse(w, response, serverPrivateKey, metrics.DirectSessions, metrics.ErrorMetrics.WriteResponseFailure, metrics.ErrorMetrics.ClientLocateFailure)

			if err := updatePortalData(redisClientPortal, packet, nnStats, directStats, len(chosenRoute.Relays), routeDecision.OnNetworkNext, serverCacheEntry.Datacenter.Name, location); err != nil {
				sentry.CaptureException(err)
				level.Error(locallogger).Log("msg", "failed to update portal data", "err", err)
			}

			if err := submitBillingEntry(biller, serverCacheEntry, sessionCacheEntry, packet, response, buyer, chosenRoute, location, storer, nil,
				routing.Decision{OnNetworkNext: false, Reason: routing.DecisionNoNearRelays}, sliceDuration, timestampStart, timestampNow, newSession); err != nil {
				sentry.CaptureException(err)
				level.Error(locallogger).Log("msg", "billing failed", "err", err)
				metrics.ErrorMetrics.BillingFailure.Add(1)
			}

			return
		}
		level.Debug(locallogger).Log("client_ip", packet.ClientAddress.IP.String(), "lat", location.Latitude, "long", location.Longitude)

		clientRelays, err := geoClient.RelaysWithin(location.Latitude, location.Longitude, 500, "mi")

		if len(clientRelays) == 0 || err != nil {
			sentry.CaptureException(err)
			level.Error(locallogger).Log("msg", "failed to locate relays near client", "err", err)
			writeSessionErrorResponse(w, response, serverPrivateKey, metrics.DirectSessions, metrics.ErrorMetrics.WriteResponseFailure, metrics.ErrorMetrics.NearRelaysLocateFailure)

			if err := updatePortalData(redisClientPortal, packet, nnStats, directStats, len(chosenRoute.Relays), routeDecision.OnNetworkNext, serverCacheEntry.Datacenter.Name, location); err != nil {
				sentry.CaptureException(err)
				level.Error(locallogger).Log("msg", "failed to update portal data", "err", err)
			}

			if err := submitBillingEntry(biller, serverCacheEntry, sessionCacheEntry, packet, response, buyer, chosenRoute, location, storer, clientRelays,
				routing.Decision{OnNetworkNext: false, Reason: routing.DecisionNoNearRelays}, sliceDuration, timestampStart, timestampNow, newSession); err != nil {
				sentry.CaptureException(err)
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

		dsRelays := rp.RelaysIn(serverCacheEntry.Datacenter)

		level.Debug(locallogger).Log("num_datacenter_relays", len(dsRelays), "num_client_relays", len(clientRelays))

		if len(dsRelays) == 0 {
			sentry.CaptureMessage(fmt.Sprintf("No relays in datacenter %s", serverCacheEntry.Datacenter.Name))
			level.Error(locallogger).Log("msg", "no relays in datacenter", "datacenter", serverCacheEntry.Datacenter.Name)
			writeSessionErrorResponse(w, response, serverPrivateKey, metrics.DirectSessions, metrics.ErrorMetrics.WriteResponseFailure, metrics.ErrorMetrics.NoRelaysInDatacenter)

			if err := updatePortalData(redisClientPortal, packet, nnStats, directStats, len(chosenRoute.Relays), routeDecision.OnNetworkNext, serverCacheEntry.Datacenter.Name, location); err != nil {
				sentry.CaptureException(err)
				level.Error(locallogger).Log("msg", "failed to update portal data", "err", err)
			}

			if err := submitBillingEntry(biller, serverCacheEntry, sessionCacheEntry, packet, response, buyer, chosenRoute, location, storer, clientRelays,
				routing.Decision{OnNetworkNext: false, Reason: routing.DecisionDatacenterHasNoRelays}, sliceDuration, timestampStart, timestampNow, newSession); err != nil {
				sentry.CaptureException(err)
				level.Error(locallogger).Log("msg", "billing failed", "err", err)
				metrics.ErrorMetrics.BillingFailure.Add(1)
			}

			return
		}

		if routing.IsVetoed(routeDecision) && sessionCacheEntry.VetoTimestamp.Before(timestampNow) {
			shouldSelect = false

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

		if buyer.RoutingRulesSettings.Mode == routing.ModeForceDirect || int64(packet.SessionID%100) > buyer.RoutingRulesSettings.SelectionPercentage {
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
		} else if buyer.RoutingRulesSettings.EnableABTest && packet.SessionID%2 == 1 {
			shouldSelect = false
			routeDecision = routing.Decision{
				OnNetworkNext: false,
				Reason:        routing.DecisionABTestDirect,
			}
		}

		if shouldSelect { // Only select a route if we should, early out for initial slice and force direct mode
			level.Debug(locallogger).Log("buyer_rtt_epsilon", buyer.RoutingRulesSettings.RTTEpsilon, "cached_route_hash", sessionCacheEntry.RouteHash)
			// Get a set of possible routes from the RouteProvider and on error ensure it falls back to direct
			routes, err := rp.Routes(dsRelays, clientRelays,
				routing.SelectAcceptableRoutesFromBestRTT(float64(buyer.RoutingRulesSettings.RTTEpsilon)),
				routing.SelectContainsRouteHash(sessionCacheEntry.RouteHash),
				routing.SelectRoutesByRandomDestRelay(rand.NewSource(rand.Int63())),
				routing.SelectRandomRoute(rand.NewSource(rand.Int63())))
			if err != nil {
				level.Error(locallogger).Log("err", err)
				writeSessionErrorResponse(w, response, serverPrivateKey, metrics.DirectSessions, metrics.ErrorMetrics.WriteResponseFailure, metrics.ErrorMetrics.RouteFailure)

				if err := updatePortalData(redisClientPortal, packet, nnStats, directStats, len(chosenRoute.Relays), routeDecision.OnNetworkNext, serverCacheEntry.Datacenter.Name, location); err != nil {
					sentry.CaptureException(err)
					level.Error(locallogger).Log("msg", "failed to update portal data", "err", err)
				}

				if err := submitBillingEntry(biller, serverCacheEntry, sessionCacheEntry, packet, response, buyer, chosenRoute, location, storer, clientRelays,
					routing.Decision{OnNetworkNext: false, Reason: routing.DecisionNoNextRoute}, sliceDuration, timestampStart, timestampNow, newSession); err != nil {
					sentry.CaptureException(err)
					level.Error(locallogger).Log("msg", "billing failed", "err", err)
					metrics.ErrorMetrics.BillingFailure.Add(1)
				}

				return
			}

			nextRoute := routes[0]

			level.Debug(locallogger).Log(
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
					routing.DecideVeto(float64(buyer.RoutingRulesSettings.RTTVeto), buyer.RoutingRulesSettings.EnablePacketLossSafety, buyer.RoutingRulesSettings.EnableYouOnlyLiveOnce),
					routing.DecideMultipath(buyer.RoutingRulesSettings.EnableMultipathForRTT, buyer.RoutingRulesSettings.EnableMultipathForJitter, buyer.RoutingRulesSettings.EnableMultipathForPacketLoss, float64(buyer.RoutingRulesSettings.RTTThreshold)),
				}

				if buyer.RoutingRulesSettings.EnableTryBeforeYouBuy {
					deciderFuncs = append(deciderFuncs,
						routing.DecideCommitted(sessionCacheEntry.RouteDecision.OnNetworkNext, uint8(buyer.RoutingRulesSettings.TryBeforeYouBuyMaxSlices), buyer.RoutingRulesSettings.EnableYouOnlyLiveOnce,
							&sessionCacheEntry.CommitPending, &sessionCacheEntry.CommitObservedSliceCounter, &sessionCacheEntry.Committed))
				}

				routeDecision = nextRoute.Decide(sessionCacheEntry.RouteDecision, nnStats, directStats, deciderFuncs...)

				if routing.IsVetoed(routeDecision) {
					// Session was vetoed this update, so set the veto timeout
					sessionCacheEntry.VetoTimestamp = timestampNow.Add(time.Hour)
				} else if sessionCacheEntry.Committed {
					// If the session is committed, set the committed flag in the response
					response.Committed = true
				}

				// If the route decision logic has decided to use multipath, then set the multipath flag in the response
				if routing.IsMultipath(routeDecision) {
					response.Multipath = true
				}
			}

			level.Debug(locallogger).Log(
				"prev_on_network_next", sessionCacheEntry.RouteDecision.OnNetworkNext,
				"prev_decision_reason", sessionCacheEntry.RouteDecision.Reason.String(),
				"on_network_next", routeDecision.OnNetworkNext,
				"decision_reason", routeDecision.Reason.String(),
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
					sentry.CaptureException(err)
					level.Error(locallogger).Log("msg", "failed to encrypt route token", "err", err)
					writeSessionErrorResponse(w, response, serverPrivateKey, metrics.DirectSessions, metrics.ErrorMetrics.WriteResponseFailure, metrics.ErrorMetrics.EncryptionFailure)
					return
				}

				level.Debug(locallogger).Log("token_type", token.Type(), "current_route_hash", chosenRoute.Hash64(), "previous_route_hash", sessionCacheEntry.RouteHash)

				// Add token info to the Session Response
				response.RouteType = int32(token.Type())
				response.NumTokens = int32(numtokens) // Num of relays + client + server
				response.Tokens = tokens

				// Fill in the near relays
				response.NumNearRelays = int32(len(clientRelays))
				response.NearRelayIDs = make([]uint64, len(clientRelays))
				response.NearRelayAddresses = make([]net.UDPAddr, len(clientRelays))
				for idx, relay := range clientRelays {
					response.NearRelayIDs[idx] = relay.ID
					response.NearRelayAddresses[idx] = relay.Addr
				}

				level.Debug(locallogger).Log("msg", "session served network next route")
			}
		}

		addRouteDecisionMetric(routeDecision, metrics)

		// Send the Session Response back to the server
		var responseData []byte
		if responseData, err = writeSessionResponse(w, response, serverPrivateKey); err != nil {
			sentry.CaptureException(err)
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
		{
			level.Debug(locallogger).Log("msg", "caching session data")
			updatedSessionCacheEntry := SessionCacheEntry{
				CustomerID:                 packet.CustomerID,
				SessionID:                  packet.SessionID,
				UserHash:                   packet.UserHash,
				Sequence:                   packet.Sequence,
				RouteHash:                  chosenRoute.Hash64(),
				RouteDecision:              routeDecision,
				CommitPending:              sessionCacheEntry.CommitPending,
				CommitObservedSliceCounter: sessionCacheEntry.CommitObservedSliceCounter,
				Committed:                  sessionCacheEntry.Committed,
				TimestampStart:             timestampStart,
				TimestampExpire:            timestampExpire,
				VetoTimestamp:              sessionCacheEntry.VetoTimestamp,
				Version:                    sessionCacheEntry.Version, //This was already incremented for the route tokens
				Response:                   responseData,
				DirectRTT:                  directStats.RTT,
				NextRTT:                    nnStats.RTT,
				Location:                   location,
			}
			result := redisClientCache.Set(sessionCacheKey, updatedSessionCacheEntry, 5*time.Minute)
			if result.Err() != nil {
				// This error case should never happen, can't produce it in test cases, but leaving it in anyway
				sentry.CaptureException(err)
				level.Error(locallogger).Log("msg", "failed to update session", "err", err)
				metrics.ErrorMetrics.UpdateSessionFailure.Add(1)
			}
		}

		// Set portal data
		if err := updatePortalData(redisClientPortal, packet, nnStats, directStats, len(chosenRoute.Relays), routeDecision.OnNetworkNext, serverCacheEntry.Datacenter.Name, location); err != nil {
			sentry.CaptureException(err)
			level.Error(locallogger).Log("msg", "failed to update portal data", "err", err)
		}

		// Submit a new billing entry
		if err := submitBillingEntry(biller, serverCacheEntry, sessionCacheEntry, packet, response, buyer, chosenRoute, location, storer, clientRelays,
			routeDecision, sliceDuration, timestampStart, timestampNow, newSession); err != nil {
			sentry.CaptureException(err)
			level.Error(locallogger).Log("msg", "billing failed", "err", err)
			metrics.ErrorMetrics.BillingFailure.Add(1)
		}
	}
}

func updatePortalData(redisClientPortal redis.Cmdable, packet SessionUpdatePacket, nnStats routing.Stats, directStats routing.Stats, relayHops int, onNetworkNext bool, datacenterName string, location routing.Location) error {
	meta := routing.SessionMeta{
		ID:         fmt.Sprintf("%x", packet.SessionID),
		UserHash:   fmt.Sprintf("%x", packet.UserHash),
		Datacenter: datacenterName,
		NextRTT:    nnStats.RTT,
		DirectRTT:  directStats.RTT,
		DeltaRTT:   directStats.RTT - nnStats.RTT,
		Location:   location,
		ClientAddr: packet.ClientAddress.String(),
		ServerAddr: packet.ServerAddress.String(),
		Hops:       relayHops,
		SDK:        packet.Version.String(),
		Connection: ConnectionTypeText(packet.ConnectionType),
	}
	// Only fill in the essential information here to then let the poral fill in additional relay info
	// so we don't spend time fetching info from storage here
	for idx := 0; idx < int(packet.NumNearRelays); idx++ {
		meta.NearbyRelays = append(meta.NearbyRelays, routing.Relay{
			ID: packet.NearRelayIDs[idx],
			ClientStats: routing.Stats{
				RTT:        float64(packet.NearRelayMeanRTT[idx]),
				Jitter:     float64(packet.NearRelayJitter[idx]),
				PacketLoss: float64(packet.NearRelayPacketLoss[idx]),
			},
		})
	}
	slice := routing.SessionSlice{
		Timestamp: time.Now(),
		Next:      nnStats,
		Direct:    directStats,
		Envelope: routing.Envelope{
			Up:   int64(packet.KbpsUp),
			Down: int64(packet.KbpsDown),
		},
	}
	point := routing.SessionMapPoint{
		Latitude:      location.Latitude,
		Longitude:     location.Longitude,
		OnNetworkNext: onNetworkNext,
	}

	tx := redisClientPortal.TxPipeline()
	tx.ZAdd("top-global", &redis.Z{Score: meta.DeltaRTT, Member: meta.ID})
	tx.ZAdd(fmt.Sprintf("top-buyer-%x", packet.CustomerID), &redis.Z{Score: meta.DeltaRTT, Member: meta.ID})
	tx.Set(fmt.Sprintf("session-%x-meta", packet.SessionID), meta, 720*time.Hour)
	tx.SAdd(fmt.Sprintf("session-%x-slices", packet.SessionID), slice)
	tx.Expire(fmt.Sprintf("session-%x-slices", packet.SessionID), 720*time.Hour)
	tx.SAdd(fmt.Sprintf("user-%x-sessions", packet.UserHash), meta.ID)
	tx.Expire(fmt.Sprintf("user-%x-sessions", packet.UserHash), 720*time.Hour)
	tx.SAdd("map-points-global", meta.ID)
	tx.SAdd(fmt.Sprintf("map-points-buyer-%x", packet.CustomerID), meta.ID)
	tx.Expire(fmt.Sprintf("map-points-buyer-%x", packet.CustomerID), 720*time.Hour)
	tx.Set(fmt.Sprintf("session-%x-point", packet.SessionID), point, 720*time.Hour)
	if _, err := tx.Exec(); err != nil {
		return err
	}

	return nil
}

func submitBillingEntry(biller billing.Biller, serverCacheEntry ServerCacheEntry, sessionCacheEntry SessionCacheEntry, request SessionUpdatePacket, response SessionResponsePacket,
	buyer routing.Buyer, chosenRoute routing.Route, location routing.Location, storer storage.Storer, clientRelays []routing.Relay, routeDecision routing.Decision,
	sliceDuration uint64, timestampStart time.Time, timestampNow time.Time, newSession bool) error {

	sameRoute := chosenRoute.Hash64() == sessionCacheEntry.RouteHash
	routeRequest := NewRouteRequest(request, &buyer, serverCacheEntry, location, storer, clientRelays)
	billingEntry := NewBillingEntry(routeRequest, &chosenRoute, int(response.RouteType), sameRoute, &buyer.RoutingRulesSettings, routeDecision, &request, sliceDuration, timestampStart, timestampNow, newSession)
	return biller.Bill(context.Background(), request.SessionID, billingEntry)
}

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

// writeInitResponse encrypts the server init response packet and sends it back to the server. Returns the marshaled response and an error.
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

// writeSessionResponse encrypts the session response packet and sends it back to the server. Returns the marshaled response and an error.
func writeSessionResponse(w io.Writer, response SessionResponsePacket, privateKey []byte) ([]byte, error) {
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

func writeSessionErrorResponse(w io.Writer, response SessionResponsePacket, privateKey []byte, directSessions metrics.Counter, writeResponseFailure metrics.Counter, errCounter metrics.Counter) {
	if _, err := writeSessionResponse(w, response, privateKey); err != nil {
		writeResponseFailure.Add(1)
		return
	}

	directSessions.Add(1)
	errCounter.Add(1)
}
