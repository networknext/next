package messages

import (
	"cloud.google.com/go/bigquery"
	// "github.com/networknext/backend/modules/encoding"
)

type RouteMatrixStatsEntry struct {
	Timestamp  uint64
	// todo
}

// todo: write

// todo: read

func (e *RouteMatrixStatsEntry) Save() (map[string]bigquery.Value, string, error) {

	bqEntry := make(map[string]bigquery.Value)

	bqEntry["timestamp"] = int(e.Timestamp)

	// todo

	return bqEntry, "", nil
}
