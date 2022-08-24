package storage_test

// todo: convert to functional test or remove

/*
import (
	"testing"

	"github.com/networknext/backend/modules/storage"
	"github.com/stretchr/testify/assert"
	"gopkg.in/auth0.v4/management"
)

func TestLocalRoleManager(t *testing.T) {
	manager := storage.NewLocalRoleManager()

	t.Run("list roles - empty", func(t *testing.T) {
		users, err := manager.List()
		assert.NoError(t, err)
		assert.Equal(t, 0, len(users.Roles))
	})

	roleIDs := []string{
		"rol_ScQpWhLvmTKRlqLU",
		"rol_8r0281hf2oC4cvrD",
		"rol_YfFrtom32or4vH89",
	}
	roleNames := []string{
		"Viewer",
		"Owner",
		"Admin",
	}
	roleDescriptions := []string{
		"Can see current sessions and the map.",
		"Can access and manage everything in an account.",
		"Can manage the Network Next system, including access to configstore.",
	}

	storage.AddRolesToRoleList([]*management.Role{
		{
			ID:          &roleIDs[0],
			Name:        &roleNames[0],
			Description: &roleDescriptions[0],
		},
		{
			ID:          &roleIDs[1],
			Name:        &roleNames[1],
			Description: &roleDescriptions[1],
		},
		{
			ID:          &roleIDs[2],
			Name:        &roleNames[2],
			Description: &roleDescriptions[2],
		},
	}, manager)

	t.Run("list roles - not empty", func(t *testing.T) {
		users, err := manager.List()
		assert.NoError(t, err)
		assert.Equal(t, 3, len(users.Roles))
	})
}
*/
