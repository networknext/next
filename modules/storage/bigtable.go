package storage

import (
	"context"
	"fmt"
	"time"
	"unicode"

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

type BigTableInstanceAdmin struct {
	Client *bigtable.InstanceAdminClient
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
func NewBigTable(ctx context.Context, gcpProjectID string, instanceID string, btTableName string, logger log.Logger, opts ...option.ClientOption) (*BigTable, error) {
	client, err := bigtable.NewClient(ctx, gcpProjectID, instanceID, opts...)
	if err != nil {
		return nil, err
	}

	if btTableName == "" {
		err := fmt.Errorf("NewBigTable() table name is empty or not defined")
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
// Admins have special abilities like creating and deleting tables
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

// Creates a new Bigtable Instance Admin
// Instance Admins have special abilities like creating instances and clusters
func NewBigTableInstanceAdmin(ctx context.Context, gcpProjectID string, logger log.Logger, opts ...option.ClientOption) (*BigTableInstanceAdmin, error) {
	client, err := bigtable.NewInstanceAdminClient(ctx, gcpProjectID, opts...)
	if err != nil {
		return nil, err
	}

	return &BigTableInstanceAdmin{
		Client: client,
		Logger: logger,
	}, nil
}

// Creates an bigtable instance with the same number of nodes per cluster
// We also only use SSD storage type and production instances
// Need to have an equal number of zones for numClusters (ideally want clusters in different zones)
func (bt *BigTableInstanceAdmin) CreateInstance(ctx context.Context, instanceID string, displayName string, zones []string, numClusters int, numNodesPerCluster int) error {
	// Verify display name length
	if len(displayName) < 4 || len(displayName) > 30 {
		return fmt.Errorf("CreateInstance() display name %s must be between 4 and 30 characters", displayName)
	}

	// Verify instance ID
	{
		if len(instanceID) < 6 || len(instanceID) > 33 {
			return fmt.Errorf("CreateInstance() instance ID %s must be between 6 and 33 characters", instanceID)
		}

		for i, r := range instanceID {
			if i == 0 {
				if !unicode.IsLower(r) || !unicode.IsLetter(r) {
					return fmt.Errorf("CreateInstance() instance ID %s must start with a lowercase letter", instanceID)
				}
			} else if !unicode.IsLower(r) && !unicode.IsNumber(r) && instanceID[i:i+1] != "-" {
				return fmt.Errorf("CreateInstance() instance ID %s must only contain hyphens, lowercase letters, and numbers", instanceID)
			}
		}
	}

	// Verify there is at least one cluster
	if numClusters < 1 {
		return fmt.Errorf("CreateInstance() need at least one cluster in the instance")
	}

	// Verify length of zones slice is the same as numClusters
	if len(zones) != numClusters {
		return fmt.Errorf("CreateInstance() need an equal of number of zones as the number of clusters")
	}

	// Verify there is at least 1 node per cluster
	if numNodesPerCluster < 1 {
		return fmt.Errorf("CreateInstance() need at least one node per cluster")
	}

	var clusterConfig []bigtable.ClusterConfig
	for i := 0; i < numClusters; i++ {
		// clusterID must be between 6 and 30 characters
		var clusterID string
		if len(instanceID) > 27 {
			clusterID = fmt.Sprintf("%s-c%d", instanceID[:27], i+1)
		} else {
			clusterID = fmt.Sprintf("%s-c%d", instanceID, i+1)
		}

		conf := bigtable.ClusterConfig{
			InstanceID:  instanceID,
			ClusterID:   clusterID,
			Zone:        zones[i],
			NumNodes:    int32(numNodesPerCluster),
			StorageType: bigtable.StorageType(0),
		}

		clusterConfig = append(clusterConfig, conf)
	}

	instanceConf := &bigtable.InstanceWithClustersConfig{
		InstanceID:   instanceID,
		DisplayName:  displayName,
		Clusters:     clusterConfig,
		InstanceType: bigtable.InstanceType(0),
	}

	return bt.Client.CreateInstanceWithClusters(ctx, instanceConf)
}

// Deletes a bigtable instance
func (bt *BigTableInstanceAdmin) DeleteInstance(ctx context.Context, instanceID string) error {
	return bt.Client.DeleteInstance(ctx, instanceID)
}

// Gets all instances for this project
func (bt *BigTableInstanceAdmin) GetInstances(ctx context.Context) ([]*bigtable.InstanceInfo, error) {
	return bt.Client.Instances(ctx)
}

// Verifies if an instance exists
func (bt *BigTableInstanceAdmin) VerifyInstanceExists(ctx context.Context, instanceID string) (bool, error) {
	// Get the instances
	instances, err := bt.GetInstances(ctx)
	if err != nil {
		return false, err
	}

	// Iterate through list of instances and identify if the instance exists
	for _, instance := range instances {
		if instanceID == instance.Name {
			return true, nil
		}
	}

	return false, nil
}

// Closes the bigtable instance admin
func (bt *BigTableInstanceAdmin) Close() error {
	if err := bt.Client.Close(); err != nil {
		return err
	}

	return nil
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

// Drops all rows from a table that start with the given prefix
func (bt *BigTableAdmin) DropRowsByPrefix(ctx context.Context, btTableName string, prefix string) error {
	return bt.Client.DropRowRange(ctx, btTableName, prefix)
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
// a map of the column name to the data stored in that cell,
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
			return fmt.Errorf("InsertRowInTable() Column name %v not present in column family map", colName)
		}
	}

	// Insert into table
	for _, rowKey := range rowKeys {
		if err := bt.SessionTable.Apply(ctx, rowKey, mut); err != nil {
			return fmt.Errorf("InsertRowInTable() Could not insert row in table %v", err)
		}
	}

	return nil
}

// Write a unique row with one cell per row key
// Works by overwriting a previous row by using a timestamp with Unix time of 0
// NOTE: MaxAgePolicy will delete any rows before the last X days, so use today's date to prevent this
// from happening until it's appropriate
// Reference: https://cloud.google.com/bigtable/docs/gc-latest-value
func (bt *BigTable) WriteRowInTable(ctx context.Context, rowKeys []string, dataMap map[string][]byte, cfMap map[string]string) error {
	// Create timestamp with only today's date (to avoid MaxAgePolicy)
	timeNow := time.Now()
	year, month, date := timeNow.Date()
	location := timeNow.Location()
	currentDateTime := bigtable.Time(time.Date(year, month, date, 0, 0, 0, 0, location)).TruncateToMilliseconds()

	// Create the mutation for the rows
	mut := bigtable.NewMutation()

	for colName, value := range dataMap {
		if cfName, ok := cfMap[colName]; ok {
			mut.Set(cfName, colName, currentDateTime, value)
		} else {
			return fmt.Errorf("WriteRowInTable() Column name %v not present in column family map", colName)
		}
	}

	// Insert into table
	for _, rowKey := range rowKeys {
		if err := bt.SessionTable.Apply(ctx, rowKey, mut); err != nil {
			return fmt.Errorf("WriteRowInTable() Could not insert row in table %v", err)
		}
	}

	return nil
}

// Write and delete a row in a table
func (bt *BigTable) WriteAndDeleteRowInTable(ctx context.Context, rowKeys []string, dataMap map[string][]byte, cfMap map[string]string) error {
	timeNow := time.Now()
	// Get the timestamp for time of insertion
	currentTimestamp := bigtable.Time(timeNow)

	// Get the timestamp from 1 millisecond ago
	deltaTimestamp := bigtable.Time(timeNow.Add(-1 * time.Millisecond)).TruncateToMilliseconds()
	// Create timestamp with unix time of 0 (1 January 1970)
	zeroTime := bigtable.Time(time.Unix(0, 0)).TruncateToMilliseconds()

	// Create slice of mutations for deleting any rows not within the past 1 millisecond
	deleteMut := bigtable.NewMutation()
	// Create slice of mutations for the new rows
	rowMut := bigtable.NewMutation()

	for colName, value := range dataMap {
		if cfName, ok := cfMap[colName]; ok {
			// Create a mutation for deleting any rows not within the past 1 millisecond
			deleteMut.DeleteTimestampRange(cfName, colName, zeroTime, deltaTimestamp)

			// Create the mutation for the replacement row
			rowMut.Set(cfName, colName, currentTimestamp, value)
		} else {
			return fmt.Errorf("WriteAndDeleteRowInTable() Column name %v not present in column family map", colName)
		}
	}

	// Add the latest rows first then delete older rows
	for _, rowKey := range rowKeys {
		if err := bt.SessionTable.Apply(ctx, rowKey, rowMut); err != nil {
			return fmt.Errorf("WriteAndDeleteRowInTable() Could not insert row in table %v", err)
		}

		if err := bt.SessionTable.Apply(ctx, rowKey, deleteMut); err != nil {
			return fmt.Errorf("WriteAndDeleteRowInTable() Could not delete row in table %v", err)
		}
	}

	return nil
}

// Inserts session data into Bigtable
func (bt *BigTable) InsertSessionMetaData(ctx context.Context,
	btCfNames []string,
	metaBinary []byte,
	rowKeys []string) error {

	// Create a map of column name to session data
	sessionDataMap := make(map[string][]byte)
	sessionDataMap["meta"] = metaBinary

	// Create a map of column name to column family
	// Always map meta to the first column family
	if len(btCfNames) == 0 {
		return fmt.Errorf("InsertSessionMetaData() Column family names slice is empty")
	}
	cfMap := make(map[string]string)
	cfMap["meta"] = btCfNames[0]

	// Decide if should write and delete row
	// or if should just write row and let compaction take care of deleting the row later
	deleteWrite, err := envvar.GetBool("BIGTABLE_WRITE_DELETE_ROW", false)
	if err != nil {
		return err
	}

	// A/B testing for above
	if deleteWrite {
		if err := bt.WriteAndDeleteRowInTable(ctx, rowKeys, sessionDataMap, cfMap); err != nil {
			return err
		}
	} else {
		if err := bt.WriteRowInTable(ctx, rowKeys, sessionDataMap, cfMap); err != nil {
			return err
		}
	}

	return nil
}

// Inserts session data into Bigtable
func (bt *BigTable) InsertSessionSliceData(ctx context.Context,
	btCfNames []string,
	sliceBinary []byte,
	rowKeys []string) error {

	// Create a map of column name to session data
	sessionDataMap := make(map[string][]byte)
	sessionDataMap["slices"] = sliceBinary

	// Create a map of column name to column family
	// Always map slice to the first column family
	if len(btCfNames) == 0 {
		return fmt.Errorf("InsertSessionSliceData() Column family names slice is empty")
	}
	cfMap := make(map[string]string)
	cfMap["slices"] = btCfNames[0]

	if err := bt.InsertRowInTable(ctx, rowKeys, sessionDataMap, cfMap); err != nil {
		return fmt.Errorf("InsertSessionSliceData() Could not insert session data into table %v", err)
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
	prefixRange := bigtable.PrefixRange(prefix)
	
	// Create a slice of all the rows to return
	values := make([]bigtable.Row, 0)

	err := bt.SessionTable.ReadRows(ctx, prefixRange, func(r bigtable.Row) bool {
		// Get the data and put it into a slice to return
		values = append(values, r)

		return true
	}, opts...)

	if err != nil {
		return nil, fmt.Errorf("GetRowsWithPrefix() Could not get rows with prefix %s: %v", prefix, err)
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
