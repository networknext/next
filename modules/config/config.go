package config

type FeatureEnum int

const (
	FEATURE_BIGTABLE          FeatureEnum = 0
	FEATURE_NEW_RELAY_BACKEND FeatureEnum = 1
	FEATURE_POSTGRES          FeatureEnum = 2
	FEATURE_VANITY_METRIC     FeatureEnum = 3
)

const NumFeatures = 4

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

var FeatureVantyMetrix = Feature{
	Name:        "FEATURE_VANITY_METRIC",
	Enum:        FEATURE_VANITY_METRIC,
	Value:       false,
	Description: "Vanity metrics for fast aggregate statistic lookup",
}
