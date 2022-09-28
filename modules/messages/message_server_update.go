package messages

import (
	"fmt"

	"cloud.google.com/go/bigquery"

	"github.com/networknext/backend/modules-old/encoding"
)

const (
	ServerUpdateMessageVersion = uint8(0)
	MaxServerUpdateMessageSize = 128
)

type ServerUpdateMessage struct {
	Version    byte
	// todo
}

func (message *ServerUpdateMessage) Read(buffer []byte) error {

	index := 0

	if !encoding.ReadUint8(buffer, &index, &message.Version) {
		return fmt.Errorf("failed to read server update version")
	}

	// todo

	return nil
}

func (message *ServerUpdateMessage) Write(buffer []byte) []byte {

	index := 0

	encoding.WriteUint8(buffer, &index, message.Version)

	// todo

	return buffer[:index]
}

func (message *ServerUpdateMessage) Save() (map[string]bigquery.Value, string, error) {

	bigquery_message := make(map[string]bigquery.Value)

	// todo

	return bigquery_message, "", nil
}
