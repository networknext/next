package packets

import (
	"github.com/networknext/backend/modules/common"
)

type Packet interface {
	Serialize(stream common.Stream) error
}

func ReadPacket[P Packet](packetData []byte, packetObject P) error {
	readStream := common.CreateReadStream(packetData)
	return packetObject.Serialize(readStream)
}
