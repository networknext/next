package storage

import (
	"fmt"
	"sync"

	"gopkg.in/auth0.v4/management"
)

type LocalUserManager struct {
	localRoles map[string][]*management.Role
	localUsers []*management.User
	rolesMutex sync.RWMutex
}

func NewLocalUserManager() *LocalUserManager {
	return &LocalUserManager{
		localRoles: make(map[string][]*management.Role),
	}
}

func (ljm *LocalUserManager) Create(user *management.User) error {
	for _, u := range ljm.localUsers {
		if u.ID == user.ID {
			return &AlreadyExistsError{resourceType: "user", resourceRef: user.ID}
		}
	}

	ljm.localUsers = append(ljm.localUsers, user)
	return nil
}
func (ljm *LocalUserManager) Delete(id string) error {
	userIndex := -1
	for i, user := range ljm.localUsers {
		if *user.ID == id {
			userIndex = i
		}
	}

	if userIndex < 0 {
		return &DoesNotExistError{resourceType: "user", resourceRef: id}
	}

	if userIndex+1 == len(ljm.localUsers) {
		ljm.localUsers = ljm.localUsers[:userIndex]
		return nil
	}

	frontSlice := ljm.localUsers[:userIndex]
	backSlice := ljm.localUsers[userIndex+1:]
	ljm.localUsers = append(frontSlice, backSlice...)
	return nil
}
func (ljm *LocalUserManager) List(opts ...management.ListOption) (*management.UserList, error) {
	var userList management.UserList
	users := make([]*management.User, len(ljm.localUsers))
	for i := range users {
		users[i] = ljm.localUsers[i]
	}

	userList.Users = users

	return &userList, nil
}
func (ljm *LocalUserManager) Read(id string) (*management.User, error) {
	for _, user := range ljm.localUsers {
		if *user.ID == id {
			return user, nil
		}
	}

	return &management.User{}, &DoesNotExistError{resourceType: "user", resourceRef: id}
}
func (ljm *LocalUserManager) AssignRoles(id string, roles ...*management.Role) error {
	ljm.rolesMutex.Lock()
	ljm.localRoles[id] = roles
	ljm.rolesMutex.Unlock()
	return nil
}
func (ljm *LocalUserManager) RemoveRoles(id string, roles ...*management.Role) error {
	ljm.rolesMutex.RLock()
	oldRoles, ok := ljm.localRoles[id]
	ljm.rolesMutex.RUnlock()

	if !ok {
		return &DoesNotExistError{resourceType: "user", resourceRef: id}
	}

	newRoles := make([]*management.Role, 0)

	for _, r := range roles {
		for _, o := range oldRoles {
			if *o.ID != *r.ID {
				newRoles = append(newRoles, o)
			}
		}
	}

	fmt.Println(newRoles)

	ljm.rolesMutex.Lock()
	ljm.localRoles[id] = newRoles
	ljm.rolesMutex.Unlock()
	return nil
}

func (ljm *LocalUserManager) Roles(id string, opts ...management.ListOption) (*management.RoleList, error) {
	ljm.rolesMutex.RLock()
	oldRoles, ok := ljm.localRoles[id]
	ljm.rolesMutex.RUnlock()

	if !ok {
		return &management.RoleList{}, &DoesNotExistError{resourceType: "user", resourceRef: id}
	}
	var roleList management.RoleList
	roleList.Roles = oldRoles
	return &roleList, nil
}

func (ljm *LocalUserManager) Update(id string, u *management.User) error {
	for i := range ljm.localUsers {
		if *ljm.localUsers[i].ID == id {
			ljm.localUsers[i] = u
			return nil
		}
	}

	return &DoesNotExistError{resourceType: "user", resourceRef: id}
}
