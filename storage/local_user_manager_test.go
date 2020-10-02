package storage_test

import (
	"testing"

	"github.com/networknext/backend/storage"
	"github.com/stretchr/testify/assert"
	"gopkg.in/auth0.v4/management"
)

func TestLocalUserManager(t *testing.T) {
	manager := storage.NewLocalUserManager()
	IDs := []string{
		"123",
		"456",
		"789",
	}
	emails := []string{
		"test@test.test",
	}
	roleNames := []string{
		"rol_ScQpWhLvmTKRlqLU",
		"rol_8r0281hf2oC4cvrD",
		"rol_YfFrtom32or4vH89",
	}
	roleTypes := []string{
		"Viewer",
		"Owner",
		"Admin",
	}
	roleDescriptions := []string{
		"Can see current sessions and the map.",
		"Can access and manage everything in an account.",
		"Can manage the Network Next system, including access to configstore.",
	}
	t.Run("add user - success", func(t *testing.T) {
		err := manager.Create(&management.User{
			ID: &IDs[0],
		})
		assert.NoError(t, err)
	})
	t.Run("add user - failure", func(t *testing.T) {
		err := manager.Create(&management.User{
			ID: &IDs[0],
		})
		assert.Error(t, err)
	})
	t.Run("update user - success", func(t *testing.T) {
		err := manager.Update(IDs[0], &management.User{
			ID:    &IDs[0],
			Email: &emails[0],
		})
		assert.NoError(t, err)

		user, err := manager.Read(IDs[0])
		assert.NoError(t, err)
		assert.Equal(t, &emails[0], user.Email)
	})
	t.Run("update user - failure", func(t *testing.T) {
		err := manager.Update(IDs[1], &management.User{
			Email: &emails[0],
		})
		assert.Error(t, err)
	})
	t.Run("read user - success", func(t *testing.T) {
		user, err := manager.Read(IDs[0])
		assert.NoError(t, err)
		assert.Equal(t, &emails[0], user.Email)
	})
	t.Run("read user - failure", func(t *testing.T) {
		_, err := manager.Read(IDs[1])
		assert.Error(t, err)
	})
	t.Run("list users", func(t *testing.T) {
		users, err := manager.List()
		assert.NoError(t, err)
		assert.Equal(t, 1, len(users.Users))
		assert.Equal(t, &IDs[0], users.Users[0].ID)
	})
	t.Run("assign roles", func(t *testing.T) {
		err := manager.AssignRoles(IDs[0], []*management.Role{
			{
				ID:          &roleTypes[0],
				Name:        &roleNames[0],
				Description: &roleDescriptions[0],
			},
		}...)
		assert.NoError(t, err)
	})
	t.Run("read roles", func(t *testing.T) {
		roles, err := manager.Roles(IDs[0])
		assert.NoError(t, err)
		assert.Equal(t, 1, len(roles.Roles))
		assert.Equal(t, roleTypes[0], *roles.Roles[0].ID)
		assert.Equal(t, roleNames[0], *roles.Roles[0].Name)
		assert.Equal(t, roleDescriptions[0], *roles.Roles[0].Description)
	})
	t.Run("remove roles - failure", func(t *testing.T) {
		err := manager.RemoveRoles(IDs[1], []*management.Role{
			{
				ID:          &roleTypes[0],
				Name:        &roleNames[0],
				Description: &roleDescriptions[0],
			},
		}...)
		assert.Error(t, err)
	})
	t.Run("remove roles - 0 roles", func(t *testing.T) {
		err := manager.RemoveRoles(IDs[0], []*management.Role{
			{
				ID:          &roleTypes[1],
				Name:        &roleNames[1],
				Description: &roleDescriptions[1],
			},
		}...)
		assert.NoError(t, err)
		roles, err := manager.Roles(IDs[0])
		assert.NoError(t, err)
		assert.Equal(t, 1, len(roles.Roles))
		assert.Equal(t, roleTypes[0], *roles.Roles[0].ID)
		assert.Equal(t, roleNames[0], *roles.Roles[0].Name)
		assert.Equal(t, roleDescriptions[0], *roles.Roles[0].Description)
	})
	t.Run("remove roles - 1 role", func(t *testing.T) {
		err := manager.RemoveRoles(IDs[0], []*management.Role{
			{
				ID:          &roleTypes[0],
				Name:        &roleNames[0],
				Description: &roleDescriptions[0],
			},
		}...)
		assert.NoError(t, err)
		roles, err := manager.Roles(IDs[0])
		assert.NoError(t, err)
		assert.Equal(t, 0, len(roles.Roles))
	})
	t.Run("delete user - success", func(t *testing.T) {
		err := manager.Delete(IDs[0])
		assert.NoError(t, err)

		users, err := manager.List()
		assert.NoError(t, err)
		assert.Equal(t, 0, len(users.Users))
	})
	t.Run("delete user - failure", func(t *testing.T) {
		err := manager.Delete(IDs[0])
		assert.Error(t, err)
	})
}
