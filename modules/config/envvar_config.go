package config

import "github.com/networknext/backend/modules/envvar"

type EnvVarConfig struct {
	Features []Feature
	LookUp   []bool
}

func NewEnvVarConfig(defaultFeatures []Feature) *EnvVarConfig {
	lookUp := make([]bool, NumFeatures)
	features := defaultFeatures
	for _, f := range defaultFeatures {
		newValue, _ := envvar.GetBool(f.Name, f.Value)
		f.Value = newValue
		lookUp[f.Enum] = newValue
	}

	return &EnvVarConfig{
		Features: features,
		LookUp:   lookUp,
	}
}

func (e *EnvVarConfig) FeatureEnabled(enum FeatureEnum) bool {
	return e.LookUp[enum]
}

func (e *EnvVarConfig) AllFeatures() []Feature {
	return e.Features
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
