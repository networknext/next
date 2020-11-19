package storage

import (
	"gopkg.in/auth0.v4/management"
)

type UserManager interface {
	AssignRoles(id string, roles ...*management.Role) error
	Create(u *management.User) error
	Delete(id string) error
	List(opts ...management.ListOption) (*management.UserList, error)
	Read(id string) (*management.User, error)
	RemoveRoles(id string, roles ...*management.Role) error
	Roles(id string, opts ...management.ListOption) (*management.RoleList, error)
	Update(id string, u *management.User) error
}

type JobManager interface {
	VerifyEmail(j *management.Job) error
}
