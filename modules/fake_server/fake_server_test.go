package fake_server

import (
	"context"
	"encoding/binary"
	"net"
	"testing"
	// "time"

	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/encoding"
	"github.com/networknext/backend/modules/transport"
	"github.com/stretchr/testify/assert"
)

const (
	testSendNormalResponse        = 0
	testSendInvalidResponse       = 1
	testSendMismatchedResponse    = 2
	testSendUnmarshalableResponse = 3
	testSendInitErrorResponse     = 4
)

func createExpectedFakeServer(t *testing.T) (*FakeServer, *net.UDPConn, *net.UDPConn) {
	_, privateKey, err := crypto.GenerateCustomerKeyPair()
	assert.NoError(t, err)

	buyerID := binary.LittleEndian.Uint64(privateKey[:8])
	privateKey = privateKey[8:]

	lc := net.ListenConfig{}

	lp, err := lc.ListenPacket(context.Background(), "udp", "0.0.0.0:0")
	assert.NoError(t, err)

	conn := lp.(*net.UDPConn)

	serverBackendLP, err := lc.ListenPacket(context.Background(), "udp", "0.0.0.0:0")
	assert.NoError(t, err)

	serverBackendConn := serverBackendLP.(*net.UDPConn)
	serverBackendAddr, err := net.ResolveUDPAddr("udp", serverBackendConn.LocalAddr().String())
	assert.NoError(t, err)

	beaconLP, err := lc.ListenPacket(context.Background(), "udp", "0.0.0.0:0")
	assert.NoError(t, err)

	beaconConn := beaconLP.(*net.UDPConn)
	beaconAddr, err := net.ResolveUDPAddr("udp", beaconConn.LocalAddr().String())
	assert.NoError(t, err)

	expectedServer := FakeServer{
		sdkVersion:         transport.SDKVersionLatest,
		buyerID:            buyerID,
		customerPrivateKey: privateKey,
		sessions:           make([]Session, 100),
		conn:               conn,
		serverBackendAddr:  serverBackendAddr,
		beaconAddr:         beaconAddr,
		dcName:             "local",
		sendBeaconPackets:  false,
	}

	return &expectedServer, serverBackendConn, beaconConn
}

func testInitResponse(t *testing.T, responseType uint32) []byte {
	responsePacket := transport.ServerInitResponsePacket{
		RequestID: 0,
		Response:  responseType,
	}

	response, err := transport.MarshalPacket(&responsePacket)
	assert.NoError(t, err)

	responseHeader := make([]byte, 1+crypto.PacketHashSize)
	responseHeader[0] = transport.PacketTypeServerInitResponse
	response = append(responseHeader, response...)

	backendPrivateKey := [32]byte{}
	response = crypto.SignPacket(backendPrivateKey[:], response)
	crypto.HashPacket(crypto.PacketHashKey, response)

	return response
}

func testSessionResponse(t *testing.T) []byte {
	responsePacket := transport.SessionResponsePacket{}

	response, err := transport.MarshalPacket(&responsePacket)
	assert.NoError(t, err)

	responseHeader := make([]byte, 1+crypto.PacketHashSize)
	responseHeader[0] = transport.PacketTypeSessionResponse
	response = append(responseHeader, response...)

	backendPrivateKey := [32]byte{}
	response = crypto.SignPacket(backendPrivateKey[:], response)
	crypto.HashPacket(crypto.PacketHashKey, response)

	return response
}

func runTestServerBackend(t *testing.T, ctx context.Context, backendConn *net.UDPConn, sendResponse int, backendRecvReady chan struct{}) {
	buffer := make([]byte, transport.DefaultMaxPacketSize)

	for {
		select {
		case <-ctx.Done():
			backendConn.Close()
			return
		default:
			backendRecvReady <- struct{}{}

			_, fromAddr, err := backendConn.ReadFromUDP(buffer)

			var response []byte

			switch sendResponse {
			case testSendInvalidResponse:
				response = []byte("bad data")

				_, err = backendConn.WriteToUDP(response, fromAddr)
				assert.NoError(t, err)
				return

			case testSendMismatchedResponse:
				switch buffer[0] {
				case transport.PacketTypeServerInitRequest:
					response = testSessionResponse(t)

				case transport.PacketTypeSessionUpdate:
					response = testInitResponse(t, transport.InitResponseOK)
				}

			case testSendUnmarshalableResponse:
				switch buffer[0] {
				case transport.PacketTypeServerInitRequest:
					response = make([]byte, 2+crypto.PacketHashSize) // We need to have at least 1 byte in the message, otherwise crypto.HashPacket will panic
					response[0] = transport.PacketTypeServerInitResponse
					crypto.HashPacket(crypto.PacketHashKey, response)

				case transport.PacketTypeSessionUpdate:
					response = make([]byte, 2+crypto.PacketHashSize) // We need to have at least 1 byte in the message, otherwise crypto.HashPacket will panic
					response[0] = transport.PacketTypeSessionResponse
					crypto.HashPacket(crypto.PacketHashKey, response)
				}

			case testSendInitErrorResponse:
				switch buffer[0] {
				case transport.PacketTypeServerInitRequest:
					response = testInitResponse(t, transport.InitResponseUnknownBuyer)

				case transport.PacketTypeSessionUpdate:
					response = testSessionResponse(t)
				}

			default:
				switch buffer[0] {
				case transport.PacketTypeServerInitRequest:
					response = testInitResponse(t, transport.InitResponseOK)

				case transport.PacketTypeSessionUpdate:
					response = testSessionResponse(t)
				}
			}

			if response != nil {
				_, err = backendConn.WriteToUDP(response, fromAddr)
				assert.NoError(t, err)
			}
		}
	}
}

func runTestBeacon(t *testing.T, ctx context.Context, beaconConn *net.UDPConn, beaconRecvReady chan struct{}) {
	dataArray := [transport.DefaultMaxPacketSize]byte{}

	var beaconPacket *transport.NextBeaconPacket

	for {
		select {
		case <-ctx.Done():
			beaconConn.Close()
			return
		default:
			beaconPacket = &transport.NextBeaconPacket{}
			data := dataArray[:]

			beaconRecvReady <- struct{}{}

			size, _, err := beaconConn.ReadFromUDP(data)
			assert.NoError(t, err)

			// Ensure the packet size is more than 1
			assert.False(t, size <= 1)

			data = data[:size]

			// Ensure we received a beacon packet
			assert.True(t, data[0] == transport.PacketTypeBeacon)

			readStream := encoding.CreateReadStream(data[1:])
			err = beaconPacket.Serialize(readStream)
			assert.NoError(t, err)
		}
	}
}

func TestNewFakeServer(t *testing.T) {
	expectedServer, _, _ := createExpectedFakeServer(t)

	actualServer, err := NewFakeServer(expectedServer.conn, expectedServer.serverBackendAddr, expectedServer.beaconAddr, len(expectedServer.sessions), expectedServer.sdkVersion, expectedServer.buyerID, expectedServer.customerPrivateKey, expectedServer.dcName, expectedServer.sendBeaconPackets)
	assert.NoError(t, err)
	assert.Equal(t, expectedServer.sdkVersion, actualServer.sdkVersion)
	assert.Equal(t, expectedServer.buyerID, actualServer.buyerID)
	assert.Equal(t, expectedServer.customerPrivateKey, actualServer.customerPrivateKey)
	assert.NotEmpty(t, actualServer.publicAddress)
	assert.NotEmpty(t, actualServer.serverRoutePublicKey)
	assert.Equal(t, expectedServer.sessions, actualServer.sessions)
	assert.Equal(t, expectedServer.conn, actualServer.conn)
	assert.Equal(t, expectedServer.serverBackendAddr, actualServer.serverBackendAddr)
	assert.Equal(t, expectedServer.beaconAddr, actualServer.beaconAddr)
	assert.Equal(t, expectedServer.dcName, actualServer.dcName)
	assert.Equal(t, expectedServer.sendBeaconPackets, actualServer.sendBeaconPackets)
}

// func TestStartLoop(t *testing.T) {
// 	server, backendConn, _ := createExpectedFakeServer(t)

// 	server, err := NewFakeServer(server.conn, server.serverBackendAddr, server.beaconAddr, len(server.sessions), server.sdkVersion, server.buyerID, server.customerPrivateKey, server.dcName, server.sendBeaconPackets)
// 	assert.NoError(t, err)

// 	backendRecvReady := make(chan struct{})

// 	ctx, cancelFunc := context.WithDeadline(context.Background(), time.Now().Add(time.Millisecond*200))
// 	defer cancelFunc()

// 	go runTestServerBackend(t, ctx, backendConn, testSendNormalResponse, backendRecvReady)
// 	<-backendRecvReady

// 	err = server.StartLoop(ctx, time.Millisecond*10, 0, 0)
// 	assert.NoError(t, err)
// }

// func TestUpdate(t *testing.T) {
// 	server, backendConn, _ := createExpectedFakeServer(t)

// 	server, err := NewFakeServer(server.conn, server.serverBackendAddr, server.beaconAddr, len(server.sessions), server.sdkVersion, server.buyerID, server.customerPrivateKey, server.dcName, server.sendBeaconPackets)
// 	assert.NoError(t, err)

// 	ctx, cancelFunc := context.WithCancel(context.Background())
// 	defer cancelFunc()

// 	backendRecvReady := make(chan struct{})

// 	go runTestServerBackend(t, ctx, backendConn, testSendNormalResponse, backendRecvReady)
// 	<-backendRecvReady

// 	err = server.update()
// 	assert.NoError(t, err)
// }

func TestSendServerInitPacket(t *testing.T) {
	t.Run("failed to marshal request", func(t *testing.T) {
		server, _, _ := createExpectedFakeServer(t)

		server, err := NewFakeServer(server.conn, server.serverBackendAddr, server.beaconAddr, len(server.sessions), server.sdkVersion, server.buyerID, server.customerPrivateKey, server.dcName, server.sendBeaconPackets)
		assert.NoError(t, err)

		// Bad version which would fail to marshal
		server.sdkVersion = transport.SDKVersion{256, 256, 256}

		err = server.sendServerInitPacket()
		assert.Error(t, err)
	})

	t.Run("failed to send request", func(t *testing.T) {
		server, _, _ := createExpectedFakeServer(t)

		server, err := NewFakeServer(server.conn, server.serverBackendAddr, server.beaconAddr, len(server.sessions), server.sdkVersion, server.buyerID, server.customerPrivateKey, server.dcName, server.sendBeaconPackets)
		assert.NoError(t, err)

		// Simulate a bad connection
		server.conn = &net.UDPConn{}

		err = server.sendServerInitPacket()
		assert.Error(t, err)
	})

	t.Run("failed to read response", func(t *testing.T) {
		server, backendConn, _ := createExpectedFakeServer(t)

		ctx, cancelFunc := context.WithCancel(context.Background())
		defer cancelFunc()

		backendRecvReady := make(chan struct{})

		// Simulate a bad response
		go runTestServerBackend(t, ctx, backendConn, testSendInvalidResponse, backendRecvReady)

		server, err := NewFakeServer(server.conn, server.serverBackendAddr, server.beaconAddr, len(server.sessions), server.sdkVersion, server.buyerID, server.customerPrivateKey, server.dcName, server.sendBeaconPackets)
		assert.NoError(t, err)

		<-backendRecvReady

		err = server.sendServerInitPacket()
		assert.Error(t, err)
	})

	t.Run("mismatched response", func(t *testing.T) {
		server, backendConn, _ := createExpectedFakeServer(t)

		ctx, cancelFunc := context.WithCancel(context.Background())
		defer cancelFunc()

		backendRecvReady := make(chan struct{})

		// Simulate a mismatched response
		go runTestServerBackend(t, ctx, backendConn, testSendMismatchedResponse, backendRecvReady)

		server, err := NewFakeServer(server.conn, server.serverBackendAddr, server.beaconAddr, len(server.sessions), server.sdkVersion, server.buyerID, server.customerPrivateKey, server.dcName, server.sendBeaconPackets)
		assert.NoError(t, err)

		<-backendRecvReady

		err = server.sendServerInitPacket()
		assert.Error(t, err)
	})

	t.Run("failed to unmarshal response", func(t *testing.T) {
		server, backendConn, _ := createExpectedFakeServer(t)

		ctx, cancelFunc := context.WithCancel(context.Background())
		defer cancelFunc()

		backendRecvReady := make(chan struct{})

		// Simulate an unmarshalable response
		go runTestServerBackend(t, ctx, backendConn, testSendUnmarshalableResponse, backendRecvReady)

		server, err := NewFakeServer(server.conn, server.serverBackendAddr, server.beaconAddr, len(server.sessions), server.sdkVersion, server.buyerID, server.customerPrivateKey, server.dcName, server.sendBeaconPackets)
		assert.NoError(t, err)

		<-backendRecvReady

		err = server.sendServerInitPacket()
		assert.Error(t, err)
	})

	t.Run("error response", func(t *testing.T) {
		server, backendConn, _ := createExpectedFakeServer(t)

		ctx, cancelFunc := context.WithCancel(context.Background())
		defer cancelFunc()

		backendRecvReady := make(chan struct{})

		// Simulate an unmarshalable response
		go runTestServerBackend(t, ctx, backendConn, testSendInitErrorResponse, backendRecvReady)

		server, err := NewFakeServer(server.conn, server.serverBackendAddr, server.beaconAddr, len(server.sessions), server.sdkVersion, server.buyerID, server.customerPrivateKey, server.dcName, server.sendBeaconPackets)
		assert.NoError(t, err)

		<-backendRecvReady

		err = server.sendServerInitPacket()
		assert.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		server, backendConn, _ := createExpectedFakeServer(t)

		ctx, cancelFunc := context.WithCancel(context.Background())
		defer cancelFunc()

		backendRecvReady := make(chan struct{})

		go runTestServerBackend(t, ctx, backendConn, testSendNormalResponse, backendRecvReady)

		server, err := NewFakeServer(server.conn, server.serverBackendAddr, server.beaconAddr, len(server.sessions), server.sdkVersion, server.buyerID, server.customerPrivateKey, server.dcName, server.sendBeaconPackets)
		assert.NoError(t, err)

		<-backendRecvReady

		err = server.sendServerInitPacket()
		assert.NoError(t, err)
	})
}

func TestSendServerUpdatePacket(t *testing.T) {
	t.Run("failed to marshal request", func(t *testing.T) {
		server, _, _ := createExpectedFakeServer(t)

		server, err := NewFakeServer(server.conn, server.serverBackendAddr, server.beaconAddr, len(server.sessions), server.sdkVersion, server.buyerID, server.customerPrivateKey, server.dcName, server.sendBeaconPackets)
		assert.NoError(t, err)

		// Bad version which would fail to marshal
		server.sdkVersion = transport.SDKVersion{256, 256, 256}

		err = server.sendServerUpdatePacket()
		assert.Error(t, err)
	})

	t.Run("failed to send request", func(t *testing.T) {
		server, _, _ := createExpectedFakeServer(t)

		server, err := NewFakeServer(server.conn, server.serverBackendAddr, server.beaconAddr, len(server.sessions), server.sdkVersion, server.buyerID, server.customerPrivateKey, server.dcName, server.sendBeaconPackets)
		assert.NoError(t, err)

		// Simulate a bad connection
		server.conn = &net.UDPConn{}

		err = server.sendServerUpdatePacket()
		assert.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		server, backendConn, _ := createExpectedFakeServer(t)

		ctx, cancelFunc := context.WithCancel(context.Background())
		defer cancelFunc()

		backendRecvReady := make(chan struct{})

		go runTestServerBackend(t, ctx, backendConn, testSendNormalResponse, backendRecvReady)

		server, err := NewFakeServer(server.conn, server.serverBackendAddr, server.beaconAddr, len(server.sessions), server.sdkVersion, server.buyerID, server.customerPrivateKey, server.dcName, server.sendBeaconPackets)
		assert.NoError(t, err)

		<-backendRecvReady

		err = server.sendServerUpdatePacket()
		assert.NoError(t, err)
	})
}

func TestSendSessionUpdatePacket(t *testing.T) {
	t.Run("failed to marshal request", func(t *testing.T) {
		server, _, _ := createExpectedFakeServer(t)

		server, err := NewFakeServer(server.conn, server.serverBackendAddr, server.beaconAddr, len(server.sessions), server.sdkVersion, server.buyerID, server.customerPrivateKey, server.dcName, server.sendBeaconPackets)
		assert.NoError(t, err)

		// Bad version which would fail to marshal
		server.sdkVersion = transport.SDKVersion{256, 256, 256}

		session, err := NewSession()
		assert.NoError(t, err)

		_, err = server.sendSessionUpdatePacket(session, false)
		assert.Error(t, err)
	})

	t.Run("failed to send request", func(t *testing.T) {
		server, _, _ := createExpectedFakeServer(t)

		server, err := NewFakeServer(server.conn, server.serverBackendAddr, server.beaconAddr, len(server.sessions), server.sdkVersion, server.buyerID, server.customerPrivateKey, server.dcName, server.sendBeaconPackets)
		assert.NoError(t, err)

		// Simulate a bad connection
		server.conn = &net.UDPConn{}

		session, err := NewSession()
		assert.NoError(t, err)

		_, err = server.sendSessionUpdatePacket(session, false)
		assert.Error(t, err)
	})

	t.Run("failed to read response", func(t *testing.T) {
		server, backendConn, _ := createExpectedFakeServer(t)

		ctx, cancelFunc := context.WithCancel(context.Background())
		defer cancelFunc()

		backendRecvReady := make(chan struct{})

		// Simulate a bad response
		go runTestServerBackend(t, ctx, backendConn, testSendInvalidResponse, backendRecvReady)

		server, err := NewFakeServer(server.conn, server.serverBackendAddr, server.beaconAddr, len(server.sessions), server.sdkVersion, server.buyerID, server.customerPrivateKey, server.dcName, server.sendBeaconPackets)
		assert.NoError(t, err)

		session, err := NewSession()
		assert.NoError(t, err)

		<-backendRecvReady
		_, err = server.sendSessionUpdatePacket(session, false)
		assert.Error(t, err)
	})

	t.Run("mismatched response", func(t *testing.T) {
		server, backendConn, _ := createExpectedFakeServer(t)

		ctx, cancelFunc := context.WithCancel(context.Background())
		defer cancelFunc()

		backendRecvReady := make(chan struct{})

		// Simulate a mismatched response
		go runTestServerBackend(t, ctx, backendConn, testSendMismatchedResponse, backendRecvReady)

		server, err := NewFakeServer(server.conn, server.serverBackendAddr, server.beaconAddr, len(server.sessions), server.sdkVersion, server.buyerID, server.customerPrivateKey, server.dcName, server.sendBeaconPackets)
		assert.NoError(t, err)

		session, err := NewSession()
		assert.NoError(t, err)

		<-backendRecvReady

		_, err = server.sendSessionUpdatePacket(session, false)
		assert.Error(t, err)
	})

	t.Run("failed to unmarshal response", func(t *testing.T) {
		server, backendConn, _ := createExpectedFakeServer(t)

		ctx, cancelFunc := context.WithCancel(context.Background())
		defer cancelFunc()

		backendRecvReady := make(chan struct{})

		// Simulate an unmarshalable response
		go runTestServerBackend(t, ctx, backendConn, testSendUnmarshalableResponse, backendRecvReady)

		server, err := NewFakeServer(server.conn, server.serverBackendAddr, server.beaconAddr, len(server.sessions), server.sdkVersion, server.buyerID, server.customerPrivateKey, server.dcName, server.sendBeaconPackets)
		assert.NoError(t, err)

		session, err := NewSession()
		assert.NoError(t, err)

		<-backendRecvReady

		_, err = server.sendSessionUpdatePacket(session, false)
		assert.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		server, backendConn, _ := createExpectedFakeServer(t)

		ctx, cancelFunc := context.WithCancel(context.Background())
		defer cancelFunc()

		backendRecvReady := make(chan struct{})

		go runTestServerBackend(t, ctx, backendConn, testSendNormalResponse, backendRecvReady)

		server, err := NewFakeServer(server.conn, server.serverBackendAddr, server.beaconAddr, len(server.sessions), server.sdkVersion, server.buyerID, server.customerPrivateKey, server.dcName, server.sendBeaconPackets)
		assert.NoError(t, err)

		session, err := NewSession()
		assert.NoError(t, err)

		<-backendRecvReady

		responsePacket, err := server.sendSessionUpdatePacket(session, false)
		assert.NoError(t, err)
		assert.NotZero(t, responsePacket)
	})
}

// func TestSendPacket(t *testing.T) {
// 	t.Run("fail to set write deadline", func(t *testing.T) {
// 		server, _, _ := createExpectedFakeServer(t)

// 		server, err := NewFakeServer(server.conn, server.serverBackendAddr, server.beaconAddr, len(server.sessions), server.sdkVersion, server.buyerID, server.customerPrivateKey, server.dcName, server.sendBeaconPackets)
// 		assert.NoError(t, err)

// 		server.conn = &net.UDPConn{}

// 		err = server.sendPacket(transport.PacketTypeServerInitRequest, nil)
// 		assert.Error(t, err)
// 	})

// 	t.Run("fail to write", func(t *testing.T) {
// 		server, _, _ := createExpectedFakeServer(t)

// 		server, err := NewFakeServer(server.conn, server.serverBackendAddr, server.beaconAddr, len(server.sessions), server.sdkVersion, server.buyerID, server.customerPrivateKey, server.dcName, server.sendBeaconPackets)
// 		assert.NoError(t, err)

// 		err = server.conn.Close()
// 		assert.NoError(t, err)

// 		err = server.sendPacket(transport.PacketTypeServerInitRequest, nil)
// 		assert.Error(t, err)
// 	})

// 	t.Run("success", func(t *testing.T) {
// 		server, backendConn, _ := createExpectedFakeServer(t)

// 		server, err := NewFakeServer(server.conn, server.serverBackendAddr, server.beaconAddr, len(server.sessions), server.sdkVersion, server.buyerID, server.customerPrivateKey, server.dcName, server.sendBeaconPackets)
// 		assert.NoError(t, err)

// 		go func() {
// 			buffer := make([]byte, transport.DefaultMaxPacketSize)
// 			_, err := backendConn.Read(buffer)
// 			assert.NoError(t, err)
// 		}()

// 		err = server.sendPacket(transport.PacketTypeServerInitRequest, nil)
// 		assert.NoError(t, err)
// 	})
// }

// func TestSendBeaconPacket(t *testing.T) {
// 	t.Run("failed to marshal request", func(t *testing.T) {
// 		server, _, _ := createExpectedFakeServer(t)
// 		server.sendBeaconPackets = true

// 		server, err := NewFakeServer(server.conn, server.serverBackendAddr, server.beaconAddr, len(server.sessions), server.sdkVersion, server.buyerID, server.customerPrivateKey, server.dcName, server.sendBeaconPackets)
// 		assert.NoError(t, err)

// 		session, err := NewSession()
// 		assert.NoError(t, err)

// 		// Bad platform type which would fail to marshal
// 		session.platformType = -1

// 		err = server.sendBeaconPacket(session)
// 		assert.Error(t, err)
// 	})

// 	t.Run("failed to set write deadline", func(t *testing.T) {
// 		server, _, _ := createExpectedFakeServer(t)
// 		server.sendBeaconPackets = true

// 		server, err := NewFakeServer(server.conn, server.serverBackendAddr, server.beaconAddr, len(server.sessions), server.sdkVersion, server.buyerID, server.customerPrivateKey, server.dcName, server.sendBeaconPackets)
// 		assert.NoError(t, err)

// 		// Simulate a bad connection
// 		server.conn = &net.UDPConn{}

// 		session, err := NewSession()
// 		assert.NoError(t, err)

// 		err = server.sendBeaconPacket(session)
// 		assert.Error(t, err)
// 	})

// 	t.Run("fail to write", func(t *testing.T) {
// 		server, _, _ := createExpectedFakeServer(t)
// 		server.sendBeaconPackets = true

// 		server, err := NewFakeServer(server.conn, server.serverBackendAddr, server.beaconAddr, len(server.sessions), server.sdkVersion, server.buyerID, server.customerPrivateKey, server.dcName, server.sendBeaconPackets)
// 		assert.NoError(t, err)

// 		err = server.conn.Close()
// 		assert.NoError(t, err)

// 		session, err := NewSession()
// 		assert.NoError(t, err)

// 		err = server.sendBeaconPacket(session)
// 		assert.Error(t, err)
// 	})

// 	t.Run("success", func(t *testing.T) {
// 		server, _, beaconConn := createExpectedFakeServer(t)
// 		server.sendBeaconPackets = true

// 		ctx, cancelFunc := context.WithCancel(context.Background())
// 		defer cancelFunc()

// 		beaconRecvReady := make(chan struct{})

// 		go runTestBeacon(t, ctx, beaconConn, beaconRecvReady)

// 		server, err := NewFakeServer(server.conn, server.serverBackendAddr, server.beaconAddr, len(server.sessions), server.sdkVersion, server.buyerID, server.customerPrivateKey, server.dcName, server.sendBeaconPackets)
// 		assert.NoError(t, err)

// 		session, err := NewSession()
// 		assert.NoError(t, err)

// 		<-beaconRecvReady

// 		err = server.sendBeaconPacket(session)
// 		assert.NoError(t, err)
// 	})
// }

// func TestReadPacket(t *testing.T) {
// 	t.Run("fail to set read deadline", func(t *testing.T) {
// 		server, _, _ := createExpectedFakeServer(t)

// 		server, err := NewFakeServer(server.conn, server.serverBackendAddr, server.beaconAddr, len(server.sessions), server.sdkVersion, server.buyerID, server.customerPrivateKey, server.dcName, server.sendBeaconPackets)
// 		assert.NoError(t, err)

// 		server.conn = &net.UDPConn{}

// 		packetType, packetData, err := server.readPacket()
// 		assert.Zero(t, packetType)
// 		assert.Nil(t, packetData)
// 		assert.Error(t, err)
// 	})

// 	t.Run("fail to read", func(t *testing.T) {
// 		server, _, _ := createExpectedFakeServer(t)

// 		server, err := NewFakeServer(server.conn, server.serverBackendAddr, server.beaconAddr, len(server.sessions), server.sdkVersion, server.buyerID, server.customerPrivateKey, server.dcName, server.sendBeaconPackets)
// 		assert.NoError(t, err)

// 		err = server.conn.Close()
// 		assert.NoError(t, err)

// 		packetType, packetData, err := server.readPacket()
// 		assert.Zero(t, packetType)
// 		assert.Nil(t, packetData)
// 		assert.Error(t, err)
// 	})

// 	t.Run("read empty packet", func(t *testing.T) {
// 		server, backendConn, _ := createExpectedFakeServer(t)

// 		server, err := NewFakeServer(server.conn, server.serverBackendAddr, server.beaconAddr, len(server.sessions), server.sdkVersion, server.buyerID, server.customerPrivateKey, server.dcName, server.sendBeaconPackets)
// 		assert.NoError(t, err)

// 		go func() {
// 			serverAddr, err := net.ResolveUDPAddr("udp", server.conn.LocalAddr().String())
// 			assert.NoError(t, err)

// 			_, err = backendConn.WriteToUDP([]byte{}, serverAddr)
// 			assert.NoError(t, err)
// 		}()

// 		packetType, packetData, err := server.readPacket()
// 		assert.Zero(t, packetType)
// 		assert.Nil(t, packetData)
// 		assert.Error(t, err)
// 	})

// 	t.Run("read non network next packet", func(t *testing.T) {
// 		server, backendConn, _ := createExpectedFakeServer(t)

// 		server, err := NewFakeServer(server.conn, server.serverBackendAddr, server.beaconAddr, len(server.sessions), server.sdkVersion, server.buyerID, server.customerPrivateKey, server.dcName, server.sendBeaconPackets)
// 		assert.NoError(t, err)

// 		go func() {
// 			serverAddr, err := net.ResolveUDPAddr("udp", server.conn.LocalAddr().String())
// 			assert.NoError(t, err)

// 			_, err = backendConn.WriteToUDP([]byte("bad packet data"), serverAddr)
// 			assert.NoError(t, err)
// 		}()

// 		packetType, packetData, err := server.readPacket()
// 		assert.Zero(t, packetType)
// 		assert.Nil(t, packetData)
// 		assert.Error(t, err)
// 	})

// 	t.Run("success", func(t *testing.T) {
// 		server, backendConn, _ := createExpectedFakeServer(t)

// 		server, err := NewFakeServer(server.conn, server.serverBackendAddr, server.beaconAddr, len(server.sessions), server.sdkVersion, server.buyerID, server.customerPrivateKey, server.dcName, server.sendBeaconPackets)
// 		assert.NoError(t, err)

// 		go func() {
// 			responseData := make([]byte, 2+crypto.PacketHashSize) // We need to have at least 1 byte in the message, otherwise crypto.HashPacket will panic
// 			responseData[0] = transport.PacketTypeServerInitResponse
// 			crypto.HashPacket(crypto.PacketHashKey, responseData)

// 			serverAddr, err := net.ResolveUDPAddr("udp", server.conn.LocalAddr().String())
// 			assert.NoError(t, err)

// 			_, err = backendConn.WriteToUDP(responseData, serverAddr)
// 			assert.NoError(t, err)
// 		}()

// 		packetType, packetData, err := server.readPacket()
// 		assert.NotZero(t, packetType)
// 		assert.Len(t, packetData, 1)
// 		assert.NoError(t, err)
// 	})
// }
