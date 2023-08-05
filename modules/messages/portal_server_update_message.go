package messages

import (
	"fmt"
	"net"

	"github.com/networknext/next/modules/encoding"
)

const (
	PortalServerUpdateMessageVersion_Min   = 1
	PortalServerUpdateMessageVersion_Max   = 1
	PortalServerUpdateMessageVersion_Write = 1
)

type PortalServerUpdateMessage struct {
	Version          byte
	Timestamp        uint64
	SDKVersion_Major byte
	SDKVersion_Minor byte
	SDKVersion_Patch byte
	MatchId          uint64
	BuyerId          uint64
	DatacenterId     uint64
	NumSessions      uint32
	StartTime        uint64
	ServerAddress    net.UDPAddr
}

func (message *PortalServerUpdateMessage) GetMaxSize() int {
	return 256
}

func (message *PortalServerUpdateMessage) Write(buffer []byte) []byte {

	index := 0

	if message.Version < PortalServerUpdateMessageVersion_Min || message.Version > PortalServerUpdateMessageVersion_Max {
		panic(fmt.Sprintf("invalid portal server update message version %d", message.Version))
	}

	encoding.WriteUint8(buffer, &index, message.Version)
	encoding.WriteUint64(buffer, &index, message.Timestamp)
	encoding.WriteUint8(buffer, &index, message.SDKVersion_Major)
	encoding.WriteUint8(buffer, &index, message.SDKVersion_Minor)
	encoding.WriteUint8(buffer, &index, message.SDKVersion_Patch)
	encoding.WriteUint64(buffer, &index, message.MatchId)
	encoding.WriteUint64(buffer, &index, message.BuyerId)
	encoding.WriteUint64(buffer, &index, message.DatacenterId)
	encoding.WriteUint32(buffer, &index, message.NumSessions)
	encoding.WriteUint64(buffer, &index, message.StartTime)
	encoding.WriteAddress(buffer, &index, &message.ServerAddress)

	return buffer[:index]
}

func (message *PortalServerUpdateMessage) Read(buffer []byte) error {

	index := 0

	if !encoding.ReadUint8(buffer, &index, &message.Version) {
		return fmt.Errorf("failed to read portal server update message version")
	}

	if message.Version < PortalServerUpdateMessageVersion_Min || message.Version > PortalServerUpdateMessageVersion_Max {
		return fmt.Errorf("invalid portal server update message version %d", message.Version)
	}

	if !encoding.ReadUint64(buffer, &index, &message.Timestamp) {
		return fmt.Errorf("failed to read timestamp")
	}

	if !encoding.ReadUint8(buffer, &index, &message.SDKVersion_Major) {
		return fmt.Errorf("failed to read sdk version major")
	}

	if !encoding.ReadUint8(buffer, &index, &message.SDKVersion_Minor) {
		return fmt.Errorf("failed to read sdk version minor")
	}

	if !encoding.ReadUint8(buffer, &index, &message.SDKVersion_Patch) {
		return fmt.Errorf("failed to read sdk version patch")
	}

	if !encoding.ReadUint64(buffer, &index, &message.MatchId) {
		return fmt.Errorf("failed to read session id")
	}

	if !encoding.ReadUint64(buffer, &index, &message.BuyerId) {
		return fmt.Errorf("failed to read buyer id")
	}

	if !encoding.ReadUint64(buffer, &index, &message.DatacenterId) {
		return fmt.Errorf("failed to read datacenter id")
	}

	if !encoding.ReadUint32(buffer, &index, &message.NumSessions) {
		return fmt.Errorf("failed to read num sessions")
	}

	if !encoding.ReadUint64(buffer, &index, &message.StartTime) {
		return fmt.Errorf("failed to read start time")
	}

	if !encoding.ReadAddress(buffer, &index, &message.ServerAddress) {
		return fmt.Errorf("failed to read server address")
	}

	return nil
}
