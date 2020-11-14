package config_test

import (
	"testing"

	"github.com/networknext/backend/modules/config"
	"github.com/stretchr/testify/assert"
)

func TestConfigInterface(t *testing.T) {
	t.Run("NewConfig", func(t *testing.T) {
		var featureConfig config.Config

		envVarConfig := config.NewEnvVarConfig([]config.Feature{
			{
				Name:        "FEATURE_BIGTABLE",
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
		})

		featureConfig = envVarConfig

		assert.NotNil(t, featureConfig.AllFeatures())
		assert.Equal(t, len(featureConfig.AllFeatures()), 3)
		assert.Equal(t, len(featureConfig.AllDisabledFeatures()), 3)
		assert.Equal(t, len(featureConfig.AllEnabledFeatures()), 0)
		assert.Equal(t, featureConfig.AllFeatures()[0].Name, "FEATURE_BIGTABLE")
		assert.Equal(t, featureConfig.AllFeatures()[0].Value, false)
		assert.Equal(t, featureConfig.AllFeatures()[0].Description, "Bigtable integration for historic session data")
		assert.Equal(t, featureConfig.AllFeatures()[0].Value, featureConfig.FeatureEnabled(config.FEATURE_BIGTABLE))
	})
}
