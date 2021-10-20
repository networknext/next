package fake_server

import (
	"context"
	"crypto/ed25519"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/transport"
)

// FakeServer represents a single fake server that simulates connected fake clients.
// It mocks server init, server update, and session update requests to load test the server backend.
type FakeServer struct {
	sdkVersion           transport.SDKVersion
	buyerID              uint64
	customerPrivateKey   []byte
	publicAddress        *net.UDPAddr
	serverRoutePublicKey []byte
	dcName               string
	sendBeaconPackets    bool

	conn              *net.UDPConn
	serverBackendAddr *net.UDPAddr
	beaconAddr        *net.UDPAddr
	sessions          []Session
}

// NewFakeServer returns a fake server with the given parameters.
func NewFakeServer(conn *net.UDPConn, serverBackendAddr *net.UDPAddr, beaconAddr *net.UDPAddr, clientCount int, sdkVersion transport.SDKVersion, buyerID uint64, customerPrivateKey []byte, dcName string, sendBeaconPackets bool) (*FakeServer, error) {
	// We need to use a random address for the server so that
	// each server instance is uniquely identifiable, so that
	// the total session count is accurate.
	// The server backend will still send the responses back to the address it came from.
	randIPBytes := make([]byte, 0)

	for i := 0; i < 4; i++ {
		randIPBytes = append(randIPBytes, byte(rand.Intn(255)))
	}

	randPort := rand.Intn(65536)

	randomAddress := net.UDPAddr{
		IP:   net.IPv4(randIPBytes[0], randIPBytes[1], randIPBytes[2], randIPBytes[3]),
		Port: randPort,
	}

	routePublicKey, _, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, err
	}

	server := &FakeServer{
		sdkVersion:           sdkVersion,
		publicAddress:        &randomAddress,
		buyerID:              buyerID,
		customerPrivateKey:   customerPrivateKey,
		serverRoutePublicKey: routePublicKey,
		dcName:               dcName,
		sendBeaconPackets:    sendBeaconPackets,
		sessions:             make([]Session, clientCount),
		conn:                 conn,
		serverBackendAddr:    serverBackendAddr,
		beaconAddr:           beaconAddr,
	}

	return server, nil
}

// StartLoop starts sending and receiving packets to and from the server backend.
// This function blocks, so call it in a separate goroutine.
func (server *FakeServer) StartLoop(ctx context.Context, updateRate time.Duration, readBufferSize int, writeBufferSize int) error {
	for i := 0; i < len(server.sessions); i++ {
		session, err := NewSession()
		if err != nil {
			return err
		}

		server.sessions[i] = session
	}

	if err := server.sendServerInitPacket(); err != nil {
		return err
	}

	ticker := time.NewTicker(updateRate)
	defer ticker.Stop()

	if err := server.update(); err != nil {
		return err
	}

	for {
		select {
		case <-ticker.C:
			if err := server.update(); err != nil {
				return err
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func (server *FakeServer) update() error {
	if err := server.sendServerUpdatePacket(); err != nil {
		return err
	}

	for i := range server.sessions {
		if time.Since(server.sessions[i].startTime) > server.sessions[i].duration {
			core.Debug("session %016x expired", server.sessions[i].sessionID)

			// Send session update with client ping timeout to write summary slice
			var err error
			_, err = server.sendSessionUpdatePacket(server.sessions[i], true)
			if err != nil {
				return err
			}

			server.sessions[i], err = NewSession()
			if err != nil {
				return err
			}
		}

		if server.sendBeaconPackets {
			if err := server.sendBeaconPacket(server.sessions[i]); err != nil {
				return err
			}
		}

		responsePacket, err := server.sendSessionUpdatePacket(server.sessions[i], false)
		if err != nil {
			return err
		}

		server.sessions[i].Advance(responsePacket)
	}

	return nil
}

// sendServerInitPacket is responsible for sending the server init packet and receiving the response.
func (server *FakeServer) sendServerInitPacket() error {
	requestPacket := transport.ServerInitRequestPacket{
		Version:        server.sdkVersion,
		BuyerID:        server.buyerID,
		DatacenterID:   crypto.HashID(server.dcName),
		RequestID:      rand.Uint64(),
		DatacenterName: server.dcName,
	}

	packetData, err := transport.MarshalPacket(&requestPacket)
	if err != nil {
		return err
	}

	if err := server.sendPacket(transport.PacketTypeServerInitRequest, packetData); err != nil {
		return err
	}

	core.Debug("sent server init request packet")

	incomingPacketType, incomingPacketData, err := server.readPacket()
	if err != nil {
		return err
	}

	if incomingPacketType != transport.PacketTypeServerInitResponse {
		return fmt.Errorf("received non server init response packet type: ID %d", incomingPacketType)
	}

	initResponse := &transport.ServerInitResponsePacket{}
	if err := transport.UnmarshalPacket(initResponse, incomingPacketData); err != nil {
		return err
	}

	if initResponse.Response != transport.InitResponseOK {
		return fmt.Errorf("failed to init: received a server init response of %d", initResponse.Response)
	}

	core.Debug("received OK server init response")
	return nil
}

// sendServerUpdatePacket is responsible for sending the server update packet.
func (server *FakeServer) sendServerUpdatePacket() error {
	requestPacket := transport.ServerUpdatePacket{
		Version:       server.sdkVersion,
		BuyerID:       server.buyerID,
		DatacenterID:  crypto.HashID(server.dcName),
		NumSessions:   uint32(len(server.sessions)),
		ServerAddress: *server.publicAddress,
	}

	packetData, err := transport.MarshalPacket(&requestPacket)
	if err != nil {
		return err
	}

	if err := server.sendPacket(transport.PacketTypeServerUpdate, packetData); err != nil {
		return err
	}

	core.Debug("sent server update packet")
	return nil
}

// sendSessionUpdatePacket is responsible for sending the session update request packet and receiving the response.
func (server *FakeServer) sendSessionUpdatePacket(session Session, clientPingTimedOut bool) (transport.SessionResponsePacket, error) {
	var sessionResponse transport.SessionResponsePacket
	sessionResponse.Version = server.sdkVersion

	requestPacket := transport.SessionUpdatePacket{
		Version:                         server.sdkVersion,
		BuyerID:                         server.buyerID,
		DatacenterID:                    crypto.HashID(server.dcName),
		SessionID:                       session.sessionID,
		SliceNumber:                     session.sliceNumber,
		SessionDataBytes:                session.sessionDataBytes,
		SessionData:                     session.sessionData,
		ClientAddress:                   session.clientAddress,
		ServerAddress:                   *server.publicAddress,
		ClientRoutePublicKey:            session.clientRoutePublicKey,
		ServerRoutePublicKey:            server.serverRoutePublicKey,
		UserHash:                        session.userHash,
		PlatformType:                    session.platformType,
		ConnectionType:                  session.connectionType,
		Next:                            session.next,
		Committed:                       session.committed,
		Reported:                        false,
		FallbackToDirect:                false,
		ClientBandwidthOverLimit:        false,
		ServerBandwidthOverLimit:        false,
		ClientPingTimedOut:              clientPingTimedOut,
		NumTags:                         0,
		Tags:                            [transport.MaxTags]uint64{},
		Flags:                           0,
		UserFlags:                       0,
		DirectMinRTT:                    session.directRTT, // todo: might be worth upgrading session to have support for min/max/prime direct RTT?
		DirectMaxRTT:                    session.directRTT,
		DirectPrimeRTT:                  session.directRTT,
		DirectJitter:                    session.directJitter,
		DirectPacketLoss:                session.directPacketLoss,
		NextRTT:                         session.nextRTT,
		NextJitter:                      session.nextJitter,
		NextPacketLoss:                  session.nextPacketLoss,
		NumNearRelays:                   session.numNearRelays,
		NearRelayIDs:                    session.nearRelayIDs,
		NearRelayRTT:                    session.nearRelayRTT,
		NearRelayJitter:                 session.nearRelayJitter,
		NearRelayPacketLoss:             session.nearRelayPacketLoss,
		NextKbpsUp:                      session.nextKbpsUp,
		NextKbpsDown:                    session.nextKbpsDown,
		PacketsSentClientToServer:       session.packetsSent,
		PacketsSentServerToClient:       session.packetsSent,
		PacketsLostClientToServer:       session.packetsLost,
		PacketsLostServerToClient:       session.packetsLost,
		PacketsOutOfOrderClientToServer: 0,
		PacketsOutOfOrderServerToClient: 0,
		JitterClientToServer:            session.jitter,
		JitterServerToClient:            session.jitter,
	}

	packetData, err := transport.MarshalPacket(&requestPacket)
	if err != nil {
		return sessionResponse, err
	}

	if err := server.sendPacket(transport.PacketTypeSessionUpdate, packetData); err != nil {
		return sessionResponse, err
	}

	core.Debug("sent session update request packet")

	if clientPingTimedOut {
		// If the session is expired, no need to receive the response packet
		return transport.SessionResponsePacket{}, nil
	}

	incomingPacketType, incomingPacketData, err := server.readPacket()
	if err != nil {
		return sessionResponse, err
	}

	if incomingPacketType != transport.PacketTypeSessionResponse {
		return sessionResponse, fmt.Errorf("received non session update response packet type: ID %d", incomingPacketType)
	}

	if err := transport.UnmarshalPacket(&sessionResponse, incomingPacketData); err != nil {
		return sessionResponse, err
	}

	core.Debug("received session update response")

	switch sessionResponse.RouteType {
	case routing.RouteTypeDirect:
		core.Debug("session %016x - taking direct route", session.sessionID)

	case routing.RouteTypeNew:
		core.Debug("session %016x - taking network next route", session.sessionID)

	case routing.RouteTypeContinue:
		core.Debug("session %016x - continuing network next route", session.sessionID)
	}

	return sessionResponse, nil
}

// sendPacket sends a given packet type to the server backend.
func (server *FakeServer) sendPacket(packetType byte, packetData []byte) error {
	packetDataHeader := make([]byte, 1+crypto.PacketHashSize)
	packetDataHeader[0] = packetType
	packetData = append(packetDataHeader, packetData...)

	packetData = crypto.SignPacket(server.customerPrivateKey[8:], packetData)
	crypto.HashPacket(crypto.PacketHashKey, packetData)

	if err := server.conn.SetWriteDeadline(time.Now().Add(time.Second * 10)); err != nil {
		return err
	}

	if _, err := server.conn.WriteToUDP(packetData, server.serverBackendAddr); err != nil {
		return fmt.Errorf("failed to write to UDP: %v", err)
	}

	return nil
}

// sendBeaconPacket sends a packet to the beacon service.
func (server *FakeServer) sendBeaconPacket(session Session) error {
	beaconPacket := transport.NextBeaconPacket{
		Version:          uint32(transport.BeaconPacketVersion),
		BuyerID:          server.buyerID,
		DatacenterID:     crypto.HashID(server.dcName),
		UserHash:         session.userHash,
		AddressHash:      crypto.HashID(session.clientAddress.String()),
		SessionID:        session.sessionID,
		PlatformID:       session.platformType,
		ConnectionType:   session.connectionType,
		Enabled:          true, // We never have network next disabled locally or in staging
		Upgraded:         session.upgraded,
		Next:             session.next,
		FallbackToDirect: false,
	}

	packetData, err := transport.WriteBeaconEntry(&beaconPacket)
	if err != nil {
		return err
	}

	packetDataHeader := make([]byte, 1)
	packetDataHeader[0] = transport.PacketTypeBeacon
	packetData = append(packetDataHeader, packetData...)

	if err := server.conn.SetWriteDeadline(time.Now().Add(time.Second * 10)); err != nil {
		return err
	}

	if _, err := server.conn.WriteToUDP(packetData, server.beaconAddr); err != nil {
		return fmt.Errorf("failed to write to UDP: %v", err)
	}

	return nil
}

// readPacket receives a packet from the server backend.
func (server FakeServer) readPacket() (byte, []byte, error) {
	if err := server.conn.SetReadDeadline(time.Now().Add(time.Second * 10)); err != nil {
		return 0, nil, err
	}

	incomingPacketDataArray := [transport.DefaultMaxPacketSize]byte{}
	incomingPacketData := incomingPacketDataArray[:]

	n, _, err := server.conn.ReadFromUDP(incomingPacketData)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to read from UDP: %v", err)
	}

	if n <= 0 {
		return 0, nil, errors.New("failed to read from UDP: read packet with no size")
	}

	incomingPacketData = incomingPacketDataArray[:n]

	if !crypto.IsNetworkNextPacket(crypto.PacketHashKey, incomingPacketData) {
		return 0, nil, errors.New("received non network next packet")
	}

	incomingPacketType := incomingPacketData[0]
	incomingPacketData = incomingPacketData[crypto.PacketHashSize+1 : n]

	return incomingPacketType, incomingPacketData, nil
}
