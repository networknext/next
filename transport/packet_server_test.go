package transport_test

import (
	"crypto/ed25519"
	"encoding/base64"
	"net"
	"testing"

	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/transport"
	"github.com/stretchr/testify/assert"
)

func TestServerUpdatePacket(t *testing.T) {
	t.Parallel()

	t.Run("crypto/ed25519", func(t *testing.T) {
		customerPublicKey, _ := base64.StdEncoding.DecodeString("leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw==")
		customerPrivateKey, _ := base64.StdEncoding.DecodeString("leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn")

		// Create a ServerUpdatePacket like the game server does
		outgoing := transport.ServerUpdatePacket{
			Sequence:             1,
			CustomerID:           4,
			DatacenterID:         5,
			NumSessionsPending:   6,
			NumSessionsUpgraded:  7,
			ServerAddress:        net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 2323},
			ServerPrivateAddress: net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 2323},
			ServerRoutePublicKey: make([]byte, crypto.KeySize),

			Version: routing.SDKVersion{1, 2, 3},
		}

		// Sign the packet and set it to the signature
		// using the customer's private key that is on
		// their game server
		outgoing.Signature = ed25519.Sign(customerPrivateKey[8:], outgoing.GetSignData())

		// Marshal the whole packet to binary to send it over the network
		data, err := outgoing.MarshalBinary()
		assert.NoError(t, err)

		// Unmarshal the data from binary like the server backend receives it
		var incoming transport.ServerUpdatePacket
		err = incoming.UnmarshalBinary(data)
		assert.NoError(t, err)

		// Verify the incoming packet's signed data with the signature
		// with the customer's public key we would get from configstore
		ed25519.Verify(customerPublicKey[8:], incoming.GetSignData(), incoming.Signature)
	})

	t.Run("libsodium", func(t *testing.T) {
		customerPublicKey, _ := base64.StdEncoding.DecodeString("leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw==")
		customerPrivateKey, _ := base64.StdEncoding.DecodeString("leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn")

		// Create a ServerUpdatePacket like the game server does
		outgoing := transport.ServerUpdatePacket{
			Sequence:             1,
			CustomerID:           4,
			DatacenterID:         5,
			NumSessionsPending:   6,
			NumSessionsUpgraded:  7,
			ServerAddress:        net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 2323},
			ServerPrivateAddress: net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 2323},
			ServerRoutePublicKey: make([]byte, crypto.KeySize),

			Version: routing.SDKVersion{1, 2, 3},
		}

		// Sign the packet and set it to the signature
		// using the customer's private key that is on
		// their game server
		outgoing.Signature = crypto.Sign(customerPrivateKey[8:], outgoing.GetSignData())

		// Marshal the whole packet to binary to send it over the network
		data, err := outgoing.MarshalBinary()
		assert.NoError(t, err)

		// Unmarshal the data from binary like the server backend receives it
		var incoming transport.ServerUpdatePacket
		err = incoming.UnmarshalBinary(data)
		assert.NoError(t, err)

		// Verify the incoming packet's signed data with the signature
		// with the customer's public key we would get from configstore
		verified := crypto.Verify(customerPublicKey[8:], incoming.GetSignData(), incoming.Signature)
		assert.True(t, verified)
	})

	t.Run("crypto", func(t *testing.T) {
		customerPublicKey, _ := base64.StdEncoding.DecodeString("leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw==")
		customerPrivateKey, _ := base64.StdEncoding.DecodeString("leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn")

		// Create a ServerUpdatePacket like the game server does
		outgoing := transport.ServerUpdatePacket{
			Sequence:             1,
			CustomerID:           4,
			DatacenterID:         5,
			NumSessionsPending:   6,
			NumSessionsUpgraded:  7,
			ServerAddress:        net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 2323},
			ServerPrivateAddress: net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 2323},
			ServerRoutePublicKey: make([]byte, crypto.KeySize),

			Version: routing.SDKVersion{1, 2, 3},
		}

		// Sign the packet and set it to the signature
		// using the customer's private key that is on
		// their game server
		outgoing.Signature = crypto.Sign(customerPrivateKey[8:], outgoing.GetSignData())

		// Marshal the whole packet to binary to send it over the network
		data, err := outgoing.MarshalBinary()
		assert.NoError(t, err)

		// Unmarshal the data from binary like the server backend receives it
		var incoming transport.ServerUpdatePacket
		err = incoming.UnmarshalBinary(data)
		assert.NoError(t, err)

		// Verify the incoming packet's signed data with the signature
		// with the customer's public key we would get from configstore
		verified := crypto.Verify(customerPublicKey[8:], incoming.GetSignData(), incoming.Signature)
		assert.True(t, verified)
	})
}

func TestSessionUpdatePacket(t *testing.T) {
	t.Parallel()

	customerPublicKey, _ := base64.StdEncoding.DecodeString("leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw==")
	customerPrivateKey, _ := base64.StdEncoding.DecodeString("leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn")

	SessionUpdatePackets := []transport.SessionUpdatePacket{
		{
			Version: routing.SDKVersion{3, 3, 2},

			Sequence:                  1,
			CustomerID:                2,
			SessionID:                 3,
			UserHash:                  4,
			PlatformID:                5,
			Tag:                       6,
			Flagged:                   true,
			FallbackToDirect:          true,
			TryBeforeYouBuy:           true,
			ConnectionType:            1,
			OnNetworkNext:             true,
			DirectMinRTT:              1.5,
			DirectMaxRTT:              2.5,
			DirectMeanRTT:             3.5,
			DirectJitter:              4.5,
			DirectPacketLoss:          5.5,
			NextMinRTT:                6.5,
			NextMaxRTT:                7.5,
			NextMeanRTT:               8.5,
			NextJitter:                9.5,
			NextPacketLoss:            10.5,
			NumNearRelays:             3,
			NearRelayIDs:              []uint64{1, 2, 3},
			NearRelayMinRTT:           []float32{1.5, 2.5, 3.5},
			NearRelayMaxRTT:           []float32{1.5, 2.5, 3.5},
			NearRelayMeanRTT:          []float32{1.5, 2.5, 3.5},
			NearRelayJitter:           []float32{1.5, 2.5, 3.5},
			NearRelayPacketLoss:       []float32{1.5, 2.5, 3.5},
			ClientAddress:             net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 2323},
			ServerAddress:             net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 2323},
			ClientRoutePublicKey:      make([]byte, crypto.KeySize),
			KbpsUp:                    10,
			KbpsDown:                  11,
			PacketsLostClientToServer: 12,
			PacketsLostServerToClient: 13,
		},
		{
			Version: routing.SDKVersion{3, 3, 3},

			Sequence:                  1,
			CustomerID:                2,
			SessionID:                 3,
			UserHash:                  4,
			PlatformID:                5,
			Tag:                       6,
			Flagged:                   true,
			FallbackToDirect:          true,
			TryBeforeYouBuy:           true,
			ConnectionType:            1,
			OnNetworkNext:             true,
			DirectMinRTT:              1.5,
			DirectMaxRTT:              2.5,
			DirectMeanRTT:             3.5,
			DirectJitter:              4.5,
			DirectPacketLoss:          5.5,
			NextMinRTT:                6.5,
			NextMaxRTT:                7.5,
			NextMeanRTT:               8.5,
			NextJitter:                9.5,
			NextPacketLoss:            10.5,
			NumNearRelays:             3,
			NearRelayIDs:              []uint64{1, 2, 3},
			NearRelayMinRTT:           []float32{1.5, 2.5, 3.5},
			NearRelayMaxRTT:           []float32{1.5, 2.5, 3.5},
			NearRelayMeanRTT:          []float32{1.5, 2.5, 3.5},
			NearRelayJitter:           []float32{1.5, 2.5, 3.5},
			NearRelayPacketLoss:       []float32{1.5, 2.5, 3.5},
			ClientAddress:             net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 2323},
			ServerAddress:             net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 2323},
			ClientRoutePublicKey:      make([]byte, crypto.KeySize),
			KbpsUp:                    10,
			KbpsDown:                  11,
			PacketsLostClientToServer: 12,
			PacketsLostServerToClient: 13,
		},
		{
			Version: routing.SDKVersion{3, 3, 4},

			Sequence:                  1,
			CustomerID:                2,
			SessionID:                 3,
			UserHash:                  4,
			PlatformID:                5,
			Tag:                       6,
			Flags:                     1023,
			Flagged:                   true,
			FallbackToDirect:          true,
			TryBeforeYouBuy:           true,
			ConnectionType:            1,
			OnNetworkNext:             true,
			DirectMinRTT:              1.5,
			DirectMaxRTT:              2.5,
			DirectMeanRTT:             3.5,
			DirectJitter:              4.5,
			DirectPacketLoss:          5.5,
			NextMinRTT:                6.5,
			NextMaxRTT:                7.5,
			NextMeanRTT:               8.5,
			NextJitter:                9.5,
			NextPacketLoss:            10.5,
			NumNearRelays:             3,
			NearRelayIDs:              []uint64{1, 2, 3},
			NearRelayMinRTT:           []float32{1.5, 2.5, 3.5},
			NearRelayMaxRTT:           []float32{1.5, 2.5, 3.5},
			NearRelayMeanRTT:          []float32{1.5, 2.5, 3.5},
			NearRelayJitter:           []float32{1.5, 2.5, 3.5},
			NearRelayPacketLoss:       []float32{1.5, 2.5, 3.5},
			ClientAddress:             net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 2323},
			ServerAddress:             net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 2323},
			ClientRoutePublicKey:      make([]byte, crypto.KeySize),
			KbpsUp:                    10,
			KbpsDown:                  11,
			PacketsLostClientToServer: 12,
			PacketsLostServerToClient: 13,
		},
		{
			Version: routing.SDKVersion{3, 4, 0},

			Sequence:                  1,
			CustomerID:                2,
			SessionID:                 3,
			UserHash:                  4,
			PlatformID:                5,
			Tag:                       6,
			Flags:                     2047,
			Flagged:                   true,
			FallbackToDirect:          true,
			ConnectionType:            1,
			OnNetworkNext:             true,
			Committed:                 true,
			DirectMinRTT:              1.5,
			DirectMaxRTT:              2.5,
			DirectMeanRTT:             3.5,
			DirectJitter:              4.5,
			DirectPacketLoss:          5.5,
			NextMinRTT:                6.5,
			NextMaxRTT:                7.5,
			NextMeanRTT:               8.5,
			NextJitter:                9.5,
			NextPacketLoss:            10.5,
			NumNearRelays:             3,
			NearRelayIDs:              []uint64{1, 2, 3},
			NearRelayMinRTT:           []float32{1.5, 2.5, 3.5},
			NearRelayMaxRTT:           []float32{1.5, 2.5, 3.5},
			NearRelayMeanRTT:          []float32{1.5, 2.5, 3.5},
			NearRelayJitter:           []float32{1.5, 2.5, 3.5},
			NearRelayPacketLoss:       []float32{1.5, 2.5, 3.5},
			ClientAddress:             net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 2323},
			ServerAddress:             net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 2323},
			ClientRoutePublicKey:      make([]byte, crypto.KeySize),
			KbpsUp:                    10,
			KbpsDown:                  11,
			PacketsLostClientToServer: 12,
			PacketsLostServerToClient: 13,
			UserFlags:                 14,
		},
		{
			Version: routing.SDKVersion{3, 4, 5},

			Sequence:                  1,
			CustomerID:                2,
			SessionID:                 3,
			UserHash:                  4,
			PlatformID:                5,
			Tag:                       6,
			Flags:                     2047,
			Flagged:                   true,
			FallbackToDirect:          true,
			ConnectionType:            1,
			OnNetworkNext:             true,
			Committed:                 true,
			DirectMinRTT:              1.5,
			DirectMaxRTT:              2.5,
			DirectMeanRTT:             3.5,
			DirectJitter:              4.5,
			DirectPacketLoss:          5.5,
			NextMinRTT:                6.5,
			NextMaxRTT:                7.5,
			NextMeanRTT:               8.5,
			NextJitter:                9.5,
			NextPacketLoss:            10.5,
			NumNearRelays:             3,
			NearRelayIDs:              []uint64{1, 2, 3},
			NearRelayMinRTT:           []float32{1.5, 2.5, 3.5},
			NearRelayMaxRTT:           []float32{1.5, 2.5, 3.5},
			NearRelayMeanRTT:          []float32{1.5, 2.5, 3.5},
			NearRelayJitter:           []float32{1.5, 2.5, 3.5},
			NearRelayPacketLoss:       []float32{1.5, 2.5, 3.5},
			ClientAddress:             net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 2323},
			ServerAddress:             net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 2323},
			ClientRoutePublicKey:      make([]byte, crypto.KeySize),
			KbpsUp:                    10,
			KbpsDown:                  11,
			PacketsSentClientToServer: 12,
			PacketsSentServerToClient: 13,
			PacketsLostClientToServer: 14,
			PacketsLostServerToClient: 15,
			UserFlags:                 16,
		},
	}

	for _, packet := range SessionUpdatePackets {
		t.Run(packet.Version.String(), func(t *testing.T) {
			// Sign the packet
			packet.Signature = crypto.Sign(customerPrivateKey[8:], packet.GetSignData())

			// Marshal the whole packet to binary to send it over the network
			data, err := packet.MarshalBinary()
			assert.NoError(t, err)

			// Unmarshal the data from binary like the server backend receives it
			var newpacket transport.SessionUpdatePacket
			newpacket.Version = packet.Version

			err = newpacket.UnmarshalBinary(data)
			assert.NoError(t, err)

			// Verify the incoming packet's signed data with the signature
			// with the customer's public key we would get from configstore
			verified := crypto.Verify(customerPublicKey[8:], newpacket.GetSignData(), newpacket.Signature)
			assert.True(t, verified)

			// Make sure the data was preserved during serialization and deserialization
			assert.EqualValues(t, packet, newpacket)
		})
	}
}

func TestSessionResponsePacket(t *testing.T) {
	t.Parallel()

	customerPublicKey, _ := base64.StdEncoding.DecodeString("leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw==")
	customerPrivateKey, _ := base64.StdEncoding.DecodeString("leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn")

	SessionResponsePackets := []transport.SessionResponsePacket{
		// 3.3.2 Packets
		{
			Version:       routing.SDKVersion{3, 3, 2},
			Sequence:      1,
			SessionID:     2,
			NumNearRelays: 3,
			NearRelayIDs:  []uint64{1, 2, 3},
			NearRelayAddresses: []net.UDPAddr{
				{IP: net.ParseIP("127.0.0.1"), Port: 1000},
				{IP: net.ParseIP("127.0.0.2"), Port: 2000},
				{IP: net.ParseIP("127.0.0.3"), Port: 3000},
			},
			RouteType:            routing.RouteTypeDirect,
			Multipath:            false,
			NumTokens:            0,
			Tokens:               nil,
			ServerRoutePublicKey: make([]byte, ed25519.PublicKeySize),
		},
		{
			Version:       routing.SDKVersion{3, 3, 2},
			Sequence:      1,
			SessionID:     2,
			NumNearRelays: 3,
			NearRelayIDs:  []uint64{1, 2, 3},
			NearRelayAddresses: []net.UDPAddr{
				{IP: net.ParseIP("127.0.0.1"), Port: 1000},
				{IP: net.ParseIP("127.0.0.2"), Port: 2000},
				{IP: net.ParseIP("127.0.0.3"), Port: 3000},
			},
			RouteType:            routing.RouteTypeNew,
			Multipath:            true,
			NumTokens:            3,
			Tokens:               make([]byte, routing.EncryptedNextRouteTokenSize*3),
			ServerRoutePublicKey: make([]byte, ed25519.PublicKeySize),
		},
		{
			Version:       routing.SDKVersion{3, 3, 2},
			Sequence:      1,
			SessionID:     2,
			NumNearRelays: 3,
			NearRelayIDs:  []uint64{1, 2, 3},
			NearRelayAddresses: []net.UDPAddr{
				{IP: net.ParseIP("127.0.0.1"), Port: 1000},
				{IP: net.ParseIP("127.0.0.2"), Port: 2000},
				{IP: net.ParseIP("127.0.0.3"), Port: 3000},
			},
			RouteType:            routing.RouteTypeContinue,
			Multipath:            true,
			NumTokens:            3,
			Tokens:               make([]byte, routing.EncryptedContinueRouteTokenSize*3),
			ServerRoutePublicKey: make([]byte, ed25519.PublicKeySize),
		},
		// 3.3.3 Packets
		{
			Version:       routing.SDKVersion{3, 3, 3},
			Sequence:      1,
			SessionID:     2,
			NumNearRelays: 3,
			NearRelayIDs:  []uint64{1, 2, 3},
			NearRelayAddresses: []net.UDPAddr{
				{IP: net.ParseIP("127.0.0.1"), Port: 1000},
				{IP: net.ParseIP("127.0.0.2"), Port: 2000},
				{IP: net.ParseIP("127.0.0.3"), Port: 3000},
			},
			RouteType:            routing.RouteTypeDirect,
			Multipath:            false,
			NumTokens:            0,
			Tokens:               nil,
			ServerRoutePublicKey: make([]byte, ed25519.PublicKeySize),
		},
		{
			Version:       routing.SDKVersion{3, 3, 3},
			Sequence:      1,
			SessionID:     2,
			NumNearRelays: 3,
			NearRelayIDs:  []uint64{1, 2, 3},
			NearRelayAddresses: []net.UDPAddr{
				{IP: net.ParseIP("127.0.0.1"), Port: 1000},
				{IP: net.ParseIP("127.0.0.2"), Port: 2000},
				{IP: net.ParseIP("127.0.0.3"), Port: 3000},
			},
			RouteType:            routing.RouteTypeNew,
			Multipath:            true,
			NumTokens:            3,
			Tokens:               make([]byte, routing.EncryptedNextRouteTokenSize*3),
			ServerRoutePublicKey: make([]byte, ed25519.PublicKeySize),
		},
		{
			Version:       routing.SDKVersion{3, 3, 3},
			Sequence:      1,
			SessionID:     2,
			NumNearRelays: 3,
			NearRelayIDs:  []uint64{1, 2, 3},
			NearRelayAddresses: []net.UDPAddr{
				{IP: net.ParseIP("127.0.0.1"), Port: 1000},
				{IP: net.ParseIP("127.0.0.2"), Port: 2000},
				{IP: net.ParseIP("127.0.0.3"), Port: 3000},
			},
			RouteType:            routing.RouteTypeContinue,
			Multipath:            true,
			NumTokens:            3,
			Tokens:               make([]byte, routing.EncryptedContinueRouteTokenSize*3),
			ServerRoutePublicKey: make([]byte, ed25519.PublicKeySize),
		},
		// 3.3.4 Packets
		{
			Version:       routing.SDKVersion{3, 3, 4},
			Sequence:      1,
			SessionID:     2,
			NumNearRelays: 3,
			NearRelayIDs:  []uint64{1, 2, 3},
			NearRelayAddresses: []net.UDPAddr{
				{IP: net.ParseIP("127.0.0.1"), Port: 1000},
				{IP: net.ParseIP("127.0.0.2"), Port: 2000},
				{IP: net.ParseIP("127.0.0.3"), Port: 3000},
			},
			RouteType:            routing.RouteTypeDirect,
			Multipath:            false,
			NumTokens:            0,
			Tokens:               nil,
			ServerRoutePublicKey: make([]byte, ed25519.PublicKeySize),
		},
		{
			Version:       routing.SDKVersion{3, 3, 4},
			Sequence:      1,
			SessionID:     2,
			NumNearRelays: 3,
			NearRelayIDs:  []uint64{1, 2, 3},
			NearRelayAddresses: []net.UDPAddr{
				{IP: net.ParseIP("127.0.0.1"), Port: 1000},
				{IP: net.ParseIP("127.0.0.2"), Port: 2000},
				{IP: net.ParseIP("127.0.0.3"), Port: 3000},
			},
			RouteType:            routing.RouteTypeNew,
			Multipath:            true,
			NumTokens:            3,
			Tokens:               make([]byte, routing.EncryptedNextRouteTokenSize*3),
			ServerRoutePublicKey: make([]byte, ed25519.PublicKeySize),
		},
		{
			Version:       routing.SDKVersion{3, 3, 4},
			Sequence:      1,
			SessionID:     2,
			NumNearRelays: 3,
			NearRelayIDs:  []uint64{1, 2, 3},
			NearRelayAddresses: []net.UDPAddr{
				{IP: net.ParseIP("127.0.0.1"), Port: 1000},
				{IP: net.ParseIP("127.0.0.2"), Port: 2000},
				{IP: net.ParseIP("127.0.0.3"), Port: 3000},
			},
			RouteType:            routing.RouteTypeContinue,
			Multipath:            true,
			NumTokens:            3,
			Tokens:               make([]byte, routing.EncryptedContinueRouteTokenSize*3),
			ServerRoutePublicKey: make([]byte, ed25519.PublicKeySize),
		},
		// 3.4.0 Packets
		{
			Version:       routing.SDKVersion{3, 4, 0},
			Sequence:      1,
			SessionID:     2,
			NumNearRelays: 3,
			NearRelayIDs:  []uint64{1, 2, 3},
			NearRelayAddresses: []net.UDPAddr{
				{IP: net.ParseIP("127.0.0.1"), Port: 1000},
				{IP: net.ParseIP("127.0.0.2"), Port: 2000},
				{IP: net.ParseIP("127.0.0.3"), Port: 3000},
			},
			RouteType:            routing.RouteTypeDirect,
			Multipath:            false,
			Committed:            false,
			NumTokens:            0,
			Tokens:               nil,
			ServerRoutePublicKey: make([]byte, ed25519.PublicKeySize),
		},
		{
			Version:       routing.SDKVersion{3, 4, 0},
			Sequence:      1,
			SessionID:     2,
			NumNearRelays: 3,
			NearRelayIDs:  []uint64{1, 2, 3},
			NearRelayAddresses: []net.UDPAddr{
				{IP: net.ParseIP("127.0.0.1"), Port: 1000},
				{IP: net.ParseIP("127.0.0.2"), Port: 2000},
				{IP: net.ParseIP("127.0.0.3"), Port: 3000},
			},
			RouteType:            routing.RouteTypeNew,
			Multipath:            true,
			Committed:            true,
			NumTokens:            3,
			Tokens:               make([]byte, routing.EncryptedNextRouteTokenSize*3),
			ServerRoutePublicKey: make([]byte, ed25519.PublicKeySize),
		},
		{
			Version:       routing.SDKVersion{3, 4, 0},
			Sequence:      1,
			SessionID:     2,
			NumNearRelays: 3,
			NearRelayIDs:  []uint64{1, 2, 3},
			NearRelayAddresses: []net.UDPAddr{
				{IP: net.ParseIP("127.0.0.1"), Port: 1000},
				{IP: net.ParseIP("127.0.0.2"), Port: 2000},
				{IP: net.ParseIP("127.0.0.3"), Port: 3000},
			},
			RouteType:            routing.RouteTypeContinue,
			Multipath:            true,
			Committed:            true,
			NumTokens:            3,
			Tokens:               make([]byte, routing.EncryptedContinueRouteTokenSize*3),
			ServerRoutePublicKey: make([]byte, ed25519.PublicKeySize),
		}}

	for _, packet := range SessionResponsePackets {
		t.Run(packet.Version.String(), func(t *testing.T) {
			// Sign the packet
			packet.Signature = crypto.Sign(customerPrivateKey[8:], packet.GetSignData())

			// Marshal the whole packet to binary to send it over the network
			data, err := packet.MarshalBinary()
			assert.NoError(t, err)

			// Unmarshal the data from binary like the server backend receives it
			var newpacket transport.SessionResponsePacket
			newpacket.Version = packet.Version

			err = newpacket.UnmarshalBinary(data)
			assert.NoError(t, err)

			// Verify the incoming packet's signed data with the signature
			// with the customer's public key we would get from configstore
			verified := crypto.Verify(customerPublicKey[8:], newpacket.GetSignData(), newpacket.Signature)
			assert.True(t, verified)

			// Make sure the data was preserved during serialization and deserialization
			assert.EqualValues(t, packet, newpacket)
		})
	}
}
