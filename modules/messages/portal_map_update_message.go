package messages

import (
	"fmt"

	"github.com/networknext/accelerate/modules/encoding"
)

const (
	PortalMapUpdateMessageVersion_Min   = 1
	PortalMapUpdateMessageVersion_Max   = 1
	PortalMapUpdateMessageVersion_Write = 1
)

type PortalMapUpdateMessage struct {
	Version   byte
	SessionId uint64
	Latitude  float32
	Longitude float32
	Next      bool
}

func (message *PortalMapUpdateMessage) GetMaxSize() int {
	return 32
}

func (message *PortalMapUpdateMessage) Write(buffer []byte) []byte {

	index := 0

	if message.Version < PortalMapUpdateMessageVersion_Min || message.Version > PortalMapUpdateMessageVersion_Max {
		panic(fmt.Sprintf("invalid portal map update message version %d", message.Version))
	}

	encoding.WriteUint8(buffer, &index, message.Version)
	encoding.WriteUint64(buffer, &index, message.SessionId)
	encoding.WriteFloat32(buffer, &index, message.Latitude)
	encoding.WriteFloat32(buffer, &index, message.Longitude)
	encoding.WriteBool(buffer, &index, message.Next)

	return buffer[:index]
}

func (message *PortalMapUpdateMessage) Read(buffer []byte) error {

	index := 0

	if !encoding.ReadUint8(buffer, &index, &message.Version) {
		return fmt.Errorf("failed to read portal map update message version")
	}

	if message.Version < PortalMapUpdateMessageVersion_Min || message.Version > PortalMapUpdateMessageVersion_Max {
		return fmt.Errorf("invalid portal map update message version %d", message.Version)
	}

	if !encoding.ReadUint64(buffer, &index, &message.SessionId) {
		return fmt.Errorf("failed to read session id")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.Latitude) {
		return fmt.Errorf("failed to read latitude")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.Longitude) {
		return fmt.Errorf("failed to read longitude")
	}

	if !encoding.ReadBool(buffer, &index, &message.Next) {
		return fmt.Errorf("failed to read next")
	}

	return nil
}
