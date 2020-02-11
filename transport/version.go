package transport

import "fmt"

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

var (
	SDKVersionMin = SDKVersion{3, 3, 2}
	SDKVersionMax = SDKVersion{254, 1023, 254}
)

func (a SDKVersion) Compare(b SDKVersion) int {
	if a.Major > b.Major {
		return SDKVersionNewer
	}
	if a.Major == b.Major {
		if a.Minor > b.Minor {
			return SDKVersionNewer
		}

		if a.Minor == b.Minor {
			if a.Patch > b.Patch {
				return SDKVersionNewer
			}

			return SDKVersionEqual
		}
	}

	return SDKVersionOlder
}

func (v SDKVersion) IsInternal() bool {
	if v.Major == 0 && v.Minor == 0 && v.Patch == 0 {
		return true
	}
	return false
}

func (v SDKVersion) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func (a SDKVersion) AtLeast(b SDKVersion) bool {
	return a.Compare(b) == SDKVersionNewer || a.Compare(b) == SDKVersionEqual
}
