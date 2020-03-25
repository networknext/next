package transport

import (
	"encoding/base64"
	"encoding/json"
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

	PacketSizeRelayInitResponse = 4 + 8 + crypto.KeySize
)

// RelayInitRequest is the struct that describes the packets comming into the relay_init endpoint
type RelayInitRequest struct {
	Magic          uint32
	Version        uint32
	Nonce          []byte
	Address        net.UDPAddr
	EncryptedToken []byte
}

func (r *RelayInitRequest) UnmarshalJSON(buf []byte) error {
	var err error
	data := make(map[string]interface{})

	if err := json.Unmarshal(buf, &data); err != nil {
		return err
	}

	if magic, ok := data["magic_request_protection"].(float64); ok {
		r.Magic = uint32(magic)
	}

	if version, ok := data["version"]; ok {
		if v, ok := version.(float64); ok {
			r.Version = uint32(v)
		}
	}

	if addr, ok := data["relay_address"].(string); ok {
		host, _, err := net.SplitHostPort(addr)
		if err != nil {
			return err
		}
		r.Address.IP = net.ParseIP(host)
		if r.Address.IP == nil {
			return errors.New("invalid relay_address")
		}
	}

	if port, ok := data["relay_port"].(float64); ok {
		r.Address.Port = int(port)
	}

	if nonce, ok := data["nonce"].(string); ok {
		if r.Nonce, err = base64.RawStdEncoding.DecodeString(nonce); err != nil {
			return err
		}
	}

	if token, ok := data["encrypted_token"].(string); ok {
		if r.EncryptedToken, err = base64.RawStdEncoding.DecodeString(token); err != nil {
			return err
		}
	}

	return nil
}

// UnmarshalBinary decodes binary data into a RelayInitRequest struct
func (r *RelayInitRequest) UnmarshalBinary(buf []byte) error {
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
func (r RelayInitRequest) MarshalBinary() ([]byte, error) {
	data := make([]byte, 4+4+crypto.NonceSize+4+len(r.Address.String())+routing.EncryptedTokenSize)
	index := 0
	encoding.WriteUint32(data, &index, r.Magic)
	encoding.WriteUint32(data, &index, r.Version)
	encoding.WriteBytes(data, &index, r.Nonce, crypto.NonceSize)
	encoding.WriteString(data, &index, r.Address.String(), uint32(len(r.Address.String())))
	encoding.WriteBytes(data, &index, r.EncryptedToken, routing.EncryptedTokenSize)

	return data, nil
}

// RelayInitResponse ...
type RelayInitResponse struct {
	Version   uint32
	Timestamp uint64
	PublicKey []byte
}

func (r RelayInitResponse) MarshalJSON() ([]byte, error) {
	data := make(map[string]interface{})

	data["Timestamp"] = r.Timestamp

	return json.Marshal(data)
}

func (r RelayInitResponse) MarshalBinary() ([]byte, error) {
	index := 0
	responseData := make([]byte, PacketSizeRelayInitResponse)

	encoding.WriteUint32(responseData, &index, VersionNumberInitResponse)
	encoding.WriteUint64(responseData, &index, r.Timestamp)
	encoding.WriteBytes(responseData, &index, r.PublicKey, crypto.KeySize)

	return responseData, nil
}

// RelayUpdatePacket is the struct wrapping a update packet
type RelayUpdatePacket struct {
	Version   uint32
	Address   net.UDPAddr
	Token     []byte
	NumRelays uint32

	PingStats []routing.RelayStatsPing

	BytesReceived uint64
}

type packetMetadata struct {
	ID          uint64 `json:"Id"` // OLD RELAY ONLY: not the right id for typical use, this is a fnv1a-64 hash of its name (firestore id), not the address which reference/new relays use
	PingKey     []byte `json:"PingKey"`
	Role        string `json:"Role"`
	Shutdown    bool   `json:"Shutdown"`
	Group       string `json:"Group"`
	TokenBase64 string `json:"PublicKey"`
}

type trafficStats struct {
	BytesPaidTx        uint64 `json:"BytesPaidTx"`
	BytesPaidRx        uint64 `json:"BytesPaidRx"`
	BytesManagementTx  uint64 `json:"BytesManagementTx"`
	BytesManagementRx  uint64 `json:"BytesManagementRx"`
	BytesMeasurementTx uint64 `json:"BytesMeasurementTx"`
	BytesMeasurementRx uint64 `json:"BytesMeasurementRx"`
	BytesInvalidRx     uint64 `json:"BytesInvalidRx"`
	SessionCount       uint64 `json:"SessionCount"`
	FlowCount          uint64 `json:"FlowCount"`
}

type RelayUpdateRequestJSON struct {
	Version      uint32                   `json:"version"`
	StringAddr   string                   `json:"relay_address"`
	Metadata     packetMetadata           `json:"Metadata"`
	Timestamp    uint64                   `json:"Timestamp"`
	Signature    []byte                   `json:"Signature"`
	Usage        float32                  `json:"Usage"`
	TrafficStats trafficStats             `json:"TrafficStats"`
	PingStats    []routing.RelayStatsPing `json:"PingStats"`
	RelayName    string                   `json:"relay_name"`
}
type RelayUpdateResponseJSON struct {
	Version      uint32                   `json:"version"`
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

	packet.BytesReceived = j.TrafficStats.BytesMeasurementRx

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

	if !encoding.ReadUint64(buff, &index, &r.BytesReceived) {
		return errors.New("invalid packet")
	}

	return nil
}

// MarshalBinary ...
func (r RelayUpdatePacket) MarshalBinary() ([]byte, error) {
	data := make([]byte, r.size())

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

	encoding.WriteUint64(data, &index, r.BytesReceived)

	return data, nil
}

func (r *RelayUpdatePacket) size() uint {
	return uint(4 + 4 + len(r.Address.String()) + routing.EncryptedTokenSize + 4 + 20*len(r.PingStats) + 8)
}
