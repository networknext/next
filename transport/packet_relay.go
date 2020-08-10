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

	MaxRelayVersionLength = 10
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

func (r RelayInitRequest) MarshalJSON() ([]byte, error) {
	data := make(map[string]interface{})

	data["magic_request_protection"] = r.Magic
	data["version"] = r.Version
	data["nonce"] = r.Nonce
	data["relay_address"] = r.Address.String()
	data["encrypted_token"] = r.EncryptedToken

	return json.Marshal(data)
}

// UnmarshalBinary decodes binary data into a RelayInitRequest struct
func (r *RelayInitRequest) UnmarshalBinary(buf []byte) error {
	index := 0
	var addr string
	if !(encoding.ReadUint32(buf, &index, &r.Magic) &&
		encoding.ReadUint32(buf, &index, &r.Version) &&
		encoding.ReadBytes(buf, &index, &r.Nonce, crypto.NonceSize) &&
		encoding.ReadString(buf, &index, &addr, routing.MaxRelayAddressLength) &&
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

	data["Version"] = VersionNumberInitResponse
	data["Timestamp"] = r.Timestamp
	data["PublicKey"] = r.PublicKey

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

func (r *RelayInitResponse) UnmarshalBinary(buf []byte) error {
	indx := 0

	var version uint32
	if !encoding.ReadUint32(buf, &indx, &version) {
		return errors.New("failed to unmarshal relay init response version")
	}

	var timestamp uint64
	if !encoding.ReadUint64(buf, &indx, &timestamp) {
		return errors.New("failed to unmarshal relay init response timestamp")
	}

	var publicKey []byte
	if !encoding.ReadBytes(buf, &indx, &publicKey, crypto.KeySize) {
		return errors.New("failed to unmarshal relay init response public key")
	}

	r.Version = version
	r.Timestamp = timestamp
	r.PublicKey = publicKey

	return nil
}

type RelayUpdateRequest struct {
	Version      uint32
	RelayVersion string
	Address      net.UDPAddr
	Token        []byte

	PingStats    []routing.RelayStatsPing
	TrafficStats routing.RelayTrafficStats

	ShuttingDown bool

	CPUUsage float64
	MemUsage float64
}

func (r *RelayUpdateRequest) UnmarshalJSON(buff []byte) error {
	var err error

	doc := gjson.ParseBytes(buff)

	r.Version = uint32(doc.Get("version").Int())
	r.RelayVersion = doc.Get("relay_version").String()

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

	r.TrafficStats.SessionCount = uint64(doc.Get("TrafficStats.SessionCount").Int())
	r.TrafficStats.BytesSent = uint64(doc.Get("TrafficStats.BytesMeasurementTx").Int())
	r.TrafficStats.BytesReceived = uint64(doc.Get("TrafficStats.BytesMeasurementRx").Int())

	r.PingStats = make([]routing.RelayStatsPing, 0)
	if err := json.Unmarshal([]byte(doc.Get("PingStats").Raw), &r.PingStats); err != nil {
		return err
	}

	r.ShuttingDown = doc.Get("shutting_down").Bool()

	r.CPUUsage = doc.Get("sys_stats.cpu_usage").Float()
	r.MemUsage = doc.Get("sys_stats.mem_usage").Float()

	return nil
}

func (r *RelayUpdateRequest) UnmarshalBinary(buff []byte) error {
	var numRelays uint32

	index := 0
	var addr string
	if !(encoding.ReadUint32(buff, &index, &r.Version) &&
		encoding.ReadString(buff, &index, &addr, routing.MaxRelayAddressLength) &&
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

	if !encoding.ReadUint64(buff, &index, &r.TrafficStats.SessionCount) {
		return errors.New("invalid packet, could not read session count")
	}

	if !encoding.ReadUint64(buff, &index, &r.TrafficStats.BytesSent) {
		return errors.New("invalid packet, could not read bytes sent")
	}

	if !encoding.ReadUint64(buff, &index, &r.TrafficStats.BytesReceived) {
		return errors.New("invalid packet, could not read bytes received")
	}

	var shuttingDown uint8
	if !encoding.ReadUint8(buff, &index, &shuttingDown) {
		return errors.New("invalid packet, could not read shutdown flag")
	}

	r.ShuttingDown = shuttingDown != 0

	if !encoding.ReadFloat64(buff, &index, &r.CPUUsage) {
		return errors.New("invalid packet, could not read cpu usage")
	}

	if !encoding.ReadFloat64(buff, &index, &r.MemUsage) {
		return errors.New("invalid packet, could not read memory usage")
	}

	if !encoding.ReadString(buff, &index, &r.RelayVersion, MaxRelayVersionLength) {
		return errors.New("invalid packet, could not read relay version")
	}

	return nil
}

func (r RelayUpdateRequest) MarshalJSON() ([]byte, error) {
	data := make(map[string]interface{})

	data["version"] = r.Version
	data["relay_address"] = r.Address.String()

	metadata := make(map[string]interface{})
	metadata["PublicKey"] = r.Token
	data["Metadata"] = metadata

	data["PingStats"] = r.PingStats

	trafficStats := make(map[string]interface{})
	trafficStats["SessionCount"] = r.TrafficStats.SessionCount
	trafficStats["BytesMeasurementTx"] = r.TrafficStats.BytesSent
	trafficStats["BytesMeasurementRx"] = r.TrafficStats.BytesReceived
	data["TrafficStats"] = trafficStats

	data["shutting_down"] = r.ShuttingDown

	return json.Marshal(data)
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

	encoding.WriteUint64(data, &index, r.TrafficStats.SessionCount)
	encoding.WriteUint64(data, &index, r.TrafficStats.BytesSent)
	encoding.WriteUint64(data, &index, r.TrafficStats.BytesReceived)
	var shutdownFlag uint8
	if r.ShuttingDown {
		shutdownFlag = 1
	} else {
		shutdownFlag = 0
	}
	encoding.WriteUint8(data, &index, shutdownFlag)
	encoding.WriteFloat64(data, &index, r.CPUUsage)
	encoding.WriteFloat64(data, &index, r.MemUsage)
	encoding.WriteString(data, &index, r.RelayVersion, uint32(len(r.RelayVersion)))

	return data[:index], nil
}

func (r *RelayUpdateRequest) size() uint {
	return uint(4 + 4 + len(r.Address.String()) + crypto.KeySize + 4 + 20*len(r.PingStats) + 8 + 8 + 8 + 1 + 8 + 8 + 4 + len(r.RelayVersion))
}

type RelayUpdateResponse struct {
	Timestamp    int64
	RelaysToPing []routing.LegacyPingData `json:"ping_data"`
}

func (r *RelayUpdateResponse) UnmarshalBinary(buff []byte) error {
	index := 0
	var version uint32
	if !encoding.ReadUint32(buff, &index, &version) {
		return errors.New("failed to unmarshal relay update response version")
	}

	var timestamp uint64
	if !encoding.ReadUint64(buff, &index, &timestamp) {
		return errors.New("failed to unmarshal relay update response timestamp")
	}
	r.Timestamp = int64(timestamp)

	var numRelaysToPing uint32
	if !encoding.ReadUint32(buff, &index, &numRelaysToPing) {
		return errors.New("failed to unmarshal relay update response number of relays to ping")
	}

	for i := 0; uint32(i) < numRelaysToPing; i++ {
		var id uint64
		if !encoding.ReadUint64(buff, &index, &id) {
			return errors.New("failed to unmarshal relay update response relay id")
		}

		var addr string
		if !encoding.ReadString(buff, &index, &addr, routing.MaxRelayAddressLength) {
			return errors.New("failed to unmarshal relay update response relay address")
		}

		r.RelaysToPing = append(r.RelaysToPing, routing.LegacyPingData{
			RelayPingData: routing.RelayPingData{
				ID:      id,
				Address: addr,
			},
		})
	}

	return nil
}

func (r RelayUpdateResponse) MarshalJSON() ([]byte, error) {
	data := make(map[string]interface{})

	data["version"] = VersionNumberUpdateResponse
	data["ping_data"] = r.RelaysToPing
	data["timestamp"] = r.Timestamp

	return json.Marshal(data)
}

func (r RelayUpdateResponse) MarshalBinary() ([]byte, error) {
	index := 0
	responseData := make([]byte, r.size())

	encoding.WriteUint32(responseData, &index, VersionNumberUpdateResponse)
	encoding.WriteUint64(responseData, &index, uint64(r.Timestamp))
	encoding.WriteUint32(responseData, &index, uint32(len(r.RelaysToPing)))
	for i := range r.RelaysToPing {
		encoding.WriteUint64(responseData, &index, r.RelaysToPing[i].RelayPingData.ID)
		encoding.WriteString(responseData, &index, r.RelaysToPing[i].RelayPingData.Address, routing.MaxRelayAddressLength)
	}

	return responseData[:index], nil
}

func (r *RelayUpdateResponse) size() int {
	return 4 + 8 + 4 + len(r.RelaysToPing)*(8+routing.MaxRelayAddressLength)
}
