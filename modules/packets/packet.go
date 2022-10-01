package packets

import (
    "github.com/networknext/backend/modules/encoding"
)

type Packet interface {
    Serialize(stream encoding.Stream) error
}

func ReadPacket[P Packet](packetData []byte, packetObject P) error {
    readStream := encoding.CreateReadStream(packetData)
    return packetObject.Serialize(readStream)
}
