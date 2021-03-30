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
				// values different to show override working
				Name:        "FEATURE_BIG_TABLE",
				Enum:        config.FEATURE_BIGTABLE,
				Value:       true,
				Description: "Bigtable integration for historic session data with override",
			},
			{
				Name:        "FEATURE_NEW_RELAY_BACKEND",
				Enum:        config.FEATURE_NEW_RELAY_BACKEND,
				Value:       false,
				Description: "New relay backend architectural changes",
			},
			{
				Name:        "FEATURE_POSTGRES",
				Enum:        config.FEATURE_POSTGRES,
				Value:       false,
				Description: "Postgres implementation to replace Firestore",
			},
		})

		featureConfig = envVarConfig

		assert.NotNil(t, featureConfig.AllFeatures())
		// check bigtable
		assert.Equal(t, featureConfig.AllFeatures()[0].Name, "FEATURE_BIG_TABLE")
		assert.Equal(t, featureConfig.AllFeatures()[0].Enum, config.FEATURE_BIGTABLE)
		assert.Equal(t, featureConfig.AllFeatures()[0].Value, true)
		assert.Equal(t, featureConfig.AllFeatures()[0].Description, "Bigtable integration for historic session data with override")
		assert.Equal(t, featureConfig.AllFeatures()[0].Value, featureConfig.FeatureEnabled(config.FEATURE_BIGTABLE))
		assert.True(t, featureConfig.FeatureEnabled(config.FEATURE_BIGTABLE))

		// check vanity
		assert.Equal(t, featureConfig.AllFeatures()[3].Name, "FEATURE_VANITY_METRIC")
		assert.Equal(t, featureConfig.AllFeatures()[3].Enum, config.FEATURE_VANITY_METRIC)
		assert.Equal(t, featureConfig.AllFeatures()[3].Value, false)
		assert.Equal(t, featureConfig.AllFeatures()[3].Description, "Vanity metrics for fast aggregate statistic lookup")
		assert.Equal(t, featureConfig.AllFeatures()[3].Value, featureConfig.FeatureEnabled(config.FEATURE_VANITY_METRIC))
		assert.False(t, featureConfig.FeatureEnabled(config.FEATURE_VANITY_METRIC))
	})
}
