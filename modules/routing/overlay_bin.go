package routing

import "time"

const (
	MaxOverlayBinWrapperSize = 100000000
)

type OverlayBinWrapper struct {
	CreationTime time.Time
	BuyerMap     map[uint64]Buyer
}

func CreateEmptyOverlayBinWrapper() *OverlayBinWrapper {
	wrapper := &OverlayBinWrapper{
		BuyerMap: make(map[uint64]Buyer),
	}

	return wrapper
}

func (wrapper OverlayBinWrapper) IsEmpty() bool {
	return len(wrapper.BuyerMap) == 0
}
