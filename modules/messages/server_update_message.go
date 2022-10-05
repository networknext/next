package messages

import (
    "fmt"

    "cloud.google.com/go/bigquery"

    "github.com/networknext/backend/modules/encoding"
)

const (
    ServerUpdateMessageVersion          = 0
    MaxServerUpdateMessageSize          = 128
    ServerUpdateMaxDatacenterNameLength = 256
)

type ServerUpdateMessage struct {
    MessageVersion   byte
    SDKVersion_Major byte
    SDKVersion_Minor byte
    SDKVersion_Patch byte
    BuyerId          uint64
    DatacenterId     uint64
    DatacenterName   string
}

func (message *ServerUpdateMessage) Read(buffer []byte) error {

    index := 0

    if !encoding.ReadUint8(buffer, &index, &message.MessageVersion) {
        return fmt.Errorf("failed to read server update message version")
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

    if !encoding.ReadString(buffer, &index, &message.DatacenterName, ServerUpdateMaxDatacenterNameLength) {
        return fmt.Errorf("failed to read datacenter name")
    }

    return nil
}

func (message *ServerUpdateMessage) Write(buffer []byte) []byte {

    index := 0

    encoding.WriteUint8(buffer, &index, message.MessageVersion)
    encoding.WriteUint8(buffer, &index, message.SDKVersion_Major)
    encoding.WriteUint8(buffer, &index, message.SDKVersion_Minor)
    encoding.WriteUint8(buffer, &index, message.SDKVersion_Patch)
    encoding.WriteUint64(buffer, &index, message.BuyerId)
    encoding.WriteUint64(buffer, &index, message.DatacenterId)
    encoding.WriteString(buffer, &index, message.DatacenterName, ServerUpdateMaxDatacenterNameLength)

    return buffer[:index]
}

func (message *ServerUpdateMessage) Save() (map[string]bigquery.Value, string, error) {

    bigquery_message := make(map[string]bigquery.Value)

    // todo: code save method

    return bigquery_message, "", nil
}
