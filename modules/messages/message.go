package messages

import (
    "cloud.google.com/go/bigquery"
)

type Message interface {
    Write(buffer []byte) []byte

    Read(buffer []byte) error

    Save() (map[string]bigquery.Value, string, error)
}
