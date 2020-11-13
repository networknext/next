package config

type FeatureEnum int

const (
	FEATURE_BIGTABLE          FeatureEnum = 0
	FEATURE_NEW_RELAY_BACKEND FeatureEnum = 1
	FEATURE_POSTGRES          FeatureEnum = 2
)

const NumFeatures = 3

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
