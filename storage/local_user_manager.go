package storage

import (
	"fmt"
	"math/rand"
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

func (lum *LocalUserManager) Create(user *management.User) error {
	emptyName := ""
	if user.ID == nil {
		newID := fmt.Sprintf("%d", rand.Intn(10000))
		user.ID = &newID
		user.Identities = []*management.UserIdentity{
			{
				UserID: &newID,
			},
		}
		user.Name = &emptyName
	}
	for _, u := range lum.localUsers {
		if u.ID == user.ID {
			return &AlreadyExistsError{resourceType: "user", resourceRef: user.ID}
		}
	}

	lum.localUsers = append(lum.localUsers, user)
	return nil
}
func (lum *LocalUserManager) Delete(id string) error {
	userIndex := -1
	for i, user := range lum.localUsers {
		if *user.ID == id {
			userIndex = i
		}
	}

	if userIndex < 0 {
		return &DoesNotExistError{resourceType: "user", resourceRef: id}
	}

	if userIndex+1 == len(lum.localUsers) {
		lum.localUsers = lum.localUsers[:userIndex]
		return nil
	}

	frontSlice := lum.localUsers[:userIndex]
	backSlice := lum.localUsers[userIndex+1:]
	lum.localUsers = append(frontSlice, backSlice...)
	return nil
}
func (lum *LocalUserManager) List(opts ...management.ListOption) (*management.UserList, error) {
	var userList management.UserList
	users := make([]*management.User, len(lum.localUsers))
	for i := range users {
		users[i] = lum.localUsers[i]
	}

	userList.Users = users

	return &userList, nil
}
func (lum *LocalUserManager) Read(id string) (*management.User, error) {
	for _, user := range lum.localUsers {
		if *user.ID == id {
			return user, nil
		}
	}

	return &management.User{}, &DoesNotExistError{resourceType: "user", resourceRef: id}
}
func (lum *LocalUserManager) AssignRoles(id string, roles ...*management.Role) error {
	lum.rolesMutex.Lock()
	lum.localRoles[id] = roles
	lum.rolesMutex.Unlock()
	return nil
}
func (lum *LocalUserManager) RemoveRoles(id string, roles ...*management.Role) error {
	lum.rolesMutex.RLock()
	oldRoles, ok := lum.localRoles[id]
	lum.rolesMutex.RUnlock()

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

	lum.rolesMutex.Lock()
	lum.localRoles[id] = newRoles
	lum.rolesMutex.Unlock()
	return nil
}

func (lum *LocalUserManager) Roles(id string, opts ...management.ListOption) (*management.RoleList, error) {
	lum.rolesMutex.RLock()
	oldRoles, ok := lum.localRoles[id]
	lum.rolesMutex.RUnlock()

	if !ok {
		return &management.RoleList{}, &DoesNotExistError{resourceType: "user", resourceRef: id}
	}
	var roleList management.RoleList
	roleList.Roles = oldRoles
	return &roleList, nil
}

func (lum *LocalUserManager) Update(id string, u *management.User) error {
	for i := range lum.localUsers {
		if *lum.localUsers[i].ID == id {
			lum.localUsers[i] = u
			return nil
		}
	}

	return &DoesNotExistError{resourceType: "user", resourceRef: id}
}
