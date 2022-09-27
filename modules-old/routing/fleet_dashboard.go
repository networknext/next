package routing

import "time"

// Fleet dashboard data will grow with time to include more types
// and tables.

// DatabaseBinFileMetaData is stored in the database_bin_meta table and
// makes up part of the data sent to the Admin UI dashboard view.
type DatabaseBinFileMetaData struct {
	DatabaseBinFileAuthor       string
	DatabaseBinFileCreationTime time.Time
}

func (dmfmd *DatabaseBinFileMetaData) String() string {
	data := "\nrouting.DatabaseBinFileMetaData:\n"
	data += "  DatabaseBinFileAuthor      : " + dmfmd.DatabaseBinFileAuthor + "\n"
	data += "  DatabaseBinFileCreationTime: " + dmfmd.DatabaseBinFileCreationTime.String() + "\n"

	return data
}
