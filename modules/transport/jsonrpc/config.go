package jsonrpc

import (
	"net/http"

	"github.com/networknext/backend/modules/storage"
)

type ConfigService struct {
	Storage storage.Storer
}

type FeatureFlag struct {
	Name        string `json:"name"`
	Value       bool   `json:"value"`
	Description string `json:"description"`
}

type FeatureFlagArgs struct {
	Name  string `json:"name"`
	Value bool   `json:"value"`
}

type FeatureFlagReply struct {
	Flags map[string]bool `json:"flags"`
}

// These are just stubbed out for the time being until postgres is implemented

func (s *ConfigService) AllFeatureFlags(r *http.Request, args *FeatureFlagArgs, reply *FeatureFlagReply) error {
	return nil
}

func (s *ConfigService) FeatureFlagByName(r *http.Request, args *FeatureFlagArgs, reply *FeatureFlagReply) error {
	return nil
}
