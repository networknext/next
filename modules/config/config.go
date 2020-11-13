package config

type FeatureEnum string

const (
	FEATURE_BIGTABLE          FeatureEnum = "FEATURE_BIGTABLE"
	FEATURE_NEW_RELAY_BACKEND FeatureEnum = "FEATURE_NEW_RELAY_BACKEND"
	FEATURE_POSTGRES          FeatureEnum = "FEATURE_POSTGRES"
)

type Config interface {
	FeatureEnabled(enum FeatureEnum) bool
	AllFeatures() map[FeatureEnum]bool
	AllEnabledFeatures() map[FeatureEnum]bool
	AllDisabledFeatures() map[FeatureEnum]bool
}

func (e FeatureEnum) ToString() string {
	return string(e)
}
