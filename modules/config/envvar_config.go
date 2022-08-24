package config

import "github.com/networknext/backend/modules/envvar"

type EnvVarConfig struct {
	Features []Feature
}

func NewEnvVarConfig(overrideFeatures []Feature) *EnvVarConfig {
	features := make([]Feature, NumFeatures)
	for _, f := range defaultFeatures {
		newValue := envvar.GetBool(f.Name, f.Value)
		f.Value = newValue
		features[f.Enum] = f
	}

	for _, f := range overrideFeatures {
		newValue := envvar.GetBool(f.Name, f.Value)
		f.Value = newValue
		features[f.Enum] = f
	}

	return &EnvVarConfig{
		Features: features,
	}
}

func (e *EnvVarConfig) FeatureEnabled(enum FeatureEnum) bool {
	return e.Features[enum].Value
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
