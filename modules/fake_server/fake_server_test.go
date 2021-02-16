package fake_server

import (
	"context"
	"encoding/binary"
	"net"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/transport"
	"github.com/stretchr/testify/assert"
)

func createExpectedFakeServer(t *testing.T) (*FakeServer, net.Conn) {
	_, privateKey, err := crypto.GenerateCustomerKeyPair()
	assert.NoError(t, err)

	customerID := binary.LittleEndian.Uint64(privateKey[:8])
	privateKey = privateKey[8:]

	serverConn, backendConn := net.Pipe()

	expectedServer := FakeServer{
		sdkVersion:         transport.SDKVersionMin,
		customerID:         customerID,
		customerPrivateKey: privateKey,
		logger:             log.NewNopLogger(),
		sessions:           make([]Session, 100),
		conn:               serverConn,
	}

	return &expectedServer, backendConn
}

func runTestServerBackend(t *testing.T, backendConn net.Conn) {
	buffer := make([]byte, transport.DefaultMaxPacketSize)

	for {
		_, err := backendConn.Read(buffer)

		switch buffer[0] {
		case transport.PacketTypeServerInitRequest:
			responsePacket := transport.ServerInitResponsePacket{
				RequestID: 0,
				Response:  transport.InitResponseOK,
			}

			response, err := transport.MarshalPacket(&responsePacket)
			assert.NoError(t, err)

			response = append([]byte{transport.PacketTypeServerInitResponse, 0, 0, 0, 0, 0, 0, 0, 0}, response...)

			backendPrivateKey := [32]byte{}
			response = crypto.SignPacket(backendPrivateKey[:], response)
			crypto.HashPacket(crypto.PacketHashKey, response)

			_, err = backendConn.Write(response)
			assert.NoError(t, err)

		case transport.PacketTypeSessionUpdate:
			responsePacket := transport.SessionResponsePacket{}

			response, err := transport.MarshalPacket(&responsePacket)
			assert.NoError(t, err)

			response = append([]byte{transport.PacketTypeSessionResponse, 0, 0, 0, 0, 0, 0, 0, 0}, response...)

			backendPrivateKey := [32]byte{}
			response = crypto.SignPacket(backendPrivateKey[:], response)
			crypto.HashPacket(crypto.PacketHashKey, response)

			_, err = backendConn.Write(response)
			assert.NoError(t, err)
		}

		assert.NoError(t, err)
	}
}

func TestNewFakeServer(t *testing.T) {
	expectedServer, _ := createExpectedFakeServer(t)

	actualServer, err := NewFakeServer(expectedServer.conn, len(expectedServer.sessions), expectedServer.sdkVersion, expectedServer.logger, expectedServer.customerID, expectedServer.customerPrivateKey)
	assert.NoError(t, err)
	assert.Equal(t, expectedServer.sdkVersion, actualServer.sdkVersion)
	assert.Equal(t, expectedServer.customerID, actualServer.customerID)
	assert.Equal(t, expectedServer.customerPrivateKey, actualServer.customerPrivateKey)
	assert.NotEmpty(t, actualServer.publicAddress)
	assert.Equal(t, expectedServer.logger, actualServer.logger)
	assert.NotEmpty(t, actualServer.serverRoutePublicKey)
	assert.Equal(t, expectedServer.sessions, actualServer.sessions)
	assert.Equal(t, expectedServer.conn, actualServer.conn)
}

func TestStartLoop(t *testing.T) {
	server, backendConn := createExpectedFakeServer(t)

	server, err := NewFakeServer(server.conn, len(server.sessions), server.sdkVersion, server.logger, server.customerID, server.customerPrivateKey)
	assert.NoError(t, err)

	go runTestServerBackend(t, backendConn)

	ctx, cancelFunc := context.WithDeadline(context.Background(), time.Now().Add(time.Millisecond*15))
	defer cancelFunc()

	err = server.StartLoop(ctx, time.Millisecond*10, 0, 0)
	assert.NoError(t, err)
}

func TestUpdate(t *testing.T) {
	server, backendConn := createExpectedFakeServer(t)

	server, err := NewFakeServer(server.conn, len(server.sessions), server.sdkVersion, server.logger, server.customerID, server.customerPrivateKey)
	assert.NoError(t, err)

	go runTestServerBackend(t, backendConn)

	err = server.update()
	assert.NoError(t, err)
}

func TestSendServerInitPacket(t *testing.T) {
	server, backendConn := createExpectedFakeServer(t)

	server, err := NewFakeServer(server.conn, len(server.sessions), server.sdkVersion, server.logger, server.customerID, server.customerPrivateKey)
	assert.NoError(t, err)

	t.Run("failed to marshal request", func(t *testing.T) {

	})

	t.Run("failed to send request", func(t *testing.T) {

	})

	t.Run("failed to read response", func(t *testing.T) {

	})

	t.Run("wrong response", func(t *testing.T) {

	})

	t.Run("failed to unmarshal response", func(t *testing.T) {

	})

	t.Run("error response", func(t *testing.T) {

	})

	t.Run("success", func(t *testing.T) {
		go runTestServerBackend(t, backendConn)

		err = server.sendServerInitPacket()
		assert.NoError(t, err)
	})
}
