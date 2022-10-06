package messages

import (
    "fmt"

    "cloud.google.com/go/bigquery"

    "github.com/networknext/backend/modules/encoding"
)

const (
    PingStatsMessageVersion = 4
    MaxPingStatsMessageSize = 128
)

type PingStatsMessage struct {
    Version    uint8
    Timestamp  uint64
    RelayA     uint64
    RelayB     uint64
    RTT        float32
    Jitter     float32
    PacketLoss float32
    Routable   bool
}

func (message *PingStatsMessage) Read(buffer []byte) error {

    index := 0

    if !encoding.ReadUint8(buffer, &index, &message.Version) {
        return fmt.Errorf("failed to read ping stats Version")
    }

    if !encoding.ReadUint64(buffer, &index, &message.Timestamp) {
        return fmt.Errorf("failed to read ping stats Timestamp")
    }

    if !encoding.ReadUint64(buffer, &index, &message.RelayA) {
        return fmt.Errorf("failed to read ping stats RelayA")
    }

    if !encoding.ReadUint64(buffer, &index, &message.RelayB) {
        return fmt.Errorf("failed to read ping stats RelayB")
    }

    if !encoding.ReadFloat32(buffer, &index, &message.RTT) {
        return fmt.Errorf("failed to read ping stats RTT")
    }

    if !encoding.ReadFloat32(buffer, &index, &message.Jitter) {
        return fmt.Errorf("failed to read ping stats Jitter")
    }

    if !encoding.ReadFloat32(buffer, &index, &message.PacketLoss) {
        return fmt.Errorf("failed to read ping stats PacketLoss")
    }

    if message.Version >= 2 {
        if !encoding.ReadBool(buffer, &index, &message.Routable) {
            return fmt.Errorf("failed to read ping stats Routable")
        }
    }

    return nil
}

func (message *PingStatsMessage) Write(buffer []byte) []byte {

    index := 0

    encoding.WriteUint8(buffer, &index, message.Version)
    encoding.WriteUint64(buffer, &index, message.Timestamp)
    encoding.WriteUint64(buffer, &index, message.RelayA)
    encoding.WriteUint64(buffer, &index, message.RelayB)
    encoding.WriteFloat32(buffer, &index, message.RTT)
    encoding.WriteFloat32(buffer, &index, message.Jitter)
    encoding.WriteFloat32(buffer, &index, message.PacketLoss)
    encoding.WriteBool(buffer, &index, message.Routable)

    return buffer[:index]
}

func (message *PingStatsMessage) Save() (map[string]bigquery.Value, string, error) {

    bigquery_message := make(map[string]bigquery.Value)

    bigquery_message["timestamp"] = int(message.Timestamp)
    bigquery_message["relay_a"] = int(message.RelayA)
    bigquery_message["relay_b"] = int(message.RelayB)
    bigquery_message["rtt"] = message.RTT
    bigquery_message["jitter"] = message.Jitter
    bigquery_message["packet_loss"] = message.PacketLoss
    bigquery_message["routable"] = message.Routable

    return bigquery_message, "", nil
}
