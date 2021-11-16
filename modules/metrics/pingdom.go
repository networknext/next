package metrics

import "context"

type PingdomStatus struct {
	// Service Information
	ServiceName string `json:"service_name"`
	GitHash     string `json:"git_hash"`
	Started     string `json:"started"`
	Uptime      string `json:"uptime"`

	// Service Metrics
	Goroutines      int     `json:"goroutines"`
	MemoryAllocated float64 `json:"mb_allocated"`

	// Sucess Metrics
	CreatePingdomUptime  int `json:"create_pingdom_uptime"`
	BigQueryWriteSuccess int `json:"bigquery_write_success"`

	// Error Metrics
	PingdomAPICallFailure        int `json:"pingdom_api_call_failure"`
	BigQueryReadFailure          int `json:"bigquery_read_failure"`
	BigQueryWriteFailure         int `json:"bigquery_write_failure"`
	BadSummaryPerformanceRequest int `json:"bad_summary_performance_request"`
}

type PingdomMetrics struct {
	PingdomServiceMetrics *ServiceMetrics
	CreatePingdomUptime   Counter
	BigQueryWriteSuccess  Counter
	ErrorMetrics          PingdomErrorMetrics
}

var EmptyPingdomMetrics = &PingdomMetrics{
	PingdomServiceMetrics: &EmptyServiceMetrics,
	CreatePingdomUptime:   &EmptyCounter{},
	BigQueryWriteSuccess:  &EmptyCounter{},
	ErrorMetrics:          EmptyPingdomErrorMetrics,
}

type PingdomErrorMetrics struct {
	PingdomAPICallFailure        Counter
	BigQueryReadFailure          Counter
	BigQueryWriteFailure         Counter
	BadSummaryPerformanceRequest Counter
}

var EmptyPingdomErrorMetrics = PingdomErrorMetrics{
	PingdomAPICallFailure:        &EmptyCounter{},
	BigQueryReadFailure:          &EmptyCounter{},
	BigQueryWriteFailure:         &EmptyCounter{},
	BadSummaryPerformanceRequest: &EmptyCounter{},
}

func NewPingdomMetrics(ctx context.Context, metricsHandler Handler, serviceName string, handlerID string, handlerName string) (*PingdomMetrics, error) {
	m := new(PingdomMetrics)

	var err error

	m.PingdomServiceMetrics, err = NewServiceMetrics(ctx, metricsHandler, serviceName)
	if err != nil {
		return nil, err
	}

	m.CreatePingdomUptime, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Create Pingdom Uptime",
		ServiceName: serviceName,
		ID:          handlerID + ".create_pingdom_uptime",
		Unit:        "count",
		Description: "The number of successful pingdom uptime structs were created",
	})
	if err != nil {
		return nil, err
	}

	m.BigQueryWriteSuccess, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " BigQuery Write Success",
		ServiceName: serviceName,
		ID:          handlerID + ".bigquery_write_success",
		Unit:        "count",
		Description: "The number of pingdom uptime structs to that were successfully written to the pingdom BigQuery table",
	})
	if err != nil {
		return nil, err
	}

	m.ErrorMetrics.PingdomAPICallFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Pingdom API Call Failure",
		ServiceName: serviceName,
		ID:          handlerID + ".pingdom_api_call_failure",
		Unit:        "errors",
		Description: "The number of failed calls to the Pingdom API",
	})
	if err != nil {
		return nil, err
	}

	m.ErrorMetrics.BigQueryReadFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " BigQuery Read Failure",
		ServiceName: serviceName,
		ID:          handlerID + ".bigquery_read_failure",
		Unit:        "errors",
		Description: "The number of failed reads from the pingdom BigQuery table",
	})
	if err != nil {
		return nil, err
	}

	m.ErrorMetrics.BigQueryWriteFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " BigQuery Write Failure",
		ServiceName: serviceName,
		ID:          handlerID + ".bigquery_write_failure",
		Unit:        "errors",
		Description: "The number of failed writes to the pingdom BigQuery table",
	})
	if err != nil {
		return nil, err
	}

	m.ErrorMetrics.BadSummaryPerformanceRequest, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Bad Summary Performance Request",
		ServiceName: serviceName,
		ID:          handlerID + ".bad_summary_performance_request",
		Unit:        "errors",
		Description: "The number of malformed summary performance requests",
	})
	if err != nil {
		return nil, err
	}

	return m, nil
}
