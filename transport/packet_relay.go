package transport

import (
	"encoding/base64"
	"errors"
	"fmt"
	"math"
	"net"

	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/encoding"
	"github.com/networknext/backend/routing"
)

const (
	VersionNumberInitRequest    = 0
	VersionNumberInitResponse   = 0
	VersionNumberUpdateRequest  = 0
	VersionNumberUpdateResponse = 0
)

// RelayInitPacket is the struct that describes the packets comming into the relay_init endpoint
type RelayInitPacket struct {
	Magic          uint32
	Version        uint32
	Nonce          []byte
	Address        net.UDPAddr
	EncryptedToken []byte
}

type RelayInitRequestJSON struct {
	Magic                uint32 `json:"magic_request_protection"`
	Version              uint32 `json:"version"`
	StringAddr           string `json:"relay_address"`
	PortNum              int    `json:"relay_port"`
	NonceBase64          string `json:"nonce"`
	EncryptedTokenBase64 string `json:"encrypted_token"`
}

// RelayInitResponseJSON ...
type RelayInitResponseJSON struct {
	Timestamp uint64 `json:"Timestamp"`
}

func (j *RelayInitRequestJSON) ToInitPacket(packet *RelayInitPacket) error {
	var err error

	packet.Magic = j.Magic
	packet.Version = j.Version

	var nonce []byte
	{
		if nonce, err = base64.RawStdEncoding.DecodeString(j.NonceBase64); err != nil {
			return err
		}
	}

	packet.Nonce = make([]byte, len(nonce))
	copy(packet.Nonce, nonce)

	var addr *net.UDPAddr
	{
		if addr, err = net.ResolveUDPAddr("udp", j.StringAddr); err != nil {
			return err
		}
	}

	packet.Address = *addr

	var token []byte
	{
		if token, err = base64.RawStdEncoding.DecodeString(j.EncryptedTokenBase64); err != nil {
			return err
		}
	}

	packet.EncryptedToken = make([]byte, len(token))
	copy(packet.EncryptedToken, token)

	return nil
}

// UnmarshalBinary decodes binary data into a RelayInitPacket struct
func (r *RelayInitPacket) UnmarshalBinary(buf []byte) error {
	index := 0
	var addr string
	if !(encoding.ReadUint32(buf, &index, &r.Magic) &&
		encoding.ReadUint32(buf, &index, &r.Version) &&
		encoding.ReadBytes(buf, &index, &r.Nonce, crypto.NonceSize) &&
		encoding.ReadString(buf, &index, &addr, MaxRelayAddressLength) &&
		encoding.ReadBytes(buf, &index, &r.EncryptedToken, routing.EncryptedTokenSize)) {
		return errors.New("invalid packet")
	}

	if udp, err := net.ResolveUDPAddr("udp", addr); udp != nil && err == nil {
		r.Address = *udp
	} else {
		return fmt.Errorf("could not resolve init packet with address '%s' with reason: %v", addr, err)
	}

	return nil
}

// MarshalBinary ...
func (r RelayInitPacket) MarshalBinary() ([]byte, error) {
	data := make([]byte, 4+4+crypto.NonceSize+4+len(r.Address.String())+routing.EncryptedTokenSize)
	index := 0
	encoding.WriteUint32(data, &index, r.Magic)
	encoding.WriteUint32(data, &index, r.Version)
	encoding.WriteBytes(data, &index, r.Nonce, crypto.NonceSize)
	encoding.WriteString(data, &index, r.Address.String(), uint32(len(r.Address.String())))
	encoding.WriteBytes(data, &index, r.EncryptedToken, routing.EncryptedTokenSize)

	return data, nil
}

// RelayUpdatePacket is the struct wrapping a update packet
type RelayUpdatePacket struct {
	Version   uint32
	Address   net.UDPAddr
	Token     []byte
	NumRelays uint32

	PingStats []routing.RelayStatsPing
}

type packetMetadata struct {
	TokenBase64 string `json:"PublicKey"`
}

type RelayUpdateRequestJSON struct {
	Version    uint32                   `json:"version"`
	StringAddr string                   `json:"relay_address"`
	Metadata   packetMetadata           `json:"Metadata"`
	PingStats  []routing.RelayStatsPing `json:"PingStats"`
}

type RelayUpdateResponseJSON struct {
	RelaysToPing []routing.LegacyPingData `json:"ping_data"`
}

func (j *RelayUpdateRequestJSON) ToUpdatePacket(packet *RelayUpdatePacket) error {
	var err error

	packet.Version = j.Version

	var addr *net.UDPAddr
	{
		if addr, err = net.ResolveUDPAddr("udp", j.StringAddr); err != nil {
			return err
		}
	}

	packet.Address = *addr

	var token []byte
	{
		if token, err = base64.StdEncoding.DecodeString(j.Metadata.TokenBase64); err != nil {
			return err
		}
	}

	packet.Token = make([]byte, len(token))
	copy(packet.Token, token)

	packet.NumRelays = uint32(len(j.PingStats))
	packet.PingStats = make([]routing.RelayStatsPing, packet.NumRelays)
	copy(packet.PingStats, j.PingStats)

	return nil
}

// UnmarshalBinary decodes the binary data into a RelayUpdatePacket struct
func (r *RelayUpdatePacket) UnmarshalBinary(buff []byte) error {
	index := 0
	var addr string
	if !(encoding.ReadUint32(buff, &index, &r.Version) &&
		encoding.ReadString(buff, &index, &addr, MaxRelayAddressLength) &&
		encoding.ReadBytes(buff, &index, &r.Token, crypto.KeySize) &&
		encoding.ReadUint32(buff, &index, &r.NumRelays)) {
		return errors.New("invalid packet")
	}

	if udp, err := net.ResolveUDPAddr("udp", addr); udp != nil && err == nil {
		r.Address = *udp
	} else {
		return fmt.Errorf("could not resolve init packet with address '%s' with reason: %v", addr, err)
	}

	r.PingStats = make([]routing.RelayStatsPing, r.NumRelays)
	for i := 0; i < int(r.NumRelays); i++ {
		stats := &r.PingStats[i]

		if !(encoding.ReadUint64(buff, &index, &stats.RelayID) &&
			encoding.ReadFloat32(buff, &index, &stats.RTT) &&
			encoding.ReadFloat32(buff, &index, &stats.Jitter) &&
			encoding.ReadFloat32(buff, &index, &stats.PacketLoss)) {
			return errors.New("invalid packet")
		}
	}

	return nil
}

// MarshalBinary ...
func (r RelayUpdatePacket) MarshalBinary() ([]byte, error) {
	data := make([]byte, 4+4+len(r.Address.String())+routing.EncryptedTokenSize+4+20*len(r.PingStats))

	index := 0
	encoding.WriteUint32(data, &index, r.Version)
	encoding.WriteString(data, &index, r.Address.String(), math.MaxInt32)
	encoding.WriteBytes(data, &index, r.Token, crypto.KeySize)
	encoding.WriteUint32(data, &index, r.NumRelays)

	for i := 0; i < int(r.NumRelays); i++ {
		stats := &r.PingStats[i]

		encoding.WriteUint64(data, &index, stats.RelayID)
		encoding.WriteUint32(data, &index, math.Float32bits(stats.RTT))
		encoding.WriteUint32(data, &index, math.Float32bits(stats.Jitter))
		encoding.WriteUint32(data, &index, math.Float32bits(stats.PacketLoss))
	}

	return data, nil
}
