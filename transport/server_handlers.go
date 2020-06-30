package transport

import (
	"bytes"
	"context"
	// "encoding/binary"
	"errors"
	"fmt"
	// fnv "hash/fnv"
	"io"
	// "math"
	// "math/rand"
	"net"
	"sync"
	"runtime"
	"time"

	// "github.com/go-kit/kit/log"
	// "github.com/go-kit/kit/log/level"
	jsoniter "github.com/json-iterator/go"

	// "github.com/go-redis/redis/v7"
	// "github.com/networknext/backend/billing"
	"github.com/networknext/backend/crypto"
	// "github.com/networknext/backend/metrics"
	"github.com/networknext/backend/routing"
	// "github.com/networknext/backend/storage"
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

func ServerInitHandlerFunc(serverPrivateKey []byte) UDPHandlerFunc {
	return func(w io.Writer, incoming *UDPPacket) {
		var packet ServerInitRequestPacket
		if err := packet.UnmarshalBinary(incoming.Data); err != nil {
			fmt.Printf("could not read server init request packet\n")
			return
		}
		response := ServerInitResponsePacket{
			RequestID: packet.RequestID,
			Response:  InitResponseOK,
			Version:   packet.Version,
		}
		if err := writeInitResponse(w, response, serverPrivateKey); err != nil {
			fmt.Printf("could not write server init response packet\n")
			return
		}
	}
}

// ==========================================================================================

/*
func ServerInitHandlerFunc(logger log.Logger, redisClient redis.Cmdable, storer storage.Storer, metrics *metrics.ServerInitMetrics, serverPrivateKey []byte) UDPHandlerFunc {

	// todo: temporarily disabled

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
			Version:   packet.Version,
		}
		defer func() {
			if err := writeInitResponse(w, response, serverPrivateKey); err != nil {
				level.Error(locallogger).Log("msg", "failed to write init response", "err", err)
			}
		}()

		if !incoming.SourceAddr.IP.IsLoopback() && !packet.Version.AtLeast(SDKVersionMin) {
			level.Error(locallogger).Log("msg", "sdk version is too old")
			response.Response = InitResponseOldSDKVersion
			metrics.ErrorMetrics.SDKTooOld.Add(1)
			return
		}

		_, err := storer.Datacenter(packet.DatacenterID)
		if err != nil {
			// Log and track the missing datacenter metric, but don't respond with an error to the SDK
			// as to allow the ServerUpdateHandlerFunc and SessionUpdateHandlerFunc to carry on working

			level.Error(locallogger).Log("msg", "failed to get datacenter from storage", "err", err)
			metrics.ErrorMetrics.DatacenterNotFound.Add(1)
		}

		buyer, err := storer.Buyer(packet.CustomerID)
		if err != nil {
			level.Error(locallogger).Log("msg", "failed to get buyer from storage", "err", err)
			response.Response = InitResponseUnknownCustomer
			metrics.ErrorMetrics.BuyerNotFound.Add(1)
			return
		}

		if !crypto.Verify(buyer.PublicKey, packet.GetSignData(), packet.Signature) {
			level.Error(locallogger).Log("msg", "signature verification failed")
			response.Response = InitResponseSignatureCheckFailed
			metrics.ErrorMetrics.VerificationFailure.Add(1)
			return
		}

		// Check if a cache entry exists for this server already in redis, and if so, remove it
		// so that the server can update probably without packet sequence issues.
		// This prevents the case where a server restarts but doesn't give the backend enough
		// time to expire the entry in redis.
		serverCacheKey := fmt.Sprintf("SERVER-%d-%s", packet.CustomerID, incoming.SourceAddr.String())
		result := redisClient.Get(serverCacheKey)
		if result.Err() != nil && result.Err() != redis.Nil {
			level.Error(locallogger).Log("msg", "failed to get server in init", "err", result.Err())
			return
		}

		// If there was no error, then the entry was found, so remove it
		if err == nil && result.Val() != "" {
			result := redisClient.Del(serverCacheKey)
			if result.Err() != nil {
				level.Error(locallogger).Log("msg", "failed to delete server cache entry in init", "server", serverCacheKey, "err", result.Err())
				return
			}

			if result.Val() == 0 {
				level.Error(locallogger).Log("msg", "could not find server cache entry in init to delete", "server", serverCacheKey, "err", result.Err())
				return
			}
		}
	}
}
*/

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

// =============================================================================

func ServerUpdateHandlerFunc() UDPHandlerFunc {
	return func(w io.Writer, incoming *UDPPacket) {
		var packet ServerUpdatePacket
		if err := packet.UnmarshalBinary(incoming.Data); err != nil {
			fmt.Printf("could not read server update packet!\n")
			return
		}
	}
}

// =============================================================================

// todo: cut down
/*
func ServerUpdateHandlerFunc(logger log.Logger, redisClient redis.Cmdable, storer storage.Storer, metrics *metrics.ServerUpdateMetrics) UDPHandlerFunc {

	return func(w io.Writer, incoming *UDPPacket) {
		var packet ServerUpdatePacket
		if err := packet.UnmarshalBinary(incoming.Data); err != nil {
			fmt.Printf("could not read server update packet!\n")
			return
		}
		serverAddress := packet.ServerAddress.String()
		// todo: store the info we need by server address in the server map
		_ = serverAddress
	}

	/*
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
			metrics.ErrorMetrics.UnserviceableUpdate.Add(1)
			metrics.ErrorMetrics.UnmarshalFailure.Add(1)
			return
		}

		serverCacheKey := fmt.Sprintf("SERVER-%d-%s", packet.CustomerID, packet.ServerAddress.String())

		locallogger := log.With(logger, "src_addr", incoming.SourceAddr.String(), "server_addr", packet.ServerAddress.String())

		// Drop the packet if version is older that the minimun sdk version
		if !incoming.SourceAddr.IP.IsLoopback() && !packet.Version.AtLeast(SDKVersionMin) {
			level.Error(locallogger).Log("msg", "sdk version is too old", "sdk", packet.Version.String())
			metrics.ErrorMetrics.UnserviceableUpdate.Add(1)
			metrics.ErrorMetrics.SDKTooOld.Add(1)
			return
		}

		locallogger = log.With(locallogger, "sdk", packet.Version.String())

		// Get the buyer information for the id in the packet
		buyer, err := storer.Buyer(packet.CustomerID)
		if err != nil {
			level.Error(locallogger).Log("msg", "failed to get buyer from storage", "err", err, "customer_id", packet.CustomerID)
			metrics.ErrorMetrics.UnserviceableUpdate.Add(1)
			metrics.ErrorMetrics.BuyerNotFound.Add(1)
			return
		}

		datacenter, err := storer.Datacenter(packet.DatacenterID)
		if err != nil {
			// Check if there is a datacenter with this alias
			var datacenterAliasFound bool
			allDatacenters := storer.Datacenters()
			for _, d := range allDatacenters {
				if packet.DatacenterID == crypto.HashID(d.AliasName) {
					datacenter = d
					datacenterAliasFound = true
					break
				}
			}

			if !datacenterAliasFound {
				level.Error(locallogger).Log("msg", "failed to get datacenter from storage", "err", err, "customer_id", packet.CustomerID)
				metrics.ErrorMetrics.UnserviceableUpdate.Add(1)
				metrics.ErrorMetrics.DatacenterNotFound.Add(1)

				// Don't return early, just set an UnknownDatacenter so the ServerCacheEntry gets sets so its used by SessionUpdateHandlerFunc
				datacenter = routing.UnknownDatacenter
				datacenter.ID = packet.DatacenterID
			}
		}

		locallogger = log.With(locallogger, "customer_id", packet.CustomerID)

		// Drop the packet if the signed packet data cannot be verified with the buyers public key
		if !crypto.Verify(buyer.PublicKey, packet.GetSignData(), packet.Signature) {
			level.Error(locallogger).Log("msg", "signature verification failed")
			metrics.ErrorMetrics.UnserviceableUpdate.Add(1)
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
			level.Error(locallogger).Log("msg", "packet too old", "packet sequence", packet.Sequence, "lastest sequence", serverentry.Sequence)
			metrics.ErrorMetrics.UnserviceableUpdate.Add(1)
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
		result := redisClient.Set(serverCacheKey, serverentry, 60*time.Second)
		if result.Err() != nil {
			level.Error(locallogger).Log("msg", "failed to update server", "err", result.Err())
			return
		}

		level.Debug(locallogger).Log("msg", "updated server")
	}
}
*/

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

/*
type RouteProvider interface {
	ResolveRelay(uint64) (routing.Relay, error)
	RelaysIn(routing.Datacenter) []routing.Relay
	Routes([]routing.Relay, []int, []routing.Relay, ...routing.SelectorFunc) ([]routing.Route, error)
}
*/

// =========================================================================================================

var sessions = map[uint64]int64{}
var sessionsMutex sync.Mutex

func TimeoutSessions() {
	for {
		timeout := time.Now().Unix() - 30
		sessionsMutex.Lock()
		numSessions := 0
		for k,v := range sessions {
			if v < timeout {
				delete(sessions, k)
			} else {
				numSessions++
			}
		}
		sessionsMutex.Unlock()
		fmt.Printf("%d sessions\n")
		time.Sleep(time.Second * 10)
	}
}

func SessionUpdateHandlerFunc(serverPrivateKey []byte) UDPHandlerFunc {
	return func(w io.Writer, incoming *UDPPacket) {
		var header SessionUpdatePacketHeader
		if err := header.UnmarshalBinary(incoming.Data); err != nil {
			fmt.Printf("could not read session update packet header!\n")
			return
		}
		response := SessionResponsePacket{
			Sequence:  header.Sequence,
			SessionID: header.SessionID,
			RouteType: int32(routing.RouteTypeDirect),
		}
		sessionsMutex.Lock()
		sessions[header.SessionID] = time.Now().Unix()
		sessionsMutex.Unlock()
		if _, err := writeSessionResponse(w, response, serverPrivateKey); err != nil {
			fmt.Printf("could not write session update response packet!\n")
			return
		}
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
			serverCacheData, err := serverCacheCmd.Bytes()
			if err != nil {

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

// todo: disabled
/*
func updatePortalData(redisClientPortal redis.Cmdable, redisClientPortalExp time.Duration, packet SessionUpdatePacket, nnStats routing.Stats, directStats routing.Stats, relayHops []routing.Relay, onNetworkNext bool, datacenterName string, location routing.Location, sessionTime time.Time, isMultiPath bool) error {
	if (nnStats.RTT == 0 && directStats.RTT == 0) || (onNetworkNext && nnStats.RTT == 0) {
		return nil
	}

	clientAddr := AnonymizeAddr(packet.ClientAddress)

	if clientAddr.IP == nil {
		return fmt.Errorf("failed to anonymize client addr")
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
		deltaRTT = directStats.RTT - nnStats.RTT
	}

	meta := routing.SessionMeta{
		ID:            fmt.Sprintf("%016x", packet.SessionID),
		UserHash:      hashedID,
		Datacenter:    datacenterName,
		OnNetworkNext: onNetworkNext,
		NextRTT:       nnStats.RTT,
		DirectRTT:     directStats.RTT,
		DeltaRTT:      deltaRTT,
		Location:      location,
		ClientAddr:    clientAddr.String(),
		ServerAddr:    packet.ServerAddress.String(),
		Hops:          relayHops,
		SDK:           packet.Version.String(),
		Connection:    ConnectionTypeText(packet.ConnectionType),
		Platform:      PlatformTypeText(packet.PlatformID),
		BuyerID:       fmt.Sprintf("%016x", packet.CustomerID),
	}

	meta.NearbyRelays = make([]routing.Relay, 0)

	// Only fill in the essential information here to then let the portal fill in additional relay info
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
		Timestamp: sessionTime,
		Next:      nnStats,
		Direct:    directStats,
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

	// set the map point key and global sessions with expiration on the entire key set for safety
	tx.Set(fmt.Sprintf("session-%016x-point", packet.SessionID), point, redisClientPortalExp)
	tx.SAdd("map-points-global", meta.ID)
	tx.Expire("map-points-global", redisClientPortalExp)

	if _, err := tx.Exec(); err != nil {
		return err
	}

	return nil
}
*/

// todo: disabled
/*
func submitBillingEntry(biller billing.Biller, serverCacheEntry ServerCacheEntry, sessionCacheEntry SessionCacheEntry, request SessionUpdatePacket, response SessionResponsePacket,
	buyer routing.Buyer, chosenRoute routing.Route, location routing.Location, storer storage.Storer, clientRelays []routing.Relay, routeDecision routing.Decision,
	sliceDuration uint64, timestampStart time.Time, timestampNow time.Time, newSession bool) error {

	sameRoute := chosenRoute.Hash64() == sessionCacheEntry.RouteHash
	routeRequest := NewRouteRequest(request, &buyer, serverCacheEntry, location, storer, clientRelays)
	billingEntry := NewBillingEntry(routeRequest, &chosenRoute, int(response.RouteType), sameRoute, &buyer.RoutingRulesSettings, routeDecision, &request, sliceDuration, timestampStart, timestampNow, newSession)
	return biller.Bill(context.Background(), request.SessionID, billingEntry)
}

func updateCacheEntries(redisClient redis.Cmdable, sessionCacheKey string, vetoCacheKey string, sessionCacheEntry SessionCacheEntry, vetoCacheEntry VetoCacheEntry, packet SessionUpdatePacket, chosenRouteHash uint64,
	routeDecision routing.Decision, timestampStart time.Time, timestampExpire time.Time, responseData []byte, directRTT float64, nextRTT float64, location routing.Location) error {
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
		Location:                   location,
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
*/

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

// todo: disabled
/*
func writeSessionErrorResponse(w io.Writer, response SessionResponsePacket, privateKey []byte, directSessions metrics.Counter, unserviceableUpdateCounter metrics.Counter, errCounter metrics.Counter) ([]byte, error) {
	responseData, err := writeSessionResponse(w, response, privateKey)
	if err != nil {
		return nil, err
	}

	directSessions.Add(1)

	unserviceableUpdateCounter.Add(1)
	errCounter.Add(1)

	return responseData, nil
}
*/

