package storage_test

import (
	"testing"

	"github.com/networknext/backend/storage"
	"github.com/stretchr/testify/assert"
	"gopkg.in/auth0.v4/management"
)

func TestLocalJobManager(t *testing.T) {
	manager := storage.LocalJobManager{}
	emptyID := ""
	job := management.Job{
		UserID: &emptyID,
	}

	t.Run("failure - no user ID", func(t *testing.T) {
		err := manager.VerifyEmail(&job)
		assert.Error(t, err)
	})

	*job.UserID = "123"

	t.Run("success", func(t *testing.T) {
		err := manager.VerifyEmail(&job)
		assert.NoError(t, err)
	})
}
