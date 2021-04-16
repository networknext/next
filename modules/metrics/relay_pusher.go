package metrics

import "context"

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
	SuccessfulMaxmindUpdates       Counter
	MaxmindSuccessfulHTTPCallsISP  Counter
	MaxmindSuccessfulHTTPCallsCity Counter
	DBBinaryTotalUpdateDuration    Gauge
	MaxmindDBTotalUpdateDuration   Gauge
	MaxmindDBCityUpdateDuration    Gauge
	MaxmindDBISPUpdateDuration     Gauge
	ErrorMetrics                   RelayPusherErrorMetrics
}

// EmptyRelayPusherMetrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptyRelayPusherMetrics RelayPusherMetrics = RelayPusherMetrics{
	SuccessfulMaxmindUpdates:       &EmptyCounter{},
	MaxmindSuccessfulHTTPCallsISP:  &EmptyCounter{},
	MaxmindSuccessfulHTTPCallsCity: &EmptyCounter{},
	DBBinaryTotalUpdateDuration:    &EmptyGauge{},
	MaxmindDBTotalUpdateDuration:   &EmptyGauge{},
	MaxmindDBCityUpdateDuration:    &EmptyGauge{},
	MaxmindDBISPUpdateDuration:     &EmptyGauge{},
	ErrorMetrics:                   EmptyRelayPusherErrorMetrics,
}

// RelayPusherErrorMetrics defines a set of metrics for recording errors for the beacon insertion service.
type RelayPusherErrorMetrics struct {
	MaxmindHTTPFailureISP             Counter
	MaxmindHTTPFailureCity            Counter
	MaxmindGZIPReadFailure            Counter
	MaxmindTempFileWriteFailure       Counter
	MaxmindSCPWriteFailure            Counter
	DatabaseSCPWriteFailure           Counter
	ServerBackendInstanceCountFailure Counter
}

// EmptyRelayPusherErrorMetrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptyRelayPusherErrorMetrics RelayPusherErrorMetrics = RelayPusherErrorMetrics{
	MaxmindHTTPFailureISP:             &EmptyCounter{},
	MaxmindHTTPFailureCity:            &EmptyCounter{},
	MaxmindGZIPReadFailure:            &EmptyCounter{},
	MaxmindTempFileWriteFailure:       &EmptyCounter{},
	MaxmindSCPWriteFailure:            &EmptyCounter{},
	DatabaseSCPWriteFailure:           &EmptyCounter{},
	ServerBackendInstanceCountFailure: &EmptyCounter{},
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

	RelayPusherServiceMetrics.RelayPusherMetrics.SuccessfulMaxmindUpdates, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Successful IP2Location Updates",
		ServiceName: "relay_pusher",
		ID:          "successful_ip_2_location_updates.count",
		Unit:        "updates",
		Description: "The total number of successful IP2Location updates in number of updates.",
	})
	if err != nil {
		return nil, err
	}

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

	RelayPusherServiceMetrics.RelayPusherMetrics.MaxmindDBTotalUpdateDuration, err = metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Maxmind DB Update Duration",
		ServiceName: "relay_pusher",
		ID:          "maxmind_db_update.duration",
		Unit:        "ms",
		Description: "The total amount of time it takes to update both maxmind files on all server backends in ms.",
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

	RelayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.ServerBackendInstanceCountFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Server Backend Instance Count Failures",
		ServiceName: "relay_pusher",
		ID:          "server_backend_instance_count_failure.count",
		Unit:        "failures",
		Description: "The total number of gcloud instance count call failures.",
	})
	if err != nil {
		return nil, err
	}

	return RelayPusherServiceMetrics, nil
}
