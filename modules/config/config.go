package config

type FeatureEnum int

const (
	FEATURE_BIGTABLE             FeatureEnum = 0
	FEATURE_NEW_RELAY_BACKEND    FeatureEnum = 1
	FEATURE_POSTGRES             FeatureEnum = 2
	FEATURE_VANITY_METRIC        FeatureEnum = 3
	FEATURE_LOAD_TEST            FeatureEnum = 4
	FEATURE_ENABLE_PPROF         FeatureEnum = 5
	FEATURE_ROUTE_MATRIX_STATS   FeatureEnum = 6
	FEATURE_MATRIX_CLOUDSTORE    FeatureEnum = 7
	FEATURE_VALVE_MATRIX         FeatureEnum = 8
	FEATURE_BILLING              FeatureEnum = 9
	FEATURE_BILLING2             FeatureEnum = 10
	FEATURE_RELAY_FULL_BANDWIDTH FeatureEnum = 11
)

// NumFeatures is always one more than the highest FeatureEnum
var NumFeatures = 12

type Feature struct {
	Name        string
	Enum        FeatureEnum
	Value       bool
	Description string
}

type Config interface {
	FeatureEnabled(enum FeatureEnum) bool
	AllFeatures() []Feature
	AllEnabledFeatures() []Feature
	AllDisabledFeatures() []Feature
}

var defaultFeatures = []Feature{
	{
		Name:        "FEATURE_BIGTABLE",
		Enum:        FEATURE_BIGTABLE,
		Value:       false,
		Description: "Bigtable integration for historic session data",
	},
	{
		Name:        "FEATURE_NEW_RELAY_BACKEND_ENABLED",
		Enum:        FEATURE_NEW_RELAY_BACKEND,
		Value:       false,
		Description: "Enables the New Relay Backend project if true",
	},
	{
		Name:        "FEATURE_POSTGRES",
		Enum:        FEATURE_POSTGRES,
		Value:       false,
		Description: "Postgres implementation to replace Firestore",
	},
	{
		Name:        "FEATURE_VANITY_METRIC",
		Enum:        FEATURE_VANITY_METRIC,
		Value:       false,
		Description: "Vanity metrics for fast aggregate statistic lookup",
	},
	{
		Name:        "FEATURE_LOAD_TEST",
		Enum:        FEATURE_LOAD_TEST,
		Value:       false,
		Description: "Disables pubsub and storer usage when true",
	},
	{
		Name:        "FEATURE_ENABLE_PPROF",
		Enum:        FEATURE_ENABLE_PPROF,
		Value:       false,
		Description: "Allows access to PPROF http handlers when true",
	},
	{
		Name:        "FEATURE_ROUTE_MATRIX_STATS",
		Enum:        FEATURE_ROUTE_MATRIX_STATS,
		Value:       false,
		Description: "Writes Route Matrix Stats to pubsub when true",
	},
	{
		Name:        "FEATURE_MATRIX_CLOUDSTORE",
		Enum:        FEATURE_MATRIX_CLOUDSTORE,
		Value:       false,
		Description: "Writes Route Matrix to cloudstore when true",
	},
	{
		Name:        "FEATURE_VALVE_MATRIX",
		Enum:        FEATURE_VALVE_MATRIX,
		Value:       false,
		Description: "Creates the valve matrix when true",
	},
	{
		Name:        "FEATURE_BILLING",
		Enum:        FEATURE_BILLING,
		Value:       false,
		Description: "Inserts and writes BillingEntry to BigQuery",
	},
	{
		Name:        "FEATURE_BILLING2",
		Enum:        FEATURE_BILLING2,
		Value:       true,
		Description: "Inserts and writes BillingEntry2 to BigQuery",
	},
	{
		Name:        "FEATURE_RELAY_FULL_BANDWIDTH",
		Enum:        FEATURE_RELAY_FULL_BANDWIDTH,
		Value:       false,
		Description: "Consider a relay as full based on its bandwidth usage",
	},
}
