package storage

import (
	"context"

	"github.com/go-kit/kit/log"
	"gopkg.in/auth0.v4/management"
)

type Auth0 struct {
	Manager *management.Management
	Logger  log.Logger
}

func NewAuth0Manager(ctx context.Context, logger log.Logger) (*Auth0, error) {
	manager, err := management.New(
		"networknext.auth0.com",
		"NIwrWYmG9U3tCQP6QxJqCx8n2xGSTCvf",
		"GZ9l7xF0dggtvz-jxbG7_-yX2YlvkGas4sIq2RJK4glxkHvT0t-WwMtyJlP5qix0",
	)
	if err != nil {
		return nil, err
	}

	return &Auth0{
		Manager: manager,
		Logger:  logger,
	}, nil
}
