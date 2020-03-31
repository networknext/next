package transport

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net"
	"strconv"

	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/encoding"
	"github.com/networknext/backend/routing"
	"github.com/tidwall/gjson"
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
		if udpAddr, err := net.ResolveUDPAddr("udp", addr); err == nil {
			r.Address = *udpAddr
		} else {
			return err
		}
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
		encoding.ReadBytes(buf, &index, &r.EncryptedToken, routing.EncryptedRelayTokenSize)) {
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
	data := make([]byte, 4+4+crypto.NonceSize+4+len(r.Address.String())+routing.EncryptedRelayTokenSize)
	index := 0
	encoding.WriteUint32(data, &index, r.Magic)
	encoding.WriteUint32(data, &index, r.Version)
	encoding.WriteBytes(data, &index, r.Nonce, crypto.NonceSize)
	encoding.WriteString(data, &index, r.Address.String(), uint32(len(r.Address.String())))
	encoding.WriteBytes(data, &index, r.EncryptedToken, routing.EncryptedRelayTokenSize)

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

type RelayUpdateRequest struct {
	Version uint32
	Address net.UDPAddr
	Token   []byte

	PingStats []routing.RelayStatsPing

	BytesReceived uint64

	ShuttingDown bool
}

func (r RelayUpdateRequest) MarshalJSON() ([]byte, error) {
	data := make(map[string]interface{})

	data["version"] = r.Version
	data["relay_address"] = r.Address.String()

	meta := make(map[string]interface{})
	meta["PublicKey"] = base64.StdEncoding.EncodeToString(r.Token)
	data["Metadata"] = meta

	stats := make(map[string]interface{})
	stats["BytesMeasurementRx"] = r.BytesReceived
	data["TrafficStats"] = stats

	data["PingStats"] = r.PingStats

	data["shutting_down"] = r.ShuttingDown

	return json.Marshal(data)
}

func (r *RelayUpdateRequest) UnmarshalJSON(buff []byte) error {
	var err error

	doc := gjson.ParseBytes(buff)

	r.Version = uint32(doc.Get("version").Int())

	addr := doc.Get("relay_address").String()
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return err
	}
	r.Address.IP = net.ParseIP(host)
	if r.Address.IP == nil {
		return errors.New("invalid relay_address")
	}
	iport, err := strconv.ParseInt(port, 10, 64)
	if err != nil {
		return err
	}
	r.Address.Port = int(iport)

	r.Token, err = base64.StdEncoding.DecodeString(doc.Get("Metadata.PublicKey").String())
	if err != nil {
		return err
	}

	if len(r.Token) != crypto.KeySize {
		return errors.New("invalid token size")
	}

	r.BytesReceived = uint64(doc.Get("TrafficStats.BytesMeasurementRx").Int())

	r.PingStats = make([]routing.RelayStatsPing, 0)
	if err := json.Unmarshal([]byte(doc.Get("PingStats").Raw), &r.PingStats); err != nil {
		return err
	}

	r.ShuttingDown = doc.Get("shutting_down").Bool()

	return nil
}

func (r *RelayUpdateRequest) UnmarshalBinary(buff []byte) error {
	var numRelays uint32

	index := 0
	var addr string
	if !(encoding.ReadUint32(buff, &index, &r.Version) &&
		encoding.ReadString(buff, &index, &addr, MaxRelayAddressLength) &&
		encoding.ReadBytes(buff, &index, &r.Token, crypto.KeySize) &&
		encoding.ReadUint32(buff, &index, &numRelays)) {
		return errors.New("invalid packet")
	}

	if udp, err := net.ResolveUDPAddr("udp", addr); udp != nil && err == nil {
		r.Address = *udp
	} else {
		return fmt.Errorf("could not resolve init packet with address '%s' with reason: %v", addr, err)
	}

	r.PingStats = make([]routing.RelayStatsPing, numRelays)
	for i := 0; i < int(numRelays); i++ {
		stats := &r.PingStats[i]

		if !(encoding.ReadUint64(buff, &index, &stats.RelayID) &&
			encoding.ReadFloat32(buff, &index, &stats.RTT) &&
			encoding.ReadFloat32(buff, &index, &stats.Jitter) &&
			encoding.ReadFloat32(buff, &index, &stats.PacketLoss)) {
			return errors.New("invalid packet, could not read a ping stat")
		}
	}

	if !encoding.ReadUint64(buff, &index, &r.BytesReceived) {
		return errors.New("invalid packet, could not read bytes received")
	}

	if index+1 > len(buff) {
		return errors.New("invalid packet, could not read shutdown flag")
	}

	r.ShuttingDown = buff[index] != 0

	return nil
}

// MarshalBinary ...
func (r RelayUpdateRequest) MarshalBinary() ([]byte, error) {
	data := make([]byte, r.size())

	index := 0
	encoding.WriteUint32(data, &index, r.Version)
	encoding.WriteString(data, &index, r.Address.String(), math.MaxInt32)
	encoding.WriteBytes(data, &index, r.Token, crypto.KeySize)
	encoding.WriteUint32(data, &index, uint32(len(r.PingStats)))

	for i := 0; i < len(r.PingStats); i++ {
		stats := &r.PingStats[i]

		encoding.WriteUint64(data, &index, stats.RelayID)
		encoding.WriteUint32(data, &index, math.Float32bits(stats.RTT))
		encoding.WriteUint32(data, &index, math.Float32bits(stats.Jitter))
		encoding.WriteUint32(data, &index, math.Float32bits(stats.PacketLoss))
	}

	encoding.WriteUint64(data, &index, r.BytesReceived)

	if r.ShuttingDown {
		data[index] = 1
	}

	return data, nil
}

func (r *RelayUpdateRequest) size() uint {
	return uint(4 + 4 + len(r.Address.String()) + routing.EncryptedRelayTokenSize + 4 + 20*len(r.PingStats) + 8)
}

type RelayUpdateResponse struct {
	RelaysToPing []routing.LegacyPingData `json:"ping_data"`
}

func (r RelayUpdateResponse) MarshalJSON() ([]byte, error) {
	data := make(map[string]interface{})

	data["ping_data"] = r.RelaysToPing

	return json.Marshal(data)
}

func (r RelayUpdateResponse) MarshalBinary() ([]byte, error) {
	index := 0
	responseData := make([]byte, r.size())

	encoding.WriteUint32(responseData, &index, VersionNumberUpdateResponse)
	encoding.WriteUint32(responseData, &index, uint32(len(r.RelaysToPing)))
	for i := range r.RelaysToPing {
		encoding.WriteUint64(responseData, &index, r.RelaysToPing[i].ID)
		encoding.WriteString(responseData, &index, r.RelaysToPing[i].Address, MaxRelayAddressLength)
	}

	return responseData, nil
}

func (r *RelayUpdateResponse) size() int {
	return 8 + (8+MaxRelayAddressLength)*len(r.RelaysToPing)
}
