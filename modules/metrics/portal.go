package metrics

// PortalStatus defines the metrics reported by the service's status endpoint
type PortalStatus struct {
	// Service Information
	ServiceName string `json:"service_name"`
	GitHash     string `json:"git_hash"`
	Started     string `json:"started"`
	Uptime      string `json:"uptime"`

	// Metrics
	Goroutines      int     `json:"goroutines"`
	MemoryAllocated float64 `json:"mb_allocated"`

	// Bigtable Counts
	ReadMetaSuccessCount  int `json:"big_table_read_meta_success_count"`
	ReadSliceSuccessCount int `json:"big_table_read_slice_success_count"`

	// Bigtable Errors
	ReadMetaFailureCount  int `json:"big_table_read_meta_failure_count"`
	ReadSliceFailureCount int `json:"big_table_read_slice_failure_count"`

	// BuyerEndpoint Errors
	NoSlicesFailure int `json:"buyer_endpoint_no_slices_failure"`
}
