package messages

import (
	"fmt"

	"cloud.google.com/go/bigquery"
	"github.com/networknext/backend/modules-old/encoding"
)

const (
	ServerInitMessageVersion = uint8(0)
	MaxServerInitMessageSize = 128
)

type ServerInitMessage struct {
	MessageVersion   byte
	SDKVersion_Major byte
	SDKVersion_Minor byte
	SDKVersion_Patch byte
	BuyerId          uint64
	DatacenterId     uint64
	DatacenterName   string
}

func (message *ServerInitMessage) Read(buffer []byte) error {

	index := 0

	if !encoding.ReadUint8(buffer, &index, &message.MessageVersion) {
		return fmt.Errorf("failed to read server init message version")
	}

	// todo: code rest of read

	return nil
}

func (message *ServerInitMessage) Write(buffer []byte) []byte {

	index := 0

	encoding.WriteUint8(buffer, &index, message.MessageVersion)

	// todo: code rest of write

	return buffer[:index]
}

func (message *ServerInitMessage) Save() (map[string]bigquery.Value, string, error) {

	bigquery_message := make(map[string]bigquery.Value)

	// todo: code save method

	return bigquery_message, "", nil
}
