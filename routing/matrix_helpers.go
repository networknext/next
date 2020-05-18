package routing

import (
	"errors"

	"github.com/networknext/backend/encoding"
)

func readIDOld(data []byte, index *int, storage *uint64, errmsg string) error {
	var tmp uint32
	if !encoding.ReadUint32(data, index, &tmp) {
		return errors.New(errmsg + " - ver < 3")
	}
	*storage = uint64(tmp)
	return nil
}

func readIDNew(data []byte, index *int, storage *uint64, errmsg string) error {
	if !encoding.ReadUint64(data, index, storage) {
		return errors.New(errmsg + " - ver >= v3")
	}
	return nil
}

func readBytesOld(data []byte, index *int, storage *[]byte, length uint32, errmsg string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(errmsg + " - ver < 3")
		}
	}()

	var bytesRead int
	*storage, bytesRead = encoding.ReadBytesOld(data[*index:])
	*index += bytesRead
	return err
}

func readBytesNew(data []byte, index *int, storage *[]byte, length uint32, errmsg string) error {
	if !encoding.ReadBytes(data, index, storage, length) {
		return errors.New(errmsg + " - ver >= v3")
	}
	return nil
}
