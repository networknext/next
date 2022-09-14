package packets

import (
	"github.com/networknext/backend/modules/encoding"
)

type Packet interface {

	Serialize(stream encoding.Stream) error	
}
