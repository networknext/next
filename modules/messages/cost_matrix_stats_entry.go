package messages

import (
	"cloud.google.com/go/bigquery"
	// "github.com/networknext/backend/modules/encoding"
)

type CostMatrixStatsEntry struct {
	Timestamp  uint64
	// todo
}

// todo: write

// todo: read

func (e *CostMatrixStatsEntry) Save() (map[string]bigquery.Value, string, error) {

	bqEntry := make(map[string]bigquery.Value)

	bqEntry["timestamp"] = int(e.Timestamp)

	// todo

	return bqEntry, "", nil
}
