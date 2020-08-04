package transport

import (
	"bytes"
	"context"
	"encoding/binary"
	"hash/fnv"
	"os"
	"strconv"
	"syscall"

	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"

	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/networknext/backend/billing"
	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/metrics"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
	"github.com/networknext/backend/transport/pubsub"
	"golang.org/x/sys/unix"

	"github.com/panjf2000/ants/v2"
)

type UDPPacket struct {
	SourceAddr net.UDPAddr
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

type UDPServerMux2 struct {
	Logger              log.Logger
	SessionErrorMetrics *metrics.SessionErrorMetrics
	PortalPublisher     pubsub.Publisher
	Biller              billing.Biller
	MaxPacketSize       int
	Port                int64

	ServerInitHandlerFunc    UDPHandlerFunc
	ServerUpdateHandlerFunc  UDPHandlerFunc
	SessionUpdateHandlerFunc func(io.Writer, *UDPPacket, PostSessionUpdateFunc)
}

// Start begins accepting UDP packets from the UDP connection and will block
func (m *UDPServerMux) Start(ctx context.Context) error {
	if m.Conn == nil {
		return errors.New("udp connection cannot be nil")
	}

	// todo: fucks up on 96 core otherwise
	numThreads := 8

	for i := 0; i < numThreads; i++ {
		go m.handler(ctx)
	}

	<-ctx.Done()

	return nil
}

// Start begins accepting UDP packets from the UDP connection and will block
func (m *UDPServerMux2) Start(ctx context.Context) error {
	// Create a post session handler to handle the post process of session updates.
	// This way, we can quickly return from the session update handler and not spawn a
	// ton of goroutines if things get backed up.
	postSessionHandler := NewPostSessionHandler(100, 1000000, m.PortalPublisher, m.Biller, m.Logger, m.SessionErrorMetrics)
	postSessionHandler.StartProcessing(ctx)

	numThreads := 8
	numSockets, ok := os.LookupEnv("NUM_UDP_SOCKETS")
	if ok {
		iNumSockets, err := strconv.ParseInt(numSockets, 10, 64)
		if err == nil {
			numThreads = int(iNumSockets)
		}
	}

	if b, err := strconv.ParseBool(os.Getenv("USE_THREAD_POOL")); err == nil && b {
		numPktThreads := 256
		if t, err := strconv.ParseUint(os.Getenv("NUM_PACKET_PROCESSING_THREADS"), 10, 64); err == nil && t > 0 {
			numPktThreads = int(t)
		}

		pools := make([]*ants.Pool, 0)
		for i := 0; i < numThreads; i++ {
			procPool, err := ants.NewPool(numPktThreads)
			if err != nil {
				level.Error(m.Logger).Log("msg", "could not create pkt recv thread pool", "err", err)
				os.Exit(1)
			}

			go m.handler(ctx, threadPoolHandlerFunc(m, postSessionHandler, procPool))

			pools = append(pools, procPool)
		}

		<-ctx.Done()

		for _, pool := range pools {
			pool.Release()
		}
	} else {
		for i := 0; i < numThreads; i++ {
			go m.handler(ctx, goroutineHandlerFunc(m, postSessionHandler))
		}
		<-ctx.Done()
	}

	return nil
}

func (m *UDPServerMux) handler(ctx context.Context) {
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

			packet := UDPPacket{SourceAddr: *from, Data: packet_data}

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

				m.Conn.WriteToUDP(res, from)
			}

		}(data, size, addr)
	}
}

type packetHandlerFunc = func(conn *net.UDPConn, packet_data []byte, packet_size int, from *net.UDPAddr)

// returns a function that handles udp packets through normal goroutines
func goroutineHandlerFunc(m *UDPServerMux2, postSessionHandler *PostSessionHandler) packetHandlerFunc {
	return func(conn *net.UDPConn, packetData []byte, packetSize int, from *net.UDPAddr) {
		go func() {
			m.handlePacket(conn, packetData, packetSize, from, goroutinePostSessionUpdateFunc(postSessionHandler))
		}()
	}
}

// returns a function that handles udp packets through a thread pool
func threadPoolHandlerFunc(m *UDPServerMux2, postSessionHandler *PostSessionHandler, procPool *ants.Pool) packetHandlerFunc {
	return func(conn *net.UDPConn, packetData []byte, packetSize int, from *net.UDPAddr) {
		procPool.Submit(func() {
			m.handlePacket(conn, packetData, packetSize, from, goroutinePostSessionUpdateFunc(postSessionHandler))
		})
	}
}

func (m *UDPServerMux2) handler(ctx context.Context, handleFunc packetHandlerFunc) {
	var conn *net.UDPConn
	// Initialize UDP connection
	{
		lc := net.ListenConfig{
			Control: func(network, address string, c syscall.RawConn) error {
				var opErr error
				err := c.Control(func(fd uintptr) {
					opErr = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
				})
				if err != nil {
					return err
				}
				return opErr
			},
		}

		lp, err := lc.ListenPacket(context.Background(), "udp", fmt.Sprintf("0.0.0.0:%d", m.Port))
		if err != nil {
			level.Error(m.Logger).Log("udp", "listenPacket", "msg", "could not bind", "err", err)
			os.Exit(1)
		}

		conn = lp.(*net.UDPConn)

		readBufferString, ok := os.LookupEnv("READ_BUFFER")
		if ok {
			readBuffer, err := strconv.ParseInt(readBufferString, 10, 64)
			if err != nil {
				level.Error(m.Logger).Log("envvar", "READ_BUFFER", "msg", "could not parse", "err", err)
				os.Exit(1)
			}
			conn.SetReadBuffer(int(readBuffer))
		}

		writeBufferString, ok := os.LookupEnv("WRITE_BUFFER")
		if ok {
			writeBuffer, err := strconv.ParseInt(writeBufferString, 10, 64)
			if err != nil {
				level.Error(m.Logger).Log("envvar", "WRITE_BUFFER", "msg", "could not parse", "err", err)
				os.Exit(1)
			}
			conn.SetWriteBuffer(int(writeBuffer))
		}
	}

	for {
		data := make([]byte, m.MaxPacketSize)

		size, addr, _ := conn.ReadFromUDP(data)
		if size <= 0 {
			continue
		}

		data = data[:size]

		handleFunc(conn, data, size, addr)
	}
}

func (m *UDPServerMux2) handlePacket(conn *net.UDPConn, packetData []byte, packetSize int, from *net.UDPAddr, postSessionUpdateFunc PostSessionUpdateFunc) {
	// Check the packet hash is legit and remove the hash from the beginning of the packet
	// to continue processing the packet as normal
	hashedPacket := crypto.Check(crypto.PacketHashKey, packetData)
	switch hashedPacket {
	case true:
		packetData = packetData[crypto.PacketHashSize:packetSize]
	default:
		// todo: once everybody has upgraded to SDK 3.4.5 or greater, this is an error. ignore packet.
		packetData = packetData[:packetSize]
	}

	packet := UDPPacket{SourceAddr: *from, Data: packetData}

	var buf bytes.Buffer

	switch packet.Data[0] {
	case PacketTypeServerInitRequest:
		m.ServerInitHandlerFunc(&buf, &packet)
	case PacketTypeServerUpdate:
		m.ServerUpdateHandlerFunc(&buf, &packet)
	case PacketTypeSessionUpdate:
		m.SessionUpdateHandlerFunc(&buf, &packet, postSessionUpdateFunc)
	}

	if buf.Len() > 0 {
		res := buf.Bytes()

		// If the hash checks out above then hash the response to the sender
		if hashedPacket {
			res = crypto.Hash(crypto.PacketHashKey, res)
		}

		conn.WriteToUDP(res, from)
	}
}

// ==========================================================================================

type ServerInitParams struct {
	ServerPrivateKey  []byte
	Storer            storage.Storer
	Metrics           *metrics.ServerInitMetrics
	Logger            log.Logger
	DatacenterTracker *DatacenterTracker
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

		params.Metrics.Invocations.Add(1)

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

		// IMPORTANT: Make sure that if we can't find the datacenter or alias, we *still* continue as normal
		// and create an UnknownDatacenter so that we can respond to sessions on that game server with a direct route!

		datacenter, err := params.Storer.Datacenter(packet.DatacenterID)
		if err != nil {
			// search the list of aliases created by/for this buyer
			datacenterAliases := params.Storer.GetDatacenterMapsForBuyer(packet.CustomerID)
			if len(datacenterAliases) == 0 {
				params.Metrics.ErrorMetrics.DatacenterNotFound.Add(1)
				writeServerInitResponse(params, w, &packet, InitResponseUnknownDatacenter)
			} else {
				for _, dcMap := range datacenterAliases {
					if packet.DatacenterID == crypto.HashID(dcMap.Alias) {
						datacenter, err = params.Storer.Datacenter(dcMap.Datacenter)
						if err != nil {
							params.Metrics.ErrorMetrics.DatacenterNotFound.Add(1)
							writeServerInitResponse(params, w, &packet, InitResponseUnknownDatacenter)
						}
						datacenter.AliasName = dcMap.Alias
					}
				}
			}
		}

		// Track datacenter IDs we don't know about so we can work with customers to add them to our database.
		if datacenter.ID == routing.UnknownDatacenter.ID {
			params.DatacenterTracker.AddUnknownDatacenter(packet.DatacenterID)
		}

		// If we get down here, all checks have passed and this server is OK to init.
		// Once a server inits, it goes into a mode where it can potentially monitor and accelerate sessions.
		// After 10 seconds, if the server fails to init, it will fall back to direct and not monitor or accelerate
		// sessions until it is restarted.

		// IMPORTANT: In a future SDK version, it is probably important that we extend the server code to retry initialization,
		// since right now it only re-initializes if that server is restarted, and we can't rely on all our customers to regularly
		// restart their servers (although Psyonix does do this).

		writeServerInitResponse(params, w, &packet, InitResponseOK)
	}
}

type ServerUpdateParams struct {
	Storer            storage.Storer
	Metrics           *metrics.ServerUpdateMetrics
	Logger            log.Logger
	ServerMap         *ServerMap
	DatacenterTracker *DatacenterTracker
}

func ServerUpdateHandlerFunc(params *ServerUpdateParams) UDPHandlerFunc {

	return func(w io.Writer, incoming *UDPPacket) {

		params.Metrics.Invocations.Add(1)

		// Read the entire server update packet. We can do this all at once because the packet contains the SDK version in it.

		var packet ServerUpdatePacket
		if err := packet.UnmarshalBinary(incoming.Data); err != nil {
			level.Error(params.Logger).Log("msg", "could not read server update packet", "err", err)
			params.Metrics.ErrorMetrics.UnserviceableUpdate.Add(1)
			params.Metrics.ErrorMetrics.ReadPacketFailure.Add(1)
			return
		}

		// Ignore the server update if the SDK version is too old
		// When we ignore updates like this all sessions going to that server will simply go direct.
		// This lets us deprecate old versions of the SDK.

		// todo: in the old code we checked if we were running on a buyer account with "internal" set, and allowed 0.0.0 there only.
		// this is much better than checking the loopback address here. please fix ryan
		if !incoming.SourceAddr.IP.IsLoopback() && !packet.Version.AtLeast(SDKVersionMin) {
			level.Error(params.Logger).Log("msg", "ignoring old sdk version", "version", packet.Version.String())
			params.Metrics.ErrorMetrics.UnserviceableUpdate.Add(1)
			params.Metrics.ErrorMetrics.SDKTooOld.Add(1)
			return
		}

		// Get the buyer information for the customer id in the packet.
		// If the buyer does not exist, this is not a server we care about.

		buyer, err := params.Storer.Buyer(packet.CustomerID)
		if err != nil {
			level.Error(params.Logger).Log("msg", "failed to get buyer from storage", "err", err, "customer_id", packet.CustomerID)
			params.Metrics.ErrorMetrics.UnserviceableUpdate.Add(1)
			params.Metrics.ErrorMetrics.BuyerNotFound.Add(1)
			return
		}

		// Check the server update is signed by the private key of the buyer.
		// If the signature does not match, this is not a server we care about.

		if !crypto.Verify(buyer.PublicKey, packet.GetSignData(), packet.Signature) {
			level.Error(params.Logger).Log("msg", "signature verification failed")
			params.Metrics.ErrorMetrics.UnserviceableUpdate.Add(1)
			params.Metrics.ErrorMetrics.VerificationFailure.Add(1)
			return
		}

		// Look up the datacenter by id and make sure it exists.
		// Sometimes the customer has datacenter aliases, eg: "multiplay.newyork" -> "inap.newyork".
		// To support this, when we can't find a datacenter directly by id, we look it up by alias instead.
		// Datacenter aliases are per-customer. Different customers have different datacenter aliases.

		datacenter, err := params.Storer.Datacenter(packet.DatacenterID)
		if err != nil {
			datacenter = routing.UnknownDatacenter

			datacenterAliases := params.Storer.GetDatacenterMapsForBuyer(packet.CustomerID)
			if len(datacenterAliases) == 0 {
				level.Error(params.Logger).Log("err", "no datacenter map found", "customerID", fmt.Sprintf("%016x", packet.CustomerID))
				params.Metrics.ErrorMetrics.DatacenterNotFound.Add(1)
				params.Metrics.ErrorMetrics.UnserviceableUpdate.Add(1)
			} else {
				aliasFound := false
				for _, dcMap := range datacenterAliases {
					if packet.DatacenterID == crypto.HashID(dcMap.Alias) {
						datacenter, err = params.Storer.Datacenter(dcMap.Datacenter)
						if err != nil {
							datacenter = routing.UnknownDatacenter

							level.Error(params.Logger).Log("msg", "datacenter alias found but could not retrieve datacenter", "err", err, "customerID", fmt.Sprintf("%016x", packet.CustomerID))
							params.Metrics.ErrorMetrics.DatacenterNotFound.Add(1)
							params.Metrics.ErrorMetrics.UnserviceableUpdate.Add(1)
							aliasFound = true // Set this to true to avoid double counting metrics
							break

						}
						datacenter.AliasName = dcMap.Alias
						aliasFound = true
						break
					}
				}

				if !aliasFound {
					datacenter = routing.UnknownDatacenter

					level.Error(params.Logger).Log("msg", "datacenter alias map does not contain datacenter", "datacenterID", fmt.Sprintf("%016x", packet.DatacenterID), "customerID", fmt.Sprintf("%016x", packet.CustomerID))
					params.Metrics.ErrorMetrics.DatacenterNotFound.Add(1)
					params.Metrics.ErrorMetrics.UnserviceableUpdate.Add(1)
				}
			}
		}

		// Track datacenter IDs we don't know about so we can work with customers to add them to our database.
		if datacenter.ID == routing.UnknownDatacenter.ID {
			params.DatacenterTracker.AddUnknownDatacenter(packet.DatacenterID)
		}

		// UDP packets may arrive out of order. So that we don't have stale server update packets arriving late and
		// ruining our server map with stale information, we must check the server update sequence number, and discard
		// any server updates that are the same sequence number or older than the current server entry in the server map.

		var sequence uint64

		serverAddress := packet.ServerAddress.String()

		params.ServerMap.Lock(buyer.ID, serverAddress)
		defer params.ServerMap.Unlock(buyer.ID, serverAddress)

		serverDataReadOnly := params.ServerMap.GetServerData(buyer.ID, serverAddress)
		if serverDataReadOnly != nil {
			sequence = serverDataReadOnly.Sequence
		}

		// todo: disable as a test
		_ = sequence
		/*
			if packet.Sequence < sequence {
				level.Error(params.Logger).Log("handler", "server", "msg", "packet too old", "packet sequence", packet.Sequence, "lastest sequence", serverDataReadOnly.sequence)
				params.Metrics.ErrorMetrics.UnserviceableUpdate.Add(1)
				params.Metrics.ErrorMetrics.PacketSequenceTooOld.Add(1)
				return
			}
		*/

		// Each one of our customer's servers reports to us with this server update packet every 10 seconds.
		// Therefore we must update the server data each time we receive an update, to keep this server entry live in our server map.
		// When we don't receive an update for a server for a certain period of time (for example 30 seconds), that server entry times out.

		server := ServerData{
			Timestamp:      time.Now().Unix(),
			RoutePublicKey: packet.ServerRoutePublicKey,
			Version:        packet.Version,
			Datacenter:     datacenter,
			Sequence:       packet.Sequence,
		}

		params.ServerMap.UpdateServerData(buyer.ID, serverAddress, &server)
	}
}

type RouteProvider interface {
	ResolveRelay(id uint64) (routing.Relay, error)
	GetDatacenterRelays(datacenter routing.Datacenter) []routing.Relay
	GetRoutes(near []routing.Relay, dest []routing.Relay) ([]routing.Route, error)
	GetNearRelays(latitude float64, longitude float64, maxNearRelays int) ([]routing.Relay, error)
}

type SessionUpdateParams struct {
	ServerPrivateKey  []byte
	RouterPrivateKey  []byte
	GetRouteProvider  func() RouteProvider
	GetIPLocator      func() routing.IPLocator
	Storer            storage.Storer
	Biller            billing.Biller
	Metrics           *metrics.SessionMetrics
	Logger            log.Logger
	VetoMap           *VetoMap
	ServerMap         *ServerMap
	SessionMap        *SessionMap
	DatacenterTracker *DatacenterTracker
	PortalPublisher   pubsub.Publisher
	InstanceID        uint64
}

func SessionUpdateHandlerFunc(params *SessionUpdateParams) func(io.Writer, *UDPPacket, PostSessionUpdateFunc) {

	return func(w io.Writer, incoming *UDPPacket, postSessionUpdateFunc PostSessionUpdateFunc) {

		params.Metrics.Invocations.Add(1)

		// First, read the session update packet header.
		// We have to read only the header first, because the rest of the session update packet depends on SDK version
		// and we don't know the version yet, since that's stored in the server data for this session, not in the packet.

		var header SessionUpdatePacketHeader
		if err := header.UnmarshalBinary(incoming.Data); err != nil {
			level.Error(params.Logger).Log("msg", "could not read session update packet header", "err", err)
			params.Metrics.ErrorMetrics.UnserviceableUpdate.Add(1)
			params.Metrics.ErrorMetrics.ReadPacketHeaderFailure.Add(1)
			return
		}

		// Look up the buyer entry by the customer id. At this point if we can't find it, just ignore the session and don't respond.
		// If somebody is sending us a session update with an invalid customer id, we don't need to waste any bandwidth responding to it.

		buyer, err := params.Storer.Buyer(header.CustomerID)
		if err != nil {
			level.Error(params.Logger).Log("msg", "failed to get buyer from storage", "err", err, "customer_id", header.CustomerID)
			params.Metrics.ErrorMetrics.UnserviceableUpdate.Add(1)
			params.Metrics.ErrorMetrics.BuyerNotFound.Add(1)
			return
		}

		// Grab the server data corresponding to the server this session is talking to.
		// The server data is necessary for us to read the rest of the session update packet.

		params.ServerMap.RLock(buyer.ID, header.ServerAddress.String())
		serverDataReadOnly := params.ServerMap.GetServerData(buyer.ID, header.ServerAddress.String())
		params.ServerMap.RUnlock(buyer.ID, header.ServerAddress.String())
		if serverDataReadOnly == nil {
			level.Error(params.Logger).Log("msg", "server data missing")
			params.Metrics.ErrorMetrics.UnserviceableUpdate.Add(1)
			params.Metrics.ErrorMetrics.ServerDataMissing.Add(1)
			return
		}

		// Now that we have the server data, we know the SDK version, so we can read the rest of the session update packet.

		var packet SessionUpdatePacket
		packet.Version = serverDataReadOnly.Version
		if err := packet.UnmarshalBinary(incoming.Data); err != nil {
			level.Error(params.Logger).Log("msg", "could not read session update packet", "err", err)
			params.Metrics.ErrorMetrics.UnserviceableUpdate.Add(1)
			params.Metrics.ErrorMetrics.ReadPacketFailure.Add(1)
			return
		}

		// Get the session data from the SDK map
		// Since we write back into this map, we must lock and unlock at the end of this function
		// otherwise under heavy contention weird stuff happens with the session map.

		params.SessionMap.Lock(header.SessionID)
		defer params.SessionMap.Unlock(header.SessionID)

		sessionDataReadOnly := params.SessionMap.GetSessionData(header.SessionID)
		if sessionDataReadOnly == nil {
			sessionDataReadOnly = NewSessionData()
		}

		// todo: disable for now as a test
		/*
			// Check the packet sequence number vs. the most recent sequence number in redis.
			// The packet sequence number must be at least as old as the current session sequence #
			// otherwise this is a stale session update packet from an older slice so we ignore it!

			if packet.Sequence < sessionDataReadOnly.sequence {
				level.Error(params.Logger).Log("handler", "session", "msg", "packet too old", "packet sequence", packet.Sequence, "lastest sequence", sessionDataReadOnly.sequence)
				params.Metrics.ErrorMetrics.UnserviceableUpdate.Add(1)
				params.Metrics.ErrorMetrics.OldSequence.Add(1)
				return
			}
		*/

		// Check the session update packet is properly signed with the customer private key.
		// Any session update not signed is invalid, so we don't waste bandwidth responding to it.

		if !crypto.Verify(buyer.PublicKey, packet.GetSignData(), packet.Signature) {
			level.Error(params.Logger).Log("err", "could not verify session update packet", "customer_id", packet.CustomerID)
			params.Metrics.ErrorMetrics.VerifyFailure.Add(1)
			return
		}

		// When multiple session updates are in flight, especially under a retry storm, there can be simultaneous calls
		// to this handler for the same session and slice. It is *extremely important* that we don't generate multiple route
		// responses in this case, otherwise we'll bill our customers multiple times for the same slice!. Instead, we implement
		// a locking system here, such that if the same slices is already being processed in another handler, we block until
		// the other handler completes, then send down the cached session response.

		// IMPORTANT: This ensures we bill our customers only once per-slice!

		// todo: disable session locks for a test
		/*
			sliceMutexes := sessionDataReadOnly.sliceMutexes
			sliceMutex := &sliceMutexes[header.Sequence%uint64(NumSessionSliceMutexes)]
			sliceMutex.Lock()
			defer sliceMutex.Unlock()
			if header.Sequence == sessionDataReadOnly.sequence {
				if _, err := w.Write(sessionDataReadOnly.cachedResponse); err != nil {
					level.Error(params.Logger).Log("msg", "failed to write cached response", "err", err)
					params.Metrics.ErrorMetrics.WriteCachedResponseFailure.Add(1)
					return
				}
			}
		*/

		// Create the default response packet with a direct route and same SDK version as the server data.
		// This makes sure that we respond to the session update with the packet version the SDK expects.

		response := SessionResponsePacket{
			Version:              serverDataReadOnly.Version,
			Sequence:             header.Sequence,
			SessionID:            header.SessionID,
			RouteType:            int32(routing.RouteTypeDirect),
			ServerRoutePublicKey: serverDataReadOnly.RoutePublicKey,
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

		// todo: all the super manual send route responses below are really gross
		// todo: it should be much easier to send a route response. maybe package up all the data we need into a struct
		// and add some helper functions based around that struct, and common things we do below.
		// todo: having to remember to manually do the yolo each response is super dangerous as well :(

		// Set up with all data we need to make a routing decision.
		// We make copies of some read only session data here, and will store any modifications
		// we make back into the session map later on.

		params.VetoMap.RLock(header.SessionID)
		vetoReason := params.VetoMap.GetVeto(header.SessionID)
		params.VetoMap.RUnlock(header.SessionID)

		newSession := packet.Sequence == 1
		nearRelays := make([]routing.Relay, len(sessionDataReadOnly.NearRelays))
		copy(nearRelays, sessionDataReadOnly.NearRelays)
		routeExpireTimestamp := sessionDataReadOnly.RouteExpireTimestamp
		location := sessionDataReadOnly.Location
		routeDecision := sessionDataReadOnly.RouteDecision
		nextSliceCounter := sessionDataReadOnly.NextSliceCounter
		committedData := sessionDataReadOnly.CommittedData
		committedData.Committed = !buyer.RoutingRulesSettings.EnableTryBeforeYouBuy

		// Run IP2Location on the session IP address.
		// We use the lat/long to find a set of relays near the client,
		// and other information like ISP name is shown in the portal.

		timestamp := time.Now()

		if location.IsZero() {
			var err error
			location, err = params.GetIPLocator().LocateIP(packet.ClientAddress.IP)

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

				sendRouteResponse(w, &directRoute, params, &packet, &response, serverDataReadOnly, &buyer, &lastNextStats, &lastDirectStats, &location, nearRelays, routeDecision, sessionDataReadOnly.RouteDecision, sessionDataReadOnly.Initial, vetoReason, nextSliceCounter,
					committedData, sessionDataReadOnly.RouteHash, sessionDataReadOnly.RouteDecision.OnNetworkNext, timestamp, routeExpireTimestamp, sessionDataReadOnly.TokenVersion, params.RouterPrivateKey, nil, postSessionUpdateFunc) //sliceMutexes)

				return
			}
		}

		// IMPORTANT: Immediately after ip2location we *must* anonymize the IP address so there is no chance
		// we accidentally use or store the non-anonymized IP address past this point. This is an important
		// business requirement because IP addresses are considered private identifiable information according
		// to the GDRP and CCPA. We must *never* collect or store non-anonymized IP addresses!

		// todo: anonymize address should work in place instead, and not have a failure case.
		// i mean this whole code here is dead code it's never ever going to run...

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

			sendRouteResponse(w, &directRoute, params, &packet, &response, serverDataReadOnly, &buyer, &lastNextStats, &lastDirectStats, &location, nearRelays, routeDecision, sessionDataReadOnly.RouteDecision, sessionDataReadOnly.Initial, vetoReason, nextSliceCounter,
				committedData, sessionDataReadOnly.RouteHash, sessionDataReadOnly.RouteDecision.OnNetworkNext, timestamp, routeExpireTimestamp, sessionDataReadOnly.TokenVersion, params.RouterPrivateKey, nil, postSessionUpdateFunc) //sliceMutexes)
			return
		}

		// Use the route matrix to get a list of relays closest to the lat/long of the client.
		// These near relays are returned back down to the SDK for this slice. The SDK pings them
		// and reports the results back up to us in the next session update.

		routeMatrix := params.GetRouteProvider()

		if newSession {

			// If this is a new session, get the near relays from the route matrix to send down to the client.
			// Because this is an expensive function, we only want to do this on the first slice.

			if nearRelays, err = routeMatrix.GetNearRelays(location.Latitude, location.Longitude, MaxNearRelays); err != nil {
				routeDecision = routing.Decision{
					OnNetworkNext: false,
					Reason:        routing.DecisionNoNearRelays,
				}

				if buyer.RoutingRulesSettings.EnableYouOnlyLiveOnce {
					routeDecision.Reason |= routing.DecisionVetoYOLO
				}

				params.Metrics.ErrorMetrics.NearRelaysLocateFailure.Add(1)
				sendRouteResponse(w, &directRoute, params, &packet, &response, serverDataReadOnly, &buyer, &lastNextStats, &lastDirectStats, &location, nearRelays, routeDecision, sessionDataReadOnly.RouteDecision, sessionDataReadOnly.Initial, vetoReason, nextSliceCounter,
					committedData, sessionDataReadOnly.RouteHash, sessionDataReadOnly.RouteDecision.OnNetworkNext, timestamp, routeExpireTimestamp, sessionDataReadOnly.TokenVersion, params.RouterPrivateKey, nil, postSessionUpdateFunc) //, sliceMutexes)
				return
			}

		} else {

			// Update the near relay stats with the reported ping stats from the previous slice.

			for i, nearRelay := range nearRelays {
				for j, clientNearRelayID := range packet.NearRelayIDs {
					if nearRelay.ID == clientNearRelayID {
						nearRelays[i].ClientStats.RTT = float64(packet.NearRelayMinRTT[j])
						nearRelays[i].ClientStats.Jitter = float64(packet.NearRelayJitter[j])
						nearRelays[i].ClientStats.PacketLoss = float64(packet.NearRelayPacketLoss[j])
					}
				}
			}

		}

		// Fill out the near relay response to send down to the client
		// This tells the client what near relays to ping for the next 10 seconds.

		response.NumNearRelays = int32(len(nearRelays))
		response.NearRelayIDs = make([]uint64, len(nearRelays))
		response.NearRelayAddresses = make([]net.UDPAddr, len(nearRelays))
		for i := range nearRelays {
			response.NearRelayIDs[i] = nearRelays[i].ID
			response.NearRelayAddresses[i] = nearRelays[i].Addr
		}

		// Don't allow customers who aren't marked as "Live" to get a network next route. They have to pay first!

		if !buyer.Live {
			routeDecision = routing.Decision{
				OnNetworkNext: false,
				Reason:        routing.DecisionBuyerNotLive,
			}

			sendRouteResponse(w, &directRoute, params, &packet, &response, serverDataReadOnly, &buyer, &lastNextStats, &lastDirectStats, &location, nearRelays, routeDecision, sessionDataReadOnly.RouteDecision, sessionDataReadOnly.Initial, vetoReason, nextSliceCounter,
				committedData, sessionDataReadOnly.RouteHash, sessionDataReadOnly.RouteDecision.OnNetworkNext, timestamp, routeExpireTimestamp, sessionDataReadOnly.TokenVersion, params.RouterPrivateKey, nil, postSessionUpdateFunc) //, sliceMutexes
			return
		}

		// If the session has fallen back to direct, just give them a direct route.

		if packet.FallbackToDirect {
			routeDecision = routing.Decision{
				OnNetworkNext: false,
				Reason:        routing.DecisionFallbackToDirect,
			}

			sendRouteResponse(w, &directRoute, params, &packet, &response, serverDataReadOnly, &buyer, &lastNextStats, &lastDirectStats, &location, nearRelays, routeDecision, sessionDataReadOnly.RouteDecision, sessionDataReadOnly.Initial, vetoReason, nextSliceCounter,
				committedData, sessionDataReadOnly.RouteHash, sessionDataReadOnly.RouteDecision.OnNetworkNext, timestamp, routeExpireTimestamp, sessionDataReadOnly.TokenVersion, params.RouterPrivateKey, nil, postSessionUpdateFunc) //, sliceMutexes)
			return
		}

		// Has this session been vetoed? A vetoed session must always go direct, because for some reason we have found a problem
		// when they are going across network next. Perhaps we made packet loss or latency worse for this player? To make sure
		// this doesn't happen repeatedly, the session is vetoed from taking network next until they connect to a new server.

		if vetoReason != routing.DecisionNoReason {
			routeDecision = routing.Decision{
				OnNetworkNext: false,
				Reason:        vetoReason,
			}

			sendRouteResponse(w, &directRoute, params, &packet, &response, serverDataReadOnly, &buyer, &lastNextStats, &lastDirectStats, &location, nearRelays, routeDecision, sessionDataReadOnly.RouteDecision, sessionDataReadOnly.Initial, vetoReason, nextSliceCounter,
				committedData, sessionDataReadOnly.RouteHash, sessionDataReadOnly.RouteDecision.OnNetworkNext, timestamp, routeExpireTimestamp, sessionDataReadOnly.TokenVersion, params.RouterPrivateKey, nil, postSessionUpdateFunc) //sliceMutexes)

			return
		}

		// Force direct mode sends all sessions direct.
		// It's useful for disabling acceleration for a customer when something goes wrong.

		if buyer.RoutingRulesSettings.Mode == routing.ModeForceDirect {
			routeDecision = routing.Decision{
				OnNetworkNext: false,
				Reason:        routing.DecisionForceDirect,
			}

			sendRouteResponse(w, &directRoute, params, &packet, &response, serverDataReadOnly, &buyer, &lastNextStats, &lastDirectStats, &location, nearRelays, routeDecision, sessionDataReadOnly.RouteDecision, sessionDataReadOnly.Initial, vetoReason, nextSliceCounter,
				committedData, sessionDataReadOnly.RouteHash, sessionDataReadOnly.RouteDecision.OnNetworkNext, timestamp, routeExpireTimestamp, sessionDataReadOnly.TokenVersion, params.RouterPrivateKey, nil, postSessionUpdateFunc) //, sliceMutexes)
			return
		}

		// The selection percentage is used to accelerate only a certain percentage of sessions.
		// Selection percentage of 100% means all sessions are considered for acceleration.
		// Selection percentage of 10% means that only 10% of sessions are.

		if buyer.RoutingRulesSettings.Mode == routing.ModeForceDirect || (header.SessionID%100) >= uint64(buyer.RoutingRulesSettings.SelectionPercentage) {
			routeDecision = routing.Decision{
				OnNetworkNext: false,
				Reason:        routing.DecisionForceDirect,
			}

			sendRouteResponse(w, &directRoute, params, &packet, &response, serverDataReadOnly, &buyer, &lastNextStats, &lastDirectStats, &location, nearRelays, routeDecision, sessionDataReadOnly.RouteDecision, sessionDataReadOnly.Initial, vetoReason, nextSliceCounter,
				committedData, sessionDataReadOnly.RouteHash, sessionDataReadOnly.RouteDecision.OnNetworkNext, timestamp, routeExpireTimestamp, sessionDataReadOnly.TokenVersion, params.RouterPrivateKey, nil, postSessionUpdateFunc) //, sliceMutexes)
			return
		}

		// If the buyer's route shader has AB test enabled, send all odd numbered sessions direct.
		// This lets us show customers the difference between network next enabled and disabled.

		if buyer.RoutingRulesSettings.EnableABTest && header.SessionID%2 == 1 {
			routeDecision = routing.Decision{
				OnNetworkNext: false,
				Reason:        routing.DecisionABTestDirect,
			}

			sendRouteResponse(w, &directRoute, params, &packet, &response, serverDataReadOnly, &buyer, &lastNextStats, &lastDirectStats, &location, nearRelays, routeDecision, sessionDataReadOnly.RouteDecision, sessionDataReadOnly.Initial, vetoReason, nextSliceCounter,
				committedData, sessionDataReadOnly.RouteHash, sessionDataReadOnly.RouteDecision.OnNetworkNext, timestamp, routeExpireTimestamp, sessionDataReadOnly.TokenVersion, params.RouterPrivateKey, nil, postSessionUpdateFunc) //, sliceMutexes)
			return
		}

		// Get all relays in the datacenter where the server is hosted.
		// Routes are planned between the near relays for this session,
		// and the set of dest relays in the datacenter.

		datacenterRelays := routeMatrix.GetDatacenterRelays(serverDataReadOnly.Datacenter)
		if len(datacenterRelays) == 0 {
			routeDecision = routing.Decision{
				OnNetworkNext: false,
				Reason:        routing.DecisionDatacenterHasNoRelays,
			}

			if buyer.RoutingRulesSettings.EnableYouOnlyLiveOnce {
				routeDecision.Reason |= routing.DecisionVetoYOLO
			}

			params.Metrics.ErrorMetrics.NoRelaysInDatacenter.Add(1)

			params.DatacenterTracker.AddEmptyDatacenter(serverDataReadOnly.Datacenter.Name)

			sendRouteResponse(w, &directRoute, params, &packet, &response, serverDataReadOnly, &buyer, &lastNextStats, &lastDirectStats, &location, nearRelays, routeDecision, sessionDataReadOnly.RouteDecision, sessionDataReadOnly.Initial, vetoReason, nextSliceCounter,
				committedData, sessionDataReadOnly.RouteHash, sessionDataReadOnly.RouteDecision.OnNetworkNext, timestamp, routeExpireTimestamp, sessionDataReadOnly.TokenVersion, params.RouterPrivateKey, nil, postSessionUpdateFunc) //, sliceMutexes)
			return
		}

		// The first slice in a session always goes direct.

		if newSession {
			routeDecision = routing.Decision{
				OnNetworkNext: false,
				Reason:        routing.DecisionInitialSlice,
			}

			sendRouteResponse(w, &directRoute, params, &packet, &response, serverDataReadOnly, &buyer, &lastNextStats, &lastDirectStats, &location, nearRelays, routeDecision, sessionDataReadOnly.RouteDecision, sessionDataReadOnly.Initial, vetoReason, nextSliceCounter,
				committedData, sessionDataReadOnly.RouteHash, sessionDataReadOnly.RouteDecision.OnNetworkNext, timestamp, routeExpireTimestamp, sessionDataReadOnly.TokenVersion, params.RouterPrivateKey, nil, postSessionUpdateFunc) //, sliceMutexes)
			return
		}

		// Get the best route. This can be a network next route or a direct route.

		var bestRoute *routing.Route
		bestRoute, routeDecision = GetBestRoute(routeMatrix, nearRelays, datacenterRelays, &params.Metrics.ErrorMetrics, &buyer,
			sessionDataReadOnly.RouteHash, sessionDataReadOnly.RouteDecision, &lastNextStats, &lastDirectStats, nextSliceCounter, &committedData, &directRoute)

		if routeDecision.OnNetworkNext {
			nextSliceCounter++
		} else {
			nextSliceCounter = 0
		}

		// Send a session update response back to the SDK.

		sendRouteResponse(w, bestRoute, params, &packet, &response, serverDataReadOnly, &buyer, &lastNextStats, &lastDirectStats, &location, nearRelays, routeDecision, sessionDataReadOnly.RouteDecision, sessionDataReadOnly.Initial, vetoReason, nextSliceCounter,
			committedData, sessionDataReadOnly.RouteHash, sessionDataReadOnly.RouteDecision.OnNetworkNext, timestamp, routeExpireTimestamp, sessionDataReadOnly.TokenVersion, params.RouterPrivateKey, nil, postSessionUpdateFunc) //sliceMutexes)
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
		routing.DecideMultipath(buyer.RoutingRulesSettings.EnableMultipathForRTT, buyer.RoutingRulesSettings.EnableMultipathForJitter, buyer.RoutingRulesSettings.EnableMultipathForPacketLoss, float64(buyer.RoutingRulesSettings.RTTThreshold), float64(buyer.RoutingRulesSettings.MultipathPacketLossThreshold)),
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

func GetNextRoute(routeMatrix RouteProvider, nearRelays []routing.Relay, datacenterRelays []routing.Relay, errorMetrics *metrics.SessionErrorMetrics, buyer *routing.Buyer, prevRouteHash uint64) *routing.Route {

	// First, Get all possible routes between all near relays and all relays in the datacenter

	routes, err := routeMatrix.GetRoutes(nearRelays, datacenterRelays)
	if err != nil {
		errorMetrics.RouteFailure.Add(1)
		return nil
	}

	// Now pick the best route from all possible routes:

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

func CalculateNextBytesUpAndDown(envelopeKbpsUp uint64, envelopeKbpsDown uint64, sliceDuration uint64) (uint64, uint64) {
	envelopeBytesUp := (((1000 * envelopeKbpsUp) / 8) * sliceDuration)
	envelopeBytesDown := (((1000 * envelopeKbpsDown) / 8) * sliceDuration)
	return envelopeBytesUp, envelopeBytesDown
}

func CalculateTotalPriceNibblins(chosenRoute *routing.Route, envelopeBytesUp uint64, envelopeBytesDown uint64) routing.Nibblin {

	if len(chosenRoute.Relays) == 0 {
		return 0
	}

	envelopeUpGB := float64(envelopeBytesUp) / 1000000000.0
	envelopeDownGB := float64(envelopeBytesDown) / 1000000000.0

	sellerPriceNibblinsPerGB := routing.Nibblin(0)
	for _, relay := range chosenRoute.Relays {
		sellerPriceNibblinsPerGB += relay.Seller.EgressPriceNibblinsPerGB
	}

	nextPriceNibblinsPerGB := routing.Nibblin(1e9)
	totalPriceNibblins := float64(sellerPriceNibblinsPerGB+nextPriceNibblinsPerGB) * (envelopeUpGB + envelopeDownGB)

	return routing.Nibblin(totalPriceNibblins)
}

func CalculateRouteRelaysPrice(chosenRoute *routing.Route, envelopeBytesUp uint64, envelopeBytesDown uint64) []routing.Nibblin {
	if len(chosenRoute.Relays) == 0 {
		return nil
	}

	relayPrices := make([]routing.Nibblin, len(chosenRoute.Relays))

	envelopeUpGB := float64(envelopeBytesUp) / 1000000000.0
	envelopeDownGB := float64(envelopeBytesDown) / 1000000000.0

	for i, relay := range chosenRoute.Relays {
		relayPriceNibblins := float64(relay.Seller.EgressPriceNibblinsPerGB) * (envelopeUpGB + envelopeDownGB)
		relayPrices[i] = routing.Nibblin(relayPriceNibblins)
	}

	return relayPrices
}

type PostSessionUpdateParams struct {
	sessionUpdateParams *SessionUpdateParams
	packet              *SessionUpdatePacket
	response            *SessionResponsePacket
	serverDataReadOnly  *ServerData
	routeRelays         []routing.Relay
	lastNextStats       *routing.Stats
	lastDirectStats     *routing.Stats
	prevRouteDecision   routing.Decision
	location            *routing.Location
	nearRelays          []routing.Relay
	routeDecision       routing.Decision
	timeNow             time.Time
	totalPriceNibblins  routing.Nibblin
	nextRelaysPrice     []routing.Nibblin
	nextBytesUp         uint64
	nextBytesDown       uint64
	prevInitial         bool
}

type PostSessionUpdateFunc = func(params *PostSessionUpdateParams)

func goroutinePostSessionUpdateFunc(postSessionHandler *PostSessionHandler) PostSessionUpdateFunc {
	return func(params *PostSessionUpdateParams) {
		go PostSessionUpdate(postSessionHandler, params)
	}
}

func threadPoolPostSessionUpdateFunc(postSessionHandler *PostSessionHandler, pool *ants.Pool) PostSessionUpdateFunc {
	return func(params *PostSessionUpdateParams) {
		pool.Submit(func() {
			PostSessionUpdate(postSessionHandler, params)
		})
	}
}

func PostSessionUpdate(postSessionHandler *PostSessionHandler, params *PostSessionUpdateParams) {
	// IMPORTANT: we actually need to display the true datacenter name in the demo and demo plus views,
	// while in the customer view of the portal, we need to display the alias. this is because aliases will
	// shortly become per-customer, thus there is really no global concept of "multiplay.losangeles", for example.

	datacenterName := params.serverDataReadOnly.Datacenter.Name
	datacenterAlias := params.serverDataReadOnly.Datacenter.AliasName

	// Send a large amount of data to the portal via ZeroMQ to the portal cruncher.
	// This drives all the stuff you see in the portal, including the map and top sessions list.
	// We send it via ZeroMQ to the portal cruncher because google pubsub is not able to deliver data quickly enough,
	// and writing all to redis would stall the session update.

	isMultipath := routing.IsMultipath(params.prevRouteDecision)

	hops := make([]RelayHop, len(params.routeRelays))
	for i := range hops {
		hops[i] = RelayHop{
			ID:   params.routeRelays[i].ID,
			Name: params.routeRelays[i].Name,
		}
	}

	nearRelayData := make([]NearRelayPortalData, len(params.nearRelays))
	for i := range nearRelayData {
		nearRelayData[i] = NearRelayPortalData{
			ID:          params.nearRelays[i].ID,
			Name:        params.nearRelays[i].Name,
			ClientStats: params.nearRelays[i].ClientStats,
		}
	}

	postSessionData := PostSessionData{
		PortalData: buildPortalData(params.packet, params.lastNextStats, params.lastDirectStats, hops, params.packet.OnNetworkNext, datacenterName, params.location, nearRelayData, params.timeNow, isMultipath, datacenterAlias),
		PortalCountData: &SessionCountData{
			InstanceID:                params.sessionUpdateParams.InstanceID,
			TotalNumDirectSessions:    params.sessionUpdateParams.SessionMap.GetDirectSessionCount(),
			TotalNumNextSessions:      params.sessionUpdateParams.SessionMap.GetNextSessionCount(),
			NumDirectSessionsPerBuyer: params.sessionUpdateParams.SessionMap.GetDirectSessionCountPerBuyer(),
			NumNextSessionsPerBuyer:   params.sessionUpdateParams.SessionMap.GetNextSessionCountPerBuyer(),
		},
		BillingEntry: buildBillingEntry(params),
	}

	postSessionHandler.Send(&postSessionData)
}

func buildPortalData(packet *SessionUpdatePacket, lastNNStats *routing.Stats, lastDirectStats *routing.Stats, relayHops []RelayHop,
	onNetworkNext bool, datacenterName string, location *routing.Location, nearRelays []NearRelayPortalData, sessionTime time.Time, isMultiPath bool, datacenterAlias string) *SessionPortalData {

	var hashedID uint64
	if !packet.Version.IsInternal() && packet.Version.Compare(SDKVersion{3, 4, 5}) == SDKVersionOlder {
		hash := fnv.New64a()
		byteArray := make([]byte, 8)
		binary.LittleEndian.PutUint64(byteArray, packet.UserHash)
		hash.Write(byteArray)
		hashedID = hash.Sum64()
	} else {
		hashedID = packet.UserHash
	}

	var deltaRTT float64
	if !onNetworkNext {
		deltaRTT = 0
	} else {
		deltaRTT = lastDirectStats.RTT - lastNNStats.RTT
	}

	return &SessionPortalData{
		Meta: SessionMeta{
			ID:              packet.SessionID,
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
			Connection:      uint8(packet.ConnectionType),
			NearbyRelays:    nearRelays,
			Platform:        uint8(packet.PlatformID),
			BuyerID:         packet.CustomerID,
		},
		Slice: SessionSlice{
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
		},
		Point: SessionMapPoint{
			Latitude:      location.Latitude,
			Longitude:     location.Longitude,
			OnNetworkNext: onNetworkNext,
		},
	}
}

func buildBillingEntry(params *PostSessionUpdateParams) *billing.BillingEntry {
	isMultipath := routing.IsMultipath(params.prevRouteDecision)

	nextRelays := [billing.BillingEntryMaxRelays]uint64{}
	for i := 0; i < len(params.routeRelays) && i < len(nextRelays); i++ {
		nextRelays[i] = params.routeRelays[i].ID
	}

	nextRelaysPriceArray := [billing.BillingEntryMaxRelays]uint64{}
	for i := 0; i < len(nextRelaysPriceArray) && i < len(params.nextRelaysPrice); i++ {
		nextRelaysPriceArray[i] = uint64(params.nextRelaysPrice[i])
	}

	return &billing.BillingEntry{
		BuyerID:                   params.packet.CustomerID,
		UserHash:                  params.packet.UserHash,
		SessionID:                 params.packet.SessionID,
		SliceNumber:               uint32(params.packet.Sequence),
		DirectRTT:                 float32(params.lastDirectStats.RTT),
		DirectJitter:              float32(params.lastDirectStats.Jitter),
		DirectPacketLoss:          float32(params.lastDirectStats.PacketLoss),
		Next:                      params.packet.OnNetworkNext,
		NextRTT:                   float32(params.lastNextStats.RTT),
		NextJitter:                float32(params.lastNextStats.Jitter),
		NextPacketLoss:            float32(params.lastNextStats.PacketLoss),
		NumNextRelays:             uint8(len(params.routeRelays)),
		NextRelays:                nextRelays,
		TotalPrice:                uint64(params.totalPriceNibblins),
		ClientToServerPacketsLost: params.packet.PacketsLostClientToServer,
		ServerToClientPacketsLost: params.packet.PacketsLostServerToClient,
		Committed:                 params.packet.Committed,
		Flagged:                   params.packet.Flagged,
		Multipath:                 isMultipath,
		Initial:                   params.prevInitial,
		NextBytesUp:               params.nextBytesUp,
		NextBytesDown:             params.nextBytesDown,
		DatacenterID:              params.serverDataReadOnly.Datacenter.ID,
		RTTReduction:              params.prevRouteDecision.Reason&routing.DecisionRTTReduction != 0 || params.prevRouteDecision.Reason&routing.DecisionRTTReductionMultipath != 0,
		PacketLossReduction:       params.prevRouteDecision.Reason&routing.DecisionHighPacketLossMultipath != 0,
		NextRelaysPrice:           nextRelaysPriceArray,
	}
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
	case routing.DecisionBuyerNotLive:
		m.DecisionMetrics.BuyerNotLive.Add(1)
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

func marshalResponse(response *SessionResponsePacket, privateKey []byte) ([]byte, error) {
	// Sign the response
	response.Signature = crypto.Sign(privateKey, response.GetSignData())

	// Marshal the packet
	responseData, err := response.MarshalBinary()
	if err != nil {
		return nil, err
	}

	return responseData, nil
}

func sendRouteResponse(w io.Writer, chosenRoute *routing.Route, params *SessionUpdateParams, packet *SessionUpdatePacket, response *SessionResponsePacket, serverDataReadOnly *ServerData,
	buyer *routing.Buyer, lastNextStats *routing.Stats, lastDirectStats *routing.Stats, location *routing.Location, nearRelays []routing.Relay, routeDecision routing.Decision, prevRouteDecision routing.Decision, prevInitial bool, vetoReason routing.DecisionReason,
	onNNSliceCounter uint64, committedData routing.CommittedData, prevRouteHash uint64, prevOnNetworkNext bool, timeNow time.Time, routeExpireTimestamp uint64, tokenVersion uint8, routerPrivateKey []byte, sliceMutexes []sync.Mutex, postSessionUpdateFunc PostSessionUpdateFunc) {
	// Update response data
	{
		if committedData.Committed {
			response.Committed = true
		}

		if routing.IsMultipath(routeDecision) {
			response.Multipath = true
			response.Committed = true // Always commit to multipath routes
		}
	}

	if routeExpireTimestamp == 0 {
		routeExpireTimestamp = uint64(timeNow.Unix())
	}

	routeExpireTimestamp += billing.BillingSliceSeconds

	// Tokenize the route
	if routeDecision.OnNetworkNext {
		var token routing.Token

		if chosenRoute.Hash64() == prevRouteHash {
			token = &routing.ContinueRouteToken{
				Expires: routeExpireTimestamp,

				SessionID: packet.SessionID,

				SessionVersion: tokenVersion,

				Client: routing.Client{
					Addr:      packet.ClientAddress,
					PublicKey: packet.ClientRoutePublicKey,
				},

				Server: routing.Server{
					Addr:      packet.ServerAddress,
					PublicKey: serverDataReadOnly.RoutePublicKey,
				},

				Relays: chosenRoute.Relays,
			}
		} else {
			tokenVersion++
			routeExpireTimestamp += billing.BillingSliceSeconds // Add another slice duration for a new network next route
			token = &routing.NextRouteToken{
				Expires: routeExpireTimestamp,

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
					PublicKey: serverDataReadOnly.RoutePublicKey,
				},

				Relays: chosenRoute.Relays,
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

	responseData, err := marshalResponse(response, params.ServerPrivateKey)
	if err != nil {
		level.Error(params.Logger).Log("msg", "could not marshal session update response packet", "err", err)
		params.Metrics.ErrorMetrics.MarshalResponseFailure.Add(1)
		return
	}

	// Update the session data
	session := SessionData{
		Timestamp:            timeNow.Unix(),
		BuyerID:              buyer.ID,
		Location:             *location,
		Sequence:             packet.Sequence,
		NearRelays:           nearRelays,
		RouteHash:            chosenRoute.Hash64(),
		Initial:              response.RouteType == routing.RouteTypeNew,
		RouteDecision:        routeDecision,
		NextSliceCounter:     onNNSliceCounter,
		CommittedData:        committedData,
		RouteExpireTimestamp: routeExpireTimestamp,
		TokenVersion:         tokenVersion,
		CachedResponse:       responseData,
		SliceMutexes:         sliceMutexes,
	}
	params.SessionMap.UpdateSessionData(packet.SessionID, &session)

	// If the session was vetoed this slice, update the veto data
	if routing.IsVetoed(routeDecision) && vetoReason == routing.DecisionNoReason {
		params.VetoMap.SetVeto(packet.SessionID, routeDecision.Reason)
	}

	if response.RouteType == routing.RouteTypeDirect {
		params.Metrics.DirectSessions.Add(1)
	} else {
		params.Metrics.NextSessions.Add(1)
	}

	addRouteDecisionMetric(routeDecision, params.Metrics)

	// If the last slice was newly on NN, then we want to extend the slice duration to 20 seconds
	// so that we calculate the usage and envelope bytes correctly.
	lastSliceDuration := uint64(billing.BillingSliceSeconds)
	if prevInitial {
		lastSliceDuration *= 2
	}

	usageBytesUp, usageBytesDown := CalculateNextBytesUpAndDown(uint64(packet.KbpsUp), uint64(packet.KbpsDown), lastSliceDuration)
	envelopeBytesUp, envelopeBytesDown := CalculateNextBytesUpAndDown(uint64(buyer.RoutingRulesSettings.EnvelopeKbpsUp), uint64(buyer.RoutingRulesSettings.EnvelopeKbpsDown), lastSliceDuration)

	// Calculate the total price for the billing entry
	totalPriceNibblins := CalculateTotalPriceNibblins(chosenRoute, envelopeBytesUp, envelopeBytesDown)

	nextRelaysPrice := CalculateRouteRelaysPrice(chosenRoute, envelopeBytesUp, envelopeBytesDown)

	// IMPORTANT: run post in parallel so it doesn't block the response
	postSessionUpdateFunc(&PostSessionUpdateParams{
		sessionUpdateParams: params,
		packet:              packet,
		response:            response,
		serverDataReadOnly:  serverDataReadOnly,
		routeRelays:         chosenRoute.Relays,
		lastNextStats:       lastNextStats,
		lastDirectStats:     lastDirectStats,
		prevRouteDecision:   prevRouteDecision,
		location:            location,
		nearRelays:          nearRelays,
		routeDecision:       routeDecision,
		timeNow:             timeNow,
		totalPriceNibblins:  totalPriceNibblins,
		nextRelaysPrice:     nextRelaysPrice,
		nextBytesUp:         usageBytesUp,
		nextBytesDown:       usageBytesDown,
		prevInitial:         prevInitial,
	})

	// Send the Session Response back to the server
	if _, err := w.Write(responseData); err != nil {
		level.Error(params.Logger).Log("msg", "could not write session update response packet", "err", err)
		params.Metrics.ErrorMetrics.WriteResponseFailure.Add(1)
		return
	}
}
