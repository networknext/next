package jsonrpc

import (
	"net/http"

	"github.com/go-kit/kit/log"

	"github.com/networknext/backend/storage"
)

type ConfigService struct {
	Storage storage.Storer
	Logger  log.Logger
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
	reply.Flags = map[string]bool{
		"FEATURE_EXPLORE": false,
	}
	return nil
}

func (s *ConfigService) FeatureFlagByName(r *http.Request, args *FeatureFlagArgs, reply *FeatureFlagReply) error {
	reply.Flags = map[string]bool{
		"FEATURE_EXPLORE": false,
	}
	return nil
}

func (s *ConfigService) FeatureFlagsByValue(r *http.Request, args *FeatureFlagArgs, reply *FeatureFlagReply) error {
	reply.Flags = map[string]bool{
		"FEATURE_EXPLORE": false,
	}
	return nil
}
