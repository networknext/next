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
			level.Error(logger).Log("msg", "could not read packet", "err", err)
			return
		}

		serverCacheKey := fmt.Sprintf("SERVER-%d-%s", packet.CustomerID, packet.ServerAddress.String())

		locallogger := log.With(logger, "src_addr", incoming.SourceAddr.String(), "server_addr", packet.ServerAddress.String())

		// Drop the packet if version is older that the minimun sdk version
		if !incoming.SourceAddr.IP.IsLoopback() && !packet.Version.AtLeast(SDKVersionMin) {
			level.Error(locallogger).Log("msg", "sdk version is too old", "sdk", packet.Version.String())
			return
		}

		locallogger = log.With(locallogger, "sdk", packet.Version.String())

		// Get the buyer information for the id in the packet
		buyer, ok := storer.Buyer(packet.CustomerID)
		if !ok {
			level.Error(locallogger).Log("msg", "failed to get buyer", "customer_id", packet.CustomerID)
			return
		}

		locallogger = log.With(locallogger, "customer_id", packet.CustomerID)

		// Drop the packet if the signed packet data cannot be verified with the buyers public key
		if !crypto.Verify(buyer.PublicKey, packet.GetSignData(), packet.Signature) {
			level.Error(locallogger).Log("msg", "signature verification failed")
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

		// Drop the packet if the sequence number is older than the previously cache sequence number
		// if packet.Sequence < serverentry.Sequence {
		// 	log.Printf("packet too old: (packet) %d < %d (Redis)", packet.Sequence, serverentry.Sequence)
		// 	return
		// }

		// Save some of the packet information to be used in SessionUpdateHandlerFunc
		serverentry = ServerCacheEntry{
			Sequence:   packet.Sequence,
			Server:     routing.Server{Addr: packet.ServerPrivateAddress, PublicKey: packet.ServerRoutePublicKey},
			Datacenter: routing.Datacenter{ID: packet.DatacenterID},
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
	SessionID       uint64
	Sequence        uint64
	RouteHash       uint64
	RouteDecision   routing.Decision
	TimestampStart  time.Time
	TimestampExpire time.Time
	VetoTimestamp   time.Time
	Version         uint8
	Response        []byte
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
func SessionUpdateHandlerFunc(logger log.Logger, redisClient redis.Cmdable, storer storage.Storer, rp RouteProvider, iploc routing.IPLocator, geoClient *routing.GeoClient, metrics *metrics.SessionMetrics, biller billing.Biller, serverPrivateKey []byte, routerPrivateKey []byte) UDPHandlerFunc {
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
			level.Error(logger).Log("msg", "could not read packet", "err", err)
			metrics.SessionErrorMetrics.ReadPacketFailure.Add(1)
			return
		}

		serverCacheKey := fmt.Sprintf("SERVER-%d-%s", packet.CustomerID, packet.ServerAddress.String())
		sessionCacheKey := fmt.Sprintf("SESSION-%d-%d", packet.CustomerID, packet.SessionID)

		locallogger := log.With(logger, "src_addr", incoming.SourceAddr.String(), "server_addr", packet.ServerAddress.String(), "client_addr", packet.ClientAddress.String(), "session_id", packet.SessionID)

		if packet.FallbackToDirect {
			level.Error(logger).Log("err", "fallback to direct")
			metrics.SessionErrorMetrics.FallbackToDirect.Add(1)
			return
		}

		var serverCacheEntry ServerCacheEntry
		var sessionCacheEntry SessionCacheEntry

		// Start building session response packet, defaulting to a direct route
		response := SessionResponsePacket{
			Sequence:  packet.Sequence,
			SessionID: packet.SessionID,
			RouteType: int32(routing.RouteTypeDirect),
		}

		// Build a redis transaction to make a single network call
		tx := redisClient.TxPipeline()
		{
			serverCacheCmd := tx.Get(serverCacheKey)
			sessionCacheCmd := tx.Get(sessionCacheKey)
			if _, err := tx.Exec(); err != nil && err != redis.Nil {
				level.Error(locallogger).Log("msg", "failed to execute redis pipeline", "err", err)
				metrics.SessionErrorMetrics.PipelineExecFailure.Add(1)
				return
			}

			// Note that if we fail to retrieve the server data, we don't bother responding since server will ignore response without ServerRoutePublicKey set
			// See next_server_internal_process_packet in next.cpp for full requirements of response packet
			serverCacheData, err := serverCacheCmd.Bytes()
			if err != nil {
				level.Error(locallogger).Log("msg", "failed to get server bytes", "err", err)
				metrics.SessionErrorMetrics.GetServerDataFailure.Add(1)
				return
			}
			if err := serverCacheEntry.UnmarshalBinary(serverCacheData); err != nil {
				level.Error(locallogger).Log("msg", "failed to unmarshal server bytes", "err", err)
				metrics.SessionErrorMetrics.UnmarshalServerDataFailure.Add(1)
				return
			}

			// Set public key on response as soon as we get it
			response.ServerRoutePublicKey = serverCacheEntry.Server.PublicKey

			if sessionCacheCmd.Err() != redis.Nil {
				sessionCacheData, err := sessionCacheCmd.Bytes()
				if err != nil {
					// This error case should never happen, can't produce it in test cases, but leaving it in anyway
					level.Error(locallogger).Log("msg", "failed to get session bytes", "err", err)
					writeSessionErrorResponse(w, response, serverPrivateKey, metrics.DirectSessions, metrics.SessionErrorMetrics.WriteResponseFailure, metrics.SessionErrorMetrics.GetSessionDataFailure)
					return
				}

				if len(sessionCacheData) != 0 {
					if err := sessionCacheEntry.UnmarshalBinary(sessionCacheData); err != nil {
						level.Error(locallogger).Log("msg", "failed to unmarshal session bytes", "err", err)
						writeSessionErrorResponse(w, response, serverPrivateKey, metrics.DirectSessions, metrics.SessionErrorMetrics.WriteResponseFailure, metrics.SessionErrorMetrics.UnmarshalSessionDataFailure)
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

		buyer, ok := storer.Buyer(packet.CustomerID)
		if !ok {
			err := fmt.Errorf("failed to get buyer with customer ID %v", packet.CustomerID)
			level.Error(locallogger).Log("err", err)
			writeSessionErrorResponse(w, response, serverPrivateKey, metrics.DirectSessions, metrics.SessionErrorMetrics.WriteResponseFailure, metrics.SessionErrorMetrics.BuyerNotFound)
			return
		}

		locallogger = log.With(locallogger, "customer_id", packet.CustomerID)

		if !crypto.Verify(buyer.PublicKey, packet.GetSignData(), packet.Signature) {
			err := errors.New("failed to verify packet signature with buyer public key")
			level.Error(locallogger).Log("err", err)
			writeSessionErrorResponse(w, response, serverPrivateKey, metrics.DirectSessions, metrics.SessionErrorMetrics.WriteResponseFailure, metrics.SessionErrorMetrics.VerifyFailure)
			return
		}

		switch seq := packet.Sequence; {
		case seq < sessionCacheEntry.Sequence:
			err := fmt.Errorf("packet sequence too old. current_sequence %v, previous sequence %v", packet.Sequence, sessionCacheEntry.Sequence)
			level.Error(locallogger).Log("err", err)
			writeSessionErrorResponse(w, response, serverPrivateKey, metrics.DirectSessions, metrics.SessionErrorMetrics.WriteResponseFailure, metrics.SessionErrorMetrics.OldSequence)
			return
		case seq == sessionCacheEntry.Sequence:
			if _, err := w.Write(sessionCacheEntry.Response); err != nil {
				level.Error(locallogger).Log("err", err)
				metrics.SessionErrorMetrics.WriteCachedResponseFailure.Add(1)
			}
			return
		}

		location, err := iploc.LocateIP(packet.ClientAddress.IP)
		if err != nil {
			level.Error(locallogger).Log("msg", "failed to locate client", "err", err)
			writeSessionErrorResponse(w, response, serverPrivateKey, metrics.DirectSessions, metrics.SessionErrorMetrics.WriteResponseFailure, metrics.SessionErrorMetrics.ClientLocateFailure)
			return
		}
		level.Debug(locallogger).Log("client_ip", packet.ClientAddress.IP.String(), "lat", location.Latitude, "long", location.Longitude)

		clientrelays, err := geoClient.RelaysWithin(location.Latitude, location.Longitude, 500, "mi")

		if len(clientrelays) == 0 || err != nil {
			level.Error(locallogger).Log("msg", "failed to locate relays near client", "err", err)
			writeSessionErrorResponse(w, response, serverPrivateKey, metrics.DirectSessions, metrics.SessionErrorMetrics.WriteResponseFailure, metrics.SessionErrorMetrics.NearRelaysLocateFailure)
			return
		}

		// Clamp relay count to max
		if len(clientrelays) > int(MaxNearRelays) {
			clientrelays = clientrelays[:MaxNearRelays]
		}

		// We need to do this because RelaysWithin only has the ID of the relay and we need the Addr and PublicKey too
		// Maybe we consider a nicer way to do this in the future
		for idx := range clientrelays {
			clientrelays[idx], _ = rp.ResolveRelay(clientrelays[idx].ID)
		}

		dsrelays := rp.RelaysIn(serverCacheEntry.Datacenter)

		level.Debug(locallogger).Log("num_datacenter_relays", len(dsrelays), "num_client_relays", len(clientrelays))

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

		if routing.IsVetoed(routeDecision) && sessionCacheEntry.VetoTimestamp.Before(timestampNow) {
			// Veto expired, bring the session back on with an initial slice
			routeDecision = routing.Decision{
				OnNetworkNext: false,
				Reason:        routing.DecisionInitialSlice,
			}
			shouldSelect = false
			newSession = true
		}

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

		if buyer.RoutingRulesSettings.Mode == routing.ModeForceDirect {
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
		}

		if shouldSelect { // Only select a route if we should, early out for initial slice and force direct mode
			level.Debug(locallogger).Log("buyer_rtt_epsilon", buyer.RoutingRulesSettings.RTTEpsilon, "cached_route_hash", sessionCacheEntry.RouteHash)
			// Get a set of possible routes from the RouteProvider and on error ensure it falls back to direct
			routes, err := rp.Routes(dsrelays, clientrelays,
				routing.SelectAcceptableRoutesFromBestRTT(float64(buyer.RoutingRulesSettings.RTTEpsilon)),
				routing.SelectContainsRouteHash(sessionCacheEntry.RouteHash),
				routing.SelectRoutesByRandomDestRelay(rand.NewSource(rand.Int63())),
				routing.SelectRandomRoute(rand.NewSource(rand.Int63())))
			if err != nil {
				level.Error(locallogger).Log("err", err)
				writeSessionErrorResponse(w, response, serverPrivateKey, metrics.DirectSessions, metrics.SessionErrorMetrics.WriteResponseFailure, metrics.SessionErrorMetrics.RouteFailure)
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
				routeDecision = nextRoute.Decide(sessionCacheEntry.RouteDecision, nnStats, directStats,
					routing.DecideUpgradeRTT(float64(buyer.RoutingRulesSettings.RTTThreshold)),
					routing.DecideDowngradeRTT(float64(buyer.RoutingRulesSettings.RTTHysteresis), buyer.RoutingRulesSettings.EnableYouOnlyLiveOnce),
					routing.DecideVeto(float64(buyer.RoutingRulesSettings.RTTVeto), buyer.RoutingRulesSettings.EnablePacketLossSafety, buyer.RoutingRulesSettings.EnableYouOnlyLiveOnce),
				)

				if routing.IsVetoed(routeDecision) {
					// Session was vetoed this update, so set the veto timeout
					sessionCacheEntry.VetoTimestamp = timestampNow.Add(time.Hour)
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
					level.Error(locallogger).Log("msg", "failed to encrypt route token", "err", err)
					writeSessionErrorResponse(w, response, serverPrivateKey, metrics.DirectSessions, metrics.SessionErrorMetrics.WriteResponseFailure, metrics.SessionErrorMetrics.EncryptionFailure)
					return
				}

				level.Debug(locallogger).Log("token_type", token.Type(), "current_route_hash", chosenRoute.Hash64(), "previous_route_hash", sessionCacheEntry.RouteHash)

				// Add token info to the Session Response
				response.RouteType = int32(token.Type())
				response.NumTokens = int32(numtokens) // Num of relays + client + server
				response.Tokens = tokens

				// Fill in the near relays
				response.NumNearRelays = int32(len(clientrelays))
				response.NearRelayIDs = make([]uint64, len(clientrelays))
				response.NearRelayAddresses = make([]net.UDPAddr, len(clientrelays))
				for idx, relay := range clientrelays {
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
			level.Error(locallogger).Log("msg", "failed to write session response", "err", err)
			metrics.SessionErrorMetrics.WriteResponseFailure.Add(1)
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
				SessionID:       packet.SessionID,
				Sequence:        packet.Sequence,
				RouteHash:       chosenRoute.Hash64(),
				RouteDecision:   routeDecision,
				TimestampStart:  timestampStart,
				TimestampExpire: timestampExpire,
				VetoTimestamp:   sessionCacheEntry.VetoTimestamp,
				Version:         sessionCacheEntry.Version, //This was already incremented for the route tokens
				Response:        responseData,
			}
			result := redisClient.Set(sessionCacheKey, updatedSessionCacheEntry, 5*time.Minute)
			if result.Err() != nil {
				// This error case should never happen, can't produce it in test cases, but leaving it in anyway
				level.Error(locallogger).Log("msg", "failed to update session", "err", err)
				metrics.SessionErrorMetrics.UpdateSessionFailure.Add(1)
			}
		}

		// Submit a new billing entry
		{
			sameRoute := chosenRoute.Hash64() == sessionCacheEntry.RouteHash
			routeRequest := NewRouteRequest(packet, buyer, serverCacheEntry, location, storer, clientrelays)
			billingEntry := NewBillingEntry(routeRequest, &chosenRoute, int(response.RouteType), sameRoute, &buyer.RoutingRulesSettings, routeDecision, &packet, sliceDuration, timestampStart, timestampNow, newSession)
			if err := biller.Bill(context.Background(), packet.SessionID, billingEntry); err != nil {
				level.Error(locallogger).Log("msg", "billing failed", "err", err)
				metrics.SessionErrorMetrics.BillingFailure.Add(1)
			}
		}
	}
}

func addRouteDecisionMetric(d routing.Decision, m *metrics.SessionMetrics) {
	switch d.Reason {
	case routing.DecisionNoChange:
		m.DecisionMetrics.NoChange.Add(1)
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
	case routing.DecisionPacketLossMultipath:
		m.DecisionMetrics.PacketLossMultipath.Add(1)
	case routing.DecisionJitterMultipath:
		m.DecisionMetrics.JitterMultipath.Add(1)
	case routing.DecisionVetoRTT:
		m.DecisionMetrics.VetoRTT.Add(1)
	case routing.DecisionRTTMultipath:
		m.DecisionMetrics.RTTMultipath.Add(1)
	case routing.DecisionVetoPacketLoss:
		m.DecisionMetrics.VetoPacketLoss.Add(1)
	case routing.DecisionFallbackToDirect:
		m.DecisionMetrics.FallbackToDirect.Add(1)
	case routing.DecisionVetoYOLO:
		m.DecisionMetrics.VetoYOLO.Add(1)
	case routing.DecisionVetoNoRoute:
		m.DecisionMetrics.VetoNoRoute.Add(1)
	case routing.DecisionInitialSlice:
		m.DecisionMetrics.InitialSlice.Add(1)
	case routing.DecisionVetoRTT | routing.DecisionVetoYOLO:
		m.DecisionMetrics.VetoRTTYOLO.Add(1)
	case routing.DecisionVetoPacketLoss | routing.DecisionVetoYOLO:
		m.DecisionMetrics.VetoPacketLossYOLO.Add(1)
	case routing.DecisionRTTIncrease:
		m.DecisionMetrics.RTTIncrease.Add(1)
	}
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
