package storage

import "fmt"

type UnmarshalError struct {
	err error
}

func (e *UnmarshalError) Error() string {
	return fmt.Sprintf("unmarshal error: %v", e.err)
}

type DoesNotExistError struct {
	resourceType string
	resourceRef  interface{}
}

func (e *DoesNotExistError) Error() string {
	return fmt.Sprintf("%s with reference %v not found", e.resourceType, e.resourceRef)
}

type AlreadyExistsError struct {
	resourceType string
	resourceRef  interface{}
}

func (e *AlreadyExistsError) Error() string {
	return fmt.Sprintf("%s with reference %v already exists", e.resourceType, e.resourceRef)
}

type HexStringConversionError struct {
	hexString string
}

func (e *HexStringConversionError) Error() string {
	return fmt.Sprintf("error converting hex string %s to uint64", e.hexString)
}

type SequenceNumbersOutOfSync struct {
	localSequenceNumber  int64
	remoteSequenceNumber int64
}

func (e *SequenceNumbersOutOfSync) Error() string {
	return fmt.Sprintf("sequence number out of sync: remote %d != local %d", e.remoteSequenceNumber, e.localSequenceNumber)
}
