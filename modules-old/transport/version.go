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
	SDKVersionMin    = SDKVersion{4, 0, 0}
	SDKVersionLatest = SDKVersion{4, 20, 0}
	SDKVersionMax    = SDKVersion{254, 1023, 254}
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
			if a.Patch == b.Patch {
				return SDKVersionEqual
			}

			if a.Patch > b.Patch {
				return SDKVersionNewer
			}

			if a.Patch < b.Patch {
				return SDKVersionOlder
			}
		}
	}
	return SDKVersionOlder
}

func (v SDKVersion) String() string {
	if v.Major == 255 {
		return "internal"
	} else {
		return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	}
}

func (a SDKVersion) AtLeast(b SDKVersion) bool {
	return a.Compare(b) != SDKVersionOlder
}
