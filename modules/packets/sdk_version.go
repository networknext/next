package packets

import (
	"fmt"
	"github.com/networknext/accelerate/modules/encoding"
)

const (
	SDKVersionEqual = iota
	SDKVersionOlder
	SDKVersionNewer
)

type SDKVersion struct {
	Major int32
	Minor int32
	Patch int32
}

func (version *SDKVersion) Serialize(stream encoding.Stream) error {
	stream.SerializeInteger(&version.Major, 0, 255)
	stream.SerializeInteger(&version.Minor, 0, 255)
	stream.SerializeInteger(&version.Patch, 0, 255)
	return stream.Error()
}

func (a SDKVersion) Compare(b SDKVersion) int {

	if a.Major > b.Major {
		return SDKVersionNewer
	} else if a.Major < b.Major {
		return SDKVersionOlder
	}

	if a.Minor > b.Minor {
		return SDKVersionNewer
	} else if a.Minor < b.Minor {
		return SDKVersionOlder
	}

	if a.Patch > b.Patch {
		return SDKVersionNewer
	} else if a.Patch < b.Patch {
		return SDKVersionOlder
	}

	return SDKVersionEqual
}

func (a SDKVersion) AtLeast(b SDKVersion) bool {
	return a.Compare(b) != SDKVersionOlder
}

func (v SDKVersion) String() string {
	if v.Major == 255 {
		return "internal"
	} else {
		return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	}
}
