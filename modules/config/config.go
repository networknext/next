package config

type FeatureEnum string

const (
	FEATURE_BIGTABLE          FeatureEnum = "FEATURE_BIGTABLE"
	FEATURE_NEW_RELAY_BACKEND FeatureEnum = "FEATURE_NEW_RELAY_BACKEND"
	FEATURE_POSTGRES          FeatureEnum = "FEATURE_POSTGRES"
)

type Feature struct {
	Name        string
	Value       bool
	Description string
}

type Config interface {
	FeatureEnabled(enum FeatureEnum) bool
	AllFeatures() []Feature
	FeatureByName(name string) Feature
	AllEnabledFeatures() []Feature
	AllDisabledFeatures() []Feature
}

func (e FeatureEnum) ToString() string {
	return string(e)
}
