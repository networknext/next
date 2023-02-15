package messages

import (
	"cloud.google.com/go/bigquery"
)

type Message interface {
	GetMaxSize() int
	Write(buffer []byte) []byte
	Read(buffer []byte) error
}

type BigQueryMessage interface {
	GetMaxSize() int
	Write(buffer []byte) []byte
	Read(buffer []byte) error
	Save() (map[string]bigquery.Value, string, error)
}
