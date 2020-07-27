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

// RelayPingStats describes the measured relay ping statistics to another relay
type RelayPingStats struct {
	ID         uint64  `json:"id"`
	Address    string  `json:"address"`
	RTT        float32 `json:"rtt"`
	Jitter     float32 `json:"jitter"`
	PacketLoss float32 `json:"packet_loss"`
}

// RelayRequest describes the packets coming into and going out of the relays endpoint
type RelayRequest struct {
	Address      net.UDPAddr
	PingStats    []RelayPingStats
	TrafficStats routing.RelayTrafficStats
	ShuttingDown bool
}

func (r *RelayRequest) UnmarshalJSON(buf []byte) error {
	data := make(map[string]interface{})

	// Unmarshal the JSON into map
	if err := json.Unmarshal(buf, &data); err != nil {
		return err
	}

	// Get the address from the map
	if addr, ok := data["address"].(string); ok {
		if udpAddr, err := net.ResolveUDPAddr("udp", addr); err == nil {
			r.Address = *udpAddr
		} else {
			return err
		}
	}

	// Get the ping stats from the map
	pingStats := make([]RelayPingStats, 0)
	if pingStatsArray, ok := data["ping_stats"].([]interface{}); ok {
		for _, v := range pingStatsArray { // Loop through the array of ping stats
			if mapStats, ok := v.(map[string]interface{}); ok { // If we find one, then parse each field
				stats := RelayPingStats{}

				if id, ok := mapStats["id"].(string); ok {
					id, err := strconv.ParseUint(id, 10, 64)
					if err != nil {
						return err
					}

					stats.ID = id
				}

				if addr, ok := mapStats["address"].(string); ok {
					stats.Address = addr
				}

				// Validate the address is correctly formatted
				if addr, ok := mapStats["address"].(string); ok {
					if udpAddr, err := net.ResolveUDPAddr("udp", addr); err == nil {
						stats.Address = udpAddr.String()
					} else {
						return err
					}
				}

				if rtt, ok := mapStats["rtt"].(float64); ok {
					stats.RTT = float32(rtt)
				}

				if jitter, ok := mapStats["jitter"].(float64); ok {
					stats.Jitter = float32(jitter)
				}

				if packetLoss, ok := mapStats["packet_loss"].(float64); ok {
					stats.PacketLoss = float32(packetLoss)
				}

				// Add the ping stat to the working list of ping stats
				pingStats = append(pingStats, stats)
			}
		}

		// Finally, save the working list of ping stats to the relay request
		r.PingStats = pingStats
	}

	// Get the traffic stats from the map
	if trafficStatsMap, ok := data["traffic_stats"].(map[string]interface{}); ok {
		if sessionCount, ok := trafficStatsMap["session_count"].(float64); ok {
			r.TrafficStats.SessionCount = uint64(sessionCount)
		}

		if bytesTx, ok := trafficStatsMap["bytes_tx"].(float64); ok {
			r.TrafficStats.BytesSent = uint64(bytesTx)
		}

		if bytesRx, ok := trafficStatsMap["bytes_rx"].(float64); ok {
			r.TrafficStats.BytesReceived = uint64(bytesRx)
		}
	}

	// Get the shutting down flag from the map
	if shuttingDown, ok := data["shutting_down"].(bool); ok {
		r.ShuttingDown = shuttingDown
	}

	return nil
}

func (r RelayRequest) MarshalJSON() ([]byte, error) {
	data := make(map[string]interface{})

	data["address"] = r.Address.String()

	pingStats := make([]map[string]interface{}, 0)
	for _, stat := range r.PingStats {
		stats := make(map[string]interface{})
		stats["id"] = strconv.FormatUint(stat.ID, 10)
		stats["address"] = stat.Address
		stats["rtt"] = stat.RTT
		stats["jitter"] = stat.Jitter
		stats["packet_loss"] = stat.PacketLoss
		pingStats = append(pingStats, stats)
	}
	data["ping_stats"] = pingStats

	trafficStats := make(map[string]interface{})
	trafficStats["session_count"] = r.TrafficStats.SessionCount
	trafficStats["bytes_tx"] = r.TrafficStats.BytesSent
	trafficStats["bytes_rx"] = r.TrafficStats.BytesReceived
	data["traffic_stats"] = trafficStats
	data["shutting_down"] = r.ShuttingDown

	return json.Marshal(data)
}

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

	CPUUsage float64
	MemUsage float64

	ShuttingDown bool
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

	if index+1 > len(buff) {
		return errors.New("invalid packet, could not read shutdown flag")
	}

	r.ShuttingDown = buff[index] != 0

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

	if r.ShuttingDown {
		data[index] = 1
	}

	return data, nil
}

func (r *RelayUpdateRequest) size() uint {
	return uint(4 + 4 + len(r.Address.String()) + crypto.KeySize + 4 + 20*len(r.PingStats) + 8 + 8 + 8 + 1)
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
	encoding.WriteUint32(responseData, &index, uint32(len(r.RelaysToPing)))
	for i := range r.RelaysToPing {
		encoding.WriteUint64(responseData, &index, r.RelaysToPing[i].RelayPingData.ID)
		encoding.WriteString(responseData, &index, r.RelaysToPing[i].RelayPingData.Address, routing.MaxRelayAddressLength)
	}

	return responseData, nil
}

func (r *RelayUpdateResponse) size() int {
	return 4 + 4 + (4+routing.MaxRelayAddressLength)*len(r.RelaysToPing)
}
