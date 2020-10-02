package storage_test

import (
	"testing"

	"github.com/networknext/backend/storage"
	"github.com/stretchr/testify/assert"
	"gopkg.in/auth0.v4/management"
)

func TestLocalJobManager(t *testing.T) {
	manager := storage.LocalJobManager{}
	fail := "FAIL"
	success := "SUCCESS"
	job := management.Job{
		Status: &fail,
	}
	t.Run("failure", func(t *testing.T) {
		err := manager.VerifyEmail(&job)
		assert.Error(t, err)
	})
	job = management.Job{
		Status: &success,
	}
	t.Run("success", func(t *testing.T) {
		err := manager.VerifyEmail(&job)
		assert.NoError(t, err)
	})
}
