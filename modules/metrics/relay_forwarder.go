package metrics

import "context"

// RelayForwarderStatus defines the metrics reported by the service's status endpoint
type RelayForwarderStatus struct {
	// Service Information
	ServiceName string `json:"service_name"`
	GitHash     string `json:"git_hash"`
	Started     string `json:"started"`
	Uptime      string `json:"uptime"`

	// Service Metrics
	Goroutines      int     `json:"goroutines"`
	MemoryAllocated float64 `json:"mb_allocated"`

	// Handler Metrics
	Invocations int     `json:"invocations"`
	DurationMs  float64 `json:"duration_ms"`

	// Error Metrics
	ParseURLError    int `json:"parse_url_error"`
	ForwardPostError int `json:"forward_port_error"`
}

type RelayForwarderMetrics struct {
	HandlerMetrics          *PacketHandlerMetrics
	ForwarderServiceMetrics *ServiceMetrics
	ErrorMetrics            RelayForwarderErrorMetrics
}

var EmptyRelayForwarderMetrics = &RelayForwarderMetrics{
	HandlerMetrics:          &EmptyPacketHandlerMetrics,
	ForwarderServiceMetrics: &EmptyServiceMetrics,
	ErrorMetrics:            *EmptyRelayForwarderErrorMetrics,
}

type RelayForwarderErrorMetrics struct {
	ParseURLError    Counter
	ForwardPostError Counter
}

var EmptyRelayForwarderErrorMetrics = &RelayForwarderErrorMetrics{
	ParseURLError:    &EmptyCounter{},
	ForwardPostError: &EmptyCounter{},
}

func NewRelayForwarderMetrics(ctx context.Context, metricsHandler Handler, serviceName string, handlerID string, handlerName string, packetDescription string) (*RelayForwarderMetrics, error) {
	m := new(RelayForwarderMetrics)

	var err error

	m.HandlerMetrics, err = NewPacketHandlerMetrics(ctx, metricsHandler, serviceName, handlerID, handlerName, packetDescription)
	if err != nil {
		return nil, err
	}

	m.ForwarderServiceMetrics, err = NewServiceMetrics(ctx, metricsHandler, serviceName)
	if err != nil {
		return nil, err
	}

	m.ErrorMetrics.ParseURLError, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Parse URL Error",
		ServiceName: serviceName,
		ID:          handlerID + ".parse_url_error",
		Unit:        "errors",
		Description: "The Number of times the " + serviceName + " failed to parse the request's remote address as a URL",
	})
	if err != nil {
		return nil, err
	}

	m.ErrorMetrics.ForwardPostError, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Forward Post Error",
		ServiceName: serviceName,
		ID:          handlerID + ".forward_post_error",
		Unit:        "errors",
		Description: "The Number of times the " + serviceName + " failed to serve a reverse proxy post request to the relay_gateway",
	})
	if err != nil {
		return nil, err
	}

	return m, nil
}
