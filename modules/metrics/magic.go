package metrics

import "context"

type MagicStatus struct {
	// Service Information
	ServiceName    string `json:"service_name"`
	GitHash        string `json:"git_hash"`
	Started        string `json:"started"`
	Uptime         string `json:"uptime"`
	OldestInstance bool   `json:"oldest_instance"`

	// Service Metrics
	Goroutines      int     `json:"goroutines"`
	MemoryAllocated float64 `json:"mb_allocated"`

	// Success Metrics
	InsertInstanceMetadataSuccess int `json:"insert_instance_metdata_success"`
	UpdateMagicValuesSuccess      int `json:"update_magic_values_success"`
	GetMagicValueSuccess          int `json:"get_magic_values_success"`
	SetMagicValueSuccess          int `json:"set_magic_values_success"`

	// Error Metrics
	InsertInstanceMetadataFailure int `json:"insert_instance_metadata_failure"`
	UpdateMagicValuesFailure      int `json:"update_magic_values_failure"`
	GetMagicValueFailure          int `json:"get_magic_values_failure"`
	SetMagicValueFailure          int `json:"set_magic_values_failure"`
	ReadFromRedisFailure          int `json:"read_from_redis_failure"`
	MarshalFailure                int `json:"marshal_failure"`
	UnmarshalFailure              int `json:"unmarshal_failure"`
}

type MagicMetrics struct {
	MagicServiceMetrics           *ServiceMetrics
	InsertInstanceMetadataSuccess Counter
	UpdateMagicValuesSuccess      Counter
	GetMagicValueSuccess          Counter
	SetMagicValueSuccess          Counter
	ErrorMetrics                  MagicErrorMetrics
}

var EmptyMagicMetrics = &MagicMetrics{
	MagicServiceMetrics:           &EmptyServiceMetrics,
	InsertInstanceMetadataSuccess: &EmptyCounter{},
	UpdateMagicValuesSuccess:      &EmptyCounter{},
	GetMagicValueSuccess:          &EmptyCounter{},
	SetMagicValueSuccess:          &EmptyCounter{},
	ErrorMetrics:                  EmptyMagicErrorMetrics,
}

type MagicErrorMetrics struct {
	InsertInstanceMetadataFailure Counter
	UpdateMagicValuesFailure      Counter
	GetMagicValueFailure          Counter
	SetMagicValueFailure          Counter
	ReadFromRedisFailure          Counter
	MarshalFailure                Counter
	UnmarshalFailure              Counter
}

var EmptyMagicErrorMetrics = MagicErrorMetrics{
	InsertInstanceMetadataFailure: &EmptyCounter{},
	UpdateMagicValuesFailure:      &EmptyCounter{},
	GetMagicValueFailure:          &EmptyCounter{},
	SetMagicValueFailure:          &EmptyCounter{},
	ReadFromRedisFailure:          &EmptyCounter{},
	MarshalFailure:                &EmptyCounter{},
	UnmarshalFailure:              &EmptyCounter{},
}

func NewMagicBackendMetrics(ctx context.Context, metricsHandler Handler, serviceName string, handlerID string, handlerName string) (*MagicMetrics, error) {
	m := new(MagicMetrics)

	var err error

	m.MagicServiceMetrics, err = NewServiceMetrics(ctx, metricsHandler, serviceName)
	if err != nil {
		return nil, err
	}

	m.InsertInstanceMetadataSuccess, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Insert Instance Metadata Success",
		ServiceName: serviceName,
		ID:          handlerID + ".insert_instance_metadata_success",
		Unit:        "count",
		Description: "The number of times the instance updated its instance metadata in redis",
	})
	if err != nil {
		return nil, err
	}

	m.UpdateMagicValuesSuccess, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Update Magic Values Success",
		ServiceName: serviceName,
		ID:          handlerID + ".update_magic_values_success",
		Unit:        "count",
		Description: "The number of times this instance successfully updated the magic values in redis",
	})
	if err != nil {
		return nil, err
	}

	m.GetMagicValueSuccess, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Get Magic Value Success",
		ServiceName: serviceName,
		ID:          handlerID + ".get_magic_value_success",
		Unit:        "count",
		Description: "The number of times this instance successfully got a magic value from redis",
	})
	if err != nil {
		return nil, err
	}

	m.SetMagicValueSuccess, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Set Magic Value Success",
		ServiceName: serviceName,
		ID:          handlerID + ".set_magic_value_success",
		Unit:        "count",
		Description: "The number of times this instance successfully set a the magic value in redis",
	})
	if err != nil {
		return nil, err
	}

	m.ErrorMetrics.InsertInstanceMetadataFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Insert Instance Metadata Failure",
		ServiceName: serviceName,
		ID:          handlerID + ".insert_instance_metadata_failure",
		Unit:        "count",
		Description: "The number of times the instance failed to update its instance metadata in redis",
	})
	if err != nil {
		return nil, err
	}

	m.ErrorMetrics.UpdateMagicValuesFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Update Magic Values Failure",
		ServiceName: serviceName,
		ID:          handlerID + ".update_magic_values_failure",
		Unit:        "count",
		Description: "The number of times this instance failed to update the magic values in redis",
	})
	if err != nil {
		return nil, err
	}

	m.ErrorMetrics.GetMagicValueFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Get Magic Value Failure",
		ServiceName: serviceName,
		ID:          handlerID + ".get_magic_value_failure",
		Unit:        "count",
		Description: "The number of times this instance failed to get a magic value from redis",
	})
	if err != nil {
		return nil, err
	}

	m.ErrorMetrics.SetMagicValueFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Set Magic Value Failure",
		ServiceName: serviceName,
		ID:          handlerID + ".set_magic_value_failure",
		Unit:        "count",
		Description: "The number of times this instance failed to set a the magic value in redis",
	})
	if err != nil {
		return nil, err
	}

	m.ErrorMetrics.ReadFromRedisFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Read From Redis Failure",
		ServiceName: serviceName,
		ID:          handlerID + ".read_from_redis_failure",
		Unit:        "count",
		Description: "The number of times this instance failed to read from redis",
	})
	if err != nil {
		return nil, err
	}

	m.ErrorMetrics.MarshalFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Marshal Failure",
		ServiceName: serviceName,
		ID:          handlerID + ".marshal_failure",
		Unit:        "count",
		Description: "The number of times this instance failed to marshal structs into json",
	})
	if err != nil {
		return nil, err
	}

	m.ErrorMetrics.UnmarshalFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Unmarshal Failure",
		ServiceName: serviceName,
		ID:          handlerID + ".unmarshal_failure",
		Unit:        "count",
		Description: "The number of times this instance failed to unmarshal values from redis",
	})
	if err != nil {
		return nil, err
	}

	return m, nil
}

func NewMagicFrontendMetrics(ctx context.Context, metricsHandler Handler, serviceName string, handlerID string, handlerName string) (*MagicMetrics, error) {
	m := new(MagicMetrics)

	var err error

	m.MagicServiceMetrics, err = NewServiceMetrics(ctx, metricsHandler, serviceName)
	if err != nil {
		return nil, err
	}

	m.InsertInstanceMetadataSuccess = &EmptyCounter{}
	m.UpdateMagicValuesSuccess = &EmptyCounter{}

	m.GetMagicValueSuccess, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Get Magic Value Success",
		ServiceName: serviceName,
		ID:          handlerID + ".get_magic_value_success",
		Unit:        "count",
		Description: "The number of times this instance successfully got a magic value from redis",
	})
	if err != nil {
		return nil, err
	}

	m.SetMagicValueSuccess = &EmptyCounter{}
	m.ErrorMetrics.InsertInstanceMetadataFailure = &EmptyCounter{}
	m.ErrorMetrics.UpdateMagicValuesFailure = &EmptyCounter{}

	m.ErrorMetrics.GetMagicValueFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Get Magic Value Failure",
		ServiceName: serviceName,
		ID:          handlerID + ".get_magic_value_failure",
		Unit:        "count",
		Description: "The number of times this instance failed to get a magic value from redis",
	})
	if err != nil {
		return nil, err
	}

	m.ErrorMetrics.SetMagicValueFailure = &EmptyCounter{}

	m.ErrorMetrics.ReadFromRedisFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Read From Redis Failure",
		ServiceName: serviceName,
		ID:          handlerID + ".read_from_redis_failure",
		Unit:        "count",
		Description: "The number of times this instance failed to read from redis",
	})
	if err != nil {
		return nil, err
	}

	m.ErrorMetrics.MarshalFailure = &EmptyCounter{}

	m.ErrorMetrics.UnmarshalFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Unmarshal Failure",
		ServiceName: serviceName,
		ID:          handlerID + ".unmarshal_failure",
		Unit:        "count",
		Description: "The number of times this instance failed to unmarshal values from redis",
	})
	if err != nil {
		return nil, err
	}

	return m, nil
}
