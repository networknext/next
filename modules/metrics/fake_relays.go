package metrics

import "context"

// FakeRelayStatus defines the metrics reported by the service's status endpoint
type FakeRelayStatus struct {
	// Service Information
	ServiceName string `json:"service_name"`
	GitHash     string `json:"git_hash"`
	Started     string `json:"started"`
	Uptime      string `json:"uptime"`

	// Service Metrics
	Goroutines      int     `json:"goroutines"`
	MemoryAllocated float64 `json:"mb_allocated"`

	// Invocations
	UpdateInvocations           int `json:"update_invocations"`
	SuccessfulUpdateInvocations int `json:"successful_update_invocations"`

	// Error Metrics
	MarshalBinaryError     int `json:"marshal_binary_error"`
	UnmarshalBinaryError   int `json:"unmarshal_binary_error"`
	UpdatePostError        int `json:"update_post_error"`
	NotOKResponseError     int `json:"not_ok_response_error"`
	ResolveUDPAddressError int `json:"resolve_udp_address_error"`
}

type FakeRelayMetrics struct {
	FakeRelayServiceMetrics     *ServiceMetrics
	UpdateInvocations           Counter
	SuccessfulUpdateInvocations Counter
	ErrorMetrics                FakeRelayErrorMetrics
}

var EmptyFakeRelayMetrics = &FakeRelayMetrics{
	FakeRelayServiceMetrics:     &EmptyServiceMetrics,
	UpdateInvocations:           &EmptyCounter{},
	SuccessfulUpdateInvocations: &EmptyCounter{},
	ErrorMetrics:                *EmptyFakeRelayErrorMetrics,
}

type FakeRelayErrorMetrics struct {
	MarshalBinaryError     Counter
	UnmarshalBinaryError   Counter
	UpdatePostError        Counter
	NotOKResponseError     Counter
	ResolveUDPAddressError Counter
}

var EmptyFakeRelayErrorMetrics = &FakeRelayErrorMetrics{
	MarshalBinaryError:     &EmptyCounter{},
	UnmarshalBinaryError:   &EmptyCounter{},
	UpdatePostError:        &EmptyCounter{},
	NotOKResponseError:     &EmptyCounter{},
	ResolveUDPAddressError: &EmptyCounter{},
}

func NewFakeRelayMetrics(ctx context.Context, metricsHandler Handler, serviceName string, handlerID string, handlerName string, packetDescription string) (*FakeRelayMetrics, error) {
	m := new(FakeRelayMetrics)

	var err error

	m.FakeRelayServiceMetrics, err = NewServiceMetrics(ctx, metricsHandler, serviceName)
	if err != nil {
		return nil, err
	}

	m.UpdateInvocations, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Update Invocations",
		ServiceName: serviceName,
		ID:          handlerID + ".update_invocations",
		Unit:        "invocations",
		Description: "The number of times the " + serviceName + " sent a relay update request to the Relay Gateway",
	})
	if err != nil {
		return nil, err
	}

	m.SuccessfulUpdateInvocations, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Successful Update Invocations",
		ServiceName: serviceName,
		ID:          handlerID + ".successful_update_invocations",
		Unit:        "invocations",
		Description: "The number of times the " + serviceName + " successfully sent a relay update request to the Relay Gateway",
	})
	if err != nil {
		return nil, err
	}

	m.ErrorMetrics.MarshalBinaryError, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Marshal Binary Error",
		ServiceName: serviceName,
		ID:          handlerID + ".marshal_binary_error",
		Unit:        "errors",
		Description: "The number of times the " + serviceName + " failed to marshal the relay update request",
	})
	if err != nil {
		return nil, err
	}

	m.ErrorMetrics.UnmarshalBinaryError, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Unmarshal Binary Error",
		ServiceName: serviceName,
		ID:          handlerID + ".unmarshal_binary_error",
		Unit:        "errors",
		Description: "The number of times the " + serviceName + " failed to unmarshal the relay update response from the Relay Gateway",
	})
	if err != nil {
		return nil, err
	}

	m.ErrorMetrics.UpdatePostError, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Update Post Error",
		ServiceName: serviceName,
		ID:          handlerID + ".update_post_error",
		Unit:        "errors",
		Description: "The number of times the " + serviceName + " failed to HTTP POST to the Relay Gateway",
	})
	if err != nil {
		return nil, err
	}

	m.ErrorMetrics.NotOKResponseError, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Not OK Response Error",
		ServiceName: serviceName,
		ID:          handlerID + ".not_ok_response_error",
		Unit:        "errors",
		Description: "The number of times the " + serviceName + " received a non-200 response from the Relay Gateway",
	})
	if err != nil {
		return nil, err
	}

	m.ErrorMetrics.ResolveUDPAddressError, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Resolve UDP Address Error",
		ServiceName: serviceName,
		ID:          handlerID + ".resolve_udp_address_error",
		Unit:        "errors",
		Description: "The number of times the " + serviceName + " failed to resolve the routing relay's UDP address",
	})
	if err != nil {
		return nil, err
	}

	return m, nil
}
