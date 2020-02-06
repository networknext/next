package transport

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	jsoniter "github.com/json-iterator/go"

	"github.com/go-redis/redis/v7"
	"github.com/networknext/backend/crypto"
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
func (m *UDPServerMux) Start(ctx context.Context, handlers int) error {
	if m.Conn == nil {
		return errors.New("relay server cannot be nil")
	}

	for i := 0; i < handlers; i++ {
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

		var buf bytes.Buffer
		packet := UDPPacket{
			SourceAddr: addr,
			Data:       data[:numbytes],
		}

		switch data[0] {
		case PacketTypeServerUpdate:
			m.ServerUpdateHandlerFunc(&buf, &packet)
		case PacketTypeSessionUpdate:
			m.SessionUpdateHandlerFunc(&buf, &packet)
		}

		if buf.Len() > 0 {
			m.Conn.WriteToUDP(buf.Bytes(), addr)
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

type BuyerProvider interface {
	GetAndCheckBySdkVersion3PublicKeyId(key uint64) (*storage.Buyer, bool)
}

// ServerUpdateHandlerFunc ...
func ServerUpdateHandlerFunc(logger log.Logger, redisClient redis.Cmdable, bp BuyerProvider) UDPHandlerFunc {
	logger = log.With(logger, "handler", "server")

	return func(w io.Writer, incoming *UDPPacket) {
		var packet ServerUpdatePacket
		if err := packet.UnmarshalBinary(incoming.Data); err != nil {
			level.Error(logger).Log("msg", "could not read packet", "err", err)
			return
		}

		locallogger := log.With(logger, "src_addr", incoming.SourceAddr.String(), "server_addr", packet.ServerAddress.String())

		// Drop the packet if version is older that the minimun sdk version
		psdkv := SDKVersion{packet.VersionMajor, packet.VersionMinor, packet.VersionPatch}
		if !incoming.SourceAddr.IP.IsLoopback() &&
			psdkv.Compare(SDKVersionMin) == SDKVersionOlder {
			level.Error(locallogger).Log("msg", "sdk version is too old", "sdk", psdkv.String())
			return
		}

		locallogger = log.With(locallogger, "sdk", psdkv.String())

		// Get the buyer information for the id in the packet
		buyer, ok := bp.GetAndCheckBySdkVersion3PublicKeyId(packet.CustomerId)
		if !ok {
			level.Error(locallogger).Log("msg", "failed to get buyer", "customer_id", packet.CustomerId)
			return
		}

		locallogger = log.With(locallogger, "customer_id", packet.CustomerId)

		// This was in the Router, but no sense having that buried in there when we already have
		// a Buyer to check before requesting a Route
		if !buyer.GetActive() {
			level.Warn(locallogger).Log("msg", "buyer is inactive")
			return
		}

		buyerPublicKey := buyer.SdkVersion3PublicKeyData
		// Drop the packet if the buyer is not an admin and they are using an internal build
		// if !buyer.GetA && psdkv.Compare(SDKVersionInternal) == SDKVersionEqual {
		// 	log.Printf("non-admin buyer using an internal sdk")
		// 	return
		// }

		// Drop the packet if the signed packet data cannot be verified with the buyers public key
		if !crypto.Verify(buyerPublicKey, packet.GetSignData(), packet.Signature) {
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
			Datacenter: routing.Datacenter{ID: packet.DatacenterId},
			SDKVersion: SDKVersion{packet.VersionMajor, packet.VersionMinor, packet.VersionPatch},
		}
		result := redisClient.Set("SERVER-"+incoming.SourceAddr.String(), serverentry, 5*time.Minute)
		if result.Err() != nil {
			level.Error(locallogger).Log("msg", "failed to update server", "err", result.Err())
			return
		}

		level.Debug(locallogger).Log("msg", "updated server")
	}
}

type SessionCacheEntry struct {
	SessionID uint64
	Sequence  uint64
	RouteHash uint64
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
	Routes([]routing.Relay, []routing.Relay) []routing.Route
}

// SessionUpdateHandlerFunc ...
func SessionUpdateHandlerFunc(logger log.Logger, redisClient redis.Cmdable, bp BuyerProvider, rp RouteProvider, iploc routing.IPLocator, geoClient *routing.GeoClient, serverPrivateKey []byte, routerPrivateKey []byte) UDPHandlerFunc {
	logger = log.With(logger, "handler", "session")

	return func(w io.Writer, incoming *UDPPacket) {
		// Deserialize the Session packet
		var packet SessionUpdatePacket
		if err := packet.UnmarshalBinary(incoming.Data); err != nil {
			level.Error(logger).Log("msg", "could not read packet", "err", err)
			return
		}

		locallogger := log.With(logger, "src_addr", incoming.SourceAddr.String(), "server_addr", packet.ServerAddress.String(), "client_addr", packet.ClientAddress.String(), "session_id", packet.SessionId)

		var serverCacheEntry ServerCacheEntry
		var sessionCacheEntry SessionCacheEntry

		// Build a redis transaction to make a single network call
		tx := redisClient.TxPipeline()
		{
			serverCacheCmd := tx.Get("SERVER-" + incoming.SourceAddr.String())
			sessionCacheCmd := tx.Get(fmt.Sprintf("SESSION-%d", packet.SessionId))
			tx.Exec()

			serverCacheData, err := serverCacheCmd.Bytes()
			if err != nil {
				level.Error(locallogger).Log("msg", "failed to get server bytes", "err", err)
				return
			}
			if err := serverCacheEntry.UnmarshalBinary(serverCacheData); err != nil {
				level.Error(locallogger).Log("msg", "failed to unmarshal server bytes", "err", err)
				return
			}

			if sessionCacheCmd.Err() != redis.Nil {
				sessionCacheData, err := sessionCacheCmd.Bytes()
				if err != nil {
					level.Error(locallogger).Log("msg", "failed to get session bytes", "err", err)
					return
				}

				if err := sessionCacheEntry.UnmarshalBinary(sessionCacheData); err != nil {
					level.Error(locallogger).Log("msg", "failed to unmarshal session bytes", "err", err)
					return
				}
			}
		}

		locallogger = log.With(locallogger, "datacenter_id", serverCacheEntry.Datacenter.ID)

		buyer, ok := bp.GetAndCheckBySdkVersion3PublicKeyId(packet.CustomerId)
		if !ok {
			level.Error(locallogger).Log("msg", "failed to get buyer", "customer_id", packet.CustomerId)
			return
		}

		locallogger = log.With(locallogger, "customer_id", packet.CustomerId)

		buyerServerPublicKey := buyer.SdkVersion3PublicKeyData

		if !crypto.Verify(buyerServerPublicKey, packet.GetSignData(serverCacheEntry.SDKVersion), packet.Signature) {
			level.Error(locallogger).Log("msg", "signature verification failed")
			return
		}

		if packet.Sequence < sessionCacheEntry.Sequence {
			level.Warn(locallogger).Log("msg", "packet sequency too old", "current_sequence", packet.Sequence, "previous_sequence", sessionCacheEntry.Sequence)
			return
		}

		location, err := iploc.LocateIP(packet.ClientAddress.IP)
		if err != nil {
			level.Error(locallogger).Log("msg", "failed to locate client")
			return
		}
		level.Debug(locallogger).Log("lat", location.Latitude, "long", location.Longitude)

		clientrelays, err := geoClient.RelaysWithin(location.Latitude, location.Longitude, 500, "mi")
		if err != nil {
			level.Error(locallogger).Log("msg", "failed to locate relays near client")
			return
		}

		// We need to do this because RelaysWithin only has the ID of the relay and we need the Addr and PublicKey too
		// Maybe we consider a nicer way to do this in the future
		for idx := range clientrelays {
			clientrelays[idx], _ = rp.ResolveRelay(clientrelays[idx].ID)
		}

		dsrelays := rp.RelaysIn(serverCacheEntry.Datacenter)

		level.Debug(locallogger).Log("num_datacenter_relays", len(dsrelays), "num_client_relays", len(clientrelays))

		// Get a set of possible routes from the RouteProvider an on error ensure it falls back to direct
		routes := rp.Routes(dsrelays, clientrelays)
		if routes == nil {
			level.Error(locallogger).Log("msg", "failed to find routes")
			return
		}
		chosenRoute := routes[0] // Just take the first one it find regardless of optimization

		// Check if the chosen route is the same one as we previously sent

		var token routing.Token
		{
			if chosenRoute.Hash64() == sessionCacheEntry.RouteHash {
				token = &routing.ContinueRouteToken{
					Expires: uint64(time.Now().Add(10 * time.Second).Unix()),

					SessionId: packet.SessionId,

					SessionVersion: 0, // Haven't figured out what this is for
					SessionFlags:   0, // Haven't figured out what this is for

					Client: routing.Client{
						Addr:      packet.ClientAddress,
						PublicKey: packet.ClientRoutePublicKey,
					},

					Server: routing.Server{
						Addr:      packet.ServerAddress,
						PublicKey: buyerServerPublicKey,
					},

					Relays: chosenRoute.Relays,
				}
			} else {
				token = &routing.NextRouteToken{
					Expires: uint64(time.Now().Add(10 * time.Second).Unix()),

					SessionId: packet.SessionId,

					SessionVersion: 0, // Haven't figured out what this is for
					SessionFlags:   0, // Haven't figured out what this is for

					Client: routing.Client{
						Addr:      packet.ClientAddress,
						PublicKey: packet.ClientRoutePublicKey,
					},

					Server: routing.Server{
						Addr:      packet.ServerAddress,
						PublicKey: buyerServerPublicKey,
					},

					Relays: chosenRoute.Relays,
				}
			}
		}

		tokens, numtokens, err := token.Encrypt(routerPrivateKey)
		if err != nil {
			level.Error(locallogger).Log("msg", "failed to encrypt route token")
			return
		}

		// Create the Session Response for the server
		response := SessionResponsePacket{
			Sequence:             packet.Sequence,
			SessionId:            packet.SessionId,
			RouteType:            int32(token.Type()),
			NumTokens:            int32(numtokens), // Num of relays + client + server
			Tokens:               tokens,
			Multipath:            false, // Haven't figured out what this is for
			ServerRoutePublicKey: serverCacheEntry.Server.PublicKey,
		}

		// Fill in the near relays
		response.NumNearRelays = int32(len(clientrelays))
		response.NearRelayIds = make([]uint64, len(clientrelays))
		response.NearRelayAddresses = make([]net.UDPAddr, len(clientrelays))
		for idx, relay := range clientrelays {
			response.NearRelayIds[idx] = relay.ID
			response.NearRelayAddresses[idx] = relay.Addr
		}

		// Sign the response
		response.Signature = crypto.Sign(serverPrivateKey, response.GetSignData())

		// Send the Session Response back to the server
		res, err := response.MarshalBinary()
		if err != nil {
			level.Error(locallogger).Log("msg", "failed to marshal session response", "err", err)
			return
		}

		if _, err := w.Write(res); err != nil {
			level.Error(locallogger).Log("msg", "failed to write session response", "err", err)
		}
	}
}
