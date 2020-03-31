package jsonrpc

import (
	"errors"
	"net/http"

	"github.com/go-redis/redis/v7"
	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
)

type OpsService struct {
	RedisClient redis.Cmdable
	Storage     storage.Storer
}

type RelaysArgs struct {
	Addr string `json:"addr"`
}

type RelaysReply struct {
	Relays []routing.Relay
}

func (s *OpsService) Relays(r *http.Request, args *RelaysArgs, reply *RelaysReply) error {
	if args.Addr != "" {
		relay, ok := s.Storage.Relay(crypto.HashID(args.Addr))
		if !ok {
			return errors.New("not found")
		}

		reply.Relays = []routing.Relay{*relay}

		return nil
	}

	reply.Relays = s.Storage.Relays()

	return nil
}

type CreateRelayArgs struct {
	Addr       string `json:"addr"`
	PublicKey  string `json:"public_key"`
	Datacenter string `json:"datacenter"`
}

type CreateRelayReply struct{}

func (s *OpsService) CreateRelay(r *http.Request, args *RelaysArgs, reply *RelaysReply) error {
	if args.Addr != "" {
		relay, ok := s.Storage.Relay(crypto.HashID(args.Addr))
		if !ok {
			return errors.New("not found")
		}

		reply.Relays = []routing.Relay{*relay}

		return nil
	}

	reply.Relays = s.Storage.Relays()

	return nil
}
