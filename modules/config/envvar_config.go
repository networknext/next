package config

import "github.com/networknext/backend/modules/envvar"

type EnvVarConfig struct {
	Features map[FeatureEnum]bool
}

func NewEnvVarConfig(defaultFeatures map[FeatureEnum]bool) *EnvVarConfig {
	features := defaultFeatures
	for key, value := range defaultFeatures {
		newValue, _ := envvar.GetBool(key.ToString(), value)
		features[key] = newValue
	}

	return &EnvVarConfig{
		Features: features,
	}
}

func (e *EnvVarConfig) FeatureEnabled(enum FeatureEnum) bool {
	return e.Features[enum]
}

func (e *EnvVarConfig) AllFeatures() map[FeatureEnum]bool {
	return e.Features
}

func (e *EnvVarConfig) AllEnabledFeatures() map[FeatureEnum]bool {
	features := make(map[FeatureEnum]bool, 0)
	for key, value := range e.Features {
		if value == true {
			features[key] = value
		}
	}
	return features
}

func (e *EnvVarConfig) AllDisabledFeatures() map[FeatureEnum]bool {
	features := make(map[FeatureEnum]bool, 0)
	for key, value := range e.Features {
		if value == false {
			features[key] = value
		}
	}
	return features
}
