package config

const (
	FEATURE_BIGTABLE          = 0
	FEATURE_NEW_RELAY_BACKEND = 1
	FEATURE_POSTGRES          = 2
)

type Feature struct {
	Name        string
	Value       bool
	Description string
}

type Config interface {
	FeatureEnabled(enum int) bool
	AllFeatures() []Feature
	FeatureByName(name string) Feature
	AllEnabledFeatures() []Feature
	AllDisabledFeatures() []Feature
}
