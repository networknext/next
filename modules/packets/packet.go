package packets

import (
	"github.com/networknext/accelerate/modules/encoding"
)

type Packet interface {
	Serialize(stream encoding.Stream) error
}

func WritePacket[P Packet](packetData []byte, packetObject P) ([]byte, error) {
	writeStream := encoding.CreateWriteStream(packetData)
	err := packetObject.Serialize(writeStream)
	writeStream.Flush()
	return packetData[:writeStream.GetBytesProcessed()], err
}

func ReadPacket[P Packet](packetData []byte, packetObject P) error {
	readStream := encoding.CreateReadStream(packetData)
	return packetObject.Serialize(readStream)
}
