package config

import "github.com/networknext/backend/modules/envvar"

type EnvVarConfig struct {
	Features []Feature
}

func NewEnvVarConfig(defaultFeatures []Feature) *EnvVarConfig {
	features := make([]Feature, 0)
	for _, f := range defaultFeatures {
		value, _ := envvar.GetBool(f.Name, false)
		f.Value = value
		features = append(features, f)
	}

	return &EnvVarConfig{
		Features: features,
	}
}

func (e *EnvVarConfig) FeatureEnabled(enum int) bool {
	if enum > len(e.Features) {
		return false
	}
	feature := e.Features[enum]
	return feature.Value
}

func (e *EnvVarConfig) AllFeatures() []Feature {
	return e.Features
}

func (e *EnvVarConfig) FeatureByName(name string) Feature {
	for _, f := range e.Features {
		if f.Name == name {
			return f
		}
	}
	return Feature{}
}

func (e *EnvVarConfig) AllEnabledFeatures() []Feature {
	features := make([]Feature, 0)
	for _, f := range e.Features {
		if f.Value == true {
			features = append(features, f)
		}
	}
	return features
}

func (e *EnvVarConfig) AllDisabledFeatures() []Feature {
	features := make([]Feature, 0)
	for _, f := range e.Features {
		if f.Value == false {
			features = append(features, f)
		}
	}
	return features
}
