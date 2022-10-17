package messages

import (
	"fmt"

	"cloud.google.com/go/bigquery"

	"github.com/networknext/backend/modules/encoding"
)

const (
	ServerInitMessageVersion_Min   = 1
	ServerInitMessageVersion_Max   = 1
	ServerInitMessageVersion_Write = 1

	MaxServerInitMessageSize = 128

	ServerInitMaxDatacenterNameLength = 256
)

type ServerInitMessage struct {
	Version          byte
	SDKVersion_Major byte
	SDKVersion_Minor byte
	SDKVersion_Patch byte
	BuyerId          uint64
	DatacenterId     uint64
	DatacenterName   string
}

func (message *ServerInitMessage) Read(buffer []byte) error {

	index := 0

	if !encoding.ReadUint8(buffer, &index, &message.Version) {
		return fmt.Errorf("failed to read server init message version")
	}

	if message.Version < ServerInitMessageVersion_Min || message.Version > ServerInitMessageVersion_Max {
		return fmt.Errorf("invalid server init message version %d", message.Version)
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

	if !encoding.ReadString(buffer, &index, &message.DatacenterName, ServerInitMaxDatacenterNameLength) {
		return fmt.Errorf("failed to read datacenter name")
	}

	return nil
}

func (message *ServerInitMessage) Write(buffer []byte) []byte {

	index := 0

	if message.Version < ServerInitMessageVersion_Min || message.Version > ServerInitMessageVersion_Max {
		panic(fmt.Sprintf("invalid server init message version %d", message.Version))
	}

	encoding.WriteUint8(buffer, &index, message.Version)
	encoding.WriteUint8(buffer, &index, message.SDKVersion_Major)
	encoding.WriteUint8(buffer, &index, message.SDKVersion_Minor)
	encoding.WriteUint8(buffer, &index, message.SDKVersion_Patch)
	encoding.WriteUint64(buffer, &index, message.BuyerId)
	encoding.WriteUint64(buffer, &index, message.DatacenterId)
	encoding.WriteString(buffer, &index, message.DatacenterName, ServerInitMaxDatacenterNameLength)

	return buffer[:index]
}

func (message *ServerInitMessage) Save() (map[string]bigquery.Value, string, error) {

	bigquery_message := make(map[string]bigquery.Value)

	// todo: code save method

	return bigquery_message, "", nil
}
