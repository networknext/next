package messages

import (
	"fmt"
	"net"

	"cloud.google.com/go/bigquery"

	"github.com/networknext/next/modules/constants"
	"github.com/networknext/next/modules/encoding"
)

const (
	AnalyticsServerInitMessageVersion_Min   = 1
	AnalyticsServerInitMessageVersion_Max   = 2
	AnalyticsServerInitMessageVersion_Write = 1
)

type AnalyticsServerInitMessage struct {
	Version          byte
	Timestamp        uint64
	SDKVersion_Major byte
	SDKVersion_Minor byte
	SDKVersion_Patch byte
	BuyerId          uint64
	DatacenterId     uint64
	DatacenterName   string
	ServerAddress    net.UDPAddr
}

func (message *AnalyticsServerInitMessage) GetMaxSize() int {
	return 64 + constants.MaxDatacenterNameLength
}

func (message *AnalyticsServerInitMessage) Read(buffer []byte) error {

	index := 0

	if !encoding.ReadUint8(buffer, &index, &message.Version) {
		return fmt.Errorf("failed to read analytics server init message version")
	}

	if message.Version < AnalyticsServerInitMessageVersion_Min || message.Version > AnalyticsServerInitMessageVersion_Max {
		return fmt.Errorf("invalid analytics server init message version %d", message.Version)
	}

	if !encoding.ReadUint64(buffer, &index, &message.Timestamp) {
		return fmt.Errorf("failed to read timestamp")
	}

	if !encoding.ReadUint8(buffer, &index, &message.SDKVersion_Major) {
		return fmt.Errorf("failed to read sdk version major")
	}

	if !encoding.ReadUint8(buffer, &index, &message.SDKVersion_Minor) {
		return fmt.Errorf("failed to read sdk version major")
	}

	if !encoding.ReadUint8(buffer, &index, &message.SDKVersion_Patch) {
		return fmt.Errorf("failed to read sdk version major")
	}

	if !encoding.ReadUint64(buffer, &index, &message.BuyerId) {
		return fmt.Errorf("failed to read buyer id")
	}

	if !encoding.ReadUint64(buffer, &index, &message.DatacenterId) {
		return fmt.Errorf("failed to read datacenter id")
	}

	if !encoding.ReadString(buffer, &index, &message.DatacenterName, constants.MaxDatacenterNameLength) {
		return fmt.Errorf("failed to read datacenter name")
	}

	if message.Version >= 2 {

		if !encoding.ReadAddress(buffer, &index, &message.ServerAddress) {
			return fmt.Errorf("failed to read server address")
		}

	}

	return nil
}

func (message *AnalyticsServerInitMessage) Write(buffer []byte) []byte {

	index := 0

	if message.Version < AnalyticsServerInitMessageVersion_Min || message.Version > AnalyticsServerInitMessageVersion_Max {
		panic(fmt.Sprintf("invalid analytics server init message version %d", message.Version))
	}

	encoding.WriteUint8(buffer, &index, message.Version)
	encoding.WriteUint64(buffer, &index, message.Timestamp)
	encoding.WriteUint8(buffer, &index, message.SDKVersion_Major)
	encoding.WriteUint8(buffer, &index, message.SDKVersion_Minor)
	encoding.WriteUint8(buffer, &index, message.SDKVersion_Patch)
	encoding.WriteUint64(buffer, &index, message.BuyerId)
	encoding.WriteUint64(buffer, &index, message.DatacenterId)
	encoding.WriteString(buffer, &index, message.DatacenterName, constants.MaxDatacenterNameLength)

	if message.Version >= 2 {
		encoding.WriteAddress(buffer, &index, &message.ServerAddress)
	}

	return buffer[:index]
}

func (message *AnalyticsServerInitMessage) Save() (map[string]bigquery.Value, string, error) {
	bigquery_entry := make(map[string]bigquery.Value)
	bigquery_entry["timestamp"] = int(message.Timestamp)
	bigquery_entry["sdk_version_major"] = int(message.SDKVersion_Major)
	bigquery_entry["sdk_version_minor"] = int(message.SDKVersion_Minor)
	bigquery_entry["sdk_version_patch"] = int(message.SDKVersion_Patch)
	bigquery_entry["buyer_id"] = int(message.BuyerId)
	bigquery_entry["datacenter_id"] = int(message.DatacenterId)
	bigquery_entry["datacenter_name"] = message.DatacenterName
	bigquery_entry["server_address"] = message.ServerAddress.String()
	return bigquery_entry, "", nil
}
