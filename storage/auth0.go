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
	manager, err := management.New("https://networknext.auth0.com/oauth/token", "8nDIPIrV9NAQHcYSGhH7BjXkGSyzB6Ja", "7E511bx4Pyk7_0fbCsX12hAoGCht1RHgS3LGO30oscxIwtVsz7GclWAR7cu7YlTy")
	if err != nil {
		return nil, err
	}

	return &Auth0{
		Manager: manager,
		Logger:  logger,
	}, nil
}
