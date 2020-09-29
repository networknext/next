package admin

import (
	"errors"
)

var (
	ErrInsufficientPrivileges = errors.New("insufficient privileges")
)

const (
	TopSessionsSize = 1000
)
