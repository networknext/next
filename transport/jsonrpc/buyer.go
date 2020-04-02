package jsonrpc

import (
	"net/http"
)

type BuyersService struct{}

type MapArgs struct {
	BuyerID uint64 `json:"buyer_id"`
}

type MapReply struct {
	Clusters []cluster `json:"clusters"`
}

type cluster struct {
	Country   string  `json:"country"`
	Region    string  `json:"region"`
	City      string  `json:"city"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Count     int     `json:"count"`
}

func (s *BuyersService) SessionsMap(r *http.Request, args *MapArgs, reply *MapReply) error {
	reply.Clusters = []cluster{
		{Country: "United States", Region: "NY", City: "Troy", Latitude: 42.7273, Longitude: -73.6696, Count: 10},
		{Country: "United States", Region: "NY", City: "Saratoga Springs", Latitude: 43.0034, Longitude: -73.842, Count: 5},
		{Country: "United States", Region: "NY", City: "Albany", Latitude: 42.6701, Longitude: -73.7754, Count: 200},
	}

	return nil
}

type SessionsArgs struct {
	BuyerID uint64 `json:"buyer_id"`
}

type SessionsReply struct {
	Sessions []session `json:"sessions"`
}

type session struct{}

func (s *BuyersService) Sessions(r *http.Request, args *SessionsArgs, reply *SessionsReply) error {
	reply.Sessions = make([]session, 0)

	return nil
}
