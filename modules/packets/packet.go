package packets

import (
	"github.com/networknext/backend/modules/common"
)

type Packet interface {
	Serialize(stream common.Stream) error
}
