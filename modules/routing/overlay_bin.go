package routing

import (
	"bytes"
	"encoding/gob"
	"io/ioutil"
	"os"
)

const (
	MaxOverlayBinWrapperSize = 100000000
)

type OverlayBinWrapper struct {
	CreationTime string
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

func (wrapper OverlayBinWrapper) WriteOverlayBinFile(outputPath string) error {
	var buffer bytes.Buffer

	err := gob.NewEncoder(&buffer).Encode(wrapper)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(outputPath, buffer.Bytes(), 0644); err != nil {
		return err
	}

	return nil
}

// This function is essentially the same as DecodeOverlayWrapper in modules/backend/helpers.go
func (wrapper *OverlayBinWrapper) ReadOverlayBinFile(overlayFilePath string) error {
	overlayFile, err := os.Open(overlayFilePath)
	if err != nil {
		return err
	}
	defer overlayFile.Close()

	return gob.NewDecoder(overlayFile).Decode(wrapper)
}
