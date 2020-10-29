package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/networknext/backend/modules/envvar"

	"cloud.google.com/go/bigtable"
	"google.golang.org/api/option"
)

type BigTable struct {
	Client       *bigtable.Client
	Logger       log.Logger
	SessionTable *bigtable.Table
}

type BigTableAdmin struct {
	Client *bigtable.AdminClient
	Logger log.Logger
}

type BigTableError struct {
	err error
}

func (e *BigTableError) Error() string {
	return fmt.Sprintf("unknown BigTable error: %v", e.err)
}

// Creates a new Bigtable object
// Mainly used for opening tables in the instance
func NewBigTable(ctx context.Context, gcpProjectID string, instanceID string, logger log.Logger, opts ...option.ClientOption) (*BigTable, error) {
	client, err := bigtable.NewClient(ctx, gcpProjectID, instanceID, opts...)
	if err != nil {
		return nil, err
	}
	btTableName := envvar.Get("GOOGLE_BIGTABLE_TABLE_NAME", "")

	if btTableName == "" {
		err := fmt.Errorf("NewBigTable() GOOGLE_BIGTABLE_TABLE_NAME is not defined")
		level.Error(logger).Log("err", err)
		return nil, err
	}
	table := client.Open(btTableName)

	return &BigTable{
		Client:       client,
		Logger:       logger,
		SessionTable: table,
	}, nil
}

// Creates a new Bigtable Admin
// Admins have special abilities like creating an deleting tables
func NewBigTableAdmin(ctx context.Context, gcpProjectID string, instanceID string, logger log.Logger, opts ...option.ClientOption) (*BigTableAdmin, error) {
	client, err := bigtable.NewAdminClient(ctx, gcpProjectID, instanceID, opts...)
	if err != nil {
		return nil, err
	}

	return &BigTableAdmin{
		Client: client,
		Logger: logger,
	}, nil
}

// Gets a list of tables for the instance
func (bt *BigTableAdmin) GetTableList(ctx context.Context) ([]string, error) {
	return bt.Client.Tables(ctx)
}

// Checks if a table exists in the instance
func (bt *BigTableAdmin) VerifyTableExists(ctx context.Context, tableName string) (bool, error) {
	tableList, err := bt.GetTableList(ctx)
	if err != nil {
		return false, err
	}

	if len(tableList) == 0 {
		return false, nil
	}

	for _, tblName := range tableList {
		if tblName == tableName {
			return true, nil
		}
	}

	return false, nil
}

// Creates a table with the given column families
func (bt *BigTableAdmin) CreateTable(ctx context.Context, btTableName string, btCfNames []string) error {
	// Create a table with the given name
	if err := bt.Client.CreateTable(ctx, btTableName); err != nil {
		return err
	}

	// Create column families for the table
	for _, btCfName := range btCfNames {
		if err := bt.Client.CreateColumnFamily(ctx, btTableName, btCfName); err != nil {
			return err
		}
	}

	return nil
}

// Deletes a table and all its data
func (bt *BigTableAdmin) DeleteTable(ctx context.Context, btTableName string) error {
	return bt.Client.DeleteTable(ctx, btTableName)
}

// Sets a garbage collection policy on column families listed in a table
func (bt *BigTableAdmin) SetMaxAgePolicy(ctx context.Context, btTableName string, btCfNames []string, maxAge time.Duration) error {
	maxAgePolicy := bigtable.MaxAgePolicy(maxAge)

	for _, btCfName := range btCfNames {
		if err := bt.Client.SetGCPolicy(ctx, btTableName, btCfName, maxAgePolicy); err != nil {
			return err
		}
	}

	return nil
}

// Closes the bigtable admin
func (bt *BigTableAdmin) Close() error {
	if err := bt.Client.Close(); err != nil {
		return err
	}

	return nil
}

// Closes the bigtable object
func (bt *BigTable) Close() error {
	if err := bt.Client.Close(); err != nil {
		return err
	}

	return nil
}

// Gets a table in the instance
func (bt *BigTable) GetTable(btTableName string) *bigtable.Table {
	return bt.Client.Open(btTableName)
}

// Inserts a row into a table given a slice of row keys,
// a map of the column name to the data stored in thaat cell,
// and a map of the column name to the column family
func (bt *BigTable) InsertRowInTable(ctx context.Context, rowKeys []string, dataMap map[string][]byte, cfMap map[string]string) error {

	// Get the timestamp for time of insertion
	currentTimestamp := bigtable.Now()

	// Create the mutation for the rows
	mut := bigtable.NewMutation()

	for colName, value := range dataMap {
		if cfName, ok := cfMap[colName]; ok {
			mut.Set(cfName, colName, currentTimestamp, value)
		} else {
			return fmt.Errorf("Column name %v not present in column family map", colName)
		}
	}

	// Insert into table
	for _, rowKey := range rowKeys {
		if err := bt.SessionTable.Apply(ctx, rowKey, mut); err != nil {
			return err
		}
	}

	return nil
}

// Inserts session data into Bigtable
func (bt *BigTable) InsertSessionData(ctx context.Context,
	btCfNames []string,
	metaBinary []byte,
	sliceBinary []byte,
	rowKeys []string) error {

	// Create a map of column name to session data
	sessionDataMap := make(map[string][]byte)
	sessionDataMap["meta"] = metaBinary
	sessionDataMap["slice"] = sliceBinary

	// Create a map of column name to column family
	// Always map meta and slice to the first column family
	if len(btCfNames) == 0 {
		return fmt.Errorf("Column family names slice is empty")
	} 
	cfMap := make(map[string]string)
	cfMap["meta"] = btCfNames[0]
	cfMap["slice"] = btCfNames[0]

	if err := bt.InsertRowInTable(ctx, rowKeys, sessionDataMap, cfMap); err != nil {
		return err
	}

	return nil
}

// Gets all rows starting with a prefix (i.e. session ID)
// Can provide a ReadOption, which can include various filters
// See: https://godoc.org/cloud.google.com/go/bigtable#ReadOption
// Returns a slice of Row structs, which is a map of a column family name as the key
// to a slice of ReadItem structs
// See: https://godoc.org/cloud.google.com/go/bigtable#Row
// type ReadItem struct {
//     Row, Column string
//     Timestamp   Timestamp
//     Value       []byte
//     Labels      []string
// }
func (bt *BigTable) GetRowsWithPrefix(ctx context.Context, prefix string, opts ...bigtable.ReadOption) ([]bigtable.Row, error) {
	// Get a range of all rows starting with a prefix
	rowRange := bigtable.PrefixRange(prefix)

	// Create a slice of all the rows to return
	values := make([]bigtable.Row, 0)

	err := bt.SessionTable.ReadRows(ctx, rowRange, func(r bigtable.Row) bool {
		// Get the data and put it into a slice to return
		values = append(values, r)

		return true
	}, opts...)

	if err != nil {
		return nil, err
	}

	return values, nil
}

// Gets a row given a row key and a slice of column family names
// Can provide a ReadOption, which can include various filters
// See: https://godoc.org/cloud.google.com/go/bigtable#ReadOption
// Returns a Row struct, which is a map of a column family name as the key
// to a slice of ReadItem structs
// NOTE: Missing rows (or rows that do not exist) return a zero-length map
// See: https://godoc.org/cloud.google.com/go/bigtable#Row
// type ReadItem struct {
//     Row, Column string
//     Timestamp   Timestamp
//     Value       []byte
//     Labels      []string
// }
func (bt *BigTable) GetRowWithRowKey(ctx context.Context, rowKey string, opts ...bigtable.ReadOption) (bigtable.Row, error) {

	r, err := bt.SessionTable.ReadRow(ctx, rowKey, opts...)
	if err != nil {
		return nil, err
	}

	return r, nil
}
