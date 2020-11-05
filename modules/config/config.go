package config

type Feature struct {
	Name        string
	Value       bool
	Description string
}

var Features []Feature = []Feature{
	{
		Name:        "ENABLE_BIGTABLE",
		Value:       false,
		Description: "Bigtable integration for historic session data",
	},
	{
		Name:        "FEATURE_NEW_RELAY_BACKEND",
		Value:       false,
		Description: "New relay backend architectural changes",
	},
	{
		Name:        "PGSQL",
		Value:       false,
		Description: "Postgres implementation to replace Firestore",
	},
}

type Config interface {
	FeatureEnabled(name string) bool
	AllFeatures() []Feature
	FeatureByName(name string) Feature
	AllEnabledFeatures() []Feature
	AllDisabledFeatures() []Feature
}
