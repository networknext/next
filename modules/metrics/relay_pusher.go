package metrics

import "context"

type RelayPusherStatus struct {
	// Service Information
	ServiceName string `json:"service_name"`
	GitHash     string `json:"git_hash"`
	Started     string `json:"started"`
	Uptime      string `json:"uptime"`

	// Service Metrics
	Goroutines      int     `json:"goroutines"`
	MemoryAllocated float64 `json:"mb_allocated"`

	// Success Metrics
	MaxmindSuccessfulHTTPCallsISP       int `json:"maxmind_successful_http_calls_isp"`
	MaxmindSuccessfulHTTPCallsCity      int `json:"maxmind_successful_http_calls_city"`
	MaxmindSuccessfulISPSCP             int `json:"maxmind_successful_isp_scp"`
	MaxmindSuccessfulCitySCP            int `json:"maxmind_successful_city_scp"`
	MaxmindSuccessfulISPStorageUploads  int `json:"maxmind_successful_isp_storage_uploads"`
	MaxmindSuccessfulCityStorageUploads int `json:"maxmind_successful_isp_city_uploads"`

	// Error Metrics
	MaxmindHTTPFailureISP           int `json:"maxmind_http_failure_isp"`
	MaxmindHTTPFailureCity          int `json:"maxmind_http_failure_city"`
	MaxmindGZIPReadFailure          int `json:"maxmind_gzip_read_failure"`
	MaxmindTempFileWriteFailure     int `json:"maxmind_temp_file_write_failure"`
	MaxmindSCPWriteFailure          int `json:"maxmind_scp_write_failure"`
	MaxmindStorageUploadFailureISP  int `json:"maxmind_storage_upload_failure_isp"`
	MaxmindStorageUploadFailureCity int `json:"maxmind_storage_upload_failure_city"`
	DatabaseSCPWriteFailure         int `json:"database_scp_write_failure"`

	// Durations
	DBBinaryTotalUpdateDurationMs float64 `json:"db_binary_total_update_duration_ms"`
	MaxmindDBCityUpdateDurationMs float64 `json:"maxmind_db_city_update_duration_ms"`
	MaxmindDBISPUpdateDurationMs  float64 `json:"maxmind_db_isp_update_duration_ms"`
}

// RelayPusherServiceMetrics defines a set of metrics for the beacon insertion service.
type RelayPusherServiceMetrics struct {
	ServiceMetrics     *ServiceMetrics
	RelayPusherMetrics *RelayPusherMetrics
}

// EmptyRelayPusherServiceMetrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptyRelayPusherServiceMetrics RelayPusherServiceMetrics = RelayPusherServiceMetrics{
	ServiceMetrics:     &EmptyServiceMetrics,
	RelayPusherMetrics: &EmptyRelayPusherMetrics,
}

// RelayPusherMetrics defines a set of metrics for monitoring the beacon insertion service.
type RelayPusherMetrics struct {
	MaxmindSuccessfulHTTPCallsISP       Counter
	MaxmindSuccessfulHTTPCallsCity      Counter
	MaxmindSuccessfulISPSCP             Counter
	MaxmindSuccessfulCitySCP            Counter
	MaxmindSuccessfulISPStorageUploads  Counter
	MaxmindSuccessfulCityStorageUploads Counter
	DBBinaryTotalUpdateDuration         Gauge
	MaxmindDBCityUpdateDuration         Gauge
	MaxmindDBISPUpdateDuration          Gauge
	ErrorMetrics                        RelayPusherErrorMetrics
}

// EmptyRelayPusherMetrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptyRelayPusherMetrics RelayPusherMetrics = RelayPusherMetrics{
	MaxmindSuccessfulHTTPCallsISP:       &EmptyCounter{},
	MaxmindSuccessfulHTTPCallsCity:      &EmptyCounter{},
	MaxmindSuccessfulISPSCP:             &EmptyCounter{},
	MaxmindSuccessfulCitySCP:            &EmptyCounter{},
	MaxmindSuccessfulISPStorageUploads:  &EmptyCounter{},
	MaxmindSuccessfulCityStorageUploads: &EmptyCounter{},
	DBBinaryTotalUpdateDuration:         &EmptyGauge{},
	MaxmindDBCityUpdateDuration:         &EmptyGauge{},
	MaxmindDBISPUpdateDuration:          &EmptyGauge{},
	ErrorMetrics:                        EmptyRelayPusherErrorMetrics,
}

// RelayPusherErrorMetrics defines a set of metrics for recording errors for the beacon insertion service.
type RelayPusherErrorMetrics struct {
	MaxmindHTTPFailureISP           Counter
	MaxmindHTTPFailureCity          Counter
	MaxmindGZIPReadFailure          Counter
	MaxmindTempFileWriteFailure     Counter
	MaxmindSCPWriteFailure          Counter
	MaxmindStorageUploadFailureISP  Counter
	MaxmindStorageUploadFailureCity Counter
	DatabaseSCPWriteFailure         Counter
}

// EmptyRelayPusherErrorMetrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptyRelayPusherErrorMetrics RelayPusherErrorMetrics = RelayPusherErrorMetrics{
	MaxmindHTTPFailureISP:           &EmptyCounter{},
	MaxmindHTTPFailureCity:          &EmptyCounter{},
	MaxmindGZIPReadFailure:          &EmptyCounter{},
	MaxmindTempFileWriteFailure:     &EmptyCounter{},
	MaxmindSCPWriteFailure:          &EmptyCounter{},
	MaxmindStorageUploadFailureISP:  &EmptyCounter{},
	MaxmindStorageUploadFailureCity: &EmptyCounter{},
	DatabaseSCPWriteFailure:         &EmptyCounter{},
}

// NewRelayPusherServiceMetrics creates the metrics that the beacon insertion service will use.
func NewRelayPusherServiceMetrics(ctx context.Context, metricsHandler Handler) (*RelayPusherServiceMetrics, error) {
	RelayPusherServiceMetrics := &RelayPusherServiceMetrics{}
	var err error

	RelayPusherServiceMetrics.ServiceMetrics, err = NewServiceMetrics(ctx, metricsHandler, "relay_pusher")
	if err != nil {
		return nil, err
	}

	RelayPusherServiceMetrics.RelayPusherMetrics = &RelayPusherMetrics{}

	RelayPusherServiceMetrics.RelayPusherMetrics.MaxmindSuccessfulHTTPCallsCity, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Successful Maxmind City HTTP GETs",
		ServiceName: "relay_pusher",
		ID:          "successful_maxmind_http_get_city.count",
		Unit:        "calls",
		Description: "The total number of successful Maxmind HTTP calls for IP to city in number of calls.",
	})
	if err != nil {
		return nil, err
	}

	RelayPusherServiceMetrics.RelayPusherMetrics.MaxmindSuccessfulHTTPCallsISP, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Successful Maxmind ISP HTTP GETs",
		ServiceName: "relay_pusher",
		ID:          "successful_maxmind_http_get_isp.count",
		Unit:        "calls",
		Description: "The total number of successful Maxmind HTTP calls for IP to ISP in number of calls.",
	})
	if err != nil {
		return nil, err
	}

	RelayPusherServiceMetrics.RelayPusherMetrics.MaxmindSuccessfulISPSCP, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Successful Maxmind ISP SCP",
		ServiceName: "relay_pusher",
		ID:          "successful_maxmind_isp_scp.count",
		Unit:        "calls",
		Description: "The total number of successful SCP calls to VMs for the maxmind ISP file.",
	})
	if err != nil {
		return nil, err
	}

	RelayPusherServiceMetrics.RelayPusherMetrics.MaxmindSuccessfulCitySCP, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Successful Maxmind City SCP",
		ServiceName: "relay_pusher",
		ID:          "successful_maxmind_city_scp.count",
		Unit:        "calls",
		Description: "The total number of successful SCP calls to VMs for the maxmind City file.",
	})
	if err != nil {
		return nil, err
	}

	RelayPusherServiceMetrics.RelayPusherMetrics.MaxmindSuccessfulCityStorageUploads, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Successful Maxmind City Storage Uploads",
		ServiceName: "relay_pusher",
		ID:          "successful_maxmind_city_storage_uploads.count",
		Unit:        "uploads",
		Description: "The total number of successful Maxmind City uploads to cloud storage.",
	})
	if err != nil {
		return nil, err
	}

	RelayPusherServiceMetrics.RelayPusherMetrics.MaxmindSuccessfulISPStorageUploads, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Successful Maxmind ISP Storage Uploads",
		ServiceName: "relay_pusher",
		ID:          "successful_maxmind_isp_storage_uploads.count",
		Unit:        "uploads",
		Description: "The total number of successful Maxmind ISP uploads to cloud storage.",
	})
	if err != nil {
		return nil, err
	}

	RelayPusherServiceMetrics.RelayPusherMetrics.MaxmindDBISPUpdateDuration, err = metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Maxmind DB ISP Update Duration",
		ServiceName: "relay_pusher",
		ID:          "maxmind_db_isp_update.duration",
		Unit:        "ms",
		Description: "The amount of time it takes to update ISP files on all server backends in ms.",
	})
	if err != nil {
		return nil, err
	}

	RelayPusherServiceMetrics.RelayPusherMetrics.MaxmindDBCityUpdateDuration, err = metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Maxmind DB City Update Duration",
		ServiceName: "relay_pusher",
		ID:          "maxmind_db_city_update.duration",
		Unit:        "ms",
		Description: "The amount of time it takes to update city files on all server backends in ms.",
	})
	if err != nil {
		return nil, err
	}

	RelayPusherServiceMetrics.RelayPusherMetrics.DBBinaryTotalUpdateDuration, err = metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "DB Update Duration",
		ServiceName: "relay_pusher",
		ID:          "db_update.duration",
		Unit:        "ms",
		Description: "The total amount of time it takes to update database binary file on all relay backends in ms.",
	})
	if err != nil {
		return nil, err
	}

	RelayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics = RelayPusherErrorMetrics{}

	RelayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.MaxmindTempFileWriteFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Maxmind Temp File Write Failures",
		ServiceName: "relay_pusher",
		ID:          "temp_file_write_failure.count",
		Unit:        "failures",
		Description: "The total number of Maxmind temporary file write failures.",
	})
	if err != nil {
		return nil, err
	}

	RelayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.MaxmindGZIPReadFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Maxmind GZIP Read Failures",
		ServiceName: "relay_pusher",
		ID:          "gzip_read_failure.count",
		Unit:        "failures",
		Description: "The total number of GZIP read failures.",
	})
	if err != nil {
		return nil, err
	}

	RelayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.MaxmindHTTPFailureCity, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Maxmind HTTP City Call Failures",
		ServiceName: "relay_pusher",
		ID:          "maxmind_http_city_failure.count",
		Unit:        "failures",
		Description: "The total number of Maxmind http city call failures.",
	})
	if err != nil {
		return nil, err
	}

	RelayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.MaxmindHTTPFailureISP, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Maxmind HTTP ISP Call Failures",
		ServiceName: "relay_pusher",
		ID:          "maxmind_http_isp_failure.count",
		Unit:        "failures",
		Description: "The total number of Maxmind http ISP call failures.",
	})
	if err != nil {
		return nil, err
	}

	RelayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.MaxmindSCPWriteFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Maxmind SCP Call Failures",
		ServiceName: "relay_pusher",
		ID:          "maxmind_scp_call_failure.count",
		Unit:        "failures",
		Description: "The total number of Maxmind SCP file copy failures.",
	})
	if err != nil {
		return nil, err
	}

	RelayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.MaxmindStorageUploadFailureISP, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Maxmind ISP Storage Upload Failures",
		ServiceName: "relay_pusher",
		ID:          "maxmind_upload_isp_failure.count",
		Unit:        "failures",
		Description: "The total number of Maxmind ISP storage upload failures.",
	})
	if err != nil {
		return nil, err
	}

	RelayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.MaxmindStorageUploadFailureCity, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Maxmind City Storage Upload Failures",
		ServiceName: "relay_pusher",
		ID:          "maxmind_upload_city_failure.count",
		Unit:        "failures",
		Description: "The total number of Maxmind City storage upload failures.",
	})
	if err != nil {
		return nil, err
	}

	RelayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.DatabaseSCPWriteFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Database SCP Call Failures",
		ServiceName: "relay_pusher",
		ID:          "database_scp_call_failure.count",
		Unit:        "failures",
		Description: "The total number of database SCP file copy failures.",
	})
	if err != nil {
		return nil, err
	}

	return RelayPusherServiceMetrics, nil
}
