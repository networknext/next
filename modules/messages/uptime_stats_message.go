package messages

import (
	"fmt"

	"cloud.google.com/go/bigquery"

	"github.com/networknext/backend/modules/encoding"
)

const (
	UptimeStatsMessageVersion_Min   = 1
	UptimeStatsMessageVersion_Max   = 1
	UptimeStatsMessageVersion_Write = 1

	MaxUptimeStatsMessageBytes = 1024

	MaxServiceNameLength = 256
)

type UptimeStatsMessage struct {
	Version      uint8
	Timestamp    uint64
	ServiceName  string
	Up           bool
	ResponseTime int
}

func (message *UptimeStatsMessage) Read(buffer []byte) error {

	index := 0

	if !encoding.ReadUint8(buffer, &index, &message.Version) {
		return fmt.Errorf("failed to read version")
	}

	if message.Version < UptimeStatsMessageVersion_Min || message.Version > UptimeStatsMessageVersion_Max {
		return fmt.Errorf("invalid uptime stats message version %d", message.Version)
	}

	if !encoding.ReadUint64(buffer, &index, &message.Timestamp) {
		return fmt.Errorf("failed to read timestamp")
	}

	if !encoding.ReadString(buffer, &index, &message.ServiceName, MaxServiceNameLength) {
		return fmt.Errorf("failed to read up status")
	}

	if !encoding.ReadBool(buffer, &index, &message.Up) {
		return fmt.Errorf("failed to read up status")
	}

	if !encoding.ReadInt(buffer, &index, &message.ResponseTime) {
		return fmt.Errorf("failed to read response time")
	}

	return nil
}

func (message *UptimeStatsMessage) Write(buffer []byte) []byte {

	index := 0

	if message.Version < UptimeStatsMessageVersion_Min || message.Version > UptimeStatsMessageVersion_Max {
		panic(fmt.Sprintf("invalid uptime stats message version %d", message.Version))
	}

	encoding.WriteUint8(buffer, &index, message.Version)
	encoding.WriteUint64(buffer, &index, message.Timestamp)
	encoding.ReadString(buffer, &index, &message.ServiceName, MaxServiceNameLength)
	encoding.WriteBool(buffer, &index, message.Up)
	encoding.WriteInt(buffer, &index, message.ResponseTime)

	return buffer[:index]
}

func (message *UptimeStatsMessage) Save() (map[string]bigquery.Value, string, error) {

	bigquery_message := make(map[string]bigquery.Value)

	bigquery_message["timestamp"] = int(message.Timestamp)
	bigquery_message["service_name"] = message.ServiceName
	bigquery_message["up"] = message.Up
	bigquery_message["response_time"] = message.ResponseTime

	return bigquery_message, "", nil
}
