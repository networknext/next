package transport

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"time"
	"math/rand"

	gkmetrics "github.com/go-kit/kit/metrics"

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
		return errors.New("relay server cannot be nil")
	}

	data := make([]byte, m.MaxPacketSize)
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			numbytes, addr, _ := m.Conn.ReadFromUDP(data)
			if numbytes <= 0 {
				continue
			}

			packet := UDPPacket{
				SourceAddr: addr,
				Data:       data[:numbytes],
			}

			go m.processPacket(ctx, packet)
		}
	}
}

func (m *UDPServerMux) processPacket(ctx context.Context, packet UDPPacket) {
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
func ServerUpdateHandlerFunc(logger log.Logger, redisClient redis.Cmdable, storer storage.Storer, duration gkmetrics.Histogram, counter gkmetrics.Counter, metricsHandler metrics.Handler) UDPHandlerFunc {
	logger = log.With(logger, "handler", "server")

	return func(w io.Writer, incoming *UDPPacket) {
		timer := gkmetrics.NewTimer(duration.With("method", "ServerUpdateHandlerFunc"))
		timer.Unit(time.Millisecond)
		defer func() {
			timer.ObserveDuration()
		}()

		var packet ServerUpdatePacket
		if err := packet.UnmarshalBinary(incoming.Data); err != nil {
			level.Error(logger).Log("msg", "could not read packet", "err", err)
			return
		}

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
			result := redisClient.Get("SERVER-" + incoming.SourceAddr.String())
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
		result := redisClient.Set("SERVER-"+incoming.SourceAddr.String(), serverentry, 5*time.Minute)
		if result.Err() != nil {
			level.Error(locallogger).Log("msg", "failed to update server", "err", result.Err())
			return
		}

		level.Debug(locallogger).Log("msg", "updated server")
		counter.Add(1)
	}
}

type SessionCacheEntry struct {
	SessionID       uint64
	Sequence        uint64
	RouteHash       uint64
	TimestampStart  time.Time
	TimestampExpire time.Time
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
func SessionUpdateHandlerFunc(logger log.Logger, redisClient redis.Cmdable, storer storage.Storer, duration gkmetrics.Histogram, counter gkmetrics.Counter, rp RouteProvider, iploc routing.IPLocator, geoClient *routing.GeoClient, metricsHandler metrics.Handler, biller billing.Biller, serverPrivateKey []byte, routerPrivateKey []byte) UDPHandlerFunc {
	logger = log.With(logger, "handler", "session")

	return func(w io.Writer, incoming *UDPPacket) {
		timer := gkmetrics.NewTimer(duration.With("method", "SessionUpdateHandlerFunc"))
		timer.Unit(time.Millisecond)
		defer func() {
			timer.ObserveDuration()
		}()

		timestampNow := time.Now()

		// Deserialize the Session packet
		var packet SessionUpdatePacket
		if err := packet.UnmarshalBinary(incoming.Data); err != nil {
			level.Error(logger).Log("msg", "could not read packet", "err", err)
			return // TODO: direct here?
		}

		locallogger := log.With(logger, "src_addr", incoming.SourceAddr.String(), "server_addr", packet.ServerAddress.String(), "client_addr", packet.ClientAddress.String(), "session_id", packet.SessionID)

		var serverCacheEntry ServerCacheEntry
		sessionCacheEntry := SessionCacheEntry{
			TimestampStart: timestampNow,
		}

		// Start building session response packet, defaulting to a direct route
		response := SessionResponsePacket{
			Sequence:  packet.Sequence,
			SessionID: packet.SessionID,
			RouteType: int32(routing.RouteTypeDirect),
		}

		// Build a redis transaction to make a single network call
		tx := redisClient.TxPipeline()
		{
			serverCacheCmd := tx.Get("SERVER-" + incoming.SourceAddr.String())
			sessionCacheCmd := tx.Get(fmt.Sprintf("SESSION-%d", packet.SessionID))
			tx.Exec()

			// Note that if we fail to retrieve the server data, we don't bother responding since server will ignore response without ServerRoutePublicKey set
			// See next_server_internal_process_packet in next.cpp for full requirements of response packet
			serverCacheData, err := serverCacheCmd.Bytes()
			if err != nil {
				level.Error(locallogger).Log("msg", "failed to get server bytes", "err", err)
				return
			}
			if err := serverCacheEntry.UnmarshalBinary(serverCacheData); err != nil {
				level.Error(locallogger).Log("msg", "failed to unmarshal server bytes", "err", err)
				return
			}

			// Set public key on response as soon as we get it
			response.ServerRoutePublicKey = serverCacheEntry.Server.PublicKey

			if sessionCacheCmd.Err() != redis.Nil {
				sessionCacheData, err := sessionCacheCmd.Bytes()
				if err != nil {
					level.Error(locallogger).Log("msg", "failed to get session bytes", "err", err)
					handleError(w, response, serverPrivateKey, err)
					return
				}

				if sessionCacheData == nil || len(sessionCacheData) == 0 {

				} else if err := sessionCacheEntry.UnmarshalBinary(sessionCacheData); err != nil {
					level.Error(locallogger).Log("msg", "failed to unmarshal session bytes", "err", err)
					handleError(w, response, serverPrivateKey, err)
					return
				}
			}
		}

		locallogger = log.With(locallogger, "datacenter_id", serverCacheEntry.Datacenter.ID)

		buyer, ok := storer.Buyer(packet.CustomerID)
		if !ok {
			err := fmt.Errorf("failed to get buyer with customer ID %v", packet.CustomerID)
			level.Error(locallogger).Log("err", err)
			handleError(w, response, serverPrivateKey, err)
			return
		}

		locallogger = log.With(locallogger, "customer_id", packet.CustomerID)

		if !crypto.Verify(buyer.PublicKey, packet.GetSignData(), packet.Signature) {
			err := errors.New("failed to verify packet signature with buyer public key")
			level.Error(locallogger).Log("err", err)
			handleError(w, response, serverPrivateKey, err)
			return
		}

		switch seq := packet.Sequence; {
		case seq < sessionCacheEntry.Sequence:
			err := fmt.Errorf("packet sequence too old. current_sequence %v, previous sequence %v", packet.Sequence, sessionCacheEntry.Sequence)
			level.Error(locallogger).Log("err", err)
			handleError(w, response, serverPrivateKey, err)
			return
		case seq == sessionCacheEntry.Sequence:
			if _, err := w.Write(sessionCacheEntry.Response); err != nil {
				level.Error(locallogger).Log("err", err)
				handleError(w, response, serverPrivateKey, err)
			}
			return
		}

		location, err := iploc.LocateIP(packet.ClientAddress.IP)
		if err != nil {
			level.Error(locallogger).Log("msg", "failed to locate client", "err", err)
			handleError(w, response, serverPrivateKey, err)
			return
		}
		level.Debug(locallogger).Log("lat", location.Latitude, "long", location.Longitude)

		clientrelays, err := geoClient.RelaysWithin(location.Latitude, location.Longitude, 500, "mi")
		if len(clientrelays) == 0 || err != nil {
			level.Error(locallogger).Log("msg", "failed to locate relays near client", "err", err)
			handleError(w, response, serverPrivateKey, err)
			return
		}

		// We need to do this because RelaysWithin only has the ID of the relay and we need the Addr and PublicKey too
		// Maybe we consider a nicer way to do this in the future
		for idx := range clientrelays {
			clientrelays[idx], _ = rp.ResolveRelay(clientrelays[idx].ID)
		}

		dsrelays := rp.RelaysIn(serverCacheEntry.Datacenter)

		level.Debug(locallogger).Log("num_datacenter_relays", len(dsrelays), "num_client_relays", len(clientrelays))

		// Get a set of possible routes from the RouteProvider and on error ensure it falls back to direct
		routes, err := rp.Routes(dsrelays, clientrelays,
			routing.SelectAcceptableRoutesFromBestRTT(float64(buyer.RoutingRulesSettings.RTTEpsilon)),
			routing.SelectContainsRouteHash(sessionCacheEntry.RouteHash),
			routing.SelectRoutesByRandomDestRelay(rand.NewSource(rand.Int63())),
			routing.SelectRandomRoute(rand.NewSource(rand.Int63())))
		if err != nil {
			level.Error(locallogger).Log("err", err)
			handleError(w, response, serverPrivateKey, err)
			return
		}

		if len(routes) <= 0 {
			level.Warn(locallogger).Log("msg", "no acceptable route available")
		}

		nextRoute := routes[0]

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

		// If this is the initial slice, always serve a direct route first
		routeDecision := routing.Decision{
			Reason: routing.DecisionInitialSlice,
		}
		if sessionCacheEntry.TimestampStart.Before(timestampNow) && packet.NumNearRelays > 0 {
			routeDecision = nextRoute.Decide(routeDecision, nnStats, directStats,
				routing.DecideUpgradeRTT(float64(buyer.RoutingRulesSettings.RTTThreshold)),
				routing.DecideDowngradeRTT(float64(buyer.RoutingRulesSettings.RTTHysteresis)),
				routing.DecideVeto(float64(buyer.RoutingRulesSettings.RTTVeto), buyer.RoutingRulesSettings.EnablePacketLossSafety, buyer.RoutingRulesSettings.EnableYouOnlyLiveOnce),
			)
		}

		locallogger = log.With(locallogger, "on_network_next", routeDecision.OnNetworkNext, "decision_reason", routeDecision.Reason)

		chosenRoute := routing.Route{
			Stats: directStats,
		}

		var token routing.Token
		if routeDecision.OnNetworkNext {
			// Route decision logic decided to serve a next route

			chosenRoute = nextRoute

			if nextRoute.Hash64() == sessionCacheEntry.RouteHash {
				token = &routing.ContinueRouteToken{
					Expires: uint64(time.Now().Add(billing.BillingSliceSeconds * time.Second).Unix()),

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
					Expires: uint64(time.Now().Add(billing.BillingSliceSeconds * 2 * time.Second).Unix()),

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
				handleError(w, response, serverPrivateKey, err)
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

		// Send the Session Response back to the server
		var responseData []byte
		if responseData, err = writeSessionResponse(w, response, serverPrivateKey); err != nil {
			level.Error(locallogger).Log("msg", "failed to write session response", "err", err)
			return
		}

		level.Debug(locallogger).Log("msg", "caching session data")
		if sessionCacheEntry, err = cacheSessionData(redisClient, &sessionCacheEntry, &packet, chosenRoute.Hash64(), time.Now(), responseData); err != nil {
			level.Error(locallogger).Log("msg", "failed to update session", "err", err)
			return
		}

		billingEntry := newBillingEntry(&chosenRoute, int(response.RouteType), &buyer.RoutingRulesSettings, routeDecision.Reason, &packet, sessionCacheEntry.TimestampStart, timestampNow)
		if err := biller.Bill(context.Background(), packet.SessionID, billingEntry); err != nil {
			level.Error(locallogger).Log("msg", "billing failed", "err", err)
		}

		counter.Add(1)
	}
}

// writeSessionResponse encrypts the session response packet and sends it back to the server. Returns the marshaled response and an error.
func writeSessionResponse(w io.Writer, packet SessionResponsePacket, privateKey []byte) ([]byte, error) {
	// Sign the response
	packet.Signature = crypto.Sign(privateKey, packet.GetSignData())

	// Marshal the packet
	responseData, err := packet.MarshalBinary()
	if err != nil {
		return nil, err
	}

	// Send the Session Response back to the server
	if _, err := w.Write(responseData); err != nil {
		return nil, err
	}

	return responseData, nil
}

// handleError forces the packet to direct and collects error metrics.
func handleError(w io.Writer, packet SessionResponsePacket, privateKey []byte, err error) {
	// Force packet to direct route
	packet.RouteType = routing.RouteTypeDirect
	writeSessionResponse(w, packet, privateKey)

	// Eventually we'll also pipe the error passed through to here up to stackdriver and do any cleanup required
}

func cacheSessionData(redisClient redis.Cmdable, prevCacheEntry *SessionCacheEntry, packet *SessionUpdatePacket, routeHash uint64, timestampExpire time.Time, responseData []byte) (SessionCacheEntry, error) {
	// Save some of the packet information to be used in SessionUpdateHandlerFunc
	sessionCacheEntry := SessionCacheEntry{
		SessionID:       packet.SessionID,
		Sequence:        packet.Sequence,
		RouteHash:       routeHash,
		TimestampStart:  prevCacheEntry.TimestampStart,
		TimestampExpire: timestampExpire,
		Version:         prevCacheEntry.Version, //This was already incremented for the route tokens
		Response:        responseData,
	}
	result := redisClient.Set(fmt.Sprintf("SESSION-%d", packet.SessionID), sessionCacheEntry, 5*time.Minute)
	if result.Err() != nil {
		return SessionCacheEntry{}, result.Err()
	}

	return sessionCacheEntry, nil
}

func newBillingEntry(
	route *routing.Route,
	routeType int,
	routingRulesSettings *routing.RoutingRulesSettings,
	decisionReason routing.DecisionReason,
	packet *SessionUpdatePacket,
	timestampStart time.Time,
	timestampNow time.Time) *billing.Entry {
	// Create billing slice flags
	sliceFlags := billing.RouteSliceFlagNone
	if routeType == routing.RouteTypeNew || routeType == routing.RouteTypeContinue {
		sliceFlags |= billing.RouteSliceFlagNext
	}

	if (decisionReason&routing.DecisionVetoRTT) != 0 ||
		(decisionReason&routing.DecisionVetoPacketLoss) != 0 ||
		(decisionReason&routing.DecisionVetoYOLO) != 0 ||
		(decisionReason&routing.DecisionVetoNoRoute) != 0 {
		sliceFlags |= billing.RouteSliceFlagVetoed
	}

	if packet.Flagged {
		sliceFlags |= billing.RouteSliceFlagReported
	}

	if packet.FallbackToDirect {
		sliceFlags |= billing.RouteSliceFlagFallbackToDirect
	}

	if (decisionReason & routing.DecisionPacketLossMultipath) != 0 {
		sliceFlags |= billing.RouteSliceFlagPacketLossMultipath
	}

	if (decisionReason & routing.DecisionJitterMultipath) != 0 {
		sliceFlags |= billing.RouteSliceFlagJitterMultipath
	}

	if (decisionReason & routing.DecisionRTTMultipath) != 0 {
		sliceFlags |= billing.RouteSliceFlagRTTMultipath
	}

	// Create slice duration
	var sliceDuration uint64
	if routeType == routing.RouteTypeContinue {
		sliceDuration = billing.BillingSliceSeconds
	} else {
		sliceDuration = billing.BillingSliceSeconds * 2
	}

	return &billing.Entry{
		Request:              nil,
		Route:                nil,
		RouteDecision:        uint64(decisionReason),
		Duration:             sliceDuration,
		UsageBytesUp:         (1000 * uint64(packet.KbpsUp)) / 8 * sliceDuration,   // Converts Kbps to bytes
		UsageBytesDown:       (1000 * uint64(packet.KbpsDown)) / 8 * sliceDuration, // Converts Kbps to bytes
		Timestamp:            uint64(timestampNow.Unix()),
		TimestampStart:       uint64(timestampStart.Unix()),
		PredictedRTT:         float32(route.Stats.RTT),
		PredictedJitter:      float32(route.Stats.Jitter),
		PredictedPacketLoss:  float32(route.Stats.PacketLoss),
		RouteChanged:         routeType != routing.RouteTypeContinue,
		NetworkNextAvailable: routeType == routing.RouteTypeNew || routeType == routing.RouteTypeContinue,
		Initial:              routeType == routing.RouteTypeNew,
		EnvelopeBytesUp:      (1000 * uint64(routingRulesSettings.EnvelopeKbpsUp)) / 8 * sliceDuration,   // Converts Kbps to bytes
		EnvelopeBytesDown:    (1000 * uint64(routingRulesSettings.EnvelopeKbpsDown)) / 8 * sliceDuration, // Converts Kbps to bytes
		ConsideredRoutes:     nil,
		AcceptableRoutes:     nil,
		SameRoute:            routeType == routing.RouteTypeContinue,
		OnNetworkNext:        packet.OnNetworkNext,
		SliceFlags:           uint64(sliceFlags),
	}
}
