package transport_test

import (
	"crypto/ed25519"
	"encoding/base64"
	"net"
	"testing"

	"github.com/networknext/backend/core"
	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/encoding"
	"github.com/networknext/backend/transport"
	"github.com/stretchr/testify/assert"
)

func TestServerUpdatePacket(t *testing.T) {
	t.Run("crypto/ed25519", func(t *testing.T) {
		customerPublicKey, _ := base64.StdEncoding.DecodeString("leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw==")
		customerPrivateKey, _ := base64.StdEncoding.DecodeString("leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn")

		// Create a ServerUpdatePacket like the game server does
		outgoing := transport.ServerUpdatePacket{
			Sequence:             1,
			VersionMajor:         1,
			VersionMinor:         2,
			VersionPatch:         3,
			CustomerId:           4,
			DatacenterId:         5,
			NumSessionsPending:   6,
			NumSessionsUpgraded:  7,
			ServerAddress:        net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 2323},
			ServerPrivateAddress: net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 2323},
			ServerRoutePublicKey: make([]byte, crypto.KeySize),
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
			VersionMajor:         1,
			VersionMinor:         2,
			VersionPatch:         3,
			CustomerId:           4,
			DatacenterId:         5,
			NumSessionsPending:   6,
			NumSessionsUpgraded:  7,
			ServerAddress:        net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 2323},
			ServerPrivateAddress: net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 2323},
			ServerRoutePublicKey: make([]byte, crypto.KeySize),
		}

		// Sign the packet and set it to the signature
		// using the customer's private key that is on
		// their game server
		outgoing.Signature = core.CryptoSignCreate(outgoing.GetSignData(), customerPrivateKey[8:])

		// Marshal the whole packet to binary to send it over the network
		data, err := outgoing.MarshalBinary()
		assert.NoError(t, err)

		// Unmarshal the data from binary like the server backend receives it
		var incoming transport.ServerUpdatePacket
		err = incoming.UnmarshalBinary(data)
		assert.NoError(t, err)

		// Verify the incoming packet's signed data with the signature
		// with the customer's public key we would get from configstore
		verified := core.CryptoSignVerify(incoming.GetSignData(), incoming.Signature, customerPublicKey[8:])
		assert.True(t, verified)
	})

	t.Run("crypto", func(t *testing.T) {
		customerPublicKey, _ := base64.StdEncoding.DecodeString("leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw==")
		customerPrivateKey, _ := base64.StdEncoding.DecodeString("leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn")

		// Create a ServerUpdatePacket like the game server does
		outgoing := transport.ServerUpdatePacket{
			Sequence:             1,
			VersionMajor:         1,
			VersionMinor:         2,
			VersionPatch:         3,
			CustomerId:           4,
			DatacenterId:         5,
			NumSessionsPending:   6,
			NumSessionsUpgraded:  7,
			ServerAddress:        net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 2323},
			ServerPrivateAddress: net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 2323},
			ServerRoutePublicKey: make([]byte, crypto.KeySize),
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

func GetTestSessionUpdatePacket() transport.SessionUpdatePacket {
	return transport.SessionUpdatePacket{
		Sequence:                  1,
		CustomerId:                2,
		SessionId:                 3,
		UserHash:                  4,
		PlatformId:                5,
		Tag:                       6,
		Flags:                     7,
		Flagged:                   true,
		FallbackToDirect:          true,
		TryBeforeYouBuy:           true,
		ConnectionType:            1,
		OnNetworkNext:             true,
		Committed:                 true,
		DirectMinRtt:              1.5,
		DirectMaxRtt:              2.5,
		DirectMeanRtt:             3.5,
		DirectJitter:              4.5,
		DirectPacketLoss:          5.5,
		NextMinRtt:                6.5,
		NextMaxRtt:                7.5,
		NextMeanRtt:               8.5,
		NextJitter:                9.5,
		NextPacketLoss:            10.5,
		NumNearRelays:             3,
		NearRelayIds:              []uint64{1, 2, 3},
		NearRelayMinRtt:           []float32{1.5, 2.5, 3.5},
		NearRelayMaxRtt:           []float32{1.5, 2.5, 3.5},
		NearRelayMeanRtt:          []float32{1.5, 2.5, 3.5},
		NearRelayJitter:           []float32{1.5, 2.5, 3.5},
		NearRelayPacketLoss:       []float32{1.5, 2.5, 3.5},
		ClientAddress:             net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 2323},
		ServerAddress:             net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 2323},
		ClientRoutePublicKey:      make([]byte, crypto.KeySize),
		KbpsUp:                    10,
		KbpsDown:                  11,
		PacketsLostClientToServer: 12,
		PacketsLostServerToClient: 13,
	}
}

// Added as temp solution to test since "func (packet *SessionUpdatePacket) MarshalBinary()" currently forces MinSDKVersion
// Will need to figure out how we want to determine SDK version of session update packets properly and replace this with regular Marshal call
func MarshalSessionUpdatePacket(packet *transport.SessionUpdatePacket, version transport.SDKVersion) ([]byte, error) {
	ws, err := encoding.CreateWriteStream(transport.DefaultMaxPacketSize)
	if err != nil {
		return nil, err
	}

	if err := packet.Serialize(ws, version); err != nil {
		return nil, err
	}
	ws.Flush()

	return ws.GetData(), nil
}

// Added as temp solution to test since "func (packet *SessionUpdatePacket) UnmarshalBinary()" currently forces MinSDKVersion
// Will need to figure out how we want to determine SDK version of session update packets properly and replace this with regular Marshal call
func UnmarshalSessionUpdatePacket(data []byte, version transport.SDKVersion) (*transport.SessionUpdatePacket, error) {
	packet := &transport.SessionUpdatePacket{}
	if err := packet.Serialize(encoding.CreateReadStream(data), version); err != nil {
		return nil, err
	}
	return packet, nil
}

var SessionUpdatePackets = []struct {
	version transport.SDKVersion
	packet  transport.SessionUpdatePacket
}{
	{
		transport.SDKVersion{3, 4, 0}, transport.SessionUpdatePacket{
			Sequence:         1,
			CustomerId:       2,
			SessionId:        3,
			UserHash:         4,
			PlatformId:       5,
			Tag:              6,
			Flags:            7,
			Flagged:          true,
			FallbackToDirect: true,
			//TryBeforeYouBuy:           true,
			ConnectionType:            1,
			OnNetworkNext:             true,
			Committed:                 true,
			DirectMinRtt:              1.5,
			DirectMaxRtt:              2.5,
			DirectMeanRtt:             3.5,
			DirectJitter:              4.5,
			DirectPacketLoss:          5.5,
			NextMinRtt:                6.5,
			NextMaxRtt:                7.5,
			NextMeanRtt:               8.5,
			NextJitter:                9.5,
			NextPacketLoss:            10.5,
			NumNearRelays:             3,
			NearRelayIds:              []uint64{1, 2, 3},
			NearRelayMinRtt:           []float32{1.5, 2.5, 3.5},
			NearRelayMaxRtt:           []float32{1.5, 2.5, 3.5},
			NearRelayMeanRtt:          []float32{1.5, 2.5, 3.5},
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
	},
}

func TestSessionUpdatePacket(t *testing.T) {

	customerPublicKey, _ := base64.StdEncoding.DecodeString("leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw==")
	customerPrivateKey, _ := base64.StdEncoding.DecodeString("leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn")

	for _, test := range SessionUpdatePackets {
		t.Run(test.version.String(), func(t *testing.T) {

			// Create a SessionUpdatePacket like the game server does
			outgoing := test.packet

			// Sign the packet
			outgoing.Signature = crypto.Sign(customerPrivateKey[8:], outgoing.GetSignData(test.version))

			// Marshal the whole packet to binary to send it over the network
			data, err := MarshalSessionUpdatePacket(&outgoing, test.version)
			assert.NoError(t, err)

			// Unmarshal the data from binary like the server backend receives it
			incoming, err := UnmarshalSessionUpdatePacket(data, test.version)
			assert.NoError(t, err)

			// Verify the incoming packet's signed data with the signature
			// with the customer's public key we would get from configstore
			verified := crypto.Verify(customerPublicKey[8:], incoming.GetSignData(test.version), incoming.Signature)
			assert.True(t, verified)

			assert.EqualValues(t, outgoing, *incoming)
		})
	}
}
