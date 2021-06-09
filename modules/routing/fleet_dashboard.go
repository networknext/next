package routing

import "time"

// FleetDashboardData will grow with time to include more fields. This
// data is stored in the fleet_data table and makes up part of the
// data sent to the Admin UI dashboard view.
type FleetDashboardData struct {
	DatabaseBinFileAuthor       string
	DatabaseBinFileCreationTime time.Time
}
