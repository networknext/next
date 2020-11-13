package config_test

import (
	"testing"

	"github.com/networknext/backend/modules/config"
	"github.com/stretchr/testify/assert"
)

func TestConfigInterface(t *testing.T) {
	t.Run("NewConfig", func(t *testing.T) {
		var featureConfig config.Config

		envVarConfig := config.NewEnvVarConfig(map[config.FeatureEnum]bool{
			config.FEATURE_BIGTABLE:          false,
			config.FEATURE_NEW_RELAY_BACKEND: false,
			config.FEATURE_POSTGRES:          false,
		})

		featureConfig = envVarConfig

		assert.NotNil(t, featureConfig.AllFeatures())
		assert.Equal(t, len(featureConfig.AllFeatures()), 3)
		assert.Equal(t, len(featureConfig.AllDisabledFeatures()), 3)
		assert.Equal(t, len(featureConfig.AllEnabledFeatures()), 0)
		assert.Equal(t, featureConfig.AllFeatures()[config.FEATURE_BIGTABLE], false)
	})
}
