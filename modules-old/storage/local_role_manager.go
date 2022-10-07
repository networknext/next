package storage

import (
	"sync"

	"gopkg.in/auth0.v4/management"
)

type LocalRoleManager struct {
	localRoleList *management.RoleList
	rolesMutex    sync.RWMutex
}

func NewLocalRoleManager() *LocalRoleManager {
	return &LocalRoleManager{
		localRoleList: &management.RoleList{
			Roles: make([]*management.Role, 0),
		},
	}
}

func (lrm *LocalRoleManager) List(opts ...management.ListOption) (r *management.RoleList, err error) {
	lrm.rolesMutex.Lock()
	defer lrm.rolesMutex.Unlock()

	return lrm.localRoleList, nil
}

// Function just for testing (Not part of the Auth0 Role manager interface) - TBD for future use...
func AddRolesToRoleList(roles []*management.Role, manager *LocalRoleManager) {
	manager.localRoleList.Roles = roles
}
