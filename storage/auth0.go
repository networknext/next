package storage

import (
	"github.com/go-kit/kit/log"
	"gopkg.in/auth0.v4/management"
)

type Auth0 struct {
	Manager *management.Management
	Logger  log.Logger
}
